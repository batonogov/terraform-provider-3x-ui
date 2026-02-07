package provider

import "encoding/json"

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
