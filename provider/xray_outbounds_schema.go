package provider

import "encoding/json"

func buildXrayOutboundsJSON(d map[string]any) any {
	v, ok := d["outbound"]
	if !ok {
		return []any{}
	}
	list, ok := v.([]any)
	if !ok {
		return []any{}
	}
	return expandOutbounds(list)
}

func expandOutbounds(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		protocol := ""

		if v, ok := m["tag"].(string); ok && v != "" {
			entry["tag"] = v
		}
		if v, ok := m["protocol"].(string); ok && v != "" {
			protocol = v
			entry["protocol"] = v
		}
		if v, ok := m["send_through"].(string); ok && v != "" {
			entry["sendThrough"] = v
		}

		// Mux
		if v, ok := m["mux"]; ok {
			if mux := expandOutboundMux(v.([]any)); mux != nil {
				entry["mux"] = mux
			}
		}

		// Protocol-specific settings
		if settings := expandOutboundSettings(m, protocol); settings != nil {
			entry["settings"] = settings
		}

		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	return out
}

func expandOutboundMux(list []any) map[string]any {
	if len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["enabled"].(bool); ok {
		out["enabled"] = v
	}
	if v, ok := item["concurrency"].(int); ok && v != 0 {
		out["concurrency"] = v
	}
	if v, ok := item["xudp_concurrency"].(int); ok && v != 0 {
		out["xudpConcurrency"] = v
	}
	if v, ok := item["xudp_proxy_udp443"].(string); ok && v != "" {
		out["xudpProxyUDP443"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func expandOutboundSettings(m map[string]any, protocol string) map[string]any {
	switch protocol {
	case "freedom":
		return expandFreedomSettings(m)
	case "blackhole":
		return expandBlackholeSettings(m)
	case "dns":
		return expandOutboundDNSSettings(m)
	case "vmess":
		return expandVmessOutSettings(m)
	case "vless":
		return expandVlessOutSettings(m)
	case "trojan":
		return expandTrojanOutSettings(m)
	case "shadowsocks":
		return expandShadowsocksOutSettings(m)
	case "socks":
		return expandSocksOutSettings(m)
	case "http":
		return expandHTTPOutSettings(m)
	case "wireguard":
		return expandWireguardOutSettings(m)
	case "hysteria", "hysteria2":
		return expandHysteriaOutSettings(m)
	default:
		return nil
	}
}

func expandFreedomSettings(m map[string]any) map[string]any {
	list, ok := m["freedom_settings"].([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["domain_strategy"].(string); ok && v != "" {
		out["domainStrategy"] = v
	}
	if v, ok := item["redirect"].(string); ok && v != "" {
		out["redirect"] = v
	}
	if v, ok := item["fragment"]; ok {
		if f := expandFreedomFragment(v.([]any)); f != nil {
			out["fragment"] = f
		}
	}
	if v, ok := item["noises"]; ok {
		if n := expandFreedomNoises(v.([]any)); n != nil {
			out["noises"] = n
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func expandFreedomFragment(list []any) map[string]any {
	if len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["packets"].(string); ok && v != "" {
		out["packets"] = v
	}
	if v, ok := item["length"].(string); ok && v != "" {
		out["length"] = v
	}
	if v, ok := item["interval"].(string); ok && v != "" {
		out["interval"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func expandFreedomNoises(list []any) []any {
	if len(list) == 0 {
		return nil
	}
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
		if v, ok := m["packet"].(string); ok && v != "" {
			entry["packet"] = v
		}
		if v, ok := m["delay"].(string); ok && v != "" {
			entry["delay"] = v
		}
		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func expandBlackholeSettings(m map[string]any) map[string]any {
	list, ok := m["blackhole_settings"].([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["response_type"].(string); ok && v != "" {
		out["response"] = map[string]any{"type": v}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func expandOutboundDNSSettings(m map[string]any) map[string]any {
	list, ok := m["dns_settings"].([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["network"].(string); ok && v != "" {
		out["network"] = v
	}
	if v, ok := item["address"].(string); ok && v != "" {
		out["address"] = v
	}
	if v, ok := item["port"].(int); ok && v != 0 {
		out["port"] = v
	}
	if v, ok := item["non_ip_query"].(string); ok && v != "" {
		out["nonIPQuery"] = v
	}
	if v, ok := item["block_types"].([]any); ok && len(v) > 0 {
		out["blockTypes"] = expandIntList(v)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func expandVmessOutSettings(m map[string]any) map[string]any {
	list, ok := m["vmess_settings"].([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	server := map[string]any{}
	if v, ok := item["address"].(string); ok && v != "" {
		server["address"] = v
	}
	if v, ok := item["port"].(int); ok && v != 0 {
		server["port"] = v
	}
	user := map[string]any{}
	if v, ok := item["id"].(string); ok && v != "" {
		user["id"] = v
	}
	if v, ok := item["security"].(string); ok && v != "" {
		user["security"] = v
	}
	if len(user) > 0 {
		server["users"] = []any{user}
	}
	return map[string]any{"vnext": []any{server}}
}

func expandVlessOutSettings(m map[string]any) map[string]any {
	list, ok := m["vless_settings"].([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	server := map[string]any{}
	if v, ok := item["address"].(string); ok && v != "" {
		server["address"] = v
	}
	if v, ok := item["port"].(int); ok && v != 0 {
		server["port"] = v
	}
	user := map[string]any{}
	if v, ok := item["id"].(string); ok && v != "" {
		user["id"] = v
	}
	if v, ok := item["flow"].(string); ok && v != "" {
		user["flow"] = v
	}
	if v, ok := item["encryption"].(string); ok && v != "" {
		user["encryption"] = v
	}
	if len(user) > 0 {
		server["users"] = []any{user}
	}
	return map[string]any{"vnext": []any{server}}
}

func expandTrojanOutSettings(m map[string]any) map[string]any {
	list, ok := m["trojan_settings"].([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	server := map[string]any{}
	if v, ok := item["address"].(string); ok && v != "" {
		server["address"] = v
	}
	if v, ok := item["port"].(int); ok && v != 0 {
		server["port"] = v
	}
	if v, ok := item["password"].(string); ok && v != "" {
		server["password"] = v
	}
	return map[string]any{"servers": []any{server}}
}

func expandShadowsocksOutSettings(m map[string]any) map[string]any {
	list, ok := m["shadowsocks_settings"].([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	server := map[string]any{}
	if v, ok := item["address"].(string); ok && v != "" {
		server["address"] = v
	}
	if v, ok := item["port"].(int); ok && v != 0 {
		server["port"] = v
	}
	if v, ok := item["password"].(string); ok && v != "" {
		server["password"] = v
	}
	if v, ok := item["method"].(string); ok && v != "" {
		server["method"] = v
	}
	if v, ok := item["uot"].(bool); ok {
		server["uot"] = v
	}
	if v, ok := item["uot_version"].(int); ok && v != 0 {
		server["UoTVersion"] = v
	}
	return map[string]any{"servers": []any{server}}
}

func expandSocksOutSettings(m map[string]any) map[string]any {
	list, ok := m["socks_settings"].([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	server := map[string]any{}
	if v, ok := item["address"].(string); ok && v != "" {
		server["address"] = v
	}
	if v, ok := item["port"].(int); ok && v != 0 {
		server["port"] = v
	}
	user := map[string]any{}
	if v, ok := item["user"].(string); ok && v != "" {
		user["user"] = v
	}
	if v, ok := item["pass"].(string); ok && v != "" {
		user["pass"] = v
	}
	if len(user) > 0 {
		server["users"] = []any{user}
	}
	return map[string]any{"servers": []any{server}}
}

func expandHTTPOutSettings(m map[string]any) map[string]any {
	list, ok := m["http_settings"].([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	server := map[string]any{}
	if v, ok := item["address"].(string); ok && v != "" {
		server["address"] = v
	}
	if v, ok := item["port"].(int); ok && v != 0 {
		server["port"] = v
	}
	user := map[string]any{}
	if v, ok := item["user"].(string); ok && v != "" {
		user["user"] = v
	}
	if v, ok := item["pass"].(string); ok && v != "" {
		user["pass"] = v
	}
	if len(user) > 0 {
		server["users"] = []any{user}
	}
	return map[string]any{"servers": []any{server}}
}

func expandWireguardOutSettings(m map[string]any) map[string]any {
	list, ok := m["wireguard_settings"].([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["mtu"].(int); ok && v != 0 {
		out["mtu"] = v
	}
	if v, ok := item["secret_key"].(string); ok && v != "" {
		out["secretKey"] = v
	}
	if v, ok := item["address"].([]any); ok && len(v) > 0 {
		out["address"] = expandStringList(v)
	}
	if v, ok := item["workers"].(int); ok && v != 0 {
		out["workers"] = v
	}
	if v, ok := item["domain_strategy"].(string); ok && v != "" {
		out["domainStrategy"] = v
	}
	if v, ok := item["reserved"].([]any); ok && len(v) > 0 {
		out["reserved"] = expandIntList(v)
	}
	if v, ok := item["no_kernel_tun"].(bool); ok {
		out["noKernelTun"] = v
	}
	if v, ok := item["peer"]; ok {
		if peers := expandWireguardPeers(v.([]any)); peers != nil {
			out["peers"] = peers
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func expandWireguardPeers(list []any) []any {
	if len(list) == 0 {
		return nil
	}
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		if v, ok := m["public_key"].(string); ok && v != "" {
			entry["publicKey"] = v
		}
		if v, ok := m["pre_shared_key"].(string); ok && v != "" {
			entry["preSharedKey"] = v
		}
		if v, ok := m["allowed_ips"].([]any); ok && len(v) > 0 {
			entry["allowedIPs"] = expandStringList(v)
		}
		if v, ok := m["endpoint"].(string); ok && v != "" {
			entry["endpoint"] = v
		}
		if v, ok := m["keep_alive"].(int); ok && v != 0 {
			entry["keepAlive"] = v
		}
		if len(entry) > 0 {
			out = append(out, entry)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func expandHysteriaOutSettings(m map[string]any) map[string]any {
	list, ok := m["hysteria_settings"].([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	server := map[string]any{}
	if v, ok := item["address"].(string); ok && v != "" {
		server["address"] = v
	}
	if v, ok := item["port"].(int); ok && v != 0 {
		server["port"] = v
	}
	if v, ok := item["version"].(int); ok && v != 0 {
		server["version"] = v
	}
	return map[string]any{"servers": []any{server}}
}

// --- Flatten ---

func flattenXrayOutboundsToMap(data any) map[string]any {
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

	out["outbound"] = flattenOutbounds(list)
	return out
}

func flattenOutbounds(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		protocol := ""

		if v, ok := m["tag"].(string); ok {
			entry["tag"] = v
		}
		if v, ok := m["protocol"].(string); ok {
			protocol = v
			entry["protocol"] = v
		}
		if v, ok := m["sendThrough"].(string); ok {
			entry["send_through"] = v
		}

		// Mux
		if v, ok := m["mux"].(map[string]any); ok {
			if mux := flattenOutboundMux(v); mux != nil {
				entry["mux"] = []any{mux}
			}
		}

		// Protocol-specific settings
		settings, _ := m["settings"].(map[string]any)
		flattenOutboundProtocolSettings(entry, settings, protocol)

		out = append(out, entry)
	}
	return out
}

func flattenOutboundMux(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := map[string]any{}
	if v, ok := in["enabled"].(bool); ok {
		out["enabled"] = v
	}
	if v, ok := in["concurrency"]; ok {
		out["concurrency"] = intValue(v)
	}
	if v, ok := in["xudpConcurrency"]; ok {
		out["xudp_concurrency"] = intValue(v)
	}
	if v, ok := in["xudpProxyUDP443"].(string); ok {
		out["xudp_proxy_udp443"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func flattenOutboundProtocolSettings(entry map[string]any, settings map[string]any, protocol string) {
	if settings == nil {
		return
	}
	switch protocol {
	case "freedom":
		entry["freedom_settings"] = []any{flattenFreedomSettings(settings)}
	case "blackhole":
		entry["blackhole_settings"] = []any{flattenBlackholeSettings(settings)}
	case "dns":
		entry["dns_settings"] = []any{flattenOutboundDNSSettings(settings)}
	case "vmess":
		entry["vmess_settings"] = []any{flattenVmessOutSettings(settings)}
	case "vless":
		entry["vless_settings"] = []any{flattenVlessOutSettings(settings)}
	case "trojan":
		entry["trojan_settings"] = []any{flattenTrojanOutSettings(settings)}
	case "shadowsocks":
		entry["shadowsocks_settings"] = []any{flattenShadowsocksOutSettings(settings)}
	case "socks":
		entry["socks_settings"] = []any{flattenSocksOutSettings(settings)}
	case "http":
		entry["http_settings"] = []any{flattenHTTPOutSettings(settings)}
	case "wireguard":
		entry["wireguard_settings"] = []any{flattenWireguardOutSettings(settings)}
	case "hysteria", "hysteria2":
		entry["hysteria_settings"] = []any{flattenHysteriaOutSettings(settings)}
	}
}

func flattenFreedomSettings(in map[string]any) map[string]any {
	out := map[string]any{}
	if v, ok := in["domainStrategy"].(string); ok {
		out["domain_strategy"] = v
	}
	if v, ok := in["redirect"].(string); ok {
		out["redirect"] = v
	}
	if v, ok := in["fragment"].(map[string]any); ok {
		f := map[string]any{}
		if p, ok := v["packets"].(string); ok {
			f["packets"] = p
		}
		if l, ok := v["length"].(string); ok {
			f["length"] = l
		}
		if i, ok := v["interval"].(string); ok {
			f["interval"] = i
		}
		out["fragment"] = []any{f}
	}
	if v, ok := in["noises"].([]any); ok {
		noises := make([]any, 0, len(v))
		for _, n := range v {
			nm, ok := n.(map[string]any)
			if !ok {
				continue
			}
			entry := map[string]any{}
			if t, ok := nm["type"].(string); ok {
				entry["type"] = t
			}
			if p, ok := nm["packet"].(string); ok {
				entry["packet"] = p
			}
			if d, ok := nm["delay"].(string); ok {
				entry["delay"] = d
			}
			noises = append(noises, entry)
		}
		out["noises"] = noises
	}
	return out
}

func flattenBlackholeSettings(in map[string]any) map[string]any {
	out := map[string]any{}
	if v, ok := in["response"].(map[string]any); ok {
		if t, ok := v["type"].(string); ok {
			out["response_type"] = t
		}
	}
	return out
}

func flattenOutboundDNSSettings(in map[string]any) map[string]any {
	out := map[string]any{}
	if v, ok := in["network"].(string); ok {
		out["network"] = v
	}
	if v, ok := in["address"].(string); ok {
		out["address"] = v
	}
	if v, ok := in["port"]; ok {
		out["port"] = intValue(v)
	}
	if v, ok := in["nonIPQuery"].(string); ok {
		out["non_ip_query"] = v
	}
	if v, ok := in["blockTypes"].([]any); ok {
		out["block_types"] = flattenIntList(v)
	}
	return out
}

func flattenVnextFirstUser(in map[string]any, fields ...string) map[string]any {
	out := map[string]any{}
	vnext, ok := in["vnext"].([]any)
	if !ok || len(vnext) == 0 {
		return out
	}
	server, ok := vnext[0].(map[string]any)
	if !ok {
		return out
	}
	if v, ok := server["address"].(string); ok {
		out["address"] = v
	}
	if v, ok := server["port"]; ok {
		out["port"] = intValue(v)
	}
	users, ok := server["users"].([]any)
	if ok && len(users) > 0 {
		user, ok := users[0].(map[string]any)
		if ok {
			for _, f := range fields {
				if v, ok := user[f]; ok {
					out[f] = v
				}
			}
		}
	}
	return out
}

func flattenVmessOutSettings(in map[string]any) map[string]any {
	return flattenVnextFirstUser(in, "id", "security")
}

func flattenVlessOutSettings(in map[string]any) map[string]any {
	return flattenVnextFirstUser(in, "id", "flow", "encryption")
}

func flattenServersFirst(in map[string]any, fields ...string) map[string]any {
	out := map[string]any{}
	servers, ok := in["servers"].([]any)
	if !ok || len(servers) == 0 {
		return out
	}
	server, ok := servers[0].(map[string]any)
	if !ok {
		return out
	}
	if v, ok := server["address"].(string); ok {
		out["address"] = v
	}
	if v, ok := server["port"]; ok {
		out["port"] = intValue(v)
	}
	for _, f := range fields {
		if v, ok := server[f]; ok {
			out[f] = v
		}
	}
	return out
}

func flattenTrojanOutSettings(in map[string]any) map[string]any {
	return flattenServersFirst(in, "password")
}

func flattenShadowsocksOutSettings(in map[string]any) map[string]any {
	out := flattenServersFirst(in, "password", "method")
	servers, ok := in["servers"].([]any)
	if ok && len(servers) > 0 {
		server, ok := servers[0].(map[string]any)
		if ok {
			if v, ok := server["uot"].(bool); ok {
				out["uot"] = v
			}
			if v, ok := server["UoTVersion"]; ok {
				out["uot_version"] = intValue(v)
			}
		}
	}
	return out
}

func flattenSocksOutSettings(in map[string]any) map[string]any {
	out := map[string]any{}
	servers, ok := in["servers"].([]any)
	if !ok || len(servers) == 0 {
		return out
	}
	server, ok := servers[0].(map[string]any)
	if !ok {
		return out
	}
	if v, ok := server["address"].(string); ok {
		out["address"] = v
	}
	if v, ok := server["port"]; ok {
		out["port"] = intValue(v)
	}
	users, ok := server["users"].([]any)
	if ok && len(users) > 0 {
		user, ok := users[0].(map[string]any)
		if ok {
			if v, ok := user["user"].(string); ok {
				out["user"] = v
			}
			if v, ok := user["pass"].(string); ok {
				out["pass"] = v
			}
		}
	}
	return out
}

func flattenHTTPOutSettings(in map[string]any) map[string]any {
	return flattenSocksOutSettings(in) // same structure
}

func flattenWireguardOutSettings(in map[string]any) map[string]any {
	out := map[string]any{}
	if v, ok := in["mtu"]; ok {
		out["mtu"] = intValue(v)
	}
	if v, ok := in["secretKey"].(string); ok {
		out["secret_key"] = v
	}
	if v, ok := in["address"].([]any); ok {
		out["address"] = v
	}
	if v, ok := in["workers"]; ok {
		out["workers"] = intValue(v)
	}
	if v, ok := in["domainStrategy"].(string); ok {
		out["domain_strategy"] = v
	}
	if v, ok := in["reserved"].([]any); ok {
		out["reserved"] = flattenIntList(v)
	}
	if v, ok := in["noKernelTun"].(bool); ok {
		out["no_kernel_tun"] = v
	}
	if v, ok := in["peers"].([]any); ok {
		out["peer"] = flattenWireguardPeers(v)
	}
	return out
}

func flattenWireguardPeers(list []any) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{}
		if v, ok := m["publicKey"].(string); ok {
			entry["public_key"] = v
		}
		if v, ok := m["preSharedKey"].(string); ok {
			entry["pre_shared_key"] = v
		}
		if v, ok := m["allowedIPs"].([]any); ok {
			entry["allowed_ips"] = v
		}
		if v, ok := m["endpoint"].(string); ok {
			entry["endpoint"] = v
		}
		if v, ok := m["keepAlive"]; ok {
			entry["keep_alive"] = intValue(v)
		}
		out = append(out, entry)
	}
	return out
}

func flattenHysteriaOutSettings(in map[string]any) map[string]any {
	out := flattenServersFirst(in)
	servers, ok := in["servers"].([]any)
	if ok && len(servers) > 0 {
		server, ok := servers[0].(map[string]any)
		if ok {
			if v, ok := server["version"]; ok {
				out["version"] = intValue(v)
			}
		}
	}
	return out
}
