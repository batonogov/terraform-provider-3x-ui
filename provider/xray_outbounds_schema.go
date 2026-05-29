package provider

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// Typed models
// ---------------------------------------------------------------------------

type XrayOutboundsModel struct {
	ID       types.String        `tfsdk:"id"`
	Outbound []XrayOutboundEntry `tfsdk:"outbound"`
}

type XrayOutboundEntry struct {
	Tag                 types.String                 `tfsdk:"tag"`
	Protocol            types.String                 `tfsdk:"protocol"`
	SendThrough         types.String                 `tfsdk:"send_through"`
	Mux                 []XrayOutboundMux            `tfsdk:"mux"`
	FreedomSettings     []XrayFreedomSettings        `tfsdk:"freedom_settings"`
	BlackholeSettings   []XrayBlackholeSettings      `tfsdk:"blackhole_settings"`
	DNSSettings         []XrayOutboundDNSSettings    `tfsdk:"dns_settings"`
	VmessSettings       []XrayVmessOutSettings       `tfsdk:"vmess_settings"`
	VlessSettings       []XrayVlessOutSettings       `tfsdk:"vless_settings"`
	TrojanSettings      []XrayTrojanOutSettings      `tfsdk:"trojan_settings"`
	ShadowsocksSettings []XrayShadowsocksOutSettings `tfsdk:"shadowsocks_settings"`
	SocksSettings       []XraySocksOutSettings       `tfsdk:"socks_settings"`
	HTTPSettings        []XrayHTTPOutSettings        `tfsdk:"http_settings"`
	WireguardSettings   []XrayWireguardOutSettings   `tfsdk:"wireguard_settings"`
	HysteriaSettings    []XrayHysteriaOutSettings    `tfsdk:"hysteria_settings"`
}

type XrayOutboundMux struct {
	Enabled         types.Bool   `tfsdk:"enabled"`
	Concurrency     types.Int64  `tfsdk:"concurrency"`
	XudpConcurrency types.Int64  `tfsdk:"xudp_concurrency"`
	XudpProxyUDP443 types.String `tfsdk:"xudp_proxy_udp443"`
}

type XrayFreedomSettings struct {
	DomainStrategy types.String           `tfsdk:"domain_strategy"`
	Redirect       types.String           `tfsdk:"redirect"`
	Fragment       []XrayFreedomFragment  `tfsdk:"fragment"`
	Noises         []XrayFreedomNoise     `tfsdk:"noises"`
	FinalRules     []XrayFreedomFinalRule `tfsdk:"final_rule"`
	IPsBlocked     types.List             `tfsdk:"ips_blocked"` // list of string
}

type XrayFreedomFragment struct {
	Packets  types.String `tfsdk:"packets"`
	Length   types.String `tfsdk:"length"`
	Interval types.String `tfsdk:"interval"`
}

type XrayFreedomNoise struct {
	Type   types.String `tfsdk:"type"`
	Packet types.String `tfsdk:"packet"`
	Delay  types.String `tfsdk:"delay"`
}

type XrayFreedomFinalRule struct {
	Action     types.String `tfsdk:"action"`
	Network    types.String `tfsdk:"network"`
	Port       types.String `tfsdk:"port"`
	IP         types.List   `tfsdk:"ip"` // list of string
	BlockDelay types.String `tfsdk:"block_delay"`
}

type XrayBlackholeSettings struct {
	ResponseType types.String `tfsdk:"response_type"`
}

type XrayOutboundDNSSettings struct {
	Network    types.String `tfsdk:"network"`
	Address    types.String `tfsdk:"address"`
	Port       types.Int64  `tfsdk:"port"`
	NonIPQuery types.String `tfsdk:"non_ip_query"`
	BlockTypes types.List   `tfsdk:"block_types"` // list of int64
}

type XrayVmessOutSettings struct {
	Address  types.String `tfsdk:"address"`
	Port     types.Int64  `tfsdk:"port"`
	ID       types.String `tfsdk:"id"`
	Security types.String `tfsdk:"security"`
}

type XrayVlessOutSettings struct {
	Address    types.String `tfsdk:"address"`
	Port       types.Int64  `tfsdk:"port"`
	ID         types.String `tfsdk:"id"`
	Flow       types.String `tfsdk:"flow"`
	Encryption types.String `tfsdk:"encryption"`
	ReverseTag types.String `tfsdk:"reverse_tag"`
}

type XrayTrojanOutSettings struct {
	Address  types.String `tfsdk:"address"`
	Port     types.Int64  `tfsdk:"port"`
	Password types.String `tfsdk:"password"`
}

type XrayShadowsocksOutSettings struct {
	Address    types.String `tfsdk:"address"`
	Port       types.Int64  `tfsdk:"port"`
	Password   types.String `tfsdk:"password"`
	Method     types.String `tfsdk:"method"`
	UOT        types.Bool   `tfsdk:"uot"`
	UOTVersion types.Int64  `tfsdk:"uot_version"`
}

type XraySocksOutSettings struct {
	Address types.String `tfsdk:"address"`
	Port    types.Int64  `tfsdk:"port"`
	User    types.String `tfsdk:"user"`
	Pass    types.String `tfsdk:"pass"`
}

type XrayHTTPOutSettings struct {
	Address types.String `tfsdk:"address"`
	Port    types.Int64  `tfsdk:"port"`
	User    types.String `tfsdk:"user"`
	Pass    types.String `tfsdk:"pass"`
}

type XrayWireguardOutSettings struct {
	MTU            types.Int64         `tfsdk:"mtu"`
	SecretKey      types.String        `tfsdk:"secret_key"`
	Address        types.List          `tfsdk:"address"` // list of strings
	Workers        types.Int64         `tfsdk:"workers"`
	DomainStrategy types.String        `tfsdk:"domain_strategy"`
	Reserved       types.List          `tfsdk:"reserved"` // list of int64
	NoKernelTun    types.Bool          `tfsdk:"no_kernel_tun"`
	Peer           []XrayWireguardPeer `tfsdk:"peer"`
}

type XrayWireguardPeer struct {
	PublicKey    types.String `tfsdk:"public_key"`
	PreSharedKey types.String `tfsdk:"pre_shared_key"`
	AllowedIPs   types.List   `tfsdk:"allowed_ips"` // list of strings
	Endpoint     types.String `tfsdk:"endpoint"`
	KeepAlive    types.Int64  `tfsdk:"keep_alive"`
}

type XrayHysteriaOutSettings struct {
	Address types.String `tfsdk:"address"`
	Port    types.Int64  `tfsdk:"port"`
	Version types.Int64  `tfsdk:"version"`
}

// ---------------------------------------------------------------------------
// Int64 list helpers (local to this file)
// ---------------------------------------------------------------------------

// expandInt64List converts a types.List of Int64Type to []any ([]int) for the
// untyped map format.
func expandInt64List(l types.List) []any {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	elems := l.Elements()
	out := make([]any, 0, len(elems))
	for _, e := range elems {
		if iv, ok := e.(types.Int64); ok && !iv.IsNull() && !iv.IsUnknown() {
			out = append(out, int(iv.ValueInt64()))
		}
	}
	return out
}

// flattenToInt64List converts a []any of numbers (from the untyped map) to a
// types.List of Int64Type. Returns types.ListNull if the slice is nil or empty.
func flattenToInt64List(v any) types.List {
	slice, ok := v.([]any)
	if !ok || len(slice) == 0 {
		return types.ListNull(types.Int64Type)
	}
	elems := make([]attr.Value, 0, len(slice))
	for _, item := range slice {
		elems = append(elems, types.Int64Value(int64(intValue(item))))
	}
	if len(elems) == 0 {
		return types.ListNull(types.Int64Type)
	}
	return types.ListValueMust(types.Int64Type, elems)
}

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

func xrayOutboundsSchema() schema.Schema {
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
			"outbound": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"tag": schema.StringAttribute{
							Optional: true, Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"protocol": schema.StringAttribute{
							Required:    true,
							Description: "Outbound protocol (freedom, blackhole, dns, vmess, vless, trojan, shadowsocks, socks, http, wireguard, hysteria).",
						},
						"send_through": schema.StringAttribute{
							Optional: true, Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"mux": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"enabled": schema.BoolAttribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.UseStateForUnknown(),
										},
									},
									"concurrency": schema.Int64Attribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.UseStateForUnknown(),
										},
									},
									"xudp_concurrency": schema.Int64Attribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.UseStateForUnknown(),
										},
									},
									"xudp_proxy_udp443": schema.StringAttribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						},
						"freedom_settings": schema.ListNestedBlock{
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
									},
								},
								Blocks: map[string]schema.Block{
									"fragment": schema.ListNestedBlock{
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
									},
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
						},
						"blackhole_settings": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"response_type": schema.StringAttribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						},
						"dns_settings": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"network": schema.StringAttribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
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
									"non_ip_query": schema.StringAttribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"block_types": schema.ListAttribute{
										Optional:    true,
										Computed:    true,
										ElementType: types.Int64Type,
										PlanModifiers: []planmodifier.List{
											listplanmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						},
						"vmess_settings": schema.ListNestedBlock{
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
									"security": schema.StringAttribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						},
						"vless_settings": schema.ListNestedBlock{
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
						},
						"trojan_settings": schema.ListNestedBlock{
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
									"password": schema.StringAttribute{
										Optional: true, Computed: true, Sensitive: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						},
						"shadowsocks_settings": schema.ListNestedBlock{
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
									"password": schema.StringAttribute{
										Optional: true, Computed: true, Sensitive: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"method": schema.StringAttribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"uot": schema.BoolAttribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.UseStateForUnknown(),
										},
									},
									"uot_version": schema.Int64Attribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						},
						"socks_settings": schema.ListNestedBlock{
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
									"user": schema.StringAttribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"pass": schema.StringAttribute{
										Optional: true, Computed: true, Sensitive: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						},
						"http_settings": schema.ListNestedBlock{
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
									"user": schema.StringAttribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"pass": schema.StringAttribute{
										Optional: true, Computed: true, Sensitive: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						},
						"wireguard_settings": schema.ListNestedBlock{
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
						},
						"hysteria_settings": schema.ListNestedBlock{
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
						},
					},
				},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Typed model -> untyped map (for buildXrayOutboundsJSON)
// ---------------------------------------------------------------------------

// hasNonEmptyEntries returns true if at least one entry in the slice is a
// non-empty map. This prevents emitting keys whose expand produced only empty
// maps (e.g. when all typed fields are null/unknown).
func hasNonEmptyEntries(list []any) bool {
	for _, item := range list {
		if m, ok := item.(map[string]any); ok && len(m) > 0 {
			return true
		}
	}
	return false
}

// expandXrayOutbounds converts the typed model to the untyped map format that
// buildXrayOutboundsJSON expects.
func expandXrayOutbounds(m *XrayOutboundsModel) map[string]any {
	payload := map[string]any{}
	if m.Outbound == nil {
		return payload
	}

	outbounds := make([]any, 0, len(m.Outbound))
	for _, ob := range m.Outbound {
		entry := map[string]any{}

		if !ob.Tag.IsNull() && !ob.Tag.IsUnknown() {
			entry["tag"] = ob.Tag.ValueString()
		}
		if !ob.Protocol.IsNull() && !ob.Protocol.IsUnknown() {
			entry["protocol"] = ob.Protocol.ValueString()
		}
		if !ob.SendThrough.IsNull() && !ob.SendThrough.IsUnknown() {
			entry["send_through"] = ob.SendThrough.ValueString()
		}

		// Mux
		if len(ob.Mux) > 0 {
			if result := expandOutboundMuxFromModel(ob.Mux); hasNonEmptyEntries(result) {
				entry["mux"] = result
			}
		}

		// Protocol-specific settings
		if len(ob.FreedomSettings) > 0 {
			if result := expandFreedomSettingsFromModel(ob.FreedomSettings); hasNonEmptyEntries(result) {
				entry["freedom_settings"] = result
			}
		}
		if len(ob.BlackholeSettings) > 0 {
			if result := expandBlackholeSettingsFromModel(ob.BlackholeSettings); hasNonEmptyEntries(result) {
				entry["blackhole_settings"] = result
			}
		}
		if len(ob.DNSSettings) > 0 {
			if result := expandDNSSettingsFromModel(ob.DNSSettings); hasNonEmptyEntries(result) {
				entry["dns_settings"] = result
			}
		}
		if len(ob.VmessSettings) > 0 {
			if result := expandVmessSettingsFromModel(ob.VmessSettings); hasNonEmptyEntries(result) {
				entry["vmess_settings"] = result
			}
		}
		if len(ob.VlessSettings) > 0 {
			if result := expandVlessSettingsFromModel(ob.VlessSettings); hasNonEmptyEntries(result) {
				entry["vless_settings"] = result
			}
		}
		if len(ob.TrojanSettings) > 0 {
			if result := expandTrojanSettingsFromModel(ob.TrojanSettings); hasNonEmptyEntries(result) {
				entry["trojan_settings"] = result
			}
		}
		if len(ob.ShadowsocksSettings) > 0 {
			if result := expandShadowsocksSettingsFromModel(ob.ShadowsocksSettings); hasNonEmptyEntries(result) {
				entry["shadowsocks_settings"] = result
			}
		}
		if len(ob.SocksSettings) > 0 {
			if result := expandSocksSettingsFromModel(ob.SocksSettings); hasNonEmptyEntries(result) {
				entry["socks_settings"] = result
			}
		}
		if len(ob.HTTPSettings) > 0 {
			if result := expandHTTPSettingsFromModel(ob.HTTPSettings); hasNonEmptyEntries(result) {
				entry["http_settings"] = result
			}
		}
		if len(ob.WireguardSettings) > 0 {
			if result := expandWireguardSettingsFromModel(ob.WireguardSettings); hasNonEmptyEntries(result) {
				entry["wireguard_settings"] = result
			}
		}
		if len(ob.HysteriaSettings) > 0 {
			if result := expandHysteriaSettingsFromModel(ob.HysteriaSettings); hasNonEmptyEntries(result) {
				entry["hysteria_settings"] = result
			}
		}

		if len(entry) > 0 {
			outbounds = append(outbounds, entry)
		}
	}

	payload["outbound"] = outbounds
	return payload
}

func expandOutboundMuxFromModel(muxList []XrayOutboundMux) []any {
	out := make([]any, 0, len(muxList))
	for _, mux := range muxList {
		entry := map[string]any{}
		if !mux.Enabled.IsNull() && !mux.Enabled.IsUnknown() {
			entry["enabled"] = mux.Enabled.ValueBool()
		}
		if !mux.Concurrency.IsNull() && !mux.Concurrency.IsUnknown() {
			entry["concurrency"] = int(mux.Concurrency.ValueInt64())
		}
		if !mux.XudpConcurrency.IsNull() && !mux.XudpConcurrency.IsUnknown() {
			entry["xudp_concurrency"] = int(mux.XudpConcurrency.ValueInt64())
		}
		if !mux.XudpProxyUDP443.IsNull() && !mux.XudpProxyUDP443.IsUnknown() {
			entry["xudp_proxy_udp443"] = mux.XudpProxyUDP443.ValueString()
		}
		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	return out
}

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

func expandBlackholeSettingsFromModel(list []XrayBlackholeSettings) []any {
	out := make([]any, 0, len(list))
	for _, bh := range list {
		entry := map[string]any{}
		if !bh.ResponseType.IsNull() && !bh.ResponseType.IsUnknown() {
			entry["response_type"] = bh.ResponseType.ValueString()
		}
		out = append(out, entry)
	}
	return out
}

func expandDNSSettingsFromModel(list []XrayOutboundDNSSettings) []any {
	out := make([]any, 0, len(list))
	for _, ds := range list {
		entry := map[string]any{}
		if !ds.Network.IsNull() && !ds.Network.IsUnknown() {
			entry["network"] = ds.Network.ValueString()
		}
		if !ds.Address.IsNull() && !ds.Address.IsUnknown() {
			entry["address"] = ds.Address.ValueString()
		}
		if !ds.Port.IsNull() && !ds.Port.IsUnknown() {
			entry["port"] = int(ds.Port.ValueInt64())
		}
		if !ds.NonIPQuery.IsNull() && !ds.NonIPQuery.IsUnknown() {
			entry["non_ip_query"] = ds.NonIPQuery.ValueString()
		}
		if !ds.BlockTypes.IsNull() && !ds.BlockTypes.IsUnknown() {
			entry["block_types"] = expandInt64List(ds.BlockTypes)
		}
		out = append(out, entry)
	}
	return out
}

func expandVmessSettingsFromModel(list []XrayVmessOutSettings) []any {
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
		if !vs.Security.IsNull() && !vs.Security.IsUnknown() {
			entry["security"] = vs.Security.ValueString()
		}
		out = append(out, entry)
	}
	return out
}

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

func expandTrojanSettingsFromModel(list []XrayTrojanOutSettings) []any {
	out := make([]any, 0, len(list))
	for _, ts := range list {
		entry := map[string]any{}
		if !ts.Address.IsNull() && !ts.Address.IsUnknown() {
			entry["address"] = ts.Address.ValueString()
		}
		if !ts.Port.IsNull() && !ts.Port.IsUnknown() {
			entry["port"] = int(ts.Port.ValueInt64())
		}
		if !ts.Password.IsNull() && !ts.Password.IsUnknown() {
			entry["password"] = ts.Password.ValueString()
		}
		out = append(out, entry)
	}
	return out
}

func expandShadowsocksSettingsFromModel(list []XrayShadowsocksOutSettings) []any {
	out := make([]any, 0, len(list))
	for _, ss := range list {
		entry := map[string]any{}
		if !ss.Address.IsNull() && !ss.Address.IsUnknown() {
			entry["address"] = ss.Address.ValueString()
		}
		if !ss.Port.IsNull() && !ss.Port.IsUnknown() {
			entry["port"] = int(ss.Port.ValueInt64())
		}
		if !ss.Password.IsNull() && !ss.Password.IsUnknown() {
			entry["password"] = ss.Password.ValueString()
		}
		if !ss.Method.IsNull() && !ss.Method.IsUnknown() {
			entry["method"] = ss.Method.ValueString()
		}
		if !ss.UOT.IsNull() && !ss.UOT.IsUnknown() {
			entry["uot"] = ss.UOT.ValueBool()
		}
		if !ss.UOTVersion.IsNull() && !ss.UOTVersion.IsUnknown() {
			entry["uot_version"] = int(ss.UOTVersion.ValueInt64())
		}
		out = append(out, entry)
	}
	return out
}

func expandSocksSettingsFromModel(list []XraySocksOutSettings) []any {
	out := make([]any, 0, len(list))
	for _, ss := range list {
		entry := map[string]any{}
		if !ss.Address.IsNull() && !ss.Address.IsUnknown() {
			entry["address"] = ss.Address.ValueString()
		}
		if !ss.Port.IsNull() && !ss.Port.IsUnknown() {
			entry["port"] = int(ss.Port.ValueInt64())
		}
		if !ss.User.IsNull() && !ss.User.IsUnknown() {
			entry["user"] = ss.User.ValueString()
		}
		if !ss.Pass.IsNull() && !ss.Pass.IsUnknown() {
			entry["pass"] = ss.Pass.ValueString()
		}
		out = append(out, entry)
	}
	return out
}

func expandHTTPSettingsFromModel(list []XrayHTTPOutSettings) []any {
	out := make([]any, 0, len(list))
	for _, hs := range list {
		entry := map[string]any{}
		if !hs.Address.IsNull() && !hs.Address.IsUnknown() {
			entry["address"] = hs.Address.ValueString()
		}
		if !hs.Port.IsNull() && !hs.Port.IsUnknown() {
			entry["port"] = int(hs.Port.ValueInt64())
		}
		if !hs.User.IsNull() && !hs.User.IsUnknown() {
			entry["user"] = hs.User.ValueString()
		}
		if !hs.Pass.IsNull() && !hs.Pass.IsUnknown() {
			entry["pass"] = hs.Pass.ValueString()
		}
		out = append(out, entry)
	}
	return out
}

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

// ---------------------------------------------------------------------------
// Untyped map -> typed model (from flattenXrayOutboundsToMap output)
// ---------------------------------------------------------------------------

// flattenXrayOutbounds converts the output of flattenXrayOutboundsToMap back
// to the typed model.
func flattenXrayOutbounds(data map[string]any) *XrayOutboundsModel {
	m := &XrayOutboundsModel{
		ID: types.StringValue("xray_outbounds"),
	}

	v, ok := data["outbound"]
	if !ok {
		return m
	}

	list, ok := v.([]any)
	if !ok {
		return m
	}

	outbounds := make([]XrayOutboundEntry, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}

		entry := XrayOutboundEntry{}

		// Tag
		if v, ok := raw["tag"].(string); ok && v != "" {
			entry.Tag = types.StringValue(v)
		} else {
			entry.Tag = types.StringNull()
		}

		// Protocol
		if v, ok := raw["protocol"].(string); ok && v != "" {
			entry.Protocol = types.StringValue(v)
		} else {
			entry.Protocol = types.StringNull()
		}

		// SendThrough
		if v, ok := raw["send_through"].(string); ok && v != "" {
			entry.SendThrough = types.StringValue(v)
		} else {
			entry.SendThrough = types.StringNull()
		}

		// Mux
		if v, ok := raw["mux"].([]any); ok && len(v) > 0 {
			entry.Mux = flattenOutboundMuxToModel(v)
		}

		// Protocol-specific settings
		if v, ok := raw["freedom_settings"].([]any); ok && len(v) > 0 {
			entry.FreedomSettings = flattenFreedomSettingsToModel(v)
		}
		if v, ok := raw["blackhole_settings"].([]any); ok && len(v) > 0 {
			entry.BlackholeSettings = flattenBlackholeSettingsToModel(v)
		}
		if v, ok := raw["dns_settings"].([]any); ok && len(v) > 0 {
			entry.DNSSettings = flattenDNSSettingsToModel(v)
		}
		if v, ok := raw["vmess_settings"].([]any); ok && len(v) > 0 {
			entry.VmessSettings = flattenVmessSettingsToModel(v)
		}
		if v, ok := raw["vless_settings"].([]any); ok && len(v) > 0 {
			entry.VlessSettings = flattenVlessSettingsToModel(v)
		}
		if v, ok := raw["trojan_settings"].([]any); ok && len(v) > 0 {
			entry.TrojanSettings = flattenTrojanSettingsToModel(v)
		}
		if v, ok := raw["shadowsocks_settings"].([]any); ok && len(v) > 0 {
			entry.ShadowsocksSettings = flattenShadowsocksSettingsToModel(v)
		}
		if v, ok := raw["socks_settings"].([]any); ok && len(v) > 0 {
			entry.SocksSettings = flattenSocksSettingsToModel(v)
		}
		if v, ok := raw["http_settings"].([]any); ok && len(v) > 0 {
			entry.HTTPSettings = flattenHTTPSettingsToModel(v)
		}
		if v, ok := raw["wireguard_settings"].([]any); ok && len(v) > 0 {
			entry.WireguardSettings = flattenWireguardSettingsToModel(v)
		}
		if v, ok := raw["hysteria_settings"].([]any); ok && len(v) > 0 {
			entry.HysteriaSettings = flattenHysteriaSettingsToModel(v)
		}

		outbounds = append(outbounds, entry)
	}

	m.Outbound = outbounds
	return m
}

func flattenOutboundMuxToModel(list []any) []XrayOutboundMux {
	out := make([]XrayOutboundMux, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		mux := XrayOutboundMux{}

		if v, ok := raw["enabled"].(bool); ok {
			mux.Enabled = types.BoolValue(v)
		} else {
			mux.Enabled = types.BoolNull()
		}

		if v, ok := raw["concurrency"]; ok {
			mux.Concurrency = types.Int64Value(int64(intValue(v)))
		} else {
			mux.Concurrency = types.Int64Null()
		}

		if v, ok := raw["xudp_concurrency"]; ok {
			mux.XudpConcurrency = types.Int64Value(int64(intValue(v)))
		} else {
			mux.XudpConcurrency = types.Int64Null()
		}

		if v, ok := raw["xudp_proxy_udp443"].(string); ok && v != "" {
			mux.XudpProxyUDP443 = types.StringValue(v)
		} else {
			mux.XudpProxyUDP443 = types.StringNull()
		}

		out = append(out, mux)
	}
	return out
}

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

func flattenBlackholeSettingsToModel(list []any) []XrayBlackholeSettings {
	out := make([]XrayBlackholeSettings, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		bh := XrayBlackholeSettings{}
		if v, ok := raw["response_type"].(string); ok && v != "" {
			bh.ResponseType = types.StringValue(v)
		} else {
			bh.ResponseType = types.StringNull()
		}
		out = append(out, bh)
	}
	return out
}

func flattenDNSSettingsToModel(list []any) []XrayOutboundDNSSettings {
	out := make([]XrayOutboundDNSSettings, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		ds := XrayOutboundDNSSettings{}

		if v, ok := raw["network"].(string); ok && v != "" {
			ds.Network = types.StringValue(v)
		} else {
			ds.Network = types.StringNull()
		}

		if v, ok := raw["address"].(string); ok && v != "" {
			ds.Address = types.StringValue(v)
		} else {
			ds.Address = types.StringNull()
		}

		if v, ok := raw["port"]; ok {
			ds.Port = types.Int64Value(int64(intValue(v)))
		} else {
			ds.Port = types.Int64Null()
		}

		if v, ok := raw["non_ip_query"].(string); ok && v != "" {
			ds.NonIPQuery = types.StringValue(v)
		} else {
			ds.NonIPQuery = types.StringNull()
		}

		if v, ok := raw["block_types"]; ok {
			ds.BlockTypes = flattenToInt64List(v)
		} else {
			ds.BlockTypes = types.ListNull(types.Int64Type)
		}

		out = append(out, ds)
	}
	return out
}

func flattenVmessSettingsToModel(list []any) []XrayVmessOutSettings {
	out := make([]XrayVmessOutSettings, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		vs := XrayVmessOutSettings{}

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

		if v, ok := raw["security"].(string); ok && v != "" {
			vs.Security = types.StringValue(v)
		} else {
			vs.Security = types.StringNull()
		}

		out = append(out, vs)
	}
	return out
}

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

func flattenTrojanSettingsToModel(list []any) []XrayTrojanOutSettings {
	out := make([]XrayTrojanOutSettings, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		ts := XrayTrojanOutSettings{}

		if v, ok := raw["address"].(string); ok && v != "" {
			ts.Address = types.StringValue(v)
		} else {
			ts.Address = types.StringNull()
		}

		if v, ok := raw["port"]; ok {
			ts.Port = types.Int64Value(int64(intValue(v)))
		} else {
			ts.Port = types.Int64Null()
		}

		if v, ok := raw["password"].(string); ok && v != "" {
			ts.Password = types.StringValue(v)
		} else {
			ts.Password = types.StringNull()
		}

		out = append(out, ts)
	}
	return out
}

func flattenShadowsocksSettingsToModel(list []any) []XrayShadowsocksOutSettings {
	out := make([]XrayShadowsocksOutSettings, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		ss := XrayShadowsocksOutSettings{}

		if v, ok := raw["address"].(string); ok && v != "" {
			ss.Address = types.StringValue(v)
		} else {
			ss.Address = types.StringNull()
		}

		if v, ok := raw["port"]; ok {
			ss.Port = types.Int64Value(int64(intValue(v)))
		} else {
			ss.Port = types.Int64Null()
		}

		if v, ok := raw["password"].(string); ok && v != "" {
			ss.Password = types.StringValue(v)
		} else {
			ss.Password = types.StringNull()
		}

		if v, ok := raw["method"].(string); ok && v != "" {
			ss.Method = types.StringValue(v)
		} else {
			ss.Method = types.StringNull()
		}

		if v, ok := raw["uot"].(bool); ok {
			ss.UOT = types.BoolValue(v)
		} else {
			ss.UOT = types.BoolNull()
		}

		if v, ok := raw["uot_version"]; ok {
			ss.UOTVersion = types.Int64Value(int64(intValue(v)))
		} else {
			ss.UOTVersion = types.Int64Null()
		}

		out = append(out, ss)
	}
	return out
}

func flattenSocksSettingsToModel(list []any) []XraySocksOutSettings {
	out := make([]XraySocksOutSettings, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		ss := XraySocksOutSettings{}

		if v, ok := raw["address"].(string); ok && v != "" {
			ss.Address = types.StringValue(v)
		} else {
			ss.Address = types.StringNull()
		}

		if v, ok := raw["port"]; ok {
			ss.Port = types.Int64Value(int64(intValue(v)))
		} else {
			ss.Port = types.Int64Null()
		}

		if v, ok := raw["user"].(string); ok && v != "" {
			ss.User = types.StringValue(v)
		} else {
			ss.User = types.StringNull()
		}

		if v, ok := raw["pass"].(string); ok && v != "" {
			ss.Pass = types.StringValue(v)
		} else {
			ss.Pass = types.StringNull()
		}

		out = append(out, ss)
	}
	return out
}

func flattenHTTPSettingsToModel(list []any) []XrayHTTPOutSettings {
	out := make([]XrayHTTPOutSettings, 0, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}
		hs := XrayHTTPOutSettings{}

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

		if v, ok := raw["user"].(string); ok && v != "" {
			hs.User = types.StringValue(v)
		} else {
			hs.User = types.StringNull()
		}

		if v, ok := raw["pass"].(string); ok && v != "" {
			hs.Pass = types.StringValue(v)
		} else {
			hs.Pass = types.StringNull()
		}

		out = append(out, hs)
	}
	return out
}

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

// ---------------------------------------------------------------------------
// Existing build/flatten functions (untyped map <-> Xray JSON)
// ---------------------------------------------------------------------------

func buildXrayOutboundsJSON(d map[string]any) any {
	v, ok := d["outbound"]
	if !ok {
		return []any{}
	}
	list, ok := v.([]any)
	if !ok {
		return []any{}
	}
	return expandOutbounds(list)
}

func expandOutbounds(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		protocol := ""

		if v, ok := m["tag"].(string); ok && v != "" {
			entry["tag"] = v
		}
		if v, ok := m["protocol"].(string); ok && v != "" {
			protocol = v
			entry["protocol"] = v
		}
		if v, ok := m["send_through"].(string); ok && v != "" {
			entry["sendThrough"] = v
		}

		// Mux
		if v, ok := m["mux"]; ok {
			if mux := expandOutboundMux(v.([]any)); mux != nil {
				entry["mux"] = mux
			}
		}

		// Protocol-specific settings
		if settings := expandOutboundSettings(m, protocol); settings != nil {
			entry["settings"] = settings
		}

		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	return out
}

func expandOutboundMux(list []any) map[string]any {
	if len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["enabled"].(bool); ok {
		out["enabled"] = v
	}
	if v, ok := item["concurrency"].(int); ok && v != 0 {
		out["concurrency"] = v
	}
	if v, ok := item["xudp_concurrency"].(int); ok && v != 0 {
		out["xudpConcurrency"] = v
	}
	if v, ok := item["xudp_proxy_udp443"].(string); ok && v != "" {
		out["xudpProxyUDP443"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func expandOutboundSettings(m map[string]any, protocol string) map[string]any {
	switch protocol {
	case "freedom":
		return expandFreedomSettings(m)
	case "blackhole":
		return expandBlackholeSettings(m)
	case "dns":
		return expandOutboundDNSSettings(m)
	case "vmess":
		return expandVmessOutSettings(m)
	case "vless":
		return expandVlessOutSettings(m)
	case "trojan":
		return expandTrojanOutSettings(m)
	case "shadowsocks":
		return expandShadowsocksOutSettings(m)
	case "socks":
		return expandSocksOutSettings(m)
	case "http":
		return expandHTTPOutSettings(m)
	case "wireguard":
		return expandWireguardOutSettings(m)
	case "hysteria", "hysteria2":
		return expandHysteriaOutSettings(m)
	default:
		return nil
	}
}

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

func expandBlackholeSettings(m map[string]any) map[string]any {
	list, ok := m["blackhole_settings"].([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["response_type"].(string); ok && v != "" {
		out["response"] = map[string]any{"type": v}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func expandOutboundDNSSettings(m map[string]any) map[string]any {
	list, ok := m["dns_settings"].([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["network"].(string); ok && v != "" {
		out["network"] = v
	}
	if v, ok := item["address"].(string); ok && v != "" {
		out["address"] = v
	}
	if v, ok := item["port"].(int); ok && v != 0 {
		out["port"] = v
	}
	if v, ok := item["non_ip_query"].(string); ok && v != "" {
		out["nonIPQuery"] = v
	}
	if v, ok := item["block_types"].([]any); ok && len(v) > 0 {
		out["blockTypes"] = expandIntList(v)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func expandVmessOutSettings(m map[string]any) map[string]any {
	list, ok := m["vmess_settings"].([]any)
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
	if v, ok := item["security"].(string); ok && v != "" {
		user["security"] = v
	}
	if len(user) > 0 {
		server["users"] = []any{user}
	}
	return map[string]any{"vnext": []any{server}}
}

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

func expandTrojanOutSettings(m map[string]any) map[string]any {
	list, ok := m["trojan_settings"].([]any)
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
	if v, ok := item["password"].(string); ok && v != "" {
		server["password"] = v
	}
	return map[string]any{"servers": []any{server}}
}

func expandShadowsocksOutSettings(m map[string]any) map[string]any {
	list, ok := m["shadowsocks_settings"].([]any)
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
	if v, ok := item["password"].(string); ok && v != "" {
		server["password"] = v
	}
	if v, ok := item["method"].(string); ok && v != "" {
		server["method"] = v
	}
	if v, ok := item["uot"].(bool); ok {
		server["uot"] = v
	}
	if v, ok := item["uot_version"].(int); ok && v != 0 {
		server["UoTVersion"] = v
	}
	return map[string]any{"servers": []any{server}}
}

func expandSocksOutSettings(m map[string]any) map[string]any {
	list, ok := m["socks_settings"].([]any)
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
	if v, ok := item["user"].(string); ok && v != "" {
		user["user"] = v
	}
	if v, ok := item["pass"].(string); ok && v != "" {
		user["pass"] = v
	}
	if len(user) > 0 {
		server["users"] = []any{user}
	}
	return map[string]any{"servers": []any{server}}
}

func expandHTTPOutSettings(m map[string]any) map[string]any {
	list, ok := m["http_settings"].([]any)
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
	if v, ok := item["user"].(string); ok && v != "" {
		user["user"] = v
	}
	if v, ok := item["pass"].(string); ok && v != "" {
		user["pass"] = v
	}
	if len(user) > 0 {
		server["users"] = []any{user}
	}
	return map[string]any{"servers": []any{server}}
}

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
		out["reserved"] = expandIntList(v)
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

// --- Flatten ---

func flattenXrayOutboundsToMap(data any) map[string]any {
	out := map[string]any{}
	if data == nil {
		return out
	}

	var list []any
	switch v := data.(type) {
	case []any:
		list = v
	case string:
		if err := json.Unmarshal([]byte(v), &list); err != nil {
			return out
		}
	default:
		return out
	}

	out["outbound"] = flattenOutbounds(list)
	return out
}

func flattenOutbounds(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		protocol := ""

		if v, ok := m["tag"].(string); ok {
			entry["tag"] = v
		}
		if v, ok := m["protocol"].(string); ok {
			protocol = v
			entry["protocol"] = v
		}
		if v, ok := m["sendThrough"].(string); ok {
			entry["send_through"] = v
		}

		// Mux
		if v, ok := m["mux"].(map[string]any); ok {
			if mux := flattenOutboundMux(v); mux != nil {
				entry["mux"] = []any{mux}
			}
		}

		// Protocol-specific settings
		settings, _ := m["settings"].(map[string]any)
		flattenOutboundProtocolSettings(entry, settings, protocol)

		out = append(out, entry)
	}
	return out
}

func flattenOutboundMux(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := map[string]any{}
	if v, ok := in["enabled"].(bool); ok {
		out["enabled"] = v
	}
	if v, ok := in["concurrency"]; ok {
		out["concurrency"] = intValue(v)
	}
	if v, ok := in["xudpConcurrency"]; ok {
		out["xudp_concurrency"] = intValue(v)
	}
	if v, ok := in["xudpProxyUDP443"].(string); ok {
		out["xudp_proxy_udp443"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func flattenOutboundProtocolSettings(entry map[string]any, settings map[string]any, protocol string) {
	if settings == nil {
		return
	}
	switch protocol {
	case "freedom":
		entry["freedom_settings"] = []any{flattenFreedomSettings(settings)}
	case "blackhole":
		entry["blackhole_settings"] = []any{flattenBlackholeSettings(settings)}
	case "dns":
		entry["dns_settings"] = []any{flattenOutboundDNSSettings(settings)}
	case "vmess":
		entry["vmess_settings"] = []any{flattenVmessOutSettings(settings)}
	case "vless":
		entry["vless_settings"] = []any{flattenVlessOutSettings(settings)}
	case "trojan":
		entry["trojan_settings"] = []any{flattenTrojanOutSettings(settings)}
	case "shadowsocks":
		entry["shadowsocks_settings"] = []any{flattenShadowsocksOutSettings(settings)}
	case "socks":
		entry["socks_settings"] = []any{flattenSocksOutSettings(settings)}
	case "http":
		entry["http_settings"] = []any{flattenHTTPOutSettings(settings)}
	case "wireguard":
		entry["wireguard_settings"] = []any{flattenWireguardOutSettings(settings)}
	case "hysteria", "hysteria2":
		entry["hysteria_settings"] = []any{flattenHysteriaOutSettings(settings)}
	}
}

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

func flattenBlackholeSettings(in map[string]any) map[string]any {
	out := map[string]any{}
	if v, ok := in["response"].(map[string]any); ok {
		if t, ok := v["type"].(string); ok {
			out["response_type"] = t
		}
	}
	if _, ok := out["response_type"]; !ok {
		out["response_type"] = "none"
	}
	return out
}

func flattenOutboundDNSSettings(in map[string]any) map[string]any {
	out := map[string]any{}
	if v, ok := in["network"].(string); ok {
		out["network"] = v
	}
	if v, ok := in["address"].(string); ok {
		out["address"] = v
	}
	if v, ok := in["port"]; ok {
		out["port"] = intValue(v)
	}
	if v, ok := in["nonIPQuery"].(string); ok {
		out["non_ip_query"] = v
	}
	if v, ok := in["blockTypes"].([]any); ok {
		out["block_types"] = flattenIntList(v)
	}
	return out
}

func flattenVnextFirstUser(in map[string]any, fields ...string) map[string]any {
	out := map[string]any{}
	server := firstVnextServer(in)
	if server == nil {
		return out
	}
	if v, ok := server["address"].(string); ok {
		out["address"] = v
	}
	if v, ok := server["port"]; ok {
		out["port"] = intValue(v)
	}
	users, ok := server["users"].([]any)
	if ok && len(users) > 0 {
		user, ok := users[0].(map[string]any)
		if ok {
			for _, f := range fields {
				if v, ok := user[f]; ok {
					out[f] = v
				}
			}
		}
	}
	return out
}

func firstVnextServer(in map[string]any) map[string]any {
	vnext, ok := in["vnext"].([]any)
	if !ok || len(vnext) == 0 {
		return nil
	}
	server, ok := vnext[0].(map[string]any)
	if !ok {
		return nil
	}
	return server
}

func flattenVmessOutSettings(in map[string]any) map[string]any {
	return flattenVnextFirstUser(in, "id", "security")
}

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

func flattenServersFirst(in map[string]any, fields ...string) map[string]any {
	out := map[string]any{}
	servers, ok := in["servers"].([]any)
	if !ok || len(servers) == 0 {
		return out
	}
	server, ok := servers[0].(map[string]any)
	if !ok {
		return out
	}
	if v, ok := server["address"].(string); ok {
		out["address"] = v
	}
	if v, ok := server["port"]; ok {
		out["port"] = intValue(v)
	}
	for _, f := range fields {
		if v, ok := server[f]; ok {
			out[f] = v
		}
	}
	return out
}

func flattenTrojanOutSettings(in map[string]any) map[string]any {
	return flattenServersFirst(in, "password")
}

func flattenShadowsocksOutSettings(in map[string]any) map[string]any {
	out := flattenServersFirst(in, "password", "method")
	servers, ok := in["servers"].([]any)
	if ok && len(servers) > 0 {
		server, ok := servers[0].(map[string]any)
		if ok {
			if v, ok := server["uot"].(bool); ok {
				out["uot"] = v
			}
			if v, ok := server["UoTVersion"]; ok {
				out["uot_version"] = intValue(v)
			}
		}
	}
	return out
}

func flattenSocksOutSettings(in map[string]any) map[string]any {
	out := map[string]any{}
	servers, ok := in["servers"].([]any)
	if !ok || len(servers) == 0 {
		return out
	}
	server, ok := servers[0].(map[string]any)
	if !ok {
		return out
	}
	if v, ok := server["address"].(string); ok {
		out["address"] = v
	}
	if v, ok := server["port"]; ok {
		out["port"] = intValue(v)
	}
	users, ok := server["users"].([]any)
	if ok && len(users) > 0 {
		user, ok := users[0].(map[string]any)
		if ok {
			if v, ok := user["user"].(string); ok {
				out["user"] = v
			}
			if v, ok := user["pass"].(string); ok {
				out["pass"] = v
			}
		}
	}
	return out
}

func flattenHTTPOutSettings(in map[string]any) map[string]any {
	return flattenSocksOutSettings(in) // same structure
}

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
