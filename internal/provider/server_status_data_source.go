package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/batonogov/terraform-provider-3x-ui/internal/client"
)

var _ datasource.DataSource = &serverStatusDataSource{}

func NewServerStatusDataSource() datasource.DataSource {
	return &serverStatusDataSource{}
}

type serverStatusDataSource struct {
	api client.Client
}

type serverStatusModel struct {
	CPU        types.Float64 `tfsdk:"cpu"`
	CPUCores   types.Int64   `tfsdk:"cpu_cores"`
	LogicalPro types.Int64   `tfsdk:"logical_processors"`
	CPUSpeed   types.Float64 `tfsdk:"cpu_speed_mhz"`

	Mem  capacityModel `tfsdk:"mem"`
	Swap capacityModel `tfsdk:"swap"`
	Disk capacityModel `tfsdk:"disk"`

	Xray       xrayModel       `tfsdk:"xray"`
	Uptime     types.Int64     `tfsdk:"uptime"`
	Loads      []types.Float64 `tfsdk:"loads"`
	TCPCount   types.Int64     `tfsdk:"tcp_count"`
	UDPCount   types.Int64     `tfsdk:"udp_count"`
	NetIO      netIOModel      `tfsdk:"net_io"`
	NetTraffic netTrafficModel `tfsdk:"net_traffic"`
	PublicIP   publicIPModel   `tfsdk:"public_ip"`
	AppStats   appStatsModel   `tfsdk:"app_stats"`
}

type capacityModel struct {
	Current types.Int64 `tfsdk:"current"`
	Total   types.Int64 `tfsdk:"total"`
}

type xrayModel struct {
	State    types.String `tfsdk:"state"`
	ErrorMsg types.String `tfsdk:"error_msg"`
	Version  types.String `tfsdk:"version"`
}

type netIOModel struct {
	Up   types.Int64 `tfsdk:"up"`
	Down types.Int64 `tfsdk:"down"`
}

type netTrafficModel struct {
	Sent types.Int64 `tfsdk:"sent"`
	Recv types.Int64 `tfsdk:"recv"`
}

type publicIPModel struct {
	IPv4 types.String `tfsdk:"ipv4"`
	IPv6 types.String `tfsdk:"ipv6"`
}

type appStatsModel struct {
	Threads types.Int64 `tfsdk:"threads"`
	Mem     types.Int64 `tfsdk:"mem"`
	Uptime  types.Int64 `tfsdk:"uptime"`
}

func (d *serverStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_status"
}

func (d *serverStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*providerData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected DataSource Configure Type", "provider data has unexpected type")
		return
	}
	d.api = data.API
}

func (d *serverStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Возвращает текущий статус сервера (CPU, память, Xray, сетевые метрики).",
		Attributes: map[string]schema.Attribute{
			"cpu": schema.Float64Attribute{
				Computed:            true,
				MarkdownDescription: "Текущая загрузка CPU.",
			},
			"cpu_cores": schema.Int64Attribute{
				Computed: true,
			},
			"logical_processors": schema.Int64Attribute{
				Computed: true,
			},
			"cpu_speed_mhz": schema.Float64Attribute{
				Computed: true,
			},
			"mem": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"current": schema.Int64Attribute{Computed: true},
					"total":   schema.Int64Attribute{Computed: true},
				},
			},
			"swap": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"current": schema.Int64Attribute{Computed: true},
					"total":   schema.Int64Attribute{Computed: true},
				},
			},
			"disk": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"current": schema.Int64Attribute{Computed: true},
					"total":   schema.Int64Attribute{Computed: true},
				},
			},
			"xray": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"state":     schema.StringAttribute{Computed: true},
					"error_msg": schema.StringAttribute{Computed: true},
					"version":   schema.StringAttribute{Computed: true},
				},
			},
			"uptime": schema.Int64Attribute{Computed: true},
			"loads": schema.ListAttribute{
				ElementType: types.Float64Type,
				Computed:    true,
			},
			"tcp_count": schema.Int64Attribute{Computed: true},
			"udp_count": schema.Int64Attribute{Computed: true},
			"net_io": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"up":   schema.Int64Attribute{Computed: true},
					"down": schema.Int64Attribute{Computed: true},
				},
			},
			"net_traffic": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"sent": schema.Int64Attribute{Computed: true},
					"recv": schema.Int64Attribute{Computed: true},
				},
			},
			"public_ip": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"ipv4": schema.StringAttribute{Computed: true},
					"ipv6": schema.StringAttribute{Computed: true},
				},
			},
			"app_stats": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"threads": schema.Int64Attribute{Computed: true},
					"mem":     schema.Int64Attribute{Computed: true},
					"uptime":  schema.Int64Attribute{Computed: true},
				},
			},
		},
	}
}

func (d *serverStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.api == nil {
		resp.Diagnostics.AddError("Unconfigured provider", "The provider client is not configured")
		return
	}
	status, err := d.api.ServerStatus(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch server status", err.Error())
		return
	}

	model := convertServerStatus(status)
	if diags := resp.State.Set(ctx, &model); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}

func convertServerStatus(status *client.ServerStatus) serverStatusModel {
	model := serverStatusModel{
		CPU:        types.Float64Value(status.CPU),
		CPUCores:   types.Int64Value(int64(status.CPUCores)),
		LogicalPro: types.Int64Value(int64(status.LogicalPro)),
		CPUSpeed:   types.Float64Value(status.CPUSpeedMHz),
		Mem: capacityModel{
			Current: types.Int64Value(int64(status.Mem.Current)),
			Total:   types.Int64Value(int64(status.Mem.Total)),
		},
		Swap: capacityModel{
			Current: types.Int64Value(int64(status.Swap.Current)),
			Total:   types.Int64Value(int64(status.Swap.Total)),
		},
		Disk: capacityModel{
			Current: types.Int64Value(int64(status.Disk.Current)),
			Total:   types.Int64Value(int64(status.Disk.Total)),
		},
		Xray: xrayModel{
			State:    types.StringValue(status.Xray.State),
			ErrorMsg: types.StringValue(status.Xray.ErrorMsg),
			Version:  types.StringValue(status.Xray.Version),
		},
		Uptime:   types.Int64Value(int64(status.Uptime)),
		Loads:    floatSliceToTypes(status.Loads),
		TCPCount: types.Int64Value(int64(status.TCPCount)),
		UDPCount: types.Int64Value(int64(status.UDPCount)),
		NetIO: netIOModel{
			Up:   types.Int64Value(int64(status.NetIO.Up)),
			Down: types.Int64Value(int64(status.NetIO.Down)),
		},
		NetTraffic: netTrafficModel{
			Sent: types.Int64Value(int64(status.NetTraffic.Sent)),
			Recv: types.Int64Value(int64(status.NetTraffic.Recv)),
		},
		PublicIP: publicIPModel{
			IPv4: types.StringValue(status.PublicIP.IPv4),
			IPv6: types.StringValue(status.PublicIP.IPv6),
		},
		AppStats: appStatsModel{
			Threads: types.Int64Value(int64(status.AppStats.Threads)),
			Mem:     types.Int64Value(int64(status.AppStats.Mem)),
			Uptime:  types.Int64Value(int64(status.AppStats.Uptime)),
		},
	}
	return model
}

func floatSliceToTypes(values []float64) []types.Float64 {
	result := make([]types.Float64, len(values))
	for i, v := range values {
		result[i] = types.Float64Value(v)
	}
	return result
}
