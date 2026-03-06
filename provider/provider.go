package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &ThreeXUIProvider{}

type ThreeXUIProvider struct {
	version string
}

type ThreeXUIProviderModel struct {
	Endpoint           types.String `tfsdk:"endpoint"`
	BasePath           types.String `tfsdk:"base_path"`
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	TwoFactorCode      types.String `tfsdk:"two_factor_code"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
	RequestTimeout     types.String `tfsdk:"request_timeout"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ThreeXUIProvider{version: version}
	}
}

func (p *ThreeXUIProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "threexui"
	resp.Version = p.version
}

func (p *ThreeXUIProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Required:    true,
				Description: "Base URL of the 3x-ui panel, e.g. http://localhost:2053.",
			},
			"base_path": schema.StringAttribute{
				Optional:    true,
				Description: "Base path configured in 3x-ui (webBasePath). Default is '/'.",
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "3x-ui username.",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "3x-ui password.",
			},
			"two_factor_code": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Optional 2FA code for login.",
			},
			"insecure_skip_verify": schema.BoolAttribute{
				Optional:    true,
				Description: "Skip TLS certificate verification (useful for self-signed certs).",
			},
			"request_timeout": schema.StringAttribute{
				Optional:    true,
				Description: "HTTP request timeout (e.g. 30s, 1m).",
			},
		},
	}
}

func (p *ThreeXUIProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ThreeXUIProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := config.Endpoint.ValueString()
	basePath := "/"
	if !config.BasePath.IsNull() && !config.BasePath.IsUnknown() {
		basePath = config.BasePath.ValueString()
	}
	username := "admin"
	if !config.Username.IsNull() && !config.Username.IsUnknown() {
		username = config.Username.ValueString()
	}
	password := "admin"
	if !config.Password.IsNull() && !config.Password.IsUnknown() {
		password = config.Password.ValueString()
	}
	twoFactorCode := ""
	if !config.TwoFactorCode.IsNull() && !config.TwoFactorCode.IsUnknown() {
		twoFactorCode = config.TwoFactorCode.ValueString()
	}
	insecureSkipVerify := false
	if !config.InsecureSkipVerify.IsNull() && !config.InsecureSkipVerify.IsUnknown() {
		insecureSkipVerify = config.InsecureSkipVerify.ValueBool()
	}
	timeoutStr := "30s"
	if !config.RequestTimeout.IsNull() && !config.RequestTimeout.IsUnknown() {
		timeoutStr = config.RequestTimeout.ValueString()
	}

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid request_timeout", err.Error())
		return
	}

	client, err := NewClient(ClientConfig{
		Endpoint:           endpoint,
		BasePath:           basePath,
		Username:           username,
		Password:           password,
		TwoFactorCode:      twoFactorCode,
		InsecureSkipVerify: insecureSkipVerify,
		Timeout:            timeout,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client init failed", err.Error())
		return
	}

	if err := client.Login(ctx); err != nil {
		resp.Diagnostics.AddError("Login failed", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ThreeXUIProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewInboundResource,
		NewInboundClientResource,
		NewPanelGeneralResource,
		NewPanelSecurityResource,
		NewPanelUserResource,
		NewPanelTelegramResource,
		NewPanelSubscriptionResource,
		NewXrayBasicsResource,
		NewXrayDNSResource,
		NewXrayRoutingResource,
		NewXrayBalancersResource,
		NewXrayReverseResource,
		NewXrayOutboundsResource,
	}
}

func (p *ThreeXUIProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewInboundsDataSource,
		NewServerStatusDataSource,
		NewXrayVersionsDataSource,
		NewXrayConfigDataSource,
		NewSettingsDataSource,
		NewOnlineClientsDataSource,
	}
}
