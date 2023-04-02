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

func (u *userResourceModel) toSFTPGo(ctx context.Context) (*sdk.User, diag.Diagnostics) {
	user := &sdk.User{
		BaseUser: sdk.BaseUser{
			Username:             u.Username.ValueString(),
			Status:               int(u.Status.ValueInt64()),
			Email:                u.Email.ValueString(),
			ExpirationDate:       u.ExpirationDate.ValueInt64(),
			Password:             u.Password.ValueString(),
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
			Role:                 u.AdditionalInfo.ValueString(),
		},
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

func (u *userResourceModel) fromSFTPGo(ctx context.Context, user *sdk.User) diag.Diagnostics {
	u.Username = types.StringValue(user.Username)
	u.Status = types.Int64Value(int64(user.Status))
	u.Email = getOptionalString(user.Email)
	u.ExpirationDate = getOptionalInt64(user.ExpirationDate)
	u.Password = types.StringValue(user.Password)
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
	var groups []userGroupMapping
	for _, g := range user.Groups {
		groups = append(groups, userGroupMapping{
			Name: types.StringValue(g.Name),
			Type: types.Int64Value(int64(g.Type)),
		})
	}
	u.Groups = groups

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

type hooksFilter struct {
	ExternalAuthDisabled  types.Bool `tfsdk:"external_auth_disabled"`
	PreLoginDisabled      types.Bool `tfsdk:"pre_login_disabled"`
	CheckPasswordDisabled types.Bool `tfsdk:"check_password_disabled"`
}

type bandwidthLimit struct {
	Sources           types.List  `tfsdk:"sources"`
	UploadBandwidth   types.Int64 `tfsdk:"upload_bandwidth"`
	DownloadBandwidth types.Int64 `tfsdk:"download_bandwidth"`
}

type dataTransferLimit struct {
	Sources              types.List  `tfsdk:"sources"`
	UploadDataTransfer   types.Int64 `tfsdk:"upload_data_transfer"`
	DownloadDataTransfer types.Int64 `tfsdk:"download_data_transfer"`
	TotalDataTransfer    types.Int64 `tfsdk:"total_data_transfer"`
}

type baseSecret struct {
	Status  types.String `tfsdk:"status"`
	Payload types.String `tfsdk:"payload"`
}

func (s *baseSecret) getTFObject() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"status":  types.StringType,
			"payload": types.StringType,
		},
	}
}

func (s *baseSecret) toSFTPGo() kms.BaseSecret {
	return kms.BaseSecret{
		Status:  s.Status.ValueString(),
		Payload: s.Payload.ValueString(),
	}
}

func (s *baseSecret) fromSFTPGo(secret *kms.BaseSecret) {
	s.Status = getOptionalString(secret.Status)
	s.Payload = getOptionalString(secret.Payload)
}

type baseUserFilters struct {
	AllowedIP               types.List          `tfsdk:"allowed_ip"`
	DeniedIP                types.List          `tfsdk:"denied_ip"`
	DeniedLoginMethods      types.List          `tfsdk:"denied_login_methods"`
	DeniedProtocols         types.List          `tfsdk:"denied_protocols"`
	FilePatterns            []patternsFilter    `tfsdk:"file_patterns"`
	MaxUploadFileSize       types.Int64         `tfsdk:"max_upload_file_size"`
	TLSUsername             types.String        `tfsdk:"tls_username"`
	Hooks                   hooksFilter         `tfsdk:"hooks"`
	DisableFsChecks         types.Bool          `tfsdk:"disable_fs_checks"`
	WebClient               types.List          `tfsdk:"web_client"`
	AllowAPIKeyAuth         types.Bool          `tfsdk:"allow_api_key_auth"`
	UserType                types.String        `tfsdk:"user_type"`
	BandwidthLimits         []bandwidthLimit    `tfsdk:"bandwidth_limits"`
	DataTransferLimits      []dataTransferLimit `tfsdk:"data_transfer_limits"`
	ExternalAuthCacheTime   types.Int64         `tfsdk:"external_auth_cache_time"`
	StartDirectory          types.String        `tfsdk:"start_directory"`
	TwoFactorAuthProtocols  types.List          `tfsdk:"two_factor_protocols"`
	FTPSecurity             types.Int64         `tfsdk:"ftp_security"`
	IsAnonymous             types.Bool          `tfsdk:"is_anonymous"`
	DefaultSharesExpiration types.Int64         `tfsdk:"default_shares_expiration"`
	PasswordExpiration      types.Int64         `tfsdk:"password_expiration"`
	PasswordStrength        types.Int64         `tfsdk:"password_strength"`
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
		"max_upload_file_size": types.Int64Type,
		"tls_username":         types.StringType,
		"hooks": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"external_auth_disabled":  types.BoolType,
				"pre_login_disabled":      types.BoolType,
				"check_password_disabled": types.BoolType,
			},
		},
		"disable_fs_checks": types.BoolType,
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
		"data_transfer_limits": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"sources": types.ListType{
						ElemType: types.StringType,
					},
					"upload_data_transfer":   types.Int64Type,
					"download_data_transfer": types.Int64Type,
					"total_data_transfer":    types.Int64Type,
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
		"password_expiration":       types.Int64Type,
		"password_strength":         types.Int64Type,
	}
}

func (f *baseUserFilters) toSFTPGo(ctx context.Context) (sdk.BaseUserFilters, diag.Diagnostics) {
	filters := sdk.BaseUserFilters{
		MaxUploadFileSize: f.MaxUploadFileSize.ValueInt64(),
		TLSUsername:       sdk.TLSUsername(f.TLSUsername.ValueString()),
		Hooks: sdk.HooksFilter{
			ExternalAuthDisabled:  f.Hooks.ExternalAuthDisabled.ValueBool(),
			PreLoginDisabled:      f.Hooks.PreLoginDisabled.ValueBool(),
			CheckPasswordDisabled: f.Hooks.CheckPasswordDisabled.ValueBool(),
		},
		DisableFsChecks:         f.DisableFsChecks.ValueBool(),
		AllowAPIKeyAuth:         f.AllowAPIKeyAuth.ValueBool(),
		UserType:                f.UserType.ValueString(),
		ExternalAuthCacheTime:   f.ExternalAuthCacheTime.ValueInt64(),
		StartDirectory:          f.StartDirectory.ValueString(),
		FTPSecurity:             int(f.FTPSecurity.ValueInt64()),
		IsAnonymous:             f.IsAnonymous.ValueBool(),
		DefaultSharesExpiration: int(f.DefaultSharesExpiration.ValueInt64()),
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
	for _, l := range f.DataTransferLimits {
		limits := sdk.DataTransferLimit{
			UploadDataTransfer:   l.UploadDataTransfer.ValueInt64(),
			DownloadDataTransfer: l.DownloadDataTransfer.ValueInt64(),
			TotalDataTransfer:    l.TotalDataTransfer.ValueInt64(),
		}
		if !l.Sources.IsNull() {
			diags := l.Sources.ElementsAs(ctx, &limits.Sources, false)
			if diags.HasError() {
				return filters, diags
			}
		}
		filters.DataTransferLimits = append(filters.DataTransferLimits, limits)
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
	f.Hooks = hooksFilter{
		ExternalAuthDisabled:  getOptionalBool(filters.Hooks.ExternalAuthDisabled),
		PreLoginDisabled:      getOptionalBool(filters.Hooks.PreLoginDisabled),
		CheckPasswordDisabled: getOptionalBool(filters.Hooks.CheckPasswordDisabled),
	}
	f.DisableFsChecks = getOptionalBool(filters.DisableFsChecks)
	webClient, diags := types.ListValueFrom(ctx, types.StringType, filters.WebClient)
	if diags.HasError() {
		return diags
	}
	f.WebClient = webClient
	f.AllowAPIKeyAuth = getOptionalBool(filters.AllowAPIKeyAuth)
	f.UserType = getOptionalString(filters.UserType)
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
	for _, limit := range filters.DataTransferLimits {
		sources, diags := types.ListValueFrom(ctx, types.StringType, limit.Sources)
		if diags.HasError() {
			return diags
		}
		f.DataTransferLimits = append(f.DataTransferLimits, dataTransferLimit{
			Sources:              sources,
			UploadDataTransfer:   types.Int64Value(limit.UploadDataTransfer),
			DownloadDataTransfer: types.Int64Value(limit.DownloadDataTransfer),
			TotalDataTransfer:    types.Int64Value(limit.TotalDataTransfer),
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
	f.PasswordExpiration = getOptionalInt64(int64(filters.PasswordExpiration))
	f.PasswordExpiration = getOptionalInt64(int64(filters.PasswordStrength))
	return nil
}

type userFilters struct {
	// embedded structs are not supported
	//baseUserFilters
	AllowedIP               types.List          `tfsdk:"allowed_ip"`
	DeniedIP                types.List          `tfsdk:"denied_ip"`
	DeniedLoginMethods      types.List          `tfsdk:"denied_login_methods"`
	DeniedProtocols         types.List          `tfsdk:"denied_protocols"`
	FilePatterns            []patternsFilter    `tfsdk:"file_patterns"`
	MaxUploadFileSize       types.Int64         `tfsdk:"max_upload_file_size"`
	TLSUsername             types.String        `tfsdk:"tls_username"`
	Hooks                   hooksFilter         `tfsdk:"hooks"`
	DisableFsChecks         types.Bool          `tfsdk:"disable_fs_checks"`
	WebClient               types.List          `tfsdk:"web_client"`
	AllowAPIKeyAuth         types.Bool          `tfsdk:"allow_api_key_auth"`
	UserType                types.String        `tfsdk:"user_type"`
	BandwidthLimits         []bandwidthLimit    `tfsdk:"bandwidth_limits"`
	DataTransferLimits      []dataTransferLimit `tfsdk:"data_transfer_limits"`
	ExternalAuthCacheTime   types.Int64         `tfsdk:"external_auth_cache_time"`
	StartDirectory          types.String        `tfsdk:"start_directory"`
	TwoFactorAuthProtocols  types.List          `tfsdk:"two_factor_protocols"`
	FTPSecurity             types.Int64         `tfsdk:"ftp_security"`
	IsAnonymous             types.Bool          `tfsdk:"is_anonymous"`
	DefaultSharesExpiration types.Int64         `tfsdk:"default_shares_expiration"`
	PasswordExpiration      types.Int64         `tfsdk:"password_expiration"`
	PasswordStrength        types.Int64         `tfsdk:"password_strength"`
	RequirePasswordChange   types.Bool          `tfsdk:"require_password_change"`
}

func (f *userFilters) getTFAttributes() map[string]attr.Type {
	baseFilters := baseUserFilters{}
	base := baseFilters.getTFAttributes()

	filters := map[string]attr.Type{
		"require_password_change": types.BoolType,
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
		Hooks:                   f.Hooks,
		DisableFsChecks:         f.DisableFsChecks,
		WebClient:               f.WebClient,
		AllowAPIKeyAuth:         f.AllowAPIKeyAuth,
		UserType:                f.UserType,
		BandwidthLimits:         f.BandwidthLimits,
		DataTransferLimits:      f.DataTransferLimits,
		ExternalAuthCacheTime:   f.ExternalAuthCacheTime,
		StartDirectory:          f.StartDirectory,
		TwoFactorAuthProtocols:  f.TwoFactorAuthProtocols,
		FTPSecurity:             f.FTPSecurity,
		IsAnonymous:             f.IsAnonymous,
		DefaultSharesExpiration: f.DefaultSharesExpiration,
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
	f.Hooks = filters.Hooks
	f.DisableFsChecks = filters.DisableFsChecks
	f.WebClient = filters.WebClient
	f.AllowAPIKeyAuth = filters.AllowAPIKeyAuth
	f.UserType = filters.UserType
	f.BandwidthLimits = filters.BandwidthLimits
	f.DataTransferLimits = filters.DataTransferLimits
	f.ExternalAuthCacheTime = filters.ExternalAuthCacheTime
	f.StartDirectory = filters.StartDirectory
	f.TwoFactorAuthProtocols = filters.TwoFactorAuthProtocols
	f.FTPSecurity = filters.FTPSecurity
	f.IsAnonymous = filters.IsAnonymous
	f.DefaultSharesExpiration = filters.DefaultSharesExpiration
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
	return nil
}

type s3FsConfig struct {
	Bucket              types.String `tfsdk:"bucket"`
	KeyPrefix           types.String `tfsdk:"key_prefix"`
	Region              types.String `tfsdk:"region"`
	AccessKey           types.String `tfsdk:"access_key"`
	AccessSecret        baseSecret   `tfsdk:"access_secret"`
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
}

type gcsFsConfig struct {
	Bucket               types.String `tfsdk:"bucket"`
	KeyPrefix            types.String `tfsdk:"key_prefix"`
	Credentials          baseSecret   `tfsdk:"credentials"`
	AutomaticCredentials types.Int64  `tfsdk:"automatic_credentials"`
	StorageClass         types.String `tfsdk:"storage_class"`
	ACL                  types.String `tfsdk:"acl"`
	UploadPartSize       types.Int64  `tfsdk:"upload_part_size"`
	UploadPartMaxTime    types.Int64  `tfsdk:"upload_part_max_time"`
}

type azBlobFsConfig struct {
	Container           types.String `tfsdk:"container"`
	AccountName         types.String `tfsdk:"account_name"`
	AccountKey          baseSecret   `tfsdk:"account_key"`
	SASURL              baseSecret   `tfsdk:"sas_url"`
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
	Passphrase baseSecret `tfsdk:"passphrase"`
}

type sftpFsConfig struct {
	Endpoint                types.String `tfsdk:"endpoint"`
	Username                types.String `tfsdk:"username"`
	Password                baseSecret   `tfsdk:"password"`
	PrivateKey              baseSecret   `tfsdk:"private_key"`
	Fingerprints            types.List   `tfsdk:"fingerprints"`
	Prefix                  types.String `tfsdk:"prefix"`
	DisableCouncurrentReads types.Bool   `tfsdk:"disable_concurrent_reads"`
	BufferSize              types.Int64  `tfsdk:"buffer_size"`
	EqualityCheckMode       types.Int64  `tfsdk:"equality_check_mode"`
}

type httpFsConfig struct {
	Endpoint          types.String `tfsdk:"endpoint"`
	Username          types.String `tfsdk:"username"`
	Password          baseSecret   `tfsdk:"password"`
	APIKey            baseSecret   `tfsdk:"api_key"`
	SkipTLSVerify     types.Bool   `tfsdk:"skip_tls_verify"`
	EqualityCheckMode types.Int64  `tfsdk:"equality_check_mode"`
}

type filesystem struct {
	Provider     types.Int64    `tfsdk:"provider"`
	S3Config     s3FsConfig     `tfsdk:"s3config"`
	GCSConfig    gcsFsConfig    `tfsdk:"gcsconfig"`
	AzBlobConfig azBlobFsConfig `tfsdk:"azblobconfig"`
	CryptConfig  cryptFsConfig  `tfsdk:"cryptconfig"`
	SFTPConfig   sftpFsConfig   `tfsdk:"sftpconfig"`
	HTTPConfig   httpFsConfig   `tfsdk:"httpconfig"`
}

func (f *filesystem) getTFAttributes() map[string]attr.Type {
	return map[string]attr.Type{
		"provider": types.Int64Type,
		"s3config": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"bucket":                 types.StringType,
				"key_prefix":             types.StringType,
				"region":                 types.StringType,
				"access_key":             types.StringType,
				"access_secret":          f.S3Config.AccessSecret.getTFObject(),
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
			},
		},
		"gcsconfig": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"bucket":                types.StringType,
				"key_prefix":            types.StringType,
				"credentials":           f.GCSConfig.Credentials.getTFObject(),
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
				"account_key":          f.AzBlobConfig.AccountKey.getTFObject(),
				"sas_url":              f.AzBlobConfig.SASURL.getTFObject(),
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
				"passphrase": f.CryptConfig.Passphrase.getTFObject(),
			},
		},
		"sftpconfig": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"endpoint":    types.StringType,
				"username":    types.StringType,
				"password":    f.SFTPConfig.Password.getTFObject(),
				"private_key": f.SFTPConfig.PrivateKey.getTFObject(),
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
				"password":            f.HTTPConfig.Password.getTFObject(),
				"api_key":             f.HTTPConfig.APIKey.getTFObject(),
				"skip_tls_verify":     types.BoolType,
				"equality_check_mode": types.Int64Type,
			},
		},
	}
}

func (f *filesystem) toSFTPGo(ctx context.Context) (sdk.Filesystem, diag.Diagnostics) {
	fs := sdk.Filesystem{
		Provider: sdk.FilesystemProvider(f.Provider.ValueInt64()),
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
			},
			AccessSecret: f.S3Config.AccessSecret.toSFTPGo(),
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
			Credentials: f.GCSConfig.Credentials.toSFTPGo(),
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
			AccountKey: f.AzBlobConfig.AccountKey.toSFTPGo(),
			SASURL:     f.AzBlobConfig.SASURL.toSFTPGo(),
		},
		CryptConfig: sdk.CryptFsConfig{
			Passphrase: f.CryptConfig.Passphrase.toSFTPGo(),
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
			Password:   f.SFTPConfig.Password.toSFTPGo(),
			PrivateKey: f.SFTPConfig.PrivateKey.toSFTPGo(),
		},
		HTTPConfig: sdk.HTTPFsConfig{
			BaseHTTPFsConfig: sdk.BaseHTTPFsConfig{
				Endpoint:          f.HTTPConfig.Endpoint.ValueString(),
				Username:          f.HTTPConfig.Username.ValueString(),
				SkipTLSVerify:     f.HTTPConfig.SkipTLSVerify.ValueBool(),
				EqualityCheckMode: int(f.HTTPConfig.EqualityCheckMode.ValueInt64()),
			},
			Password: f.HTTPConfig.Password.toSFTPGo(),
			APIKey:   f.HTTPConfig.APIKey.toSFTPGo(),
		},
	}

	if !f.SFTPConfig.Fingerprints.IsNull() {
		diags := f.SFTPConfig.Fingerprints.ElementsAs(ctx, fs.SFTPConfig.Fingerprints, false)
		if diags.HasError() {
			return fs, diags
		}
	}
	return fs, nil
}

func (f *filesystem) fromSFTPGo(ctx context.Context, fs *sdk.Filesystem) diag.Diagnostics {
	f.Provider = types.Int64Value(int64(fs.Provider))
	f.S3Config = s3FsConfig{
		Bucket:              getOptionalString(fs.S3Config.Bucket),
		KeyPrefix:           getOptionalString(fs.S3Config.KeyPrefix),
		Region:              getOptionalString(fs.S3Config.Region),
		AccessKey:           getOptionalString(fs.S3Config.AccessKey),
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
	}
	f.S3Config.AccessSecret.fromSFTPGo(&fs.S3Config.AccessSecret)
	f.GCSConfig = gcsFsConfig{
		Bucket:               getOptionalString(fs.GCSConfig.Bucket),
		KeyPrefix:            getOptionalString(fs.GCSConfig.KeyPrefix),
		AutomaticCredentials: getOptionalInt64(int64(fs.GCSConfig.AutomaticCredentials)),
		StorageClass:         getOptionalString(fs.GCSConfig.StorageClass),
		ACL:                  getOptionalString(fs.GCSConfig.ACL),
		UploadPartSize:       getOptionalInt64(fs.GCSConfig.UploadPartSize),
		UploadPartMaxTime:    getOptionalInt64(int64(fs.GCSConfig.UploadPartMaxTime)),
	}
	f.GCSConfig.Credentials.fromSFTPGo(&fs.GCSConfig.Credentials)
	f.AzBlobConfig = azBlobFsConfig{
		Container:           getOptionalString(fs.AzBlobConfig.Container),
		AccountName:         getOptionalString(fs.AzBlobConfig.AccountName),
		Endpoint:            getOptionalString(fs.AzBlobConfig.Endpoint),
		KeyPrefix:           getOptionalString(fs.AzBlobConfig.KeyPrefix),
		UploadPartSize:      getOptionalInt64(fs.AzBlobConfig.UploadPartSize),
		UploadConcurrency:   getOptionalInt64(int64(fs.AzBlobConfig.UploadConcurrency)),
		DownloadPartSize:    getOptionalInt64(fs.AzBlobConfig.DownloadPartSize),
		DownloadConcurrency: getOptionalInt64(int64(fs.AzBlobConfig.DownloadConcurrency)),
		UseEmulator:         getOptionalBool(fs.AzBlobConfig.UseEmulator),
		AccessTier:          getOptionalString(fs.AzBlobConfig.AccessTier),
	}
	f.AzBlobConfig.AccountKey.fromSFTPGo(&fs.AzBlobConfig.AccountKey)
	f.AzBlobConfig.SASURL.fromSFTPGo(&fs.AzBlobConfig.SASURL)
	f.CryptConfig.Passphrase.fromSFTPGo(&fs.CryptConfig.Passphrase)
	f.SFTPConfig = sftpFsConfig{
		Endpoint:                getOptionalString(fs.SFTPConfig.Endpoint),
		Username:                getOptionalString(fs.SFTPConfig.Username),
		Prefix:                  getOptionalString(fs.SFTPConfig.Prefix),
		DisableCouncurrentReads: getOptionalBool(fs.SFTPConfig.DisableCouncurrentReads),
		BufferSize:              getOptionalInt64(fs.SFTPConfig.BufferSize),
		EqualityCheckMode:       getOptionalInt64(int64(fs.SFTPConfig.EqualityCheckMode)),
	}
	f.SFTPConfig.Password.fromSFTPGo(&fs.SFTPConfig.Password)
	f.SFTPConfig.PrivateKey.fromSFTPGo(&fs.SFTPConfig.PrivateKey)
	fingerprints, diags := types.ListValueFrom(ctx, types.StringType, fs.SFTPConfig.Fingerprints)
	if diags.HasError() {
		return diags
	}
	f.SFTPConfig.Fingerprints = fingerprints
	f.HTTPConfig = httpFsConfig{
		Endpoint:          getOptionalString(fs.HTTPConfig.Endpoint),
		Username:          getOptionalString(fs.HTTPConfig.Username),
		SkipTLSVerify:     getOptionalBool(fs.HTTPConfig.SkipTLSVerify),
		EqualityCheckMode: getOptionalInt64(int64(fs.HTTPConfig.EqualityCheckMode)),
	}
	f.HTTPConfig.Password.fromSFTPGo(&fs.HTTPConfig.Password)
	f.HTTPConfig.APIKey.fromSFTPGo(&fs.HTTPConfig.APIKey)
	return nil
}

type virtualFolderResourceModel struct {
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
	f.MappedPath = getOptionalString(folder.MappedPath)
	f.Description = getOptionalString(folder.Description)
	f.UsedQuotaSize = getOptionalInt64(folder.UsedQuotaSize)
	f.UsedQuotaFiles = getOptionalInt64(int64(folder.UsedQuotaFiles))
	f.LastQuotaUpdate = getOptionalInt64(folder.LastQuotaUpdate)

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

type adminFilters struct {
	AllowList       types.List       `tfsdk:"allow_list"`
	AllowAPIKeyAuth types.Bool       `tfsdk:"allow_api_key_auth"`
	Preferences     adminPreferences `tfsdk:"preferences"`
}

func (f *adminFilters) getTFAttributes() map[string]attr.Type {
	return map[string]attr.Type{
		"allow_list": types.ListType{
			ElemType: types.StringType,
		},
		"allow_api_key_auth": types.BoolType,
		"preferences": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"hide_user_page_sections":  types.Int64Type,
				"default_users_expiration": types.Int64Type,
			},
		},
	}
}

func (f *adminFilters) toSFTPGo(ctx context.Context) (client.AdminFilters, diag.Diagnostics) {
	filters := client.AdminFilters{
		AllowAPIKeyAuth: f.AllowAPIKeyAuth.ValueBool(),
		Preferences: client.AdminPreferences{
			HideUserPageSections:   int(f.Preferences.HideUserPageSections.ValueInt64()),
			DefaultUsersExpiration: int(f.Preferences.DefaultUsersExpiration.ValueInt64()),
		},
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
	f.Preferences = adminPreferences{
		HideUserPageSections:   getOptionalInt64(int64(filters.Preferences.HideUserPageSections)),
		DefaultUsersExpiration: getOptionalInt64(int64(filters.Preferences.DefaultUsersExpiration)),
	}
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
	Username       types.String        `tfsdk:"username"`
	Status         types.Int64         `tfsdk:"status"`
	Email          types.String        `tfsdk:"email"`
	Password       types.String        `tfsdk:"password"`
	Permissions    types.List          `tfsdk:"permissions"`
	Filters        types.Object        `tfsdk:"filters"`
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
