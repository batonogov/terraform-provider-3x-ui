package provider

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/batonogov/terraform-provider-3x-ui/internal/client"
)

// Ensure Provider satisfies the expected interfaces.
var (
	_ provider.Provider = &Provider{}
)

// Provider implements the terraform-plugin-framework Provider interface.
type Provider struct {
	version string
}

// providerData is shared between resources and data sources.
type providerData struct {
	API client.Client
}

// New returns a function that instantiates the provider, used by the server entrypoint.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Provider{
			version: version,
		}
	}
}

// Metadata returns the provider type name and version.
func (p *Provider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "3xui"
	resp.Version = p.version
}

// Schema defines the provider-level configuration schema.
func (p *Provider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = providerschema.Schema{
		MarkdownDescription: "Configure access to a 3x-ui panel instance.",
		Attributes: map[string]providerschema.Attribute{
			"base_url": providerschema.StringAttribute{
				MarkdownDescription: "Base URL of the 3x-ui panel, including scheme and port.",
				Required:            true,
			},
			"username": providerschema.StringAttribute{
				MarkdownDescription: "Panel username. Required unless `session_cookie` is provided.",
				Optional:            true,
			},
			"password": providerschema.StringAttribute{
				MarkdownDescription: "Panel password. Required unless `session_cookie` is provided.",
				Optional:            true,
				Sensitive:           true,
			},
			"session_cookie": providerschema.StringAttribute{
				MarkdownDescription: "Pre-existing session cookie (`session=...`). If set, `username`/`password` can be omitted.",
				Optional:            true,
				Sensitive:           true,
			},
			"tls_skip_verify": providerschema.BoolAttribute{
				MarkdownDescription: "Skip TLS certificate verification (use only for development/self-signed certificates).",
				Optional:            true,
			},
			"request_timeout": providerschema.StringAttribute{
				MarkdownDescription: "HTTP request timeout (Go duration). Default: `30s`.",
				Optional:            true,
			},
			"poll_interval": providerschema.StringAttribute{
				MarkdownDescription: "Polling interval for long-running operations. Default: `5s`.",
				Optional:            true,
			},
			"max_retries": providerschema.Int64Attribute{
				MarkdownDescription: "Number of retry attempts for transient API errors. Default: `3`.",
				Optional:            true,
			},
		},
	}
}

// Configure validates provider settings and shares them with resources/data sources.
func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerConfig

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	providerData, err := p.buildProviderData(ctx, config)
	if err != nil {
		resp.Diagnostics.AddError("Invalid provider configuration", err.Error())
		return
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

// Resources returns provider resources. To be populated in future steps.
func (p *Provider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewInboundResource,
		NewUserResource,
	}
}

// DataSources returns provider data sources. To be populated in future steps.
func (p *Provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewServerStatusDataSource,
	}
}

func (p *Provider) buildProviderData(ctx context.Context, config providerConfig) (*providerData, error) {
	baseURLValue := config.BaseURL.ValueString()
	if baseURLValue == "" {
		return nil, fmt.Errorf("base_url must be provided")
	}

	parsedURL, err := url.Parse(baseURLValue)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base_url: %w", err)
	}

	var username *string
	if !config.Username.IsNull() && !config.Username.IsUnknown() {
		value := config.Username.ValueString()
		username = &value
	}

	var password *string
	if !config.Password.IsNull() && !config.Password.IsUnknown() {
		value := config.Password.ValueString()
		password = &value
	}

	var sessionCookie *string
	if !config.SessionCookie.IsNull() && !config.SessionCookie.IsUnknown() {
		value := config.SessionCookie.ValueString()
		sessionCookie = &value
	}

	if sessionCookie == nil {
		if username == nil || password == nil {
			return nil, fmt.Errorf("username and password must be set when session_cookie is not provided")
		}
	}

	requestTimeout := 30 * time.Second
	if !config.RequestTimeout.IsNull() && !config.RequestTimeout.IsUnknown() && config.RequestTimeout.ValueString() != "" {
		dur, err := time.ParseDuration(config.RequestTimeout.ValueString())
		if err != nil {
			return nil, fmt.Errorf("invalid request_timeout: %w", err)
		}
		requestTimeout = dur
	}

	pollInterval := 5 * time.Second
	if !config.PollInterval.IsNull() && !config.PollInterval.IsUnknown() && config.PollInterval.ValueString() != "" {
		dur, err := time.ParseDuration(config.PollInterval.ValueString())
		if err != nil {
			return nil, fmt.Errorf("invalid poll_interval: %w", err)
		}
		pollInterval = dur
	}

	maxRetries := int64(3)
	if !config.MaxRetries.IsNull() && !config.MaxRetries.IsUnknown() {
		maxRetries = config.MaxRetries.ValueInt64()
	}

	tlsSkip := false
	if !config.TLSSkipVerify.IsNull() && !config.TLSSkipVerify.IsUnknown() {
		tlsSkip = config.TLSSkipVerify.ValueBool()
	}

	apiClient, err := client.New(client.Config{
		BaseURL:        parsedURL,
		Username:       username,
		Password:       password,
		SessionCookie:  sessionCookie,
		TLSSkipVerify:  tlsSkip,
		RequestTimeout: requestTimeout,
		MaxRetries:     int(maxRetries),
		PollInterval:   pollInterval,
		UserAgent:      fmt.Sprintf("terraform-provider-3x-ui/%s", p.version),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize API client: %w", err)
	}

	if err := apiClient.Ready(ctx); err != nil {
		return nil, fmt.Errorf("failed to authenticate against 3x-ui: %w", err)
	}

	return &providerData{
		API: apiClient,
	}, nil
}

// providerConfig mirrors the provider schema for decoding configuration values.
type providerConfig struct {
	BaseURL        types.String `tfsdk:"base_url"`
	Username       types.String `tfsdk:"username"`
	Password       types.String `tfsdk:"password"`
	SessionCookie  types.String `tfsdk:"session_cookie"`
	TLSSkipVerify  types.Bool   `tfsdk:"tls_skip_verify"`
	RequestTimeout types.String `tfsdk:"request_timeout"`
	PollInterval   types.String `tfsdk:"poll_interval"`
	MaxRetries     types.Int64  `tfsdk:"max_retries"`
}
