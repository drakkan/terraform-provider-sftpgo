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
	_ datasource.DataSource              = &usersDataSource{}
	_ datasource.DataSourceWithConfigure = &usersDataSource{}
)

// NewUsersDataSource is a helper function to simplify the provider implementation.
func NewUsersDataSource() datasource.DataSource {
	return &usersDataSource{}
}

// usersDataSource is the data source implementation.
type usersDataSource struct {
	client *client.Client
}

// Metadata returns the data source type name.
func (d *usersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

// Schema defines the schema for the data source.
func (d *usersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of users.",
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of users.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"username": schema.StringAttribute{
							Computed:    true,
							Description: "Unique username.",
						},
						"status": schema.Int64Attribute{
							Computed:    true,
							Description: "1 enabled, 0 disabled (login is not allowed).",
						},
						"email": schema.StringAttribute{
							Computed: true,
						},
						"expiration_date": schema.Int64Attribute{
							Computed:    true,
							Description: "Account expiration date as unix timestamp in milliseconds. An expired account cannot login.",
						},
						"password": schema.StringAttribute{
							Computed: true,
						},
						"public_keys": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "List of public keys.",
						},
						"home_dir": schema.StringAttribute{
							Computed:    true,
							Description: "The user cannot upload or download files outside this directory. Must be an absolute path.",
						},
						"uid": schema.Int64Attribute{
							Computed:    true,
							Description: "If SFTPGo runs as root system user then the created files and directories will be assigned to this system UID.",
						},
						"gid": schema.Int64Attribute{
							Computed:    true,
							Description: "If SFTPGo runs as root system user then the created files and directories will be assigned to this system GID.",
						},
						"max_sessions": schema.Int64Attribute{
							Computed:    true,
							Description: "Maximum concurrent sessions. Not set means no limit.",
						},
						"quota_size": schema.Int64Attribute{
							Computed:    true,
							Description: "Maximum size allowed as bytes. Not set means no limit.",
						},
						"quota_files": schema.Int64Attribute{
							Computed:    true,
							Description: "Maximum number of files allowed. Not set means no limit.",
						},
						"permissions": schema.MapAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "Comma separated, per-directory, permissions.",
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
							Description: "Last quota update as unix timestamp in milliseconds.",
						},
						"upload_bandwidth": schema.Int64Attribute{
							Computed:    true,
							Description: "Maximum upload bandwidth as KB/s. Not set means unlimited. This is the default if no per-source limit match.",
						},
						"download_bandwidth": schema.Int64Attribute{
							Computed:    true,
							Description: "Maximum download bandwidth as KB/s. Not set means unlimited. This is the default if no per-source limit match.",
						},
						"upload_data_transfer": schema.Int64Attribute{
							Computed:    true,
							Description: "Maximum data transfer allowed for uploads as MB. Not set means no limit.",
						},
						"download_data_transfer": schema.Int64Attribute{
							Computed:    true,
							Description: "Maximum data transfer allowed for downloads as MB. Not set means no limit.",
						},
						"total_data_transfer": schema.Int64Attribute{
							Computed:    true,
							Description: "Maximum total data transfer as MB. Not set means unlimited. You can set a total data transfer instead of the individual values for uploads and downloads.",
						},
						"used_upload_data_transfer": schema.Int64Attribute{
							Computed:    true,
							Description: "Uploaded size, as bytes, since the last reset.",
						},
						"used_download_data_transfer": schema.Int64Attribute{
							Computed:    true,
							Description: "Downloaded size, as bytes, since the last reset.",
						},
						"last_login": schema.Int64Attribute{
							Computed:    true,
							Description: "Last login as unix timestamp in milliseconds.",
						},
						"created_at": schema.Int64Attribute{
							Computed:    true,
							Description: "Creation time as unix timestamp in milliseconds.",
						},
						"updated_at": schema.Int64Attribute{
							Computed:    true,
							Description: "Last update time as unix timestamp in milliseconds.",
						},
						"first_download": schema.Int64Attribute{
							Computed:    true,
							Description: "First download time as unix timestamp in milliseconds.",
						},
						"first_upload": schema.Int64Attribute{
							Computed:    true,
							Description: "First upload time as unix timestamp in milliseconds.",
						},
						"last_password_change": schema.Int64Attribute{
							Computed:    true,
							Description: "Last password change as unix timestamp in milliseconds.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Optional description.",
						},
						"additional_info": schema.StringAttribute{
							Computed:    true,
							Description: "Free form text field.",
						},
						"role": schema.StringAttribute{
							Computed:    true,
							Description: "Role name.",
						},
						"groups": schema.ListNestedAttribute{
							Computed:    true,
							Description: "Groups.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Computed:    true,
										Description: "Group name.",
									},
									"type": schema.Int64Attribute{
										Computed:    true,
										Description: "Group type. 1 = Primary, 2 = Secondary, 3 = Membership only.",
									},
								},
							},
						},
						"filters":         getComputedSchemaForUserFilters(false),
						"virtual_folders": getComputedSchemaForVirtualFolders(),
						"filesystem":      getComputedSchemaForFilesystem(),
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *usersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*client.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *usersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state usersDataSourceModel

	users, err := d.client.GetUsers()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read SFTPGo Users",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, user := range users {
		var userState userResourceModel
		diags := userState.fromSFTPGo(ctx, &user)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		state.Users = append(state.Users, userState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// usersDataSourceModel maps the data source schema data.
type usersDataSourceModel struct {
	Users []userResourceModel `tfsdk:"users"`
}
