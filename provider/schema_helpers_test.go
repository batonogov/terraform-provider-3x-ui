package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func listNestedBlock(t *testing.T, blocks map[string]schema.Block, name string) schema.ListNestedBlock {
	t.Helper()
	block, ok := blocks[name].(schema.ListNestedBlock)
	if !ok {
		t.Fatalf("%s must be a ListNestedBlock", name)
	}
	return block
}

func validateBlockSize(block schema.ListNestedBlock, size int) bool {
	elements := make([]attr.Value, size)
	for i := range elements {
		elements[i] = types.StringValue("element")
	}
	req := validator.ListRequest{
		Path:           path.Root("test"),
		PathExpression: path.MatchRoot("test"),
		ConfigValue:    types.ListValueMust(types.StringType, elements),
	}
	resp := validator.ListResponse{}
	for _, v := range block.Validators {
		v.ValidateList(context.Background(), req, &resp)
	}
	return resp.Diagnostics.HasError()
}

func assertSingletonBlock(t *testing.T, block schema.ListNestedBlock) {
	t.Helper()
	if validateBlockSize(block, 1) {
		t.Fatal("one block must be accepted")
	}
	if !validateBlockSize(block, 2) {
		t.Fatal("two blocks must be rejected")
	}
}

func assertRepeatableBlock(t *testing.T, block schema.ListNestedBlock) {
	t.Helper()
	if validateBlockSize(block, 2) {
		t.Fatal("repeatable block must accept two elements")
	}
}

func TestXrayBasicsSingletonBlocks(t *testing.T) {
	s := xrayBasicsSchema()
	for _, name := range []string{"log", "policy", "api", "stats", "metrics"} {
		t.Run(name, func(t *testing.T) {
			assertSingletonBlock(t, listNestedBlock(t, s.Blocks, name))
		})
	}

	policy := listNestedBlock(t, s.Blocks, "policy")
	assertSingletonBlock(t, listNestedBlock(t, policy.NestedObject.Blocks, "system"))
	assertRepeatableBlock(t, listNestedBlock(t, policy.NestedObject.Blocks, "level"))
	assertRepeatableBlock(t, listNestedBlock(t, s.Blocks, "env"))
}

func TestXrayOutboundsSingletonBlocks(t *testing.T) {
	s := xrayOutboundsSchema()
	outbound := listNestedBlock(t, s.Blocks, "outbound")
	assertRepeatableBlock(t, outbound)

	for _, name := range []string{
		"mux", "freedom_settings", "blackhole_settings", "dns_settings",
		"vmess_settings", "vless_settings", "trojan_settings",
		"shadowsocks_settings", "socks_settings", "http_settings",
		"wireguard_settings", "hysteria_settings",
	} {
		t.Run(name, func(t *testing.T) {
			assertSingletonBlock(t, listNestedBlock(t, outbound.NestedObject.Blocks, name))
		})
	}

	freedom := listNestedBlock(t, outbound.NestedObject.Blocks, "freedom_settings")
	assertSingletonBlock(t, listNestedBlock(t, freedom.NestedObject.Blocks, "fragment"))
	assertRepeatableBlock(t, listNestedBlock(t, freedom.NestedObject.Blocks, "noises"))
	assertRepeatableBlock(t, listNestedBlock(t, freedom.NestedObject.Blocks, "final_rule"))

	wireguard := listNestedBlock(t, outbound.NestedObject.Blocks, "wireguard_settings")
	assertRepeatableBlock(t, listNestedBlock(t, wireguard.NestedObject.Blocks, "peer"))
}

func TestXrayBalancerSingletonBlocks(t *testing.T) {
	s := xrayBalancersSchema()
	balancer := listNestedBlock(t, s.Blocks, "balancer")
	assertRepeatableBlock(t, balancer)
	strategy := listNestedBlock(t, balancer.NestedObject.Blocks, "strategy")
	assertSingletonBlock(t, strategy)
	settings := listNestedBlock(t, strategy.NestedObject.Blocks, "settings")
	assertSingletonBlock(t, settings)
	assertRepeatableBlock(t, listNestedBlock(t, settings.NestedObject.Blocks, "costs"))
}

func TestBurstObservatoryPingConfigSingletonBlock(t *testing.T) {
	s := xrayObservatorySchema()
	burst := listNestedBlock(t, s.Blocks, "burst_observatory")
	assertRepeatableBlock(t, burst)
	assertSingletonBlock(t, listNestedBlock(t, burst.NestedObject.Blocks, "ping_config"))
}
