package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// trafficCounterModifier is an Int64 plan modifier for computed traffic-counter
// fields (up, down, all_time, last_traffic_reset_time) that change continuously
// outside of Terraform.  During update it marks the planned value as unknown so
// the framework accepts whatever the Read returns, preventing
// "Provider produced inconsistent result after apply" errors.
type trafficCounterModifier struct{}

func (trafficCounterModifier) Description(_ context.Context) string {
	return "Marks traffic counter as unknown during update to accept externally-driven changes"
}

func (trafficCounterModifier) MarkdownDescription(_ context.Context) string {
	return "Marks traffic counter as unknown during update to accept externally-driven changes"
}

func (m trafficCounterModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	// Create: no prior state — use standard UseStateForUnknown behaviour.
	if req.StateValue.IsNull() {
		if resp.PlanValue.IsUnknown() {
			resp.PlanValue = req.StateValue
		}
		return
	}

	// Update: mark as unknown so Terraform accepts any value from Read.
	resp.PlanValue = types.Int64Unknown()
}
