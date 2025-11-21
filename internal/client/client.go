package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Config defines the options required to build a Client instance.
type Config struct {
	BaseURL       *url.URL
	Username      *string
	Password      *string
	SessionCookie *string
	TLSSkipVerify bool

	RequestTimeout time.Duration
	MaxRetries     int
	PollInterval   time.Duration
	UserAgent      string
	HTTPClient     httpClient
}

// Client is the high-level interface consumed by provider resources/data sources.
type Client interface {
	Ready(ctx context.Context) error
	ListInbounds(ctx context.Context) ([]Inbound, error)
	ServerStatus(ctx context.Context) (*ServerStatus, error)
	GetInbound(ctx context.Context, id int) (*Inbound, error)
	CreateInbound(ctx context.Context, payload InboundPayload) (*Inbound, error)
	UpdateInbound(ctx context.Context, id int, payload InboundPayload) (*Inbound, error)
	DeleteInbound(ctx context.Context, id int) error
	AddClient(ctx context.Context, inboundID int, client InboundClient) error
	UpdateClient(ctx context.Context, inboundID int, clientID string, client InboundClient) error
	DeleteClient(ctx context.Context, inboundID int, clientID string) error
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// client is the concrete implementation.
type client struct {
	cfg          Config
	httpClient   httpClient
	nativeClient *http.Client

	loginMu  sync.Mutex
	loggedIn bool
}

// New creates a new API client with the provided configuration.
func New(cfg Config) (Client, error) {
	if cfg.BaseURL == nil {
		return nil, fmt.Errorf("base URL must be set")
	}

	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 30 * time.Second
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 1 * time.Second
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "terraform-provider-3x-ui/dev"
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	if cfg.TLSSkipVerify {
		configureInsecureTransport(transport)
	}

	var httpClient httpClient
	var nativeClient *http.Client

	if cfg.HTTPClient != nil {
		httpClient = cfg.HTTPClient
		if c, ok := cfg.HTTPClient.(*http.Client); ok {
			nativeClient = c
		}
	} else {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, fmt.Errorf("unable to init cookie jar: %w", err)
		}

		nativeClient = &http.Client{
			Transport: transport,
			Jar:       jar,
			Timeout:   cfg.RequestTimeout,
		}
		httpClient = nativeClient
	}

	c := &client{
		cfg:          cfg,
		httpClient:   httpClient,
		nativeClient: nativeClient,
	}

	if cfg.SessionCookie != nil {
		c.setSessionCookie(*cfg.SessionCookie)
		c.loggedIn = true
	}

	return c, nil
}

func (c *client) Ready(ctx context.Context) error {
	c.loginMu.Lock()
	defer c.loginMu.Unlock()

	if c.loggedIn {
		return nil
	}

	if c.cfg.SessionCookie != nil {
		c.loggedIn = true
		return nil
	}

	if err := c.performLogin(ctx); err != nil {
		return err
	}
	c.loggedIn = true
	return nil
}

func (c *client) performLogin(ctx context.Context) error {
	if c.cfg.SessionCookie != nil {
		return nil
	}

	if c.cfg.Username == nil || c.cfg.Password == nil {
		return fmt.Errorf("username and password are required for login")
	}

	form := url.Values{}
	form.Set("username", *c.cfg.Username)
	form.Set("password", *c.cfg.Password)

	loginURL := c.cfg.BaseURL.ResolveReference(&url.URL{Path: "/login"})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("build login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if userAgent := c.cfg.UserAgent; userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read login response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed: status=%d body=%s", resp.StatusCode, truncateBody(data))
	}

	var envelope responseEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return fmt.Errorf("decode login response: %w", err)
	}
	if !envelope.Success {
		return fmt.Errorf("login failed: %s", envelope.Msg)
	}

	return nil
}

func (c *client) setSessionCookie(value string) {
	if c.nativeClient == nil || c.nativeClient.Jar == nil {
		return
	}
	cookie := &http.Cookie{
		Name:  "session",
		Value: value,
		Path:  "/",
	}
	c.nativeClient.Jar.SetCookies(c.cfg.BaseURL, []*http.Cookie{cookie})
}

func configureInsecureTransport(t *http.Transport) {
	if t.TLSClientConfig == nil {
		t.TLSClientConfig = &tls.Config{}
	}
	t.TLSClientConfig.InsecureSkipVerify = true
}

func (c *client) ListInbounds(ctx context.Context) ([]Inbound, error) {
	var result []Inbound
	if err := c.getJSON(ctx, "/panel/api/inbounds/list", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *client) ServerStatus(ctx context.Context) (*ServerStatus, error) {
	var status ServerStatus
	if err := c.getJSON(ctx, "/panel/api/server/status", nil, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

func (c *client) getJSON(ctx context.Context, path string, query url.Values, out any) error {
	return c.request(ctx, http.MethodGet, path, query, nil, "", out)
}

func (c *client) postJSON(ctx context.Context, path string, payload any, out any) error {
	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal payload: %w", err)
		}
	}
	return c.request(ctx, http.MethodPost, path, nil, body, "application/json", out)
}

func (c *client) request(ctx context.Context, method, path string, query url.Values, body []byte, contentType string, out any) error {
	if err := c.Ready(ctx); err != nil {
		return err
	}

	maxAttempts := c.cfg.MaxRetries + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := c.buildRequest(ctx, method, path, query, body, contentType)
		if err != nil {
			return err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt == maxAttempts || !isRetryableError(err) {
				return fmt.Errorf("request failed: %w", err)
			}
			time.Sleep(backoffDuration(attempt))
			continue
		}

		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("read response: %w", readErr)
		}

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusNotFound {
			if err := c.handleUnauthorized(ctx); err != nil {
				return err
			}
			continue
		}

		if shouldRetryStatus(resp.StatusCode) {
			lastErr = fmt.Errorf("retryable status: %d", resp.StatusCode)
			if attempt == maxAttempts {
				return fmt.Errorf("request failed: status=%d body=%s", resp.StatusCode, truncateBody(data))
			}
			time.Sleep(backoffDuration(attempt))
			continue
		}

		if resp.StatusCode >= 400 {
			return fmt.Errorf("request failed: status=%d body=%s", resp.StatusCode, truncateBody(data))
		}

		var envelope responseEnvelope
		if err := json.Unmarshal(data, &envelope); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
		if !envelope.Success {
			return &APIError{StatusCode: resp.StatusCode, Message: envelope.Msg}
		}
		if out != nil && len(envelope.Obj) > 0 {
			if err := json.Unmarshal(envelope.Obj, out); err != nil {
				return fmt.Errorf("decode payload: %w", err)
			}
		}
		return nil
	}

	return lastErr
}

func (c *client) buildRequest(ctx context.Context, method, path string, query url.Values, body []byte, contentType string) (*http.Request, error) {
	endpoint := c.cfg.BaseURL.ResolveReference(&url.URL{Path: path})
	if len(query) > 0 {
		endpoint.RawQuery = query.Encode()
	}

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), reader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	if contentType != "" && body != nil {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Accept", "application/json")
	if c.cfg.UserAgent != "" {
		req.Header.Set("User-Agent", c.cfg.UserAgent)
	}

	return req, nil
}

func (c *client) handleUnauthorized(ctx context.Context) error {
	c.invalidateSession()
	if c.cfg.SessionCookie != nil {
		return fmt.Errorf("provided session cookie is invalid or expired")
	}
	return c.Ready(ctx)
}

func (c *client) invalidateSession() {
	c.loginMu.Lock()
	defer c.loginMu.Unlock()
	c.loggedIn = false
	if c.nativeClient != nil && c.nativeClient.Jar != nil {
		c.nativeClient.Jar.SetCookies(c.cfg.BaseURL, nil)
	}
}

func shouldRetryStatus(status int) bool {
	if status == http.StatusTooManyRequests {
		return true
	}
	return status >= 500 && status <= 599
}

func isRetryableError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout() || netErr.Temporary()
	}
	return false
}

func backoffDuration(attempt int) time.Duration {
	base := 500 * time.Millisecond
	jitter := time.Duration(rand.Int63n(int64(250 * time.Millisecond)))
	return time.Duration(attempt)*base + jitter
}

func truncateBody(data []byte) string {
	limit := 256
	if len(data) > limit {
		data = data[:limit]
	}
	return strings.TrimSpace(string(data))
}
