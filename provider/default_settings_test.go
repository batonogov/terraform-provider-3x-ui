package provider

import (
	"testing"
)

func TestDefaultSettingsForProtocol(t *testing.T) {
	tests := []struct {
		protocol string
		wantNil  bool
		wantKey  string
	}{
		{"vless", false, "decryption"},
		{"vmess", true, ""},
		{"trojan", true, ""},
		{"shadowsocks", true, ""},
		{"http", true, ""},
		{"unknown", true, ""},
		{"", true, ""},
	}
	for _, tt := range tests {
		result, err := defaultSettingsForProtocol(tt.protocol)
		if err != nil {
			t.Fatalf("defaultSettingsForProtocol(%q): unexpected error: %v", tt.protocol, err)
		}
		if tt.wantNil && result != nil {
			t.Fatalf("defaultSettingsForProtocol(%q): expected nil, got %v", tt.protocol, result)
		}
		if !tt.wantNil && result == nil {
			t.Fatalf("defaultSettingsForProtocol(%q): expected non-nil", tt.protocol)
		}
		if tt.wantKey != "" {
			if _, ok := result[tt.wantKey]; !ok {
				t.Fatalf("defaultSettingsForProtocol(%q): missing key %q", tt.protocol, tt.wantKey)
			}
		}
	}
}

func TestDefaultSettingsForProtocol_VlessFields(t *testing.T) {
	result, err := defaultSettingsForProtocol("vless")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["decryption"] != "none" {
		t.Fatalf("expected decryption=none, got %v", result["decryption"])
	}
	if result["encryption"] != "none" {
		t.Fatalf("expected encryption=none, got %v", result["encryption"])
	}
	testseed, ok := result["testseed"].([]any)
	if !ok || len(testseed) != 4 {
		t.Fatalf("expected testseed with 4 elements, got %v", result["testseed"])
	}
}

func TestProtocolUsesClients(t *testing.T) {
	tests := []struct {
		protocol string
		want     bool
	}{
		{"vmess", true},
		{"vless", true},
		{"trojan", true},
		{"shadowsocks", true},
		{"http", false},
		{"mixed", false},
		{"", false},
		{"wireguard", false},
	}
	for _, tt := range tests {
		got := protocolUsesClients(tt.protocol)
		if got != tt.want {
			t.Fatalf("protocolUsesClients(%q) = %v, want %v", tt.protocol, got, tt.want)
		}
	}
}

func TestApplyDefaultInboundSettings_Nil(t *testing.T) {
	err := applyDefaultInboundSettings(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyDefaultInboundSettings_EmptySettings(t *testing.T) {
	inbound := &Inbound{Protocol: "vless", Settings: ""}
	err := applyDefaultInboundSettings(inbound)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inbound.Settings == "" || inbound.Settings == "{}" {
		t.Fatalf("expected non-empty settings for vless, got %q", inbound.Settings)
	}
}

func TestApplyDefaultInboundSettings_ExistingSettings(t *testing.T) {
	inbound := &Inbound{Protocol: "vless", Settings: `{"decryption":"none","clients":[]}`}
	err := applyDefaultInboundSettings(inbound)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyDefaultInboundSettings_HttpProtocol(t *testing.T) {
	inbound := &Inbound{Protocol: "http", Settings: ""}
	err := applyDefaultInboundSettings(inbound)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inbound.Settings != "{}" {
		t.Fatalf("expected {} for http, got %q", inbound.Settings)
	}
}
