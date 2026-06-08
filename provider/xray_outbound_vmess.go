package provider

// ---------------------------------------------------------------------------
// VMess outbound: schema, expand (typed model -> untyped map),
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

func vmessSettingsBlock() schema.ListNestedBlock {
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
				"id": schema.StringAttribute{
					Optional: true, Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"security": schema.StringAttribute{
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

func expandVmessSettingsFromModel(list []XrayVmessOutSettings) []any {
	out := make([]any, 0, len(list))
	for _, vs := range list {
		entry := map[string]any{}
		if !vs.Address.IsNull() && !vs.Address.IsUnknown() {
			entry["address"] = vs.Address.ValueString()
		}
		if !vs.Port.IsNull() && !vs.Port.IsUnknown() {
			entry["port"] = int(vs.Port.ValueInt64())
		}
		if !vs.ID.IsNull() && !vs.ID.IsUnknown() {
			entry["id"] = vs.ID.ValueString()
		}
		if !vs.Security.IsNull() && !vs.Security.IsUnknown() {
			entry["security"] = vs.Security.ValueString()
		}
		out = append(out, entry)
	}
	return out
}

// --- Untyped map -> typed model ---

func flattenVmessSettingsToModel(list []any) []XrayVmessOutSettings {
	out := make([]XrayVmessOutSettings, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		vs := XrayVmessOutSettings{}

		if v, ok := raw["address"].(string); ok && v != "" {
			vs.Address = types.StringValue(v)
		} else {
			vs.Address = types.StringNull()
		}

		if v, ok := raw["port"]; ok {
			vs.Port = types.Int64Value(int64(intValue(v)))
		} else {
			vs.Port = types.Int64Null()
		}

		if v, ok := raw["id"].(string); ok && v != "" {
			vs.ID = types.StringValue(v)
		} else {
			vs.ID = types.StringNull()
		}

		if v, ok := raw["security"].(string); ok && v != "" {
			vs.Security = types.StringValue(v)
		} else {
			vs.Security = types.StringNull()
		}

		out = append(out, vs)
	}
	return out
}

// --- Untyped map -> Xray JSON ---

func expandVmessOutSettings(m map[string]any) map[string]any {
	list, ok := m["vmess_settings"].([]any)
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
	user := map[string]any{}
	if v, ok := item["id"].(string); ok && v != "" {
		user["id"] = v
	}
	if v, ok := item["security"].(string); ok && v != "" {
		user["security"] = v
	}
	if len(user) > 0 {
		server["users"] = []any{user}
	}
	return map[string]any{"vnext": []any{server}}
}

// --- Xray JSON -> untyped map ---

func flattenVmessOutSettings(in map[string]any) map[string]any {
	return flattenVnextFirstUser(in, "id", "security")
}
