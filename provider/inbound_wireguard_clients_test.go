package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestExpandWireguardClientsFromModel verifies the typed -> camelCase JSON
// serialisation for the WireGuard multi-client array (3x-ui v3.4.2+). camelCase
// keys are mandatory here, unlike the legacy peer array (snake_case).
func TestExpandWireguardClientsFromModel(t *testing.T) {
	t.Run("full client round-trips to camelCase keys", func(t *testing.T) {
		m := []InboundWireguardClientModel{{
			PrivateKey:   types.StringValue("privA"),
			PublicKey:    types.StringValue("pubA"),
			PreSharedKey: types.StringValue("pskA"),
			AllowedIPs:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("10.0.0.2/32")}),
			KeepAlive:    types.Int64Value(25),
			Email:        types.StringValue("user@example.com"),
			LimitIP:      types.Int64Value(3),
			TotalGB:      types.Int64Value(100),
			ExpiryTime:   types.Int64Value(0),
			Enable:       types.BoolValue(true),
			TgID:         types.Int64Value(0),
			SubID:        types.StringValue("sub1"),
			Comment:      types.StringValue("c1"),
			Reset:        types.Int64Value(0),
		}}
		got := expandWireguardClientsFromModel(m)
		if len(got) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(got))
		}
		entry, ok := got[0].(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", got[0])
		}
		// camelCase keys — must NOT be snake_case (legacy peer style).
		for _, k := range []string{"privateKey", "publicKey", "preSharedKey", "allowedIPs", "keepAlive", "email", "limitIp", "totalGB", "expiryTime", "enable", "subId", "comment"} {
			if _, has := entry[k]; !has {
				t.Errorf("expected camelCase key %q in expanded entry, got: %v", k, entry)
			}
		}
		// snake_case keys must NOT leak through.
		for _, k := range []string{"private_key", "public_key", "pre_shared_key", "allowed_ips", "keep_alive"} {
			if _, has := entry[k]; has {
				t.Errorf("legacy snake_case key %q must not appear in WG client (camelCase contract): %v", k, entry)
			}
		}
		if entry["privateKey"] != "privA" {
			t.Errorf("expected privateKey=privA, got %v", entry["privateKey"])
		}
		if entry["limitIp"] != 3 {
			t.Errorf("expected limitIp=3, got %v", entry["limitIp"])
		}
	})

	t.Run("null/unknown fields are omitted, zero-value entries dropped", func(t *testing.T) {
		m := []InboundWireguardClientModel{{
			// All null/unknown — should produce no entry.
		}}
		got := expandWireguardClientsFromModel(m)
		if len(got) != 0 {
			t.Fatalf("expected 0 entries for all-null client, got %d: %v", len(got), got)
		}
	})
}

// TestFlattenInboundWireguardClientsToModel verifies camelCase JSON -> typed
// parsing, including the backward-compat path where old panels (≤ v3.4.1) never
// carry the clients key.
func TestFlattenInboundWireguardClientsToModel(t *testing.T) {
	t.Run("parses camelCase upstream payload", func(t *testing.T) {
		list := []any{
			map[string]any{
				"privateKey":   "privA",
				"publicKey":    "pubA",
				"preSharedKey": "pskA",
				"allowedIPs":   []any{"10.0.0.2/32"},
				"keepAlive":    25,
				"email":        "user@example.com",
				"limitIp":      3,
				"totalGB":      100,
				"expiryTime":   0,
				"enable":       true,
				"tgId":         0,
				"subId":        "sub1",
				"comment":      "c1",
				"reset":        0,
			},
		}
		got := flattenInboundWireguardClientsToModel(list)
		if len(got) != 1 {
			t.Fatalf("expected 1 client, got %d", len(got))
		}
		c := got[0]
		if c.PrivateKey.ValueString() != "privA" {
			t.Errorf("expected privateKey=privA, got %q", c.PrivateKey)
		}
		if c.PublicKey.ValueString() != "pubA" {
			t.Errorf("expected publicKey=pubA, got %q", c.PublicKey)
		}
		if c.Email.ValueString() != "user@example.com" {
			t.Errorf("expected email=user@example.com, got %q", c.Email)
		}
		if c.LimitIP.ValueInt64() != 3 {
			t.Errorf("expected limitIP=3, got %v", c.LimitIP)
		}
		if !c.Enable.ValueBool() {
			t.Errorf("expected enable=true, got %v", c.Enable)
		}
	})

	t.Run("empty/nil array yields empty slice (old panels)", func(t *testing.T) {
		// Old panels (v3.4.1 and earlier) never carry wireguard_settings.clients.
		// flattenWireguardInboundSettings guards on the key being present and
		// non-empty before calling this, so an empty input here must not panic.
		if got := flattenInboundWireguardClientsToModel(nil); len(got) != 0 {
			t.Fatalf("expected empty slice for nil input, got %v", got)
		}
		if got := flattenInboundWireguardClientsToModel([]any{}); len(got) != 0 {
			t.Fatalf("expected empty slice for empty input, got %v", got)
		}
	})
}

// TestWireguardSettingsClientsRoundTrip exercises the full WG inbound settings
// expand/flatten path to confirm the clients array survives a write-read cycle
// and does not interfere with the legacy peer array.
func TestWireguardSettingsClientsRoundTrip(t *testing.T) {
	model := &InboundWireguardSettingsModel{
		SecretKey: types.StringValue("serverSecret"),
		Peer: []InboundWireguardPeerModel{{
			PublicKey: types.StringValue("legacyPeerPub"),
		}},
		Clients: []InboundWireguardClientModel{{
			PublicKey: types.StringValue("clientPub"),
			Email:     types.StringValue("c@x.test"),
		}},
	}
	expanded := expandWireguardInboundSettings(model)
	if _, ok := expanded["clients"]; !ok {
		t.Fatalf("expected clients key in expanded settings: %v", expanded)
	}
	if _, ok := expanded["peers"]; !ok {
		t.Fatalf("expected peers key in expanded settings: %v", expanded)
	}
	// Round-trip back through flatten.
	flat := flattenWireguardInboundSettings(expanded)
	if flat == nil {
		t.Fatal("expected non-nil flattened settings")
		return
	}
	if len(flat.Clients) != 1 {
		t.Fatalf("expected 1 client after round-trip, got %d", len(flat.Clients))
	}
	if flat.Clients[0].PublicKey.ValueString() != "clientPub" {
		t.Errorf("expected clientPub after round-trip, got %q", flat.Clients[0].PublicKey)
	}
	if len(flat.Peer) != 1 {
		t.Fatalf("expected 1 peer after round-trip, got %d", len(flat.Peer))
	}
}

// TestFlattenWireguardSettings_OldPanelNoClientsKey verifies backward
// compatibility: a v3.4.1-and-earlier WG inbound has no clients key, so flatten
// must leave the Clients field as a nil slice (no drift, no panic).
func TestFlattenWireguardSettings_OldPanelNoClientsKey(t *testing.T) {
	oldPanelData := map[string]any{
		"secret_key": "serverSecret",
		"peers": []any{
			map[string]any{"public_key": "legacyPeerPub"},
		},
		// NOTE: no "clients" key — this is the pre-v3.4.2 shape.
	}
	flat := flattenWireguardInboundSettings(oldPanelData)
	if flat == nil {
		t.Fatal("expected non-nil flattened settings")
		return
	}
	if len(flat.Clients) != 0 {
		t.Fatalf("expected 0 clients on old panel, got %d: %v", len(flat.Clients), flat.Clients)
	}
	if len(flat.Peer) != 1 {
		t.Fatalf("expected legacy peer preserved, got %d", len(flat.Peer))
	}
}
