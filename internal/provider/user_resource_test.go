package provider

import (
	"context"
	"testing"

	"github.com/batonogov/terraform-provider-3x-ui/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestUserToClientValidation(t *testing.T) {
	model := userModel{
		ClientID:   types.StringValue("uuid"),
		Email:      types.StringValue("user@example.com"),
		Password:   types.StringValue("secret"),
		LimitIP:    types.Int64Value(2),
		TotalGB:    types.Int64Value(10),
		ExpiryTime: types.Int64Value(1234567890),
		Enable:     types.BoolValue(true),
		Comment:    types.StringValue("demo"),
		SubID:      types.StringValue("sub"),
		Reset:      types.Int64Value(1),
		TgID:       types.Int64Value(1234),
	}

	payload, err := model.toClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.Email != "user@example.com" || payload.Password != "secret" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if !payload.Enable {
		t.Fatalf("expected enable true")
	}
}

func TestUserBuildState(t *testing.T) {
	api := newStubClient()
	resource := &userResource{api: api}
	ctx := context.Background()

	state, err := resource.buildState(ctx, 1, "vless", "uuid")
	if err != nil {
		t.Fatalf("build state failed: %v", err)
	}

	if state.Email.ValueString() != "user@example.com" {
		t.Fatalf("unexpected email: %s", state.Email.ValueString())
	}
	if state.Protocol.ValueString() != "vless" {
		t.Fatalf("unexpected protocol: %s", state.Protocol.ValueString())
	}
}

func TestDeriveClientIdentifier(t *testing.T) {
	client := client.InboundClient{
		ID:       "uuid",
		Password: "trojan-pass",
		Email:    "user@example.com",
	}

	if got := deriveClientIdentifier("trojan", client); got != "trojan-pass" {
		t.Fatalf("expected trojan password identifier")
	}
	if got := deriveClientIdentifier("shadowsocks", client); got != "user@example.com" {
		t.Fatalf("expected email identifier")
	}
	if got := deriveClientIdentifier("vless", client); got != "uuid" {
		t.Fatalf("expected id identifier")
	}
}
