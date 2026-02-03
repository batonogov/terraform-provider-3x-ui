package provider

import "testing"

func TestNewUUID(t *testing.T) {
	first, err := newUUID()
	if err != nil {
		t.Fatalf("newUUID error: %v", err)
	}
	second, err := newUUID()
	if err != nil {
		t.Fatalf("newUUID error: %v", err)
	}
	if first == second {
		t.Fatalf("expected unique UUIDs")
	}
	if len(first) != 36 {
		t.Fatalf("unexpected uuid length: %d", len(first))
	}
}
