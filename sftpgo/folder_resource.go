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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &folderResource{}
	_ resource.ResourceWithConfigure = &folderResource{}
)

// NewFolderResource is a helper function to simplify the provider implementation.
func NewFolderResource() resource.Resource {
	return &folderResource{}
}

// folderResource is the resource implementation.
type folderResource struct {
	client *client.Client
}

// Configure adds the provider configured client to the resource.
func (r *folderResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Metadata returns the resource type name.
func (r *folderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_folder"
}

// Schema defines the schema for the resource.
func (r *folderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Virtual folder",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Required to use the test framework. Matches the folder name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique name",
			},
			"mapped_path": schema.StringAttribute{
				Optional:    true,
				Description: "Absolute path to a local directory. This is the folder root path for local storage provider. For non-local filesystems it will store temporary files.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Optional description.",
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
				Description: "Last quota update as unix timestamp in milliseconds",
			},
			"filesystem": getSchemaForFilesystem(),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *folderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan virtualFolderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	folder, diags := plan.toSFTPGo(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	folder, err := r.client.CreateFolder(*folder)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating folder",
			"Could not create folder, unexpected error: "+err.Error(),
		)
		return
	}
	var state virtualFolderResourceModel
	diags = state.fromSFTPGo(ctx, folder)
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
func (r *folderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state virtualFolderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	folder, err := r.client.GetFolder(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SFTPGo Folder",
			"Could not read SFTPGo Folder "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	var newState virtualFolderResourceModel
	diags = newState.fromSFTPGo(ctx, folder)
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
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *folderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan virtualFolderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	folder, diags := plan.toSFTPGo(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateFolder(*folder)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating folder",
			"Could not update folder, unexpected error: "+err.Error(),
		)
		return
	}

	folder, err = r.client.GetFolder(plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SFTPGo Folder",
			"Could not read SFTPGo Folder "+plan.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	var state virtualFolderResourceModel
	diags = state.fromSFTPGo(ctx, folder)
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
func (r *folderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state virtualFolderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing folder
	err := r.client.DeleteFolder(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting SFTPGo folder",
			"Could not delete folder, unexpected error: "+err.Error(),
		)
		return
	}
}

func (*folderResource) preservePlanFields(ctx context.Context, plan, state *virtualFolderResourceModel) diag.Diagnostics {
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
