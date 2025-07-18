---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "sftpgo_user Resource - sftpgo"
subcategory: ""
description: |-
  User
---

# sftpgo_user (Resource)

User



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `filesystem` (Attributes) Filesystem configuration. (see [below for nested schema](#nestedatt--filesystem))
- `home_dir` (String) The user cannot upload or download files outside this directory. Must be an absolute path.
- `permissions` (Map of String) Comma separated, per-directory, permissions.
- `status` (Number) 1 enabled, 0 disabled (login is not allowed).
- `username` (String) Unique username.

### Optional

- `additional_info` (String) Free form text field.
- `description` (String) Optional description.
- `download_bandwidth` (Number) Maximum download bandwidth as KB/s. Not set means unlimited. This is the default if no per-source limit match.
- `download_data_transfer` (Number) Maximum data transfer allowed for downloads as MB. Not set means no limit.
- `email` (String)
- `expiration_date` (Number) Account expiration date as unix timestamp in milliseconds. An expired account cannot login.
- `filters` (Attributes) (see [below for nested schema](#nestedatt--filters))
- `gid` (Number) If SFTPGo runs as root system user then the created files and directories will be assigned to this system GID. Default not set.
- `groups` (Attributes List) Groups. (see [below for nested schema](#nestedatt--groups))
- `max_sessions` (Number) Maximum concurrent sessions. Not set means no limit.
- `password` (String, Sensitive) Plain text password or hash format supported by SFTPGo. Set to empty to remove the password.
- `public_keys` (List of String) List of public keys in OpenSSH format.
- `quota_files` (Number) Maximum number of files allowed. Not set means no limit.
- `quota_size` (Number) Maximum size allowed as bytes. Not set means no limit.
- `role` (String) Role name.
- `total_data_transfer` (Number) Maximum total data transfer as MB. Not set means unlimited. You can set a total data transfer instead of the individual values for uploads and downloads.
- `uid` (Number) If SFTPGo runs as root system user then the created files and directories will be assigned to this system UID. Default not set.
- `upload_bandwidth` (Number) Maximum upload bandwidth as KB/s. Not set means unlimited. This is the default if no per-source limit match.
- `upload_data_transfer` (Number) Maximum data transfer allowed for uploads as MB. Not set means no limit.
- `virtual_folders` (Attributes List) (see [below for nested schema](#nestedatt--virtual_folders))

### Read-Only

- `created_at` (Number) Creation time as unix timestamp in milliseconds.
- `first_download` (Number) First download time as unix timestamp in milliseconds.
- `first_upload` (Number) First upload time as unix timestamp in milliseconds.
- `id` (String) Required to use the test framework. Matches the username.
- `last_login` (Number) Last login as unix timestamp in milliseconds.
- `last_password_change` (Number) Last password change as unix timestamp in milliseconds.
- `last_quota_update` (Number) Last quota update as unix timestamp in milliseconds.
- `updated_at` (Number) Last update time as unix timestamp in milliseconds.
- `used_download_data_transfer` (Number) Downloaded size, as bytes, since the last reset.
- `used_quota_files` (Number) Used quota as number of files.
- `used_quota_size` (Number) Used quota as bytes.
- `used_upload_data_transfer` (Number) Uploaded size, as bytes, since the last reset.

<a id="nestedatt--filesystem"></a>
### Nested Schema for `filesystem`

Required:

- `provider` (Number) Provider. 0 = local filesystem, 1 = S3 Compatible, 2 = Google Cloud, 3 = Azure Blob, 4 = Local encrypted, 5 = SFTP, 6 = HTTP

Optional:

- `azblobconfig` (Attributes) (see [below for nested schema](#nestedatt--filesystem--azblobconfig))
- `cryptconfig` (Attributes) (see [below for nested schema](#nestedatt--filesystem--cryptconfig))
- `gcsconfig` (Attributes) (see [below for nested schema](#nestedatt--filesystem--gcsconfig))
- `httpconfig` (Attributes) (see [below for nested schema](#nestedatt--filesystem--httpconfig))
- `osconfig` (Attributes) (see [below for nested schema](#nestedatt--filesystem--osconfig))
- `s3config` (Attributes) (see [below for nested schema](#nestedatt--filesystem--s3config))
- `sftpconfig` (Attributes) (see [below for nested schema](#nestedatt--filesystem--sftpconfig))

<a id="nestedatt--filesystem--azblobconfig"></a>
### Nested Schema for `filesystem.azblobconfig`

Optional:

- `access_tier` (String) Blob Access Tier. Not set means the container default.
- `account_key` (String, Sensitive) Plain text account key. If you set a string in SFTPGo secret format, SFTPGo will keep the current secret on updates while the Terraform plan will save your value. Don't do this unless you are sure the values match (e.g because you imported an existing resource).
- `account_name` (String)
- `container` (String)
- `download_concurrency` (Number) How many parts are downloaded in parallel. Default: 5.
- `download_part_size` (Number) The buffer size (in MB) to use for multipart downloads. If this value is not set, the default value (5MB) will be used.
- `endpoint` (String) Optional endpoint. Default is "blob.core.windows.net". If you use the emulator the endpoint must include the protocol, for example "http://127.0.0.1:10000".
- `key_prefix` (String) If specified then the SFTPGo user will be restricted to objects starting with the specified prefix. The prefix must not start with "/" and must end with "/"
- `sas_url` (String, Sensitive) Plain text SAS URL. If you set a string in SFTPGo secret format, SFTPGo will keep the current secret on updates while the Terraform plan will save your value. Don't do this unless you are sure the values match (e.g because you imported an existing resource).
- `upload_concurrency` (Number) How many parts are uploaded in parallel. Default: 5.
- `upload_part_size` (Number) The buffer size (in MB) to use for multipart uploads. If this value is not set, the default value (5MB) will be used.
- `use_emulator` (Boolean)


<a id="nestedatt--filesystem--cryptconfig"></a>
### Nested Schema for `filesystem.cryptconfig`

Optional:

- `passphrase` (String, Sensitive) Plain text passphrase. If you set a string in SFTPGo secret format, SFTPGo will keep the current secret on updates while the Terraform plan will save your value. Don't do this unless you are sure the values match (e.g because you imported an existing resource).
- `read_buffer_size` (Number) Optional read buffer size, as MB, to use for downloads. Omit to disable buffering, that's fine in most use cases.
- `write_buffer_size` (Number) Optional write buffer size, as MB, to use for uploads. Omit to disable buffering, that's fine in most use cases.


<a id="nestedatt--filesystem--gcsconfig"></a>
### Nested Schema for `filesystem.gcsconfig`

Required:

- `bucket` (String)

Optional:

- `acl` (String) The ACL to apply to uploaded objects. Not set means the bucket default.
- `automatic_credentials` (Number)
- `credentials` (String, Sensitive) Plain text credentials. If you set a string in SFTPGo secret format, SFTPGo will keep the current secret on updates while the Terraform plan will save your value. Don't do this unless you are sure the values match (e.g because you imported an existing resource).
- `hns` (Number) Set to 1 if Hierarchical namespace is enabled for the bucket. Available in the Enterprise edition.
- `key_prefix` (String) If specified then the SFTPGo user will be restricted to objects starting with the specified prefix. The prefix must not start with "/" and must end with "/"
- `storage_class` (String) The storage class to use when storing objects. Leave not set for default.
- `upload_part_max_time` (Number) The maximum time allowed, in seconds, to upload a single chunk. The default value is 32. Not set means use the default.
- `upload_part_size` (Number) The buffer size (in MB) to use for multipart uploads. The default value is 16MB. Not set means use the default.


<a id="nestedatt--filesystem--httpconfig"></a>
### Nested Schema for `filesystem.httpconfig`

Required:

- `endpoint` (String)

Optional:

- `api_key` (String, Sensitive) Plain text API key. If you set a string in SFTPGo secret format, SFTPGo will keep the current secret on updates while the Terraform plan will save your value. Don't do this unless you are sure the values match (e.g because you imported an existing resource).
- `equality_check_mode` (Number)
- `password` (String, Sensitive) Plain text password. If you set a string in SFTPGo secret format, SFTPGo will keep the current secret on updates while the Terraform plan will save your value. Don't do this unless you are sure the values match (e.g because you imported an existing resource).
- `skip_tls_verify` (Boolean)
- `username` (String)


<a id="nestedatt--filesystem--osconfig"></a>
### Nested Schema for `filesystem.osconfig`

Optional:

- `read_buffer_size` (Number) Optional read buffer size, as MB, to use for downloads. Omit to disable buffering, that's fine in most use cases.
- `write_buffer_size` (Number) Optional write buffer size, as MB, to use for uploads. Omit to disable no buffering, that's fine in most use cases.


<a id="nestedatt--filesystem--s3config"></a>
### Nested Schema for `filesystem.s3config`

Required:

- `bucket` (String)

Optional:

- `access_key` (String)
- `access_secret` (String, Sensitive) Plain text access secret. If you set a string in SFTPGo secret format, SFTPGo will keep the current secret on updates while the Terraform plan will save your value. Don't do this unless you are sure the values match (e.g because you imported an existing resource).
- `acl` (String) The canned ACL to apply to uploaded objects. Not set means the bucket default.
- `download_concurrency` (Number) How many parts are downloaded in parallel. Not set means the default (5). Ignored for partial downloads.
- `download_part_max_time` (Number) The maximum time allowed, in seconds, to download a single chunk. Not set means no timeout. Ignored for partial downloads.
- `download_part_size` (Number) The buffer size (in MB) to use for multipart downloads. If this value is not set, the default value (5MB) will be used.
- `endpoint` (String) The endpoint is generally required for S3 compatible backends. For AWS S3, leave not set to use the default endpoint for the specified region.
- `force_path_style` (Boolean) If set path-style addressing is used, i.e. http://s3.amazonaws.com/BUCKET/KEY
- `key_prefix` (String) If specified then the SFTPGo user will be restricted to objects starting with the specified prefix. The prefix must not start with "/" and must end with "/"
- `region` (String)
- `role_arn` (String) Optional IAM Role ARN to assume.
- `session_token` (String) Optional Session token that is a part of temporary security credentials provisioned by AWS STS.
- `skip_tls_verify` (Boolean) If set the S3 client accepts any TLS certificate presented by the server and any host name in that certificate. In this mode, TLS is susceptible to man-in-the-middle attacks. This should be used only for testing.
- `sse_customer_key` (String, Sensitive) Plain text Server-Side encryption key. If you set a string in SFTPGo secret format, SFTPGo will keep the current secret on updates while the Terraform plan will save your value. Don't do this unless you are sure the values match (e.g because you imported an existing resource).
- `storage_class` (String) The storage class to use when storing objects. Leave not set for default.
- `upload_concurrency` (Number) How many parts are uploaded in parallel. Not set means the default (5).
- `upload_part_max_time` (Number) The maximum time allowed, in seconds, to upload a single chunk. Not set means no timeout.
- `upload_part_size` (Number) The buffer size (in MB) to use for multipart uploads. If this value is not set, the default value (5MB) will be used.


<a id="nestedatt--filesystem--sftpconfig"></a>
### Nested Schema for `filesystem.sftpconfig`

Required:

- `endpoint` (String) SFTP endpoint as host:port. Port is always required.
- `prefix` (String) Similar to a chroot for local filesystem. Example: "/somedir/subdir".
- `username` (String)

Optional:

- `buffer_size` (Number) The buffer size (in MB) to use for uploads/downloads. Buffering could improve performance for high latency networks. With buffering enabled upload resume is not supported and a file cannot be opened for both reading and writing at the same time. Not set means disabled.
- `disable_concurrent_reads` (Boolean) Concurrent reads are safe to use and disabling them will degrade performance so they are enabled by default. Some servers automatically delete files once they are downloaded. Using concurrent reads is problematic with such servers.
- `equality_check_mode` (Number) Defines how to check if this config points to the same server as another config. By default both the endpoint and the username must match. 1 means that only the endpoint must match. If different configs point to the same server the renaming between the fs configs is allowed.
- `fingerprints` (List of String) SHA256 fingerprints to validate when connecting to the external SFTP server. If not set any host key will be accepted: this is a security risk.
- `key_passphrase` (String, Sensitive) Plain text passphrase for the private key. If you set a string in SFTPGo secret format, SFTPGo will keep the current secret on updates while the Terraform plan will save your value. Don't do this unless you are sure the values match (e.g because you imported an existing resource).
- `password` (String, Sensitive) Plain text password. If you set a string in SFTPGo secret format, SFTPGo will keep the current secret on updates while the Terraform plan will save your value. Don't do this unless you are sure the values match (e.g because you imported an existing resource).
- `private_key` (String, Sensitive) Plain text private key. If you set a string in SFTPGo secret format, SFTPGo will keep the current secret on updates while the Terraform plan will save your value. Don't do this unless you are sure the values match (e.g because you imported an existing resource).
- `socks5_password` (String, Sensitive) Plain text SOCKS5 password. If you set a string in SFTPGo secret format, SFTPGo will keep the current secret on updates while the Terraform plan will save your value. Don't do this unless you are sure the values match (e.g because you imported an existing resource). Available in the Enterprise edition.
- `socks5_proxy` (String) The address of the SOCKS5 proxy server, including the hostname or IP and the port number. Available in the Enterprise edition.
- `socks5_username` (String) The optional SOCKS5 username. Available in the Enterprise edition.



<a id="nestedatt--filters"></a>
### Nested Schema for `filters`

Optional:

- `access_time` (Attributes List) Time periods in which access is allowed (see [below for nested schema](#nestedatt--filters--access_time))
- `additional_emails` (List of String) Additional email addresses.
- `allow_api_key_auth` (Boolean) If set, API Key authentication is allowed.
- `allowed_ip` (List of String) Only connections from these IP/Mask are allowed. IP/Mask must be in CIDR notation as defined in RFC 4632 and RFC 4291 for example "192.0.2.0/24" or "2001:db8::/32"
- `bandwidth_limits` (Attributes List) (see [below for nested schema](#nestedatt--filters--bandwidth_limits))
- `check_password_disabled` (Boolean) If set, check password hook will not be executed.
- `custom1` (String) An extra placeholder value available for use in group configurations. It can be referenced as %custom1%. Available in the Enterprise edition.
- `default_shares_expiration` (Number) Default expiration for newly created shares as number of days. Not set means no default expiration.
- `denied_ip` (List of String) Connections from these IP/Mask are allowed. Denied rules will be evaluated before allowed ones.
- `denied_login_methods` (List of String) Disabled login methods. Valid values: "publickey", "password", "password-over-SSH", "keyboard-interactive", "publickey+password", "publickey+keyboard-interactive", "TLSCertificate", "TLSCertificate+password"
- `denied_protocols` (List of String) Disabled protocols. Valid values: SSH, FTP, DAV, HTTP
- `disable_fs_checks` (Boolean) Disable checks for existence and automatic creation of home directory and virtual folders after user login.
- `enforce_secure_algorithms` (Boolean) If enabled, only secure algorithms are allowed. This setting is currently enforced for SSH/SFTP. Available in the Enterprise edition.
- `external_auth_cache_time` (Number) Defines the cache time, in seconds, for users authenticated using an external auth hook. Not set means no cache.
- `external_auth_disabled` (Boolean) If set, external auth hook will not be executed.
- `file_patterns` (Attributes List) (see [below for nested schema](#nestedatt--filters--file_patterns))
- `ftp_security` (Number) FTP security mode. Set to 1 to require TLS for both data and control connection.
- `is_anonymous` (Boolean) If enabled the user can login with any password or no password at all. Anonymous users are supported for FTP and WebDAV protocols and permissions will be automatically set to "list" and "download" (read only)
- `max_shares_expiration` (Number) Maximum allowed expiration, as a number of days, when a user creates or updates a share. Not set means that non-expiring shares are allowed.
- `max_upload_file_size` (Number) Max size allowed for a single upload. Unset means no limit.
- `password_expiration` (Number) The password expires after the defined number of days. Not set means no expiration
- `password_strength` (Number) Minimum password strength. Not set means disabled, any password will be accepted. Values in the 50-70 range are suggested for common use cases.
- `pre_login_disabled` (Boolean) If set, external pre-login hook will not be executed.
- `require_password_change` (Boolean) If set, user must change their password from WebClient/REST API at next login.
- `start_directory` (String) Alternate starting directory. If not set, the default is "/". This option is supported for SFTP/SCP, FTP and HTTP (WebClient/REST API) protocols. Relative paths will use this directory as base.
- `tls_certs` (List of String) TLS certificates for mutual authentication. If provided will be checked before TLS username.
- `tls_username` (String) TLS certificate attribute to use as username. For FTP clients it must match the name provided using the "USER" command. For WebDAV, if no username is provided, the CN will be used as username. For WebDAV clients it must match the implicit or provided username.
- `two_factor_protocols` (List of String) Defines protocols that require two factor authentication. Valid values: SSH, FTP, HTTP
- `user_type` (String) Hint for authentication plugins. Valid values: LDAPUser, OSUser
- `web_client` (List of String) Web Client/user REST API restrictions. Valid values: write-disabled, password-change-disabled, password-reset-disabled, publickey-change-disabled, tls-cert-change-disabled, mfa-disabled, api-key-auth-change-disabled, info-change-disabled, shares-disabled, shares-without-password-disabled, shares-require-email-auth, wopi-disabled, rest-api-disabled. Only available in the Enterprise version: shares-require-email-auth, wopi-disabled, rest-api-disabled

<a id="nestedatt--filters--access_time"></a>
### Nested Schema for `filters.access_time`

Required:

- `day_of_week` (Number) Day of week, 0 Sunday, 6 Saturday
- `from` (String) Start time in HH:MM format
- `to` (String) End time in HH:MM format


<a id="nestedatt--filters--bandwidth_limits"></a>
### Nested Schema for `filters.bandwidth_limits`

Required:

- `sources` (List of String) Source networks in CIDR notation as defined in RFC 4632 and RFC 4291 for example "192.0.2.0/24" or "2001:db8::/32". The limit applies if the defined networks contain the client IP.

Optional:

- `download_bandwidth` (Number) Maximum download bandwidth as KB/s.
- `upload_bandwidth` (Number) Maximum upload bandwidth as KB/s.


<a id="nestedatt--filters--file_patterns"></a>
### Nested Schema for `filters.file_patterns`

Required:

- `path` (String) Virtual path, if no other specific filter is defined, the filter applies for sub directories too.

Optional:

- `allowed_patterns` (List of String) Files/directories with these, case insensitive, patterns are allowed. Allowed file patterns are evaluated before the denied ones.
- `denied_patterns` (List of String) Files/directories with these, case insensitive, patterns are not allowed.
- `deny_policy` (Number) Set to 1 to hide denied files/directories in directory listing.



<a id="nestedatt--groups"></a>
### Nested Schema for `groups`

Required:

- `name` (String) Group name.
- `type` (Number) Group type. 1 = Primary, 2 = Secondary, 3 = Membership only.


<a id="nestedatt--virtual_folders"></a>
### Nested Schema for `virtual_folders`

Required:

- `name` (String) Unique folder name
- `quota_files` (Number) Maximum number of files allowed. Not set means unlimited, -1 included in user quota
- `quota_size` (Number) Maximum size allowed as bytes. Not set means unlimited, -1 included in user quota
- `virtual_path` (String) The folder will be available on this path.

Optional:

- `description` (String) Optional description.
- `last_quota_update` (Number) Last quota update as unix timestamp in milliseconds
- `mapped_path` (String) Absolute path to a local directory. This is the folder root path for local storage provider. For non-local filesystems it will store temporary files.
- `used_quota_files` (Number) Used quota as number of files.
- `used_quota_size` (Number) Used quota as bytes.

Read-Only:

- `filesystem` (Attributes) Filesystem configuration. (see [below for nested schema](#nestedatt--virtual_folders--filesystem))

<a id="nestedatt--virtual_folders--filesystem"></a>
### Nested Schema for `virtual_folders.filesystem`

Read-Only:

- `azblobconfig` (Attributes) (see [below for nested schema](#nestedatt--virtual_folders--filesystem--azblobconfig))
- `cryptconfig` (Attributes) (see [below for nested schema](#nestedatt--virtual_folders--filesystem--cryptconfig))
- `gcsconfig` (Attributes) (see [below for nested schema](#nestedatt--virtual_folders--filesystem--gcsconfig))
- `httpconfig` (Attributes) (see [below for nested schema](#nestedatt--virtual_folders--filesystem--httpconfig))
- `osconfig` (Attributes) (see [below for nested schema](#nestedatt--virtual_folders--filesystem--osconfig))
- `provider` (Number) Provider. 0 = local filesystem, 1 = S3 Compatible, 2 = Google Cloud, 3 = Azure Blob, 4 = Local encrypted, 5 = SFTP, 6 = HTTP
- `s3config` (Attributes) (see [below for nested schema](#nestedatt--virtual_folders--filesystem--s3config))
- `sftpconfig` (Attributes) (see [below for nested schema](#nestedatt--virtual_folders--filesystem--sftpconfig))

<a id="nestedatt--virtual_folders--filesystem--azblobconfig"></a>
### Nested Schema for `virtual_folders.filesystem.azblobconfig`

Read-Only:

- `access_tier` (String)
- `account_key` (String) SFTPGo secret formatted as string: "$<status>$<key>$<additional data length>$<additional data><payload>".
- `account_name` (String)
- `container` (String)
- `download_concurrency` (Number) How many parts are downloaded in parallel.
- `download_part_size` (Number) The buffer size (in MB) to use for multipart downloads.
- `endpoint` (String) Optional endpoint
- `key_prefix` (String) If specified then the SFTPGo user will be restricted to objects starting with this prefix.
- `sas_url` (String) SFTPGo secret formatted as string: "$<status>$<key>$<additional data length>$<additional data><payload>".
- `upload_concurrency` (Number) How many parts are uploaded in parallel.
- `upload_part_size` (Number) The buffer size (in MB) to use for multipart uploads.
- `use_emulator` (Boolean)


<a id="nestedatt--virtual_folders--filesystem--cryptconfig"></a>
### Nested Schema for `virtual_folders.filesystem.cryptconfig`

Read-Only:

- `passphrase` (String) SFTPGo secret formatted as string: "$<status>$<key>$<additional data length>$<additional data><payload>".
- `read_buffer_size` (Number) Optional read buffer size, as MB, to use for downloads.
- `write_buffer_size` (Number) Optional write buffer size, as MB, to use for uploads.


<a id="nestedatt--virtual_folders--filesystem--gcsconfig"></a>
### Nested Schema for `virtual_folders.filesystem.gcsconfig`

Read-Only:

- `acl` (String) The ACL to apply to uploaded objects. Empty means the bucket default.
- `automatic_credentials` (Number) If set to 1 SFTPGo will use credentials from the environment
- `bucket` (String)
- `credentials` (String) SFTPGo secret formatted as string: "$<status>$<key>$<additional data length>$<additional data><payload>".
- `hns` (Number) 1 if Hierarchical namespace support is enabled for the bucket. Available in the Enterprise edition.
- `key_prefix` (String) If specified then the SFTPGo user will be restricted to objects starting with this prefix.
- `storage_class` (String)
- `upload_part_max_time` (Number) The maximum time allowed, in seconds, to upload a single chunk. The default value is 32. Not set means use the default.
- `upload_part_size` (Number) The buffer size (in MB) to use for multipart uploads. The default value is 16MB. Not set means use the default.


<a id="nestedatt--virtual_folders--filesystem--httpconfig"></a>
### Nested Schema for `virtual_folders.filesystem.httpconfig`

Read-Only:

- `api_key` (String) SFTPGo secret formatted as string: "$<status>$<key>$<additional data length>$<additional data><payload>".
- `endpoint` (String)
- `equality_check_mode` (Number)
- `password` (String) SFTPGo secret formatted as string: "$<status>$<key>$<additional data length>$<additional data><payload>".
- `skip_tls_verify` (Boolean)
- `username` (String)


<a id="nestedatt--virtual_folders--filesystem--osconfig"></a>
### Nested Schema for `virtual_folders.filesystem.osconfig`

Read-Only:

- `read_buffer_size` (Number) Optional read buffer size, as MB, to use for downloads.
- `write_buffer_size` (Number) Optional write buffer size, as MB, to use for uploads.


<a id="nestedatt--virtual_folders--filesystem--s3config"></a>
### Nested Schema for `virtual_folders.filesystem.s3config`

Read-Only:

- `access_key` (String)
- `access_secret` (String) SFTPGo secret formatted as string: "$<status>$<key>$<additional data length>$<additional data><payload>".
- `acl` (String) The canned ACL to apply to uploaded objects. Empty means the bucket default.
- `bucket` (String)
- `download_concurrency` (Number) How many parts are downloaded in parallel. Ignored for partial downloads.
- `download_part_max_time` (Number) The maximum time allowed, in seconds, to download a single chunk. Not set means no timeout.
- `download_part_size` (Number) The buffer size (in MB) to use for multipart downloads.
- `endpoint` (String) The endpoint is generally required for S3 compatible backends.
- `force_path_style` (Boolean) If enabled path-style addressing is used, i.e. http://s3.amazonaws.com/BUCKET/KEY
- `key_prefix` (String) If specified then the SFTPGo user will be restricted to objects starting with this prefix.
- `region` (String)
- `role_arn` (String) IAM Role ARN to assume.
- `session_token` (String) Optional Session token that is a part of temporary security credentials provisioned by AWS STS.
- `skip_tls_verify` (Boolean) If set the S3 client accepts any TLS certificate presented by the server and any host name in that certificate. In this mode, TLS is susceptible to man-in-the-middle attacks. This should be used only for testing.
- `sse_customer_key` (String) SFTPGo secret formatted as string: "$<status>$<key>$<additional data length>$<additional data><payload>".
- `storage_class` (String)
- `upload_concurrency` (Number) How many parts are uploaded in parallel. Not set means the default (5).
- `upload_part_max_time` (Number) The maximum time allowed, in seconds, to upload a single chunk. Not set means no timeout.
- `upload_part_size` (Number) The buffer size (in MB) to use for multipart uploads.


<a id="nestedatt--virtual_folders--filesystem--sftpconfig"></a>
### Nested Schema for `virtual_folders.filesystem.sftpconfig`

Read-Only:

- `buffer_size` (Number) The buffer size (in MB) to use for uploads/downloads.
- `disable_concurrent_reads` (Boolean)
- `endpoint` (String) SFTP endpoint as host:port.
- `equality_check_mode` (Number)
- `fingerprints` (List of String) SHA256 fingerprints to validate when connecting to the external SFTP server.
- `key_passphrase` (String) SFTPGo secret formatted as string: "$<status>$<key>$<additional data length>$<additional data><payload>".
- `password` (String) SFTPGo secret formatted as string: "$<status>$<key>$<additional data length>$<additional data><payload>".
- `prefix` (String) Restrict access to this path.
- `private_key` (String) SFTPGo secret formatted as string: "$<status>$<key>$<additional data length>$<additional data><payload>".
- `socks5_password` (String) SFTPGo secret formatted as string: "$<status>$<key>$<additional data length>$<additional data><payload>". Available in the Enterprise edition.
- `socks5_proxy` (String) The address of the SOCKS5 proxy server, including the hostname or IP and the port number. Available in the Enterprise edition.
- `socks5_username` (String) The optional SOCKS5 username. Available in the Enterprise edition.
- `username` (String)
