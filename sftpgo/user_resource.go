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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &userResource{}
	_ resource.ResourceWithConfigure = &userResource{}
)

// NewUserResource is a helper function to simplify the provider implementation.
func NewUserResource() resource.Resource {
	return &userResource{}
}

// userResource is the resource implementation.
type userResource struct {
	client *client.Client
}

// Configure adds the provider configured client to the resource.
func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Metadata returns the resource type name.
func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema defines the schema for the resource.
func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "User",
		Attributes: map[string]schema.Attribute{
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
			"expiration_date": schema.Int64Attribute{
				Optional:    true,
				Description: "Account expiration date as unix timestamp in milliseconds. An expired account cannot login.",
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"public_keys": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "List of public keys in OpenSSH format.",
			},
			"home_dir": schema.StringAttribute{
				Required:    true,
				Description: "The user cannot upload or download files outside this directory. Must be an absolute path.",
			},
			"email": schema.StringAttribute{
				Optional: true,
			},
			"uid": schema.Int64Attribute{
				Optional:    true,
				Description: "If SFTPGo runs as root system user then the created files and directories will be assigned to this system UID. Default not set.",
			},
			"gid": schema.Int64Attribute{
				Optional:    true,
				Description: "If SFTPGo runs as root system user then the created files and directories will be assigned to this system GID. Default not set.",
			},
			"max_sessions": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum concurrent sessions. Not set means no limit.",
			},
			"quota_size": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum size allowed as bytes. Not set means no limit.",
			},
			"quota_files": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum number of files allowed. Not set means no limit.",
			},
			"permissions": schema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Comma separated, per-directory, permissions.",
			},
			"used_quota_size": schema.Int64Attribute{
				Computed:    true,
				Description: "Used quota as bytes.",
			},
			"used_quota_files": schema.Int64Attribute{
				Computed:    true,
				Description: "Used quota as number of files.",
			},
			"last_quota_update": schema.Int64Attribute{
				Computed:    true,
				Description: "Last quota update as unix timestamp in milliseconds.",
			},
			"upload_bandwidth": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum upload bandwidth as KB/s. Not set means unlimited. This is the default if no per-source limit match.",
			},
			"download_bandwidth": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum download bandwidth as KB/s. Not set means unlimited. This is the default if no per-source limit match.",
			},
			"upload_data_transfer": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum data transfer allowed for uploads as MB. Not set means no limit.",
			},
			"download_data_transfer": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum data transfer allowed for downloads as MB. Not set means no limit.",
			},
			"total_data_transfer": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum total data transfer as MB. Not set means unlimited. You can set a total data transfer instead of the individual values for uploads and downloads.",
			},
			"used_upload_data_transfer": schema.Int64Attribute{
				Computed:    true,
				Description: "Uploaded size, as bytes, since the last reset.",
			},
			"used_download_data_transfer": schema.Int64Attribute{
				Computed:    true,
				Description: "Downloaded size, as bytes, since the last reset.",
			},
			"last_login": schema.Int64Attribute{
				Computed:    true,
				Description: "Last login as unix timestamp in milliseconds.",
			},
			"created_at": schema.Int64Attribute{
				Computed:    true,
				Description: "Creation time as unix timestamp in milliseconds.",
			},
			"updated_at": schema.Int64Attribute{
				Computed:    true,
				Description: "Last update time as unix timestamp in milliseconds.",
			},
			"first_download": schema.Int64Attribute{
				Computed:    true,
				Description: "First download time as unix timestamp in milliseconds.",
			},
			"first_upload": schema.Int64Attribute{
				Computed:    true,
				Description: "First upload time as unix timestamp in milliseconds.",
			},
			"last_password_change": schema.Int64Attribute{
				Computed:    true,
				Description: "Last password change as unix timestamp in milliseconds.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Optional description.",
			},
			"additional_info": schema.StringAttribute{
				Optional:    true,
				Description: "Free form text field.",
			},
			"role": schema.StringAttribute{
				Optional:    true,
				Description: "Role name.",
			},
			"groups": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Groups.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Group name.",
						},
						"type": schema.Int64Attribute{
							Required:    true,
							Description: "Group type. 1 = Primary, 2 = Secondary, 3 = Membership only.",
							Validators: []validator.Int64{
								int64validator.Between(1, 3),
							},
						},
					},
				},
			},
			"filters":         getSchemaForUserFilters(false),
			"virtual_folders": getSchemaForVirtualFolders(),
			"filesystem":      getSchemaForFilesystem(),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan userResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, diags := plan.toSFTPGo(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.CreateUser(*user)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating user",
			"Could not create user, unexpected error: "+err.Error(),
		)
		return
	}
	var state userResourceModel
	diags = state.fromSFTPGo(ctx, user)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = r.preservePlanFields(ctx, &plan, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state userResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(state.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SFTPGo User",
			"Could not read SFTPGo User "+state.Username.ValueString()+": "+err.Error(),
		)
		return
	}

	var newState userResourceModel
	diags = newState.fromSFTPGo(ctx, user)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = r.preservePlanFields(ctx, &state, &newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan userResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	user, diags := plan.toSFTPGo(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateUser(*user)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating user",
			"Could not update user, unexpected error: "+err.Error(),
		)
		return
	}

	user, err = r.client.GetUser(plan.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SFTPGo User",
			"Could not read SFTPGo User "+plan.Username.ValueString()+": "+err.Error(),
		)
		return
	}

	var state userResourceModel
	diags = state.fromSFTPGo(ctx, user)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = r.preservePlanFields(ctx, &plan, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing user
	err := r.client.DeleteUser(state.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting SFTPGo user",
			"Could not delete user, unexpected error: "+err.Error(),
		)
		return
	}
}

func (*userResource) preservePlanFields(ctx context.Context, plan, state *userResourceModel) diag.Diagnostics {
	state.Password = plan.Password

	var fsPlan filesystem
	diags := plan.FsConfig.As(ctx, &fsPlan, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return diags
	}
	var fsState filesystem
	diags = state.FsConfig.As(ctx, &fsState, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return diags
	}

	fs, diags := preserveFsConfigPlanFields(ctx, fsPlan, fsState)
	if diags.HasError() {
		return diags
	}
	state.FsConfig = fs

	return nil
}
