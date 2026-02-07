package provider

import (
	"encoding/json"
	"fmt"
)

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
