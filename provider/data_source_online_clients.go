package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &OnlineClientsDataSource{}

type OnlineClientsDataSource struct {
	client *Client
}

type OnlineClientsDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Clients []string     `tfsdk:"clients"`
}

func NewOnlineClientsDataSource() datasource.DataSource {
	return &OnlineClientsDataSource{}
}

func (d *OnlineClientsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_online_clients"
}

func (d *OnlineClientsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves the list of currently online client emails.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"clients": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of online client emails.",
			},
		},
	}
}

func (d *OnlineClientsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OnlineClientsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	clients, err := d.client.GetOnlineClients(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get online clients", err.Error())
		return
	}

	var state OnlineClientsDataSourceModel
	state.Clients = clients
	state.ID = types.StringValue(strconv.Itoa(len(clients)))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
