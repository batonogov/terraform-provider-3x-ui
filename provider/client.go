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
}

type Client struct {
	baseURL    *url.URL
	basePath   string
	username   string
	password   string
	twoFactor  string
	httpClient *http.Client
	maxRetries int
}

// SetBasePath updates the client's base path to match a new webBasePath.
func (c *Client) SetBasePath(p string) {
	c.basePath = normalizeBasePath(p)
}

type apiResponse struct {
	Success bool            `json:"success"`
	Msg     string          `json:"msg"`
	Obj     json.RawMessage `json:"obj"`
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

	return &Client{
		baseURL:    baseURL,
		basePath:   basePath,
		username:   cfg.Username,
		password:   cfg.Password,
		twoFactor:  cfg.TwoFactorCode,
		httpClient: client,
		maxRetries: maxRetries,
	}, nil
}

func (c *Client) Login(ctx context.Context) error {
	form := url.Values{}
	form.Set("username", c.username)
	form.Set("password", c.password)
	if c.twoFactor != "" {
		form.Set("twoFactorCode", c.twoFactor)
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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return err
	}
	if !apiResp.Success {
		if apiResp.Msg == "" {
			return errors.New("login failed")
		}
		return errors.New(apiResp.Msg)
	}
	return nil
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
	return c.doForm(ctx, http.MethodPost, relPath, url.Values{}, nil)
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

func (c *Client) UpdateInboundClient(ctx context.Context, inboundID int, clientID string, client map[string]any) error {
	if inboundID == 0 {
		return errors.New("inbound id is required for update client")
	}
	if clientID == "" {
		return errors.New("client id is required for update client")
	}
	if client == nil {
		return errors.New("client data is required")
	}
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

func (c *Client) DeleteInboundClient(ctx context.Context, inboundID int, clientID string) error {
	if inboundID == 0 {
		return errors.New("inbound id is required for delete client")
	}
	if clientID == "" {
		return errors.New("client id is required for delete client")
	}
	relPath := fmt.Sprintf("panel/api/inbounds/%d/delClient/%s", inboundID, clientID)
	return c.doForm(ctx, http.MethodPost, relPath, url.Values{}, nil)
}

func (c *Client) GetServerStatus(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.doJSON(ctx, http.MethodGet, "panel/api/server/status", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetXrayVersions(ctx context.Context) ([]string, error) {
	var out []string
	if err := c.doJSON(ctx, http.MethodGet, "panel/api/server/getXrayVersion", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
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
	if err := c.doJSON(ctx, http.MethodPost, "panel/setting/updateUser", payload, nil); err != nil {
		return err
	}
	c.username = newUsername
	c.password = newPassword
	return nil
}

func (c *Client) GetSettings(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.doForm(ctx, http.MethodPost, "panel/setting/all", url.Values{}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) UpdateSettings(ctx context.Context, settings map[string]any) error {
	if settings == nil {
		return errors.New("settings payload is required")
	}
	return c.doJSONRetryable(ctx, http.MethodPost, "panel/setting/update", settings, nil)
}

// SendRestart sends the restart request but does not wait for readiness.
func (c *Client) SendRestart(ctx context.Context) error {
	// The panel may close the connection mid-response, so ignore EOF.
	err := c.doForm(ctx, http.MethodPost, "panel/setting/restartPanel", url.Values{}, nil)
	if err != nil && err.Error() == "EOF" {
		return nil
	}
	return err
}

// RestartPanel sends a restart and waits for the panel to become ready.
func (c *Client) RestartPanel(ctx context.Context) error {
	if err := c.SendRestart(ctx); err != nil {
		return err
	}
	return c.WaitForReady(ctx)
}

// WaitForReady polls the panel until it responds successfully or the context
// is cancelled. The panel needs a few seconds to come back after a restart.
func (c *Client) WaitForReady(ctx context.Context) error {
	const (
		interval = 2 * time.Second
		timeout  = 30 * time.Second
	)
	deadline := time.Now().Add(timeout)
	for {
		if time.Now().After(deadline) {
			return errors.New("panel did not become ready after restart")
		}
		// Try to login — this also verifies the panel is reachable.
		if err := c.Login(ctx); err == nil {
			return nil
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
	if err := c.doForm(ctx, http.MethodPost, "panel/xray", url.Values{}, &raw); err != nil {
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
	return c.doFormRetryable(ctx, http.MethodPost, "panel/xray/update", form, nil)
}

func (c *Client) GetXrayOutboundTestURL(ctx context.Context) (string, error) {
	var raw string
	if err := c.doForm(ctx, http.MethodPost, "panel/xray", url.Values{}, &raw); err != nil {
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
	return c.doFormRetryable(ctx, http.MethodPost, "panel/xray/update", form, nil)
}

func (c *Client) doRequest(ctx context.Context, method, endpoint, contentType string, body []byte, out any) error {
	resp, err := c.doRequestOnce(ctx, method, endpoint, contentType, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusNotFound {
		if err := c.Login(ctx); err != nil {
			return err
		}
		resp.Body.Close()
		resp, err = c.doRequestOnce(ctx, method, endpoint, contentType, body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
	}

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
// panel state out of sync), DeleteInbound (3x-ui's DelInbound calls
// GetInbound first and errors on a missing row, so a retry after a
// successful-but-5xx delete turns success into failure).
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
			"backoff":      defaultRetryBackoff.String(),
		})
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(defaultRetryBackoff):
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
// the budget is exhausted. It is distinct from withRetry: that one absorbs
// transient HTTP 5xx, this one absorbs application-layer "success but the
// row is not visible yet" gaps. Callers pass an opName for telemetry.
//
// fn returns (found, err). A non-nil err aborts immediately (no retry — if
// the read itself failed, the panel is in worse shape than just slow).
// found=false triggers a backoff and another attempt.
func (c *Client) WithReadAfterWriteRetry(ctx context.Context, opName string, fn func() (bool, error)) error {
	for attempt := 0; attempt < readAfterWriteAttempts; attempt++ {
		found, err := fn()
		if err != nil {
			return err
		}
		if found {
			return nil
		}
		if attempt == readAfterWriteAttempts-1 {
			break
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		tflog.Warn(ctx, "retrying read-after-write", map[string]any{
			"operation":    opName,
			"attempt":      attempt + 1,
			"max_attempts": readAfterWriteAttempts,
			"backoff":      readAfterWriteBackoff.String(),
		})
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(readAfterWriteBackoff):
		}
	}
	return fmt.Errorf("%s: row not visible after %d attempts (%s total)", opName, readAfterWriteAttempts, readAfterWriteBackoff*time.Duration(readAfterWriteAttempts-1))
}

func (c *Client) doRequestOnce(ctx context.Context, method, endpoint, contentType string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return c.httpClient.Do(req)
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
			msg := strings.TrimSpace(string(body))
			if len(msg) > 1024 {
				msg = msg[:1024] + "...(truncated)"
			}
			return &HTTPStatusError{StatusCode: resp.StatusCode, Body: msg}
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
	return form
}
