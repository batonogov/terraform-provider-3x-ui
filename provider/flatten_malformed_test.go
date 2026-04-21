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

func TestInboundToModel_MalformedSettings_FailHard(t *testing.T) {
	inbound := &Inbound{
		ID:       1,
		Settings: `{broken`,
		Protocol: "vless",
	}
	_, diags := inboundToModel(inbound, true)
	if !diags.HasError() {
		t.Fatal("expected diagnostics error for malformed settings")
	}
}

func TestInboundToModel_MalformedStreamSettings_FailHard(t *testing.T) {
	inbound := &Inbound{
		ID:             1,
		Settings:       `{"decryption":"none"}`,
		StreamSettings: `{broken`,
		Protocol:       "vless",
	}
	_, diags := inboundToModel(inbound, true)
	if !diags.HasError() {
		t.Fatal("expected diagnostics error for malformed stream_settings")
	}
}

func TestInboundToModel_MalformedSniffing_FailHard(t *testing.T) {
	inbound := &Inbound{
		ID:             1,
		Settings:       `{"decryption":"none"}`,
		StreamSettings: `{"network":"tcp"}`,
		Sniffing:       `{broken`,
		Protocol:       "vless",
	}
	_, diags := inboundToModel(inbound, true)
	if !diags.HasError() {
		t.Fatal("expected diagnostics error for malformed sniffing")
	}
}

func TestInboundToModel_MalformedSettings_Soft(t *testing.T) {
	inbound := &Inbound{
		ID:       1,
		Settings: `{broken`,
		Protocol: "vless",
	}
	m, diags := inboundToModel(inbound, false)
	if diags.HasError() {
		t.Fatal("expected warnings, not errors, for soft mode")
	}
	if len(diags) == 0 {
		t.Fatal("expected at least one warning diagnostic")
	}
	if m != nil {
		if m.ID.ValueString() != "1" {
			t.Fatalf("expected ID=1, got %s", m.ID.ValueString())
		}
	} else {
		t.Fatal("expected model to be returned in soft mode")
	}
}
