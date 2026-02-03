package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"
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
	c, err := NewClient(ClientConfig{
		Endpoint:           endpoint,
		BasePath:           "/",
		Username:           "admin",
		Password:           "admin",
		InsecureSkipVerify: true,
		Timeout:            2 * time.Second,
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
