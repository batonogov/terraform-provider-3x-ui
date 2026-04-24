package provider

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &InboundResource{}
	_ resource.ResourceWithImportState = &InboundResource{}
	_ resource.ResourceWithModifyPlan  = &InboundResource{}
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
	Tag                  types.String `tfsdk:"tag"`

	// Per-protocol settings (typed blocks)
	VlessSettings       *InboundVlessSettingsModel       `tfsdk:"vless_settings"`
	TrojanSettings      *InboundTrojanSettingsModel      `tfsdk:"trojan_settings"`
	ShadowsocksSettings *InboundShadowsocksSettingsModel `tfsdk:"shadowsocks_settings"`
	HTTPSettings        *InboundHTTPSettingsModel        `tfsdk:"http_settings"`
	SocksSettings       *InboundSocksSettingsModel       `tfsdk:"socks_settings"`
	WireguardSettings   *InboundWireguardSettingsModel   `tfsdk:"wireguard_settings"`
	DokodemoSettings    *InboundDokodemoSettingsModel    `tfsdk:"dokodemo_settings"`
	HysteriaSettings    *InboundHysteriaSettingsModel    `tfsdk:"hysteria_settings"`

	// Stream settings (typed block)
	StreamSettings *InboundStreamSettingsModel `tfsdk:"stream_settings"`

	// Sniffing (typed block)
	Sniffing *InboundSniffingModel `tfsdk:"sniffing"`
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
				Default:     int64default.StaticInt64(0),
				Description: "Total traffic limit (bytes). 0 means unlimited.",
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
				Computed:    true,
				Default:     stringdefault.StaticString(""),
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
				Description: "Listen address.",
			},
			"port": schema.Int64Attribute{
				Required:    true,
				Description: "Listen port.",
			},
			"protocol": schema.StringAttribute{
				Required:    true,
				Description: "Protocol (vless, vmess, trojan, shadowsocks, http, socks, wireguard, tunnel, dokodemo-door, etc.).",
			},
			"tag": schema.StringAttribute{
				Computed:    true,
				Description: "Xray inbound tag (auto-generated by the panel).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: func() map[string]schema.Block {
			blocks := inboundSettingsBlockSchemas()
			blocks["stream_settings"] = inboundStreamSettingsBlockSchema()
			blocks["sniffing"] = inboundSniffingBlockSchema()
			return blocks
		}(),
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

	state, diags := inboundToModel(created, false)
	resp.Diagnostics.Append(diags...)
	// Preserve plan block presence: if the user did not specify a block in
	// config, nil it out even if the API returned data for it.  This avoids
	// the "was absent, but now present" inconsistency error from Terraform.
	alignBlocksWithPlan(state, &plan)
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

	newState, diags := inboundToModel(inbound, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Preserve block presence from prior state so that Read does not
	// introduce blocks the user never specified, which would cause
	// perpetual diffs.  Skip alignment when coming from import (where
	// protocol is not yet set in the prior state).
	if !state.Protocol.IsNull() && !state.Protocol.IsUnknown() {
		alignBlocksWithPlan(newState, &state)
	}
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

	newState, diags := inboundToModel(updated, false)
	resp.Diagnostics.Append(diags...)
	alignBlocksWithPlan(newState, &plan)
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

// ModifyPlan preserves nested Optional blocks from state when the user did not
// specify them in config.  Without this, Terraform plans to remove these blocks
// after import, producing false drift.
func (r *InboundResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Skip on create (no prior state) or destroy (no plan).
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

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

	changed := false

	// Preserve reality_settings.settings block from state when the user
	// specified reality_settings but omitted the inner settings block.
	//
	// Other stream_settings sub-blocks (tcp_settings, ws_settings, etc.) do
	// not need this — alignBlocksWithPlan nils them when the user omits them,
	// so state stays consistent.  But reality_settings.settings is a
	// sub-sub-block: alignBlocksWithPlan only checks whether reality_settings
	// itself is present (not whether its child "settings" block is), so the
	// child block survives in state and causes drift unless we preserve it
	// in the plan here.
	if plan.StreamSettings != nil &&
		plan.StreamSettings.RealitySettings != nil &&
		plan.StreamSettings.RealitySettings.Settings == nil &&
		state.StreamSettings != nil &&
		state.StreamSettings.RealitySettings != nil &&
		state.StreamSettings.RealitySettings.Settings != nil {
		plan.StreamSettings.RealitySettings.Settings = state.StreamSettings.RealitySettings.Settings
		changed = true
	}

	if changed {
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
	}
}

// ---------------------------------------------------------------------------
// Model <-> Inbound conversion
// ---------------------------------------------------------------------------

func expandInboundFromModel(m *InboundResourceModel) *Inbound {
	protocol := m.Protocol.ValueString()

	// Settings: typed blocks -> untyped map -> JSON
	var settingsJSON string
	if settingsMap := expandSettingsFromModel(protocol, m); len(settingsMap) > 0 {
		settingsJSON = buildSettingsJSON(settingsMap)
	} else {
		settingsJSON = "{}"
	}

	// Stream settings: typed block -> untyped map -> JSON
	var streamSettingsJSON string
	if ssMap := expandStreamSettingsFromModel(m.StreamSettings); len(ssMap) > 0 {
		streamSettingsJSON = buildStreamSettingsJSON(ssMap)
	} else {
		streamSettingsJSON = "{}"
	}

	// Sniffing: typed block -> untyped map -> JSON
	var sniffingJSON string
	if snMap := expandSniffingFromModel(m.Sniffing); len(snMap) > 0 {
		sniffingJSON = buildSniffingJSON(snMap)
	} else {
		sniffingJSON = "{}"
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
		Protocol:             protocol,
		Settings:             settingsJSON,
		StreamSettings:       streamSettingsJSON,
		Sniffing:             sniffingJSON,
	}
}

// inboundToModel converts an API Inbound into a Terraform model.
// When failHard is true (Read/Import), parse errors are reported as errors.
// When false (Create/Update), parse errors are reported as warnings so that
// the model with basic fields is still returned — this prevents leaving an
// unmanaged resource after a successful write to the API.
func inboundToModel(inbound *Inbound, failHard bool) (*InboundResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	m := &InboundResourceModel{
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
		Listen:               stringValueOrNull(inbound.Listen),
		Port:                 types.Int64Value(int64(inbound.Port)),
		Protocol:             types.StringValue(inbound.Protocol),
		Tag:                  types.StringValue(inbound.Tag),
	}

	addDiag := diags.AddWarning
	if failHard {
		addDiag = diags.AddError
	}

	// Settings: JSON -> untyped map -> typed model
	settingsMap, err := flattenSettingsToMap(inbound.Settings)
	if err != nil {
		addDiag("Failed to parse inbound settings", err.Error())
	} else if settingsMap != nil {
		flattenSettingsToModel(inbound.Protocol, settingsMap, m)
	}

	// Stream settings: JSON -> untyped map -> typed model
	ssMap, err := flattenStreamSettingsToMap(inbound.StreamSettings)
	if err != nil {
		addDiag("Failed to parse inbound stream_settings", err.Error())
	} else if ssMap != nil {
		m.StreamSettings = flattenStreamSettingsToModel(ssMap)
	}

	// Sniffing: JSON -> untyped map -> typed model
	snMap, err := flattenSniffingToMap(inbound.Sniffing)
	if err != nil {
		addDiag("Failed to parse inbound sniffing", err.Error())
	} else if snMap != nil {
		m.Sniffing = flattenSniffingToModel(snMap)
	}

	return m, diags
}

// alignBlocksWithPlan nils out blocks on the state that were not present in the
// plan.  This prevents the "was absent, but now present" inconsistency error
// that Terraform raises when a Computed block appears in the state but was not
// in the configuration.
func alignBlocksWithPlan(state *InboundResourceModel, plan *InboundResourceModel) {
	if plan.VlessSettings == nil {
		state.VlessSettings = nil
	}
	if plan.TrojanSettings == nil {
		state.TrojanSettings = nil
	}
	if plan.ShadowsocksSettings == nil {
		state.ShadowsocksSettings = nil
	}
	if plan.HTTPSettings == nil {
		state.HTTPSettings = nil
	}
	if plan.SocksSettings == nil {
		state.SocksSettings = nil
	}
	if plan.WireguardSettings == nil {
		state.WireguardSettings = nil
	}
	if plan.DokodemoSettings == nil {
		state.DokodemoSettings = nil
	}
	if plan.HysteriaSettings == nil {
		state.HysteriaSettings = nil
	}
	if plan.StreamSettings == nil {
		state.StreamSettings = nil
	} else if state.StreamSettings != nil {
		// Align nested stream_settings sub-blocks
		if plan.StreamSettings.RealitySettings == nil {
			state.StreamSettings.RealitySettings = nil
		} else if state.StreamSettings.RealitySettings != nil {
			if plan.StreamSettings.RealitySettings.Settings == nil {
				state.StreamSettings.RealitySettings.Settings = nil
			}
		}
		if plan.StreamSettings.TCPSettings == nil {
			state.StreamSettings.TCPSettings = nil
		}
		if plan.StreamSettings.WSSettings == nil {
			state.StreamSettings.WSSettings = nil
		}
		if plan.StreamSettings.GRPCSettings == nil {
			state.StreamSettings.GRPCSettings = nil
		}
		if plan.StreamSettings.HTTPUpgradeSettings == nil {
			state.StreamSettings.HTTPUpgradeSettings = nil
		}
		if plan.StreamSettings.XHTTPSettings == nil {
			state.StreamSettings.XHTTPSettings = nil
		}
		if plan.StreamSettings.KCPSettings == nil {
			state.StreamSettings.KCPSettings = nil
		}
		if plan.StreamSettings.HysteriaSettings == nil {
			state.StreamSettings.HysteriaSettings = nil
		}
		if plan.StreamSettings.Sockopt == nil {
			state.StreamSettings.Sockopt = nil
		}
	}
	if plan.Sniffing == nil {
		state.Sniffing = nil
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
