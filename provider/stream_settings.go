package provider

import (
	"encoding/json"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func streamSettingsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"network": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"security": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"external_proxy": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"dest": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"port": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"remark": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"force_tls": {
					Type:     schema.TypeString,
					Optional: true,
				},
			}},
		},
		"reality_settings": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"show": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"xver": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"target": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					DiffSuppressFunc: suppressIfNewEmpty,
				},
				"server_names": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"private_key": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					Sensitive:        true,
					DiffSuppressFunc: suppressIfNewEmpty,
				},
				"min_client_ver": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"max_client_ver": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"max_timediff": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"short_ids": {
					Type:             schema.TypeList,
					Optional:         true,
					Computed:         true,
					DiffSuppressFunc: suppressIfNewEmpty,
					Elem:             &schema.Schema{Type: schema.TypeString},
				},
				"mldsa65_seed": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"settings": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{Schema: map[string]*schema.Schema{
						"public_key": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							Sensitive:        true,
							DiffSuppressFunc: suppressIfNewEmpty,
						},
						"fingerprint": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"server_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"spider_x": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"mldsa65_verify": {
							Type:     schema.TypeString,
							Optional: true,
						},
					}},
				},
			}},
		},
		"tcp_settings": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"accept_proxy_protocol": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"header": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
						},
					}},
				},
			}},
		},
	}
}

func buildStreamSettingsJSON(d *schema.ResourceData) string {
	raw, ok := d.GetOk("stream_settings")
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
	if v, ok := item["network"]; ok {
		payload["network"] = v.(string)
	}
	if v, ok := item["security"]; ok {
		payload["security"] = v.(string)
	}
	if v, ok := item["external_proxy"]; ok {
		payload["externalProxy"] = expandExternalProxy(v.([]any))
	}
	if v, ok := item["reality_settings"]; ok {
		if rs := expandRealitySettings(v.([]any)); rs != nil {
			payload["realitySettings"] = rs
		}
	}
	if v, ok := item["tcp_settings"]; ok {
		if ts := expandTCPSettings(v.([]any)); ts != nil {
			payload["tcpSettings"] = ts
		}
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

func flattenStreamSettings(stream string) []any {
	if strings.TrimSpace(stream) == "" {
		return []any{}
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(stream), &payload); err != nil {
		return []any{}
	}
	out := map[string]any{}
	if v, ok := payload["network"].(string); ok {
		out["network"] = v
	}
	if v, ok := payload["security"].(string); ok {
		out["security"] = v
	}
	if v, ok := payload["externalProxy"].([]any); ok {
		out["external_proxy"] = flattenExternalProxy(v)
	}
	if v, ok := payload["realitySettings"].(map[string]any); ok {
		if rs := flattenRealitySettings(v); rs != nil {
			out["reality_settings"] = []any{rs}
		}
	}
	if v, ok := payload["tcpSettings"].(map[string]any); ok {
		if ts := flattenTCPSettings(v); ts != nil {
			out["tcp_settings"] = []any{ts}
		}
	}
	if len(out) == 0 {
		return []any{}
	}
	return []any{out}
}

func expandExternalProxy(list []any) []any {
	if len(list) == 0 {
		return []any{}
	}
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		if v, ok := m["dest"].(string); ok && v != "" {
			entry["dest"] = v
		}
		if v, ok := m["port"].(int); ok && v != 0 {
			entry["port"] = v
		}
		if v, ok := m["remark"].(string); ok && v != "" {
			entry["remark"] = v
		}
		if v, ok := m["force_tls"].(string); ok && v != "" {
			entry["forceTls"] = v
		}
		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	return out
}

func flattenExternalProxy(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		if v, ok := m["dest"].(string); ok {
			entry["dest"] = v
		}
		if v, ok := m["port"]; ok {
			entry["port"] = intValue(v)
		}
		if v, ok := m["remark"].(string); ok {
			entry["remark"] = v
		}
		if v, ok := m["forceTls"].(string); ok {
			entry["force_tls"] = v
		}
		out = append(out, entry)
	}
	return out
}

func expandRealitySettings(list []any) map[string]any {
	if len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	rs := map[string]any{}
	target := ""
	if v, ok := item["show"].(bool); ok {
		rs["show"] = v
	}
	if v, ok := item["xver"].(int); ok {
		rs["xver"] = v
	}
	if v, ok := item["target"].(string); ok && v != "" {
		target = v
		rs["target"] = v
	}
	if v, ok := item["server_names"]; ok {
		rs["serverNames"] = expandStringList(v.([]any))
	}
	if v, ok := item["private_key"].(string); ok && v != "" {
		rs["privateKey"] = v
	}
	if v, ok := item["min_client_ver"].(string); ok && v != "" {
		rs["minClientVer"] = v
	}
	if v, ok := item["max_client_ver"].(string); ok && v != "" {
		rs["maxClientVer"] = v
	}
	if v, ok := item["max_timediff"].(int); ok {
		rs["maxTimediff"] = v
	}
	if v, ok := item["short_ids"]; ok {
		rs["shortIds"] = expandStringList(v.([]any))
	}
	if v, ok := item["mldsa65_seed"].(string); ok && v != "" {
		rs["mldsa65Seed"] = v
	}
	if v, ok := item["settings"]; ok {
		if s := expandRealityInnerSettings(v.([]any)); s != nil {
			rs["settings"] = s
		}
	}
	if !hasRealityServerNames(rs) {
		if target != "" {
			host := strings.Split(target, ":")[0]
			if host != "" {
				rs["serverNames"] = []any{host}
			}
		}
	}
	if !hasRealityServerNames(rs) {
		rs["target"] = "www.apple.com:443"
		rs["serverNames"] = []any{"www.apple.com", "apple.com"}
	}
	if len(rs) == 0 {
		return nil
	}
	return rs
}

func expandRealityInnerSettings(list []any) map[string]any {
	if len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["public_key"].(string); ok && v != "" {
		out["publicKey"] = v
	}
	if v, ok := item["fingerprint"].(string); ok && v != "" {
		out["fingerprint"] = v
	}
	if v, ok := item["server_name"].(string); ok && v != "" {
		out["serverName"] = v
	}
	if v, ok := item["spider_x"].(string); ok && v != "" {
		out["spiderX"] = v
	}
	if v, ok := item["mldsa65_verify"].(string); ok && v != "" {
		out["mldsa65Verify"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func flattenRealitySettings(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := map[string]any{}
	if v, ok := in["show"].(bool); ok {
		out["show"] = v
	}
	if v, ok := in["xver"]; ok {
		out["xver"] = intValue(v)
	}
	if v, ok := in["target"].(string); ok {
		out["target"] = v
	}
	if v, ok := in["serverNames"].([]any); ok {
		out["server_names"] = v
	}
	if v, ok := in["privateKey"].(string); ok {
		out["private_key"] = v
	}
	if v, ok := in["minClientVer"].(string); ok {
		out["min_client_ver"] = v
	}
	if v, ok := in["maxClientVer"].(string); ok {
		out["max_client_ver"] = v
	}
	if v, ok := in["maxTimediff"]; ok {
		out["max_timediff"] = intValue(v)
	}
	if v, ok := in["shortIds"].([]any); ok {
		out["short_ids"] = v
	}
	if v, ok := in["mldsa65Seed"].(string); ok {
		out["mldsa65_seed"] = v
	}
	if v, ok := in["settings"].(map[string]any); ok {
		if s := flattenRealityInnerSettings(v); s != nil {
			out["settings"] = []any{s}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func flattenRealityInnerSettings(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := map[string]any{}
	if v, ok := in["publicKey"].(string); ok {
		out["public_key"] = v
	}
	if v, ok := in["fingerprint"].(string); ok {
		out["fingerprint"] = v
	}
	if v, ok := in["serverName"].(string); ok {
		out["server_name"] = v
	}
	if v, ok := in["spiderX"].(string); ok {
		out["spider_x"] = v
	}
	if v, ok := in["mldsa65Verify"].(string); ok {
		out["mldsa65_verify"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func expandTCPSettings(list []any) map[string]any {
	if len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["accept_proxy_protocol"].(bool); ok {
		out["acceptProxyProtocol"] = v
	}
	if v, ok := item["header"]; ok {
		if h := expandTCPHeader(v.([]any)); h != nil {
			out["header"] = h
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func expandTCPHeader(list []any) map[string]any {
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

func flattenTCPSettings(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := map[string]any{}
	if v, ok := in["acceptProxyProtocol"].(bool); ok {
		out["accept_proxy_protocol"] = v
	}
	if v, ok := in["header"].(map[string]any); ok {
		if h := flattenTCPHeader(v); h != nil {
			out["header"] = []any{h}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func flattenTCPHeader(in map[string]any) map[string]any {
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
