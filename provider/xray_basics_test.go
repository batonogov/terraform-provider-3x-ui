package provider

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
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
