package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

type InboundSniffingModel struct {
	Enabled         types.Bool `tfsdk:"enabled"`
	DestOverride    types.List `tfsdk:"dest_override"` // list of string
	MetadataOnly    types.Bool `tfsdk:"metadata_only"`
	RouteOnly       types.Bool `tfsdk:"route_only"`
	IpsExcluded     types.List `tfsdk:"ips_excluded"`     // list of string
	DomainsExcluded types.List `tfsdk:"domains_excluded"` // list of string
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
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"dest_override": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Destination override protocols (e.g. http, tls, quic, fakedns).",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"metadata_only": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Only sniff metadata.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"route_only": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Only use sniffing for routing.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"ips_excluded": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "IPs/CIDRs excluded from sniffing (e.g. geoip:private).",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"domains_excluded": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Domains excluded from sniffing (e.g. domain:example.com).",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
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
	if !m.IpsExcluded.IsNull() && !m.IpsExcluded.IsUnknown() {
		out["ips_excluded"] = typesListToAnySlice(m.IpsExcluded)
	}
	if !m.DomainsExcluded.IsNull() && !m.DomainsExcluded.IsUnknown() {
		out["domains_excluded"] = typesListToAnySlice(m.DomainsExcluded)
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

	if v, ok := data["ips_excluded"]; ok {
		m.IpsExcluded = anySliceToTypesList(v)
	} else {
		m.IpsExcluded = types.ListNull(types.StringType)
	}
	if v, ok := data["domains_excluded"]; ok {
		m.DomainsExcluded = anySliceToTypesList(v)
	} else {
		m.DomainsExcluded = types.ListNull(types.StringType)
	}

	return m
}

// ---------------------------------------------------------------------------
// JSON string -> untyped map (wraps existing flattenSniffing)
// ---------------------------------------------------------------------------

func flattenSniffingToMap(sniffingJSON string) (map[string]any, error) {
	flat, err := flattenSniffing(sniffingJSON)
	if err != nil {
		return nil, err
	}
	if len(flat) == 0 {
		return nil, nil
	}
	m, ok := flat[0].(map[string]any)
	if !ok {
		return nil, nil
	}
	return m, nil
}
