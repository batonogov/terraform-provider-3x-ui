package provider

import (
	"reflect"
	"testing"
)

func TestParseJSONField_Empty(t *testing.T) {
	out, err := ParseJSONField("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil && len(out) != 0 {
		t.Fatalf("expected empty map or nil, got: %#v", out)
	}
}

func TestParseJSONField_OK(t *testing.T) {
	input := `{"a":1,"b":"x","c":true}`
	out, err := ParseJSONField(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := map[string]any{"a": float64(1), "b": "x", "c": true}
	if !reflect.DeepEqual(out, expected) {
		t.Fatalf("unexpected result: %#v", out)
	}
}

func TestParseJSONField_Bad(t *testing.T) {
	_, err := ParseJSONField("{")
	if err == nil {
		t.Fatalf("expected error")
	}
}
