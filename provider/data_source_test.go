package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientGetClientTraffics(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Go's net/http decodes percent-encoded path segments, so r.URL.Path contains the decoded form.
		if r.URL.Path != "/panel/api/inbounds/getClientTraffics/test@example.com" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(okResponse(ClientTraffic{
			ID: 1, InboundID: 5, Email: "test@example.com",
			Up: 100, Down: 200, Total: 1000, ExpiryTime: 9999, Enable: true,
		}))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	traffic, err := client.GetClientTraffics(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("GetClientTraffics failed: %v", err)
	}
	if traffic.ID != 1 || traffic.Email != "test@example.com" || traffic.Up != 100 || traffic.InboundID != 5 {
		t.Fatalf("unexpected traffic: %#v", traffic)
	}
}

func TestClientGetClientTrafficsPathEscape(t *testing.T) {
	var receivedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.RawPath
		if receivedPath == "" {
			receivedPath = r.URL.Path
		}
		w.Write(okResponse(ClientTraffic{ID: 1, Email: "user/slash"}))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	setLegacyClientAPI(client)
	_, err := client.GetClientTraffics(context.Background(), "user/slash")
	if err != nil {
		t.Fatalf("GetClientTraffics failed: %v", err)
	}
	expected := "/panel/api/inbounds/getClientTraffics/user%2Fslash"
	if receivedPath != expected {
		t.Fatalf("path not escaped: got %q, want %q", receivedPath, expected)
	}
}

func TestClientGetClientTrafficsNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write(okResponse(nil))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	_, err := client.GetClientTraffics(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent client, got nil")
	}
}

func TestClientGetClientTrafficsEmptyEmail(t *testing.T) {
	client := newTestClient(t, "http://localhost:0")
	_, err := client.GetClientTraffics(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty email, got nil")
	}
}

func TestClientGetInbounds(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/inbounds/list" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(okResponse([]Inbound{{ID: 1, Port: 1234, Protocol: "vmess", Settings: "{}"}}))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	inbounds, err := client.GetInbounds(context.Background())
	if err != nil {
		t.Fatalf("GetInbounds failed: %v", err)
	}
	if len(inbounds) != 1 || inbounds[0].ID != 1 {
		t.Fatalf("unexpected inbounds: %#v", inbounds)
	}
}

func TestClientGetServerStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/server/status" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(okResponse(map[string]any{"cpu": 12}))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	status, err := client.GetServerStatus(context.Background())
	if err != nil {
		t.Fatalf("GetServerStatus failed: %v", err)
	}
	if status["cpu"] != float64(12) {
		t.Fatalf("unexpected status: %#v", status)
	}
}

func TestClientGetXrayVersions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/server/getXrayVersion" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(okResponse([]string{"v1", "v2"}))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	versions, err := client.GetXrayVersions(context.Background())
	if err != nil {
		t.Fatalf("GetXrayVersions failed: %v", err)
	}
	if len(versions) != 2 || versions[0] != "v1" {
		t.Fatalf("unexpected versions: %#v", versions)
	}
}

func TestClientGetXrayConfig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/server/getConfigJson" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(okResponse(map[string]any{"inbounds": []any{}}))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	config, err := client.GetXrayConfig(context.Background())
	if err != nil {
		t.Fatalf("GetXrayConfig failed: %v", err)
	}
	if _, ok := config["inbounds"]; !ok {
		t.Fatalf("unexpected config: %#v", config)
	}
}

func TestClientGetSettings(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/setting/all" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(okResponse(map[string]any{"webPort": 2053}))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	settings, err := client.GetSettings(context.Background())
	if err != nil {
		t.Fatalf("GetSettings failed: %v", err)
	}
	if settings["webPort"] != float64(2053) {
		t.Fatalf("unexpected settings: %#v", settings)
	}
}
