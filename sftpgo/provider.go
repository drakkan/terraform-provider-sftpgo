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
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo/client"
)

const (
	placeholderID = "placeholder"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &sftpgoProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New() provider.Provider {
	return &sftpgoProvider{}
}

// sftpgoProviderModel maps provider schema data to a Go type.
type sftpgoProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	APIKey   types.String `tfsdk:"api_key"`
	Headers  []keyValue   `tfsdk:"headers"`
	Edition  types.Int64  `tfsdk:"edition"`
}

// sftpgoProvider is the provider implementation.
type sftpgoProvider struct{}

// Metadata returns the provider type name.
func (p *sftpgoProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sftpgo"
}

// Schema defines the provider-level schema for configuration data.
func (p *sftpgoProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with SFTPGo.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:    true,
				Description: "URI for SFTPGo API. May also be provided via SFTPGO_HOST environment variable.",
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "Username for SFTPGo API. May also be provided via SFTPGO_USERNAME environment variable.",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Password for SFTPGo API. May also be provided via SFTPGO_PASSWORD environment variable.",
			},
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "SFTPGo API key. May also be provided via SFTPGO_API_KEY environment variable. You must provide an API key or username and password. If both an API key and username and password are provided, the API key will be used.",
			},
			"headers": schema.ListNestedAttribute{
				Optional:    true,
				Description: `Headers to add to the HTTP request.`,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required:    true,
							Description: "The header name. May also be provided via SFTPGO_HEADERS__0__KEY, SFTPGO_HEADERS__1__KEY, ... SFTPGO_HEADERS__9__KEY environment variables.",
						},
						"value": schema.StringAttribute{
							Required:    true,
							Description: "The header value. May also be provided via SFTPGO_HEADERS__0__VALUE, SFTPGO_HEADERS__1__VALUE, ... SFTPGO_HEADERS__9__VALUE environment variables.",
						},
					},
				},
			},
			"edition": schema.Int64Attribute{
				Optional:    true,
				Description: "SFTPGo edition. 0 = Open Source, 1 = Enterprise",
				Validators: []validator.Int64{
					int64validator.Between(0, 1),
				},
			},
		},
	}
}

// Configure prepares a SFTPGo API client for data sources and resources.
func (p *sftpgoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	ctx = tflog.SetField(ctx, "Version", getVersion())
	tflog.Info(ctx, "Configuring SFTPGo client")

	// Retrieve provider data from configuration
	var config sftpgoProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown SFTPGo API Host",
			"The provider cannot create the SFTPGo API client as there is an unknown configuration value for the SFTPGo API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the SFTPGO_HOST environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown SFTPGo API Username",
			"The provider cannot create the SFTPGo API client as there is an unknown configuration value for the SFTPGo API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the SFTPGO_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown SFTPGo API Password",
			"The provider cannot create the SFTPGo API client as there is an unknown configuration value for the SFTPGo API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the SFTPGO_PASSWORD environment variable.",
		)
	}

	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown SFTPGo API Key",
			"The provider cannot create the SFTPGo API client as there is an unknown configuration value for the SFTPGo API key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the SFTPGO_API_KEY environment variable.",
		)
	}

	if config.Edition.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("edition"),
			"Unknown SFTPGo edition",
			"The provider cannot create the SFTPGo API client as there is an unknown configuration value for the SFTPGo edition. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the SFTPGO_EDITION environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("SFTPGO_HOST")
	username := os.Getenv("SFTPGO_USERNAME")
	password := os.Getenv("SFTPGO_PASSWORD")
	apiKey := os.Getenv("SFTPGO_API_KEY")
	headers := getHeadersFromEnv()
	edition := getIntFromEnv("SFTPGO_EDITION", 0)

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	if !config.Edition.IsNull() {
		edition = config.Edition.ValueInt64()
	}

	if len(config.Headers) > 0 {
		headers = nil
		for _, h := range config.Headers {
			headers = append(headers, client.KeyValue{
				Key:   h.Key.ValueString(),
				Value: h.Value.ValueString(),
			})
		}
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing SFTPGo API Host",
			"The provider cannot create the SFTPGo API client as there is a missing or empty value for the SFTPGo API host. "+
				"Set the host value in the configuration or use the SFTPGO_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if apiKey == "" {
		if username == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("username"),
				"Missing SFTPGo API Username",
				"The provider cannot create the SFTPGo API client as there is a missing or empty value for the SFTPGo API username. "+
					"Set the username value in the configuration or use the SFTPGO_USERNAME environment variable. "+
					"If either is already set, ensure the value is not empty.",
			)
		}

		if password == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("password"),
				"Missing SFTPGo API Password",
				"The provider cannot create the SFTPGo API client as there is a missing or empty value for the SFTPGo API password. "+
					"Set the password value in the configuration or use the SFTPGO_PASSWORD environment variable. "+
					"If either is already set, ensure the value is not empty.",
			)
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "SFTPGo_host", host)
	ctx = tflog.SetField(ctx, "SFTPGo_username", username)
	ctx = tflog.SetField(ctx, "SFTPGo_password", password)
	ctx = tflog.SetField(ctx, "SFTPGo_api_key", apiKey)
	ctx = tflog.SetField(ctx, "SFTPGo_headers", headers)
	ctx = tflog.SetField(ctx, "SFTPGo edition", edition)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "SFTPGo_password")
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "SFTPGo_api_key")
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "SFTPGo_headers")

	tflog.Debug(ctx, "Creating SFTPGo client")

	// Create a new SFTPGo client using the configuration values
	client, err := client.NewClient(host, username, password, apiKey, headers, edition)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create SFTPGo API Client",
			"An unexpected error occurred when creating the SFTPGo API client. "+
				"If the error is not clear, please check the SFTPGo logs.\n\n"+
				"SFTPGo Client Error: "+err.Error(),
		)
		return
	}

	// Make the SFTPGo client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured SFTPGo client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *sftpgoProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewUsersDataSource,
		NewRolesDataSource,
		NewFoldersDataSource,
		NewGroupsDataSource,
		NewAdminsDataSource,
		NewDefenderEntriesDataSource,
		NewAllowListEntriesDataSource,
		NewRlSafeListEntriesDataSource,
		NewActionsDataSource,
		NewRulesDataSource,
		NewLicenseDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *sftpgoProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewUserResource,
		NewRoleResource,
		NewFolderResource,
		NewGroupResource,
		NewAdminResource,
		NewDefenderEntryResource,
		NewAllowListEntryResource,
		NewRlSafeListEntryResource,
		NewActionResource,
		NewRuleResource,
		NewLicenseResource,
	}
}

func getIntFromEnv(name string, defaultValue int64) int64 {
	val, err := strconv.ParseInt(os.Getenv(name), 10, 64)
	if err != nil {
		return defaultValue
	}
	return val
}

func getHeadersFromEnv() []client.KeyValue {
	var headers []client.KeyValue

	for idx := 0; idx < 10; idx++ {
		key := strings.TrimSpace(os.Getenv(fmt.Sprintf("SFTPGO_HEADERS__%d__KEY", idx)))
		value := strings.TrimSpace(os.Getenv(fmt.Sprintf("SFTPGO_HEADERS__%d__VALUE", idx)))
		if key != "" && value != "" {
			headers = append(headers, client.KeyValue{
				Key:   key,
				Value: value,
			})
		}
	}
	return headers
}
