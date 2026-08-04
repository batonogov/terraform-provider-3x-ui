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
	ruDomains := strList("geosite:category-ru")
	geoipPrivate := strList("geoip:private")
	geoipCN := strList("geoip:cn")

	// Config (step 2): [0] RU-domains→blocked, [1] geoip:cn→direct.
	config := XrayRoutingModel{
		ID:             types.StringValue("xray_routing"),
		DomainStrategy: types.StringValue("AsIs"),
		Rule: []XrayRoutingRule{
			{Type: types.StringValue("field"), OutboundTag: types.StringValue("blocked"), Domain: ruDomains},
			{Type: types.StringValue("field"), OutboundTag: types.StringValue("direct"), IP: geoipCN},
		},
	}
	// Plan after schema plan modifiers: each new rule inherited stale fields
	// from the prior rule at the same index. Rule 0 gained stale ip:geoip:private
	// (from prior index-0), rule 1 gained stale domain:geosite:category-ru
	// (from prior index-1). The provider must not write these.
	plan := XrayRoutingModel{
		ID:             types.StringValue("xray_routing"),
		DomainStrategy: types.StringValue("AsIs"),
		Rule: []XrayRoutingRule{
			{Type: types.StringValue("field"), OutboundTag: types.StringValue("blocked"), Domain: ruDomains, IP: geoipPrivate},
			{Type: types.StringValue("field"), OutboundTag: types.StringValue("direct"), IP: geoipCN, Domain: ruDomains},
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
// the resource writes them to the panel. The state/index mapping mirrors the
// acceptance test: step1 [ip:private→direct, domain:category-ru→blocked,
// network:tcp,udp→blocked] → step2 [domain:category-ru→blocked,
// ip:geoip:cn→direct, network:tcp,udp→blocked].
func TestXrayRoutingResource_ModifyPlan_StripsStaleCarryForward(t *testing.T) {
	r := &XrayRoutingResource{}
	ctx := context.Background()
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	ruleListType := objType.AttributeTypes["rule"].(tftypes.List)
	ruleObjType := ruleListType.ElementType.(tftypes.Object)

	// Config (step 2): clean 2-rule list — domain→blocked, ip:geoip:cn→direct.
	cfgRules := tftypes.NewValue(ruleListType, []tftypes.Value{
		rtRuleValue(t, ruleObjType, "blocked", "", []string{"geosite:category-ru"}, nil),
		rtRuleValue(t, ruleObjType, "direct", "", nil, []string{"geoip:cn"}),
	})
	// Plan: the schema plan modifiers carried the prior state's per-index
	// fields into the new rules — rule 0 gained stale ip:geoip:private (from
	// prior index-0) and rule 1 gained stale domain:geosite:category-ru (from
	// prior index-1).
	planRules := tftypes.NewValue(ruleListType, []tftypes.Value{
		rtRuleValue(t, ruleObjType, "blocked", "", []string{"geosite:category-ru"}, []string{"geoip:private"}),
		rtRuleValue(t, ruleObjType, "direct", "", []string{"geosite:category-ru"}, []string{"geoip:cn"}),
	})
	// State: the prior 3-rule list mirroring step 1.
	stateRules := tftypes.NewValue(ruleListType, []tftypes.Value{
		rtRuleValue(t, ruleObjType, "direct", "", nil, []string{"geoip:private"}),
		rtRuleValue(t, ruleObjType, "blocked", "", []string{"geosite:category-ru"}, nil),
		rtRuleValue(t, ruleObjType, "blocked", "tcp,udp", nil, nil),
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
	// rule 1 must keep its configured ip and must NOT carry the stale domain.
	if !out.Rule[1].Domain.IsNull() {
		t.Fatalf("rule 1 must not inherit stale domain:geosite:category-ru, got %v", out.Rule[1].Domain)
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

// TestXrayRoutingResource_ModifyPlan_NoOpWhenPlanMirrorsConfig ensures
// ModifyPlan does not rewrite the plan when reconcileRoutingPlan reports no
// change (plan already mirrors config) — covers the reconcile-false branch.
func TestXrayRoutingResource_ModifyPlan_NoOpWhenPlanMirrorsConfig(t *testing.T) {
	r := &XrayRoutingResource{}
	ctx := context.Background()
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	ruleListType := objType.AttributeTypes["rule"].(tftypes.List)
	ruleObjType := ruleListType.ElementType.(tftypes.Object)

	rules := tftypes.NewValue(ruleListType, []tftypes.Value{
		rtRuleValue(t, ruleObjType, "direct", "", nil, []string{"geoip:private"}),
	})
	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: routingRaw(t, schemaResp, rules)}
	config := tfsdk.Config{Schema: schemaResp.Schema, Raw: routingRaw(t, schemaResp, rules)}
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: routingRaw(t, schemaResp, rules)}

	resp := &resource.ModifyPlanResponse{Plan: tfsdk.Plan{Schema: schemaResp.Schema, Raw: tftypes.NewValue(objType, nil)}}
	r.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan, Config: config, State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
	// Plan must not have been rewritten — resp.Plan.Raw stays the sentinel null.
	if !resp.Plan.Raw.IsNull() {
		t.Fatalf("ModifyPlan must not rewrite the plan when reconcile is a no-op, got %v", resp.Plan.Raw)
	}
}

// TestXrayRoutingResource_ModifyPlan_SkipsUnknownRuleCollection ensures
// ModifyPlan defers to the schema plan modifiers when the configured `rule`
// collection is itself unknown (e.g. a computed `dynamic` block whose
// for_each is not yet known). Decoding an unknown list into []XrayRoutingRule
// would raise a Value Conversion Error (hashicorp/terraform-plugin-framework#1025);
// the guard reads `rule` as types.List and returns early when IsUnknown().
func TestXrayRoutingResource_ModifyPlan_SkipsUnknownRuleCollection(t *testing.T) {
	r := &XrayRoutingResource{}
	ctx := context.Background()
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	ruleListType := objType.AttributeTypes["rule"].(tftypes.List)
	ruleObjType := ruleListType.ElementType.(tftypes.Object)

	// Config with an unknown `rule` collection (simulates a computed dynamic block).
	unknownRules := tftypes.NewValue(ruleListType, tftypes.UnknownValue)
	cfgRaw := routingRaw(t, schemaResp, unknownRules)

	// Plan and State are concrete and non-null so the first guard passes.
	planRules := tftypes.NewValue(ruleListType, []tftypes.Value{
		rtRuleValue(t, ruleObjType, "direct", "", nil, []string{"geoip:private"}),
	})
	stateRules := tftypes.NewValue(ruleListType, []tftypes.Value{
		rtRuleValue(t, ruleObjType, "direct", "", nil, []string{"geoip:private"}),
	})

	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: routingRaw(t, schemaResp, planRules)}
	config := tfsdk.Config{Schema: schemaResp.Schema, Raw: cfgRaw}
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: routingRaw(t, schemaResp, stateRules)}

	resp := &resource.ModifyPlanResponse{Plan: tfsdk.Plan{Schema: schemaResp.Schema, Raw: routingRaw(t, schemaResp, planRules)}}
	r.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan, Config: config, State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("ModifyPlan must not error on an unknown rule collection: %v", resp.Diagnostics)
	}
	// Plan must be unchanged — the guard returned without reconciling.
	var out XrayRoutingModel
	resp.Plan.Get(ctx, &out)
	if len(out.Rule) != 1 {
		t.Fatalf("expected 1 rule (unchanged plan), got %d", len(out.Rule))
	}
}

// TestXrayRoutingResource_ModifyPlan_ConfigDecodeError ensures ModifyPlan
// bails gracefully (no panic) when Config.GetAttribute("rule") fails — here
// the Config Raw is an Object whose `rule` value is a String (wrong type), so
// ValueFromTerraform raises a Value Conversion Error. Covers the GetAttribute
// HasError guard.
func TestXrayRoutingResource_ModifyPlan_ConfigDecodeError(t *testing.T) {
	r := &XrayRoutingResource{}
	ctx := context.Background()
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	ruleListType := objType.AttributeTypes["rule"].(tftypes.List)

	// Build a Config Object with a custom type where `rule` is String instead
	// of List — tftypes.NewValue validates, so we need a matching Object type.
	customAttrTypes := make(map[string]tftypes.Type, len(objType.AttributeTypes))
	for k, v := range objType.AttributeTypes {
		customAttrTypes[k] = v
	}
	customAttrTypes["rule"] = tftypes.String
	customObjType := tftypes.Object{AttributeTypes: customAttrTypes}

	customVals := make(map[string]tftypes.Value, len(customAttrTypes))
	for k, ty := range customAttrTypes {
		switch k {
		case "rule":
			customVals[k] = tftypes.NewValue(tftypes.String, "not-a-list")
		case "id":
			customVals[k] = tftypes.NewValue(tftypes.String, "xray_routing")
		default:
			customVals[k] = tftypes.NewValue(ty, nil)
		}
	}
	badCfgRaw := tftypes.NewValue(customObjType, customVals)

	// Valid Plan and State (non-null).
	planRules := tftypes.NewValue(ruleListType, []tftypes.Value{})
	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: routingRaw(t, schemaResp, planRules)}
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: routingRaw(t, schemaResp, planRules)}
	badConfig := tfsdk.Config{Schema: schemaResp.Schema, Raw: badCfgRaw}

	resp := &resource.ModifyPlanResponse{Plan: plan}
	r.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan, Config: badConfig, State: state}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics error when Config rule has the wrong type")
	}
}

// TestXrayRoutingResource_ModifyPlan_ConfigGetDecodeError ensures ModifyPlan
// bails gracefully when GetAttribute("rule") succeeds but Config.Get fails —
// here `rule` is a valid List but `id` is a Number (schema expects String),
// so decoding the whole Object into XrayRoutingModel raises a Value
// Conversion Error. Covers the Config.Get HasError guard.
func TestXrayRoutingResource_ModifyPlan_ConfigGetDecodeError(t *testing.T) {
	r := &XrayRoutingResource{}
	ctx := context.Background()
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	ruleListType := objType.AttributeTypes["rule"].(tftypes.List)

	// Build a Config Object with a custom type where `rule` is a valid List
	// (so GetAttribute succeeds) but `id` is Bool (schema expects String).
	customAttrTypes := make(map[string]tftypes.Type, len(objType.AttributeTypes))
	for k, v := range objType.AttributeTypes {
		customAttrTypes[k] = v
	}
	customAttrTypes["id"] = tftypes.Bool
	customObjType := tftypes.Object{AttributeTypes: customAttrTypes}

	customVals := make(map[string]tftypes.Value, len(customAttrTypes))
	for k, ty := range customAttrTypes {
		switch k {
		case "rule":
			customVals[k] = tftypes.NewValue(ruleListType, []tftypes.Value{})
		case "id":
			customVals[k] = tftypes.NewValue(tftypes.Bool, true)
		case "domain_strategy":
			customVals[k] = tftypes.NewValue(tftypes.String, "AsIs")
		default:
			customVals[k] = tftypes.NewValue(ty, nil)
		}
	}
	badCfgRaw := tftypes.NewValue(customObjType, customVals)

	// Valid Plan and State (non-null).
	planRules := tftypes.NewValue(ruleListType, []tftypes.Value{})
	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: routingRaw(t, schemaResp, planRules)}
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: routingRaw(t, schemaResp, planRules)}
	badConfig := tfsdk.Config{Schema: schemaResp.Schema, Raw: badCfgRaw}

	resp := &resource.ModifyPlanResponse{Plan: plan}
	r.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan, Config: badConfig, State: state}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics error when Config.Get fails to decode")
	}
}

// TestXrayRoutingResource_ModifyPlan_PlanDecodeError ensures ModifyPlan bails
// gracefully (no panic) when Plan.Raw has the wrong tftypes type and Plan.Get
// cannot decode into XrayRoutingModel. Config is valid with a known rule list
// so the GetAttribute + IsUnknown guards pass first.
func TestXrayRoutingResource_ModifyPlan_PlanDecodeError(t *testing.T) {
	r := &XrayRoutingResource{}
	ctx := context.Background()
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	ruleListType := objType.AttributeTypes["rule"].(tftypes.List)

	// Valid Config with a known rule list, valid State (non-null), bad Plan.
	cfgRules := tftypes.NewValue(ruleListType, []tftypes.Value{})
	config := tfsdk.Config{Schema: schemaResp.Schema, Raw: routingRaw(t, schemaResp, cfgRules)}
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: routingRaw(t, schemaResp, cfgRules)}
	badPlan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: tftypes.NewValue(tftypes.String, "not-an-object")}

	resp := &resource.ModifyPlanResponse{Plan: badPlan}
	r.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: badPlan, Config: config, State: state}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics error when Plan.Raw has the wrong type")
	}
}
