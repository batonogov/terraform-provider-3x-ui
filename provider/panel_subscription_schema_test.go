package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestPanelSubscriptionSchemaSubUpdates(t *testing.T) {
	t.Parallel()

	attr, ok := panelSubscriptionSchema().Attributes["sub_updates"].(schema.Int64Attribute)
	if !ok {
		t.Fatal("sub_updates must be an Int64Attribute")
	}
	if !attr.IsOptional() || !attr.IsComputed() {
		t.Fatal("sub_updates must remain Optional+Computed")
	}
	if len(attr.Validators) == 0 {
		t.Fatal("sub_updates must retain its range validator")
	}
}
