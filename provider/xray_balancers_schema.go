package provider

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	Tag      types.String           `tfsdk:"tag"`
	Selector types.List             `tfsdk:"selector"` // list of strings
	Strategy []XrayBalancerStrategy `tfsdk:"strategy"`
}

type XrayBalancerStrategy struct {
	Type types.String `tfsdk:"type"`
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
					},
					Blocks: map[string]schema.Block{
						"strategy": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
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

		if len(b.Strategy) > 0 {
			strategies := make([]any, 0, len(b.Strategy))
			for _, s := range b.Strategy {
				sEntry := map[string]any{}
				if !s.Type.IsNull() && !s.Type.IsUnknown() {
					sEntry["type"] = s.Type.ValueString()
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
	if len(out) == 0 {
		return nil
	}
	return out
}
