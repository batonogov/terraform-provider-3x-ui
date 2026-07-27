package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The REALITY client-version gate fields (minClientVer/maxClientVer/maxTimediff)
// must survive the typed-model <-> untyped-map conversion. The map <-> JSON layer
// is covered by stream_settings_test.go.

func TestExpandRealitySettingsFromModel_ClientVerFields(t *testing.T) {
	m := &InboundRealitySettingsModel{
		Target:       types.StringValue("example.com:443"),
		MinClientVer: types.StringValue("26.3.27"),
		MaxClientVer: types.StringValue("26.9.9"),
		MaxTimediff:  types.Int64Value(60000),
	}
	out := expandRealitySettingsFromModel(m)
	if out["min_client_ver"] != "26.3.27" {
		t.Fatalf("min_client_ver: %v", out["min_client_ver"])
	}
	if out["max_client_ver"] != "26.9.9" {
		t.Fatalf("max_client_ver: %v", out["max_client_ver"])
	}
	if out["max_timediff"] != 60000 {
		t.Fatalf("max_timediff: %v", out["max_timediff"])
	}
}

func TestFlattenRealitySettingsToModel_ClientVerFields(t *testing.T) {
	data := map[string]any{
		"min_client_ver": "26.3.27",
		"max_client_ver": "26.9.9",
		"max_timediff":   60000,
	}
	m := flattenRealitySettingsToModel(data)
	if m.MinClientVer.ValueString() != "26.3.27" {
		t.Fatalf("MinClientVer: %v", m.MinClientVer)
	}
	if m.MaxClientVer.ValueString() != "26.9.9" {
		t.Fatalf("MaxClientVer: %v", m.MaxClientVer)
	}
	if m.MaxTimediff.ValueInt64() != 60000 {
		t.Fatalf("MaxTimediff: %v", m.MaxTimediff)
	}
}

// Absent gate fields must flatten to null so the Optional+Computed attributes do
// not force a spurious diff against a panel that omits them.
func TestFlattenRealitySettingsToModel_ClientVerAbsent(t *testing.T) {
	m := flattenRealitySettingsToModel(map[string]any{})
	if !m.MinClientVer.IsNull() || !m.MaxClientVer.IsNull() || !m.MaxTimediff.IsNull() {
		t.Fatalf("expected null client-ver fields when absent, got %v/%v/%v",
			m.MinClientVer, m.MaxClientVer, m.MaxTimediff)
	}
}
