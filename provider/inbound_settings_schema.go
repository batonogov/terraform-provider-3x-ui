package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// Per-protocol models
// ---------------------------------------------------------------------------

type InboundVlessSettingsModel struct {
	Decryption   types.String           `tfsdk:"decryption"`
	Encryption   types.String           `tfsdk:"encryption"`
	SelectedAuth types.String           `tfsdk:"selected_auth"`
	Fallback     []InboundFallbackModel `tfsdk:"fallback"`
}

type InboundTrojanSettingsModel struct {
	Fallback []InboundFallbackModel `tfsdk:"fallback"`
}

type InboundShadowsocksSettingsModel struct {
	Method   types.String `tfsdk:"method"`
	Password types.String `tfsdk:"password"`
	Network  types.String `tfsdk:"network"`
	IVCheck  types.Bool   `tfsdk:"iv_check"`
}

type InboundHTTPSettingsModel struct {
	Auth             types.String          `tfsdk:"auth"`
	AllowTransparent types.Bool            `tfsdk:"allow_transparent"`
	Account          []InboundAccountModel `tfsdk:"account"`
}

type InboundSocksSettingsModel struct {
	Auth    types.String          `tfsdk:"auth"`
	UDP     types.Bool            `tfsdk:"udp"`
	IP      types.String          `tfsdk:"ip"`
	Account []InboundAccountModel `tfsdk:"account"`
}

type InboundWireguardSettingsModel struct {
	MTU         types.Int64                 `tfsdk:"mtu"`
	SecretKey   types.String                `tfsdk:"secret_key"`
	NoKernelTun types.Bool                  `tfsdk:"no_kernel_tun"`
	Peer        []InboundWireguardPeerModel `tfsdk:"peer"`
}

type InboundDokodemoSettingsModel struct {
	Address        types.String `tfsdk:"address"`
	Port           types.Int64  `tfsdk:"port"`
	PortMap        types.Map    `tfsdk:"port_map"` // map of string
	Network        types.String `tfsdk:"network"`
	FollowRedirect types.Bool   `tfsdk:"follow_redirect"`
}

// Sub-models

type InboundAccountModel struct {
	User types.String `tfsdk:"user"`
	Pass types.String `tfsdk:"pass"`
}

type InboundFallbackModel struct {
	Name types.String `tfsdk:"name"`
	Alpn types.String `tfsdk:"alpn"`
	Path types.String `tfsdk:"path"`
	Dest types.String `tfsdk:"dest"`
	Xver types.Int64  `tfsdk:"xver"`
}

type InboundWireguardPeerModel struct {
	PrivateKey   types.String `tfsdk:"private_key"`
	PublicKey    types.String `tfsdk:"public_key"`
	PreSharedKey types.String `tfsdk:"pre_shared_key"`
	AllowedIPs   types.List   `tfsdk:"allowed_ips"` // list of string
	KeepAlive    types.Int64  `tfsdk:"keep_alive"`
}

// ---------------------------------------------------------------------------
// Schema blocks
// ---------------------------------------------------------------------------

func inboundSettingsBlockSchemas() map[string]schema.Block {
	return map[string]schema.Block{
		"vless_settings": schema.SingleNestedBlock{
			Description: "Settings for VLESS protocol.",
			Attributes: map[string]schema.Attribute{
				"decryption": schema.StringAttribute{
					Optional: true, Computed: true,
					Description: "Decryption method (usually 'none').",
				},
				"encryption": schema.StringAttribute{
					Optional: true, Computed: true,
					Description: "Encryption method.",
				},
				"selected_auth": schema.StringAttribute{
					Optional: true, Computed: true,
					Description: "Selected authentication type.",
				},
			},
			Blocks: map[string]schema.Block{
				"fallback": schema.ListNestedBlock{
					Description: "Fallback destinations.",
					NestedObject: schema.NestedBlockObject{
						Attributes: inboundFallbackAttributes(),
					},
				},
			},
		},
		"trojan_settings": schema.SingleNestedBlock{
			Description: "Settings for Trojan protocol.",
			Blocks: map[string]schema.Block{
				"fallback": schema.ListNestedBlock{
					Description: "Fallback destinations.",
					NestedObject: schema.NestedBlockObject{
						Attributes: inboundFallbackAttributes(),
					},
				},
			},
		},
		"shadowsocks_settings": schema.SingleNestedBlock{
			Description: "Settings for Shadowsocks protocol.",
			Attributes: map[string]schema.Attribute{
				"method": schema.StringAttribute{
					Optional: true, Computed: true,
					Description: "Encryption method (e.g. aes-256-gcm, chacha20-ietf-poly1305).",
				},
				"password": schema.StringAttribute{
					Optional: true, Computed: true, Sensitive: true,
					Description: "Password for encryption.",
				},
				"network": schema.StringAttribute{
					Optional: true, Computed: true,
					Description: "Network type (e.g. 'tcp,udp').",
				},
				"iv_check": schema.BoolAttribute{
					Optional: true, Computed: true,
					Description: "Enable IV check.",
				},
			},
		},
		"http_settings": schema.SingleNestedBlock{
			Description: "Settings for HTTP proxy protocol.",
			Attributes: map[string]schema.Attribute{
				"auth": schema.StringAttribute{
					Optional: true, Computed: true,
					Description: "Authentication type (e.g. 'password', 'noauth').",
				},
				"allow_transparent": schema.BoolAttribute{
					Optional: true, Computed: true,
					Description: "Allow transparent proxy.",
				},
			},
			Blocks: map[string]schema.Block{
				"account": schema.ListNestedBlock{
					Description: "User accounts for authentication.",
					NestedObject: schema.NestedBlockObject{
						Attributes: inboundAccountAttributes(),
					},
				},
			},
		},
		"socks_settings": schema.SingleNestedBlock{
			Description: "Settings for SOCKS proxy protocol.",
			Attributes: map[string]schema.Attribute{
				"auth": schema.StringAttribute{
					Optional: true, Computed: true,
					Description: "Authentication type (e.g. 'password', 'noauth').",
				},
				"udp": schema.BoolAttribute{
					Optional: true, Computed: true,
					Description: "Enable UDP support.",
				},
				"ip": schema.StringAttribute{
					Optional: true, Computed: true,
					Description: "IP address for UDP.",
				},
			},
			Blocks: map[string]schema.Block{
				"account": schema.ListNestedBlock{
					Description: "User accounts for authentication.",
					NestedObject: schema.NestedBlockObject{
						Attributes: inboundAccountAttributes(),
					},
				},
			},
		},
		"wireguard_settings": schema.SingleNestedBlock{
			Description: "Settings for WireGuard protocol.",
			Attributes: map[string]schema.Attribute{
				"mtu": schema.Int64Attribute{
					Optional: true, Computed: true,
					Description: "MTU size.",
				},
				"secret_key": schema.StringAttribute{
					Optional: true, Computed: true, Sensitive: true,
					Description: "WireGuard secret key.",
				},
				"no_kernel_tun": schema.BoolAttribute{
					Optional: true, Computed: true,
					Description: "Disable kernel TUN.",
				},
			},
			Blocks: map[string]schema.Block{
				"peer": schema.ListNestedBlock{
					Description: "WireGuard peers.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"private_key": schema.StringAttribute{
								Optional: true, Computed: true, Sensitive: true,
							},
							"public_key": schema.StringAttribute{
								Optional: true, Computed: true,
							},
							"pre_shared_key": schema.StringAttribute{
								Optional: true, Computed: true, Sensitive: true,
							},
							"allowed_ips": schema.ListAttribute{
								Optional:    true,
								Computed:    true,
								ElementType: types.StringType,
							},
							"keep_alive": schema.Int64Attribute{
								Optional: true, Computed: true,
							},
						},
					},
				},
			},
		},
		"dokodemo_settings": schema.SingleNestedBlock{
			Description: "Settings for Dokodemo-door / tunnel protocol.",
			Attributes: map[string]schema.Attribute{
				"address": schema.StringAttribute{
					Optional: true, Computed: true,
					Description: "Destination address.",
				},
				"port": schema.Int64Attribute{
					Optional: true, Computed: true,
					Description: "Destination port.",
				},
				"port_map": schema.MapAttribute{
					Optional:    true,
					Computed:    true,
					ElementType: types.StringType,
					Description: "Port mapping (e.g. {\"80\": \"127.0.0.1:8080\"}).",
				},
				"network": schema.StringAttribute{
					Optional: true, Computed: true,
					Description: "Network type (e.g. 'tcp', 'udp', 'tcp,udp').",
				},
				"follow_redirect": schema.BoolAttribute{
					Optional: true, Computed: true,
					Description: "Follow redirect.",
				},
			},
		},
	}
}

func inboundFallbackAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Optional: true, Computed: true,
		},
		"alpn": schema.StringAttribute{
			Optional: true, Computed: true,
		},
		"path": schema.StringAttribute{
			Optional: true, Computed: true,
		},
		"dest": schema.StringAttribute{
			Optional: true, Computed: true,
		},
		"xver": schema.Int64Attribute{
			Optional: true, Computed: true,
		},
	}
}

func inboundAccountAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"user": schema.StringAttribute{
			Optional: true, Computed: true,
		},
		"pass": schema.StringAttribute{
			Optional: true, Computed: true, Sensitive: true,
		},
	}
}

// ---------------------------------------------------------------------------
// Typed model -> untyped map (for buildSettingsJSON)
// ---------------------------------------------------------------------------

func expandSettingsFromModel(protocol string, m *InboundResourceModel) map[string]any {
	switch protocol {
	case "vless":
		return expandVlessInboundSettings(m.VlessSettings)
	case "trojan":
		return expandTrojanInboundSettings(m.TrojanSettings)
	case "shadowsocks":
		return expandShadowsocksInboundSettings(m.ShadowsocksSettings)
	case "http":
		return expandHTTPInboundSettings(m.HTTPSettings)
	case "socks":
		return expandSocksInboundSettings(m.SocksSettings)
	case "wireguard":
		return expandWireguardInboundSettings(m.WireguardSettings)
	case "dokodemo-door", "tunnel":
		return expandDokodemoInboundSettings(m.DokodemoSettings)
	default:
		return nil
	}
}

func expandVlessInboundSettings(m *InboundVlessSettingsModel) map[string]any {
	if m == nil {
		return nil
	}
	out := map[string]any{}
	if !m.Decryption.IsNull() && !m.Decryption.IsUnknown() {
		out["decryption"] = m.Decryption.ValueString()
	}
	if !m.Encryption.IsNull() && !m.Encryption.IsUnknown() {
		out["encryption"] = m.Encryption.ValueString()
	}
	if !m.SelectedAuth.IsNull() && !m.SelectedAuth.IsUnknown() {
		out["selected_auth"] = m.SelectedAuth.ValueString()
	}
	if len(m.Fallback) > 0 {
		out["fallbacks"] = expandFallbacksFromModel(m.Fallback)
	}
	return out
}

func expandTrojanInboundSettings(m *InboundTrojanSettingsModel) map[string]any {
	if m == nil {
		return nil
	}
	out := map[string]any{}
	if len(m.Fallback) > 0 {
		out["fallbacks"] = expandFallbacksFromModel(m.Fallback)
	}
	return out
}

func expandShadowsocksInboundSettings(m *InboundShadowsocksSettingsModel) map[string]any {
	if m == nil {
		return nil
	}
	out := map[string]any{}
	if !m.Method.IsNull() && !m.Method.IsUnknown() {
		out["method"] = m.Method.ValueString()
	}
	if !m.Password.IsNull() && !m.Password.IsUnknown() {
		out["password"] = m.Password.ValueString()
	}
	if !m.Network.IsNull() && !m.Network.IsUnknown() {
		out["network"] = m.Network.ValueString()
	}
	if !m.IVCheck.IsNull() && !m.IVCheck.IsUnknown() {
		out["iv_check"] = m.IVCheck.ValueBool()
	}
	return out
}

func expandHTTPInboundSettings(m *InboundHTTPSettingsModel) map[string]any {
	if m == nil {
		return nil
	}
	out := map[string]any{}
	if !m.Auth.IsNull() && !m.Auth.IsUnknown() {
		out["auth"] = m.Auth.ValueString()
	}
	if !m.AllowTransparent.IsNull() && !m.AllowTransparent.IsUnknown() {
		out["allow_transparent"] = m.AllowTransparent.ValueBool()
	}
	if len(m.Account) > 0 {
		out["accounts"] = expandAccountsFromModel(m.Account)
	}
	return out
}

func expandSocksInboundSettings(m *InboundSocksSettingsModel) map[string]any {
	if m == nil {
		return nil
	}
	out := map[string]any{}
	if !m.Auth.IsNull() && !m.Auth.IsUnknown() {
		out["auth"] = m.Auth.ValueString()
	}
	if !m.UDP.IsNull() && !m.UDP.IsUnknown() {
		out["udp"] = m.UDP.ValueBool()
	}
	if !m.IP.IsNull() && !m.IP.IsUnknown() {
		out["ip"] = m.IP.ValueString()
	}
	if len(m.Account) > 0 {
		out["accounts"] = expandAccountsFromModel(m.Account)
	}
	return out
}

func expandWireguardInboundSettings(m *InboundWireguardSettingsModel) map[string]any {
	if m == nil {
		return nil
	}
	out := map[string]any{}
	if !m.MTU.IsNull() && !m.MTU.IsUnknown() {
		out["mtu"] = int(m.MTU.ValueInt64())
	}
	if !m.SecretKey.IsNull() && !m.SecretKey.IsUnknown() {
		out["secret_key"] = m.SecretKey.ValueString()
	}
	if !m.NoKernelTun.IsNull() && !m.NoKernelTun.IsUnknown() {
		out["no_kernel_tun"] = m.NoKernelTun.ValueBool()
	}
	if len(m.Peer) > 0 {
		out["peers"] = expandWireguardPeersFromModel(m.Peer)
	}
	return out
}

func expandDokodemoInboundSettings(m *InboundDokodemoSettingsModel) map[string]any {
	if m == nil {
		return nil
	}
	out := map[string]any{}
	if !m.Address.IsNull() && !m.Address.IsUnknown() {
		out["address"] = m.Address.ValueString()
	}
	if !m.Port.IsNull() && !m.Port.IsUnknown() {
		out["port"] = int(m.Port.ValueInt64())
	}
	if !m.PortMap.IsNull() && !m.PortMap.IsUnknown() {
		pm := map[string]any{}
		for k, v := range m.PortMap.Elements() {
			if sv, ok := v.(types.String); ok {
				pm[k] = sv.ValueString()
			}
		}
		if len(pm) > 0 {
			out["port_map"] = pm
		}
	}
	if !m.Network.IsNull() && !m.Network.IsUnknown() {
		out["network"] = m.Network.ValueString()
	}
	if !m.FollowRedirect.IsNull() && !m.FollowRedirect.IsUnknown() {
		out["follow_redirect"] = m.FollowRedirect.ValueBool()
	}
	return out
}

func expandFallbacksFromModel(list []InboundFallbackModel) []any {
	out := make([]any, 0, len(list))
	for _, fb := range list {
		entry := map[string]any{}
		if !fb.Name.IsNull() && !fb.Name.IsUnknown() {
			entry["name"] = fb.Name.ValueString()
		}
		if !fb.Alpn.IsNull() && !fb.Alpn.IsUnknown() {
			entry["alpn"] = fb.Alpn.ValueString()
		}
		if !fb.Path.IsNull() && !fb.Path.IsUnknown() {
			entry["path"] = fb.Path.ValueString()
		}
		if !fb.Dest.IsNull() && !fb.Dest.IsUnknown() {
			entry["dest"] = fb.Dest.ValueString()
		}
		if !fb.Xver.IsNull() && !fb.Xver.IsUnknown() {
			entry["xver"] = int(fb.Xver.ValueInt64())
		}
		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	return out
}

func expandAccountsFromModel(list []InboundAccountModel) []any {
	out := make([]any, 0, len(list))
	for _, acc := range list {
		entry := map[string]any{}
		if !acc.User.IsNull() && !acc.User.IsUnknown() {
			entry["user"] = acc.User.ValueString()
		}
		if !acc.Pass.IsNull() && !acc.Pass.IsUnknown() {
			entry["pass"] = acc.Pass.ValueString()
		}
		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	return out
}

func expandWireguardPeersFromModel(list []InboundWireguardPeerModel) []any {
	out := make([]any, 0, len(list))
	for _, p := range list {
		entry := map[string]any{}
		if !p.PrivateKey.IsNull() && !p.PrivateKey.IsUnknown() {
			entry["private_key"] = p.PrivateKey.ValueString()
		}
		if !p.PublicKey.IsNull() && !p.PublicKey.IsUnknown() {
			entry["public_key"] = p.PublicKey.ValueString()
		}
		if !p.PreSharedKey.IsNull() && !p.PreSharedKey.IsUnknown() {
			entry["pre_shared_key"] = p.PreSharedKey.ValueString()
		}
		if !p.AllowedIPs.IsNull() && !p.AllowedIPs.IsUnknown() {
			entry["allowed_ips"] = typesListToAnySlice(p.AllowedIPs)
		}
		if !p.KeepAlive.IsNull() && !p.KeepAlive.IsUnknown() {
			entry["keep_alive"] = int(p.KeepAlive.ValueInt64())
		}
		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Untyped map -> typed model (from flattenSettings output)
// ---------------------------------------------------------------------------

func flattenSettingsToModel(protocol string, data map[string]any, m *InboundResourceModel) {
	switch protocol {
	case "vless":
		m.VlessSettings = flattenVlessInboundSettings(data)
	case "trojan":
		m.TrojanSettings = flattenTrojanInboundSettings(data)
	case "shadowsocks":
		m.ShadowsocksSettings = flattenShadowsocksInboundSettings(data)
	case "http":
		m.HTTPSettings = flattenHTTPInboundSettings(data)
	case "socks":
		m.SocksSettings = flattenSocksInboundSettings(data)
	case "wireguard":
		m.WireguardSettings = flattenWireguardInboundSettings(data)
	case "dokodemo-door", "tunnel":
		m.DokodemoSettings = flattenDokodemoInboundSettings(data)
	}
}

func flattenVlessInboundSettings(data map[string]any) *InboundVlessSettingsModel {
	if len(data) == 0 {
		return nil
	}
	m := &InboundVlessSettingsModel{}
	if v, ok := data["decryption"].(string); ok {
		m.Decryption = types.StringValue(v)
	} else {
		m.Decryption = types.StringNull()
	}
	if v, ok := data["encryption"].(string); ok {
		m.Encryption = types.StringValue(v)
	} else {
		m.Encryption = types.StringNull()
	}
	if v, ok := data["selected_auth"].(string); ok && v != "" {
		m.SelectedAuth = types.StringValue(v)
	} else {
		m.SelectedAuth = types.StringNull()
	}
	if v, ok := data["fallbacks"].([]any); ok && len(v) > 0 {
		m.Fallback = flattenFallbacksToModel(v)
	}
	return m
}

func flattenTrojanInboundSettings(data map[string]any) *InboundTrojanSettingsModel {
	if len(data) == 0 {
		return nil
	}
	m := &InboundTrojanSettingsModel{}
	if v, ok := data["fallbacks"].([]any); ok && len(v) > 0 {
		m.Fallback = flattenFallbacksToModel(v)
	}
	return m
}

func flattenShadowsocksInboundSettings(data map[string]any) *InboundShadowsocksSettingsModel {
	if len(data) == 0 {
		return nil
	}
	m := &InboundShadowsocksSettingsModel{}
	if v, ok := data["method"].(string); ok {
		m.Method = types.StringValue(v)
	} else {
		m.Method = types.StringNull()
	}
	if v, ok := data["password"].(string); ok {
		m.Password = types.StringValue(v)
	} else {
		m.Password = types.StringNull()
	}
	if v, ok := data["network"].(string); ok {
		m.Network = types.StringValue(v)
	} else {
		m.Network = types.StringNull()
	}
	if v, ok := data["iv_check"].(bool); ok {
		m.IVCheck = types.BoolValue(v)
	} else {
		m.IVCheck = types.BoolNull()
	}
	return m
}

func flattenHTTPInboundSettings(data map[string]any) *InboundHTTPSettingsModel {
	if len(data) == 0 {
		return nil
	}
	m := &InboundHTTPSettingsModel{}
	if v, ok := data["auth"].(string); ok {
		m.Auth = types.StringValue(v)
	} else {
		m.Auth = types.StringNull()
	}
	if v, ok := data["allow_transparent"].(bool); ok {
		m.AllowTransparent = types.BoolValue(v)
	} else {
		m.AllowTransparent = types.BoolNull()
	}
	if v, ok := data["accounts"].([]any); ok && len(v) > 0 {
		m.Account = flattenAccountsToModel(v)
	}
	return m
}

func flattenSocksInboundSettings(data map[string]any) *InboundSocksSettingsModel {
	if len(data) == 0 {
		return nil
	}
	m := &InboundSocksSettingsModel{}
	if v, ok := data["auth"].(string); ok {
		m.Auth = types.StringValue(v)
	} else {
		m.Auth = types.StringNull()
	}
	if v, ok := data["udp"].(bool); ok {
		m.UDP = types.BoolValue(v)
	} else {
		m.UDP = types.BoolNull()
	}
	if v, ok := data["ip"].(string); ok && v != "" {
		m.IP = types.StringValue(v)
	} else {
		m.IP = types.StringNull()
	}
	if v, ok := data["accounts"].([]any); ok && len(v) > 0 {
		m.Account = flattenAccountsToModel(v)
	}
	return m
}

func flattenWireguardInboundSettings(data map[string]any) *InboundWireguardSettingsModel {
	if len(data) == 0 {
		return nil
	}
	m := &InboundWireguardSettingsModel{}
	if v, ok := data["mtu"]; ok {
		m.MTU = types.Int64Value(int64(intValue(v)))
	} else {
		m.MTU = types.Int64Null()
	}
	if v, ok := data["secret_key"].(string); ok && v != "" {
		m.SecretKey = types.StringValue(v)
	} else {
		m.SecretKey = types.StringNull()
	}
	if v, ok := data["no_kernel_tun"].(bool); ok {
		m.NoKernelTun = types.BoolValue(v)
	} else {
		m.NoKernelTun = types.BoolNull()
	}
	if v, ok := data["peers"].([]any); ok && len(v) > 0 {
		m.Peer = flattenInboundWireguardPeersToModel(v)
	}
	return m
}

func flattenDokodemoInboundSettings(data map[string]any) *InboundDokodemoSettingsModel {
	if len(data) == 0 {
		return nil
	}
	m := &InboundDokodemoSettingsModel{}
	if v, ok := data["address"].(string); ok && v != "" {
		m.Address = types.StringValue(v)
	} else {
		m.Address = types.StringNull()
	}
	if v, ok := data["port"]; ok {
		m.Port = types.Int64Value(int64(intValue(v)))
	} else {
		m.Port = types.Int64Null()
	}
	if v, ok := data["port_map"]; ok {
		switch pm := v.(type) {
		case map[string]any:
			m.PortMap = anyMapToTypesMap(pm)
		case map[string]string:
			elems := make(map[string]any, len(pm))
			for k, s := range pm {
				elems[k] = s
			}
			m.PortMap = anyMapToTypesMap(elems)
		default:
			m.PortMap = types.MapNull(types.StringType)
		}
	} else {
		m.PortMap = types.MapNull(types.StringType)
	}
	if v, ok := data["network"].(string); ok && v != "" {
		m.Network = types.StringValue(v)
	} else {
		m.Network = types.StringNull()
	}
	if v, ok := data["follow_redirect"].(bool); ok {
		m.FollowRedirect = types.BoolValue(v)
	} else {
		m.FollowRedirect = types.BoolNull()
	}
	return m
}

func flattenFallbacksToModel(list []any) []InboundFallbackModel {
	out := make([]InboundFallbackModel, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		fb := InboundFallbackModel{}
		if v, ok := raw["name"].(string); ok && v != "" {
			fb.Name = types.StringValue(v)
		} else {
			fb.Name = types.StringNull()
		}
		if v, ok := raw["alpn"].(string); ok && v != "" {
			fb.Alpn = types.StringValue(v)
		} else {
			fb.Alpn = types.StringNull()
		}
		if v, ok := raw["path"].(string); ok && v != "" {
			fb.Path = types.StringValue(v)
		} else {
			fb.Path = types.StringNull()
		}
		if v, ok := raw["dest"].(string); ok && v != "" {
			fb.Dest = types.StringValue(v)
		} else {
			fb.Dest = types.StringNull()
		}
		if v, ok := raw["xver"]; ok {
			fb.Xver = types.Int64Value(int64(intValue(v)))
		} else {
			fb.Xver = types.Int64Null()
		}
		out = append(out, fb)
	}
	return out
}

func flattenAccountsToModel(list []any) []InboundAccountModel {
	out := make([]InboundAccountModel, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		acc := InboundAccountModel{}
		if v, ok := raw["user"].(string); ok && v != "" {
			acc.User = types.StringValue(v)
		} else {
			acc.User = types.StringNull()
		}
		if v, ok := raw["pass"].(string); ok && v != "" {
			acc.Pass = types.StringValue(v)
		} else {
			acc.Pass = types.StringNull()
		}
		out = append(out, acc)
	}
	return out
}

func flattenInboundWireguardPeersToModel(list []any) []InboundWireguardPeerModel {
	out := make([]InboundWireguardPeerModel, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		p := InboundWireguardPeerModel{}
		if v, ok := raw["private_key"].(string); ok && v != "" {
			p.PrivateKey = types.StringValue(v)
		} else {
			p.PrivateKey = types.StringNull()
		}
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
		if v, ok := raw["keep_alive"]; ok {
			p.KeepAlive = types.Int64Value(int64(intValue(v)))
		} else {
			p.KeepAlive = types.Int64Null()
		}
		out = append(out, p)
	}
	return out
}

// ---------------------------------------------------------------------------
// JSON string -> untyped map (wraps existing flattenSettings)
// ---------------------------------------------------------------------------

func flattenSettingsToMap(settingsJSON string) (map[string]any, error) {
	flat, err := flattenSettings(settingsJSON)
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
