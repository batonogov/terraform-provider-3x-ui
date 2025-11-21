package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/batonogov/terraform-provider-3x-ui/internal/client"
)

// Ensure implementation satisfies interfaces
var (
	_ resource.Resource                = &inboundResource{}
	_ resource.ResourceWithImportState = &inboundResource{}
)

func NewInboundResource() resource.Resource {
	return &inboundResource{}
}

type inboundResource struct {
	api client.Client
}

func (r *inboundResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_inbound"
}

func (r *inboundResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *inboundResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Управляет inbound-записями панели 3x-ui. Настройки протокола и транспорта передаются в виде JSON.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Идентификатор inbound в панели.",
			},
			"remark": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Комментарий/название inbound.",
			},
			"listen": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Адрес прослушивания (по умолчанию пусто).",
			},
			"port": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Порт прослушивания (1-65535).",
			},
			"protocol": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Протокол inbound (например, vless, vmess, trojan).",
			},
			"enable": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Включить/отключить inbound.",
			},
			"settings_json": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "JSON конфигурация для конкретного протокола (`settings`).",
			},
			"stream_settings_json": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "JSON настроек транспорта (`streamSettings`).",
			},
			"sniffing_json": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "JSON блока `sniffing`.",
			},
			"tag": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Автоматически присвоенный тег inbound.",
			},
			"up": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Исходящий трафик (байты).",
			},
			"down": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Входящий трафик (байты).",
			},
			"total": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Лимит трафика (байты).",
			},
			"expiry_time": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "UNIX-время истечения inbound.",
			},
			"traffic_reset": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Правило сброса трафика (например, `never`, `daily`).",
			},
			"last_traffic_reset_time": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Последний момент сброса трафика (UNIX).",
			},
		},
	}
}

func (r *inboundResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan inboundResourceModel
	if diags := req.Plan.Get(ctx, &plan); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	payload, err := plan.toPayload()
	if err != nil {
		resp.Diagnostics.AddError("Invalid configuration", err.Error())
		return
	}

	inbound, err := r.api.CreateInbound(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Create inbound failed", err.Error())
		return
	}

	state := mapInboundToModel(inbound)
	state.mergePlan(plan)

	if diags := resp.State.Set(ctx, &state); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}

func (r *inboundResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state inboundResourceModel
	if diags := req.State.Get(ctx, &state); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if state.ID.IsNull() {
		return
	}

	inbound, err := r.api.GetInbound(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Read inbound failed", err.Error())
		return
	}

	newState := mapInboundToModel(inbound)
	newState.mergePlan(state)

	if diags := resp.State.Set(ctx, &newState); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}

func (r *inboundResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan inboundResourceModel
	if diags := req.Plan.Get(ctx, &plan); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if plan.ID.IsNull() {
		resp.Diagnostics.AddError("Missing ID", "State is missing inbound ID")
		return
	}

	payload, err := plan.toPayload()
	if err != nil {
		resp.Diagnostics.AddError("Invalid configuration", err.Error())
		return
	}

	inbound, err := r.api.UpdateInbound(ctx, int(plan.ID.ValueInt64()), payload)
	if err != nil {
		resp.Diagnostics.AddError("Update inbound failed", err.Error())
		return
	}

	state := mapInboundToModel(inbound)
	state.mergePlan(plan)

	if diags := resp.State.Set(ctx, &state); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}

func (r *inboundResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state inboundResourceModel
	if diags := req.State.Get(ctx, &state); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if state.ID.IsNull() {
		return
	}

	if err := r.api.DeleteInbound(ctx, int(state.ID.ValueInt64())); err != nil {
		resp.Diagnostics.AddError("Delete inbound failed", err.Error())
	}
}

func (r *inboundResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid inbound ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

type inboundResourceModel struct {
	ID                   types.Int64  `tfsdk:"id"`
	Remark               types.String `tfsdk:"remark"`
	Listen               types.String `tfsdk:"listen"`
	Port                 types.Int64  `tfsdk:"port"`
	Protocol             types.String `tfsdk:"protocol"`
	Enable               types.Bool   `tfsdk:"enable"`
	SettingsJSON         types.String `tfsdk:"settings_json"`
	StreamSettingsJSON   types.String `tfsdk:"stream_settings_json"`
	SniffingJSON         types.String `tfsdk:"sniffing_json"`
	Tag                  types.String `tfsdk:"tag"`
	Up                   types.Int64  `tfsdk:"up"`
	Down                 types.Int64  `tfsdk:"down"`
	Total                types.Int64  `tfsdk:"total"`
	ExpiryTime           types.Int64  `tfsdk:"expiry_time"`
	TrafficReset         types.String `tfsdk:"traffic_reset"`
	LastTrafficResetTime types.Int64  `tfsdk:"last_traffic_reset_time"`
}

func (m inboundResourceModel) toPayload() (client.InboundPayload, error) {
	var payload client.InboundPayload

	if m.Remark.IsNull() || m.Remark.ValueString() == "" {
		return payload, fmt.Errorf("remark must be provided")
	}
	if m.Protocol.IsNull() || m.Protocol.ValueString() == "" {
		return payload, fmt.Errorf("protocol must be provided")
	}
	if m.Port.IsNull() {
		return payload, fmt.Errorf("port must be provided")
	}
	portVal := m.Port.ValueInt64()
	if portVal < 1 || portVal > 65535 {
		return payload, fmt.Errorf("port must be between 1 and 65535")
	}

	settingsRaw, err := parseJSONField(m.SettingsJSON, "settings_json")
	if err != nil {
		return payload, err
	}
	streamRaw, err := parseOptionalJSON(m.StreamSettingsJSON)
	if err != nil {
		return payload, fmt.Errorf("invalid stream_settings_json: %w", err)
	}
	sniffingRaw, err := parseOptionalJSON(m.SniffingJSON)
	if err != nil {
		return payload, fmt.Errorf("invalid sniffing_json: %w", err)
	}

	payload = client.InboundPayload{
		Remark:               m.Remark.ValueString(),
		Listen:               valueString(m.Listen),
		Port:                 int(portVal),
		Protocol:             m.Protocol.ValueString(),
		Enable:               valueBoolDefault(m.Enable, true),
		Settings:             settingsRaw,
		StreamSettings:       streamRaw,
		Sniffing:             sniffingRaw,
		Up:                   valueInt64(m.Up),
		Down:                 valueInt64(m.Down),
		Total:                valueInt64(m.Total),
		ExpiryTime:           valueInt64(m.ExpiryTime),
		TrafficReset:         valueStringDefault(m.TrafficReset, "never"),
		LastTrafficResetTime: valueInt64(m.LastTrafficResetTime),
	}
	return payload, nil
}

func mapInboundToModel(inbound *client.Inbound) inboundResourceModel {
	return inboundResourceModel{
		ID:                   types.Int64Value(int64(inbound.ID)),
		Remark:               types.StringValue(inbound.Remark),
		Listen:               types.StringValue(inbound.Listen),
		Port:                 types.Int64Value(int64(inbound.Port)),
		Protocol:             types.StringValue(inbound.Protocol),
		Enable:               types.BoolValue(inbound.Enable),
		SettingsJSON:         rawMessageToStringValue(inbound.Settings),
		StreamSettingsJSON:   rawMessageToStringValue(inbound.StreamSettings),
		SniffingJSON:         rawMessageToStringValue(inbound.Sniffing),
		Tag:                  types.StringValue(inbound.Tag),
		Up:                   types.Int64Value(inbound.Up),
		Down:                 types.Int64Value(inbound.Down),
		Total:                types.Int64Value(inbound.Total),
		ExpiryTime:           types.Int64Value(inbound.ExpiryTime),
		TrafficReset:         types.StringValue(inbound.TrafficReset),
		LastTrafficResetTime: types.Int64Value(inbound.LastTrafficResetTime),
	}
}

func (m *inboundResourceModel) mergePlan(plan inboundResourceModel) {
	if plan.SettingsJSON.ValueString() != "" {
		m.SettingsJSON = plan.SettingsJSON
	}
	if plan.StreamSettingsJSON.ValueString() != "" {
		m.StreamSettingsJSON = plan.StreamSettingsJSON
	}
	if plan.SniffingJSON.ValueString() != "" {
		m.SniffingJSON = plan.SniffingJSON
	}
	if !plan.Enable.IsNull() {
		m.Enable = plan.Enable
	}
	if !plan.Total.IsNull() {
		m.Total = plan.Total
	}
	if !plan.ExpiryTime.IsNull() {
		m.ExpiryTime = plan.ExpiryTime
	}
	if !plan.TrafficReset.IsNull() {
		m.TrafficReset = plan.TrafficReset
	}
	if !plan.LastTrafficResetTime.IsNull() {
		m.LastTrafficResetTime = plan.LastTrafficResetTime
	}
}

func parseJSONField(value types.String, field string) (json.RawMessage, error) {
	if value.IsNull() || value.ValueString() == "" {
		return nil, fmt.Errorf("%s must be a valid JSON string", field)
	}
	raw := []byte(value.ValueString())
	if !json.Valid(raw) {
		return nil, fmt.Errorf("%s must contain valid JSON", field)
	}
	return json.RawMessage(raw), nil
}

func parseOptionalJSON(value types.String) (json.RawMessage, error) {
	if value.IsNull() || value.ValueString() == "" {
		return nil, nil
	}
	raw := []byte(value.ValueString())
	if !json.Valid(raw) {
		return nil, fmt.Errorf("invalid JSON")
	}
	return json.RawMessage(raw), nil
}

func valueInt64(v types.Int64) int64 {
	if v.IsNull() {
		return 0
	}
	return v.ValueInt64()
}

func valueStringDefault(v types.String, def string) string {
	if v.IsNull() || v.ValueString() == "" {
		return def
	}
	return v.ValueString()
}

func valueString(v types.String) string {
	if v.IsNull() {
		return ""
	}
	return v.ValueString()
}

func valueBoolDefault(v types.Bool, def bool) bool {
	if v.IsNull() {
		return def
	}
	return v.ValueBool()
}

func rawMessageToStringValue(raw json.RawMessage) types.String {
	if len(raw) == 0 {
		return types.StringNull()
	}
	return types.StringValue(string(raw))
}
