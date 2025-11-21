package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/batonogov/terraform-provider-3x-ui/internal/client"
)

var (
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
)

func NewUserResource() resource.Resource {
	return &userResource{}
}

type userResource struct {
	api client.Client
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*providerData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *providerData, got %T", req.ProviderData))
		return
	}
	r.api = data.API
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Управляет клиентом (user) внутри конкретного inbound.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Terraform ID ресурса `<inbound_id>/<identifier>`.",
			},
			"inbound_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "ID inbound, которому принадлежит клиент.",
			},
			"client_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Идентификатор клиента (UUID для VLESS/VMESS; пароль для Trojan; email для Shadowsocks).",
			},
			"email": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Email пользователя (используется панелью для статистики и уведомлений).",
			},
			"password": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Пароль клиента (обязателен для Trojan).",
			},
			"security": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Метод шифрования (например, auto, aes-128-gcm).",
			},
			"flow": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Flow (например, xtls-rprx-vision).",
			},
			"limit_ip": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Максимум одновременно разрешённых IP.",
			},
			"total_gb": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Лимит трафика в гигабайтах.",
			},
			"expiry_time": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "UNIX время истечения клиента.",
			},
			"enable": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Включить/отключить клиента.",
			},
			"comment": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Комментарий.",
			},
			"sub_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Идентификатор подписки.",
			},
			"reset": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Период сброса (дни).",
			},
			"tg_id": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Telegram ID для уведомлений.",
			},
			"protocol": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Протокол inbound (определяется автоматически).",
			},
		},
	}
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userModel
	if diags := req.Plan.Get(ctx, &plan); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	inbound, err := r.api.GetInbound(ctx, int(plan.InboundID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to load inbound", err.Error())
		return
	}

	clientPayload, err := plan.toClient()
	if err != nil {
		resp.Diagnostics.AddError("Invalid client configuration", err.Error())
		return
	}

	if err := r.api.AddClient(ctx, inbound.ID, clientPayload); err != nil {
		resp.Diagnostics.AddError("Add client failed", err.Error())
		return
	}

	state, err := r.buildState(ctx, inbound.ID, inbound.Protocol, plan.ClientID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to refresh client", err.Error())
		return
	}
	state.mergePlan(plan)
	state.Protocol = types.StringValue(inbound.Protocol)
	state.ID = types.StringValue(stateID(inbound.ID, state.ClientID.ValueString()))

	if diags := resp.State.Set(ctx, &state); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userModel
	if diags := req.State.Get(ctx, &state); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	refreshed, err := r.buildState(ctx, int(state.InboundID.ValueInt64()), state.Protocol.ValueString(), state.ClientID.ValueString())
	if err != nil {
		if isNotFoundErr(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read client", err.Error())
		return
	}

	refreshed.mergePlan(state)
	refreshed.Protocol = state.Protocol
	refreshed.ID = types.StringValue(stateID(int(state.InboundID.ValueInt64()), refreshed.ClientID.ValueString()))

	if diags := resp.State.Set(ctx, &refreshed); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userModel
	if diags := req.Plan.Get(ctx, &plan); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	inbound, err := r.api.GetInbound(ctx, int(plan.InboundID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to load inbound", err.Error())
		return
	}

	payload, err := plan.toClient()
	if err != nil {
		resp.Diagnostics.AddError("Invalid client configuration", err.Error())
		return
	}

	clientKey := plan.ClientID.ValueString()
	if err := r.api.UpdateClient(ctx, inbound.ID, clientKey, payload); err != nil {
		resp.Diagnostics.AddError("Update client failed", err.Error())
		return
	}

	state, err := r.buildState(ctx, inbound.ID, inbound.Protocol, clientKey)
	if err != nil {
		resp.Diagnostics.AddError("Failed to refresh client", err.Error())
		return
	}
	state.mergePlan(plan)
	state.Protocol = types.StringValue(inbound.Protocol)
	state.ID = types.StringValue(stateID(inbound.ID, state.ClientID.ValueString()))

	if diags := resp.State.Set(ctx, &state); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userModel
	if diags := req.State.Get(ctx, &state); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	if state.InboundID.IsNull() || state.ClientID.IsNull() {
		return
	}
	if err := r.api.DeleteClient(ctx, int(state.InboundID.ValueInt64()), state.ClientID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete client failed", err.Error())
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format <inbound_id>/<client_id>")
		return
	}
	inboundID, err := strconv.Atoi(parts[0])
	if err != nil {
		resp.Diagnostics.AddError("Invalid inbound ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("inbound_id"), int64(inboundID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("client_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

type userModel struct {
	ID         types.String `tfsdk:"id"`
	InboundID  types.Int64  `tfsdk:"inbound_id"`
	ClientID   types.String `tfsdk:"client_id"`
	Email      types.String `tfsdk:"email"`
	Password   types.String `tfsdk:"password"`
	Security   types.String `tfsdk:"security"`
	Flow       types.String `tfsdk:"flow"`
	LimitIP    types.Int64  `tfsdk:"limit_ip"`
	TotalGB    types.Int64  `tfsdk:"total_gb"`
	ExpiryTime types.Int64  `tfsdk:"expiry_time"`
	Enable     types.Bool   `tfsdk:"enable"`
	Comment    types.String `tfsdk:"comment"`
	SubID      types.String `tfsdk:"sub_id"`
	Reset      types.Int64  `tfsdk:"reset"`
	TgID       types.Int64  `tfsdk:"tg_id"`
	Protocol   types.String `tfsdk:"protocol"`
}

func (m userModel) toClient() (client.InboundClient, error) {
	if m.ClientID.IsNull() || m.ClientID.ValueString() == "" {
		return client.InboundClient{}, fmt.Errorf("client_id must be provided")
	}
	if m.Email.IsNull() || m.Email.ValueString() == "" {
		return client.InboundClient{}, fmt.Errorf("email must be provided")
	}

	return client.InboundClient{
		ID:         m.ClientID.ValueString(),
		Email:      m.Email.ValueString(),
		Password:   valueStringDefault(m.Password, m.ClientID.ValueString()),
		Security:   valueString(m.Security),
		Flow:       valueString(m.Flow),
		LimitIP:    int(valueInt64(m.LimitIP)),
		TotalGB:    valueInt64(m.TotalGB),
		ExpiryTime: valueInt64(m.ExpiryTime),
		Enable:     valueBoolDefault(m.Enable, true),
		Comment:    valueString(m.Comment),
		SubID:      valueString(m.SubID),
		Reset:      int(valueInt64(m.Reset)),
		TgID:       valueInt64(m.TgID),
	}, nil
}

func (m *userModel) mergePlan(plan userModel) {
	if !plan.Email.IsNull() {
		m.Email = plan.Email
	}
	if !plan.Security.IsNull() {
		m.Security = plan.Security
	}
	if !plan.Flow.IsNull() {
		m.Flow = plan.Flow
	}
	if !plan.LimitIP.IsNull() {
		m.LimitIP = plan.LimitIP
	}
	if !plan.TotalGB.IsNull() {
		m.TotalGB = plan.TotalGB
	}
	if !plan.ExpiryTime.IsNull() {
		m.ExpiryTime = plan.ExpiryTime
	}
	if !plan.Enable.IsNull() {
		m.Enable = plan.Enable
	}
	if !plan.Comment.IsNull() {
		m.Comment = plan.Comment
	}
	if !plan.SubID.IsNull() {
		m.SubID = plan.SubID
	}
	if !plan.Reset.IsNull() {
		m.Reset = plan.Reset
	}
	if !plan.TgID.IsNull() {
		m.TgID = plan.TgID
	}
	if !plan.Password.IsNull() {
		m.Password = plan.Password
	}
}

func (r *userResource) buildState(ctx context.Context, inboundID int, protocol, clientID string) (userModel, error) {
	inbound, err := r.api.GetInbound(ctx, inboundID)
	if err != nil {
		return userModel{}, err
	}
	if protocol == "" {
		protocol = inbound.Protocol
	}

	clients, err := decodeClients(inbound.Settings)
	if err != nil {
		return userModel{}, err
	}
	target, ok := findClient(protocol, clients, clientID)
	if !ok {
		return userModel{}, fmt.Errorf("client %s not found", clientID)
	}

	return userModel{
		InboundID:  types.Int64Value(int64(inboundID)),
		ClientID:   types.StringValue(clientID),
		Email:      types.StringValue(target.Email),
		Password:   types.StringValue(target.Password),
		Security:   types.StringValue(target.Security),
		Flow:       types.StringValue(target.Flow),
		LimitIP:    types.Int64Value(int64(target.LimitIP)),
		TotalGB:    types.Int64Value(target.TotalGB),
		ExpiryTime: types.Int64Value(target.ExpiryTime),
		Enable:     types.BoolValue(target.Enable),
		Comment:    types.StringValue(target.Comment),
		SubID:      types.StringValue(target.SubID),
		Reset:      types.Int64Value(int64(target.Reset)),
		TgID:       types.Int64Value(target.TgID),
		Protocol:   types.StringValue(protocol),
	}, nil
}

func decodeClients(raw json.RawMessage) ([]client.InboundClient, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var data struct {
		Clients []client.InboundClient `json:"clients"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	return data.Clients, nil
}

func findClient(protocol string, clients []client.InboundClient, identifier string) (client.InboundClient, bool) {
	for _, c := range clients {
		if deriveClientIdentifier(protocol, c) == identifier {
			return c, true
		}
	}
	return client.InboundClient{}, false
}

func deriveClientIdentifier(protocol string, c client.InboundClient) string {
	switch strings.ToLower(protocol) {
	case "trojan":
		if c.Password != "" {
			return c.Password
		}
	case "shadowsocks":
		if c.Email != "" {
			return c.Email
		}
	}
	if c.ID != "" {
		return c.ID
	}
	return c.Email
}

func stateID(inboundID int, clientID string) string {
	return fmt.Sprintf("%d/%s", inboundID, clientID)
}

func isNotFoundErr(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "not found")
}
