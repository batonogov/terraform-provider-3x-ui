package provider

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &NodesDataSource{}

type NodesDataSource struct {
	client *Client
}

type NodesDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Nodes types.String `tfsdk:"nodes"`
}

func NewNodesDataSource() datasource.DataSource {
	return &NodesDataSource{}
}

func (d *NodesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nodes"
}

func (d *NodesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"nodes": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "JSON array of cluster node objects (3x-ui multi-node surface, GET /panel/api/nodes). Marked Sensitive because the payload may include each node's pinnedCertSha256, which the panel returns raw. Starting with 3x-ui v3.6.0 (#5613) the apiToken is write-only and is no longer returned on GET. The array is the full node tree, including transitive sub-nodes surfaced from downstream panels (objects with id == 0 and transitive == true, which are read-only projections).",
			},
		},
	}
}

func (d *NodesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NodesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	nodes, err := d.client.GetNodes(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list cluster nodes", err.Error())
		return
	}

	// The nodes payload may include each node's pinnedCertSha256 (raw, no
	// upstream redaction); the `nodes` attribute is Sensitive so Terraform
	// never prints it. Since 3x-ui v3.6.0 (#5613) the apiToken is write-only
	// and omitted from the response. G117 flags marshaling a struct with a
	// secret-named field — exposure here is by design, matching the existing
	// threexui_inbounds pattern.
	payload, err := json.Marshal(nodes) //nolint:gosec // G117: intentional, attr is Sensitive
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal cluster nodes", err.Error())
		return
	}

	var state NodesDataSourceModel
	state.Nodes = types.StringValue(string(payload))
	if len(nodes) == 0 {
		state.ID = types.StringValue("0")
	} else {
		state.ID = types.StringValue(strconv.Itoa(nodes[0].Id))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
