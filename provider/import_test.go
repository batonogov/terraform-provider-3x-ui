package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestInboundImporter(t *testing.T) {
	r := resourceInbound()
	if r.Importer == nil || r.Importer.StateContext == nil {
		t.Fatalf("inbound importer not configured")
	}
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{})
	d.SetId("123")
	if _, err := r.Importer.StateContext(nil, d, nil); err != nil {
		t.Fatalf("import state failed: %v", err)
	}
}

func TestInboundClientImporter(t *testing.T) {
	r := resourceInboundClient()
	if r.Importer == nil || r.Importer.StateContext == nil {
		t.Fatalf("inbound_client importer not configured")
	}
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{})
	d.SetId("10:client-id")
	if _, err := r.Importer.StateContext(nil, d, nil); err != nil {
		t.Fatalf("import state failed: %v", err)
	}
}
