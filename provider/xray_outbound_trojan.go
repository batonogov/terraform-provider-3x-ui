package provider

// ---------------------------------------------------------------------------
// Trojan outbound: schema, expand (typed model -> untyped map),
// flatten (untyped map -> typed model), JSON expand/flatten
// ---------------------------------------------------------------------------

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- Schema ---

func trojanSettingsBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
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
				"password": schema.StringAttribute{
					Optional: true, Computed: true, Sensitive: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
			},
		},
	}
}

// --- Typed model -> untyped map ---

func expandTrojanSettingsFromModel(list []XrayTrojanOutSettings) []any {
	out := make([]any, 0, len(list))
	for _, ts := range list {
		entry := map[string]any{}
		if !ts.Address.IsNull() && !ts.Address.IsUnknown() {
			entry["address"] = ts.Address.ValueString()
		}
		if !ts.Port.IsNull() && !ts.Port.IsUnknown() {
			entry["port"] = int(ts.Port.ValueInt64())
		}
		if !ts.Password.IsNull() && !ts.Password.IsUnknown() {
			entry["password"] = ts.Password.ValueString()
		}
		out = append(out, entry)
	}
	return out
}

// --- Untyped map -> typed model ---

func flattenTrojanSettingsToModel(list []any) []XrayTrojanOutSettings {
	out := make([]XrayTrojanOutSettings, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		ts := XrayTrojanOutSettings{}

		if v, ok := raw["address"].(string); ok && v != "" {
			ts.Address = types.StringValue(v)
		} else {
			ts.Address = types.StringNull()
		}

		if v, ok := raw["port"]; ok {
			ts.Port = types.Int64Value(int64(intValue(v)))
		} else {
			ts.Port = types.Int64Null()
		}

		if v, ok := raw["password"].(string); ok && v != "" {
			ts.Password = types.StringValue(v)
		} else {
			ts.Password = types.StringNull()
		}

		out = append(out, ts)
	}
	return out
}

// --- Untyped map -> Xray JSON ---

func expandTrojanOutSettings(m map[string]any) map[string]any {
	list, ok := m["trojan_settings"].([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	server := map[string]any{}
	if v, ok := item["address"].(string); ok && v != "" {
		server["address"] = v
	}
	if v, ok := item["port"].(int); ok && v != 0 {
		server["port"] = v
	}
	if v, ok := item["password"].(string); ok && v != "" {
		server["password"] = v
	}
	return map[string]any{"servers": []any{server}}
}

// --- Xray JSON -> untyped map ---

func flattenTrojanOutSettings(in map[string]any) map[string]any {
	return flattenServersFirst(in, "password")
}
