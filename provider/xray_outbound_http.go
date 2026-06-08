package provider

// ---------------------------------------------------------------------------
// HTTP outbound: schema, expand (typed model -> untyped map),
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

func httpSettingsBlock() schema.ListNestedBlock {
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

func expandHTTPSettingsFromModel(list []XrayHTTPOutSettings) []any {
	out := make([]any, 0, len(list))
	for _, hs := range list {
		entry := map[string]any{}
		if !hs.Address.IsNull() && !hs.Address.IsUnknown() {
			entry["address"] = hs.Address.ValueString()
		}
		if !hs.Port.IsNull() && !hs.Port.IsUnknown() {
			entry["port"] = int(hs.Port.ValueInt64())
		}
		if !hs.User.IsNull() && !hs.User.IsUnknown() {
			entry["user"] = hs.User.ValueString()
		}
		if !hs.Pass.IsNull() && !hs.Pass.IsUnknown() {
			entry["pass"] = hs.Pass.ValueString()
		}
		out = append(out, entry)
	}
	return out
}

// --- Untyped map -> typed model ---

func flattenHTTPSettingsToModel(list []any) []XrayHTTPOutSettings {
	out := make([]XrayHTTPOutSettings, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		hs := XrayHTTPOutSettings{}

		if v, ok := raw["address"].(string); ok && v != "" {
			hs.Address = types.StringValue(v)
		} else {
			hs.Address = types.StringNull()
		}

		if v, ok := raw["port"]; ok {
			hs.Port = types.Int64Value(int64(intValue(v)))
		} else {
			hs.Port = types.Int64Null()
		}

		if v, ok := raw["user"].(string); ok && v != "" {
			hs.User = types.StringValue(v)
		} else {
			hs.User = types.StringNull()
		}

		if v, ok := raw["pass"].(string); ok && v != "" {
			hs.Pass = types.StringValue(v)
		} else {
			hs.Pass = types.StringNull()
		}

		out = append(out, hs)
	}
	return out
}

// --- Untyped map -> Xray JSON ---

func expandHTTPOutSettings(m map[string]any) map[string]any {
	list, ok := m["http_settings"].([]any)
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

func flattenHTTPOutSettings(in map[string]any) map[string]any {
	return flattenSocksOutSettings(in) // same structure
}
