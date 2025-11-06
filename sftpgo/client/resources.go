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

package client

import (
	"github.com/sftpgo/sdk"
	"github.com/sftpgo/sdk/kms"
)

const (
	FTPFilesystemProvider = 7
)

var (
	// EnterpriseWebClientOptions defines the additional WebClientOptions available in the Enterprise version.
	EnterpriseWebClientOptions = []string{"shares-require-email-auth", "wopi-disabled", "rest-api-disabled"}
	WebClientOptions           = append(sdk.WebClientOptions, EnterpriseWebClientOptions...)
)

// Filesystem defines filesystem details
type Filesystem struct {
	Provider     sdk.FilesystemProvider `json:"provider"`
	OSConfig     sdk.OSFsConfig         `json:"osconfig,omitempty"`
	S3Config     sdk.S3FsConfig         `json:"s3config,omitempty"`
	GCSConfig    GCSFsConfig            `json:"gcsconfig,omitempty"`
	AzBlobConfig sdk.AzBlobFsConfig     `json:"azblobconfig,omitempty"`
	CryptConfig  sdk.CryptFsConfig      `json:"cryptconfig,omitempty"`
	SFTPConfig   SFTPFsConfig           `json:"sftpconfig,omitempty"`
	FTPConfig    FTPFsConfig            `json:"ftpconfig,omitempty"`
	HTTPConfig   sdk.HTTPFsConfig       `json:"httpconfig,omitempty"`
}

// BaseSFTPFsConfig defines the base configuration for SFTP based filesystem
type BaseSFTPFsConfig struct {
	Endpoint                string   `json:"endpoint,omitempty"`
	Username                string   `json:"username,omitempty"`
	Fingerprints            []string `json:"fingerprints,omitempty"`
	Prefix                  string   `json:"prefix,omitempty"`
	DisableCouncurrentReads bool     `json:"disable_concurrent_reads,omitempty"`
	BufferSize              int64    `json:"buffer_size,omitempty"`
	EqualityCheckMode       int      `json:"equality_check_mode,omitempty"`
	SocksProxy              string   `json:"socks_proxy,omitempty"`
	SocksUsername           string   `json:"socks_username,omitempty"`
}

// SFTPFsConfig defines the configuration for SFTP based filesystem
type SFTPFsConfig struct {
	BaseSFTPFsConfig
	Password      kms.BaseSecret `json:"password,omitempty"`
	PrivateKey    kms.BaseSecret `json:"private_key,omitempty"`
	KeyPassphrase kms.BaseSecret `json:"key_passphrase,omitempty"`
	SocksPassword kms.BaseSecret `json:"socks_password,omitempty"`
}

// BaseGCSFsConfig defines the base configuration for Google Cloud Storage based filesystems
type BaseGCSFsConfig struct {
	Bucket                string `json:"bucket,omitempty"`
	KeyPrefix             string `json:"key_prefix,omitempty"`
	CredentialFile        string `json:"-"`
	AutomaticCredentials  int    `json:"automatic_credentials,omitempty"`
	HierarchicalNamespace int    `json:"hns,omitempty"`
	StorageClass          string `json:"storage_class,omitempty"`
	ACL                   string `json:"acl,omitempty"`
	UploadPartSize        int64  `json:"upload_part_size,omitempty"`
	UploadPartMaxTime     int    `json:"upload_part_max_time,omitempty"`
}

// GCSFsConfig defines the configuration for Google Cloud Storage based filesystems
type GCSFsConfig struct {
	BaseGCSFsConfig
	Credentials kms.BaseSecret `json:"credentials,omitempty"`
}

// FTPFsConfig defines the configuration for FTP based filesystem
type FTPFsConfig struct {
	Endpoint string         `json:"endpoint,omitempty"`
	Username string         `json:"username,omitempty"`
	Password kms.BaseSecret `json:"password,omitempty"`
	// 0 disabled, 1 explicit, 2 implicit
	TLSMode       int  `json:"tls_mode,omitempty"`
	SkipTLSVerify bool `json:"skip_tls_verify,omitempty"`
}

type BaseVirtualFolder struct {
	sdk.BaseVirtualFolder
	FsConfig Filesystem `json:"filesystem"`
}

type VirtualFolder struct {
	BaseVirtualFolder
	VirtualPath string `json:"virtual_path"`
	// Maximum size allowed as bytes. 0 means unlimited, -1 included in user quota
	QuotaSize int64 `json:"quota_size"`
	// Maximum number of files allowed. 0 means unlimited, -1 included in user quota
	QuotaFiles int `json:"quota_files"`
}

// GroupUserSettings defines the settings to apply to users
type GroupUserSettings struct {
	sdk.BaseGroupUserSettings
	FsConfig Filesystem      `json:"filesystem"`
	Filters  BaseUserFilters `json:"filters"`
}

// Group defines an SFTPGo group.
// Groups are used to easily configure similar users
type Group struct {
	sdk.BaseGroup
	UserSettings   GroupUserSettings `json:"user_settings,omitempty"`
	VirtualFolders []VirtualFolder   `json:"virtual_folders,omitempty"`
}

// PasswordPolicy defines static password validation rules
type PasswordPolicy struct {
	Length   int `json:"length,omitempty"`
	Uppers   int `json:"uppers,omitempty"`
	Lowers   int `json:"lowers,omitempty"`
	Digits   int `json:"digits,omitempty"`
	Specials int `json:"specials,omitempty"`
}

// IsSet reports whether at least one rule is defined
func (p *PasswordPolicy) IsSet() bool {
	return p.Length > 0 || p.Uppers > 0 || p.Lowers > 0 || p.Digits > 0 || p.Specials > 0
}

// BaseUserFilters defines additional restrictions for a user
type BaseUserFilters struct {
	// only clients connecting from these IP/Mask are allowed.
	// IP/Mask must be in CIDR notation as defined in RFC 4632 and RFC 4291
	// for example "192.0.2.0/24" or "2001:db8::/32"
	AllowedIP []string `json:"allowed_ip,omitempty"`
	// clients connecting from these IP/Mask are not allowed.
	// Denied rules will be evaluated before allowed ones
	DeniedIP []string `json:"denied_ip,omitempty"`
	// these login methods are not allowed.
	// If null or empty any available login method is allowed
	DeniedLoginMethods []string `json:"denied_login_methods,omitempty"`
	// these protocols are not allowed.
	// If null or empty any available protocol is allowed
	DeniedProtocols []string `json:"denied_protocols,omitempty"`
	// filter based on shell patterns.
	// Please note that these restrictions can be easily bypassed.
	FilePatterns []sdk.PatternsFilter `json:"file_patterns,omitempty"`
	// max size allowed for a single upload, 0 means unlimited
	MaxUploadFileSize int64 `json:"max_upload_file_size,omitempty"`
	// TLS certificate attribute to use as username.
	// For FTP clients it must match the name provided using the
	// "USER" command
	TLSUsername sdk.TLSUsername `json:"tls_username,omitempty"`
	// TLSCerts defines the allowed TLS certificates for mutual authentication.
	// If provided will be checked before TLSUsername
	TLSCerts []string `json:"tls_certs,omitempty"`
	// user specific hook overrides
	Hooks sdk.HooksFilter `json:"hooks,omitempty"`
	// Disable checks for existence and automatic creation of home directory
	// and virtual folders.
	// SFTPGo requires that the user's home directory, virtual folder root,
	// and intermediate paths to virtual folders exist to work properly.
	// If you already know that the required directories exist, disabling
	// these checks will speed up login.
	// You could, for example, disable these checks after the first login
	DisableFsChecks bool `json:"disable_fs_checks,omitempty"`
	// WebClient related configuration options
	WebClient []string `json:"web_client,omitempty"`
	// API key auth allows to impersonate this user with an API key
	AllowAPIKeyAuth bool `json:"allow_api_key_auth,omitempty"`
	// UserType is an hint for authentication plugins.
	// It is ignored when using SFTPGo internal authentication
	UserType string `json:"user_type,omitempty"`
	// Per-source bandwidth limits
	BandwidthLimits []sdk.BandwidthLimit `json:"bandwidth_limits,omitempty"`
	// Defines the cache time, in seconds, for users authenticated using
	// an external auth hook. 0 means no cache
	ExternalAuthCacheTime int64 `json:"external_auth_cache_time,omitempty"`
	// Specifies an alternate starting directory. If not set, the default is "/".
	// This option is supported for SFTP/SCP, FTP and HTTP (WebClient/REST API) protocols.
	// Relative paths will use this directory as base
	StartDirectory string `json:"start_directory,omitempty"`
	// TwoFactorAuthProtocols defines protocols that require two factor authentication
	TwoFactorAuthProtocols []string `json:"two_factor_protocols,omitempty"`
	// Define the FTP security mode. Set to 1 to require TLS for both data and control
	// connection. This setting is useful if you want to allow both encrypted and plain text
	// FTP sessions globally and then you want to require encrypted sessions on a per-user
	// basis.
	// It has no effect if TLS is already required for all users in the configuration file.
	FTPSecurity int `json:"ftp_security,omitempty"`
	// If enabled the user can login with any password or no password at all.
	// Anonymous users are supported for FTP and WebDAV protocols and
	// permissions will be automatically set to "list" and "download" (read only)
	IsAnonymous bool `json:"is_anonymous,omitempty"`
	// Defines the default expiration for newly created shares as number of days.
	// 0 means no expiration
	DefaultSharesExpiration int `json:"default_shares_expiration,omitempty"`
	// Defines the maximum sharing expiration as a number of days. If set, users
	// must set an expiration for their shares and it must be less than or equal
	// to this number of days. 0 means any expiration
	MaxSharesExpiration int `json:"max_shares_expiration,omitempty"`
	// The password expires after the defined number of days. 0 means no expiration
	PasswordExpiration int `json:"password_expiration,omitempty"`
	// PasswordStrength defines the minimum password strength.
	// 0 means disabled, any password will be accepted. Values in the 50-70
	// range are suggested for common use cases.
	PasswordStrength int `json:"password_strength,omitempty"`
	// PasswordPolicy defines static password complexity requirements. Whenever
	// possible, prefer using the entropy-based approach provided by
	// PasswordStrength.
	PasswordPolicy PasswordPolicy `json:"password_policy,omitempty"`
	// AccessTime defines the time periods in which access is allowed
	AccessTime []sdk.TimePeriod `json:"access_time,omitempty"`
	// If enabled, only secure algorithms are allowed. This setting is currently enforced for SSH/SFTP
	EnforceSecureAlgorithms bool `json:"enforce_secure_algorithms"`
}

type UserFilters struct {
	BaseUserFilters
	RequirePasswordChange bool               `json:"require_password_change,omitempty"`
	AdditionalEmails      []string           `json:"additional_emails,omitempty"`
	TOTPConfig            sdk.TOTPConfig     `json:"totp_config,omitempty"`
	RecoveryCodes         []sdk.RecoveryCode `json:"recovery_codes,omitempty"`
	CustomPlaceholder1    string             `json:"custom1,omitempty"`
}

// User defines a SFTPGo user
type User struct {
	sdk.User
	// we remove the omitempty attribute from the password
	// otherwise setting an empty password will preserve the current one
	// on update
	Password       string          `json:"password"`
	Filters        UserFilters     `json:"filters"`
	VirtualFolders []VirtualFolder `json:"virtual_folders,omitempty"`
	FsConfig       Filesystem      `json:"filesystem"`
}
