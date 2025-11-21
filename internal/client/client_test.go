package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestClientListInbounds(t *testing.T) {
	handler := newMockHandler()
	httpClient := newHandlerClient(handler)

	baseURL, _ := url.Parse("http://example.com")

	cfg := Config{
		BaseURL:        baseURL,
		Username:       ptr("admin"),
		Password:       ptr("secret"),
		RequestTimeout: time.Second,
		MaxRetries:     1,
		PollInterval:   100 * time.Millisecond,
		UserAgent:      "test-client",
		HTTPClient:     httpClient,
	}

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	ctx := context.Background()
	inbounds, err := client.ListInbounds(ctx)
	if err != nil {
		t.Fatalf("list inbounds: %v", err)
	}

	if len(inbounds) != 1 {
		t.Fatalf("expected 1 inbound, got %d", len(inbounds))
	}
	if inbounds[0].Remark != "demo" {
		t.Fatalf("unexpected inbound remark: %s", inbounds[0].Remark)
	}
}

func TestClientServerStatus(t *testing.T) {
	handler := newMockHandler()
	httpClient := newHandlerClient(handler)

	baseURL, _ := url.Parse("http://example.com")

	client, err := New(Config{
		BaseURL:        baseURL,
		Username:       ptr("admin"),
		Password:       ptr("secret"),
		RequestTimeout: time.Second,
		UserAgent:      "test-client",
		HTTPClient:     httpClient,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	status, err := client.ServerStatus(context.Background())
	if err != nil {
		t.Fatalf("server status: %v", err)
	}

	if status.CPU <= 0 {
		t.Fatalf("expected cpu usage > 0")
	}
	if status.Xray.Version != "1.0.0" {
		t.Fatalf("unexpected xray version: %s", status.Xray.Version)
	}
}

func newHandlerClient(handler http.Handler) *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Transport: handlerRoundTripper{handler: handler},
		Jar:       jar,
	}
}

func newMockHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.FormValue("username") != "admin" || r.FormValue("password") != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "valid", Path: "/"})
		writeEnvelope(w, map[string]any{"ok": true})
	})

	mux.HandleFunc("/panel/api/inbounds/list", func(w http.ResponseWriter, r *http.Request) {
		if !hasValidSession(r) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		writeEnvelope(w, []Inbound{
			{
				ID:     1,
				Remark: "demo",
				Port:   443,
			},
		})
	})

	mux.HandleFunc("/panel/api/server/status", func(w http.ResponseWriter, r *http.Request) {
		if !hasValidSession(r) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		writeEnvelope(w, ServerStatus{
			CPU:      1.2,
			CPUCores: 2,
			Xray: XrayState{
				Version: "1.0.0",
				State:   "running",
			},
		})
	})

	return mux
}

func hasValidSession(r *http.Request) bool {
	cookie, err := r.Cookie("session")
	return err == nil && cookie.Value == "valid"
}

type handlerRoundTripper struct {
	handler http.Handler
}

func (rt handlerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	rt.handler.ServeHTTP(rec, req)
	resp := rec.Result()
	resp.Request = req
	return resp, nil
}

func writeEnvelope(w http.ResponseWriter, obj any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"msg":     "ok",
		"obj":     obj,
	})
}

func ptr[T any](v T) *T {
	return &v
}
