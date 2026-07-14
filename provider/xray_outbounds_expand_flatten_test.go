package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// Table-driven round-trip tests for xray outbound per-protocol expand/flatten.
//
// There are TWO expand/flatten pairs per protocol:
//  1. expandXxxSettingsFromModel ([]TypedModel -> []any) / flattenXxxSettingsToModel ([]any -> []TypedModel)
//  2. expandXxxOutSettings (map[settings_key]->Xray JSON) / flattenXxxOutSettings (Xray JSON -> map)
//
// Both are tested here.
// ---------------------------------------------------------------------------

// --- VLESS ---

func TestExpandFlatten_VlessOutModel_RoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		input []XrayVlessOutSettings
	}{
		{name: "empty", input: nil},
		{
			name: "full",
			input: []XrayVlessOutSettings{
				{
					Address:    types.StringValue("1.2.3.4"),
					Port:       types.Int64Value(443),
					ID:         types.StringValue("uuid-1234"),
					Flow:       types.StringValue("xtls-rprx-vision"),
					Encryption: types.StringValue("none"),
					ReverseTag: types.StringValue("rev"),
				},
			},
		},
		{
			name: "nulls",
			input: []XrayVlessOutSettings{{}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expanded := expandVlessSettingsFromModel(tc.input)
			flat := flattenVlessSettingsToModel(expanded)
			if len(flat) != len(tc.input) {
				t.Fatalf("count mismatch: got %d want %d", len(flat), len(tc.input))
			}
			for i, vs := range tc.input {
				if flat[i].Address.ValueString() != vs.Address.ValueString() {
					t.Fatalf("address[%d] mismatch: got %q want %q", i, flat[i].Address.ValueString(), vs.Address.ValueString())
				}
				if flat[i].ID.ValueString() != vs.ID.ValueString() {
					t.Fatalf("id[%d] mismatch: got %q want %q", i, flat[i].ID.ValueString(), vs.ID.ValueString())
				}
			}
		})
	}
}

func TestExpandFlatten_VlessOutJSON_RoundTrip(t *testing.T) {
	cases := []struct {
		name string
		data map[string]any
	}{
		{name: "empty", data: map[string]any{}},
		{
			name: "full",
			data: map[string]any{
				"vless_settings": []any{
					map[string]any{
						"address":     "1.2.3.4",
						"port":        443,
						"id":          "uuid-1234",
						"flow":        "xtls-rprx-vision",
						"encryption":  "none",
						"reverse_tag": "rev",
					},
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			xrayJSON := expandVlessOutSettings(tc.data)
			flat := flattenVlessOutSettings(xrayJSON)
			// For empty input expandVlessOutSettings returns nil
			if tc.name == "empty" {
				if xrayJSON != nil {
					t.Fatalf("expected nil for empty, got %v", xrayJSON)
				}
				return
			}
			if flat["address"] != "1.2.3.4" {
				t.Fatalf("address mismatch: got %v want %v", flat["address"], "1.2.3.4")
			}
			if flat["id"] != "uuid-1234" {
				t.Fatalf("id mismatch: got %v want %v", flat["id"], "uuid-1234")
			}
			if flat["reverse_tag"] != "rev" {
				t.Fatalf("reverse_tag mismatch: got %v want %v", flat["reverse_tag"], "rev")
			}
		})
	}
}

// --- VMess ---

func TestExpandFlatten_VmessOutModel_RoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		input []XrayVmessOutSettings
	}{
		{name: "empty", input: nil},
		{
			name: "full",
			input: []XrayVmessOutSettings{
				{
					Address:  types.StringValue("5.6.7.8"),
					Port:     types.Int64Value(8443),
					ID:       types.StringValue("vmess-uuid"),
					Security: types.StringValue("aes-128-gcm"),
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expanded := expandVmessSettingsFromModel(tc.input)
			flat := flattenVmessSettingsToModel(expanded)
			if len(flat) != len(tc.input) {
				t.Fatalf("count mismatch: got %d want %d", len(flat), len(tc.input))
			}
			for i, vs := range tc.input {
				if flat[i].Address.ValueString() != vs.Address.ValueString() {
					t.Fatalf("address[%d] mismatch: got %q want %q", i, flat[i].Address.ValueString(), vs.Address.ValueString())
				}
				if flat[i].Security.ValueString() != vs.Security.ValueString() {
					t.Fatalf("security[%d] mismatch: got %q want %q", i, flat[i].Security.ValueString(), vs.Security.ValueString())
				}
			}
		})
	}
}

func TestExpandFlatten_VmessOutJSON_RoundTrip(t *testing.T) {
	data := map[string]any{
		"vmess_settings": []any{
			map[string]any{
				"address":  "5.6.7.8",
				"port":     8443,
				"id":       "uuid-vmess",
				"security": "auto",
			},
		},
	}
	xrayJSON := expandVmessOutSettings(data)
	if xrayJSON == nil {
		t.Fatal("expected non-nil xrayJSON")
	}
	flat := flattenVmessOutSettings(xrayJSON)
	if flat["address"] != "5.6.7.8" {
		t.Fatalf("address mismatch: got %v want %v", flat["address"], "5.6.7.8")
	}
	if flat["security"] != "auto" {
		t.Fatalf("security mismatch: got %v want %v", flat["security"], "auto")
	}
}

// --- Trojan ---

func TestExpandFlatten_TrojanOutModel_RoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		input []XrayTrojanOutSettings
	}{
		{name: "empty", input: nil},
		{
			name: "full",
			input: []XrayTrojanOutSettings{
				{
					Address:  types.StringValue("trojan.example.com"),
					Port:     types.Int64Value(443),
					Password: types.StringValue("pass123"),
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expanded := expandTrojanSettingsFromModel(tc.input)
			flat := flattenTrojanSettingsToModel(expanded)
			if len(flat) != len(tc.input) {
				t.Fatalf("count mismatch: got %d want %d", len(flat), len(tc.input))
			}
			for i, ts := range tc.input {
				if flat[i].Address.ValueString() != ts.Address.ValueString() {
					t.Fatalf("address[%d] mismatch: got %q want %q", i, flat[i].Address.ValueString(), ts.Address.ValueString())
				}
				if flat[i].Password.ValueString() != ts.Password.ValueString() {
					t.Fatalf("password[%d] mismatch: got %q want %q", i, flat[i].Password.ValueString(), ts.Password.ValueString())
				}
			}
		})
	}
}

func TestExpandFlatten_TrojanOutJSON_RoundTrip(t *testing.T) {
	data := map[string]any{
		"trojan_settings": []any{
			map[string]any{
				"address":  "trojan.example.com",
				"port":     443,
				"password": "secret",
			},
		},
	}
	xrayJSON := expandTrojanOutSettings(data)
	if xrayJSON == nil {
		t.Fatal("expected non-nil xrayJSON")
	}
	flat := flattenTrojanOutSettings(xrayJSON)
	if flat["address"] != "trojan.example.com" {
		t.Fatalf("address mismatch: got %v want %v", flat["address"], "trojan.example.com")
	}
	if flat["password"] != "secret" {
		t.Fatalf("password mismatch: got %v want %v", flat["password"], "secret")
	}
}

// --- Shadowsocks ---

func TestExpandFlatten_ShadowsocksOutModel_RoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		input []XrayShadowsocksOutSettings
	}{
		{name: "empty", input: nil},
		{
			name: "full",
			input: []XrayShadowsocksOutSettings{
				{
					Address:    types.StringValue("ss.example.com"),
					Port:       types.Int64Value(8388),
					Password:   types.StringValue("ss-pass"),
					Method:     types.StringValue("aes-256-gcm"),
					UOT:        types.BoolValue(true),
					UOTVersion: types.Int64Value(2),
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expanded := expandShadowsocksSettingsFromModel(tc.input)
			flat := flattenShadowsocksSettingsToModel(expanded)
			if len(flat) != len(tc.input) {
				t.Fatalf("count mismatch: got %d want %d", len(flat), len(tc.input))
			}
			for i, ss := range tc.input {
				if flat[i].Method.ValueString() != ss.Method.ValueString() {
					t.Fatalf("method[%d] mismatch: got %q want %q", i, flat[i].Method.ValueString(), ss.Method.ValueString())
				}
				if flat[i].UOT.ValueBool() != ss.UOT.ValueBool() {
					t.Fatalf("uot[%d] mismatch: got %v want %v", i, flat[i].UOT.ValueBool(), ss.UOT.ValueBool())
				}
			}
		})
	}
}

func TestExpandFlatten_ShadowsocksOutJSON_RoundTrip(t *testing.T) {
	data := map[string]any{
		"shadowsocks_settings": []any{
			map[string]any{
				"address":     "ss.example.com",
				"port":        8388,
				"password":    "ss-pass",
				"method":      "chacha20-ietf-poly1305",
				"uot":         true,
				"uot_version": 2,
			},
		},
	}
	xrayJSON := expandShadowsocksOutSettings(data)
	if xrayJSON == nil {
		t.Fatal("expected non-nil xrayJSON")
	}
	flat := flattenShadowsocksOutSettings(xrayJSON)
	if flat["address"] != "ss.example.com" {
		t.Fatalf("address mismatch: got %v want %v", flat["address"], "ss.example.com")
	}
	if flat["method"] != "chacha20-ietf-poly1305" {
		t.Fatalf("method mismatch: got %v want %v", flat["method"], "chacha20-ietf-poly1305")
	}
	if flat["uot"] != true {
		t.Fatalf("uot mismatch: got %v want %v", flat["uot"], true)
	}
}

// --- SOCKS ---

func TestExpandFlatten_SocksOutModel_RoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		input []XraySocksOutSettings
	}{
		{name: "empty", input: nil},
		{
			name: "full",
			input: []XraySocksOutSettings{
				{
					Address: types.StringValue("socks.example.com"),
					Port:    types.Int64Value(1080),
					User:    types.StringValue("socksuser"),
					Pass:    types.StringValue("sockspass"),
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expanded := expandSocksSettingsFromModel(tc.input)
			flat := flattenSocksSettingsToModel(expanded)
			if len(flat) != len(tc.input) {
				t.Fatalf("count mismatch: got %d want %d", len(flat), len(tc.input))
			}
			for i, ss := range tc.input {
				if flat[i].User.ValueString() != ss.User.ValueString() {
					t.Fatalf("user[%d] mismatch: got %q want %q", i, flat[i].User.ValueString(), ss.User.ValueString())
				}
			}
		})
	}
}

func TestExpandFlatten_SocksOutJSON_RoundTrip(t *testing.T) {
	data := map[string]any{
		"socks_settings": []any{
			map[string]any{
				"address": "socks.example.com",
				"port":    1080,
				"user":    "su",
				"pass":    "sp",
			},
		},
	}
	xrayJSON := expandSocksOutSettings(data)
	if xrayJSON == nil {
		t.Fatal("expected non-nil xrayJSON")
	}
	flat := flattenSocksOutSettings(xrayJSON)
	if flat["address"] != "socks.example.com" {
		t.Fatalf("address mismatch: got %v want %v", flat["address"], "socks.example.com")
	}
	if flat["user"] != "su" {
		t.Fatalf("user mismatch: got %v want %v", flat["user"], "su")
	}
}

// --- HTTP ---

func TestExpandFlatten_HTTPOutModel_RoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		input []XrayHTTPOutSettings
	}{
		{name: "empty", input: nil},
		{
			name: "full",
			input: []XrayHTTPOutSettings{
				{
					Address: types.StringValue("http.example.com"),
					Port:    types.Int64Value(8080),
					User:    types.StringValue("httpuser"),
					Pass:    types.StringValue("httppass"),
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expanded := expandHTTPSettingsFromModel(tc.input)
			flat := flattenHTTPSettingsToModel(expanded)
			if len(flat) != len(tc.input) {
				t.Fatalf("count mismatch: got %d want %d", len(flat), len(tc.input))
			}
			for i, hs := range tc.input {
				if flat[i].User.ValueString() != hs.User.ValueString() {
					t.Fatalf("user[%d] mismatch: got %q want %q", i, flat[i].User.ValueString(), hs.User.ValueString())
				}
			}
		})
	}
}

func TestExpandFlatten_HTTPOutJSON_RoundTrip(t *testing.T) {
	data := map[string]any{
		"http_settings": []any{
			map[string]any{
				"address": "http.example.com",
				"port":    8080,
				"user":    "hu",
				"pass":    "hp",
			},
		},
	}
	xrayJSON := expandHTTPOutSettings(data)
	if xrayJSON == nil {
		t.Fatal("expected non-nil xrayJSON")
	}
	flat := flattenHTTPOutSettings(xrayJSON)
	if flat["address"] != "http.example.com" {
		t.Fatalf("address mismatch: got %v want %v", flat["address"], "http.example.com")
	}
	if flat["user"] != "hu" {
		t.Fatalf("user mismatch: got %v want %v", flat["user"], "hu")
	}
}

// --- Hysteria ---

func TestExpandFlatten_HysteriaOutModel_RoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		input []XrayHysteriaOutSettings
	}{
		{name: "empty", input: nil},
		{
			name: "full",
			input: []XrayHysteriaOutSettings{
				{
					Address: types.StringValue("hy.example.com"),
					Port:    types.Int64Value(36712),
					Version: types.Int64Value(2),
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expanded := expandHysteriaSettingsFromModel(tc.input)
			flat := flattenHysteriaSettingsToModel(expanded)
			if len(flat) != len(tc.input) {
				t.Fatalf("count mismatch: got %d want %d", len(flat), len(tc.input))
			}
			for i, hs := range tc.input {
				if flat[i].Version.ValueInt64() != hs.Version.ValueInt64() {
					t.Fatalf("version[%d] mismatch: got %d want %d", i, flat[i].Version.ValueInt64(), hs.Version.ValueInt64())
				}
			}
		})
	}
}

func TestExpandFlatten_HysteriaOutJSON_RoundTrip(t *testing.T) {
	data := map[string]any{
		"hysteria_settings": []any{
			map[string]any{
				"address": "hy.example.com",
				"port":    36712,
				"version": 2,
			},
		},
	}
	xrayJSON := expandHysteriaOutSettings(data)
	if xrayJSON == nil {
		t.Fatal("expected non-nil xrayJSON")
	}
	flat := flattenHysteriaOutSettings(xrayJSON)
	if flat["address"] != "hy.example.com" {
		t.Fatalf("address mismatch: got %v want %v", flat["address"], "hy.example.com")
	}
	if flat["version"] != 2 {
		t.Fatalf("version mismatch: got %v want %v", flat["version"], 2)
	}
}

// --- Blackhole ---

func TestExpandFlatten_BlackholeOutModel_RoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		input []XrayBlackholeSettings
	}{
		{name: "empty", input: nil},
		{
			name: "with_type",
			input: []XrayBlackholeSettings{
				{ResponseType: types.StringValue("http")},
			},
		},
		{
			name:  "null_type",
			input: []XrayBlackholeSettings{{ResponseType: types.StringNull()}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expanded := expandBlackholeSettingsFromModel(tc.input)
			flat := flattenBlackholeSettingsToModel(expanded)
			if len(flat) != len(tc.input) {
				t.Fatalf("count mismatch: got %d want %d", len(flat), len(tc.input))
			}
		})
	}
}

func TestExpandFlatten_BlackholeOutJSON_RoundTrip(t *testing.T) {
	// Expand with a response_type
	data := map[string]any{
		"blackhole_settings": []any{
			map[string]any{"response_type": "none"},
		},
	}
	xrayJSON := expandBlackholeSettings(data)
	if xrayJSON == nil {
		t.Fatal("expected non-nil xrayJSON")
	}
	flat := flattenBlackholeSettings(xrayJSON)
	if flat["response_type"] != "none" {
		t.Fatalf("response_type mismatch: got %v want %v", flat["response_type"], "none")
	}

	// Flatten empty input should default to "none"
	flat2 := flattenBlackholeSettings(map[string]any{})
	if flat2["response_type"] != "none" {
		t.Fatalf("default response_type mismatch: got %v want %v", flat2["response_type"], "none")
	}
}

// --- DNS ---

func TestExpandFlatten_DNSOutModel_RoundTrip(t *testing.T) {
	blockTypes, _ := types.ListValue(types.Int64Type, []attr.Value{
		types.Int64Value(1),
		types.Int64Value(2),
	})
	cases := []struct {
		name  string
		input []XrayOutboundDNSSettings
	}{
		{name: "empty", input: nil},
		{
			name: "full",
			input: []XrayOutboundDNSSettings{
				{
					Network:    types.StringValue("tcp"),
					Address:    types.StringValue("1.1.1.1"),
					Port:       types.Int64Value(53),
					NonIPQuery: types.StringValue("drop"),
					BlockTypes: blockTypes,
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expanded := expandDNSSettingsFromModel(tc.input)
			flat := flattenDNSSettingsToModel(expanded)
			if len(flat) != len(tc.input) {
				t.Fatalf("count mismatch: got %d want %d", len(flat), len(tc.input))
			}
			for i, ds := range tc.input {
				if flat[i].Network.ValueString() != ds.Network.ValueString() {
					t.Fatalf("network[%d] mismatch: got %q want %q", i, flat[i].Network.ValueString(), ds.Network.ValueString())
				}
			}
		})
	}
}

func TestExpandFlatten_DNSOutJSON_RoundTrip(t *testing.T) {
	data := map[string]any{
		"dns_settings": []any{
			map[string]any{
				"network":      "tcp",
				"address":      "8.8.8.8",
				"port":         53,
				"non_ip_query": "skip",
				"block_types":  []any{float64(1), float64(3)},
			},
		},
	}
	xrayJSON := expandOutboundDNSSettings(data)
	if xrayJSON == nil {
		t.Fatal("expected non-nil xrayJSON")
	}
	flat := flattenOutboundDNSSettings(xrayJSON)
	if flat["network"] != "tcp" {
		t.Fatalf("network mismatch: got %v want %v", flat["network"], "tcp")
	}
	if flat["non_ip_query"] != "skip" {
		t.Fatalf("non_ip_query mismatch: got %v want %v", flat["non_ip_query"], "skip")
	}
}

// --- Freedom ---

func TestExpandFlatten_FreedomOutModel_RoundTrip(t *testing.T) {
	ipsBlocked, _ := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("geoip:cn"),
		types.StringValue("geoip:private"),
	})
	cases := []struct {
		name  string
		input []XrayFreedomSettings
	}{
		{name: "empty", input: nil},
		{
			name: "full",
			input: []XrayFreedomSettings{
				{
					DomainStrategy: types.StringValue("UseIPv4"),
					Redirect:       types.StringValue(":0"),
					Fragment: []XrayFreedomFragment{
						{Packets: types.StringValue("tlshello"), Length: types.StringValue("100-200"), Interval: types.StringValue("10-20")},
					},
					IPsBlocked: ipsBlocked,
				},
			},
		},
		{
			name:  "all_null",
			input: []XrayFreedomSettings{{}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expanded := expandFreedomSettingsFromModel(tc.input)
			flat := flattenFreedomSettingsToModel(expanded)
			if len(flat) != len(tc.input) {
				t.Fatalf("count mismatch: got %d want %d", len(flat), len(tc.input))
			}
			for i, fs := range tc.input {
				if flat[i].DomainStrategy.ValueString() != fs.DomainStrategy.ValueString() {
					t.Fatalf("domain_strategy[%d] mismatch: got %q want %q", i, flat[i].DomainStrategy.ValueString(), fs.DomainStrategy.ValueString())
				}
			}
		})
	}
}

func TestExpandFlatten_FreedomOutJSON_RoundTrip(t *testing.T) {
	data := map[string]any{
		"freedom_settings": []any{
			map[string]any{
				"domain_strategy": "AsIs",
				"redirect":        "127.0.0.1:0",
				"fragment": []any{
					map[string]any{
						"packets":  "tlshello",
						"length":   "100-200",
						"interval": "10-20",
					},
				},
			},
		},
	}
	xrayJSON := expandFreedomSettings(data)
	if xrayJSON == nil {
		t.Fatal("expected non-nil xrayJSON")
	}
	if xrayJSON["domainStrategy"] != "AsIs" {
		t.Fatalf("domainStrategy mismatch: got %v want %v", xrayJSON["domainStrategy"], "AsIs")
	}
	flat := flattenFreedomSettings(xrayJSON)
	if flat["domain_strategy"] != "AsIs" {
		t.Fatalf("domain_strategy mismatch: got %v want %v", flat["domain_strategy"], "AsIs")
	}
}

// --- Wireguard ---

func TestExpandFlatten_WireguardOutModel_RoundTrip(t *testing.T) {
	addr, _ := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("172.16.0.1/32"),
	})
	reserved, _ := types.ListValue(types.Int64Type, []attr.Value{
		types.Int64Value(10),
		types.Int64Value(20),
	})
	allowedIPs, _ := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("0.0.0.0/0"),
	})
	cases := []struct {
		name  string
		input []XrayWireguardOutSettings
	}{
		{name: "empty", input: nil},
		{
			name: "full",
			input: []XrayWireguardOutSettings{
				{
					MTU:            types.Int64Value(1280),
					SecretKey:      types.StringValue("secret"),
					Address:        addr,
					Workers:        types.Int64Value(2),
					DomainStrategy: types.StringValue("UseIPv4v6"),
					Reserved:       reserved,
					NoKernelTun:    types.BoolValue(true),
					Peer: []XrayWireguardPeer{
						{PublicKey: types.StringValue("pub"), PreSharedKey: types.StringValue("psk"), AllowedIPs: allowedIPs, Endpoint: types.StringValue("1.2.3.4:51820"), KeepAlive: types.Int64Value(0)},
					},
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expanded := expandWireguardSettingsFromModel(tc.input)
			flat := flattenWireguardSettingsToModel(expanded)
			if len(flat) != len(tc.input) {
				t.Fatalf("count mismatch: got %d want %d", len(flat), len(tc.input))
			}
			for i, wg := range tc.input {
				if flat[i].MTU.ValueInt64() != wg.MTU.ValueInt64() {
					t.Fatalf("mtu[%d] mismatch: got %d want %d", i, flat[i].MTU.ValueInt64(), wg.MTU.ValueInt64())
				}
			}
		})
	}
}

func TestExpandFlatten_WireguardOutJSON_RoundTrip(t *testing.T) {
	data := map[string]any{
		"wireguard_settings": []any{
			map[string]any{
				"mtu":            1280,
				"secret_key":     "priv",
				"address":        []any{"172.16.0.1/32"},
				"workers":        2,
				"domain_strategy": "UseIP",
				"reserved":       []any{float64(10), float64(20)},
				"no_kernel_tun":  true,
			},
		},
	}
	xrayJSON := expandWireguardOutSettings(data)
	if xrayJSON == nil {
		t.Fatal("expected non-nil xrayJSON")
	}
	flat := flattenWireguardOutSettings(xrayJSON)
	if flat["secret_key"] != "priv" {
		t.Fatalf("secret_key mismatch: got %v want %v", flat["secret_key"], "priv")
	}
	if flat["no_kernel_tun"] != true {
		t.Fatalf("no_kernel_tun mismatch: got %v want %v", flat["no_kernel_tun"], true)
	}
}
