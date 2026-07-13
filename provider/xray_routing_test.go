package provider

import (
	"testing"
)

// TestIsManagedDnsAllowRule mirrors 3x-ui v3.5.0's service.dnsAllowRuleShape.
// The panel auto-inserts these direct allow-rules for private DNS servers
// before the geoip:private block on every xray-template save; the provider
// must filter them out of Read to avoid permanent drift.
func TestIsManagedDnsAllowRule(t *testing.T) {
	cases := []struct {
		name string
		rule map[string]any
		want bool
	}{
		{
			name: "canonical managed rule",
			rule: map[string]any{
				"type": "field", "outboundTag": "direct",
				"ip": []any{"10.0.0.53"}, "port": "53",
			},
			want: true,
		},
		{
			name: "managed rule with enabled=true tolerated",
			rule: map[string]any{
				"type": "field", "outboundTag": "direct",
				"ip": []any{"10.0.0.53"}, "port": "53",
				"enabled": true,
			},
			want: true,
		},
		{
			name: "snake_case outbound_tag also recognised",
			rule: map[string]any{
				"type": "field", "outbound_tag": "direct",
				"ip": []any{"10.0.0.53"}, "port": "53",
			},
			want: true,
		},
		{
			name: "enabled=false is no longer managed",
			rule: map[string]any{
				"type": "field", "outboundTag": "direct",
				"ip": []any{"10.0.0.53"}, "port": "53",
				"enabled": false,
			},
			want: false,
		},
		{
			name: "extra matcher (domain) is NOT managed",
			rule: map[string]any{
				"type": "field", "outboundTag": "direct",
				"ip": []any{"10.0.0.53"}, "port": "53",
				"domain": []any{"example.com"},
			},
			want: false,
		},
		{
			name: "missing ip is not managed",
			rule: map[string]any{
				"type": "field", "outboundTag": "direct", "port": "53",
			},
			want: false,
		},
		{
			name: "missing port is not managed",
			rule: map[string]any{
				"type": "field", "outboundTag": "direct", "ip": []any{"10.0.0.53"},
			},
			want: false,
		},
		{
			name: "outboundTag not direct is not managed",
			rule: map[string]any{
				"type": "field", "outboundTag": "blocked",
				"ip": []any{"geoip:private"}, "port": "53",
			},
			want: false,
		},
		{
			name: "type not field is not managed",
			rule: map[string]any{
				"type": "default", "outboundTag": "direct",
				"ip": []any{"10.0.0.53"}, "port": "53",
			},
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isManagedDnsAllowRule(tc.rule); got != tc.want {
				t.Fatalf("isManagedDnsAllowRule(%v) = %v, want %v", tc.rule, got, tc.want)
			}
		})
	}
}

// TestFlattenRoutingRulesFiltersManagedDns confirms the wire→untyped flatten
// path drops both API routing rules and managed DNS allow-rules so they never
// surface as drift in threexui_xray_routing's rules block.
func TestFlattenRoutingRulesFiltersManagedDns(t *testing.T) {
	list := []any{
		// user rule — must survive
		map[string]any{
			"type": "field", "outboundTag": "blocked",
			"ip": []any{"geoip:private"},
		},
		// managed DNS allow-rule — must be dropped
		map[string]any{
			"type": "field", "outboundTag": "direct",
			"ip": []any{"10.0.0.53"}, "port": "53",
		},
		// API routing rule — must be dropped
		map[string]any{
			"type": "field", "inboundTag": []string{"api"}, "outboundTag": "api",
		},
	}
	out := flattenRoutingRules(list)
	if len(out) != 1 {
		t.Fatalf("expected 1 surviving rule (user rule only), got %d: %#v", len(out), out)
	}
	survivor, _ := out[0].(map[string]any)
	if survivor["outbound_tag"] != "blocked" {
		t.Fatalf("expected the user geoip:private block rule to survive, got %v", survivor["outbound_tag"])
	}
}

// TestExpandRoutingRulesFiltersManagedDns mirrors the flatten-path test above
// for the write/expand path. expandRoutingRules reads snake_case keys (model
// representation) and is the other call site that filters managed rules;
// without this test the expand-path `|| isManagedDnsAllowRule(m)` clause would
// be uncovered. It also exercises the snake_case `outbound_tag` recognition in
// isManagedDnsAllowRule (flatten-path rules arrive in camelCase, so the snake
// branch is only reached here).
func TestExpandRoutingRulesFiltersManagedDns(t *testing.T) {
	list := []any{
		// user rule (snake_case input) — must survive; expandRoutingRules
		// translates outbound_tag→outboundTag (snake→camel) on output.
		map[string]any{
			"type": "field", "outbound_tag": "blocked",
			"ip": []any{"geoip:private"},
		},
		// managed DNS allow-rule (snake_case) — must be dropped
		map[string]any{
			"type": "field", "outbound_tag": "direct",
			"ip": []any{"10.0.0.53"}, "port": "53",
		},
	}
	out := expandRoutingRules(list)
	if len(out) != 1 {
		t.Fatalf("expected 1 surviving rule, got %d: %#v", len(out), out)
	}
	survivor, _ := out[0].(map[string]any)
	if survivor["outboundTag"] != "blocked" {
		t.Fatalf("expected the user block rule to survive, got %v", survivor["outboundTag"])
	}
}
