package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns a terraform resource provider for 3x-ui.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Base URL of the 3x-ui panel, e.g. http://localhost:2053.",
			},
			"base_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "/",
				Description: "Base path configured in 3x-ui (webBasePath). Default is '/'.",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "admin",
				Description: "3x-ui username.",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "admin",
				Sensitive:   true,
				Description: "3x-ui password.",
			},
			"two_factor_code": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Optional 2FA code for login.",
			},
			"insecure_skip_verify": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Skip TLS certificate verification (useful for self-signed certs).",
			},
			"request_timeout": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "30s",
				Description: "HTTP request timeout (e.g. 30s, 1m).",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"threexui_inbound":            resourceInbound(),
			"threexui_inbound_client":     resourceInboundClient(),
			"threexui_panel_general":      resourcePanelSettings(),
			"threexui_panel_security":     resourceAccountSettings(),
			"threexui_panel_telegram":     resourceTelegramSettings(),
			"threexui_panel_subscription": resourceSubscriptionSettings(),
			"threexui_xray_basics":        resourceXrayBasics(),
			"threexui_xray_dns":           resourceXrayDNS(),
			"threexui_xray_routing":       resourceXrayRouting(),
			"threexui_xray_balancers":     resourceXrayBalancers(),
			"threexui_xray_reverse":       resourceXrayReverse(),
			"threexui_xray_outbounds":     resourceXrayOutbounds(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"threexui_inbounds":      dataSourceInbounds(),
			"threexui_server_status": dataSourceServerStatus(),
			"threexui_xray_versions": dataSourceXrayVersions(),
			"threexui_xray_config":   dataSourceXrayConfig(),
			"threexui_settings":      dataSourceSettings(),
		},
		ConfigureContextFunc: configureClient,
	}
}

func configureClient(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	timeoutStr := d.Get("request_timeout").(string)
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return nil, diag.Errorf("invalid request_timeout: %v", err)
	}

	client, err := NewClient(ClientConfig{
		Endpoint:           d.Get("endpoint").(string),
		BasePath:           d.Get("base_path").(string),
		Username:           d.Get("username").(string),
		Password:           d.Get("password").(string),
		TwoFactorCode:      d.Get("two_factor_code").(string),
		InsecureSkipVerify: d.Get("insecure_skip_verify").(bool),
		Timeout:            timeout,
	})
	if err != nil {
		return nil, diag.Errorf("client init failed: %v", err)
	}

	if err := client.Login(ctx); err != nil {
		return nil, diag.Errorf("login failed: %v", err)
	}

	return client, diags
}
