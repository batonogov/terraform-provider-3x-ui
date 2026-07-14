package provider

// ---------------------------------------------------------------------------
// DNS outbound: schema, expand (typed model -> untyped map),
// flatten (untyped map -> typed model), JSON expand/flatten
// ---------------------------------------------------------------------------

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- Schema ---

func dnsSettingsBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"network": schema.StringAttribute{
					Optional: true, Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"address": schema.StringAttribute{
					Optional: true, Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"port": schema.Int64Attribute{
					Optional: true, Computed: true,
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.UseStateForUnknown(),
					},
				},
				"non_ip_query": schema.StringAttribute{
					Optional: true, Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"block_types": schema.ListAttribute{
					Optional:    true,
					Computed:    true,
					ElementType: types.Int64Type,
					PlanModifiers: []planmodifier.List{
						listplanmodifier.UseStateForUnknown(),
					},
				},
			},
		},
	}
}

// --- Typed model -> untyped map ---

func expandDNSSettingsFromModel(list []XrayOutboundDNSSettings) []any {
	out := make([]any, 0, len(list))
	for _, ds := range list {
		entry := map[string]any{}
		if !ds.Network.IsNull() && !ds.Network.IsUnknown() {
			entry["network"] = ds.Network.ValueString()
		}
		if !ds.Address.IsNull() && !ds.Address.IsUnknown() {
			entry["address"] = ds.Address.ValueString()
		}
		if !ds.Port.IsNull() && !ds.Port.IsUnknown() {
			entry["port"] = int(ds.Port.ValueInt64())
		}
		if !ds.NonIPQuery.IsNull() && !ds.NonIPQuery.IsUnknown() {
			entry["non_ip_query"] = ds.NonIPQuery.ValueString()
		}
		if !ds.BlockTypes.IsNull() && !ds.BlockTypes.IsUnknown() {
			entry["block_types"] = expandInt64List(ds.BlockTypes)
		}
		out = append(out, entry)
	}
	return out
}

// --- Untyped map -> typed model ---

func flattenDNSSettingsToModel(list []any) []XrayOutboundDNSSettings {
	out := make([]XrayOutboundDNSSettings, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		ds := XrayOutboundDNSSettings{}

		if v, ok := raw["network"].(string); ok && v != "" {
			ds.Network = types.StringValue(v)
		} else {
			ds.Network = types.StringNull()
		}

		if v, ok := raw["address"].(string); ok && v != "" {
			ds.Address = types.StringValue(v)
		} else {
			ds.Address = types.StringNull()
		}

		if v, ok := raw["port"]; ok {
			ds.Port = types.Int64Value(int64(intValue(v)))
		} else {
			ds.Port = types.Int64Null()
		}

		if v, ok := raw["non_ip_query"].(string); ok && v != "" {
			ds.NonIPQuery = types.StringValue(v)
		} else {
			ds.NonIPQuery = types.StringNull()
		}

		if v, ok := raw["block_types"]; ok {
			ds.BlockTypes = flattenToInt64List(v)
		} else {
			ds.BlockTypes = types.ListNull(types.Int64Type)
		}

		out = append(out, ds)
	}
	return out
}

// --- Untyped map -> Xray JSON ---

func expandOutboundDNSSettings(m map[string]any) map[string]any {
	list, ok := m["dns_settings"].([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["network"].(string); ok && v != "" {
		out["network"] = v
	}
	if v, ok := item["address"].(string); ok && v != "" {
		out["address"] = v
	}
	if v, ok := item["port"].(int); ok && v != 0 {
		out["port"] = v
	}
	if v, ok := item["non_ip_query"].(string); ok && v != "" {
		out["nonIPQuery"] = v
	}
	if v, ok := item["block_types"].([]any); ok && len(v) > 0 {
		out["blockTypes"] = flattenIntList(v)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// --- Xray JSON -> untyped map ---

func flattenOutboundDNSSettings(in map[string]any) map[string]any {
	out := map[string]any{}
	if v, ok := in["network"].(string); ok {
		out["network"] = v
	}
	if v, ok := in["address"].(string); ok {
		out["address"] = v
	}
	if v, ok := in["port"]; ok {
		out["port"] = intValue(v)
	}
	if v, ok := in["nonIPQuery"].(string); ok {
		out["non_ip_query"] = v
	}
	if v, ok := in["blockTypes"].([]any); ok {
		out["block_types"] = flattenIntList(v)
	}
	return out
}
