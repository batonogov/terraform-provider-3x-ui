package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

func TestFlattenInbound(t *testing.T) {
	in := Inbound{ID: 1, Port: 1234, Protocol: "vmess", Settings: "{}", Enable: true}
	out := flattenInbound(in)
	if out["id"].(int) != 1 {
		t.Fatalf("unexpected id: %#v", out)
	}
	if out["port"].(int) != 1234 {
		t.Fatalf("unexpected port: %#v", out)
	}
}
