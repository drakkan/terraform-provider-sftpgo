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
	_ datasource.DataSource              = &allowListEntriesDataSource{}
	_ datasource.DataSourceWithConfigure = &allowListEntriesDataSource{}
)

// NewAllowListEntriesDataSource is a helper function to simplify the provider implementation.
func NewAllowListEntriesDataSource() datasource.DataSource {
	return &allowListEntriesDataSource{}
}

// allowListEntriesDataSource is the data source implementation.
type allowListEntriesDataSource struct {
	client *client.Client
}

// Metadata returns the data source type name.
func (d *allowListEntriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_allowlist_entries"
}

// Schema defines the schema for the data source.
func (d *allowListEntriesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the allow list entries.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Required to use the test framework. Just a placeholder.",
			},
			"entries": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of entries.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"ipornet": schema.StringAttribute{
							Computed:    true,
							Description: `IP address or network in CIDR format, for example "192.168.1.2/32", "192.168.0.0/24", "2001:db8::/32"`,
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Optional description.",
						},
						"protocols": schema.Int64Attribute{
							Computed:    true,
							Description: "Defines the protocol the entry applies to. 0 means all the supported protocols, 1 SSH, 2 FTP, 4 WebDAV, 8 HTTP. Protocols can be combined, for example 3 means SSH and FTP.",
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
func (d *allowListEntriesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*client.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *allowListEntriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state allowListEntriesDataSourceModel

	entries, err := d.client.GetIPListEntries(1)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read SFTPGo allow list entries",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, entry := range entries {
		var entryState allowListEntryResourceModel
		diags := entryState.fromSFTPGo(ctx, &entry)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		state.Entries = append(state.Entries, entryState)
	}

	state.ID = types.StringValue(placeholderID)

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// allowListEntriesDataSourceModel maps the data source schema data.
type allowListEntriesDataSourceModel struct {
	ID      types.String                  `tfsdk:"id"`
	Entries []allowListEntryResourceModel `tfsdk:"entries"`
}
