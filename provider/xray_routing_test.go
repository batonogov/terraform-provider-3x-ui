package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// TestIsManagedDnsAllowRule mirrors 3x-ui v3.5.0's service.dnsAllowRuleShape.
// The panel auto-inserts these direct allow-rules for private DNS servers
// before the geoip:private block on every xray-template save; the provider
// must filter them out of Read to avoid permanent drift.
func TestIsManagedDnsAllowRule(t *testing.T) {
	cases := []struct {
		name string
		rule map[string]any
		want bool
	}{
		{
			name: "canonical managed rule",
			rule: map[string]any{
				"type": "field", "outboundTag": "direct",
				"ip": []any{"10.0.0.53"}, "port": "53",
			},
			want: true,
		},
		{
			name: "managed rule with enabled=true tolerated",
			rule: map[string]any{
				"type": "field", "outboundTag": "direct",
				"ip": []any{"10.0.0.53"}, "port": "53",
				"enabled": true,
			},
			want: true,
		},
		{
			name: "snake_case outbound_tag also recognised",
			rule: map[string]any{
				"type": "field", "outbound_tag": "direct",
				"ip": []any{"10.0.0.53"}, "port": "53",
			},
			want: true,
		},
		{
			name: "enabled=false is no longer managed",
			rule: map[string]any{
				"type": "field", "outboundTag": "direct",
				"ip": []any{"10.0.0.53"}, "port": "53",
				"enabled": false,
			},
			want: false,
		},
		{
			name: "extra matcher (domain) is NOT managed",
			rule: map[string]any{
				"type": "field", "outboundTag": "direct",
				"ip": []any{"10.0.0.53"}, "port": "53",
				"domain": []any{"example.com"},
			},
			want: false,
		},
		{
			name: "missing ip is not managed",
			rule: map[string]any{
				"type": "field", "outboundTag": "direct", "port": "53",
			},
			want: false,
		},
		{
			name: "missing port is not managed",
			rule: map[string]any{
				"type": "field", "outboundTag": "direct", "ip": []any{"10.0.0.53"},
			},
			want: false,
		},
		{
			name: "outboundTag not direct is not managed",
			rule: map[string]any{
				"type": "field", "outboundTag": "blocked",
				"ip": []any{"geoip:private"}, "port": "53",
			},
			want: false,
		},
		{
			name: "type not field is not managed",
			rule: map[string]any{
				"type": "default", "outboundTag": "direct",
				"ip": []any{"10.0.0.53"}, "port": "53",
			},
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isManagedDnsAllowRule(tc.rule); got != tc.want {
				t.Fatalf("isManagedDnsAllowRule(%v) = %v, want %v", tc.rule, got, tc.want)
			}
		})
	}
}

// TestFlattenRoutingRulesFiltersManagedDns confirms the wire→untyped flatten
// path drops both API routing rules and managed DNS allow-rules so they never
// surface as drift in threexui_xray_routing's rules block.
func TestFlattenRoutingRulesFiltersManagedDns(t *testing.T) {
	list := []any{
		// user rule — must survive
		map[string]any{
			"type": "field", "outboundTag": "blocked",
			"ip": []any{"geoip:private"},
		},
		// managed DNS allow-rule — must be dropped
		map[string]any{
			"type": "field", "outboundTag": "direct",
			"ip": []any{"10.0.0.53"}, "port": "53",
		},
		// API routing rule — must be dropped
		map[string]any{
			"type": "field", "inboundTag": []string{"api"}, "outboundTag": "api",
		},
	}
	out := flattenRoutingRules(list)
	if len(out) != 1 {
		t.Fatalf("expected 1 surviving rule (user rule only), got %d: %#v", len(out), out)
	}
	survivor, _ := out[0].(map[string]any)
	if survivor["outbound_tag"] != "blocked" {
		t.Fatalf("expected the user geoip:private block rule to survive, got %v", survivor["outbound_tag"])
	}
}

// TestExpandRoutingRulesFiltersManagedDns mirrors the flatten-path test above
// for the write/expand path. expandRoutingRules reads snake_case keys (model
// representation) and is the other call site that filters managed rules;
// without this test the expand-path `|| isManagedDnsAllowRule(m)` clause would
// be uncovered. It also exercises the snake_case `outbound_tag` recognition in
// isManagedDnsAllowRule (flatten-path rules arrive in camelCase, so the snake
// branch is only reached here).
func TestExpandRoutingRulesFiltersManagedDns(t *testing.T) {
	list := []any{
		// user rule (snake_case input) — must survive; expandRoutingRules
		// translates outbound_tag→outboundTag (snake→camel) on output.
		map[string]any{
			"type": "field", "outbound_tag": "blocked",
			"ip": []any{"geoip:private"},
		},
		// managed DNS allow-rule (snake_case) — must be dropped
		map[string]any{
			"type": "field", "outbound_tag": "direct",
			"ip": []any{"10.0.0.53"}, "port": "53",
		},
	}
	out := expandRoutingRules(list)
	if len(out) != 1 {
		t.Fatalf("expected 1 surviving rule, got %d: %#v", len(out), out)
	}
	survivor, _ := out[0].(map[string]any)
	if survivor["outboundTag"] != "blocked" {
		t.Fatalf("expected the user block rule to survive, got %v", survivor["outboundTag"])
	}
}

// ---------------------------------------------------------------------------
// ModifyPlan: prevent stale field carry-forward when rules are reordered/removed
// ---------------------------------------------------------------------------
//
// The nested `rule` block attributes are Optional+Computed with
// UseStateForUnknown (added by #228 to silence known-after-apply drift).
// ListNestedBlock elements are matched by index, so that modifier copies the
// prior rule's unset fields into the new rule occupying the same index,
// bleeding stale matchers across rules on reorder/removal. reconcileRoutingPlan
// (called from XrayRoutingResource.ModifyPlan) makes the configuration
// authoritative so the carried values are dropped.

func strList(vs ...string) types.List {
	if len(vs) == 0 {
		return types.ListNull(types.StringType)
	}
	out := make([]attr.Value, len(vs))
	for i, v := range vs {
		out[i] = types.StringValue(v)
	}
	return types.ListValueMust(types.StringType, out)
}

func TestReconcileRoutingPlan_StripsStaleCarryForward(t *testing.T) {
	ruDomains := strList("geosite:category-ru", "ru")
	geoipPrivate := strList("geoip:private")
	geoipCN := strList("geoip:cn")

	config := XrayRoutingModel{
		ID:             types.StringValue("xray_routing"),
		DomainStrategy: types.StringValue("AsIs"),
		Rule: []XrayRoutingRule{
			{Type: types.StringValue("field"), OutboundTag: types.StringValue("direct"), Domain: ruDomains},
			{Type: types.StringValue("field"), OutboundTag: types.StringValue("direct"), IP: geoipCN},
		},
	}
	// Plan after schema plan modifiers: each new rule inherited stale fields
	// from the prior rule at the same index. The provider must not write these.
	plan := XrayRoutingModel{
		ID:             types.StringValue("xray_routing"),
		DomainStrategy: types.StringValue("AsIs"),
		Rule: []XrayRoutingRule{
			{Type: types.StringValue("field"), OutboundTag: types.StringValue("direct"), Domain: ruDomains, IP: geoipPrivate},
			{Type: types.StringValue("field"), OutboundTag: types.StringValue("direct"), IP: geoipCN, Network: types.StringValue("tcp,udp")},
		},
	}

	changed := reconcileRoutingPlan(&plan, config)
	if !changed {
		t.Fatal("expected reconcile to report a change when the plan carried stale fields")
	}
	if !reflect.DeepEqual(plan.Rule, config.Rule) {
		t.Fatalf("plan rules must mirror config after reconcile\ngot:  %#v\nwant: %#v", plan.Rule, config.Rule)
	}
}

func TestReconcileRoutingPlan_NoOpWhenUnchanged(t *testing.T) {
	rule := []XrayRoutingRule{
		{Type: types.StringValue("field"), OutboundTag: types.StringValue("direct"), IP: strList("geoip:private")},
	}
	config := XrayRoutingModel{Rule: rule}
	plan := XrayRoutingModel{Rule: rule}

	changed := reconcileRoutingPlan(&plan, config)
	if changed {
		t.Fatal("expected no change when the plan already mirrors config")
	}
	if !reflect.DeepEqual(plan.Rule, rule) {
		t.Fatal("plan rules must be untouched when reconcile is a no-op")
	}
}

// rtRuleValue builds a single `rule` block element as a tftypes.Value, setting
// only the supplied matchers and leaving every other attribute null.
func rtRuleValue(t *testing.T, ruleObjType tftypes.Object, outboundTag, network string, domain, ip []string) tftypes.Value {
	t.Helper()
	vals := map[string]tftypes.Value{}
	for k, ty := range ruleObjType.AttributeTypes {
		vals[k] = tftypes.NewValue(ty, nil)
	}
	vals["type"] = tftypes.NewValue(tftypes.String, "field")
	if outboundTag != "" {
		vals["outbound_tag"] = tftypes.NewValue(tftypes.String, outboundTag)
	}
	if network != "" {
		vals["network"] = tftypes.NewValue(tftypes.String, network)
	}
	if len(domain) > 0 {
		elems := make([]tftypes.Value, len(domain))
		for i, d := range domain {
			elems[i] = tftypes.NewValue(tftypes.String, d)
		}
		vals["domain"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, elems)
	}
	if len(ip) > 0 {
		elems := make([]tftypes.Value, len(ip))
		for i, x := range ip {
			elems[i] = tftypes.NewValue(tftypes.String, x)
		}
		vals["ip"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, elems)
	}
	return tftypes.NewValue(ruleObjType, vals)
}

// routingRaw builds the top-level resource object tftypes.Value with the given
// `rule` list and defaults for the other (string) attributes.
func routingRaw(t *testing.T, schemaResp resource.SchemaResponse, ruleListVal tftypes.Value) tftypes.Value {
	t.Helper()
	ctx := context.Background()
	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, ty := range objType.AttributeTypes {
		switch k {
		case "id":
			vals[k] = tftypes.NewValue(tftypes.String, "xray_routing")
		case "domain_strategy":
			vals[k] = tftypes.NewValue(tftypes.String, "AsIs")
		case "rule":
			vals[k] = ruleListVal
		default:
			vals[k] = tftypes.NewValue(ty, nil)
		}
	}
	return tftypes.NewValue(objType, vals)
}

// TestXrayRoutingResource_ModifyPlan_StripsStaleCarryForward drives the full
// ModifyPlan path (decode plan/config → reconcile → set plan) to ensure the
// plugin-framework plan modifiers' carried-forward values are erased before
// the resource writes them to the panel.
func TestXrayRoutingResource_ModifyPlan_StripsStaleCarryForward(t *testing.T) {
	r := &XrayRoutingResource{}
	ctx := context.Background()
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	ruleListType := objType.AttributeTypes["rule"].(tftypes.List)
	ruleObjType := ruleListType.ElementType.(tftypes.Object)

	// Config: clean 2-rule list — RU-domains→direct, geoip:cn→direct.
	cfgRules := tftypes.NewValue(ruleListType, []tftypes.Value{
		rtRuleValue(t, ruleObjType, "direct", "", []string{"geosite:category-ru", "ru"}, nil),
		rtRuleValue(t, ruleObjType, "direct", "", nil, []string{"geoip:cn"}),
	})
	// Plan: the schema plan modifiers carried the prior state's per-index
	// fields into the new rules — rule 0 gained stale ip:geoip:private and
	// rule 1 gained stale network:"tcp,udp".
	planRules := tftypes.NewValue(ruleListType, []tftypes.Value{
		rtRuleValue(t, ruleObjType, "direct", "", []string{"geosite:category-ru", "ru"}, []string{"geoip:private"}),
		rtRuleValue(t, ruleObjType, "direct", "tcp,udp", nil, []string{"geoip:cn"}),
	})
	// State: the prior 3-rule list (only needs to be non-null for ModifyPlan's guard).
	stateRules := tftypes.NewValue(ruleListType, []tftypes.Value{
		rtRuleValue(t, ruleObjType, "direct", "", nil, []string{"geoip:private"}),
		rtRuleValue(t, ruleObjType, "direct", "", []string{"geosite:category-ru", "ru"}, nil),
		rtRuleValue(t, ruleObjType, "proxy", "tcp,udp", nil, nil),
	})

	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: routingRaw(t, schemaResp, planRules)}
	config := tfsdk.Config{Schema: schemaResp.Schema, Raw: routingRaw(t, schemaResp, cfgRules)}
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: routingRaw(t, schemaResp, stateRules)}

	resp := &resource.ModifyPlanResponse{Plan: tfsdk.Plan{Schema: schemaResp.Schema, Raw: tftypes.NewValue(objType, nil)}}
	r.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan, Config: config, State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}

	var out XrayRoutingModel
	resp.Plan.Get(ctx, &out)
	if len(out.Rule) != 2 {
		t.Fatalf("expected 2 rules in reconciled plan, got %d", len(out.Rule))
	}
	// rule 0 must keep its configured domain and must NOT carry the stale ip.
	if !out.Rule[0].IP.IsNull() {
		t.Fatalf("rule 0 must not inherit stale ip:geoip:private, got %v", out.Rule[0].IP)
	}
	if out.Rule[0].Domain.IsNull() {
		t.Fatal("rule 0 must keep its configured domain")
	}
	// rule 1 must keep its configured ip and must NOT carry the stale network.
	if !out.Rule[1].Network.IsNull() {
		t.Fatalf("rule 1 must not inherit stale network:tcp,udp, got %v", out.Rule[1].Network)
	}
	if out.Rule[1].IP.IsNull() {
		t.Fatal("rule 1 must keep its configured ip:geoip:cn")
	}
}

// TestXrayRoutingResource_ModifyPlan_NoOpOnCreate ensures ModifyPlan leaves the
// plan untouched when there is no prior state (Create) — the guard short-circuits
// before decoding, so the empty plan passed in is returned verbatim.
func TestXrayRoutingResource_ModifyPlan_NoOpOnCreate(t *testing.T) {
	r := &XrayRoutingResource{}
	ctx := context.Background()
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	tt := schemaResp.Schema.Type().TerraformType(ctx)
	emptyPlan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: tftypes.NewValue(tt, nil)}
	emptyConfig := tfsdk.Config{Schema: schemaResp.Schema, Raw: tftypes.NewValue(tt, nil)}
	// null state → Create path; guard returns without touching the plan.
	resp := &resource.ModifyPlanResponse{Plan: emptyPlan}
	r.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: emptyPlan, Config: emptyConfig, State: tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(tt, nil)}}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
	if !resp.Plan.Raw.IsNull() {
		t.Fatalf("ModifyPlan must not rewrite the plan on Create, got %v", resp.Plan.Raw)
	}
}
