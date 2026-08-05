package provider

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// Typed models
// ---------------------------------------------------------------------------

type XrayBasicsModel struct {
	ID      types.String        `tfsdk:"id"`
	Log     []XrayBasicsLog     `tfsdk:"log"`
	Policy  []XrayBasicsPolicy  `tfsdk:"policy"`
	API     []XrayBasicsAPI     `tfsdk:"api"`
	Stats   []XrayBasicsStats   `tfsdk:"stats"`
	Metrics []XrayBasicsMetrics `tfsdk:"metrics"`
	Env     []XrayBasicsEnv     `tfsdk:"env"`
}

type XrayBasicsLog struct {
	Loglevel types.String `tfsdk:"loglevel"`
	Access   types.String `tfsdk:"access"`
	Error    types.String `tfsdk:"error"`
	DNSLog   types.Bool   `tfsdk:"dns_log"`
}

type XrayBasicsPolicy struct {
	System []XrayBasicsPolicySystem `tfsdk:"system"`
	Level  []XrayBasicsPolicyLevel  `tfsdk:"level"`
}

type XrayBasicsPolicySystem struct {
	StatsInboundDownlink  types.Bool `tfsdk:"stats_inbound_downlink"`
	StatsInboundUplink    types.Bool `tfsdk:"stats_inbound_uplink"`
	StatsOutboundDownlink types.Bool `tfsdk:"stats_outbound_downlink"`
	StatsOutboundUplink   types.Bool `tfsdk:"stats_outbound_uplink"`
}

type XrayBasicsPolicyLevel struct {
	ID                types.Int64 `tfsdk:"id"`
	Handshake         types.Int64 `tfsdk:"handshake"`
	ConnIdle          types.Int64 `tfsdk:"conn_idle"`
	UplinkOnly        types.Int64 `tfsdk:"uplink_only"`
	DownlinkOnly      types.Int64 `tfsdk:"downlink_only"`
	StatsUserUplink   types.Bool  `tfsdk:"stats_user_uplink"`
	StatsUserDownlink types.Bool  `tfsdk:"stats_user_downlink"`
	BufferSize        types.Int64 `tfsdk:"buffer_size"`
}

type XrayBasicsAPI struct {
	Tag      types.String `tfsdk:"tag"`
	Services types.List   `tfsdk:"services"`
}

type XrayBasicsStats struct{}

type XrayBasicsMetrics struct {
	Tag    types.String `tfsdk:"tag"`
	Listen types.String `tfsdk:"listen"`
}

// XrayBasicsEnv represents one entry of the xray-core top-level `env` map
// (3x-ui v3.5.0+, xray-core v26.7.11+). xray-core expects a map[string]string;
// the provider models it as a repeated block so state stays deterministic.
// Keys are arbitrary environment-variable names (e.g. XRAY_LOG_LEVEL) and are
// written to the wire verbatim — no snake_case/camelCase translation.
type XrayBasicsEnv struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

func xrayBasicsSchema() schema.Schema {
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
			"log": singletonListNestedBlock(schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"loglevel": schema.StringAttribute{
							Optional: true, Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"access": schema.StringAttribute{
							Optional: true, Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"error": schema.StringAttribute{
							Optional: true, Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"dns_log": schema.BoolAttribute{
							Optional: true, Computed: true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			}),
			"policy": singletonListNestedBlock(schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"system": singletonListNestedBlock(schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"stats_inbound_downlink": schema.BoolAttribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.UseStateForUnknown(),
										},
									},
									"stats_inbound_uplink": schema.BoolAttribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.UseStateForUnknown(),
										},
									},
									"stats_outbound_downlink": schema.BoolAttribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.UseStateForUnknown(),
										},
									},
									"stats_outbound_uplink": schema.BoolAttribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						}),
						"level": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.Int64Attribute{
										Required: true,
									},
									"handshake": schema.Int64Attribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.UseStateForUnknown(),
										},
									},
									"conn_idle": schema.Int64Attribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.UseStateForUnknown(),
										},
									},
									"uplink_only": schema.Int64Attribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.UseStateForUnknown(),
										},
									},
									"downlink_only": schema.Int64Attribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.UseStateForUnknown(),
										},
									},
									"stats_user_uplink": schema.BoolAttribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.UseStateForUnknown(),
										},
									},
									"stats_user_downlink": schema.BoolAttribute{
										Optional: true, Computed: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.UseStateForUnknown(),
										},
									},
									"buffer_size": schema.Int64Attribute{
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
			}),
			"api": singletonListNestedBlock(schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"tag": schema.StringAttribute{
							Optional: true, Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"services": schema.ListAttribute{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			}),
			"stats": singletonListNestedBlock(schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{},
				},
			}),
			"metrics": singletonListNestedBlock(schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"tag": schema.StringAttribute{
							Optional: true, Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"listen": schema.StringAttribute{
							Optional: true, Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			}),
			"env": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required:    true,
							Description: "Environment variable name (e.g. XRAY_LOG_LEVEL). Written verbatim to the xray-core template. Duplicate keys are not enforced at the schema level — the last entry wins on the wire.",
						},
						"value": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Environment variable value.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// alignBasicsBlocksWithPlan nils out blocks on the state that were not present
// in the plan. This prevents the "was absent, but now present" inconsistency
// error that Terraform raises when 3x-ui returns default values for omitted
// optional blocks.
// ---------------------------------------------------------------------------

func alignBasicsBlocksWithPlan(state, plan *XrayBasicsModel) {
	if len(plan.Log) == 0 {
		state.Log = nil
	}
	if len(plan.Policy) == 0 {
		state.Policy = nil
	} else {
		planPol := plan.Policy[0]
		if len(state.Policy) > 0 {
			statePol := &state.Policy[0]
			if len(planPol.System) == 0 {
				statePol.System = nil
			}
			if len(planPol.Level) == 0 {
				statePol.Level = nil
			}
		}
	}
	if len(plan.API) == 0 {
		state.API = nil
	}
	if len(plan.Stats) == 0 {
		state.Stats = nil
	}
	if len(plan.Metrics) == 0 {
		state.Metrics = nil
	}
	if len(plan.Env) == 0 {
		state.Env = nil
	}
}

// ---------------------------------------------------------------------------
// Expand: typed model -> untyped map (for buildXrayBasicsJSON)
// ---------------------------------------------------------------------------

func expandXrayBasics(m *XrayBasicsModel) map[string]any {
	out := map[string]any{}

	if len(m.Log) > 0 {
		log := m.Log[0]
		logMap := map[string]any{}
		if !log.Loglevel.IsNull() && !log.Loglevel.IsUnknown() {
			logMap["loglevel"] = log.Loglevel.ValueString()
		}
		if !log.Access.IsNull() && !log.Access.IsUnknown() {
			logMap["access"] = log.Access.ValueString()
		}
		if !log.Error.IsNull() && !log.Error.IsUnknown() {
			logMap["error"] = log.Error.ValueString()
		}
		if !log.DNSLog.IsNull() && !log.DNSLog.IsUnknown() {
			logMap["dns_log"] = log.DNSLog.ValueBool()
		}
		out["log"] = logMap
	}

	if len(m.Policy) > 0 {
		pol := m.Policy[0]
		polMap := map[string]any{}

		if len(pol.System) > 0 {
			sys := pol.System[0]
			sysMap := map[string]any{}
			if !sys.StatsInboundDownlink.IsNull() && !sys.StatsInboundDownlink.IsUnknown() {
				sysMap["stats_inbound_downlink"] = sys.StatsInboundDownlink.ValueBool()
			}
			if !sys.StatsInboundUplink.IsNull() && !sys.StatsInboundUplink.IsUnknown() {
				sysMap["stats_inbound_uplink"] = sys.StatsInboundUplink.ValueBool()
			}
			if !sys.StatsOutboundDownlink.IsNull() && !sys.StatsOutboundDownlink.IsUnknown() {
				sysMap["stats_outbound_downlink"] = sys.StatsOutboundDownlink.ValueBool()
			}
			if !sys.StatsOutboundUplink.IsNull() && !sys.StatsOutboundUplink.IsUnknown() {
				sysMap["stats_outbound_uplink"] = sys.StatsOutboundUplink.ValueBool()
			}
			polMap["system"] = sysMap
		}

		if len(pol.Level) > 0 {
			levels := make([]any, 0, len(pol.Level))
			for _, lvl := range pol.Level {
				entry := map[string]any{}
				if !lvl.ID.IsNull() && !lvl.ID.IsUnknown() {
					entry["id"] = int(lvl.ID.ValueInt64())
				}
				if !lvl.Handshake.IsNull() && !lvl.Handshake.IsUnknown() {
					entry["handshake"] = int(lvl.Handshake.ValueInt64())
				}
				if !lvl.ConnIdle.IsNull() && !lvl.ConnIdle.IsUnknown() {
					entry["conn_idle"] = int(lvl.ConnIdle.ValueInt64())
				}
				if !lvl.UplinkOnly.IsNull() && !lvl.UplinkOnly.IsUnknown() {
					entry["uplink_only"] = int(lvl.UplinkOnly.ValueInt64())
				}
				if !lvl.DownlinkOnly.IsNull() && !lvl.DownlinkOnly.IsUnknown() {
					entry["downlink_only"] = int(lvl.DownlinkOnly.ValueInt64())
				}
				if !lvl.StatsUserUplink.IsNull() && !lvl.StatsUserUplink.IsUnknown() {
					entry["stats_user_uplink"] = lvl.StatsUserUplink.ValueBool()
				}
				if !lvl.StatsUserDownlink.IsNull() && !lvl.StatsUserDownlink.IsUnknown() {
					entry["stats_user_downlink"] = lvl.StatsUserDownlink.ValueBool()
				}
				if !lvl.BufferSize.IsNull() && !lvl.BufferSize.IsUnknown() {
					entry["buffer_size"] = int(lvl.BufferSize.ValueInt64())
				}
				levels = append(levels, entry)
			}
			polMap["level"] = levels
		}

		out["policy"] = polMap
	}

	if len(m.API) > 0 {
		api := m.API[0]
		apiMap := map[string]any{}
		if !api.Tag.IsNull() && !api.Tag.IsUnknown() {
			apiMap["tag"] = api.Tag.ValueString()
		}
		if !api.Services.IsNull() && !api.Services.IsUnknown() {
			elems := api.Services.Elements()
			svcList := make([]any, 0, len(elems))
			for _, e := range elems {
				if sv, ok := e.(types.String); ok {
					svcList = append(svcList, sv.ValueString())
				}
			}
			apiMap["services"] = svcList
		}
		out["api"] = apiMap
	}

	if len(m.Stats) > 0 {
		out["stats"] = map[string]any{}
	}

	if len(m.Metrics) > 0 {
		metrics := m.Metrics[0]
		metricsMap := map[string]any{}
		if !metrics.Tag.IsNull() && !metrics.Tag.IsUnknown() {
			metricsMap["tag"] = metrics.Tag.ValueString()
		}
		if !metrics.Listen.IsNull() && !metrics.Listen.IsUnknown() {
			metricsMap["listen"] = metrics.Listen.ValueString()
		}
		out["metrics"] = metricsMap
	}

	if len(m.Env) > 0 {
		envMap := map[string]any{}
		for _, e := range m.Env {
			if e.Key.IsNull() || e.Key.IsUnknown() {
				continue
			}
			k := e.Key.ValueString()
			v := ""
			if !e.Value.IsNull() && !e.Value.IsUnknown() {
				v = e.Value.ValueString()
			}
			envMap[k] = v
		}
		if len(envMap) > 0 {
			out["env"] = envMap
		}
	}

	return out
}

// ---------------------------------------------------------------------------
// Flatten: untyped map (from flattenXrayBasicsToMap) -> typed model
// ---------------------------------------------------------------------------

func flattenXrayBasics(data map[string]any) *XrayBasicsModel {
	m := &XrayBasicsModel{
		ID: types.StringValue(xraySectionBasics.id),
	}

	if logRaw, ok := data["log"]; ok {
		if logMap, ok := logRaw.(map[string]any); ok && len(logMap) > 0 {
			log := XrayBasicsLog{}
			if v, ok := logMap["loglevel"].(string); ok && v != "" {
				log.Loglevel = types.StringValue(v)
			} else {
				log.Loglevel = types.StringNull()
			}
			if v, ok := logMap["access"].(string); ok && v != "" {
				log.Access = types.StringValue(v)
			} else {
				log.Access = types.StringNull()
			}
			if v, ok := logMap["error"].(string); ok && v != "" {
				log.Error = types.StringValue(v)
			} else {
				log.Error = types.StringNull()
			}
			if v, ok := logMap["dns_log"].(bool); ok {
				log.DNSLog = types.BoolValue(v)
			} else {
				log.DNSLog = types.BoolNull()
			}
			m.Log = []XrayBasicsLog{log}
		}
	}

	if polRaw, ok := data["policy"]; ok {
		if polMap, ok := polRaw.(map[string]any); ok && len(polMap) > 0 {
			pol := XrayBasicsPolicy{}

			if sysRaw, ok := polMap["system"]; ok {
				if sysMap, ok := sysRaw.(map[string]any); ok && len(sysMap) > 0 {
					sys := XrayBasicsPolicySystem{}
					if v, ok := sysMap["stats_inbound_downlink"].(bool); ok {
						sys.StatsInboundDownlink = types.BoolValue(v)
					} else {
						sys.StatsInboundDownlink = types.BoolNull()
					}
					if v, ok := sysMap["stats_inbound_uplink"].(bool); ok {
						sys.StatsInboundUplink = types.BoolValue(v)
					} else {
						sys.StatsInboundUplink = types.BoolNull()
					}
					if v, ok := sysMap["stats_outbound_downlink"].(bool); ok {
						sys.StatsOutboundDownlink = types.BoolValue(v)
					} else {
						sys.StatsOutboundDownlink = types.BoolNull()
					}
					if v, ok := sysMap["stats_outbound_uplink"].(bool); ok {
						sys.StatsOutboundUplink = types.BoolValue(v)
					} else {
						sys.StatsOutboundUplink = types.BoolNull()
					}
					pol.System = []XrayBasicsPolicySystem{sys}
				}
			}

			if levelsRaw, ok := polMap["level"]; ok {
				if levelsList, ok := levelsRaw.([]any); ok && len(levelsList) > 0 {
					levels := make([]XrayBasicsPolicyLevel, 0, len(levelsList))
					for _, item := range levelsList {
						lm, ok := item.(map[string]any)
						if !ok {
							continue
						}
						lvl := XrayBasicsPolicyLevel{}
						lvl.ID = types.Int64Value(int64(intValue(lm["id"])))
						if v, ok := lm["handshake"]; ok {
							lvl.Handshake = types.Int64Value(int64(intValue(v)))
						} else {
							lvl.Handshake = types.Int64Null()
						}
						if v, ok := lm["conn_idle"]; ok {
							lvl.ConnIdle = types.Int64Value(int64(intValue(v)))
						} else {
							lvl.ConnIdle = types.Int64Null()
						}
						if v, ok := lm["uplink_only"]; ok {
							lvl.UplinkOnly = types.Int64Value(int64(intValue(v)))
						} else {
							lvl.UplinkOnly = types.Int64Null()
						}
						if v, ok := lm["downlink_only"]; ok {
							lvl.DownlinkOnly = types.Int64Value(int64(intValue(v)))
						} else {
							lvl.DownlinkOnly = types.Int64Null()
						}
						if v, ok := lm["stats_user_uplink"].(bool); ok {
							lvl.StatsUserUplink = types.BoolValue(v)
						} else {
							lvl.StatsUserUplink = types.BoolNull()
						}
						if v, ok := lm["stats_user_downlink"].(bool); ok {
							lvl.StatsUserDownlink = types.BoolValue(v)
						} else {
							lvl.StatsUserDownlink = types.BoolNull()
						}
						if v, ok := lm["buffer_size"]; ok {
							lvl.BufferSize = types.Int64Value(int64(intValue(v)))
						} else {
							lvl.BufferSize = types.Int64Null()
						}
						levels = append(levels, lvl)
					}
					pol.Level = levels
				}
			}

			m.Policy = []XrayBasicsPolicy{pol}
		}
	}

	if apiRaw, ok := data["api"]; ok {
		if apiMap, ok := apiRaw.(map[string]any); ok && len(apiMap) > 0 {
			api := XrayBasicsAPI{}
			if v, ok := apiMap["tag"].(string); ok {
				api.Tag = types.StringValue(v)
			} else {
				api.Tag = types.StringNull()
			}
			if v, ok := apiMap["services"].([]any); ok && len(v) > 0 {
				elems := make([]attr.Value, 0, len(v))
				for _, s := range v {
					if str, ok := s.(string); ok {
						elems = append(elems, types.StringValue(str))
					}
				}
				api.Services = types.ListValueMust(types.StringType, elems)
			} else {
				api.Services = types.ListNull(types.StringType)
			}
			m.API = []XrayBasicsAPI{api}
		}
	}

	if _, ok := data["stats"]; ok {
		m.Stats = []XrayBasicsStats{{}}
	}

	if metricsRaw, ok := data["metrics"]; ok {
		if metricsMap, ok := metricsRaw.(map[string]any); ok && len(metricsMap) > 0 {
			metrics := XrayBasicsMetrics{}
			if v, ok := metricsMap["tag"].(string); ok && v != "" {
				metrics.Tag = types.StringValue(v)
			} else {
				metrics.Tag = types.StringNull()
			}
			if v, ok := metricsMap["listen"].(string); ok && v != "" {
				metrics.Listen = types.StringValue(v)
			} else {
				metrics.Listen = types.StringNull()
			}
			m.Metrics = []XrayBasicsMetrics{metrics}
		}
	}

	if envRaw, ok := data["env"]; ok {
		if envMap, ok := envRaw.(map[string]any); ok && len(envMap) > 0 {
			env := make([]XrayBasicsEnv, 0, len(envMap))
			for k, raw := range envMap {
				e := XrayBasicsEnv{Key: types.StringValue(k)}
				if v, ok := raw.(string); ok && v != "" {
					e.Value = types.StringValue(v)
				} else {
					e.Value = types.StringNull()
				}
				env = append(env, e)
			}
			// Deterministic order: Go randomises map iteration.
			sort.Slice(env, func(i, j int) bool {
				return env[i].Key.ValueString() < env[j].Key.ValueString()
			})
			if len(env) > 0 {
				m.Env = env
			}
		}
	}

	return m
}

// ---------------------------------------------------------------------------
// Existing build/flatten functions (untyped map <-> Xray JSON)
// ---------------------------------------------------------------------------

func buildXrayBasicsJSON(d map[string]any) any {
	payload := map[string]any{}

	if v, ok := d["log"]; ok {
		if m, ok := v.(map[string]any); ok {
			if log := expandBasicsLog(m); log != nil {
				payload["log"] = log
			}
		}
	}
	if v, ok := d["policy"]; ok {
		if m, ok := v.(map[string]any); ok {
			if policy := expandBasicsPolicy(m); policy != nil {
				payload["policy"] = policy
			}
		}
	}
	if v, ok := d["api"]; ok {
		if m, ok := v.(map[string]any); ok {
			if api := expandBasicsAPI(m); api != nil {
				payload["api"] = api
			}
		}
	}
	if _, ok := d["stats"]; ok {
		payload["stats"] = map[string]any{}
	}
	if v, ok := d["metrics"]; ok {
		if m, ok := v.(map[string]any); ok {
			if metrics := expandBasicsMetrics(m); metrics != nil {
				payload["metrics"] = metrics
			}
		}
	}
	// env is a map[string]string carried verbatim — keys are arbitrary
	// environment-variable names, not translated to camelCase like the
	// nested log/policy/api attributes.
	if v, ok := d["env"]; ok {
		if m, ok := v.(map[string]any); ok && len(m) > 0 {
			payload["env"] = m
		}
	}

	return payload
}

func expandBasicsLog(item map[string]any) map[string]any {
	if len(item) == 0 {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["loglevel"].(string); ok && v != "" {
		out["loglevel"] = v
	}
	if v, ok := item["access"].(string); ok && v != "" {
		out["access"] = v
	}
	if v, ok := item["error"].(string); ok && v != "" {
		out["error"] = v
	}
	if v, ok := item["dns_log"]; ok {
		out["dnsLog"] = boolValue(v)
	}
	return out
}

func expandBasicsPolicy(item map[string]any) map[string]any {
	if len(item) == 0 {
		return nil
	}
	out := map[string]any{}

	if v, ok := item["system"]; ok {
		if m, ok := v.(map[string]any); ok {
			if sys := expandBasicsPolicySystem(m); sys != nil {
				out["system"] = sys
			}
		}
	}
	if v, ok := item["level"]; ok {
		if list, ok := v.([]any); ok {
			if levels := expandBasicsPolicyLevels(list); levels != nil {
				out["levels"] = levels
			}
		}
	}

	if len(out) == 0 {
		return nil
	}
	return out
}

func expandBasicsPolicySystem(item map[string]any) map[string]any {
	if len(item) == 0 {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["stats_inbound_downlink"]; ok {
		out["statsInboundDownlink"] = boolValue(v)
	}
	if v, ok := item["stats_inbound_uplink"]; ok {
		out["statsInboundUplink"] = boolValue(v)
	}
	if v, ok := item["stats_outbound_downlink"]; ok {
		out["statsOutboundDownlink"] = boolValue(v)
	}
	if v, ok := item["stats_outbound_uplink"]; ok {
		out["statsOutboundUplink"] = boolValue(v)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// expandBasicsPolicyLevels converts TF level blocks to Xray policy.levels map.
// Xray uses string keys like "0", "1" for levels.
func expandBasicsPolicyLevels(list []any) map[string]any {
	if len(list) == 0 {
		return nil
	}
	out := map[string]any{}
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		id := intValue(m["id"])
		entry := map[string]any{}
		if v, ok := m["handshake"]; ok {
			entry["handshake"] = intValue(v)
		}
		if v, ok := m["conn_idle"]; ok {
			entry["connIdle"] = intValue(v)
		}
		if v, ok := m["uplink_only"]; ok {
			entry["uplinkOnly"] = intValue(v)
		}
		if v, ok := m["downlink_only"]; ok {
			entry["downlinkOnly"] = intValue(v)
		}
		if v, ok := m["stats_user_uplink"]; ok {
			entry["statsUserUplink"] = boolValue(v)
		}
		if v, ok := m["stats_user_downlink"]; ok {
			entry["statsUserDownlink"] = boolValue(v)
		}
		if v, ok := m["buffer_size"]; ok {
			entry["bufferSize"] = intValue(v)
		}
		out[fmt.Sprintf("%d", id)] = entry
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func flattenXrayBasicsToMap(data any) map[string]any {
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

	if v, ok := payload["log"].(map[string]any); ok {
		if log := flattenBasicsLog(v); log != nil {
			out["log"] = log
		}
	}
	if v, ok := payload["policy"].(map[string]any); ok {
		if policy := flattenBasicsPolicy(v); policy != nil {
			out["policy"] = policy
		}
	}
	if v, ok := payload["api"].(map[string]any); ok {
		if api := flattenBasicsAPI(v); api != nil {
			out["api"] = api
		}
	}
	if _, ok := payload["stats"]; ok {
		out["stats"] = map[string]any{}
	}
	if v, ok := payload["metrics"].(map[string]any); ok {
		if metrics := flattenBasicsMetrics(v); metrics != nil {
			out["metrics"] = metrics
		}
	}
	if v, ok := payload["env"].(map[string]any); ok && len(v) > 0 {
		out["env"] = v
	}

	return out
}

func expandBasicsAPI(item map[string]any) map[string]any {
	if len(item) == 0 {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["tag"].(string); ok && v != "" {
		out["tag"] = v
	}
	if v, ok := item["services"]; ok {
		if list, ok := v.([]any); ok {
			out["services"] = expandStringList(list)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func flattenBasicsLog(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := map[string]any{}
	if v, ok := in["loglevel"].(string); ok {
		out["loglevel"] = v
	}
	if v, ok := in["access"].(string); ok {
		out["access"] = v
	}
	if v, ok := in["error"].(string); ok {
		out["error"] = v
	}
	if v, ok := in["dnsLog"].(bool); ok {
		out["dns_log"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func flattenBasicsPolicy(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := map[string]any{}

	if v, ok := in["system"].(map[string]any); ok {
		if sys := flattenBasicsPolicySystem(v); sys != nil {
			out["system"] = sys
		}
	}
	if v, ok := in["levels"].(map[string]any); ok {
		if levels := flattenBasicsPolicyLevels(v); levels != nil {
			out["level"] = levels
		}
	}

	if len(out) == 0 {
		return nil
	}
	return out
}

func flattenBasicsPolicySystem(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := map[string]any{}
	if v, ok := in["statsInboundDownlink"].(bool); ok {
		out["stats_inbound_downlink"] = v
	}
	if v, ok := in["statsInboundUplink"].(bool); ok {
		out["stats_inbound_uplink"] = v
	}
	if v, ok := in["statsOutboundDownlink"].(bool); ok {
		out["stats_outbound_downlink"] = v
	}
	if v, ok := in["statsOutboundUplink"].(bool); ok {
		out["stats_outbound_uplink"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// flattenBasicsPolicyLevels converts Xray policy.levels map to TF level blocks.
func flattenBasicsPolicyLevels(in map[string]any) []any {
	if len(in) == 0 {
		return nil
	}
	keys := make([]string, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]any, 0, len(in))
	for _, key := range keys {
		val := in[key]
		m, ok := val.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		// Parse level ID from map key
		id := 0
		if _, err := fmt.Sscanf(key, "%d", &id); err == nil {
			entry["id"] = id
		}
		if v, ok := m["handshake"]; ok {
			entry["handshake"] = intValue(v)
		}
		if v, ok := m["connIdle"]; ok {
			entry["conn_idle"] = intValue(v)
		}
		if v, ok := m["uplinkOnly"]; ok {
			entry["uplink_only"] = intValue(v)
		}
		if v, ok := m["downlinkOnly"]; ok {
			entry["downlink_only"] = intValue(v)
		}
		if v, ok := m["statsUserUplink"].(bool); ok {
			entry["stats_user_uplink"] = v
		}
		if v, ok := m["statsUserDownlink"].(bool); ok {
			entry["stats_user_downlink"] = v
		}
		if v, ok := m["bufferSize"]; ok {
			entry["buffer_size"] = intValue(v)
		}
		out = append(out, entry)
	}
	return out
}

func flattenBasicsAPI(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := map[string]any{}
	if v, ok := in["tag"].(string); ok {
		out["tag"] = v
	}
	if v, ok := in["services"].([]any); ok {
		out["services"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func expandBasicsMetrics(item map[string]any) map[string]any {
	if len(item) == 0 {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["tag"].(string); ok && v != "" {
		out["tag"] = v
	}
	if v, ok := item["listen"].(string); ok && v != "" {
		out["listen"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func flattenBasicsMetrics(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := map[string]any{}
	if v, ok := in["tag"].(string); ok {
		out["tag"] = v
	}
	if v, ok := in["listen"].(string); ok {
		out["listen"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
