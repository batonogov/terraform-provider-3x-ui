package provider

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDeepMergeJSON_Flat(t *testing.T) {
	dst := map[string]any{"a": "1"}
	src := map[string]any{"b": "2"}
	result := deepMergeJSON(dst, src)
	if result["a"] != "1" || result["b"] != "2" {
		t.Fatalf("unexpected result: %v", result)
	}
}

func TestDeepMergeJSON_Overlap(t *testing.T) {
	dst := map[string]any{"a": "old"}
	src := map[string]any{"a": "new"}
	result := deepMergeJSON(dst, src)
	if result["a"] != "new" {
		t.Fatalf("expected overwrite, got %v", result["a"])
	}
}

func TestDeepMergeJSON_Nested(t *testing.T) {
	dst := map[string]any{"a": map[string]any{"x": 1}}
	src := map[string]any{"a": map[string]any{"y": 2}}
	result := deepMergeJSON(dst, src)
	inner, ok := result["a"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested map")
	}
	if inner["x"] != 1 || inner["y"] != 2 {
		t.Fatalf("unexpected nested result: %v", inner)
	}
}

func TestDeepMergeJSON_NilDst(t *testing.T) {
	src := map[string]any{"a": "1"}
	result := deepMergeJSON(nil, src)
	if result["a"] != "1" {
		t.Fatalf("unexpected result: %v", result)
	}
}

func TestDeepMergeJSON_MapReplacedByScalar(t *testing.T) {
	dst := map[string]any{"a": map[string]any{"x": 1}}
	src := map[string]any{"a": "scalar"}
	result := deepMergeJSON(dst, src)
	if result["a"] != "scalar" {
		t.Fatalf("expected scalar, got %v", result["a"])
	}
}

func TestSetJSONPath_SingleLevel(t *testing.T) {
	root := map[string]any{}
	setJSONPath(root, []string{"key"}, "val")
	if root["key"] != "val" {
		t.Fatalf("unexpected: %v", root)
	}
}

func TestSetJSONPath_MultiLevel(t *testing.T) {
	root := map[string]any{}
	setJSONPath(root, []string{"a", "b", "c"}, 42)
	a, _ := root["a"].(map[string]any)
	b, _ := a["b"].(map[string]any)
	if b["c"] != 42 {
		t.Fatalf("unexpected: %v", root)
	}
}

func TestSetJSONPath_CreateIntermediates(t *testing.T) {
	root := map[string]any{"x": "keep"}
	setJSONPath(root, []string{"a", "b"}, "new")
	if root["x"] != "keep" {
		t.Fatalf("existing key lost")
	}
	a, _ := root["a"].(map[string]any)
	if a["b"] != "new" {
		t.Fatalf("unexpected: %v", root)
	}
}

func TestSetJSONPath_Overwrite(t *testing.T) {
	root := map[string]any{"a": "old"}
	setJSONPath(root, []string{"a"}, "new")
	if root["a"] != "new" {
		t.Fatalf("expected overwrite")
	}
}

func TestGetJSONPath_Existing(t *testing.T) {
	root := map[string]any{"a": map[string]any{"b": "val"}}
	got := getJSONPath(root, []string{"a", "b"})
	if got != "val" {
		t.Fatalf("expected val, got %v", got)
	}
}

func TestGetJSONPath_Missing(t *testing.T) {
	root := map[string]any{"a": "val"}
	got := getJSONPath(root, []string{"b"})
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestGetJSONPath_NonMapIntermediate(t *testing.T) {
	root := map[string]any{"a": "string"}
	got := getJSONPath(root, []string{"a", "b"})
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestCloneJSONMap_Nil(t *testing.T) {
	result := cloneJSONMap(nil)
	if result == nil || len(result) != 0 {
		t.Fatalf("expected empty map, got %v", result)
	}
}

func TestCloneJSONMap_Independence(t *testing.T) {
	original := map[string]any{"a": "1", "b": "2"}
	clone := cloneJSONMap(original)
	clone["a"] = "changed"
	if original["a"] != "1" {
		t.Fatalf("clone modified original")
	}
}

func TestDeepEqualJSON_EqualMaps(t *testing.T) {
	a := map[string]any{"x": float64(1)}
	b := map[string]any{"x": float64(1)}
	if !deepEqualJSON(a, b) {
		t.Fatalf("expected equal")
	}
}

func TestDeepEqualJSON_DifferentMaps(t *testing.T) {
	a := map[string]any{"x": float64(1)}
	b := map[string]any{"x": float64(2)}
	if deepEqualJSON(a, b) {
		t.Fatalf("expected not equal")
	}
}

func TestDeepEqualJSON_Arrays(t *testing.T) {
	a := []any{float64(1), float64(2)}
	b := []any{float64(1), float64(2)}
	if !deepEqualJSON(a, b) {
		t.Fatalf("expected equal")
	}
}

func TestDeepEqualJSON_DifferentTypes(t *testing.T) {
	if deepEqualJSON("string", float64(1)) {
		t.Fatalf("expected not equal for different types")
	}
}

func TestExtractXraySection_MergeRoot(t *testing.T) {
	current := map[string]any{
		"log":       map[string]any{"loglevel": "debug"},
		"policy":    map[string]any{},
		"api":       map[string]any{"tag": "api"},
		"stats":     map[string]any{},
		"dns":       map[string]any{"servers": []any{}},
		"outbounds": []any{},
	}
	result := extractXraySection(current, xraySectionBasics)
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map")
	}
	if _, ok := m["log"]; !ok {
		t.Fatalf("missing log")
	}
	if _, ok := m["api"]; !ok {
		t.Fatalf("missing api")
	}
	if _, ok := m["stats"]; !ok {
		t.Fatalf("missing stats")
	}
	if _, ok := m["dns"]; ok {
		t.Fatalf("dns should not be in basics")
	}
	if _, ok := m["outbounds"]; ok {
		t.Fatalf("outbounds should not be in basics")
	}
}

func TestExtractXraySection_SetPath(t *testing.T) {
	current := map[string]any{"dns": map[string]any{"servers": []any{"8.8.8.8"}}}
	result := extractXraySection(current, xraySectionDNS)
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map")
	}
	if _, ok := m["servers"]; !ok {
		t.Fatalf("missing servers")
	}
}

func TestApplyXraySection_MergeRoot(t *testing.T) {
	current := map[string]any{"log": map[string]any{"loglevel": "info"}}
	desired := map[string]any{"log": map[string]any{"loglevel": "debug"}}
	result, err := applyXraySection(current, desired, xraySectionBasics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	log, _ := result["log"].(map[string]any)
	if log["loglevel"] != "debug" {
		t.Fatalf("expected debug, got %v", log["loglevel"])
	}
}

func TestApplyXraySection_SetPath(t *testing.T) {
	current := map[string]any{}
	desired := map[string]any{"servers": []any{"8.8.8.8"}}
	result, err := applyXraySection(current, desired, xraySectionDNS)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dns, ok := result["dns"].(map[string]any)
	if !ok {
		t.Fatalf("expected dns key")
	}
	if _, ok := dns["servers"]; !ok {
		t.Fatalf("missing servers in dns")
	}
}

func TestApplyXraySection_MergeRoot_NotObject(t *testing.T) {
	_, err := applyXraySection(map[string]any{}, "string", xraySectionBasics)
	if err == nil {
		t.Fatalf("expected error for non-object")
	}
}

func TestApplyXraySection_SetPath_EmptyPath(t *testing.T) {
	section := xraySection{id: "test", mode: xraySectionSetPath, path: []string{}}
	_, err := applyXraySection(map[string]any{}, map[string]any{}, section)
	if err == nil {
		t.Fatalf("expected error for empty path")
	}
}

// --- Build/Flatten unit tests ---

func TestFlattenXrayReverseToMap(t *testing.T) {
	data := map[string]any{
		"bridges": []any{
			map[string]any{"tag": "b1", "domain": "test.com"},
		},
		"portals": []any{
			map[string]any{"tag": "p1", "domain": "test.com"},
		},
	}
	result := flattenXrayReverseToMap(data)
	bridges, ok := result["bridge"].([]any)
	if !ok || len(bridges) != 1 {
		t.Fatalf("expected 1 bridge, got %v", result["bridge"])
	}
	b := bridges[0].(map[string]any)
	if b["tag"] != "b1" {
		t.Fatalf("expected tag b1, got %v", b["tag"])
	}
}

func TestFlattenXrayBalancersToMap(t *testing.T) {
	data := []any{
		map[string]any{
			"tag":      "bal1",
			"selector": []any{"proxy-*"},
			"strategy": map[string]any{"type": "leastPing"},
		},
	}
	result := flattenXrayBalancersToMap(data)
	balancers, ok := result["balancer"].([]any)
	if !ok || len(balancers) != 1 {
		t.Fatalf("expected 1 balancer, got %v", result["balancer"])
	}
}

func TestFlattenXrayDNSToMap(t *testing.T) {
	data := map[string]any{
		"servers": []any{
			"8.8.8.8",
			map[string]any{
				"address":     "localhost",
				"port":        float64(53),
				"domains":     []any{"geosite:cn"},
				"expectedIPs": []any{"geoip:cn"},
			},
		},
		"queryStrategy": "UseIP",
	}
	result := flattenXrayDNSToMap(data)
	servers, ok := result["server"].([]any)
	if !ok || len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %v", result["server"])
	}
	// First server is string-only
	s0 := servers[0].(map[string]any)
	if s0["address"] != "8.8.8.8" {
		t.Fatalf("expected 8.8.8.8, got %v", s0["address"])
	}
	// Second server is object
	s1 := servers[1].(map[string]any)
	if s1["address"] != "localhost" {
		t.Fatalf("expected localhost, got %v", s1["address"])
	}
	if result["query_strategy"] != "UseIP" {
		t.Fatalf("expected UseIP, got %v", result["query_strategy"])
	}
}

func TestExpandDNSServers_StringOnly(t *testing.T) {
	list := []any{
		map[string]any{"address": "8.8.8.8"},
	}
	result := expandDNSServers(list)
	if len(result) != 1 {
		t.Fatalf("expected 1 server")
	}
	// Should be a plain string
	if s, ok := result[0].(string); !ok || s != "8.8.8.8" {
		t.Fatalf("expected string 8.8.8.8, got %v", result[0])
	}
}

func TestExpandDNSServers_WithPort(t *testing.T) {
	list := []any{
		map[string]any{"address": "localhost", "port": 53},
	}
	result := expandDNSServers(list)
	if len(result) != 1 {
		t.Fatalf("expected 1 server")
	}
	m, ok := result[0].(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", result[0])
	}
	if m["address"] != "localhost" || m["port"] != 53 {
		t.Fatalf("unexpected: %v", m)
	}
}

func TestFlattenXrayRoutingToMap(t *testing.T) {
	data := map[string]any{
		"domainStrategy": "AsIs",
		"rules": []any{
			map[string]any{
				"type":        "field",
				"ip":          []any{"geoip:private"},
				"outboundTag": "blocked",
			},
		},
	}
	result := flattenXrayRoutingToMap(data)
	if result["domain_strategy"] != "AsIs" {
		t.Fatalf("expected AsIs, got %v", result["domain_strategy"])
	}
	rules, ok := result["rule"].([]any)
	if !ok || len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %v", result["rule"])
	}
	r := rules[0].(map[string]any)
	if r["outbound_tag"] != "blocked" {
		t.Fatalf("expected blocked, got %v", r["outbound_tag"])
	}
}

func TestFlattenXrayBasicsToMap(t *testing.T) {
	data := map[string]any{
		"log": map[string]any{
			"loglevel": "warning",
			"dnsLog":   false,
		},
		"api": map[string]any{
			"tag":      "api",
			"services": []any{"HandlerService", "StatsService"},
		},
		"stats": map[string]any{},
	}
	result := flattenXrayBasicsToMap(data)
	log, ok := result["log"].(map[string]any)
	if !ok {
		t.Fatalf("expected log map, got %v", result["log"])
	}
	if log["loglevel"] != "warning" {
		t.Fatalf("expected warning, got %v", log["loglevel"])
	}
	if _, ok := result["stats"]; !ok {
		t.Fatalf("expected stats block")
	}
}

func TestFlattenXrayOutboundsToMap(t *testing.T) {
	data := []any{
		map[string]any{
			"tag":      "direct",
			"protocol": "freedom",
			"settings": map[string]any{
				"domainStrategy": "AsIs",
			},
		},
		map[string]any{
			"tag":      "blocked",
			"protocol": "blackhole",
			"settings": map[string]any{
				"response": map[string]any{"type": "none"},
			},
		},
	}
	result := flattenXrayOutboundsToMap(data)
	outbounds, ok := result["outbound"].([]any)
	if !ok || len(outbounds) != 2 {
		t.Fatalf("expected 2 outbounds, got %v", result["outbound"])
	}

	o0 := outbounds[0].(map[string]any)
	if o0["tag"] != "direct" || o0["protocol"] != "freedom" {
		t.Fatalf("unexpected first outbound: %v", o0)
	}
	freedomList, ok := o0["freedom_settings"].([]any)
	if !ok || len(freedomList) != 1 {
		t.Fatalf("expected freedom_settings, got %v", o0["freedom_settings"])
	}
	freedom := freedomList[0].(map[string]any)
	if freedom["domain_strategy"] != "AsIs" {
		t.Fatalf("expected AsIs, got %v", freedom["domain_strategy"])
	}

	o1 := outbounds[1].(map[string]any)
	bhList, ok := o1["blackhole_settings"].([]any)
	if !ok || len(bhList) != 1 {
		t.Fatalf("expected blackhole_settings, got %v", o1["blackhole_settings"])
	}
	bh := bhList[0].(map[string]any)
	if bh["response_type"] != "none" {
		t.Fatalf("expected none, got %v", bh["response_type"])
	}
}

func TestExpandReverseEntries(t *testing.T) {
	list := []any{
		map[string]any{"tag": "b1", "domain": "test.com"},
	}
	result := expandReverseEntries(list)
	if len(result) != 1 {
		t.Fatalf("expected 1 entry")
	}
	m := result[0].(map[string]any)
	if m["tag"] != "b1" || m["domain"] != "test.com" {
		t.Fatalf("unexpected: %v", m)
	}
}

func TestExpandRoutingRules(t *testing.T) {
	list := []any{
		map[string]any{
			"type":         "field",
			"ip":           []any{"geoip:private"},
			"outbound_tag": "blocked",
		},
	}
	result := expandRoutingRules(list)
	if len(result) != 1 {
		t.Fatalf("expected 1 rule")
	}
	r := result[0].(map[string]any)
	if r["type"] != "field" {
		t.Fatalf("expected field type")
	}
	if r["outboundTag"] != "blocked" {
		t.Fatalf("expected outboundTag blocked, got %v", r["outboundTag"])
	}
	ips, ok := r["ip"].([]string)
	if !ok || len(ips) != 1 || ips[0] != "geoip:private" {
		t.Fatalf("unexpected ips: %v", r["ip"])
	}
}

func TestFlattenWireguardOutSettings(t *testing.T) {
	in := map[string]any{
		"secretKey":      "test-key",
		"address":        []any{"10.0.0.2/32"},
		"mtu":            float64(1420),
		"workers":        float64(2),
		"domainStrategy": "ForceIPv6v4",
		"reserved":       []any{float64(1), float64(2), float64(3)},
		"noKernelTun":    false,
		"peers": []any{
			map[string]any{
				"publicKey":  "pub-key",
				"endpoint":   "engage.cloudflareclient.com:2408",
				"allowedIPs": []any{"0.0.0.0/0", "::/0"},
				"keepAlive":  float64(30),
			},
		},
	}
	result := flattenWireguardOutSettings(in)
	if result["secret_key"] != "test-key" {
		t.Fatalf("expected test-key, got %v", result["secret_key"])
	}
	if result["domain_strategy"] != "ForceIPv6v4" {
		t.Fatalf("expected ForceIPv6v4, got %v", result["domain_strategy"])
	}
	peers, ok := result["peer"].([]any)
	if !ok || len(peers) != 1 {
		t.Fatalf("expected 1 peer, got %v", result["peer"])
	}
	p := peers[0].(map[string]any)
	if p["public_key"] != "pub-key" {
		t.Fatalf("expected pub-key, got %v", p["public_key"])
	}
	reserved, ok := result["reserved"].([]int)
	if !ok || !reflect.DeepEqual(reserved, []int{1, 2, 3}) {
		t.Fatalf("expected [1,2,3], got %v", result["reserved"])
	}
}

func TestFlattenBasicsPolicyLevels(t *testing.T) {
	in := map[string]any{
		"0": map[string]any{
			"handshake":         float64(4),
			"connIdle":          float64(300),
			"statsUserUplink":   false,
			"statsUserDownlink": false,
		},
	}
	result := flattenBasicsPolicyLevels(in)
	if len(result) != 1 {
		t.Fatalf("expected 1 level, got %d", len(result))
	}
	level := result[0].(map[string]any)
	if level["id"] != 0 {
		t.Fatalf("expected id 0, got %v", level["id"])
	}
	if level["handshake"] != 4 {
		t.Fatalf("expected handshake 4, got %v", level["handshake"])
	}
}

// --- Expand unit tests ---

func TestExpandBasicsLog(t *testing.T) {
	item := map[string]any{
		"loglevel": "warning",
		"access":   "/var/log/access.log",
		"error":    "/var/log/error.log",
		"dns_log":  true,
	}
	result := expandBasicsLog(item)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result["loglevel"] != "warning" {
		t.Fatalf("expected warning, got %v", result["loglevel"])
	}
	if result["access"] != "/var/log/access.log" {
		t.Fatalf("expected access path, got %v", result["access"])
	}
	if result["error"] != "/var/log/error.log" {
		t.Fatalf("expected error path, got %v", result["error"])
	}
	if result["dnsLog"] != true {
		t.Fatalf("expected dnsLog true, got %v", result["dnsLog"])
	}
}

func TestExpandBasicsLog_Empty(t *testing.T) {
	result := expandBasicsLog(map[string]any{})
	if result != nil {
		t.Fatalf("expected nil for empty map, got %v", result)
	}
}

func TestExpandBasicsPolicy(t *testing.T) {
	item := map[string]any{
		"system": map[string]any{
			"stats_inbound_downlink":  true,
			"stats_inbound_uplink":    true,
			"stats_outbound_downlink": false,
			"stats_outbound_uplink":   false,
		},
		"level": []any{
			map[string]any{
				"id":                  0,
				"handshake":           4,
				"conn_idle":           300,
				"stats_user_uplink":   true,
				"stats_user_downlink": true,
			},
		},
	}
	result := expandBasicsPolicy(item)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	sys, ok := result["system"].(map[string]any)
	if !ok {
		t.Fatalf("expected system map, got %T", result["system"])
	}
	if sys["statsInboundDownlink"] != true {
		t.Fatalf("expected statsInboundDownlink true, got %v", sys["statsInboundDownlink"])
	}

	levels, ok := result["levels"].(map[string]any)
	if !ok {
		t.Fatalf("expected levels map, got %T", result["levels"])
	}
	level0, ok := levels["0"].(map[string]any)
	if !ok {
		t.Fatalf("expected level 0 map")
	}
	if level0["handshake"] != 4 {
		t.Fatalf("expected handshake 4, got %v", level0["handshake"])
	}
	if level0["connIdle"] != 300 {
		t.Fatalf("expected connIdle 300, got %v", level0["connIdle"])
	}
}

func TestExpandBasicsAPI(t *testing.T) {
	item := map[string]any{
		"tag":      "api",
		"services": []any{"HandlerService", "StatsService"},
	}
	result := expandBasicsAPI(item)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result["tag"] != "api" {
		t.Fatalf("expected tag api, got %v", result["tag"])
	}
	services, ok := result["services"].([]string)
	if !ok || len(services) != 2 {
		t.Fatalf("expected 2 services, got %v", result["services"])
	}
	if services[0] != "HandlerService" {
		t.Fatalf("expected HandlerService, got %v", services[0])
	}
}

func TestExpandBasicsAPI_Empty(t *testing.T) {
	result := expandBasicsAPI(map[string]any{})
	if result != nil {
		t.Fatalf("expected nil for empty map, got %v", result)
	}
}

func TestExpandBalancers(t *testing.T) {
	list := []any{
		map[string]any{
			"tag":      "bal1",
			"selector": []any{"proxy-*"},
			"strategy": []any{
				map[string]any{"type": "leastPing"},
			},
		},
	}
	result := expandBalancers(list)
	if len(result) != 1 {
		t.Fatalf("expected 1 balancer, got %d", len(result))
	}
	m := result[0].(map[string]any)
	if m["tag"] != "bal1" {
		t.Fatalf("expected bal1, got %v", m["tag"])
	}
	sel, ok := m["selector"].([]string)
	if !ok || len(sel) != 1 || sel[0] != "proxy-*" {
		t.Fatalf("unexpected selector: %v", m["selector"])
	}
	strategy, ok := m["strategy"].(map[string]any)
	if !ok || strategy["type"] != "leastPing" {
		t.Fatalf("unexpected strategy: %v", m["strategy"])
	}
}

func TestExpandOutbounds_Freedom(t *testing.T) {
	list := []any{
		map[string]any{
			"tag":      "direct",
			"protocol": "freedom",
			"freedom_settings": []any{
				map[string]any{
					"domain_strategy": "AsIs",
					"fragment": []any{
						map[string]any{
							"packets":  "tlshello",
							"length":   "100-200",
							"interval": "10-20",
						},
					},
					"noises": []any{
						map[string]any{
							"type":   "rand",
							"packet": "10-20",
							"delay":  "10-16",
						},
					},
				},
			},
		},
	}
	result := expandOutbounds(list)
	if len(result) != 1 {
		t.Fatalf("expected 1 outbound, got %d", len(result))
	}
	m := result[0].(map[string]any)
	if m["protocol"] != "freedom" {
		t.Fatalf("expected freedom, got %v", m["protocol"])
	}
	settings, ok := m["settings"].(map[string]any)
	if !ok {
		t.Fatalf("expected settings map, got %T", m["settings"])
	}
	if settings["domainStrategy"] != "AsIs" {
		t.Fatalf("expected AsIs, got %v", settings["domainStrategy"])
	}
	fragment, ok := settings["fragment"].(map[string]any)
	if !ok {
		t.Fatalf("expected fragment map, got %T", settings["fragment"])
	}
	if fragment["packets"] != "tlshello" {
		t.Fatalf("expected tlshello, got %v", fragment["packets"])
	}
	noises, ok := settings["noises"].([]any)
	if !ok || len(noises) != 1 {
		t.Fatalf("expected 1 noise, got %v", settings["noises"])
	}
}

func TestExpandOutbounds_Vmess(t *testing.T) {
	list := []any{
		map[string]any{
			"tag":      "vmess-out",
			"protocol": "vmess",
			"vmess_settings": []any{
				map[string]any{
					"address":  "example.com",
					"port":     443,
					"id":       "test-uuid",
					"security": "auto",
				},
			},
		},
	}
	result := expandOutbounds(list)
	if len(result) != 1 {
		t.Fatalf("expected 1 outbound")
	}
	settings := result[0].(map[string]any)["settings"].(map[string]any)
	vnext, ok := settings["vnext"].([]any)
	if !ok || len(vnext) != 1 {
		t.Fatalf("expected vnext with 1 server, got %v", settings["vnext"])
	}
	server := vnext[0].(map[string]any)
	if server["address"] != "example.com" {
		t.Fatalf("expected example.com, got %v", server["address"])
	}
	users := server["users"].([]any)
	if len(users) != 1 {
		t.Fatalf("expected 1 user")
	}
	user := users[0].(map[string]any)
	if user["id"] != "test-uuid" || user["security"] != "auto" {
		t.Fatalf("unexpected user: %v", user)
	}
}

func TestExpandOutbounds_Vless(t *testing.T) {
	list := []any{
		map[string]any{
			"tag":      "vless-out",
			"protocol": "vless",
			"vless_settings": []any{
				map[string]any{
					"address":    "example.com",
					"port":       443,
					"id":         "test-uuid",
					"flow":       "xtls-rprx-vision",
					"encryption": "none",
				},
			},
		},
	}
	result := expandOutbounds(list)
	settings := result[0].(map[string]any)["settings"].(map[string]any)
	vnext := settings["vnext"].([]any)
	server := vnext[0].(map[string]any)
	user := server["users"].([]any)[0].(map[string]any)
	if user["flow"] != "xtls-rprx-vision" {
		t.Fatalf("expected xtls-rprx-vision, got %v", user["flow"])
	}
	if user["encryption"] != "none" {
		t.Fatalf("expected none, got %v", user["encryption"])
	}
}

func TestExpandOutbounds_Trojan(t *testing.T) {
	list := []any{
		map[string]any{
			"tag":      "trojan-out",
			"protocol": "trojan",
			"trojan_settings": []any{
				map[string]any{
					"address":  "example.com",
					"port":     443,
					"password": "secret",
				},
			},
		},
	}
	result := expandOutbounds(list)
	settings := result[0].(map[string]any)["settings"].(map[string]any)
	servers := settings["servers"].([]any)
	if len(servers) != 1 {
		t.Fatalf("expected 1 server")
	}
	server := servers[0].(map[string]any)
	if server["password"] != "secret" {
		t.Fatalf("expected secret, got %v", server["password"])
	}
}

func TestExpandOutbounds_Socks(t *testing.T) {
	list := []any{
		map[string]any{
			"tag":      "socks-out",
			"protocol": "socks",
			"socks_settings": []any{
				map[string]any{
					"address": "127.0.0.1",
					"port":    1080,
					"user":    "admin",
					"pass":    "password",
				},
			},
		},
	}
	result := expandOutbounds(list)
	settings := result[0].(map[string]any)["settings"].(map[string]any)
	servers := settings["servers"].([]any)
	server := servers[0].(map[string]any)
	if server["address"] != "127.0.0.1" {
		t.Fatalf("expected 127.0.0.1, got %v", server["address"])
	}
	users := server["users"].([]any)
	user := users[0].(map[string]any)
	if user["user"] != "admin" || user["pass"] != "password" {
		t.Fatalf("unexpected user: %v", user)
	}
}

func TestExpandOutbounds_HTTP(t *testing.T) {
	list := []any{
		map[string]any{
			"tag":      "http-out",
			"protocol": "http",
			"http_settings": []any{
				map[string]any{
					"address": "proxy.example.com",
					"port":    8080,
					"user":    "user1",
					"pass":    "pass1",
				},
			},
		},
	}
	result := expandOutbounds(list)
	settings := result[0].(map[string]any)["settings"].(map[string]any)
	servers := settings["servers"].([]any)
	server := servers[0].(map[string]any)
	if server["address"] != "proxy.example.com" {
		t.Fatalf("expected proxy.example.com, got %v", server["address"])
	}
}

func TestExpandOutbounds_Hysteria(t *testing.T) {
	list := []any{
		map[string]any{
			"tag":      "hysteria-out",
			"protocol": "hysteria",
			"hysteria_settings": []any{
				map[string]any{
					"address": "example.com",
					"port":    443,
					"version": 2,
				},
			},
		},
	}
	result := expandOutbounds(list)
	settings := result[0].(map[string]any)["settings"].(map[string]any)
	servers := settings["servers"].([]any)
	server := servers[0].(map[string]any)
	if server["version"] != 2 {
		t.Fatalf("expected version 2, got %v", server["version"])
	}
}

func TestFlattenOutbounds_DNS(t *testing.T) {
	data := []any{
		map[string]any{
			"tag":      "dns-out",
			"protocol": "dns",
			"settings": map[string]any{
				"network": "udp",
				"address": "1.1.1.1",
				"port":    float64(53),
			},
		},
	}
	result := flattenXrayOutboundsToMap(data)
	outbounds := result["outbound"].([]any)
	o := outbounds[0].(map[string]any)
	dnsSettings := o["dns_settings"].([]any)[0].(map[string]any)
	if dnsSettings["network"] != "udp" {
		t.Fatalf("expected udp, got %v", dnsSettings["network"])
	}
	if dnsSettings["address"] != "1.1.1.1" {
		t.Fatalf("expected 1.1.1.1, got %v", dnsSettings["address"])
	}
	if dnsSettings["port"] != 53 {
		t.Fatalf("expected 53, got %v", dnsSettings["port"])
	}
}

func TestFlattenOutbounds_Vmess(t *testing.T) {
	data := []any{
		map[string]any{
			"tag":      "vmess-out",
			"protocol": "vmess",
			"settings": map[string]any{
				"vnext": []any{
					map[string]any{
						"address": "example.com",
						"port":    float64(443),
						"users": []any{
							map[string]any{"id": "uuid-1", "security": "auto"},
						},
					},
				},
			},
		},
	}
	result := flattenXrayOutboundsToMap(data)
	outbounds := result["outbound"].([]any)
	o := outbounds[0].(map[string]any)
	vmess := o["vmess_settings"].([]any)[0].(map[string]any)
	if vmess["address"] != "example.com" {
		t.Fatalf("expected example.com, got %v", vmess["address"])
	}
	if vmess["id"] != "uuid-1" {
		t.Fatalf("expected uuid-1, got %v", vmess["id"])
	}
	if vmess["security"] != "auto" {
		t.Fatalf("expected auto, got %v", vmess["security"])
	}
}

func TestFlattenOutbounds_Vless(t *testing.T) {
	data := []any{
		map[string]any{
			"tag":      "vless-out",
			"protocol": "vless",
			"settings": map[string]any{
				"vnext": []any{
					map[string]any{
						"address": "example.com",
						"port":    float64(443),
						"users": []any{
							map[string]any{"id": "uuid-2", "flow": "xtls-rprx-vision", "encryption": "none"},
						},
					},
				},
			},
		},
	}
	result := flattenXrayOutboundsToMap(data)
	outbounds := result["outbound"].([]any)
	vless := outbounds[0].(map[string]any)["vless_settings"].([]any)[0].(map[string]any)
	if vless["flow"] != "xtls-rprx-vision" {
		t.Fatalf("expected xtls-rprx-vision, got %v", vless["flow"])
	}
}

func TestFlattenOutbounds_Trojan(t *testing.T) {
	data := []any{
		map[string]any{
			"tag":      "trojan-out",
			"protocol": "trojan",
			"settings": map[string]any{
				"servers": []any{
					map[string]any{"address": "example.com", "port": float64(443), "password": "secret"},
				},
			},
		},
	}
	result := flattenXrayOutboundsToMap(data)
	trojan := result["outbound"].([]any)[0].(map[string]any)["trojan_settings"].([]any)[0].(map[string]any)
	if trojan["password"] != "secret" {
		t.Fatalf("expected secret, got %v", trojan["password"])
	}
}

func TestFlattenOutbounds_Socks(t *testing.T) {
	data := []any{
		map[string]any{
			"tag":      "socks-out",
			"protocol": "socks",
			"settings": map[string]any{
				"servers": []any{
					map[string]any{
						"address": "127.0.0.1",
						"port":    float64(1080),
						"users": []any{
							map[string]any{"user": "admin", "pass": "pass"},
						},
					},
				},
			},
		},
	}
	result := flattenXrayOutboundsToMap(data)
	socks := result["outbound"].([]any)[0].(map[string]any)["socks_settings"].([]any)[0].(map[string]any)
	if socks["user"] != "admin" {
		t.Fatalf("expected admin, got %v", socks["user"])
	}
	if socks["pass"] != "pass" {
		t.Fatalf("expected pass, got %v", socks["pass"])
	}
}

func TestFlattenOutbounds_HTTP(t *testing.T) {
	data := []any{
		map[string]any{
			"tag":      "http-out",
			"protocol": "http",
			"settings": map[string]any{
				"servers": []any{
					map[string]any{
						"address": "proxy.example.com",
						"port":    float64(8080),
						"users": []any{
							map[string]any{"user": "user1", "pass": "pass1"},
						},
					},
				},
			},
		},
	}
	result := flattenXrayOutboundsToMap(data)
	http := result["outbound"].([]any)[0].(map[string]any)["http_settings"].([]any)[0].(map[string]any)
	if http["user"] != "user1" {
		t.Fatalf("expected user1, got %v", http["user"])
	}
}

func TestFlattenOutbounds_Hysteria(t *testing.T) {
	data := []any{
		map[string]any{
			"tag":      "hysteria-out",
			"protocol": "hysteria",
			"settings": map[string]any{
				"servers": []any{
					map[string]any{"address": "example.com", "port": float64(443), "version": float64(2)},
				},
			},
		},
	}
	result := flattenXrayOutboundsToMap(data)
	hysteria := result["outbound"].([]any)[0].(map[string]any)["hysteria_settings"].([]any)[0].(map[string]any)
	if hysteria["version"] != 2 {
		t.Fatalf("expected 2, got %v", hysteria["version"])
	}
}

func TestExpandOutbounds_Blackhole(t *testing.T) {
	list := []any{
		map[string]any{
			"tag":      "blocked",
			"protocol": "blackhole",
			"blackhole_settings": []any{
				map[string]any{"response_type": "http"},
			},
		},
	}
	result := expandOutbounds(list)
	if len(result) != 1 {
		t.Fatalf("expected 1 outbound, got %d", len(result))
	}
	settings := result[0].(map[string]any)["settings"].(map[string]any)
	resp := settings["response"].(map[string]any)
	if resp["type"] != "http" {
		t.Fatalf("expected http, got %v", resp["type"])
	}
}

func TestExpandOutbounds_DNS(t *testing.T) {
	list := []any{
		map[string]any{
			"tag":      "dns-out",
			"protocol": "dns",
			"dns_settings": []any{
				map[string]any{"network": "udp", "address": "1.1.1.1", "port": 53},
			},
		},
	}
	result := expandOutbounds(list)
	if len(result) != 1 {
		t.Fatalf("expected 1 outbound, got %d", len(result))
	}
	settings := result[0].(map[string]any)["settings"].(map[string]any)
	if settings["network"] != "udp" {
		t.Fatalf("expected udp, got %v", settings["network"])
	}
	if settings["address"] != "1.1.1.1" {
		t.Fatalf("expected 1.1.1.1, got %v", settings["address"])
	}
	if settings["port"] != 53 {
		t.Fatalf("expected 53, got %v", settings["port"])
	}
}

func TestExpandOutbounds_Shadowsocks(t *testing.T) {
	list := []any{
		map[string]any{
			"tag":      "ss-out",
			"protocol": "shadowsocks",
			"shadowsocks_settings": []any{
				map[string]any{
					"address": "ss.example.com", "port": 8388,
					"password": "secret", "method": "aes-256-gcm", "uot": true,
				},
			},
		},
	}
	result := expandOutbounds(list)
	server := result[0].(map[string]any)["settings"].(map[string]any)["servers"].([]any)[0].(map[string]any)
	if server["address"] != "ss.example.com" {
		t.Fatalf("expected ss.example.com, got %v", server["address"])
	}
	if server["method"] != "aes-256-gcm" {
		t.Fatalf("expected aes-256-gcm, got %v", server["method"])
	}
	if server["uot"] != true {
		t.Fatalf("expected uot true, got %v", server["uot"])
	}
}

func TestExpandOutbounds_Wireguard(t *testing.T) {
	list := []any{
		map[string]any{
			"tag": "wg-out", "protocol": "wireguard",
			"wireguard_settings": []any{
				map[string]any{
					"secret_key": "wg-secret", "address": []any{"10.0.0.2/32"},
					"mtu": 1420, "domain_strategy": "ForceIPv4",
					"peer": []any{
						map[string]any{
							"public_key": "wg-pub", "endpoint": "wg.example.com:51820",
							"allowed_ips": []any{"0.0.0.0/0"}, "keep_alive": 25,
						},
					},
				},
			},
		},
	}
	result := expandOutbounds(list)
	settings := result[0].(map[string]any)["settings"].(map[string]any)
	if settings["secretKey"] != "wg-secret" {
		t.Fatalf("expected wg-secret, got %v", settings["secretKey"])
	}
	if settings["mtu"] != 1420 {
		t.Fatalf("expected 1420, got %v", settings["mtu"])
	}
	peers := settings["peers"].([]any)
	if len(peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(peers))
	}
	if peers[0].(map[string]any)["publicKey"] != "wg-pub" {
		t.Fatalf("expected wg-pub, got %v", peers[0].(map[string]any)["publicKey"])
	}
}

func TestFlattenOutbounds_Shadowsocks(t *testing.T) {
	data := []any{
		map[string]any{
			"tag": "ss-out", "protocol": "shadowsocks",
			"settings": map[string]any{
				"servers": []any{
					map[string]any{
						"address": "ss.example.com", "port": float64(8388),
						"password": "secret", "method": "aes-256-gcm", "uot": true,
					},
				},
			},
		},
	}
	result := flattenXrayOutboundsToMap(data)
	ss := result["outbound"].([]any)[0].(map[string]any)["shadowsocks_settings"].([]any)[0].(map[string]any)
	if ss["address"] != "ss.example.com" {
		t.Fatalf("expected ss.example.com, got %v", ss["address"])
	}
	if ss["method"] != "aes-256-gcm" {
		t.Fatalf("expected aes-256-gcm, got %v", ss["method"])
	}
	if ss["uot"] != true {
		t.Fatalf("expected uot true, got %v", ss["uot"])
	}
}

func TestExpandOutboundMux(t *testing.T) {
	list := []any{
		map[string]any{
			"enabled": true, "concurrency": 8,
			"xudp_concurrency": 16, "xudp_proxy_udp443": "reject",
		},
	}
	result := expandOutboundMux(list)
	if result == nil {
		t.Fatal("expected non-nil mux")
	}
	if result["enabled"] != true {
		t.Fatalf("expected enabled true, got %v", result["enabled"])
	}
	if result["concurrency"] != 8 {
		t.Fatalf("expected 8, got %v", result["concurrency"])
	}
	if result["xudpConcurrency"] != 16 {
		t.Fatalf("expected 16, got %v", result["xudpConcurrency"])
	}
	if result["xudpProxyUDP443"] != "reject" {
		t.Fatalf("expected reject, got %v", result["xudpProxyUDP443"])
	}
}

func TestFlattenOutboundMux(t *testing.T) {
	in := map[string]any{
		"enabled": true, "concurrency": float64(8),
		"xudpConcurrency": float64(16), "xudpProxyUDP443": "reject",
	}
	result := flattenOutboundMux(in)
	if result == nil {
		t.Fatal("expected non-nil mux")
	}
	if result["enabled"] != true {
		t.Fatalf("expected true, got %v", result["enabled"])
	}
	if result["concurrency"] != 8 {
		t.Fatalf("expected 8, got %v", result["concurrency"])
	}
	if result["xudp_concurrency"] != 16 {
		t.Fatalf("expected 16, got %v", result["xudp_concurrency"])
	}
	if result["xudp_proxy_udp443"] != "reject" {
		t.Fatalf("expected reject, got %v", result["xudp_proxy_udp443"])
	}
}

func TestExpandDNSServers_WithAllFields(t *testing.T) {
	list := []any{
		map[string]any{
			"address": "dns.example.com", "port": 53,
			"domains":       []any{"example.com", "example.org"},
			"expect_ips":    []any{"1.2.3.0/24"},
			"skip_fallback": true, "query_strategy": "UseIPv4",
		},
	}
	result := expandDNSServers(list)
	m := result[0].(map[string]any)
	if m["address"] != "dns.example.com" {
		t.Fatalf("expected dns.example.com, got %v", m["address"])
	}
	if m["port"] != 53 {
		t.Fatalf("expected 53, got %v", m["port"])
	}
	if m["skipFallback"] != true {
		t.Fatalf("expected skipFallback true, got %v", m["skipFallback"])
	}
	if m["queryStrategy"] != "UseIPv4" {
		t.Fatalf("expected UseIPv4, got %v", m["queryStrategy"])
	}
}

func TestBuildXrayDNSJSON_Roundtrip(t *testing.T) {
	input := map[string]any{
		"server": []any{
			map[string]any{"address": "8.8.8.8"},
			map[string]any{"address": "localhost", "port": 53, "domains": []any{"example.com"}},
		},
		"query_strategy": "UseIP",
	}
	flattened := flattenXrayDNSToMap(buildXrayDNSJSON(input))
	if flattened["query_strategy"] != "UseIP" {
		t.Fatalf("expected UseIP, got %v", flattened["query_strategy"])
	}
	servers := flattened["server"].([]any)
	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}
	if servers[0].(map[string]any)["address"] != "8.8.8.8" {
		t.Fatalf("expected 8.8.8.8, got %v", servers[0].(map[string]any)["address"])
	}
}

func TestBuildXrayRoutingJSON_Roundtrip(t *testing.T) {
	input := map[string]any{
		"domain_strategy": "IPIfNonMatch",
		"rule": []any{
			map[string]any{"type": "field", "ip": []any{"geoip:private"}, "outbound_tag": "direct"},
		},
	}
	flattened := flattenXrayRoutingToMap(buildXrayRoutingJSON(input))
	if flattened["domain_strategy"] != "IPIfNonMatch" {
		t.Fatalf("expected IPIfNonMatch, got %v", flattened["domain_strategy"])
	}
	rules := flattened["rule"].([]any)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].(map[string]any)["outbound_tag"] != "direct" {
		t.Fatalf("expected direct, got %v", rules[0].(map[string]any)["outbound_tag"])
	}
}

func TestBuildXrayBasicsJSON_Roundtrip(t *testing.T) {
	input := map[string]any{
		"log":   map[string]any{"loglevel": "debug", "dns_log": true},
		"api":   map[string]any{"tag": "api", "services": []any{"HandlerService", "StatsService"}},
		"stats": map[string]any{},
	}
	flattened := flattenXrayBasicsToMap(buildXrayBasicsJSON(input))
	log := flattened["log"].(map[string]any)
	if log["loglevel"] != "debug" {
		t.Fatalf("expected debug, got %v", log["loglevel"])
	}
	if log["dns_log"] != true {
		t.Fatalf("expected dns_log true, got %v", log["dns_log"])
	}
	if _, ok := flattened["stats"]; !ok {
		t.Fatalf("expected stats block")
	}
}

func TestBuildXrayReverseJSON_Roundtrip(t *testing.T) {
	input := map[string]any{
		"bridge": []any{map[string]any{"tag": "b1", "domain": "bridge.example.com"}},
		"portal": []any{map[string]any{"tag": "p1", "domain": "portal.example.com"}},
	}
	flattened := flattenXrayReverseToMap(buildXrayReverseJSON(input))
	b := flattened["bridge"].([]any)[0].(map[string]any)
	if b["tag"] != "b1" || b["domain"] != "bridge.example.com" {
		t.Fatalf("unexpected bridge: %v", b)
	}
	p := flattened["portal"].([]any)[0].(map[string]any)
	if p["tag"] != "p1" || p["domain"] != "portal.example.com" {
		t.Fatalf("unexpected portal: %v", p)
	}
}

func TestBuildXrayBalancersJSON_Roundtrip(t *testing.T) {
	input := map[string]any{
		"balancer": []any{
			map[string]any{
				"tag": "bal1", "selector": []any{"proxy-*"},
				"strategy": []any{map[string]any{"type": "random"}},
			},
		},
	}
	flattened := flattenXrayBalancersToMap(buildXrayBalancersJSON(input))
	bal := flattened["balancer"].([]any)[0].(map[string]any)
	if bal["tag"] != "bal1" {
		t.Fatalf("expected bal1, got %v", bal["tag"])
	}
	strategy := bal["strategy"].([]any)[0].(map[string]any)
	if strategy["type"] != "random" {
		t.Fatalf("expected random, got %v", strategy["type"])
	}
}

func TestBuildXrayOutboundsJSON_Roundtrip(t *testing.T) {
	input := map[string]any{
		"outbound": []any{
			map[string]any{
				"tag": "direct", "protocol": "freedom",
				"freedom_settings": []any{map[string]any{"domain_strategy": "AsIs"}},
			},
			map[string]any{
				"tag": "blocked", "protocol": "blackhole",
				"blackhole_settings": []any{map[string]any{"response_type": "none"}},
			},
		},
	}
	flattened := flattenXrayOutboundsToMap(buildXrayOutboundsJSON(input))
	outbounds := flattened["outbound"].([]any)
	if len(outbounds) != 2 {
		t.Fatalf("expected 2 outbounds, got %d", len(outbounds))
	}
	freedom := outbounds[0].(map[string]any)["freedom_settings"].([]any)[0].(map[string]any)
	if freedom["domain_strategy"] != "AsIs" {
		t.Fatalf("expected AsIs, got %v", freedom["domain_strategy"])
	}
	bh := outbounds[1].(map[string]any)["blackhole_settings"].([]any)[0].(map[string]any)
	if bh["response_type"] != "none" {
		t.Fatalf("expected none, got %v", bh["response_type"])
	}
}

// --- Tests for review fixes ---

func TestFlattenBasicsPolicyLevels_Sorted(t *testing.T) {
	in := map[string]any{
		"2": map[string]any{"handshake": float64(8)},
		"0": map[string]any{"handshake": float64(4)},
		"1": map[string]any{"handshake": float64(6)},
	}
	// Run 20 times to catch non-determinism.
	for i := 0; i < 20; i++ {
		result := flattenBasicsPolicyLevels(in)
		if len(result) != 3 {
			t.Fatalf("expected 3 levels, got %d", len(result))
		}
		ids := make([]int, len(result))
		for j, item := range result {
			ids[j] = item.(map[string]any)["id"].(int)
		}
		if ids[0] != 0 || ids[1] != 1 || ids[2] != 2 {
			t.Fatalf("iteration %d: expected sorted [0,1,2], got %v", i, ids)
		}
	}
}

func TestExpandInt64List_WithNullAndUnknown(t *testing.T) {
	elems := []attr.Value{
		types.Int64Value(1),
		types.Int64Null(),
		types.Int64Unknown(),
		types.Int64Value(3),
	}
	l, diags := types.ListValue(types.Int64Type, elems)
	if diags.HasError() {
		t.Fatalf("failed to create list: %v", diags)
	}
	result := expandInt64List(l)
	expected := []any{1, 3}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestExpandInt64List_NullList(t *testing.T) {
	l := types.ListNull(types.Int64Type)
	result := expandInt64List(l)
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

func TestExpandInt64List_UnknownList(t *testing.T) {
	l := types.ListUnknown(types.Int64Type)
	result := expandInt64List(l)
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

func TestExpandXrayOutbounds_EmptyMuxBlock(t *testing.T) {
	m := &XrayOutboundsModel{
		Outbound: []XrayOutboundEntry{
			{
				Tag:      types.StringValue("test"),
				Protocol: types.StringValue("freedom"),
				Mux: []XrayOutboundMux{
					{
						Enabled:         types.BoolNull(),
						Concurrency:     types.Int64Null(),
						XudpConcurrency: types.Int64Null(),
						XudpProxyUDP443: types.StringNull(),
					},
				},
			},
		},
	}
	result := expandXrayOutbounds(m)
	outbounds := result["outbound"].([]any)
	if len(outbounds) != 1 {
		t.Fatalf("expected 1 outbound, got %d", len(outbounds))
	}
	entry := outbounds[0].(map[string]any)
	if _, ok := entry["mux"]; ok {
		t.Fatalf("mux with all-null fields should not be in result")
	}
}

func TestExpandXrayOutbounds_EmptySettingsBlock(t *testing.T) {
	m := &XrayOutboundsModel{
		Outbound: []XrayOutboundEntry{
			{
				Tag:      types.StringValue("test"),
				Protocol: types.StringValue("blackhole"),
				BlackholeSettings: []XrayBlackholeSettings{
					{ResponseType: types.StringNull()},
				},
			},
		},
	}
	result := expandXrayOutbounds(m)
	outbounds := result["outbound"].([]any)
	entry := outbounds[0].(map[string]any)
	if _, ok := entry["blackhole_settings"]; ok {
		t.Fatalf("blackhole_settings with all-null fields should not be in result")
	}
}

func TestExpandOutbounds_FreedomIPsBlocked(t *testing.T) {
	list := []any{
		map[string]any{
			"tag":      "direct",
			"protocol": "freedom",
			"freedom_settings": []any{
				map[string]any{
					"domain_strategy": "AsIs",
					"ips_blocked":     []any{"geoip:cn", "10.0.0.0/8"},
				},
			},
		},
	}
	result := expandOutbounds(list)
	if len(result) != 1 {
		t.Fatalf("expected 1 outbound, got %d", len(result))
	}
	m := result[0].(map[string]any)
	settings, ok := m["settings"].(map[string]any)
	if !ok {
		t.Fatalf("expected settings map, got %T", m["settings"])
	}
	ipsBlocked, ok := settings["ipsBlocked"].([]any)
	if !ok || len(ipsBlocked) != 2 {
		t.Fatalf("expected 2 ipsBlocked entries, got %v", settings["ipsBlocked"])
	}
	if ipsBlocked[0] != "geoip:cn" || ipsBlocked[1] != "10.0.0.0/8" {
		t.Fatalf("unexpected ipsBlocked values: %v", ipsBlocked)
	}
}

func TestFlattenOutbounds_FreedomIPsBlocked(t *testing.T) {
	data := []any{
		map[string]any{
			"tag":      "direct",
			"protocol": "freedom",
			"settings": map[string]any{
				"domainStrategy": "AsIs",
				"ipsBlocked":     []any{"geoip:cn", "10.0.0.0/8"},
			},
		},
	}
	result := flattenXrayOutboundsToMap(data)
	outbounds := result["outbound"].([]any)
	if len(outbounds) != 1 {
		t.Fatalf("expected 1 outbound, got %d", len(outbounds))
	}
	freedom := outbounds[0].(map[string]any)["freedom_settings"].([]any)[0].(map[string]any)
	ipsBlocked, ok := freedom["ips_blocked"].([]any)
	if !ok || len(ipsBlocked) != 2 {
		t.Fatalf("expected 2 ips_blocked entries, got %v", freedom["ips_blocked"])
	}
	if ipsBlocked[0] != "geoip:cn" || ipsBlocked[1] != "10.0.0.0/8" {
		t.Fatalf("unexpected ips_blocked values: %v", ipsBlocked)
	}
}

func TestFreedomIPsBlocked_Roundtrip(t *testing.T) {
	input := map[string]any{
		"outbound": []any{
			map[string]any{
				"tag": "direct", "protocol": "freedom",
				"freedom_settings": []any{map[string]any{
					"domain_strategy": "AsIs",
					"ips_blocked":     []any{"geoip:cn", "192.168.0.0/16"},
				}},
			},
		},
	}
	flattened := flattenXrayOutboundsToMap(buildXrayOutboundsJSON(input))
	outbounds := flattened["outbound"].([]any)
	if len(outbounds) != 1 {
		t.Fatalf("expected 1 outbound, got %d", len(outbounds))
	}
	freedom := outbounds[0].(map[string]any)["freedom_settings"].([]any)[0].(map[string]any)
	ipsBlocked, ok := freedom["ips_blocked"].([]any)
	if !ok || len(ipsBlocked) != 2 {
		t.Fatalf("expected 2 ips_blocked entries after roundtrip, got %v", freedom["ips_blocked"])
	}
	if ipsBlocked[0] != "geoip:cn" || ipsBlocked[1] != "192.168.0.0/16" {
		t.Fatalf("unexpected ips_blocked values: %v", ipsBlocked)
	}
}

func TestFreedomIPsBlocked_TypedModelRoundtrip(t *testing.T) {
	ipsBlockedList := types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("geoip:cn"),
		types.StringValue("10.0.0.0/8"),
	})

	model := &XrayOutboundsModel{
		Outbound: []XrayOutboundEntry{
			{
				Tag:      types.StringValue("direct"),
				Protocol: types.StringValue("freedom"),
				FreedomSettings: []XrayFreedomSettings{
					{
						DomainStrategy: types.StringValue("AsIs"),
						Redirect:       types.StringNull(),
						IPsBlocked:     ipsBlockedList,
					},
				},
			},
		},
	}

	// Typed model -> untyped map
	expanded := expandXrayOutbounds(model)
	outbounds := expanded["outbound"].([]any)
	entry := outbounds[0].(map[string]any)
	fsList := entry["freedom_settings"].([]any)
	fs := fsList[0].(map[string]any)
	ips, ok := fs["ips_blocked"].([]any)
	if !ok || len(ips) != 2 {
		t.Fatalf("expected 2 ips_blocked in expanded map, got %v", fs["ips_blocked"])
	}
	if ips[0] != "geoip:cn" || ips[1] != "10.0.0.0/8" {
		t.Fatalf("unexpected expanded ips_blocked: %v", ips)
	}

	// Untyped map -> typed model (flatten back)
	flatModel := flattenXrayOutbounds(expanded)
	if len(flatModel.Outbound) != 1 {
		t.Fatalf("expected 1 outbound in flattened model, got %d", len(flatModel.Outbound))
	}
	flatFS := flatModel.Outbound[0].FreedomSettings
	if len(flatFS) != 1 {
		t.Fatalf("expected 1 freedom_settings, got %d", len(flatFS))
	}
	if flatFS[0].IPsBlocked.IsNull() || flatFS[0].IPsBlocked.IsUnknown() {
		t.Fatalf("expected non-null ips_blocked in flattened model")
	}
	elems := flatFS[0].IPsBlocked.Elements()
	if len(elems) != 2 {
		t.Fatalf("expected 2 elements in ips_blocked, got %d", len(elems))
	}
	if elems[0].(types.String).ValueString() != "geoip:cn" {
		t.Fatalf("expected geoip:cn, got %v", elems[0])
	}
	if elems[1].(types.String).ValueString() != "10.0.0.0/8" {
		t.Fatalf("expected 10.0.0.0/8, got %v", elems[1])
	}
}

func TestFreedomIPsBlocked_NullHandling(t *testing.T) {
	model := &XrayOutboundsModel{
		Outbound: []XrayOutboundEntry{
			{
				Tag:      types.StringValue("direct"),
				Protocol: types.StringValue("freedom"),
				FreedomSettings: []XrayFreedomSettings{
					{
						DomainStrategy: types.StringValue("AsIs"),
						Redirect:       types.StringNull(),
						IPsBlocked:     types.ListNull(types.StringType),
					},
				},
			},
		},
	}

	expanded := expandXrayOutbounds(model)
	outbounds := expanded["outbound"].([]any)
	entry := outbounds[0].(map[string]any)
	fsList := entry["freedom_settings"].([]any)
	fs := fsList[0].(map[string]any)
	if _, ok := fs["ips_blocked"]; ok {
		t.Fatalf("null ips_blocked should not appear in expanded map")
	}

	// Flatten with missing ips_blocked -> should get null list
	untypedMap := map[string]any{
		"outbound": []any{
			map[string]any{
				"protocol": "freedom",
				"freedom_settings": []any{
					map[string]any{
						"domain_strategy": "AsIs",
					},
				},
			},
		},
	}
	flatModel := flattenXrayOutbounds(untypedMap)
	flatFS := flatModel.Outbound[0].FreedomSettings[0]
	if !flatFS.IPsBlocked.IsNull() {
		t.Fatalf("expected null ips_blocked when key is missing, got %v", flatFS.IPsBlocked)
	}
}
