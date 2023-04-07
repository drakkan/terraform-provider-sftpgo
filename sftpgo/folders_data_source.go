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
	_ datasource.DataSource              = &foldersDataSource{}
	_ datasource.DataSourceWithConfigure = &foldersDataSource{}
)

// NewFoldersDataSource is a helper function to simplify the provider implementation.
func NewFoldersDataSource() datasource.DataSource {
	return &foldersDataSource{}
}

// foldersDataSource is the data source implementation.
type foldersDataSource struct {
	client *client.Client
}

// Metadata returns the data source type name.
func (d *foldersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_folders"
}

// Schema defines the schema for the data source.
func (d *foldersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of virtual folders.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Required to use the test framework. Just a placeholder.",
			},
			"folders": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of virtual folders.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Unique name",
						},
						"mapped_path": schema.StringAttribute{
							Computed:    true,
							Description: "Absolute path to a local directory. This is the folder root path for local storage provider. For non-local filesystems it will store temporary files.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Optional description.",
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
							Description: "Last quota update as unix timestamp in milliseconds",
						},
						"filesystem": getComputedSchemaForFilesystem(),
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *foldersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*client.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *foldersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state foldersDataSourceModel

	folders, err := d.client.GetFolders()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read SFTPGo Virtual Folders",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, folder := range folders {
		var folderState virtualFolderResourceModel
		diags := folderState.fromSFTPGo(ctx, &folder)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		state.Folders = append(state.Folders, folderState)
	}

	state.ID = types.StringValue(placeholderID)

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// foldersDataSourceModel maps the data source schema data.
type foldersDataSourceModel struct {
	ID      types.String                 `tfsdk:"id"`
	Folders []virtualFolderResourceModel `tfsdk:"folders"`
}
