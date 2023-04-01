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

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/sftpgo/terraform-provider-sftpgo/sftpgo/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &groupResource{}
	_ resource.ResourceWithConfigure   = &groupResource{}
	_ resource.ResourceWithImportState = &groupResource{}
)

// NewGroupResource is a helper function to simplify the provider implementation.
func NewGroupResource() resource.Resource {
	return &groupResource{}
}

// groupResource is the resource implementation.
type groupResource struct {
	client *client.Client
}

// Configure adds the provider configured client to the resource.
func (r *groupResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Metadata returns the resource type name.
func (r *groupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

// Schema defines the schema for the resource.
func (r *groupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Group",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Optional description.",
			},
			"created_at": schema.Int64Attribute{
				Computed:    true,
				Description: "Creation time as unix timestamp in milliseconds.",
			},
			"updated_at": schema.Int64Attribute{
				Computed:    true,
				Description: "Last update time as unix timestamp in milliseconds.",
			},
			"user_settings": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Settings to apply to users",
				Attributes: map[string]schema.Attribute{
					"home_dir": schema.StringAttribute{
						Optional:    true,
						Description: "If not set and the filesystem provider is local (0), the root filesystem will not be overridden.",
					},
					"max_sessions": schema.Int64Attribute{
						Optional:    true,
						Description: "Maximum concurrent sessions.",
					},
					"quota_size": schema.Int64Attribute{
						Optional:    true,
						Description: "Maximum size allowed as bytes.",
					},
					"quota_files": schema.Int64Attribute{
						Optional:    true,
						Description: "Maximum number of files allowed",
					},
					"permissions": schema.MapAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Description: "Comma separated, per-directory, permissions.",
					},
					"upload_bandwidth": schema.Int64Attribute{
						Optional:    true,
						Description: "Maximum upload bandwidth as KB/s. This is the default if no per-source limit match.",
					},
					"download_bandwidth": schema.Int64Attribute{
						Optional:    true,
						Description: "Maximum download bandwidth as KB/s. This is the default if no per-source limit match.",
					},
					"upload_data_transfer": schema.Int64Attribute{
						Optional:    true,
						Description: "Maximum data transfer allowed for uploads as MB.",
					},
					"download_data_transfer": schema.Int64Attribute{
						Optional:    true,
						Description: "Maximum data transfer allowed for downloads as MB",
					},
					"total_data_transfer": schema.Int64Attribute{
						Optional:    true,
						Description: "Maximum total data transfer as MB. You can set a total data transfer instead of the individual values for uploads and downloads.",
					},
					"expires_in": schema.Int64Attribute{
						Optional:    true,
						Description: "Defines account expiration in number of days from creation. Not set means no expiration.",
					},
					"filters":    getSchemaForUserFilters(true),
					"filesystem": getSchemaForFilesystem(),
				},
			},
			"virtual_folders": getSchemaForVirtualFolders(),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan groupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, diags := plan.toSFTPGo(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.CreateGroup(*group)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating group",
			"Could not create group, unexpected error: "+err.Error(),
		)
		return
	}
	var state groupResourceModel
	diags = state.fromSFTPGo(ctx, group)
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
func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state groupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.GetGroup(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SFTPGo Group",
			"Could not read SFTPGo Group "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	diags = state.fromSFTPGo(ctx, group)
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
func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan groupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	group, diags := plan.toSFTPGo(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateGroup(*group)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating group",
			"Could not update group, unexpected error: "+err.Error(),
		)
		return
	}

	group, err = r.client.GetGroup(plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SFTPGo Group",
			"Could not read SFTPGo Group "+plan.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	var state groupResourceModel
	diags = state.fromSFTPGo(ctx, group)
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
func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state groupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing group
	err := r.client.DeleteGroup(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting SFTPGo group",
			"Could not delete group, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing the resource and save the Terraform state
func (r *groupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import name and save to name attribute
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (*groupResource) preservePlanFields(ctx context.Context, plan, state *groupResourceModel) diag.Diagnostics {
	var settingsPlan groupUserSettings
	diags := plan.UserSettings.As(ctx, &settingsPlan, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return diags
	}
	var settingsState groupUserSettings
	diags = state.UserSettings.As(ctx, &settingsState, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return diags
	}

	var fsPlan filesystem
	diags = settingsPlan.FsConfig.As(ctx, &fsPlan, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return diags
	}
	var fsState filesystem
	diags = settingsState.FsConfig.As(ctx, &fsState, basetypes.ObjectAsOptions{
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
	settingsState.FsConfig = fs

	settings, diags := types.ObjectValueFrom(ctx, settingsState.getTFAttributes(), settingsState)
	if diags.HasError() {
		return diags
	}
	state.UserSettings = settings

	return nil
}
