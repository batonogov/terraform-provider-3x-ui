package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestPanelGeneralSchemaAttributeShape asserts that the v3.2.0/v3.3.1 panel
// egress attributes exist with the expected Optional+Computed shape. It also
// exercises panelGeneralSchema() so the schema literal counts towards unit
// coverage (otherwise every panel_* attribute reads as uncovered by Codecov
// because no unit test instantiates the resource Schema()).
func TestPanelGeneralSchemaAttributeShape(t *testing.T) {
	s := panelGeneralSchema()
	requireStringAttr := func(name string) {
		t.Helper()
		attr, ok := s.Attributes[name]
		if !ok {
			t.Fatalf("attribute %q missing on panel_general schema", name)
		}
		str, ok := attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("attribute %q is not a StringAttribute", name)
		}
		if !str.Optional {
			t.Errorf("attribute %q must be Optional", name)
		}
		if !str.Computed {
			t.Errorf("attribute %q must be Computed", name)
		}
	}

	// Sanity-check that a well-known string attribute is present (confirms
	// the schema loaded, not just the two egress attrs).
	requireStringAttr("web_domain")

	// v3.2.0–v3.3.0 proxy URL (kept for backward compat).
	requireStringAttr("panel_proxy")

	// v3.3.1 outbound egress bridge (Xray outbound / balancer tag).
	requireStringAttr("panel_outbound")
}

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

func TestPanelSettingsNeedRestart_SubscriptionServerKey(t *testing.T) {
	// Toggling/changing the subscription server must trigger a panel restart — the sub
	// server is initialised at panel startup, so otherwise the subscription URL 404s.
	if !panelSettingsNeedRestart(map[string]any{"subEnable": false}, map[string]any{"subEnable": true}) {
		t.Fatalf("expected restart when the subscription server is enabled")
	}
	if !panelSettingsNeedRestart(map[string]any{"subPath": "/old/"}, map[string]any{"subPath": "/new/"}) {
		t.Fatalf("expected restart when the subscription path changes")
	}
	if !panelSettingsNeedRestart(map[string]any{"subPort": float64(2096)}, map[string]any{"subPort": float64(2097)}) {
		t.Fatalf("expected restart when the subscription port changes")
	}
}

func TestPanelSettingsNeedRestart_SubscriptionLinkKeyNoRestart(t *testing.T) {
	// Link-generation settings (read at request time) must NOT force a restart.
	if panelSettingsNeedRestart(map[string]any{"subURI": "https://a/"}, map[string]any{"subURI": "https://b/"}) {
		t.Fatalf("expected no restart for subURI (link-generation only)")
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
	if result["a"] != "1" {
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

func TestFlattenPanelSecurity(t *testing.T) {
	in := map[string]any{
		"twoFactorEnable": true,
		"twoFactorToken":  "secret",
	}
	m := flattenPanelSecurity(in)
	if m.TwoFactorEnable.ValueBool() != true {
		t.Fatalf("unexpected two_factor_enable: %v", m.TwoFactorEnable)
	}
	if m.TwoFactorToken.ValueString() != "secret" {
		t.Fatalf("unexpected two_factor_token: %v", m.TwoFactorToken)
	}
}

func TestFlattenPanelTelegram(t *testing.T) {
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
	m := flattenPanelTelegram(in)
	if m.TgBotEnable.ValueBool() != true {
		t.Fatalf("unexpected tg_bot_enable: %v", m.TgBotEnable)
	}
	if m.TgBotToken.ValueString() != "tok" {
		t.Fatalf("unexpected tg_bot_token: %v", m.TgBotToken)
	}
	if m.TgCPU.ValueInt64() != 80 {
		t.Fatalf("unexpected tg_cpu: %v", m.TgCPU)
	}
}

func TestFlattenPanelSubscription(t *testing.T) {
	in := map[string]any{
		"subEnable":        true,
		"subJsonEnable":    false,
		"subTitle":         "My Sub",
		"subPort":          float64(443),
		"subEncrypt":       true,
		"subEmailInRemark": false,
	}
	m := flattenPanelSubscription(in)
	if m.SubEnable.ValueBool() != true {
		t.Fatalf("unexpected sub_enable: %v", m.SubEnable)
	}
	if m.SubTitle.ValueString() != "My Sub" {
		t.Fatalf("unexpected sub_title: %v", m.SubTitle)
	}
	if m.SubPort.ValueInt64() != 443 {
		t.Fatalf("unexpected sub_port: %v", m.SubPort)
	}
	if m.SubEmailInRemark.ValueBool() {
		t.Fatalf("unexpected sub_email_in_remark: %v", m.SubEmailInRemark)
	}
}

func TestFlattenPanelGeneral(t *testing.T) {
	in := map[string]any{
		"webListen":                  "0.0.0.0",
		"webPort":                    float64(2053),
		"webBasePath":                "/panel/",
		"trustedProxyCIDRs":          "127.0.0.1/32,10.0.0.0/8",
		"pageSize":                   float64(25),
		"restartXrayOnClientDisable": true,
	}
	m := flattenPanelGeneral(in)
	if m.WebListen.ValueString() != "0.0.0.0" {
		t.Fatalf("unexpected web_listen: %v", m.WebListen)
	}
	if m.WebPort.ValueInt64() != 2053 {
		t.Fatalf("unexpected web_port: %v", m.WebPort)
	}
	if m.WebBasePath.ValueString() != "/panel/" {
		t.Fatalf("unexpected web_base_path: %v", m.WebBasePath)
	}
	if m.TrustedProxyCIDRs.ValueString() != "127.0.0.1/32,10.0.0.0/8" {
		t.Fatalf("unexpected trusted_proxy_cidrs: %v", m.TrustedProxyCIDRs)
	}
	if !m.RestartXrayOnClientDisable.ValueBool() {
		t.Fatalf("unexpected restart_xray_on_client_disable: %v", m.RestartXrayOnClientDisable)
	}
	expanded := expandPanelGeneral(m)
	if expanded["trustedProxyCIDRs"] != "127.0.0.1/32,10.0.0.0/8" {
		t.Fatalf("unexpected expanded trustedProxyCIDRs: %v", expanded["trustedProxyCIDRs"])
	}
	if expanded["restartXrayOnClientDisable"] != true {
		t.Fatalf("unexpected expanded restartXrayOnClientDisable: %v", expanded["restartXrayOnClientDisable"])
	}
}

func TestPanelProxyExpandFlatten(t *testing.T) {
	t.Run("proxy set", func(t *testing.T) {
		m := &PanelGeneralModel{
			PanelProxy: types.StringValue("socks5://proxy:1080"),
		}
		expanded := expandPanelGeneral(m)
		if expanded["panelProxy"] != "socks5://proxy:1080" {
			t.Fatalf("expected panelProxy=socks5://proxy:1080, got %v", expanded["panelProxy"])
		}
	})

	t.Run("proxy null", func(t *testing.T) {
		m := &PanelGeneralModel{
			PanelProxy: types.StringNull(),
		}
		expanded := expandPanelGeneral(m)
		if _, ok := expanded["panelProxy"]; ok {
			t.Fatalf("expected no panelProxy key, got %v", expanded["panelProxy"])
		}
	})

	t.Run("flatten with value", func(t *testing.T) {
		in := map[string]any{"panelProxy": "http://proxy:8080"}
		m := flattenPanelGeneral(in)
		if m.PanelProxy.ValueString() != "http://proxy:8080" {
			t.Fatalf("expected http://proxy:8080, got %q", m.PanelProxy)
		}
	})

	t.Run("flatten empty string", func(t *testing.T) {
		in := map[string]any{"panelProxy": ""}
		m := flattenPanelGeneral(in)
		if m.PanelProxy.ValueString() != "" {
			t.Fatalf("expected empty string for empty panelProxy, got %q", m.PanelProxy)
		}
	})

	t.Run("flatten missing key", func(t *testing.T) {
		in := map[string]any{}
		m := flattenPanelGeneral(in)
		if !m.PanelProxy.IsNull() {
			t.Fatalf("expected null for missing panelProxy, got %q", m.PanelProxy)
		}
	})
}

func TestPanelOutboundExpandFlatten(t *testing.T) {
	t.Run("outbound set", func(t *testing.T) {
		m := &PanelGeneralModel{
			PanelOutbound: types.StringValue("proxy-out"),
		}
		expanded := expandPanelGeneral(m)
		if expanded["panelOutbound"] != "proxy-out" {
			t.Fatalf("expected panelOutbound=proxy-out, got %v", expanded["panelOutbound"])
		}
	})

	t.Run("outbound null", func(t *testing.T) {
		m := &PanelGeneralModel{
			PanelOutbound: types.StringNull(),
		}
		expanded := expandPanelGeneral(m)
		if _, ok := expanded["panelOutbound"]; ok {
			t.Fatalf("expected no panelOutbound key, got %v", expanded["panelOutbound"])
		}
	})

	t.Run("flatten with value", func(t *testing.T) {
		in := map[string]any{"panelOutbound": "proxy-out"}
		m := flattenPanelGeneral(in)
		if m.PanelOutbound.ValueString() != "proxy-out" {
			t.Fatalf("expected proxy-out, got %q", m.PanelOutbound)
		}
	})

	t.Run("flatten empty string", func(t *testing.T) {
		in := map[string]any{"panelOutbound": ""}
		m := flattenPanelGeneral(in)
		if m.PanelOutbound.ValueString() != "" {
			t.Fatalf("expected empty string for empty panelOutbound, got %q", m.PanelOutbound)
		}
	})

	t.Run("flatten missing key", func(t *testing.T) {
		in := map[string]any{}
		m := flattenPanelGeneral(in)
		if !m.PanelOutbound.IsNull() {
			t.Fatalf("expected null for missing panelOutbound, got %q", m.PanelOutbound)
		}
	})
}

func TestPreserveSettingSecret_ConfiguredEmptyObservedNonEmpty(t *testing.T) {
	got := preserveSettingSecret(types.StringValue("existing-secret"), types.StringValue(""))
	if got.ValueString() != "" {
		t.Fatalf("expected configured empty string, got %q", got.ValueString())
	}
}

func TestPreserveSettingSecret_ConfiguredNonEmptyObservedEmpty(t *testing.T) {
	got := preserveSettingSecret(types.StringValue(""), types.StringValue("configured-secret"))
	if got.ValueString() != "configured-secret" {
		t.Fatalf("expected configured secret, got %q", got.ValueString())
	}
}

func TestPreserveSettingSecret_ConfiguredNonEmptyObservedRedacted(t *testing.T) {
	got := preserveSettingSecret(types.StringValue("********"), types.StringValue("configured-secret"))
	if got.ValueString() != "configured-secret" {
		t.Fatalf("expected configured secret, got %q", got.ValueString())
	}
}

func TestPreserveSettingSecret_ObservedDifferentNonEmpty(t *testing.T) {
	got := preserveSettingSecret(types.StringValue("remote-secret"), types.StringValue("state-secret"))
	if got.ValueString() != "remote-secret" {
		t.Fatalf("expected observed secret, got %q", got.ValueString())
	}
}

// preserveRemoved* echoes the configured value when the panel returns null for
// a field it dropped upstream (remarkModel, tgBotLoginNotify, subShowInfo,
// panelProxy, ...). It must return the configured value when observed is
// null/unknown (the v3.4.0 / v3.3.1+ case) and the observed value otherwise
// (older panels that still return the field).

func TestPreserveRemovedString_ObservedNull(t *testing.T) {
	got := preserveRemovedString(types.StringNull(), types.StringValue("-ieo"))
	if got.ValueString() != "-ieo" {
		t.Fatalf("expected configured echoed, got %q", got.ValueString())
	}
}

func TestPreserveRemovedString_ObservedUnknown(t *testing.T) {
	got := preserveRemovedString(types.StringUnknown(), types.StringValue("-ieo"))
	if got.ValueString() != "-ieo" {
		t.Fatalf("expected configured echoed, got %q", got.ValueString())
	}
}

func TestPreserveRemovedString_ObservedNonEmpty(t *testing.T) {
	// older panels still return the field — observed wins, no drift
	got := preserveRemovedString(types.StringValue("remote"), types.StringValue("state"))
	if got.ValueString() != "remote" {
		t.Fatalf("expected observed to win on panels that return it, got %q", got.ValueString())
	}
}

func TestPreserveRemovedBool_ObservedNull(t *testing.T) {
	got := preserveRemovedBool(types.BoolNull(), types.BoolValue(true))
	if !got.ValueBool() {
		t.Fatalf("expected configured true echoed, got false")
	}
}

func TestPreserveRemovedBool_ObservedUnknown(t *testing.T) {
	got := preserveRemovedBool(types.BoolUnknown(), types.BoolValue(true))
	if !got.ValueBool() {
		t.Fatalf("expected configured true echoed, got false")
	}
}

func TestPreserveRemovedBool_ObservedSet(t *testing.T) {
	got := preserveRemovedBool(types.BoolValue(false), types.BoolValue(true))
	if got.ValueBool() {
		t.Fatalf("expected observed to win on panels that return it, got true")
	}
}

func TestMergeSettingsForUpdate_PreservesCachedRedactedSecret(t *testing.T) {
	client := &Client{}
	client.rememberConfiguredSettingSecrets(map[string]any{"tgBotToken": "configured-token"})

	existing := map[string]any{
		"pageSize":     float64(25),
		"tgBotToken":   "",
		"ldapPassword": "********",
	}
	client.rememberConfiguredSettingSecrets(map[string]any{"ldapPassword": "configured-password"})

	merged := mergeSettingsForUpdate(client, existing, map[string]any{"pageSize": 50})
	if merged["tgBotToken"] != "configured-token" {
		t.Fatalf("expected cached tgBotToken, got %v", merged["tgBotToken"])
	}
	if merged["ldapPassword"] != "configured-password" {
		t.Fatalf("expected cached ldapPassword, got %v", merged["ldapPassword"])
	}
	if existing["tgBotToken"] != "" {
		t.Fatalf("mergeSettingsForUpdate mutated existing tgBotToken: %v", existing["tgBotToken"])
	}
}

func TestMergeSettingsForUpdate_DoesNotOverrideDesiredSecret(t *testing.T) {
	client := &Client{}
	client.rememberConfiguredSettingSecrets(map[string]any{"tgBotToken": "configured-token"})

	merged := mergeSettingsForUpdate(
		client,
		map[string]any{"pageSize": float64(25), "tgBotToken": ""},
		map[string]any{"pageSize": 50, "tgBotToken": ""},
	)
	if merged["tgBotToken"] != "" {
		t.Fatalf("expected desired empty tgBotToken, got %v", merged["tgBotToken"])
	}
}

func TestExpandPanelSecurity(t *testing.T) {
	m := &PanelSecurityModel{}
	m.TwoFactorEnable = typeBoolValue(false)
	m.TwoFactorToken = typeStringValue("tok")
	result := expandPanelSecurity(m)
	if result["twoFactorEnable"] != false {
		t.Fatalf("unexpected twoFactorEnable: %v", result["twoFactorEnable"])
	}
	if result["twoFactorToken"] != "tok" {
		t.Fatalf("unexpected twoFactorToken: %v", result["twoFactorToken"])
	}
}

func TestExpandPanelTelegram(t *testing.T) {
	m := &PanelTelegramModel{}
	m.TgBotEnable = typeBoolValue(true)
	m.TgBotToken = typeStringValue("token")
	m.TgCPU = typeInt64Value(80)
	result := expandPanelTelegram(m)
	if result["tgBotEnable"] != true {
		t.Fatalf("unexpected tgBotEnable: %v", result["tgBotEnable"])
	}
	if result["tgCpu"] != 80 {
		t.Fatalf("unexpected tgCpu: %v", result["tgCpu"])
	}
}

func TestFlattenPanelSubscription_ClashFields(t *testing.T) {
	in := map[string]any{
		"subEnable":             true,
		"subJsonEnable":         false,
		"subClashEnable":        true,
		"subClashPath":          "/clash/",
		"subClashURI":           "https://example.com/clash/",
		"subClashEnableRouting": true,
		"subClashRules":         "rule1,rule2",
		"subJsonFinalMask":      `{"tcp":"mask1"}`,
	}
	m := flattenPanelSubscription(in)
	if !m.SubClashEnable.ValueBool() {
		t.Fatalf("expected sub_clash_enable true")
	}
	if m.SubClashPath.ValueString() != "/clash/" {
		t.Fatalf("unexpected sub_clash_path: %s", m.SubClashPath.ValueString())
	}
	if m.SubClashURI.ValueString() != "https://example.com/clash/" {
		t.Fatalf("unexpected sub_clash_uri: %s", m.SubClashURI.ValueString())
	}
	if !m.SubClashEnableRouting.ValueBool() {
		t.Fatalf("expected sub_clash_enable_routing true")
	}
	if m.SubClashRules.ValueString() != "rule1,rule2" {
		t.Fatalf("unexpected sub_clash_rules: %s", m.SubClashRules.ValueString())
	}
	if m.SubJsonFinalMask.ValueString() != `{"tcp":"mask1"}` {
		t.Fatalf("unexpected sub_json_final_mask: %s", m.SubJsonFinalMask.ValueString())
	}
}

func TestExpandPanelSubscription_v328Fields(t *testing.T) {
	m := &PanelSubscriptionModel{
		SubEnable:             typeBoolValue(true),
		SubClashEnableRouting: typeBoolValue(true),
		SubClashRules:         typeStringValue("rule1,rule2"),
		SubJsonFinalMask:      typeStringValue(`{"tcp":"mask1"}`),
	}
	result := expandPanelSubscription(m)
	if result["subClashEnableRouting"] != true {
		t.Fatalf("expected subClashEnableRouting true, got %v", result["subClashEnableRouting"])
	}
	if result["subClashRules"] != "rule1,rule2" {
		t.Fatalf("unexpected subClashRules: %v", result["subClashRules"])
	}
	if result["subJsonFinalMask"] != `{"tcp":"mask1"}` {
		t.Fatalf("unexpected subJsonFinalMask: %v", result["subJsonFinalMask"])
	}
}

// Helper functions for creating typed values in tests
func typeBoolValue(v bool) types.Bool       { return types.BoolValue(v) }
func typeStringValue(v string) types.String { return types.StringValue(v) }
func typeInt64Value(v int64) types.Int64    { return types.Int64Value(v) }
