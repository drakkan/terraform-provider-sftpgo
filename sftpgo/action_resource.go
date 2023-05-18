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
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &actionResource{}
	_ resource.ResourceWithConfigure   = &actionResource{}
	_ resource.ResourceWithImportState = &actionResource{}
)

// NewActionResource is a helper function to simplify the provider implementation.
func NewActionResource() resource.Resource {
	return &actionResource{}
}

// actionResource is the resource implementation.
type actionResource struct {
	client *client.Client
}

// Configure adds the provider configured client to the resource.
func (r *actionResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Metadata returns the resource type name.
func (r *actionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_action"
}

// Schema defines the schema for the resource.
func (r *actionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Event action",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Required to use the test framework. Matches the action name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Optional description.",
			},
			"type": schema.Int64Attribute{
				Required:    true,
				Description: "Action type. 1 = HTTP, 2 = Command, 3 = Email, 4 = Backup, 5 = User quota reset, 6 = Folder quota reset, 7 = Transfer quota reset, 8 = Data retention check, 9 = Filesystem, 10 = Metadata check, 11 = Password expiration check, 12 = User expiration check, 13 = Identity Provider account check.",
				Validators: []validator.Int64{
					int64validator.Between(1, 13),
				},
			},
			"options": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Configuration options specific for the action type.",
				Attributes: map[string]schema.Attribute{
					"http_config": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "HTTP action configurations.",
						Attributes: map[string]schema.Attribute{
							"endpoint": schema.StringAttribute{
								Required:    true,
								Description: "HTTP endpoint to invoke.",
							},
							"username": schema.StringAttribute{
								Optional: true,
							},
							"password": schema.StringAttribute{
								Optional:    true,
								Sensitive:   true,
								Description: computedSecretDescription,
							},
							"headers": schema.ListNestedAttribute{
								Optional:    true,
								Description: `Headers to add to the HTTP request.`,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"key": schema.StringAttribute{
											Required: true,
										},
										"value": schema.StringAttribute{
											Required: true,
										},
									},
								},
							},
							"timeout": schema.Int64Attribute{
								Optional:    true,
								Description: "Time limit for the request in seconds. Ignored for multipart requests with files as attachments. For non multipart requests is required Ignored for multipart requests with files as attachments otherwise required and must be between 1 and 120",
							},
							"skip_tls_verify": schema.BoolAttribute{
								Optional:    true,
								Description: "If enabled any certificate presented by the server and any host name in that certificate are accepted. In this mode, TLS is susceptible to machine-in-the-middle attacks.",
							},
							"method": schema.StringAttribute{
								Required:    true,
								Description: "HTTP method.",
								Validators: []validator.String{
									stringvalidator.OneOf(http.MethodGet, http.MethodPost, http.MethodPut),
								},
							},
							"query_parameters": schema.ListNestedAttribute{
								Optional:    true,
								Description: `Query parameters to add to the HTTP request.`,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"key": schema.StringAttribute{
											Required: true,
										},
										"value": schema.StringAttribute{
											Required: true,
										},
									},
								},
							},
							"body": schema.StringAttribute{
								Optional:    true,
								Description: "Request body for POST/PUT.",
							},
							"parts": schema.ListNestedAttribute{
								Optional:    true,
								Description: `Multipart requests allow to combine one or more sets of data into a single body. For each part, you can set a file path or a body as text. Placeholders are supported in file path, body, header values.`,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Required: true,
										},
										"headers": schema.ListNestedAttribute{
											Optional: true,
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"key": schema.StringAttribute{
														Required: true,
													},
													"value": schema.StringAttribute{
														Required: true,
													},
												},
											},
										},
										"filepath": schema.StringAttribute{
											Optional:    true,
											Description: `Path to the file to be sent as an attachment.`,
										},
										"body": schema.StringAttribute{
											Optional: true,
										},
									},
								},
							},
						},
					},
					"cmd_config": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "External command action configurations.",
						Attributes: map[string]schema.Attribute{
							"cmd": schema.StringAttribute{
								Required:    true,
								Description: "Absolute path to the command to execute.",
							},
							"args": schema.ListAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: "Command line arguments.",
							},
							"timeout": schema.Int64Attribute{
								Required:    true,
								Description: "Time limit for the command in seconds.",
								Validators: []validator.Int64{
									int64validator.Between(1, 120),
								},
							},
							"env_vars": schema.ListNestedAttribute{
								Optional:    true,
								Description: "Environment variables to set for the external command.",
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"key": schema.StringAttribute{
											Required: true,
										},
										"value": schema.StringAttribute{
											Required: true,
										},
									},
								},
							},
						},
					},
					"email_config": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Email action configurations.",
						Attributes: map[string]schema.Attribute{
							"recipients": schema.ListAttribute{
								ElementType: types.StringType,
								Required:    true,
								Validators: []validator.List{
									listvalidator.UniqueValues(),
								},
							},
							"subject": schema.StringAttribute{
								Required: true,
							},
							"content_type": schema.Int64Attribute{
								Optional:    true,
								Description: "Optional content type. 0 means text/plain, 1 means text/html. If omitted, text/plain is assumed.",
								Validators: []validator.Int64{
									int64validator.Between(0, 1),
								},
							},
							"body": schema.StringAttribute{
								Required: true,
							},
							"attachments": schema.ListAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: "Paths to attach. The total size is limited to 10 MB.",
								Validators: []validator.List{
									listvalidator.UniqueValues(),
								},
							},
						},
					},
					"retention_config": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Data retention action configurations.",
						Attributes: map[string]schema.Attribute{
							"folders": schema.ListNestedAttribute{
								Optional:    true,
								Description: "Folders to apply data retention rules to.",
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"path": schema.StringAttribute{
											Required:    true,
											Description: "Path for which to apply the retention rules.",
										},
										"retention": schema.Int64Attribute{
											Required:    true,
											Description: "Retention as hours. 0 as retention means excluding the specified path.",
											Validators: []validator.Int64{
												int64validator.AtLeast(0),
											},
										},
										"delete_empty_dirs": schema.BoolAttribute{
											Optional:    true,
											Description: "If enabled, empty directories will be deleted.",
										},
										"ignore_user_permissions": schema.BoolAttribute{
											Optional:    true,
											Description: `Defines whether to delete files even if the user does not have the "delete" permission, by default files will be skipped if the user does not have the "delete" permission.`,
										},
									},
								},
							},
						},
					},
					"fs_config": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Filesystem action configurations.",
						Attributes: map[string]schema.Attribute{
							"type": schema.Int64Attribute{
								Required:    true,
								Description: `1 = Rename, 2 = Delete, 3 = Mkdir, 4 = Exist, 5 = Compress, 6 = Copy.`,
								Validators: []validator.Int64{
									int64validator.Between(1, 6),
								},
							},
							"renames": schema.ListNestedAttribute{
								Optional:    true,
								Description: "Paths to rename. The key is the source path, the value is the target.",
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"key": schema.StringAttribute{
											Required: true,
										},
										"value": schema.StringAttribute{
											Required: true,
										},
									},
								},
							},
							"mkdirs": schema.ListAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: "Directories paths to create.",
								Validators: []validator.List{
									listvalidator.UniqueValues(),
								},
							},
							"deletes": schema.ListAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: "Paths to delete.",
								Validators: []validator.List{
									listvalidator.UniqueValues(),
								},
							},
							"exist": schema.ListAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: "Paths to check for existence.",
								Validators: []validator.List{
									listvalidator.UniqueValues(),
								},
							},
							"copy": schema.ListNestedAttribute{
								Optional:    true,
								Description: "Paths to copy. The key is the source path, the value is the target.",
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"key": schema.StringAttribute{
											Required: true,
										},
										"value": schema.StringAttribute{
											Required: true,
										},
									},
								},
							},
							"compress": schema.SingleNestedAttribute{
								Optional:    true,
								Description: "Configuration for paths to compress as zip.",
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Required:    true,
										Description: `Full path to the zip file.`,
									},
									"paths": schema.ListAttribute{
										ElementType: types.StringType,
										Required:    true,
										Description: "Paths to include in the compressed archive.",
										Validators: []validator.List{
											listvalidator.UniqueValues(),
										},
									},
								},
							},
						},
					},
					"pwd_expiration_config": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Password expiration action configurations.",
						Attributes: map[string]schema.Attribute{
							"threshold": schema.Int64Attribute{
								Required:    true,
								Description: `An email notification will be generated for users whose password expires in a number of days less than or equal to this threshold.`,
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
								},
							},
						},
					},
					"idp_config": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Identity Provider account check action configurations.",
						Attributes: map[string]schema.Attribute{
							"mode": schema.Int64Attribute{
								Required:    true,
								Description: `0 means create or update the account, 1 means create the account if it doesn't exist.`,
								Validators: []validator.Int64{
									int64validator.Between(0, 1),
								},
							},
							"template_user": schema.StringAttribute{
								Optional:    true,
								Description: `SFTPGo user template in JSON format.`,
							},
							"template_admin": schema.StringAttribute{
								Optional:    true,
								Description: `SFTPGo admin template in JSON format.`,
							},
						},
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *actionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan eventActionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	action, diags := plan.toSFTPGo(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	action, err := r.client.CreateAction(*action)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating event action",
			"Could not create event action, unexpected error: "+err.Error(),
		)
		return
	}
	var state eventActionResourceModel
	diags = state.fromSFTPGo(ctx, action)
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
func (r *actionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state eventActionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	action, err := r.client.GetAction(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SFTPGo Event Action",
			"Could not read SFTPGo Event Action "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	var newState eventActionResourceModel
	diags = newState.fromSFTPGo(ctx, action)
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
func (r *actionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan eventActionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	action, diags := plan.toSFTPGo(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateAction(*action)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating event action",
			"Could not update event action, unexpected error: "+err.Error(),
		)
		return
	}

	action, err = r.client.GetAction(plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SFTPGo Event Action",
			"Could not read SFTPGo Event Action "+plan.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	var state eventActionResourceModel
	diags = state.fromSFTPGo(ctx, action)
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
func (r *actionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state eventActionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing action
	err := r.client.DeleteAction(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting SFTPGo Event Action",
			"Could not delete event action, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing the resource and save the Terraform state
func (*actionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import name and save to name attribute
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (*actionResource) preservePlanFields(ctx context.Context, plan, state *eventActionResourceModel) diag.Diagnostics {
	if plan.Options.IsNull() {
		return nil
	}
	// only HTTP config has a secret to preserve
	if plan.Type.ValueInt64() != 1 {
		return nil
	}

	var optionsPlan eventActionOptions
	diags := plan.Options.As(ctx, &optionsPlan, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return diags
	}

	var optionsState eventActionOptions
	diags = state.Options.As(ctx, &optionsState, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return diags
	}

	optionsState.HTTPConfig.Password = optionsPlan.HTTPConfig.Password
	optionsStateObj, diags := types.ObjectValueFrom(ctx, optionsState.getTFAttributes(), optionsState)
	if diags.HasError() {
		return diags
	}
	state.Options = optionsStateObj

	return nil
}
