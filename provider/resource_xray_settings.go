package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
	xraySectionBasics    = xraySection{id: "xray_basics", mode: xraySectionMergeRoot}
	xraySectionDNS       = xraySection{id: "xray_dns", mode: xraySectionSetPath, path: []string{"dns"}}
	xraySectionRouting   = xraySection{id: "xray_routing", mode: xraySectionSetPath, path: []string{"routing"}}
	xraySectionBalancers = xraySection{id: "xray_balancers", mode: xraySectionSetPath, path: []string{"routing", "balancers"}}
	xraySectionReverse   = xraySection{id: "xray_reverse", mode: xraySectionSetPath, path: []string{"reverse"}}
	xraySectionOutbounds = xraySection{id: "xray_outbounds", mode: xraySectionSetPath, path: []string{"outbounds"}}
)

type xrayBuildFunc func(d map[string]any) any
type xrayFlattenFunc func(data any) map[string]any

// xrayResourceModel is shared by all xray resources: they store config as JSON.
type xrayResourceModel struct {
	ID   types.String `tfsdk:"id"`
	JSON types.String `tfsdk:"json"`
}

func xrayResourceSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"json": schema.StringAttribute{
				Required:    true,
				Description: "JSON representation of the xray section config.",
				PlanModifiers: []planmodifier.String{
					jsonSubsetPlanModifier{},
				},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Shared CRUD helpers
// ---------------------------------------------------------------------------

func xrayApplyHelper(
	ctx context.Context,
	jsonStr string,
	diags *diag.Diagnostics,
	client *Client,
	section xraySection,
	build xrayBuildFunc,
	flatten xrayFlattenFunc,
) *xrayResourceModel {
	var input map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
		diags.AddError("Invalid JSON", err.Error())
		return nil
	}

	desired := build(input)

	xrayTemplateMu.Lock()
	defer xrayTemplateMu.Unlock()

	current, err := client.GetXrayTemplate(ctx)
	if err != nil {
		diags.AddError("Failed to get xray template", err.Error())
		return nil
	}

	updated, err := applyXraySection(current, desired, section)
	if err != nil {
		diags.AddError("Failed to apply xray section", err.Error())
		return nil
	}

	if err := client.UpdateXrayTemplate(ctx, updated); err != nil {
		diags.AddError("Failed to update xray template", err.Error())
		return nil
	}

	result := xrayReadHelperLocked(ctx, diags, client, section, flatten)
	if result == nil {
		return nil
	}

	// Preserve plan JSON if it is a subset of the flatten result.
	// This prevents "inconsistent result" when flatten adds API defaults
	// (e.g. xray_basics flatten may add "error":"" to log).
	var planVal, stateVal any
	if json.Unmarshal([]byte(jsonStr), &planVal) == nil {
		if json.Unmarshal([]byte(result.JSON.ValueString()), &stateVal) == nil {
			if isSubset(planVal, stateVal) {
				result.JSON = types.StringValue(jsonStr)
			}
		}
	}

	return result
}

// xrayReadHelperLocked reads the xray template without acquiring the mutex.
// The caller must hold xrayTemplateMu if concurrent access is possible.
func xrayReadHelperLocked(
	ctx context.Context,
	diags *diag.Diagnostics,
	client *Client,
	section xraySection,
	flatten xrayFlattenFunc,
) *xrayResourceModel {
	current, err := client.GetXrayTemplate(ctx)
	if err != nil {
		diags.AddError("Failed to get xray template", err.Error())
		return nil
	}

	value := extractXraySection(current, section)
	flat := flatten(value)

	payload, err := json.Marshal(flat)
	if err != nil {
		diags.AddError("Failed to marshal state", err.Error())
		return nil
	}

	return &xrayResourceModel{
		ID:   types.StringValue(section.id),
		JSON: types.StringValue(string(payload)),
	}
}

// xrayReadHelper reads the xray template (acquires no mutex — Read is safe to
// call without the write-side lock).
func xrayReadHelper(
	ctx context.Context,
	diags *diag.Diagnostics,
	client *Client,
	section xraySection,
	flatten xrayFlattenFunc,
) *xrayResourceModel {
	return xrayReadHelperLocked(ctx, diags, client, section, flatten)
}

func xrayImportState(
	ctx context.Context,
	_ resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
	section xraySection,
	flatten xrayFlattenFunc,
	client *Client,
) {
	state := xrayReadHelper(ctx, &resp.Diagnostics, client, section, flatten)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
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
	resp.Schema = xrayResourceSchema()
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
	var plan xrayResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := xrayApplyHelper(ctx, plan.JSON.ValueString(), &resp.Diagnostics, r.client, xraySectionBasics,
		buildXrayBasicsJSON,
		flattenXrayBasicsToMap)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *XrayBasicsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var cur xrayResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &cur)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := xrayReadHelper(ctx, &resp.Diagnostics, r.client, xraySectionBasics, flattenXrayBasicsToMap)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *XrayBasicsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan xrayResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := xrayApplyHelper(ctx, plan.JSON.ValueString(), &resp.Diagnostics, r.client, xraySectionBasics,
		buildXrayBasicsJSON,
		flattenXrayBasicsToMap)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *XrayBasicsResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *XrayBasicsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	xrayImportState(ctx, req, resp, xraySectionBasics, flattenXrayBasicsToMap, r.client)
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
	resp.Schema = xrayResourceSchema()
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
	var plan xrayResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := xrayApplyHelper(ctx, plan.JSON.ValueString(), &resp.Diagnostics, r.client, xraySectionDNS,
		buildXrayDNSJSON,
		flattenXrayDNSToMap)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *XrayDNSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var cur xrayResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &cur)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := xrayReadHelper(ctx, &resp.Diagnostics, r.client, xraySectionDNS, flattenXrayDNSToMap)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *XrayDNSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan xrayResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := xrayApplyHelper(ctx, plan.JSON.ValueString(), &resp.Diagnostics, r.client, xraySectionDNS,
		buildXrayDNSJSON,
		flattenXrayDNSToMap)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *XrayDNSResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *XrayDNSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	xrayImportState(ctx, req, resp, xraySectionDNS, flattenXrayDNSToMap, r.client)
}

// ---------------------------------------------------------------------------
// XrayRoutingResource
// ---------------------------------------------------------------------------

var (
	_ resource.Resource                = &XrayRoutingResource{}
	_ resource.ResourceWithImportState = &XrayRoutingResource{}
)

type XrayRoutingResource struct{ client *Client }

func NewXrayRoutingResource() resource.Resource { return &XrayRoutingResource{} }

func (r *XrayRoutingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_xray_routing"
}

func (r *XrayRoutingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = xrayResourceSchema()
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
	var plan xrayResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := xrayApplyHelper(ctx, plan.JSON.ValueString(), &resp.Diagnostics, r.client, xraySectionRouting,
		buildXrayRoutingJSON,
		flattenXrayRoutingToMap)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *XrayRoutingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var cur xrayResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &cur)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := xrayReadHelper(ctx, &resp.Diagnostics, r.client, xraySectionRouting, flattenXrayRoutingToMap)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *XrayRoutingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan xrayResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := xrayApplyHelper(ctx, plan.JSON.ValueString(), &resp.Diagnostics, r.client, xraySectionRouting,
		buildXrayRoutingJSON,
		flattenXrayRoutingToMap)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *XrayRoutingResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *XrayRoutingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	xrayImportState(ctx, req, resp, xraySectionRouting, flattenXrayRoutingToMap, r.client)
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
	resp.Schema = xrayResourceSchema()
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
	var plan xrayResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := xrayApplyHelper(ctx, plan.JSON.ValueString(), &resp.Diagnostics, r.client, xraySectionBalancers,
		buildXrayBalancersJSON,
		flattenXrayBalancersToMap)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *XrayBalancersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var cur xrayResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &cur)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := xrayReadHelper(ctx, &resp.Diagnostics, r.client, xraySectionBalancers, flattenXrayBalancersToMap)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *XrayBalancersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan xrayResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := xrayApplyHelper(ctx, plan.JSON.ValueString(), &resp.Diagnostics, r.client, xraySectionBalancers,
		buildXrayBalancersJSON,
		flattenXrayBalancersToMap)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *XrayBalancersResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *XrayBalancersResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	xrayImportState(ctx, req, resp, xraySectionBalancers, flattenXrayBalancersToMap, r.client)
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
	resp.Schema = xrayResourceSchema()
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
	var plan xrayResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := xrayApplyHelper(ctx, plan.JSON.ValueString(), &resp.Diagnostics, r.client, xraySectionReverse,
		buildXrayReverseJSON,
		flattenXrayReverseToMap)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *XrayReverseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var cur xrayResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &cur)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := xrayReadHelper(ctx, &resp.Diagnostics, r.client, xraySectionReverse, flattenXrayReverseToMap)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *XrayReverseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan xrayResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := xrayApplyHelper(ctx, plan.JSON.ValueString(), &resp.Diagnostics, r.client, xraySectionReverse,
		buildXrayReverseJSON,
		flattenXrayReverseToMap)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *XrayReverseResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *XrayReverseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	xrayImportState(ctx, req, resp, xraySectionReverse, flattenXrayReverseToMap, r.client)
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
	resp.Schema = xrayResourceSchema()
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
	var plan xrayResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := xrayApplyHelper(ctx, plan.JSON.ValueString(), &resp.Diagnostics, r.client, xraySectionOutbounds,
		buildXrayOutboundsJSON,
		flattenXrayOutboundsToMap)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *XrayOutboundsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var cur xrayResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &cur)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := xrayReadHelper(ctx, &resp.Diagnostics, r.client, xraySectionOutbounds, flattenXrayOutboundsToMap)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *XrayOutboundsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan xrayResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := xrayApplyHelper(ctx, plan.JSON.ValueString(), &resp.Diagnostics, r.client, xraySectionOutbounds,
		buildXrayOutboundsJSON,
		flattenXrayOutboundsToMap)
	if state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *XrayOutboundsResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *XrayOutboundsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	xrayImportState(ctx, req, resp, xraySectionOutbounds, flattenXrayOutboundsToMap, r.client)
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
		for _, key := range []string{"log", "policy", "api", "stats"} {
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
