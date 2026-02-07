package provider

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// Typed models
// ---------------------------------------------------------------------------

type XrayRoutingModel struct {
	ID             types.String      `tfsdk:"id"`
	DomainStrategy types.String      `tfsdk:"domain_strategy"`
	DomainMatcher  types.String      `tfsdk:"domain_matcher"`
	Rule           []XrayRoutingRule `tfsdk:"rule"`
}

type XrayRoutingRule struct {
	Type        types.String `tfsdk:"type"`
	Domain      types.List   `tfsdk:"domain"`
	IP          types.List   `tfsdk:"ip"`
	Port        types.String `tfsdk:"port"`
	SourcePort  types.String `tfsdk:"source_port"`
	Network     types.String `tfsdk:"network"`
	Source      types.List   `tfsdk:"source"`
	User        types.List   `tfsdk:"user"`
	InboundTag  types.List   `tfsdk:"inbound_tag"`
	Protocol    types.List   `tfsdk:"protocol"`
	Attrs       types.String `tfsdk:"attrs"`
	OutboundTag types.String `tfsdk:"outbound_tag"`
	BalancerTag types.String `tfsdk:"balancer_tag"`
}

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

func xrayRoutingSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain_strategy": schema.StringAttribute{
				Optional: true,
			},
			"domain_matcher": schema.StringAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"rule": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Optional: true,
						},
						"domain": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
						},
						"ip": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
						},
						"port": schema.StringAttribute{
							Optional: true,
						},
						"source_port": schema.StringAttribute{
							Optional: true,
						},
						"network": schema.StringAttribute{
							Optional: true,
						},
						"source": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
						},
						"user": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
						},
						"inbound_tag": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
						},
						"protocol": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
						},
						"attrs": schema.StringAttribute{
							Optional: true,
						},
						"outbound_tag": schema.StringAttribute{
							Optional: true,
						},
						"balancer_tag": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Typed model <-> untyped map conversion
// ---------------------------------------------------------------------------

// expandXrayRouting converts the typed model to the untyped map format that
// buildXrayRoutingJSON expects.
func expandXrayRouting(m *XrayRoutingModel) map[string]any {
	out := map[string]any{}

	if !m.DomainStrategy.IsNull() && !m.DomainStrategy.IsUnknown() {
		out["domain_strategy"] = m.DomainStrategy.ValueString()
	}
	if !m.DomainMatcher.IsNull() && !m.DomainMatcher.IsUnknown() {
		out["domain_matcher"] = m.DomainMatcher.ValueString()
	}

	if len(m.Rule) > 0 {
		rules := make([]any, 0, len(m.Rule))
		for _, r := range m.Rule {
			entry := map[string]any{}

			if !r.Type.IsNull() && !r.Type.IsUnknown() {
				entry["type"] = r.Type.ValueString()
			}
			if !r.Domain.IsNull() && !r.Domain.IsUnknown() {
				entry["domain"] = typesListToAnySlice(r.Domain)
			}
			if !r.IP.IsNull() && !r.IP.IsUnknown() {
				entry["ip"] = typesListToAnySlice(r.IP)
			}
			if !r.Port.IsNull() && !r.Port.IsUnknown() {
				entry["port"] = r.Port.ValueString()
			}
			if !r.SourcePort.IsNull() && !r.SourcePort.IsUnknown() {
				entry["source_port"] = r.SourcePort.ValueString()
			}
			if !r.Network.IsNull() && !r.Network.IsUnknown() {
				entry["network"] = r.Network.ValueString()
			}
			if !r.Source.IsNull() && !r.Source.IsUnknown() {
				entry["source"] = typesListToAnySlice(r.Source)
			}
			if !r.User.IsNull() && !r.User.IsUnknown() {
				entry["user"] = typesListToAnySlice(r.User)
			}
			if !r.InboundTag.IsNull() && !r.InboundTag.IsUnknown() {
				entry["inbound_tag"] = typesListToAnySlice(r.InboundTag)
			}
			if !r.Protocol.IsNull() && !r.Protocol.IsUnknown() {
				entry["protocol"] = typesListToAnySlice(r.Protocol)
			}
			if !r.Attrs.IsNull() && !r.Attrs.IsUnknown() {
				entry["attrs"] = r.Attrs.ValueString()
			}
			if !r.OutboundTag.IsNull() && !r.OutboundTag.IsUnknown() {
				entry["outbound_tag"] = r.OutboundTag.ValueString()
			}
			if !r.BalancerTag.IsNull() && !r.BalancerTag.IsUnknown() {
				entry["balancer_tag"] = r.BalancerTag.ValueString()
			}

			rules = append(rules, entry)
		}
		out["rule"] = rules
	}

	return out
}

// flattenXrayRouting converts the output of flattenXrayRoutingToMap back to
// the typed model. Input has keys like domain_strategy, domain_matcher, rule.
func flattenXrayRouting(data map[string]any) *XrayRoutingModel {
	m := &XrayRoutingModel{
		ID: types.StringValue("xray_routing"),
	}

	if v, ok := data["domain_strategy"].(string); ok && v != "" {
		m.DomainStrategy = types.StringValue(v)
	} else {
		m.DomainStrategy = types.StringNull()
	}

	if v, ok := data["domain_matcher"].(string); ok && v != "" {
		m.DomainMatcher = types.StringValue(v)
	} else {
		m.DomainMatcher = types.StringNull()
	}

	if v, ok := data["rule"].([]any); ok && len(v) > 0 {
		m.Rule = make([]XrayRoutingRule, 0, len(v))
		for _, item := range v {
			rm, ok := item.(map[string]any)
			if !ok {
				continue
			}
			rule := XrayRoutingRule{}

			if s, ok := rm["type"].(string); ok && s != "" {
				rule.Type = types.StringValue(s)
			} else {
				rule.Type = types.StringNull()
			}

			rule.Domain = anySliceToTypesList(rm["domain"])
			rule.IP = anySliceToTypesList(rm["ip"])

			if s, ok := rm["port"].(string); ok && s != "" {
				rule.Port = types.StringValue(s)
			} else {
				rule.Port = types.StringNull()
			}
			if s, ok := rm["source_port"].(string); ok && s != "" {
				rule.SourcePort = types.StringValue(s)
			} else {
				rule.SourcePort = types.StringNull()
			}
			if s, ok := rm["network"].(string); ok && s != "" {
				rule.Network = types.StringValue(s)
			} else {
				rule.Network = types.StringNull()
			}

			rule.Source = anySliceToTypesList(rm["source"])
			rule.User = anySliceToTypesList(rm["user"])
			rule.InboundTag = anySliceToTypesList(rm["inbound_tag"])
			rule.Protocol = anySliceToTypesList(rm["protocol"])

			if s, ok := rm["attrs"].(string); ok && s != "" {
				rule.Attrs = types.StringValue(s)
			} else {
				rule.Attrs = types.StringNull()
			}
			if s, ok := rm["outbound_tag"].(string); ok && s != "" {
				rule.OutboundTag = types.StringValue(s)
			} else {
				rule.OutboundTag = types.StringNull()
			}
			if s, ok := rm["balancer_tag"].(string); ok && s != "" {
				rule.BalancerTag = types.StringValue(s)
			} else {
				rule.BalancerTag = types.StringNull()
			}

			m.Rule = append(m.Rule, rule)
		}
	}

	return m
}

// ---------------------------------------------------------------------------
// List conversion helpers
// ---------------------------------------------------------------------------

// typesListToAnySlice converts a types.List of StringType to []any for the
// untyped map format used by buildXrayRoutingJSON / expandRoutingRules.
func typesListToAnySlice(l types.List) []any {
	elems := l.Elements()
	out := make([]any, 0, len(elems))
	for _, e := range elems {
		if sv, ok := e.(types.String); ok {
			out = append(out, sv.ValueString())
		}
	}
	return out
}

// anySliceToTypesList converts a []any of strings (from flattenRoutingRules)
// to a types.List of StringType. Returns types.ListNull if the slice is
// nil or empty.
func anySliceToTypesList(v any) types.List {
	slice, ok := v.([]any)
	if !ok || len(slice) == 0 {
		return types.ListNull(types.StringType)
	}
	elems := make([]attr.Value, 0, len(slice))
	for _, item := range slice {
		switch s := item.(type) {
		case string:
			elems = append(elems, types.StringValue(s))
		default:
			// best-effort: skip non-string entries
			continue
		}
	}
	if len(elems) == 0 {
		return types.ListNull(types.StringType)
	}
	return types.ListValueMust(types.StringType, elems)
}

// ---------------------------------------------------------------------------
// Existing build/flatten functions (untyped map <-> Xray JSON)
// ---------------------------------------------------------------------------

func buildXrayRoutingJSON(d map[string]any) any {
	payload := map[string]any{}

	if v, ok := d["domain_strategy"].(string); ok && v != "" {
		payload["domainStrategy"] = v
	}
	if v, ok := d["domain_matcher"].(string); ok && v != "" {
		payload["domainMatcher"] = v
	}
	if v, ok := d["rule"]; ok {
		if list, ok := v.([]any); ok {
			payload["rules"] = expandRoutingRules(list)
		}
	}

	return payload
}

func expandRoutingRules(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}

		if v, ok := m["type"].(string); ok && v != "" {
			entry["type"] = v
		}
		if v, ok := m["domain"].([]any); ok && len(v) > 0 {
			entry["domain"] = expandStringList(v)
		}
		if v, ok := m["ip"].([]any); ok && len(v) > 0 {
			entry["ip"] = expandStringList(v)
		}
		if v, ok := m["port"].(string); ok && v != "" {
			entry["port"] = v
		}
		if v, ok := m["source_port"].(string); ok && v != "" {
			entry["sourcePort"] = v
		}
		if v, ok := m["network"].(string); ok && v != "" {
			entry["network"] = v
		}
		if v, ok := m["source"].([]any); ok && len(v) > 0 {
			entry["source"] = expandStringList(v)
		}
		if v, ok := m["user"].([]any); ok && len(v) > 0 {
			entry["user"] = expandStringList(v)
		}
		if v, ok := m["inbound_tag"].([]any); ok && len(v) > 0 {
			entry["inboundTag"] = expandStringList(v)
		}
		if v, ok := m["protocol"].([]any); ok && len(v) > 0 {
			entry["protocol"] = expandStringList(v)
		}
		if v, ok := m["attrs"].(string); ok && v != "" {
			entry["attrs"] = v
		}
		if v, ok := m["outbound_tag"].(string); ok && v != "" {
			entry["outboundTag"] = v
		}
		if v, ok := m["balancer_tag"].(string); ok && v != "" {
			entry["balancerTag"] = v
		}

		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	return out
}

func flattenXrayRoutingToMap(data any) map[string]any {
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

	if v, ok := payload["domainStrategy"].(string); ok {
		out["domain_strategy"] = v
	}
	if v, ok := payload["domainMatcher"].(string); ok {
		out["domain_matcher"] = v
	}
	if v, ok := payload["rules"].([]any); ok {
		out["rule"] = flattenRoutingRules(v)
	}

	return out
}

func flattenRoutingRules(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}

		if v, ok := m["type"].(string); ok {
			entry["type"] = v
		}
		if v, ok := m["domain"].([]any); ok {
			entry["domain"] = v
		}
		if v, ok := m["ip"].([]any); ok {
			entry["ip"] = v
		}
		if v, ok := m["port"].(string); ok {
			entry["port"] = v
		}
		if v, ok := m["sourcePort"].(string); ok {
			entry["source_port"] = v
		}
		if v, ok := m["network"].(string); ok {
			entry["network"] = v
		}
		if v, ok := m["source"].([]any); ok {
			entry["source"] = v
		}
		if v, ok := m["user"].([]any); ok {
			entry["user"] = v
		}
		if v, ok := m["inboundTag"].([]any); ok {
			entry["inbound_tag"] = v
		}
		if v, ok := m["protocol"].([]any); ok {
			entry["protocol"] = v
		}
		if v, ok := m["attrs"].(string); ok {
			entry["attrs"] = v
		}
		if v, ok := m["outboundTag"].(string); ok {
			entry["outbound_tag"] = v
		}
		if v, ok := m["balancerTag"].(string); ok {
			entry["balancer_tag"] = v
		}

		out = append(out, entry)
	}
	return out
}
