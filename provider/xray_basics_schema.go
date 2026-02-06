package provider

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func xrayBasicsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"log": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"loglevel": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"access": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"error": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"dns_log": {
					Type:     schema.TypeBool,
					Optional: true,
				},
			}},
		},
		"policy": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"system": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{Schema: map[string]*schema.Schema{
						"stats_inbound_downlink": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"stats_inbound_uplink": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"stats_outbound_downlink": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"stats_outbound_uplink": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					}},
				},
				"level": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"handshake": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"conn_idle": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"uplink_only": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"downlink_only": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"stats_user_uplink": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"stats_user_downlink": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"buffer_size": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					}},
				},
			}},
		},
		"api": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"tag": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"services": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
			}},
		},
		"stats": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem:     &schema.Resource{Schema: map[string]*schema.Schema{}},
		},
	}
}

func buildXrayBasicsJSON(d *schema.ResourceData) (map[string]any, error) {
	payload := map[string]any{}

	if v, ok := d.GetOk("log"); ok {
		if log := expandBasicsLog(v.([]any)); log != nil {
			payload["log"] = log
		}
	}
	if v, ok := d.GetOk("policy"); ok {
		if policy := expandBasicsPolicy(v.([]any)); policy != nil {
			payload["policy"] = policy
		}
	}
	if v, ok := d.GetOk("api"); ok {
		if api := expandBasicsAPI(v.([]any)); api != nil {
			payload["api"] = api
		}
	}
	if _, ok := d.GetOk("stats"); ok {
		payload["stats"] = map[string]any{}
	}

	return payload, nil
}

func expandBasicsLog(list []any) map[string]any {
	if len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
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
	if v, ok := item["dns_log"].(bool); ok {
		out["dnsLog"] = v
	}
	return out
}

func expandBasicsPolicy(list []any) map[string]any {
	if len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}

	if v, ok := item["system"]; ok {
		if sys := expandBasicsPolicySystem(v.([]any)); sys != nil {
			out["system"] = sys
		}
	}
	if v, ok := item["level"]; ok {
		if levels := expandBasicsPolicyLevels(v.([]any)); levels != nil {
			out["levels"] = levels
		}
	}

	if len(out) == 0 {
		return nil
	}
	return out
}

func expandBasicsPolicySystem(list []any) map[string]any {
	if len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["stats_inbound_downlink"].(bool); ok {
		out["statsInboundDownlink"] = v
	}
	if v, ok := item["stats_inbound_uplink"].(bool); ok {
		out["statsInboundUplink"] = v
	}
	if v, ok := item["stats_outbound_downlink"].(bool); ok {
		out["statsOutboundDownlink"] = v
	}
	if v, ok := item["stats_outbound_uplink"].(bool); ok {
		out["statsOutboundUplink"] = v
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
		if v, ok := m["handshake"].(int); ok {
			entry["handshake"] = v
		}
		if v, ok := m["conn_idle"].(int); ok {
			entry["connIdle"] = v
		}
		if v, ok := m["uplink_only"].(int); ok {
			entry["uplinkOnly"] = v
		}
		if v, ok := m["downlink_only"].(int); ok {
			entry["downlinkOnly"] = v
		}
		if v, ok := m["stats_user_uplink"].(bool); ok {
			entry["statsUserUplink"] = v
		}
		if v, ok := m["stats_user_downlink"].(bool); ok {
			entry["statsUserDownlink"] = v
		}
		if v, ok := m["buffer_size"].(int); ok {
			entry["bufferSize"] = v
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
			out["log"] = []any{log}
		}
	}
	if v, ok := payload["policy"].(map[string]any); ok {
		if policy := flattenBasicsPolicy(v); policy != nil {
			out["policy"] = []any{policy}
		}
	}
	if v, ok := payload["api"].(map[string]any); ok {
		if api := flattenBasicsAPI(v); api != nil {
			out["api"] = []any{api}
		}
	}
	if _, ok := payload["stats"]; ok {
		out["stats"] = []any{map[string]any{}}
	}

	return out
}

func expandBasicsAPI(list []any) map[string]any {
	if len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["tag"].(string); ok && v != "" {
		out["tag"] = v
	}
	if v, ok := item["services"]; ok {
		out["services"] = expandStringList(v.([]any))
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
			out["system"] = []any{sys}
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
	out := make([]any, 0, len(in))
	for key, val := range in {
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
