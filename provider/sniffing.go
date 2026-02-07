package provider

import (
	"encoding/json"
	"strings"
)

func buildSniffingJSON(item map[string]any) string {
	if item == nil {
		return "{}"
	}

	payload := map[string]any{}
	if v, ok := item["enabled"]; ok {
		payload["enabled"] = boolValue(v)
	}
	if v, ok := item["dest_override"]; ok {
		if list, ok := v.([]any); ok {
			payload["destOverride"] = expandStringList(list)
		}
	}
	if v, ok := item["metadata_only"]; ok {
		payload["metadataOnly"] = boolValue(v)
	}
	if v, ok := item["route_only"]; ok {
		payload["routeOnly"] = boolValue(v)
	}

	if len(payload) == 0 {
		return "{}"
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func flattenSniffing(sniffing string) []any {
	if strings.TrimSpace(sniffing) == "" {
		return []any{}
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(sniffing), &payload); err != nil {
		return []any{}
	}
	out := map[string]any{}
	if v, ok := payload["enabled"].(bool); ok {
		out["enabled"] = v
	}
	if v, ok := payload["destOverride"].([]any); ok {
		out["dest_override"] = v
	}
	if v, ok := payload["metadataOnly"].(bool); ok {
		out["metadata_only"] = v
	}
	if v, ok := payload["routeOnly"].(bool); ok {
		out["route_only"] = v
	}
	if len(out) == 0 {
		return []any{}
	}
	return []any{out}
}

func expandStringList(list []any) []string {
	out := make([]string, 0, len(list))
	for _, v := range list {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
