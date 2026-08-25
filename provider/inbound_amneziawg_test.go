package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// awgFullServerModel is a server block with every attribute populated, so the
// round-trip tests below fail on a key that is dropped anywhere along
// model → untyped map → JSON → untyped map → model.
func awgFullServerModel() *InboundAmneziawgServerModel {
	return &InboundAmneziawgServerModel{
		PrivateKey:             types.StringValue("privateKeyForTestsOnly000000000000000000000="),
		PublicKey:              types.StringValue("publicKeyForTestsOnly0000000000000000000000="),
		SubnetIP:               types.StringValue("10.8.1.0"),
		SubnetCIDR:             types.Int64Value(24),
		MTU:                    types.Int64Value(1380),
		PrimaryDNS:             types.StringValue("8.8.8.8"),
		SecondaryDNS:           types.StringValue("8.8.4.4"),
		ExternalInterface:      types.StringValue("eth0"),
		IPv6Enabled:            types.BoolValue(true),
		IPv6Subnet:             types.StringValue("fd00:8:1::/64"),
		IPv6ExternalInterface:  types.StringValue("eth1"),
		RouteThroughXray:       types.BoolValue(true),
		Jc:                     types.Int64Value(4),
		Jmin:                   types.Int64Value(56),
		Jmax:                   types.Int64Value(190),
		S1:                     types.Int64Value(88),
		S2:                     types.Int64Value(33),
		S3:                     types.Int64Value(41),
		S4:                     types.Int64Value(19),
		H1:                     types.StringValue("1553264530-1553265530"),
		H2:                     types.StringValue("2553264530-2553265530"),
		H3:                     types.StringValue("3553264530-3553265530"),
		H4:                     types.StringValue("4553264530-4553265530"),
		I1:                     types.StringValue("<r 64>"),
		I2:                     types.StringValue("<r 48>"),
		I3:                     types.StringValue("<r 40>"),
		I4:                     types.StringValue("<r 36>"),
		I5:                     types.StringValue("<r 32>"),
		HeaderProtectionKey:    types.StringValue("MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="),
		ContentPaddingAddition: types.StringValue("12-40"),
		RekeyAfterTime:         types.StringValue("110-140"),
		RekeyTimeout:           types.StringValue("4-7"),
		RejectAfterTime:        types.StringValue("180-220"),
		KeepaliveTimeout:       types.StringValue("9-13"),
		MaxHandshakeAttempts:   types.StringValue("18-24"),
		RandomTrailers:         types.BoolValue(true),
		DisableCookies:         types.BoolValue(false),
	}
}

func awgClientModel() InboundAmneziawgClientModel {
	return InboundAmneziawgClientModel{
		Email:          types.StringValue("peer-one@test.com"),
		PrivateKey:     types.StringValue("clientPrivateKey0000000000000000000000000000="),
		PublicKey:      types.StringValue("clientPublicKey00000000000000000000000000000="),
		PreSharedKey:   types.StringValue("presharedKey000000000000000000000000000000000="),
		AllowedIPs:     anySliceToTypesList([]any{"10.8.1.2/32", "fd00:8:1::2/128"}),
		KeepAlive:      types.Int64Value(25),
		ForwardedPorts: types.StringValue("80,443,8000-8100"),
		Enable:         types.BoolValue(true),
		LimitIP:        types.Int64Value(2),
		TotalGB:        types.Int64Value(107374182400),
		ExpiryTime:     types.Int64Value(1767225600000),
		TgID:           types.Int64Value(12345),
		SubID:          types.StringValue("abcdef0123456789"),
		Comment:        types.StringValue("first peer"),
		Reset:          types.Int64Value(0),
	}
}

// The full path an inbound settings blob travels on apply and read back:
// typed model → expandSettingsFromModel → buildSettingsJSON → flattenSettings →
// flattenSettingsToModel. Every attribute must survive it unchanged.
func TestExpandFlattenAmneziawgSettingsModel_RoundTrip(t *testing.T) {
	original := &InboundResourceModel{
		AmneziawgSettings: &InboundAmneziawgSettingsModel{
			Server:  awgFullServerModel(),
			Clients: []InboundAmneziawgClientModel{awgClientModel()},
		},
	}

	expanded := expandSettingsFromModel("amneziawg", original)
	if expanded == nil {
		t.Fatal("expected non-nil expansion for amneziawg")
	}
	server, ok := expanded["server"].(map[string]any)
	if !ok {
		t.Fatalf("expected a nested server map, got %T", expanded["server"])
	}
	// The expander must emit the upstream camelCase spelling directly — settings.go
	// forwards this object verbatim rather than translating it.
	if server["subnetIp"] != "10.8.1.0" {
		t.Errorf("expected camelCase subnetIp, got %v (keys: %v)", server["subnetIp"], mapKeys(server))
	}
	if server["headerProtectionKey"] == nil {
		t.Error("headerProtectionKey missing from the expanded server block")
	}

	raw := buildSettingsJSON(expanded, "amneziawg")

	var wire map[string]any
	if err := json.Unmarshal([]byte(raw), &wire); err != nil {
		t.Fatalf("buildSettingsJSON produced invalid JSON: %v (%s)", err, raw)
	}
	if _, ok := wire["server"]; !ok {
		t.Fatalf("server block missing from the wire payload: %s", raw)
	}
	if _, ok := wire["clients"]; !ok {
		t.Fatalf("clients array missing from the wire payload: %s", raw)
	}

	flattened, err := flattenSettings(raw, "amneziawg")
	if err != nil {
		t.Fatalf("flattenSettings: %v", err)
	}
	if len(flattened) != 1 {
		t.Fatalf("expected one settings entry, got %d", len(flattened))
	}

	restoredModel := &InboundResourceModel{}
	flattenSettingsToModel("amneziawg", flattened[0].(map[string]any), restoredModel)

	got := restoredModel.AmneziawgSettings
	if got == nil {
		t.Fatal("amneziawg_settings was dropped on the way back")
	}
	if got.Server == nil {
		t.Fatal("server block was dropped on the way back")
	}
	if *got.Server != *original.AmneziawgSettings.Server {
		t.Errorf("server block changed across the round trip:\n before: %+v\n after:  %+v",
			*original.AmneziawgSettings.Server, *got.Server)
	}
	if len(got.Clients) != 1 {
		t.Fatalf("expected one client, got %d", len(got.Clients))
	}
	before, after := original.AmneziawgSettings.Clients[0], got.Clients[0]
	if !before.AllowedIPs.Equal(after.AllowedIPs) {
		t.Errorf("allowed_ips changed: %v -> %v", before.AllowedIPs, after.AllowedIPs)
	}
	if before.ForwardedPorts != after.ForwardedPorts {
		t.Errorf("forwarded_ports changed: %v -> %v", before.ForwardedPorts, after.ForwardedPorts)
	}
	if before.Email != after.Email || before.PrivateKey != after.PrivateKey ||
		before.PublicKey != after.PublicKey || before.PreSharedKey != after.PreSharedKey ||
		before.KeepAlive != after.KeepAlive || before.Enable != after.Enable ||
		before.LimitIP != after.LimitIP || before.TotalGB != after.TotalGB ||
		before.ExpiryTime != after.ExpiryTime || before.TgID != after.TgID ||
		before.SubID != after.SubID || before.Comment != after.Comment ||
		before.Reset != after.Reset {
		t.Errorf("client changed across the round trip:\n before: %+v\n after:  %+v", before, after)
	}
}

// A blank private key makes the panel generate a NEW server keypair, silently
// invalidating every existing client config. Nulls must therefore be omitted
// from the payload rather than sent as "".
func TestExpandAmneziawgServer_OmitsUnsetFields(t *testing.T) {
	server := expandAmneziawgServerFromModel(&InboundAmneziawgServerModel{
		SubnetIP:   types.StringValue("10.9.0.0"),
		PrivateKey: types.StringNull(),
		MTU:        types.Int64Unknown(),
	})

	if _, ok := server["privateKey"]; ok {
		t.Error("a null private_key must not be sent — the panel would rotate the server keypair")
	}
	if _, ok := server["mtu"]; ok {
		t.Error("an unknown mtu must not be sent")
	}
	if server["subnetIp"] != "10.9.0.0" {
		t.Errorf("configured value lost: %v", server["subnetIp"])
	}
}

// Upstream declares ipv6Enabled/routeThroughXray/disableCookies with omitempty,
// so a false value is stripped from the stored blob. Reading a stripped key back
// as null would break a configuration that sets it to false explicitly
// ("inconsistent result after apply"), so absence must read as false.
func TestFlattenAmneziawgServer_StrippedBoolsReadAsFalse(t *testing.T) {
	server := flattenAmneziawgServerToModel(map[string]any{"subnetIp": "10.8.1.0"})
	if server == nil {
		t.Fatal("expected a server model")
	}
	for name, got := range map[string]types.Bool{
		"ipv6_enabled":       server.IPv6Enabled,
		"route_through_xray": server.RouteThroughXray,
		"random_trailers":    server.RandomTrailers,
		"disable_cookies":    server.DisableCookies,
	} {
		if got.IsNull() || got.ValueBool() {
			t.Errorf("%s should read as false when the panel strips the key, got %v", name, got)
		}
	}
	// String and int fields keep null for absent keys — the panel fills them in.
	if !server.MTU.IsNull() {
		t.Errorf("absent mtu should stay null, got %v", server.MTU)
	}
	if !server.HeaderProtectionKey.IsNull() {
		t.Errorf("absent header_protection_key should stay null, got %v", server.HeaderProtectionKey)
	}
}

// primaryDns/secondaryDns and h1-h4 are NOT omitempty upstream: an empty string
// is a meaningful cleared value there, and must survive as "" rather than
// collapsing to null.
func TestFlattenAmneziawgServer_KeepsClearedStrings(t *testing.T) {
	server := flattenAmneziawgServerToModel(map[string]any{
		"primaryDns":   "",
		"secondaryDns": "",
		"h3":           "",
	})
	if server == nil {
		t.Fatal("expected a server model")
	}
	for name, got := range map[string]types.String{
		"primary_dns":   server.PrimaryDNS,
		"secondary_dns": server.SecondaryDNS,
		"h3":            server.H3,
	} {
		if got.IsNull() {
			t.Errorf("%s must read back as an empty string, not null", name)
		}
	}
}

// An AmneziaWG inbound with no settings at all is valid: the panel generates the
// whole server block on save. The expander must not invent one.
func TestExpandAmneziawgSettings_EmptyModel(t *testing.T) {
	if got := expandAmneziawgInboundSettings(nil); got != nil {
		t.Errorf("nil model should expand to nil, got %v", got)
	}
	got := expandAmneziawgInboundSettings(&InboundAmneziawgSettingsModel{})
	if len(got) != 0 {
		t.Errorf("empty model should expand to an empty map, got %v", got)
	}
	if flattenAmneziawgInboundSettings(map[string]any{}) != nil {
		t.Error("empty settings should flatten to nil")
	}
}

// clients[] is protocol-gated in settings.go: forwarding it for a protocol whose
// clients belong to threexui_inbound_client would make the two resources fight.
// AmneziaWG must be on the same side of that gate as WireGuard (#342).
func TestAmneziawgClientsSurviveSettingsJSON(t *testing.T) {
	item := map[string]any{
		"server":  map[string]any{"subnetIp": "10.8.1.0"},
		"clients": []any{map[string]any{"email": "peer@test.com"}},
	}

	raw := buildSettingsJSON(item, "amneziawg")
	var wire map[string]any
	if err := json.Unmarshal([]byte(raw), &wire); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := wire["clients"]; !ok {
		t.Fatalf("clients[] stripped for amneziawg: %s", raw)
	}

	// The same payload under a client-resource-owned protocol must lose clients[].
	rawVless := buildSettingsJSON(item, "vless")
	var wireVless map[string]any
	if err := json.Unmarshal([]byte(rawVless), &wireVless); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := wireVless["clients"]; ok {
		t.Errorf("clients[] must stay stripped for vless: %s", rawVless)
	}
	if _, ok := wireVless["server"]; ok {
		t.Errorf("the amneziawg server block must not leak into other protocols: %s", rawVless)
	}
}

// The nested server block is a schema.SingleNestedBlock, and clients is a list
// block — the shapes docs and plan modifiers depend on.
func TestAmneziawgSettingsBlockShape(t *testing.T) {
	block := amneziawgSettingsBlock()

	server, ok := block.Blocks["server"].(schema.SingleNestedBlock)
	if !ok {
		t.Fatalf("server must be a SingleNestedBlock, got %T", block.Blocks["server"])
	}
	if _, ok := block.Blocks["clients"].(schema.ListNestedBlock); !ok {
		t.Fatalf("clients must be a ListNestedBlock, got %T", block.Blocks["clients"])
	}

	for _, name := range []string{"private_key", "header_protection_key"} {
		if !server.Attributes[name].IsSensitive() {
			t.Errorf("server.%s must be Sensitive", name)
		}
	}

	// Every documented server field must exist in the schema, with the tfsdk
	// model and the schema agreeing on the attribute set.
	want := []string{
		"private_key", "public_key", "subnet_ip", "subnet_cidr", "mtu",
		"primary_dns", "secondary_dns", "external_interface",
		"ipv6_enabled", "ipv6_subnet", "ipv6_external_interface", "route_through_xray",
		"jc", "jmin", "jmax", "s1", "s2", "s3", "s4",
		"h1", "h2", "h3", "h4", "i1", "i2", "i3", "i4", "i5",
		"header_protection_key", "content_padding_addition",
		"rekey_after_time", "rekey_timeout", "reject_after_time",
		"keepalive_timeout", "max_handshake_attempts",
		"random_trailers", "disable_cookies",
	}
	for _, name := range want {
		if _, ok := server.Attributes[name]; !ok {
			t.Errorf("server block is missing attribute %q", name)
		}
	}
	if len(server.Attributes) != len(want) {
		t.Errorf("server block has %d attributes, expected %d — update the model, the docs and this list together",
			len(server.Attributes), len(want))
	}
}

// The validator is the only thing standing between a practitioner and a silent
// server-key rotation, so it has to fire on exactly the right configurations.
func TestAmneziawgServerRequiredValidator(t *testing.T) {
	cases := []struct {
		name      string
		model     InboundResourceModel
		wantError bool
	}{
		{
			name: "amneziawg without a settings block is rejected",
			model: InboundResourceModel{
				Protocol: types.StringValue("amneziawg"),
			},
			wantError: true,
		},
		{
			name: "amneziawg with settings but no server block is rejected",
			model: InboundResourceModel{
				Protocol:          types.StringValue("amneziawg"),
				AmneziawgSettings: &InboundAmneziawgSettingsModel{},
			},
			wantError: true,
		},
		{
			name: "an empty server block is enough — the panel fills it in",
			model: InboundResourceModel{
				Protocol: types.StringValue("amneziawg"),
				AmneziawgSettings: &InboundAmneziawgSettingsModel{
					Server: &InboundAmneziawgServerModel{},
				},
			},
		},
		{
			name: "other protocols are untouched",
			model: InboundResourceModel{
				Protocol: types.StringValue("wireguard"),
			},
		},
		{
			name: "an unknown protocol cannot be checked here",
			model: InboundResourceModel{
				Protocol: types.StringUnknown(),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotError := amneziawgConfigRejected(t, tc.model)
			if gotError != tc.wantError {
				t.Errorf("validator error = %v, want %v", gotError, tc.wantError)
			}
		})
	}
}

// amneziawgConfigRejected runs the validator against a config built from model
// and reports whether it produced an error diagnostic.
func amneziawgConfigRejected(t *testing.T, model InboundResourceModel) bool {
	t.Helper()

	var schemaResp resource.SchemaResponse
	(&InboundResource{}).Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("building schema: %v", schemaResp.Diagnostics)
	}

	ctx := context.Background()
	obj, diags := types.ObjectValueFrom(ctx, schemaResp.Schema.Type().(types.ObjectType).AttrTypes, model)
	if diags.HasError() {
		t.Fatalf("building config object: %v", diags)
	}
	raw, err := obj.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("converting config object: %v", err)
	}

	req := resource.ValidateConfigRequest{
		Config: tfsdk.Config{Raw: raw, Schema: schemaResp.Schema},
	}
	resp := &resource.ValidateConfigResponse{}
	amneziawgServerRequiredValidator{}.ValidateResource(ctx, req, resp)
	return resp.Diagnostics.HasError()
}

// The validator has to be wired into the resource, or it never runs.
func TestInboundResourceRegistersAmneziawgValidator(t *testing.T) {
	validators := (&InboundResource{}).ConfigValidators(context.Background())
	for _, v := range validators {
		if _, ok := v.(amneziawgServerRequiredValidator); ok {
			if v.Description(context.Background()) == "" {
				t.Error("validator must describe itself")
			}
			return
		}
	}
	t.Fatalf("threexui_inbound does not register the amneziawg server validator (got %d validators)", len(validators))
}
