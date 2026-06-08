package provider

// ---------------------------------------------------------------------------
// Socks outbound: schema, expand (typed model -> untyped map),
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

func socksSettingsBlock() schema.ListNestedBlock {
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
				"user": schema.StringAttribute{
					Optional: true, Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"pass": schema.StringAttribute{
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

func expandSocksSettingsFromModel(list []XraySocksOutSettings) []any {
	out := make([]any, 0, len(list))
	for _, ss := range list {
		entry := map[string]any{}
		if !ss.Address.IsNull() && !ss.Address.IsUnknown() {
			entry["address"] = ss.Address.ValueString()
		}
		if !ss.Port.IsNull() && !ss.Port.IsUnknown() {
			entry["port"] = int(ss.Port.ValueInt64())
		}
		if !ss.User.IsNull() && !ss.User.IsUnknown() {
			entry["user"] = ss.User.ValueString()
		}
		if !ss.Pass.IsNull() && !ss.Pass.IsUnknown() {
			entry["pass"] = ss.Pass.ValueString()
		}
		out = append(out, entry)
	}
	return out
}

// --- Untyped map -> typed model ---

func flattenSocksSettingsToModel(list []any) []XraySocksOutSettings {
	out := make([]XraySocksOutSettings, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		ss := XraySocksOutSettings{}

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

		if v, ok := raw["user"].(string); ok && v != "" {
			ss.User = types.StringValue(v)
		} else {
			ss.User = types.StringNull()
		}

		if v, ok := raw["pass"].(string); ok && v != "" {
			ss.Pass = types.StringValue(v)
		} else {
			ss.Pass = types.StringNull()
		}

		out = append(out, ss)
	}
	return out
}

// --- Untyped map -> Xray JSON ---

func expandSocksOutSettings(m map[string]any) map[string]any {
	list, ok := m["socks_settings"].([]any)
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
	if v, ok := item["user"].(string); ok && v != "" {
		user["user"] = v
	}
	if v, ok := item["pass"].(string); ok && v != "" {
		user["pass"] = v
	}
	if len(user) > 0 {
		server["users"] = []any{user}
	}
	return map[string]any{"servers": []any{server}}
}

// --- Xray JSON -> untyped map ---

func flattenSocksOutSettings(in map[string]any) map[string]any {
	out := map[string]any{}
	servers, ok := in["servers"].([]any)
	if !ok || len(servers) == 0 {
		return out
	}
	server, ok := servers[0].(map[string]any)
	if !ok {
		return out
	}
	if v, ok := server["address"].(string); ok {
		out["address"] = v
	}
	if v, ok := server["port"]; ok {
		out["port"] = intValue(v)
	}
	users, ok := server["users"].([]any)
	if ok && len(users) > 0 {
		user, ok := users[0].(map[string]any)
		if ok {
			if v, ok := user["user"].(string); ok {
				out["user"] = v
			}
			if v, ok := user["pass"].(string); ok {
				out["pass"] = v
			}
		}
	}
	return out
}
