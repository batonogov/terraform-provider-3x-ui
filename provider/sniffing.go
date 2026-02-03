package provider

import (
	"encoding/json"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func buildSniffingJSON(d *schema.ResourceData) string {
	raw, ok := d.GetOk("sniffing")
	if !ok {
		return "{}"
	}
	list, ok := raw.([]any)
	if !ok || len(list) == 0 {
		return "{}"
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return "{}"
	}

	payload := map[string]any{}
	if v, ok := item["enabled"]; ok {
		payload["enabled"] = v.(bool)
	}
	if v, ok := item["dest_override"]; ok {
		payload["destOverride"] = expandStringList(v.([]any))
	}
	if v, ok := item["metadata_only"]; ok {
		payload["metadataOnly"] = v.(bool)
	}
	if v, ok := item["route_only"]; ok {
		payload["routeOnly"] = v.(bool)
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
