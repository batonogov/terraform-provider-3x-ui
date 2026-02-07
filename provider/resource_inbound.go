package provider

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &InboundResource{}
	_ resource.ResourceWithImportState = &InboundResource{}
)

type InboundResource struct {
	client *Client
}

type InboundResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Up                   types.Int64  `tfsdk:"up"`
	Down                 types.Int64  `tfsdk:"down"`
	Total                types.Int64  `tfsdk:"total"`
	AllTime              types.Int64  `tfsdk:"all_time"`
	Remark               types.String `tfsdk:"remark"`
	Enable               types.Bool   `tfsdk:"enable"`
	ExpiryTime           types.Int64  `tfsdk:"expiry_time"`
	TrafficReset         types.String `tfsdk:"traffic_reset"`
	LastTrafficResetTime types.Int64  `tfsdk:"last_traffic_reset_time"`
	Listen               types.String `tfsdk:"listen"`
	Port                 types.Int64  `tfsdk:"port"`
	Protocol             types.String `tfsdk:"protocol"`
	Settings             types.String `tfsdk:"settings"`
	StreamSettings       types.String `tfsdk:"stream_settings"`
	Sniffing             types.String `tfsdk:"sniffing"`
	Tag                  types.String `tfsdk:"tag"`
}

func NewInboundResource() resource.Resource {
	return &InboundResource{}
}

func (r *InboundResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_inbound"
}

func (r *InboundResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a 3x-ui inbound.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Numeric ID of the inbound.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"up": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Upload traffic (bytes).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"down": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Download traffic (bytes).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"total": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Total traffic limit (bytes). 0 means unlimited.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"all_time": schema.Int64Attribute{
				Computed:    true,
				Description: "All-time accumulated traffic (bytes).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"remark": schema.StringAttribute{
				Optional:    true,
				Description: "Remark / display name for the inbound.",
			},
			"enable": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the inbound is enabled.",
			},
			"expiry_time": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Expiry time in milliseconds since epoch. 0 means never.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"traffic_reset": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("never"),
				Description: "Traffic reset interval (e.g. 'never', 'day', 'week', 'month', 'year').",
			},
			"last_traffic_reset_time": schema.Int64Attribute{
				Computed:    true,
				Description: "Last traffic reset time in milliseconds since epoch.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"listen": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Listen address.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"port": schema.Int64Attribute{
				Required:    true,
				Description: "Listen port.",
			},
			"protocol": schema.StringAttribute{
				Required:    true,
				Description: "Protocol (vless, vmess, trojan, shadowsocks, http, socks, wireguard, etc.).",
			},
			"settings": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Inbound settings as a JSON string.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					jsonSubsetPlanModifier{},
				},
			},
			"stream_settings": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Stream settings as a JSON string.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					jsonSubsetPlanModifier{},
				},
			},
			"sniffing": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Sniffing settings as a JSON string.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					jsonSubsetPlanModifier{},
				},
			},
			"tag": schema.StringAttribute{
				Computed:    true,
				Description: "Xray inbound tag (auto-generated by the panel).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *InboundResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("Expected *Client, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *InboundResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan InboundResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inbound := expandInboundFromModel(&plan)

	settingsJSON, err := ensureVlessEncFromAuth(ctx, r.client, inbound.Settings, inbound.Protocol)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve VLESS encryption", err.Error())
		return
	}
	inbound.Settings = settingsJSON

	if err := applyDefaultInboundSettings(inbound); err != nil {
		resp.Diagnostics.AddError("Failed to apply default inbound settings", err.Error())
		return
	}
	if err := ensureRealityKeys(ctx, r.client, inbound, nil); err != nil {
		resp.Diagnostics.AddError("Failed to ensure Reality keys", err.Error())
		return
	}
	if err := ensureInboundClientIDs(inbound); err != nil {
		resp.Diagnostics.AddError("Failed to ensure inbound client IDs", err.Error())
		return
	}

	created, err := r.client.AddInbound(ctx, inbound)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create inbound", err.Error())
		return
	}
	if created == nil {
		resp.Diagnostics.AddError("Empty response", "API returned nil inbound")
		return
	}

	state := inboundToModel(created)
	// Preserve plan values for JSON fields when the user specified them.
	// API may add defaults (e.g. clients:[] in settings, tcpSettings in
	// stream_settings). The plan modifier handles Read→Plan diff suppression.
	if !plan.Settings.IsNull() && !plan.Settings.IsUnknown() {
		state.Settings = plan.Settings
	}
	if !plan.StreamSettings.IsNull() && !plan.StreamSettings.IsUnknown() {
		state.StreamSettings = plan.StreamSettings
	}
	if !plan.Sniffing.IsNull() && !plan.Sniffing.IsUnknown() {
		state.Sniffing = plan.Sniffing
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *InboundResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state InboundResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := parseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", err.Error())
		return
	}

	inbound, err := r.client.GetInbound(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read inbound", err.Error())
		return
	}

	newState := inboundToModel(inbound)
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *InboundResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan InboundResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state InboundResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := parseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", err.Error())
		return
	}

	existing, err := r.client.GetInbound(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read existing inbound", err.Error())
		return
	}

	inbound := expandInboundFromModel(&plan)

	settingsJSON, err := ensureVlessEncFromAuth(ctx, r.client, inbound.Settings, inbound.Protocol)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve VLESS encryption", err.Error())
		return
	}
	inbound.Settings = settingsJSON

	if err := applyDefaultInboundSettings(inbound); err != nil {
		resp.Diagnostics.AddError("Failed to apply default inbound settings", err.Error())
		return
	}
	if err := ensureRealityKeys(ctx, r.client, inbound, existing); err != nil {
		resp.Diagnostics.AddError("Failed to ensure Reality keys", err.Error())
		return
	}
	if err := preserveInboundSettings(inbound, existing); err != nil {
		resp.Diagnostics.AddError("Failed to preserve inbound settings", err.Error())
		return
	}
	inbound.ID = id

	updated, err := r.client.UpdateInbound(ctx, inbound)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update inbound", err.Error())
		return
	}

	newState := inboundToModel(updated)
	// Preserve plan values for JSON fields when the user specified them.
	if !plan.Settings.IsNull() && !plan.Settings.IsUnknown() {
		newState.Settings = plan.Settings
	}
	if !plan.StreamSettings.IsNull() && !plan.StreamSettings.IsUnknown() {
		newState.StreamSettings = plan.StreamSettings
	}
	if !plan.Sniffing.IsNull() && !plan.Sniffing.IsUnknown() {
		newState.Sniffing = plan.Sniffing
	}
	// Tag is Computed-only (auto-generated by panel). The API may return
	// empty tag on update; preserve the plan value (prior state).
	if newState.Tag.ValueString() == "" && !plan.Tag.IsNull() && !plan.Tag.IsUnknown() {
		newState.Tag = plan.Tag
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *InboundResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state InboundResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := parseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", err.Error())
		return
	}

	if err := r.client.DeleteInbound(ctx, id); err != nil {
		resp.Diagnostics.AddError("Failed to delete inbound", err.Error())
		return
	}
}

func (r *InboundResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ---------------------------------------------------------------------------
// Model <-> Inbound conversion
// ---------------------------------------------------------------------------

func expandInboundFromModel(m *InboundResourceModel) *Inbound {
	settings := m.Settings.ValueString()
	if settings == "" {
		settings = "{}"
	}
	streamSettings := m.StreamSettings.ValueString()
	if streamSettings == "" {
		streamSettings = "{}"
	}
	sniffing := m.Sniffing.ValueString()
	if sniffing == "" {
		sniffing = "{}"
	}

	return &Inbound{
		Up:                   m.Up.ValueInt64(),
		Down:                 m.Down.ValueInt64(),
		Total:                m.Total.ValueInt64(),
		Remark:               m.Remark.ValueString(),
		Enable:               m.Enable.ValueBool(),
		ExpiryTime:           m.ExpiryTime.ValueInt64(),
		TrafficReset:         m.TrafficReset.ValueString(),
		LastTrafficResetTime: m.LastTrafficResetTime.ValueInt64(),
		Listen:               m.Listen.ValueString(),
		Port:                 int(m.Port.ValueInt64()),
		Protocol:             m.Protocol.ValueString(),
		Settings:             settings,
		StreamSettings:       streamSettings,
		Sniffing:             sniffing,
	}
}

func inboundToModel(inbound *Inbound) *InboundResourceModel {
	return &InboundResourceModel{
		ID:                   types.StringValue(fmt.Sprintf("%d", inbound.ID)),
		Up:                   types.Int64Value(inbound.Up),
		Down:                 types.Int64Value(inbound.Down),
		Total:                types.Int64Value(inbound.Total),
		AllTime:              types.Int64Value(inbound.AllTime),
		Remark:               types.StringValue(inbound.Remark),
		Enable:               types.BoolValue(inbound.Enable),
		ExpiryTime:           types.Int64Value(inbound.ExpiryTime),
		TrafficReset:         types.StringValue(inbound.TrafficReset),
		LastTrafficResetTime: types.Int64Value(inbound.LastTrafficResetTime),
		Listen:               types.StringValue(inbound.Listen),
		Port:                 types.Int64Value(int64(inbound.Port)),
		Protocol:             types.StringValue(inbound.Protocol),
		Settings:             types.StringValue(inbound.Settings),
		StreamSettings:       types.StringValue(inbound.StreamSettings),
		Sniffing:             types.StringValue(inbound.Sniffing),
		Tag:                  types.StringValue(inbound.Tag),
	}
}

// ---------------------------------------------------------------------------
// ensureVlessEncFromAuth — resolves VLESS decryption/encryption from the
// panel's auth endpoint when selectedAuth is set but decryption/encryption
// are missing in the settings JSON.
// ---------------------------------------------------------------------------

func ensureVlessEncFromAuth(ctx context.Context, client *Client, settingsJSON string, protocol string) (string, error) {
	if protocol != "vless" || client == nil {
		return settingsJSON, nil
	}
	if strings.TrimSpace(settingsJSON) == "" {
		return settingsJSON, nil
	}

	settings, err := ParseJSONField(settingsJSON)
	if err != nil {
		return settingsJSON, err
	}

	selected := stringValue(settings["selectedAuth"])
	if selected == "" {
		return settingsJSON, nil
	}

	decryptionMissing := stringValue(settings["decryption"]) == ""
	encryptionMissing := stringValue(settings["encryption"]) == ""
	if !decryptionMissing && !encryptionMissing {
		return settingsJSON, nil
	}

	auths, err := client.GetNewVlessEnc(ctx)
	if err != nil {
		return settingsJSON, err
	}

	var match *VlessEncAuth
	for i := range auths {
		if auths[i].Label == selected {
			match = &auths[i]
			break
		}
	}
	if match == nil {
		return settingsJSON, fmt.Errorf("no auth block for selected_auth %q", selected)
	}

	if decryptionMissing {
		settings["decryption"] = match.Decryption
	}
	if encryptionMissing {
		settings["encryption"] = match.Encryption
	}

	updated, err := json.Marshal(settings)
	if err != nil {
		return settingsJSON, err
	}
	return string(updated), nil
}

// ---------------------------------------------------------------------------
// isSubset — standalone utility function (no SDK dependency)
// ---------------------------------------------------------------------------

func isSubset(desired, actual any) bool {
	switch dv := desired.(type) {
	case map[string]any:
		av, ok := actual.(map[string]any)
		if !ok {
			return false
		}
		for k, dval := range dv {
			aval, ok := av[k]
			if !ok {
				return false
			}
			if !isSubset(dval, aval) {
				return false
			}
		}
		return true
	case []any:
		av, ok := actual.([]any)
		if !ok {
			return false
		}
		if len(dv) == 0 {
			return true
		}
		if len(dv) > len(av) {
			return false
		}
		for i := range dv {
			found := false
			for j := range av {
				if isSubset(dv[i], av[j]) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(desired, actual)
	}
}

// ---------------------------------------------------------------------------
// jsonSubsetPlanModifier suppresses diffs when the config value is a JSON
// subset of the prior state value. This handles API-added defaults (like
// clients:[] in settings, tcpSettings defaults in stream_settings).
// ---------------------------------------------------------------------------

type jsonSubsetPlanModifier struct{}

func (m jsonSubsetPlanModifier) Description(_ context.Context) string {
	return "Suppresses diffs when config is a JSON subset of prior state."
}

func (m jsonSubsetPlanModifier) MarkdownDescription(_ context.Context) string {
	return "Suppresses diffs when config is a JSON subset of prior state."
}

func (m jsonSubsetPlanModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// No prior state (Create) or config is null/unknown — nothing to do.
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() ||
		req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	configJSON := req.ConfigValue.ValueString()
	stateJSON := req.StateValue.ValueString()

	var configVal, stateVal any
	if err := json.Unmarshal([]byte(configJSON), &configVal); err != nil {
		return
	}
	if err := json.Unmarshal([]byte(stateJSON), &stateVal); err != nil {
		return
	}

	if isSubset(configVal, stateVal) {
		resp.PlanValue = req.StateValue
	}
}

// ---------------------------------------------------------------------------
// preserveInboundSettings / preserveSettingsKey — preserve clients and
// testseed from existing inbound during update (no SDK dependency)
// ---------------------------------------------------------------------------

func preserveInboundSettings(desired *Inbound, existing *Inbound) error {
	if desired == nil || existing == nil {
		return nil
	}
	if strings.TrimSpace(desired.Settings) == "" || strings.TrimSpace(existing.Settings) == "" {
		return nil
	}
	desiredSettings, err := ParseJSONField(desired.Settings)
	if err != nil {
		return err
	}
	existingSettings, err := ParseJSONField(existing.Settings)
	if err != nil {
		return err
	}
	if !preserveSettingsKey(desiredSettings, existingSettings, "clients") {
		return nil
	}
	_ = preserveSettingsKey(desiredSettings, existingSettings, "testseed")
	updated, err := json.Marshal(desiredSettings)
	if err != nil {
		return err
	}
	desired.Settings = string(updated)
	return nil
}

func preserveSettingsKey(desired, existing map[string]any, key string) bool {
	if existing == nil {
		return false
	}
	existingVal, ok := existing[key]
	if !ok {
		return false
	}
	switch v := existingVal.(type) {
	case []any:
		if len(v) == 0 {
			return false
		}
	}
	if desired == nil {
		return false
	}
	if desiredVal, ok := desired[key]; ok {
		if list, ok := desiredVal.([]any); ok && len(list) > 0 {
			return false
		}
	}
	desired[key] = existingVal
	return true
}

// ---------------------------------------------------------------------------
// ensureInboundClientIDs — auto-generate UUIDs for clients without id
// (no SDK dependency)
// ---------------------------------------------------------------------------

func ensureInboundClientIDs(inbound *Inbound) error {
	if inbound == nil {
		return nil
	}
	settings, err := ParseJSONField(inbound.Settings)
	if err != nil {
		return err
	}
	clientsRaw, ok := settings["clients"]
	if !ok {
		return nil
	}
	clients, ok := clientsRaw.([]any)
	if !ok {
		return nil
	}
	changed := false
	for i := range clients {
		clientMap, ok := clients[i].(map[string]any)
		if !ok {
			continue
		}
		id, _ := clientMap["id"].(string)
		if id == "" {
			newID, err := newUUID()
			if err != nil {
				return err
			}
			clientMap["id"] = newID
			clients[i] = clientMap
			changed = true
		}
	}
	if !changed {
		return nil
	}
	settings["clients"] = clients
	updated, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	inbound.Settings = string(updated)
	return nil
}

// ---------------------------------------------------------------------------
// Reality key helpers (no SDK dependency)
// ---------------------------------------------------------------------------

func ensureRealityKeys(ctx context.Context, client *Client, inbound *Inbound, existing *Inbound) error {
	if inbound == nil || inbound.StreamSettings == "" {
		return nil
	}
	payload, err := ParseJSONField(inbound.StreamSettings)
	if err != nil {
		return err
	}
	security := stringValue(payload["security"])
	if security != "reality" {
		return nil
	}
	rs, _ := payload["realitySettings"].(map[string]any)
	if rs == nil {
		rs = map[string]any{}
	}
	mergeRealityFromExisting(existing, rs)
	ensureRealityDefaults(rs)
	if !hasRealityShortIDs(rs) {
		rs["shortIds"] = randomShortIDs()
	}
	if pk, ok := rs["privateKey"].(string); ok && pk != "" {
		return nil
	}
	cert, err := client.GetNewX25519Cert(ctx)
	if err != nil {
		return err
	}
	privateKey := stringValue(cert["privateKey"])
	publicKey := stringValue(cert["publicKey"])
	if privateKey == "" {
		return fmt.Errorf("generated reality privateKey is empty")
	}
	rs["privateKey"] = privateKey
	settings, _ := rs["settings"].(map[string]any)
	if settings == nil {
		settings = map[string]any{}
	}
	if settings["publicKey"] == nil || stringValue(settings["publicKey"]) == "" {
		settings["publicKey"] = publicKey
	}
	rs["settings"] = settings
	payload["realitySettings"] = rs
	updated, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	inbound.StreamSettings = string(updated)
	return nil
}

func ensureRealityDefaults(reality map[string]any) {
	if reality == nil {
		return
	}
	if hasStringListValues(reality["serverNames"]) {
		return
	}
	target := stringValue(reality["target"])
	if target != "" {
		host := strings.Split(target, ":")[0]
		if host != "" {
			reality["serverNames"] = []any{host}
			return
		}
	}
	reality["target"] = "www.apple.com:443"
	reality["serverNames"] = []any{"www.apple.com", "apple.com"}
}

func mergeRealityFromExisting(existing *Inbound, reality map[string]any) {
	if existing == nil || existing.StreamSettings == "" {
		return
	}
	payload, err := ParseJSONField(existing.StreamSettings)
	if err != nil {
		return
	}
	rs, _ := payload["realitySettings"].(map[string]any)
	if rs == nil {
		return
	}
	if stringValue(reality["privateKey"]) == "" {
		if pk := stringValue(rs["privateKey"]); pk != "" {
			reality["privateKey"] = pk
		}
	}
	if !hasRealityShortIDs(reality) {
		if raw, ok := rs["shortIds"]; ok {
			if list, ok := raw.([]any); ok && len(list) > 0 {
				reality["shortIds"] = list
			}
		}
	}
	settings, _ := reality["settings"].(map[string]any)
	if settings == nil {
		settings = map[string]any{}
	}
	if stringValue(settings["publicKey"]) == "" {
		if s, ok := rs["settings"].(map[string]any); ok {
			if pk := stringValue(s["publicKey"]); pk != "" {
				settings["publicKey"] = pk
			}
		}
	}
	reality["settings"] = settings
}

func hasRealityShortIDs(reality map[string]any) bool {
	if reality == nil {
		return false
	}
	return hasStringListValues(reality["shortIds"])
}

func hasRealityServerNames(reality map[string]any) bool {
	if reality == nil {
		return false
	}
	return hasStringListValues(reality["serverNames"])
}

func hasStringListValues(raw any) bool {
	switch v := raw.(type) {
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				return true
			}
		}
	case []string:
		for _, s := range v {
			if s != "" {
				return true
			}
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Random helpers (no SDK dependency)
// ---------------------------------------------------------------------------

func randomHex(length int) string {
	if length <= 0 {
		return ""
	}
	buf := make([]byte, (length+1)/2)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	out := hex.EncodeToString(buf)
	if len(out) > length {
		out = out[:length]
	}
	return out
}

func randomShortIDs() []any {
	lengths := []int{2, 4, 6, 8, 10, 12, 14, 16}
	out := make([]any, 0, len(lengths))
	for _, l := range lengths {
		out = append(out, randomHex(l))
	}
	return out
}

// ---------------------------------------------------------------------------
// parseID — parses a numeric string ID (no SDK dependency)
// ---------------------------------------------------------------------------

func parseID(id string) (int, error) {
	var parsed int
	_, err := fmt.Sscanf(id, "%d", &parsed)
	if err != nil || parsed == 0 {
		return 0, fmt.Errorf("invalid id: %s", id)
	}
	return parsed, nil
}
