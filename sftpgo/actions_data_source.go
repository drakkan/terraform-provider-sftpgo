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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &actionsDataSource{}
	_ datasource.DataSourceWithConfigure = &actionsDataSource{}
)

// NewActionsDataSource is a helper function to simplify the provider implementation.
func NewActionsDataSource() datasource.DataSource {
	return &actionsDataSource{}
}

// actionsDataSource is the data source implementation.
type actionsDataSource struct {
	client *client.Client
}

// Metadata returns the data source type name.
func (d *actionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_actions"
}

// Schema defines the schema for the data source.
func (d *actionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of event actions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Required to use the test framework. Just a placeholder.",
			},
			"actions": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of event actions.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Unique name.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Optional description.",
						},
						"type": schema.Int64Attribute{
							Computed:    true,
							Description: "Action type. 1 = HTTP, 2 = Command, 3 = Email, 4 = Backup, 5 = User quota reset, 6 = Folder quota reset, 7 = Transfer quota reset, 8 = Data retention check, 9 = Filesystem, 11 = Password expiration check, 12 = User expiration check, 13 = Identity Provider account check, 14 = User inactivity check, 15 = Rotate log file.",
						},
						"options": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Configuration options specific for the action type.",
							Attributes: map[string]schema.Attribute{
								"http_config": schema.SingleNestedAttribute{
									Computed:    true,
									Description: "HTTP action configurations.",
									Attributes: map[string]schema.Attribute{
										"endpoint": schema.StringAttribute{
											Computed:    true,
											Description: "HTTP endpoint to invoke.",
										},
										"username": schema.StringAttribute{
											Computed: true,
										},
										"password": schema.StringAttribute{
											Computed:    true,
											Description: computedSecretDescription,
										},
										"headers": schema.ListNestedAttribute{
											Computed:    true,
											Description: `Headers to add to the HTTP request.`,
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"key": schema.StringAttribute{
														Computed: true,
													},
													"value": schema.StringAttribute{
														Computed: true,
													},
												},
											},
										},
										"timeout": schema.Int64Attribute{
											Computed:    true,
											Description: "Time limit for the request in seconds. Ignored for multipart requests with files as attachments.",
										},
										"skip_tls_verify": schema.BoolAttribute{
											Computed:    true,
											Description: "If enabled any certificate presented by the server and any host name in that certificate are accepted. In this mode, TLS is susceptible to machine-in-the-middle attacks.",
										},
										"method": schema.StringAttribute{
											Computed:    true,
											Description: "HTTP method: GET, POST, PUT, DELETE.",
										},
										"query_parameters": schema.ListNestedAttribute{
											Computed:    true,
											Description: `Query parameters to add to the HTTP request.`,
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"key": schema.StringAttribute{
														Computed: true,
													},
													"value": schema.StringAttribute{
														Computed: true,
													},
												},
											},
										},
										"body": schema.StringAttribute{
											Computed:    true,
											Description: "Request body for POST/PUT.",
										},
										"parts": schema.ListNestedAttribute{
											Computed:    true,
											Description: `Multipart requests allow to combine one or more sets of data into a single body. For each part, you can set a file path or a body as text. Placeholders are supported in file path, body, header values.`,
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Computed: true,
													},
													"headers": schema.ListNestedAttribute{
														Computed: true,
														NestedObject: schema.NestedAttributeObject{
															Attributes: map[string]schema.Attribute{
																"key": schema.StringAttribute{
																	Computed: true,
																},
																"value": schema.StringAttribute{
																	Computed: true,
																},
															},
														},
													},
													"filepath": schema.StringAttribute{
														Computed:    true,
														Description: `Path to the file to be sent as an attachment.`,
													},
													"body": schema.StringAttribute{
														Computed: true,
													},
												},
											},
										},
									},
								},
								"cmd_config": schema.SingleNestedAttribute{
									Computed:    true,
									Description: "External command action configurations.",
									Attributes: map[string]schema.Attribute{
										"cmd": schema.StringAttribute{
											Computed:    true,
											Description: "Absolute path to the command to execute.",
										},
										"args": schema.ListAttribute{
											ElementType: types.StringType,
											Computed:    true,
											Description: "Command line arguments.",
										},
										"timeout": schema.Int64Attribute{
											Computed:    true,
											Description: "Time limit for the command in seconds.",
										},
										"env_vars": schema.ListNestedAttribute{
											Computed:    true,
											Description: "Environment variables to set for the external command.",
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"key": schema.StringAttribute{
														Computed: true,
													},
													"value": schema.StringAttribute{
														Computed: true,
													},
												},
											},
										},
									},
								},
								"email_config": schema.SingleNestedAttribute{
									Computed:    true,
									Description: "Email action configurations.",
									Attributes: map[string]schema.Attribute{
										"recipients": schema.ListAttribute{
											ElementType: types.StringType,
											Computed:    true,
										},
										"bcc": schema.ListAttribute{
											ElementType: types.StringType,
											Computed:    true,
										},
										"subject": schema.StringAttribute{
											Computed: true,
										},
										"body": schema.StringAttribute{
											Computed: true,
										},
										"content_type": schema.Int64Attribute{
											Computed:    true,
											Description: "1 means text/html 0 or omitted means text/plain.",
										},
										"attachments": schema.ListAttribute{
											ElementType: types.StringType,
											Computed:    true,
											Description: "Paths to attach. The total size is limited to 10 MB.",
										},
									},
								},
								"retention_config": schema.SingleNestedAttribute{
									Computed:    true,
									Description: "Data retention action configurations.",
									Attributes: map[string]schema.Attribute{
										"folders": schema.ListNestedAttribute{
											Computed:    true,
											Description: "Folders to apply data retention rules to.",
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"path": schema.StringAttribute{
														Computed:    true,
														Description: "Path for which to apply the retention rules.",
													},
													"retention": schema.Int64Attribute{
														Computed:    true,
														Description: "Retention as hours. 0 as retention means excluding the specified path.",
													},
													"delete_empty_dirs": schema.BoolAttribute{
														Computed:    true,
														Description: "If enabled, empty directories will be deleted.",
													},
												},
											},
										},
										"archive_folder": schema.StringAttribute{
											Computed:    true,
											Description: `Virtual folder name. If set, files will be moved there instead of being deleted. ` + enterpriseFeatureNote + ".",
										},
										"archive_path": schema.StringAttribute{
											Computed:    true,
											Description: `The base path where archived files will be stored. Placeholders are supported. ` + enterpriseFeatureNote + ".",
										},
									},
								},
								"fs_config": schema.SingleNestedAttribute{
									Computed:    true,
									Description: "Filesystem action configurations.",
									Attributes: map[string]schema.Attribute{
										"type": schema.Int64Attribute{
											Computed:    true,
											Description: `1 = Rename, 2 = Delete, 3 = Mkdir, 4 = Exist, 5 = Compress, 6 = Copy, 7 = PGP (` + enterpriseFeatureNote + `), ` + `8 Metadata Check (` + enterpriseFeatureNote + `).`,
										},
										"renames": schema.ListNestedAttribute{
											Computed:    true,
											Description: "Paths to rename. The key is the source path, the value is the target.",
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"key": schema.StringAttribute{
														Computed: true,
													},
													"value": schema.StringAttribute{
														Computed: true,
													},
													"update_modtime": schema.BoolAttribute{
														Optional:    true,
														Description: "Update modification time. This setting is not recursive and only applies to storage providers that support changing modification times.",
													},
												},
											},
										},
										"mkdirs": schema.ListAttribute{
											ElementType: types.StringType,
											Computed:    true,
											Description: "Directories paths to create.",
										},
										"deletes": schema.ListAttribute{
											ElementType: types.StringType,
											Computed:    true,
											Description: "Paths to delete.",
										},
										"exist": schema.ListAttribute{
											ElementType: types.StringType,
											Computed:    true,
											Description: "Paths to check for existence.",
										},
										"copy": schema.ListNestedAttribute{
											Computed:    true,
											Description: "Paths to copy. The key is the source path, the value is the target.",
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"key": schema.StringAttribute{
														Computed: true,
													},
													"value": schema.StringAttribute{
														Computed: true,
													},
												},
											},
										},
										"compress": schema.SingleNestedAttribute{
											Computed:    true,
											Description: "Configuration for paths to compress as zip.",
											Attributes: map[string]schema.Attribute{
												"name": schema.StringAttribute{
													Computed:    true,
													Description: `Full path to the zip file.`,
												},
												"paths": schema.ListAttribute{
													ElementType: types.StringType,
													Computed:    true,
													Description: "Paths to include in the compressed archive.",
												},
											},
										},
										"pgp": schema.SingleNestedAttribute{
											Computed:    true,
											Description: "Configuration for PGP actions. " + enterpriseFeatureNote + ".",
											Attributes: map[string]schema.Attribute{
												"mode": schema.Int64Attribute{
													Computed:    true,
													Description: `1 = Encrypt, 2 = Decrypt.`,
												},
												"profile": schema.Int64Attribute{
													Computed:    true,
													Description: `0 = Default, 1 = RFC 4880, 2 = RFC 9580. Don't set to use the default.`,
												},
												"paths": schema.ListNestedAttribute{
													Computed:    true,
													Description: "Paths to encrypt or decrypt.",
													NestedObject: schema.NestedAttributeObject{
														Attributes: map[string]schema.Attribute{
															"key": schema.StringAttribute{
																Computed: true,
															},
															"value": schema.StringAttribute{
																Computed: true,
															},
														},
													},
												},
												"password": schema.StringAttribute{
													Computed:    true,
													Description: computedSecretDescription,
												},
												"private_key": schema.StringAttribute{
													Computed:    true,
													Description: computedSecretDescription,
												},
												"passphrase": schema.StringAttribute{
													Computed:    true,
													Description: computedSecretDescription,
												},
												"public_key": schema.StringAttribute{
													Computed: true,
												},
											},
										},
										"metadata_check": schema.SingleNestedAttribute{
											Computed:    true,
											Description: "Configuration for Metadata Check actions. " + enterpriseFeatureNote + ".",
											Attributes: map[string]schema.Attribute{
												"path": schema.StringAttribute{
													Computed: true,
												},
												"metadata": schema.SingleNestedAttribute{
													Computed: true,
													Attributes: map[string]schema.Attribute{
														"key": schema.StringAttribute{
															Computed: true,
														},
														"value": schema.StringAttribute{
															Computed: true,
														},
													},
												},
												"timeout": schema.Int64Attribute{
													Computed: true,
												},
											},
										},
										"folder": schema.StringAttribute{
											Computed:    true,
											Description: enterpriseFeatureNote + ".",
										},
										"target_folder": schema.StringAttribute{
											Computed:    true,
											Description: enterpriseFeatureNote + ".",
										},
									},
								},
								"pwd_expiration_config": schema.SingleNestedAttribute{
									Computed:    true,
									Description: "Password expiration action configurations.",
									Attributes: map[string]schema.Attribute{
										"threshold": schema.Int64Attribute{
											Computed:    true,
											Description: `An email notification will be generated for users whose password expires in a number of days less than or equal to this threshold.`,
										},
									},
								},
								"user_inactivity_config": schema.SingleNestedAttribute{
									Computed:    true,
									Description: "User inactivity check configurations.",
									Attributes: map[string]schema.Attribute{
										"disable_threshold": schema.Int64Attribute{
											Computed:    true,
											Description: `Inactivity in days, since the last login before disabling the account.`,
										},
										"delete_threshold": schema.Int64Attribute{
											Computed:    true,
											Description: `Inactivity in days, since the last login before deleting the account.`,
										},
									},
								},
								"idp_config": schema.SingleNestedAttribute{
									Computed:    true,
									Description: "Identity Provider account check action configurations.",
									Attributes: map[string]schema.Attribute{
										"mode": schema.Int64Attribute{
											Computed:    true,
											Description: `0 means create or update the account, 1 means create the account if it doesn't exist.`,
										},
										"template_user": schema.StringAttribute{
											Computed:    true,
											Description: `SFTPGo user template in JSON format.`,
										},
										"template_admin": schema.StringAttribute{
											Computed:    true,
											Description: `SFTPGo admin template in JSON format.`,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *actionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*client.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *actionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state actionsDataSourceModel

	actions, err := d.client.GetActions()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read SFTPGo Actions",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, action := range actions {
		var actionState eventActionResourceModel
		diags := actionState.fromSFTPGo(ctx, &action)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		state.Actions = append(state.Actions, actionState)
	}

	state.ID = types.StringValue(placeholderID)

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// actionsDataSourceModel maps the data source schema data.
type actionsDataSourceModel struct {
	ID      types.String               `tfsdk:"id"`
	Actions []eventActionResourceModel `tfsdk:"actions"`
}
