package provider

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func xrayReverseSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"bridge": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"tag": {
					Type:     schema.TypeString,
					Required: true,
				},
				"domain": {
					Type:     schema.TypeString,
					Required: true,
				},
			}},
		},
		"portal": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"tag": {
					Type:     schema.TypeString,
					Required: true,
				},
				"domain": {
					Type:     schema.TypeString,
					Required: true,
				},
			}},
		},
	}
}

func buildXrayReverseJSON(d *schema.ResourceData) (map[string]any, error) {
	payload := map[string]any{}

	if v, ok := d.GetOk("bridge"); ok {
		payload["bridges"] = expandReverseEntries(v.([]any))
	}
	if v, ok := d.GetOk("portal"); ok {
		payload["portals"] = expandReverseEntries(v.([]any))
	}

	return payload, nil
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
