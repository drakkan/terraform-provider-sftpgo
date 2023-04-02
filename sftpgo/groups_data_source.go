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
	_ datasource.DataSource              = &groupsDataSource{}
	_ datasource.DataSourceWithConfigure = &groupsDataSource{}
)

// NewGroupsDataSource is a helper function to simplify the provider implementation.
func NewGroupsDataSource() datasource.DataSource {
	return &groupsDataSource{}
}

// groupsDataSource is the data source implementation.
type groupsDataSource struct {
	client *client.Client
}

// Metadata returns the data source type name.
func (d *groupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

// Schema defines the schema for the data source.
func (d *groupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of groups.",
		Attributes: map[string]schema.Attribute{
			"groups": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of groups.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Unique name",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Optional description.",
						},
						"created_at": schema.Int64Attribute{
							Computed:    true,
							Description: "Creation time as unix timestamp in milliseconds.",
						},
						"updated_at": schema.Int64Attribute{
							Computed:    true,
							Description: "Last update time as unix timestamp in milliseconds.",
						},
						"user_settings": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"home_dir": schema.StringAttribute{
									Computed:    true,
									Description: "If not set and the filesystem provider is local (0), the root filesystem will not be overridden.",
								},
								"max_sessions": schema.Int64Attribute{
									Computed:    true,
									Description: "Maximum concurrent sessions.",
								},
								"quota_size": schema.Int64Attribute{
									Computed:    true,
									Description: "Maximum size allowed as bytes.",
								},
								"quota_files": schema.Int64Attribute{
									Computed:    true,
									Description: "Maximum number of files allowed",
								},
								"permissions": schema.MapAttribute{
									Computed:    true,
									ElementType: types.StringType,
									Description: "Comma separated, per-directory, permissions.",
								},
								"upload_bandwidth": schema.Int64Attribute{
									Computed:    true,
									Description: "Maximum upload bandwidth as KB/s. This is the default if no per-source limit match.",
								},
								"download_bandwidth": schema.Int64Attribute{
									Computed:    true,
									Description: "Maximum download bandwidth as KB/s. This is the default if no per-source limit match.",
								},
								"upload_data_transfer": schema.Int64Attribute{
									Computed:    true,
									Description: "Maximum data transfer allowed for uploads as MB.",
								},
								"download_data_transfer": schema.Int64Attribute{
									Computed:    true,
									Description: "Maximum data transfer allowed for downloads as MB",
								},
								"total_data_transfer": schema.Int64Attribute{
									Computed:    true,
									Description: "Maximum total data transfer as MB. You can set a total data transfer instead of the individual values for uploads and downloads.",
								},
								"expires_in": schema.Int64Attribute{
									Computed:    true,
									Description: "Defines account expiration in number of days from creation. Not set means no expiration.",
								},
								"filters":    getComputedSchemaForUserFilters(true),
								"filesystem": getComputedSchemaForFilesystem(),
							},
						},
						"virtual_folders": getComputedSchemaForVirtualFolders(),
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *groupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*client.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *groupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state groupsDataSourceModel

	groups, err := d.client.GetGroups()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read SFTPGo Groups",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, group := range groups {
		var groupState groupResourceModel
		diags := groupState.fromSFTPGo(ctx, &group)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		state.Groups = append(state.Groups, groupState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// groupsDataSourceModel maps the data source schema data.
type groupsDataSourceModel struct {
	Groups []groupResourceModel `tfsdk:"groups"`
}
