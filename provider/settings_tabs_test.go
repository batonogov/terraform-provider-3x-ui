package provider

import (
	"reflect"
	"testing"
)

func TestPanelSettingsNeedRestart_ChangedKey(t *testing.T) {
	existing := map[string]any{"webPort": float64(2053)}
	desired := map[string]any{"webPort": float64(8080)}
	if !panelSettingsNeedRestart(existing, desired) {
		t.Fatalf("expected restart needed")
	}
}

func TestPanelSettingsNeedRestart_UnchangedKey(t *testing.T) {
	existing := map[string]any{"webPort": float64(2053)}
	desired := map[string]any{"webPort": float64(2053)}
	if panelSettingsNeedRestart(existing, desired) {
		t.Fatalf("expected no restart")
	}
}

func TestPanelSettingsNeedRestart_NonRestartKey(t *testing.T) {
	existing := map[string]any{"pageSize": float64(25)}
	desired := map[string]any{"pageSize": float64(50)}
	if panelSettingsNeedRestart(existing, desired) {
		t.Fatalf("expected no restart for pageSize")
	}
}

func TestPanelSettingsNeedRestart_NewRestartKey(t *testing.T) {
	existing := map[string]any{}
	desired := map[string]any{"webBasePath": "/new/"}
	if !panelSettingsNeedRestart(existing, desired) {
		t.Fatalf("expected restart for new restart key")
	}
}

func TestPanelSettingsNeedRestart_NoRestartKeys(t *testing.T) {
	existing := map[string]any{"webPort": float64(2053)}
	desired := map[string]any{"remarkModel": "-ieo"}
	if panelSettingsNeedRestart(existing, desired) {
		t.Fatalf("expected no restart")
	}
}

func TestSettingsValueEqual_Strings(t *testing.T) {
	if !settingsValueEqual("abc", "abc") {
		t.Fatalf("expected equal")
	}
	if settingsValueEqual("abc", "def") {
		t.Fatalf("expected not equal")
	}
}

func TestSettingsValueEqual_Bools(t *testing.T) {
	if !settingsValueEqual(true, true) {
		t.Fatalf("expected equal")
	}
	if settingsValueEqual(true, false) {
		t.Fatalf("expected not equal")
	}
}

func TestSettingsValueEqual_FloatAndInt(t *testing.T) {
	if !settingsValueEqual(float64(42), int(42)) {
		t.Fatalf("expected equal")
	}
	if !settingsValueEqual(float64(42), int64(42)) {
		t.Fatalf("expected equal for int64")
	}
}

func TestSettingsValueEqual_IntAndFloat(t *testing.T) {
	if !settingsValueEqual(int(42), float64(42)) {
		t.Fatalf("expected equal")
	}
}

func TestSettingsValueEqual_Nil(t *testing.T) {
	if !settingsValueEqual(nil, nil) {
		t.Fatalf("expected nil == nil")
	}
	if settingsValueEqual(nil, "x") {
		t.Fatalf("expected nil != string")
	}
}

func TestSettingsValueEqual_DifferentTypes(t *testing.T) {
	if settingsValueEqual("42", float64(42)) {
		t.Fatalf("expected not equal for string vs number")
	}
}

func TestNumberValueEqual_Float64(t *testing.T) {
	if !numberValueEqual(3.14, float64(3.14)) {
		t.Fatalf("expected equal")
	}
}

func TestNumberValueEqual_Int(t *testing.T) {
	if !numberValueEqual(42.0, int(42)) {
		t.Fatalf("expected equal")
	}
}

func TestNumberValueEqual_Int64(t *testing.T) {
	if !numberValueEqual(100.0, int64(100)) {
		t.Fatalf("expected equal")
	}
}

func TestNumberValueEqual_Uint(t *testing.T) {
	if !numberValueEqual(10.0, uint(10)) {
		t.Fatalf("expected equal")
	}
}

func TestNumberValueEqual_NonNumeric(t *testing.T) {
	if numberValueEqual(42.0, "42") {
		t.Fatalf("expected not equal for string")
	}
}

func TestMergeSettings_BothNil(t *testing.T) {
	result := mergeSettings(nil, nil)
	if len(result) != 0 {
		t.Fatalf("expected empty map, got %v", result)
	}
}

func TestMergeSettings_NilExisting(t *testing.T) {
	desired := map[string]any{"a": "1"}
	result := mergeSettings(nil, desired)
	if !reflect.DeepEqual(result, desired) {
		t.Fatalf("expected desired, got %v", result)
	}
}

func TestMergeSettings_Overlap(t *testing.T) {
	existing := map[string]any{"a": "old", "b": "keep"}
	desired := map[string]any{"a": "new"}
	result := mergeSettings(existing, desired)
	if result["a"] != "new" {
		t.Fatalf("expected override, got %v", result["a"])
	}
	if result["b"] != "keep" {
		t.Fatalf("expected keep, got %v", result["b"])
	}
}

func TestMergeSettings_Union(t *testing.T) {
	existing := map[string]any{"a": "1"}
	desired := map[string]any{"b": "2"}
	result := mergeSettings(existing, desired)
	if result["a"] != "1" || result["b"] != "2" {
		t.Fatalf("expected union, got %v", result)
	}
}

func TestFlattenPanelSettingsFields_Full(t *testing.T) {
	in := map[string]any{
		"webListen":   "0.0.0.0",
		"webPort":     float64(2053),
		"webBasePath": "/panel/",
		"pageSize":    float64(25),
	}
	out := flattenPanelSettingsFields(in)
	if out["web_listen"] != "0.0.0.0" {
		t.Fatalf("unexpected web_listen: %v", out["web_listen"])
	}
	if out["web_port"] != 2053 {
		t.Fatalf("unexpected web_port: %v", out["web_port"])
	}
	if out["web_base_path"] != "/panel/" {
		t.Fatalf("unexpected web_base_path: %v", out["web_base_path"])
	}
}

func TestFlattenPanelSettingsFields_Empty(t *testing.T) {
	out := flattenPanelSettingsFields(map[string]any{})
	if len(out) != 0 {
		t.Fatalf("expected empty, got %v", out)
	}
}

func TestFlattenAccountSettingsFields_Full(t *testing.T) {
	in := map[string]any{
		"twoFactorEnable": true,
		"twoFactorToken":  "secret",
	}
	out := flattenAccountSettingsFields(in)
	if out["two_factor_enable"] != true {
		t.Fatalf("unexpected: %v", out)
	}
	if out["two_factor_token"] != "secret" {
		t.Fatalf("unexpected: %v", out)
	}
}

func TestFlattenAccountSettingsFields_Empty(t *testing.T) {
	out := flattenAccountSettingsFields(map[string]any{})
	if len(out) != 0 {
		t.Fatalf("expected empty, got %v", out)
	}
}

func TestFlattenTelegramSettingsFields_Full(t *testing.T) {
	in := map[string]any{
		"tgBotEnable":      true,
		"tgBotToken":       "tok",
		"tgBotProxy":       "proxy",
		"tgBotAPIServer":   "api",
		"tgBotChatId":      "123",
		"tgLang":           "en",
		"tgRunTime":        "@daily",
		"tgBotBackup":      true,
		"tgBotLoginNotify": false,
		"tgCpu":            float64(80),
	}
	out := flattenTelegramSettingsFields(in)
	if out["tg_bot_enable"] != true {
		t.Fatalf("unexpected tg_bot_enable: %v", out["tg_bot_enable"])
	}
	if out["tg_bot_token"] != "tok" {
		t.Fatalf("unexpected tg_bot_token: %v", out["tg_bot_token"])
	}
	if out["tg_cpu"] != 80 {
		t.Fatalf("unexpected tg_cpu: %v", out["tg_cpu"])
	}
}

func TestFlattenTelegramSettingsFields_Empty(t *testing.T) {
	out := flattenTelegramSettingsFields(map[string]any{})
	if len(out) != 0 {
		t.Fatalf("expected empty, got %v", out)
	}
}

func TestFlattenSubscriptionSettingsFields_Full(t *testing.T) {
	in := map[string]any{
		"subEnable":     true,
		"subJsonEnable": false,
		"subTitle":      "My Sub",
		"subPort":       float64(443),
		"subEncrypt":    true,
	}
	out := flattenSubscriptionSettingsFields(in)
	if out["sub_enable"] != true {
		t.Fatalf("unexpected sub_enable: %v", out["sub_enable"])
	}
	if out["sub_title"] != "My Sub" {
		t.Fatalf("unexpected sub_title: %v", out["sub_title"])
	}
	if out["sub_port"] != 443 {
		t.Fatalf("unexpected sub_port: %v", out["sub_port"])
	}
}

func TestFlattenSubscriptionSettingsFields_Empty(t *testing.T) {
	out := flattenSubscriptionSettingsFields(map[string]any{})
	if len(out) != 0 {
		t.Fatalf("expected empty, got %v", out)
	}
}
