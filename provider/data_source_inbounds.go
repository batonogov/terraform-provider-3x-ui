package provider

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &InboundsDataSource{}

type InboundsDataSource struct {
	client *Client
}

type InboundsDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Inbounds types.String `tfsdk:"inbounds"`
}

func NewInboundsDataSource() datasource.DataSource {
	return &InboundsDataSource{}
}

func (d *InboundsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_inbounds"
}

func (d *InboundsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"inbounds": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "JSON array of inbound objects. Marked Sensitive because the payload includes client UUIDs/passwords, Reality privateKey, WireGuard secretKey, and similar credentials.",
			},
		},
	}
}

func (d *InboundsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *InboundsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	inbounds, err := d.client.GetInbounds(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list inbounds", err.Error())
		return
	}

	payload, err := json.Marshal(inbounds)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal inbounds", err.Error())
		return
	}

	var state InboundsDataSourceModel
	state.Inbounds = types.StringValue(string(payload))
	if len(inbounds) == 0 {
		state.ID = types.StringValue("0")
	} else {
		state.ID = types.StringValue(strconv.Itoa(inbounds[0].ID))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
