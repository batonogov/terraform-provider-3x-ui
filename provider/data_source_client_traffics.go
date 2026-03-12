package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ClientTrafficsDataSource{}

type ClientTrafficsDataSource struct {
	client *Client
}

type ClientTrafficsDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Email      types.String `tfsdk:"email"`
	InboundID  types.Int64  `tfsdk:"inbound_id"`
	Enable     types.Bool   `tfsdk:"enable"`
	Up         types.Int64  `tfsdk:"up"`
	Down       types.Int64  `tfsdk:"down"`
	Total      types.Int64  `tfsdk:"total"`
	ExpiryTime types.Int64  `tfsdk:"expiry_time"`
}

func NewClientTrafficsDataSource() datasource.DataSource {
	return &ClientTrafficsDataSource{}
}

func (d *ClientTrafficsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client_traffics"
}

func (d *ClientTrafficsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves traffic statistics for a client by email. " +
			"The 3x-ui panel enforces email uniqueness per client in the client_traffics table, " +
			"so email is the canonical lookup key for this API endpoint.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal traffic record ID.",
			},
			"email": schema.StringAttribute{
				Required:    true,
				Description: "Client email to look up. Must match an existing client email in the panel.",
			},
			"inbound_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Associated inbound ID.",
			},
			"enable": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the client is enabled.",
			},
			"up": schema.Int64Attribute{
				Computed:    true,
				Description: "Upload bytes.",
			},
			"down": schema.Int64Attribute{
				Computed:    true,
				Description: "Download bytes.",
			},
			"total": schema.Int64Attribute{
				Computed:    true,
				Description: "Traffic limit in bytes.",
			},
			"expiry_time": schema.Int64Attribute{
				Computed:    true,
				Description: "Expiration timestamp (milliseconds since epoch).",
			},
		},
	}
}

func (d *ClientTrafficsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", "Expected *Client")
		return
	}
	d.client = client
}

func (d *ClientTrafficsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ClientTrafficsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	email := config.Email.ValueString()
	traffic, err := d.client.GetClientTraffics(ctx, email)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get client traffics", err.Error())
		return
	}

	state := ClientTrafficsDataSourceModel{
		ID:         types.StringValue(fmt.Sprintf("%d", traffic.ID)),
		Email:      types.StringValue(traffic.Email),
		InboundID:  types.Int64Value(int64(traffic.InboundID)),
		Enable:     types.BoolValue(traffic.Enable),
		Up:         types.Int64Value(traffic.Up),
		Down:       types.Int64Value(traffic.Down),
		Total:      types.Int64Value(traffic.Total),
		ExpiryTime: types.Int64Value(traffic.ExpiryTime),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
