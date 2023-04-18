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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &defenderEntryResource{}
	_ resource.ResourceWithConfigure   = &defenderEntryResource{}
	_ resource.ResourceWithImportState = &defenderEntryResource{}
)

// NewDefenderEntryResource is a helper function to simplify the provider implementation.
func NewDefenderEntryResource() resource.Resource {
	return &defenderEntryResource{}
}

// defenderEntryResource is the resource implementation.
type defenderEntryResource struct {
	client *client.Client
}

// Configure adds the provider configured client to the resource.
func (r *defenderEntryResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Metadata returns the resource type name.
func (r *defenderEntryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_defender_entry"
}

// Schema defines the schema for the resource.
func (r *defenderEntryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Defender entry",
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
			"mode": schema.Int64Attribute{
				Required:    true,
				Description: "1 = allow, 2 = deny.",
				Validators: []validator.Int64{
					int64validator.Between(1, 2),
				},
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
func (r *defenderEntryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan defenderEntryResourceModel
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
			"Error creating defender entry",
			"Could not create defender entry, unexpected error: "+err.Error(),
		)
		return
	}
	var state defenderEntryResourceModel
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
func (r *defenderEntryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state defenderEntryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	entry, err := r.client.GetIPListEntry(2, state.IPOrNet.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SFTPGo defender entry",
			"Could not read SFTPGo defender entry "+state.IPOrNet.ValueString()+": "+err.Error(),
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
func (r *defenderEntryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan defenderEntryResourceModel
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
			"Error updating defender entry",
			"Could not update defender entry, unexpected error: "+err.Error(),
		)
		return
	}

	entry, err = r.client.GetIPListEntry(2, plan.IPOrNet.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SFTPGo defender entry",
			"Could not read SFTPGo defender entry "+plan.IPOrNet.ValueString()+": "+err.Error(),
		)
		return
	}

	var state defenderEntryResourceModel
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
func (r *defenderEntryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state defenderEntryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing entry
	err := r.client.DeleteIPListEntry(2, state.IPOrNet.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting SFTPGo defender entry",
			"Could not delete defender entry, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing the resource and save the Terraform state
func (*defenderEntryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ipornet and save to ipornet attribute
	resource.ImportStatePassthroughID(ctx, path.Root("ipornet"), req, resp)
}
