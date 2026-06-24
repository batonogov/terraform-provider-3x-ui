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
// version or the timeout is exceeded. Although UpdateXray in 3x-ui is
// synchronous (download → extract → RestartXray), the restarted xray process
// may not report any version for a few seconds via getXrayVersion — during
// that gap the panel returns "Unknown", which surfaces as
// ErrXrayVersionUnknown. We treat that as "still restarting" and keep
// polling rather than failing the apply (issue #157).
//
// On slow CI runners (PostgreSQL backend) the stall can last minutes: 3x-ui's
// refreshVersion() runs `xray -version` via exec.Command(...).Output() with
// NO timeout, and under IO pressure that exec blocks with p.version="Unknown"
// (PR #306 PostgreSQL run: 4.5 min of continuous "Unknown"). To ride this out
// the budget was raised (90s → 180s) and, after nudgeAfter seconds of
// continuous Unknown, we issue a single RestartXrayService under a bounded
// context — it re-triggers refreshVersion() and is far cheaper than
// InstallXray (no GitHub download). RestartXrayService honours the context
// via doJSON, so the nudge is itself time-bounded and cannot extend the stall.
func (r *XrayVersionResource) waitForXrayVersion(ctx context.Context, version string) (string, error) {
	const (
		maxAttempts       = 180
		retryInstallAfter = 30
		nudgeAfter        = 60
		nudgeTimeout      = 30 * time.Second
	)
	var lastSeen string
	var retried bool
	var nudged bool
	for i := 0; i < maxAttempts; i++ {
		current, err := r.client.GetCurrentXrayVersion(ctx)
		if err == nil {
			lastSeen = current
			if current == version {
				return current, nil
			}
		} else if !errors.Is(err, ErrXrayVersionUnknown) {
			return "", err
		}
		// If the panel reports a stale version after retryInstallAfter polls,
		// re-issue InstallXray once — the first install may not have taken
		// effect (observed on 3x-ui v3.2.6–v3.2.7, see #262).
		if i == retryInstallAfter && !retried && lastSeen != "" && lastSeen != version {
			if installErr := r.client.InstallXray(ctx, version); installErr != nil {
				return "", fmt.Errorf("re-issuing InstallXray(%s): %w", version, installErr)
			}
			retried = true
		}
		// After nudgeAfter seconds of continuous Unknown, force a single cheap
		// restart under a bounded context to re-trigger refreshVersion(). A failed
		// nudge is non-fatal — we log nothing (no logger here) and keep polling.
		if !nudged && i >= nudgeAfter {
			nudgeCtx, cancel := context.WithTimeout(ctx, nudgeTimeout)
			nudgeErr := r.client.RestartXrayService(nudgeCtx)
			cancel()
			nudged = true
			if nudgeErr != nil {
				// Transient nudge failure: keep polling, the version may still
				// come back on its own once the panel's refreshVersion() returns.
				continue
			}
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(time.Second):
		}
	}
	if lastSeen != "" {
		return lastSeen, nil
	}
	return "", ErrXrayVersionUnknown
}

func (r *XrayVersionResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"Xray version cannot be reverted",
		"Removing this resource from state does not change the installed Xray version. The current version remains active on the panel.",
	)
}
