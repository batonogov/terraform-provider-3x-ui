package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// setLegacyClientAPI forces the client to use old v2.9.x/v3.0.x endpoints,
// skipping the v3.1.0+ probe.
func setLegacyClientAPI(c *Client) {
	v := false
	c.newClientAPI = &v
}

func TestClientAddInboundClient(t *testing.T) {
	var gotSettings string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/inbounds/addClient" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		r.ParseForm()
		gotSettings = r.FormValue("settings")
		w.Write(okResponse(nil))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	setLegacyClientAPI(client)
	data := map[string]any{"id": "uuid", "email": "a@b"}
	if err := client.AddInboundClient(context.Background(), 1, data); err != nil {
		t.Fatalf("AddInboundClient failed: %v", err)
	}
	var payload map[string][]map[string]any
	if err := json.Unmarshal([]byte(gotSettings), &payload); err != nil {
		t.Fatalf("invalid settings json: %v", err)
	}
	if len(payload["clients"]) != 1 {
		t.Fatalf("expected 1 client, got %d", len(payload["clients"]))
	}
}

func TestClientAddInboundClientNewAPI(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.Method == http.MethodPost && r.URL.Path == "/panel/api/clients/add" {
			json.NewDecoder(r.Body).Decode(&gotBody)
		}
		w.Write(okResponse(nil))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	data := map[string]any{"id": "uuid", "email": "a@b"}
	if err := client.AddInboundClient(context.Background(), 1, data); err != nil {
		t.Fatalf("AddInboundClient failed: %v", err)
	}
	if gotPath != "/panel/api/clients/add" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotBody["client"] == nil {
		t.Fatal("expected client field in payload")
	}
	ids, ok := gotBody["inboundIds"].([]any)
	if !ok || len(ids) != 1 {
		t.Fatalf("expected inboundIds with 1 element, got %v", gotBody["inboundIds"])
	}
}

func TestClientUpdateInboundClient(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Write(okResponse(nil))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	setLegacyClientAPI(client)
	data := map[string]any{"id": "uuid", "email": "a@b"}
	if err := client.UpdateInboundClient(context.Background(), 2, "uuid", data); err != nil {
		t.Fatalf("UpdateInboundClient failed: %v", err)
	}
	if gotPath != "/panel/api/inbounds/updateClient/uuid" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
}

func TestClientUpdateInboundClientNewAPI(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Write(okResponse(nil))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	data := map[string]any{"id": "uuid", "email": "a@b"}
	if err := client.UpdateInboundClient(context.Background(), 2, "uuid", data); err != nil {
		t.Fatalf("UpdateInboundClient failed: %v", err)
	}
	if gotPath != "/panel/api/clients/update/a@b" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
}

func TestClientDeleteInboundClient(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Write(okResponse(nil))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	setLegacyClientAPI(client)
	if err := client.DeleteInboundClient(context.Background(), 9, "cid", "test@example.com"); err != nil {
		t.Fatalf("DeleteInboundClient failed: %v", err)
	}
	if gotPath != "/panel/api/inbounds/9/delClient/cid" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
}

func TestClientDeleteInboundClientNewAPI(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Write(okResponse(nil))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	if err := client.DeleteInboundClient(context.Background(), 9, "cid", "test@example.com"); err != nil {
		t.Fatalf("DeleteInboundClient failed: %v", err)
	}
	if gotPath != "/panel/api/clients/del/test@example.com" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
}

func TestClientDeleteInboundClientIgnoresMissingClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(failResponse("Client Not Found In Inbound For ID: cid"))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	setLegacyClientAPI(client)
	if err := client.DeleteInboundClient(context.Background(), 9, "cid", "test@example.com"); err != nil {
		t.Fatalf("expected missing client delete to be idempotent, got %v", err)
	}
}
