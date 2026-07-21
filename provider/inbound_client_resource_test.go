package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestInboundClientResource_Schema exercises the full schema definition
// (incl. the v3.5.0 secret/ad_tag attributes) so the schema declaration lines
// count toward Codecov patch coverage.
func TestInboundClientResource_Schema(t *testing.T) {
	r := NewInboundClientResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema errors: %v", resp.Diagnostics)
	}
	for _, attr := range []string{"secret", "ad_tag", "email"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Fatalf("attribute %q missing from inbound_client schema", attr)
		}
	}
	for _, name := range []string{"id", "client_id", "secret"} {
		if !resp.Schema.Attributes[name].IsSensitive() {
			t.Errorf("%s attribute must be Sensitive", name)
		}
	}
}

func TestFindClientByID(t *testing.T) {
	clients := []map[string]any{
		{"id": "uuid-1", "email": "a@example.com"},
		{"password": "pw", "email": "b@example.com"},
		{"auth": "hysteria-secret", "email": "d@example.com"},
		{"email": "c@example.com"},
	}

	if found := findClientByID(clients, "uuid-1"); found == nil {
		t.Fatalf("expected to find client by id")
	}
	if found := findClientByID(clients, "pw"); found == nil {
		t.Fatalf("expected to find client by password")
	}
	if found := findClientByID(clients, "hysteria-secret"); found == nil {
		t.Fatalf("expected to find client by auth")
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

	t.Run("auth fallback", func(t *testing.T) {
		m := &InboundClientResourceModel{
			ClientID: types.StringUnknown(),
			Email:    types.StringValue("user@test.com"),
		}
		got := getClientIDFromModel(m, map[string]any{"auth": "hysteria-secret", "email": "user@test.com"})
		if got != "hysteria-secret" {
			t.Fatalf("expected hysteria-secret, got %q", got)
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

func TestInboundClientReverseTagExpandFlatten(t *testing.T) {
	model := &InboundClientResourceModel{
		InboundID:  types.Int64Value(1),
		ClientID:   types.StringValue("uuid"),
		Email:      types.StringValue("user@test.com"),
		ReverseTag: types.StringValue("reverse-a"),
	}

	expanded := expandInboundClientFromModel(model)
	reverse, ok := expanded["reverse"].(map[string]any)
	if !ok {
		t.Fatalf("expected reverse map, got %T", expanded["reverse"])
	}
	if reverse["tag"] != "reverse-a" {
		t.Fatalf("unexpected reverse tag: %v", reverse["tag"])
	}

	flattened := inboundClientToModel(1, "uuid", expanded)
	if flattened.ReverseTag.ValueString() != "reverse-a" {
		t.Fatalf("expected reverse-a, got %q", flattened.ReverseTag.ValueString())
	}
}

func TestInboundClientGroupExpandFlatten(t *testing.T) {
	t.Run("group set", func(t *testing.T) {
		model := &InboundClientResourceModel{
			InboundID: types.Int64Value(1),
			ClientID:  types.StringValue("uuid"),
			Email:     types.StringValue("user@test.com"),
			Group:     types.StringValue("premium"),
		}
		expanded := expandInboundClientFromModel(model)
		if expanded["group"] != "premium" {
			t.Fatalf("expected group=premium, got %v", expanded["group"])
		}
		flattened := inboundClientToModel(1, "uuid", expanded)
		if flattened.Group.ValueString() != "premium" {
			t.Fatalf("expected premium, got %q", flattened.Group)
		}
	})

	t.Run("group null", func(t *testing.T) {
		model := &InboundClientResourceModel{
			InboundID: types.Int64Value(1),
			ClientID:  types.StringValue("uuid"),
			Email:     types.StringValue("user@test.com"),
			Group:     types.StringNull(),
		}
		expanded := expandInboundClientFromModel(model)
		if _, ok := expanded["group"]; ok {
			t.Fatalf("expected no group key, got %v", expanded["group"])
		}
	})

	t.Run("group empty from API", func(t *testing.T) {
		client := map[string]any{
			"id": "uuid", "email": "user@test.com", "group": "",
		}
		flattened := inboundClientToModel(1, "uuid", client)
		if !flattened.Group.IsNull() {
			t.Fatalf("expected null for empty group, got %q", flattened.Group)
		}
	})
}

func TestInboundClientMtprotoExpandFlatten(t *testing.T) {
	// v3.5.0 mtg-multi: MTProto clients carry a per-client FakeTLS secret and
	// an optional advertising tag. Both are Optional on the model and only
	// written to the wire map when set, mirroring password/group handling.
	t.Run("secret and ad_tag set", func(t *testing.T) {
		model := &InboundClientResourceModel{
			InboundID: types.Int64Value(1),
			ClientID:  types.StringValue("uuid"),
			Email:     types.StringValue("mtproto@test.com"),
			Secret:    types.StringValue("ee1234567890abcdef1234567890abcd7777772e636c6f7564666c6172652e636f6d"),
			AdTag:     types.StringValue("0123456789abcdef0123456789abcdef"),
		}
		expanded := expandInboundClientFromModel(model)
		if expanded["secret"] != model.Secret.ValueString() {
			t.Fatalf("expected secret on wire, got %v", expanded["secret"])
		}
		// ad_tag tfsdk maps to upstream camelCase "adTag" key.
		if expanded["adTag"] != model.AdTag.ValueString() {
			t.Fatalf("expected adTag on wire, got %v", expanded["adTag"])
		}
		flattened := inboundClientToModel(1, "uuid", expanded)
		if flattened.Secret.ValueString() != model.Secret.ValueString() {
			t.Fatalf("secret round-trip failed: got %q", flattened.Secret)
		}
		if flattened.AdTag.ValueString() != model.AdTag.ValueString() {
			t.Fatalf("ad_tag round-trip failed: got %q", flattened.AdTag)
		}
	})

	t.Run("secret and ad_tag null are omitted", func(t *testing.T) {
		model := &InboundClientResourceModel{
			InboundID: types.Int64Value(1),
			ClientID:  types.StringValue("uuid"),
			Email:     types.StringValue("vless@test.com"),
			Secret:    types.StringNull(),
			AdTag:     types.StringNull(),
		}
		expanded := expandInboundClientFromModel(model)
		if _, ok := expanded["secret"]; ok {
			t.Fatalf("non-MTProto client must not carry a secret key")
		}
		if _, ok := expanded["adTag"]; ok {
			t.Fatalf("non-MTProto client must not carry an adTag key")
		}
	})

	t.Run("empty from API flattens to null", func(t *testing.T) {
		client := map[string]any{
			"id": "uuid", "email": "mtproto@test.com", "secret": "", "adTag": "",
		}
		flattened := inboundClientToModel(1, "uuid", client)
		if !flattened.Secret.IsNull() {
			t.Fatalf("expected null secret, got %q", flattened.Secret)
		}
		if !flattened.AdTag.IsNull() {
			t.Fatalf("expected null ad_tag, got %q", flattened.AdTag)
		}
	})
}
