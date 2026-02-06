package provider

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func xrayRoutingSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"domain_strategy": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"domain_matcher": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"rule": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"type": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "field",
				},
				"domain": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"ip": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"port": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"source_port": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"network": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"source": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"user": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"inbound_tag": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"protocol": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"attrs": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"outbound_tag": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"balancer_tag": {
					Type:     schema.TypeString,
					Optional: true,
				},
			}},
		},
	}
}

func buildXrayRoutingJSON(d *schema.ResourceData) (map[string]any, error) {
	payload := map[string]any{}

	if v, ok := d.GetOk("domain_strategy"); ok {
		payload["domainStrategy"] = v.(string)
	}
	if v, ok := d.GetOk("domain_matcher"); ok {
		payload["domainMatcher"] = v.(string)
	}
	if v, ok := d.GetOk("rule"); ok {
		payload["rules"] = expandRoutingRules(v.([]any))
	}

	return payload, nil
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
