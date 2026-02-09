package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

type InboundSniffingModel struct {
	Enabled      types.Bool `tfsdk:"enabled"`
	DestOverride types.List `tfsdk:"dest_override"` // list of string
	MetadataOnly types.Bool `tfsdk:"metadata_only"`
	RouteOnly    types.Bool `tfsdk:"route_only"`
}

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

func inboundSniffingBlockSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Sniffing settings for the inbound.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether sniffing is enabled.",
			},
			"dest_override": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Destination override protocols (e.g. http, tls, quic, fakedns).",
			},
			"metadata_only": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Only sniff metadata.",
			},
			"route_only": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Only use sniffing for routing.",
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Typed model -> untyped map
// ---------------------------------------------------------------------------

func expandSniffingFromModel(m *InboundSniffingModel) map[string]any {
	if m == nil {
		return nil
	}
	out := map[string]any{}
	if !m.Enabled.IsNull() && !m.Enabled.IsUnknown() {
		out["enabled"] = m.Enabled.ValueBool()
	}
	if !m.DestOverride.IsNull() && !m.DestOverride.IsUnknown() {
		out["dest_override"] = typesListToAnySlice(m.DestOverride)
	}
	if !m.MetadataOnly.IsNull() && !m.MetadataOnly.IsUnknown() {
		out["metadata_only"] = m.MetadataOnly.ValueBool()
	}
	if !m.RouteOnly.IsNull() && !m.RouteOnly.IsUnknown() {
		out["route_only"] = m.RouteOnly.ValueBool()
	}
	return out
}

// ---------------------------------------------------------------------------
// Untyped map -> typed model
// ---------------------------------------------------------------------------

func flattenSniffingToModel(data map[string]any) *InboundSniffingModel {
	if len(data) == 0 {
		return nil
	}
	m := &InboundSniffingModel{}

	if v, ok := data["enabled"].(bool); ok {
		m.Enabled = types.BoolValue(v)
	} else {
		m.Enabled = types.BoolNull()
	}

	if v, ok := data["dest_override"]; ok {
		m.DestOverride = anySliceToTypesList(v)
	} else {
		m.DestOverride = types.ListNull(types.StringType)
	}

	if v, ok := data["metadata_only"].(bool); ok {
		m.MetadataOnly = types.BoolValue(v)
	} else {
		m.MetadataOnly = types.BoolNull()
	}

	if v, ok := data["route_only"].(bool); ok {
		m.RouteOnly = types.BoolValue(v)
	} else {
		m.RouteOnly = types.BoolNull()
	}

	return m
}

// ---------------------------------------------------------------------------
// JSON string -> untyped map (wraps existing flattenSniffing)
// ---------------------------------------------------------------------------

func flattenSniffingToMap(sniffingJSON string) map[string]any {
	flat := flattenSniffing(sniffingJSON)
	if len(flat) == 0 {
		return nil
	}
	m, ok := flat[0].(map[string]any)
	if !ok {
		return nil
	}
	return m
}
