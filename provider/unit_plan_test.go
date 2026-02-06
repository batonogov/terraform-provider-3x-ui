package provider

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestNormalizeBasePath(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", "/"},
		{"   ", "/"},
		{"xui", "/xui/"},
		{"/xui", "/xui/"},
		{"/xui/", "/xui/"},
	}
	for _, tt := range tests {
		if got := normalizeBasePath(tt.in); got != tt.want {
			t.Fatalf("normalizeBasePath(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestResolvePathWithBaseURLPath(t *testing.T) {
	client := newTestClient(t, "http://example.com/base")
	client.basePath = "/xui/"

	endpoint, err := client.resolvePath("panel/api/inbounds/list")
	if err != nil {
		t.Fatalf("resolvePath error: %v", err)
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if parsed.Path != "/base/xui/panel/api/inbounds/list" {
		t.Fatalf("unexpected path: %s", parsed.Path)
	}
}

func TestResolvePathEmptyRel(t *testing.T) {
	client := newTestClient(t, "http://example.com")
	if _, err := client.resolvePath(""); err == nil {
		t.Fatalf("expected error for empty rel path")
	}
}

func TestParseID(t *testing.T) {
	if got, err := parseID("123"); err != nil || got != 123 {
		t.Fatalf("parseID valid failed: got %d err %v", got, err)
	}
	if _, err := parseID("0"); err == nil {
		t.Fatalf("expected error for id 0")
	}
	if _, err := parseID("abc"); err == nil {
		t.Fatalf("expected error for non-numeric id")
	}
}

func TestMakeInboundClientID(t *testing.T) {
	got := makeInboundClientID(10, "client-id")
	if got != "10:client-id" {
		t.Fatalf("makeInboundClientID = %q", got)
	}
}

func TestNewUUIDFormat(t *testing.T) {
	id, err := newUUID()
	if err != nil {
		t.Fatalf("newUUID error: %v", err)
	}
	if len(id) != 36 {
		t.Fatalf("unexpected UUID length: %d (%s)", len(id), id)
	}
	if id[8] != '-' || id[13] != '-' || id[18] != '-' || id[23] != '-' {
		t.Fatalf("unexpected UUID format: %s", id)
	}
}

func TestJSONSubsetDiffSuppress(t *testing.T) {
	old := `{"a":1,"b":{"c":2},"arr":[{"x":1,"y":2},{"x":2}]}`
	newSubset := `{"b":{"c":2},"arr":[{"x":2}]}`
	if !jsonSubsetDiffSuppress("k", old, newSubset, nil) {
		t.Fatalf("expected subset diff suppress")
	}
	if jsonSubsetDiffSuppress("k", old, "{", nil) {
		t.Fatalf("expected false on invalid JSON")
	}
	if !jsonSubsetDiffSuppress("k", old, "   ", nil) {
		t.Fatalf("expected true when new is empty")
	}
}

func TestBuildAndFlattenSettings(t *testing.T) {
	r := resourceInbound()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"settings": []any{
			map[string]any{
				"decryption": "none",
				"clients": []any{
					map[string]any{
						"email":    "a@example.com",
						"enable":   true,
						"flow":     "xtls",
						"limit_ip": 1,
					},
				},
			},
		},
	})
	settingsJSON := buildSettingsJSON(d)
	if settingsJSON == "{}" {
		t.Fatalf("expected settings JSON, got {}")
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(settingsJSON), &payload); err != nil {
		t.Fatalf("unmarshal settings error: %v", err)
	}
	if payload["decryption"] != "none" {
		t.Fatalf("unexpected decryption: %#v", payload["decryption"])
	}
	clients, ok := payload["clients"].([]any)
	if !ok || len(clients) != 1 {
		t.Fatalf("expected 1 client, got %#v", payload["clients"])
	}

	flattened := flattenSettings(`{"decryption":"none","clients":[{"id":"id1","email":"a@example.com","limitIp":2,"expiryTime":10,"enable":true}]}`)
	if len(flattened) != 1 {
		t.Fatalf("expected 1 settings item, got %d", len(flattened))
	}
	out := flattened[0].(map[string]any)
	if out["decryption"] != "none" {
		t.Fatalf("unexpected flattened decryption: %#v", out["decryption"])
	}
	outClients := out["clients"].([]any)
	outClient := outClients[0].(map[string]any)
	if outClient["id"] != "id1" || outClient["limit_ip"].(int) != 2 {
		t.Fatalf("unexpected flattened client: %#v", outClient)
	}
}

func TestBuildAndFlattenStreamSettings(t *testing.T) {
	r := resourceInbound()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"stream_settings": []any{
			map[string]any{
				"network":  "tcp",
				"security": "reality",
				"external_proxy": []any{
					map[string]any{
						"dest":   "example.com",
						"port":   443,
						"remark": "edge",
					},
				},
				"tcp_settings": []any{
					map[string]any{
						"accept_proxy_protocol": true,
						"header": []any{
							map[string]any{"type": "none"},
						},
					},
				},
			},
		},
	})
	streamJSON := buildStreamSettingsJSON(d)
	if streamJSON == "{}" {
		t.Fatalf("expected stream settings JSON, got {}")
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(streamJSON), &payload); err != nil {
		t.Fatalf("unmarshal stream settings error: %v", err)
	}
	if payload["network"] != "tcp" || payload["security"] != "reality" {
		t.Fatalf("unexpected stream settings payload: %#v", payload)
	}

	flattened := flattenStreamSettings(`{"network":"tcp","security":"reality","externalProxy":[{"dest":"example.com","port":443,"remark":"edge"}],"tcpSettings":{"acceptProxyProtocol":true,"header":{"type":"none"}}}`)
	if len(flattened) != 1 {
		t.Fatalf("expected 1 stream settings item, got %d", len(flattened))
	}
	out := flattened[0].(map[string]any)
	if out["network"] != "tcp" || out["security"] != "reality" {
		t.Fatalf("unexpected flattened stream settings: %#v", out)
	}
}

func TestBuildAndFlattenSniffing(t *testing.T) {
	r := resourceInbound()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"sniffing": []any{
			map[string]any{
				"enabled":       true,
				"dest_override": []any{"http", "tls"},
				"metadata_only": false,
				"route_only":    true,
			},
		},
	})
	sniffingJSON := buildSniffingJSON(d)
	if sniffingJSON == "{}" {
		t.Fatalf("expected sniffing JSON, got {}")
	}
	flattened := flattenSniffing(`{"enabled":true,"destOverride":["http","tls"],"metadataOnly":false,"routeOnly":true}`)
	if len(flattened) != 1 {
		t.Fatalf("expected 1 sniffing item, got %d", len(flattened))
	}
	out := flattened[0].(map[string]any)
	if out["enabled"] != true || out["route_only"] != true {
		t.Fatalf("unexpected flattened sniffing: %#v", out)
	}
}

func TestEnsureInboundClientIDs_NoChange(t *testing.T) {
	inbound := &Inbound{
		Settings: `{"clients":[{"id":"fixed-id","email":"a@example.com"}]}`,
	}
	original := inbound.Settings
	if err := ensureInboundClientIDs(inbound); err != nil {
		t.Fatalf("ensureInboundClientIDs error: %v", err)
	}
	if inbound.Settings != original {
		t.Fatalf("expected settings unchanged")
	}
}

func TestApplyDefaultInboundSettings_Empty(t *testing.T) {
	inbound := &Inbound{
		Protocol: "vless",
		Settings: "",
	}
	if err := applyDefaultInboundSettings(inbound); err != nil {
		t.Fatalf("applyDefaultInboundSettings error: %v", err)
	}
	if inbound.Settings == "" || inbound.Settings == "{}" {
		t.Fatalf("expected default settings, got %q", inbound.Settings)
	}
	var settings map[string]any
	if err := json.Unmarshal([]byte(inbound.Settings), &settings); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if settings["decryption"] != "none" {
		t.Fatalf("expected decryption=none, got %#v", settings["decryption"])
	}
}

func TestApplyDefaultInboundSettings_InvalidJSON(t *testing.T) {
	inbound := &Inbound{
		Protocol: "vless",
		Settings: "{",
	}
	if err := applyDefaultInboundSettings(inbound); err == nil {
		t.Fatalf("expected error for invalid JSON settings")
	}
}

func TestFlattenSniffingEmpty(t *testing.T) {
	if out := flattenSniffing(""); len(out) != 0 {
		t.Fatalf("expected empty sniffing, got %#v", out)
	}
}
