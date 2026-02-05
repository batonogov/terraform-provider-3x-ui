package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const remarkModelDefaultAPI = "-ieo"

func resourcePanelSettings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePanelSettingsApply,
		ReadContext:   resourcePanelSettingsRead,
		UpdateContext: resourcePanelSettingsApply,
		DeleteContext: resourceSettingsDelete,
		Schema:        panelSettingsSchemaFields(),
	}
}

func resourceAccountSettings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAccountSettingsApply,
		ReadContext:   resourceAccountSettingsRead,
		UpdateContext: resourceAccountSettingsApply,
		DeleteContext: resourceSettingsDelete,
		Schema:        accountSettingsSchemaFields(),
	}
}

func resourceTelegramSettings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTelegramSettingsApply,
		ReadContext:   resourceTelegramSettingsRead,
		UpdateContext: resourceTelegramSettingsApply,
		DeleteContext: resourceSettingsDelete,
		Schema:        telegramSettingsSchemaFields(),
	}
}

func resourceSubscriptionSettings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSubscriptionSettingsApply,
		ReadContext:   resourceSubscriptionSettingsRead,
		UpdateContext: resourceSubscriptionSettingsApply,
		DeleteContext: resourceSettingsDelete,
		Schema:        subscriptionSettingsSchemaFields(),
	}
}

func resourcePanelSettingsApply(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*Client)
	desired, ok, err := expandPanelSettingsFields(d)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ok {
		if d.Id() == "" {
			d.SetId("settings")
		}
		return resourceSettingsReadWith(ctx, d, meta, flattenPanelSettingsFields)
	}

	existing, err := client.GetSettings(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	needRestart := panelSettingsNeedRestart(existing, desired)
	merged := mergeSettings(existing, desired)
	if err := client.UpdateSettings(ctx, merged); err != nil {
		return diag.FromErr(err)
	}
	if needRestart {
		if err := client.RestartPanel(ctx); err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId("settings")
	return resourceSettingsReadWith(ctx, d, meta, flattenPanelSettingsFields)
}

func resourceAccountSettingsApply(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return resourceSettingsApplyWith(ctx, d, meta, expandAccountSettingsFields, flattenAccountSettingsFields)
}

func resourceTelegramSettingsApply(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return resourceSettingsApplyWith(ctx, d, meta, expandTelegramSettingsFields, flattenTelegramSettingsFields)
}

func resourceSubscriptionSettingsApply(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return resourceSettingsApplyWith(ctx, d, meta, expandSubscriptionSettingsFields, flattenSubscriptionSettingsFields)
}

func resourcePanelSettingsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return resourceSettingsReadWith(ctx, d, meta, flattenPanelSettingsFields)
}

func resourceAccountSettingsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return resourceSettingsReadWith(ctx, d, meta, flattenAccountSettingsFields)
}

func resourceTelegramSettingsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return resourceSettingsReadWith(ctx, d, meta, flattenTelegramSettingsFields)
}

func resourceSubscriptionSettingsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return resourceSettingsReadWith(ctx, d, meta, flattenSubscriptionSettingsFields)
}

func resourceSettingsApplyWith(
	ctx context.Context,
	d *schema.ResourceData,
	meta any,
	expand func(*schema.ResourceData) (map[string]any, bool, error),
	flatten func(map[string]any) map[string]any,
) diag.Diagnostics {
	client := meta.(*Client)
	desired, ok, err := expand(d)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ok {
		if d.Id() == "" {
			d.SetId("settings")
		}
		return resourceSettingsReadWith(ctx, d, meta, flatten)
	}

	existing, err := client.GetSettings(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	merged := mergeSettings(existing, desired)
	if err := client.UpdateSettings(ctx, merged); err != nil {
		return diag.FromErr(err)
	}
	d.SetId("settings")
	return resourceSettingsReadWith(ctx, d, meta, flatten)
}

func resourceSettingsReadWith(
	ctx context.Context,
	d *schema.ResourceData,
	meta any,
	flatten func(map[string]any) map[string]any,
) diag.Diagnostics {
	client := meta.(*Client)
	settings, err := client.GetSettings(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if flatten != nil {
		if err := setSettingsFields(d, flatten(settings)); err != nil {
			return diag.FromErr(err)
		}
	}
	if d.Id() == "" {
		d.SetId("settings")
	}
	return nil
}

func panelSettingsNeedRestart(existing, desired map[string]any) bool {
	restartKeys := []string{
		"webListen",
		"webDomain",
		"webPort",
		"webBasePath",
		"webCertFile",
		"webKeyFile",
		"sessionMaxAge",
	}
	for _, key := range restartKeys {
		newVal, ok := desired[key]
		if !ok {
			continue
		}
		oldVal, ok := existing[key]
		if !ok {
			return true
		}
		if !settingsValueEqual(oldVal, newVal) {
			return true
		}
	}
	return false
}

func settingsValueEqual(a, b any) bool {
	switch av := a.(type) {
	case nil:
		return b == nil
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case float64:
		return numberValueEqual(av, b)
	case float32:
		return numberValueEqual(float64(av), b)
	case int:
		return numberValueEqual(float64(av), b)
	case int8:
		return numberValueEqual(float64(av), b)
	case int16:
		return numberValueEqual(float64(av), b)
	case int32:
		return numberValueEqual(float64(av), b)
	case int64:
		return numberValueEqual(float64(av), b)
	case uint:
		return numberValueEqual(float64(av), b)
	case uint8:
		return numberValueEqual(float64(av), b)
	case uint16:
		return numberValueEqual(float64(av), b)
	case uint32:
		return numberValueEqual(float64(av), b)
	case uint64:
		return numberValueEqual(float64(av), b)
	default:
		return false
	}
}

func numberValueEqual(a float64, b any) bool {
	switch bv := b.(type) {
	case float64:
		return a == bv
	case float32:
		return a == float64(bv)
	case int:
		return a == float64(bv)
	case int8:
		return a == float64(bv)
	case int16:
		return a == float64(bv)
	case int32:
		return a == float64(bv)
	case int64:
		return a == float64(bv)
	case uint:
		return a == float64(bv)
	case uint8:
		return a == float64(bv)
	case uint16:
		return a == float64(bv)
	case uint32:
		return a == float64(bv)
	case uint64:
		return a == float64(bv)
	default:
		return false
	}
}

func setSettingsFields(d *schema.ResourceData, fields map[string]any) error {
	for k, v := range fields {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

func panelSettingsSchemaFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"web_listen": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"web_domain": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"web_port": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  2053,
		},
		"web_base_path": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "/",
		},
		"session_max_age": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  360,
		},
		"page_size": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  25,
		},
		"remark_model": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  remarkModelDefaultAPI,
		},
		"date_picker": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "gregorian",
		},
		"time_location": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "Local",
		},
		"expire_diff": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  0,
		},
		"traffic_diff": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  0,
		},
		"web_cert_file": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"web_key_file": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"external_traffic_inform_enable": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"external_traffic_inform_uri": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"ldap_enable": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"ldap_host": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"ldap_port": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  389,
		},
		"ldap_use_tls": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"ldap_bind_dn": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"ldap_password": {
			Type:      schema.TypeString,
			Optional:  true,
			Sensitive: true,
			Default:   "",
		},
		"ldap_base_dn": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"ldap_user_filter": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "(objectClass=person)",
		},
		"ldap_user_attr": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "mail",
		},
		"ldap_vless_field": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "vless_enabled",
		},
		"ldap_sync_cron": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "@every 1m",
		},
		"ldap_flag_field": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"ldap_truthy_values": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "true,1,yes,on",
		},
		"ldap_invert_flag": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"ldap_inbound_tags": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"ldap_auto_create": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"ldap_auto_delete": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"ldap_default_total_gb": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  0,
		},
		"ldap_default_expiry_days": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  0,
		},
		"ldap_default_limit_ip": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  0,
		},
	}
}

func accountSettingsSchemaFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"two_factor_enable": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		"two_factor_token": {
			Type:      schema.TypeString,
			Optional:  true,
			Computed:  true,
			Sensitive: true,
		},
	}
}

func telegramSettingsSchemaFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tg_bot_enable": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		"tg_bot_token": {
			Type:      schema.TypeString,
			Optional:  true,
			Computed:  true,
			Sensitive: true,
		},
		"tg_bot_proxy": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"tg_bot_api_server": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"tg_bot_chat_id": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"tg_lang": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"tg_run_time": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"tg_bot_backup": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		"tg_bot_login_notify": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		"tg_cpu": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
	}
}

func subscriptionSettingsSchemaFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"sub_enable": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		"sub_json_enable": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		"sub_title": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"sub_support_url": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"sub_profile_url": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"sub_announce": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"sub_enable_routing": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		"sub_routing_rules": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"sub_listen": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"sub_port": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"sub_path": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"sub_domain": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"sub_cert_file": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"sub_key_file": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"sub_updates": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"sub_encrypt": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		"sub_show_info": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		"sub_uri": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"sub_json_path": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"sub_json_uri": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"sub_json_fragment": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"sub_json_noises": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"sub_json_mux": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"sub_json_rules": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
	}
}

func expandPanelSettingsFields(d *schema.ResourceData) (map[string]any, bool, error) {
	payload := map[string]any{}
	if v, ok := getStringField(d, "web_listen"); ok {
		payload["webListen"] = v
	}
	if v, ok := getStringField(d, "web_domain"); ok {
		payload["webDomain"] = v
	}
	if v, ok, err := getPortField(d, "web_port"); err != nil {
		return nil, false, err
	} else if ok {
		payload["webPort"] = v
	}
	if v, ok := getStringField(d, "web_base_path"); ok {
		payload["webBasePath"] = v
	}
	if v, ok := getIntField(d, "session_max_age"); ok {
		payload["sessionMaxAge"] = v
	}
	if v, ok := getIntField(d, "page_size"); ok {
		payload["pageSize"] = v
	}
	if v, ok := getStringField(d, "remark_model"); ok {
		payload["remarkModel"] = v
	}
	if v, ok := getStringField(d, "date_picker"); ok {
		payload["datepicker"] = v
	}
	if v, ok := getStringField(d, "time_location"); ok {
		payload["timeLocation"] = v
	}
	if v, ok := getIntField(d, "expire_diff"); ok {
		payload["expireDiff"] = v
	}
	if v, ok := getIntField(d, "traffic_diff"); ok {
		payload["trafficDiff"] = v
	}
	if v, ok := getStringField(d, "web_cert_file"); ok {
		payload["webCertFile"] = v
	}
	if v, ok := getStringField(d, "web_key_file"); ok {
		payload["webKeyFile"] = v
	}
	if v, ok := getBoolField(d, "external_traffic_inform_enable"); ok {
		payload["externalTrafficInformEnable"] = v
	}
	if v, ok := getStringField(d, "external_traffic_inform_uri"); ok {
		payload["externalTrafficInformURI"] = v
	}
	if v, ok := getBoolField(d, "ldap_enable"); ok {
		payload["ldapEnable"] = v
	}
	if v, ok := getStringField(d, "ldap_host"); ok {
		payload["ldapHost"] = v
	}
	if v, ok, err := getPortField(d, "ldap_port"); err != nil {
		return nil, false, err
	} else if ok {
		payload["ldapPort"] = v
	}
	if v, ok := getBoolField(d, "ldap_use_tls"); ok {
		payload["ldapUseTLS"] = v
	}
	if v, ok := getStringField(d, "ldap_bind_dn"); ok {
		payload["ldapBindDN"] = v
	}
	if v, ok := getStringField(d, "ldap_password"); ok {
		payload["ldapPassword"] = v
	}
	if v, ok := getStringField(d, "ldap_base_dn"); ok {
		payload["ldapBaseDN"] = v
	}
	if v, ok := getStringField(d, "ldap_user_filter"); ok {
		payload["ldapUserFilter"] = v
	}
	if v, ok := getStringField(d, "ldap_user_attr"); ok {
		payload["ldapUserAttr"] = v
	}
	if v, ok := getStringField(d, "ldap_vless_field"); ok {
		payload["ldapVlessField"] = v
	}
	if v, ok := getStringField(d, "ldap_sync_cron"); ok {
		payload["ldapSyncCron"] = v
	}
	if v, ok := getStringField(d, "ldap_flag_field"); ok {
		payload["ldapFlagField"] = v
	}
	if v, ok := getStringField(d, "ldap_truthy_values"); ok {
		payload["ldapTruthyValues"] = v
	}
	if v, ok := getBoolField(d, "ldap_invert_flag"); ok {
		payload["ldapInvertFlag"] = v
	}
	if v, ok := getStringField(d, "ldap_inbound_tags"); ok {
		payload["ldapInboundTags"] = v
	}
	if v, ok := getBoolField(d, "ldap_auto_create"); ok {
		payload["ldapAutoCreate"] = v
	}
	if v, ok := getBoolField(d, "ldap_auto_delete"); ok {
		payload["ldapAutoDelete"] = v
	}
	if v, ok := getIntField(d, "ldap_default_total_gb"); ok {
		payload["ldapDefaultTotalGB"] = v
	}
	if v, ok := getIntField(d, "ldap_default_expiry_days"); ok {
		payload["ldapDefaultExpiryDays"] = v
	}
	if v, ok := getIntField(d, "ldap_default_limit_ip"); ok {
		payload["ldapDefaultLimitIP"] = v
	}
	if len(payload) == 0 {
		return nil, false, nil
	}
	return payload, true, nil
}

func expandAccountSettingsFields(d *schema.ResourceData) (map[string]any, bool, error) {
	payload := map[string]any{}
	if v, ok := getBoolField(d, "two_factor_enable"); ok {
		payload["twoFactorEnable"] = v
	}
	if v, ok := getStringField(d, "two_factor_token"); ok {
		payload["twoFactorToken"] = v
	}
	if len(payload) == 0 {
		return nil, false, nil
	}
	return payload, true, nil
}

func expandTelegramSettingsFields(d *schema.ResourceData) (map[string]any, bool, error) {
	payload := map[string]any{}
	if v, ok := getBoolField(d, "tg_bot_enable"); ok {
		payload["tgBotEnable"] = v
	}
	if v, ok := getStringField(d, "tg_bot_token"); ok {
		payload["tgBotToken"] = v
	}
	if v, ok := getStringField(d, "tg_bot_proxy"); ok {
		payload["tgBotProxy"] = v
	}
	if v, ok := getStringField(d, "tg_bot_api_server"); ok {
		payload["tgBotAPIServer"] = v
	}
	if v, ok := getStringField(d, "tg_bot_chat_id"); ok {
		payload["tgBotChatId"] = v
	}
	if v, ok := getStringField(d, "tg_lang"); ok {
		payload["tgLang"] = v
	}
	if v, ok := getStringField(d, "tg_run_time"); ok {
		payload["tgRunTime"] = v
	}
	if v, ok := getBoolField(d, "tg_bot_backup"); ok {
		payload["tgBotBackup"] = v
	}
	if v, ok := getBoolField(d, "tg_bot_login_notify"); ok {
		payload["tgBotLoginNotify"] = v
	}
	if v, ok := getIntField(d, "tg_cpu"); ok {
		payload["tgCpu"] = v
	}
	if len(payload) == 0 {
		return nil, false, nil
	}
	return payload, true, nil
}

func expandSubscriptionSettingsFields(d *schema.ResourceData) (map[string]any, bool, error) {
	payload := map[string]any{}
	if v, ok := getBoolField(d, "sub_enable"); ok {
		payload["subEnable"] = v
	}
	if v, ok := getBoolField(d, "sub_json_enable"); ok {
		payload["subJsonEnable"] = v
	}
	if v, ok := getStringField(d, "sub_title"); ok {
		payload["subTitle"] = v
	}
	if v, ok := getStringField(d, "sub_support_url"); ok {
		payload["subSupportUrl"] = v
	}
	if v, ok := getStringField(d, "sub_profile_url"); ok {
		payload["subProfileUrl"] = v
	}
	if v, ok := getStringField(d, "sub_announce"); ok {
		payload["subAnnounce"] = v
	}
	if v, ok := getBoolField(d, "sub_enable_routing"); ok {
		payload["subEnableRouting"] = v
	}
	if v, ok := getStringField(d, "sub_routing_rules"); ok {
		payload["subRoutingRules"] = v
	}
	if v, ok := getStringField(d, "sub_listen"); ok {
		payload["subListen"] = v
	}
	if v, ok, err := getPortField(d, "sub_port"); err != nil {
		return nil, false, err
	} else if ok {
		payload["subPort"] = v
	}
	if v, ok := getStringField(d, "sub_path"); ok {
		payload["subPath"] = v
	}
	if v, ok := getStringField(d, "sub_domain"); ok {
		payload["subDomain"] = v
	}
	if v, ok := getStringField(d, "sub_cert_file"); ok {
		payload["subCertFile"] = v
	}
	if v, ok := getStringField(d, "sub_key_file"); ok {
		payload["subKeyFile"] = v
	}
	if v, ok := getIntField(d, "sub_updates"); ok {
		payload["subUpdates"] = v
	}
	if v, ok := getBoolField(d, "sub_encrypt"); ok {
		payload["subEncrypt"] = v
	}
	if v, ok := getBoolField(d, "sub_show_info"); ok {
		payload["subShowInfo"] = v
	}
	if v, ok := getStringField(d, "sub_uri"); ok {
		payload["subURI"] = v
	}
	if v, ok := getStringField(d, "sub_json_path"); ok {
		payload["subJsonPath"] = v
	}
	if v, ok := getStringField(d, "sub_json_uri"); ok {
		payload["subJsonURI"] = v
	}
	if v, ok := getStringField(d, "sub_json_fragment"); ok {
		payload["subJsonFragment"] = v
	}
	if v, ok := getStringField(d, "sub_json_noises"); ok {
		payload["subJsonNoises"] = v
	}
	if v, ok := getStringField(d, "sub_json_mux"); ok {
		payload["subJsonMux"] = v
	}
	if v, ok := getStringField(d, "sub_json_rules"); ok {
		payload["subJsonRules"] = v
	}
	if len(payload) == 0 {
		return nil, false, nil
	}
	return payload, true, nil
}

func flattenPanelSettingsFields(in map[string]any) map[string]any {
	out := map[string]any{}
	if v, ok := in["webListen"]; ok {
		out["web_listen"] = stringValue(v)
	}
	if v, ok := in["webDomain"]; ok {
		out["web_domain"] = stringValue(v)
	}
	if v, ok := in["webPort"]; ok {
		out["web_port"] = intValue(v)
	}
	if v, ok := in["webBasePath"]; ok {
		out["web_base_path"] = stringValue(v)
	}
	if v, ok := in["sessionMaxAge"]; ok {
		out["session_max_age"] = intValue(v)
	}
	if v, ok := in["pageSize"]; ok {
		out["page_size"] = intValue(v)
	}
	if v, ok := in["remarkModel"]; ok {
		out["remark_model"] = stringValue(v)
	}
	if v, ok := in["datepicker"]; ok {
		out["date_picker"] = stringValue(v)
	}
	if v, ok := in["timeLocation"]; ok {
		out["time_location"] = stringValue(v)
	}
	if v, ok := in["expireDiff"]; ok {
		out["expire_diff"] = intValue(v)
	}
	if v, ok := in["trafficDiff"]; ok {
		out["traffic_diff"] = intValue(v)
	}
	if v, ok := in["webCertFile"]; ok {
		out["web_cert_file"] = stringValue(v)
	}
	if v, ok := in["webKeyFile"]; ok {
		out["web_key_file"] = stringValue(v)
	}
	if v, ok := in["externalTrafficInformEnable"]; ok {
		out["external_traffic_inform_enable"] = boolValue(v)
	}
	if v, ok := in["externalTrafficInformURI"]; ok {
		out["external_traffic_inform_uri"] = stringValue(v)
	}
	if v, ok := in["ldapEnable"]; ok {
		out["ldap_enable"] = boolValue(v)
	}
	if v, ok := in["ldapHost"]; ok {
		out["ldap_host"] = stringValue(v)
	}
	if v, ok := in["ldapPort"]; ok {
		out["ldap_port"] = intValue(v)
	}
	if v, ok := in["ldapUseTLS"]; ok {
		out["ldap_use_tls"] = boolValue(v)
	}
	if v, ok := in["ldapBindDN"]; ok {
		out["ldap_bind_dn"] = stringValue(v)
	}
	if v, ok := in["ldapPassword"]; ok {
		out["ldap_password"] = stringValue(v)
	}
	if v, ok := in["ldapBaseDN"]; ok {
		out["ldap_base_dn"] = stringValue(v)
	}
	if v, ok := in["ldapUserFilter"]; ok {
		out["ldap_user_filter"] = stringValue(v)
	}
	if v, ok := in["ldapUserAttr"]; ok {
		out["ldap_user_attr"] = stringValue(v)
	}
	if v, ok := in["ldapVlessField"]; ok {
		out["ldap_vless_field"] = stringValue(v)
	}
	if v, ok := in["ldapSyncCron"]; ok {
		out["ldap_sync_cron"] = stringValue(v)
	}
	if v, ok := in["ldapFlagField"]; ok {
		out["ldap_flag_field"] = stringValue(v)
	}
	if v, ok := in["ldapTruthyValues"]; ok {
		out["ldap_truthy_values"] = stringValue(v)
	}
	if v, ok := in["ldapInvertFlag"]; ok {
		out["ldap_invert_flag"] = boolValue(v)
	}
	if v, ok := in["ldapInboundTags"]; ok {
		out["ldap_inbound_tags"] = stringValue(v)
	}
	if v, ok := in["ldapAutoCreate"]; ok {
		out["ldap_auto_create"] = boolValue(v)
	}
	if v, ok := in["ldapAutoDelete"]; ok {
		out["ldap_auto_delete"] = boolValue(v)
	}
	if v, ok := in["ldapDefaultTotalGB"]; ok {
		out["ldap_default_total_gb"] = intValue(v)
	}
	if v, ok := in["ldapDefaultExpiryDays"]; ok {
		out["ldap_default_expiry_days"] = intValue(v)
	}
	if v, ok := in["ldapDefaultLimitIP"]; ok {
		out["ldap_default_limit_ip"] = intValue(v)
	}
	return out
}

func flattenAccountSettingsFields(in map[string]any) map[string]any {
	out := map[string]any{}
	if v, ok := in["twoFactorEnable"]; ok {
		out["two_factor_enable"] = boolValue(v)
	}
	if v, ok := in["twoFactorToken"]; ok {
		out["two_factor_token"] = stringValue(v)
	}
	return out
}

func flattenTelegramSettingsFields(in map[string]any) map[string]any {
	out := map[string]any{}
	if v, ok := in["tgBotEnable"]; ok {
		out["tg_bot_enable"] = boolValue(v)
	}
	if v, ok := in["tgBotToken"]; ok {
		out["tg_bot_token"] = stringValue(v)
	}
	if v, ok := in["tgBotProxy"]; ok {
		out["tg_bot_proxy"] = stringValue(v)
	}
	if v, ok := in["tgBotAPIServer"]; ok {
		out["tg_bot_api_server"] = stringValue(v)
	}
	if v, ok := in["tgBotChatId"]; ok {
		out["tg_bot_chat_id"] = stringValue(v)
	}
	if v, ok := in["tgLang"]; ok {
		out["tg_lang"] = stringValue(v)
	}
	if v, ok := in["tgRunTime"]; ok {
		out["tg_run_time"] = stringValue(v)
	}
	if v, ok := in["tgBotBackup"]; ok {
		out["tg_bot_backup"] = boolValue(v)
	}
	if v, ok := in["tgBotLoginNotify"]; ok {
		out["tg_bot_login_notify"] = boolValue(v)
	}
	if v, ok := in["tgCpu"]; ok {
		out["tg_cpu"] = intValue(v)
	}
	return out
}

func flattenSubscriptionSettingsFields(in map[string]any) map[string]any {
	out := map[string]any{}
	if v, ok := in["subEnable"]; ok {
		out["sub_enable"] = boolValue(v)
	}
	if v, ok := in["subJsonEnable"]; ok {
		out["sub_json_enable"] = boolValue(v)
	}
	if v, ok := in["subTitle"]; ok {
		out["sub_title"] = stringValue(v)
	}
	if v, ok := in["subSupportUrl"]; ok {
		out["sub_support_url"] = stringValue(v)
	}
	if v, ok := in["subProfileUrl"]; ok {
		out["sub_profile_url"] = stringValue(v)
	}
	if v, ok := in["subAnnounce"]; ok {
		out["sub_announce"] = stringValue(v)
	}
	if v, ok := in["subEnableRouting"]; ok {
		out["sub_enable_routing"] = boolValue(v)
	}
	if v, ok := in["subRoutingRules"]; ok {
		out["sub_routing_rules"] = stringValue(v)
	}
	if v, ok := in["subListen"]; ok {
		out["sub_listen"] = stringValue(v)
	}
	if v, ok := in["subPort"]; ok {
		out["sub_port"] = intValue(v)
	}
	if v, ok := in["subPath"]; ok {
		out["sub_path"] = stringValue(v)
	}
	if v, ok := in["subDomain"]; ok {
		out["sub_domain"] = stringValue(v)
	}
	if v, ok := in["subCertFile"]; ok {
		out["sub_cert_file"] = stringValue(v)
	}
	if v, ok := in["subKeyFile"]; ok {
		out["sub_key_file"] = stringValue(v)
	}
	if v, ok := in["subUpdates"]; ok {
		out["sub_updates"] = intValue(v)
	}
	if v, ok := in["subEncrypt"]; ok {
		out["sub_encrypt"] = boolValue(v)
	}
	if v, ok := in["subShowInfo"]; ok {
		out["sub_show_info"] = boolValue(v)
	}
	if v, ok := in["subURI"]; ok {
		out["sub_uri"] = stringValue(v)
	}
	if v, ok := in["subJsonPath"]; ok {
		out["sub_json_path"] = stringValue(v)
	}
	if v, ok := in["subJsonURI"]; ok {
		out["sub_json_uri"] = stringValue(v)
	}
	if v, ok := in["subJsonFragment"]; ok {
		out["sub_json_fragment"] = stringValue(v)
	}
	if v, ok := in["subJsonNoises"]; ok {
		out["sub_json_noises"] = stringValue(v)
	}
	if v, ok := in["subJsonMux"]; ok {
		out["sub_json_mux"] = stringValue(v)
	}
	if v, ok := in["subJsonRules"]; ok {
		out["sub_json_rules"] = stringValue(v)
	}
	return out
}
