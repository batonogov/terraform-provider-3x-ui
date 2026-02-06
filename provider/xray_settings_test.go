package provider

import (
	"reflect"
	"testing"
)

func TestDeepMergeJSON_Flat(t *testing.T) {
	dst := map[string]any{"a": "1"}
	src := map[string]any{"b": "2"}
	result := deepMergeJSON(dst, src)
	if result["a"] != "1" || result["b"] != "2" {
		t.Fatalf("unexpected result: %v", result)
	}
}

func TestDeepMergeJSON_Overlap(t *testing.T) {
	dst := map[string]any{"a": "old"}
	src := map[string]any{"a": "new"}
	result := deepMergeJSON(dst, src)
	if result["a"] != "new" {
		t.Fatalf("expected overwrite, got %v", result["a"])
	}
}

func TestDeepMergeJSON_Nested(t *testing.T) {
	dst := map[string]any{"a": map[string]any{"x": 1}}
	src := map[string]any{"a": map[string]any{"y": 2}}
	result := deepMergeJSON(dst, src)
	inner, ok := result["a"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested map")
	}
	if inner["x"] != 1 || inner["y"] != 2 {
		t.Fatalf("unexpected nested result: %v", inner)
	}
}

func TestDeepMergeJSON_NilDst(t *testing.T) {
	src := map[string]any{"a": "1"}
	result := deepMergeJSON(nil, src)
	if result["a"] != "1" {
		t.Fatalf("unexpected result: %v", result)
	}
}

func TestDeepMergeJSON_MapReplacedByScalar(t *testing.T) {
	dst := map[string]any{"a": map[string]any{"x": 1}}
	src := map[string]any{"a": "scalar"}
	result := deepMergeJSON(dst, src)
	if result["a"] != "scalar" {
		t.Fatalf("expected scalar, got %v", result["a"])
	}
}

func TestSetJSONPath_SingleLevel(t *testing.T) {
	root := map[string]any{}
	setJSONPath(root, []string{"key"}, "val")
	if root["key"] != "val" {
		t.Fatalf("unexpected: %v", root)
	}
}

func TestSetJSONPath_MultiLevel(t *testing.T) {
	root := map[string]any{}
	setJSONPath(root, []string{"a", "b", "c"}, 42)
	a, _ := root["a"].(map[string]any)
	b, _ := a["b"].(map[string]any)
	if b["c"] != 42 {
		t.Fatalf("unexpected: %v", root)
	}
}

func TestSetJSONPath_CreateIntermediates(t *testing.T) {
	root := map[string]any{"x": "keep"}
	setJSONPath(root, []string{"a", "b"}, "new")
	if root["x"] != "keep" {
		t.Fatalf("existing key lost")
	}
	a, _ := root["a"].(map[string]any)
	if a["b"] != "new" {
		t.Fatalf("unexpected: %v", root)
	}
}

func TestSetJSONPath_Overwrite(t *testing.T) {
	root := map[string]any{"a": "old"}
	setJSONPath(root, []string{"a"}, "new")
	if root["a"] != "new" {
		t.Fatalf("expected overwrite")
	}
}

func TestGetJSONPath_Existing(t *testing.T) {
	root := map[string]any{"a": map[string]any{"b": "val"}}
	got := getJSONPath(root, []string{"a", "b"})
	if got != "val" {
		t.Fatalf("expected val, got %v", got)
	}
}

func TestGetJSONPath_Missing(t *testing.T) {
	root := map[string]any{"a": "val"}
	got := getJSONPath(root, []string{"b"})
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestGetJSONPath_NonMapIntermediate(t *testing.T) {
	root := map[string]any{"a": "string"}
	got := getJSONPath(root, []string{"a", "b"})
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestCloneJSONMap_Nil(t *testing.T) {
	result := cloneJSONMap(nil)
	if result == nil || len(result) != 0 {
		t.Fatalf("expected empty map, got %v", result)
	}
}

func TestCloneJSONMap_Independence(t *testing.T) {
	original := map[string]any{"a": "1", "b": "2"}
	clone := cloneJSONMap(original)
	clone["a"] = "changed"
	if original["a"] != "1" {
		t.Fatalf("clone modified original")
	}
}

func TestNormalizeJSONString_Valid(t *testing.T) {
	result := normalizeJSONString(`{ "b": 1, "a": 2 }`)
	if result != `{"a":2,"b":1}` {
		t.Fatalf("unexpected: %q", result)
	}
}

func TestNormalizeJSONString_Invalid(t *testing.T) {
	result := normalizeJSONString("{bad")
	if result != "{bad" {
		t.Fatalf("expected raw string back, got %q", result)
	}
}

func TestNormalizeJSONString_Empty(t *testing.T) {
	result := normalizeJSONString("")
	if result != "" {
		t.Fatalf("expected empty, got %q", result)
	}
}

func TestJsonEqualDiffSuppress_EqualFormatting(t *testing.T) {
	if !jsonEqualDiffSuppress("k", `{"a":1}`, `{ "a": 1 }`, nil) {
		t.Fatalf("expected equal")
	}
}

func TestJsonEqualDiffSuppress_DifferentValues(t *testing.T) {
	if jsonEqualDiffSuppress("k", `{"a":1}`, `{"a":2}`, nil) {
		t.Fatalf("expected not equal")
	}
}

func TestJsonEqualDiffSuppress_BothEmpty(t *testing.T) {
	if !jsonEqualDiffSuppress("k", "", "", nil) {
		t.Fatalf("expected equal for empty")
	}
}

func TestJsonEqualDiffSuppress_OneEmpty(t *testing.T) {
	if jsonEqualDiffSuppress("k", `{"a":1}`, "", nil) {
		t.Fatalf("expected not equal when one empty")
	}
}

func TestJsonEqualDiffSuppress_InvalidJSON(t *testing.T) {
	if jsonEqualDiffSuppress("k", "{bad", `{"a":1}`, nil) {
		t.Fatalf("expected false for invalid JSON")
	}
}

func TestDeepEqualJSON_EqualMaps(t *testing.T) {
	a := map[string]any{"x": float64(1)}
	b := map[string]any{"x": float64(1)}
	if !deepEqualJSON(a, b) {
		t.Fatalf("expected equal")
	}
}

func TestDeepEqualJSON_DifferentMaps(t *testing.T) {
	a := map[string]any{"x": float64(1)}
	b := map[string]any{"x": float64(2)}
	if deepEqualJSON(a, b) {
		t.Fatalf("expected not equal")
	}
}

func TestDeepEqualJSON_Arrays(t *testing.T) {
	a := []any{float64(1), float64(2)}
	b := []any{float64(1), float64(2)}
	if !deepEqualJSON(a, b) {
		t.Fatalf("expected equal")
	}
}

func TestDeepEqualJSON_DifferentTypes(t *testing.T) {
	if deepEqualJSON("string", float64(1)) {
		t.Fatalf("expected not equal for different types")
	}
}

func TestExtractXraySection_MergeRoot(t *testing.T) {
	current := map[string]any{
		"log":       map[string]any{"level": "debug"},
		"policy":    map[string]any{},
		"routing":   map[string]any{"rules": []any{}},
		"outbounds": []any{},
		"dns":       map[string]any{"servers": []any{}},
	}
	result := extractXraySection(current, xraySectionBasics)
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map")
	}
	if _, ok := m["log"]; !ok {
		t.Fatalf("missing log")
	}
	if _, ok := m["dns"]; ok {
		t.Fatalf("dns should not be in basics")
	}
}

func TestExtractXraySection_SetPath(t *testing.T) {
	current := map[string]any{"dns": map[string]any{"servers": []any{"8.8.8.8"}}}
	result := extractXraySection(current, xraySectionDNS)
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map")
	}
	if _, ok := m["servers"]; !ok {
		t.Fatalf("missing servers")
	}
}

func TestExtractXraySection_ReplaceAll(t *testing.T) {
	current := map[string]any{"a": "1", "b": "2"}
	result := extractXraySection(current, xraySectionAdvanced)
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map")
	}
	if !reflect.DeepEqual(m, current) {
		t.Fatalf("expected full config")
	}
}

func TestApplyXraySection_MergeRoot(t *testing.T) {
	current := map[string]any{"log": map[string]any{"level": "info"}}
	desired := map[string]any{"log": map[string]any{"level": "debug"}}
	result, err := applyXraySection(current, desired, xraySectionBasics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	log, _ := result["log"].(map[string]any)
	if log["level"] != "debug" {
		t.Fatalf("expected debug, got %v", log["level"])
	}
}

func TestApplyXraySection_SetPath(t *testing.T) {
	current := map[string]any{}
	desired := map[string]any{"servers": []any{"8.8.8.8"}}
	result, err := applyXraySection(current, desired, xraySectionDNS)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dns, ok := result["dns"].(map[string]any)
	if !ok {
		t.Fatalf("expected dns key")
	}
	if _, ok := dns["servers"]; !ok {
		t.Fatalf("missing servers in dns")
	}
}

func TestApplyXraySection_ReplaceAll(t *testing.T) {
	current := map[string]any{"old": "data"}
	desired := map[string]any{"new": "data"}
	result, err := applyXraySection(current, desired, xraySectionAdvanced)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result["old"]; ok {
		t.Fatalf("old data should be gone")
	}
	if result["new"] != "data" {
		t.Fatalf("expected new data")
	}
}

func TestApplyXraySection_MergeRoot_NotObject(t *testing.T) {
	_, err := applyXraySection(map[string]any{}, "string", xraySectionBasics)
	if err == nil {
		t.Fatalf("expected error for non-object")
	}
}

func TestApplyXraySection_SetPath_EmptyPath(t *testing.T) {
	section := xraySection{id: "test", mode: xraySectionSetPath, path: []string{}}
	_, err := applyXraySection(map[string]any{}, map[string]any{}, section)
	if err == nil {
		t.Fatalf("expected error for empty path")
	}
}
