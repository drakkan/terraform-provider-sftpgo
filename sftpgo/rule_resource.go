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

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
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
	_ resource.Resource                = &ruleResource{}
	_ resource.ResourceWithConfigure   = &ruleResource{}
	_ resource.ResourceWithImportState = &ruleResource{}
)

// NewRuleResource is a helper function to simplify the provider implementation.
func NewRuleResource() resource.Resource {
	return &ruleResource{}
}

// ruleResource is the resource implementation.
type ruleResource struct {
	client *client.Client
}

// Configure adds the provider configured client to the resource.
func (r *ruleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Metadata returns the resource type name.
func (r *ruleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rule"
}

// Schema defines the schema for the resource.
func (r *ruleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Event rule",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Required to use the test framework. Matches the rule name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique name.",
			},
			"status": schema.Int64Attribute{
				Required:    true,
				Description: "1 enabled, 0 disabled.",
				Validators: []validator.Int64{
					int64validator.Between(0, 1),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Optional description.",
			},
			"trigger": schema.Int64Attribute{
				Required:    true,
				Description: "Event trigger. 1 = Filesystem event, 2 = Provider event, 3 = Schedule, 4 = IP Blocked, 5 = Certificate renewal, 6 = On demand, 7 = Identity Provider login.",
				Validators: []validator.Int64{
					int64validator.Between(1, 7),
				},
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
			"conditions": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Defines the conditions that trigger the rule.",
				Attributes: map[string]schema.Attribute{
					"fs_events": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: `Filesystem events that trigger the rule. Supported values: "upload", "pre-upload", "first-upload", "download", "pre-download", "first-download", "delete", "pre-delete", "rename", "mkdir", "rmdir", "copy", "ssh_cmd"`,
						Validators: []validator.List{
							listvalidator.UniqueValues(),
							listvalidator.ValueStringsAre(stringvalidator.OneOf("upload", "pre-upload", "first-upload", "download",
								"pre-download", "first-download", "delete", "pre-delete", "rename", "mkdir", "rmdir", "copy",
								"ssh_cmd")),
						},
					},
					"provider_events": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: `Provider events that trigger the rule. Supported values: "add", "update", "delete".`,
						Validators: []validator.List{
							listvalidator.UniqueValues(),
							listvalidator.ValueStringsAre(stringvalidator.OneOf("add", "update", "delete")),
						},
					},
					"schedules": schema.ListNestedAttribute{
						Optional:    true,
						Description: "List of schedules that trigger the rule. Hours: 0-23. Day of week: 0-6 (Sun-Sat). Day of month: 1-31. Month: 1-12. Asterisk (*) indicates a match for all the values of the field. e.g. every day of week, every day of month and so on.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"hour": schema.StringAttribute{
									Required: true,
								},
								"day_of_week": schema.StringAttribute{
									Required: true,
								},
								"day_of_month": schema.StringAttribute{
									Required: true,
								},
								"month": schema.StringAttribute{
									Required: true,
								},
							},
						},
					},
					"idp_login_event": schema.Int64Attribute{
						Optional:    true,
						Description: `Identity Provider login event that trigger the rule. 0 any, 1 user, 2 admin.`,
						Validators: []validator.Int64{
							int64validator.Between(0, 2),
						},
					},
					"options": schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Options for event conditions.",
						Attributes: map[string]schema.Attribute{
							"names": schema.ListNestedAttribute{
								Optional:    true,
								Description: `Shell-like pattern filters for usernames, folder names. For example "user*"" will match names starting with "user". For provider events, this filter is applied to the username of the admin executing the event.`,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"pattern": schema.StringAttribute{
											Required: true,
										},
										"inverse_match": schema.BoolAttribute{
											Optional: true,
										},
									},
								},
							},
							"group_names": schema.ListNestedAttribute{
								Optional:    true,
								Description: `Shell-like pattern filters for group names. For example "group*"" will match group names starting with "group".`,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"pattern": schema.StringAttribute{
											Required: true,
										},
										"inverse_match": schema.BoolAttribute{
											Optional: true,
										},
									},
								},
							},
							"role_names": schema.ListNestedAttribute{
								Optional:    true,
								Description: `Shell-like pattern filters for role names. For example "role*"" will match role names starting with "role".`,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"pattern": schema.StringAttribute{
											Required: true,
										},
										"inverse_match": schema.BoolAttribute{
											Optional: true,
										},
									},
								},
							},
							"fs_paths": schema.ListNestedAttribute{
								Optional:    true,
								Description: `Shell-like pattern filters for filesystem events. For example "/adir/*.txt"" will match paths in the "/adir" directory ending with ".txt". Double asterisk is supported, for example "/**/*.txt" will match any file ending with ".txt". "/mydir/**" will match any entry in "/mydir".`,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"pattern": schema.StringAttribute{
											Required: true,
										},
										"inverse_match": schema.BoolAttribute{
											Optional: true,
										},
									},
								},
							},
							"protocols": schema.ListAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: `The filesystem event rule will be triggered only for the specified protocols. Empty means any protocol. Supported values: "SFTP", "SCP", "SSH", "FTP", "DAV", "HTTP", "HTTPShare","OIDC"`,
								Validators: []validator.List{
									listvalidator.UniqueValues(),
									listvalidator.ValueStringsAre(stringvalidator.OneOf("SFTP", "SCP", "SSH", "FTP", "DAV", "HTTP",
										"HTTPShare", "OIDC")),
								},
							},
							"provider_objects": schema.ListAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: `The provider event rule will be triggered only for the specified provider objects. Empty means any provider object. Supported values: "user", "folder", "group", "admin", "api_key", "share", "event_rule", "event_action".`,
								Validators: []validator.List{
									listvalidator.UniqueValues(),
									listvalidator.ValueStringsAre(stringvalidator.OneOf("user", "folder", "group", "admin", "api_key", "share",
										"event_rule", "event_action")),
								},
							},
							"min_size": schema.Int64Attribute{
								Optional:    true,
								Description: `Minimum file size as bytes.`,
								Validators: []validator.Int64{
									int64validator.AtLeast(0),
								},
							},
							"max_size": schema.Int64Attribute{
								Optional:    true,
								Description: `Maximum file size as bytes.`,
								Validators: []validator.Int64{
									int64validator.AtLeast(0),
								},
							},
							"event_statuses": schema.ListAttribute{
								ElementType: types.Int32Type,
								Optional:    true,
								Description: `The filesystem event rules will be triggered only for actions with the specified status. Empty means any status. Suported values: 1 (OK), 2 (Failed), 3 (Failed for a quota exceeded error).`,
								Validators: []validator.List{
									listvalidator.UniqueValues(),
									listvalidator.ValueInt32sAre(int32validator.OneOf(1, 2, 3)),
								},
							},
							"concurrent_execution": schema.BoolAttribute{
								Optional:    true,
								Description: `If enabled, allow to execute scheduled tasks concurrently from multiple SFTPGo instances.`,
							},
						},
					},
				},
			},
			"actions": schema.ListNestedAttribute{
				Required:    true,
				Description: `List of actions to execute.`,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true,
						},
						"is_failure_action": schema.BoolAttribute{
							Optional: true,
						},
						"stop_on_failure": schema.BoolAttribute{
							Optional: true,
						},
						"execute_sync": schema.BoolAttribute{
							Optional:    true,
							Description: `Supported for upload events and required for pre-* events and Identity provider login events if the action checks the account.`,
						},
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *ruleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan eventRuleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, diags := plan.toSFTPGo(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.CreateRule(*rule)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating rule",
			"Could not create rule, unexpected error: "+err.Error(),
		)
		return
	}
	var state eventRuleResourceModel
	diags = state.fromSFTPGo(ctx, rule)
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
func (r *ruleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state eventRuleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.GetRule(state.Name.ValueString())
	if err != nil {
		// Check if the rule was not found (404 error)
		if statusErr, ok := err.(client.StatusError); ok && statusErr.StatusCode == 404 {
			// Resource has been deleted outside of Terraform, remove it from state
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading SFTPGo Event Rule",
			"Could not read SFTPGo Event Rule "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	diags = state.fromSFTPGo(ctx, rule)
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
func (r *ruleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan eventRuleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	rule, diags := plan.toSFTPGo(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateRule(*rule)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating event rule",
			"Could not update event rule, unexpected error: "+err.Error(),
		)
		return
	}

	rule, err = r.client.GetRule(plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SFTPGo Event Rule",
			"Could not read SFTPGo Event Rule "+plan.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	var state eventRuleResourceModel
	diags = state.fromSFTPGo(ctx, rule)
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
func (r *ruleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state eventRuleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing event rule
	err := r.client.DeleteRule(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting SFTPGo event rule",
			"Could not delete event rule, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing the resource and save the Terraform state
func (*ruleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import name and save to name attribute
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
