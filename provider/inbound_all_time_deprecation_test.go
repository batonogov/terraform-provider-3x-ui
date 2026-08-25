package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// inboundSchemaAttr fetches one top-level attribute off the threexui_inbound schema.
func inboundSchemaAttr(t *testing.T, name string) schema.Attribute {
	t.Helper()

	var resp resource.SchemaResponse
	(&InboundResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned diagnostics: %v", resp.Diagnostics)
	}

	attr, ok := resp.Schema.Attributes[name]
	if !ok {
		t.Fatalf("threexui_inbound has no %q attribute", name)
	}
	return attr
}

// all_time is a dead attribute: no 3x-ui release ever sent an `allTime` field on
// the inbound API, so it always reads 0 (#442). It stays in the schema for state
// compatibility until the next major release, but must carry a deprecation
// notice so the Registry and `terraform providers schema` flag it.
func TestInboundAllTimeIsDeprecated(t *testing.T) {
	attr := inboundSchemaAttr(t, "all_time")

	msg := attr.GetDeprecationMessage()
	if msg == "" {
		t.Fatal("all_time must carry a DeprecationMessage")
	}
	if !strings.Contains(msg, "always 0") {
		t.Errorf("DeprecationMessage should say the value is always 0, got %q", msg)
	}
	if !strings.Contains(msg, "threexui_client_traffics") {
		t.Errorf("DeprecationMessage should point at the replacement, got %q", msg)
	}
	// The field did exist upstream (v2.6.7 through v3.0.2) — the message has to
	// say it was removed, not that it never existed, or the next reader of the
	// drop migration in inbound_migration.go will think the provider is confused.
	if !strings.Contains(msg, "v3.1.0") {
		t.Errorf("DeprecationMessage should name the version that removed the field, got %q", msg)
	}
	if !strings.Contains(strings.ToLower(attr.GetDescription()), "deprecated") {
		t.Errorf("Description should mark the attribute deprecated, got %q", attr.GetDescription())
	}
}

// The live traffic counters must NOT be swept up by the deprecation.
func TestInboundTrafficCountersNotDeprecated(t *testing.T) {
	for _, name := range []string{"up", "down", "total"} {
		if msg := inboundSchemaAttr(t, name).GetDeprecationMessage(); msg != "" {
			t.Errorf("%s must not be deprecated, got %q", name, msg)
		}
	}
}
