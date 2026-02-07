package provider

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ServerStatusDataSource{}

type ServerStatusDataSource struct {
	client *Client
}

type ServerStatusDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	JSON types.String `tfsdk:"json"`
}

func NewServerStatusDataSource() datasource.DataSource {
	return &ServerStatusDataSource{}
}

func (d *ServerStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_status"
}

func (d *ServerStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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

func (d *ServerStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServerStatusDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	status, err := d.client.GetServerStatus(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get server status", err.Error())
		return
	}

	payload, err := json.Marshal(status)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal server status", err.Error())
		return
	}

	var state ServerStatusDataSourceModel
	state.JSON = types.StringValue(string(payload))
	state.ID = types.StringValue(strconv.FormatInt(int64(len(payload)), 10))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
