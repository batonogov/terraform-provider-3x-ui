package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// defaultRetryBackoff is the fixed delay between retry attempts on transient
// 5xx responses. It is intentionally short — a longer wait would mask real
// bugs, since this retry exists only to absorb sub-second contention spikes
// in 3x-ui's SQLite write path on older versions.
const defaultRetryBackoff = 500 * time.Millisecond

// readAfterWriteAttempts is the maximum number of times a post-write read
// will be retried while the just-written row is not yet visible. 3x-ui
// occasionally returns success from add/update endpoints before the SQLite
// commit becomes visible to a follow-up GET — see issue #157. The matrix
// acceptance test sustains write pressure for tens of seconds, and CI
// runners are slower than local hardware, so the budget is sized for the
// worst-case lag observed in CI (~10s — local lag is sub-second).
const readAfterWriteAttempts = 20

// readAfterWriteBackoff is the delay between read-after-write retry attempts.
// Same rationale as defaultRetryBackoff: long enough to absorb a typical
// SQLite contention spike, short enough not to mask real bugs.
const readAfterWriteBackoff = 500 * time.Millisecond

const csrfHeaderName = "X-CSRF-Token"

type ClientConfig struct {
	Endpoint           string
	BasePath           string
	Username           string
	Password           string
	TwoFactorCode      string
	InsecureSkipVerify bool
	Timeout            time.Duration
	// MaxRetries is the maximum number of *additional* attempts on transient
	// 5xx for idempotent write endpoints. 0 disables retry entirely; 1 (the
	// default applied by the provider) means up to one retry after a backoff.
	MaxRetries int
	// The fields below override production retry/backoff timings for tests.
	// Zero means use the production default.
	RetryBackoff           time.Duration // default: 500ms
	ReadAfterWriteAttempts int           // default: 20
	ReadAfterWriteBackoff  time.Duration // default: 500ms
	VersionRetryAttempts   int           // default: 4
	VersionRetryBackoff    time.Duration // default: 2s
}

type Client struct {
	baseURL                 *url.URL
	basePath                string
	username                string
	password                string
	twoFactor               string
	httpClient              *http.Client
	maxRetries              int
	retryBackoff            time.Duration
	rawAttempts             int
	rawBackoff              time.Duration
	versionRetryAttempts    int
	versionRetryBaseBackoff time.Duration
	authMu                  sync.Mutex
	csrfToken               string
	settingsSecretMu        sync.Mutex
	settingsSecrets         map[string]string
	newClientMu             sync.Mutex
	newClientAPI            *bool // nil=undetected, true=v3.1.0+ /panel/api/clients/*, false=old
	settingsAPIMu           sync.Mutex
	settingsUnderAPI        *bool // nil=undetected, true=v3.3.0+ /panel/api/setting/*, false=old /panel/setting/*
}

// SetBasePath updates the client's base path to match a new webBasePath.
func (c *Client) SetBasePath(p string) {
	c.basePath = normalizeBasePath(p)
}

// ReadAfterWriteConfig returns the read-after-write retry budget (attempts and
// backoff) so callers like waitForInboundDeletion can align their own polling
// loops with the same budget without hard-coding the constants.
func (c *Client) ReadAfterWriteConfig() (attempts int, backoff time.Duration) {
	return c.rawAttempts, c.rawBackoff
}

type apiResponse struct {
	Success bool            `json:"success"`
	Msg     string          `json:"msg"`
	Obj     json.RawMessage `json:"obj"`
}

type loginFailedError struct {
	message string
}

func (e *loginFailedError) Error() string {
	return e.message
}

func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.Endpoint == "" {
		return nil, errors.New("endpoint is required")
	}
	baseURL, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, err
	}
	if baseURL.Scheme == "" {
		return nil, errors.New("endpoint must include scheme (http or https)")
	}

	basePath := normalizeBasePath(cfg.BasePath)

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	// #nosec G402 -- InsecureSkipVerify is intentional: the provider manages
	// self-hosted panels that frequently use self-signed certificates. The
	// user explicitly opts in via the insecure_skip_verify provider attribute.
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: cfg.InsecureSkipVerify}

	client := &http.Client{
		Jar:       jar,
		Timeout:   cfg.Timeout,
		Transport: transport,
	}

	maxRetries := cfg.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	rb := cfg.RetryBackoff
	if rb == 0 {
		rb = defaultRetryBackoff
	}
	ra := cfg.ReadAfterWriteAttempts
	if ra == 0 {
		ra = readAfterWriteAttempts
	}
	raBackoff := cfg.ReadAfterWriteBackoff
	if raBackoff == 0 {
		raBackoff = readAfterWriteBackoff
	}
	vAttempts := cfg.VersionRetryAttempts
	if vAttempts == 0 {
		vAttempts = versionRetryAttempts
	}
	vBackoff := cfg.VersionRetryBackoff
	if vBackoff == 0 {
		vBackoff = versionRetryBaseBackoff
	}

	return &Client{
		baseURL:                 baseURL,
		basePath:                basePath,
		username:                cfg.Username,
		password:                cfg.Password,
		twoFactor:               cfg.TwoFactorCode,
		httpClient:              client,
		maxRetries:              maxRetries,
		retryBackoff:            rb,
		rawAttempts:             ra,
		rawBackoff:              raBackoff,
		versionRetryAttempts:    vAttempts,
		versionRetryBaseBackoff: vBackoff,
	}, nil
}

func (c *Client) Login(ctx context.Context) error {
	c.authMu.Lock()
	defer c.authMu.Unlock()
	return c.loginLocked(ctx)
}

type loginCredentialAttempt struct {
	username  string
	password  string
	bootstrap bool
	label     string
}

// LoginWithBootstrapCredentials tries primary and opt-in bootstrap credentials
// in the safest order for the detected panel generation. 3x-ui v2.9 logs the
// password from failed login attempts, so panels without anonymous CSRF support
// try bootstrap credentials first and avoid probing with the desired steady
// state password during fresh-panel bootstrap. 3x-ui v3 has anonymous CSRF
// bootstrap and redacted failed-login logs, so it keeps the steady-state
// primary-first behavior.
func (c *Client) LoginWithBootstrapCredentials(ctx context.Context, bootstrapUsername, bootstrapPassword string) (bool, error) {
	c.authMu.Lock()
	defer c.authMu.Unlock()

	primaryUsername := c.username
	primaryPassword := c.password

	csrfToken, err := c.fetchCSRFToken(ctx, "csrf-token", true)
	if err != nil {
		return false, err
	}
	c.csrfToken = csrfToken

	primary := loginCredentialAttempt{
		username: primaryUsername,
		password: primaryPassword,
		label:    "primary",
	}
	bootstrap := loginCredentialAttempt{
		username:  bootstrapUsername,
		password:  bootstrapPassword,
		bootstrap: true,
		label:     "bootstrap",
	}

	attempts := []loginCredentialAttempt{primary, bootstrap}
	if csrfToken == "" {
		attempts = []loginCredentialAttempt{bootstrap, primary}
	}

	usedBootstrap, err := c.loginWithCredentialOrderLocked(ctx, attempts)
	if err != nil {
		c.username = primaryUsername
		c.password = primaryPassword
		return false, err
	}

	return usedBootstrap, nil
}

func (c *Client) loginWithCredentialOrderLocked(ctx context.Context, attempts []loginCredentialAttempt) (bool, error) {
	var firstErr error
	var firstLabel string
	var lastErr error

	for _, attempt := range attempts {
		c.username = attempt.username
		c.password = attempt.password

		err := c.loginLocked(ctx)
		if err == nil {
			return attempt.bootstrap, nil
		}
		if firstErr == nil {
			firstErr = err
			firstLabel = attempt.label
		}
		lastErr = err

		var loginErr *loginFailedError
		if !errors.As(err, &loginErr) {
			return false, err
		}
	}

	if len(attempts) == 0 {
		return false, errors.New("no login credentials configured")
	}
	lastLabel := attempts[len(attempts)-1].label
	return false, fmt.Errorf("%s login failed; %s login also failed: %w", firstLabel, lastLabel, lastErr)
}

func (c *Client) loginLocked(ctx context.Context) error {
	csrfToken, err := c.fetchCSRFToken(ctx, "csrf-token", true)
	if err != nil {
		return err
	}
	c.csrfToken = csrfToken

	form := url.Values{}
	form.Set("username", c.username)
	form.Set("password", c.password)
	if c.twoFactor != "" {
		form.Set("twoFactorCode", c.twoFactor)
	}
	if csrfToken != "" {
		form.Set("_csrf", csrfToken)
	}

	endpoint, err := c.resolvePath("login")
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if csrfToken != "" {
		req.Header.Set(csrfHeaderName, csrfToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return httpStatusError(resp.StatusCode, body)
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return err
	}
	if !apiResp.Success {
		if apiResp.Msg == "" {
			return &loginFailedError{message: "login failed"}
		}
		return &loginFailedError{message: apiResp.Msg}
	}
	return nil
}

func (c *Client) fetchCSRFToken(ctx context.Context, relPath string, optional bool) (string, error) {
	endpoint, err := c.resolvePath(relPath)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		if optional && (resp.StatusCode == http.StatusUnauthorized ||
			resp.StatusCode == http.StatusForbidden ||
			resp.StatusCode == http.StatusNotFound) {
			return "", nil
		}
		return "", httpStatusError(resp.StatusCode, body)
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		if optional {
			return "", nil
		}
		return "", err
	}
	if !apiResp.Success {
		if optional {
			return "", nil
		}
		if apiResp.Msg == "" {
			return "", errors.New("csrf token request failed")
		}
		return "", errors.New(apiResp.Msg)
	}
	if apiResp.Obj == nil {
		return "", nil
	}

	var token string
	if err := json.Unmarshal(apiResp.Obj, &token); err != nil {
		if optional {
			return "", nil
		}
		return "", err
	}
	return token, nil
}

func (c *Client) refreshCSRFToken(ctx context.Context) (bool, error) {
	c.authMu.Lock()
	defer c.authMu.Unlock()

	for _, relPath := range []string{"panel/csrf-token", "csrf-token"} {
		token, err := c.fetchCSRFToken(ctx, relPath, true)
		if err != nil {
			return false, err
		}
		if token != "" {
			c.csrfToken = token
			return true, nil
		}
	}
	c.csrfToken = ""
	return false, nil
}

func (c *Client) doJSON(ctx context.Context, method, relPath string, body any, out any) error {
	endpoint, err := c.resolvePath(relPath)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return err
		}
	}
	contentType := ""
	if body != nil {
		contentType = "application/json"
	}
	return c.doRequest(ctx, method, endpoint, contentType, buf.Bytes(), out)
}

func (c *Client) doForm(ctx context.Context, method, relPath string, form url.Values, out any) error { //nolint:unparam // method kept for API consistency with doJSON
	endpoint, err := c.resolvePath(relPath)
	if err != nil {
		return err
	}

	return c.doRequest(ctx, method, endpoint, "application/x-www-form-urlencoded", []byte(form.Encode()), out)
}

func (c *Client) AddInbound(ctx context.Context, inbound *Inbound) (*Inbound, error) {
	var out Inbound
	if err := c.doForm(ctx, http.MethodPost, "panel/api/inbounds/add", inboundToForm(inbound), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateInbound(ctx context.Context, inbound *Inbound) (*Inbound, error) {
	if inbound == nil {
		return nil, errors.New("inbound is nil")
	}
	if inbound.ID == 0 {
		return nil, errors.New("inbound id is required for update")
	}
	relPath := fmt.Sprintf("panel/api/inbounds/update/%d", inbound.ID)
	var out Inbound
	if err := c.doFormRetryable(ctx, http.MethodPost, relPath, inboundToForm(inbound), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteInbound(ctx context.Context, id int) error {
	if id == 0 {
		return errors.New("inbound id is required for delete")
	}
	relPath := fmt.Sprintf("panel/api/inbounds/del/%d", id)
	op := "DELETE " + relPath

	// 3x-ui's DelInbound is multi-step (xray API call → traffic cleanup →
	// row delete). On a transient panic the handler returns 5xx after the
	// row has already been removed — a naive `withRetry` would then hit
	// the `GetInbound first, error on missing row` path inside DelInbound
	// and turn success into failure (issue #161). So on 5xx we verify
	// with GetInbounds before deciding to retry.
	err := c.doForm(ctx, http.MethodPost, relPath, url.Values{}, nil)
	if err == nil {
		return nil
	}
	code, transient := transient5xxStatus(err)
	if !transient {
		return err
	}

	// First verify-and-maybe-retry pass.
	if c.deleteVerifyAbsent(ctx, op, id, code) {
		return nil
	}

	// Row still present (or verify itself failed). Retry the DELETE
	// once. If it succeeds, we are done.
	retryErr := c.doForm(ctx, http.MethodPost, relPath, url.Values{}, nil)
	if retryErr == nil {
		return nil
	}

	// Retry also failed. If it was another 5xx, the same panic-after-
	// commit case may have just played out a second time — verify once
	// more before propagating the error.
	if retryCode, retryTransient := transient5xxStatus(retryErr); retryTransient {
		if c.deleteVerifyAbsent(ctx, op, id, retryCode) {
			return nil
		}
	}
	return retryErr
}

// deleteVerifyAbsent calls GetInbounds and reports whether the inbound is
// gone. Logs both the verify attempt and any verify failure so operators
// can distinguish "row gone" from "could not check". Returns false on a
// verify-call error — caller treats that as "still present" and proceeds
// to retry the DELETE.
func (c *Client) deleteVerifyAbsent(ctx context.Context, op string, id, statusCode int) bool {
	tflog.Warn(ctx, "verifying delete after transient 5xx", map[string]any{
		"operation":   op,
		"status_code": statusCode,
	})
	gone, verifyErr := c.inboundAbsent(ctx, id)
	if verifyErr != nil {
		tflog.Warn(ctx, "delete verification failed; will retry DELETE", map[string]any{
			"operation": op,
			"error":     verifyErr.Error(),
		})
		return false
	}
	return gone
}

// inboundAbsent reports whether the inbound with id is no longer present in
// the panel's list. A list-call error is propagated — callers must not treat
// "could not check" as "row gone".
func (c *Client) inboundAbsent(ctx context.Context, id int) (bool, error) {
	inbounds, err := c.GetInbounds(ctx)
	if err != nil {
		return false, err
	}
	for _, in := range inbounds {
		if in.ID == id {
			return false, nil
		}
	}
	return true, nil
}

func (c *Client) GetInbound(ctx context.Context, id int) (*Inbound, error) {
	if id == 0 {
		return nil, errors.New("inbound id is required for get")
	}
	relPath := fmt.Sprintf("panel/api/inbounds/get/%d", id)
	var out Inbound
	if err := c.doJSON(ctx, http.MethodGet, relPath, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetInbounds(ctx context.Context) ([]Inbound, error) {
	var out []Inbound
	if err := c.doJSON(ctx, http.MethodGet, "panel/api/inbounds/list", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) AddInboundClient(ctx context.Context, inboundID int, client map[string]any) error {
	if inboundID == 0 {
		return errors.New("inbound id is required for add client")
	}
	if client == nil {
		return errors.New("client data is required")
	}

	if c.useNewClientAPI(ctx) {
		payload := map[string]any{
			"client":     client,
			"inboundIds": []int{inboundID},
		}
		err := c.doJSON(ctx, http.MethodPost, "panel/api/clients/add", payload, nil)
		if err == nil {
			return nil
		}
		if isHTTPNotFound(err) {
			c.markLegacyClientAPI()
		} else {
			return err
		}
	}

	// Old endpoint (v2.9.x, v3.0.x)
	payload := map[string]any{"clients": []map[string]any{client}}
	settings, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	form := url.Values{}
	form.Set("id", strconv.Itoa(inboundID))
	form.Set("settings", string(settings))
	return c.doForm(ctx, http.MethodPost, "panel/api/inbounds/addClient", form, nil)
}

func (c *Client) UpdateInboundClient(ctx context.Context, inboundID int, clientID string, currentEmail string, client map[string]any) error {
	if inboundID == 0 {
		return errors.New("inbound id is required for update client")
	}
	if clientID == "" {
		return errors.New("client id is required for update client")
	}
	if client == nil {
		return errors.New("client data is required")
	}

	if c.useNewClientAPI(ctx) {
		if currentEmail == "" {
			currentEmail, _ = client["email"].(string)
		}
		if currentEmail == "" {
			return errors.New("client email is required for v3.1.0+ update")
		}
		relPath := fmt.Sprintf("panel/api/clients/update/%s", url.PathEscape(currentEmail))
		err := c.doJSONRetryable(ctx, http.MethodPost, relPath, client, nil)
		if err == nil {
			return nil
		}
		if isHTTPNotFound(err) {
			c.markLegacyClientAPI()
		} else {
			return err
		}
	}

	// Old endpoint (v2.9.x, v3.0.x)
	payload := map[string]any{"clients": []map[string]any{client}}
	settings, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	form := url.Values{}
	form.Set("id", strconv.Itoa(inboundID))
	form.Set("settings", string(settings))
	relPath := fmt.Sprintf("panel/api/inbounds/updateClient/%s", clientID)
	return c.doFormRetryable(ctx, http.MethodPost, relPath, form, nil)
}

func (c *Client) DeleteInboundClient(ctx context.Context, inboundID int, clientID string, email string) error {
	if inboundID == 0 {
		return errors.New("inbound id is required for delete client")
	}
	if clientID == "" {
		return errors.New("client id is required for delete client")
	}

	if c.useNewClientAPI(ctx) && email != "" {
		relPath := fmt.Sprintf("panel/api/clients/del/%s", url.PathEscape(email))
		err := c.doForm(ctx, http.MethodPost, relPath, url.Values{}, nil)
		if err == nil {
			return nil
		}
		if isHTTPNotFound(err) {
			c.markLegacyClientAPI()
		} else if strings.Contains(strings.ToLower(err.Error()), "client not found") {
			return nil
		} else {
			return err
		}
	}

	// Old endpoint (v2.9.x, v3.0.x)
	relPath := fmt.Sprintf("panel/api/inbounds/%d/delClient/%s", inboundID, clientID)
	err := c.doForm(ctx, http.MethodPost, relPath, url.Values{}, nil)
	if err != nil && strings.Contains(err.Error(), "Client Not Found") {
		return nil
	}
	return err
}

func (c *Client) GetServerStatus(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.doJSON(ctx, http.MethodGet, "panel/api/server/status", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

const (
	// versionRetryAttempts is the maximum number of retries for GetXrayVersions
	// when the upstream GitHub API rate-limits the 3x-ui panel's internal
	// cache fetch. Exponential backoff: 2s, 4s, 8s, 16s.
	versionRetryAttempts    = 4
	versionRetryBaseBackoff = 2 * time.Second
)

// isUpstreamRateLimitError reports whether err originated from 3x-ui's
// getXrayVersion handler failing due to an upstream GitHub API rate limit.
// The panel returns success:false with a message containing "rate limit".
func isUpstreamRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "rate limit")
}

func (c *Client) GetXrayVersions(ctx context.Context) ([]string, error) {
	var out []string
	var lastErr error
	backoff := c.versionRetryBaseBackoff
	for attempt := 0; attempt <= c.versionRetryAttempts; attempt++ {
		err := c.doJSON(ctx, http.MethodGet, "panel/api/server/getXrayVersion", nil, &out)
		if err == nil {
			return out, nil
		}
		lastErr = err
		if !isUpstreamRateLimitError(err) {
			return nil, err
		}
		if attempt == c.versionRetryAttempts {
			break
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		tflog.Warn(ctx, "retrying getXrayVersion after upstream rate limit", map[string]any{
			"attempt":      attempt + 1,
			"max_attempts": c.versionRetryAttempts,
			"backoff":      backoff.String(),
		})
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
		}
	}
	return nil, fmt.Errorf("getXrayVersion: upstream rate limit persisted after %d retries: %w", c.versionRetryAttempts, lastErr)
}

// ErrXrayVersionUnknown is returned when the 3x-ui API reports the Xray
// version as "Unknown", which typically means Xray is not running.
var ErrXrayVersionUnknown = errors.New("xray version is unknown (Xray may not be running)")

// GetCurrentXrayVersion returns the installed Xray version with "v" prefix.
// The server status API returns version without "v" (e.g. "26.2.6"),
// but installXray and getXrayVersion use "v"-prefixed tags (e.g. "v26.2.6").
// When Xray is not running, the API may return "Unknown" — this is treated as
// an error because the resource cannot determine the actual installed version.
func (c *Client) GetCurrentXrayVersion(ctx context.Context) (string, error) {
	status, err := c.GetServerStatus(ctx)
	if err != nil {
		return "", err
	}
	xray, ok := status["xray"].(map[string]any)
	if !ok {
		return "", errors.New("xray section not found in server status")
	}
	version, ok := xray["version"].(string)
	if !ok {
		return "", errors.New("xray version not found in server status")
	}
	if version == "Unknown" {
		return "", ErrXrayVersionUnknown
	}
	return normalizeXrayVersion(version), nil
}

// normalizeXrayVersion ensures the version string has a "v" prefix.
func normalizeXrayVersion(v string) string {
	if v != "" && v[0] != 'v' {
		return "v" + v
	}
	return v
}

func (c *Client) InstallXray(ctx context.Context, version string) error {
	return c.doJSON(ctx, http.MethodPost, "panel/api/server/installXray/"+version, nil, nil)
}

func (c *Client) GetXrayConfig(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.doJSON(ctx, http.MethodGet, "panel/api/server/getConfigJson", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetOnlineClients(ctx context.Context) ([]string, error) {
	if c.useNewClientAPI(ctx) {
		var out []string
		err := c.doJSON(ctx, http.MethodPost, "panel/api/clients/onlines", nil, &out)
		if err == nil {
			if out == nil {
				out = []string{}
			}
			return out, nil
		}
		if isHTTPNotFound(err) {
			c.markLegacyClientAPI()
		} else {
			return nil, err
		}
	}

	// Old endpoint (v2.9.x, v3.0.x)
	var out []string
	if err := c.doJSON(ctx, http.MethodPost, "panel/api/inbounds/onlines", nil, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = []string{}
	}
	return out, nil
}

func (c *Client) GetClientTraffics(ctx context.Context, email string) (*ClientTraffic, error) {
	if email == "" {
		return nil, errors.New("email is required for get client traffics")
	}

	if c.useNewClientAPI(ctx) {
		relPath := fmt.Sprintf("panel/api/clients/traffic/%s", url.PathEscape(email))
		var out ClientTraffic
		err := c.doJSON(ctx, http.MethodGet, relPath, nil, &out)
		if err == nil {
			if out.Email == "" {
				return nil, fmt.Errorf("client with email %q not found", email)
			}
			return &out, nil
		}
		if isHTTPNotFound(err) {
			c.markLegacyClientAPI()
		} else {
			return nil, err
		}
	}

	// Old endpoint (v2.9.x, v3.0.x)
	relPath := fmt.Sprintf("panel/api/inbounds/getClientTraffics/%s", url.PathEscape(email))
	var out ClientTraffic
	if err := c.doJSON(ctx, http.MethodGet, relPath, nil, &out); err != nil {
		return nil, err
	}
	if out.Email == "" {
		return nil, fmt.Errorf("client with email %q not found", email)
	}
	return &out, nil
}

func (c *Client) GetNewX25519Cert(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.doJSON(ctx, http.MethodGet, "panel/api/server/getNewX25519Cert", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

type VlessEncAuth struct {
	Label      string `json:"label"`
	Decryption string `json:"decryption"`
	Encryption string `json:"encryption"`
}

func (c *Client) GetNewVlessEnc(ctx context.Context) ([]VlessEncAuth, error) {
	var out struct {
		Auths []VlessEncAuth `json:"auths"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "panel/api/server/getNewVlessEnc", nil, &out); err != nil {
		return nil, err
	}
	return out.Auths, nil
}

func (c *Client) UpdateUser(ctx context.Context, oldUsername, oldPassword, newUsername, newPassword string) error {
	payload := map[string]any{
		"oldUsername": oldUsername,
		"oldPassword": oldPassword,
		"newUsername": newUsername,
		"newPassword": newPassword,
	}
	err := c.withSettingsFallback(ctx, func() error {
		return c.doJSON(ctx, http.MethodPost, c.settingPath(ctx, "updateUser"), payload, nil)
	})
	if err != nil {
		return err
	}
	c.username = newUsername
	c.password = newPassword
	return nil
}

func (c *Client) GetSettings(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	err := c.withSettingsFallback(ctx, func() error {
		return c.doForm(ctx, http.MethodPost, c.settingPath(ctx, "all"), url.Values{}, &out)
	})
	return out, err
}

func (c *Client) UpdateSettings(ctx context.Context, settings map[string]any) error {
	if settings == nil {
		return errors.New("settings payload is required")
	}
	return c.withSettingsFallback(ctx, func() error {
		return c.doJSONRetryable(ctx, http.MethodPost, c.settingPath(ctx, "update"), settings, nil)
	})
}

// SendRestart sends the restart request but does not wait for readiness.
func (c *Client) SendRestart(ctx context.Context) error {
	// The panel may close the connection mid-response, so ignore EOF.
	var firstErr error
	err := c.withSettingsFallback(ctx, func() error {
		e := c.doForm(ctx, http.MethodPost, c.settingPath(ctx, "restartPanel"), url.Values{}, nil)
		if e != nil && e.Error() == "EOF" {
			firstErr = e
			return nil
		}
		return e
	})
	if err != nil {
		return err
	}
	if firstErr != nil {
		return nil
	}
	return nil
}

// RestartPanel sends a restart and waits for the panel to become ready.
func (c *Client) RestartPanel(ctx context.Context) error {
	if err := c.SendRestart(ctx); err != nil {
		return err
	}
	return c.WaitForReady(ctx)
}

// RestartXrayService restarts the Xray core via the 3x-ui API.
func (c *Client) RestartXrayService(ctx context.Context) error {
	return c.doJSON(ctx, http.MethodPost, "panel/api/server/restartXrayService", nil, nil)
}

// WaitForReady polls the panel until it responds successfully or the context
// is cancelled. The panel needs a few seconds to come back after a restart.
func (c *Client) WaitForReady(ctx context.Context) error {
	const (
		interval = 2 * time.Second
		timeout  = 30 * time.Second
		// 3x-ui restarts its web server asynchronously after /setting/restartPanel:
		// the SIGHUP is delivered to a channel and the main loop does Stop()+Start()
		// in a goroutine (see main.go), so the old socket can briefly still answer
		// before being torn down. A single successful login is therefore racy — a
		// caller can observe "ready" on the dying socket and then hit a connection
		// reset a moment later. Require a few CONSECUTIVE successes so the restart
		// has actually settled.
		consecutiveRequired = 3
	)
	deadline := time.Now().Add(timeout)
	streak := 0
	for {
		if time.Now().After(deadline) {
			return errors.New("panel did not become ready after restart")
		}
		// Try to login — this also verifies the panel is reachable.
		if err := c.Login(ctx); err == nil {
			streak++
			if streak >= consecutiveRequired {
				return nil
			}
		} else {
			streak = 0
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}

func (c *Client) GetXrayTemplate(ctx context.Context) (map[string]any, error) {
	var raw string
	err := c.withSettingsFallback(ctx, func() error {
		return c.doForm(ctx, http.MethodPost, c.xraySettingPath(ctx, ""), url.Values{}, &raw)
	})
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(raw) == "" {
		return map[string]any{}, nil
	}
	var payload struct {
		XraySetting any `json:"xraySetting"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, err
	}
	if payload.XraySetting == nil {
		return map[string]any{}, nil
	}
	settings, ok := payload.XraySetting.(map[string]any)
	if !ok {
		return nil, errors.New("xraySetting is not an object")
	}
	return settings, nil
}

func (c *Client) UpdateXrayTemplate(ctx context.Context, settings map[string]any) error {
	if settings == nil {
		return errors.New("xraySetting payload is required")
	}
	// Preserve outboundTestUrl so it is not reset by 3x-ui when only
	// the xray template is being updated.
	testURL, err := c.GetXrayOutboundTestURL(ctx)
	if err != nil {
		return fmt.Errorf("failed to read outboundTestUrl: %w", err)
	}

	payload, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	form := url.Values{}
	form.Set("xraySetting", string(payload))
	if testURL != "" {
		form.Set("outboundTestUrl", testURL)
	}
	return c.withSettingsFallback(ctx, func() error {
		return c.doFormRetryable(ctx, http.MethodPost, c.xraySettingPath(ctx, "update"), form, nil)
	})
}

func (c *Client) GetXrayOutboundTestURL(ctx context.Context) (string, error) {
	var raw string
	err := c.withSettingsFallback(ctx, func() error {
		return c.doForm(ctx, http.MethodPost, c.xraySettingPath(ctx, ""), url.Values{}, &raw)
	})
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(raw) == "" {
		return "", nil
	}
	var payload struct {
		OutboundTestURL string `json:"outboundTestUrl"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return "", err
	}
	return payload.OutboundTestURL, nil
}

func (c *Client) SetXrayOutboundTestURL(ctx context.Context, testURL string) error {
	// Read current xray template to preserve it.
	tmpl, err := c.GetXrayTemplate(ctx)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(tmpl)
	if err != nil {
		return err
	}
	form := url.Values{}
	form.Set("xraySetting", string(payload))
	form.Set("outboundTestUrl", testURL)
	return c.withSettingsFallback(ctx, func() error {
		return c.doFormRetryable(ctx, http.MethodPost, c.xraySettingPath(ctx, "update"), form, nil)
	})
}

func (c *Client) doRequest(ctx context.Context, method, endpoint, contentType string, body []byte, out any) error {
	resp, err := c.doRequestOnce(ctx, method, endpoint, contentType, body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusForbidden && requiresCSRF(method) {
		// #nosec G104 -- discarding body before retry; Close error is not actionable
		resp.Body.Close()
		refreshed, refreshErr := c.refreshCSRFToken(ctx)
		if refreshErr != nil {
			return refreshErr
		}
		if !refreshed {
			return &HTTPStatusError{StatusCode: http.StatusForbidden}
		}
		resp, err = c.doRequestOnce(ctx, method, endpoint, contentType, body)
		if err != nil {
			return err
		}
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusNotFound {
		// #nosec G104 -- discarding body before re-login; Close error is not actionable
		resp.Body.Close()
		if err := c.Login(ctx); err != nil {
			return err
		}
		resp, err = c.doRequestOnce(ctx, method, endpoint, contentType, body)
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusForbidden && requiresCSRF(method) {
			// #nosec G104 -- discarding body before retry; Close error is not actionable
			resp.Body.Close()
			refreshed, refreshErr := c.refreshCSRFToken(ctx)
			if refreshErr != nil {
				return refreshErr
			}
			if !refreshed {
				return &HTTPStatusError{StatusCode: http.StatusForbidden}
			}
			resp, err = c.doRequestOnce(ctx, method, endpoint, contentType, body)
			if err != nil {
				return err
			}
		}
	}
	defer resp.Body.Close()

	return decodeAPIResponse(resp, out)
}

// HTTPStatusError carries the HTTP status code from an upstream response so
// callers can distinguish transient 5xx (retryable) from 4xx (do not retry).
type HTTPStatusError struct {
	StatusCode int
	Body       string
}

func (e *HTTPStatusError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("request failed: status %d", e.StatusCode)
	}
	return fmt.Sprintf("request failed: status %d, body: %s", e.StatusCode, e.Body)
}

func httpStatusError(statusCode int, body []byte) error {
	msg := strings.TrimSpace(string(body))
	if len(msg) > 1024 {
		msg = msg[:1024] + "...(truncated)"
	}
	return &HTTPStatusError{StatusCode: statusCode, Body: msg}
}

// transient5xxStatus reports whether err is an *HTTPStatusError with a 5xx
// code, returning the code alongside the boolean for callers that want to
// log it without re-running errors.As. 5xx on a write endpoint typically
// originates from gin's recovery middleware after a panic in the handler —
// the panel's controllers otherwise return 200 with success:false. A panic
// during a SQLite transaction is the canonical "transient" failure this
// retry targets.
func transient5xxStatus(err error) (int, bool) {
	var httpErr *HTTPStatusError
	if !errors.As(err, &httpErr) {
		return 0, false
	}
	if httpErr.StatusCode < 500 || httpErr.StatusCode >= 600 {
		return httpErr.StatusCode, false
	}
	return httpErr.StatusCode, true
}

// withRetry runs fn up to (1 + c.maxRetries) times, retrying only on a
// transient 5xx response from an idempotent endpoint. It is the single
// retry policy in the client; doFormRetryable / doJSONRetryable wrap it
// so callers do not have to construct logging fields themselves.
//
// Not safe for non-idempotent endpoints: AddInbound (would create a
// duplicate), AddInboundClient (duplicate), UpdateUser (the second call
// would run with stale credentials and could leave provider state and
// panel state out of sync). DeleteInbound has its own retry-with-verify
// path — see DeleteInbound — because a naive retry would turn a
// successful-but-5xx delete into a failure (DelInbound errors on a
// missing row).
//
// Retries are visible: every retry emits a tflog.Warn so operators can
// detect upstream flakiness instead of having it silently absorbed.
func (c *Client) withRetry(ctx context.Context, op string, fn func() error) error {
	var err error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}
		statusCode, retryable := transient5xxStatus(err)
		if !retryable || attempt == c.maxRetries {
			return err
		}
		// Honor cancellation before logging or sleeping so a cancelled
		// context does not produce a "retrying" entry that we never act on.
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		tflog.Warn(ctx, "retrying transient 5xx", map[string]any{
			"operation":    op,
			"attempt":      attempt + 1,
			"max_attempts": c.maxRetries,
			"status_code":  statusCode,
			"backoff":      c.retryBackoff.String(),
		})
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(c.retryBackoff):
		}
	}
	return err
}

func (c *Client) doFormRetryable(ctx context.Context, method, relPath string, form url.Values, out any) error { //nolint:unparam // method kept for API symmetry with doJSONRetryable / doForm
	return c.withRetry(ctx, method+" "+relPath, func() error {
		return c.doForm(ctx, method, relPath, form, out)
	})
}

func (c *Client) doJSONRetryable(ctx context.Context, method, relPath string, body any, out any) error {
	return c.withRetry(ctx, method+" "+relPath, func() error {
		return c.doJSON(ctx, method, relPath, body, out)
	})
}

// WithReadAfterWriteRetry polls fn until the row is visible (found=true) or
// the budget is exhausted. It handles two transient conditions:
//
//   - found=false: the write succeeded but SQLite hasn't made the row visible yet.
//   - err is a transient 5xx: the panel is temporarily unavailable under load.
//
// Non-transient errors (4xx, auth failures, etc.) abort immediately.
func (c *Client) WithReadAfterWriteRetry(ctx context.Context, opName string, fn func() (bool, error)) error {
	for attempt := 0; attempt < c.rawAttempts; attempt++ {
		found, err := fn()
		if err != nil {
			if _, retryable := transient5xxStatus(err); !retryable || attempt == c.rawAttempts-1 {
				return err
			}
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
			}
			tflog.Warn(ctx, "retrying read-after-write (transient error)", map[string]any{
				"operation":    opName,
				"attempt":      attempt + 1,
				"max_attempts": c.rawAttempts,
				"backoff":      c.rawBackoff.String(),
			})
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.rawBackoff):
			}
			continue
		}
		if found {
			return nil
		}
		if attempt == c.rawAttempts-1 {
			break
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		tflog.Warn(ctx, "retrying read-after-write", map[string]any{
			"operation":    opName,
			"attempt":      attempt + 1,
			"max_attempts": c.rawAttempts,
			"backoff":      c.rawBackoff.String(),
		})
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(c.rawBackoff):
		}
	}
	return fmt.Errorf("%s: row not visible after %d attempts (%s total)", opName, c.rawAttempts, c.rawBackoff*time.Duration(c.rawAttempts-1))
}

func (c *Client) doRequestOnce(ctx context.Context, method, endpoint, contentType string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if requiresCSRF(method) {
		c.authMu.Lock()
		token := c.csrfToken
		c.authMu.Unlock()
		if token != "" {
			req.Header.Set(csrfHeaderName, token)
		}
	}
	return c.httpClient.Do(req)
}

func requiresCSRF(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return false
	default:
		return true
	}
}

func decodeAPIResponse(resp *http.Response, out any) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		if resp.StatusCode >= 400 {
			return &HTTPStatusError{StatusCode: resp.StatusCode}
		}
		return nil
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		if resp.StatusCode >= 400 {
			return httpStatusError(resp.StatusCode, body)
		}
		return err
	}
	if !apiResp.Success {
		if apiResp.Msg == "" {
			return fmt.Errorf("request failed: status %d", resp.StatusCode)
		}
		return fmt.Errorf("request failed: status %d, msg: %s", resp.StatusCode, apiResp.Msg)
	}

	if out == nil || apiResp.Obj == nil {
		return nil
	}

	return json.Unmarshal(apiResp.Obj, out)
}

func (c *Client) resolvePath(rel string) (string, error) {
	if rel == "" {
		return "", errors.New("empty path")
	}

	base := *c.baseURL
	basePath := strings.TrimSuffix(base.Path, "/")
	merged := basePath + c.basePath
	if !strings.HasSuffix(merged, "/") {
		merged += "/"
	}
	merged += strings.TrimPrefix(rel, "/")
	base.Path = merged
	return base.String(), nil
}

func normalizeBasePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}
	return p
}

func inboundToForm(in *Inbound) url.Values {
	form := url.Values{}
	if in == nil {
		return form
	}
	form.Set("id", strconv.Itoa(in.ID))
	form.Set("up", strconv.FormatInt(in.Up, 10))
	form.Set("down", strconv.FormatInt(in.Down, 10))
	form.Set("total", strconv.FormatInt(in.Total, 10))
	form.Set("remark", in.Remark)
	form.Set("enable", strconv.FormatBool(in.Enable))
	form.Set("expiryTime", strconv.FormatInt(in.ExpiryTime, 10))
	form.Set("trafficReset", in.TrafficReset)
	form.Set("lastTrafficResetTime", strconv.FormatInt(in.LastTrafficResetTime, 10))
	form.Set("listen", in.Listen)
	form.Set("port", strconv.Itoa(in.Port))
	form.Set("protocol", in.Protocol)
	form.Set("settings", in.Settings)
	form.Set("streamSettings", in.StreamSettings)
	form.Set("sniffing", in.Sniffing)
	if in.NodeID != nil {
		form.Set("nodeId", strconv.Itoa(*in.NodeID))
	}
	return form
}

// useNewClientAPI returns true if the v3.1.0+ /panel/api/clients/* surface
// should be tried. On the first call it probes the panel to detect the
// available API surface. Subsequent calls use the cached result.
// The probe goes through doRequest so it benefits from auto re-login
// on expired sessions (a raw HTTP GET would get 404 on unauthenticated
// v3.1.0+ and incorrectly mark the API as old).
func (c *Client) useNewClientAPI(ctx context.Context) bool {
	c.newClientMu.Lock()
	defer c.newClientMu.Unlock()
	if c.newClientAPI != nil {
		return *c.newClientAPI
	}

	var out json.RawMessage
	err := c.doJSON(ctx, http.MethodGet, "panel/api/clients/list", nil, &out)
	isNew := err == nil
	c.newClientAPI = &isNew
	if isNew {
		tflog.Info(ctx, "detected 3x-ui v3.1.0+ client API surface")
	} else {
		tflog.Info(ctx, "detected 3x-ui v2.9.x/v3.0.x client API surface")
	}
	return isNew
}

func (c *Client) markLegacyClientAPI() {
	v := false
	c.newClientMu.Lock()
	c.newClientAPI = &v
	c.newClientMu.Unlock()
}

// useSettingsAPI returns true if the v3.3.0+ /panel/api/setting/* surface
// should be used. On the first call it probes the panel to detect the
// available API surface. Subsequent calls use the cached result.
func (c *Client) useSettingsAPI(ctx context.Context) bool {
	c.settingsAPIMu.Lock()
	defer c.settingsAPIMu.Unlock()
	if c.settingsUnderAPI != nil {
		return *c.settingsUnderAPI
	}

	var out json.RawMessage
	err := c.doForm(ctx, http.MethodPost, "panel/api/setting/all", url.Values{}, &out)
	isNew := err == nil
	c.settingsUnderAPI = &isNew
	if isNew {
		tflog.Info(ctx, "detected 3x-ui v3.3.0+ settings API surface (/panel/api/setting/*)")
	} else {
		tflog.Info(ctx, "detected pre-v3.3.0 settings API surface (/panel/setting/*)")
	}
	return isNew
}

func (c *Client) markLegacySettingsAPI(ctx context.Context) {
	v := false
	c.settingsAPIMu.Lock()
	c.settingsUnderAPI = &v
	c.settingsAPIMu.Unlock()
	tflog.Warn(ctx, "settings API returned 404, falling back to pre-v3.3.0 API surface (/panel/setting/*)")
}

// withSettingsFallback executes fn. If the v3.3.0+ API was detected but the
// request returns 404, it marks the API as legacy and retries once.
func (c *Client) withSettingsFallback(ctx context.Context, fn func() error) error {
	err := fn()
	if err == nil || !isHTTPNotFound(err) || !c.useSettingsAPI(ctx) {
		return err
	}
	c.markLegacySettingsAPI(ctx)
	return fn()
}

// settingPath returns the correct API path for a settings endpoint based on
// the detected 3x-ui version.
func (c *Client) settingPath(ctx context.Context, suffix string) string {
	if c.useSettingsAPI(ctx) {
		return "panel/api/setting/" + suffix
	}
	return "panel/setting/" + suffix
}

// xraySettingPath returns the correct API path for an xray settings endpoint
// based on the detected 3x-ui version.
func (c *Client) xraySettingPath(ctx context.Context, suffix string) string {
	prefix := "panel/xray"
	if c.useSettingsAPI(ctx) {
		prefix = "panel/api/xray"
	}
	if suffix == "" {
		return prefix
	}
	return prefix + "/" + suffix
}

// isHTTPNotFound reports whether err originated from an HTTP 404 response.
func isHTTPNotFound(err error) bool {
	var httpErr *HTTPStatusError
	if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
		return true
	}
	return false
}
