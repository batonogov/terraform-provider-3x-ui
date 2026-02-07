package provider

import "encoding/json"

func buildXrayBalancersJSON(d map[string]any) any {
	v, ok := d["balancer"]
	if !ok {
		return []any{}
	}
	list, ok := v.([]any)
	if !ok {
		return []any{}
	}
	return expandBalancers(list)
}

func expandBalancers(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		if v, ok := m["tag"].(string); ok && v != "" {
			entry["tag"] = v
		}
		if v, ok := m["selector"]; ok {
			if list, ok := v.([]any); ok {
				entry["selector"] = expandStringList(list)
			}
		}
		if v, ok := m["strategy"]; ok {
			if list, ok := v.([]any); ok {
				if s := expandBalancerStrategy(list); s != nil {
					entry["strategy"] = s
				}
			}
		}
		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	return out
}

func expandBalancerStrategy(list []any) map[string]any {
	if len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["type"].(string); ok && v != "" {
		out["type"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func flattenXrayBalancersToMap(data any) map[string]any {
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

	out["balancer"] = flattenBalancers(list)
	return out
}

func flattenBalancers(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		if v, ok := m["tag"].(string); ok {
			entry["tag"] = v
		}
		if v, ok := m["selector"].([]any); ok {
			entry["selector"] = v
		}
		if v, ok := m["strategy"].(map[string]any); ok {
			if s := flattenBalancerStrategy(v); s != nil {
				entry["strategy"] = []any{s}
			}
		}
		out = append(out, entry)
	}
	return out
}

func flattenBalancerStrategy(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := map[string]any{}
	if v, ok := in["type"].(string); ok {
		out["type"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
