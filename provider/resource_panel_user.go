package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                     = &PanelUserResource{}
	_ resource.ResourceWithConfigure        = &PanelUserResource{}
	_ resource.ResourceWithImportState      = &PanelUserResource{}
	_ resource.ResourceWithConfigValidators = &PanelUserResource{}
)

type PanelUserResource struct {
	client *Client
}

type PanelUserModel struct {
	ID                types.String `tfsdk:"id"`
	Username          types.String `tfsdk:"username"`
	Password          types.String `tfsdk:"password"`
	PasswordWO        types.String `tfsdk:"password_wo"`
	PasswordWOVersion types.Int64  `tfsdk:"password_wo_version"`
}

func NewPanelUserResource() resource.Resource {
	return &PanelUserResource{}
}

func (r *PanelUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_panel_user"
}

func (r *PanelUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the admin username and password for the 3x-ui panel.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "The desired admin username.",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The desired admin password.",
				Validators: []validator.String{
					stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("password_wo")),
				},
			},
			"password_wo": schema.StringAttribute{
				Optional:    true,
				WriteOnly:   true,
				Description: "Write-only version of password.",
			},
			"password_wo_version": schema.Int64Attribute{
				Optional:    true,
				Description: "Increment this to trigger a password update when using password_wo.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *PanelUserResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(path.MatchRoot("password"), path.MatchRoot("password_wo")),
	}
}

func (r *PanelUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PanelUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PanelUserModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config PanelUserModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newUsername := plan.Username.ValueString()
	newPassword := resolvePanelUserPassword(plan, config)

	oldUsername := r.client.username
	oldPassword := r.client.password

	if err := r.client.UpdateUser(ctx, oldUsername, oldPassword, newUsername, newPassword); err != nil {
		resp.Diagnostics.AddError("Failed to update user", err.Error())
		return
	}

	r.warnCredentialsChanged(&resp.Diagnostics)

	plan.ID = types.StringValue("user")
	if !config.PasswordWO.IsNull() {
		plan.Password = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PanelUserResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
}

func (r *PanelUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PanelUserModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PanelUserModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config PanelUserModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newUsername := plan.Username.ValueString()
	newPassword := resolvePanelUserPasswordUpdate(plan, state, config)

	oldUsername := state.Username.ValueString()
	oldPassword := state.Password.ValueString()
	if oldPassword == "" {
		oldUsername = r.client.username
		oldPassword = r.client.password
	}

	if err := r.client.UpdateUser(ctx, oldUsername, oldPassword, newUsername, newPassword); err != nil {
		resp.Diagnostics.AddError("Failed to update user", err.Error())
		return
	}

	r.warnCredentialsChanged(&resp.Diagnostics)

	plan.ID = types.StringValue("user")
	if !config.PasswordWO.IsNull() {
		plan.Password = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PanelUserResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"Panel user credentials cannot be reverted",
		"Removing this resource from state does not revert the admin credentials. The current username and password remain active on the panel.",
	)
}

func (r *PanelUserResource) warnCredentialsChanged(diags *diag.Diagnostics) {
	diags.AddWarning(
		"Admin credentials changed",
		"The panel admin credentials have been updated. Update the provider's username and password to match the new values, otherwise the provider will fail to authenticate on the next run.",
	)
}

func (r *PanelUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func resolvePanelUserPassword(plan, config PanelUserModel) string {
	if !config.PasswordWO.IsNull() {
		return config.PasswordWO.ValueString()
	}
	return plan.Password.ValueString()
}

func resolvePanelUserPasswordUpdate(plan, state, config PanelUserModel) string {
	if !config.PasswordWO.IsNull() {
		return config.PasswordWO.ValueString()
	}
	if !plan.Password.IsNull() {
		return plan.Password.ValueString()
	}
	return state.Password.ValueString()
}
