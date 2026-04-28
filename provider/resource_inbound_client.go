package provider

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var inboundClientMu sync.Mutex

// ---------------------------------------------------------------------------
// Interface checks
// ---------------------------------------------------------------------------

var (
	_ resource.Resource                = &InboundClientResource{}
	_ resource.ResourceWithImportState = &InboundClientResource{}
)

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

type InboundClientResourceModel struct {
	ID         types.String `tfsdk:"id"`
	InboundID  types.Int64  `tfsdk:"inbound_id"`
	ClientID   types.String `tfsdk:"client_id"`
	Email      types.String `tfsdk:"email"`
	Security   types.String `tfsdk:"security"`
	Password   types.String `tfsdk:"password"`
	Flow       types.String `tfsdk:"flow"`
	Auth       types.String `tfsdk:"auth"`
	LimitIP    types.Int64  `tfsdk:"limit_ip"`
	TotalGB    types.Int64  `tfsdk:"total_gb"`
	ExpiryTime types.Int64  `tfsdk:"expiry_time"`
	Enable     types.Bool   `tfsdk:"enable"`
	TgID       types.Int64  `tfsdk:"tg_id"`
	SubID      types.String `tfsdk:"sub_id"`
	Comment    types.String `tfsdk:"comment"`
	Reset      types.Int64  `tfsdk:"reset"`
}

// ---------------------------------------------------------------------------
// Resource
// ---------------------------------------------------------------------------

type InboundClientResource struct {
	client *Client
}

func NewInboundClientResource() resource.Resource {
	return &InboundClientResource{}
}

func (r *InboundClientResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_inbound_client"
}

func (r *InboundClientResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"inbound_id": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"client_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				Required: true,
			},
			"security": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"flow": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"auth": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Auth password for Hysteria clients.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"limit_ip": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"total_gb": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"expiry_time": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"enable": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"tg_id": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"sub_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"comment": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"reset": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *InboundClientResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *Client")
		return
	}
	r.client = c
}

// ---------------------------------------------------------------------------
// CRUD
// ---------------------------------------------------------------------------

func (r *InboundClientResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan InboundClientResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inboundID := int(plan.InboundID.ValueInt64())

	inboundClientMu.Lock()
	defer inboundClientMu.Unlock()

	if err := ensureInboundClientsKey(ctx, r.client, inboundID); err != nil {
		resp.Diagnostics.AddError("Failed to ensure clients key", err.Error())
		return
	}

	clientData := expandInboundClientFromModel(&plan)
	clientID := getClientIDFromModel(&plan, clientData)
	if clientID == "" {
		var err error
		clientID, err = newUUID()
		if err != nil {
			resp.Diagnostics.AddError("Failed to generate UUID", err.Error())
			return
		}
	}
	clientData["id"] = clientID

	// Generate subId if not present (API does not auto-generate it; the web UI does).
	if stringValue(clientData["subId"]) == "" {
		subID, err := newSubID()
		if err != nil {
			resp.Diagnostics.AddError("Failed to generate sub_id", err.Error())
			return
		}
		clientData["subId"] = subID
	}

	if err := r.client.AddInboundClient(ctx, inboundID, clientData); err != nil {
		resp.Diagnostics.AddError("Failed to add inbound client", err.Error())
		return
	}

	// Read back from API to populate state. The just-added client may not
	// be visible to a subsequent GET if SQLite is contended (issue #157),
	// so poll until it appears or the budget is exhausted.
	state, err := r.readClientStateWithRetry(ctx, inboundID, clientID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read created inbound client", err.Error())
		return
	}
	if state == nil {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *InboundClientResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var cur InboundClientResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &cur)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inboundID, clientID, err := splitInboundClientID(cur.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid resource ID", err.Error())
		return
	}

	state := r.readClientState(ctx, &resp.Diagnostics, inboundID, clientID)
	if state == nil {
		// Client not found — remove from state.
		if !resp.Diagnostics.HasError() {
			resp.State.RemoveResource(ctx)
		}
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *InboundClientResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan InboundClientResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var cur InboundClientResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &cur)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inboundID, clientID, err := splitInboundClientID(cur.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid resource ID", err.Error())
		return
	}

	inboundClientMu.Lock()
	defer inboundClientMu.Unlock()

	if err := ensureInboundClientsKey(ctx, r.client, inboundID); err != nil {
		resp.Diagnostics.AddError("Failed to ensure clients key", err.Error())
		return
	}

	clientData := expandInboundClientFromModel(&plan)
	clientData["id"] = clientID

	if err := r.client.UpdateInboundClient(ctx, inboundID, clientID, clientData); err != nil {
		resp.Diagnostics.AddError("Failed to update inbound client", err.Error())
		return
	}

	state, err := r.readClientStateWithRetry(ctx, inboundID, clientID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read updated inbound client", err.Error())
		return
	}
	if state == nil {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *InboundClientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var cur InboundClientResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &cur)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inboundID, clientID, err := splitInboundClientID(cur.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid resource ID", err.Error())
		return
	}

	inboundClientMu.Lock()
	defer inboundClientMu.Unlock()

	if err := ensureInboundClientsKey(ctx, r.client, inboundID); err != nil {
		resp.Diagnostics.AddError("Failed to ensure clients key", err.Error())
		return
	}

	if err := r.client.DeleteInboundClient(ctx, inboundID, clientID); err != nil {
		if strings.Contains(err.Error(), "no client remained in Inbound") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to delete inbound client", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

// ---------------------------------------------------------------------------
// Import
// ---------------------------------------------------------------------------

func (r *InboundClientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	inboundID, clientID, err := splitInboundClientID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected format inbound_id:client_id, got: %s", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), makeInboundClientID(inboundID, clientID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("inbound_id"), int64(inboundID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("client_id"), clientID)...)
}

// ---------------------------------------------------------------------------
// Read helper
// ---------------------------------------------------------------------------

func (r *InboundClientResource) readClientState(ctx context.Context, diags *diag.Diagnostics, inboundID int, clientID string) *InboundClientResourceModel {
	inbound, err := r.client.GetInbound(ctx, inboundID)
	if err != nil {
		diags.AddError("Failed to read inbound", err.Error())
		return nil
	}

	settings, err := parseInboundSettings(inbound.Settings)
	if err != nil {
		diags.AddError("Failed to parse inbound settings", err.Error())
		return nil
	}

	found := findClientByID(settings.Clients, clientID)
	if found == nil {
		return nil
	}

	return inboundClientToModel(inboundID, clientID, found)
}

// readClientStateWithRetry is the post-write variant of readClientState: it
// treats "client not yet visible" as a transient condition and polls until
// the row appears or the budget is exhausted (issue #157). Use it only
// after a successful AddInboundClient/UpdateInboundClient — for plain Read
// the absence of a client is meaningful and should not be retried.
func (r *InboundClientResource) readClientStateWithRetry(ctx context.Context, inboundID int, clientID string) (*InboundClientResourceModel, error) {
	var state *InboundClientResourceModel
	err := r.client.WithReadAfterWriteRetry(ctx, fmt.Sprintf("read inbound %d client %s", inboundID, clientID), func() (bool, error) {
		inbound, getErr := r.client.GetInbound(ctx, inboundID)
		if getErr != nil {
			return false, getErr
		}
		settings, parseErr := parseInboundSettings(inbound.Settings)
		if parseErr != nil {
			return false, parseErr
		}
		found := findClientByID(settings.Clients, clientID)
		if found == nil {
			return false, nil
		}
		state = inboundClientToModel(inboundID, clientID, found)
		return true, nil
	})
	if err != nil {
		return nil, err
	}
	return state, nil
}

// ---------------------------------------------------------------------------
// Expand / Flatten
// ---------------------------------------------------------------------------

func expandInboundClientFromModel(m *InboundClientResourceModel) map[string]any {
	client := map[string]any{}

	if !m.Email.IsNull() {
		client["email"] = m.Email.ValueString()
	}
	if !m.Security.IsNull() && !m.Security.IsUnknown() {
		client["security"] = m.Security.ValueString()
	}
	if !m.Password.IsNull() && !m.Password.IsUnknown() {
		client["password"] = m.Password.ValueString()
	}
	if !m.Flow.IsNull() && !m.Flow.IsUnknown() {
		client["flow"] = m.Flow.ValueString()
	}
	if !m.Auth.IsNull() && !m.Auth.IsUnknown() {
		client["auth"] = m.Auth.ValueString()
	}
	if !m.LimitIP.IsNull() && !m.LimitIP.IsUnknown() {
		client["limitIp"] = int(m.LimitIP.ValueInt64())
	}
	if !m.TotalGB.IsNull() && !m.TotalGB.IsUnknown() {
		client["totalGB"] = int(m.TotalGB.ValueInt64())
	}
	if !m.ExpiryTime.IsNull() && !m.ExpiryTime.IsUnknown() {
		client["expiryTime"] = int(m.ExpiryTime.ValueInt64())
	}
	if !m.Enable.IsNull() && !m.Enable.IsUnknown() {
		client["enable"] = m.Enable.ValueBool()
	}
	if !m.TgID.IsNull() && !m.TgID.IsUnknown() {
		client["tgId"] = int(m.TgID.ValueInt64())
	}
	if !m.SubID.IsNull() && !m.SubID.IsUnknown() {
		client["subId"] = m.SubID.ValueString()
	}
	if !m.Comment.IsNull() && !m.Comment.IsUnknown() {
		client["comment"] = m.Comment.ValueString()
	}
	if !m.Reset.IsNull() && !m.Reset.IsUnknown() {
		client["reset"] = int(m.Reset.ValueInt64())
	}
	if !m.ClientID.IsNull() && !m.ClientID.IsUnknown() {
		client["id"] = m.ClientID.ValueString()
	}

	return client
}

func inboundClientToModel(inboundID int, clientID string, client map[string]any) *InboundClientResourceModel {
	return &InboundClientResourceModel{
		ID:         types.StringValue(makeInboundClientID(inboundID, clientID)),
		InboundID:  types.Int64Value(int64(inboundID)),
		ClientID:   types.StringValue(clientID),
		Email:      types.StringValue(stringValue(client["email"])),
		Security:   types.StringValue(stringValue(client["security"])),
		Password:   types.StringValue(stringValue(client["password"])),
		Flow:       types.StringValue(stringValue(client["flow"])),
		Auth:       types.StringValue(stringValue(client["auth"])),
		LimitIP:    types.Int64Value(int64(intValue(client["limitIp"]))),
		TotalGB:    types.Int64Value(int64(intValue(client["totalGB"]))),
		ExpiryTime: types.Int64Value(int64(intValue(client["expiryTime"]))),
		Enable:     types.BoolValue(boolValue(client["enable"])),
		TgID:       types.Int64Value(int64(intValue(client["tgId"]))),
		SubID:      types.StringValue(stringValue(client["subId"])),
		Comment:    types.StringValue(stringValue(client["comment"])),
		Reset:      types.Int64Value(int64(intValue(client["reset"]))),
	}
}

func getClientIDFromModel(m *InboundClientResourceModel, client map[string]any) string {
	if !m.ClientID.IsNull() && !m.ClientID.IsUnknown() && m.ClientID.ValueString() != "" {
		return m.ClientID.ValueString()
	}
	if v := stringValue(client["id"]); v != "" {
		return v
	}
	if v := stringValue(client["password"]); v != "" {
		return v
	}
	if v := stringValue(client["auth"]); v != "" {
		return v
	}
	return ""
}

// ---------------------------------------------------------------------------
// Helper types and functions
// ---------------------------------------------------------------------------

type inboundSettings struct {
	Clients []map[string]any `json:"clients"`
}

func parseInboundSettings(settings string) (*inboundSettings, error) {
	if strings.TrimSpace(settings) == "" {
		return &inboundSettings{}, nil
	}
	var out inboundSettings
	if err := json.Unmarshal([]byte(settings), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func ensureInboundClientsKey(ctx context.Context, client *Client, inboundID int) error {
	if client == nil || inboundID == 0 {
		return nil
	}
	inbound, err := client.GetInbound(ctx, inboundID)
	if err != nil {
		return err
	}
	settings, err := ParseJSONField(inbound.Settings)
	if err != nil {
		return err
	}
	if settings == nil {
		settings = map[string]any{}
	}
	if _, ok := settings["clients"]; ok {
		return nil
	}
	settings["clients"] = []any{}
	updated, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	inbound.Settings = string(updated)
	_, err = client.UpdateInbound(ctx, inbound)
	return err
}

func findClientByID(clients []map[string]any, clientID string) map[string]any {
	for _, client := range clients {
		if stringValue(client["id"]) == clientID || stringValue(client["password"]) == clientID || stringValue(client["auth"]) == clientID || stringValue(client["email"]) == clientID {
			return client
		}
	}
	return nil
}

func makeInboundClientID(inboundID int, clientID string) string {
	return fmt.Sprintf("%d:%s", inboundID, clientID)
}

func newSubID() (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = chars[b[i]%byte(len(chars))]
	}
	return string(b), nil
}

func newUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4],
		b[4:6],
		b[6:8],
		b[8:10],
		b[10:16],
	), nil
}

func splitInboundClientID(id string) (int, string, error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid inbound_client id: %s", id)
	}
	inboundID, err := strconv.Atoi(parts[0])
	if err != nil || inboundID == 0 {
		return 0, "", fmt.Errorf("invalid inbound id: %s", id)
	}
	if parts[1] == "" {
		return 0, "", fmt.Errorf("invalid client id: %s", id)
	}
	return inboundID, parts[1], nil
}

func stringValue(v any) string {
	s, _ := v.(string)
	return s
}

func stringValueOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func intValue(v any) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case float32:
		return int(val)
	default:
		return 0
	}
}

func boolValue(v any) bool {
	b, _ := v.(bool)
	return b
}
