package provider

import "testing"

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
