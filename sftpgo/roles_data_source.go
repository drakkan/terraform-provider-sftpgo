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

	"github.com/sftpgo/terraform-provider-sftpgo/sftpgo/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &rolesDataSource{}
	_ datasource.DataSourceWithConfigure = &rolesDataSource{}
)

// NewRolesDataSource is a helper function to simplify the provider implementation.
func NewRolesDataSource() datasource.DataSource {
	return &rolesDataSource{}
}

// rolesDataSource is the data source implementation.
type rolesDataSource struct {
	client *client.Client
}

// Metadata returns the data source type name.
func (d *rolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_roles"
}

// Schema defines the schema for the data source.
func (d *rolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of roles.",
		Attributes: map[string]schema.Attribute{
			"roles": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of roles.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Unique name.",
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
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *rolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*client.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *rolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state rolesDataSourceModel

	roles, err := d.client.GetRoles()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read SFTPGo Roles",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, role := range roles {
		var roleState roleResourceModel
		diags := roleState.fromSFTPGo(ctx, &role)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		state.Roles = append(state.Roles, roleState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// rolesDataSourceModel maps the data source schema data.
type rolesDataSourceModel struct {
	Roles []roleResourceModel `tfsdk:"roles"`
}
