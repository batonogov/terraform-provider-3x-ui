package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &PanelUserResource{}
	_ resource.ResourceWithConfigure = &PanelUserResource{}
)

type PanelUserResource struct {
	client *Client
}

type PanelUserModel struct {
	ID       types.String `tfsdk:"id"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
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
				Required:    true,
				Sensitive:   true,
				Description: "The desired admin password.",
			},
		},
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

	newUsername := plan.Username.ValueString()
	newPassword := plan.Password.ValueString()

	// Use the provider's current credentials as old credentials.
	oldUsername := r.client.username
	oldPassword := r.client.password

	if err := r.client.UpdateUser(ctx, oldUsername, oldPassword, newUsername, newPassword); err != nil {
		resp.Diagnostics.AddError("Failed to update user", err.Error())
		return
	}

	r.warnCredentialsChanged(&resp.Diagnostics)

	plan.ID = types.StringValue("user")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PanelUserResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
	// No API to read user credentials; state is preserved as-is.
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

	newUsername := plan.Username.ValueString()
	newPassword := plan.Password.ValueString()

	// Use the previous state credentials as old credentials.
	oldUsername := state.Username.ValueString()
	oldPassword := state.Password.ValueString()

	if err := r.client.UpdateUser(ctx, oldUsername, oldPassword, newUsername, newPassword); err != nil {
		resp.Diagnostics.AddError("Failed to update user", err.Error())
		return
	}

	r.warnCredentialsChanged(&resp.Diagnostics)

	plan.ID = types.StringValue("user")
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
