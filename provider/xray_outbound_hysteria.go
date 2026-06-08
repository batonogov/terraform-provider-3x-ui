package provider

// ---------------------------------------------------------------------------
// Hysteria outbound: schema, expand (typed model -> untyped map),
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

func hysteriaSettingsBlock() schema.ListNestedBlock {
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
				"version": schema.Int64Attribute{
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

func expandHysteriaSettingsFromModel(list []XrayHysteriaOutSettings) []any {
	out := make([]any, 0, len(list))
	for _, hs := range list {
		entry := map[string]any{}
		if !hs.Address.IsNull() && !hs.Address.IsUnknown() {
			entry["address"] = hs.Address.ValueString()
		}
		if !hs.Port.IsNull() && !hs.Port.IsUnknown() {
			entry["port"] = int(hs.Port.ValueInt64())
		}
		if !hs.Version.IsNull() && !hs.Version.IsUnknown() {
			entry["version"] = int(hs.Version.ValueInt64())
		}
		out = append(out, entry)
	}
	return out
}

// --- Untyped map -> typed model ---

func flattenHysteriaSettingsToModel(list []any) []XrayHysteriaOutSettings {
	out := make([]XrayHysteriaOutSettings, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		hs := XrayHysteriaOutSettings{}

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

		if v, ok := raw["version"]; ok {
			hs.Version = types.Int64Value(int64(intValue(v)))
		} else {
			hs.Version = types.Int64Null()
		}

		out = append(out, hs)
	}
	return out
}

// --- Untyped map -> Xray JSON ---

func expandHysteriaOutSettings(m map[string]any) map[string]any {
	list, ok := m["hysteria_settings"].([]any)
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
	if v, ok := item["version"].(int); ok && v != 0 {
		server["version"] = v
	}
	return map[string]any{"servers": []any{server}}
}

// --- Xray JSON -> untyped map ---

func flattenHysteriaOutSettings(in map[string]any) map[string]any {
	out := flattenServersFirst(in)
	servers, ok := in["servers"].([]any)
	if ok && len(servers) > 0 {
		server, ok := servers[0].(map[string]any)
		if ok {
			if v, ok := server["version"]; ok {
				out["version"] = intValue(v)
			}
		}
	}
	return out
}
