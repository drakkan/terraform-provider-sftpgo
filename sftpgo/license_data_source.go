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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &licenseDataSource{}
	_ datasource.DataSourceWithConfigure = &licenseDataSource{}
)

// NewLicenseDataSource is a helper function to simplify the provider implementation.
func NewLicenseDataSource() datasource.DataSource {
	return &licenseDataSource{}
}

// licenseDataSource is the data source implementation.
type licenseDataSource struct {
	client *client.Client
}

// Metadata returns the data source type name.
func (d *licenseDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_license_info"
}

// Schema defines the schema for the data source.
func (d *licenseDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches License Information. " + enterpriseFeatureNote + ".",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Required to use the test framework. Just a placeholder.",
			},
			"license": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "License details.",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed: true,
					},
					"key": schema.StringAttribute{
						Computed:    true,
						Description: "License key",
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
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *licenseDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*client.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *licenseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state licenseDataSourceModel

	license, err := d.client.GetLicense()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read SFTPGo License",
			err.Error(),
		)
		return
	}

	var licenseResource licenseResourceModel
	diags := licenseResource.fromSFTPGo(ctx, license)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.License = licenseResource
	state.ID = types.StringValue(placeholderID)

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// licenseDataSourceModel maps the data source schema data.
type licenseDataSourceModel struct {
	ID      types.String         `tfsdk:"id"`
	License licenseResourceModel `tfsdk:"license"`
}
