package provider

import (
	"encoding/json"
	"testing"
)

func TestEnsureInboundClientIDs(t *testing.T) {
	inbound := &Inbound{
		Settings: `{"clients":[{"email":"a@example.com"}]}`,
	}
	if err := ensureInboundClientIDs(inbound); err != nil {
		t.Fatalf("ensureInboundClientIDs error: %v", err)
	}
	var settings map[string]any
	if err := json.Unmarshal([]byte(inbound.Settings), &settings); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	clients := settings["clients"].([]any)
	client := clients[0].(map[string]any)
	id, _ := client["id"].(string)
	if id == "" {
		t.Fatalf("expected generated id")
	}
}
