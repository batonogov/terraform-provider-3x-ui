package provider

import (
	"strings"
	"testing"
)

func TestFlattenSettings_MalformedJSON(t *testing.T) {
	_, err := flattenSettings(`{invalid`)
	if err == nil {
		t.Fatal("expected error for malformed settings JSON")
	}
	if !strings.Contains(err.Error(), "failed to parse settings JSON") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestFlattenStreamSettings_MalformedJSON(t *testing.T) {
	_, err := flattenStreamSettings(`not json`)
	if err == nil {
		t.Fatal("expected error for malformed stream_settings JSON")
	}
	if !strings.Contains(err.Error(), "failed to parse stream_settings JSON") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestFlattenSniffing_MalformedJSON(t *testing.T) {
	_, err := flattenSniffing(`[broken`)
	if err == nil {
		t.Fatal("expected error for malformed sniffing JSON")
	}
	if !strings.Contains(err.Error(), "failed to parse sniffing JSON") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestFlattenSettingsToMap_MalformedJSON(t *testing.T) {
	_, err := flattenSettingsToMap(`{bad`)
	if err == nil {
		t.Fatal("expected error for malformed settings JSON")
	}
}

func TestFlattenStreamSettingsToMap_MalformedJSON(t *testing.T) {
	_, err := flattenStreamSettingsToMap(`{bad`)
	if err == nil {
		t.Fatal("expected error for malformed stream_settings JSON")
	}
}

func TestFlattenSniffingToMap_MalformedJSON(t *testing.T) {
	_, err := flattenSniffingToMap(`{bad`)
	if err == nil {
		t.Fatal("expected error for malformed sniffing JSON")
	}
}

func TestFlattenSettings_EmptyString(t *testing.T) {
	result, err := flattenSettings("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %v", result)
	}
}

func TestFlattenStreamSettings_EmptyString(t *testing.T) {
	result, err := flattenStreamSettings("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %v", result)
	}
}

func TestFlattenSniffing_EmptyString(t *testing.T) {
	result, err := flattenSniffing("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %v", result)
	}
}

func TestInboundToModel_MalformedSettings(t *testing.T) {
	inbound := &Inbound{
		ID:       1,
		Settings: `{broken`,
		Protocol: "vless",
	}
	_, diags := inboundToModel(inbound)
	if !diags.HasError() {
		t.Fatal("expected diagnostics error for malformed settings")
	}
}

func TestInboundToModel_MalformedStreamSettings(t *testing.T) {
	inbound := &Inbound{
		ID:             1,
		Settings:       `{"decryption":"none"}`,
		StreamSettings: `{broken`,
		Protocol:       "vless",
	}
	_, diags := inboundToModel(inbound)
	if !diags.HasError() {
		t.Fatal("expected diagnostics error for malformed stream_settings")
	}
}

func TestInboundToModel_MalformedSniffing(t *testing.T) {
	inbound := &Inbound{
		ID:             1,
		Settings:       `{"decryption":"none"}`,
		StreamSettings: `{"network":"tcp"}`,
		Sniffing:       `{broken`,
		Protocol:       "vless",
	}
	_, diags := inboundToModel(inbound)
	if !diags.HasError() {
		t.Fatal("expected diagnostics error for malformed sniffing")
	}
}
