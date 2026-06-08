package provider

// ---------------------------------------------------------------------------
// Blackhole outbound: schema, expand (typed model -> untyped map),
// flatten (untyped map -> typed model), JSON expand/flatten
// ---------------------------------------------------------------------------

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- Schema ---

func blackholeSettingsBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"response_type": schema.StringAttribute{
					Optional: true, Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
			},
		},
	}
}

// --- Typed model -> untyped map ---

func expandBlackholeSettingsFromModel(list []XrayBlackholeSettings) []any {
	out := make([]any, 0, len(list))
	for _, bh := range list {
		entry := map[string]any{}
		if !bh.ResponseType.IsNull() && !bh.ResponseType.IsUnknown() {
			entry["response_type"] = bh.ResponseType.ValueString()
		}
		out = append(out, entry)
	}
	return out
}

// --- Untyped map -> typed model ---

func flattenBlackholeSettingsToModel(list []any) []XrayBlackholeSettings {
	out := make([]XrayBlackholeSettings, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		bh := XrayBlackholeSettings{}
		if v, ok := raw["response_type"].(string); ok && v != "" {
			bh.ResponseType = types.StringValue(v)
		} else {
			bh.ResponseType = types.StringNull()
		}
		out = append(out, bh)
	}
	return out
}

// --- Untyped map -> Xray JSON ---

func expandBlackholeSettings(m map[string]any) map[string]any {
	list, ok := m["blackhole_settings"].([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["response_type"].(string); ok && v != "" {
		out["response"] = map[string]any{"type": v}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// --- Xray JSON -> untyped map ---

func flattenBlackholeSettings(in map[string]any) map[string]any {
	out := map[string]any{}
	if v, ok := in["response"].(map[string]any); ok {
		if t, ok := v["type"].(string); ok {
			out["response_type"] = t
		}
	}
	if _, ok := out["response_type"]; !ok {
		out["response_type"] = "none"
	}
	return out
}
