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
	if err := client.UpdateInboundClient(context.Background(), 2, "uuid", "", data); err != nil {
		t.Fatalf("UpdateInboundClient failed: %v", err)
	}
	if gotPath != "/panel/api/inbounds/updateClient/uuid" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
}

func TestClientUpdateInboundClientNewAPI(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.Method == http.MethodPost && r.URL.Path == "/panel/api/clients/update/a@b" {
			json.NewDecoder(r.Body).Decode(&gotBody)
		}
		w.Write(okResponse(nil))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	data := map[string]any{"id": "uuid", "email": "a@b"}
	if err := client.UpdateInboundClient(context.Background(), 2, "uuid", "a@b", data); err != nil {
		t.Fatalf("UpdateInboundClient failed: %v", err)
	}
	if gotPath != "/panel/api/clients/update/a@b" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotBody["id"] != "uuid" || gotBody["email"] != "a@b" {
		t.Fatalf("expected client data in body, got %v", gotBody)
	}
}

func TestClientUpdateInboundClientNewAPIEmailChange(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.Method == http.MethodPost {
			json.NewDecoder(r.Body).Decode(&gotBody)
		}
		w.Write(okResponse(nil))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	data := map[string]any{"id": "uuid", "email": "new@b"}
	if err := client.UpdateInboundClient(context.Background(), 2, "uuid", "old@b", data); err != nil {
		t.Fatalf("UpdateInboundClient failed: %v", err)
	}
	// URL must use the OLD email to look up the existing record
	if gotPath != "/panel/api/clients/update/old@b" {
		t.Fatalf("unexpected path: %s (expected /panel/api/clients/update/old@b)", gotPath)
	}
	// Body must contain the NEW email
	if gotBody["email"] != "new@b" {
		t.Fatalf("expected new email in body, got %v", gotBody["email"])
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

func TestClientDeleteInboundClientNewAPIIgnoresMissingClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panel/api/clients/del/test@example.com" {
			w.Write(failResponse("Client Not Found"))
			return
		}
		w.Write(okResponse(nil))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	if err := client.DeleteInboundClient(context.Background(), 9, "cid", "test@example.com"); err != nil {
		t.Fatalf("expected missing client delete to be idempotent, got %v", err)
	}
}

func TestClientDeleteInboundClientEmptyEmailUsesOldEndpoint(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Write(okResponse(nil))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	if err := client.DeleteInboundClient(context.Background(), 9, "cid", ""); err != nil {
		t.Fatalf("DeleteInboundClient failed: %v", err)
	}
	if gotPath != "/panel/api/inbounds/9/delClient/cid" {
		t.Fatalf("expected old endpoint, got path: %s", gotPath)
	}
}

func TestGetOnlineClientsNewAPI(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.URL.Path == "/panel/api/clients/onlines" {
			w.Write(okResponse([]string{"a@b", "c@d"}))
			return
		}
		w.Write(okResponse(nil))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	clients, err := client.GetOnlineClients(context.Background())
	if err != nil {
		t.Fatalf("GetOnlineClients failed: %v", err)
	}
	if gotPath != "/panel/api/clients/onlines" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if len(clients) != 2 || clients[0] != "a@b" || clients[1] != "c@d" {
		t.Fatalf("unexpected clients: %v", clients)
	}
}

func TestGetOnlineClientsNewAPIEmptyResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panel/api/clients/onlines" {
			w.Write(okResponse([]string{}))
			return
		}
		w.Write(okResponse(nil))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	clients, err := client.GetOnlineClients(context.Background())
	if err != nil {
		t.Fatalf("GetOnlineClients failed: %v", err)
	}
	if clients == nil || len(clients) != 0 {
		t.Fatalf("expected empty non-nil slice, got %v", clients)
	}
}

func TestGetClientTrafficsNewAPI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panel/api/clients/traffic/test@example.com" {
			w.Write(okResponse(ClientTraffic{
				ID: 1, InboundID: 5, Email: "test@example.com",
				Up: 100, Down: 200, Total: 1000, ExpiryTime: 9999, Enable: true,
			}))
			return
		}
		w.Write(okResponse(nil))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	traffic, err := client.GetClientTraffics(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("GetClientTraffics failed: %v", err)
	}
	if traffic.ID != 1 || traffic.Email != "test@example.com" || traffic.Up != 100 {
		t.Fatalf("unexpected traffic: %#v", traffic)
	}
}

func TestGetClientTrafficsNewAPINotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(okResponse(nil))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	_, err := client.GetClientTraffics(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent client, got nil")
	}
}
