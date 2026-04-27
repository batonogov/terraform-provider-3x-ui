package provider

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExpandFallbacks_Empty(t *testing.T) {
	result := expandFallbacks([]any{})
	if len(result) != 0 {
		t.Fatalf("expected empty, got %v", result)
	}
}

func TestExpandFallbacks_Single(t *testing.T) {
	input := []any{
		map[string]any{"name": "fb1", "alpn": "h2", "path": "/ws", "dest": "127.0.0.1:8080", "xver": 1},
	}
	result := expandFallbacks(input)
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	m, _ := result[0].(map[string]any)
	if m["name"] != "fb1" {
		t.Fatalf("unexpected name: %v", m["name"])
	}
	if m["xver"] != 1 {
		t.Fatalf("unexpected xver: %v", m["xver"])
	}
}

func TestExpandFallbacks_Multiple(t *testing.T) {
	input := []any{
		map[string]any{"dest": "a"},
		map[string]any{"dest": "b"},
	}
	result := expandFallbacks(input)
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
}

func TestExpandFallbacks_MissingOptional(t *testing.T) {
	input := []any{
		map[string]any{"dest": "127.0.0.1:80"},
	}
	result := expandFallbacks(input)
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	m, _ := result[0].(map[string]any)
	if _, ok := m["name"]; ok {
		t.Fatalf("name should not be set")
	}
}

func TestFlattenFallbacks_Empty(t *testing.T) {
	result := flattenFallbacks([]any{})
	if len(result) != 0 {
		t.Fatalf("expected empty, got %v", result)
	}
}

func TestFlattenFallbacks_Single(t *testing.T) {
	input := []any{
		map[string]any{"name": "fb", "alpn": "h2", "path": "/", "dest": "target", "xver": float64(2)},
	}
	result := flattenFallbacks(input)
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	m, _ := result[0].(map[string]any)
	if m["name"] != "fb" {
		t.Fatalf("unexpected name: %v", m["name"])
	}
	if m["xver"] != 2 {
		t.Fatalf("unexpected xver: %v", m["xver"])
	}
}

func TestExpandAccounts_Empty(t *testing.T) {
	result := expandAccounts([]any{})
	if len(result) != 0 {
		t.Fatalf("expected empty, got %v", result)
	}
}

func TestExpandAccounts_Single(t *testing.T) {
	input := []any{
		map[string]any{"user": "admin", "pass": "secret"},
	}
	result := expandAccounts(input)
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	m, _ := result[0].(map[string]any)
	if m["user"] != "admin" || m["pass"] != "secret" {
		t.Fatalf("unexpected: %v", m)
	}
}

func TestFlattenAccounts_Empty(t *testing.T) {
	result := flattenAccounts([]any{})
	if len(result) != 0 {
		t.Fatalf("expected empty, got %v", result)
	}
}

func TestFlattenAccounts_Single(t *testing.T) {
	input := []any{
		map[string]any{"user": "u", "pass": "p"},
	}
	result := flattenAccounts(input)
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	m, _ := result[0].(map[string]any)
	if m["user"] != "u" || m["pass"] != "p" {
		t.Fatalf("unexpected: %v", m)
	}
}

func TestExpandPeers_Empty(t *testing.T) {
	result := expandPeers([]any{})
	if len(result) != 0 {
		t.Fatalf("expected empty, got %v", result)
	}
}

func TestExpandPeers_Full(t *testing.T) {
	input := []any{
		map[string]any{
			"private_key":    "priv",
			"public_key":     "pub",
			"pre_shared_key": "psk",
			"allowed_ips":    []any{"10.0.0.0/24"},
			"keep_alive":     30,
		},
	}
	result := expandPeers(input)
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	m, _ := result[0].(map[string]any)
	if m["privateKey"] != "priv" {
		t.Fatalf("unexpected privateKey: %v", m["privateKey"])
	}
	if m["publicKey"] != "pub" {
		t.Fatalf("unexpected publicKey: %v", m["publicKey"])
	}
	if m["keepAlive"] != 30 {
		t.Fatalf("unexpected keepAlive: %v", m["keepAlive"])
	}
}

func TestExpandPeers_Minimal(t *testing.T) {
	input := []any{
		map[string]any{"public_key": "pub", "allowed_ips": []any{}},
	}
	result := expandPeers(input)
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
}

func TestFlattenPeers_Empty(t *testing.T) {
	result := flattenPeers([]any{})
	if len(result) != 0 {
		t.Fatalf("expected empty, got %v", result)
	}
}

func TestFlattenPeers_Full(t *testing.T) {
	input := []any{
		map[string]any{
			"privateKey":   "priv",
			"publicKey":    "pub",
			"preSharedKey": "psk",
			"allowedIPs":   []any{"10.0.0.0/24"},
			"keepAlive":    float64(30),
		},
	}
	result := flattenPeers(input)
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	m, _ := result[0].(map[string]any)
	if m["private_key"] != "priv" {
		t.Fatalf("unexpected private_key: %v", m["private_key"])
	}
	if m["keep_alive"] != 30 {
		t.Fatalf("unexpected keep_alive: %v", m["keep_alive"])
	}
}

func TestFlattenStringMap_Normal(t *testing.T) {
	in := map[string]any{"a": "1", "b": "2"}
	out := flattenStringMap(in)
	expected := map[string]string{"a": "1", "b": "2"}
	if !reflect.DeepEqual(out, expected) {
		t.Fatalf("unexpected: %v", out)
	}
}

func TestFlattenStringMap_MixedTypes(t *testing.T) {
	in := map[string]any{"a": "str", "b": 42}
	out := flattenStringMap(in)
	if out["a"] != "str" {
		t.Fatalf("expected str for a")
	}
	if _, ok := out["b"]; ok {
		t.Fatalf("non-string should be skipped")
	}
}

func TestFlattenIntList_Empty(t *testing.T) {
	result := flattenIntList([]any{})
	if len(result) != 0 {
		t.Fatalf("expected empty, got %v", result)
	}
}

func TestFlattenIntList_Float64Values(t *testing.T) {
	input := []any{float64(1), float64(2), float64(3)}
	result := flattenIntList(input)
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestBuildFlattenSettings_DokodemoPortMap(t *testing.T) {
	input := map[string]any{
		"address":         "127.0.0.1",
		"port":            80,
		"port_map":        map[string]any{"80": "http", "443": "https"},
		"network":         "tcp,udp",
		"follow_redirect": true,
	}
	jsonStr := buildSettingsJSON(input)
	result, err := flattenSettings(jsonStr)
	if err != nil {
		t.Fatalf("flattenSettings error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	m, ok := result[0].(map[string]any)
	if !ok {
		t.Fatal("expected map[string]any")
	}
	if m["address"] != "127.0.0.1" {
		t.Fatalf("unexpected address: %v", m["address"])
	}
	if m["follow_redirect"] != true {
		t.Fatalf("unexpected follow_redirect: %v", m["follow_redirect"])
	}
	pm, ok := m["port_map"].(map[string]string)
	if !ok {
		t.Fatalf("port_map not map[string]string: %T", m["port_map"])
	}
	if pm["80"] != "http" || pm["443"] != "https" {
		t.Fatalf("unexpected port_map: %v", pm)
	}
}

func TestBuildFlattenSettings_Mixed(t *testing.T) {
	input := map[string]any{
		"auth":     "password",
		"accounts": []any{map[string]any{"user": "mixeduser", "pass": "mixedpass"}},
		"udp":      true,
		"ip":       "127.0.0.1",
	}
	jsonStr := buildSettingsJSON(input)
	result, err := flattenSettings(jsonStr)
	if err != nil {
		t.Fatalf("flattenSettings error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	m, ok := result[0].(map[string]any)
	if !ok {
		t.Fatal("expected map[string]any")
	}
	if m["auth"] != "password" {
		t.Fatalf("unexpected auth: %v", m["auth"])
	}
	if m["udp"] != true {
		t.Fatalf("unexpected udp: %v", m["udp"])
	}
	if m["ip"] != "127.0.0.1" {
		t.Fatalf("unexpected ip: %v", m["ip"])
	}
	accounts, ok := m["accounts"].([]any)
	if !ok || len(accounts) != 1 {
		t.Fatalf("unexpected accounts: %v", m["accounts"])
	}
	acc, _ := accounts[0].(map[string]any)
	if acc["user"] != "mixeduser" || acc["pass"] != "mixedpass" {
		t.Fatalf("unexpected account: %v", acc)
	}
}

func TestExpandFlattenMixedSettingsModel_RoundTrip(t *testing.T) {
	model := &InboundResourceModel{
		MixedSettings: &InboundMixedSettingsModel{
			Auth: types.StringValue("password"),
			UDP:  types.BoolValue(true),
			IP:   types.StringValue("127.0.0.1"),
			Account: []InboundAccountModel{
				{User: types.StringValue("u1"), Pass: types.StringValue("p1")},
			},
		},
	}

	expanded := expandSettingsFromModel("mixed", model)
	if expanded == nil {
		t.Fatal("expected non-nil result for mixed protocol")
	}
	if expanded["auth"] != "password" {
		t.Fatalf("unexpected auth: %v", expanded["auth"])
	}
	if expanded["udp"] != true {
		t.Fatalf("unexpected udp: %v", expanded["udp"])
	}
	if expanded["ip"] != "127.0.0.1" {
		t.Fatalf("unexpected ip: %v", expanded["ip"])
	}

	// Build JSON and flatten back
	jsonStr := buildSettingsJSON(expanded)
	flatMap, err := flattenSettingsToMap(jsonStr)
	if err != nil {
		t.Fatalf("flattenSettingsToMap error: %v", err)
	}

	result := &InboundResourceModel{}
	flattenSettingsToModel("mixed", flatMap, result)
	if result.MixedSettings == nil {
		t.Fatal("expected MixedSettings to be set")
	}
	if result.MixedSettings.Auth.ValueString() != "password" {
		t.Fatalf("unexpected auth: %s", result.MixedSettings.Auth.ValueString())
	}
	if result.MixedSettings.UDP.ValueBool() != true {
		t.Fatalf("unexpected udp: %v", result.MixedSettings.UDP.ValueBool())
	}
	if result.MixedSettings.IP.ValueString() != "127.0.0.1" {
		t.Fatalf("unexpected ip: %s", result.MixedSettings.IP.ValueString())
	}
	if len(result.MixedSettings.Account) != 1 {
		t.Fatalf("expected 1 account, got %d", len(result.MixedSettings.Account))
	}
	if result.MixedSettings.Account[0].User.ValueString() != "u1" {
		t.Fatalf("unexpected user: %s", result.MixedSettings.Account[0].User.ValueString())
	}
}

func TestExpandSettingsFromModel_MixedNil(t *testing.T) {
	model := &InboundResourceModel{
		MixedSettings: nil,
	}
	result := expandSettingsFromModel("mixed", model)
	if result != nil {
		t.Fatalf("expected nil for nil MixedSettings, got %v", result)
	}
}

func TestFlattenSettingsToModel_MixedEmptyData(t *testing.T) {
	model := &InboundResourceModel{}
	flattenSettingsToModel("mixed", map[string]any{}, model)
	if model.MixedSettings != nil {
		t.Fatal("expected nil MixedSettings for empty data")
	}
}

// TestExpandFlattenSettingsModel_Hysteria2Alias locks in that "hysteria2"
// (introduced in 3x-ui v2.9.3) routes through the same settings model as
// "hysteria". This is the static counterpart of an acceptance test — the
// panel may normalize the literal back to "hysteria" + settings.version=2,
// which would cause plan-drift in an acc test, so the roundtrip is validated
// here at the unit layer instead.
func TestExpandFlattenSettingsModel_Hysteria2Alias(t *testing.T) {
	model := &InboundResourceModel{
		HysteriaSettings: &InboundHysteriaSettingsModel{
			Version: types.Int64Value(2),
		},
	}

	expanded := expandSettingsFromModel("hysteria2", model)
	if expanded == nil {
		t.Fatal("expected non-nil result for hysteria2 protocol")
	}
	if expanded["version"] != 2 {
		t.Fatalf("unexpected version: %v", expanded["version"])
	}

	jsonStr := buildSettingsJSON(expanded)
	flatMap, err := flattenSettingsToMap(jsonStr)
	if err != nil {
		t.Fatalf("flattenSettingsToMap error: %v", err)
	}

	result := &InboundResourceModel{}
	flattenSettingsToModel("hysteria2", flatMap, result)
	if result.HysteriaSettings == nil {
		t.Fatal("expected HysteriaSettings to be set for hysteria2 protocol")
	}
	if result.HysteriaSettings.Version.ValueInt64() != 2 {
		t.Fatalf("unexpected version: %d", result.HysteriaSettings.Version.ValueInt64())
	}
}

func TestExpandSettingsFromModel_TunnelProtocol(t *testing.T) {
	model := &InboundResourceModel{
		DokodemoSettings: &InboundDokodemoSettingsModel{},
	}
	// tunnel should dispatch to the same expand as dokodemo-door
	result := expandSettingsFromModel("tunnel", model)
	if result == nil {
		t.Fatal("expected non-nil result for tunnel protocol")
	}
}

func TestFlattenSettingsToModel_TunnelProtocol(t *testing.T) {
	data := map[string]any{
		"address": "10.0.0.1",
		"port":    8080,
		"network": "tcp",
	}
	model := &InboundResourceModel{}
	flattenSettingsToModel("tunnel", data, model)
	if model.DokodemoSettings == nil {
		t.Fatal("expected DokodemoSettings to be set for tunnel protocol")
	}
	if model.DokodemoSettings.Address.ValueString() != "10.0.0.1" {
		t.Fatalf("unexpected address: %s", model.DokodemoSettings.Address.ValueString())
	}
}
