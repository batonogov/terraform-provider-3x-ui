package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// Shared model and schema
// ---------------------------------------------------------------------------

type settingsResourceModel struct {
	ID   types.String `tfsdk:"id"`
	JSON types.String `tfsdk:"json"`
}

func settingsResourceSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"json": schema.StringAttribute{
				Required:    true,
				Description: "JSON object with settings fields (snake_case keys).",
			},
		},
	}
}

// ---------------------------------------------------------------------------
// settingsApplyHelper: shared apply logic for all settings resources
// ---------------------------------------------------------------------------

func settingsApplyHelper(
	ctx context.Context,
	jsonStr string,
	diags *diag.Diagnostics,
	client *Client,
	expand func(map[string]any) (map[string]any, bool),
	flatten func(map[string]any) map[string]any,
) *settingsResourceModel {
	var input map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
		diags.AddError("Invalid JSON", err.Error())
		return nil
	}

	desired, ok := expand(input)
	if !ok {
		// Nothing to apply; read current state.
		return settingsReadHelper(ctx, diags, client, flatten)
	}

	existing, err := client.GetSettings(ctx)
	if err != nil {
		diags.AddError("Failed to get settings", err.Error())
		return nil
	}

	merged := mergeSettings(existing, desired)
	if err := client.UpdateSettings(ctx, merged); err != nil {
		diags.AddError("Failed to update settings", err.Error())
		return nil
	}

	return settingsReadHelper(ctx, diags, client, flatten)
}

func settingsReadHelper(
	ctx context.Context,
	diags *diag.Diagnostics,
	client *Client,
	flatten func(map[string]any) map[string]any,
) *settingsResourceModel {
	settings, err := client.GetSettings(ctx)
	if err != nil {
		diags.AddError("Failed to get settings", err.Error())
		return nil
	}

	flat := flatten(settings)
	payload, err := json.Marshal(flat)
	if err != nil {
		diags.AddError("Failed to marshal settings", err.Error())
		return nil
	}

	return &settingsResourceModel{
		ID:   types.StringValue("settings"),
		JSON: types.StringValue(string(payload)),
	}
}

// ---------------------------------------------------------------------------
// PanelGeneralResource (threexui_panel_general)
// ---------------------------------------------------------------------------

var (
	_ resource.Resource                = &PanelGeneralResource{}
	_ resource.ResourceWithConfigure   = &PanelGeneralResource{}
	_ resource.ResourceWithImportState = &PanelGeneralResource{}
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
	resp.Schema = settingsResourceSchema()
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

func (r *PanelGeneralResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan settingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.applyPanelGeneral(ctx, plan.JSON.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := settingsReadHelper(ctx, &resp.Diagnostics, r.client, flattenPanelSettingsFields)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *PanelGeneralResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	state := settingsReadHelper(ctx, &resp.Diagnostics, r.client, flattenPanelSettingsFields)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *PanelGeneralResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan settingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.applyPanelGeneral(ctx, plan.JSON.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := settingsReadHelper(ctx, &resp.Diagnostics, r.client, flattenPanelSettingsFields)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *PanelGeneralResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *PanelGeneralResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	state := settingsReadHelper(ctx, &resp.Diagnostics, r.client, flattenPanelSettingsFields)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *PanelGeneralResource) applyPanelGeneral(ctx context.Context, jsonStr string, diags *diag.Diagnostics) {
	var input map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
		diags.AddError("Invalid JSON", err.Error())
		return
	}

	desired, ok := expandPanelSettingsFields(input)
	if !ok {
		return
	}

	// Warn about web_base_path change.
	if _, hasBasePath := desired["webBasePath"]; hasBasePath {
		diags.AddWarning(
			"Changing web_base_path requires updating provider config",
			"The provider's base_path must match the panel's web_base_path. After this change, update the provider configuration to use the new base_path, otherwise the provider will not be able to connect.",
		)
	}

	existing, err := r.client.GetSettings(ctx)
	if err != nil {
		diags.AddError("Failed to get settings", err.Error())
		return
	}

	needRestart := panelSettingsNeedRestart(existing, desired)
	merged := mergeSettings(existing, desired)
	if err := r.client.UpdateSettings(ctx, merged); err != nil {
		diags.AddError("Failed to update settings", err.Error())
		return
	}

	if needRestart {
		if err := r.client.RestartPanel(ctx); err != nil {
			diags.AddError("Failed to restart panel", err.Error())
			return
		}
	}
}

// ---------------------------------------------------------------------------
// PanelSecurityResource (threexui_panel_security)
// ---------------------------------------------------------------------------

var (
	_ resource.Resource                = &PanelSecurityResource{}
	_ resource.ResourceWithConfigure   = &PanelSecurityResource{}
	_ resource.ResourceWithImportState = &PanelSecurityResource{}
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
	resp.Schema = settingsResourceSchema()
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
	var plan settingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.warnIfTwoFactor(plan.JSON.ValueString(), &resp.Diagnostics)

	state := settingsApplyHelper(ctx, plan.JSON.ValueString(), &resp.Diagnostics, r.client,
		expandAccountSettingsFields, flattenAccountSettingsFields)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *PanelSecurityResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	state := settingsReadHelper(ctx, &resp.Diagnostics, r.client, flattenAccountSettingsFields)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *PanelSecurityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan settingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.warnIfTwoFactor(plan.JSON.ValueString(), &resp.Diagnostics)

	state := settingsApplyHelper(ctx, plan.JSON.ValueString(), &resp.Diagnostics, r.client,
		expandAccountSettingsFields, flattenAccountSettingsFields)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *PanelSecurityResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *PanelSecurityResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	state := settingsReadHelper(ctx, &resp.Diagnostics, r.client, flattenAccountSettingsFields)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *PanelSecurityResource) warnIfTwoFactor(jsonStr string, diags *diag.Diagnostics) {
	var input map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
		return
	}
	if v, ok := input["two_factor_enable"]; ok {
		if b, isBool := v.(bool); isBool && b {
			diags.AddWarning(
				"Enabling 2FA will block provider authentication",
				"The provider does not support two-factor authentication codes during login. Enabling 2FA will prevent the provider from connecting to the panel. You will need to disable 2FA via the API or UI to restore provider access.",
			)
		}
	}
}

// ---------------------------------------------------------------------------
// PanelTelegramResource (threexui_panel_telegram)
// ---------------------------------------------------------------------------

var (
	_ resource.Resource                = &PanelTelegramResource{}
	_ resource.ResourceWithConfigure   = &PanelTelegramResource{}
	_ resource.ResourceWithImportState = &PanelTelegramResource{}
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
	resp.Schema = settingsResourceSchema()
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
	var plan settingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := settingsApplyHelper(ctx, plan.JSON.ValueString(), &resp.Diagnostics, r.client,
		expandTelegramSettingsFields, flattenTelegramSettingsFields)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *PanelTelegramResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	state := settingsReadHelper(ctx, &resp.Diagnostics, r.client, flattenTelegramSettingsFields)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *PanelTelegramResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan settingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := settingsApplyHelper(ctx, plan.JSON.ValueString(), &resp.Diagnostics, r.client,
		expandTelegramSettingsFields, flattenTelegramSettingsFields)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *PanelTelegramResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *PanelTelegramResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	state := settingsReadHelper(ctx, &resp.Diagnostics, r.client, flattenTelegramSettingsFields)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
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
	resp.Schema = settingsResourceSchema()
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
	var plan settingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.applySubscription(ctx, plan.JSON.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := settingsReadHelper(ctx, &resp.Diagnostics, r.client, flattenSubscriptionSettingsFields)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *PanelSubscriptionResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	state := settingsReadHelper(ctx, &resp.Diagnostics, r.client, flattenSubscriptionSettingsFields)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *PanelSubscriptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan settingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.applySubscription(ctx, plan.JSON.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := settingsReadHelper(ctx, &resp.Diagnostics, r.client, flattenSubscriptionSettingsFields)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *PanelSubscriptionResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *PanelSubscriptionResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	state := settingsReadHelper(ctx, &resp.Diagnostics, r.client, flattenSubscriptionSettingsFields)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

// applySubscription applies subscription settings twice to work around a 3x-ui
// bug where subJsonEnable is not persisted when subEnable changes in the same
// request.
func (r *PanelSubscriptionResource) applySubscription(ctx context.Context, jsonStr string, diags *diag.Diagnostics) {
	var input map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
		diags.AddError("Invalid JSON", err.Error())
		return
	}

	desired, ok := expandSubscriptionSettingsFields(input)
	if !ok {
		return
	}

	// First apply.
	existing, err := r.client.GetSettings(ctx)
	if err != nil {
		diags.AddError("Failed to get settings", err.Error())
		return
	}
	merged := mergeSettings(existing, desired)
	if err := r.client.UpdateSettings(ctx, merged); err != nil {
		diags.AddError("Failed to update settings", err.Error())
		return
	}

	// Second apply (workaround for 3x-ui bug).
	existing2, err := r.client.GetSettings(ctx)
	if err != nil {
		diags.AddError("Failed to get settings (second apply)", err.Error())
		return
	}
	merged2 := mergeSettings(existing2, desired)
	if err := r.client.UpdateSettings(ctx, merged2); err != nil {
		diags.AddError("Failed to update settings (second apply)", err.Error())
		return
	}
}

// ---------------------------------------------------------------------------
// Expand functions: snake_case map[string]any -> camelCase API payload
// ---------------------------------------------------------------------------

func expandPanelSettingsFields(input map[string]any) (map[string]any, bool) {
	payload := map[string]any{}
	if v, ok := input["web_listen"].(string); ok {
		payload["webListen"] = v
	}
	if v, ok := input["web_domain"].(string); ok {
		payload["webDomain"] = v
	}
	if v, ok := input["web_port"]; ok {
		payload["webPort"] = jsonNumber(v)
	}
	if v, ok := input["web_base_path"].(string); ok {
		payload["webBasePath"] = v
	}
	if v, ok := input["session_max_age"]; ok {
		payload["sessionMaxAge"] = jsonNumber(v)
	}
	if v, ok := input["page_size"]; ok {
		payload["pageSize"] = jsonNumber(v)
	}
	if v, ok := input["remark_model"].(string); ok {
		payload["remarkModel"] = v
	}
	if v, ok := input["date_picker"].(string); ok {
		payload["datepicker"] = v
	}
	if v, ok := input["time_location"].(string); ok {
		payload["timeLocation"] = v
	}
	if v, ok := input["expire_diff"]; ok {
		payload["expireDiff"] = jsonNumber(v)
	}
	if v, ok := input["traffic_diff"]; ok {
		payload["trafficDiff"] = jsonNumber(v)
	}
	if v, ok := input["web_cert_file"].(string); ok {
		payload["webCertFile"] = v
	}
	if v, ok := input["web_key_file"].(string); ok {
		payload["webKeyFile"] = v
	}
	if v, ok := input["external_traffic_inform_enable"].(bool); ok {
		payload["externalTrafficInformEnable"] = v
	}
	if v, ok := input["external_traffic_inform_uri"].(string); ok {
		payload["externalTrafficInformURI"] = v
	}
	if v, ok := input["ldap_enable"].(bool); ok {
		payload["ldapEnable"] = v
	}
	if v, ok := input["ldap_host"].(string); ok {
		payload["ldapHost"] = v
	}
	if v, ok := input["ldap_port"]; ok {
		payload["ldapPort"] = jsonNumber(v)
	}
	if v, ok := input["ldap_use_tls"].(bool); ok {
		payload["ldapUseTLS"] = v
	}
	if v, ok := input["ldap_bind_dn"].(string); ok {
		payload["ldapBindDN"] = v
	}
	if v, ok := input["ldap_password"].(string); ok {
		payload["ldapPassword"] = v
	}
	if v, ok := input["ldap_base_dn"].(string); ok {
		payload["ldapBaseDN"] = v
	}
	if v, ok := input["ldap_user_filter"].(string); ok {
		payload["ldapUserFilter"] = v
	}
	if v, ok := input["ldap_user_attr"].(string); ok {
		payload["ldapUserAttr"] = v
	}
	if v, ok := input["ldap_vless_field"].(string); ok {
		payload["ldapVlessField"] = v
	}
	if v, ok := input["ldap_sync_cron"].(string); ok {
		payload["ldapSyncCron"] = v
	}
	if v, ok := input["ldap_flag_field"].(string); ok {
		payload["ldapFlagField"] = v
	}
	if v, ok := input["ldap_truthy_values"].(string); ok {
		payload["ldapTruthyValues"] = v
	}
	if v, ok := input["ldap_invert_flag"].(bool); ok {
		payload["ldapInvertFlag"] = v
	}
	if v, ok := input["ldap_inbound_tags"].(string); ok {
		payload["ldapInboundTags"] = v
	}
	if v, ok := input["ldap_auto_create"].(bool); ok {
		payload["ldapAutoCreate"] = v
	}
	if v, ok := input["ldap_auto_delete"].(bool); ok {
		payload["ldapAutoDelete"] = v
	}
	if v, ok := input["ldap_default_total_gb"]; ok {
		payload["ldapDefaultTotalGB"] = jsonNumber(v)
	}
	if v, ok := input["ldap_default_expiry_days"]; ok {
		payload["ldapDefaultExpiryDays"] = jsonNumber(v)
	}
	if v, ok := input["ldap_default_limit_ip"]; ok {
		payload["ldapDefaultLimitIP"] = jsonNumber(v)
	}
	return payload, len(payload) > 0
}

func expandAccountSettingsFields(input map[string]any) (map[string]any, bool) {
	payload := map[string]any{}
	if v, ok := input["two_factor_enable"].(bool); ok {
		payload["twoFactorEnable"] = v
	}
	if v, ok := input["two_factor_token"].(string); ok {
		payload["twoFactorToken"] = v
	}
	return payload, len(payload) > 0
}

func expandTelegramSettingsFields(input map[string]any) (map[string]any, bool) {
	payload := map[string]any{}
	if v, ok := input["tg_bot_enable"].(bool); ok {
		payload["tgBotEnable"] = v
	}
	if v, ok := input["tg_bot_token"].(string); ok {
		payload["tgBotToken"] = v
	}
	if v, ok := input["tg_bot_proxy"].(string); ok {
		payload["tgBotProxy"] = v
	}
	if v, ok := input["tg_bot_api_server"].(string); ok {
		payload["tgBotAPIServer"] = v
	}
	if v, ok := input["tg_bot_chat_id"].(string); ok {
		payload["tgBotChatId"] = v
	}
	if v, ok := input["tg_lang"].(string); ok {
		payload["tgLang"] = v
	}
	if v, ok := input["tg_run_time"].(string); ok {
		payload["tgRunTime"] = v
	}
	if v, ok := input["tg_bot_backup"].(bool); ok {
		payload["tgBotBackup"] = v
	}
	if v, ok := input["tg_bot_login_notify"].(bool); ok {
		payload["tgBotLoginNotify"] = v
	}
	if v, ok := input["tg_cpu"]; ok {
		payload["tgCpu"] = jsonNumber(v)
	}
	return payload, len(payload) > 0
}

func expandSubscriptionSettingsFields(input map[string]any) (map[string]any, bool) {
	payload := map[string]any{}
	if v, ok := input["sub_enable"].(bool); ok {
		payload["subEnable"] = v
	}
	if v, ok := input["sub_json_enable"].(bool); ok {
		payload["subJsonEnable"] = v
	}
	if v, ok := input["sub_title"].(string); ok {
		payload["subTitle"] = v
	}
	if v, ok := input["sub_support_url"].(string); ok {
		payload["subSupportUrl"] = v
	}
	if v, ok := input["sub_profile_url"].(string); ok {
		payload["subProfileUrl"] = v
	}
	if v, ok := input["sub_announce"].(string); ok {
		payload["subAnnounce"] = v
	}
	if v, ok := input["sub_enable_routing"].(bool); ok {
		payload["subEnableRouting"] = v
	}
	if v, ok := input["sub_routing_rules"].(string); ok {
		payload["subRoutingRules"] = v
	}
	if v, ok := input["sub_listen"].(string); ok {
		payload["subListen"] = v
	}
	if v, ok := input["sub_port"]; ok {
		payload["subPort"] = jsonNumber(v)
	}
	if v, ok := input["sub_path"].(string); ok {
		payload["subPath"] = v
	}
	if v, ok := input["sub_domain"].(string); ok {
		payload["subDomain"] = v
	}
	if v, ok := input["sub_cert_file"].(string); ok {
		payload["subCertFile"] = v
	}
	if v, ok := input["sub_key_file"].(string); ok {
		payload["subKeyFile"] = v
	}
	if v, ok := input["sub_updates"]; ok {
		payload["subUpdates"] = jsonNumber(v)
	}
	if v, ok := input["sub_encrypt"].(bool); ok {
		payload["subEncrypt"] = v
	}
	if v, ok := input["sub_show_info"].(bool); ok {
		payload["subShowInfo"] = v
	}
	if v, ok := input["sub_uri"].(string); ok {
		payload["subURI"] = v
	}
	if v, ok := input["sub_json_path"].(string); ok {
		payload["subJsonPath"] = v
	}
	if v, ok := input["sub_json_uri"].(string); ok {
		payload["subJsonURI"] = v
	}
	if v, ok := input["sub_json_fragment"].(string); ok {
		payload["subJsonFragment"] = v
	}
	if v, ok := input["sub_json_noises"].(string); ok {
		payload["subJsonNoises"] = v
	}
	if v, ok := input["sub_json_mux"].(string); ok {
		payload["subJsonMux"] = v
	}
	if v, ok := input["sub_json_rules"].(string); ok {
		payload["subJsonRules"] = v
	}
	return payload, len(payload) > 0
}

// jsonNumber converts a JSON-decoded number (float64) to int for the API.
// JSON numbers from json.Unmarshal are always float64.
func jsonNumber(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case float32:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	default:
		return 0
	}
}

// ---------------------------------------------------------------------------
// Flatten functions: camelCase API response -> snake_case map[string]any
// These have no SDK dependency and are preserved as-is.
// ---------------------------------------------------------------------------

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
