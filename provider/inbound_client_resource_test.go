package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFindClientByID(t *testing.T) {
	clients := []map[string]any{
		{"id": "uuid-1", "email": "a@example.com"},
		{"password": "pw", "email": "b@example.com"},
		{"email": "c@example.com"},
	}

	if found := findClientByID(clients, "uuid-1"); found == nil {
		t.Fatalf("expected to find client by id")
	}
	if found := findClientByID(clients, "pw"); found == nil {
		t.Fatalf("expected to find client by password")
	}
	if found := findClientByID(clients, "c@example.com"); found == nil {
		t.Fatalf("expected to find client by email")
	}
	if found := findClientByID(clients, "missing"); found != nil {
		t.Fatalf("expected missing client")
	}
}

func TestSplitInboundClientID(t *testing.T) {
	inboundID, clientID, err := splitInboundClientID("10:abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inboundID != 10 || clientID != "abc" {
		t.Fatalf("unexpected result: %d %s", inboundID, clientID)
	}
	if _, _, err := splitInboundClientID("bad"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestGetClientIDFromModel(t *testing.T) {
	t.Run("explicit client_id wins", func(t *testing.T) {
		m := &InboundClientResourceModel{
			ClientID: types.StringValue("my-uuid"),
			Email:    types.StringValue("user@test.com"),
		}
		got := getClientIDFromModel(m, map[string]any{"email": "user@test.com"})
		if got != "my-uuid" {
			t.Fatalf("expected my-uuid, got %q", got)
		}
	})

	t.Run("password fallback", func(t *testing.T) {
		m := &InboundClientResourceModel{
			ClientID: types.StringUnknown(),
			Email:    types.StringValue("user@test.com"),
		}
		got := getClientIDFromModel(m, map[string]any{"password": "trojan-pass", "email": "user@test.com"})
		if got != "trojan-pass" {
			t.Fatalf("expected trojan-pass, got %q", got)
		}
	})

	t.Run("no fallback to email", func(t *testing.T) {
		m := &InboundClientResourceModel{
			ClientID: types.StringUnknown(),
			Email:    types.StringValue("user@test.com"),
		}
		got := getClientIDFromModel(m, map[string]any{"email": "user@test.com"})
		if got != "" {
			t.Fatalf("expected empty string (UUID will be generated), got %q", got)
		}
	})

	t.Run("null client_id returns empty", func(t *testing.T) {
		m := &InboundClientResourceModel{
			ClientID: types.StringNull(),
			Email:    types.StringValue("user@test.com"),
		}
		got := getClientIDFromModel(m, map[string]any{"email": "user@test.com"})
		if got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
	})
}
