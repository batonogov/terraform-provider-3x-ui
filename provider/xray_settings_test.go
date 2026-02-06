package provider

import (
	"reflect"
	"testing"
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
	logList, ok := result["log"].([]any)
	if !ok || len(logList) != 1 {
		t.Fatalf("expected log block, got %v", result["log"])
	}
	log := logList[0].(map[string]any)
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
