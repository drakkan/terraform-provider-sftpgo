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
				Required: true,
				MarkdownDescription: "Action type.\n\n" +
					"Supported values:\n" +
					"* `1`: HTTP\n" +
					"* `2`: Command\n" +
					"* `3`: Email\n" +
					"* `4`: Backup\n" +
					"* `5`: User quota reset\n" +
					"* `6`: Folder quota reset\n" +
					"* `7`: Transfer quota reset\n" +
					"* `8`: Data retention check\n" +
					"* `9`: Filesystem\n" +
					"* `11`: Password expiration check\n" +
					"* `12`: User expiration check\n" +
					"* `13`: Identity Provider account check\n" +
					"* `14`: User inactivity check\n" +
					"* `15`: Rotate log file\n" +
					"* `16`: IMAP\n" +
					"* `17`: ICAP\n" +
					"* `18`: Share expiration check",
				Validators: []validator.Int64{
					int64validator.Any(
						int64validator.Between(1, 9),
						int64validator.Between(11, 18),
					),
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
									stringvalidator.OneOf(http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete),
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
							"bcc": schema.ListAttribute{
								ElementType: types.StringType,
								Optional:    true,
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
									},
								},
							},
							"archive_folder": schema.StringAttribute{
								Optional:    true,
								Description: `Virtual folder name. If set, files will be moved there instead of being deleted. ` + enterpriseFeatureNote + ".",
							},
							"archive_path": schema.StringAttribute{
								Optional:    true,
								Description: `The base path where archived files will be stored. Placeholders are supported. ` + enterpriseFeatureNote + ".",
							},
						},
					},
					"fs_config": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Filesystem action configurations.",
						Attributes: map[string]schema.Attribute{
							"type": schema.Int64Attribute{
								Required:    true,
								Description: `1 = Rename, 2 = Delete, 3 = Mkdir, 4 = Exist, 5 = Compress, 6 = Copy, 7 = PGP (` + enterpriseFeatureNote + `), ` + `8 Metadata Check (` + enterpriseFeatureNote + `), ` + `9 Decompress (` + enterpriseFeatureNote + `).`,
								Validators: []validator.Int64{
									int64validator.Between(1, 9),
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
										"update_modtime": schema.BoolAttribute{
											Optional:    true,
											Description: "Update modification time. This setting is not recursive and only applies to storage providers that support changing modification times.",
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
							"decompress": schema.SingleNestedAttribute{
								Optional:    true,
								Description: "Configuration for archive to extract. " + enterpriseFeatureNote,
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Required:    true,
										Description: `Full path to the zip file.`,
									},
									"extract_dir": schema.StringAttribute{
										Required:    true,
										Description: `Directory to extract the archive into.`,
									},
								},
							},
							"pgp": schema.SingleNestedAttribute{
								Optional:    true,
								Description: "Configuration for PGP actions. Either a password or a key pair is required. For encryption, the public key is required, and the private, if provided, will be used for signing. For decryption, the private key is required, and the public key, if provided, will be used for signature verification. " + enterpriseFeatureNote + ".",
								Attributes: map[string]schema.Attribute{
									"mode": schema.Int64Attribute{
										Required:    true,
										Description: `1 = Encrypt, 2 = Decrypt.`,
										Validators: []validator.Int64{
											int64validator.Between(1, 2),
										},
									},
									"profile": schema.Int64Attribute{
										Optional:    true,
										Description: `Algorithms to use. 0 = Default (widely implemented algorithms), 1 = RFC 4880, 2 = RFC 9580. Don't set to use the default.`,
										Validators: []validator.Int64{
											int64validator.Between(0, 2),
										},
									},
									"paths": schema.ListNestedAttribute{
										Required:    true,
										Description: "Paths to encrypt or decrypt.",
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
									"password": schema.StringAttribute{
										Optional:    true,
										Sensitive:   true,
										Description: computedSecretDescription,
									},
									"private_key": schema.StringAttribute{
										Optional:    true,
										Sensitive:   true,
										Description: computedSecretDescription,
									},
									"passphrase": schema.StringAttribute{
										Optional:    true,
										Sensitive:   true,
										Description: computedSecretDescription,
									},
									"public_key": schema.StringAttribute{
										Optional: true,
									},
								},
							},
							"metadata_check": schema.SingleNestedAttribute{
								Optional:    true,
								Description: "This action verifies whether the metadata key matches the configured value or is absent for the specified path. Optionally, it can retry periodically until the specified timeout (in seconds) is reached. " + enterpriseFeatureNote + ".",
								Attributes: map[string]schema.Attribute{
									"path": schema.StringAttribute{
										Required: true,
									},
									"metadata": schema.SingleNestedAttribute{
										Required: true,
										Attributes: map[string]schema.Attribute{
											"key": schema.StringAttribute{
												Required: true,
											},
											"value": schema.StringAttribute{
												Optional: true,
											},
										},
									},
									"timeout": schema.Int64Attribute{
										Optional: true,
									},
								},
							},
							"folder": schema.StringAttribute{
								Optional:    true,
								Description: "Actions triggered by filesystem events, such as uploads or downloads, use the filesystem associated with the user. By specifying a folder, you can control which filesystem is used. This is especially useful for events that aren't tied to a user, such as scheduled tasks and advanced workflows. " + enterpriseFeatureNote + ".",
							},
							"target_folder": schema.StringAttribute{
								Optional:    true,
								Description: "By specifying a target folder, you can use a different filesystem for target paths than the one associated with the user who triggered the action. This is useful for moving files to another storage backend, such as a different S3 bucket or an external SFTP server, accessing restricted areas of the same storage backend, supporting scheduled actions, or enabling more advanced workflows. " + enterpriseFeatureNote + ".",
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
					"user_inactivity_config": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "User inactivity check configurations.",
						Attributes: map[string]schema.Attribute{
							"disable_threshold": schema.Int64Attribute{
								Optional:    true,
								Description: `Inactivity in days, since the last login before disabling the account.`,
								Validators: []validator.Int64{
									int64validator.AtLeast(0),
								},
							},
							"delete_threshold": schema.Int64Attribute{
								Optional:    true,
								Description: `Inactivity in days, since the last login before deleting the account.`,
								Validators: []validator.Int64{
									int64validator.AtLeast(0),
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
					"imap_config": schema.SingleNestedAttribute{
						Optional:            true,
						MarkdownDescription: "Enables automatic retrieval of email attachments from IMAP mailboxes. " + enterpriseFeatureNote,
						Attributes: map[string]schema.Attribute{
							"endpoint": schema.StringAttribute{
								Required:            true,
								MarkdownDescription: "IMAP endpoint in the format `schema://host:port`. Supported schemas: `imap`, `imaps`.",
							},
							"username": schema.StringAttribute{
								Optional: true,
							},
							"password": schema.StringAttribute{
								Optional:    true,
								Sensitive:   true,
								Description: computedSecretDescription,
							},
							"auth_type": schema.Int64Attribute{
								Optional: true,
								MarkdownDescription: "Authentication type.\n\n" +
									"Supported values:\n" +
									"* `0`: Login\n" +
									"* `1`: Plain",
								Validators: []validator.Int64{
									int64validator.Between(0, 1),
								},
							},
							"oauth2": schema.SingleNestedAttribute{
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"provider": schema.Int64Attribute{
										Optional: true,
										MarkdownDescription: "OAuth2 Provider.\n\n" +
											"Supported values:\n" +
											"* `0`: Google\n" +
											"* `1`: Microsoft",
										Validators: []validator.Int64{
											int64validator.Between(0, 1),
										},
									},
									"tenant": schema.StringAttribute{
										Optional:    true,
										Description: "OAuth2 Tenant.",
									},
									"client_id": schema.StringAttribute{
										Optional:    true,
										Description: "OAuth2 Client ID.",
									},
									"client_secret": schema.StringAttribute{
										Optional:    true,
										Sensitive:   true,
										Description: computedSecretDescription,
									},
									"refresh_token": schema.StringAttribute{
										Optional:    true,
										Sensitive:   true,
										Description: computedSecretDescription,
									},
								},
							},
							"mailbox": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "Mailbox to check, e.g. `INBOX`.",
							},
							"path": schema.StringAttribute{
								Optional:    true,
								Description: "Directory path where downloaded attachments will be stored.",
							},
							"post_process_action": schema.Int64Attribute{
								Optional: true,
								MarkdownDescription: "Action to perform on the email after processing.\n\n" +
									"Supported values:\n" +
									"* `0`: Mark messages as read\n" +
									"* `1`: Delete messages",
								Validators: []validator.Int64{
									int64validator.Between(0, 1),
								},
							},
							"target_folder": schema.StringAttribute{
								Optional: true,
								MarkdownDescription: "If specified, attachments are stored directly in this target virtual folder. " +
									"If not specified, attachments are downloaded to the user account associated with the action.\n\n" +
									"**Note:** IMAP actions are executed for a single user only, so a target folder or an appropriate filter in the related rule is recommended.",
							},
						},
					},
					"icap_config": schema.SingleNestedAttribute{
						Optional:            true,
						MarkdownDescription: "Enables integration with ICAP servers to perform antivirus scanning and DLP checks. " + enterpriseFeatureNote,
						Attributes: map[string]schema.Attribute{
							"endpoint": schema.StringAttribute{
								Required: true,
								MarkdownDescription: "ICAP endpoint in the format `schema://host:port`. Supported schemas: `icap`, `icaps`. " +
									"If the port is omitted, port `1344` is used.",
							},
							"timeout": schema.Int64Attribute{
								Optional:    true,
								Description: "Timeout in seconds for ICAP requests.",
							},
							"skip_tls_verify": schema.BoolAttribute{
								Optional: true,
								MarkdownDescription: "If enabled, any certificate presented by the server and any host name in that certificate are accepted. " +
									"**Warning:** In this mode, TLS is susceptible to machine-in-the-middle attacks.",
							},
							"paths": schema.ListAttribute{
								ElementType:         types.StringType,
								Required:            true,
								MarkdownDescription: "List of virtual paths to scan. Placeholders are supported, e.g. `{{.VirtualPath}}`.",
								Validators: []validator.List{
									listvalidator.UniqueValues(),
								},
							},
							"method": schema.StringAttribute{
								Required:            true,
								MarkdownDescription: "ICAP method to use. Currently `REQMOD` is supported.",
								Validators: []validator.String{
									stringvalidator.OneOf("REQMOD"),
								},
							},
							"headers": schema.ListNestedAttribute{
								Optional:    true,
								Description: "Headers to add to the ICAP request.",
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
							"block_action": schema.Int64Attribute{
								Optional: true,
								MarkdownDescription: "Action executed when the ICAP server detects an infected file.\n\n" +
									"Supported values:\n" +
									"* `1`: Ignore (allow file)\n" +
									"* `2`: Delete (reject file)\n" +
									"* `3`: Quarantine",
								Validators: []validator.Int64{
									int64validator.Between(1, 3),
								},
							},
							"adapt_action": schema.Int64Attribute{
								Optional: true,
								MarkdownDescription: "Action executed when the ICAP server returns a modified file.\n\n" +
									"Supported values:\n" +
									"* `1`: Ignore (keep original)\n" +
									"* `2`: Delete (reject)\n" +
									"* `3`: Quarantine\n" +
									"* `4`: Overwrite (use modified file)",
								Validators: []validator.Int64{
									int64validator.Between(1, 4),
								},
							},
							"failure_policy": schema.Int64Attribute{
								Optional: true,
								MarkdownDescription: "Action executed when the ICAP scan fails (e.g. server unreachable).\n\n" +
									"Supported values:\n" +
									"* `1`: Ignore (allow file)\n" +
									"* `2`: Delete (reject file)\n" +
									"* `3`: Quarantine",
								Validators: []validator.Int64{
									int64validator.Between(1, 3),
								},
							},
							"quarantine_folder": schema.StringAttribute{
								Optional:    true,
								Description: "The name of the virtual folder where quarantined files will be stored.",
							},
							"quarantine_path": schema.StringAttribute{
								Optional:    true,
								Description: "The virtual path where quarantined files will be stored. Placeholders are supported.",
							},
						},
					},
					"share_expiration_config": schema.SingleNestedAttribute{
						Optional:            true,
						MarkdownDescription: "Automated lifecycle management for shares based on inactivity, expiration, or max tokens. " + enterpriseFeatureNote,
						Attributes: map[string]schema.Attribute{
							"advance_notice": schema.Int64Attribute{
								Optional:    true,
								Description: "Number of days before the share expiration (real or calculated) to trigger the expiration event. Set to 0 to disable.",
							},
							"grace_period": schema.Int64Attribute{
								Optional:    true,
								Description: "Number of days a share is kept in the database after it has expired. Set to 0 to disable share deletion.",
							},
							"inactivity_threshold": schema.Int64Attribute{
								Optional:    true,
								Description: "Validity period (in days) for shares without an explicit expiration date. Checks last use or creation time. Set to 0 to disable.",
							},
							"split_events": schema.BoolAttribute{
								Optional: true,
								MarkdownDescription: "If true, events are split. " +
									"For example, email actions will send a separate notification for each share instead of a cumulative report.",
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
		if client.IsNotFound(err) {
			// Resource has been deleted outside of Terraform, remove it from state
			resp.State.RemoveResource(ctx)
			return
		}
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
		if client.IsNotFound(err) {
			return
		}
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
	// only HTTP and PGP config have a secret to preserve
	actionType := plan.Type.ValueInt64()
	if actionType != 1 && actionType != 9 && actionType != 16 {
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

	if actionType == 1 && optionsPlan.HTTPConfig != nil {
		optionsState.HTTPConfig.Password = optionsPlan.HTTPConfig.Password
	}
	if actionType == 9 && optionsPlan.FsConfig != nil && optionsPlan.FsConfig.Type.ValueInt64() == 7 {
		optionsState.FsConfig.PGP.Password = optionsPlan.FsConfig.PGP.Password
		optionsState.FsConfig.PGP.PrivateKey = optionsPlan.FsConfig.PGP.PrivateKey
		optionsState.FsConfig.PGP.Passphrase = optionsPlan.FsConfig.PGP.Passphrase
	}
	if actionType == 16 && optionsPlan.IMAPConfig != nil {
		optionsState.IMAPConfig.Password = optionsPlan.IMAPConfig.Password
		if optionsPlan.IMAPConfig.OAuth2 != nil {
			optionsState.IMAPConfig.OAuth2.ClientSecret = optionsPlan.IMAPConfig.OAuth2.ClientSecret
			optionsState.IMAPConfig.OAuth2.RefreshToken = optionsPlan.IMAPConfig.OAuth2.RefreshToken
		}
	}

	optionsStateObj, diags := types.ObjectValueFrom(ctx, optionsState.getTFAttributes(), optionsState)
	if diags.HasError() {
		return diags
	}
	state.Options = optionsStateObj

	return nil
}
