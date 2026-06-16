package provider

import (
	"context"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// Panel Security model, schema, expand/flatten
// ---------------------------------------------------------------------------

type PanelSecurityModel struct {
	ID                      types.String `tfsdk:"id"`
	TwoFactorEnable         types.Bool   `tfsdk:"two_factor_enable"`
	TwoFactorToken          types.String `tfsdk:"two_factor_token"`
	TwoFactorTokenWO        types.String `tfsdk:"two_factor_token_wo"`
	TwoFactorTokenWOVersion types.Int64  `tfsdk:"two_factor_token_wo_version"`
}

func panelSecuritySchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"two_factor_enable": schema.BoolAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"two_factor_token": schema.StringAttribute{
				Optional: true, Computed: true, Sensitive: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Validators: []validator.String{
					stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("two_factor_token_wo")),
				},
			},
			"two_factor_token_wo": schema.StringAttribute{
				Optional:  true,
				WriteOnly: true,
			},
			"two_factor_token_wo_version": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AlsoRequires(path.MatchRoot("two_factor_token_wo")),
				},
			},
		},
	}
}

func expandPanelSecurity(m *PanelSecurityModel) map[string]any {
	payload := map[string]any{}
	if !m.TwoFactorEnable.IsNull() && !m.TwoFactorEnable.IsUnknown() {
		payload["twoFactorEnable"] = m.TwoFactorEnable.ValueBool()
	}
	if !m.TwoFactorTokenWO.IsNull() && !m.TwoFactorTokenWO.IsUnknown() {
		payload["twoFactorToken"] = m.TwoFactorTokenWO.ValueString()
	} else if !m.TwoFactorToken.IsNull() && !m.TwoFactorToken.IsUnknown() {
		payload["twoFactorToken"] = m.TwoFactorToken.ValueString()
	}
	return payload
}

func flattenPanelSecurity(in map[string]any) *PanelSecurityModel {
	m := &PanelSecurityModel{
		ID: types.StringValue("settings"),
	}
	if v, ok := in["twoFactorEnable"]; ok {
		m.TwoFactorEnable = types.BoolValue(boolValue(v))
	}
	if v, ok := in["twoFactorToken"]; ok {
		m.TwoFactorToken = types.StringValue(stringValue(v))
	}
	return m
}

// ---------------------------------------------------------------------------
// Panel Telegram model, schema, expand/flatten
// ---------------------------------------------------------------------------

type PanelTelegramModel struct {
	ID                  types.String `tfsdk:"id"`
	TgBotEnable         types.Bool   `tfsdk:"tg_bot_enable"`
	TgBotToken          types.String `tfsdk:"tg_bot_token"`
	TgBotTokenWO        types.String `tfsdk:"tg_bot_token_wo"`
	TgBotTokenWOVersion types.Int64  `tfsdk:"tg_bot_token_wo_version"`
	TgBotProxy          types.String `tfsdk:"tg_bot_proxy"`
	TgBotAPIServer      types.String `tfsdk:"tg_bot_api_server"`
	TgBotChatID         types.String `tfsdk:"tg_bot_chat_id"`
	TgLang              types.String `tfsdk:"tg_lang"`
	TgRunTime           types.String `tfsdk:"tg_run_time"`
	TgBotBackup         types.Bool   `tfsdk:"tg_bot_backup"`
	TgBotLoginNotify    types.Bool   `tfsdk:"tg_bot_login_notify"`
	TgCPU               types.Int64  `tfsdk:"tg_cpu"`
}

func panelTelegramSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tg_bot_enable": schema.BoolAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"tg_bot_token": schema.StringAttribute{
				Optional: true, Computed: true, Sensitive: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Validators: []validator.String{
					stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("tg_bot_token_wo")),
				},
			},
			"tg_bot_token_wo": schema.StringAttribute{
				Optional:  true,
				WriteOnly: true,
			},
			"tg_bot_token_wo_version": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AlsoRequires(path.MatchRoot("tg_bot_token_wo")),
				},
			},
			"tg_bot_proxy": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tg_bot_api_server": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tg_bot_chat_id": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tg_lang": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tg_run_time": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tg_bot_backup": schema.BoolAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"tg_bot_login_notify": schema.BoolAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"tg_cpu": schema.Int64Attribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
		},
	}
}

func expandPanelTelegram(m *PanelTelegramModel) map[string]any {
	payload := map[string]any{}
	if !m.TgBotEnable.IsNull() && !m.TgBotEnable.IsUnknown() {
		payload["tgBotEnable"] = m.TgBotEnable.ValueBool()
	}
	if !m.TgBotTokenWO.IsNull() && !m.TgBotTokenWO.IsUnknown() {
		payload["tgBotToken"] = m.TgBotTokenWO.ValueString()
	} else if !m.TgBotToken.IsNull() && !m.TgBotToken.IsUnknown() {
		payload["tgBotToken"] = m.TgBotToken.ValueString()
	}
	if !m.TgBotProxy.IsNull() && !m.TgBotProxy.IsUnknown() {
		payload["tgBotProxy"] = m.TgBotProxy.ValueString()
	}
	if !m.TgBotAPIServer.IsNull() && !m.TgBotAPIServer.IsUnknown() {
		payload["tgBotAPIServer"] = m.TgBotAPIServer.ValueString()
	}
	if !m.TgBotChatID.IsNull() && !m.TgBotChatID.IsUnknown() {
		payload["tgBotChatId"] = m.TgBotChatID.ValueString()
	}
	if !m.TgLang.IsNull() && !m.TgLang.IsUnknown() {
		payload["tgLang"] = m.TgLang.ValueString()
	}
	if !m.TgRunTime.IsNull() && !m.TgRunTime.IsUnknown() {
		payload["tgRunTime"] = m.TgRunTime.ValueString()
	}
	if !m.TgBotBackup.IsNull() && !m.TgBotBackup.IsUnknown() {
		payload["tgBotBackup"] = m.TgBotBackup.ValueBool()
	}
	if !m.TgBotLoginNotify.IsNull() && !m.TgBotLoginNotify.IsUnknown() {
		payload["tgBotLoginNotify"] = m.TgBotLoginNotify.ValueBool()
	}
	if !m.TgCPU.IsNull() && !m.TgCPU.IsUnknown() {
		payload["tgCpu"] = int(m.TgCPU.ValueInt64())
	}
	return payload
}

func flattenPanelTelegram(in map[string]any) *PanelTelegramModel {
	m := &PanelTelegramModel{
		ID: types.StringValue("settings"),
	}
	if v, ok := in["tgBotEnable"]; ok {
		m.TgBotEnable = types.BoolValue(boolValue(v))
	}
	if v, ok := in["tgBotToken"]; ok {
		m.TgBotToken = types.StringValue(stringValue(v))
	}
	if v, ok := in["tgBotProxy"]; ok {
		m.TgBotProxy = types.StringValue(stringValue(v))
	}
	if v, ok := in["tgBotAPIServer"]; ok {
		m.TgBotAPIServer = types.StringValue(stringValue(v))
	}
	if v, ok := in["tgBotChatId"]; ok {
		m.TgBotChatID = types.StringValue(stringValue(v))
	}
	if v, ok := in["tgLang"]; ok {
		m.TgLang = types.StringValue(stringValue(v))
	}
	if v, ok := in["tgRunTime"]; ok {
		m.TgRunTime = types.StringValue(stringValue(v))
	}
	if v, ok := in["tgBotBackup"]; ok {
		m.TgBotBackup = types.BoolValue(boolValue(v))
	}
	if v, ok := in["tgBotLoginNotify"]; ok {
		m.TgBotLoginNotify = types.BoolValue(boolValue(v))
	}
	if v, ok := in["tgCpu"]; ok {
		m.TgCPU = types.Int64Value(int64(intValue(v)))
	}
	return m
}

// ---------------------------------------------------------------------------
// Panel Subscription model, schema, expand/flatten
// ---------------------------------------------------------------------------

type PanelSubscriptionModel struct {
	ID                    types.String `tfsdk:"id"`
	SubEnable             types.Bool   `tfsdk:"sub_enable"`
	SubJsonEnable         types.Bool   `tfsdk:"sub_json_enable"`
	SubTitle              types.String `tfsdk:"sub_title"`
	SubSupportURL         types.String `tfsdk:"sub_support_url"`
	SubProfileURL         types.String `tfsdk:"sub_profile_url"`
	SubAnnounce           types.String `tfsdk:"sub_announce"`
	SubEnableRouting      types.Bool   `tfsdk:"sub_enable_routing"`
	SubRoutingRules       types.String `tfsdk:"sub_routing_rules"`
	SubListen             types.String `tfsdk:"sub_listen"`
	SubPort               types.Int64  `tfsdk:"sub_port"`
	SubPath               types.String `tfsdk:"sub_path"`
	SubDomain             types.String `tfsdk:"sub_domain"`
	SubCertFile           types.String `tfsdk:"sub_cert_file"`
	SubKeyFile            types.String `tfsdk:"sub_key_file"`
	SubUpdates            types.Int64  `tfsdk:"sub_updates"`
	SubEncrypt            types.Bool   `tfsdk:"sub_encrypt"`
	SubShowInfo           types.Bool   `tfsdk:"sub_show_info"`
	SubEmailInRemark      types.Bool   `tfsdk:"sub_email_in_remark"`
	SubURI                types.String `tfsdk:"sub_uri"`
	SubJsonPath           types.String `tfsdk:"sub_json_path"`
	SubJsonURI            types.String `tfsdk:"sub_json_uri"`
	SubJsonFragment       types.String `tfsdk:"sub_json_fragment"`
	SubJsonNoises         types.String `tfsdk:"sub_json_noises"`
	SubJsonMux            types.String `tfsdk:"sub_json_mux"`
	SubJsonRules          types.String `tfsdk:"sub_json_rules"`
	SubClashEnable        types.Bool   `tfsdk:"sub_clash_enable"`
	SubClashPath          types.String `tfsdk:"sub_clash_path"`
	SubClashURI           types.String `tfsdk:"sub_clash_uri"`
	SubClashEnableRouting types.Bool   `tfsdk:"sub_clash_enable_routing"`
	SubClashRules         types.String `tfsdk:"sub_clash_rules"`
	SubJsonFinalMask      types.String `tfsdk:"sub_json_final_mask"`
}

func panelSubscriptionSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages subscription settings in the 3x-ui panel.\n\n" +
			"~> **Note:** When `sub_port` (default `2096`) differs from the main panel port and the panel " +
			"runs behind a reverse proxy, the proxy must be configured to forward subscription path requests " +
			"to the subscription port. Without this, subscription URLs will return 404. " +
			"For example, in Caddy: `handle /sub/* { reverse_proxy 3x-ui:2096 }`. " +
			"In Nginx: `location /sub/ { proxy_pass http://3x-ui:2096; }`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sub_enable": schema.BoolAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"sub_json_enable": schema.BoolAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"sub_title": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"sub_support_url": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"sub_profile_url": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"sub_announce": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"sub_enable_routing": schema.BoolAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"sub_routing_rules": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"sub_listen": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"sub_port": schema.Int64Attribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"sub_path": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"sub_domain": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"sub_cert_file": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"sub_key_file": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"sub_updates": schema.Int64Attribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"sub_encrypt": schema.BoolAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"sub_show_info": schema.BoolAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"sub_email_in_remark": schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Include the client email in subscription profile names.",
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"sub_uri": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"sub_json_path": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"sub_json_uri": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"sub_json_fragment": schema.StringAttribute{
				Optional: true, Computed: true,
				Description: "JSON fragment settings for subscription. " +
					"**v2.9.2+:** only the fragment parameters object, e.g. " +
					"`{\"packets\":\"tlshello\",\"length\":\"100-200\",\"interval\":\"10-20\"}`. " +
					"**v2.9.1 and earlier:** full outbound object with tag, protocol, settings and streamSettings. " +
					"Deprecated in 3x-ui v3.2.8 — replaced by sub_clash_enable_routing.",
				PlanModifiers:      []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				DeprecationMessage: "Deprecated in 3x-ui v3.2.8. Use sub_clash_enable_routing instead.",
			},
			"sub_json_noises": schema.StringAttribute{
				Optional: true, Computed: true,
				Description: "JSON noise settings for subscription. " +
					"**v2.9.2+:** only the noises array, e.g. " +
					"`[{\"type\":\"rand\",\"packet\":\"10-20\",\"delay\":\"10-16\"}]`. " +
					"**v2.9.1 and earlier:** full outbound object with tag, protocol, settings and streamSettings. " +
					"Deprecated in 3x-ui v3.2.8 — replaced by sub_clash_rules.",
				PlanModifiers:      []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				DeprecationMessage: "Deprecated in 3x-ui v3.2.8. Use sub_clash_rules instead.",
			},
			"sub_json_mux": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"sub_json_rules": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"sub_clash_enable": schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Enable Clash/Mihomo subscription endpoint.",
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"sub_clash_path": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Path for Clash/Mihomo subscription endpoint.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"sub_clash_uri": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Clash/Mihomo subscription server URI.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"sub_clash_enable_routing": schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Enable global routing rules for Clash/Mihomo subscriptions (3x-ui v3.2.8+).",
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"sub_clash_rules": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Clash/Mihomo global routing rules (3x-ui v3.2.8+).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"sub_json_final_mask": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "JSON subscription global finalmask — tcp/udp masks and quicParams (3x-ui v3.2.8+).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func expandPanelSubscription(m *PanelSubscriptionModel) map[string]any {
	payload := map[string]any{}
	if !m.SubEnable.IsNull() && !m.SubEnable.IsUnknown() {
		payload["subEnable"] = m.SubEnable.ValueBool()
	}
	if !m.SubJsonEnable.IsNull() && !m.SubJsonEnable.IsUnknown() {
		payload["subJsonEnable"] = m.SubJsonEnable.ValueBool()
	}
	if !m.SubTitle.IsNull() && !m.SubTitle.IsUnknown() {
		payload["subTitle"] = m.SubTitle.ValueString()
	}
	if !m.SubSupportURL.IsNull() && !m.SubSupportURL.IsUnknown() {
		payload["subSupportUrl"] = m.SubSupportURL.ValueString()
	}
	if !m.SubProfileURL.IsNull() && !m.SubProfileURL.IsUnknown() {
		payload["subProfileUrl"] = m.SubProfileURL.ValueString()
	}
	if !m.SubAnnounce.IsNull() && !m.SubAnnounce.IsUnknown() {
		payload["subAnnounce"] = m.SubAnnounce.ValueString()
	}
	if !m.SubEnableRouting.IsNull() && !m.SubEnableRouting.IsUnknown() {
		payload["subEnableRouting"] = m.SubEnableRouting.ValueBool()
	}
	if !m.SubRoutingRules.IsNull() && !m.SubRoutingRules.IsUnknown() {
		payload["subRoutingRules"] = m.SubRoutingRules.ValueString()
	}
	if !m.SubListen.IsNull() && !m.SubListen.IsUnknown() {
		payload["subListen"] = m.SubListen.ValueString()
	}
	if !m.SubPort.IsNull() && !m.SubPort.IsUnknown() {
		payload["subPort"] = int(m.SubPort.ValueInt64())
	}
	if !m.SubPath.IsNull() && !m.SubPath.IsUnknown() {
		payload["subPath"] = m.SubPath.ValueString()
	}
	if !m.SubDomain.IsNull() && !m.SubDomain.IsUnknown() {
		payload["subDomain"] = m.SubDomain.ValueString()
	}
	if !m.SubCertFile.IsNull() && !m.SubCertFile.IsUnknown() {
		payload["subCertFile"] = m.SubCertFile.ValueString()
	}
	if !m.SubKeyFile.IsNull() && !m.SubKeyFile.IsUnknown() {
		payload["subKeyFile"] = m.SubKeyFile.ValueString()
	}
	if !m.SubUpdates.IsNull() && !m.SubUpdates.IsUnknown() {
		payload["subUpdates"] = int(m.SubUpdates.ValueInt64())
	}
	if !m.SubEncrypt.IsNull() && !m.SubEncrypt.IsUnknown() {
		payload["subEncrypt"] = m.SubEncrypt.ValueBool()
	}
	if !m.SubShowInfo.IsNull() && !m.SubShowInfo.IsUnknown() {
		payload["subShowInfo"] = m.SubShowInfo.ValueBool()
	}
	if !m.SubEmailInRemark.IsNull() && !m.SubEmailInRemark.IsUnknown() {
		payload["subEmailInRemark"] = m.SubEmailInRemark.ValueBool()
	}
	if !m.SubURI.IsNull() && !m.SubURI.IsUnknown() {
		payload["subURI"] = m.SubURI.ValueString()
	}
	if !m.SubJsonPath.IsNull() && !m.SubJsonPath.IsUnknown() {
		payload["subJsonPath"] = m.SubJsonPath.ValueString()
	}
	if !m.SubJsonURI.IsNull() && !m.SubJsonURI.IsUnknown() {
		payload["subJsonURI"] = m.SubJsonURI.ValueString()
	}
	if !m.SubJsonFragment.IsNull() && !m.SubJsonFragment.IsUnknown() {
		payload["subJsonFragment"] = m.SubJsonFragment.ValueString()
	}
	if !m.SubJsonNoises.IsNull() && !m.SubJsonNoises.IsUnknown() {
		payload["subJsonNoises"] = m.SubJsonNoises.ValueString()
	}
	if !m.SubJsonMux.IsNull() && !m.SubJsonMux.IsUnknown() {
		payload["subJsonMux"] = m.SubJsonMux.ValueString()
	}
	if !m.SubJsonRules.IsNull() && !m.SubJsonRules.IsUnknown() {
		payload["subJsonRules"] = m.SubJsonRules.ValueString()
	}
	if !m.SubClashEnable.IsNull() && !m.SubClashEnable.IsUnknown() {
		payload["subClashEnable"] = m.SubClashEnable.ValueBool()
	}
	if !m.SubClashPath.IsNull() && !m.SubClashPath.IsUnknown() {
		payload["subClashPath"] = m.SubClashPath.ValueString()
	}
	if !m.SubClashURI.IsNull() && !m.SubClashURI.IsUnknown() {
		payload["subClashURI"] = m.SubClashURI.ValueString()
	}
	if !m.SubClashEnableRouting.IsNull() && !m.SubClashEnableRouting.IsUnknown() {
		payload["subClashEnableRouting"] = m.SubClashEnableRouting.ValueBool()
	}
	if !m.SubClashRules.IsNull() && !m.SubClashRules.IsUnknown() {
		payload["subClashRules"] = m.SubClashRules.ValueString()
	}
	if !m.SubJsonFinalMask.IsNull() && !m.SubJsonFinalMask.IsUnknown() {
		payload["subJsonFinalMask"] = m.SubJsonFinalMask.ValueString()
	}
	return payload
}

func flattenPanelSubscription(in map[string]any) *PanelSubscriptionModel {
	m := &PanelSubscriptionModel{
		ID: types.StringValue("settings"),
	}
	if v, ok := in["subEnable"]; ok {
		m.SubEnable = types.BoolValue(boolValue(v))
	}
	if v, ok := in["subJsonEnable"]; ok {
		m.SubJsonEnable = types.BoolValue(boolValue(v))
	}
	if v, ok := in["subTitle"]; ok {
		m.SubTitle = types.StringValue(stringValue(v))
	}
	if v, ok := in["subSupportUrl"]; ok {
		m.SubSupportURL = types.StringValue(stringValue(v))
	}
	if v, ok := in["subProfileUrl"]; ok {
		m.SubProfileURL = types.StringValue(stringValue(v))
	}
	if v, ok := in["subAnnounce"]; ok {
		m.SubAnnounce = types.StringValue(stringValue(v))
	}
	if v, ok := in["subEnableRouting"]; ok {
		m.SubEnableRouting = types.BoolValue(boolValue(v))
	}
	if v, ok := in["subRoutingRules"]; ok {
		m.SubRoutingRules = types.StringValue(stringValue(v))
	}
	if v, ok := in["subListen"]; ok {
		m.SubListen = types.StringValue(stringValue(v))
	}
	if v, ok := in["subPort"]; ok {
		m.SubPort = types.Int64Value(int64(intValue(v)))
	}
	if v, ok := in["subPath"]; ok {
		m.SubPath = types.StringValue(stringValue(v))
	}
	if v, ok := in["subDomain"]; ok {
		m.SubDomain = types.StringValue(stringValue(v))
	}
	if v, ok := in["subCertFile"]; ok {
		m.SubCertFile = types.StringValue(stringValue(v))
	}
	if v, ok := in["subKeyFile"]; ok {
		m.SubKeyFile = types.StringValue(stringValue(v))
	}
	if v, ok := in["subUpdates"]; ok {
		m.SubUpdates = types.Int64Value(int64(intValue(v)))
	}
	if v, ok := in["subEncrypt"]; ok {
		m.SubEncrypt = types.BoolValue(boolValue(v))
	}
	if v, ok := in["subShowInfo"]; ok {
		m.SubShowInfo = types.BoolValue(boolValue(v))
	}
	if v, ok := in["subEmailInRemark"]; ok {
		m.SubEmailInRemark = types.BoolValue(boolValue(v))
	}
	if v, ok := in["subURI"]; ok {
		m.SubURI = types.StringValue(stringValue(v))
	}
	if v, ok := in["subJsonPath"]; ok {
		m.SubJsonPath = types.StringValue(stringValue(v))
	}
	if v, ok := in["subJsonURI"]; ok {
		m.SubJsonURI = types.StringValue(stringValue(v))
	}
	if v, ok := in["subJsonFragment"]; ok {
		m.SubJsonFragment = types.StringValue(stringValue(v))
	} else {
		m.SubJsonFragment = types.StringValue("")
	}
	if v, ok := in["subJsonNoises"]; ok {
		m.SubJsonNoises = types.StringValue(stringValue(v))
	} else {
		m.SubJsonNoises = types.StringValue("")
	}
	if v, ok := in["subJsonMux"]; ok {
		m.SubJsonMux = types.StringValue(stringValue(v))
	}
	if v, ok := in["subJsonRules"]; ok {
		m.SubJsonRules = types.StringValue(stringValue(v))
	}
	if v, ok := in["subClashEnable"]; ok {
		m.SubClashEnable = types.BoolValue(boolValue(v))
	}
	if v, ok := in["subClashPath"]; ok {
		m.SubClashPath = types.StringValue(stringValue(v))
	} else {
		m.SubClashPath = types.StringNull()
	}
	if v, ok := in["subClashURI"]; ok {
		m.SubClashURI = types.StringValue(stringValue(v))
	} else {
		m.SubClashURI = types.StringNull()
	}
	if v, ok := in["subClashEnableRouting"]; ok {
		m.SubClashEnableRouting = types.BoolValue(boolValue(v))
	}
	if v, ok := in["subClashRules"]; ok {
		m.SubClashRules = types.StringValue(stringValue(v))
	}
	if v, ok := in["subJsonFinalMask"]; ok {
		m.SubJsonFinalMask = types.StringValue(stringValue(v))
	}
	return m
}

// modifyPlanWOVersion marks the plain secret attribute as Unknown during plan
// when the *_wo_version trigger changes (or is set for the first time). This
// tells Terraform to accept a new sensitive value from Apply instead of
// rejecting it as "inconsistent values for sensitive" — the prior-state value
// is otherwise carried forward by UseStateForUnknown and blocks the update.
//
// On no-op plans (version unchanged, no _wo in config) it does nothing.
func modifyPlanWOVersion[T any](
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
	woVersion func(T) types.Int64,
	setPlain func(*T, types.String),
) {
	if req.Plan.Raw.IsNull() {
		return
	}
	if req.State.Raw.IsNull() {
		return
	}

	var plan, state T
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !woVersionTriggered(woVersion(plan), woVersion(state)) {
		return
	}

	setPlain(&plan, types.StringUnknown())
	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

// woVersionTriggered reports whether the *_wo_version value represents a
// change relative to the prior state: the values differ, or the prior state
// has none while the plan introduces one.
func woVersionTriggered(plan, state types.Int64) bool {
	if plan.IsNull() || plan.IsUnknown() {
		return false
	}
	if state.IsNull() || state.IsUnknown() {
		return true
	}
	return plan.ValueInt64() != state.ValueInt64()
}

// ---------------------------------------------------------------------------
// Panel General model, schema, expand/flatten
// ---------------------------------------------------------------------------

type PanelGeneralModel struct {
	ID                          types.String `tfsdk:"id"`
	WebListen                   types.String `tfsdk:"web_listen"`
	WebDomain                   types.String `tfsdk:"web_domain"`
	WebPort                     types.Int64  `tfsdk:"web_port"`
	WebBasePath                 types.String `tfsdk:"web_base_path"`
	SessionMaxAge               types.Int64  `tfsdk:"session_max_age"`
	TrustedProxyCIDRs           types.String `tfsdk:"trusted_proxy_cidrs"`
	PageSize                    types.Int64  `tfsdk:"page_size"`
	RemarkModel                 types.String `tfsdk:"remark_model"`
	DatePicker                  types.String `tfsdk:"date_picker"`
	TimeLocation                types.String `tfsdk:"time_location"`
	ExpireDiff                  types.Int64  `tfsdk:"expire_diff"`
	TrafficDiff                 types.Int64  `tfsdk:"traffic_diff"`
	WebCertFile                 types.String `tfsdk:"web_cert_file"`
	WebKeyFile                  types.String `tfsdk:"web_key_file"`
	ExternalTrafficInformEnable types.Bool   `tfsdk:"external_traffic_inform_enable"`
	ExternalTrafficInformURI    types.String `tfsdk:"external_traffic_inform_uri"`
	RestartXrayOnClientDisable  types.Bool   `tfsdk:"restart_xray_on_client_disable"`
	LDAPEnable                  types.Bool   `tfsdk:"ldap_enable"`
	LDAPHost                    types.String `tfsdk:"ldap_host"`
	LDAPPort                    types.Int64  `tfsdk:"ldap_port"`
	LDAPUseTLS                  types.Bool   `tfsdk:"ldap_use_tls"`
	LDAPBindDN                  types.String `tfsdk:"ldap_bind_dn"`
	LDAPPassword                types.String `tfsdk:"ldap_password"`
	LDAPPasswordWO              types.String `tfsdk:"ldap_password_wo"`
	LDAPPasswordWOVersion       types.Int64  `tfsdk:"ldap_password_wo_version"`
	LDAPBaseDN                  types.String `tfsdk:"ldap_base_dn"`
	LDAPUserFilter              types.String `tfsdk:"ldap_user_filter"`
	LDAPUserAttr                types.String `tfsdk:"ldap_user_attr"`
	LDAPVlessField              types.String `tfsdk:"ldap_vless_field"`
	LDAPSyncCron                types.String `tfsdk:"ldap_sync_cron"`
	LDAPFlagField               types.String `tfsdk:"ldap_flag_field"`
	LDAPTruthyValues            types.String `tfsdk:"ldap_truthy_values"`
	LDAPInvertFlag              types.Bool   `tfsdk:"ldap_invert_flag"`
	LDAPInboundTags             types.String `tfsdk:"ldap_inbound_tags"`
	LDAPAutoCreate              types.Bool   `tfsdk:"ldap_auto_create"`
	LDAPAutoDelete              types.Bool   `tfsdk:"ldap_auto_delete"`
	LDAPDefaultTotalGB          types.Int64  `tfsdk:"ldap_default_total_gb"`
	LDAPDefaultExpiryDays       types.Int64  `tfsdk:"ldap_default_expiry_days"`
	LDAPDefaultLimitIP          types.Int64  `tfsdk:"ldap_default_limit_ip"`
	XrayOutboundTestURL         types.String `tfsdk:"xray_outbound_test_url"`
	PanelProxy                  types.String `tfsdk:"panel_proxy"`
	PanelOutbound               types.String `tfsdk:"panel_outbound"`
}

func panelGeneralSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"web_listen": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"web_domain": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"web_port": schema.Int64Attribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"web_base_path": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"session_max_age": schema.Int64Attribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"trusted_proxy_cidrs": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "Comma-separated trusted reverse proxy IPs/CIDRs used by 3x-ui when honoring " +
					"X-Forwarded-For, X-Forwarded-Host, and X-Real-IP headers.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"page_size": schema.Int64Attribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"remark_model": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"date_picker": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"time_location": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"expire_diff": schema.Int64Attribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"traffic_diff": schema.Int64Attribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"web_cert_file": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"web_key_file": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"external_traffic_inform_enable": schema.BoolAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"external_traffic_inform_uri": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"restart_xray_on_client_disable": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Restart Xray when clients are automatically disabled by expiry or traffic limit.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"ldap_enable": schema.BoolAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"ldap_host": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ldap_port": schema.Int64Attribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"ldap_use_tls": schema.BoolAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"ldap_bind_dn": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ldap_password": schema.StringAttribute{
				Optional: true, Computed: true, Sensitive: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Validators: []validator.String{
					stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("ldap_password_wo")),
				},
			},
			"ldap_password_wo": schema.StringAttribute{
				Optional:  true,
				WriteOnly: true,
			},
			"ldap_password_wo_version": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AlsoRequires(path.MatchRoot("ldap_password_wo")),
				},
			},
			"ldap_base_dn": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ldap_user_filter": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ldap_user_attr": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ldap_vless_field": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ldap_sync_cron": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ldap_flag_field": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ldap_truthy_values": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ldap_invert_flag": schema.BoolAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"ldap_inbound_tags": schema.StringAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ldap_auto_create": schema.BoolAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"ldap_auto_delete": schema.BoolAttribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"ldap_default_total_gb": schema.Int64Attribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"ldap_default_expiry_days": schema.Int64Attribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"ldap_default_limit_ip": schema.Int64Attribute{
				Optional: true, Computed: true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"xray_outbound_test_url": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "URL used for testing outbound connectivity (default: https://www.google.com/generate_204).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"panel_proxy": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "HTTP/SOCKS5 proxy URL for the panel's own outbound requests (xray version " +
					"checks, Telegram bot, outbound testing). Available on 3x-ui v3.2.0 through v3.3.0; " +
					"superseded by panel_outbound (outbound egress bridge) on v3.3.1+. " +
					"Ignored by v3.3.1+ panels.",
				PlanModifiers:      []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				DeprecationMessage: "Superseded by panel_outbound on 3x-ui v3.3.1+. Use panel_outbound for new configurations.",
			},
			"panel_outbound": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Xray outbound tag (or balancer tag) used for the panel's own outbound HTTP (update checks/downloads, Telegram, geo updates, outbound-subscription fetches). Available on 3x-ui v3.3.1+. Ignored by older panels; use panel_proxy on v3.2.0 through v3.3.0.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func expandPanelGeneral(m *PanelGeneralModel) map[string]any {
	payload := map[string]any{}
	if !m.WebListen.IsNull() && !m.WebListen.IsUnknown() {
		payload["webListen"] = m.WebListen.ValueString()
	}
	if !m.WebDomain.IsNull() && !m.WebDomain.IsUnknown() {
		payload["webDomain"] = m.WebDomain.ValueString()
	}
	if !m.WebPort.IsNull() && !m.WebPort.IsUnknown() {
		payload["webPort"] = int(m.WebPort.ValueInt64())
	}
	if !m.WebBasePath.IsNull() && !m.WebBasePath.IsUnknown() {
		payload["webBasePath"] = m.WebBasePath.ValueString()
	}
	if !m.SessionMaxAge.IsNull() && !m.SessionMaxAge.IsUnknown() {
		payload["sessionMaxAge"] = int(m.SessionMaxAge.ValueInt64())
	}
	if !m.TrustedProxyCIDRs.IsNull() && !m.TrustedProxyCIDRs.IsUnknown() {
		payload["trustedProxyCIDRs"] = m.TrustedProxyCIDRs.ValueString()
	}
	if !m.PageSize.IsNull() && !m.PageSize.IsUnknown() {
		payload["pageSize"] = int(m.PageSize.ValueInt64())
	}
	if !m.RemarkModel.IsNull() && !m.RemarkModel.IsUnknown() {
		payload["remarkModel"] = m.RemarkModel.ValueString()
	}
	if !m.DatePicker.IsNull() && !m.DatePicker.IsUnknown() {
		payload["datepicker"] = m.DatePicker.ValueString()
	}
	if !m.TimeLocation.IsNull() && !m.TimeLocation.IsUnknown() {
		payload["timeLocation"] = m.TimeLocation.ValueString()
	}
	if !m.ExpireDiff.IsNull() && !m.ExpireDiff.IsUnknown() {
		payload["expireDiff"] = int(m.ExpireDiff.ValueInt64())
	}
	if !m.TrafficDiff.IsNull() && !m.TrafficDiff.IsUnknown() {
		payload["trafficDiff"] = int(m.TrafficDiff.ValueInt64())
	}
	if !m.WebCertFile.IsNull() && !m.WebCertFile.IsUnknown() {
		payload["webCertFile"] = m.WebCertFile.ValueString()
	}
	if !m.WebKeyFile.IsNull() && !m.WebKeyFile.IsUnknown() {
		payload["webKeyFile"] = m.WebKeyFile.ValueString()
	}
	if !m.ExternalTrafficInformEnable.IsNull() && !m.ExternalTrafficInformEnable.IsUnknown() {
		payload["externalTrafficInformEnable"] = m.ExternalTrafficInformEnable.ValueBool()
	}
	if !m.ExternalTrafficInformURI.IsNull() && !m.ExternalTrafficInformURI.IsUnknown() {
		payload["externalTrafficInformURI"] = m.ExternalTrafficInformURI.ValueString()
	}
	if !m.RestartXrayOnClientDisable.IsNull() && !m.RestartXrayOnClientDisable.IsUnknown() {
		payload["restartXrayOnClientDisable"] = m.RestartXrayOnClientDisable.ValueBool()
	}
	if !m.LDAPEnable.IsNull() && !m.LDAPEnable.IsUnknown() {
		payload["ldapEnable"] = m.LDAPEnable.ValueBool()
	}
	if !m.LDAPHost.IsNull() && !m.LDAPHost.IsUnknown() {
		payload["ldapHost"] = m.LDAPHost.ValueString()
	}
	if !m.LDAPPort.IsNull() && !m.LDAPPort.IsUnknown() {
		payload["ldapPort"] = int(m.LDAPPort.ValueInt64())
	}
	if !m.LDAPUseTLS.IsNull() && !m.LDAPUseTLS.IsUnknown() {
		payload["ldapUseTLS"] = m.LDAPUseTLS.ValueBool()
	}
	if !m.LDAPBindDN.IsNull() && !m.LDAPBindDN.IsUnknown() {
		payload["ldapBindDN"] = m.LDAPBindDN.ValueString()
	}
	if !m.LDAPPasswordWO.IsNull() && !m.LDAPPasswordWO.IsUnknown() {
		payload["ldapPassword"] = m.LDAPPasswordWO.ValueString()
	} else if !m.LDAPPassword.IsNull() && !m.LDAPPassword.IsUnknown() {
		payload["ldapPassword"] = m.LDAPPassword.ValueString()
	}
	if !m.LDAPBaseDN.IsNull() && !m.LDAPBaseDN.IsUnknown() {
		payload["ldapBaseDN"] = m.LDAPBaseDN.ValueString()
	}
	if !m.LDAPUserFilter.IsNull() && !m.LDAPUserFilter.IsUnknown() {
		payload["ldapUserFilter"] = m.LDAPUserFilter.ValueString()
	}
	if !m.LDAPUserAttr.IsNull() && !m.LDAPUserAttr.IsUnknown() {
		payload["ldapUserAttr"] = m.LDAPUserAttr.ValueString()
	}
	if !m.LDAPVlessField.IsNull() && !m.LDAPVlessField.IsUnknown() {
		payload["ldapVlessField"] = m.LDAPVlessField.ValueString()
	}
	if !m.LDAPSyncCron.IsNull() && !m.LDAPSyncCron.IsUnknown() {
		payload["ldapSyncCron"] = m.LDAPSyncCron.ValueString()
	}
	if !m.LDAPFlagField.IsNull() && !m.LDAPFlagField.IsUnknown() {
		payload["ldapFlagField"] = m.LDAPFlagField.ValueString()
	}
	if !m.LDAPTruthyValues.IsNull() && !m.LDAPTruthyValues.IsUnknown() {
		payload["ldapTruthyValues"] = m.LDAPTruthyValues.ValueString()
	}
	if !m.LDAPInvertFlag.IsNull() && !m.LDAPInvertFlag.IsUnknown() {
		payload["ldapInvertFlag"] = m.LDAPInvertFlag.ValueBool()
	}
	if !m.LDAPInboundTags.IsNull() && !m.LDAPInboundTags.IsUnknown() {
		payload["ldapInboundTags"] = m.LDAPInboundTags.ValueString()
	}
	if !m.LDAPAutoCreate.IsNull() && !m.LDAPAutoCreate.IsUnknown() {
		payload["ldapAutoCreate"] = m.LDAPAutoCreate.ValueBool()
	}
	if !m.LDAPAutoDelete.IsNull() && !m.LDAPAutoDelete.IsUnknown() {
		payload["ldapAutoDelete"] = m.LDAPAutoDelete.ValueBool()
	}
	if !m.LDAPDefaultTotalGB.IsNull() && !m.LDAPDefaultTotalGB.IsUnknown() {
		payload["ldapDefaultTotalGB"] = int(m.LDAPDefaultTotalGB.ValueInt64())
	}
	if !m.LDAPDefaultExpiryDays.IsNull() && !m.LDAPDefaultExpiryDays.IsUnknown() {
		payload["ldapDefaultExpiryDays"] = int(m.LDAPDefaultExpiryDays.ValueInt64())
	}
	if !m.LDAPDefaultLimitIP.IsNull() && !m.LDAPDefaultLimitIP.IsUnknown() {
		payload["ldapDefaultLimitIP"] = int(m.LDAPDefaultLimitIP.ValueInt64())
	}
	if !m.PanelProxy.IsNull() && !m.PanelProxy.IsUnknown() {
		payload["panelProxy"] = m.PanelProxy.ValueString()
	}
	if !m.PanelOutbound.IsNull() && !m.PanelOutbound.IsUnknown() {
		payload["panelOutbound"] = m.PanelOutbound.ValueString()
	}
	return payload
}

func flattenPanelGeneral(in map[string]any) *PanelGeneralModel {
	m := &PanelGeneralModel{
		ID: types.StringValue("settings"),
	}
	if v, ok := in["webListen"]; ok {
		m.WebListen = types.StringValue(stringValue(v))
	}
	if v, ok := in["webDomain"]; ok {
		m.WebDomain = types.StringValue(stringValue(v))
	}
	if v, ok := in["webPort"]; ok {
		m.WebPort = types.Int64Value(int64(intValue(v)))
	}
	if v, ok := in["webBasePath"]; ok {
		m.WebBasePath = types.StringValue(stringValue(v))
	}
	if v, ok := in["sessionMaxAge"]; ok {
		m.SessionMaxAge = types.Int64Value(int64(intValue(v)))
	}
	if v, ok := in["trustedProxyCIDRs"]; ok {
		m.TrustedProxyCIDRs = types.StringValue(stringValue(v))
	}
	if v, ok := in["pageSize"]; ok {
		m.PageSize = types.Int64Value(int64(intValue(v)))
	}
	if v, ok := in["remarkModel"]; ok {
		m.RemarkModel = types.StringValue(stringValue(v))
	}
	if v, ok := in["datepicker"]; ok {
		m.DatePicker = types.StringValue(stringValue(v))
	}
	if v, ok := in["timeLocation"]; ok {
		m.TimeLocation = types.StringValue(stringValue(v))
	}
	if v, ok := in["expireDiff"]; ok {
		m.ExpireDiff = types.Int64Value(int64(intValue(v)))
	}
	if v, ok := in["trafficDiff"]; ok {
		m.TrafficDiff = types.Int64Value(int64(intValue(v)))
	}
	if v, ok := in["webCertFile"]; ok {
		m.WebCertFile = types.StringValue(stringValue(v))
	}
	if v, ok := in["webKeyFile"]; ok {
		m.WebKeyFile = types.StringValue(stringValue(v))
	}
	if v, ok := in["externalTrafficInformEnable"]; ok {
		m.ExternalTrafficInformEnable = types.BoolValue(boolValue(v))
	}
	if v, ok := in["externalTrafficInformURI"]; ok {
		m.ExternalTrafficInformURI = types.StringValue(stringValue(v))
	}
	if v, ok := in["restartXrayOnClientDisable"]; ok {
		m.RestartXrayOnClientDisable = types.BoolValue(boolValue(v))
	}
	if v, ok := in["ldapEnable"]; ok {
		m.LDAPEnable = types.BoolValue(boolValue(v))
	}
	if v, ok := in["ldapHost"]; ok {
		m.LDAPHost = types.StringValue(stringValue(v))
	}
	if v, ok := in["ldapPort"]; ok {
		m.LDAPPort = types.Int64Value(int64(intValue(v)))
	}
	if v, ok := in["ldapUseTLS"]; ok {
		m.LDAPUseTLS = types.BoolValue(boolValue(v))
	}
	if v, ok := in["ldapBindDN"]; ok {
		m.LDAPBindDN = types.StringValue(stringValue(v))
	}
	if v, ok := in["ldapPassword"]; ok {
		m.LDAPPassword = types.StringValue(stringValue(v))
	}
	if v, ok := in["ldapBaseDN"]; ok {
		m.LDAPBaseDN = types.StringValue(stringValue(v))
	}
	if v, ok := in["ldapUserFilter"]; ok {
		m.LDAPUserFilter = types.StringValue(stringValue(v))
	}
	if v, ok := in["ldapUserAttr"]; ok {
		m.LDAPUserAttr = types.StringValue(stringValue(v))
	}
	if v, ok := in["ldapVlessField"]; ok {
		m.LDAPVlessField = types.StringValue(stringValue(v))
	}
	if v, ok := in["ldapSyncCron"]; ok {
		m.LDAPSyncCron = types.StringValue(stringValue(v))
	}
	if v, ok := in["ldapFlagField"]; ok {
		m.LDAPFlagField = types.StringValue(stringValue(v))
	}
	if v, ok := in["ldapTruthyValues"]; ok {
		m.LDAPTruthyValues = types.StringValue(stringValue(v))
	}
	if v, ok := in["ldapInvertFlag"]; ok {
		m.LDAPInvertFlag = types.BoolValue(boolValue(v))
	}
	if v, ok := in["ldapInboundTags"]; ok {
		m.LDAPInboundTags = types.StringValue(stringValue(v))
	}
	if v, ok := in["ldapAutoCreate"]; ok {
		m.LDAPAutoCreate = types.BoolValue(boolValue(v))
	}
	if v, ok := in["ldapAutoDelete"]; ok {
		m.LDAPAutoDelete = types.BoolValue(boolValue(v))
	}
	if v, ok := in["ldapDefaultTotalGB"]; ok {
		m.LDAPDefaultTotalGB = types.Int64Value(int64(intValue(v)))
	}
	if v, ok := in["ldapDefaultExpiryDays"]; ok {
		m.LDAPDefaultExpiryDays = types.Int64Value(int64(intValue(v)))
	}
	if v, ok := in["ldapDefaultLimitIP"]; ok {
		m.LDAPDefaultLimitIP = types.Int64Value(int64(intValue(v)))
	}
	if v, ok := in["panelProxy"]; ok {
		m.PanelProxy = types.StringValue(stringValue(v))
	}
	if v, ok := in["panelOutbound"]; ok {
		m.PanelOutbound = types.StringValue(stringValue(v))
	}
	return m
}

// ---------------------------------------------------------------------------
// Shared typed settings helper: apply settings to API and read back
// ---------------------------------------------------------------------------

var settingsMu sync.Mutex

var panelSettingSecretKeys = []string{
	"ldapPassword",
	"twoFactorToken",
	"tgBotToken",
}

func settingsApplyTyped(
	ctx context.Context,
	desired map[string]any,
	diags *diag.Diagnostics,
	client *Client,
) {
	if len(desired) == 0 {
		return
	}

	settingsMu.Lock()
	defer settingsMu.Unlock()

	existing, err := client.GetSettings(ctx)
	if err != nil {
		diags.AddError("Failed to get settings", err.Error())
		return
	}

	merged := mergeSettingsForUpdate(client, existing, desired)
	if err := client.UpdateSettings(ctx, merged); err != nil {
		diags.AddError("Failed to update settings", err.Error())
		return
	}
	client.rememberConfiguredSettingSecrets(desired)
}

func settingsReadTyped(
	ctx context.Context,
	diags *diag.Diagnostics,
	client *Client,
) map[string]any {
	settings, err := client.GetSettings(ctx)
	if err != nil {
		diags.AddError("Failed to get settings", err.Error())
		return nil
	}
	return settings
}

func mergeSettingsForUpdate(client *Client, existing, desired map[string]any) map[string]any {
	if client != nil {
		existing = client.preserveCachedSettingSecrets(existing, desired)
	}
	return mergeSettings(existing, desired)
}

func (c *Client) rememberConfiguredSettingSecrets(settings map[string]any) {
	if c == nil || len(settings) == 0 {
		return
	}

	c.settingsSecretMu.Lock()
	defer c.settingsSecretMu.Unlock()

	for _, key := range panelSettingSecretKeys {
		value, ok := settings[key]
		if !ok {
			continue
		}

		secret := stringValue(value)
		if secret == "" || isRedactedSettingSecretValue(secret) {
			delete(c.settingsSecrets, key)
			continue
		}

		if c.settingsSecrets == nil {
			c.settingsSecrets = make(map[string]string)
		}
		c.settingsSecrets[key] = secret
	}
}

func (c *Client) preserveCachedSettingSecrets(existing, desired map[string]any) map[string]any {
	if c == nil || len(existing) == 0 {
		return existing
	}

	c.settingsSecretMu.Lock()
	defer c.settingsSecretMu.Unlock()

	if len(c.settingsSecrets) == 0 {
		return existing
	}

	out := existing
	copied := false
	for _, key := range panelSettingSecretKeys {
		if _, configured := desired[key]; configured {
			continue
		}

		cached := c.settingsSecrets[key]
		if cached == "" || !isRedactedSettingSecret(existing[key]) {
			continue
		}

		if !copied {
			out = make(map[string]any, len(existing))
			for k, v := range existing {
				out[k] = v
			}
			copied = true
		}
		out[key] = cached
	}
	return out
}

func isRedactedSettingSecret(value any) bool {
	secret, ok := value.(string)
	if !ok {
		return value == nil
	}
	return isRedactedSettingSecretValue(secret)
}

func isRedactedSettingSecretValue(secret string) bool {
	trimmed := strings.TrimSpace(secret)
	if trimmed == "" {
		return true
	}
	if strings.EqualFold(trimmed, "redacted") || strings.EqualFold(trimmed, "<redacted>") {
		return true
	}
	if len(trimmed) >= 3 && strings.Trim(trimmed, "*") == "" {
		return true
	}
	return false
}

func preserveSettingSecret(observed, configured types.String) types.String {
	if configured.IsNull() || configured.IsUnknown() {
		return observed
	}

	configuredValue := configured.ValueString()
	if observed.IsNull() || observed.IsUnknown() {
		return configured
	}

	observedValue := observed.ValueString()
	if configuredValue == "" && observedValue != "" {
		return configured
	}
	if configuredValue != "" && isRedactedSettingSecretValue(observedValue) {
		return configured
	}

	return observed
}

// preserveWOVersion carries the *_wo_version trigger into state during Apply.
// UseStateForUnknown on the schema handles the Plan phase; this handles Apply,
// where flatten* (which has no _wo_version field) would otherwise overwrite
// the planned value with null.
func preserveWOVersion(observed, configured types.Int64) types.Int64 {
	if observed.IsNull() || observed.IsUnknown() {
		return configured
	}
	return observed
}

func preservePanelGeneralSecrets(state, configured *PanelGeneralModel) {
	if state == nil || configured == nil {
		return
	}
	state.LDAPPassword = preserveSettingSecret(state.LDAPPassword, configured.LDAPPassword)
	state.LDAPPasswordWOVersion = preserveWOVersion(state.LDAPPasswordWOVersion, configured.LDAPPasswordWOVersion)
}

func preservePanelSecuritySecrets(state, configured *PanelSecurityModel) {
	if state == nil || configured == nil {
		return
	}
	state.TwoFactorToken = preserveSettingSecret(state.TwoFactorToken, configured.TwoFactorToken)
	state.TwoFactorTokenWOVersion = preserveWOVersion(state.TwoFactorTokenWOVersion, configured.TwoFactorTokenWOVersion)
}

func preservePanelTelegramSecrets(state, configured *PanelTelegramModel) {
	if state == nil || configured == nil {
		return
	}
	state.TgBotToken = preserveSettingSecret(state.TgBotToken, configured.TgBotToken)
	state.TgBotTokenWOVersion = preserveWOVersion(state.TgBotTokenWOVersion, configured.TgBotTokenWOVersion)
}

// ---------------------------------------------------------------------------
// PanelGeneralResource (threexui_panel_general)
// ---------------------------------------------------------------------------

var (
	_ resource.Resource                = &PanelGeneralResource{}
	_ resource.ResourceWithConfigure   = &PanelGeneralResource{}
	_ resource.ResourceWithImportState = &PanelGeneralResource{}
	_ resource.ResourceWithModifyPlan  = &PanelGeneralResource{}
)

type PanelGeneralResource struct {
	client *Client
}

func NewPanelGeneralResource() resource.Resource {
	return &PanelGeneralResource{}
}

func (r *PanelGeneralResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_panel_general"
}

func (r *PanelGeneralResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = panelGeneralSchema()
}

func (r *PanelGeneralResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *Client")
		return
	}
	r.client = client
}

func (r *PanelGeneralResource) readPanelGeneralState(ctx context.Context, diags *diag.Diagnostics) *PanelGeneralModel {
	settings := settingsReadTyped(ctx, diags, r.client)
	if settings == nil {
		return nil
	}
	state := flattenPanelGeneral(settings)

	// xrayOutboundTestUrl is served via the xray endpoint, not the settings API.
	testURL, err := r.client.GetXrayOutboundTestURL(ctx)
	if err != nil {
		diags.AddError("Failed to get xray outbound test URL", err.Error())
		return nil
	}
	if testURL != "" {
		state.XrayOutboundTestURL = types.StringValue(testURL)
	}
	return state
}

func (r *PanelGeneralResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PanelGeneralModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config PanelGeneralModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resolveGeneralLDAPPasswordWO(&plan, config)

	r.applyPanelGeneral(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := r.readPanelGeneralState(ctx, &resp.Diagnostics)
	if state == nil {
		return
	}
	preservePanelGeneralSecrets(state, &plan)
	r.client.rememberConfiguredSettingSecrets(expandPanelGeneral(state))
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *PanelGeneralResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var prior PanelGeneralModel
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := r.readPanelGeneralState(ctx, &resp.Diagnostics)
	if state == nil {
		return
	}
	preservePanelGeneralSecrets(state, &prior)
	r.client.rememberConfiguredSettingSecrets(expandPanelGeneral(state))
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *PanelGeneralResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PanelGeneralModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var priorState PanelGeneralModel
	resp.Diagnostics.Append(req.State.Get(ctx, &priorState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config PanelGeneralModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resolveGeneralLDAPPasswordWOUpdate(&plan, priorState, config)

	r.applyPanelGeneral(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := r.readPanelGeneralState(ctx, &resp.Diagnostics)
	if state == nil {
		return
	}
	preservePanelGeneralSecrets(state, &plan)
	r.client.rememberConfiguredSettingSecrets(expandPanelGeneral(state))
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *PanelGeneralResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *PanelGeneralResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifyPlanWOVersion(
		ctx, req, resp,
		func(m PanelGeneralModel) types.Int64 { return m.LDAPPasswordWOVersion },
		func(m *PanelGeneralModel, v types.String) { m.LDAPPassword = v },
	)
}

func (r *PanelGeneralResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	state := r.readPanelGeneralState(ctx, &resp.Diagnostics)
	if state == nil {
		return
	}
	r.client.rememberConfiguredSettingSecrets(expandPanelGeneral(state))
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func resolveGeneralLDAPPasswordWO(plan *PanelGeneralModel, config PanelGeneralModel) {
	if !config.LDAPPasswordWO.IsNull() {
		plan.LDAPPassword = config.LDAPPasswordWO
	}
}

func resolveGeneralLDAPPasswordWOUpdate(plan *PanelGeneralModel, state PanelGeneralModel, config PanelGeneralModel) {
	if config.LDAPPasswordWO.IsNull() {
		return
	}
	if woVersionTriggered(plan.LDAPPasswordWOVersion, state.LDAPPasswordWOVersion) {
		plan.LDAPPassword = config.LDAPPasswordWO
	}
}

func (r *PanelGeneralResource) applyPanelGeneral(ctx context.Context, plan *PanelGeneralModel, diags *diag.Diagnostics) {
	desired := expandPanelGeneral(plan)

	// Warn about web_base_path change.
	if _, hasBasePath := desired["webBasePath"]; hasBasePath {
		diags.AddWarning(
			"Changing web_base_path requires updating provider config",
			"The provider's base_path must match the panel's web_base_path. After this change, update the provider configuration to use the new base_path, otherwise the provider will not be able to connect.",
		)
	}

	if len(desired) > 0 {
		settingsMu.Lock()
		existing, err := r.client.GetSettings(ctx)
		if err != nil {
			settingsMu.Unlock()
			diags.AddError("Failed to get settings", err.Error())
			return
		}

		needRestart := panelSettingsNeedRestart(existing, desired)
		merged := mergeSettingsForUpdate(r.client, existing, desired)
		if err := r.client.UpdateSettings(ctx, merged); err != nil {
			settingsMu.Unlock()
			diags.AddError("Failed to update settings", err.Error())
			return
		}
		r.client.rememberConfiguredSettingSecrets(desired)
		settingsMu.Unlock()

		if needRestart {
			// Send the restart request while basePath still points to the
			// old path (where the panel is currently listening).
			if err := r.client.SendRestart(ctx); err != nil {
				diags.AddError("Failed to restart panel", err.Error())
				return
			}

			// Now update basePath so waitForReady polls the new path.
			if newPath, ok := desired["webBasePath"]; ok {
				r.client.SetBasePath(stringValue(newPath))
			}

			if err := r.client.WaitForReady(ctx); err != nil {
				diags.AddError("Panel did not become ready after restart", err.Error())
				return
			}
		} else if newPath, ok := desired["webBasePath"]; ok {
			// No restart needed, but basePath still needs updating.
			r.client.SetBasePath(stringValue(newPath))
		}
	}

	// xrayOutboundTestUrl is managed via xray endpoint, not settings API.
	if !plan.XrayOutboundTestURL.IsNull() && !plan.XrayOutboundTestURL.IsUnknown() {
		xrayTemplateMu.Lock()
		if err := r.client.SetXrayOutboundTestURL(ctx, plan.XrayOutboundTestURL.ValueString()); err != nil {
			xrayTemplateMu.Unlock()
			diags.AddError("Failed to set xray outbound test URL", err.Error())
			return
		}
		xrayTemplateMu.Unlock()
	}
}

// ---------------------------------------------------------------------------
// PanelSecurityResource (threexui_panel_security)
// ---------------------------------------------------------------------------

var (
	_ resource.Resource                = &PanelSecurityResource{}
	_ resource.ResourceWithConfigure   = &PanelSecurityResource{}
	_ resource.ResourceWithImportState = &PanelSecurityResource{}
	_ resource.ResourceWithModifyPlan  = &PanelSecurityResource{}
)

type PanelSecurityResource struct {
	client *Client
}

func NewPanelSecurityResource() resource.Resource {
	return &PanelSecurityResource{}
}

func (r *PanelSecurityResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_panel_security"
}

func (r *PanelSecurityResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = panelSecuritySchema()
}

func (r *PanelSecurityResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *Client")
		return
	}
	r.client = client
}

func (r *PanelSecurityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PanelSecurityModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config PanelSecurityModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resolveSecurityTokenWO(&plan, config)

	r.warnIfTwoFactor(&plan, &resp.Diagnostics)

	desired := expandPanelSecurity(&plan)
	settingsApplyTyped(ctx, desired, &resp.Diagnostics, r.client)
	if resp.Diagnostics.HasError() {
		return
	}

	settings := settingsReadTyped(ctx, &resp.Diagnostics, r.client)
	if settings == nil {
		return
	}
	state := flattenPanelSecurity(settings)
	preservePanelSecuritySecrets(state, &plan)
	r.client.rememberConfiguredSettingSecrets(expandPanelSecurity(state))
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *PanelSecurityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var prior PanelSecurityModel
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings := settingsReadTyped(ctx, &resp.Diagnostics, r.client)
	if settings == nil {
		return
	}
	state := flattenPanelSecurity(settings)
	preservePanelSecuritySecrets(state, &prior)
	r.client.rememberConfiguredSettingSecrets(expandPanelSecurity(state))
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *PanelSecurityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PanelSecurityModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var priorState PanelSecurityModel
	resp.Diagnostics.Append(req.State.Get(ctx, &priorState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config PanelSecurityModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resolveSecurityTokenWOUpdate(&plan, priorState, config)

	r.warnIfTwoFactor(&plan, &resp.Diagnostics)

	desired := expandPanelSecurity(&plan)
	settingsApplyTyped(ctx, desired, &resp.Diagnostics, r.client)
	if resp.Diagnostics.HasError() {
		return
	}

	settings := settingsReadTyped(ctx, &resp.Diagnostics, r.client)
	if settings == nil {
		return
	}
	state := flattenPanelSecurity(settings)
	preservePanelSecuritySecrets(state, &plan)
	r.client.rememberConfiguredSettingSecrets(expandPanelSecurity(state))
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *PanelSecurityResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *PanelSecurityResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifyPlanWOVersion(
		ctx, req, resp,
		func(m PanelSecurityModel) types.Int64 { return m.TwoFactorTokenWOVersion },
		func(m *PanelSecurityModel, v types.String) { m.TwoFactorToken = v },
	)
}

func (r *PanelSecurityResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	settings := settingsReadTyped(ctx, &resp.Diagnostics, r.client)
	if settings == nil {
		return
	}
	state := flattenPanelSecurity(settings)
	r.client.rememberConfiguredSettingSecrets(expandPanelSecurity(state))
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func resolveSecurityTokenWO(plan *PanelSecurityModel, config PanelSecurityModel) {
	if !config.TwoFactorTokenWO.IsNull() {
		plan.TwoFactorToken = config.TwoFactorTokenWO
	}
}

func resolveSecurityTokenWOUpdate(plan *PanelSecurityModel, state PanelSecurityModel, config PanelSecurityModel) {
	if config.TwoFactorTokenWO.IsNull() {
		return
	}
	if woVersionTriggered(plan.TwoFactorTokenWOVersion, state.TwoFactorTokenWOVersion) {
		plan.TwoFactorToken = config.TwoFactorTokenWO
	}
}

func (r *PanelSecurityResource) warnIfTwoFactor(plan *PanelSecurityModel, diags *diag.Diagnostics) {
	if !plan.TwoFactorEnable.IsNull() && !plan.TwoFactorEnable.IsUnknown() && plan.TwoFactorEnable.ValueBool() {
		diags.AddWarning(
			"2FA enabled — automatic re-login will not work",
			"The provider can send a TOTP code with the initial login (via the two_factor_code provider attribute), but TOTP codes expire every 30 seconds. Automatic re-login on session expiry will fail once the code is no longer valid. Ensure you supply a fresh two_factor_code for each run, or disable 2FA to allow unattended operation.",
		)
	}
}

// ---------------------------------------------------------------------------
// PanelTelegramResource (threexui_panel_telegram)
// ---------------------------------------------------------------------------

var (
	_ resource.Resource                = &PanelTelegramResource{}
	_ resource.ResourceWithConfigure   = &PanelTelegramResource{}
	_ resource.ResourceWithImportState = &PanelTelegramResource{}
	_ resource.ResourceWithModifyPlan  = &PanelTelegramResource{}
)

type PanelTelegramResource struct {
	client *Client
}

func NewPanelTelegramResource() resource.Resource {
	return &PanelTelegramResource{}
}

func (r *PanelTelegramResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_panel_telegram"
}

func (r *PanelTelegramResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = panelTelegramSchema()
}

func (r *PanelTelegramResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *Client")
		return
	}
	r.client = client
}

func (r *PanelTelegramResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PanelTelegramModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config PanelTelegramModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resolveTelegramTokenWO(&plan, config)

	desired := expandPanelTelegram(&plan)
	settingsApplyTyped(ctx, desired, &resp.Diagnostics, r.client)
	if resp.Diagnostics.HasError() {
		return
	}

	settings := settingsReadTyped(ctx, &resp.Diagnostics, r.client)
	if settings == nil {
		return
	}
	state := flattenPanelTelegram(settings)
	preservePanelTelegramSecrets(state, &plan)
	r.client.rememberConfiguredSettingSecrets(expandPanelTelegram(state))
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *PanelTelegramResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var prior PanelTelegramModel
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings := settingsReadTyped(ctx, &resp.Diagnostics, r.client)
	if settings == nil {
		return
	}
	state := flattenPanelTelegram(settings)
	preservePanelTelegramSecrets(state, &prior)
	r.client.rememberConfiguredSettingSecrets(expandPanelTelegram(state))
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *PanelTelegramResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PanelTelegramModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var priorState PanelTelegramModel
	resp.Diagnostics.Append(req.State.Get(ctx, &priorState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config PanelTelegramModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resolveTelegramTokenWOUpdate(&plan, priorState, config)

	desired := expandPanelTelegram(&plan)
	settingsApplyTyped(ctx, desired, &resp.Diagnostics, r.client)
	if resp.Diagnostics.HasError() {
		return
	}

	settings := settingsReadTyped(ctx, &resp.Diagnostics, r.client)
	if settings == nil {
		return
	}
	state := flattenPanelTelegram(settings)
	preservePanelTelegramSecrets(state, &plan)
	r.client.rememberConfiguredSettingSecrets(expandPanelTelegram(state))
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func resolveTelegramTokenWO(plan *PanelTelegramModel, config PanelTelegramModel) {
	if !config.TgBotTokenWO.IsNull() {
		plan.TgBotToken = config.TgBotTokenWO
	}
}

func resolveTelegramTokenWOUpdate(plan *PanelTelegramModel, state PanelTelegramModel, config PanelTelegramModel) {
	if config.TgBotTokenWO.IsNull() {
		return
	}
	if woVersionTriggered(plan.TgBotTokenWOVersion, state.TgBotTokenWOVersion) {
		plan.TgBotToken = config.TgBotTokenWO
	}
}

func (r *PanelTelegramResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *PanelTelegramResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifyPlanWOVersion(
		ctx, req, resp,
		func(m PanelTelegramModel) types.Int64 { return m.TgBotTokenWOVersion },
		func(m *PanelTelegramModel, v types.String) { m.TgBotToken = v },
	)
}

func (r *PanelTelegramResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	settings := settingsReadTyped(ctx, &resp.Diagnostics, r.client)
	if settings == nil {
		return
	}
	state := flattenPanelTelegram(settings)
	r.client.rememberConfiguredSettingSecrets(expandPanelTelegram(state))
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// ---------------------------------------------------------------------------
// PanelSubscriptionResource (threexui_panel_subscription)
// ---------------------------------------------------------------------------

var (
	_ resource.Resource                = &PanelSubscriptionResource{}
	_ resource.ResourceWithConfigure   = &PanelSubscriptionResource{}
	_ resource.ResourceWithImportState = &PanelSubscriptionResource{}
)

type PanelSubscriptionResource struct {
	client *Client
}

func NewPanelSubscriptionResource() resource.Resource {
	return &PanelSubscriptionResource{}
}

func (r *PanelSubscriptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_panel_subscription"
}

func (r *PanelSubscriptionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = panelSubscriptionSchema()
}

func (r *PanelSubscriptionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *Client")
		return
	}
	r.client = client
}

func (r *PanelSubscriptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PanelSubscriptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.applySubscription(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	settings := settingsReadTyped(ctx, &resp.Diagnostics, r.client)
	if settings == nil {
		return
	}
	state := flattenPanelSubscription(settings)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *PanelSubscriptionResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	settings := settingsReadTyped(ctx, &resp.Diagnostics, r.client)
	if settings == nil {
		return
	}
	state := flattenPanelSubscription(settings)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *PanelSubscriptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PanelSubscriptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.applySubscription(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	settings := settingsReadTyped(ctx, &resp.Diagnostics, r.client)
	if settings == nil {
		return
	}
	state := flattenPanelSubscription(settings)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *PanelSubscriptionResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *PanelSubscriptionResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	settings := settingsReadTyped(ctx, &resp.Diagnostics, r.client)
	if settings == nil {
		return
	}
	state := flattenPanelSubscription(settings)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// applySubscription applies subscription settings twice to work around a 3x-ui
// bug where subJsonEnable is not persisted when subEnable changes in the same
// request.
func (r *PanelSubscriptionResource) applySubscription(ctx context.Context, plan *PanelSubscriptionModel, diags *diag.Diagnostics) {
	desired := expandPanelSubscription(plan)
	if len(desired) == 0 {
		return
	}

	settingsMu.Lock()
	defer settingsMu.Unlock()

	// First apply.
	existing, err := r.client.GetSettings(ctx)
	if err != nil {
		diags.AddError("Failed to get settings", err.Error())
		return
	}
	merged := mergeSettingsForUpdate(r.client, existing, desired)
	if err := r.client.UpdateSettings(ctx, merged); err != nil {
		diags.AddError("Failed to update settings", err.Error())
		return
	}
	r.client.rememberConfiguredSettingSecrets(desired)

	// Second apply (workaround for 3x-ui bug).
	existing2, err := r.client.GetSettings(ctx)
	if err != nil {
		diags.AddError("Failed to get settings (second apply)", err.Error())
		return
	}
	merged2 := mergeSettingsForUpdate(r.client, existing2, desired)
	if err := r.client.UpdateSettings(ctx, merged2); err != nil {
		diags.AddError("Failed to update settings (second apply)", err.Error())
		return
	}
	r.client.rememberConfiguredSettingSecrets(desired)
}

// ---------------------------------------------------------------------------
// Helper functions (no SDK dependency)
// ---------------------------------------------------------------------------

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
