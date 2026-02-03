package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestConfigureClientInvalidTimeout(t *testing.T) {
	p := Provider()
	d := schema.TestResourceDataRaw(t, p.Schema, map[string]any{
		"endpoint":        "http://example.com",
		"username":        "admin",
		"password":        "admin",
		"request_timeout": "not-a-duration",
	})
	if _, diags := configureClient(context.Background(), d); len(diags) == 0 {
		t.Fatalf("expected diagnostics for invalid request_timeout")
	}
}

func TestConfigureClientEndpointRequiresScheme(t *testing.T) {
	p := Provider()
	d := schema.TestResourceDataRaw(t, p.Schema, map[string]any{
		"endpoint":        "localhost:2053",
		"username":        "admin",
		"password":        "admin",
		"request_timeout": "1s",
	})
	if _, diags := configureClient(context.Background(), d); len(diags) == 0 {
		t.Fatalf("expected diagnostics for endpoint without scheme")
	}
}

func TestConfigureClientBasePathNormalization(t *testing.T) {
	p := Provider()
	d := schema.TestResourceDataRaw(t, p.Schema, map[string]any{
		"endpoint":        "http://example.com",
		"base_path":       "xui",
		"username":        "admin",
		"password":        "admin",
		"request_timeout": "1s",
	})

	// Stub Login by using a local client and overriding endpoint to non-routable.
	// We only verify base_path normalization in the constructed client.
	client, err := NewClient(ClientConfig{
		Endpoint:           d.Get("endpoint").(string),
		BasePath:           d.Get("base_path").(string),
		Username:           d.Get("username").(string),
		Password:           d.Get("password").(string),
		TwoFactorCode:      "",
		InsecureSkipVerify: false,
		Timeout:            0,
	})
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	if client.basePath != "/xui/" {
		t.Fatalf("expected normalized base path '/xui/', got %q", client.basePath)
	}
}
