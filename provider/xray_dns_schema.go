package provider

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func xrayDNSSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"server": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"address": {
					Type:     schema.TypeString,
					Required: true,
				},
				"port": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"domains": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"expect_ips": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"unexpected_ips": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"skip_fallback": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"query_strategy": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"disable_cache": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"final_query": {
					Type:     schema.TypeBool,
					Optional: true,
				},
			}},
		},
		"hosts": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"query_strategy": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"tag": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"disable_cache": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"disable_fallback": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"disable_fallback_if_match": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"client_ip": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}

func buildXrayDNSJSON(d *schema.ResourceData) (map[string]any, error) { //nolint:unparam // error required by buildFunc interface
	payload := map[string]any{}

	if v, ok := d.GetOk("server"); ok {
		payload["servers"] = expandDNSServers(v.([]any))
	}
	if v, ok := d.GetOk("hosts"); ok {
		payload["hosts"] = expandStringMap(v.(map[string]any))
	}
	if v, ok := d.GetOk("query_strategy"); ok {
		payload["queryStrategy"] = v.(string)
	}
	if v, ok := d.GetOk("tag"); ok {
		payload["tag"] = v.(string)
	}
	if v, ok := d.GetOkExists("disable_cache"); ok { //nolint:staticcheck // needed for zero-value booleans
		payload["disableCache"] = v.(bool)
	}
	if v, ok := d.GetOkExists("disable_fallback"); ok { //nolint:staticcheck // needed for zero-value booleans
		payload["disableFallback"] = v.(bool)
	}
	if v, ok := d.GetOkExists("disable_fallback_if_match"); ok { //nolint:staticcheck // needed for zero-value booleans
		payload["disableFallbackIfMatch"] = v.(bool)
	}
	if v, ok := d.GetOk("client_ip"); ok {
		payload["clientIp"] = v.(string)
	}

	return payload, nil
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
		if v, ok := m["port"].(int); ok && v != 0 {
			hasExtra = true
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
		if v, ok := m["skip_fallback"].(bool); ok && v {
			hasExtra = true
		}
		if v, ok := m["query_strategy"].(string); ok && v != "" {
			hasExtra = true
		}
		if v, ok := m["disable_cache"].(bool); ok && v {
			hasExtra = true
		}
		if v, ok := m["final_query"].(bool); ok && v {
			hasExtra = true
		}

		if !hasExtra {
			out = append(out, address)
			continue
		}

		entry := map[string]any{
			"address": address,
		}
		if v, ok := m["port"].(int); ok && v != 0 {
			entry["port"] = v
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
		if v, ok := m["skip_fallback"].(bool); ok {
			entry["skipFallback"] = v
		}
		if v, ok := m["query_strategy"].(string); ok && v != "" {
			entry["queryStrategy"] = v
		}
		if v, ok := m["disable_cache"].(bool); ok {
			entry["disableCache"] = v
		}
		if v, ok := m["final_query"].(bool); ok {
			entry["finalQuery"] = v
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
