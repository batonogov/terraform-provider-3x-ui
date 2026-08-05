package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestXrayBalancerFallbackTagAndSettingsRoundTrip covers the v3.5.0-relevant
// balancer fields added to xray_balancers: fallback_tag (balancer-to-balancer
// fallback) and strategy.settings (leastPing/leastLoad tuning, including the
// costs array). Exercises all four conversion layers:
// model → untyped (expandXrayBalancers) → wire (expandBalancers) →
// untyped (flattenBalancers) → model (flattenXrayBalancers).
func TestXrayBalancerFallbackTagAndSettingsRoundTrip(t *testing.T) {
	model := &XrayBalancersModel{
		Balancer: []XrayBalancerEntry{
			{
				Tag: types.StringValue("bal"),
				Selector: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("out-1"),
					types.StringValue("out-2"),
				}),
				FallbackTag: types.StringValue("fallback-bal"),
				Strategy: []XrayBalancerStrategy{
					{
						Type: types.StringValue("leastLoad"),
						Settings: []XrayBalancerStrategySettings{
							{
								Expected:  types.Int64Value(3),
								MaxRTT:    types.StringValue("500ms"),
								Tolerance: types.Float64Value(0.5),
								Baselines: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("base-1")}),
								Costs: []XrayBalancerCost{
									{Regexp: types.BoolValue(true), Match: types.StringValue(".*"), Value: types.Float64Value(1.5)},
								},
							},
						},
					},
				},
			},
		},
	}

	// 1. model → untyped
	expanded := expandXrayBalancers(model)
	list, _ := expanded["balancer"].([]any)
	entry, _ := list[0].(map[string]any)
	if entry["fallbackTag"] != "fallback-bal" {
		t.Fatalf("model→untyped: expected fallbackTag=fallback-bal, got %v", entry["fallbackTag"])
	}
	strategy, _ := entry["strategy"].([]any)
	stratMap, _ := strategy[0].(map[string]any)
	if _, ok := stratMap["settings"].(map[string]any); !ok {
		t.Fatalf("model→untyped: expected settings map, got %T", stratMap["settings"])
	}

	// 2. untyped → wire
	wire := expandBalancers(list)
	wireEntry, _ := wire[0].(map[string]any)
	if wireEntry["fallbackTag"] != "fallback-bal" {
		t.Fatalf("untyped→wire: expected fallbackTag, got %v", wireEntry["fallbackTag"])
	}
	wireStrat, _ := wireEntry["strategy"].(map[string]any)
	wireSettings, _ := wireStrat["settings"].(map[string]any)
	if wireSettings["expected"] != 3 {
		t.Fatalf("untyped→wire: expected settings.expected=3, got %v", wireSettings["expected"])
	}
	wireCosts, _ := wireSettings["costs"].([]any)
	if len(wireCosts) != 1 {
		t.Fatalf("untyped→wire: expected 1 cost, got %d", len(wireCosts))
	}

	// 3. wire → untyped (snake_case keys)
	flat := flattenBalancers(wire)
	flatEntry, _ := flat[0].(map[string]any)
	if flatEntry["fallback_tag"] != "fallback-bal" {
		t.Fatalf("wire→untyped: expected fallback_tag, got %v", flatEntry["fallback_tag"])
	}

	// 4. untyped → model
	model2 := flattenXrayBalancers(map[string]any{"balancer": flat})
	if len(model2.Balancer) != 1 {
		t.Fatalf("untyped→model: expected 1 balancer, got %d", len(model2.Balancer))
	}
	b := model2.Balancer[0]
	if b.FallbackTag.ValueString() != "fallback-bal" {
		t.Fatalf("untyped→model: expected FallbackTag=fallback-bal, got %q", b.FallbackTag)
	}
	if len(b.Strategy) != 1 || len(b.Strategy[0].Settings) != 1 {
		t.Fatalf("untyped→model: expected 1 strategy with 1 settings block, got %+v", b.Strategy)
	}
	st := b.Strategy[0].Settings[0]
	if st.Expected.ValueInt64() != 3 {
		t.Fatalf("untyped→model: Expected round-trip failed: %d", st.Expected.ValueInt64())
	}
	if st.Tolerance.ValueFloat64() != 0.5 {
		t.Fatalf("untyped→model: Tolerance round-trip failed: %f", st.Tolerance.ValueFloat64())
	}
	if st.MaxRTT.ValueString() != "500ms" {
		t.Fatalf("untyped→model: MaxRTT round-trip failed: %q", st.MaxRTT)
	}
	if len(st.Costs) != 1 || st.Costs[0].Value.ValueFloat64() != 1.5 {
		t.Fatalf("untyped→model: Costs round-trip failed: %+v", st.Costs)
	}
	if !st.Costs[0].Regexp.ValueBool() {
		t.Fatalf("untyped→model: Costs.Regexp round-trip failed")
	}
}

// TestXrayBalancerFallbackTagOmittedWhenEmpty confirms a balancer with no
// fallback tag and no strategy settings omits those keys on the wire.
func TestXrayBalancerFallbackTagOmittedWhenEmpty(t *testing.T) {
	model := &XrayBalancersModel{
		Balancer: []XrayBalancerEntry{
			{
				Tag:      types.StringValue("plain"),
				Selector: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("out-1")}),
			},
		},
	}
	expanded := expandXrayBalancers(model)
	list, _ := expanded["balancer"].([]any)
	entry, _ := list[0].(map[string]any)
	if _, ok := entry["fallbackTag"]; ok {
		t.Fatalf("unset fallback_tag must not appear on the wire")
	}
	if _, ok := entry["strategy"]; ok {
		t.Fatalf("unset strategy must not appear on the wire")
	}
}

// TestToFloat64 confirms the numeric-coercion helper accepts every numeric
// kind encoding/json / Go maps may carry (JSON decodes all numbers as float64,
// but in-process round-trip values may be int/int64). Non-numeric values must
// return ok=false so flatten leaves the field null rather than panicking.
func TestToFloat64(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want float64
		ok   bool
	}{
		{"float64", float64(1.5), 1.5, true},
		{"int", 3, 3.0, true},
		{"int64", int64(7), 7.0, true},
		{"int32", int32(2), 2.0, true},
		{"float32", float32(0.25), 0.25, true},
		{"string", "1.5", 0, false},
		{"nil", nil, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := toFloat64(tc.in)
			if ok != tc.ok {
				t.Fatalf("toFloat64(%v) ok = %v, want %v", tc.in, ok, tc.ok)
			}
			if ok && got != tc.want {
				t.Fatalf("toFloat64(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

// TestXrayBalancersSchema exercises the full schema definition (incl. the
// v3.5.0 fallback_tag + strategy.settings nested blocks) so the schema helper
// lines count toward Codecov patch coverage.
func TestXrayBalancersSchema(t *testing.T) {
	s := xrayBalancersSchema()
	if s.Attributes["id"] == nil {
		t.Fatal("expected id attribute")
	}
	if s.Blocks["balancer"] == nil {
		t.Fatal("expected balancer block")
	}
}

// TestFlattenXrayBalancersEdgeCases covers the defensive branches in
// flattenXrayBalancers: missing balancer key, non-list balancer, non-map item,
// and a balancer entry lacking tag/selector.
func TestFlattenXrayBalancersEdgeCases(t *testing.T) {
	cases := []struct {
		name string
		data map[string]any
		want int // expected number of balancer entries
	}{
		{"no balancer key", map[string]any{}, 0},
		{"balancer not a list", map[string]any{"balancer": "oops"}, 0},
		{"non-map item skipped", map[string]any{"balancer": []any{"str"}}, 0},
		{"entry without tag/selector", map[string]any{"balancer": []any{
			map[string]any{"fallbackTag": "fb"},
		}}, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := flattenXrayBalancers(tc.data)
			if len(m.Balancer) != tc.want {
				t.Fatalf("expected %d balancers, got %d", tc.want, len(m.Balancer))
			}
		})
	}
}

// TestFlattenBalancerStrategyEdgeCases covers the nil/empty branches.
func TestFlattenBalancerStrategyEdgeCases(t *testing.T) {
	if flattenBalancerStrategy(map[string]any{}) != nil {
		t.Fatal("empty strategy map must return nil")
	}
	// Strategy with only a non-type key → out stays empty → nil.
	if flattenBalancerStrategy(map[string]any{"settings": map[string]any{}}) != nil {
		t.Fatal("strategy with only empty settings must return nil")
	}
	// Type without settings → out has type only.
	got := flattenBalancerStrategy(map[string]any{"type": "random"})
	if got == nil || got["type"] != "random" {
		t.Fatalf("expected type=random, got %v", got)
	}
}

// TestFlattenBalancerStrategySettingsIntExpected confirms expected survives a
// Go int (in-process round-trip), not just JSON float64.
func TestFlattenBalancerStrategySettingsIntExpected(t *testing.T) {
	res := flattenBalancerStrategySettings(map[string]any{
		"settings": map[string]any{
			"expected":  5, // int, not float64
			"tolerance": 0.25,
			"baselines": []any{"b1"},
			"costs": []any{
				map[string]any{"regexp": true, "match": ".*"}, // cost without value
			},
		},
	})
	if len(res) != 1 {
		t.Fatalf("expected 1 settings block, got %d", len(res))
	}
	st := res[0]
	if st.Expected.ValueInt64() != 5 {
		t.Fatalf("expected int expected=5, got %d", st.Expected.ValueInt64())
	}
	if st.Tolerance.ValueFloat64() != 0.25 {
		t.Fatalf("expected tolerance=0.25, got %f", st.Tolerance.ValueFloat64())
	}
	if len(st.Costs) != 1 || !st.Costs[0].Regexp.ValueBool() {
		t.Fatalf("expected 1 cost with regexp=true, got %+v", st.Costs)
	}
}

func TestFlattenBalancerStrategySettingsBaselinesDefaultToTypedNull(t *testing.T) {
	tests := map[string]map[string]any{
		"absent": {"expected": 2},
		"empty":  {"expected": 2, "baselines": []any{}},
	}

	for name, settings := range tests {
		t.Run(name, func(t *testing.T) {
			res := flattenBalancerStrategySettings(map[string]any{"settings": settings})
			if len(res) != 1 {
				t.Fatalf("expected 1 settings block, got %d", len(res))
			}

			baselines := res[0].Baselines
			if !baselines.IsNull() {
				t.Fatalf("expected baselines to be null, got %s", baselines)
			}
			if got := baselines.ElementType(context.Background()); !got.Equal(types.StringType) {
				t.Fatalf("expected baselines element type %s, got %s", types.StringType, got)
			}
		})
	}
}

// TestExpandBalancerStrategyEdgeCases covers the empty/non-map branches.
func TestExpandBalancerStrategyEdgeCases(t *testing.T) {
	if expandBalancerStrategy(nil) != nil {
		t.Fatal("nil list must return nil")
	}
	if expandBalancerStrategy([]any{"not-a-map"}) != nil {
		t.Fatal("non-map item must return nil")
	}
	if expandBalancerStrategy([]any{map[string]any{}}) != nil {
		t.Fatal("empty map (no type/settings) must return nil")
	}
}

// TestBuildXrayBalancersJSONEdgeCases covers the missing/non-list balancer key.
func TestBuildXrayBalancersJSONEdgeCases(t *testing.T) {
	res := buildXrayBalancersJSON(map[string]any{})
	arr, ok := res.([]any)
	if !ok || len(arr) != 0 {
		t.Fatalf("missing balancer key must yield empty []any, got %#v", res)
	}
	res = buildXrayBalancersJSON(map[string]any{"balancer": "oops"})
	arr, _ = res.([]any)
	if len(arr) != 0 {
		t.Fatalf("non-list balancer must yield empty []any, got %#v", res)
	}
}
