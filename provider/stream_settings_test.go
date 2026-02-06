package provider

import (
	"testing"
)

func TestExpandRealitySettings_Empty(t *testing.T) {
	result := expandRealitySettings([]any{})
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

func TestExpandRealitySettings_Full(t *testing.T) {
	input := []any{
		map[string]any{
			"show":           true,
			"xver":           2,
			"target":         "example.com:443",
			"server_names":   []any{"example.com"},
			"private_key":    "privkey",
			"min_client_ver": "1.0",
			"max_client_ver": "2.0",
			"max_timediff":   60,
			"short_ids":      []any{"abc123"},
			"mldsa65_seed":   "seed",
			"settings":       []any{map[string]any{"public_key": "pubkey", "fingerprint": "chrome"}},
		},
	}
	result := expandRealitySettings(input)
	if result == nil {
		t.Fatalf("expected non-nil")
	}
	if result["show"] != true {
		t.Fatalf("unexpected show: %v", result["show"])
	}
	if result["target"] != "example.com:443" {
		t.Fatalf("unexpected target: %v", result["target"])
	}
	if result["privateKey"] != "privkey" {
		t.Fatalf("unexpected privateKey: %v", result["privateKey"])
	}
	settings, ok := result["settings"].(map[string]any)
	if !ok {
		t.Fatalf("expected settings map")
	}
	if settings["publicKey"] != "pubkey" {
		t.Fatalf("unexpected publicKey: %v", settings["publicKey"])
	}
}

func TestExpandRealitySettings_NoTargetDefaults(t *testing.T) {
	input := []any{
		map[string]any{},
	}
	result := expandRealitySettings(input)
	if result == nil {
		t.Fatalf("expected non-nil")
	}
	if result["target"] != "www.apple.com:443" {
		t.Fatalf("expected default target, got %v", result["target"])
	}
}

func TestExpandRealitySettings_TargetDerivesServerNames(t *testing.T) {
	input := []any{
		map[string]any{
			"target": "myhost.com:443",
		},
	}
	result := expandRealitySettings(input)
	if result == nil {
		t.Fatalf("expected non-nil")
	}
	names, ok := result["serverNames"].([]any)
	if !ok {
		t.Fatalf("expected serverNames")
	}
	if len(names) == 0 {
		t.Fatalf("expected at least one server name")
	}
	if names[0] != "myhost.com" {
		t.Fatalf("expected myhost.com, got %v", names[0])
	}
}

func TestFlattenRealitySettings_Nil(t *testing.T) {
	result := flattenRealitySettings(nil)
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

func TestFlattenRealitySettings_Empty(t *testing.T) {
	result := flattenRealitySettings(map[string]any{})
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

func TestFlattenRealitySettings_Full(t *testing.T) {
	in := map[string]any{
		"show":         true,
		"xver":         float64(2),
		"target":       "example.com:443",
		"serverNames":  []any{"example.com"},
		"privateKey":   "pk",
		"minClientVer": "1.0",
		"maxClientVer": "2.0",
		"maxTimediff":  float64(60),
		"shortIds":     []any{"abc"},
		"mldsa65Seed":  "seed",
		"settings":     map[string]any{"publicKey": "pub", "fingerprint": "chrome"},
	}
	out := flattenRealitySettings(in)
	if out == nil {
		t.Fatalf("expected non-nil")
	}
	if out["show"] != true {
		t.Fatalf("unexpected show: %v", out["show"])
	}
	if out["target"] != "example.com:443" {
		t.Fatalf("unexpected target: %v", out["target"])
	}
	if out["private_key"] != "pk" {
		t.Fatalf("unexpected private_key: %v", out["private_key"])
	}
	if out["xver"] != 2 {
		t.Fatalf("unexpected xver: %v", out["xver"])
	}
	settings, ok := out["settings"].([]any)
	if !ok || len(settings) != 1 {
		t.Fatalf("expected settings list with 1 element")
	}
}

func TestFlattenRealitySettings_Partial(t *testing.T) {
	in := map[string]any{
		"target":     "example.com:443",
		"privateKey": "pk",
	}
	out := flattenRealitySettings(in)
	if out == nil {
		t.Fatalf("expected non-nil")
	}
	if out["target"] != "example.com:443" {
		t.Fatalf("unexpected target")
	}
}

func TestExpandRealityInnerSettings_Empty(t *testing.T) {
	result := expandRealityInnerSettings([]any{})
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

func TestExpandRealityInnerSettings_Full(t *testing.T) {
	input := []any{
		map[string]any{
			"public_key":     "pub",
			"fingerprint":    "chrome",
			"server_name":    "example.com",
			"spider_x":       "/",
			"mldsa65_verify": "verify",
		},
	}
	result := expandRealityInnerSettings(input)
	if result == nil {
		t.Fatalf("expected non-nil")
	}
	if result["publicKey"] != "pub" {
		t.Fatalf("unexpected publicKey: %v", result["publicKey"])
	}
	if result["fingerprint"] != "chrome" {
		t.Fatalf("unexpected fingerprint: %v", result["fingerprint"])
	}
	if result["serverName"] != "example.com" {
		t.Fatalf("unexpected serverName")
	}
	if result["spiderX"] != "/" {
		t.Fatalf("unexpected spiderX")
	}
}

func TestExpandRealityInnerSettings_PublicKeyOnly(t *testing.T) {
	input := []any{
		map[string]any{"public_key": "pub"},
	}
	result := expandRealityInnerSettings(input)
	if result == nil {
		t.Fatalf("expected non-nil")
	}
	if result["publicKey"] != "pub" {
		t.Fatalf("unexpected publicKey")
	}
	if _, ok := result["fingerprint"]; ok {
		t.Fatalf("fingerprint should not be set")
	}
}

func TestFlattenRealityInnerSettings_Empty(t *testing.T) {
	result := flattenRealityInnerSettings(map[string]any{})
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

func TestFlattenRealityInnerSettings_Full(t *testing.T) {
	in := map[string]any{
		"publicKey":     "pub",
		"fingerprint":   "chrome",
		"serverName":    "example.com",
		"spiderX":       "/",
		"mldsa65Verify": "verify",
	}
	out := flattenRealityInnerSettings(in)
	if out == nil {
		t.Fatalf("expected non-nil")
	}
	if out["public_key"] != "pub" {
		t.Fatalf("unexpected public_key")
	}
	if out["fingerprint"] != "chrome" {
		t.Fatalf("unexpected fingerprint")
	}
	if out["server_name"] != "example.com" {
		t.Fatalf("unexpected server_name")
	}
	if out["spider_x"] != "/" {
		t.Fatalf("unexpected spider_x")
	}
	if out["mldsa65_verify"] != "verify" {
		t.Fatalf("unexpected mldsa65_verify")
	}
}
