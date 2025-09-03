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
	_ resource.Resource                = &allowListEntryResource{}
	_ resource.ResourceWithConfigure   = &allowListEntryResource{}
	_ resource.ResourceWithImportState = &allowListEntryResource{}
)

// NewAllowListEntryResource is a helper function to simplify the provider implementation.
func NewAllowListEntryResource() resource.Resource {
	return &allowListEntryResource{}
}

// allowListEntryResource is the resource implementation.
type allowListEntryResource struct {
	client *client.Client
}

// Configure adds the provider configured client to the resource.
func (r *allowListEntryResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Metadata returns the resource type name.
func (r *allowListEntryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_allowlist_entry"
}

// Schema defines the schema for the resource.
func (r *allowListEntryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Allow list entry",
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
				Description: `IP address or network in CIDR format, for example "192.168.1.2/32", "192.168.0.0/24", "2001:db8::/32"`,
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
func (r *allowListEntryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan allowListEntryResourceModel
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
			"Error creating allow list entry",
			"Could not create allow list entry, unexpected error: "+err.Error(),
		)
		return
	}
	var state allowListEntryResourceModel
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
func (r *allowListEntryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state allowListEntryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	entry, err := r.client.GetIPListEntry(1, state.IPOrNet.ValueString())
	if err != nil {
		// Check if the entry was not found (404 error)
		if statusErr, ok := err.(client.StatusError); ok && statusErr.StatusCode == 404 {
			// Resource has been deleted outside of Terraform, remove it from state
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading SFTPGo allow list entry",
			"Could not read SFTPGo allow list entry "+state.IPOrNet.ValueString()+": "+err.Error(),
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
func (r *allowListEntryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan allowListEntryResourceModel
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
			"Error updating allow list entry",
			"Could not update allow list entry, unexpected error: "+err.Error(),
		)
		return
	}

	entry, err = r.client.GetIPListEntry(1, plan.IPOrNet.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SFTPGo allow list entry",
			"Could not read SFTPGo allow list entry "+plan.IPOrNet.ValueString()+": "+err.Error(),
		)
		return
	}

	var state allowListEntryResourceModel
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
func (r *allowListEntryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state allowListEntryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing entry
	err := r.client.DeleteIPListEntry(1, state.IPOrNet.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting SFTPGo allow list entry",
			"Could not delete allow list entry, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing the resource and save the Terraform state
func (*allowListEntryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ipornet and save to ipornet attribute
	resource.ImportStatePassthroughID(ctx, path.Root("ipornet"), req, resp)
}
