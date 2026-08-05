package provider

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// Typed models
// ---------------------------------------------------------------------------

type XrayBalancersModel struct {
	ID       types.String        `tfsdk:"id"`
	Balancer []XrayBalancerEntry `tfsdk:"balancer"`
}

type XrayBalancerEntry struct {
	Tag         types.String           `tfsdk:"tag"`
	Selector    types.List             `tfsdk:"selector"` // list of strings
	FallbackTag types.String           `tfsdk:"fallback_tag"`
	Strategy    []XrayBalancerStrategy `tfsdk:"strategy"`
}

type XrayBalancerStrategy struct {
	Type     types.String                   `tfsdk:"type"`
	Settings []XrayBalancerStrategySettings `tfsdk:"settings"`
}

// XrayBalancerStrategySettings mirrors xray-core's balancer strategy settings
// (frontend BalancerStrategySettingsSchema). Used by the leastPing/leastLoad
// strategies to tune observatory-based selection.
type XrayBalancerStrategySettings struct {
	Expected  types.Int64        `tfsdk:"expected"`
	MaxRTT    types.String       `tfsdk:"max_rtt"`
	Tolerance types.Float64      `tfsdk:"tolerance"`
	Baselines types.List         `tfsdk:"baselines"` // list of strings
	Costs     []XrayBalancerCost `tfsdk:"costs"`
}

// XrayBalancerCost mirrors xray-core's balancer cost object
// (frontend BalancerCostObjectSchema): a regexp/keyword match → cost value.
type XrayBalancerCost struct {
	Regexp types.Bool    `tfsdk:"regexp"`
	Match  types.String  `tfsdk:"match"`
	Value  types.Float64 `tfsdk:"value"`
}

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

func xrayBalancersSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"balancer": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"tag": schema.StringAttribute{
							Optional: true, Computed: true,
						},
						"selector": schema.ListAttribute{
							Required:    true,
							ElementType: types.StringType,
						},
						"fallback_tag": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Fallback balancer tag used when this balancer has no healthy outbound (xray-core balancer-to-balancer fallback). Empty means no fallback.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"strategy": singletonListNestedBlock(schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"settings": singletonListNestedBlock(schema.ListNestedBlock{
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"expected": schema.Int64Attribute{
													Optional: true, Computed: true,
													Description: "Number of expected alive outbounds (leastPing/leastLoad).",
													PlanModifiers: []planmodifier.Int64{
														int64planmodifier.UseStateForUnknown(),
													},
												},
												"max_rtt": schema.StringAttribute{
													Optional: true, Computed: true,
													Description: "Max acceptable round-trip time (xray duration string, e.g. \"500ms\").",
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.UseStateForUnknown(),
													},
												},
												"tolerance": schema.Float64Attribute{
													Optional: true, Computed: true,
													Description: "Selection tolerance 0.0–1.0 (leastLoad).",
													PlanModifiers: []planmodifier.Float64{
														float64planmodifier.UseStateForUnknown(),
													},
												},
												"baselines": schema.ListAttribute{
													Optional: true, Computed: true,
													ElementType: types.StringType,
													PlanModifiers: []planmodifier.List{
														listplanmodifier.UseStateForUnknown(),
													},
												},
											},
											Blocks: map[string]schema.Block{
												"costs": schema.ListNestedBlock{
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"regexp": schema.BoolAttribute{
																Optional: true, Computed: true,
																PlanModifiers: []planmodifier.Bool{
																	boolplanmodifier.UseStateForUnknown(),
																},
															},
															"match": schema.StringAttribute{
																Optional: true, Computed: true,
																PlanModifiers: []planmodifier.String{
																	stringplanmodifier.UseStateForUnknown(),
																},
															},
															"value": schema.Float64Attribute{
																Optional: true, Computed: true,
																PlanModifiers: []planmodifier.Float64{
																	float64planmodifier.UseStateForUnknown(),
																},
															},
														},
													},
												},
											},
										},
									}),
								},
							},
						}),
					},
				},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Typed expand: model -> untyped map (for buildXrayBalancersJSON)
// ---------------------------------------------------------------------------

func expandXrayBalancers(m *XrayBalancersModel) map[string]any {
	payload := map[string]any{}
	if m.Balancer == nil {
		return payload
	}

	balancers := make([]any, 0, len(m.Balancer))
	for _, b := range m.Balancer {
		entry := map[string]any{}

		if !b.Tag.IsNull() && !b.Tag.IsUnknown() {
			entry["tag"] = b.Tag.ValueString()
		}

		if !b.Selector.IsNull() && !b.Selector.IsUnknown() {
			elems := b.Selector.Elements()
			sel := make([]any, 0, len(elems))
			for _, e := range elems {
				sv, ok := e.(types.String)
				if ok && !sv.IsNull() && !sv.IsUnknown() {
					sel = append(sel, sv.ValueString())
				}
			}
			entry["selector"] = sel
		}

		if !b.FallbackTag.IsNull() && !b.FallbackTag.IsUnknown() {
			entry["fallbackTag"] = b.FallbackTag.ValueString()
		}

		if len(b.Strategy) > 0 {
			strategies := make([]any, 0, len(b.Strategy))
			for _, s := range b.Strategy {
				sEntry := map[string]any{}
				if !s.Type.IsNull() && !s.Type.IsUnknown() {
					sEntry["type"] = s.Type.ValueString()
				}
				if settings := expandBalancerStrategySettings(s.Settings); settings != nil {
					sEntry["settings"] = settings
				}
				if len(sEntry) > 0 {
					strategies = append(strategies, sEntry)
				}
			}
			if len(strategies) > 0 {
				entry["strategy"] = strategies
			}
		}

		if len(entry) > 0 {
			balancers = append(balancers, entry)
		}
	}

	payload["balancer"] = balancers
	return payload
}

// ---------------------------------------------------------------------------
// Typed flatten: untyped map (from flattenXrayBalancersToMap) -> model
// ---------------------------------------------------------------------------

func flattenXrayBalancers(data map[string]any) *XrayBalancersModel {
	m := &XrayBalancersModel{
		ID: types.StringValue("xray_balancers"),
	}

	v, ok := data["balancer"]
	if !ok {
		return m
	}

	list, ok := v.([]any)
	if !ok {
		return m
	}

	balancers := make([]XrayBalancerEntry, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}

		entry := XrayBalancerEntry{}

		if tag, ok := raw["tag"].(string); ok {
			entry.Tag = types.StringValue(tag)
		}

		if sel, ok := raw["selector"].([]any); ok {
			vals := make([]attr.Value, 0, len(sel))
			for _, s := range sel {
				if str, ok := s.(string); ok {
					vals = append(vals, types.StringValue(str))
				}
			}
			entry.Selector = types.ListValueMust(types.StringType, vals)
		} else {
			entry.Selector = types.ListValueMust(types.StringType, []attr.Value{})
		}

		if ft, ok := raw["fallback_tag"].(string); ok && ft != "" {
			entry.FallbackTag = types.StringValue(ft)
		}

		if stratList, ok := raw["strategy"].([]any); ok {
			strategies := make([]XrayBalancerStrategy, 0, len(stratList))
			for _, sItem := range stratList {
				sMap, ok := sItem.(map[string]any)
				if !ok {
					continue
				}
				s := XrayBalancerStrategy{}
				if t, ok := sMap["type"].(string); ok {
					s.Type = types.StringValue(t)
				}
				if sMap != nil {
					s.Settings = flattenBalancerStrategySettings(sMap)
				}
				strategies = append(strategies, s)
			}
			entry.Strategy = strategies
		}

		balancers = append(balancers, entry)
	}

	m.Balancer = balancers
	return m
}

// ---------------------------------------------------------------------------
// Existing untyped build/flatten functions (used by CRUD layer)
// ---------------------------------------------------------------------------

func buildXrayBalancersJSON(d map[string]any) any {
	v, ok := d["balancer"]
	if !ok {
		return []any{}
	}
	list, ok := v.([]any)
	if !ok {
		return []any{}
	}
	return expandBalancers(list)
}

func expandBalancers(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		if v, ok := m["tag"].(string); ok && v != "" {
			entry["tag"] = v
		}
		if v, ok := m["selector"]; ok {
			if list, ok := v.([]any); ok {
				entry["selector"] = expandStringList(list)
			}
		}
		if v, ok := m["fallbackTag"].(string); ok && v != "" {
			entry["fallbackTag"] = v
		}
		if v, ok := m["strategy"]; ok {
			if list, ok := v.([]any); ok {
				if s := expandBalancerStrategy(list); s != nil {
					entry["strategy"] = s
				}
			}
		}
		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	return out
}

func expandBalancerStrategy(list []any) map[string]any {
	if len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["type"].(string); ok && v != "" {
		out["type"] = v
	}
	if v, ok := item["settings"].(map[string]any); ok && len(v) > 0 {
		out["settings"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// expandBalancerStrategySettings converts the TF settings block list into the
// wire map xray-core expects (camelCase keys). Returns nil when the block is
// absent/empty so the key is omitted on the wire.
func expandBalancerStrategySettings(settings []XrayBalancerStrategySettings) map[string]any {
	if len(settings) == 0 {
		return nil
	}
	st := settings[0]
	out := map[string]any{}
	if !st.Expected.IsNull() && !st.Expected.IsUnknown() {
		out["expected"] = int(st.Expected.ValueInt64())
	}
	if !st.MaxRTT.IsNull() && !st.MaxRTT.IsUnknown() {
		out["maxRTT"] = st.MaxRTT.ValueString()
	}
	if !st.Tolerance.IsNull() && !st.Tolerance.IsUnknown() {
		out["tolerance"] = st.Tolerance.ValueFloat64()
	}
	if !st.Baselines.IsNull() && !st.Baselines.IsUnknown() {
		elems := st.Baselines.Elements()
		bl := make([]any, 0, len(elems))
		for _, e := range elems {
			if sv, ok := e.(types.String); ok && !sv.IsNull() && !sv.IsUnknown() {
				bl = append(bl, sv.ValueString())
			}
		}
		// Preserve an explicitly configured empty list. Omitting this key would
		// make the panel read-back indistinguishable from a null value and cause
		// Terraform to reject the post-apply state.
		out["baselines"] = bl
	}
	if len(st.Costs) > 0 {
		costs := make([]any, 0, len(st.Costs))
		for _, c := range st.Costs {
			cEntry := map[string]any{}
			if !c.Regexp.IsNull() && !c.Regexp.IsUnknown() {
				cEntry["regexp"] = c.Regexp.ValueBool()
			}
			if !c.Match.IsNull() && !c.Match.IsUnknown() {
				cEntry["match"] = c.Match.ValueString()
			}
			if !c.Value.IsNull() && !c.Value.IsUnknown() {
				cEntry["value"] = c.Value.ValueFloat64()
			}
			if len(cEntry) > 0 {
				costs = append(costs, cEntry)
			}
		}
		if len(costs) > 0 {
			out["costs"] = costs
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func flattenXrayBalancersToMap(data any) map[string]any {
	out := map[string]any{}
	if data == nil {
		return out
	}

	var list []any
	switch v := data.(type) {
	case []any:
		list = v
	case string:
		if err := json.Unmarshal([]byte(v), &list); err != nil {
			return out
		}
	default:
		return out
	}

	out["balancer"] = flattenBalancers(list)
	return out
}

func flattenBalancers(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		if v, ok := m["tag"].(string); ok {
			entry["tag"] = v
		}
		if v, ok := m["selector"].([]any); ok {
			entry["selector"] = v
		}
		if v, ok := m["fallbackTag"].(string); ok {
			entry["fallback_tag"] = v
		}
		if v, ok := m["strategy"].(map[string]any); ok {
			if s := flattenBalancerStrategy(v); s != nil {
				entry["strategy"] = []any{s}
			}
		}
		out = append(out, entry)
	}
	return out
}

func flattenBalancerStrategy(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := map[string]any{}
	if v, ok := in["type"].(string); ok {
		out["type"] = v
	}
	if v, ok := in["settings"].(map[string]any); ok && len(v) > 0 {
		out["settings"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// flattenBalancerStrategySettings converts the wire settings map (camelCase)
// back into the TF settings block list. JSON numbers arrive as float64.
func flattenBalancerStrategySettings(in map[string]any) []XrayBalancerStrategySettings {
	settingsRaw, ok := in["settings"].(map[string]any)
	if !ok || len(settingsRaw) == 0 {
		return nil
	}
	st := XrayBalancerStrategySettings{
		Baselines: types.ListNull(types.StringType),
	}
	if v, ok := settingsRaw["expected"]; ok {
		switch n := v.(type) {
		case float64:
			st.Expected = types.Int64Value(int64(n))
		case int:
			st.Expected = types.Int64Value(int64(n))
		case int64:
			st.Expected = types.Int64Value(n)
		}
	}
	if v, ok := settingsRaw["maxRTT"].(string); ok && v != "" {
		st.MaxRTT = types.StringValue(v)
	}
	if v, ok := settingsRaw["tolerance"]; ok {
		if f, ok := toFloat64(v); ok {
			st.Tolerance = types.Float64Value(f)
		}
	}
	if v, ok := settingsRaw["baselines"].([]any); ok {
		vals := make([]attr.Value, 0, len(v))
		for _, b := range v {
			if s, ok := b.(string); ok {
				vals = append(vals, types.StringValue(s))
			}
		}
		st.Baselines = types.ListValueMust(types.StringType, vals)
	}
	if v, ok := settingsRaw["costs"].([]any); ok && len(v) > 0 {
		costs := make([]XrayBalancerCost, 0, len(v))
		for _, cItem := range v {
			cMap, ok := cItem.(map[string]any)
			if !ok {
				continue
			}
			c := XrayBalancerCost{}
			if r, ok := cMap["regexp"].(bool); ok {
				c.Regexp = types.BoolValue(r)
			}
			if m, ok := cMap["match"].(string); ok {
				c.Match = types.StringValue(m)
			}
			if val, ok := cMap["value"]; ok {
				if f, ok := toFloat64(val); ok {
					c.Value = types.Float64Value(f)
				}
			}
			costs = append(costs, c)
		}
		if len(costs) > 0 {
			st.Costs = costs
		}
	}
	return []XrayBalancerStrategySettings{st}
}

// toFloat64 coerces a JSON-decoded numeric value to float64. encoding/json
// decodes all numbers as float64, but values built in-process (e.g. in
// round-trip tests) may be int/int64, so accept all numeric kinds.
func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	}
	return 0, false
}
