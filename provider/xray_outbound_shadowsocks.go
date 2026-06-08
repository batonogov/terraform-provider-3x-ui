package provider

// ---------------------------------------------------------------------------
// Shadowsocks outbound: schema, expand (typed model -> untyped map),
// flatten (untyped map -> typed model), JSON expand/flatten
// ---------------------------------------------------------------------------

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- Schema ---

func shadowsocksSettingsBlock() schema.ListNestedBlock {
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
				"method": schema.StringAttribute{
					Optional: true, Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"uot": schema.BoolAttribute{
					Optional: true, Computed: true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
					},
				},
				"uot_version": schema.Int64Attribute{
					Optional: true, Computed: true,
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.UseStateForUnknown(),
					},
				},
			},
		},
	}
}

// --- Typed model -> untyped map ---

func expandShadowsocksSettingsFromModel(list []XrayShadowsocksOutSettings) []any {
	out := make([]any, 0, len(list))
	for _, ss := range list {
		entry := map[string]any{}
		if !ss.Address.IsNull() && !ss.Address.IsUnknown() {
			entry["address"] = ss.Address.ValueString()
		}
		if !ss.Port.IsNull() && !ss.Port.IsUnknown() {
			entry["port"] = int(ss.Port.ValueInt64())
		}
		if !ss.Password.IsNull() && !ss.Password.IsUnknown() {
			entry["password"] = ss.Password.ValueString()
		}
		if !ss.Method.IsNull() && !ss.Method.IsUnknown() {
			entry["method"] = ss.Method.ValueString()
		}
		if !ss.UOT.IsNull() && !ss.UOT.IsUnknown() {
			entry["uot"] = ss.UOT.ValueBool()
		}
		if !ss.UOTVersion.IsNull() && !ss.UOTVersion.IsUnknown() {
			entry["uot_version"] = int(ss.UOTVersion.ValueInt64())
		}
		out = append(out, entry)
	}
	return out
}

// --- Untyped map -> typed model ---

func flattenShadowsocksSettingsToModel(list []any) []XrayShadowsocksOutSettings {
	out := make([]XrayShadowsocksOutSettings, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		ss := XrayShadowsocksOutSettings{}

		if v, ok := raw["address"].(string); ok && v != "" {
			ss.Address = types.StringValue(v)
		} else {
			ss.Address = types.StringNull()
		}

		if v, ok := raw["port"]; ok {
			ss.Port = types.Int64Value(int64(intValue(v)))
		} else {
			ss.Port = types.Int64Null()
		}

		if v, ok := raw["password"].(string); ok && v != "" {
			ss.Password = types.StringValue(v)
		} else {
			ss.Password = types.StringNull()
		}

		if v, ok := raw["method"].(string); ok && v != "" {
			ss.Method = types.StringValue(v)
		} else {
			ss.Method = types.StringNull()
		}

		if v, ok := raw["uot"].(bool); ok {
			ss.UOT = types.BoolValue(v)
		} else {
			ss.UOT = types.BoolNull()
		}

		if v, ok := raw["uot_version"]; ok {
			ss.UOTVersion = types.Int64Value(int64(intValue(v)))
		} else {
			ss.UOTVersion = types.Int64Null()
		}

		out = append(out, ss)
	}
	return out
}

// --- Untyped map -> Xray JSON ---

func expandShadowsocksOutSettings(m map[string]any) map[string]any {
	list, ok := m["shadowsocks_settings"].([]any)
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
	if v, ok := item["method"].(string); ok && v != "" {
		server["method"] = v
	}
	if v, ok := item["uot"].(bool); ok {
		server["uot"] = v
	}
	if v, ok := item["uot_version"].(int); ok && v != 0 {
		server["UoTVersion"] = v
	}
	return map[string]any{"servers": []any{server}}
}

// --- Xray JSON -> untyped map ---

func flattenShadowsocksOutSettings(in map[string]any) map[string]any {
	out := flattenServersFirst(in, "password", "method")
	servers, ok := in["servers"].([]any)
	if ok && len(servers) > 0 {
		server, ok := servers[0].(map[string]any)
		if ok {
			if v, ok := server["uot"].(bool); ok {
				out["uot"] = v
			}
			if v, ok := server["UoTVersion"]; ok {
				out["uot_version"] = intValue(v)
			}
		}
	}
	return out
}
