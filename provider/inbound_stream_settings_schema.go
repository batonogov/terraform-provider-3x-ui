package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// Models
// ---------------------------------------------------------------------------

type InboundStreamSettingsModel struct {
	Network             types.String                     `tfsdk:"network"`
	Security            types.String                     `tfsdk:"security"`
	ExternalProxy       []InboundExternalProxyModel      `tfsdk:"external_proxy"`
	RealitySettings     *InboundRealitySettingsModel     `tfsdk:"reality_settings"`
	TCPSettings         *InboundTCPSettingsModel         `tfsdk:"tcp_settings"`
	WSSettings          *InboundWSSettingsModel          `tfsdk:"ws_settings"`
	GRPCSettings        *InboundGRPCSettingsModel        `tfsdk:"grpc_settings"`
	HTTPUpgradeSettings *InboundHTTPUpgradeSettingsModel `tfsdk:"httpupgrade_settings"`
	XHTTPSettings       *InboundXHTTPSettingsModel       `tfsdk:"xhttp_settings"`
	KCPSettings         *InboundKCPSettingsModel         `tfsdk:"kcp_settings"`
	Sockopt             *InboundSockoptModel             `tfsdk:"sockopt"`
}

type InboundExternalProxyModel struct {
	Dest     types.String `tfsdk:"dest"`
	Port     types.Int64  `tfsdk:"port"`
	Remark   types.String `tfsdk:"remark"`
	ForceTLS types.String `tfsdk:"force_tls"`
}

type InboundRealitySettingsModel struct {
	Show        types.Bool                        `tfsdk:"show"`
	Xver        types.Int64                       `tfsdk:"xver"`
	Target      types.String                      `tfsdk:"target"`
	ServerNames types.List                        `tfsdk:"server_names"` // list of string
	PrivateKey  types.String                      `tfsdk:"private_key"`
	ShortIDs    types.List                        `tfsdk:"short_ids"` // list of string
	Mldsa65Seed types.String                      `tfsdk:"mldsa65_seed"`
	Settings    *InboundRealityInnerSettingsModel `tfsdk:"settings"`
}

type InboundRealityInnerSettingsModel struct {
	PublicKey     types.String `tfsdk:"public_key"`
	Fingerprint   types.String `tfsdk:"fingerprint"`
	ServerName    types.String `tfsdk:"server_name"`
	SpiderX       types.String `tfsdk:"spider_x"`
	Mldsa65Verify types.String `tfsdk:"mldsa65_verify"`
}

type InboundTCPSettingsModel struct {
	AcceptProxyProtocol types.Bool   `tfsdk:"accept_proxy_protocol"`
	HeaderType          types.String `tfsdk:"header_type"`
}

type InboundWSSettingsModel struct {
	Path    types.String `tfsdk:"path"`
	Headers types.Map    `tfsdk:"headers"` // map of string
}

type InboundGRPCSettingsModel struct {
	ServiceName         types.String `tfsdk:"service_name"`
	MultiMode           types.Bool   `tfsdk:"multi_mode"`
	IdleTimeout         types.Int64  `tfsdk:"idle_timeout"`
	HealthCheckTimeout  types.Int64  `tfsdk:"health_check_timeout"`
	PermitWithoutStream types.Bool   `tfsdk:"permit_without_stream"`
	InitialWindowsSize  types.Int64  `tfsdk:"initial_windows_size"`
}

type InboundHTTPUpgradeSettingsModel struct {
	Path types.String `tfsdk:"path"`
	Host types.String `tfsdk:"host"`
}

type InboundXHTTPSettingsModel struct {
	Path              types.String `tfsdk:"path"`
	Mode              types.String `tfsdk:"mode"`
	NoSSEHeader       types.Bool   `tfsdk:"no_sse_header"`
	KeepAliveInterval types.Int64  `tfsdk:"keep_alive_interval"`
}

type InboundKCPSettingsModel struct {
	MTU              types.Int64  `tfsdk:"mtu"`
	TTI              types.Int64  `tfsdk:"tti"`
	UplinkCapacity   types.Int64  `tfsdk:"uplink_capacity"`
	DownlinkCapacity types.Int64  `tfsdk:"downlink_capacity"`
	Congestion       types.Bool   `tfsdk:"congestion"`
	ReadBufferSize   types.Int64  `tfsdk:"read_buffer_size"`
	WriteBufferSize  types.Int64  `tfsdk:"write_buffer_size"`
	HeaderType       types.String `tfsdk:"header_type"`
}

type InboundSockoptModel struct {
	Mark                 types.Int64  `tfsdk:"mark"`
	TCPKeepAliveInterval types.Int64  `tfsdk:"tcp_keep_alive_interval"`
	TCPNoDelay           types.Bool   `tfsdk:"tcp_no_delay"`
	TFOEnable            types.Bool   `tfsdk:"tfo_enable"`
	Tproxy               types.String `tfsdk:"tproxy"`
	DomainStrategy       types.String `tfsdk:"domain_strategy"`
}

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

func inboundStreamSettingsBlockSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Stream (transport) settings for the inbound.",
		Attributes: map[string]schema.Attribute{
			"network": schema.StringAttribute{
				Optional: true, Computed: true,
				Description: "Transport network (tcp, ws, grpc, httpupgrade, xhttp, kcp).",
			},
			"security": schema.StringAttribute{
				Optional: true, Computed: true,
				Description: "Security type (none, reality, tls).",
			},
		},
		Blocks: map[string]schema.Block{
			"external_proxy": schema.ListNestedBlock{
				Description: "External proxy entries.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"dest": schema.StringAttribute{
							Optional: true, Computed: true,
						},
						"port": schema.Int64Attribute{
							Optional: true, Computed: true,
						},
						"remark": schema.StringAttribute{
							Optional: true, Computed: true,
						},
						"force_tls": schema.StringAttribute{
							Optional: true, Computed: true,
						},
					},
				},
			},
			"reality_settings": schema.SingleNestedBlock{
				Description: "Reality settings.",
				Attributes: map[string]schema.Attribute{
					"show": schema.BoolAttribute{
						Optional: true, Computed: true,
					},
					"xver": schema.Int64Attribute{
						Optional: true, Computed: true,
					},
					"target": schema.StringAttribute{
						Optional: true, Computed: true,
						Description: "Target server (e.g. 'google.com:443').",
					},
					"server_names": schema.ListAttribute{
						Optional: true, Computed: true,
						ElementType: types.StringType,
					},
					"private_key": schema.StringAttribute{
						Optional: true, Computed: true, Sensitive: true,
						Description: "Reality private key (auto-generated if empty).",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"short_ids": schema.ListAttribute{
						Optional: true, Computed: true,
						ElementType: types.StringType,
						Description: "Short IDs (auto-generated if empty).",
					},
					"mldsa65_seed": schema.StringAttribute{
						Optional: true, Computed: true,
					},
				},
				Blocks: map[string]schema.Block{
					"settings": schema.SingleNestedBlock{
						Description: "Reality inner settings (client-side).",
						Attributes: map[string]schema.Attribute{
							"public_key": schema.StringAttribute{
								Optional: true, Computed: true,
								Description: "Reality public key (auto-generated if empty).",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"fingerprint": schema.StringAttribute{
								Optional: true, Computed: true,
							},
							"server_name": schema.StringAttribute{
								Optional: true, Computed: true,
							},
							"spider_x": schema.StringAttribute{
								Optional: true, Computed: true,
							},
							"mldsa65_verify": schema.StringAttribute{
								Optional: true, Computed: true,
							},
						},
					},
				},
			},
			"tcp_settings": schema.SingleNestedBlock{
				Description: "TCP transport settings.",
				Attributes: map[string]schema.Attribute{
					"accept_proxy_protocol": schema.BoolAttribute{
						Optional: true, Computed: true,
					},
					"header_type": schema.StringAttribute{
						Optional: true, Computed: true,
						Description: "Header type (e.g. 'none', 'http').",
					},
				},
			},
			"ws_settings": schema.SingleNestedBlock{
				Description: "WebSocket transport settings.",
				Attributes: map[string]schema.Attribute{
					"path": schema.StringAttribute{
						Optional: true, Computed: true,
					},
					"headers": schema.MapAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						Description: "Custom headers as key-value pairs.",
					},
				},
			},
			"grpc_settings": schema.SingleNestedBlock{
				Description: "gRPC transport settings.",
				Attributes: map[string]schema.Attribute{
					"service_name": schema.StringAttribute{
						Optional: true, Computed: true,
					},
					"multi_mode": schema.BoolAttribute{
						Optional: true, Computed: true,
					},
					"idle_timeout": schema.Int64Attribute{
						Optional: true, Computed: true,
					},
					"health_check_timeout": schema.Int64Attribute{
						Optional: true, Computed: true,
					},
					"permit_without_stream": schema.BoolAttribute{
						Optional: true, Computed: true,
					},
					"initial_windows_size": schema.Int64Attribute{
						Optional: true, Computed: true,
					},
				},
			},
			"httpupgrade_settings": schema.SingleNestedBlock{
				Description: "HTTP Upgrade transport settings.",
				Attributes: map[string]schema.Attribute{
					"path": schema.StringAttribute{
						Optional: true, Computed: true,
					},
					"host": schema.StringAttribute{
						Optional: true, Computed: true,
					},
				},
			},
			"xhttp_settings": schema.SingleNestedBlock{
				Description: "XHTTP transport settings.",
				Attributes: map[string]schema.Attribute{
					"path": schema.StringAttribute{
						Optional: true, Computed: true,
					},
					"mode": schema.StringAttribute{
						Optional: true, Computed: true,
					},
					"no_sse_header": schema.BoolAttribute{
						Optional: true, Computed: true,
					},
					"keep_alive_interval": schema.Int64Attribute{
						Optional: true, Computed: true,
					},
				},
			},
			"kcp_settings": schema.SingleNestedBlock{
				Description: "mKCP transport settings.",
				Attributes: map[string]schema.Attribute{
					"mtu": schema.Int64Attribute{
						Optional: true, Computed: true,
					},
					"tti": schema.Int64Attribute{
						Optional: true, Computed: true,
					},
					"uplink_capacity": schema.Int64Attribute{
						Optional: true, Computed: true,
					},
					"downlink_capacity": schema.Int64Attribute{
						Optional: true, Computed: true,
					},
					"congestion": schema.BoolAttribute{
						Optional: true, Computed: true,
					},
					"read_buffer_size": schema.Int64Attribute{
						Optional: true, Computed: true,
					},
					"write_buffer_size": schema.Int64Attribute{
						Optional: true, Computed: true,
					},
					"header_type": schema.StringAttribute{
						Optional: true, Computed: true,
						Description: "Header type (e.g. 'none', 'srtp', 'utp', 'wechat-video', 'dtls', 'wireguard').",
					},
				},
			},
			"sockopt": schema.SingleNestedBlock{
				Description: "Socket options.",
				Attributes: map[string]schema.Attribute{
					"mark": schema.Int64Attribute{
						Optional: true, Computed: true,
					},
					"tcp_keep_alive_interval": schema.Int64Attribute{
						Optional: true, Computed: true,
					},
					"tcp_no_delay": schema.BoolAttribute{
						Optional: true, Computed: true,
					},
					"tfo_enable": schema.BoolAttribute{
						Optional: true, Computed: true,
					},
					"tproxy": schema.StringAttribute{
						Optional: true, Computed: true,
					},
					"domain_strategy": schema.StringAttribute{
						Optional: true, Computed: true,
					},
				},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Typed model -> untyped map (for buildStreamSettingsJSON)
// ---------------------------------------------------------------------------

func expandStreamSettingsFromModel(m *InboundStreamSettingsModel) map[string]any {
	if m == nil {
		return nil
	}
	out := map[string]any{}

	if !m.Network.IsNull() && !m.Network.IsUnknown() {
		out["network"] = m.Network.ValueString()
	}
	if !m.Security.IsNull() && !m.Security.IsUnknown() {
		out["security"] = m.Security.ValueString()
	}
	if len(m.ExternalProxy) > 0 {
		out["external_proxy"] = expandExternalProxyFromModel(m.ExternalProxy)
	}
	if m.RealitySettings != nil {
		if rs := expandRealitySettingsFromModel(m.RealitySettings); len(rs) > 0 {
			out["reality_settings"] = []any{rs}
		}
	}
	if m.TCPSettings != nil {
		if ts := expandTCPSettingsFromModel(m.TCPSettings); len(ts) > 0 {
			out["tcp_settings"] = []any{ts}
		}
	}
	if m.WSSettings != nil {
		if ws := expandWSSettingsFromModel(m.WSSettings); len(ws) > 0 {
			out["ws_settings"] = []any{ws}
		}
	}
	if m.GRPCSettings != nil {
		if gs := expandGRPCSettingsFromModel(m.GRPCSettings); len(gs) > 0 {
			out["grpc_settings"] = []any{gs}
		}
	}
	if m.HTTPUpgradeSettings != nil {
		if hu := expandHTTPUpgradeSettingsFromModel(m.HTTPUpgradeSettings); len(hu) > 0 {
			out["httpupgrade_settings"] = []any{hu}
		}
	}
	if m.XHTTPSettings != nil {
		if xh := expandXHTTPSettingsFromModel(m.XHTTPSettings); len(xh) > 0 {
			out["xhttp_settings"] = []any{xh}
		}
	}
	if m.KCPSettings != nil {
		if kcp := expandKCPSettingsFromModel(m.KCPSettings); len(kcp) > 0 {
			out["kcp_settings"] = []any{kcp}
		}
	}
	if m.Sockopt != nil {
		if so := expandSockoptFromModel(m.Sockopt); len(so) > 0 {
			out["sockopt"] = []any{so}
		}
	}

	return out
}

func expandExternalProxyFromModel(list []InboundExternalProxyModel) []any {
	out := make([]any, 0, len(list))
	for _, ep := range list {
		entry := map[string]any{}
		if !ep.Dest.IsNull() && !ep.Dest.IsUnknown() {
			entry["dest"] = ep.Dest.ValueString()
		}
		if !ep.Port.IsNull() && !ep.Port.IsUnknown() {
			entry["port"] = int(ep.Port.ValueInt64())
		}
		if !ep.Remark.IsNull() && !ep.Remark.IsUnknown() {
			entry["remark"] = ep.Remark.ValueString()
		}
		if !ep.ForceTLS.IsNull() && !ep.ForceTLS.IsUnknown() {
			entry["force_tls"] = ep.ForceTLS.ValueString()
		}
		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	return out
}

func expandRealitySettingsFromModel(m *InboundRealitySettingsModel) map[string]any {
	if m == nil {
		return nil
	}
	out := map[string]any{}
	if !m.Show.IsNull() && !m.Show.IsUnknown() {
		out["show"] = m.Show.ValueBool()
	}
	if !m.Xver.IsNull() && !m.Xver.IsUnknown() {
		out["xver"] = int(m.Xver.ValueInt64())
	}
	if !m.Target.IsNull() && !m.Target.IsUnknown() {
		out["target"] = m.Target.ValueString()
	}
	if !m.ServerNames.IsNull() && !m.ServerNames.IsUnknown() {
		out["server_names"] = typesListToAnySlice(m.ServerNames)
	}
	if !m.PrivateKey.IsNull() && !m.PrivateKey.IsUnknown() {
		out["private_key"] = m.PrivateKey.ValueString()
	}
	if !m.ShortIDs.IsNull() && !m.ShortIDs.IsUnknown() {
		out["short_ids"] = typesListToAnySlice(m.ShortIDs)
	}
	if !m.Mldsa65Seed.IsNull() && !m.Mldsa65Seed.IsUnknown() {
		out["mldsa65_seed"] = m.Mldsa65Seed.ValueString()
	}
	if m.Settings != nil {
		if s := expandRealityInnerSettingsFromModel(m.Settings); len(s) > 0 {
			out["settings"] = []any{s}
		}
	}
	return out
}

func expandRealityInnerSettingsFromModel(m *InboundRealityInnerSettingsModel) map[string]any {
	if m == nil {
		return nil
	}
	out := map[string]any{}
	if !m.PublicKey.IsNull() && !m.PublicKey.IsUnknown() {
		out["public_key"] = m.PublicKey.ValueString()
	}
	if !m.Fingerprint.IsNull() && !m.Fingerprint.IsUnknown() {
		out["fingerprint"] = m.Fingerprint.ValueString()
	}
	if !m.ServerName.IsNull() && !m.ServerName.IsUnknown() {
		out["server_name"] = m.ServerName.ValueString()
	}
	if !m.SpiderX.IsNull() && !m.SpiderX.IsUnknown() {
		out["spider_x"] = m.SpiderX.ValueString()
	}
	if !m.Mldsa65Verify.IsNull() && !m.Mldsa65Verify.IsUnknown() {
		out["mldsa65_verify"] = m.Mldsa65Verify.ValueString()
	}
	return out
}

func expandTCPSettingsFromModel(m *InboundTCPSettingsModel) map[string]any {
	if m == nil {
		return nil
	}
	out := map[string]any{}
	if !m.AcceptProxyProtocol.IsNull() && !m.AcceptProxyProtocol.IsUnknown() {
		out["accept_proxy_protocol"] = m.AcceptProxyProtocol.ValueBool()
	}
	if !m.HeaderType.IsNull() && !m.HeaderType.IsUnknown() {
		out["header_type"] = m.HeaderType.ValueString()
	}
	return out
}

func expandWSSettingsFromModel(m *InboundWSSettingsModel) map[string]any {
	if m == nil {
		return nil
	}
	out := map[string]any{}
	if !m.Path.IsNull() && !m.Path.IsUnknown() {
		out["path"] = m.Path.ValueString()
	}
	if !m.Headers.IsNull() && !m.Headers.IsUnknown() {
		headers := map[string]any{}
		for k, v := range m.Headers.Elements() {
			if sv, ok := v.(types.String); ok && !sv.IsNull() && !sv.IsUnknown() {
				headers[k] = sv.ValueString()
			}
		}
		if len(headers) > 0 {
			out["headers"] = headers
		}
	}
	return out
}

func expandGRPCSettingsFromModel(m *InboundGRPCSettingsModel) map[string]any {
	if m == nil {
		return nil
	}
	out := map[string]any{}
	if !m.ServiceName.IsNull() && !m.ServiceName.IsUnknown() {
		out["service_name"] = m.ServiceName.ValueString()
	}
	if !m.MultiMode.IsNull() && !m.MultiMode.IsUnknown() {
		out["multi_mode"] = m.MultiMode.ValueBool()
	}
	if !m.IdleTimeout.IsNull() && !m.IdleTimeout.IsUnknown() {
		out["idle_timeout"] = int(m.IdleTimeout.ValueInt64())
	}
	if !m.HealthCheckTimeout.IsNull() && !m.HealthCheckTimeout.IsUnknown() {
		out["health_check_timeout"] = int(m.HealthCheckTimeout.ValueInt64())
	}
	if !m.PermitWithoutStream.IsNull() && !m.PermitWithoutStream.IsUnknown() {
		out["permit_without_stream"] = m.PermitWithoutStream.ValueBool()
	}
	if !m.InitialWindowsSize.IsNull() && !m.InitialWindowsSize.IsUnknown() {
		out["initial_windows_size"] = int(m.InitialWindowsSize.ValueInt64())
	}
	return out
}

func expandHTTPUpgradeSettingsFromModel(m *InboundHTTPUpgradeSettingsModel) map[string]any {
	if m == nil {
		return nil
	}
	out := map[string]any{}
	if !m.Path.IsNull() && !m.Path.IsUnknown() {
		out["path"] = m.Path.ValueString()
	}
	if !m.Host.IsNull() && !m.Host.IsUnknown() {
		out["host"] = m.Host.ValueString()
	}
	return out
}

func expandXHTTPSettingsFromModel(m *InboundXHTTPSettingsModel) map[string]any {
	if m == nil {
		return nil
	}
	out := map[string]any{}
	if !m.Path.IsNull() && !m.Path.IsUnknown() {
		out["path"] = m.Path.ValueString()
	}
	if !m.Mode.IsNull() && !m.Mode.IsUnknown() {
		out["mode"] = m.Mode.ValueString()
	}
	if !m.NoSSEHeader.IsNull() && !m.NoSSEHeader.IsUnknown() {
		out["no_sse_header"] = m.NoSSEHeader.ValueBool()
	}
	if !m.KeepAliveInterval.IsNull() && !m.KeepAliveInterval.IsUnknown() {
		out["keep_alive_interval"] = int(m.KeepAliveInterval.ValueInt64())
	}
	return out
}

func expandKCPSettingsFromModel(m *InboundKCPSettingsModel) map[string]any {
	if m == nil {
		return nil
	}
	out := map[string]any{}
	if !m.MTU.IsNull() && !m.MTU.IsUnknown() {
		out["mtu"] = int(m.MTU.ValueInt64())
	}
	if !m.TTI.IsNull() && !m.TTI.IsUnknown() {
		out["tti"] = int(m.TTI.ValueInt64())
	}
	if !m.UplinkCapacity.IsNull() && !m.UplinkCapacity.IsUnknown() {
		out["uplink_capacity"] = int(m.UplinkCapacity.ValueInt64())
	}
	if !m.DownlinkCapacity.IsNull() && !m.DownlinkCapacity.IsUnknown() {
		out["downlink_capacity"] = int(m.DownlinkCapacity.ValueInt64())
	}
	if !m.Congestion.IsNull() && !m.Congestion.IsUnknown() {
		out["congestion"] = m.Congestion.ValueBool()
	}
	if !m.ReadBufferSize.IsNull() && !m.ReadBufferSize.IsUnknown() {
		out["read_buffer_size"] = int(m.ReadBufferSize.ValueInt64())
	}
	if !m.WriteBufferSize.IsNull() && !m.WriteBufferSize.IsUnknown() {
		out["write_buffer_size"] = int(m.WriteBufferSize.ValueInt64())
	}
	if !m.HeaderType.IsNull() && !m.HeaderType.IsUnknown() {
		out["header_type"] = m.HeaderType.ValueString()
	}
	return out
}

func expandSockoptFromModel(m *InboundSockoptModel) map[string]any {
	if m == nil {
		return nil
	}
	out := map[string]any{}
	if !m.Mark.IsNull() && !m.Mark.IsUnknown() {
		out["mark"] = int(m.Mark.ValueInt64())
	}
	if !m.TCPKeepAliveInterval.IsNull() && !m.TCPKeepAliveInterval.IsUnknown() {
		out["tcp_keep_alive_interval"] = int(m.TCPKeepAliveInterval.ValueInt64())
	}
	if !m.TCPNoDelay.IsNull() && !m.TCPNoDelay.IsUnknown() {
		out["tcp_no_delay"] = m.TCPNoDelay.ValueBool()
	}
	if !m.TFOEnable.IsNull() && !m.TFOEnable.IsUnknown() {
		out["tfo_enable"] = m.TFOEnable.ValueBool()
	}
	if !m.Tproxy.IsNull() && !m.Tproxy.IsUnknown() {
		out["tproxy"] = m.Tproxy.ValueString()
	}
	if !m.DomainStrategy.IsNull() && !m.DomainStrategy.IsUnknown() {
		out["domain_strategy"] = m.DomainStrategy.ValueString()
	}
	return out
}

// ---------------------------------------------------------------------------
// Untyped map -> typed model (from flattenStreamSettings output)
// ---------------------------------------------------------------------------

func flattenStreamSettingsToModel(data map[string]any) *InboundStreamSettingsModel {
	if len(data) == 0 {
		return nil
	}
	m := &InboundStreamSettingsModel{}

	if v, ok := data["network"].(string); ok {
		m.Network = types.StringValue(v)
	} else {
		m.Network = types.StringNull()
	}

	if v, ok := data["security"].(string); ok {
		m.Security = types.StringValue(v)
	} else {
		m.Security = types.StringNull()
	}

	if v, ok := data["external_proxy"].([]any); ok && len(v) > 0 {
		m.ExternalProxy = flattenExternalProxyToModel(v)
	}

	if v, ok := data["reality_settings"].([]any); ok && len(v) > 0 {
		if raw, ok := v[0].(map[string]any); ok {
			m.RealitySettings = flattenRealitySettingsToModel(raw)
		}
	}

	if v, ok := data["tcp_settings"].([]any); ok && len(v) > 0 {
		if raw, ok := v[0].(map[string]any); ok {
			m.TCPSettings = flattenTCPSettingsToModel(raw)
		}
	}

	if v, ok := data["ws_settings"].([]any); ok && len(v) > 0 {
		if raw, ok := v[0].(map[string]any); ok {
			m.WSSettings = flattenWSSettingsToModel(raw)
		}
	}

	if v, ok := data["grpc_settings"].([]any); ok && len(v) > 0 {
		if raw, ok := v[0].(map[string]any); ok {
			m.GRPCSettings = flattenGRPCSettingsToModel(raw)
		}
	}

	if v, ok := data["httpupgrade_settings"].([]any); ok && len(v) > 0 {
		if raw, ok := v[0].(map[string]any); ok {
			m.HTTPUpgradeSettings = flattenHTTPUpgradeSettingsToModel(raw)
		}
	}

	if v, ok := data["xhttp_settings"].([]any); ok && len(v) > 0 {
		if raw, ok := v[0].(map[string]any); ok {
			m.XHTTPSettings = flattenXHTTPSettingsToModel(raw)
		}
	}

	if v, ok := data["kcp_settings"].([]any); ok && len(v) > 0 {
		if raw, ok := v[0].(map[string]any); ok {
			m.KCPSettings = flattenKCPSettingsToModel(raw)
		}
	}

	if v, ok := data["sockopt"].([]any); ok && len(v) > 0 {
		if raw, ok := v[0].(map[string]any); ok {
			m.Sockopt = flattenSockoptToModel(raw)
		}
	}

	return m
}

func flattenExternalProxyToModel(list []any) []InboundExternalProxyModel {
	out := make([]InboundExternalProxyModel, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		ep := InboundExternalProxyModel{}
		if v, ok := raw["dest"].(string); ok && v != "" {
			ep.Dest = types.StringValue(v)
		} else {
			ep.Dest = types.StringNull()
		}
		if v, ok := raw["port"]; ok {
			ep.Port = types.Int64Value(int64(intValue(v)))
		} else {
			ep.Port = types.Int64Null()
		}
		if v, ok := raw["remark"].(string); ok && v != "" {
			ep.Remark = types.StringValue(v)
		} else {
			ep.Remark = types.StringNull()
		}
		if v, ok := raw["force_tls"].(string); ok && v != "" {
			ep.ForceTLS = types.StringValue(v)
		} else {
			ep.ForceTLS = types.StringNull()
		}
		out = append(out, ep)
	}
	return out
}

func flattenRealitySettingsToModel(data map[string]any) *InboundRealitySettingsModel {
	if len(data) == 0 {
		return nil
	}
	m := &InboundRealitySettingsModel{}
	if v, ok := data["show"].(bool); ok {
		m.Show = types.BoolValue(v)
	} else {
		m.Show = types.BoolNull()
	}
	if v, ok := data["xver"]; ok {
		m.Xver = types.Int64Value(int64(intValue(v)))
	} else {
		m.Xver = types.Int64Null()
	}
	if v, ok := data["target"].(string); ok && v != "" {
		m.Target = types.StringValue(v)
	} else {
		m.Target = types.StringNull()
	}
	if v, ok := data["server_names"]; ok {
		m.ServerNames = anySliceToTypesList(v)
	} else {
		m.ServerNames = types.ListNull(types.StringType)
	}
	if v, ok := data["private_key"].(string); ok && v != "" {
		m.PrivateKey = types.StringValue(v)
	} else {
		m.PrivateKey = types.StringNull()
	}
	if v, ok := data["short_ids"]; ok {
		m.ShortIDs = anySliceToTypesList(v)
	} else {
		m.ShortIDs = types.ListNull(types.StringType)
	}
	if v, ok := data["mldsa65_seed"].(string); ok && v != "" {
		m.Mldsa65Seed = types.StringValue(v)
	} else {
		m.Mldsa65Seed = types.StringNull()
	}
	if v, ok := data["settings"].([]any); ok && len(v) > 0 {
		if raw, ok := v[0].(map[string]any); ok {
			m.Settings = flattenRealityInnerSettingsToModel(raw)
		}
	}
	return m
}

func flattenRealityInnerSettingsToModel(data map[string]any) *InboundRealityInnerSettingsModel {
	if len(data) == 0 {
		return nil
	}
	m := &InboundRealityInnerSettingsModel{}
	if v, ok := data["public_key"].(string); ok && v != "" {
		m.PublicKey = types.StringValue(v)
	} else {
		m.PublicKey = types.StringNull()
	}
	if v, ok := data["fingerprint"].(string); ok && v != "" {
		m.Fingerprint = types.StringValue(v)
	} else {
		m.Fingerprint = types.StringNull()
	}
	if v, ok := data["server_name"].(string); ok && v != "" {
		m.ServerName = types.StringValue(v)
	} else {
		m.ServerName = types.StringNull()
	}
	if v, ok := data["spider_x"].(string); ok && v != "" {
		m.SpiderX = types.StringValue(v)
	} else {
		m.SpiderX = types.StringNull()
	}
	if v, ok := data["mldsa65_verify"].(string); ok && v != "" {
		m.Mldsa65Verify = types.StringValue(v)
	} else {
		m.Mldsa65Verify = types.StringNull()
	}
	return m
}

func flattenTCPSettingsToModel(data map[string]any) *InboundTCPSettingsModel {
	if len(data) == 0 {
		return nil
	}
	m := &InboundTCPSettingsModel{}
	if v, ok := data["accept_proxy_protocol"].(bool); ok {
		m.AcceptProxyProtocol = types.BoolValue(v)
	} else {
		m.AcceptProxyProtocol = types.BoolNull()
	}
	// Flatten header: existing flattenTCPSettings returns header as []any{map[type:...]}
	if v, ok := data["header"].([]any); ok && len(v) > 0 {
		if h, ok := v[0].(map[string]any); ok {
			if t, ok := h["type"].(string); ok {
				m.HeaderType = types.StringValue(t)
			} else {
				m.HeaderType = types.StringNull()
			}
		}
	} else if v, ok := data["header_type"].(string); ok {
		m.HeaderType = types.StringValue(v)
	} else {
		m.HeaderType = types.StringNull()
	}
	return m
}

func flattenWSSettingsToModel(data map[string]any) *InboundWSSettingsModel {
	if len(data) == 0 {
		return nil
	}
	m := &InboundWSSettingsModel{}
	if v, ok := data["path"].(string); ok && v != "" {
		m.Path = types.StringValue(v)
	} else {
		m.Path = types.StringNull()
	}
	if v, ok := data["headers"].(map[string]any); ok && len(v) > 0 {
		m.Headers = anyMapToTypesMap(v)
	} else {
		m.Headers = types.MapNull(types.StringType)
	}
	return m
}

func flattenGRPCSettingsToModel(data map[string]any) *InboundGRPCSettingsModel {
	if len(data) == 0 {
		return nil
	}
	m := &InboundGRPCSettingsModel{}
	if v, ok := data["service_name"].(string); ok && v != "" {
		m.ServiceName = types.StringValue(v)
	} else {
		m.ServiceName = types.StringNull()
	}
	if v, ok := data["multi_mode"].(bool); ok {
		m.MultiMode = types.BoolValue(v)
	} else {
		m.MultiMode = types.BoolNull()
	}
	if v, ok := data["idle_timeout"]; ok {
		m.IdleTimeout = types.Int64Value(int64(intValue(v)))
	} else {
		m.IdleTimeout = types.Int64Null()
	}
	if v, ok := data["health_check_timeout"]; ok {
		m.HealthCheckTimeout = types.Int64Value(int64(intValue(v)))
	} else {
		m.HealthCheckTimeout = types.Int64Null()
	}
	if v, ok := data["permit_without_stream"].(bool); ok {
		m.PermitWithoutStream = types.BoolValue(v)
	} else {
		m.PermitWithoutStream = types.BoolNull()
	}
	if v, ok := data["initial_windows_size"]; ok {
		m.InitialWindowsSize = types.Int64Value(int64(intValue(v)))
	} else {
		m.InitialWindowsSize = types.Int64Null()
	}
	return m
}

func flattenHTTPUpgradeSettingsToModel(data map[string]any) *InboundHTTPUpgradeSettingsModel {
	if len(data) == 0 {
		return nil
	}
	m := &InboundHTTPUpgradeSettingsModel{}
	if v, ok := data["path"].(string); ok && v != "" {
		m.Path = types.StringValue(v)
	} else {
		m.Path = types.StringNull()
	}
	if v, ok := data["host"].(string); ok && v != "" {
		m.Host = types.StringValue(v)
	} else {
		m.Host = types.StringNull()
	}
	return m
}

func flattenXHTTPSettingsToModel(data map[string]any) *InboundXHTTPSettingsModel {
	if len(data) == 0 {
		return nil
	}
	m := &InboundXHTTPSettingsModel{}
	if v, ok := data["path"].(string); ok && v != "" {
		m.Path = types.StringValue(v)
	} else {
		m.Path = types.StringNull()
	}
	if v, ok := data["mode"].(string); ok && v != "" {
		m.Mode = types.StringValue(v)
	} else {
		m.Mode = types.StringNull()
	}
	if v, ok := data["no_sse_header"].(bool); ok {
		m.NoSSEHeader = types.BoolValue(v)
	} else {
		m.NoSSEHeader = types.BoolNull()
	}
	if v, ok := data["keep_alive_interval"]; ok {
		m.KeepAliveInterval = types.Int64Value(int64(intValue(v)))
	} else {
		m.KeepAliveInterval = types.Int64Null()
	}
	return m
}

func flattenKCPSettingsToModel(data map[string]any) *InboundKCPSettingsModel {
	if len(data) == 0 {
		return nil
	}
	m := &InboundKCPSettingsModel{}
	if v, ok := data["mtu"]; ok {
		m.MTU = types.Int64Value(int64(intValue(v)))
	} else {
		m.MTU = types.Int64Null()
	}
	if v, ok := data["tti"]; ok {
		m.TTI = types.Int64Value(int64(intValue(v)))
	} else {
		m.TTI = types.Int64Null()
	}
	if v, ok := data["uplink_capacity"]; ok {
		m.UplinkCapacity = types.Int64Value(int64(intValue(v)))
	} else {
		m.UplinkCapacity = types.Int64Null()
	}
	if v, ok := data["downlink_capacity"]; ok {
		m.DownlinkCapacity = types.Int64Value(int64(intValue(v)))
	} else {
		m.DownlinkCapacity = types.Int64Null()
	}
	if v, ok := data["congestion"].(bool); ok {
		m.Congestion = types.BoolValue(v)
	} else {
		m.Congestion = types.BoolNull()
	}
	if v, ok := data["read_buffer_size"]; ok {
		m.ReadBufferSize = types.Int64Value(int64(intValue(v)))
	} else {
		m.ReadBufferSize = types.Int64Null()
	}
	if v, ok := data["write_buffer_size"]; ok {
		m.WriteBufferSize = types.Int64Value(int64(intValue(v)))
	} else {
		m.WriteBufferSize = types.Int64Null()
	}
	if v, ok := data["header_type"].(string); ok && v != "" {
		m.HeaderType = types.StringValue(v)
	} else {
		m.HeaderType = types.StringNull()
	}
	return m
}

func flattenSockoptToModel(data map[string]any) *InboundSockoptModel {
	if len(data) == 0 {
		return nil
	}
	m := &InboundSockoptModel{}
	if v, ok := data["mark"]; ok {
		m.Mark = types.Int64Value(int64(intValue(v)))
	} else {
		m.Mark = types.Int64Null()
	}
	if v, ok := data["tcp_keep_alive_interval"]; ok {
		m.TCPKeepAliveInterval = types.Int64Value(int64(intValue(v)))
	} else {
		m.TCPKeepAliveInterval = types.Int64Null()
	}
	if v, ok := data["tcp_no_delay"].(bool); ok {
		m.TCPNoDelay = types.BoolValue(v)
	} else {
		m.TCPNoDelay = types.BoolNull()
	}
	if v, ok := data["tfo_enable"].(bool); ok {
		m.TFOEnable = types.BoolValue(v)
	} else {
		m.TFOEnable = types.BoolNull()
	}
	if v, ok := data["tproxy"].(string); ok && v != "" {
		m.Tproxy = types.StringValue(v)
	} else {
		m.Tproxy = types.StringNull()
	}
	if v, ok := data["domain_strategy"].(string); ok && v != "" {
		m.DomainStrategy = types.StringValue(v)
	} else {
		m.DomainStrategy = types.StringNull()
	}
	return m
}

// ---------------------------------------------------------------------------
// JSON string -> untyped map (wraps existing flattenStreamSettings)
// ---------------------------------------------------------------------------

func flattenStreamSettingsToMap(streamJSON string) map[string]any {
	flat := flattenStreamSettings(streamJSON)
	if len(flat) == 0 {
		return nil
	}
	m, ok := flat[0].(map[string]any)
	if !ok {
		return nil
	}
	return m
}

// ---------------------------------------------------------------------------
// Map helper: map[string]any -> types.Map of StringType
// ---------------------------------------------------------------------------

func anyMapToTypesMap(m map[string]any) types.Map {
	if len(m) == 0 {
		return types.MapNull(types.StringType)
	}
	elems := map[string]types.String{}
	for k, v := range m {
		if s, ok := v.(string); ok {
			elems[k] = types.StringValue(s)
		}
	}
	result, _ := types.MapValueFrom(nil, types.StringType, elems) //nolint:staticcheck // nil context ok for MapValueFrom
	return result
}
