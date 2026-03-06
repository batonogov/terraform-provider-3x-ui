package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

func TestClientUpdateInboundClient(t *testing.T) {
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
	if gotPath != "/panel/api/inbounds/updateClient/uuid" {
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
	if err := client.DeleteInboundClient(context.Background(), 9, "cid"); err != nil {
		t.Fatalf("DeleteInboundClient failed: %v", err)
	}
	if gotPath != "/panel/api/inbounds/9/delClient/cid" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
}

func TestClientDeleteInboundClientLastClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(failResponse("Something went wrong (no client remained in Inbound)"))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	err := client.DeleteInboundClient(context.Background(), 9, "cid")
	if err == nil {
		t.Fatal("expected error when deleting last client")
	}
	if !strings.Contains(err.Error(), "no client remained in Inbound") {
		t.Fatalf("unexpected error: %v", err)
	}
}
