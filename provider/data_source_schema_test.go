package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// TestDataSourceSensitiveAttributes guards the project security rule that any
// data source returning a raw JSON payload from the panel/Xray API must mark its
// JSON attribute Sensitive. Without this assertion a future refactor could
// silently drop the flag and re-expose client UUIDs/passwords, Reality keys,
// WireGuard secrets, Telegram tokens, etc. in plan output.
func TestDataSourceSensitiveAttributes(t *testing.T) {
	cases := []struct {
		name      string
		factory   func() datasource.DataSource
		attribute string
	}{
		{"inbounds", NewInboundsDataSource, "inbounds"},
		{"nodes", NewNodesDataSource, "nodes"},
		{"settings", NewSettingsDataSource, "json"},
		{"xray_config", NewXrayConfigDataSource, "json"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var resp datasource.SchemaResponse
			tc.factory().Schema(context.Background(), datasource.SchemaRequest{}, &resp)

			attr, ok := resp.Schema.Attributes[tc.attribute]
			if !ok {
				t.Fatalf("attribute %q missing on data source", tc.attribute)
			}
			if !attr.IsSensitive() {
				t.Fatalf("attribute %q must be Sensitive", tc.attribute)
			}
		})
	}
}
