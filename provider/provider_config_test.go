package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestNewClientInvalidEndpoint(t *testing.T) {
	_, err := NewClient(ClientConfig{
		Endpoint: "not-a-url",
		Username: "admin",
		Password: "admin",
		Timeout:  time.Second,
	})
	if err == nil {
		t.Fatalf("expected error for invalid endpoint")
	}
}

func TestNewClientBasePathNormalization(t *testing.T) {
	client, err := NewClient(ClientConfig{
		Endpoint: "http://example.com",
		BasePath: "xui",
		Username: "admin",
		Password: "admin",
		Timeout:  time.Second,
	})
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	if client.basePath != "/xui/" {
		t.Fatalf("expected normalized base path '/xui/', got %q", client.basePath)
	}
}

func TestProviderSensitiveAttributes(t *testing.T) {
	var resp provider.SchemaResponse
	(&ThreeXUIProvider{}).Schema(context.Background(), provider.SchemaRequest{}, &resp)

	for _, name := range []string{"password", "bootstrap_password", "two_factor_code"} {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Fatalf("attribute %q missing on provider", name)
		}
		if !attr.IsSensitive() {
			t.Fatalf("attribute %q must be Sensitive", name)
		}
	}
}
