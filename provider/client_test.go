package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflogtest"
)

func okResponse(obj any) []byte {
	resp := apiResponse{Success: true}
	if obj != nil {
		b, _ := json.Marshal(obj)
		resp.Obj = b
	}
	b, _ := json.Marshal(resp)
	return b
}

func failResponse(msg string) []byte {
	b, _ := json.Marshal(apiResponse{Success: false, Msg: msg})
	return b
}

func newTestClient(t *testing.T, endpoint string) *Client {
	t.Helper()
	// Mirror the provider's default max_retries=1 so existing tests cover
	// the retry path with the same configuration users get out of the box.
	return newTestClientWithRetries(t, endpoint, 1)
}

func newTestClientWithRetries(t *testing.T, endpoint string, maxRetries int) *Client {
	t.Helper()
	c, err := NewClient(ClientConfig{
		Endpoint:           endpoint,
		BasePath:           "/",
		Username:           "admin",
		Password:           "admin",
		InsecureSkipVerify: true,
		Timeout:            2 * time.Second,
		MaxRetries:         maxRetries,
	})
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	return c
}

func TestLoginSuccess(t *testing.T) {
	var loginCalls int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		atomic.AddInt32(&loginCalls, 1)
		r.ParseForm()
		if r.FormValue("username") != "admin" || r.FormValue("password") != "admin" {
			w.Write(failResponse("bad creds"))
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: "sess"})
		w.Write(okResponse(nil))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	if err := client.Login(context.Background()); err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if atomic.LoadInt32(&loginCalls) != 1 {
		t.Fatalf("expected 1 login call, got %d", loginCalls)
	}
}

func TestAutoReloginOn404(t *testing.T) {
	var loginCalls int32
	var apiCalls int32
	var authed atomic.Bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			atomic.AddInt32(&loginCalls, 1)
			authed.Store(true)
			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: "sess"})
			w.Write(okResponse(nil))
		case "/panel/api/server/status":
			atomic.AddInt32(&apiCalls, 1)
			if !authed.Load() {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Write(okResponse(map[string]any{"ok": true}))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)

	var out map[string]any
	if err := client.doJSON(context.Background(), http.MethodGet, "panel/api/server/status", nil, &out); err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if out["ok"] != true {
		t.Fatalf("unexpected response: %#v", out)
	}
	if atomic.LoadInt32(&loginCalls) != 1 {
		t.Fatalf("expected 1 login call, got %d", loginCalls)
	}
	if atomic.LoadInt32(&apiCalls) != 2 {
		t.Fatalf("expected 2 api calls (fail+retry), got %d", apiCalls)
	}
}

func TestAutoReloginOn401(t *testing.T) {
	var loginCalls int32
	var apiCalls int32
	var authed atomic.Bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			atomic.AddInt32(&loginCalls, 1)
			authed.Store(true)
			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: "sess"})
			w.Write(okResponse(nil))
		case "/panel/api/server/status":
			atomic.AddInt32(&apiCalls, 1)
			if !authed.Load() {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Write(okResponse(map[string]any{"ok": true}))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)

	var out map[string]any
	if err := client.doJSON(context.Background(), http.MethodGet, "panel/api/server/status", nil, &out); err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if out["ok"] != true {
		t.Fatalf("unexpected response: %#v", out)
	}
	if atomic.LoadInt32(&loginCalls) != 1 {
		t.Fatalf("expected 1 login call, got %d", loginCalls)
	}
	if atomic.LoadInt32(&apiCalls) != 2 {
		t.Fatalf("expected 2 api calls (fail+retry), got %d", apiCalls)
	}
}

func TestLoginFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(failResponse("wrongUsernameOrPassword"))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	if err := client.Login(context.Background()); err == nil {
		t.Fatalf("expected login error")
	}
}

func TestUpdateInboundMaxRetriesZeroDisablesRetry(t *testing.T) {
	// Operators must be able to opt out of retry entirely (max_retries=0)
	// so a transient 5xx surfaces immediately without any silent retry.
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/inbounds/update/7" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		atomic.AddInt32(&calls, 1)
		http.Error(w, "transient", http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestClientWithRetries(t, srv.URL, 0)
	if _, err := client.UpdateInbound(context.Background(), &Inbound{ID: 7}); err == nil {
		t.Fatalf("expected 5xx to surface when max_retries=0")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected exactly 1 call (no retry when max_retries=0), got %d", got)
	}
}

func TestUpdateInboundEmitsRetryWarnLog(t *testing.T) {
	// The retry path's promise of observability is part of the contract:
	// every retry must emit a Warn entry with operation, attempt,
	// max_attempts, status_code, and backoff. Verify it actually happens.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/inbounds/update/7" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		http.Error(w, "transient", http.StatusInternalServerError)
	}))
	defer srv.Close()

	var output bytes.Buffer
	ctx := tflogtest.RootLogger(context.Background(), &output)

	client := newTestClient(t, srv.URL)
	if _, err := client.UpdateInbound(ctx, &Inbound{ID: 7}); err == nil {
		t.Fatalf("expected persistent 5xx to surface")
	}

	entries, err := tflogtest.MultilineJSONDecode(&output)
	if err != nil {
		t.Fatalf("MultilineJSONDecode: %v", err)
	}
	var warn map[string]any
	for _, e := range entries {
		if e["@level"] == "warn" && e["@message"] == "retrying transient 5xx" {
			warn = e
			break
		}
	}
	if warn == nil {
		t.Fatalf("expected a warn entry for the retry; got entries: %v", entries)
	}
	for _, key := range []string{"operation", "attempt", "max_attempts", "status_code", "backoff"} {
		if _, ok := warn[key]; !ok {
			t.Errorf("retry warn entry missing %q field: %v", key, warn)
		}
	}
	if op, _ := warn["operation"].(string); op != "POST panel/api/inbounds/update/7" {
		t.Errorf("unexpected operation field: %v", warn["operation"])
	}
	if code, _ := warn["status_code"].(float64); code != 500 {
		t.Errorf("unexpected status_code field: %v", warn["status_code"])
	}
}

func TestUpdateInboundMaxRetriesTwo(t *testing.T) {
	// Two retries: 500, 500, 200 → succeed on third attempt, exactly 3 calls.
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/inbounds/update/7" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		n := atomic.AddInt32(&calls, 1)
		if n <= 2 {
			http.Error(w, "transient", http.StatusInternalServerError)
			return
		}
		w.Write(okResponse(&Inbound{ID: 7}))
	}))
	defer srv.Close()

	client := newTestClientWithRetries(t, srv.URL, 2)
	if _, err := client.UpdateInbound(context.Background(), &Inbound{ID: 7}); err != nil {
		t.Fatalf("UpdateInbound after 500-500-200 failed: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("expected 3 calls (initial + 2 retries), got %d", got)
	}
}

func TestUpdateInboundRetriesTransient5xx(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/inbounds/update/7" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			http.Error(w, "transient", http.StatusInternalServerError)
			return
		}
		w.Write(okResponse(&Inbound{ID: 7}))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	if _, err := client.UpdateInbound(context.Background(), &Inbound{ID: 7}); err != nil {
		t.Fatalf("UpdateInbound after 500-then-200 failed: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected 2 calls (1 retry on 5xx), got %d", got)
	}
}

func TestUpdateInboundSurfacesPersistent5xx(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		http.Error(w, "still broken", http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	_, err := client.UpdateInbound(context.Background(), &Inbound{ID: 9})
	if err == nil {
		t.Fatalf("expected error after persistent 5xx")
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected 2 calls (initial + 1 retry), got %d", got)
	}
}

func TestUpdateInbound4xxIsNotRetried(t *testing.T) {
	// Non-5xx errors must surface immediately — only transient server-side
	// contention is retryable.
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	if _, err := client.UpdateInbound(context.Background(), &Inbound{ID: 1}); err == nil {
		t.Fatalf("expected 4xx to surface")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected exactly 1 call, got %d", got)
	}
}

func TestGetInboundDoesNotRetryOn5xx(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/inbounds/get/7" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		atomic.AddInt32(&calls, 1)
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	if _, err := client.GetInbound(context.Background(), 7); err == nil {
		t.Fatalf("expected 5xx GET error to surface")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected exactly 1 call (no retry on GET), got %d", got)
	}
}

func TestAddInboundDoesNotRetryOn5xx(t *testing.T) {
	// Retrying a non-idempotent create would risk creating a duplicate
	// inbound on the panel.
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/inbounds/add" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		atomic.AddInt32(&calls, 1)
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	if _, err := client.AddInbound(context.Background(), &Inbound{Port: 1, Protocol: "vmess", Settings: "{}"}); err == nil {
		t.Fatalf("expected 5xx error to surface")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected exactly 1 call (no retry on Add), got %d", got)
	}
}

func TestUpdateUserDoesNotRetryOn5xx(t *testing.T) {
	// Retrying credentials change with stale old creds (the second call
	// would still send oldUsername/oldPassword from the first attempt) is
	// unsafe and would diverge provider state from the panel.
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/setting/updateUser" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		atomic.AddInt32(&calls, 1)
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	if err := client.UpdateUser(context.Background(), "admin", "admin", "newuser", "newpass"); err == nil {
		t.Fatalf("expected 5xx to surface")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected exactly 1 call (no retry on UpdateUser), got %d", got)
	}
}

func TestUpdateInboundRetryRespectsCtxCancel(t *testing.T) {
	// Context cancellation during the 500ms backoff must short-circuit the
	// retry, not block the caller.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "transient", http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, err := client.UpdateInbound(ctx, &Inbound{ID: 1}); err == nil {
		t.Fatalf("expected ctx error")
	}
	if elapsed := time.Since(start); elapsed > 400*time.Millisecond {
		t.Fatalf("backoff did not respect ctx cancel; elapsed=%v", elapsed)
	}
}

func TestUpdateInbound5xxThenReloginThenSuccess(t *testing.T) {
	// First call: 401 → re-login → still 500 (transient) → backoff → retry → 200.
	// Verifies the 5xx-retry composes with the existing 401-relogin flow.
	var n int32
	var loginCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			atomic.AddInt32(&loginCalls, 1)
			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: "sess"})
			w.Write(okResponse(nil))
		case "/panel/api/inbounds/update/7":
			c := atomic.AddInt32(&n, 1)
			switch c {
			case 1:
				w.WriteHeader(http.StatusUnauthorized)
			case 2:
				http.Error(w, "transient", http.StatusInternalServerError)
			default:
				w.Write(okResponse(&Inbound{ID: 7}))
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	if _, err := client.UpdateInbound(context.Background(), &Inbound{ID: 7}); err != nil {
		t.Fatalf("UpdateInbound: %v", err)
	}
	if got := atomic.LoadInt32(&n); got != 3 {
		t.Fatalf("expected 3 update calls (401, 500, 200), got %d", got)
	}
	if got := atomic.LoadInt32(&loginCalls); got != 1 {
		t.Fatalf("expected 1 login call, got %d", got)
	}
}

func TestUpdateInboundClientRetriesTransient5xx(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/inbounds/updateClient/abc" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			http.Error(w, "transient", http.StatusInternalServerError)
			return
		}
		w.Write(okResponse(nil))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	if err := client.UpdateInboundClient(context.Background(), 7, "abc", map[string]any{"id": "abc"}); err != nil {
		t.Fatalf("UpdateInboundClient after 500-then-200 failed: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected 2 calls (1 retry on 5xx), got %d", got)
	}
}

func TestUpdateSettingsRetriesTransient5xx(t *testing.T) {
	// Covers doJSONRetryable: same retry shape as the form path, but
	// exercises the JSON encoding/content-type branch.
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/setting/update" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			http.Error(w, "transient", http.StatusInternalServerError)
			return
		}
		w.Write(okResponse(nil))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	if err := client.UpdateSettings(context.Background(), map[string]any{"webPort": 2053}); err != nil {
		t.Fatalf("UpdateSettings after 500-then-200 failed: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected 2 calls (1 retry on 5xx), got %d", got)
	}
}

func TestDeleteInboundDoesNotRetryOn5xx(t *testing.T) {
	// 3x-ui's DelInbound looks up the row first and errors on a missing
	// one, so a retry after a successful-but-5xx delete would turn success
	// into failure.
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/inbounds/del/7" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		atomic.AddInt32(&calls, 1)
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	if err := client.DeleteInbound(context.Background(), 7); err == nil {
		t.Fatalf("expected 5xx to surface")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected exactly 1 call (no retry on Delete), got %d", got)
	}
}

func TestAddInboundClientDoesNotRetryOn5xx(t *testing.T) {
	// AddClient is non-idempotent — a retry could create a duplicate
	// client on the panel.
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/inbounds/addClient" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		atomic.AddInt32(&calls, 1)
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	if err := client.AddInboundClient(context.Background(), 7, map[string]any{"id": "abc"}); err == nil {
		t.Fatalf("expected 5xx to surface")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected exactly 1 call (no retry on AddClient), got %d", got)
	}
}

func TestResolvePathWithBasePath(t *testing.T) {
	client := newTestClient(t, "http://example.com")
	client.basePath = "/xui/"

	endpoint, err := client.resolvePath("panel/api/inbounds/list")
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if parsed.Path != "/xui/panel/api/inbounds/list" {
		t.Fatalf("unexpected path: %s", parsed.Path)
	}
}
