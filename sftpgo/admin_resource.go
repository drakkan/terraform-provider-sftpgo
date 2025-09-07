// Copyright (C) 2023 Nicola Murino
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sftpgo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &adminResource{}
	_ resource.ResourceWithConfigure   = &adminResource{}
	_ resource.ResourceWithImportState = &adminResource{}
)

// NewAdminResource is a helper function to simplify the provider implementation.
func NewAdminResource() resource.Resource {
	return &adminResource{}
}

// adminResource is the resource implementation.
type adminResource struct {
	client *client.Client
}

// Configure adds the provider configured client to the resource.
func (r *adminResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Metadata returns the resource type name.
func (r *adminResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin"
}

// Schema defines the schema for the resource.
func (r *adminResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Admin",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Required to use the test framework. Matches the username.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "Unique username.",
			},
			"status": schema.Int64Attribute{
				Required:    true,
				Description: "1 enabled, 0 disabled (login is not allowed).",
				Validators: []validator.Int64{
					int64validator.Between(0, 1),
				},
			},
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Plain text password or hash format supported by SFTPGo.",
			},
			"email": schema.StringAttribute{
				Optional: true,
			},
			"permissions": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "Granted permissions.",
				Validators: []validator.List{
					listvalidator.UniqueValues(),
					listvalidator.ValueStringsAre(stringvalidator.OneOf("*", "add_users", "edit_users",
						"del_users", "view_users", "view_conns", "close_conns", "view_status", "manage_admins",
						"manage_folders", "manage_groups", "manage_apikeys", "quota_scans", "manage_system",
						"manage_defender", "view_defender", "retention_checks", "metadata_checks", "view_events",
						"manage_event_rules", "manage_roles", "manage_ip_lists", "disable_mfa")),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Optional description.",
			},
			"additional_info": schema.StringAttribute{
				Optional:    true,
				Description: "Free form text field.",
			},
			"created_at": schema.Int64Attribute{
				Computed:    true,
				Description: "Creation time as unix timestamp in milliseconds.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.Int64Attribute{
				Computed:    true,
				Description: "Last update time as unix timestamp in milliseconds.",
			},
			"last_login": schema.Int64Attribute{
				Computed:    true,
				Description: "Last login as unix timestamp in milliseconds.",
			},
			"role": schema.StringAttribute{
				Optional:    true,
				Description: "Role name. If set the admin can only administer users with the same role.",
			},
			"filters": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Additional restrictions.",
				Attributes: map[string]schema.Attribute{
					"allow_list": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: `Only connections from these IP/Mask are allowed. IP/Mask must be in CIDR notation as defined in RFC 4632 and RFC 4291 for example "192.0.2.0/24" or "2001:db8::/32"`,
					},
					"allow_api_key_auth": schema.BoolAttribute{
						Optional:    true,
						Description: "If set, API Key authentication is allowed.",
					},
					"require_password_change": schema.BoolAttribute{
						Optional:    true,
						Description: "If set, two factor authentication is required.",
					},
					"require_two_factor": schema.BoolAttribute{
						Optional:    true,
						Description: "If set, API Key authentication is allowed.",
					},
					"disable_password_auth": schema.BoolAttribute{
						Optional:    true,
						Description: "If set, password authentication is disabled. The administrator can authenticate using an API key or OpenID Connect, if either is enabled. " + enterpriseFeatureNote + ".",
					},
				},
			},
			"preferences": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Admin preferences.",
				Attributes: map[string]schema.Attribute{
					"hide_user_page_sections": schema.Int64Attribute{
						Optional:    true,
						Description: "If set allow to hide some sections from the user page in the WebAdmin. 1 = groups, 2 = filesystem, 4 = virtual folders, 8 = profile, 16 = ACL, 32 = Disk and bandwidth quota limits, 64 = Advanced. Settings can be combined.",
					},
					"default_users_expiration": schema.Int64Attribute{
						Optional:    true,
						Description: "If set defines the default expiration for newly created users as number of days.",
					},
				},
			},
			"groups": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Groups automatically selected for new users created by this admin.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Group name.",
						},
						"options": schema.SingleNestedAttribute{
							Required:    true,
							Description: "Options for admin/group mapping",
							Attributes: map[string]schema.Attribute{
								"add_to_users_as": schema.Int64Attribute{
									Required:    true,
									Description: "Add to users as the specified group type. 1 = Primary, 2 = Secondary, 3 = Membership only.",
								},
							},
						},
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *adminResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan adminResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	admin, diags := plan.toSFTPGo(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	admin, err := r.client.CreateAdmin(*admin)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating admin",
			"Could not create admin, unexpected error: "+err.Error(),
		)
		return
	}
	var state adminResourceModel
	diags = state.fromSFTPGo(ctx, admin)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Password = plan.Password

	// Set state to fully populated data
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *adminResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state adminResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	admin, err := r.client.GetAdmin(state.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SFTPGo Admin",
			"Could not read SFTPGo Admin "+state.Username.ValueString()+": "+err.Error(),
		)
		return
	}

	var newState adminResourceModel
	diags = newState.fromSFTPGo(ctx, admin)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState.Password = state.Password

	// Set refreshed state
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *adminResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan adminResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	admin, diags := plan.toSFTPGo(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateAdmin(*admin)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating admin",
			"Could not update admin, unexpected error: "+err.Error(),
		)
		return
	}

	admin, err = r.client.GetAdmin(plan.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SFTPGo Admin",
			"Could not read SFTPGo Admin "+plan.Username.ValueString()+": "+err.Error(),
		)
		return
	}

	var state adminResourceModel
	diags = state.fromSFTPGo(ctx, admin)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Password = plan.Password

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *adminResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state adminResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing admin
	err := r.client.DeleteAdmin(state.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting SFTPGo admin",
			"Could not delete admin, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing the resource and save the Terraform state
func (*adminResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import username and save to username attribute
	resource.ImportStatePassthroughID(ctx, path.Root("username"), req, resp)
}
