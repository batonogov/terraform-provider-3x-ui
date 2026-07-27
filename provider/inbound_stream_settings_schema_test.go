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
	if out["max_timediff"] != int64(60000) {
		t.Fatalf("max_timediff: %T %v", out["max_timediff"], out["max_timediff"])
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

// "0.0.0" is the documented way to remove the REALITY lower bound: Xray only
// skips the gate for a literal zero version, and replaces an empty minClientVer
// with its own default. The value therefore has to survive both conversion
// layers unchanged — typed model -> snake_case map -> panel JSON and back.
func TestRealitySettings_ZeroVersionRoundTrip(t *testing.T) {
	in := &InboundRealitySettingsModel{
		Target:       types.StringValue("example.com:443"),
		MinClientVer: types.StringValue("0.0.0"),
		MaxClientVer: types.StringValue("255.255.255"),
		MaxTimediff:  types.Int64Value(0),
	}

	wire := expandRealitySettings([]any{expandRealitySettingsFromModel(in)})
	if wire["minClientVer"] != "0.0.0" {
		t.Fatalf("minClientVer lost on the wire: %v", wire["minClientVer"])
	}

	if got, ok := wire["maxTimediff"].(int64); !ok || got != 0 {
		t.Fatalf("maxTimediff should stay a concrete int64 zero, got %T %v",
			wire["maxTimediff"], wire["maxTimediff"])
	}

	out := flattenRealitySettingsToModel(flattenRealitySettings(wire))
	if got := out.MinClientVer.ValueString(); got != "0.0.0" {
		t.Fatalf("MinClientVer: want 0.0.0, got %q", got)
	}
	if got := out.MaxClientVer.ValueString(); got != "255.255.255" {
		t.Fatalf("MaxClientVer: want 255.255.255, got %q", got)
	}
	// max_timediff = 0 disables the time check; it must stay a concrete zero
	// rather than collapsing to null, which would break the Optional+Computed
	// contract on the next plan.
	if out.MaxTimediff.IsNull() || out.MaxTimediff.ValueInt64() != 0 {
		t.Fatalf("MaxTimediff: want 0, got %v", out.MaxTimediff)
	}
}

// xray-core spells the field maxTimeDiff, the panel spells it maxTimediff, and
// 3x-ui keeps streamSettings as opaque JSON text — so an inbound authored
// outside the panel can carry either. Reading only one spelling would import
// the other as null and then drop it on the next update, quietly turning the
// time-difference gate off.
func TestFlattenRealitySettings_MaxTimediffAlias(t *testing.T) {
	canonical := flattenRealitySettings(map[string]any{"maxTimeDiff": float64(45000)})
	if got := canonical["max_timediff"]; got != int64(45000) {
		t.Fatalf("canonical maxTimeDiff not read: %v", got)
	}

	panelSpelling := flattenRealitySettings(map[string]any{"maxTimediff": float64(45000)})
	if got := panelSpelling["max_timediff"]; got != int64(45000) {
		t.Fatalf("panel maxTimediff not read: %v", got)
	}

	// Both present: precedence is fixed so a plan never depends on map order.
	both := flattenRealitySettings(map[string]any{
		"maxTimeDiff": float64(1000),
		"maxTimediff": float64(2000),
	})
	if got := both["max_timediff"]; got != int64(1000) {
		t.Fatalf("canonical spelling should win, got %v", got)
	}

	if _, ok := flattenRealitySettings(map[string]any{})["max_timediff"]; ok {
		t.Fatalf("absent field must not materialise a zero")
	}
}

// An inbound imported with the canonical spelling has to reach the typed model
// with its value intact — this is the path that decides whether the next
// unrelated update preserves or drops the gate.
func TestFlattenRealitySettingsToModel_MaxTimediffAliasImport(t *testing.T) {
	m := flattenRealitySettingsToModel(
		flattenRealitySettings(map[string]any{
			"target":      "example.com:443",
			"maxTimeDiff": float64(45000),
		}),
	)
	if m.MaxTimediff.IsNull() || m.MaxTimediff.ValueInt64() != 45000 {
		t.Fatalf("MaxTimediff after import: want 45000, got %v", m.MaxTimediff)
	}

	// ...and survives being written back out under the panel spelling.
	wire := expandRealitySettings([]any{expandRealitySettingsFromModel(m)})
	if got := wire["maxTimediff"]; got != int64(45000) {
		t.Fatalf("maxTimediff dropped on re-serialise: %v", got)
	}
}

// `int` is 32 bits on the 386 and arm release targets, so the conversion path
// must not narrow: a value the schema accepts would otherwise wrap to a
// negative maxTimediff on those builds.
func TestRealityMaxTimediff_NoPlatformNarrowing(t *testing.T) {
	const beyond32Bit = int64(4_294_967_296) // 2^32

	wire := expandRealitySettings([]any{expandRealitySettingsFromModel(
		&InboundRealitySettingsModel{MaxTimediff: types.Int64Value(beyond32Bit)},
	)})
	if got := wire["maxTimediff"]; got != beyond32Bit {
		t.Fatalf("maxTimediff narrowed on write: %v", got)
	}

	m := flattenRealitySettingsToModel(flattenRealitySettings(wire))
	if got := m.MaxTimediff.ValueInt64(); got != beyond32Bit {
		t.Fatalf("maxTimediff narrowed on read: want %d, got %d", beyond32Bit, got)
	}
}
