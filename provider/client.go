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
)

type ClientConfig struct {
	Endpoint           string
	BasePath           string
	Username           string
	Password           string
	TwoFactorCode      string
	InsecureSkipVerify bool
	Timeout            time.Duration
}

type Client struct {
	baseURL    *url.URL
	basePath   string
	username   string
	password   string
	twoFactor  string
	httpClient *http.Client
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

	return &Client{
		baseURL:    baseURL,
		basePath:   basePath,
		username:   cfg.Username,
		password:   cfg.Password,
		twoFactor:  cfg.TwoFactorCode,
		httpClient: client,
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
	if err := c.doForm(ctx, http.MethodPost, relPath, inboundToForm(inbound), &out); err != nil {
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
	return c.doForm(ctx, http.MethodPost, relPath, form, nil)
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

// GetCurrentXrayVersion returns the installed Xray version with "v" prefix.
// The server status API returns version without "v" (e.g. "26.2.6"),
// but installXray and getXrayVersion use "v"-prefixed tags (e.g. "v26.2.6").
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
	return c.doJSON(ctx, http.MethodPost, "panel/setting/update", settings, nil)
}

func (c *Client) RestartPanel(ctx context.Context) error {
	return c.doForm(ctx, http.MethodPost, "panel/setting/restartPanel", url.Values{}, nil)
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
	return c.doForm(ctx, http.MethodPost, "panel/xray/update", form, nil)
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
	return c.doForm(ctx, http.MethodPost, "panel/xray/update", form, nil)
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
			return fmt.Errorf("request failed: status %d", resp.StatusCode)
		}
		return nil
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		if resp.StatusCode >= 400 {
			msg := strings.TrimSpace(string(body))
			if msg == "" {
				return fmt.Errorf("request failed: status %d", resp.StatusCode)
			}
			if len(msg) > 1024 {
				msg = msg[:1024] + "...(truncated)"
			}
			return fmt.Errorf("request failed: status %d, body: %s", resp.StatusCode, msg)
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
