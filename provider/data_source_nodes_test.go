package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetNodes verifies the client decodes the 3x-ui multi-node tree
// (GET /panel/api/nodes/list), including the managed fields and a transitive
// sub-node (read-only projection surfaced from a downstream panel).
func TestGetNodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: "sess"})
			w.Write(okResponse(nil))
			return
		case "/panel/api/nodes/list":
			// A direct node plus a transitive sub-node (Id == 0, read-only).
			w.Write(okResponse([]any{
				map[string]any{
					"id":              1,
					"name":            "de-fra-1",
					"remark":          "Frankfurt",
					"scheme":          "https",
					"address":         "node1.example.com",
					"port":            2053,
					"basePath":        "/",
					"apiToken":        "abcdef0123456789",
					"enable":          true,
					"tlsVerifyMode":   "verify",
					"inboundSyncMode": "all",
					"inboundTags":     []any{},
					"outboundTag":     "",
					"guid":            "11111111-1111-1111-1111-111111111111",
					"status":          "online",
					"latencyMs":       42,
					"xrayVersion":     "25.10.31",
					"panelVersion":    "v3.4.1",
					"transitive":      false,
					"inboundCount":    5,
					"clientCount":     27,
				},
				map[string]any{
					"id":         0,
					"name":       "de-fra-1-sub",
					"address":    "10.0.0.2",
					"port":       2053,
					"guid":       "22222222-2222-2222-2222-222222222222",
					"parentGuid": "11111111-1111-1111-1111-111111111111",
					"transitive": true,
					"status":     "online",
				},
			}))
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	nodes, err := client.GetNodes(context.Background())
	if err != nil {
		t.Fatalf("GetNodes error: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}

	direct := nodes[0]
	if direct.Id != 1 || direct.Name != "de-fra-1" || direct.Address != "node1.example.com" {
		t.Fatalf("unexpected direct node: %+v", direct)
	}
	if direct.Port != 2053 || direct.Scheme != "https" || direct.BasePath != "/" {
		t.Fatalf("unexpected direct node connection fields: %+v", direct)
	}
	if direct.ApiToken != "abcdef0123456789" {
		t.Fatalf("expected raw apiToken to round-trip, got %q", direct.ApiToken)
	}
	if direct.TlsVerifyMode != "verify" || direct.InboundSyncMode != "all" {
		t.Fatalf("unexpected managed fields: %+v", direct)
	}
	if direct.Status != "online" || direct.LatencyMs != 42 || direct.XrayVersion != "25.10.31" {
		t.Fatalf("unexpected observed fields: %+v", direct)
	}
	if direct.Transitive {
		t.Fatalf("direct node must not be transitive")
	}

	transitive := nodes[1]
	if transitive.Id != 0 || !transitive.Transitive {
		t.Fatalf("expected transitive sub-node with Id==0, got %+v", transitive)
	}
	if transitive.ParentGuid != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("unexpected parentGuid: %q", transitive.ParentGuid)
	}
}

// TestGetNodesEmpty verifies the cluster-with-no-nodes case (fresh panel).
func TestGetNodesEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: "sess"})
			w.Write(okResponse(nil))
			return
		case "/panel/api/nodes/list":
			w.Write(okResponse([]any{}))
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	nodes, err := client.GetNodes(context.Background())
	if err != nil {
		t.Fatalf("GetNodes error: %v", err)
	}
	if len(nodes) != 0 {
		t.Fatalf("expected 0 nodes on fresh panel, got %d", len(nodes))
	}
}
