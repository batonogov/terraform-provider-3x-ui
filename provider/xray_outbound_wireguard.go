package provider

// ---------------------------------------------------------------------------
// Wireguard outbound: schema, expand (typed model -> untyped map),
// flatten (untyped map -> typed model), JSON expand/flatten
// ---------------------------------------------------------------------------

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- Schema ---

func wireguardSettingsBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"mtu": schema.Int64Attribute{
					Optional: true, Computed: true,
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.UseStateForUnknown(),
					},
				},
				"secret_key": schema.StringAttribute{
					Optional: true, Computed: true, Sensitive: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"address": schema.ListAttribute{
					Optional:    true,
					Computed:    true,
					ElementType: types.StringType,
					PlanModifiers: []planmodifier.List{
						listplanmodifier.UseStateForUnknown(),
					},
				},
				"workers": schema.Int64Attribute{
					Optional: true, Computed: true,
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.UseStateForUnknown(),
					},
				},
				"domain_strategy": schema.StringAttribute{
					Optional: true, Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"reserved": schema.ListAttribute{
					Optional:    true,
					Computed:    true,
					ElementType: types.Int64Type,
					PlanModifiers: []planmodifier.List{
						listplanmodifier.UseStateForUnknown(),
					},
				},
				"no_kernel_tun": schema.BoolAttribute{
					Optional: true, Computed: true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"peer": schema.ListNestedBlock{
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"public_key": schema.StringAttribute{
								Optional: true, Computed: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"pre_shared_key": schema.StringAttribute{
								Optional: true, Computed: true, Sensitive: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"allowed_ips": schema.ListAttribute{
								Optional:    true,
								Computed:    true,
								ElementType: types.StringType,
								PlanModifiers: []planmodifier.List{
									listplanmodifier.UseStateForUnknown(),
								},
							},
							"endpoint": schema.StringAttribute{
								Optional: true, Computed: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"keep_alive": schema.Int64Attribute{
								Optional: true, Computed: true,
								PlanModifiers: []planmodifier.Int64{
									int64planmodifier.UseStateForUnknown(),
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

func expandWireguardSettingsFromModel(list []XrayWireguardOutSettings) []any {
	out := make([]any, 0, len(list))
	for _, wg := range list {
		entry := map[string]any{}
		if !wg.MTU.IsNull() && !wg.MTU.IsUnknown() {
			entry["mtu"] = int(wg.MTU.ValueInt64())
		}
		if !wg.SecretKey.IsNull() && !wg.SecretKey.IsUnknown() {
			entry["secret_key"] = wg.SecretKey.ValueString()
		}
		if !wg.Address.IsNull() && !wg.Address.IsUnknown() {
			entry["address"] = typesListToAnySlice(wg.Address)
		}
		if !wg.Workers.IsNull() && !wg.Workers.IsUnknown() {
			entry["workers"] = int(wg.Workers.ValueInt64())
		}
		if !wg.DomainStrategy.IsNull() && !wg.DomainStrategy.IsUnknown() {
			entry["domain_strategy"] = wg.DomainStrategy.ValueString()
		}
		if !wg.Reserved.IsNull() && !wg.Reserved.IsUnknown() {
			entry["reserved"] = expandInt64List(wg.Reserved)
		}
		if !wg.NoKernelTun.IsNull() && !wg.NoKernelTun.IsUnknown() {
			entry["no_kernel_tun"] = wg.NoKernelTun.ValueBool()
		}
		if len(wg.Peer) > 0 {
			peers := make([]any, 0, len(wg.Peer))
			for _, p := range wg.Peer {
				pEntry := map[string]any{}
				if !p.PublicKey.IsNull() && !p.PublicKey.IsUnknown() {
					pEntry["public_key"] = p.PublicKey.ValueString()
				}
				if !p.PreSharedKey.IsNull() && !p.PreSharedKey.IsUnknown() {
					pEntry["pre_shared_key"] = p.PreSharedKey.ValueString()
				}
				if !p.AllowedIPs.IsNull() && !p.AllowedIPs.IsUnknown() {
					pEntry["allowed_ips"] = typesListToAnySlice(p.AllowedIPs)
				}
				if !p.Endpoint.IsNull() && !p.Endpoint.IsUnknown() {
					pEntry["endpoint"] = p.Endpoint.ValueString()
				}
				if !p.KeepAlive.IsNull() && !p.KeepAlive.IsUnknown() {
					pEntry["keep_alive"] = int(p.KeepAlive.ValueInt64())
				}
				if len(pEntry) > 0 {
					peers = append(peers, pEntry)
				}
			}
			entry["peer"] = peers
		}
		out = append(out, entry)
	}
	return out
}

// --- Untyped map -> typed model ---

func flattenWireguardSettingsToModel(list []any) []XrayWireguardOutSettings {
	out := make([]XrayWireguardOutSettings, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		wg := XrayWireguardOutSettings{}

		if v, ok := raw["mtu"]; ok {
			wg.MTU = types.Int64Value(int64(intValue(v)))
		} else {
			wg.MTU = types.Int64Null()
		}

		if v, ok := raw["secret_key"].(string); ok && v != "" {
			wg.SecretKey = types.StringValue(v)
		} else {
			wg.SecretKey = types.StringNull()
		}

		if v, ok := raw["address"]; ok {
			wg.Address = anySliceToTypesList(v)
		} else {
			wg.Address = types.ListNull(types.StringType)
		}

		if v, ok := raw["workers"]; ok {
			wg.Workers = types.Int64Value(int64(intValue(v)))
		} else {
			wg.Workers = types.Int64Null()
		}

		if v, ok := raw["domain_strategy"].(string); ok && v != "" {
			wg.DomainStrategy = types.StringValue(v)
		} else {
			wg.DomainStrategy = types.StringNull()
		}

		if v, ok := raw["reserved"]; ok {
			wg.Reserved = flattenToInt64List(v)
		} else {
			wg.Reserved = types.ListNull(types.Int64Type)
		}

		if v, ok := raw["no_kernel_tun"].(bool); ok {
			wg.NoKernelTun = types.BoolValue(v)
		} else {
			wg.NoKernelTun = types.BoolNull()
		}

		if v, ok := raw["peer"].([]any); ok && len(v) > 0 {
			wg.Peer = flattenWireguardPeersToModel(v)
		}

		out = append(out, wg)
	}
	return out
}

func flattenWireguardPeersToModel(list []any) []XrayWireguardPeer {
	out := make([]XrayWireguardPeer, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		p := XrayWireguardPeer{}

		if v, ok := raw["public_key"].(string); ok && v != "" {
			p.PublicKey = types.StringValue(v)
		} else {
			p.PublicKey = types.StringNull()
		}

		if v, ok := raw["pre_shared_key"].(string); ok && v != "" {
			p.PreSharedKey = types.StringValue(v)
		} else {
			p.PreSharedKey = types.StringNull()
		}

		if v, ok := raw["allowed_ips"]; ok {
			p.AllowedIPs = anySliceToTypesList(v)
		} else {
			p.AllowedIPs = types.ListNull(types.StringType)
		}

		if v, ok := raw["endpoint"].(string); ok && v != "" {
			p.Endpoint = types.StringValue(v)
		} else {
			p.Endpoint = types.StringNull()
		}

		if v, ok := raw["keep_alive"]; ok {
			p.KeepAlive = types.Int64Value(int64(intValue(v)))
		} else {
			p.KeepAlive = types.Int64Null()
		}

		out = append(out, p)
	}
	return out
}

// --- Untyped map -> Xray JSON ---

func expandWireguardOutSettings(m map[string]any) map[string]any {
	list, ok := m["wireguard_settings"].([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["mtu"].(int); ok && v != 0 {
		out["mtu"] = v
	}
	if v, ok := item["secret_key"].(string); ok && v != "" {
		out["secretKey"] = v
	}
	if v, ok := item["address"].([]any); ok && len(v) > 0 {
		out["address"] = expandStringList(v)
	}
	if v, ok := item["workers"].(int); ok && v != 0 {
		out["workers"] = v
	}
	if v, ok := item["domain_strategy"].(string); ok && v != "" {
		out["domainStrategy"] = v
	}
	if v, ok := item["reserved"].([]any); ok && len(v) > 0 {
		out["reserved"] = flattenIntList(v)
	}
	if v, ok := item["no_kernel_tun"].(bool); ok {
		out["noKernelTun"] = v
	}
	if v, ok := item["peer"]; ok {
		if peers := expandWireguardPeers(v.([]any)); peers != nil {
			out["peers"] = peers
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func expandWireguardPeers(list []any) []any {
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
		if v, ok := m["public_key"].(string); ok && v != "" {
			entry["publicKey"] = v
		}
		if v, ok := m["pre_shared_key"].(string); ok && v != "" {
			entry["preSharedKey"] = v
		}
		if v, ok := m["allowed_ips"].([]any); ok && len(v) > 0 {
			entry["allowedIPs"] = expandStringList(v)
		}
		if v, ok := m["endpoint"].(string); ok && v != "" {
			entry["endpoint"] = v
		}
		if v, ok := m["keep_alive"].(int); ok && v != 0 {
			entry["keepAlive"] = v
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

func flattenWireguardOutSettings(in map[string]any) map[string]any {
	out := map[string]any{}
	if v, ok := in["mtu"]; ok {
		out["mtu"] = intValue(v)
	}
	if v, ok := in["secretKey"].(string); ok {
		out["secret_key"] = v
	}
	if v, ok := in["address"].([]any); ok {
		out["address"] = v
	}
	if v, ok := in["workers"]; ok {
		out["workers"] = intValue(v)
	}
	if v, ok := in["domainStrategy"].(string); ok {
		out["domain_strategy"] = v
	}
	if v, ok := in["reserved"].([]any); ok {
		out["reserved"] = flattenIntList(v)
	}
	if v, ok := in["noKernelTun"].(bool); ok {
		out["no_kernel_tun"] = v
	}
	if v, ok := in["peers"].([]any); ok {
		out["peer"] = flattenWireguardPeers(v)
	}
	return out
}

func flattenWireguardPeers(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		if v, ok := m["publicKey"].(string); ok {
			entry["public_key"] = v
		}
		if v, ok := m["preSharedKey"].(string); ok {
			entry["pre_shared_key"] = v
		}
		if v, ok := m["allowedIPs"].([]any); ok {
			entry["allowed_ips"] = v
		}
		if v, ok := m["endpoint"].(string); ok {
			entry["endpoint"] = v
		}
		if v, ok := m["keepAlive"]; ok {
			entry["keep_alive"] = intValue(v)
		}
		out = append(out, entry)
	}
	return out
}
