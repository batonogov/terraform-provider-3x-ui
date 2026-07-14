package provider

import (
	"encoding/json"
	"net/url"
	"testing"
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

func TestSetBasePath(t *testing.T) {
	client := newTestClient(t, "http://example.com")
	if client.basePath != "/" {
		t.Fatalf("initial basePath = %q, want /", client.basePath)
	}
	client.SetBasePath("/newpath/")
	if client.basePath != "/newpath/" {
		t.Fatalf("after SetBasePath(/newpath/) = %q, want /newpath/", client.basePath)
	}
	client.SetBasePath("raw")
	if client.basePath != "/raw/" {
		t.Fatalf("after SetBasePath(raw) = %q, want /raw/", client.basePath)
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

func TestIsSubset(t *testing.T) {
	var old, newVal any
	_ = json.Unmarshal([]byte(`{"a":1,"b":{"c":2},"arr":[{"x":1,"y":2},{"x":2}]}`), &old)
	_ = json.Unmarshal([]byte(`{"b":{"c":2},"arr":[{"x":2}]}`), &newVal)
	if !isSubset(newVal, old) {
		t.Fatalf("expected subset")
	}

	var notSubset any
	_ = json.Unmarshal([]byte(`{"z":99}`), &notSubset)
	if isSubset(notSubset, old) {
		t.Fatalf("expected not subset")
	}
}

func TestBuildAndFlattenSettings(t *testing.T) {
	item := map[string]any{
		"decryption": "none",
	}
	settingsJSON := buildSettingsJSON(item, "")
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

	flattened, err := flattenSettings(`{"decryption":"none","clients":[{"id":"id1","email":"a@example.com","limitIp":2,"expiryTime":10,"enable":true}]}`, "")
	if err != nil {
		t.Fatalf("flattenSettings error: %v", err)
	}
	if len(flattened) != 1 {
		t.Fatalf("expected 1 settings item, got %d", len(flattened))
	}
	out := flattened[0].(map[string]any)
	if out["decryption"] != "none" {
		t.Fatalf("unexpected flattened decryption: %#v", out["decryption"])
	}
}

func TestBuildAndFlattenStreamSettings(t *testing.T) {
	item := map[string]any{
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
	}
	streamJSON := buildStreamSettingsJSON(item)
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

	flattened, err := flattenStreamSettings(`{"network":"tcp","security":"reality","externalProxy":[{"dest":"example.com","port":443,"remark":"edge"}],"tcpSettings":{"acceptProxyProtocol":true,"header":{"type":"none"}}}`)
	if err != nil {
		t.Fatalf("flattenStreamSettings error: %v", err)
	}
	if len(flattened) != 1 {
		t.Fatalf("expected 1 stream settings item, got %d", len(flattened))
	}
	out := flattened[0].(map[string]any)
	if out["network"] != "tcp" || out["security"] != "reality" {
		t.Fatalf("unexpected flattened stream settings: %#v", out)
	}
}

func TestEmptyTCPSettingsRoundtrip(t *testing.T) {
	// Simulate user specifying tcp_settings {} (empty block).
	item := map[string]any{
		"network":      "tcp",
		"security":     "reality",
		"tcp_settings": []any{map[string]any{}},
	}
	streamJSON := buildStreamSettingsJSON(item)
	var payload map[string]any
	if err := json.Unmarshal([]byte(streamJSON), &payload); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if payload["network"] != "tcp" {
		t.Fatalf("expected network=tcp, got %v", payload["network"])
	}
	// tcpSettings must be present in the API payload (even if empty).
	if _, ok := payload["tcpSettings"]; !ok {
		t.Fatalf("expected tcpSettings in payload, got: %v", payload)
	}

	// Flatten the API response back — tcp_settings must survive.
	flattened, err := flattenStreamSettings(streamJSON)
	if err != nil {
		t.Fatalf("flattenStreamSettings error: %v", err)
	}
	out := flattened[0].(map[string]any)
	ts, ok := out["tcp_settings"].([]any)
	if !ok || len(ts) == 0 {
		t.Fatalf("expected tcp_settings block in flattened output, got: %#v", out)
	}
}

func TestEmptyNetworkSettingsRoundtrip(t *testing.T) {
	cases := []struct {
		name    string
		network string
		tfKey   string
		apiKey  string
	}{
		{"tcp", "tcp", "tcp_settings", "tcpSettings"},
		{"ws", "ws", "ws_settings", "wsSettings"},
		{"grpc", "grpc", "grpc_settings", "grpcSettings"},
		{"httpupgrade", "httpupgrade", "httpupgrade_settings", "httpupgradeSettings"},
		{"xhttp", "xhttp", "xhttp_settings", "xhttpSettings"},
		{"kcp", "kcp", "kcp_settings", "kcpSettings"},
		{"hysteria_stream", "hysteria", "hysteria_settings", "hysteriaSettings"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			item := map[string]any{
				"network":  tc.network,
				"security": "none",
				tc.tfKey:   []any{map[string]any{}},
			}
			streamJSON := buildStreamSettingsJSON(item)

			var payload map[string]any
			if err := json.Unmarshal([]byte(streamJSON), &payload); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}
			if _, ok := payload[tc.apiKey]; !ok {
				t.Fatalf("expected %s in API payload, got: %v", tc.apiKey, payload)
			}

			flattened, err := flattenStreamSettings(streamJSON)
			if err != nil {
				t.Fatalf("flattenStreamSettings error: %v", err)
			}
			out := flattened[0].(map[string]any)
			v, ok := out[tc.tfKey].([]any)
			if !ok || len(v) == 0 {
				t.Fatalf("expected %s block in flattened output, got: %#v", tc.tfKey, out)
			}
		})
	}
}

func TestBuildAndFlattenSniffing(t *testing.T) {
	item := map[string]any{
		"enabled":       true,
		"dest_override": []any{"http", "tls"},
		"metadata_only": false,
		"route_only":    true,
	}
	sniffingJSON := buildSniffingJSON(item)
	if sniffingJSON == "{}" {
		t.Fatalf("expected sniffing JSON, got {}")
	}
	flattened, err := flattenSniffing(`{"enabled":true,"destOverride":["http","tls"],"metadataOnly":false,"routeOnly":true}`)
	if err != nil {
		t.Fatalf("flattenSniffing error: %v", err)
	}
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
	out, err := flattenSniffing("")
	if err != nil {
		t.Fatalf("flattenSniffing error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty sniffing, got %#v", out)
	}
}

func TestStringValueOrNull(t *testing.T) {
	// Empty string → null
	result := stringValueOrNull("")
	if !result.IsNull() {
		t.Fatalf("expected null for empty string, got %q", result.ValueString())
	}

	// Non-empty string → value
	result = stringValueOrNull("0.0.0.0")
	if result.IsNull() {
		t.Fatal("expected non-null for '0.0.0.0'")
	}
	if result.ValueString() != "0.0.0.0" {
		t.Fatalf("expected '0.0.0.0', got %q", result.ValueString())
	}

	// Another non-empty string
	result = stringValueOrNull("127.0.0.1")
	if result.IsNull() || result.ValueString() != "127.0.0.1" {
		t.Fatalf("expected '127.0.0.1', got null=%v val=%q", result.IsNull(), result.ValueString())
	}
}

func TestNormalizeXrayVersion(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"26.2.6", "v26.2.6"},
		{"v26.2.6", "v26.2.6"},
		{"1.0.0", "v1.0.0"},
		{"v1.0.0", "v1.0.0"},
		{"", ""},
		{"Unknown", "vUnknown"},
	}
	for _, tc := range tests {
		got := normalizeXrayVersion(tc.in)
		if got != tc.want {
			t.Errorf("normalizeXrayVersion(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
