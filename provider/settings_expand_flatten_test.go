package provider

import (
	"reflect"
	"testing"
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
