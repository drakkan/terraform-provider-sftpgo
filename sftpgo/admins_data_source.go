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
	_ datasource.DataSource              = &adminsDataSource{}
	_ datasource.DataSourceWithConfigure = &adminsDataSource{}
)

// NewAdminsDataSource is a helper function to simplify the provider implementation.
func NewAdminsDataSource() datasource.DataSource {
	return &adminsDataSource{}
}

// adminsDataSource is the data source implementation.
type adminsDataSource struct {
	client *client.Client
}

// Metadata returns the data source type name.
func (d *adminsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admins"
}

// Schema defines the schema for the data source.
func (d *adminsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of admins.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Required to use the test framework. Just a placeholder.",
			},
			"admins": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of admins.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"username": schema.StringAttribute{
							Computed:    true,
							Description: "Unique username.",
						},
						"status": schema.Int64Attribute{
							Computed:    true,
							Description: "1 enabled, 0 disabled (login is not allowed).",
						},
						"password": schema.StringAttribute{
							Computed:    true,
							Description: "Password hash saved in the SFTPGo data provider.",
						},
						"email": schema.StringAttribute{
							Computed: true,
						},
						"permissions": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "Granted permissions.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Optional description.",
						},
						"additional_info": schema.StringAttribute{
							Computed:    true,
							Description: "Free form text field.",
						},
						"created_at": schema.Int64Attribute{
							Computed:    true,
							Description: "Creation time as unix timestamp in milliseconds.",
						},
						"updated_at": schema.Int64Attribute{
							Computed:    true,
							Description: "Last update time as unix timestamp in milliseconds.",
						},
						"last_login": schema.Int64Attribute{
							Computed:    true,
							Description: "Last login as unix timestamp in milliseconds.",
						},
						"role": schema.StringAttribute{
							Computed:    true,
							Description: "Role name. If set the admin can only administer users with the same role.",
						},
						"filters": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Additional restrictions.",
							Attributes: map[string]schema.Attribute{
								"allow_list": schema.ListAttribute{
									ElementType: types.StringType,
									Computed:    true,
									Description: `Only connections from these IP/Mask are allowed. IP/Mask must be in CIDR notation as defined in RFC 4632 and RFC 4291 for example "192.0.2.0/24" or "2001:db8::/32"`,
								},
								"allow_api_key_auth": schema.BoolAttribute{
									Computed:    true,
									Description: "If set, API Key authentication is allowed.",
								},
								"require_password_change": schema.BoolAttribute{
									Computed:    true,
									Description: "If set, two factor authentication is required.",
								},
								"require_two_factor": schema.BoolAttribute{
									Computed:    true,
									Description: "If set, API Key authentication is allowed.",
								},
								"disable_password_auth": schema.BoolAttribute{
									Computed:    true,
									Description: "If set, password authentication is disabled. The administrator can authenticate using an API key or OpenID Connect, if either is enabled.",
								},
							},
						},
						"preferences": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Admin preferences.",
							Attributes: map[string]schema.Attribute{
								"hide_user_page_sections": schema.Int64Attribute{
									Computed:    true,
									Description: "If set allow to hide some sections from the user page in the WebAdmin. 1 = groups, 2 = filesystem, 4 = virtual folders, 8 = profile, 16 = ACL, 32 = Disk and bandwidth quota limits, 64 = Advanced. Settings can be combined.",
								},
								"default_users_expiration": schema.Int64Attribute{
									Computed:    true,
									Description: "If set defines the default expiration for newly created users as number of days.",
								},
							},
						},
						"groups": schema.ListNestedAttribute{
							Computed:    true,
							Description: "Groups automatically selected for new users created by this admin.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Computed:    true,
										Description: "Group name.",
									},
									"options": schema.SingleNestedAttribute{
										Computed:    true,
										Description: "Options for admin/group mapping",
										Attributes: map[string]schema.Attribute{
											"add_to_users_as": schema.Int64Attribute{
												Computed:    true,
												Description: "Add to users as the specified group type. 1 = Primary, 2 = Secondary, 3 = Membership only.",
											},
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
func (d *adminsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*client.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *adminsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state adminsDataSourceModel

	admins, err := d.client.GetAdmins()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read SFTPGo Admins",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, admin := range admins {
		var adminState adminResourceModel
		diags := adminState.fromSFTPGo(ctx, &admin)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		state.Admins = append(state.Admins, adminState)
	}

	state.ID = types.StringValue(placeholderID)

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// adminsDataSourceModel maps the data source schema data.
type adminsDataSourceModel struct {
	ID     types.String         `tfsdk:"id"`
	Admins []adminResourceModel `tfsdk:"admins"`
}
