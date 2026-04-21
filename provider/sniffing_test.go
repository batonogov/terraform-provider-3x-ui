package provider

import (
	"encoding/json"
	"testing"
)

func TestBuildSniffingJSON_WithExclusions(t *testing.T) {
	input := map[string]any{
		"enabled":          true,
		"dest_override":    []any{"http", "tls"},
		"metadata_only":    false,
		"route_only":       false,
		"ips_excluded":     []any{"geoip:private"},
		"domains_excluded": []any{"domain:example.com"},
	}
	result := buildSniffingJSON(input)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	ips, ok := parsed["ipsExcluded"].([]any)
	if !ok || len(ips) != 1 || ips[0] != "geoip:private" {
		t.Fatalf("unexpected ipsExcluded: %v", parsed["ipsExcluded"])
	}
	domains, ok := parsed["domainsExcluded"].([]any)
	if !ok || len(domains) != 1 || domains[0] != "domain:example.com" {
		t.Fatalf("unexpected domainsExcluded: %v", parsed["domainsExcluded"])
	}
}

func TestFlattenSniffing_WithExclusions(t *testing.T) {
	input := `{"enabled":true,"destOverride":["http","tls"],"metadataOnly":false,"routeOnly":false,"ipsExcluded":["geoip:private"],"domainsExcluded":["domain:example.com"]}`
	result, err := flattenSniffing(input)
	if err != nil {
		t.Fatalf("flattenSniffing error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
	item := result[0].(map[string]any)
	ips, ok := item["ips_excluded"].([]any)
	if !ok || len(ips) != 1 || ips[0] != "geoip:private" {
		t.Fatalf("unexpected ips_excluded: %v", item["ips_excluded"])
	}
	domains, ok := item["domains_excluded"].([]any)
	if !ok || len(domains) != 1 || domains[0] != "domain:example.com" {
		t.Fatalf("unexpected domains_excluded: %v", item["domains_excluded"])
	}
}

func TestBuildSniffingJSON_NoExclusions(t *testing.T) {
	input := map[string]any{
		"enabled":       true,
		"dest_override": []any{"http", "tls"},
		"metadata_only": false,
		"route_only":    false,
	}
	result := buildSniffingJSON(input)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := parsed["ipsExcluded"]; ok {
		t.Fatalf("ipsExcluded should be absent when not set")
	}
	if _, ok := parsed["domainsExcluded"]; ok {
		t.Fatalf("domainsExcluded should be absent when not set")
	}
}
