package provider

// ---------------------------------------------------------------------------
// VLESS outbound: schema, expand (typed model -> untyped map),
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

func vlessSettingsBlock() schema.ListNestedBlock {
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
				"flow": schema.StringAttribute{
					Optional: true, Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"encryption": schema.StringAttribute{
					Optional: true, Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"reverse_tag": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "VLESS reverse tag. Stored in 3x-ui as reverse.tag.",
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
			},
		},
	}
}

// --- Typed model -> untyped map ---

func expandVlessSettingsFromModel(list []XrayVlessOutSettings) []any {
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
		if !vs.Flow.IsNull() && !vs.Flow.IsUnknown() {
			entry["flow"] = vs.Flow.ValueString()
		}
		if !vs.Encryption.IsNull() && !vs.Encryption.IsUnknown() {
			entry["encryption"] = vs.Encryption.ValueString()
		}
		if !vs.ReverseTag.IsNull() && !vs.ReverseTag.IsUnknown() {
			entry["reverse_tag"] = vs.ReverseTag.ValueString()
		}
		out = append(out, entry)
	}
	return out
}

// --- Untyped map -> typed model ---

func flattenVlessSettingsToModel(list []any) []XrayVlessOutSettings {
	out := make([]XrayVlessOutSettings, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		vs := XrayVlessOutSettings{}

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

		if v, ok := raw["flow"].(string); ok && v != "" {
			vs.Flow = types.StringValue(v)
		} else {
			vs.Flow = types.StringNull()
		}

		if v, ok := raw["encryption"].(string); ok && v != "" {
			vs.Encryption = types.StringValue(v)
		} else {
			vs.Encryption = types.StringNull()
		}
		if v, ok := raw["reverse_tag"].(string); ok && v != "" {
			vs.ReverseTag = types.StringValue(v)
		} else {
			vs.ReverseTag = types.StringNull()
		}

		out = append(out, vs)
	}
	return out
}

// --- Untyped map -> Xray JSON ---

func expandVlessOutSettings(m map[string]any) map[string]any {
	list, ok := m["vless_settings"].([]any)
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
	if v, ok := item["flow"].(string); ok && v != "" {
		user["flow"] = v
	}
	if v, ok := item["encryption"].(string); ok && v != "" {
		user["encryption"] = v
	}
	if v, ok := item["reverse_tag"].(string); ok && v != "" {
		user["reverse"] = map[string]any{"tag": v}
	}
	if len(user) > 0 {
		server["users"] = []any{user}
	}
	return map[string]any{"vnext": []any{server}}
}

// --- Xray JSON -> untyped map ---

func flattenVlessOutSettings(in map[string]any) map[string]any {
	out := flattenVnextFirstUser(in, "id", "flow", "encryption")
	server := firstVnextServer(in)
	if server == nil {
		return out
	}
	users, ok := server["users"].([]any)
	if ok && len(users) > 0 {
		if user, ok := users[0].(map[string]any); ok {
			if tag := reverseTagValue(user["reverse"]); tag != "" {
				out["reverse_tag"] = tag
			}
		}
	}
	return out
}
