package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &XrayVersionsDataSource{}

type XrayVersionsDataSource struct {
	client *Client
}

type XrayVersionsDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Versions types.List   `tfsdk:"versions"`
}

func NewXrayVersionsDataSource() datasource.DataSource {
	return &XrayVersionsDataSource{}
}

func (d *XrayVersionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_xray_versions"
}

func (d *XrayVersionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"versions": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *XrayVersionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *XrayVersionsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	versions, err := d.client.GetXrayVersions(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get xray versions", err.Error())
		return
	}

	elems := make([]types.String, 0, len(versions))
	for _, v := range versions {
		elems = append(elems, types.StringValue(v))
	}
	listVal, diags := types.ListValueFrom(ctx, types.StringType, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state XrayVersionsDataSourceModel
	state.Versions = listVal
	state.ID = types.StringValue(strconv.Itoa(len(versions)))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
