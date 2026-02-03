package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestClientAddInbound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/inbounds/add" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.FormValue("port") != "1234" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.FormValue("protocol") != "vmess" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.FormValue("settings") != "{}" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Write(okResponse(&Inbound{ID: 1, Port: 1234, Protocol: "vmess", Settings: "{}", Enable: true}))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	inbound := &Inbound{
		Port:     1234,
		Protocol: "vmess",
		Settings: "{}",
		Enable:   true,
	}

	created, err := client.AddInbound(context.Background(), inbound)
	if err != nil {
		t.Fatalf("AddInbound failed: %v", err)
	}
	if created.ID != 1 {
		t.Fatalf("unexpected ID: %d", created.ID)
	}
}

func TestClientUpdateInbound(t *testing.T) {
	var gotID int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/panel/api/inbounds/update/10" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		gotID, _ = strconv.Atoi(r.FormValue("id"))
		w.Write(okResponse(&Inbound{ID: 10, Port: 443, Protocol: "vless", Settings: "{}", Enable: true}))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	inbound := &Inbound{
		ID:       10,
		Port:     443,
		Protocol: "vless",
		Settings: "{}",
		Enable:   true,
	}

	updated, err := client.UpdateInbound(context.Background(), inbound)
	if err != nil {
		t.Fatalf("UpdateInbound failed: %v", err)
	}
	if updated.ID != 10 || gotID != 10 {
		t.Fatalf("unexpected IDs: updated=%d form=%d", updated.ID, gotID)
	}
}

func TestClientGetInbound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/panel/api/inbounds/get/7" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(okResponse(&Inbound{ID: 7, Port: 8080, Protocol: "vmess", Settings: "{}"}))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	inbound, err := client.GetInbound(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetInbound failed: %v", err)
	}
	if inbound.ID != 7 || inbound.Port != 8080 {
		t.Fatalf("unexpected inbound: %#v", inbound)
	}
}

func TestClientDeleteInbound(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/panel/api/inbounds/del/5" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		called = true
		w.Write(okResponse(map[string]any{"id": 5}))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	if err := client.DeleteInbound(context.Background(), 5); err != nil {
		t.Fatalf("DeleteInbound failed: %v", err)
	}
	if !called {
		t.Fatalf("delete endpoint not called")
	}
}

func TestInboundToForm(t *testing.T) {
	in := &Inbound{
		ID:             1,
		Up:             2,
		Down:           3,
		Total:          4,
		Remark:         "r",
		Enable:         true,
		ExpiryTime:     5,
		TrafficReset:   "daily",
		Listen:         "0.0.0.0",
		Port:           1234,
		Protocol:       "vmess",
		Settings:       "{}",
		StreamSettings: "{}",
		Sniffing:       "{}",
	}

	form := inboundToForm(in)
	want := url.Values{
		"id":                   []string{"1"},
		"up":                   []string{"2"},
		"down":                 []string{"3"},
		"total":                []string{"4"},
		"remark":               []string{"r"},
		"enable":               []string{"true"},
		"expiryTime":           []string{"5"},
		"trafficReset":         []string{"daily"},
		"lastTrafficResetTime": []string{"0"},
		"listen":               []string{"0.0.0.0"},
		"port":                 []string{"1234"},
		"protocol":             []string{"vmess"},
		"settings":             []string{"{}"},
		"streamSettings":       []string{"{}"},
		"sniffing":             []string{"{}"},
	}
	if form.Encode() != want.Encode() {
		t.Fatalf("unexpected form: %s", form.Encode())
	}
}
