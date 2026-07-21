package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestRealityCredentialAttributesSensitive(t *testing.T) {
	stream := inboundStreamSettingsBlockSchema()
	reality, ok := stream.Blocks["reality_settings"].(schema.SingleNestedBlock)
	if !ok {
		t.Fatal("reality_settings must be a SingleNestedBlock")
	}

	for _, name := range []string{"private_key", "short_ids", "mldsa65_seed"} {
		if !reality.Attributes[name].IsSensitive() {
			t.Errorf("reality_settings.%s must be Sensitive", name)
		}
	}
}

func TestOutboundUUIDAttributesSensitive(t *testing.T) {
	outbound, ok := xrayOutboundsSchema().Blocks["outbound"].(schema.ListNestedBlock)
	if !ok {
		t.Fatal("outbound must be a ListNestedBlock")
	}

	for _, blockName := range []string{"vless_settings", "vmess_settings"} {
		settings, ok := outbound.NestedObject.Blocks[blockName].(schema.ListNestedBlock)
		if !ok {
			t.Fatalf("%s must be a ListNestedBlock", blockName)
		}
		if !settings.NestedObject.Attributes["id"].IsSensitive() {
			t.Errorf("%s.id must be Sensitive", blockName)
		}
	}
}
