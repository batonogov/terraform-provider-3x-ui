package provider

import (
	"context"
	"reflect"
	"testing"

	resourcepath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// TestXrayBasicsEnvRoundTrip exercises the env block (3x-ui v3.5.0+, xray-core
// v26.7.11+): model → untyped map (expandXrayBasics) → wire map
// (buildXrayBasicsJSON) → untyped map (flattenXrayBasicsToMap) → model
// (flattenXrayBasics). xray-core stores env as map[string]string; the provider
// models it as a repeated {key,value} block. Keys must survive verbatim (no
// camelCase translation) and the flattened list must be deterministically
// sorted by key.
func TestXrayBasicsEnvRoundTrip(t *testing.T) {
	model := &XrayBasicsModel{
		ID: types.StringValue("xray_basics"),
		Env: []XrayBasicsEnv{
			{Key: types.StringValue("XRAY_LOG_LEVEL"), Value: types.StringValue("warning")},
			{Key: types.StringValue("XRAY_LOCATION_ASSET"), Value: types.StringValue("/usr/share/xray")},
		},
	}

	// 1. model → untyped
	expanded := expandXrayBasics(model)
	envMap, ok := expanded["env"].(map[string]any)
	if !ok {
		t.Fatalf("model→untyped: expected env map, got %T", expanded["env"])
	}
	if envMap["XRAY_LOG_LEVEL"] != "warning" {
		t.Fatalf("model→untyped: XRAY_LOG_LEVEL key not preserved verbatim, got %v", envMap["XRAY_LOG_LEVEL"])
	}

	// 2. untyped → wire (keys carried verbatim, no camelCase)
	wire := buildXrayBasicsJSON(expanded).(map[string]any)
	wireEnv, ok := wire["env"].(map[string]any)
	if !ok {
		t.Fatalf("untyped→wire: expected env map, got %T", wire["env"])
	}
	if wireEnv["XRAY_LOCATION_ASSET"] != "/usr/share/xray" {
		t.Fatalf("untyped→wire: key not preserved verbatim, got %v", wireEnv["XRAY_LOCATION_ASSET"])
	}

	// 3. wire → untyped
	flat := flattenXrayBasicsToMap(wire)
	flatEnv, ok := flat["env"].(map[string]any)
	if !ok {
		t.Fatalf("wire→untyped: expected env map, got %T", flat["env"])
	}
	if !reflect.DeepEqual(flatEnv, envMap) {
		t.Fatalf("wire→untyped: env map changed across wire round-trip:\n got  %v\n want %v", flatEnv, envMap)
	}

	// 4. untyped → model (flattened list sorted by key)
	model2 := flattenXrayBasics(flat)
	if len(model2.Env) != 2 {
		t.Fatalf("untyped→model: expected 2 env entries, got %d", len(model2.Env))
	}
	// Sorted alphabetically: XRAY_LOCATION_ASSET < XRAY_LOG_LEVEL
	if model2.Env[0].Key.ValueString() != "XRAY_LOCATION_ASSET" {
		t.Fatalf("untyped→model: expected sorted order, first key = %q",
			model2.Env[0].Key.ValueString())
	}
	if model2.Env[0].Value.ValueString() != "/usr/share/xray" {
		t.Fatalf("untyped→model: value mismatch, got %q", model2.Env[0].Value.ValueString())
	}
}

// TestXrayBasicsEnvOmittedWhenEmpty verifies that an unset env block produces
// no "env" key on the wire — xray-core treats a missing key as "no env".
func TestXrayBasicsEnvOmittedWhenEmpty(t *testing.T) {
	model := &XrayBasicsModel{
		ID: types.StringValue("xray_basics"),
		// Env intentionally nil.
	}
	expanded := expandXrayBasics(model)
	if _, ok := expanded["env"]; ok {
		t.Fatalf("nil env block must not produce an env key in the untyped map")
	}
	wire := buildXrayBasicsJSON(expanded).(map[string]any)
	if _, ok := wire["env"]; ok {
		t.Fatalf("nil env block must not be written to the wire")
	}
}

// TestXrayBasicsSchema exercises the full schema definition (incl. the v3.5.0
// env block) so the schema helper lines count toward Codecov patch coverage.
func TestXrayBasicsSchema(t *testing.T) {
	s := xrayBasicsSchema()
	if s.Blocks["env"] == nil {
		t.Fatal("expected env block in xray_basics schema")
	}
	if s.Blocks["metrics"] == nil {
		t.Fatal("expected metrics block in xray_basics schema")
	}
}

// TestXrayBasicsEnvSingleKeyEmptyValue covers the env flatten branch where a
// value is empty (ValueNull path) and the env block has a single entry.
func TestXrayBasicsEnvSingleKeyEmptyValue(t *testing.T) {
	model := &XrayBasicsModel{
		ID: types.StringValue("xray_basics"),
		Env: []XrayBasicsEnv{
			{Key: types.StringValue("ONLY"), Value: types.StringNull()},
		},
	}
	expanded := expandXrayBasics(model)
	envMap, ok := expanded["env"].(map[string]any)
	if !ok {
		t.Fatalf("expected env map, got %T", expanded["env"])
	}
	// Null value expands to empty string.
	if v, ok := envMap["ONLY"].(string); !ok || v != "" {
		t.Fatalf("expected ONLY to be empty string for null value, got %v", envMap["ONLY"])
	}
	// Round-trip back: single key preserved.
	m2 := flattenXrayBasics(flattenXrayBasicsToMap(buildXrayBasicsJSON(expanded)))
	if len(m2.Env) != 1 || m2.Env[0].Key.ValueString() != "ONLY" {
		t.Fatalf("single-key env round-trip failed, got %+v", m2.Env)
	}
}

// TestExtractXraySectionIncludesEnv confirms the merge-root extractor (used by
// threexui_xray_basics) picks up the top-level "env" key added in v3.5.0.
func TestExtractXraySectionIncludesEnv(t *testing.T) {
	current := map[string]any{
		"log": map[string]any{"loglevel": "warning"},
		"env": map[string]any{"XRAY_LOG_LEVEL": "warning"},
		"dns": map[string]any{"servers": []any{"1.1.1.1"}}, // must NOT be extracted by basics
	}
	section := xraySectionBasics
	got := extractXraySection(current, section).(map[string]any)
	if _, ok := got["env"]; !ok {
		t.Fatalf("extractXraySection must include env for merge-root basics, got %v", got)
	}
	if _, ok := got["dns"]; ok {
		t.Fatalf("extractXraySection must not leak non-basics keys like dns, got %v", got)
	}
	if _, ok := got["log"]; !ok {
		t.Fatalf("extractXraySection must still include log, got %v", got)
	}
}

func TestAlignBasicsBlocksWithPlanClearsLog(t *testing.T) {
	state := &XrayBasicsModel{
		Log: []XrayBasicsLog{{Loglevel: types.StringValue("warning")}},
	}
	plan := &XrayBasicsModel{}

	alignBasicsBlocksWithPlan(state, plan)

	if state.Log != nil {
		t.Fatalf("expected state.Log to be nil after align (plan has no log), got %v", state.Log)
	}
}

// TestAlignBasicsBlocksWithPlanClearsEnv verifies the drift-prevention path:
// when the plan has no env block but the state does, alignBasicsBlocksWithPlan
// nils out state.Env so Terraform does not raise "was absent, but now present".
func TestAlignBasicsBlocksWithPlanClearsEnv(t *testing.T) {
	state := &XrayBasicsModel{
		Env: []XrayBasicsEnv{
			{Key: types.StringValue("XRAY_LOG_LEVEL"), Value: types.StringValue("warning")},
		},
	}
	plan := &XrayBasicsModel{} // no Env block

	alignBasicsBlocksWithPlan(state, plan)

	if state.Env != nil {
		t.Fatalf("expected state.Env to be nil after align (plan has no env), got %v", state.Env)
	}
}

// TestAlignBasicsBlocksWithPlanKeepsEnv confirms that when both plan and state
// carry an env block, align leaves the state's env entries intact.
func TestAlignBasicsBlocksWithPlanKeepsEnv(t *testing.T) {
	state := &XrayBasicsModel{
		Env: []XrayBasicsEnv{
			{Key: types.StringValue("XRAY_LOG_LEVEL"), Value: types.StringValue("warning")},
		},
	}
	plan := &XrayBasicsModel{
		Env: []XrayBasicsEnv{
			{Key: types.StringValue("XRAY_LOG_LEVEL")}, // presence is what matters
		},
	}

	alignBasicsBlocksWithPlan(state, plan)

	if len(state.Env) != 1 {
		t.Fatalf("expected state.Env to be preserved when plan has env, got %v", state.Env)
	}
}

func basicsLevelRaw(t *testing.T, levelType tftypes.Object, id int64, fields map[string]tftypes.Value) tftypes.Value {
	t.Helper()
	return basicsLevelRawWithID(t, levelType, tftypes.NewValue(tftypes.Number, id), fields)
}

func basicsLevelRawWithID(t *testing.T, levelType tftypes.Object, id tftypes.Value, fields map[string]tftypes.Value) tftypes.Value {
	t.Helper()
	values := make(map[string]tftypes.Value, len(levelType.AttributeTypes))
	for name, attrType := range levelType.AttributeTypes {
		values[name] = tftypes.NewValue(attrType, nil)
	}
	values["id"] = id
	for name, value := range fields {
		values[name] = value
	}
	return tftypes.NewValue(levelType, values)
}

func basicsPolicyRaw(t *testing.T, policyType tftypes.Object, levelList tftypes.Value) tftypes.Value {
	t.Helper()
	return basicsPolicyRawWithFields(t, policyType, map[string]tftypes.Value{"level": levelList})
}

func basicsPolicyRawWithFields(t *testing.T, policyType tftypes.Object, fields map[string]tftypes.Value) tftypes.Value {
	t.Helper()
	values := make(map[string]tftypes.Value, len(policyType.AttributeTypes))
	for name, attrType := range policyType.AttributeTypes {
		values[name] = tftypes.NewValue(attrType, nil)
	}
	for name, value := range fields {
		values[name] = value
	}
	return tftypes.NewValue(policyType, values)
}

func basicsRaw(t *testing.T, schemaResp resource.SchemaResponse, policyList tftypes.Value) tftypes.Value {
	t.Helper()
	ctx := context.Background()
	resourceType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	values := make(map[string]tftypes.Value, len(resourceType.AttributeTypes))
	for name, attrType := range resourceType.AttributeTypes {
		values[name] = tftypes.NewValue(attrType, nil)
	}
	values["id"] = tftypes.NewValue(tftypes.String, "xray_basics")
	values["policy"] = policyList
	return tftypes.NewValue(resourceType, values)
}

func TestXrayBasicsResourceModifyPlanCanonicalizesReorderedConfiguredLevels(t *testing.T) {
	ctx := context.Background()
	r := &XrayBasicsResource{}
	modifier, ok := any(r).(resource.ResourceWithModifyPlan)
	if !ok {
		t.Fatal("XrayBasicsResource must implement ResourceWithModifyPlan")
	}

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	resourceType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	policyListType := resourceType.AttributeTypes["policy"].(tftypes.List)
	policyType := policyListType.ElementType.(tftypes.Object)
	levelListType := policyType.AttributeTypes["level"].(tftypes.List)
	levelType := levelListType.ElementType.(tftypes.Object)
	systemListType := policyType.AttributeTypes["system"].(tftypes.List)
	systemType := systemListType.ElementType.(tftypes.Object)

	intValue := func(value int64) tftypes.Value {
		return tftypes.NewValue(tftypes.Number, value)
	}
	configuredLevels := tftypes.NewValue(levelListType, []tftypes.Value{
		basicsLevelRaw(t, levelType, 1, map[string]tftypes.Value{"conn_idle": intValue(123)}),
		basicsLevelRaw(t, levelType, 0, map[string]tftypes.Value{"handshake": intValue(17)}),
	})
	pollutedLevels := tftypes.NewValue(levelListType, []tftypes.Value{
		basicsLevelRaw(t, levelType, 1, map[string]tftypes.Value{"handshake": intValue(17), "conn_idle": intValue(123)}),
		basicsLevelRaw(t, levelType, 0, map[string]tftypes.Value{"handshake": intValue(17), "conn_idle": intValue(123)}),
	})
	priorLevels := tftypes.NewValue(levelListType, []tftypes.Value{
		basicsLevelRaw(t, levelType, 0, map[string]tftypes.Value{"handshake": intValue(17), "conn_idle": intValue(300)}),
		basicsLevelRaw(t, levelType, 1, map[string]tftypes.Value{"handshake": intValue(4), "conn_idle": intValue(123)}),
	})
	systemValues := make(map[string]tftypes.Value, len(systemType.AttributeTypes))
	for name, attrType := range systemType.AttributeTypes {
		systemValues[name] = tftypes.NewValue(attrType, nil)
	}
	systemValues["stats_inbound_downlink"] = tftypes.NewValue(tftypes.Bool, true)
	priorSystem := tftypes.NewValue(systemListType, []tftypes.Value{tftypes.NewValue(systemType, systemValues)})
	policyList := func(levels tftypes.Value) tftypes.Value {
		return tftypes.NewValue(policyListType, []tftypes.Value{basicsPolicyRaw(t, policyType, levels)})
	}
	priorPolicyList := tftypes.NewValue(policyListType, []tftypes.Value{
		basicsPolicyRawWithFields(t, policyType, map[string]tftypes.Value{"level": priorLevels, "system": priorSystem}),
	})

	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: basicsRaw(t, schemaResp, policyList(pollutedLevels))}
	config := tfsdk.Config{Schema: schemaResp.Schema, Raw: basicsRaw(t, schemaResp, policyList(configuredLevels))}
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: basicsRaw(t, schemaResp, priorPolicyList)}
	resp := &resource.ModifyPlanResponse{Plan: plan}

	modifier.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan, Config: config, State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected ModifyPlan diagnostics: %v", resp.Diagnostics)
	}

	var got XrayBasicsModel
	resp.Diagnostics.Append(resp.Plan.Get(ctx, &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("cannot decode modified plan: %v", resp.Diagnostics)
	}
	if len(got.Policy) != 1 || len(got.Policy[0].Level) != 2 {
		t.Fatalf("expected one policy with two levels, got %#v", got.Policy)
	}
	levels := got.Policy[0].Level
	if levels[0].ID.ValueInt64() != 0 || levels[1].ID.ValueInt64() != 1 {
		t.Fatalf("planned levels must be canonicalized by ID, got [%d, %d]", levels[0].ID.ValueInt64(), levels[1].ID.ValueInt64())
	}
	if levels[0].Handshake.ValueInt64() != 17 || levels[0].ConnIdle.ValueInt64() != 300 {
		t.Fatalf("level 0 must keep its own configured value and same-ID default, got %#v", levels[0])
	}
	if levels[1].ConnIdle.ValueInt64() != 123 || levels[1].Handshake.ValueInt64() != 4 {
		t.Fatalf("level 1 must keep its own configured value and same-ID default, got %#v", levels[1])
	}
	if len(got.Policy[0].System) != 1 || !got.Policy[0].System[0].StatsInboundDownlink.ValueBool() {
		t.Fatalf("omitted policy system sibling must be preserved from state, got %#v", got.Policy[0].System)
	}
}

func TestXrayBasicsResourceModifyPlanCanonicalizesReverseOrderOnCreate(t *testing.T) {
	ctx := context.Background()
	r := &XrayBasicsResource{}
	modifier := any(r).(resource.ResourceWithModifyPlan)

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	resourceType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	policyListType := resourceType.AttributeTypes["policy"].(tftypes.List)
	policyType := policyListType.ElementType.(tftypes.Object)
	levelListType := policyType.AttributeTypes["level"].(tftypes.List)
	levelType := levelListType.ElementType.(tftypes.Object)

	configuredLevels := tftypes.NewValue(levelListType, []tftypes.Value{
		basicsLevelRaw(t, levelType, 2, map[string]tftypes.Value{"conn_idle": tftypes.NewValue(tftypes.Number, int64(120))}),
		basicsLevelRaw(t, levelType, 0, map[string]tftypes.Value{"handshake": tftypes.NewValue(tftypes.Number, int64(4))}),
	})
	policyList := tftypes.NewValue(policyListType, []tftypes.Value{
		basicsPolicyRaw(t, policyType, configuredLevels),
	})
	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: basicsRaw(t, schemaResp, policyList)}
	config := tfsdk.Config{Schema: schemaResp.Schema, Raw: basicsRaw(t, schemaResp, policyList)}
	nullState := tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(resourceType, nil)}
	resp := &resource.ModifyPlanResponse{Plan: plan}

	modifier.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan, Config: config, State: nullState}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected ModifyPlan diagnostics: %v", resp.Diagnostics)
	}

	var got XrayBasicsModel
	resp.Diagnostics.Append(resp.Plan.Get(ctx, &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("cannot decode modified plan: %v", resp.Diagnostics)
	}
	levels := got.Policy[0].Level
	if levels[0].ID.ValueInt64() != 0 || levels[1].ID.ValueInt64() != 2 {
		t.Fatalf("create plan levels must be canonicalized by ID, got [%d, %d]", levels[0].ID.ValueInt64(), levels[1].ID.ValueInt64())
	}
	if !levels[0].ConnIdle.IsUnknown() || !levels[1].Handshake.IsUnknown() ||
		!levels[0].StatsUserUplink.IsUnknown() || !levels[1].StatsUserDownlink.IsUnknown() {
		t.Fatalf("omitted computed level fields must remain unknown, got %#v", levels)
	}
}

func TestXrayBasicsResourceModifyPlanPreservesUnknownPolicyShapes(t *testing.T) {
	ctx := context.Background()
	r := &XrayBasicsResource{}
	modifier, ok := any(r).(resource.ResourceWithModifyPlan)
	if !ok {
		t.Fatal("XrayBasicsResource must implement ResourceWithModifyPlan")
	}

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	resourceType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	policyListType := resourceType.AttributeTypes["policy"].(tftypes.List)
	policyType := policyListType.ElementType.(tftypes.Object)
	levelListType := policyType.AttributeTypes["level"].(tftypes.List)
	levelType := levelListType.ElementType.(tftypes.Object)

	knownLevel := basicsLevelRaw(t, levelType, 0, nil)
	knownLevels := tftypes.NewValue(levelListType, []tftypes.Value{knownLevel})
	knownPolicy := basicsPolicyRaw(t, policyType, knownLevels)
	knownPolicyList := tftypes.NewValue(policyListType, []tftypes.Value{knownPolicy})
	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: basicsRaw(t, schemaResp, knownPolicyList)}
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: basicsRaw(t, schemaResp, knownPolicyList)}

	tests := map[string]tftypes.Value{
		"policy collection": tftypes.NewValue(policyListType, tftypes.UnknownValue),
		"policy object": tftypes.NewValue(policyListType, []tftypes.Value{
			tftypes.NewValue(policyType, tftypes.UnknownValue),
		}),
		"level collection": tftypes.NewValue(policyListType, []tftypes.Value{
			basicsPolicyRaw(t, policyType, tftypes.NewValue(levelListType, tftypes.UnknownValue)),
		}),
		"level element": tftypes.NewValue(policyListType, []tftypes.Value{
			basicsPolicyRaw(t, policyType, tftypes.NewValue(levelListType, []tftypes.Value{
				tftypes.NewValue(levelType, tftypes.UnknownValue),
			})),
		}),
		"level id": tftypes.NewValue(policyListType, []tftypes.Value{
			basicsPolicyRaw(t, policyType, tftypes.NewValue(levelListType, []tftypes.Value{
				basicsLevelRawWithID(t, levelType, tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), nil),
			})),
		}),
	}

	for name, configuredPolicy := range tests {
		t.Run(name, func(t *testing.T) {
			config := tfsdk.Config{Schema: schemaResp.Schema, Raw: basicsRaw(t, schemaResp, configuredPolicy)}
			resp := &resource.ModifyPlanResponse{Plan: plan}
			modifier.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan, Config: config, State: state}, resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("unknown %s must not produce diagnostics: %v", name, resp.Diagnostics)
			}
			if !resp.Plan.Raw.Equal(plan.Raw) {
				t.Fatalf("unknown %s must preserve the proposed plan\n got: %s\nwant: %s", name, resp.Plan.Raw, plan.Raw)
			}
		})
	}

	priorTests := map[string]tftypes.Value{
		"unknown policy object": tftypes.NewValue(policyListType, []tftypes.Value{
			tftypes.NewValue(policyType, tftypes.UnknownValue),
		}),
		"unknown level element": tftypes.NewValue(policyListType, []tftypes.Value{
			basicsPolicyRaw(t, policyType, tftypes.NewValue(levelListType, []tftypes.Value{
				tftypes.NewValue(levelType, tftypes.UnknownValue),
			})),
		}),
	}
	for name, priorPolicy := range priorTests {
		t.Run("prior "+name, func(t *testing.T) {
			config := tfsdk.Config{Schema: schemaResp.Schema, Raw: basicsRaw(t, schemaResp, knownPolicyList)}
			priorState := tfsdk.State{Schema: schemaResp.Schema, Raw: basicsRaw(t, schemaResp, priorPolicy)}
			resp := &resource.ModifyPlanResponse{Plan: plan}
			modifier.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan, Config: config, State: priorState}, resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("%s must be ignored without diagnostics: %v", name, resp.Diagnostics)
			}
		})
	}
}

func TestXrayBasicsResourceModifyPlanSortsWithUnknownNonIDLeaf(t *testing.T) {
	ctx := context.Background()
	r := &XrayBasicsResource{}
	modifier, ok := any(r).(resource.ResourceWithModifyPlan)
	if !ok {
		t.Fatal("XrayBasicsResource must implement ResourceWithModifyPlan")
	}

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	resourceType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	policyListType := resourceType.AttributeTypes["policy"].(tftypes.List)
	policyType := policyListType.ElementType.(tftypes.Object)
	levelListType := policyType.AttributeTypes["level"].(tftypes.List)
	levelType := levelListType.ElementType.(tftypes.Object)

	configuredLevels := tftypes.NewValue(levelListType, []tftypes.Value{
		basicsLevelRaw(t, levelType, 1, map[string]tftypes.Value{
			"handshake": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		}),
		basicsLevelRaw(t, levelType, 0, map[string]tftypes.Value{
			"handshake": tftypes.NewValue(tftypes.Number, int64(4)),
		}),
	})
	pollutedLevels := tftypes.NewValue(levelListType, []tftypes.Value{
		basicsLevelRaw(t, levelType, 1, map[string]tftypes.Value{
			"handshake": tftypes.NewValue(tftypes.Number, int64(99)),
		}),
		basicsLevelRaw(t, levelType, 0, map[string]tftypes.Value{
			"handshake": tftypes.NewValue(tftypes.Number, int64(4)),
		}),
	})
	policyList := func(levels tftypes.Value) tftypes.Value {
		return tftypes.NewValue(policyListType, []tftypes.Value{basicsPolicyRaw(t, policyType, levels)})
	}
	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: basicsRaw(t, schemaResp, policyList(pollutedLevels))}
	config := tfsdk.Config{Schema: schemaResp.Schema, Raw: basicsRaw(t, schemaResp, policyList(configuredLevels))}
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: basicsRaw(t, schemaResp, policyList(pollutedLevels))}
	resp := &resource.ModifyPlanResponse{Plan: plan}

	modifier.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan, Config: config, State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unknown non-ID leaf must remain representable: %v", resp.Diagnostics)
	}
	var got XrayBasicsModel
	resp.Diagnostics.Append(resp.Plan.Get(ctx, &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("cannot decode modified plan: %v", resp.Diagnostics)
	}
	levels := got.Policy[0].Level
	if levels[0].ID.ValueInt64() != 0 || levels[1].ID.ValueInt64() != 1 {
		t.Fatalf("levels with known IDs must still be sorted, got [%d, %d]", levels[0].ID.ValueInt64(), levels[1].ID.ValueInt64())
	}
	if !levels[1].Handshake.IsUnknown() {
		t.Fatalf("configured unknown handshake must replace stale planned value, got %v", levels[1].Handshake)
	}
}

func TestXrayBasicsResourceModifyPlanPreservesOmittedPolicy(t *testing.T) {
	ctx := context.Background()
	r := &XrayBasicsResource{}
	modifier := any(r).(resource.ResourceWithModifyPlan)

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	resourceType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	policyListType := resourceType.AttributeTypes["policy"].(tftypes.List)
	nullPolicy := tftypes.NewValue(policyListType, nil)
	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: basicsRaw(t, schemaResp, nullPolicy)}
	config := tfsdk.Config{Schema: schemaResp.Schema, Raw: basicsRaw(t, schemaResp, nullPolicy)}
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: basicsRaw(t, schemaResp, nullPolicy)}
	resp := &resource.ModifyPlanResponse{Plan: plan}

	modifier.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan, Config: config, State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("omitted policy must not produce diagnostics: %v", resp.Diagnostics)
	}
	var gotPolicy types.List
	resp.Diagnostics.Append(resp.Plan.GetAttribute(ctx, resourcepath.Root("policy"), &gotPolicy)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("cannot read modified policy plan: %v", resp.Diagnostics)
	}
	if !gotPolicy.IsNull() {
		t.Fatalf("omitted policy must remain null, got %v", gotPolicy)
	}
}

func TestXrayBasicsResourceModifyPlanPreservesNullDestroyPlan(t *testing.T) {
	ctx := context.Background()
	r := &XrayBasicsResource{}
	modifier := any(r).(resource.ResourceWithModifyPlan)

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	resourceType := schemaResp.Schema.Type().TerraformType(ctx)
	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: tftypes.NewValue(resourceType, nil)}
	resp := &resource.ModifyPlanResponse{Plan: plan}

	modifier.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("null destroy plan must not produce diagnostics: %v", resp.Diagnostics)
	}
	if !resp.Plan.Raw.IsNull() {
		t.Fatalf("null destroy plan must remain null, got %s", resp.Plan.Raw)
	}
}

func TestFlattenBasicsPolicyLevelsSortsIDsNumerically(t *testing.T) {
	levels := flattenBasicsPolicyLevels(map[string]any{
		"10": map[string]any{"connIdle": 10},
		"2":  map[string]any{"connIdle": 2},
		"1":  map[string]any{"connIdle": 1},
	})
	if len(levels) != 3 {
		t.Fatalf("expected three flattened levels, got %d", len(levels))
	}
	got := make([]int, len(levels))
	for i, level := range levels {
		got[i] = level.(map[string]any)["id"].(int)
	}
	want := []int{1, 2, 10}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("flattened level IDs must use numeric order, got %v, want %v", got, want)
	}
}

func TestFlattenBasicsPolicyLevelsSortsNumericBeforeNonnumericKeys(t *testing.T) {
	levels := flattenBasicsPolicyLevels(map[string]any{
		"10":    map[string]any{"connIdle": 10},
		"two":   map[string]any{"connIdle": 22},
		"1":     map[string]any{"connIdle": 1},
		"alpha": map[string]any{"connIdle": 11},
	})
	if len(levels) != 4 {
		t.Fatalf("expected four flattened levels, got %d", len(levels))
	}
	got := make([]int, len(levels))
	for i, level := range levels {
		got[i] = level.(map[string]any)["conn_idle"].(int)
	}
	want := []int{1, 10, 11, 22}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("numeric keys must sort before lexical keys, got %v, want %v", got, want)
	}
}
