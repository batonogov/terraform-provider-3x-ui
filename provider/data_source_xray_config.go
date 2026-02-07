package provider

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &XrayConfigDataSource{}

type XrayConfigDataSource struct {
	client *Client
}

type XrayConfigDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	JSON types.String `tfsdk:"json"`
}

func NewXrayConfigDataSource() datasource.DataSource {
	return &XrayConfigDataSource{}
}

func (d *XrayConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_xray_config"
}

func (d *XrayConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"json": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *XrayConfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *XrayConfigDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	config, err := d.client.GetXrayConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get xray config", err.Error())
		return
	}
	payload, err := json.Marshal(config)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal xray config", err.Error())
		return
	}

	var state XrayConfigDataSourceModel
	state.JSON = types.StringValue(string(payload))
	state.ID = types.StringValue(strconv.FormatInt(int64(len(payload)), 10))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
