package provider

// ---------------------------------------------------------------------------
// Freedom outbound: schema, expand (typed model -> untyped map),
// flatten (untyped map -> typed model), JSON expand/flatten
// ---------------------------------------------------------------------------

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- Schema ---

func freedomSettingsBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"domain_strategy": schema.StringAttribute{
					Optional: true, Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"redirect": schema.StringAttribute{
					Optional: true, Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"ips_blocked": schema.ListAttribute{
					Optional:    true,
					Computed:    true,
					ElementType: types.StringType,
					Description: "Deprecated legacy list of IPs/CIDRs to block (e.g. geoip:cn). Use final_rule on 3x-ui v2.9.4+.",
					PlanModifiers: []planmodifier.List{
						listplanmodifier.UseStateForUnknown(),
					},
					DeprecationMessage: "Deprecated. Use final_rule on 3x-ui v2.9.4+ instead.",
				},
			},
			Blocks: map[string]schema.Block{
				"fragment": singletonListNestedBlock(schema.ListNestedBlock{
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"packets": schema.StringAttribute{
								Optional: true, Computed: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"length": schema.StringAttribute{
								Optional: true, Computed: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"interval": schema.StringAttribute{
								Optional: true, Computed: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
						},
					},
				}),
				"noises": schema.ListNestedBlock{
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								Optional: true, Computed: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"packet": schema.StringAttribute{
								Optional: true, Computed: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"delay": schema.StringAttribute{
								Optional: true, Computed: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
						},
					},
				},
				"final_rule": schema.ListNestedBlock{
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"action": schema.StringAttribute{
								Optional: true, Computed: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"network": schema.StringAttribute{
								Optional: true, Computed: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"port": schema.StringAttribute{
								Optional: true, Computed: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"ip": schema.ListAttribute{
								Optional:    true,
								Computed:    true,
								ElementType: types.StringType,
								PlanModifiers: []planmodifier.List{
									listplanmodifier.UseStateForUnknown(),
								},
							},
							"block_delay": schema.StringAttribute{
								Optional: true, Computed: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
						},
					},
				},
			},
		},
	}
}

// --- Typed model -> untyped map ---

func expandFreedomSettingsFromModel(list []XrayFreedomSettings) []any {
	out := make([]any, 0, len(list))
	for _, fs := range list {
		entry := map[string]any{}
		if !fs.DomainStrategy.IsNull() && !fs.DomainStrategy.IsUnknown() {
			entry["domain_strategy"] = fs.DomainStrategy.ValueString()
		}
		if !fs.Redirect.IsNull() && !fs.Redirect.IsUnknown() {
			entry["redirect"] = fs.Redirect.ValueString()
		}
		if len(fs.Fragment) > 0 {
			frags := make([]any, 0, len(fs.Fragment))
			for _, f := range fs.Fragment {
				fEntry := map[string]any{}
				if !f.Packets.IsNull() && !f.Packets.IsUnknown() {
					fEntry["packets"] = f.Packets.ValueString()
				}
				if !f.Length.IsNull() && !f.Length.IsUnknown() {
					fEntry["length"] = f.Length.ValueString()
				}
				if !f.Interval.IsNull() && !f.Interval.IsUnknown() {
					fEntry["interval"] = f.Interval.ValueString()
				}
				if len(fEntry) > 0 {
					frags = append(frags, fEntry)
				}
			}
			entry["fragment"] = frags
		}
		if len(fs.Noises) > 0 {
			noises := make([]any, 0, len(fs.Noises))
			for _, n := range fs.Noises {
				nEntry := map[string]any{}
				if !n.Type.IsNull() && !n.Type.IsUnknown() {
					nEntry["type"] = n.Type.ValueString()
				}
				if !n.Packet.IsNull() && !n.Packet.IsUnknown() {
					nEntry["packet"] = n.Packet.ValueString()
				}
				if !n.Delay.IsNull() && !n.Delay.IsUnknown() {
					nEntry["delay"] = n.Delay.ValueString()
				}
				if len(nEntry) > 0 {
					noises = append(noises, nEntry)
				}
			}
			entry["noises"] = noises
		}
		if len(fs.FinalRules) > 0 {
			entry["final_rule"] = expandFreedomFinalRulesFromModel(fs.FinalRules)
		}
		if !fs.IPsBlocked.IsNull() && !fs.IPsBlocked.IsUnknown() {
			entry["ips_blocked"] = typesListToAnySlice(fs.IPsBlocked)
		}
		out = append(out, entry)
	}
	return out
}

func expandFreedomFinalRulesFromModel(list []XrayFreedomFinalRule) []any {
	out := make([]any, 0, len(list))
	for _, r := range list {
		entry := map[string]any{}
		if !r.Action.IsNull() && !r.Action.IsUnknown() {
			entry["action"] = r.Action.ValueString()
		}
		if !r.Network.IsNull() && !r.Network.IsUnknown() {
			entry["network"] = r.Network.ValueString()
		}
		if !r.Port.IsNull() && !r.Port.IsUnknown() {
			entry["port"] = r.Port.ValueString()
		}
		if !r.IP.IsNull() && !r.IP.IsUnknown() {
			entry["ip"] = typesListToAnySlice(r.IP)
		}
		if !r.BlockDelay.IsNull() && !r.BlockDelay.IsUnknown() {
			entry["block_delay"] = r.BlockDelay.ValueString()
		}
		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	return out
}

// --- Untyped map -> typed model ---

func flattenFreedomSettingsToModel(list []any) []XrayFreedomSettings {
	out := make([]XrayFreedomSettings, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		fs := XrayFreedomSettings{}

		if v, ok := raw["domain_strategy"].(string); ok && v != "" {
			fs.DomainStrategy = types.StringValue(v)
		} else {
			fs.DomainStrategy = types.StringNull()
		}

		if v, ok := raw["redirect"].(string); ok && v != "" {
			fs.Redirect = types.StringValue(v)
		} else {
			fs.Redirect = types.StringNull()
		}

		if v, ok := raw["fragment"].([]any); ok && len(v) > 0 {
			frags := make([]XrayFreedomFragment, 0, len(v))
			for _, fi := range v {
				fm, ok := fi.(map[string]any)
				if !ok {
					continue
				}
				f := XrayFreedomFragment{}
				if p, ok := fm["packets"].(string); ok && p != "" {
					f.Packets = types.StringValue(p)
				} else {
					f.Packets = types.StringNull()
				}
				if l, ok := fm["length"].(string); ok && l != "" {
					f.Length = types.StringValue(l)
				} else {
					f.Length = types.StringNull()
				}
				if i, ok := fm["interval"].(string); ok && i != "" {
					f.Interval = types.StringValue(i)
				} else {
					f.Interval = types.StringNull()
				}
				frags = append(frags, f)
			}
			fs.Fragment = frags
		}

		if v, ok := raw["noises"].([]any); ok && len(v) > 0 {
			noises := make([]XrayFreedomNoise, 0, len(v))
			for _, ni := range v {
				nm, ok := ni.(map[string]any)
				if !ok {
					continue
				}
				n := XrayFreedomNoise{}
				if t, ok := nm["type"].(string); ok && t != "" {
					n.Type = types.StringValue(t)
				} else {
					n.Type = types.StringNull()
				}
				if p, ok := nm["packet"].(string); ok && p != "" {
					n.Packet = types.StringValue(p)
				} else {
					n.Packet = types.StringNull()
				}
				if d, ok := nm["delay"].(string); ok && d != "" {
					n.Delay = types.StringValue(d)
				} else {
					n.Delay = types.StringNull()
				}
				noises = append(noises, n)
			}
			fs.Noises = noises
		}

		if v, ok := raw["final_rule"].([]any); ok && len(v) > 0 {
			fs.FinalRules = flattenFreedomFinalRulesToModel(v)
		}

		if v, ok := raw["ips_blocked"]; ok {
			fs.IPsBlocked = anySliceToTypesList(v)
		} else {
			fs.IPsBlocked = types.ListNull(types.StringType)
		}

		out = append(out, fs)
	}
	return out
}

func flattenFreedomFinalRulesToModel(list []any) []XrayFreedomFinalRule {
	out := make([]XrayFreedomFinalRule, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		rule := XrayFreedomFinalRule{}
		if v, ok := raw["action"].(string); ok && v != "" {
			rule.Action = types.StringValue(v)
		} else {
			rule.Action = types.StringNull()
		}
		if v, ok := raw["network"].(string); ok && v != "" {
			rule.Network = types.StringValue(v)
		} else {
			rule.Network = types.StringNull()
		}
		if v, ok := raw["port"].(string); ok && v != "" {
			rule.Port = types.StringValue(v)
		} else {
			rule.Port = types.StringNull()
		}
		rule.IP = anySliceToTypesList(raw["ip"])
		if v, ok := raw["block_delay"].(string); ok && v != "" {
			rule.BlockDelay = types.StringValue(v)
		} else {
			rule.BlockDelay = types.StringNull()
		}
		out = append(out, rule)
	}
	return out
}

// --- Untyped map -> Xray JSON ---

func expandFreedomSettings(m map[string]any) map[string]any {
	list, ok := m["freedom_settings"].([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["domain_strategy"].(string); ok && v != "" {
		out["domainStrategy"] = v
	}
	if v, ok := item["redirect"].(string); ok && v != "" {
		out["redirect"] = v
	}
	if v, ok := item["fragment"]; ok {
		if f := expandFreedomFragment(v.([]any)); f != nil {
			out["fragment"] = f
		}
	}
	if v, ok := item["noises"]; ok {
		if n := expandFreedomNoises(v.([]any)); n != nil {
			out["noises"] = n
		}
	}
	if v, ok := item["final_rule"]; ok {
		if rules, ok := v.([]any); ok && len(rules) > 0 {
			out["finalRules"] = expandFreedomFinalRules(rules)
		}
	}
	if v, ok := item["ips_blocked"]; ok {
		if list, ok := v.([]any); ok && len(list) > 0 {
			out["ipsBlocked"] = list
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func expandFreedomFinalRules(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		if v, ok := m["action"].(string); ok && v != "" {
			entry["action"] = v
		}
		if v, ok := m["network"].(string); ok && v != "" {
			entry["network"] = v
		}
		if v, ok := m["port"].(string); ok && v != "" {
			entry["port"] = v
		}
		if v, ok := m["ip"].([]any); ok && len(v) > 0 {
			entry["ip"] = expandStringList(v)
		}
		if v, ok := m["block_delay"].(string); ok && v != "" {
			entry["blockDelay"] = v
		}
		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	return out
}

func expandFreedomFragment(list []any) map[string]any {
	if len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["packets"].(string); ok && v != "" {
		out["packets"] = v
	}
	if v, ok := item["length"].(string); ok && v != "" {
		out["length"] = v
	}
	if v, ok := item["interval"].(string); ok && v != "" {
		out["interval"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func expandFreedomNoises(list []any) []any {
	if len(list) == 0 {
		return nil
	}
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
		if v, ok := m["packet"].(string); ok && v != "" {
			entry["packet"] = v
		}
		if v, ok := m["delay"].(string); ok && v != "" {
			entry["delay"] = v
		}
		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// --- Xray JSON -> untyped map ---

func flattenFreedomSettings(in map[string]any) map[string]any {
	out := map[string]any{}
	if v, ok := in["domainStrategy"].(string); ok {
		out["domain_strategy"] = v
	}
	if v, ok := in["redirect"].(string); ok {
		out["redirect"] = v
	}
	if v, ok := in["fragment"].(map[string]any); ok {
		f := map[string]any{}
		if p, ok := v["packets"].(string); ok {
			f["packets"] = p
		}
		if l, ok := v["length"].(string); ok {
			f["length"] = l
		}
		if i, ok := v["interval"].(string); ok {
			f["interval"] = i
		}
		out["fragment"] = []any{f}
	}
	if v, ok := in["noises"].([]any); ok {
		noises := make([]any, 0, len(v))
		for _, n := range v {
			nm, ok := n.(map[string]any)
			if !ok {
				continue
			}
			entry := map[string]any{}
			if t, ok := nm["type"].(string); ok {
				entry["type"] = t
			}
			if p, ok := nm["packet"].(string); ok {
				entry["packet"] = p
			}
			if d, ok := nm["delay"].(string); ok {
				entry["delay"] = d
			}
			noises = append(noises, entry)
		}
		out["noises"] = noises
	}
	if v, ok := in["finalRules"].([]any); ok && len(v) > 0 {
		out["final_rule"] = flattenFreedomFinalRules(v)
	}
	if v, ok := in["ipsBlocked"].([]any); ok && len(v) > 0 {
		out["ips_blocked"] = v
	}
	return out
}

func flattenFreedomFinalRules(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		if v, ok := m["action"].(string); ok {
			entry["action"] = v
		}
		if v, ok := m["network"].(string); ok {
			entry["network"] = v
		}
		if v, ok := m["port"].(string); ok {
			entry["port"] = v
		}
		if v, ok := m["ip"].([]any); ok {
			entry["ip"] = v
		}
		if v, ok := m["blockDelay"].(string); ok {
			entry["block_delay"] = v
		}
		out = append(out, entry)
	}
	return out
}
