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
	_ datasource.DataSource              = &rulesDataSource{}
	_ datasource.DataSourceWithConfigure = &rulesDataSource{}
)

// NewRulesDataSource is a helper function to simplify the provider implementation.
func NewRulesDataSource() datasource.DataSource {
	return &rulesDataSource{}
}

// rulesDataSource is the data source implementation.
type rulesDataSource struct {
	client *client.Client
}

// Metadata returns the data source type name.
func (d *rulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rules"
}

// Schema defines the schema for the data source.
func (d *rulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of event rules.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Required to use the test framework. Just a placeholder.",
			},
			"rules": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of event rules.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Unique name.",
						},
						"status": schema.Int64Attribute{
							Computed:    true,
							Description: "1 enabled, 0 disabled.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Optional description.",
						},
						"trigger": schema.Int64Attribute{
							Computed:    true,
							Description: "Event trigger. 1 = Filesystem event, 2 = Provider event, 3 = Schedule, 4 = IP Blocked, 5 = Certificate renewal, 6 = On demand, 7 = Identity Provider login.",
						},
						"created_at": schema.Int64Attribute{
							Computed:    true,
							Description: "Creation time as unix timestamp in milliseconds.",
						},
						"updated_at": schema.Int64Attribute{
							Computed:    true,
							Description: "Last update time as unix timestamp in milliseconds.",
						},
						"conditions": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Defines the conditions that trigger the rule.",
							Attributes: map[string]schema.Attribute{
								"fs_events": schema.ListAttribute{
									ElementType: types.StringType,
									Computed:    true,
									Description: `Filesystem events that trigger the rule. Supported values: "upload", "pre-upload", "first-upload", "download", "pre-download", "first-download", "delete", "pre-delete", "rename", "mkdir", "rmdir", "copy", "ssh_cmd"`,
								},
								"provider_events": schema.ListAttribute{
									ElementType: types.StringType,
									Computed:    true,
									Description: `Provider events that trigger the rule. Supported values: "add", "update", "delete".`,
								},
								"schedules": schema.ListNestedAttribute{
									Computed:    true,
									Description: "List of schedules that trigger the rule. Hours: 0-23. Day of week: 0-6 (Sun-Sat). Day of month: 1-31. Month: 1-12. Asterisk (*) indicates a match for all the values of the field. e.g. every day of week, every day of month and so on.",
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"minute": schema.StringAttribute{
												Computed: true,
											},
											"hour": schema.StringAttribute{
												Computed: true,
											},
											"day_of_week": schema.StringAttribute{
												Computed: true,
											},
											"day_of_month": schema.StringAttribute{
												Computed: true,
											},
											"month": schema.StringAttribute{
												Computed: true,
											},
										},
									},
								},
								"idp_login_event": schema.Int64Attribute{
									Computed:    true,
									Description: `Identity Provider login event that trigger the rule. 0 any, 1 user, 2 admin.`,
								},
								"options": schema.SingleNestedAttribute{
									Computed:    true,
									Description: "Options for event conditions.",
									Attributes: map[string]schema.Attribute{
										"names": schema.ListNestedAttribute{
											Computed:    true,
											Description: `Shell-like pattern filters for usernames, folder names. For example "user*"" will match names starting with "user". For provider events, this filter is applied to the username of the admin executing the event.`,
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"pattern": schema.StringAttribute{
														Computed: true,
													},
													"inverse_match": schema.BoolAttribute{
														Computed: true,
													},
												},
											},
										},
										"group_names": schema.ListNestedAttribute{
											Computed:    true,
											Description: `Shell-like pattern filters for group names. For example "group*"" will match group names starting with "group".`,
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"pattern": schema.StringAttribute{
														Computed: true,
													},
													"inverse_match": schema.BoolAttribute{
														Computed: true,
													},
												},
											},
										},
										"role_names": schema.ListNestedAttribute{
											Computed:    true,
											Description: `Shell-like pattern filters for role names. For example "role*"" will match role names starting with "role".`,
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"pattern": schema.StringAttribute{
														Computed: true,
													},
													"inverse_match": schema.BoolAttribute{
														Computed: true,
													},
												},
											},
										},
										"fs_paths": schema.ListNestedAttribute{
											Computed:    true,
											Description: `Shell-like pattern filters for filesystem events. For example "/adir/*.txt"" will match paths in the "/adir" directory ending with ".txt". Double asterisk is supported, for example "/**/*.txt" will match any file ending with ".txt". "/mydir/**" will match any entry in "/mydir".`,
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"pattern": schema.StringAttribute{
														Computed: true,
													},
													"inverse_match": schema.BoolAttribute{
														Computed: true,
													},
												},
											},
										},
										"protocols": schema.ListAttribute{
											ElementType: types.StringType,
											Computed:    true,
											Description: `The filesystem event rule will be triggered only for the specified protocols. Empty means any protocol. Supported values: "SFTP", "SCP", "SSH", "FTP", "DAV", "HTTP", "HTTPShare","OIDC"`,
										},
										"provider_objects": schema.ListAttribute{
											ElementType: types.StringType,
											Computed:    true,
											Description: `The provider event rule will be triggered only for the specified provider objects. Empty means any provider object. Supported values: "user", "folder", "group", "admin", "api_key", "share", "event_rule", "event_action".`,
										},
										"min_size": schema.Int64Attribute{
											Computed:    true,
											Description: `Minimum file size as bytes.`,
										},
										"max_size": schema.Int64Attribute{
											Computed:    true,
											Description: `Maximum file size as bytes.`,
										},
										"event_statuses": schema.ListAttribute{
											ElementType: types.Int32Type,
											Computed:    true,
											Description: `The filesystem event rules will be triggered only for actions with the specified status. Empty means any status. Suported values: 1 (OK), 2 (Failed), 3 (Failed for a quota exceeded error).`,
										},
										"concurrent_execution": schema.BoolAttribute{
											Computed:    true,
											Description: `If enabled, allow to execute scheduled tasks concurrently from multiple SFTPGo instances.`,
										},
									},
								},
							},
						},
						"actions": schema.ListNestedAttribute{
							Computed:    true,
							Description: `List of actions to execute.`,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Computed: true,
									},
									"is_failure_action": schema.BoolAttribute{
										Computed: true,
									},
									"stop_on_failure": schema.BoolAttribute{
										Computed: true,
									},
									"execute_sync": schema.BoolAttribute{
										Computed:    true,
										Description: `Supported for upload events and required for pre-* events and Identity provider login events if the action checks the account.`,
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
func (d *rulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*client.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *rulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state rulesDataSourceModel

	rules, err := d.client.GetRules()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read SFTPGo Rules",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, rule := range rules {
		var ruleState eventRuleResourceModel
		diags := ruleState.fromSFTPGo(ctx, &rule)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		state.Rules = append(state.Rules, ruleState)
	}

	state.ID = types.StringValue(placeholderID)

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// rulesDataSourceModel maps the data source schema data.
type rulesDataSourceModel struct {
	ID    types.String             `tfsdk:"id"`
	Rules []eventRuleResourceModel `tfsdk:"rules"`
}
