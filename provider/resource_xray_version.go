package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &XrayVersionResource{}
	_ resource.ResourceWithConfigure = &XrayVersionResource{}
)

type XrayVersionResource struct {
	client *Client
}

type XrayVersionModel struct {
	ID             types.String `tfsdk:"id"`
	Version        types.String `tfsdk:"version"`
	CurrentVersion types.String `tfsdk:"current_version"`
}

func NewXrayVersionResource() resource.Resource {
	return &XrayVersionResource{}
}

func (r *XrayVersionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_xray_version"
}

func (r *XrayVersionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the installed Xray core version on the 3x-ui panel.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				Required:    true,
				Description: "The desired Xray version to install (e.g. \"v25.1.1\"). Must include the \"v\" prefix.",
			},
			"current_version": schema.StringAttribute{
				Computed:    true,
				Description: "The currently installed Xray version (with \"v\" prefix).",
			},
		},
	}
}

func (r *XrayVersionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *Client")
		return
	}
	r.client = client
}

func (r *XrayVersionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan XrayVersionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	version := plan.Version.ValueString()

	if err := r.client.InstallXray(ctx, version); err != nil {
		resp.Diagnostics.AddError("Failed to install Xray version", err.Error())
		return
	}

	current, err := r.waitForXrayVersion(ctx, version)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get current Xray version", err.Error())
		return
	}

	if current != version {
		resp.Diagnostics.AddError(
			"Xray version mismatch after install",
			fmt.Sprintf("Requested %s but the panel reports %s.", version, current),
		)
		return
	}

	plan.ID = types.StringValue("xray_version")
	plan.CurrentVersion = types.StringValue(current)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *XrayVersionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state XrayVersionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetCurrentXrayVersion(ctx)
	if err != nil {
		if errors.Is(err, ErrXrayVersionUnknown) {
			// Xray is not running — keep existing state and warn.
			resp.Diagnostics.AddWarning(
				"Xray version is unknown",
				"The Xray process may not be running. The previously known version is preserved in state. "+
					"Restart Xray via the panel to allow Terraform to detect the actual version.",
			)
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
		resp.Diagnostics.AddError("Failed to get current Xray version", err.Error())
		return
	}

	// Update both version and current_version from the observed state
	// so that Terraform detects drift if the version was changed outside TF.
	state.Version = types.StringValue(current)
	state.CurrentVersion = types.StringValue(current)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *XrayVersionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan XrayVersionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	version := plan.Version.ValueString()

	if err := r.client.InstallXray(ctx, version); err != nil {
		resp.Diagnostics.AddError("Failed to install Xray version", err.Error())
		return
	}

	current, err := r.waitForXrayVersion(ctx, version)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get current Xray version", err.Error())
		return
	}

	if current != version {
		resp.Diagnostics.AddError(
			"Xray version mismatch after install",
			fmt.Sprintf("Requested %s but the panel reports %s.", version, current),
		)
		return
	}

	plan.ID = types.StringValue("xray_version")
	plan.CurrentVersion = types.StringValue(current)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// waitForXrayVersion polls GetCurrentXrayVersion until it matches the desired
// version or the timeout (30s) is exceeded. InstallXray is asynchronous — the
// API returns immediately while the download/install runs in the background.
func (r *XrayVersionResource) waitForXrayVersion(ctx context.Context, version string) (string, error) {
	for i := 0; i < 30; i++ {
		current, err := r.client.GetCurrentXrayVersion(ctx)
		if err != nil {
			return "", err
		}
		if current == version {
			return current, nil
		}
		time.Sleep(time.Second)
	}
	current, err := r.client.GetCurrentXrayVersion(ctx)
	if err != nil {
		return "", err
	}
	return current, nil
}

func (r *XrayVersionResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"Xray version cannot be reverted",
		"Removing this resource from state does not change the installed Xray version. The current version remains active on the panel.",
	)
}
