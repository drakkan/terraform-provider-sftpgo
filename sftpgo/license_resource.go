// Copyright (C) 2025 Nicola Murino
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &licenseResource{}
	_ resource.ResourceWithConfigure   = &licenseResource{}
	_ resource.ResourceWithImportState = &licenseResource{}
)

// NewLicenseResource is a helper function to simplify the provider implementation.
func NewLicenseResource() resource.Resource {
	return &licenseResource{}
}

// licenseResource is the resource implementation.
type licenseResource struct {
	client *client.Client
}

// Configure adds the provider configured client to the resource.
func (r *licenseResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Metadata returns the resource type name.
func (r *licenseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_license"
}

// Schema defines the schema for the resource.
func (r *licenseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "License\n\n" +
			"Note: This resource cannot be deleted. Running `terraform destroy` will remove it from Terraform state only, but not from the underlying system. " +
			enterpriseFeatureNote + ".",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Required to use the test framework. Matches the license key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "License key. Required to add or update a license key.",
			},
			"type": schema.Int64Attribute{
				Computed:    true,
				Description: "License type: 0 = Disabled, 1 = Subscription, 2 = Lifetime",
			},
			"valid_from": schema.Int64Attribute{
				Computed:    true,
				Description: "Validity start time in Unix timestamp (milliseconds).",
			},
			"valid_to": schema.Int64Attribute{
				Computed:    true,
				Description: "Validity end time in Unix timestamp (milliseconds).",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *licenseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan licenseResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.Key.ValueString() == "" {
		resp.Diagnostics.AddError("Missing license key", "The 'key' attribute is required to add or update the license.")
		return
	}

	license, diags := plan.toSFTPGo(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	license, err := r.client.AddLicense(license.Key)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding license",
			"Could not create license, unexpected error: "+err.Error(),
		)
		return
	}
	var state licenseResourceModel
	diags = state.fromSFTPGo(ctx, license)
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
func (r *licenseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state licenseResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	license, err := r.client.GetLicense()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SFTPGo License",
			"Could not read SFTPGo License "+state.Key.ValueString()+": "+err.Error(),
		)
		return
	}

	diags = state.fromSFTPGo(ctx, license)
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
func (r *licenseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan licenseResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Key.ValueString() == "" {
		resp.Diagnostics.AddError("Missing license key", "The 'key' attribute is required to add or update the license.")
		return
	}

	license, diags := plan.toSFTPGo(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	license, err := r.client.AddLicense(license.Key)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating license",
			"Could not update license, unexpected error: "+err.Error(),
		)
		return
	}

	var state licenseResourceModel
	diags = state.fromSFTPGo(ctx, license)
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
func (r *licenseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state licenseResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Since the license cannot be deleted, we simply remove it from Terraform state.
}

// ImportState imports an existing the resource and save the Terraform state
func (*licenseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import name and save to name attribute
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}
