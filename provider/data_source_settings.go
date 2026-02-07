package provider

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SettingsDataSource{}

type SettingsDataSource struct {
	client *Client
}

type SettingsDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	JSON types.String `tfsdk:"json"`
}

func NewSettingsDataSource() datasource.DataSource {
	return &SettingsDataSource{}
}

func (d *SettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_settings"
}

func (d *SettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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

func (d *SettingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SettingsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	settings, err := d.client.GetSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get settings", err.Error())
		return
	}
	payload, err := json.Marshal(settings)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal settings", err.Error())
		return
	}

	var state SettingsDataSourceModel
	state.JSON = types.StringValue(string(payload))
	state.ID = types.StringValue(strconv.FormatInt(int64(len(payload)), 10))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
