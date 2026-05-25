package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestTrafficCounterModifier_Create(t *testing.T) {
	m := trafficCounterModifier{}
	resp := &planmodifier.Int64Response{
		PlanValue: types.Int64Unknown(),
	}
	req := planmodifier.Int64Request{
		StateValue: types.Int64Null(),
	}
	m.PlanModifyInt64(context.Background(), req, resp)

	// During create (null state) the plan value stays as-is (unknown).
	if !resp.PlanValue.IsUnknown() {
		t.Fatalf("create: expected plan to remain unknown, got %v", resp.PlanValue)
	}
}

func TestTrafficCounterModifier_Update(t *testing.T) {
	m := trafficCounterModifier{}
	resp := &planmodifier.Int64Response{
		PlanValue: types.Int64Value(42),
	}
	req := planmodifier.Int64Request{
		StateValue: types.Int64Value(42),
	}
	m.PlanModifyInt64(context.Background(), req, resp)

	// During update (known state) the plan must become unknown so Terraform
	// accepts any value returned by Read.
	if !resp.PlanValue.IsUnknown() {
		t.Fatalf("update: expected plan to be unknown, got %v", resp.PlanValue)
	}
}
