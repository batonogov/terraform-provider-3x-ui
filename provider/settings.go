package provider

import (
	"encoding/json"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func settingsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"clients": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"id": {
					Type:             schema.TypeString,
					Optional:         true,
					DiffSuppressFunc: suppressIfNewEmpty,
				},
				"email":    {Type: schema.TypeString, Optional: true},
				"enable":   {Type: schema.TypeBool, Optional: true},
				"flow":     {Type: schema.TypeString, Optional: true},
				"security": {Type: schema.TypeString, Optional: true},
				"password": {
					Type:             schema.TypeString,
					Optional:         true,
					DiffSuppressFunc: suppressIfNewEmpty,
				},
				"limit_ip":    {Type: schema.TypeInt, Optional: true},
				"total_gb":    {Type: schema.TypeInt, Optional: true},
				"expiry_time": {Type: schema.TypeInt, Optional: true},
				"tg_id":       {Type: schema.TypeInt, Optional: true},
				"sub_id":      {Type: schema.TypeString, Optional: true},
				"comment":     {Type: schema.TypeString, Optional: true},
				"reset":       {Type: schema.TypeInt, Optional: true},
				"created_at":  {Type: schema.TypeInt, Computed: true},
				"updated_at":  {Type: schema.TypeInt, Computed: true},
			}},
		},
		"decryption": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"encryption": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"fallbacks": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"name": {Type: schema.TypeString, Optional: true},
				"alpn": {Type: schema.TypeString, Optional: true},
				"path": {Type: schema.TypeString, Optional: true},
				"dest": {Type: schema.TypeString, Optional: true},
				"xver": {Type: schema.TypeInt, Optional: true},
			}},
		},
		"selected_auth": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"testseed": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeInt},
		},
		"method": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"password": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"network": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"iv_check": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"allow_transparent": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"auth": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"accounts": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"user": {Type: schema.TypeString, Optional: true},
				"pass": {Type: schema.TypeString, Optional: true},
			}},
		},
		"udp": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"ip": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"address": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"port": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"port_map": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"follow_redirect": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"mtu": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"secret_key": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"no_kernel_tun": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"peers": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"private_key": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"public_key": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"pre_shared_key": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"allowed_ips": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"keep_alive": {
					Type:     schema.TypeInt,
					Optional: true,
				},
			}},
		},
		"name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"user_level": {
			Type:     schema.TypeInt,
			Optional: true,
		},
	}
}

func buildSettingsJSON(d *schema.ResourceData) string {
	raw, ok := d.GetOk("settings")
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
	if v, ok := item["clients"]; ok {
		payload["clients"] = expandClients(v.([]any))
	}
	if v, ok := item["decryption"].(string); ok && v != "" {
		payload["decryption"] = v
	}
	if v, ok := item["encryption"].(string); ok && v != "" {
		payload["encryption"] = v
	}
	if v, ok := item["fallbacks"]; ok {
		payload["fallbacks"] = expandFallbacks(v.([]any))
	}
	if v, ok := item["selected_auth"].(string); ok && v != "" {
		payload["selectedAuth"] = v
	}
	if v, ok := item["testseed"]; ok {
		payload["testseed"] = expandIntList(v.([]any))
	}
	if v, ok := item["method"].(string); ok && v != "" {
		payload["method"] = v
	}
	if v, ok := item["password"].(string); ok && v != "" {
		payload["password"] = v
	}
	if v, ok := item["network"].(string); ok && v != "" {
		payload["network"] = v
	}
	if v, ok := item["iv_check"].(bool); ok {
		payload["ivCheck"] = v
	}
	if v, ok := item["allow_transparent"].(bool); ok {
		payload["allowTransparent"] = v
	}
	if v, ok := item["auth"].(string); ok && v != "" {
		payload["auth"] = v
	}
	if v, ok := item["accounts"]; ok {
		payload["accounts"] = expandAccounts(v.([]any))
	}
	if v, ok := item["udp"].(bool); ok {
		payload["udp"] = v
	}
	if v, ok := item["ip"].(string); ok && v != "" {
		payload["ip"] = v
	}
	if v, ok := item["address"].(string); ok && v != "" {
		payload["address"] = v
	}
	if v, ok := item["port"].(int); ok && v != 0 {
		payload["port"] = v
	}
	if v, ok := item["port_map"]; ok {
		payload["portMap"] = expandStringMap(v.(map[string]any))
	}
	if v, ok := item["follow_redirect"].(bool); ok {
		payload["followRedirect"] = v
	}
	if v, ok := item["mtu"].(int); ok && v != 0 {
		payload["mtu"] = v
	}
	if v, ok := item["secret_key"].(string); ok && v != "" {
		payload["secretKey"] = v
	}
	if v, ok := item["no_kernel_tun"].(bool); ok {
		payload["noKernelTun"] = v
	}
	if v, ok := item["peers"]; ok {
		payload["peers"] = expandPeers(v.([]any))
	}
	if v, ok := item["name"].(string); ok && v != "" {
		payload["name"] = v
	}
	if v, ok := item["user_level"].(int); ok && v != 0 {
		payload["userLevel"] = v
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

func flattenSettings(settings string) []any {
	if strings.TrimSpace(settings) == "" {
		return []any{}
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(settings), &payload); err != nil {
		return []any{}
	}
	out := map[string]any{}
	if v, ok := payload["clients"].([]any); ok {
		out["clients"] = flattenClients(v)
	}
	if v, ok := payload["decryption"].(string); ok {
		out["decryption"] = v
	}
	if v, ok := payload["encryption"].(string); ok {
		out["encryption"] = v
	}
	if v, ok := payload["fallbacks"].([]any); ok {
		out["fallbacks"] = flattenFallbacks(v)
	}
	if v, ok := payload["selectedAuth"].(string); ok {
		out["selected_auth"] = v
	}
	if v, ok := payload["testseed"].([]any); ok {
		out["testseed"] = v
	}
	if v, ok := payload["method"].(string); ok {
		out["method"] = v
	}
	if v, ok := payload["password"].(string); ok {
		out["password"] = v
	}
	if v, ok := payload["network"].(string); ok {
		out["network"] = v
	}
	if v, ok := payload["ivCheck"].(bool); ok {
		out["iv_check"] = v
	}
	if v, ok := payload["allowTransparent"].(bool); ok {
		out["allow_transparent"] = v
	}
	if v, ok := payload["auth"].(string); ok {
		out["auth"] = v
	}
	if v, ok := payload["accounts"].([]any); ok {
		out["accounts"] = flattenAccounts(v)
	}
	if v, ok := payload["udp"].(bool); ok {
		out["udp"] = v
	}
	if v, ok := payload["ip"].(string); ok {
		out["ip"] = v
	}
	if v, ok := payload["address"].(string); ok {
		out["address"] = v
	}
	if v, ok := payload["port"]; ok {
		out["port"] = intValue(v)
	}
	if v, ok := payload["portMap"].(map[string]any); ok {
		out["port_map"] = flattenStringMap(v)
	}
	if v, ok := payload["followRedirect"].(bool); ok {
		out["follow_redirect"] = v
	}
	if v, ok := payload["mtu"]; ok {
		out["mtu"] = intValue(v)
	}
	if v, ok := payload["secretKey"].(string); ok {
		out["secret_key"] = v
	}
	if v, ok := payload["noKernelTun"].(bool); ok {
		out["no_kernel_tun"] = v
	}
	if v, ok := payload["peers"].([]any); ok {
		out["peers"] = flattenPeers(v)
	}
	if v, ok := payload["name"].(string); ok {
		out["name"] = v
	}
	if v, ok := payload["userLevel"]; ok {
		out["user_level"] = intValue(v)
	}
	if len(out) == 0 {
		return []any{}
	}
	return []any{out}
}

func preserveInboundClientIDs(desired *Inbound, existing *Inbound) error {
	if desired == nil || existing == nil {
		return nil
	}
	if strings.TrimSpace(desired.Settings) == "" || strings.TrimSpace(existing.Settings) == "" {
		return nil
	}

	desiredSettings, err := ParseJSONField(desired.Settings)
	if err != nil {
		return err
	}
	existingSettings, err := ParseJSONField(existing.Settings)
	if err != nil {
		return err
	}

	desiredClients, ok := desiredSettings["clients"].([]any)
	if !ok || len(desiredClients) == 0 {
		return nil
	}
	existingClients, ok := existingSettings["clients"].([]any)
	if !ok || len(existingClients) == 0 {
		return nil
	}

	existingByEmail := map[string]map[string]any{}
	for _, c := range existingClients {
		m, ok := c.(map[string]any)
		if !ok {
			continue
		}
		email := stringValue(m["email"])
		if email != "" {
			existingByEmail[email] = m
		}
	}

	changed := false
	for i, c := range desiredClients {
		m, ok := c.(map[string]any)
		if !ok {
			continue
		}
		email := stringValue(m["email"])
		if email == "" {
			continue
		}
		existingClient := existingByEmail[email]
		if existingClient == nil {
			continue
		}

		switch desired.Protocol {
		case "vless", "vmess":
			if stringValue(m["id"]) == "" {
				if id := stringValue(existingClient["id"]); id != "" {
					m["id"] = id
					changed = true
				}
			}
		case "trojan":
			if stringValue(m["password"]) == "" {
				if pw := stringValue(existingClient["password"]); pw != "" {
					m["password"] = pw
					changed = true
				}
			}
		case "shadowsocks":
			if stringValue(m["password"]) == "" {
				if pw := stringValue(existingClient["password"]); pw != "" {
					m["password"] = pw
					changed = true
				}
			}
		}

		desiredClients[i] = m
	}

	if !changed {
		return nil
	}
	desiredSettings["clients"] = desiredClients
	updated, err := json.Marshal(desiredSettings)
	if err != nil {
		return err
	}
	desired.Settings = string(updated)
	return nil
}

func expandClients(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		if v, ok := m["id"].(string); ok && v != "" {
			entry["id"] = v
		}
		if v, ok := m["email"].(string); ok && v != "" {
			entry["email"] = v
		}
		if v, ok := m["enable"].(bool); ok {
			entry["enable"] = v
		}
		if v, ok := m["flow"].(string); ok && v != "" {
			entry["flow"] = v
		}
		if v, ok := m["security"].(string); ok && v != "" {
			entry["security"] = v
		}
		if v, ok := m["password"].(string); ok && v != "" {
			entry["password"] = v
		}
		if v, ok := m["limit_ip"].(int); ok && v != 0 {
			entry["limitIp"] = v
		}
		if v, ok := m["total_gb"].(int); ok && v != 0 {
			entry["totalGB"] = v
		}
		if v, ok := m["expiry_time"].(int); ok && v != 0 {
			entry["expiryTime"] = v
		}
		if v, ok := m["tg_id"]; ok {
			switch t := v.(type) {
			case int:
				entry["tgId"] = t
			case int64:
				entry["tgId"] = t
			case float64:
				entry["tgId"] = int64(t)
			case string:
				if t != "" {
					entry["tgId"] = t
				}
			}
		}
		if v, ok := m["sub_id"].(string); ok && v != "" {
			entry["subId"] = v
		}
		if v, ok := m["comment"].(string); ok && v != "" {
			entry["comment"] = v
		}
		if v, ok := m["reset"].(int); ok && v != 0 {
			entry["reset"] = v
		}
		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	return out
}

func flattenClients(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		if v, ok := m["id"].(string); ok {
			entry["id"] = v
		}
		if v, ok := m["email"].(string); ok {
			entry["email"] = v
		}
		if v, ok := m["enable"].(bool); ok {
			entry["enable"] = v
		}
		if v, ok := m["flow"].(string); ok {
			entry["flow"] = v
		}
		if v, ok := m["security"].(string); ok {
			entry["security"] = v
		}
		if v, ok := m["password"].(string); ok {
			entry["password"] = v
		}
		if v, ok := m["limitIp"]; ok {
			entry["limit_ip"] = intValue(v)
		}
		if v, ok := m["totalGB"]; ok {
			entry["total_gb"] = intValue(v)
		}
		if v, ok := m["expiryTime"]; ok {
			entry["expiry_time"] = intValue(v)
		}
		if v, ok := m["tgId"]; ok {
			switch t := v.(type) {
			case float64:
				entry["tg_id"] = int64(t)
			default:
				entry["tg_id"] = t
			}
		}
		if v, ok := m["subId"]; ok {
			entry["sub_id"] = v
		}
		if v, ok := m["comment"]; ok {
			entry["comment"] = v
		}
		if v, ok := m["reset"]; ok {
			entry["reset"] = intValue(v)
		}
		if v, ok := m["created_at"]; ok {
			entry["created_at"] = intValue(v)
		}
		if v, ok := m["updated_at"]; ok {
			entry["updated_at"] = intValue(v)
		}
		out = append(out, entry)
	}
	return out
}

func expandFallbacks(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		if v, ok := m["name"].(string); ok && v != "" {
			entry["name"] = v
		}
		if v, ok := m["alpn"].(string); ok && v != "" {
			entry["alpn"] = v
		}
		if v, ok := m["path"].(string); ok && v != "" {
			entry["path"] = v
		}
		if v, ok := m["dest"].(string); ok && v != "" {
			entry["dest"] = v
		}
		if v, ok := m["xver"].(int); ok {
			entry["xver"] = v
		}
		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	return out
}

func flattenFallbacks(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		if v, ok := m["name"].(string); ok {
			entry["name"] = v
		}
		if v, ok := m["alpn"].(string); ok {
			entry["alpn"] = v
		}
		if v, ok := m["path"].(string); ok {
			entry["path"] = v
		}
		if v, ok := m["dest"].(string); ok {
			entry["dest"] = v
		}
		if v, ok := m["xver"]; ok {
			entry["xver"] = intValue(v)
		}
		out = append(out, entry)
	}
	return out
}

func expandAccounts(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		if v, ok := m["user"].(string); ok && v != "" {
			entry["user"] = v
		}
		if v, ok := m["pass"].(string); ok && v != "" {
			entry["pass"] = v
		}
		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	return out
}

func flattenAccounts(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		if v, ok := m["user"].(string); ok {
			entry["user"] = v
		}
		if v, ok := m["pass"].(string); ok {
			entry["pass"] = v
		}
		out = append(out, entry)
	}
	return out
}

func expandPeers(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		if v, ok := m["private_key"].(string); ok && v != "" {
			entry["privateKey"] = v
		}
		if v, ok := m["public_key"].(string); ok && v != "" {
			entry["publicKey"] = v
		}
		if v, ok := m["pre_shared_key"].(string); ok && v != "" {
			entry["preSharedKey"] = v
		}
		if v, ok := m["allowed_ips"]; ok {
			entry["allowedIPs"] = expandStringList(v.([]any))
		}
		if v, ok := m["keep_alive"].(int); ok && v != 0 {
			entry["keepAlive"] = v
		}
		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	return out
}

func flattenPeers(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		if v, ok := m["privateKey"].(string); ok {
			entry["private_key"] = v
		}
		if v, ok := m["publicKey"].(string); ok {
			entry["public_key"] = v
		}
		if v, ok := m["preSharedKey"].(string); ok {
			entry["pre_shared_key"] = v
		}
		if v, ok := m["allowedIPs"].([]any); ok {
			entry["allowed_ips"] = v
		}
		if v, ok := m["keepAlive"]; ok {
			entry["keep_alive"] = intValue(v)
		}
		out = append(out, entry)
	}
	return out
}

func expandStringMap(in map[string]any) map[string]string {
	out := map[string]string{}
	for k, v := range in {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}

func flattenStringMap(in map[string]any) map[string]string {
	out := map[string]string{}
	for k, v := range in {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}

func expandIntList(list []any) []int {
	out := make([]int, 0, len(list))
	for _, v := range list {
		out = append(out, intValue(v))
	}
	return out
}

func suppressIfNewEmpty(k, old, new string, d *schema.ResourceData) bool {
	return strings.TrimSpace(new) == ""
}
