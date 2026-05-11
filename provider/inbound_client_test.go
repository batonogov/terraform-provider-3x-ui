package provider

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
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

// TestInboundResourceWaitForDeletion exercises the post-Delete poll added in
// #136. The helper checks the inbound list for the absence of id; it does NOT
// re-issue DELETE (DelInbound is not idempotent in 3x-ui).
func TestInboundResourceWaitForDeletion(t *testing.T) {
	t.Run("immediate-absent", func(t *testing.T) {
		var lists int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/panel/api/inbounds/list" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			atomic.AddInt32(&lists, 1)
			w.Write(okResponse([]Inbound{{ID: 7}}))
		}))
		defer srv.Close()

		res := &InboundResource{client: newTestClient(t, srv.URL)}
		if err := res.waitForInboundDeletion(context.Background(), 9); err != nil {
			t.Fatalf("waitForInboundDeletion: %v", err)
		}
		if got := atomic.LoadInt32(&lists); got != 1 {
			t.Fatalf("expected exactly 1 list call, got %d", got)
		}
	})

	t.Run("becomes-absent-after-poll", func(t *testing.T) {
		var lists int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/panel/api/inbounds/list" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			n := atomic.AddInt32(&lists, 1)
			if n <= 2 {
				w.Write(okResponse([]Inbound{{ID: 3}}))
				return
			}
			w.Write(okResponse([]Inbound{}))
		}))
		defer srv.Close()

		res := &InboundResource{client: newTestClient(t, srv.URL)}
		if err := res.waitForInboundDeletion(context.Background(), 3); err != nil {
			t.Fatalf("waitForInboundDeletion: %v", err)
		}
		if got := atomic.LoadInt32(&lists); got != 3 {
			t.Fatalf("expected exactly 3 list calls, got %d", got)
		}
	})

	t.Run("never-absent-returns-error", func(t *testing.T) {
		var lists int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/panel/api/inbounds/list" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			atomic.AddInt32(&lists, 1)
			w.Write(okResponse([]Inbound{{ID: 1}}))
		}))
		defer srv.Close()

		res := &InboundResource{client: newTestClient(t, srv.URL)}
		if err := res.waitForInboundDeletion(context.Background(), 1); err == nil {
			t.Fatalf("expected error after exhausting attempts")
		}
		// destroyVisibilityAttempts changed; resource-side waitForInboundDeletion
		// is intentionally separate (20×500ms = 10s) — see provider/resource_inbound.go.
		if got := atomic.LoadInt32(&lists); got != 20 {
			t.Fatalf("expected 20 list calls, got %d", got)
		}
	})

	t.Run("transient-list-error-not-treated-as-success", func(t *testing.T) {
		// A transient 5xx must not be misread as confirmed deletion.
		var lists int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/panel/api/inbounds/list" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			atomic.AddInt32(&lists, 1)
			http.Error(w, "boom", http.StatusInternalServerError)
		}))
		defer srv.Close()

		res := &InboundResource{client: newTestClient(t, srv.URL)}
		err := res.waitForInboundDeletion(context.Background(), 5)
		if err == nil {
			t.Fatalf("expected error when list keeps failing")
		}
	})

	t.Run("ctx-canceled-during-backoff", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/panel/api/inbounds/list" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Write(okResponse([]Inbound{{ID: 11}}))
		}))
		defer srv.Close()

		res := &InboundResource{client: newTestClient(t, srv.URL)}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		start := time.Now()
		err := res.waitForInboundDeletion(ctx, 11)
		elapsed := time.Since(start)
		if err == nil {
			t.Fatalf("expected ctx error")
		}
		if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context error, got %v", err)
		}
		// Full attempt budget is 6 × 500ms ≈ 3s. Cancellation must
		// short-circuit the backoff well before that.
		if elapsed > 600*time.Millisecond {
			t.Fatalf("backoff did not respect ctx cancel; elapsed=%v", elapsed)
		}
	})
}

func TestInboundToForm(t *testing.T) {
	nodeID := 42
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
		NodeID:         &nodeID,
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
		"nodeId":               []string{"42"},
	}
	if form.Encode() != want.Encode() {
		t.Fatalf("unexpected form: %s", form.Encode())
	}
}
