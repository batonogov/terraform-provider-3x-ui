package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourcepath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var xrayTemplateMu sync.Mutex

type xraySectionMode int

const (
	xraySectionMergeRoot xraySectionMode = iota
	xraySectionSetPath
)

type xraySection struct {
	id   string
	mode xraySectionMode
	path []string
}

var (
	xraySectionBasics           = xraySection{id: "xray_basics", mode: xraySectionMergeRoot}
	xraySectionDNS              = xraySection{id: "xray_dns", mode: xraySectionSetPath, path: []string{"dns"}}
	xraySectionRouting          = xraySection{id: "xray_routing", mode: xraySectionSetPath, path: []string{"routing"}}
	xraySectionBalancers        = xraySection{id: "xray_balancers", mode: xraySectionSetPath, path: []string{"routing", "balancers"}}
	xraySectionReverse          = xraySection{id: "xray_reverse", mode: xraySectionSetPath, path: []string{"reverse"}}
	xraySectionOutbounds        = xraySection{id: "xray_outbounds", mode: xraySectionSetPath, path: []string{"outbounds"}}
	xraySectionObservatory      = xraySection{id: "xray_observatory", mode: xraySectionSetPath, path: []string{"observatory"}}
	xraySectionBurstObservatory = xraySection{id: "xray_observatory", mode: xraySectionSetPath, path: []string{"burstObservatory"}}
)

type xrayFlattenFunc func(data any) map[string]any

// ---------------------------------------------------------------------------
// Shared typed CRUD helpers
// ---------------------------------------------------------------------------

// xrayApplyTyped applies the desired value to the xray template.
// If desired is empty (empty map or empty/nil slice), the apply is skipped
// to avoid overwriting an existing section with an empty value.
//
// This means a user cannot "clear" a section by supplying an empty block in
// Terraform — the empty value is treated as a no-op. The Delete operation
// only removes the resource from Terraform state; it does not reset the
// corresponding section in the 3x-ui xray config. To clear a section
// manually, edit the xray template directly in the 3x-ui panel.
func xrayApplyTyped(
	ctx context.Context,
	desired any,
	diags *diag.Diagnostics,
	client *Client,
	section xraySection,
) {
	if isEmptyXrayValue(desired) {
		return
	}

	xrayTemplateMu.Lock()
	defer xrayTemplateMu.Unlock()

	current, err := client.GetXrayTemplate(ctx)
	if err != nil {
		diags.AddError("Failed to get xray template", err.Error())
		return
	}

	updated, err := applyXraySection(current, desired, section)
	if err != nil {
		diags.AddError("Failed to apply xray section", err.Error())
		return
	}

	if err := client.UpdateXrayTemplate(ctx, updated); err != nil {
		diags.AddError("Failed to update xray template", err.Error())
		return
	}
}

// xrayReadSection reads the xray template and extracts+flattens the specified section.
func xrayReadSection(
	ctx context.Context,
	diags *diag.Diagnostics,
	client *Client,
	section xraySection,
	flatten xrayFlattenFunc,
) map[string]any {
	current, err := client.GetXrayTemplate(ctx)
	if err != nil {
		diags.AddError("Failed to get xray template", err.Error())
		return nil
	}

	value := extractXraySection(current, section)
	return flatten(value)
}

// ---------------------------------------------------------------------------
// XrayBasicsResource
// ---------------------------------------------------------------------------

var (
	_ resource.Resource                = &XrayBasicsResource{}
	_ resource.ResourceWithImportState = &XrayBasicsResource{}
)

type XrayBasicsResource struct{ client *Client }

func NewXrayBasicsResource() resource.Resource { return &XrayBasicsResource{} }

func (r *XrayBasicsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_xray_basics"
}

func (r *XrayBasicsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = xrayBasicsSchema()
}

func (r *XrayBasicsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *Client")
		return
	}
	r.client = c
}

func (r *XrayBasicsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan XrayBasicsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := expandXrayBasics(&plan)
	desired := buildXrayBasicsJSON(input)
	xrayApplyTyped(ctx, desired, &resp.Diagnostics, r.client, xraySectionBasics)
	if resp.Diagnostics.HasError() {
		return
	}

	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionBasics, flattenXrayBasicsToMap)
	if flat == nil {
		return
	}
	state := flattenXrayBasics(flat)
	alignBasicsBlocksWithPlan(state, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *XrayBasicsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var prior XrayBasicsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}

	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionBasics, flattenXrayBasicsToMap)
	if flat == nil {
		return
	}
	state := flattenXrayBasics(flat)
	// Skip alignment during import (prior state has no blocks set).
	if prior.ID.IsNull() || prior.ID.IsUnknown() {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
		return
	}
	alignBasicsBlocksWithPlan(state, &prior)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *XrayBasicsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan XrayBasicsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := expandXrayBasics(&plan)
	desired := buildXrayBasicsJSON(input)
	xrayApplyTyped(ctx, desired, &resp.Diagnostics, r.client, xraySectionBasics)
	if resp.Diagnostics.HasError() {
		return
	}

	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionBasics, flattenXrayBasicsToMap)
	if flat == nil {
		return
	}
	state := flattenXrayBasics(flat)
	alignBasicsBlocksWithPlan(state, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *XrayBasicsResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *XrayBasicsResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionBasics, flattenXrayBasicsToMap)
	if flat == nil {
		return
	}
	state := flattenXrayBasics(flat)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// ---------------------------------------------------------------------------
// XrayDNSResource
// ---------------------------------------------------------------------------

var (
	_ resource.Resource                = &XrayDNSResource{}
	_ resource.ResourceWithImportState = &XrayDNSResource{}
)

type XrayDNSResource struct{ client *Client }

func NewXrayDNSResource() resource.Resource { return &XrayDNSResource{} }

func (r *XrayDNSResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_xray_dns"
}

func (r *XrayDNSResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = xrayDNSSchema()
}

func (r *XrayDNSResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *Client")
		return
	}
	r.client = c
}

func (r *XrayDNSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan XrayDNSModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := expandXrayDNS(&plan)
	desired := buildXrayDNSJSON(input)
	xrayApplyTyped(ctx, desired, &resp.Diagnostics, r.client, xraySectionDNS)
	if resp.Diagnostics.HasError() {
		return
	}

	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionDNS, flattenXrayDNSToMap)
	if flat == nil {
		return
	}
	state := flattenXrayDNS(flat)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *XrayDNSResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionDNS, flattenXrayDNSToMap)
	if flat == nil {
		return
	}
	state := flattenXrayDNS(flat)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *XrayDNSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan XrayDNSModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := expandXrayDNS(&plan)
	desired := buildXrayDNSJSON(input)
	xrayApplyTyped(ctx, desired, &resp.Diagnostics, r.client, xraySectionDNS)
	if resp.Diagnostics.HasError() {
		return
	}

	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionDNS, flattenXrayDNSToMap)
	if flat == nil {
		return
	}
	state := flattenXrayDNS(flat)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *XrayDNSResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *XrayDNSResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionDNS, flattenXrayDNSToMap)
	if flat == nil {
		return
	}
	state := flattenXrayDNS(flat)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// ---------------------------------------------------------------------------
// XrayRoutingResource
// ---------------------------------------------------------------------------

var (
	_ resource.Resource                = &XrayRoutingResource{}
	_ resource.ResourceWithImportState = &XrayRoutingResource{}
	_ resource.ResourceWithModifyPlan  = &XrayRoutingResource{}
)

type XrayRoutingResource struct{ client *Client }

func NewXrayRoutingResource() resource.Resource { return &XrayRoutingResource{} }

func (r *XrayRoutingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_xray_routing"
}

func (r *XrayRoutingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = xrayRoutingSchema()
}

func (r *XrayRoutingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *Client")
		return
	}
	r.client = c
}

func (r *XrayRoutingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan XrayRoutingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ensureNoAPIRoutingRules(plan.Rule, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	input := expandXrayRouting(&plan)
	desired := buildXrayRoutingJSON(input)
	xrayApplyTyped(ctx, desired, &resp.Diagnostics, r.client, xraySectionRouting)
	if resp.Diagnostics.HasError() {
		return
	}

	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionRouting, flattenXrayRoutingToMap)
	if flat == nil {
		return
	}
	state := flattenXrayRouting(flat)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *XrayRoutingResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionRouting, flattenXrayRoutingToMap)
	if flat == nil {
		return
	}
	state := flattenXrayRouting(flat)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *XrayRoutingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan XrayRoutingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ensureNoAPIRoutingRules(plan.Rule, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	input := expandXrayRouting(&plan)
	desired := buildXrayRoutingJSON(input)
	xrayApplyTyped(ctx, desired, &resp.Diagnostics, r.client, xraySectionRouting)
	if resp.Diagnostics.HasError() {
		return
	}

	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionRouting, flattenXrayRoutingToMap)
	if flat == nil {
		return
	}
	state := flattenXrayRouting(flat)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *XrayRoutingResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *XrayRoutingResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionRouting, flattenXrayRoutingToMap)
	if flat == nil {
		return
	}
	state := flattenXrayRouting(flat)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// ModifyPlan keeps the practitioner's configuration authoritative for the
// routing rule list, defeating a stale carry-forward bug in the schema plan
// modifiers.
//
// The nested `rule` block attributes are Optional + Computed with
// stringplanmodifier/listplanmodifier.UseStateForUnknown (added by #228 to
// silence "(known after apply)" drift after import). Because ListNestedBlock
// elements are matched by index, that modifier copies the prior rule's unset
// fields into the new rule occupying the same index whenever rules are
// removed or reordered — bleeding stale matchers across rules. On a reorder
// from [private→direct, RU-domains→direct, catch→proxy] to
// [RU-domains→direct, geoip:cn→direct, catch→proxy] the carried fields put
// `ip:[geoip:private]` on the RU-domains rule and `network:"tcp,udp"` on the
// geoip:cn rule, then those merged rules get written to the panel on apply.
//
// Overriding the planned `rule` list with the configured one removes the
// carry-forward while leaving truly unchanged rules (config == plan)
// untouched. The panel echoes the routing template verbatim on save and the
// provider filters the panel-managed `api` and `xui-dns-allow` rules on read,
// so there are no legitimate computed-only routing fields: configuration is
// always authoritative and this override never drops a real value.
//
// Create (no prior state) and Delete (null plan) have nothing to reconcile.
// When the configured `rule` collection is itself unknown (e.g. a computed
// `dynamic` block whose `for_each` is not yet known), defer to the schema plan
// modifiers — decoding an unknown list into `[]XrayRoutingRule` raises a
// Value Conversion Error (hashicorp/terraform-plugin-framework#1025).
func (r *XrayRoutingResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}
	// Read `rule` as types.List to tolerate an unknown collection. Decoding
	// into []XrayRoutingRule via req.Config.Get would fail with a Value
	// Conversion Error if the whole list is unknown.
	var ruleAttr types.List
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, resourcepath.Root("rule"), &ruleAttr)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if ruleAttr.IsUnknown() {
		return
	}
	var plan, config XrayRoutingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if reconcileRoutingPlan(&plan, config) {
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
	}
}

// reconcileRoutingPlan rewrites the planned routing rules to mirror the
// configuration, returning whether the plan changed. It is the unit-testable
// core of XrayRoutingResource.ModifyPlan: when the plan the schema plan
// modifiers produced carries stale per-index state that configuration does
// not, override it with the configured rule list.
func reconcileRoutingPlan(plan *XrayRoutingModel, config XrayRoutingModel) bool {
	if reflect.DeepEqual(plan.Rule, config.Rule) {
		return false
	}
	plan.Rule = config.Rule
	return true
}

// ---------------------------------------------------------------------------
// XrayBalancersResource
// ---------------------------------------------------------------------------

var (
	_ resource.Resource                = &XrayBalancersResource{}
	_ resource.ResourceWithImportState = &XrayBalancersResource{}
)

type XrayBalancersResource struct{ client *Client }

func NewXrayBalancersResource() resource.Resource { return &XrayBalancersResource{} }

func (r *XrayBalancersResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_xray_balancers"
}

func (r *XrayBalancersResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = xrayBalancersSchema()
}

func (r *XrayBalancersResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *Client")
		return
	}
	r.client = c
}

func (r *XrayBalancersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan XrayBalancersModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := expandXrayBalancers(&plan)
	desired := buildXrayBalancersJSON(input)
	xrayApplyTyped(ctx, desired, &resp.Diagnostics, r.client, xraySectionBalancers)
	if resp.Diagnostics.HasError() {
		return
	}

	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionBalancers, flattenXrayBalancersToMap)
	if flat == nil {
		return
	}
	state := flattenXrayBalancers(flat)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *XrayBalancersResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionBalancers, flattenXrayBalancersToMap)
	if flat == nil {
		return
	}
	state := flattenXrayBalancers(flat)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *XrayBalancersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan XrayBalancersModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := expandXrayBalancers(&plan)
	desired := buildXrayBalancersJSON(input)
	xrayApplyTyped(ctx, desired, &resp.Diagnostics, r.client, xraySectionBalancers)
	if resp.Diagnostics.HasError() {
		return
	}

	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionBalancers, flattenXrayBalancersToMap)
	if flat == nil {
		return
	}
	state := flattenXrayBalancers(flat)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *XrayBalancersResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *XrayBalancersResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionBalancers, flattenXrayBalancersToMap)
	if flat == nil {
		return
	}
	state := flattenXrayBalancers(flat)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// ---------------------------------------------------------------------------
// XrayReverseResource
// ---------------------------------------------------------------------------

var (
	_ resource.Resource                = &XrayReverseResource{}
	_ resource.ResourceWithImportState = &XrayReverseResource{}
)

type XrayReverseResource struct{ client *Client }

func NewXrayReverseResource() resource.Resource { return &XrayReverseResource{} }

func (r *XrayReverseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_xray_reverse"
}

func (r *XrayReverseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = xrayReverseSchema()
}

func (r *XrayReverseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *Client")
		return
	}
	r.client = c
}

func (r *XrayReverseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan XrayReverseModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := expandXrayReverse(&plan)
	desired := buildXrayReverseJSON(input)
	xrayApplyTyped(ctx, desired, &resp.Diagnostics, r.client, xraySectionReverse)
	if resp.Diagnostics.HasError() {
		return
	}

	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionReverse, flattenXrayReverseToMap)
	if flat == nil {
		return
	}
	state := flattenXrayReverse(flat)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *XrayReverseResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionReverse, flattenXrayReverseToMap)
	if flat == nil {
		return
	}
	state := flattenXrayReverse(flat)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *XrayReverseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan XrayReverseModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := expandXrayReverse(&plan)
	desired := buildXrayReverseJSON(input)
	xrayApplyTyped(ctx, desired, &resp.Diagnostics, r.client, xraySectionReverse)
	if resp.Diagnostics.HasError() {
		return
	}

	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionReverse, flattenXrayReverseToMap)
	if flat == nil {
		return
	}
	state := flattenXrayReverse(flat)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *XrayReverseResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *XrayReverseResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionReverse, flattenXrayReverseToMap)
	if flat == nil {
		return
	}
	state := flattenXrayReverse(flat)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// ---------------------------------------------------------------------------
// XrayOutboundsResource
// ---------------------------------------------------------------------------

var (
	_ resource.Resource                = &XrayOutboundsResource{}
	_ resource.ResourceWithImportState = &XrayOutboundsResource{}
)

type XrayOutboundsResource struct{ client *Client }

func NewXrayOutboundsResource() resource.Resource { return &XrayOutboundsResource{} }

func (r *XrayOutboundsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_xray_outbounds"
}

func (r *XrayOutboundsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = xrayOutboundsSchema()
}

func (r *XrayOutboundsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *Client")
		return
	}
	r.client = c
}

func (r *XrayOutboundsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan XrayOutboundsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := expandXrayOutbounds(&plan)
	desired := buildXrayOutboundsJSON(input)
	xrayApplyTyped(ctx, desired, &resp.Diagnostics, r.client, xraySectionOutbounds)
	if resp.Diagnostics.HasError() {
		return
	}

	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionOutbounds, flattenXrayOutboundsToMap)
	if flat == nil {
		return
	}
	state := flattenXrayOutbounds(flat)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *XrayOutboundsResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionOutbounds, flattenXrayOutboundsToMap)
	if flat == nil {
		return
	}
	state := flattenXrayOutbounds(flat)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *XrayOutboundsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan XrayOutboundsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := expandXrayOutbounds(&plan)
	desired := buildXrayOutboundsJSON(input)
	xrayApplyTyped(ctx, desired, &resp.Diagnostics, r.client, xraySectionOutbounds)
	if resp.Diagnostics.HasError() {
		return
	}

	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionOutbounds, flattenXrayOutboundsToMap)
	if flat == nil {
		return
	}
	state := flattenXrayOutbounds(flat)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *XrayOutboundsResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *XrayOutboundsResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	flat := xrayReadSection(ctx, &resp.Diagnostics, r.client, xraySectionOutbounds, flattenXrayOutboundsToMap)
	if flat == nil {
		return
	}
	state := flattenXrayOutbounds(flat)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// ---------------------------------------------------------------------------
// XrayObservatoryResource
// ---------------------------------------------------------------------------

var (
	_ resource.Resource                = &XrayObservatoryResource{}
	_ resource.ResourceWithImportState = &XrayObservatoryResource{}
)

type XrayObservatoryResource struct{ client *Client }

func NewXrayObservatoryResource() resource.Resource { return &XrayObservatoryResource{} }

func (r *XrayObservatoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_xray_observatory"
}

func (r *XrayObservatoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = xrayObservatorySchema()
}

func (r *XrayObservatoryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *Client")
		return
	}
	r.client = c
}

// xrayObservatoryApply writes both the "observatory" and "burstObservatory"
// top-level keys. Unlike other xray resources that use a single xraySection
// path, this resource manages two independent top-level JSON keys. It performs
// two mutex-protected read-modify-write cycles (one per key), keeping the
// existing xrayTemplateMu serialization.
func (r *XrayObservatoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan XrayObservatoryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.applyObservatory(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := r.readObservatory(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || state == nil {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *XrayObservatoryResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	state := r.readObservatory(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || state == nil {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *XrayObservatoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan XrayObservatoryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.applyObservatory(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := r.readObservatory(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || state == nil {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *XrayObservatoryResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *XrayObservatoryResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	state := r.readObservatory(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || state == nil {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// applyObservatory writes the observatory and burstObservatory sections.
// Each is applied as a set-path operation on the top-level JSON key.
//
// xrayApplyTyped stores its `desired` argument verbatim at the section path via
// setJSONPath(root, section.path, desired). For set-path sections the desired
// value must therefore be the section CONTENT (e.g. the tag-keyed object), not a
// map re-wrapped with the key — wrapping it again would double-nest the value
// (root["observatory"] = {"observatory": {...}}), which the read-back path then
// cannot decode, leaving subject_selector as an invalid zero-value types.List
// and raising a framework Value Conversion Error.
func (r *XrayObservatoryResource) applyObservatory(ctx context.Context, plan *XrayObservatoryModel, diags *diag.Diagnostics) {
	input := expandXrayObservatory(plan)
	desired := buildXrayObservatoryJSON(input).(map[string]any)

	// Apply observatory key. `obs` is already the tag-keyed object to store at
	// path ["observatory"], so pass it directly.
	if obs, ok := desired["observatory"]; ok && !isEmptyXrayValue(obs) {
		xrayApplyTyped(ctx, obs, diags, r.client, xraySectionObservatory)
	} else {
		// If user removed all observatories, clear the key
		r.clearObservatoryKey(ctx, "observatory", diags)
	}

	if diags.HasError() {
		return
	}

	// Apply burstObservatory key. `burst` is the tag-keyed object for path
	// ["burstObservatory"], passed directly (same rationale as above).
	if burst, ok := desired["burstObservatory"]; ok && !isEmptyXrayValue(burst) {
		xrayApplyTyped(ctx, burst, diags, r.client, xraySectionBurstObservatory)
	} else {
		r.clearObservatoryKey(ctx, "burstObservatory", diags)
	}
}

// clearObservatoryKey removes a top-level key from the xray template.
func (r *XrayObservatoryResource) clearObservatoryKey(ctx context.Context, key string, diags *diag.Diagnostics) {
	xrayTemplateMu.Lock()
	defer xrayTemplateMu.Unlock()

	current, err := r.client.GetXrayTemplate(ctx)
	if err != nil {
		diags.AddError("Failed to get xray template", err.Error())
		return
	}

	if _, exists := current[key]; exists {
		delete(current, key)
		if err := r.client.UpdateXrayTemplate(ctx, current); err != nil {
			diags.AddError("Failed to update xray template", err.Error())
		}
	}
}

// readObservatory reads both observatory and burstObservatory from the template.
func (r *XrayObservatoryResource) readObservatory(ctx context.Context, diags *diag.Diagnostics) *XrayObservatoryModel {
	current, err := r.client.GetXrayTemplate(ctx)
	if err != nil {
		diags.AddError("Failed to get xray template", err.Error())
		return nil
	}

	// Build a combined payload for flattenXrayObservatoryToMap
	payload := map[string]any{}
	if v, ok := current["observatory"]; ok {
		payload["observatory"] = v
	}
	if v, ok := current["burstObservatory"]; ok {
		payload["burstObservatory"] = v
	}

	flat := flattenXrayObservatoryToMap(payload)
	return flattenXrayObservatory(flat)
}

// ---------------------------------------------------------------------------
// JSON helpers (no SDK dependency)
// ---------------------------------------------------------------------------

func deepEqualJSON(a, b any) bool {
	ab, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bb, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return bytes.Equal(ab, bb)
}

func applyXraySection(current map[string]any, desired any, section xraySection) (map[string]any, error) {
	root := cloneJSONMap(current)
	switch section.mode {
	case xraySectionMergeRoot:
		desiredMap, ok := desired.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("value must be an object for %s", section.id)
		}
		root = deepMergeJSON(root, desiredMap)
	case xraySectionSetPath:
		if len(section.path) == 0 {
			return nil, fmt.Errorf("invalid section path for %s", section.id)
		}
		setJSONPath(root, section.path, desired)
	default:
		return nil, fmt.Errorf("unknown section mode for %s", section.id)
	}
	return root, nil
}

func extractXraySection(current map[string]any, section xraySection) any {
	switch section.mode {
	case xraySectionMergeRoot:
		out := map[string]any{}
		for _, key := range []string{"log", "policy", "api", "stats", "env"} {
			if v, ok := current[key]; ok {
				out[key] = v
			}
		}
		return out
	case xraySectionSetPath:
		return getJSONPath(current, section.path)
	default:
		return nil
	}
}

func cloneJSONMap(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func deepMergeJSON(dst, src map[string]any) map[string]any {
	if dst == nil {
		dst = map[string]any{}
	}
	for k, v := range src {
		if vMap, ok := v.(map[string]any); ok {
			if existing, ok := dst[k].(map[string]any); ok {
				dst[k] = deepMergeJSON(existing, vMap)
				continue
			}
		}
		dst[k] = v
	}
	return dst
}

func setJSONPath(root map[string]any, path []string, value any) {
	current := root
	for i := 0; i < len(path)-1; i++ {
		key := path[i]
		next, ok := current[key].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[key] = next
		}
		current = next
	}
	current[path[len(path)-1]] = value
}

func getJSONPath(root map[string]any, path []string) any {
	current := any(root)
	for _, key := range path {
		m, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current, ok = m[key]
		if !ok {
			return nil
		}
	}
	return current
}

// isEmptyXrayValue returns true if v is nil, an empty map, or an empty/nil slice.
func isEmptyXrayValue(v any) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case map[string]any:
		return len(val) == 0
	case []any:
		return len(val) == 0
	}
	return false
}
