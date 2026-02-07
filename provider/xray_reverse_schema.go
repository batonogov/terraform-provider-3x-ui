package provider

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// Typed models
// ---------------------------------------------------------------------------

type XrayReverseModel struct {
	ID     types.String       `tfsdk:"id"`
	Bridge []XrayReverseEntry `tfsdk:"bridge"`
	Portal []XrayReverseEntry `tfsdk:"portal"`
}

type XrayReverseEntry struct {
	Tag    types.String `tfsdk:"tag"`
	Domain types.String `tfsdk:"domain"`
}

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

func xrayReverseSchema() schema.Schema {
	entryBlock := schema.ListNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"tag": schema.StringAttribute{
					Required: true,
				},
				"domain": schema.StringAttribute{
					Required: true,
				},
			},
		},
	}

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
			"bridge": entryBlock,
			"portal": entryBlock,
		},
	}
}

// ---------------------------------------------------------------------------
// Typed model -> untyped map  (for buildXrayReverseJSON)
// ---------------------------------------------------------------------------

func expandXrayReverse(m *XrayReverseModel) map[string]any {
	out := map[string]any{}

	if entries := expandReverseEntryList(m.Bridge); len(entries) > 0 {
		out["bridge"] = entries
	}
	if entries := expandReverseEntryList(m.Portal); len(entries) > 0 {
		out["portal"] = entries
	}

	return out
}

func expandReverseEntryList(entries []XrayReverseEntry) []any {
	if len(entries) == 0 {
		return nil
	}
	out := make([]any, 0, len(entries))
	for _, e := range entries {
		entry := map[string]any{}
		if !e.Tag.IsNull() && !e.Tag.IsUnknown() {
			entry["tag"] = e.Tag.ValueString()
		}
		if !e.Domain.IsNull() && !e.Domain.IsUnknown() {
			entry["domain"] = e.Domain.ValueString()
		}
		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Untyped map -> typed model  (from flattenXrayReverseToMap output)
// ---------------------------------------------------------------------------

func flattenXrayReverse(data map[string]any) *XrayReverseModel {
	m := &XrayReverseModel{
		ID: types.StringValue(xraySectionReverse.id),
	}

	if v, ok := data["bridge"].([]any); ok {
		m.Bridge = flattenReverseEntryList(v)
	}
	if v, ok := data["portal"].([]any); ok {
		m.Portal = flattenReverseEntryList(v)
	}

	return m
}

func flattenReverseEntryList(list []any) []XrayReverseEntry {
	if len(list) == 0 {
		return nil
	}
	out := make([]XrayReverseEntry, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := XrayReverseEntry{}
		if v, ok := m["tag"].(string); ok {
			entry.Tag = types.StringValue(v)
		}
		if v, ok := m["domain"].(string); ok {
			entry.Domain = types.StringValue(v)
		}
		out = append(out, entry)
	}
	return out
}

// ---------------------------------------------------------------------------
// Legacy untyped build / flatten (used by CRUD layer)
// ---------------------------------------------------------------------------

func buildXrayReverseJSON(d map[string]any) any {
	payload := map[string]any{}

	if v, ok := d["bridge"]; ok {
		if list, ok := v.([]any); ok {
			payload["bridges"] = expandReverseEntries(list)
		}
	}
	if v, ok := d["portal"]; ok {
		if list, ok := v.([]any); ok {
			payload["portals"] = expandReverseEntries(list)
		}
	}

	return payload
}

func expandReverseEntries(list []any) []any {
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
		if v, ok := m["domain"].(string); ok && v != "" {
			entry["domain"] = v
		}
		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	return out
}

func flattenXrayReverseToMap(data any) map[string]any {
	out := map[string]any{}
	if data == nil {
		return out
	}

	var payload map[string]any
	switch v := data.(type) {
	case map[string]any:
		payload = v
	case string:
		if err := json.Unmarshal([]byte(v), &payload); err != nil {
			return out
		}
	default:
		return out
	}

	if v, ok := payload["bridges"].([]any); ok {
		out["bridge"] = flattenReverseEntries(v)
	}
	if v, ok := payload["portals"].([]any); ok {
		out["portal"] = flattenReverseEntries(v)
	}

	return out
}

func flattenReverseEntries(list []any) []any {
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
		if v, ok := m["domain"].(string); ok {
			entry["domain"] = v
		}
		out = append(out, entry)
	}
	return out
}
