package provider

import "encoding/json"

func buildXrayDNSJSON(d map[string]any) any {
	payload := map[string]any{}

	if v, ok := d["server"]; ok {
		if list, ok := v.([]any); ok {
			payload["servers"] = expandDNSServers(list)
		}
	}
	if v, ok := d["hosts"]; ok {
		if m, ok := v.(map[string]any); ok {
			payload["hosts"] = expandStringMap(m)
		}
	}
	if v, ok := d["query_strategy"].(string); ok && v != "" {
		payload["queryStrategy"] = v
	}
	if v, ok := d["tag"].(string); ok && v != "" {
		payload["tag"] = v
	}
	if v, ok := d["disable_cache"]; ok {
		payload["disableCache"] = boolValue(v)
	}
	if v, ok := d["disable_fallback"]; ok {
		payload["disableFallback"] = boolValue(v)
	}
	if v, ok := d["disable_fallback_if_match"]; ok {
		payload["disableFallbackIfMatch"] = boolValue(v)
	}
	if v, ok := d["client_ip"].(string); ok && v != "" {
		payload["clientIp"] = v
	}

	return payload
}

// expandDNSServers converts TF server blocks to JSON.
// A server with only address is serialized as a plain string.
// A server with additional fields is serialized as an object.
func expandDNSServers(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		address, _ := m["address"].(string)
		if address == "" {
			continue
		}

		// Check if server has any fields beyond address
		hasExtra := false
		if v, ok := m["port"]; ok {
			if intValue(v) != 0 {
				hasExtra = true
			}
		}
		if v, ok := m["domains"].([]any); ok && len(v) > 0 {
			hasExtra = true
		}
		if v, ok := m["expect_ips"].([]any); ok && len(v) > 0 {
			hasExtra = true
		}
		if v, ok := m["unexpected_ips"].([]any); ok && len(v) > 0 {
			hasExtra = true
		}
		if v, ok := m["skip_fallback"]; ok && boolValue(v) {
			hasExtra = true
		}
		if v, ok := m["query_strategy"].(string); ok && v != "" {
			hasExtra = true
		}
		if v, ok := m["disable_cache"]; ok && boolValue(v) {
			hasExtra = true
		}
		if v, ok := m["final_query"]; ok && boolValue(v) {
			hasExtra = true
		}

		if !hasExtra {
			out = append(out, address)
			continue
		}

		entry := map[string]any{
			"address": address,
		}
		if v, ok := m["port"]; ok {
			if p := intValue(v); p != 0 {
				entry["port"] = p
			}
		}
		if v, ok := m["domains"].([]any); ok && len(v) > 0 {
			entry["domains"] = expandStringList(v)
		}
		if v, ok := m["expect_ips"].([]any); ok && len(v) > 0 {
			entry["expectedIPs"] = expandStringList(v)
		}
		if v, ok := m["unexpected_ips"].([]any); ok && len(v) > 0 {
			entry["unexpectedIPs"] = expandStringList(v)
		}
		if v, ok := m["skip_fallback"]; ok {
			entry["skipFallback"] = boolValue(v)
		}
		if v, ok := m["query_strategy"].(string); ok && v != "" {
			entry["queryStrategy"] = v
		}
		if v, ok := m["disable_cache"]; ok {
			entry["disableCache"] = boolValue(v)
		}
		if v, ok := m["final_query"]; ok {
			entry["finalQuery"] = boolValue(v)
		}

		out = append(out, entry)
	}
	return out
}

func flattenXrayDNSToMap(data any) map[string]any {
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

	if v, ok := payload["servers"].([]any); ok {
		out["server"] = flattenDNSServers(v)
	}
	if v, ok := payload["hosts"].(map[string]any); ok {
		out["hosts"] = flattenStringMap(v)
	}
	if v, ok := payload["queryStrategy"].(string); ok {
		out["query_strategy"] = v
	}
	if v, ok := payload["tag"].(string); ok {
		out["tag"] = v
	}
	if v, ok := payload["disableCache"].(bool); ok {
		out["disable_cache"] = v
	}
	if v, ok := payload["disableFallback"].(bool); ok {
		out["disable_fallback"] = v
	}
	if v, ok := payload["disableFallbackIfMatch"].(bool); ok {
		out["disable_fallback_if_match"] = v
	}
	if v, ok := payload["clientIp"].(string); ok {
		out["client_ip"] = v
	}

	return out
}

// flattenDNSServers handles both string servers and object servers from API.
func flattenDNSServers(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		switch v := item.(type) {
		case string:
			out = append(out, map[string]any{
				"address": v,
			})
		case map[string]any:
			entry := map[string]any{}
			if addr, ok := v["address"].(string); ok {
				entry["address"] = addr
			}
			if p, ok := v["port"]; ok {
				entry["port"] = intValue(p)
			}
			if d, ok := v["domains"].([]any); ok {
				entry["domains"] = d
			}
			if e, ok := v["expectedIPs"].([]any); ok {
				entry["expect_ips"] = e
			}
			if u, ok := v["unexpectedIPs"].([]any); ok {
				entry["unexpected_ips"] = u
			}
			if s, ok := v["skipFallback"].(bool); ok {
				entry["skip_fallback"] = s
			}
			if q, ok := v["queryStrategy"].(string); ok {
				entry["query_strategy"] = q
			}
			if dc, ok := v["disableCache"].(bool); ok {
				entry["disable_cache"] = dc
			}
			if f, ok := v["finalQuery"].(bool); ok {
				entry["final_query"] = f
			}
			out = append(out, entry)
		}
	}
	return out
}
