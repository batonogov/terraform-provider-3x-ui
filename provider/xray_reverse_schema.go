package provider

import "encoding/json"

func buildXrayReverseJSON(d map[string]any) any {
	payload := map[string]any{}

	if v, ok := d["bridge"]; ok {
		if list, ok := v.([]any); ok {
			payload["bridges"] = expandReverseEntries(list)
		}
	}
	if v, ok := d["portal"]; ok {
		if list, ok := v.([]any); ok {
			payload["portals"] = expandReverseEntries(list)
		}
	}

	return payload
}

func expandReverseEntries(list []any) []any {
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
		if v, ok := m["domain"].(string); ok && v != "" {
			entry["domain"] = v
		}
		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	return out
}

func flattenXrayReverseToMap(data any) map[string]any {
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

	if v, ok := payload["bridges"].([]any); ok {
		out["bridge"] = flattenReverseEntries(v)
	}
	if v, ok := payload["portals"].([]any); ok {
		out["portal"] = flattenReverseEntries(v)
	}

	return out
}

func flattenReverseEntries(list []any) []any {
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
		if v, ok := m["domain"].(string); ok {
			entry["domain"] = v
		}
		out = append(out, entry)
	}
	return out
}
