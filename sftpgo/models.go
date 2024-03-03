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
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/sftpgo/sdk"
	"github.com/sftpgo/sdk/kms"

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo/client"
)

// userResourceModel maps users schema data.
type userResourceModel struct {
	ID                       types.String       `tfsdk:"id"`
	Username                 types.String       `tfsdk:"username"`
	Email                    types.String       `tfsdk:"email"`
	Status                   types.Int64        `tfsdk:"status"`
	ExpirationDate           types.Int64        `tfsdk:"expiration_date"`
	Password                 types.String       `tfsdk:"password"`
	PublicKeys               types.List         `tfsdk:"public_keys"`
	HomeDir                  types.String       `tfsdk:"home_dir"`
	UID                      types.Int64        `tfsdk:"uid"`
	GID                      types.Int64        `tfsdk:"gid"`
	MaxSessions              types.Int64        `tfsdk:"max_sessions"`
	QuotaSize                types.Int64        `tfsdk:"quota_size"`
	QuotaFiles               types.Int64        `tfsdk:"quota_files"`
	Permissions              types.Map          `tfsdk:"permissions"`
	UsedQuotaSize            types.Int64        `tfsdk:"used_quota_size"`
	UsedQuotaFiles           types.Int64        `tfsdk:"used_quota_files"`
	LastQuotaUpdate          types.Int64        `tfsdk:"last_quota_update"`
	UploadBandwidth          types.Int64        `tfsdk:"upload_bandwidth"`
	DownloadBandwidth        types.Int64        `tfsdk:"download_bandwidth"`
	UploadDataTransfer       types.Int64        `tfsdk:"upload_data_transfer"`
	DownloadDataTransfer     types.Int64        `tfsdk:"download_data_transfer"`
	TotalDataTransfer        types.Int64        `tfsdk:"total_data_transfer"`
	UsedUploadDataTransfer   types.Int64        `tfsdk:"used_upload_data_transfer"`
	UsedDownloadDataTransfer types.Int64        `tfsdk:"used_download_data_transfer"`
	LastLogin                types.Int64        `tfsdk:"last_login"`
	CreatedAt                types.Int64        `tfsdk:"created_at"`
	UpdatedAt                types.Int64        `tfsdk:"updated_at"`
	FirstDownload            types.Int64        `tfsdk:"first_download"`
	FirstUpload              types.Int64        `tfsdk:"first_upload"`
	LastPasswordChange       types.Int64        `tfsdk:"last_password_change"`
	Description              types.String       `tfsdk:"description"`
	AdditionalInfo           types.String       `tfsdk:"additional_info"`
	Role                     types.String       `tfsdk:"role"`
	Groups                   []userGroupMapping `tfsdk:"groups"`
	Filters                  types.Object       `tfsdk:"filters"`
	VirtualFolders           []virtualFolder    `tfsdk:"virtual_folders"`
	FsConfig                 types.Object       `tfsdk:"filesystem"`
}

func (u *userResourceModel) toSFTPGo(ctx context.Context) (*client.User, diag.Diagnostics) {
	user := &client.User{
		User: sdk.User{
			BaseUser: sdk.BaseUser{
				Username:             u.Username.ValueString(),
				Status:               int(u.Status.ValueInt64()),
				Email:                u.Email.ValueString(),
				ExpirationDate:       u.ExpirationDate.ValueInt64(),
				HomeDir:              u.HomeDir.ValueString(),
				UID:                  int(u.UID.ValueInt64()),
				GID:                  int(u.GID.ValueInt64()),
				MaxSessions:          int(u.MaxSessions.ValueInt64()),
				QuotaSize:            u.QuotaSize.ValueInt64(),
				QuotaFiles:           int(u.QuotaFiles.ValueInt64()),
				UploadBandwidth:      u.UploadBandwidth.ValueInt64(),
				DownloadBandwidth:    u.DownloadBandwidth.ValueInt64(),
				UploadDataTransfer:   u.UploadDataTransfer.ValueInt64(),
				DownloadDataTransfer: u.DownloadDataTransfer.ValueInt64(),
				TotalDataTransfer:    u.TotalDataTransfer.ValueInt64(),
				Description:          u.Description.ValueString(),
				AdditionalInfo:       u.AdditionalInfo.ValueString(),
				Role:                 u.Role.ValueString(),
			},
		},
		Password: u.Password.ValueString(),
	}
	if !u.PublicKeys.IsNull() {
		diags := u.PublicKeys.ElementsAs(ctx, &user.PublicKeys, false)
		if diags.HasError() {
			return user, diags
		}
	}
	permissions := make(map[string]string)
	if !u.Permissions.IsNull() {
		diags := u.Permissions.ElementsAs(ctx, &permissions, false)
		if diags.HasError() {
			return user, diags
		}
	}
	user.Permissions = make(map[string][]string)
	for k, v := range permissions {
		user.Permissions[k] = strings.Split(v, ",")
	}
	for _, g := range u.Groups {
		user.Groups = append(user.Groups, sdk.GroupMapping{
			Name: g.Name.ValueString(),
			Type: int(g.Type.ValueInt64()),
		})
	}
	var filters userFilters
	diags := u.Filters.As(ctx, &filters, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return user, diags
	}
	sftpgoFilters, diags := filters.toSFTPGo(ctx)
	if diags.HasError() {
		return user, diags
	}
	user.Filters = sftpgoFilters
	for _, f := range u.VirtualFolders {
		folder, diags := f.toSFTPGo(ctx)
		if diags.HasError() {
			return user, diags
		}
		user.VirtualFolders = append(user.VirtualFolders, folder)
	}
	var fs filesystem
	diags = u.FsConfig.As(ctx, &fs, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return user, diags
	}
	sftpgoFs, diags := fs.toSFTPGo(ctx)
	if diags.HasError() {
		return user, diags
	}
	user.FsConfig = sftpgoFs
	return user, nil
}

func (u *userResourceModel) fromSFTPGo(ctx context.Context, user *client.User) diag.Diagnostics {
	u.Username = types.StringValue(user.Username)
	u.ID = u.Username
	u.Status = types.Int64Value(int64(user.Status))
	u.Email = getOptionalString(user.Email)
	u.ExpirationDate = getOptionalInt64(user.ExpirationDate)
	u.Password = getOptionalString(user.Password)
	u.HomeDir = types.StringValue(user.HomeDir)
	u.UID = getOptionalInt64(int64(user.UID))
	u.GID = getOptionalInt64(int64(user.GID))
	u.MaxSessions = getOptionalInt64(int64(user.MaxSessions))
	u.QuotaSize = getOptionalInt64(user.QuotaSize)
	u.QuotaFiles = getOptionalInt64(int64(user.QuotaFiles))
	u.UsedQuotaSize = getOptionalInt64(user.UsedQuotaSize)
	u.UsedQuotaFiles = getOptionalInt64(int64(user.UsedQuotaFiles))
	u.LastQuotaUpdate = getOptionalInt64(user.LastQuotaUpdate)
	u.UploadBandwidth = getOptionalInt64(user.UploadBandwidth)
	u.DownloadBandwidth = getOptionalInt64(user.DownloadBandwidth)
	u.UploadDataTransfer = getOptionalInt64(user.UploadDataTransfer)
	u.DownloadDataTransfer = getOptionalInt64(user.DownloadDataTransfer)
	u.TotalDataTransfer = getOptionalInt64(user.TotalDataTransfer)
	u.UsedUploadDataTransfer = getOptionalInt64(user.UsedUploadDataTransfer)
	u.UsedDownloadDataTransfer = getOptionalInt64(user.UsedDownloadDataTransfer)
	u.LastLogin = getOptionalInt64(user.LastLogin)
	u.CreatedAt = types.Int64Value(user.CreatedAt)
	u.UpdatedAt = types.Int64Value(user.UpdatedAt)
	u.FirstDownload = getOptionalInt64(user.FirstDownload)
	u.FirstUpload = getOptionalInt64(user.FirstUpload)
	u.LastPasswordChange = getOptionalInt64(user.LastPasswordChange)
	u.Description = getOptionalString(user.Description)
	u.AdditionalInfo = getOptionalString(user.AdditionalInfo)
	u.Role = getOptionalString(user.Role)
	pKeys, diags := types.ListValueFrom(ctx, types.StringType, user.PublicKeys)
	if diags.HasError() {
		return diags
	}
	u.PublicKeys = pKeys

	permissions := make(map[string]string)
	for k, v := range user.Permissions {
		permissions[k] = strings.Join(v, ",")
	}
	tfMap, diags := types.MapValueFrom(ctx, types.StringType, permissions)
	if diags.HasError() {
		return diags
	}
	u.Permissions = tfMap

	u.Groups = nil
	for _, g := range user.Groups {
		u.Groups = append(u.Groups, userGroupMapping{
			Name: types.StringValue(g.Name),
			Type: types.Int64Value(int64(g.Type)),
		})
	}

	var f userFilters
	diags = f.fromSFTPGo(ctx, &user.Filters)
	if diags.HasError() {
		return diags
	}
	filters, diags := types.ObjectValueFrom(ctx, f.getTFAttributes(), f)
	if diags.HasError() {
		return diags
	}
	u.Filters = filters

	u.VirtualFolders = nil
	for _, f := range user.VirtualFolders {
		var folder virtualFolder
		diags := folder.fromSFTPGo(ctx, &f)
		if diags.HasError() {
			return diags
		}
		u.VirtualFolders = append(u.VirtualFolders, folder)
	}

	var fsConfig filesystem
	diags = fsConfig.fromSFTPGo(ctx, &user.FsConfig)
	if diags.HasError() {
		return diags
	}
	fs, diags := types.ObjectValueFrom(ctx, fsConfig.getTFAttributes(), fsConfig)
	if diags.HasError() {
		return diags
	}
	u.FsConfig = fs

	return nil
}

type userGroupMapping struct {
	Name types.String `tfsdk:"name"`
	Type types.Int64  `tfsdk:"type"`
}

type patternsFilter struct {
	Path            types.String `tfsdk:"path"`
	AllowedPatterns types.List   `tfsdk:"allowed_patterns"`
	DeniedPatterns  types.List   `tfsdk:"denied_patterns"`
	DenyPolicy      types.Int64  `tfsdk:"deny_policy"`
}

type bandwidthLimit struct {
	Sources           types.List  `tfsdk:"sources"`
	UploadBandwidth   types.Int64 `tfsdk:"upload_bandwidth"`
	DownloadBandwidth types.Int64 `tfsdk:"download_bandwidth"`
}

type baseUserFilters struct {
	AllowedIP               types.List       `tfsdk:"allowed_ip"`
	DeniedIP                types.List       `tfsdk:"denied_ip"`
	DeniedLoginMethods      types.List       `tfsdk:"denied_login_methods"`
	DeniedProtocols         types.List       `tfsdk:"denied_protocols"`
	FilePatterns            []patternsFilter `tfsdk:"file_patterns"`
	MaxUploadFileSize       types.Int64      `tfsdk:"max_upload_file_size"`
	TLSUsername             types.String     `tfsdk:"tls_username"`
	ExternalAuthDisabled    types.Bool       `tfsdk:"external_auth_disabled"`
	PreLoginDisabled        types.Bool       `tfsdk:"pre_login_disabled"`
	CheckPasswordDisabled   types.Bool       `tfsdk:"check_password_disabled"`
	DisableFsChecks         types.Bool       `tfsdk:"disable_fs_checks"`
	WebClient               types.List       `tfsdk:"web_client"`
	AllowAPIKeyAuth         types.Bool       `tfsdk:"allow_api_key_auth"`
	UserType                types.String     `tfsdk:"user_type"`
	BandwidthLimits         []bandwidthLimit `tfsdk:"bandwidth_limits"`
	ExternalAuthCacheTime   types.Int64      `tfsdk:"external_auth_cache_time"`
	StartDirectory          types.String     `tfsdk:"start_directory"`
	TwoFactorAuthProtocols  types.List       `tfsdk:"two_factor_protocols"`
	FTPSecurity             types.Int64      `tfsdk:"ftp_security"`
	IsAnonymous             types.Bool       `tfsdk:"is_anonymous"`
	DefaultSharesExpiration types.Int64      `tfsdk:"default_shares_expiration"`
	MaxSharesExpiration     types.Int64      `tfsdk:"max_shares_expiration"`
	PasswordExpiration      types.Int64      `tfsdk:"password_expiration"`
	PasswordStrength        types.Int64      `tfsdk:"password_strength"`
}

func (f *baseUserFilters) getTFAttributes() map[string]attr.Type {
	return map[string]attr.Type{
		"allowed_ip": types.ListType{
			ElemType: types.StringType,
		},
		"denied_ip": types.ListType{
			ElemType: types.StringType,
		},
		"denied_login_methods": types.ListType{
			ElemType: types.StringType,
		},
		"denied_protocols": types.ListType{
			ElemType: types.StringType,
		},
		"file_patterns": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"path": types.StringType,
					"allowed_patterns": types.ListType{
						ElemType: types.StringType,
					},
					"denied_patterns": types.ListType{
						ElemType: types.StringType,
					},
					"deny_policy": types.Int64Type,
				},
			},
		},
		"max_upload_file_size":    types.Int64Type,
		"tls_username":            types.StringType,
		"external_auth_disabled":  types.BoolType,
		"pre_login_disabled":      types.BoolType,
		"check_password_disabled": types.BoolType,
		"disable_fs_checks":       types.BoolType,
		"web_client": types.ListType{
			ElemType: types.StringType,
		},
		"allow_api_key_auth": types.BoolType,
		"user_type":          types.StringType,
		"bandwidth_limits": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"sources": types.ListType{
						ElemType: types.StringType,
					},
					"upload_bandwidth":   types.Int64Type,
					"download_bandwidth": types.Int64Type,
				},
			},
		},
		"external_auth_cache_time": types.Int64Type,
		"start_directory":          types.StringType,
		"two_factor_protocols": types.ListType{
			ElemType: types.StringType,
		},
		"ftp_security":              types.Int64Type,
		"is_anonymous":              types.BoolType,
		"default_shares_expiration": types.Int64Type,
		"max_shares_expiration":     types.Int64Type,
		"password_expiration":       types.Int64Type,
		"password_strength":         types.Int64Type,
	}
}

func (f *baseUserFilters) toSFTPGo(ctx context.Context) (sdk.BaseUserFilters, diag.Diagnostics) {
	filters := sdk.BaseUserFilters{
		MaxUploadFileSize: f.MaxUploadFileSize.ValueInt64(),
		TLSUsername:       sdk.TLSUsername(f.TLSUsername.ValueString()),
		Hooks: sdk.HooksFilter{
			ExternalAuthDisabled:  f.ExternalAuthDisabled.ValueBool(),
			PreLoginDisabled:      f.PreLoginDisabled.ValueBool(),
			CheckPasswordDisabled: f.CheckPasswordDisabled.ValueBool(),
		},
		DisableFsChecks:         f.DisableFsChecks.ValueBool(),
		AllowAPIKeyAuth:         f.AllowAPIKeyAuth.ValueBool(),
		UserType:                f.UserType.ValueString(),
		ExternalAuthCacheTime:   f.ExternalAuthCacheTime.ValueInt64(),
		StartDirectory:          f.StartDirectory.ValueString(),
		FTPSecurity:             int(f.FTPSecurity.ValueInt64()),
		IsAnonymous:             f.IsAnonymous.ValueBool(),
		DefaultSharesExpiration: int(f.DefaultSharesExpiration.ValueInt64()),
		MaxSharesExpiration:     int(f.MaxSharesExpiration.ValueInt64()),
		PasswordExpiration:      int(f.PasswordExpiration.ValueInt64()),
		PasswordStrength:        int(f.PasswordStrength.ValueInt64()),
	}
	for _, p := range f.FilePatterns {
		patterns := sdk.PatternsFilter{
			Path:       p.Path.ValueString(),
			DenyPolicy: int(p.DenyPolicy.ValueInt64()),
		}
		if !p.AllowedPatterns.IsNull() {
			diags := p.AllowedPatterns.ElementsAs(ctx, &patterns.AllowedPatterns, false)
			if diags.HasError() {
				return filters, diags
			}
		}
		if !p.DeniedPatterns.IsNull() {
			diags := p.DeniedPatterns.ElementsAs(ctx, &patterns.DeniedPatterns, false)
			if diags.HasError() {
				return filters, diags
			}
		}
		filters.FilePatterns = append(filters.FilePatterns, patterns)
	}
	for _, l := range f.BandwidthLimits {
		limits := sdk.BandwidthLimit{
			UploadBandwidth:   l.UploadBandwidth.ValueInt64(),
			DownloadBandwidth: l.DownloadBandwidth.ValueInt64(),
		}
		if !l.Sources.IsNull() {
			diags := l.Sources.ElementsAs(ctx, &limits.Sources, false)
			if diags.HasError() {
				return filters, diags
			}
		}
		filters.BandwidthLimits = append(filters.BandwidthLimits, limits)
	}
	if !f.AllowedIP.IsNull() {
		diags := f.AllowedIP.ElementsAs(ctx, &filters.AllowedIP, false)
		if diags.HasError() {
			return filters, diags
		}
	}
	if !f.DeniedIP.IsNull() {
		diags := f.DeniedIP.ElementsAs(ctx, &filters.DeniedIP, false)
		if diags.HasError() {
			return filters, diags
		}
	}
	if !f.DeniedLoginMethods.IsNull() {
		diags := f.DeniedLoginMethods.ElementsAs(ctx, &filters.DeniedLoginMethods, false)
		if diags.HasError() {
			return filters, diags
		}
	}
	if !f.DeniedProtocols.IsNull() {
		diags := f.DeniedProtocols.ElementsAs(ctx, &filters.DeniedProtocols, false)
		if diags.HasError() {
			return filters, diags
		}
	}
	if !f.WebClient.IsNull() {
		diags := f.WebClient.ElementsAs(ctx, &filters.WebClient, false)
		if diags.HasError() {
			return filters, diags
		}
	}
	if !f.TwoFactorAuthProtocols.IsNull() {
		diags := f.TwoFactorAuthProtocols.ElementsAs(ctx, &filters.TwoFactorAuthProtocols, false)
		if diags.HasError() {
			return filters, diags
		}
	}

	return filters, nil
}

func (f *baseUserFilters) fromSFTPGo(ctx context.Context, filters *sdk.BaseUserFilters) diag.Diagnostics {
	allowedIP, diags := types.ListValueFrom(ctx, types.StringType, filters.AllowedIP)
	if diags.HasError() {
		return diags
	}
	f.AllowedIP = allowedIP
	deniedIP, diags := types.ListValueFrom(ctx, types.StringType, filters.DeniedIP)
	if diags.HasError() {
		return diags
	}
	f.DeniedIP = deniedIP
	deniedLoginMethods, diags := types.ListValueFrom(ctx, types.StringType, filters.DeniedLoginMethods)
	if diags.HasError() {
		return diags
	}
	f.DeniedLoginMethods = deniedLoginMethods
	deniedProtocols, diags := types.ListValueFrom(ctx, types.StringType, filters.DeniedProtocols)
	if diags.HasError() {
		return diags
	}
	f.DeniedProtocols = deniedProtocols

	f.FilePatterns = nil
	for _, patterns := range filters.FilePatterns {
		allowedPatterns, diags := types.ListValueFrom(ctx, types.StringType, patterns.AllowedPatterns)
		if diags.HasError() {
			return diags
		}
		deniedPatterns, diags := types.ListValueFrom(ctx, types.StringType, patterns.DeniedPatterns)
		if diags.HasError() {
			return diags
		}
		f.FilePatterns = append(f.FilePatterns, patternsFilter{
			Path:            types.StringValue(patterns.Path),
			AllowedPatterns: allowedPatterns,
			DeniedPatterns:  deniedPatterns,
			DenyPolicy:      getOptionalInt64(int64(patterns.DenyPolicy)),
		})
	}

	f.MaxUploadFileSize = getOptionalInt64(filters.MaxUploadFileSize)
	f.TLSUsername = getOptionalString(string(filters.TLSUsername))
	f.ExternalAuthDisabled = getOptionalBool(filters.Hooks.ExternalAuthDisabled)
	f.PreLoginDisabled = getOptionalBool(filters.Hooks.PreLoginDisabled)
	f.CheckPasswordDisabled = getOptionalBool(filters.Hooks.CheckPasswordDisabled)
	f.DisableFsChecks = getOptionalBool(filters.DisableFsChecks)
	webClient, diags := types.ListValueFrom(ctx, types.StringType, filters.WebClient)
	if diags.HasError() {
		return diags
	}
	f.WebClient = webClient
	f.AllowAPIKeyAuth = getOptionalBool(filters.AllowAPIKeyAuth)
	f.UserType = getOptionalString(filters.UserType)

	f.BandwidthLimits = nil
	for _, limit := range filters.BandwidthLimits {
		sources, diags := types.ListValueFrom(ctx, types.StringType, limit.Sources)
		if diags.HasError() {
			return diags
		}
		f.BandwidthLimits = append(f.BandwidthLimits, bandwidthLimit{
			Sources:           sources,
			UploadBandwidth:   getOptionalInt64(limit.UploadBandwidth),
			DownloadBandwidth: getOptionalInt64(limit.DownloadBandwidth),
		})
	}

	f.ExternalAuthCacheTime = getOptionalInt64(filters.ExternalAuthCacheTime)
	f.StartDirectory = getOptionalString(filters.StartDirectory)
	twoFactorProtos, diags := types.ListValueFrom(ctx, types.StringType, filters.TwoFactorAuthProtocols)
	if diags.HasError() {
		return diags
	}
	f.TwoFactorAuthProtocols = twoFactorProtos
	f.FTPSecurity = getOptionalInt64(int64(filters.FTPSecurity))
	f.IsAnonymous = getOptionalBool(filters.IsAnonymous)
	f.DefaultSharesExpiration = getOptionalInt64(int64(filters.DefaultSharesExpiration))
	f.MaxSharesExpiration = getOptionalInt64(int64(filters.MaxSharesExpiration))
	f.PasswordExpiration = getOptionalInt64(int64(filters.PasswordExpiration))
	f.PasswordStrength = getOptionalInt64(int64(filters.PasswordStrength))
	return nil
}

type userFilters struct {
	// embedded structs are not supported
	//baseUserFilters
	AllowedIP               types.List       `tfsdk:"allowed_ip"`
	DeniedIP                types.List       `tfsdk:"denied_ip"`
	DeniedLoginMethods      types.List       `tfsdk:"denied_login_methods"`
	DeniedProtocols         types.List       `tfsdk:"denied_protocols"`
	FilePatterns            []patternsFilter `tfsdk:"file_patterns"`
	MaxUploadFileSize       types.Int64      `tfsdk:"max_upload_file_size"`
	TLSUsername             types.String     `tfsdk:"tls_username"`
	TLSCerts                types.List       `tfsdk:"tls_certs"`
	ExternalAuthDisabled    types.Bool       `tfsdk:"external_auth_disabled"`
	PreLoginDisabled        types.Bool       `tfsdk:"pre_login_disabled"`
	CheckPasswordDisabled   types.Bool       `tfsdk:"check_password_disabled"`
	DisableFsChecks         types.Bool       `tfsdk:"disable_fs_checks"`
	WebClient               types.List       `tfsdk:"web_client"`
	AllowAPIKeyAuth         types.Bool       `tfsdk:"allow_api_key_auth"`
	UserType                types.String     `tfsdk:"user_type"`
	BandwidthLimits         []bandwidthLimit `tfsdk:"bandwidth_limits"`
	ExternalAuthCacheTime   types.Int64      `tfsdk:"external_auth_cache_time"`
	StartDirectory          types.String     `tfsdk:"start_directory"`
	TwoFactorAuthProtocols  types.List       `tfsdk:"two_factor_protocols"`
	FTPSecurity             types.Int64      `tfsdk:"ftp_security"`
	IsAnonymous             types.Bool       `tfsdk:"is_anonymous"`
	DefaultSharesExpiration types.Int64      `tfsdk:"default_shares_expiration"`
	MaxSharesExpiration     types.Int64      `tfsdk:"max_shares_expiration"`
	PasswordExpiration      types.Int64      `tfsdk:"password_expiration"`
	PasswordStrength        types.Int64      `tfsdk:"password_strength"`
	RequirePasswordChange   types.Bool       `tfsdk:"require_password_change"`
}

func (f *userFilters) getTFAttributes() map[string]attr.Type {
	baseFilters := baseUserFilters{}
	base := baseFilters.getTFAttributes()

	filters := map[string]attr.Type{
		"require_password_change": types.BoolType,
		"tls_certs": types.ListType{
			ElemType: types.StringType,
		},
	}

	for k, v := range filters {
		base[k] = v
	}
	return base
}

func (f *userFilters) getBaseFilters() baseUserFilters {
	return baseUserFilters{
		AllowedIP:               f.AllowedIP,
		DeniedIP:                f.DeniedIP,
		DeniedLoginMethods:      f.DeniedLoginMethods,
		DeniedProtocols:         f.DeniedProtocols,
		FilePatterns:            f.FilePatterns,
		MaxUploadFileSize:       f.MaxUploadFileSize,
		TLSUsername:             f.TLSUsername,
		ExternalAuthDisabled:    f.ExternalAuthDisabled,
		PreLoginDisabled:        f.PreLoginDisabled,
		CheckPasswordDisabled:   f.CheckPasswordDisabled,
		DisableFsChecks:         f.DisableFsChecks,
		WebClient:               f.WebClient,
		AllowAPIKeyAuth:         f.AllowAPIKeyAuth,
		UserType:                f.UserType,
		BandwidthLimits:         f.BandwidthLimits,
		ExternalAuthCacheTime:   f.ExternalAuthCacheTime,
		StartDirectory:          f.StartDirectory,
		TwoFactorAuthProtocols:  f.TwoFactorAuthProtocols,
		FTPSecurity:             f.FTPSecurity,
		IsAnonymous:             f.IsAnonymous,
		DefaultSharesExpiration: f.DefaultSharesExpiration,
		MaxSharesExpiration:     f.MaxSharesExpiration,
		PasswordExpiration:      f.PasswordExpiration,
		PasswordStrength:        f.PasswordStrength,
	}
}

func (f *userFilters) fromBaseFilters(filters *baseUserFilters) {
	f.AllowedIP = filters.AllowedIP
	f.DeniedIP = filters.DeniedIP
	f.DeniedLoginMethods = filters.DeniedLoginMethods
	f.DeniedProtocols = filters.DeniedProtocols
	f.FilePatterns = filters.FilePatterns
	f.MaxUploadFileSize = filters.MaxUploadFileSize
	f.TLSUsername = filters.TLSUsername
	f.ExternalAuthDisabled = filters.ExternalAuthDisabled
	f.PreLoginDisabled = filters.PreLoginDisabled
	f.CheckPasswordDisabled = filters.CheckPasswordDisabled
	f.DisableFsChecks = filters.DisableFsChecks
	f.WebClient = filters.WebClient
	f.AllowAPIKeyAuth = filters.AllowAPIKeyAuth
	f.UserType = filters.UserType
	f.BandwidthLimits = filters.BandwidthLimits
	f.ExternalAuthCacheTime = filters.ExternalAuthCacheTime
	f.StartDirectory = filters.StartDirectory
	f.TwoFactorAuthProtocols = filters.TwoFactorAuthProtocols
	f.FTPSecurity = filters.FTPSecurity
	f.IsAnonymous = filters.IsAnonymous
	f.DefaultSharesExpiration = filters.DefaultSharesExpiration
	f.MaxSharesExpiration = filters.MaxSharesExpiration
	f.PasswordExpiration = filters.PasswordExpiration
	f.PasswordStrength = filters.PasswordStrength
}

func (f *userFilters) toSFTPGo(ctx context.Context) (sdk.UserFilters, diag.Diagnostics) {
	var filters sdk.UserFilters
	baseFilters := f.getBaseFilters()
	base, diags := baseFilters.toSFTPGo(ctx)
	if diags.HasError() {
		return filters, diags
	}
	filters.BaseUserFilters = base
	filters.RequirePasswordChange = f.RequirePasswordChange.ValueBool()
	if !f.TLSCerts.IsNull() {
		diags := f.TLSCerts.ElementsAs(ctx, &filters.TLSCerts, false)
		if diags.HasError() {
			return filters, diags
		}
	}

	return filters, nil
}

func (f *userFilters) fromSFTPGo(ctx context.Context, filters *sdk.UserFilters) diag.Diagnostics {
	var base baseUserFilters
	diags := base.fromSFTPGo(ctx, &filters.BaseUserFilters)
	if diags.HasError() {
		return diags
	}
	f.fromBaseFilters(&base)
	f.RequirePasswordChange = getOptionalBool(filters.RequirePasswordChange)
	tlsCerts, diags := types.ListValueFrom(ctx, types.StringType, filters.TLSCerts)
	if diags.HasError() {
		return diags
	}
	f.TLSCerts = tlsCerts

	return nil
}

type osFsConfig struct {
	ReadBufferSize  types.Int64 `tfsdk:"read_buffer_size"`
	WriteBufferSize types.Int64 `tfsdk:"write_buffer_size"`
}

type s3FsConfig struct {
	Bucket              types.String `tfsdk:"bucket"`
	KeyPrefix           types.String `tfsdk:"key_prefix"`
	Region              types.String `tfsdk:"region"`
	AccessKey           types.String `tfsdk:"access_key"`
	AccessSecret        types.String `tfsdk:"access_secret"`
	RoleARN             types.String `tfsdk:"role_arn"`
	Endpoint            types.String `tfsdk:"endpoint"`
	StorageClass        types.String `tfsdk:"storage_class"`
	ACL                 types.String `tfsdk:"acl"`
	UploadPartSize      types.Int64  `tfsdk:"upload_part_size"`
	UploadConcurrency   types.Int64  `tfsdk:"upload_concurrency"`
	DownloadPartSize    types.Int64  `tfsdk:"download_part_size"`
	UploadPartMaxTime   types.Int64  `tfsdk:"upload_part_max_time"`
	DownloadConcurrency types.Int64  `tfsdk:"download_concurrency"`
	DownloadPartMaxTime types.Int64  `tfsdk:"download_part_max_time"`
	ForcePathStyle      types.Bool   `tfsdk:"force_path_style"`
	SkipTLSVerify       types.Bool   `tfsdk:"skip_tls_verify"`
}

type gcsFsConfig struct {
	Bucket               types.String `tfsdk:"bucket"`
	KeyPrefix            types.String `tfsdk:"key_prefix"`
	Credentials          types.String `tfsdk:"credentials"`
	AutomaticCredentials types.Int64  `tfsdk:"automatic_credentials"`
	StorageClass         types.String `tfsdk:"storage_class"`
	ACL                  types.String `tfsdk:"acl"`
	UploadPartSize       types.Int64  `tfsdk:"upload_part_size"`
	UploadPartMaxTime    types.Int64  `tfsdk:"upload_part_max_time"`
}

type azBlobFsConfig struct {
	Container           types.String `tfsdk:"container"`
	AccountName         types.String `tfsdk:"account_name"`
	AccountKey          types.String `tfsdk:"account_key"`
	SASURL              types.String `tfsdk:"sas_url"`
	Endpoint            types.String `tfsdk:"endpoint"`
	KeyPrefix           types.String `tfsdk:"key_prefix"`
	UploadPartSize      types.Int64  `tfsdk:"upload_part_size"`
	UploadConcurrency   types.Int64  `tfsdk:"upload_concurrency"`
	DownloadPartSize    types.Int64  `tfsdk:"download_part_size"`
	DownloadConcurrency types.Int64  `tfsdk:"download_concurrency"`
	UseEmulator         types.Bool   `tfsdk:"use_emulator"`
	AccessTier          types.String `tfsdk:"access_tier"`
}

type cryptFsConfig struct {
	Passphrase      types.String `tfsdk:"passphrase"`
	ReadBufferSize  types.Int64  `tfsdk:"read_buffer_size"`
	WriteBufferSize types.Int64  `tfsdk:"write_buffer_size"`
}

type sftpFsConfig struct {
	Endpoint                types.String `tfsdk:"endpoint"`
	Username                types.String `tfsdk:"username"`
	Password                types.String `tfsdk:"password"`
	PrivateKey              types.String `tfsdk:"private_key"`
	Fingerprints            types.List   `tfsdk:"fingerprints"`
	Prefix                  types.String `tfsdk:"prefix"`
	DisableCouncurrentReads types.Bool   `tfsdk:"disable_concurrent_reads"`
	BufferSize              types.Int64  `tfsdk:"buffer_size"`
	EqualityCheckMode       types.Int64  `tfsdk:"equality_check_mode"`
}

type httpFsConfig struct {
	Endpoint          types.String `tfsdk:"endpoint"`
	Username          types.String `tfsdk:"username"`
	Password          types.String `tfsdk:"password"`
	APIKey            types.String `tfsdk:"api_key"`
	SkipTLSVerify     types.Bool   `tfsdk:"skip_tls_verify"`
	EqualityCheckMode types.Int64  `tfsdk:"equality_check_mode"`
}

type filesystem struct {
	Provider     types.Int64     `tfsdk:"provider"`
	OSConfig     *osFsConfig     `tfsdk:"osconfig"`
	S3Config     *s3FsConfig     `tfsdk:"s3config"`
	GCSConfig    *gcsFsConfig    `tfsdk:"gcsconfig"`
	AzBlobConfig *azBlobFsConfig `tfsdk:"azblobconfig"`
	CryptConfig  *cryptFsConfig  `tfsdk:"cryptconfig"`
	SFTPConfig   *sftpFsConfig   `tfsdk:"sftpconfig"`
	HTTPConfig   *httpFsConfig   `tfsdk:"httpconfig"`
}

func (f *filesystem) ensureNotNull() {
	if f.OSConfig == nil {
		f.OSConfig = &osFsConfig{}
	}
	if f.S3Config == nil {
		f.S3Config = &s3FsConfig{}
	}
	if f.GCSConfig == nil {
		f.GCSConfig = &gcsFsConfig{}
	}
	if f.AzBlobConfig == nil {
		f.AzBlobConfig = &azBlobFsConfig{}
	}
	if f.CryptConfig == nil {
		f.CryptConfig = &cryptFsConfig{}
	}
	if f.SFTPConfig == nil {
		f.SFTPConfig = &sftpFsConfig{}
	}
	if f.HTTPConfig == nil {
		f.HTTPConfig = &httpFsConfig{}
	}
}

func (f *filesystem) getTFAttributes() map[string]attr.Type {
	return map[string]attr.Type{
		"provider": types.Int64Type,
		"osconfig": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"read_buffer_size":  types.Int64Type,
				"write_buffer_size": types.Int64Type,
			},
		},
		"s3config": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"bucket":                 types.StringType,
				"key_prefix":             types.StringType,
				"region":                 types.StringType,
				"access_key":             types.StringType,
				"access_secret":          types.StringType,
				"role_arn":               types.StringType,
				"endpoint":               types.StringType,
				"storage_class":          types.StringType,
				"acl":                    types.StringType,
				"upload_part_size":       types.Int64Type,
				"upload_concurrency":     types.Int64Type,
				"download_part_size":     types.Int64Type,
				"upload_part_max_time":   types.Int64Type,
				"download_concurrency":   types.Int64Type,
				"download_part_max_time": types.Int64Type,
				"force_path_style":       types.BoolType,
				"skip_tls_verify":        types.BoolType,
			},
		},
		"gcsconfig": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"bucket":                types.StringType,
				"key_prefix":            types.StringType,
				"credentials":           types.StringType,
				"automatic_credentials": types.Int64Type,
				"storage_class":         types.StringType,
				"acl":                   types.StringType,
				"upload_part_size":      types.Int64Type,
				"upload_part_max_time":  types.Int64Type,
			},
		},
		"azblobconfig": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"container":            types.StringType,
				"account_name":         types.StringType,
				"account_key":          types.StringType,
				"sas_url":              types.StringType,
				"endpoint":             types.StringType,
				"key_prefix":           types.StringType,
				"upload_part_size":     types.Int64Type,
				"upload_concurrency":   types.Int64Type,
				"download_part_size":   types.Int64Type,
				"download_concurrency": types.Int64Type,
				"use_emulator":         types.BoolType,
				"access_tier":          types.StringType,
			},
		},
		"cryptconfig": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"passphrase":        types.StringType,
				"read_buffer_size":  types.Int64Type,
				"write_buffer_size": types.Int64Type,
			},
		},
		"sftpconfig": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"endpoint":    types.StringType,
				"username":    types.StringType,
				"password":    types.StringType,
				"private_key": types.StringType,
				"fingerprints": types.ListType{
					ElemType: types.StringType,
				},
				"prefix":                   types.StringType,
				"disable_concurrent_reads": types.BoolType,
				"buffer_size":              types.Int64Type,
				"equality_check_mode":      types.Int64Type,
			},
		},
		"httpconfig": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"endpoint":            types.StringType,
				"username":            types.StringType,
				"password":            types.StringType,
				"api_key":             types.StringType,
				"skip_tls_verify":     types.BoolType,
				"equality_check_mode": types.Int64Type,
			},
		},
	}
}

func (f *filesystem) toSFTPGo(ctx context.Context) (sdk.Filesystem, diag.Diagnostics) {
	f.ensureNotNull()
	fs := sdk.Filesystem{
		Provider: sdk.FilesystemProvider(f.Provider.ValueInt64()),
		OSConfig: sdk.OSFsConfig{
			ReadBufferSize:  int(f.OSConfig.ReadBufferSize.ValueInt64()),
			WriteBufferSize: int(f.OSConfig.WriteBufferSize.ValueInt64()),
		},
		S3Config: sdk.S3FsConfig{
			BaseS3FsConfig: sdk.BaseS3FsConfig{
				Bucket:              f.S3Config.Bucket.ValueString(),
				KeyPrefix:           f.S3Config.KeyPrefix.ValueString(),
				Region:              f.S3Config.Region.ValueString(),
				AccessKey:           f.S3Config.AccessKey.ValueString(),
				RoleARN:             f.S3Config.RoleARN.ValueString(),
				Endpoint:            f.S3Config.Endpoint.ValueString(),
				StorageClass:        f.S3Config.StorageClass.ValueString(),
				ACL:                 f.S3Config.ACL.ValueString(),
				UploadPartSize:      f.S3Config.UploadPartSize.ValueInt64(),
				UploadConcurrency:   int(f.S3Config.UploadConcurrency.ValueInt64()),
				DownloadPartSize:    f.S3Config.DownloadPartSize.ValueInt64(),
				UploadPartMaxTime:   int(f.S3Config.UploadPartMaxTime.ValueInt64()),
				DownloadConcurrency: int(f.S3Config.DownloadConcurrency.ValueInt64()),
				DownloadPartMaxTime: int(f.S3Config.DownloadPartMaxTime.ValueInt64()),
				ForcePathStyle:      f.S3Config.ForcePathStyle.ValueBool(),
				SkipTLSVerify:       f.S3Config.SkipTLSVerify.ValueBool(),
			},
			AccessSecret: getSFTPGoSecret(f.S3Config.AccessSecret.ValueString()),
		},
		GCSConfig: sdk.GCSFsConfig{
			BaseGCSFsConfig: sdk.BaseGCSFsConfig{
				Bucket:               f.GCSConfig.Bucket.ValueString(),
				KeyPrefix:            f.GCSConfig.KeyPrefix.ValueString(),
				AutomaticCredentials: int(f.GCSConfig.AutomaticCredentials.ValueInt64()),
				StorageClass:         f.GCSConfig.StorageClass.ValueString(),
				ACL:                  f.GCSConfig.ACL.ValueString(),
				UploadPartSize:       f.GCSConfig.UploadPartSize.ValueInt64(),
				UploadPartMaxTime:    int(f.GCSConfig.UploadPartMaxTime.ValueInt64()),
			},
			Credentials: getSFTPGoSecret(f.GCSConfig.Credentials.ValueString()),
		},
		AzBlobConfig: sdk.AzBlobFsConfig{
			BaseAzBlobFsConfig: sdk.BaseAzBlobFsConfig{
				Container:           f.AzBlobConfig.Container.ValueString(),
				AccountName:         f.AzBlobConfig.AccountName.ValueString(),
				Endpoint:            f.AzBlobConfig.Endpoint.ValueString(),
				KeyPrefix:           f.AzBlobConfig.KeyPrefix.ValueString(),
				UploadPartSize:      f.AzBlobConfig.UploadPartSize.ValueInt64(),
				UploadConcurrency:   int(f.AzBlobConfig.UploadConcurrency.ValueInt64()),
				DownloadPartSize:    f.AzBlobConfig.DownloadPartSize.ValueInt64(),
				DownloadConcurrency: int(f.AzBlobConfig.DownloadConcurrency.ValueInt64()),
				UseEmulator:         f.AzBlobConfig.UseEmulator.ValueBool(),
				AccessTier:          f.AzBlobConfig.AccessTier.ValueString(),
			},
			AccountKey: getSFTPGoSecret(f.AzBlobConfig.AccountKey.ValueString()),
			SASURL:     getSFTPGoSecret(f.AzBlobConfig.SASURL.ValueString()),
		},
		CryptConfig: sdk.CryptFsConfig{
			Passphrase: getSFTPGoSecret(f.CryptConfig.Passphrase.ValueString()),
			OSFsConfig: sdk.OSFsConfig{
				ReadBufferSize:  int(f.CryptConfig.ReadBufferSize.ValueInt64()),
				WriteBufferSize: int(f.CryptConfig.WriteBufferSize.ValueInt64()),
			},
		},
		SFTPConfig: sdk.SFTPFsConfig{
			BaseSFTPFsConfig: sdk.BaseSFTPFsConfig{
				Endpoint:                f.SFTPConfig.Endpoint.ValueString(),
				Username:                f.SFTPConfig.Username.ValueString(),
				Prefix:                  f.SFTPConfig.Prefix.ValueString(),
				DisableCouncurrentReads: f.SFTPConfig.DisableCouncurrentReads.ValueBool(),
				BufferSize:              f.SFTPConfig.BufferSize.ValueInt64(),
				EqualityCheckMode:       int(f.SFTPConfig.EqualityCheckMode.ValueInt64()),
			},
			Password:   getSFTPGoSecret(f.SFTPConfig.Password.ValueString()),
			PrivateKey: getSFTPGoSecret(f.SFTPConfig.PrivateKey.ValueString()),
		},
		HTTPConfig: sdk.HTTPFsConfig{
			BaseHTTPFsConfig: sdk.BaseHTTPFsConfig{
				Endpoint:          f.HTTPConfig.Endpoint.ValueString(),
				Username:          f.HTTPConfig.Username.ValueString(),
				SkipTLSVerify:     f.HTTPConfig.SkipTLSVerify.ValueBool(),
				EqualityCheckMode: int(f.HTTPConfig.EqualityCheckMode.ValueInt64()),
			},
			Password: getSFTPGoSecret(f.HTTPConfig.Password.ValueString()),
			APIKey:   getSFTPGoSecret(f.HTTPConfig.APIKey.ValueString()),
		},
	}

	if !f.SFTPConfig.Fingerprints.IsNull() {
		diags := f.SFTPConfig.Fingerprints.ElementsAs(ctx, &fs.SFTPConfig.Fingerprints, false)
		if diags.HasError() {
			return fs, diags
		}
	}
	return fs, nil
}

func (f *filesystem) fromSFTPGo(ctx context.Context, fs *sdk.Filesystem) diag.Diagnostics {
	f.Provider = types.Int64Value(int64(fs.Provider))
	f.OSConfig = nil
	f.S3Config = nil
	f.GCSConfig = nil
	f.AzBlobConfig = nil
	f.CryptConfig = nil
	f.SFTPConfig = nil
	f.HTTPConfig = nil
	switch fs.Provider {
	case sdk.LocalFilesystemProvider:
		if fs.OSConfig.ReadBufferSize > 0 || fs.OSConfig.WriteBufferSize > 0 {
			f.OSConfig = &osFsConfig{
				ReadBufferSize:  getOptionalInt64(int64(fs.OSConfig.ReadBufferSize)),
				WriteBufferSize: getOptionalInt64(int64(fs.OSConfig.WriteBufferSize)),
			}
		}
	case sdk.S3FilesystemProvider:
		f.S3Config = &s3FsConfig{
			Bucket:              getOptionalString(fs.S3Config.Bucket),
			KeyPrefix:           getOptionalString(fs.S3Config.KeyPrefix),
			Region:              getOptionalString(fs.S3Config.Region),
			AccessKey:           getOptionalString(fs.S3Config.AccessKey),
			AccessSecret:        getOptionalString(getSecretFromSFTPGo(fs.S3Config.AccessSecret)),
			RoleARN:             getOptionalString(fs.S3Config.RoleARN),
			Endpoint:            getOptionalString(fs.S3Config.Endpoint),
			StorageClass:        getOptionalString(fs.S3Config.StorageClass),
			ACL:                 getOptionalString(fs.S3Config.ACL),
			UploadPartSize:      getOptionalInt64(fs.S3Config.UploadPartSize),
			UploadConcurrency:   getOptionalInt64(int64(fs.S3Config.UploadConcurrency)),
			DownloadPartSize:    getOptionalInt64(fs.S3Config.DownloadPartSize),
			UploadPartMaxTime:   getOptionalInt64(int64(fs.S3Config.UploadPartMaxTime)),
			DownloadConcurrency: getOptionalInt64(int64(fs.S3Config.DownloadConcurrency)),
			DownloadPartMaxTime: getOptionalInt64(int64(fs.S3Config.DownloadPartMaxTime)),
			ForcePathStyle:      getOptionalBool(fs.S3Config.ForcePathStyle),
			SkipTLSVerify:       getOptionalBool(fs.S3Config.SkipTLSVerify),
		}
	case sdk.GCSFilesystemProvider:
		f.GCSConfig = &gcsFsConfig{
			Bucket:               getOptionalString(fs.GCSConfig.Bucket),
			KeyPrefix:            getOptionalString(fs.GCSConfig.KeyPrefix),
			Credentials:          getOptionalString(getSecretFromSFTPGo(fs.GCSConfig.Credentials)),
			AutomaticCredentials: getOptionalInt64(int64(fs.GCSConfig.AutomaticCredentials)),
			StorageClass:         getOptionalString(fs.GCSConfig.StorageClass),
			ACL:                  getOptionalString(fs.GCSConfig.ACL),
			UploadPartSize:       getOptionalInt64(fs.GCSConfig.UploadPartSize),
			UploadPartMaxTime:    getOptionalInt64(int64(fs.GCSConfig.UploadPartMaxTime)),
		}
	case sdk.AzureBlobFilesystemProvider:
		f.AzBlobConfig = &azBlobFsConfig{
			Container:           getOptionalString(fs.AzBlobConfig.Container),
			AccountName:         getOptionalString(fs.AzBlobConfig.AccountName),
			AccountKey:          getOptionalString(getSecretFromSFTPGo(fs.AzBlobConfig.AccountKey)),
			SASURL:              getOptionalString(getSecretFromSFTPGo(fs.AzBlobConfig.SASURL)),
			Endpoint:            getOptionalString(fs.AzBlobConfig.Endpoint),
			KeyPrefix:           getOptionalString(fs.AzBlobConfig.KeyPrefix),
			UploadPartSize:      getOptionalInt64(fs.AzBlobConfig.UploadPartSize),
			UploadConcurrency:   getOptionalInt64(int64(fs.AzBlobConfig.UploadConcurrency)),
			DownloadPartSize:    getOptionalInt64(fs.AzBlobConfig.DownloadPartSize),
			DownloadConcurrency: getOptionalInt64(int64(fs.AzBlobConfig.DownloadConcurrency)),
			UseEmulator:         getOptionalBool(fs.AzBlobConfig.UseEmulator),
			AccessTier:          getOptionalString(fs.AzBlobConfig.AccessTier),
		}
	case sdk.CryptedFilesystemProvider:
		f.CryptConfig = &cryptFsConfig{
			Passphrase:      getOptionalString(getSecretFromSFTPGo(fs.CryptConfig.Passphrase)),
			ReadBufferSize:  getOptionalInt64(int64(fs.CryptConfig.ReadBufferSize)),
			WriteBufferSize: getOptionalInt64(int64(fs.CryptConfig.WriteBufferSize)),
		}
	case sdk.SFTPFilesystemProvider:
		f.SFTPConfig = &sftpFsConfig{
			Endpoint:                getOptionalString(fs.SFTPConfig.Endpoint),
			Username:                getOptionalString(fs.SFTPConfig.Username),
			Password:                getOptionalString(getSecretFromSFTPGo(fs.SFTPConfig.Password)),
			PrivateKey:              getOptionalString(getSecretFromSFTPGo(fs.SFTPConfig.PrivateKey)),
			Prefix:                  getOptionalString(fs.SFTPConfig.Prefix),
			DisableCouncurrentReads: getOptionalBool(fs.SFTPConfig.DisableCouncurrentReads),
			BufferSize:              getOptionalInt64(fs.SFTPConfig.BufferSize),
			EqualityCheckMode:       getOptionalInt64(int64(fs.SFTPConfig.EqualityCheckMode)),
		}
		fingerprints, diags := types.ListValueFrom(ctx, types.StringType, fs.SFTPConfig.Fingerprints)
		if diags.HasError() {
			return diags
		}
		f.SFTPConfig.Fingerprints = fingerprints
	case sdk.HTTPFilesystemProvider:
		f.HTTPConfig = &httpFsConfig{
			Endpoint:          getOptionalString(fs.HTTPConfig.Endpoint),
			Username:          getOptionalString(fs.HTTPConfig.Username),
			Password:          getOptionalString(getSecretFromSFTPGo(fs.HTTPConfig.Password)),
			APIKey:            getOptionalString(getSecretFromSFTPGo(fs.HTTPConfig.APIKey)),
			SkipTLSVerify:     getOptionalBool(fs.HTTPConfig.SkipTLSVerify),
			EqualityCheckMode: getOptionalInt64(int64(fs.HTTPConfig.EqualityCheckMode)),
		}
	}

	return nil
}

type virtualFolderResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	MappedPath      types.String `tfsdk:"mapped_path"`
	Description     types.String `tfsdk:"description"`
	UsedQuotaSize   types.Int64  `tfsdk:"used_quota_size"`
	UsedQuotaFiles  types.Int64  `tfsdk:"used_quota_files"`
	LastQuotaUpdate types.Int64  `tfsdk:"last_quota_update"`
	FsConfig        types.Object `tfsdk:"filesystem"`
}

func (f *virtualFolderResourceModel) toSFTPGo(ctx context.Context) (*sdk.BaseVirtualFolder, diag.Diagnostics) {
	folder := &sdk.BaseVirtualFolder{
		Name:            f.Name.ValueString(),
		MappedPath:      f.MappedPath.ValueString(),
		Description:     f.Description.ValueString(),
		UsedQuotaSize:   f.UsedQuotaSize.ValueInt64(),
		UsedQuotaFiles:  int(f.UsedQuotaFiles.ValueInt64()),
		LastQuotaUpdate: f.LastQuotaUpdate.ValueInt64(),
	}
	var fs filesystem
	diags := f.FsConfig.As(ctx, &fs, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return folder, diags
	}
	sftpgoFs, diags := fs.toSFTPGo(ctx)
	if diags.HasError() {
		return folder, diags
	}
	folder.FsConfig = sftpgoFs

	return folder, nil
}

func (f *virtualFolderResourceModel) fromSFTPGo(ctx context.Context, folder *sdk.BaseVirtualFolder) diag.Diagnostics {
	f.Name = types.StringValue(folder.Name)
	f.ID = f.Name
	f.MappedPath = getOptionalString(folder.MappedPath)
	f.Description = getOptionalString(folder.Description)
	f.UsedQuotaSize = types.Int64Value(folder.UsedQuotaSize)
	f.UsedQuotaFiles = types.Int64Value(int64(folder.UsedQuotaFiles))
	f.LastQuotaUpdate = types.Int64Value(folder.LastQuotaUpdate)

	var fsConfig filesystem
	diags := fsConfig.fromSFTPGo(ctx, &folder.FsConfig)
	if diags.HasError() {
		return diags
	}
	fs, diags := types.ObjectValueFrom(ctx, fsConfig.getTFAttributes(), fsConfig)
	if diags.HasError() {
		return diags
	}
	f.FsConfig = fs

	return nil
}

type virtualFolder struct {
	// embedded structs are not supported
	//baseVirtualFolder
	Name            types.String `tfsdk:"name"`
	MappedPath      types.String `tfsdk:"mapped_path"`
	Description     types.String `tfsdk:"description"`
	UsedQuotaSize   types.Int64  `tfsdk:"used_quota_size"`
	UsedQuotaFiles  types.Int64  `tfsdk:"used_quota_files"`
	LastQuotaUpdate types.Int64  `tfsdk:"last_quota_update"`
	FsConfig        types.Object `tfsdk:"filesystem"`
	VirtualPath     types.String `tfsdk:"virtual_path"`
	QuotaSize       types.Int64  `tfsdk:"quota_size"`
	QuotaFiles      types.Int64  `tfsdk:"quota_files"`
}

func (f *virtualFolder) getBaseFolder() virtualFolderResourceModel {
	return virtualFolderResourceModel{
		Name:            f.Name,
		MappedPath:      f.MappedPath,
		Description:     f.Description,
		UsedQuotaSize:   f.UsedQuotaSize,
		UsedQuotaFiles:  f.UsedQuotaFiles,
		LastQuotaUpdate: f.LastQuotaUpdate,
		FsConfig:        f.FsConfig,
	}
}

func (f *virtualFolder) fromBaseFolder(folder *virtualFolderResourceModel) {
	f.Name = folder.Name
	f.MappedPath = folder.MappedPath
	f.Description = folder.Description
	f.UsedQuotaSize = folder.UsedQuotaSize
	f.UsedQuotaFiles = folder.UsedQuotaFiles
	f.LastQuotaUpdate = folder.LastQuotaUpdate
	f.FsConfig = folder.FsConfig
}

func (f *virtualFolder) toSFTPGo(ctx context.Context) (sdk.VirtualFolder, diag.Diagnostics) {
	folder := sdk.VirtualFolder{
		VirtualPath: f.VirtualPath.ValueString(),
		QuotaSize:   f.QuotaSize.ValueInt64(),
		QuotaFiles:  int(f.QuotaFiles.ValueInt64()),
	}
	baseFolder := f.getBaseFolder()
	base, diags := baseFolder.toSFTPGo(ctx)
	if diags.HasError() {
		return folder, diags
	}
	folder.BaseVirtualFolder = *base

	return folder, nil
}

func (f *virtualFolder) fromSFTPGo(ctx context.Context, folder *sdk.VirtualFolder) diag.Diagnostics {
	var base virtualFolderResourceModel
	diags := base.fromSFTPGo(ctx, &folder.BaseVirtualFolder)
	if diags.HasError() {
		return diags
	}
	f.fromBaseFolder(&base)
	f.VirtualPath = types.StringValue(folder.VirtualPath)
	f.QuotaSize = types.Int64Value(folder.QuotaSize)
	f.QuotaFiles = types.Int64Value(int64(folder.QuotaFiles))
	return nil
}

type roleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.Int64  `tfsdk:"created_at"`
	UpdatedAt   types.Int64  `tfsdk:"updated_at"`
}

func (r *roleResourceModel) toSFTPGo(ctx context.Context) (*client.Role, diag.Diagnostics) {
	role := &client.Role{
		Name:        r.Name.ValueString(),
		Description: r.Description.ValueString(),
		CreatedAt:   r.CreatedAt.ValueInt64(),
		UpdatedAt:   r.UpdatedAt.ValueInt64(),
	}

	return role, nil
}

func (r *roleResourceModel) fromSFTPGo(ctx context.Context, role *client.Role) diag.Diagnostics {
	r.Name = types.StringValue(role.Name)
	r.ID = r.Name
	r.Description = getOptionalString(role.Description)
	r.CreatedAt = types.Int64Value(role.CreatedAt)
	r.UpdatedAt = types.Int64Value(role.UpdatedAt)
	return nil
}

type groupUserSettings struct {
	HomeDir              types.String `tfsdk:"home_dir"`
	MaxSessions          types.Int64  `tfsdk:"max_sessions"`
	QuotaSize            types.Int64  `tfsdk:"quota_size"`
	QuotaFiles           types.Int64  `tfsdk:"quota_files"`
	Permissions          types.Map    `tfsdk:"permissions"`
	UploadBandwidth      types.Int64  `tfsdk:"upload_bandwidth"`
	DownloadBandwidth    types.Int64  `tfsdk:"download_bandwidth"`
	UploadDataTransfer   types.Int64  `tfsdk:"upload_data_transfer"`
	DownloadDataTransfer types.Int64  `tfsdk:"download_data_transfer"`
	TotalDataTransfer    types.Int64  `tfsdk:"total_data_transfer"`
	ExpiresIn            types.Int64  `tfsdk:"expires_in"`
	Filters              types.Object `tfsdk:"filters"`
	FsConfig             types.Object `tfsdk:"filesystem"`
}

func (s *groupUserSettings) getTFAttributes() map[string]attr.Type {
	filters := baseUserFilters{}
	fs := filesystem{}
	return map[string]attr.Type{
		"home_dir":     types.StringType,
		"max_sessions": types.Int64Type,
		"quota_size":   types.Int64Type,
		"quota_files":  types.Int64Type,
		"permissions": types.MapType{
			ElemType: types.StringType,
		},
		"upload_bandwidth":       types.Int64Type,
		"download_bandwidth":     types.Int64Type,
		"upload_data_transfer":   types.Int64Type,
		"download_data_transfer": types.Int64Type,
		"total_data_transfer":    types.Int64Type,
		"expires_in":             types.Int64Type,
		"filters": types.ObjectType{
			AttrTypes: filters.getTFAttributes(),
		},
		"filesystem": types.ObjectType{
			AttrTypes: fs.getTFAttributes(),
		},
	}
}

func (s *groupUserSettings) toSFTPGo(ctx context.Context) (sdk.GroupUserSettings, diag.Diagnostics) {
	settings := sdk.GroupUserSettings{
		BaseGroupUserSettings: sdk.BaseGroupUserSettings{
			HomeDir:              s.HomeDir.ValueString(),
			MaxSessions:          int(s.MaxSessions.ValueInt64()),
			QuotaSize:            s.QuotaSize.ValueInt64(),
			QuotaFiles:           int(s.QuotaFiles.ValueInt64()),
			UploadBandwidth:      s.UploadBandwidth.ValueInt64(),
			DownloadBandwidth:    s.DownloadBandwidth.ValueInt64(),
			UploadDataTransfer:   s.UploadDataTransfer.ValueInt64(),
			DownloadDataTransfer: s.DownloadDataTransfer.ValueInt64(),
			TotalDataTransfer:    s.TotalDataTransfer.ValueInt64(),
			ExpiresIn:            int(s.ExpiresIn.ValueInt64()),
		},
	}
	permissions := make(map[string]string)
	if !s.Permissions.IsNull() {
		diags := s.Permissions.ElementsAs(ctx, &permissions, false)
		if diags.HasError() {
			return settings, diags
		}
	}
	settings.Permissions = make(map[string][]string)
	for k, v := range permissions {
		settings.Permissions[k] = strings.Split(v, ",")
	}

	var filters baseUserFilters
	diags := s.Filters.As(ctx, &filters, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return settings, diags
	}
	sftpgoFilters, diags := filters.toSFTPGo(ctx)
	if diags.HasError() {
		return settings, diags
	}
	settings.Filters = sftpgoFilters

	var fs filesystem
	diags = s.FsConfig.As(ctx, &fs, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return settings, diags
	}
	sftpgoFs, diags := fs.toSFTPGo(ctx)
	if diags.HasError() {
		return settings, diags
	}
	settings.FsConfig = sftpgoFs

	return settings, nil
}

func (s *groupUserSettings) fromSFTPGo(ctx context.Context, settings *sdk.GroupUserSettings) diag.Diagnostics {
	s.HomeDir = getOptionalString(settings.HomeDir)
	s.MaxSessions = getOptionalInt64(int64(settings.MaxSessions))
	s.QuotaSize = getOptionalInt64(settings.QuotaSize)
	s.QuotaFiles = getOptionalInt64(int64(settings.QuotaFiles))

	permissions := make(map[string]string)
	for k, v := range settings.Permissions {
		permissions[k] = strings.Join(v, ",")
	}
	if len(permissions) > 0 {
		tfMap, diags := types.MapValueFrom(ctx, types.StringType, permissions)
		if diags.HasError() {
			return diags
		}
		s.Permissions = tfMap
	} else {
		s.Permissions = types.MapNull(types.StringType)
	}

	s.UploadBandwidth = getOptionalInt64(settings.UploadBandwidth)
	s.DownloadBandwidth = getOptionalInt64(settings.DownloadBandwidth)
	s.UploadDataTransfer = getOptionalInt64(settings.UploadDataTransfer)
	s.DownloadDataTransfer = getOptionalInt64(settings.DownloadDataTransfer)
	s.TotalDataTransfer = getOptionalInt64(settings.TotalDataTransfer)
	s.ExpiresIn = getOptionalInt64(int64(settings.ExpiresIn))

	var f baseUserFilters
	diags := f.fromSFTPGo(ctx, &settings.Filters)
	if diags.HasError() {
		return diags
	}
	filters, diags := types.ObjectValueFrom(ctx, f.getTFAttributes(), f)
	if diags.HasError() {
		return diags
	}
	s.Filters = filters

	var fsConfig filesystem
	diags = fsConfig.fromSFTPGo(ctx, &settings.FsConfig)
	if diags.HasError() {
		return diags
	}
	fs, diags := types.ObjectValueFrom(ctx, fsConfig.getTFAttributes(), fsConfig)
	if diags.HasError() {
		return diags
	}
	s.FsConfig = fs

	return nil
}

type groupResourceModel struct {
	ID             types.String    `tfsdk:"id"`
	Name           types.String    `tfsdk:"name"`
	Description    types.String    `tfsdk:"description"`
	CreatedAt      types.Int64     `tfsdk:"created_at"`
	UpdatedAt      types.Int64     `tfsdk:"updated_at"`
	UserSettings   types.Object    `tfsdk:"user_settings"`
	VirtualFolders []virtualFolder `tfsdk:"virtual_folders"`
}

func (g *groupResourceModel) toSFTPGo(ctx context.Context) (*sdk.Group, diag.Diagnostics) {
	group := &sdk.Group{
		BaseGroup: sdk.BaseGroup{
			Name:        g.Name.ValueString(),
			Description: g.Description.ValueString(),
			CreatedAt:   g.CreatedAt.ValueInt64(),
			UpdatedAt:   g.UpdatedAt.ValueInt64(),
		},
	}

	var settings groupUserSettings
	diags := g.UserSettings.As(ctx, &settings, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return group, diags
	}
	sftpgoSettings, diags := settings.toSFTPGo(ctx)
	if diags.HasError() {
		return group, diags
	}
	group.UserSettings = sftpgoSettings

	for _, f := range g.VirtualFolders {
		folder, diags := f.toSFTPGo(ctx)
		if diags.HasError() {
			return group, diags
		}
		group.VirtualFolders = append(group.VirtualFolders, folder)
	}

	return group, nil
}

func (g *groupResourceModel) fromSFTPGo(ctx context.Context, group *sdk.Group) diag.Diagnostics {
	g.Name = types.StringValue(group.Name)
	g.ID = g.Name
	g.Description = getOptionalString(group.Description)
	g.CreatedAt = types.Int64Value(group.CreatedAt)
	g.UpdatedAt = types.Int64Value(group.UpdatedAt)

	var s groupUserSettings
	diags := s.fromSFTPGo(ctx, &group.UserSettings)
	if diags.HasError() {
		return diags
	}
	settings, diags := types.ObjectValueFrom(ctx, s.getTFAttributes(), s)
	if diags.HasError() {
		return diags
	}
	g.UserSettings = settings

	g.VirtualFolders = nil
	for _, f := range group.VirtualFolders {
		var folder virtualFolder
		diags := folder.fromSFTPGo(ctx, &f)
		if diags.HasError() {
			return diags
		}
		g.VirtualFolders = append(g.VirtualFolders, folder)
	}

	return nil
}

type adminPreferences struct {
	HideUserPageSections   types.Int64 `tfsdk:"hide_user_page_sections"`
	DefaultUsersExpiration types.Int64 `tfsdk:"default_users_expiration"`
}

func (*adminPreferences) getTFAttributes() map[string]attr.Type {
	return map[string]attr.Type{
		"hide_user_page_sections":  types.Int64Type,
		"default_users_expiration": types.Int64Type,
	}
}

func (p *adminPreferences) toSFTPGo(ctx context.Context) (client.AdminPreferences, diag.Diagnostics) {
	return client.AdminPreferences{
		HideUserPageSections:   int(p.HideUserPageSections.ValueInt64()),
		DefaultUsersExpiration: int(p.DefaultUsersExpiration.ValueInt64()),
	}, nil
}

func (p *adminPreferences) fromSFTPGo(ctx context.Context, preferences *client.AdminPreferences) diag.Diagnostics {
	p.HideUserPageSections = getOptionalInt64(int64(preferences.HideUserPageSections))
	p.DefaultUsersExpiration = getOptionalInt64(int64(preferences.DefaultUsersExpiration))
	return nil
}

type adminFilters struct {
	AllowList             types.List `tfsdk:"allow_list"`
	AllowAPIKeyAuth       types.Bool `tfsdk:"allow_api_key_auth"`
	RequireTwoFactor      types.Bool `tfsdk:"require_two_factor"`
	RequirePasswordChange types.Bool `tfsdk:"require_password_change"`
}

func (f *adminFilters) getTFAttributes() map[string]attr.Type {
	return map[string]attr.Type{
		"allow_list": types.ListType{
			ElemType: types.StringType,
		},
		"allow_api_key_auth":      types.BoolType,
		"require_two_factor":      types.BoolType,
		"require_password_change": types.BoolType,
	}
}

func (f *adminFilters) toSFTPGo(ctx context.Context) (client.AdminFilters, diag.Diagnostics) {
	filters := client.AdminFilters{
		AllowAPIKeyAuth:       f.AllowAPIKeyAuth.ValueBool(),
		RequireTwoFactor:      f.RequireTwoFactor.ValueBool(),
		RequirePasswordChange: f.RequirePasswordChange.ValueBool(),
	}
	if !f.AllowList.IsNull() {
		diags := f.AllowList.ElementsAs(ctx, &filters.AllowList, false)
		if diags.HasError() {
			return filters, diags
		}
	}
	return filters, nil
}

func (f *adminFilters) fromSFTPGo(ctx context.Context, filters *client.AdminFilters) diag.Diagnostics {
	allowList, diags := types.ListValueFrom(ctx, types.StringType, filters.AllowList)
	if diags.HasError() {
		return diags
	}
	f.AllowList = allowList
	f.AllowAPIKeyAuth = getOptionalBool(filters.AllowAPIKeyAuth)
	f.RequireTwoFactor = getOptionalBool(filters.RequireTwoFactor)
	f.RequirePasswordChange = getOptionalBool(filters.RequirePasswordChange)
	return nil
}

type adminGroupMappingOptions struct {
	AddToUsersAs types.Int64 `tfsdk:"add_to_users_as"`
}

func (o *adminGroupMappingOptions) getTFAttributes() map[string]attr.Type {
	return map[string]attr.Type{
		"add_to_users_as": types.Int64Type,
	}
}

func (o *adminGroupMappingOptions) toSFTPGo(ctx context.Context) (client.AdminGroupMappingOptions, diag.Diagnostics) {
	options := client.AdminGroupMappingOptions{
		AddToUsersAs: int(o.AddToUsersAs.ValueInt64()),
	}
	return options, nil
}

func (o *adminGroupMappingOptions) fromSFTPGo(ctx context.Context, options *client.AdminGroupMappingOptions) diag.Diagnostics {
	o.AddToUsersAs = getOptionalInt64(int64(options.AddToUsersAs))
	return nil
}

type adminGroupMapping struct {
	Name    types.String `tfsdk:"name"`
	Options types.Object `tfsdk:"options"`
}

func (m *adminGroupMapping) toSFTPGo(ctx context.Context) (client.AdminGroupMapping, diag.Diagnostics) {
	mapping := client.AdminGroupMapping{
		Name: m.Name.ValueString(),
	}
	var options adminGroupMappingOptions
	diags := m.Options.As(ctx, &options, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return mapping, diags
	}
	sftpgoOptions, diags := options.toSFTPGo(ctx)
	if diags.HasError() {
		return mapping, diags
	}
	mapping.Options = sftpgoOptions

	return mapping, nil
}

func (m *adminGroupMapping) fromSFTPGo(ctx context.Context, mapping *client.AdminGroupMapping) diag.Diagnostics {
	m.Name = types.StringValue(mapping.Name)

	var options adminGroupMappingOptions
	diags := options.fromSFTPGo(ctx, &mapping.Options)
	if diags.HasError() {
		return diags
	}
	o, diags := types.ObjectValueFrom(ctx, options.getTFAttributes(), options)
	if diags.HasError() {
		return diags
	}
	m.Options = o

	return nil
}

type adminResourceModel struct {
	ID             types.String        `tfsdk:"id"`
	Username       types.String        `tfsdk:"username"`
	Status         types.Int64         `tfsdk:"status"`
	Email          types.String        `tfsdk:"email"`
	Password       types.String        `tfsdk:"password"`
	Permissions    types.List          `tfsdk:"permissions"`
	Filters        types.Object        `tfsdk:"filters"`
	Preferences    types.Object        `tfsdk:"preferences"`
	Description    types.String        `tfsdk:"description"`
	AdditionalInfo types.String        `tfsdk:"additional_info"`
	Groups         []adminGroupMapping `tfsdk:"groups"`
	CreatedAt      types.Int64         `tfsdk:"created_at"`
	UpdatedAt      types.Int64         `tfsdk:"updated_at"`
	LastLogin      types.Int64         `tfsdk:"last_login"`
	Role           types.String        `tfsdk:"role"`
}

func (a *adminResourceModel) toSFTPGo(ctx context.Context) (*client.Admin, diag.Diagnostics) {
	admin := &client.Admin{
		Username:       a.Username.ValueString(),
		Status:         int(a.Status.ValueInt64()),
		Email:          a.Email.ValueString(),
		Password:       a.Password.ValueString(),
		Description:    a.Description.ValueString(),
		AdditionalInfo: a.AdditionalInfo.ValueString(),
		CreatedAt:      a.CreatedAt.ValueInt64(),
		UpdatedAt:      a.UpdatedAt.ValueInt64(),
		LastLogin:      a.LastLogin.ValueInt64(),
		Role:           a.Role.ValueString(),
	}

	if !a.Permissions.IsNull() {
		diags := a.Permissions.ElementsAs(ctx, &admin.Permissions, false)
		if diags.HasError() {
			return admin, diags
		}
	}

	var filters adminFilters
	diags := a.Filters.As(ctx, &filters, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return admin, diags
	}
	sftpgoFilters, diags := filters.toSFTPGo(ctx)
	if diags.HasError() {
		return admin, diags
	}
	admin.Filters = sftpgoFilters

	var preferences adminPreferences
	diags = a.Preferences.As(ctx, &preferences, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return admin, diags
	}
	sftpgoPreferences, diags := preferences.toSFTPGo(ctx)
	if diags.HasError() {
		return admin, diags
	}
	admin.Filters.Preferences = sftpgoPreferences

	for _, g := range a.Groups {
		group, diags := g.toSFTPGo(ctx)
		if diags.HasError() {
			return admin, diags
		}
		admin.Groups = append(admin.Groups, group)
	}

	return admin, nil
}

func (a *adminResourceModel) fromSFTPGo(ctx context.Context, admin *client.Admin) diag.Diagnostics {
	a.Username = types.StringValue(admin.Username)
	a.ID = a.Username
	a.Status = types.Int64Value(int64(admin.Status))
	a.Email = getOptionalString(admin.Email)
	a.Password = getOptionalString(admin.Password)

	permissions, diags := types.ListValueFrom(ctx, types.StringType, admin.Permissions)
	if diags.HasError() {
		return diags
	}
	a.Permissions = permissions

	var filters adminFilters
	diags = filters.fromSFTPGo(ctx, &admin.Filters)
	if diags.HasError() {
		return diags
	}
	f, diags := types.ObjectValueFrom(ctx, filters.getTFAttributes(), filters)
	if diags.HasError() {
		return diags
	}
	a.Filters = f

	var preferences adminPreferences
	diags = preferences.fromSFTPGo(ctx, &admin.Filters.Preferences)
	if diags.HasError() {
		return diags
	}
	p, diags := types.ObjectValueFrom(ctx, preferences.getTFAttributes(), preferences)
	if diags.HasError() {
		return diags
	}
	a.Preferences = p

	a.Description = getOptionalString(admin.Description)
	a.AdditionalInfo = getOptionalString(admin.AdditionalInfo)

	var groups []adminGroupMapping
	for _, g := range admin.Groups {
		group := adminGroupMapping{}
		diags := group.fromSFTPGo(ctx, &g)
		if diags.HasError() {
			return diags
		}
		groups = append(groups, group)
	}
	a.Groups = groups
	a.CreatedAt = types.Int64Value(admin.CreatedAt)
	a.UpdatedAt = types.Int64Value(admin.UpdatedAt)
	a.LastLogin = types.Int64Value(admin.LastLogin)
	a.Role = getOptionalString(admin.Role)

	return nil
}

type defenderEntryResourceModel struct {
	ID          types.String `tfsdk:"id"`
	IPOrNet     types.String `tfsdk:"ipornet"`
	Description types.String `tfsdk:"description"`
	Mode        types.Int64  `tfsdk:"mode"`
	Protocols   types.Int64  `tfsdk:"protocols"`
	CreatedAt   types.Int64  `tfsdk:"created_at"`
	UpdatedAt   types.Int64  `tfsdk:"updated_at"`
}

func (e *defenderEntryResourceModel) toSFTPGo(ctx context.Context) (*client.IPListEntry, diag.Diagnostics) {
	entry := &client.IPListEntry{
		IPOrNet:     e.IPOrNet.ValueString(),
		Description: e.Description.ValueString(),
		Type:        2,
		Mode:        int(e.Mode.ValueInt64()),
		Protocols:   int(e.Protocols.ValueInt64()),
	}
	return entry, nil
}

func (e *defenderEntryResourceModel) fromSFTPGo(ctx context.Context, entry *client.IPListEntry) diag.Diagnostics {
	e.IPOrNet = types.StringValue(entry.IPOrNet)
	e.ID = e.IPOrNet
	e.Description = getOptionalString(entry.Description)
	e.Mode = types.Int64Value(int64(entry.Mode))
	e.Protocols = types.Int64Value(int64(entry.Protocols))
	e.CreatedAt = types.Int64Value(entry.CreatedAt)
	e.UpdatedAt = types.Int64Value(entry.UpdatedAt)
	return nil
}

type allowListEntryResourceModel struct {
	ID          types.String `tfsdk:"id"`
	IPOrNet     types.String `tfsdk:"ipornet"`
	Description types.String `tfsdk:"description"`
	Protocols   types.Int64  `tfsdk:"protocols"`
	CreatedAt   types.Int64  `tfsdk:"created_at"`
	UpdatedAt   types.Int64  `tfsdk:"updated_at"`
}

func (e *allowListEntryResourceModel) toSFTPGo(ctx context.Context) (*client.IPListEntry, diag.Diagnostics) {
	entry := &client.IPListEntry{
		IPOrNet:     e.IPOrNet.ValueString(),
		Description: e.Description.ValueString(),
		Type:        1,
		Mode:        1,
		Protocols:   int(e.Protocols.ValueInt64()),
	}
	return entry, nil
}

func (e *allowListEntryResourceModel) fromSFTPGo(ctx context.Context, entry *client.IPListEntry) diag.Diagnostics {
	e.IPOrNet = types.StringValue(entry.IPOrNet)
	e.ID = e.IPOrNet
	e.Description = getOptionalString(entry.Description)
	e.Protocols = types.Int64Value(int64(entry.Protocols))
	e.CreatedAt = types.Int64Value(entry.CreatedAt)
	e.UpdatedAt = types.Int64Value(entry.UpdatedAt)
	return nil
}

type rlSafeListEntryResourceModel struct {
	ID          types.String `tfsdk:"id"`
	IPOrNet     types.String `tfsdk:"ipornet"`
	Description types.String `tfsdk:"description"`
	Protocols   types.Int64  `tfsdk:"protocols"`
	CreatedAt   types.Int64  `tfsdk:"created_at"`
	UpdatedAt   types.Int64  `tfsdk:"updated_at"`
}

func (e *rlSafeListEntryResourceModel) toSFTPGo(ctx context.Context) (*client.IPListEntry, diag.Diagnostics) {
	entry := &client.IPListEntry{
		IPOrNet:     e.IPOrNet.ValueString(),
		Description: e.Description.ValueString(),
		Type:        3,
		Mode:        1,
		Protocols:   int(e.Protocols.ValueInt64()),
	}
	return entry, nil
}

func (e *rlSafeListEntryResourceModel) fromSFTPGo(ctx context.Context, entry *client.IPListEntry) diag.Diagnostics {
	e.IPOrNet = types.StringValue(entry.IPOrNet)
	e.ID = e.IPOrNet
	e.Description = getOptionalString(entry.Description)
	e.Protocols = types.Int64Value(int64(entry.Protocols))
	e.CreatedAt = types.Int64Value(entry.CreatedAt)
	e.UpdatedAt = types.Int64Value(entry.UpdatedAt)
	return nil
}

type keyValue struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func (*keyValue) getTFAttributes() map[string]attr.Type {
	return map[string]attr.Type{
		"key":   types.StringType,
		"value": types.StringType,
	}
}

type httpPart struct {
	Name     types.String `tfsdk:"name"`
	Filepath types.String `tfsdk:"filepath"`
	Headers  []keyValue   `tfsdk:"headers"`
	Body     types.String `tfsdk:"body"`
}

type eventActionHTTPConfig struct {
	Endpoint        types.String `tfsdk:"endpoint"`
	Username        types.String `tfsdk:"username"`
	Password        types.String `tfsdk:"password"`
	Headers         []keyValue   `tfsdk:"headers"`
	Timeout         types.Int64  `tfsdk:"timeout"`
	SkipTLSVerify   types.Bool   `tfsdk:"skip_tls_verify"`
	Method          types.String `tfsdk:"method"`
	QueryParameters []keyValue   `tfsdk:"query_parameters"`
	Body            types.String `tfsdk:"body"`
	Parts           []httpPart   `tfsdk:"parts"`
}

type eventActionCommandConfig struct {
	Cmd     types.String `tfsdk:"cmd"`
	Args    types.List   `tfsdk:"args"`
	Timeout types.Int64  `tfsdk:"timeout"`
	EnvVars []keyValue   `tfsdk:"env_vars"`
}

type eventActionEmailConfig struct {
	Recipients  types.List   `tfsdk:"recipients"`
	Bcc         types.List   `tfsdk:"bcc"`
	Subject     types.String `tfsdk:"subject"`
	Body        types.String `tfsdk:"body"`
	Attachments types.List   `tfsdk:"attachments"`
	ContentType types.Int64  `tfsdk:"content_type"`
}

type folderRetention struct {
	Path                  types.String `tfsdk:"path"`
	Retention             types.Int64  `tfsdk:"retention"`
	DeleteEmptyDirs       types.Bool   `tfsdk:"delete_empty_dirs"`
	IgnoreUserPermissions types.Bool   `tfsdk:"ignore_user_permissions"`
}

type eventActionDataRetentionConfig struct {
	Folders []folderRetention `tfsdk:"folders"`
}

type eventActionFsCompress struct {
	Name  types.String `tfsdk:"name"`
	Paths types.List   `tfsdk:"paths"`
}

type eventActionFilesystemConfig struct {
	Type     types.Int64            `tfsdk:"type"`
	Renames  []keyValue             `tfsdk:"renames"`
	MkDirs   types.List             `tfsdk:"mkdirs"`
	Deletes  types.List             `tfsdk:"deletes"`
	Exist    types.List             `tfsdk:"exist"`
	Copy     []keyValue             `tfsdk:"copy"`
	Compress *eventActionFsCompress `tfsdk:"compress"`
}

type eventActionPasswordExpiration struct {
	Threshold types.Int64 `tfsdk:"threshold"`
}

type eventActionIDPAccountCheck struct {
	Mode          types.Int64  `tfsdk:"mode"`
	TemplateUser  types.String `tfsdk:"template_user"`
	TemplateAdmin types.String `tfsdk:"template_admin"`
}

type eventActionOptions struct {
	HTTPConfig          *eventActionHTTPConfig          `tfsdk:"http_config"`
	CmdConfig           *eventActionCommandConfig       `tfsdk:"cmd_config"`
	EmailConfig         *eventActionEmailConfig         `tfsdk:"email_config"`
	RetentionConfig     *eventActionDataRetentionConfig `tfsdk:"retention_config"`
	FsConfig            *eventActionFilesystemConfig    `tfsdk:"fs_config"`
	PwdExpirationConfig *eventActionPasswordExpiration  `tfsdk:"pwd_expiration_config"`
	IDPConfig           *eventActionIDPAccountCheck     `tfsdk:"idp_config"`
}

func (o *eventActionOptions) ensureNotNull() {
	if o.HTTPConfig == nil {
		o.HTTPConfig = &eventActionHTTPConfig{}
	}
	if o.CmdConfig == nil {
		o.CmdConfig = &eventActionCommandConfig{}
	}
	if o.EmailConfig == nil {
		o.EmailConfig = &eventActionEmailConfig{}
	}
	if o.RetentionConfig == nil {
		o.RetentionConfig = &eventActionDataRetentionConfig{}
	}
	if o.FsConfig == nil {
		o.FsConfig = &eventActionFilesystemConfig{}
	}
	if o.FsConfig.Compress == nil {
		o.FsConfig.Compress = &eventActionFsCompress{}
	}
	if o.PwdExpirationConfig == nil {
		o.PwdExpirationConfig = &eventActionPasswordExpiration{}
	}
	if o.IDPConfig == nil {
		o.IDPConfig = &eventActionIDPAccountCheck{}
	}
}

func (*eventActionOptions) getTFAttributes() map[string]attr.Type {
	kv := keyValue{}

	return map[string]attr.Type{
		"http_config": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"endpoint": types.StringType,
				"username": types.StringType,
				"password": types.StringType,
				"headers": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"key":   types.StringType,
							"value": types.StringType,
						},
					},
				},
				"timeout":         types.Int64Type,
				"skip_tls_verify": types.BoolType,
				"method":          types.StringType,
				"query_parameters": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: kv.getTFAttributes(),
					},
				},
				"body": types.StringType,
				"parts": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"name":     types.StringType,
							"filepath": types.StringType,
							"headers": types.ListType{
								ElemType: types.ObjectType{
									AttrTypes: kv.getTFAttributes(),
								},
							},
							"body": types.StringType,
						},
					},
				},
			},
		},
		"cmd_config": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"cmd": types.StringType,
				"args": types.ListType{
					ElemType: types.StringType,
				},
				"timeout": types.Int64Type,
				"env_vars": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: kv.getTFAttributes(),
					},
				},
			},
		},
		"email_config": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"recipients": types.ListType{
					ElemType: types.StringType,
				},
				"bcc": types.ListType{
					ElemType: types.StringType,
				},
				"subject":      types.StringType,
				"content_type": types.Int64Type,
				"body":         types.StringType,
				"attachments": types.ListType{
					ElemType: types.StringType,
				},
			},
		},
		"retention_config": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"folders": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"path":                    types.StringType,
							"retention":               types.Int64Type,
							"delete_empty_dirs":       types.BoolType,
							"ignore_user_permissions": types.BoolType,
						},
					},
				},
			},
		},
		"fs_config": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"type": types.Int64Type,
				"renames": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: kv.getTFAttributes(),
					},
				},
				"mkdirs": types.ListType{
					ElemType: types.StringType,
				},
				"deletes": types.ListType{
					ElemType: types.StringType,
				},
				"exist": types.ListType{
					ElemType: types.StringType,
				},
				"copy": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: kv.getTFAttributes(),
					},
				},
				"compress": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name": types.StringType,
						"paths": types.ListType{
							ElemType: types.StringType,
						},
					},
				},
			},
		},
		"pwd_expiration_config": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"threshold": types.Int64Type,
			},
		},
		"idp_config": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"mode":           types.Int64Type,
				"template_user":  types.StringType,
				"template_admin": types.StringType,
			},
		},
	}
}

func (o *eventActionOptions) toSFTPGo(ctx context.Context) (client.EventActionOptions, diag.Diagnostics) {
	o.ensureNotNull()
	options := client.EventActionOptions{
		HTTPConfig: client.EventActionHTTPConfig{
			Endpoint:      o.HTTPConfig.Endpoint.ValueString(),
			Username:      o.HTTPConfig.Username.ValueString(),
			Password:      getSFTPGoSecret(o.HTTPConfig.Password.ValueString()),
			Timeout:       int(o.HTTPConfig.Timeout.ValueInt64()),
			SkipTLSVerify: o.HTTPConfig.SkipTLSVerify.ValueBool(),
			Method:        o.HTTPConfig.Method.ValueString(),
			Body:          o.HTTPConfig.Body.ValueString(),
		},
		CmdConfig: client.EventActionCommandConfig{
			Cmd:     o.CmdConfig.Cmd.ValueString(),
			Timeout: int(o.CmdConfig.Timeout.ValueInt64()),
		},
		EmailConfig: client.EventActionEmailConfig{
			Subject:     o.EmailConfig.Subject.ValueString(),
			Body:        o.EmailConfig.Body.ValueString(),
			ContentType: int(o.EmailConfig.ContentType.ValueInt64()),
		},
		FsConfig: client.EventActionFilesystemConfig{
			Type: int(o.FsConfig.Type.ValueInt64()),
			Compress: client.EventActionFsCompress{
				Name: o.FsConfig.Compress.Name.ValueString(),
			},
		},
		PwdExpirationConfig: client.EventActionPasswordExpiration{
			Threshold: int(o.PwdExpirationConfig.Threshold.ValueInt64()),
		},
		IDPConfig: client.EventActionIDPAccountCheck{
			Mode:          int(o.IDPConfig.Mode.ValueInt64()),
			TemplateUser:  o.IDPConfig.TemplateUser.ValueString(),
			TemplateAdmin: o.IDPConfig.TemplateAdmin.ValueString(),
		},
	}

	for _, h := range o.HTTPConfig.Headers {
		options.HTTPConfig.Headers = append(options.HTTPConfig.Headers, client.KeyValue{
			Key:   h.Key.ValueString(),
			Value: h.Value.ValueString(),
		})
	}
	for _, q := range o.HTTPConfig.QueryParameters {
		options.HTTPConfig.QueryParameters = append(options.HTTPConfig.QueryParameters, client.KeyValue{
			Key:   q.Key.ValueString(),
			Value: q.Value.ValueString(),
		})
	}
	for _, p := range o.HTTPConfig.Parts {
		var headers []client.KeyValue
		for _, h := range p.Headers {
			headers = append(headers, client.KeyValue{
				Key:   h.Key.ValueString(),
				Value: h.Value.ValueString(),
			})
		}
		options.HTTPConfig.Parts = append(options.HTTPConfig.Parts, client.HTTPPart{
			Name:     p.Name.ValueString(),
			Filepath: p.Filepath.ValueString(),
			Headers:  headers,
			Body:     p.Body.ValueString(),
		})
	}

	if !o.CmdConfig.Args.IsNull() {
		diags := o.CmdConfig.Args.ElementsAs(ctx, &options.CmdConfig.Args, false)
		if diags.HasError() {
			return options, diags
		}
	}
	for _, h := range o.CmdConfig.EnvVars {
		options.CmdConfig.EnvVars = append(options.CmdConfig.EnvVars, client.KeyValue{
			Key:   h.Key.ValueString(),
			Value: h.Value.ValueString(),
		})
	}

	if !o.EmailConfig.Recipients.IsNull() {
		diags := o.EmailConfig.Recipients.ElementsAs(ctx, &options.EmailConfig.Recipients, false)
		if diags.HasError() {
			return options, diags
		}
	}
	if !o.EmailConfig.Bcc.IsNull() {
		diags := o.EmailConfig.Bcc.ElementsAs(ctx, &options.EmailConfig.Bcc, false)
		if diags.HasError() {
			return options, diags
		}
	}
	if !o.EmailConfig.Attachments.IsNull() {
		diags := o.EmailConfig.Attachments.ElementsAs(ctx, &options.EmailConfig.Attachments, false)
		if diags.HasError() {
			return options, diags
		}
	}

	for _, folder := range o.RetentionConfig.Folders {
		options.RetentionConfig.Folders = append(options.RetentionConfig.Folders, client.FolderRetention{
			Path:                  folder.Path.ValueString(),
			Retention:             int(folder.Retention.ValueInt64()),
			DeleteEmptyDirs:       folder.DeleteEmptyDirs.ValueBool(),
			IgnoreUserPermissions: folder.IgnoreUserPermissions.ValueBool(),
		})
	}

	for _, v := range o.FsConfig.Renames {
		options.FsConfig.Renames = append(options.FsConfig.Renames, client.KeyValue{
			Key:   v.Key.ValueString(),
			Value: v.Value.ValueString(),
		})
	}
	for _, v := range o.FsConfig.Copy {
		options.FsConfig.Copy = append(options.FsConfig.Copy, client.KeyValue{
			Key:   v.Key.ValueString(),
			Value: v.Value.ValueString(),
		})
	}
	if !o.FsConfig.MkDirs.IsNull() {
		diags := o.FsConfig.MkDirs.ElementsAs(ctx, &options.FsConfig.MkDirs, false)
		if diags.HasError() {
			return options, diags
		}
	}
	if !o.FsConfig.Deletes.IsNull() {
		diags := o.FsConfig.Deletes.ElementsAs(ctx, &options.FsConfig.Deletes, false)
		if diags.HasError() {
			return options, diags
		}
	}
	if !o.FsConfig.Exist.IsNull() {
		diags := o.FsConfig.Exist.ElementsAs(ctx, &options.FsConfig.Exist, false)
		if diags.HasError() {
			return options, diags
		}
	}
	if !o.FsConfig.Compress.Paths.IsNull() {
		diags := o.FsConfig.Compress.Paths.ElementsAs(ctx, &options.FsConfig.Compress.Paths, false)
		if diags.HasError() {
			return options, diags
		}
	}

	return options, nil
}

func (o *eventActionOptions) fromSFTPGo(ctx context.Context, action *client.BaseEventAction) diag.Diagnostics {
	o.HTTPConfig = nil
	o.CmdConfig = nil
	o.EmailConfig = nil
	o.RetentionConfig = nil
	o.FsConfig = nil
	o.PwdExpirationConfig = nil
	o.IDPConfig = nil

	switch action.Type {
	case client.ActionTypeHTTP:
		o.HTTPConfig = &eventActionHTTPConfig{
			Endpoint:      getOptionalString(action.Options.HTTPConfig.Endpoint),
			Username:      getOptionalString(action.Options.HTTPConfig.Username),
			Password:      getOptionalString(getSecretFromSFTPGo(action.Options.HTTPConfig.Password)),
			Timeout:       types.Int64Value(int64(action.Options.HTTPConfig.Timeout)),
			SkipTLSVerify: getOptionalBool(action.Options.HTTPConfig.SkipTLSVerify),
			Method:        getOptionalString(action.Options.HTTPConfig.Method),
			Body:          getOptionalString(action.Options.HTTPConfig.Body),
		}
		for _, h := range action.Options.HTTPConfig.Headers {
			o.HTTPConfig.Headers = append(o.HTTPConfig.Headers, keyValue{
				Key:   types.StringValue(h.Key),
				Value: types.StringValue(h.Value),
			})
		}
		for _, q := range action.Options.HTTPConfig.QueryParameters {
			o.HTTPConfig.QueryParameters = append(o.HTTPConfig.QueryParameters, keyValue{
				Key:   types.StringValue(q.Key),
				Value: types.StringValue(q.Value),
			})
		}
		for _, p := range action.Options.HTTPConfig.Parts {
			var headers []keyValue
			for _, h := range p.Headers {
				headers = append(headers, keyValue{
					Key:   types.StringValue(h.Key),
					Value: types.StringValue(h.Value),
				})
			}
			o.HTTPConfig.Parts = append(o.HTTPConfig.Parts, httpPart{
				Name:     types.StringValue(p.Name),
				Headers:  headers,
				Filepath: getOptionalString(p.Filepath),
				Body:     getOptionalString(p.Body),
			})
		}
	case client.ActionTypeCommand:
		o.CmdConfig = &eventActionCommandConfig{
			Cmd:     types.StringValue(action.Options.CmdConfig.Cmd),
			Timeout: types.Int64Value(int64(action.Options.CmdConfig.Timeout)),
		}
		args, diags := types.ListValueFrom(ctx, types.StringType, action.Options.CmdConfig.Args)
		if diags.HasError() {
			return diags
		}
		o.CmdConfig.Args = args
		for _, e := range action.Options.CmdConfig.EnvVars {
			o.CmdConfig.EnvVars = append(o.CmdConfig.EnvVars, keyValue{
				Key:   types.StringValue(e.Key),
				Value: types.StringValue(e.Value),
			})
		}
	case client.ActionTypeEmail:
		o.EmailConfig = &eventActionEmailConfig{
			Subject:     types.StringValue(action.Options.EmailConfig.Subject),
			Body:        types.StringValue(action.Options.EmailConfig.Body),
			ContentType: getOptionalInt64(int64(action.Options.EmailConfig.ContentType)),
		}
		recipients, diags := types.ListValueFrom(ctx, types.StringType, action.Options.EmailConfig.Recipients)
		if diags.HasError() {
			return diags
		}
		o.EmailConfig.Recipients = recipients
		bcc, diags := types.ListValueFrom(ctx, types.StringType, action.Options.EmailConfig.Bcc)
		if diags.HasError() {
			return diags
		}
		o.EmailConfig.Bcc = bcc
		attachments, diags := types.ListValueFrom(ctx, types.StringType, action.Options.EmailConfig.Attachments)
		if diags.HasError() {
			return diags
		}
		o.EmailConfig.Attachments = attachments
	case client.ActionTypeDataRetentionCheck:
		o.RetentionConfig = &eventActionDataRetentionConfig{}
		for _, f := range action.Options.RetentionConfig.Folders {
			o.RetentionConfig.Folders = append(o.RetentionConfig.Folders, folderRetention{
				Path:                  types.StringValue(f.Path),
				Retention:             types.Int64Value(int64(f.Retention)),
				DeleteEmptyDirs:       getOptionalBool(f.DeleteEmptyDirs),
				IgnoreUserPermissions: getOptionalBool(f.IgnoreUserPermissions),
			})
		}
	case client.ActionTypeFilesystem:
		o.FsConfig = &eventActionFilesystemConfig{
			Type:    types.Int64Value(int64(action.Options.FsConfig.Type)),
			MkDirs:  types.ListNull(types.StringType),
			Deletes: types.ListNull(types.StringType),
			Exist:   types.ListNull(types.StringType),
		}

		switch action.Options.FsConfig.Type {
		case client.FilesystemActionRename:
			for _, v := range action.Options.FsConfig.Renames {
				o.FsConfig.Renames = append(o.FsConfig.Renames, keyValue{
					Key:   types.StringValue(v.Key),
					Value: types.StringValue(v.Value),
				})
			}
		case client.FilesystemActionDelete:
			deletes, diags := types.ListValueFrom(ctx, types.StringType, action.Options.FsConfig.Deletes)
			if diags.HasError() {
				return diags
			}
			o.FsConfig.Deletes = deletes
		case client.FilesystemActionMkdirs:
			mkdirs, diags := types.ListValueFrom(ctx, types.StringType, action.Options.FsConfig.MkDirs)
			if diags.HasError() {
				return diags
			}
			o.FsConfig.MkDirs = mkdirs
		case client.FilesystemActionExist:
			exist, diags := types.ListValueFrom(ctx, types.StringType, action.Options.FsConfig.Exist)
			if diags.HasError() {
				return diags
			}
			o.FsConfig.Exist = exist
		case client.FilesystemActionCompress:
			o.FsConfig.Compress = &eventActionFsCompress{
				Name: types.StringValue(action.Options.FsConfig.Compress.Name),
			}
			paths, diags := types.ListValueFrom(ctx, types.StringType, action.Options.FsConfig.Compress.Paths)
			if diags.HasError() {
				return diags
			}
			o.FsConfig.Compress.Paths = paths
		case client.FilesystemActionCopy:
			for _, v := range action.Options.FsConfig.Copy {
				o.FsConfig.Copy = append(o.FsConfig.Copy, keyValue{
					Key:   types.StringValue(v.Key),
					Value: types.StringValue(v.Value),
				})
			}
		}
	case client.ActionTypePasswordExpirationCheck:
		o.PwdExpirationConfig = &eventActionPasswordExpiration{
			Threshold: types.Int64Value(int64(action.Options.PwdExpirationConfig.Threshold)),
		}
	case client.ActionTypeIDPAccountCheck:
		o.IDPConfig = &eventActionIDPAccountCheck{
			Mode:          types.Int64Value(int64(action.Options.IDPConfig.Mode)),
			TemplateUser:  getOptionalString(action.Options.IDPConfig.TemplateUser),
			TemplateAdmin: getOptionalString(action.Options.IDPConfig.TemplateAdmin),
		}
	}

	return nil
}

type eventActionResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.Int64  `tfsdk:"type"`
	Options     types.Object `tfsdk:"options"` // eventActionOptions
}

func (a *eventActionResourceModel) toSFTPGo(ctx context.Context) (*client.BaseEventAction, diag.Diagnostics) {
	action := &client.BaseEventAction{
		Name:        a.Name.ValueString(),
		Description: a.Description.ValueString(),
		Type:        int(a.Type.ValueInt64()),
	}
	var options eventActionOptions
	diags := a.Options.As(ctx, &options, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return action, diags
	}
	sftpgoOptions, diags := options.toSFTPGo(ctx)
	if diags.HasError() {
		return action, diags
	}
	action.Options = sftpgoOptions

	return action, nil
}

func (a *eventActionResourceModel) fromSFTPGo(ctx context.Context, action *client.BaseEventAction) diag.Diagnostics {
	a.Name = types.StringValue(action.Name)
	a.ID = a.Name
	a.Type = types.Int64Value(int64(action.Type))
	a.Description = getOptionalString(action.Description)
	var opts eventActionOptions
	diags := opts.fromSFTPGo(ctx, action)
	if diags.HasError() {
		return diags
	}
	options, diags := types.ObjectValueFrom(ctx, opts.getTFAttributes(), opts)
	if diags.HasError() {
		return diags
	}
	a.Options = options
	return nil
}

type ruleSchedule struct {
	Hours      types.String `tfsdk:"hour"`
	DayOfWeek  types.String `tfsdk:"day_of_week"`
	DayOfMonth types.String `tfsdk:"day_of_month"`
	Month      types.String `tfsdk:"month"`
}

type ruleConditionPattern struct {
	Pattern      types.String `tfsdk:"pattern"`
	InverseMatch types.Bool   `tfsdk:"inverse_match"`
}

func (*ruleConditionPattern) getTFAttributes() map[string]attr.Type {
	return map[string]attr.Type{
		"pattern":       types.StringType,
		"inverse_match": types.BoolType,
	}
}

type ruleConditionOptions struct {
	Names               []ruleConditionPattern `tfsdk:"names"`
	GroupNames          []ruleConditionPattern `tfsdk:"group_names"`
	RoleNames           []ruleConditionPattern `tfsdk:"role_names"`
	FsPaths             []ruleConditionPattern `tfsdk:"fs_paths"`
	Protocols           types.List             `tfsdk:"protocols"`
	ProviderObjects     types.List             `tfsdk:"provider_objects"`
	MinFileSize         types.Int64            `tfsdk:"min_size"`
	MaxFileSize         types.Int64            `tfsdk:"max_size"`
	ConcurrentExecution types.Bool             `tfsdk:"concurrent_execution"`
}

type ruleConditions struct {
	FsEvents       types.List            `tfsdk:"fs_events"`
	ProviderEvents types.List            `tfsdk:"provider_events"`
	Schedules      []ruleSchedule        `tfsdk:"schedules"`
	IDPLoginEvent  types.Int64           `tfsdk:"idp_login_event"`
	Options        *ruleConditionOptions `tfsdk:"options"`
}

func (*ruleConditions) getTFAttributes() map[string]attr.Type {
	p := ruleConditionPattern{}

	return map[string]attr.Type{
		"fs_events": types.ListType{
			ElemType: types.StringType,
		},
		"provider_events": types.ListType{
			ElemType: types.StringType,
		},
		"schedules": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"hour":         types.StringType,
					"day_of_week":  types.StringType,
					"day_of_month": types.StringType,
					"month":        types.StringType,
				},
			},
		},
		"idp_login_event": types.Int64Type,
		"options": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"names": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: p.getTFAttributes(),
					},
				},
				"group_names": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: p.getTFAttributes(),
					},
				},
				"role_names": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: p.getTFAttributes(),
					},
				},
				"fs_paths": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: p.getTFAttributes(),
					},
				},
				"protocols": types.ListType{
					ElemType: types.StringType,
				},
				"provider_objects": types.ListType{
					ElemType: types.StringType,
				},
				"min_size":             types.Int64Type,
				"max_size":             types.Int64Type,
				"concurrent_execution": types.BoolType,
			},
		},
	}
}

func (c *ruleConditions) toSFTPGo(ctx context.Context) (client.EventRuleConditions, diag.Diagnostics) {
	conditions := client.EventRuleConditions{
		IDPLoginEvent: int(c.IDPLoginEvent.ValueInt64()),
	}
	if !c.FsEvents.IsNull() {
		diags := c.FsEvents.ElementsAs(ctx, &conditions.FsEvents, false)
		if diags.HasError() {
			return conditions, diags
		}
	}
	if !c.ProviderEvents.IsNull() {
		diags := c.FsEvents.ElementsAs(ctx, &conditions.ProviderEvents, false)
		if diags.HasError() {
			return conditions, diags
		}
	}
	for _, schedule := range c.Schedules {
		conditions.Schedules = append(conditions.Schedules, client.Schedule{
			Hours:      schedule.Hours.ValueString(),
			DayOfWeek:  schedule.DayOfWeek.ValueString(),
			DayOfMonth: schedule.DayOfMonth.ValueString(),
			Month:      schedule.Month.ValueString(),
		})
	}
	if c.Options != nil {
		conditions.Options = client.ConditionOptions{
			MinFileSize:         c.Options.MinFileSize.ValueInt64(),
			MaxFileSize:         c.Options.MaxFileSize.ValueInt64(),
			ConcurrentExecution: c.Options.ConcurrentExecution.ValueBool(),
		}
		for _, val := range c.Options.Names {
			conditions.Options.Names = append(conditions.Options.Names, client.ConditionPattern{
				Pattern:      val.Pattern.ValueString(),
				InverseMatch: val.InverseMatch.ValueBool(),
			})
		}
		for _, val := range c.Options.GroupNames {
			conditions.Options.GroupNames = append(conditions.Options.GroupNames, client.ConditionPattern{
				Pattern:      val.Pattern.ValueString(),
				InverseMatch: val.InverseMatch.ValueBool(),
			})
		}
		for _, val := range c.Options.RoleNames {
			conditions.Options.RoleNames = append(conditions.Options.RoleNames, client.ConditionPattern{
				Pattern:      val.Pattern.ValueString(),
				InverseMatch: val.InverseMatch.ValueBool(),
			})
		}
		for _, val := range c.Options.FsPaths {
			conditions.Options.FsPaths = append(conditions.Options.FsPaths, client.ConditionPattern{
				Pattern:      val.Pattern.ValueString(),
				InverseMatch: val.InverseMatch.ValueBool(),
			})
		}
		if !c.Options.Protocols.IsNull() {
			diags := c.Options.Protocols.ElementsAs(ctx, &conditions.Options.Protocols, false)
			if diags.HasError() {
				return conditions, diags
			}
		}
		if !c.Options.ProviderObjects.IsNull() {
			diags := c.Options.ProviderObjects.ElementsAs(ctx, &conditions.Options.ProviderObjects, false)
			if diags.HasError() {
				return conditions, diags
			}
		}
	}

	return conditions, nil
}

func (c *ruleConditions) fromSFTPGo(ctx context.Context, conditions *client.EventRuleConditions, trigger int) diag.Diagnostics {
	fsEvents, diags := types.ListValueFrom(ctx, types.StringType, conditions.FsEvents)
	if diags.HasError() {
		return diags
	}
	c.FsEvents = fsEvents

	providerEvents, diags := types.ListValueFrom(ctx, types.StringType, conditions.ProviderEvents)
	if diags.HasError() {
		return diags
	}
	c.ProviderEvents = providerEvents

	for _, schedule := range conditions.Schedules {
		c.Schedules = append(c.Schedules, ruleSchedule{
			Hours:      types.StringValue(schedule.Hours),
			DayOfWeek:  types.StringValue(schedule.DayOfWeek),
			DayOfMonth: types.StringValue(schedule.DayOfMonth),
			Month:      types.StringValue(schedule.Month),
		})
	}
	if trigger == 7 {
		c.IDPLoginEvent = types.Int64Value(int64(conditions.IDPLoginEvent))
	} else {
		c.IDPLoginEvent = getOptionalInt64(int64(conditions.IDPLoginEvent))
	}

	c.Options = &ruleConditionOptions{
		MinFileSize:         getOptionalInt64(conditions.Options.MinFileSize),
		MaxFileSize:         getOptionalInt64(conditions.Options.MaxFileSize),
		ConcurrentExecution: getOptionalBool(conditions.Options.ConcurrentExecution),
	}
	for _, val := range conditions.Options.Names {
		c.Options.Names = append(c.Options.Names, ruleConditionPattern{
			Pattern:      types.StringValue(val.Pattern),
			InverseMatch: getOptionalBool(val.InverseMatch),
		})
	}
	for _, val := range conditions.Options.GroupNames {
		c.Options.GroupNames = append(c.Options.GroupNames, ruleConditionPattern{
			Pattern:      types.StringValue(val.Pattern),
			InverseMatch: getOptionalBool(val.InverseMatch),
		})
	}
	for _, val := range conditions.Options.RoleNames {
		c.Options.RoleNames = append(c.Options.RoleNames, ruleConditionPattern{
			Pattern:      types.StringValue(val.Pattern),
			InverseMatch: getOptionalBool(val.InverseMatch),
		})
	}
	for _, val := range conditions.Options.FsPaths {
		c.Options.FsPaths = append(c.Options.FsPaths, ruleConditionPattern{
			Pattern:      types.StringValue(val.Pattern),
			InverseMatch: getOptionalBool(val.InverseMatch),
		})
	}
	protocols, diags := types.ListValueFrom(ctx, types.StringType, conditions.Options.Protocols)
	if diags.HasError() {
		return diags
	}
	c.Options.Protocols = protocols
	providerObjects, diags := types.ListValueFrom(ctx, types.StringType, conditions.Options.ProviderObjects)
	if diags.HasError() {
		return diags
	}
	c.Options.ProviderObjects = providerObjects

	return nil
}

type ruleAction struct {
	Name            types.String `tfsdk:"name"`
	IsFailureAction types.Bool   `tfsdk:"is_failure_action"`
	StopOnFailure   types.Bool   `tfsdk:"stop_on_failure"`
	ExecuteSync     types.Bool   `tfsdk:"execute_sync"`
}

type eventRuleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Status      types.Int64  `tfsdk:"status"`
	Description types.String `tfsdk:"description"`
	Trigger     types.Int64  `tfsdk:"trigger"`
	Conditions  types.Object `tfsdk:"conditions"` // ruleConditions
	Actions     []ruleAction `tfsdk:"actions"`
	CreatedAt   types.Int64  `tfsdk:"created_at"`
	UpdatedAt   types.Int64  `tfsdk:"updated_at"`
}

func (r *eventRuleResourceModel) toSFTPGo(ctx context.Context) (*client.EventRule, diag.Diagnostics) {
	rule := &client.EventRule{
		Name:        r.Name.ValueString(),
		Status:      int(r.Status.ValueInt64()),
		Description: r.Description.ValueString(),
		Trigger:     int(r.Trigger.ValueInt64()),
	}

	for idx, action := range r.Actions {
		rule.Actions = append(rule.Actions, client.EventAction{
			Name:  action.Name.ValueString(),
			Order: idx + 1,
			Options: client.EventActionRelationOptions{
				IsFailureAction: action.IsFailureAction.ValueBool(),
				StopOnFailure:   action.StopOnFailure.ValueBool(),
				ExecuteSync:     action.ExecuteSync.ValueBool(),
			},
		})
	}

	var conditions ruleConditions
	diags := r.Conditions.As(ctx, &conditions, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return rule, diags
	}
	sftpgoConditions, diags := conditions.toSFTPGo(ctx)
	if diags.HasError() {
		return rule, diags
	}
	rule.Conditions = sftpgoConditions

	return rule, nil
}

func (r *eventRuleResourceModel) fromSFTPGo(ctx context.Context, rule *client.EventRule) diag.Diagnostics {
	r.Name = types.StringValue(rule.Name)
	r.ID = r.Name
	r.Status = types.Int64Value(int64(rule.Status))
	r.Description = getOptionalString(rule.Description)
	r.Trigger = types.Int64Value(int64(rule.Trigger))
	r.CreatedAt = types.Int64Value(rule.CreatedAt)
	r.UpdatedAt = types.Int64Value(rule.UpdatedAt)

	r.Actions = nil
	for _, action := range rule.Actions {
		r.Actions = append(r.Actions, ruleAction{
			Name:            types.StringValue(action.Name),
			IsFailureAction: getOptionalBool(action.Options.IsFailureAction),
			StopOnFailure:   getOptionalBool(action.Options.StopOnFailure),
			ExecuteSync:     getOptionalBool(action.Options.ExecuteSync),
		})
	}

	var c ruleConditions
	diags := c.fromSFTPGo(ctx, &rule.Conditions, rule.Trigger)
	if diags.HasError() {
		return diags
	}
	conditions, diags := types.ObjectValueFrom(ctx, c.getTFAttributes(), c)
	if diags.HasError() {
		return diags
	}
	r.Conditions = conditions

	return nil
}

func getOptionalInt64(val int64) types.Int64 {
	if val == 0 {
		return types.Int64Null()
	}
	return types.Int64Value(val)
}

func getOptionalString(val string) types.String {
	if val == "" {
		return types.StringNull()
	}
	return types.StringValue(val)
}

func getOptionalBool(val bool) types.Bool {
	if !val {
		return types.BoolNull()
	}
	return types.BoolValue(val)
}

var supportedSecretStatues = []string{kms.SecretStatusSecretBox, kms.SecretStatusAES256GCM, kms.SecretStatusGCP,
	kms.SecretStatusAWS, kms.SecretStatusVaultTransit, kms.SecretStatusAzureKeyVault}

func getSFTPGoSecret(val string) kms.BaseSecret {
	if val == "" {
		return kms.BaseSecret{}
	}
	parts := strings.SplitN(val, "$", 5)
	if len(parts) == 5 && parts[0] == "" && contains(supportedSecretStatues, parts[1]) {
		additionalDataLen, err := strconv.Atoi(parts[3])
		if err == nil && len(parts[4]) > additionalDataLen {
			return kms.BaseSecret{
				Status:         parts[1],
				Payload:        parts[4][additionalDataLen:],
				Key:            parts[2],
				AdditionalData: parts[4][:additionalDataLen],
			}
		}
	}
	return kms.BaseSecret{
		Status:  kms.SecretStatusPlain,
		Payload: val,
	}
}

func getSecretFromSFTPGo(secret kms.BaseSecret) string {
	if secret.Status == "" {
		return ""
	}
	if secret.Status == kms.SecretStatusPlain {
		return secret.Payload
	}
	return fmt.Sprintf("$%s$%s$%d$%s%s", secret.Status, secret.Key, len(secret.AdditionalData),
		secret.AdditionalData, secret.Payload)
}
