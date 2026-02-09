package provider

import (
	"encoding/json"
	"strings"
)

func buildStreamSettingsJSON(item map[string]any) string {
	if item == nil {
		return "{}"
	}

	payload := map[string]any{}
	if v, ok := item["network"].(string); ok && v != "" {
		payload["network"] = v
	}
	if v, ok := item["security"].(string); ok && v != "" {
		payload["security"] = v
	}
	if v, ok := item["external_proxy"]; ok {
		if list, ok := v.([]any); ok {
			payload["externalProxy"] = expandExternalProxy(list)
		}
	}
	if v, ok := item["reality_settings"]; ok {
		if list, ok := v.([]any); ok {
			if rs := expandRealitySettings(list); rs != nil {
				payload["realitySettings"] = rs
			}
		}
	}
	if v, ok := item["tcp_settings"]; ok {
		if list, ok := v.([]any); ok {
			if ts := expandTCPSettings(list); ts != nil {
				payload["tcpSettings"] = ts
			}
		}
	}
	if v, ok := item["ws_settings"]; ok {
		if list, ok := v.([]any); ok {
			if ws := expandWSSettings(list); ws != nil {
				payload["wsSettings"] = ws
			}
		}
	}
	if v, ok := item["grpc_settings"]; ok {
		if list, ok := v.([]any); ok {
			if gs := expandGRPCSettings(list); gs != nil {
				payload["grpcSettings"] = gs
			}
		}
	}
	if v, ok := item["httpupgrade_settings"]; ok {
		if list, ok := v.([]any); ok {
			if hu := expandHTTPUpgradeSettings(list); hu != nil {
				payload["httpupgradeSettings"] = hu
			}
		}
	}
	if v, ok := item["xhttp_settings"]; ok {
		if list, ok := v.([]any); ok {
			if xh := expandXHTTPSettings(list); xh != nil {
				payload["xhttpSettings"] = xh
			}
		}
	}
	if v, ok := item["kcp_settings"]; ok {
		if list, ok := v.([]any); ok {
			if kcp := expandKCPSettings(list); kcp != nil {
				payload["kcpSettings"] = kcp
			}
		}
	}
	if v, ok := item["sockopt"]; ok {
		if list, ok := v.([]any); ok {
			if so := expandSockopt(list); so != nil {
				payload["sockopt"] = so
			}
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
	if v, ok := payload["wsSettings"].(map[string]any); ok {
		if ws := flattenWSSettings(v); ws != nil {
			out["ws_settings"] = []any{ws}
		}
	}
	if v, ok := payload["grpcSettings"].(map[string]any); ok {
		if gs := flattenGRPCSettings(v); gs != nil {
			out["grpc_settings"] = []any{gs}
		}
	}
	if v, ok := payload["httpupgradeSettings"].(map[string]any); ok {
		if hu := flattenHTTPUpgradeSettings(v); hu != nil {
			out["httpupgrade_settings"] = []any{hu}
		}
	}
	if v, ok := payload["xhttpSettings"].(map[string]any); ok {
		if xh := flattenXHTTPSettings(v); xh != nil {
			out["xhttp_settings"] = []any{xh}
		}
	}
	if v, ok := payload["kcpSettings"].(map[string]any); ok {
		if kcp := flattenKCPSettings(v); kcp != nil {
			out["kcp_settings"] = []any{kcp}
		}
	}
	if v, ok := payload["sockopt"].(map[string]any); ok {
		if so := flattenSockopt(v); so != nil {
			out["sockopt"] = []any{so}
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
		if v, ok := m["port"]; ok {
			if p := intValue(v); p != 0 {
				entry["port"] = p
			}
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
	if v, ok := item["show"]; ok {
		rs["show"] = boolValue(v)
	}
	if v, ok := item["xver"]; ok {
		rs["xver"] = intValue(v)
	}
	if v, ok := item["target"].(string); ok && v != "" {
		target = v
		rs["target"] = v
	}
	if v, ok := item["server_names"]; ok {
		if list, ok := v.([]any); ok {
			rs["serverNames"] = expandStringList(list)
		}
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
	if v, ok := item["max_timediff"]; ok {
		rs["maxTimediff"] = intValue(v)
	}
	if v, ok := item["short_ids"]; ok {
		if list, ok := v.([]any); ok {
			rs["shortIds"] = expandStringList(list)
		}
	}
	if v, ok := item["mldsa65_seed"].(string); ok && v != "" {
		rs["mldsa65Seed"] = v
	}
	if v, ok := item["settings"]; ok {
		if list, ok := v.([]any); ok {
			if s := expandRealityInnerSettings(list); s != nil {
				rs["settings"] = s
			}
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
	if v, ok := item["accept_proxy_protocol"]; ok {
		out["acceptProxyProtocol"] = boolValue(v)
	}
	// Support both nested header block and flat header_type string
	if v, ok := item["header"]; ok {
		if list, ok := v.([]any); ok {
			if h := expandTCPHeader(list); h != nil {
				out["header"] = h
			}
		}
	}
	if v, ok := item["header_type"].(string); ok && v != "" {
		out["header"] = map[string]any{"type": v}
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

// ---------------------------------------------------------------------------
// WebSocket
// ---------------------------------------------------------------------------

func expandWSSettings(list []any) map[string]any {
	if len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["path"].(string); ok && v != "" {
		out["path"] = v
	}
	if v, ok := item["headers"].(map[string]any); ok && len(v) > 0 {
		out["headers"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func flattenWSSettings(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := map[string]any{}
	if v, ok := in["path"].(string); ok {
		out["path"] = v
	}
	if v, ok := in["headers"].(map[string]any); ok {
		out["headers"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// ---------------------------------------------------------------------------
// gRPC
// ---------------------------------------------------------------------------

func expandGRPCSettings(list []any) map[string]any {
	if len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["service_name"].(string); ok && v != "" {
		out["serviceName"] = v
	}
	if v, ok := item["multi_mode"]; ok {
		out["multiMode"] = boolValue(v)
	}
	if v, ok := item["idle_timeout"]; ok {
		if n := intValue(v); n != 0 {
			out["idle_timeout"] = n
		}
	}
	if v, ok := item["health_check_timeout"]; ok {
		if n := intValue(v); n != 0 {
			out["health_check_timeout"] = n
		}
	}
	if v, ok := item["permit_without_stream"]; ok {
		out["permitWithoutStream"] = boolValue(v)
	}
	if v, ok := item["initial_windows_size"]; ok {
		if n := intValue(v); n != 0 {
			out["initial_windows_size"] = n
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func flattenGRPCSettings(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := map[string]any{}
	if v, ok := in["serviceName"].(string); ok {
		out["service_name"] = v
	}
	if v, ok := in["multiMode"].(bool); ok {
		out["multi_mode"] = v
	}
	if v, ok := in["idle_timeout"]; ok {
		out["idle_timeout"] = intValue(v)
	}
	if v, ok := in["health_check_timeout"]; ok {
		out["health_check_timeout"] = intValue(v)
	}
	if v, ok := in["permitWithoutStream"].(bool); ok {
		out["permit_without_stream"] = v
	}
	if v, ok := in["initial_windows_size"]; ok {
		out["initial_windows_size"] = intValue(v)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// ---------------------------------------------------------------------------
// HTTP Upgrade
// ---------------------------------------------------------------------------

func expandHTTPUpgradeSettings(list []any) map[string]any {
	if len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["path"].(string); ok && v != "" {
		out["path"] = v
	}
	if v, ok := item["host"].(string); ok && v != "" {
		out["host"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func flattenHTTPUpgradeSettings(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := map[string]any{}
	if v, ok := in["path"].(string); ok {
		out["path"] = v
	}
	if v, ok := in["host"].(string); ok {
		out["host"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// ---------------------------------------------------------------------------
// XHTTP
// ---------------------------------------------------------------------------

func expandXHTTPSettings(list []any) map[string]any {
	if len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["path"].(string); ok && v != "" {
		out["path"] = v
	}
	if v, ok := item["mode"].(string); ok && v != "" {
		out["mode"] = v
	}
	if v, ok := item["no_sse_header"]; ok {
		out["noSSEHeader"] = boolValue(v)
	}
	if v, ok := item["keep_alive_interval"]; ok {
		if n := intValue(v); n != 0 {
			out["keepAliveInterval"] = n
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func flattenXHTTPSettings(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := map[string]any{}
	if v, ok := in["path"].(string); ok {
		out["path"] = v
	}
	if v, ok := in["mode"].(string); ok {
		out["mode"] = v
	}
	if v, ok := in["noSSEHeader"].(bool); ok {
		out["no_sse_header"] = v
	}
	if v, ok := in["keepAliveInterval"]; ok {
		out["keep_alive_interval"] = intValue(v)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// ---------------------------------------------------------------------------
// KCP
// ---------------------------------------------------------------------------

func expandKCPSettings(list []any) map[string]any {
	if len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["mtu"]; ok {
		if n := intValue(v); n != 0 {
			out["mtu"] = n
		}
	}
	if v, ok := item["tti"]; ok {
		if n := intValue(v); n != 0 {
			out["tti"] = n
		}
	}
	if v, ok := item["uplink_capacity"]; ok {
		if n := intValue(v); n != 0 {
			out["uplinkCapacity"] = n
		}
	}
	if v, ok := item["downlink_capacity"]; ok {
		if n := intValue(v); n != 0 {
			out["downlinkCapacity"] = n
		}
	}
	if v, ok := item["congestion"]; ok {
		out["congestion"] = boolValue(v)
	}
	if v, ok := item["read_buffer_size"]; ok {
		if n := intValue(v); n != 0 {
			out["readBufferSize"] = n
		}
	}
	if v, ok := item["write_buffer_size"]; ok {
		if n := intValue(v); n != 0 {
			out["writeBufferSize"] = n
		}
	}
	if v, ok := item["header_type"].(string); ok && v != "" {
		out["header"] = map[string]any{"type": v}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func flattenKCPSettings(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := map[string]any{}
	if v, ok := in["mtu"]; ok {
		out["mtu"] = intValue(v)
	}
	if v, ok := in["tti"]; ok {
		out["tti"] = intValue(v)
	}
	if v, ok := in["uplinkCapacity"]; ok {
		out["uplink_capacity"] = intValue(v)
	}
	if v, ok := in["downlinkCapacity"]; ok {
		out["downlink_capacity"] = intValue(v)
	}
	if v, ok := in["congestion"].(bool); ok {
		out["congestion"] = v
	}
	if v, ok := in["readBufferSize"]; ok {
		out["read_buffer_size"] = intValue(v)
	}
	if v, ok := in["writeBufferSize"]; ok {
		out["write_buffer_size"] = intValue(v)
	}
	if v, ok := in["header"].(map[string]any); ok {
		if t, ok := v["type"].(string); ok {
			out["header_type"] = t
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// ---------------------------------------------------------------------------
// Sockopt
// ---------------------------------------------------------------------------

func expandSockopt(list []any) map[string]any {
	if len(list) == 0 {
		return nil
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]any{}
	if v, ok := item["mark"]; ok {
		if n := intValue(v); n != 0 {
			out["mark"] = n
		}
	}
	if v, ok := item["tcp_keep_alive_interval"]; ok {
		if n := intValue(v); n != 0 {
			out["tcpKeepAliveInterval"] = n
		}
	}
	if v, ok := item["tcp_no_delay"]; ok {
		out["tcpNoDelay"] = boolValue(v)
	}
	if v, ok := item["tfo_enable"]; ok {
		out["tcpFastOpen"] = boolValue(v)
	}
	if v, ok := item["tproxy"].(string); ok && v != "" {
		out["tproxy"] = v
	}
	if v, ok := item["domain_strategy"].(string); ok && v != "" {
		out["domainStrategy"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func flattenSockopt(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := map[string]any{}
	if v, ok := in["mark"]; ok {
		out["mark"] = intValue(v)
	}
	if v, ok := in["tcpKeepAliveInterval"]; ok {
		out["tcp_keep_alive_interval"] = intValue(v)
	}
	if v, ok := in["tcpNoDelay"].(bool); ok {
		out["tcp_no_delay"] = v
	}
	if v, ok := in["tcpFastOpen"].(bool); ok {
		out["tfo_enable"] = v
	}
	if v, ok := in["tproxy"].(string); ok {
		out["tproxy"] = v
	}
	if v, ok := in["domainStrategy"].(string); ok {
		out["domain_strategy"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
