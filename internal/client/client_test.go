package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestClientInboundLifecycle(t *testing.T) {
	handler, state := newMockAPIHandler()
	httpClient := newHandlerClient(handler)

	baseURL, _ := url.Parse("http://example.com")

	client, err := New(Config{
		BaseURL:        baseURL,
		Username:       ptr("admin"),
		Password:       ptr("secret"),
		RequestTimeout: time.Second,
		MaxRetries:     1,
		PollInterval:   100 * time.Millisecond,
		UserAgent:      "test-client",
		HTTPClient:     httpClient,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	ctx := context.Background()
	created, err := client.CreateInbound(ctx, InboundPayload{
		Remark:   "new inbound",
		Protocol: "vless",
		Settings: json.RawMessage(`{"clients":[]}`),
	})
	if err != nil {
		t.Fatalf("create inbound: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected non-zero ID")
	}
	if _, exists := state.inbounds[created.ID]; !exists {
		t.Fatalf("inbound not stored in state")
	}

	_, err = client.UpdateInbound(ctx, created.ID, InboundPayload{
		Remark:   "updated",
		Protocol: "vless",
		Settings: json.RawMessage(`{"clients":[]}`),
	})
	if err != nil {
		t.Fatalf("update inbound: %v", err)
	}
	if state.inbounds[created.ID].Remark != "updated" {
		t.Fatalf("remark not updated")
	}

	if err := client.DeleteInbound(ctx, created.ID); err != nil {
		t.Fatalf("delete inbound: %v", err)
	}
	if _, exists := state.inbounds[created.ID]; exists {
		t.Fatalf("inbound not deleted from state")
	}
}

func TestClientServerStatus(t *testing.T) {
	handler, _ := newMockAPIHandler()
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

func TestClientAddUpdateDeleteClient(t *testing.T) {
	handler, state := newMockAPIHandler()
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

	ctx := context.Background()
	payload := InboundPayload{
		Remark:   "state inbound",
		Protocol: "vless",
		Settings: json.RawMessage(`{"clients":[]}`),
	}
	created, err := client.CreateInbound(ctx, payload)
	if err != nil {
		t.Fatalf("create inbound: %v", err)
	}

	user := InboundClient{ID: "uuid", Email: "user@example.com", Enable: true}
	if err := client.AddClient(ctx, created.ID, user); err != nil {
		t.Fatalf("add client: %v", err)
	}
	if state.lastClientAction != "add" {
		t.Fatalf("expected add action")
	}

	user.Comment = "updated"
	if err := client.UpdateClient(ctx, created.ID, "uuid", user); err != nil {
		t.Fatalf("update client: %v", err)
	}
	if state.lastClientAction != "update" {
		t.Fatalf("expected update action")
	}

	if err := client.DeleteClient(ctx, created.ID, "uuid"); err != nil {
		t.Fatalf("delete client: %v", err)
	}
	if state.lastClientAction != "delete" {
		t.Fatalf("expected delete action")
	}
}

func newHandlerClient(handler http.Handler) *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Transport: handlerRoundTripper{handler: handler},
		Jar:       jar,
	}
}

type mockAPIState struct {
	authenticated    bool
	nextInboundID    int
	inbounds         map[int]*Inbound
	lastClientAction string
}

func newMockAPIHandler() (http.Handler, *mockAPIState) {
	state := &mockAPIState{
		nextInboundID: 2,
		inbounds: map[int]*Inbound{
			1: {
				ID:       1,
				Remark:   "demo",
				Port:     443,
				Protocol: "vless",
				Settings: json.RawMessage(`{"clients":[]}`),
			},
		},
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.FormValue("username") != "admin" || r.FormValue("password") != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		state.authenticated = true
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "valid", Path: "/"})
		writeEnvelope(w, map[string]any{"ok": true})
	})

	mux.HandleFunc("/panel/api/inbounds/list", func(w http.ResponseWriter, r *http.Request) {
		if !state.authenticated {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var list []Inbound
		for _, inb := range state.inbounds {
			list = append(list, *inb)
		}
		writeEnvelope(w, list)
	})

	mux.HandleFunc("/panel/api/inbounds/add", func(w http.ResponseWriter, r *http.Request) {
		if !state.authenticated {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var payload InboundPayload
		json.NewDecoder(r.Body).Decode(&payload)
		inb := &Inbound{
			ID:       state.nextInboundID,
			Remark:   payload.Remark,
			Protocol: payload.Protocol,
			Settings: payload.Settings,
		}
		state.inbounds[inb.ID] = inb
		state.nextInboundID++
		writeEnvelope(w, inb)
	})

	mux.HandleFunc("/panel/api/inbounds/update/", func(w http.ResponseWriter, r *http.Request) {
		if !state.authenticated {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		parts := strings.Split(r.URL.Path, "/")
		idStr := parts[len(parts)-1]
		var payload InboundPayload
		json.NewDecoder(r.Body).Decode(&payload)
		id := atoi(idStr)
		if inb, ok := state.inbounds[id]; ok {
			inb.Remark = payload.Remark
			inb.Settings = payload.Settings
			writeEnvelope(w, inb)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/panel/api/inbounds/del/", func(w http.ResponseWriter, r *http.Request) {
		if !state.authenticated {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		parts := strings.Split(r.URL.Path, "/")
		id := atoi(parts[len(parts)-1])
		delete(state.inbounds, id)
		writeEnvelope(w, map[string]any{"ok": true})
	})

	mux.HandleFunc("/panel/api/server/status", func(w http.ResponseWriter, r *http.Request) {
		if !state.authenticated {
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

	mux.HandleFunc("/panel/api/inbounds/addClient", func(w http.ResponseWriter, r *http.Request) {
		if !state.authenticated {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		state.lastClientAction = "add"
		writeEnvelope(w, map[string]any{"ok": true})
	})

	mux.HandleFunc("/panel/api/inbounds/updateClient/", func(w http.ResponseWriter, r *http.Request) {
		if !state.authenticated {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		state.lastClientAction = "update"
		writeEnvelope(w, map[string]any{"ok": true})
	})

	mux.HandleFunc("/panel/api/inbounds/", func(w http.ResponseWriter, r *http.Request) {
		if !state.authenticated {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if strings.Contains(r.URL.Path, "/delClient/") {
			state.lastClientAction = "delete"
			writeEnvelope(w, map[string]any{"ok": true})
			return
		}
	})

	return mux, state
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

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
