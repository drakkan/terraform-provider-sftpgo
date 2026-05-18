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

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &trustedListEntryResource{}
	_ resource.ResourceWithConfigure   = &trustedListEntryResource{}
	_ resource.ResourceWithImportState = &trustedListEntryResource{}
)

// NewTrustedListEntryResource is a helper function to simplify the provider implementation.
func NewTrustedListEntryResource() resource.Resource {
	return &trustedListEntryResource{}
}

// trustedListEntryResource is the resource implementation.
type trustedListEntryResource struct {
	client *client.Client
}

// Configure adds the provider configured client to the resource.
func (r *trustedListEntryResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Metadata returns the resource type name.
func (r *trustedListEntryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trustedlist_entry"
}

// Schema defines the schema for the resource.
func (r *trustedListEntryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Trusted list entry. " + enterpriseFeatureNote + ".",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Required to use the test framework. Matches the IP or network field.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ipornet": schema.StringAttribute{
				Required:    true,
				Description: `IP address or network in CIDR format, for example "192.168.1.2/32", "192.168.0.0/24", "2001:db8::/32". Changing this value after creation forces replacement because the SFTPGo API does not support renaming.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Optional description.",
			},
			"protocols": schema.Int64Attribute{
				Required:    true,
				Description: "Defines the protocol the entry applies to. 0 means all the supported protocols, 1 SSH, 2 FTP, 4 WebDAV, 8 HTTP. Protocols can be combined, for example 3 means SSH and FTP.",
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
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *trustedListEntryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan trustedListEntryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	entry, diags := plan.toSFTPGo(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	entry, err := r.client.CreateIPListEntry(*entry)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating trusted list entry",
			"Could not create trusted list entry, unexpected error: "+err.Error(),
		)
		return
	}
	var state trustedListEntryResourceModel
	diags = state.fromSFTPGo(ctx, entry)
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
func (r *trustedListEntryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state trustedListEntryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	entry, err := r.client.GetIPListEntry(4, state.IPOrNet.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			// Resource has been deleted outside of Terraform, remove it from state
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading SFTPGo trusted list entry",
			"Could not read SFTPGo trusted list entry "+state.IPOrNet.ValueString()+": "+err.Error(),
		)
		return
	}

	diags = state.fromSFTPGo(ctx, entry)
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
func (r *trustedListEntryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan trustedListEntryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	entry, diags := plan.toSFTPGo(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateIPListEntry(*entry)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating trusted list entry",
			"Could not update trusted list entry, unexpected error: "+err.Error(),
		)
		return
	}

	entry, err = r.client.GetIPListEntry(4, plan.IPOrNet.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SFTPGo trusted list entry",
			"Could not read SFTPGo trusted list entry "+plan.IPOrNet.ValueString()+": "+err.Error(),
		)
		return
	}

	var state trustedListEntryResourceModel
	diags = state.fromSFTPGo(ctx, entry)
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
func (r *trustedListEntryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state trustedListEntryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing entry
	err := r.client.DeleteIPListEntry(4, state.IPOrNet.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting SFTPGo trusted list entry",
			"Could not delete trusted list entry, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing the resource and save the Terraform state
func (*trustedListEntryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ipornet and save to ipornet attribute
	resource.ImportStatePassthroughID(ctx, path.Root("ipornet"), req, resp)
}
