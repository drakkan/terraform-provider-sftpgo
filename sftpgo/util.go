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
	"strings"

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/sftpgo/sdk"
	"github.com/sftpgo/sdk/kms"
)

const (
	computedSecretDescription    = `SFTPGo secret formatted as string: "$<status>$<key>$<additional data length>$<additional data><payload>".`
	secretDescriptionGeneric     = `If you set a string in SFTPGo secret format, SFTPGo will keep the current secret on updates while the Terraform plan will save your value. Don't do this unless you are sure the values match (e.g because you imported an existing resource).`
	writeOnlyDescriptionGeneric  = `Write-only variant of the matching attribute: the value is read from the configuration only and is never persisted to the Terraform plan or state. Requires Terraform 1.11 or later. Mutually exclusive with the non write-only attribute. Use the companion _wo_version attribute to trigger an update.`
	writeOnlyVersionDescGeneric  = `Trigger attribute for the matching write-only attribute. Because write-only values are not stored in state, Terraform cannot detect changes to them. Bump this value to force the provider to re-apply the write-only value on the next apply.`
	enterpriseFeatureNote        = `Available in the Enterprise edition`
)

func getComputedSchemaForFilesystem() dsschema.SingleNestedAttribute {
	return dsschema.SingleNestedAttribute{
		Computed:    true,
		Description: "Filesystem configuration.",
		Attributes: map[string]dsschema.Attribute{
			"provider": dsschema.Int64Attribute{
				Computed:    true,
				Description: "Provider. 0 = local filesystem, 1 = S3 Compatible, 2 = Google Cloud, 3 = Azure Blob, 4 = Local encrypted, 5 = SFTP, 6 = HTTP",
			},
			"osconfig": dsschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dsschema.Attribute{
					"read_buffer_size": dsschema.Int64Attribute{
						Computed:    true,
						Description: "Optional read buffer size, as MB, to use for downloads.",
					},
					"write_buffer_size": dsschema.Int64Attribute{
						Computed:    true,
						Description: "Optional write buffer size, as MB, to use for uploads.",
					},
				},
			},
			"s3config": dsschema.SingleNestedAttribute{
				Computed:    true,
				Description: "S3 compatible object storage configuration details.",
				Attributes: map[string]dsschema.Attribute{
					"bucket": dsschema.StringAttribute{
						Computed:    true,
						Description: "S3 bucket name.",
					},
					"region": dsschema.StringAttribute{
						Computed:    true,
						Description: "S3 region.",
					},
					"access_key": dsschema.StringAttribute{
						Computed:    true,
						Description: "AWS Access Key ID for authentication. Leave blank when using IAM roles or instance profiles.",
					},
					"access_secret": dsschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"access_secret_wo":            computedWOPlaceholder(false),
					"access_secret_wo_version":    computedWOPlaceholder(true),
					"sse_customer_key": dsschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"sse_customer_key_wo":         computedWOPlaceholder(false),
					"sse_customer_key_wo_version": computedWOPlaceholder(true),
					"key_prefix": dsschema.StringAttribute{
						Computed:    true,
						Description: `If specified then the SFTPGo user will be restricted to objects starting with this prefix.`,
					},
					"role_arn": dsschema.StringAttribute{
						Computed:    true,
						Description: "IAM Role ARN to assume.",
					},
					"session_token": dsschema.StringAttribute{
						Computed:    true,
						Description: "Optional Session token that is a part of temporary security credentials provisioned by AWS STS.",
					},
					"endpoint": dsschema.StringAttribute{
						Computed:    true,
						Description: "The endpoint is generally required for S3 compatible backends.",
					},
					"storage_class": dsschema.StringAttribute{
						Computed:    true,
						Description: "S3 storage class for uploaded objects (e.g. STANDARD, STANDARD_IA, GLACIER). Leave empty for the default storage class.",
					},
					"acl": dsschema.StringAttribute{
						Computed:    true,
						Description: "The canned ACL to apply to uploaded objects. Empty means the bucket default.",
					},
					"upload_part_size": dsschema.Int64Attribute{
						Computed:    true,
						Description: "The buffer size (in MB) to use for multipart uploads.",
					},
					"upload_concurrency": dsschema.Int64Attribute{
						Computed:    true,
						Description: "How many parts are uploaded in parallel. Not set means the default (5).",
					},
					"download_part_size": dsschema.Int64Attribute{
						Computed:    true,
						Description: "The buffer size (in MB) to use for multipart downloads.",
					},
					"upload_part_max_time": dsschema.Int64Attribute{
						Computed:    true,
						Description: "The maximum time allowed, in seconds, to upload a single chunk. Not set means no timeout.",
					},
					"download_concurrency": dsschema.Int64Attribute{
						Computed:    true,
						Description: "How many parts are downloaded in parallel. Ignored for partial downloads.",
					},
					"download_part_max_time": dsschema.Int64Attribute{
						Computed:    true,
						Description: "The maximum time allowed, in seconds, to download a single chunk. Not set means no timeout.",
					},
					"force_path_style": dsschema.BoolAttribute{
						Computed:    true,
						Description: `If enabled path-style addressing is used, i.e. http://s3.amazonaws.com/BUCKET/KEY`,
					},
					"skip_tls_verify": dsschema.BoolAttribute{
						Computed:    true,
						Description: `If set the S3 client accepts any TLS certificate presented by the server and any host name in that certificate. In this mode, TLS is susceptible to man-in-the-middle attacks. This should be used only for testing.`,
					},
					"checksum_algorithm": dsschema.StringAttribute{
						Computed:    true,
						Description: "Checksum algorithm to compute and send with uploads (PutObject, multipart upload, CopyObject) for end-to-end integrity verification. Empty means no checksum is sent. " + enterpriseFeatureNote + ".",
					},
				},
			},
			"gcsconfig": dsschema.SingleNestedAttribute{
				Computed:    true,
				Description: "Google Cloud Storage configuration details.",
				Attributes: map[string]dsschema.Attribute{
					"bucket": dsschema.StringAttribute{
						Computed:    true,
						Description: "GCS bucket name.",
					},
					"key_prefix": dsschema.StringAttribute{
						Computed:    true,
						Description: `If specified then the SFTPGo user will be restricted to objects starting with this prefix.`,
					},
					"universe_domain": dsschema.StringAttribute{
						Computed:    true,
						Description: `The universe domain to use for Google Cloud API requests. If omitted or empty, the default public domain (googleapis.com) is used. Set this value if you need to connect to a custom Google Cloud environment, such as Google Distributed Cloud or a Sovereign Cloud. ` + enterpriseFeatureNote,
					},
					"credentials": dsschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"credentials_wo":         computedWOPlaceholder(false),
					"credentials_wo_version": computedWOPlaceholder(true),
					"automatic_credentials": dsschema.Int64Attribute{
						Computed:    true,
						Description: "If set to 1 SFTPGo will use credentials from the environment",
					},
					"hns": dsschema.Int64Attribute{
						Computed:    true,
						Description: "1 if Hierarchical namespace support is enabled for the bucket. " + enterpriseFeatureNote + ".",
					},
					"storage_class": dsschema.StringAttribute{
						Computed:    true,
						Description: "Google Cloud Storage class for uploaded objects (e.g. STANDARD, NEARLINE, COLDLINE, ARCHIVE). Leave empty for the default storage class.",
					},
					"acl": dsschema.StringAttribute{
						Computed:    true,
						Description: "The ACL to apply to uploaded objects. Empty means the bucket default.",
					},
					"upload_part_size": dsschema.Int64Attribute{
						Computed:    true,
						Description: "The buffer size (in MB) to use for multipart uploads. The default value is 16MB. Not set means use the default.",
					},
					"upload_part_max_time": dsschema.Int64Attribute{
						Computed:    true,
						Description: "The maximum time allowed, in seconds, to upload a single chunk. The default value is 32. Not set means use the default.",
					},
				},
			},
			"azblobconfig": dsschema.SingleNestedAttribute{
				Computed:    true,
				Description: "Azure Blob Storage configuration details.",
				Attributes: map[string]dsschema.Attribute{
					"container": dsschema.StringAttribute{
						Computed:    true,
						Description: "Azure Blob Storage container name.",
					},
					"account_name": dsschema.StringAttribute{
						Computed:    true,
						Description: "Storage account name. Leave blank to use SAS URL.",
					},
					"account_key": dsschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"account_key_wo":         computedWOPlaceholder(false),
					"account_key_wo_version": computedWOPlaceholder(true),
					"sas_url": dsschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"sas_url_wo":         computedWOPlaceholder(false),
					"sas_url_wo_version": computedWOPlaceholder(true),
					"endpoint": dsschema.StringAttribute{
						Computed:    true,
						Description: "Optional endpoint",
					},
					"key_prefix": dsschema.StringAttribute{
						Computed:    true,
						Description: `If specified then the SFTPGo user will be restricted to objects starting with this prefix.`,
					},
					"upload_part_size": dsschema.Int64Attribute{
						Computed:    true,
						Description: "The buffer size (in MB) to use for multipart uploads.",
					},
					"upload_concurrency": dsschema.Int64Attribute{
						Computed:    true,
						Description: "How many parts are uploaded in parallel.",
					},
					"download_part_size": dsschema.Int64Attribute{
						Computed:    true,
						Description: "The buffer size (in MB) to use for multipart downloads.",
					},
					"download_concurrency": dsschema.Int64Attribute{
						Computed:    true,
						Description: "How many parts are downloaded in parallel.",
					},
					"use_emulator": dsschema.BoolAttribute{
						Computed:    true,
						Description: "If true, the Azure Storage Emulator (Azurite) is used instead of the cloud service.",
					},
					"access_tier": dsschema.StringAttribute{
						Computed:    true,
						Description: "Blob access tier. Valid values: empty, Archive, Hot, Cool.",
					},
				},
			},
			"cryptconfig": dsschema.SingleNestedAttribute{
				Computed:    true,
				Description: "Encrypted local filesystem configuration details.",
				Attributes: map[string]dsschema.Attribute{
					"passphrase": dsschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"passphrase_wo":         computedWOPlaceholder(false),
					"passphrase_wo_version": computedWOPlaceholder(true),
					"read_buffer_size": dsschema.Int64Attribute{
						Computed:    true,
						Description: "Optional read buffer size, as MB, to use for downloads.",
					},
					"write_buffer_size": dsschema.Int64Attribute{
						Computed:    true,
						Description: "Optional write buffer size, as MB, to use for uploads.",
					},
				},
			},
			"sftpconfig": dsschema.SingleNestedAttribute{
				Computed:    true,
				Description: "Remote SFTP server configuration details.",
				Attributes: map[string]dsschema.Attribute{
					"endpoint": dsschema.StringAttribute{
						Computed:    true,
						Description: "SFTP endpoint as host:port.",
					},
					"username": dsschema.StringAttribute{
						Computed:    true,
						Description: "Username for SFTP authentication.",
					},
					"password": dsschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"password_wo":         computedWOPlaceholder(false),
					"password_wo_version": computedWOPlaceholder(true),
					"private_key": dsschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"private_key_wo":         computedWOPlaceholder(false),
					"private_key_wo_version": computedWOPlaceholder(true),
					"key_passphrase": dsschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"key_passphrase_wo":         computedWOPlaceholder(false),
					"key_passphrase_wo_version": computedWOPlaceholder(true),
					"fingerprints": dsschema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "SHA256 fingerprints to validate when connecting to the external SFTP server.",
					},
					"prefix": dsschema.StringAttribute{
						Computed:    true,
						Description: "Restrict access to this path.",
					},
					"disable_concurrent_reads": dsschema.BoolAttribute{
						Computed:    true,
						Description: "Concurrent reads are safe to use and disabling them will degrade performance. Some servers automatically delete files once they are downloaded; disable concurrent reads for such servers.",
					},
					"buffer_size": dsschema.Int64Attribute{
						Computed:    true,
						Description: "The buffer size (in MB) to use for uploads/downloads.",
					},
					"equality_check_mode": dsschema.Int64Attribute{
						Computed:    true,
						Description: "Defines how to check if two configs point to the same server (enables renaming between matching configs). 0 = username and endpoint must match (default), 1 = only the endpoint must match.",
					},
					"socks_proxy": dsschema.StringAttribute{
						Computed:    true,
						Description: "The address of the SOCKS proxy server, including schema, host, and port. Examples: socks5://127.0.0.1:1080, socks4://127.0.0.1:1080, socks4a://127.0.0.1:1080. " + enterpriseFeatureNote + ".",
					},
					"socks_username": dsschema.StringAttribute{
						Computed:    true,
						Description: "The optional SOCKS username. " + enterpriseFeatureNote + ".",
					},
					"socks_password": dsschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription + " " + enterpriseFeatureNote + ".",
					},
					"socks_password_wo":         computedWOPlaceholder(false),
					"socks_password_wo_version": computedWOPlaceholder(true),
				},
			},
			"ftpconfig": dsschema.SingleNestedAttribute{
				Computed:    true,
				Description: "Remote FTP server configuration details. " + enterpriseFeatureNote,
				Attributes: map[string]dsschema.Attribute{
					"endpoint": dsschema.StringAttribute{
						Computed:    true,
						Description: "FTP endpoint as host:port.",
					},
					"username": dsschema.StringAttribute{
						Computed:    true,
						Description: "Username for FTP authentication.",
					},
					"password": dsschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"password_wo":         computedWOPlaceholder(false),
					"password_wo_version": computedWOPlaceholder(true),
					"tls_mode": dsschema.Int64Attribute{
						Computed:    true,
						Description: "0 disabled, 1 Explicit, 2 Implicit.",
					},
					"skip_tls_verify": dsschema.BoolAttribute{
						Computed:    true,
						Description: "If true, the TLS certificate of the FTP server is not verified.",
					},
				},
			},
			"httpconfig": dsschema.SingleNestedAttribute{
				Computed:    true,
				Description: "HTTP/S remote filesystem configuration details.",
				Attributes: map[string]dsschema.Attribute{
					"endpoint": dsschema.StringAttribute{
						Computed:    true,
						Description: "HTTP/S endpoint URL. SFTPGo uses this URL as base; for example for the `stat` API, SFTPGo appends `/stat/{name}`.",
					},
					"username": dsschema.StringAttribute{
						Computed:    true,
						Description: "Username for HTTP basic authentication.",
					},
					"password": dsschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"password_wo":         computedWOPlaceholder(false),
					"password_wo_version": computedWOPlaceholder(true),
					"api_key": dsschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"api_key_wo":         computedWOPlaceholder(false),
					"api_key_wo_version": computedWOPlaceholder(true),
					"skip_tls_verify": dsschema.BoolAttribute{
						Computed:    true,
						Description: "If true, the TLS certificate of the HTTP endpoint is not verified. Use with caution.",
					},
					"equality_check_mode": dsschema.Int64Attribute{
						Computed:    true,
						Description: "Defines how to check if two configs point to the same server (enables renaming between matching configs). 0 = username and endpoint must match (default), 1 = only the endpoint must match.",
					},
				},
			},
		},
	}
}

func getSchemaForFilesystem() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Required:    true,
		Description: "Filesystem configuration. This block must be set explicitly: for the default local filesystem pass `filesystem = { provider = 0 }`. Defaults are no longer auto-populated from the server because the block contains write-only attributes.",
		Attributes: map[string]schema.Attribute{
			"provider": schema.Int64Attribute{
				Required:    true,
				Description: "Provider. 0 = local filesystem, 1 = S3 Compatible, 2 = Google Cloud, 3 = Azure Blob, 4 = Local encrypted, 5 = SFTP, 6 = HTTP, 7 = FTP",
				Validators: []validator.Int64{
					int64validator.Between(0, 7),
				},
			},
			"osconfig": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"read_buffer_size": schema.Int64Attribute{
						Optional:    true,
						Description: "Optional read buffer size, as MB, to use for downloads. Omit to disable buffering, that's fine in most use cases.",
						Validators: []validator.Int64{
							int64validator.Between(0, 10),
						},
					},
					"write_buffer_size": schema.Int64Attribute{
						Optional:    true,
						Description: "Optional write buffer size, as MB, to use for uploads. Omit to disable no buffering, that's fine in most use cases.",
						Validators: []validator.Int64{
							int64validator.Between(0, 10),
						},
					},
				},
			},
			"s3config": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "S3 compatible object storage configuration details.",
				Attributes: map[string]schema.Attribute{
					"bucket": schema.StringAttribute{
						Required:    true,
						Description: "S3 bucket name.",
					},
					"region": schema.StringAttribute{
						Optional:    true,
						Description: "S3 region.",
					},
					"access_key": schema.StringAttribute{
						Optional:    true,
						Description: "AWS Access Key ID for authentication. Leave blank when using IAM roles or instance profiles.",
					},
					"access_secret": schema.StringAttribute{
						Optional:           true,
						Sensitive:          true,
						Description:        "Plain text access secret. " + secretDescriptionGeneric + " Mutually exclusive with `access_secret_wo`.",
						DeprecationMessage: legacySecretDeprecation("access_secret"),
						Validators:         []validator.String{conflictsWithWO("access_secret")},
					},
					"access_secret_wo":         writeOnlyAttr("access_secret"),
					"access_secret_wo_version": writeOnlyVersionAttr("access_secret"),
					"sse_customer_key": schema.StringAttribute{
						Optional:           true,
						Sensitive:          true,
						Description:        "Plain text Server-Side encryption key. " + secretDescriptionGeneric + " Mutually exclusive with `sse_customer_key_wo`.",
						DeprecationMessage: legacySecretDeprecation("sse_customer_key"),
						Validators:         []validator.String{conflictsWithWO("sse_customer_key")},
					},
					"sse_customer_key_wo":         writeOnlyAttr("sse_customer_key"),
					"sse_customer_key_wo_version": writeOnlyVersionAttr("sse_customer_key"),
					"key_prefix": schema.StringAttribute{
						Optional:    true,
						Description: `If specified then the SFTPGo user will be restricted to objects starting with the specified prefix. The prefix must not start with "/" and must end with "/"`,
					},
					"role_arn": schema.StringAttribute{
						Optional:    true,
						Description: "Optional IAM Role ARN to assume.",
					},
					"session_token": schema.StringAttribute{
						Optional:    true,
						Description: "Optional Session token that is a part of temporary security credentials provisioned by AWS STS.",
					},
					"endpoint": schema.StringAttribute{
						Optional:    true,
						Description: "The endpoint is generally required for S3 compatible backends. For AWS S3, leave not set to use the default endpoint for the specified region.",
					},
					"storage_class": schema.StringAttribute{
						Optional:    true,
						Description: "The storage class to use when storing objects. Leave not set for default.",
					},
					"acl": schema.StringAttribute{
						Optional:    true,
						Description: "The canned ACL to apply to uploaded objects. Not set means the bucket default.",
					},
					"upload_part_size": schema.Int64Attribute{
						Optional:    true,
						Description: "The buffer size (in MB) to use for multipart uploads. If this value is not set, the default value (5MB) will be used.",
					},
					"upload_concurrency": schema.Int64Attribute{
						Optional:    true,
						Description: "How many parts are uploaded in parallel. Not set means the default (5).",
					},
					"download_part_size": schema.Int64Attribute{
						Optional:    true,
						Description: "The buffer size (in MB) to use for multipart downloads. If this value is not set, the default value (5MB) will be used.",
					},
					"upload_part_max_time": schema.Int64Attribute{
						Optional:    true,
						Description: "The maximum time allowed, in seconds, to upload a single chunk. Not set means no timeout.",
					},
					"download_concurrency": schema.Int64Attribute{
						Optional:    true,
						Description: "How many parts are downloaded in parallel. Not set means the default (5). Ignored for partial downloads.",
					},
					"download_part_max_time": schema.Int64Attribute{
						Optional:    true,
						Description: "The maximum time allowed, in seconds, to download a single chunk. Not set means no timeout. Ignored for partial downloads.",
					},
					"force_path_style": schema.BoolAttribute{
						Optional:    true,
						Description: `If set path-style addressing is used, i.e. http://s3.amazonaws.com/BUCKET/KEY`,
					},
					"skip_tls_verify": schema.BoolAttribute{
						Optional:    true,
						Description: `If set the S3 client accepts any TLS certificate presented by the server and any host name in that certificate. In this mode, TLS is susceptible to man-in-the-middle attacks. This should be used only for testing.`,
					},
					"checksum_algorithm": schema.StringAttribute{
						Optional:    true,
						Description: `Checksum algorithm to compute and send with uploads (PutObject, multipart upload, CopyObject) for end-to-end integrity verification. Leave empty (default) for maximum compatibility with S3-compatible services. Supported values: "crc32", "crc32c", "crc64nvme", "sha1", "sha256". ` + enterpriseFeatureNote + ".",
						Validators: []validator.String{
							stringvalidator.OneOf("", "crc32", "crc32c", "crc64nvme", "sha1", "sha256"),
						},
					},
				},
			},
			"gcsconfig": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Google Cloud Storage configuration details.",
				Attributes: map[string]schema.Attribute{
					"bucket": schema.StringAttribute{
						Required:    true,
						Description: "GCS bucket name.",
					},
					"universe_domain": schema.StringAttribute{
						Optional:    true,
						Description: `The universe domain to use for Google Cloud API requests. If omitted or empty, the default public domain (googleapis.com) is used. Set this value if you need to connect to a custom Google Cloud environment, such as Google Distributed Cloud or a Sovereign Cloud. ` + enterpriseFeatureNote,
					},
					"credentials": schema.StringAttribute{
						Optional:           true,
						Sensitive:          true,
						Description:        "Plain text credentials. " + secretDescriptionGeneric + " Mutually exclusive with `credentials_wo`.",
						DeprecationMessage: legacySecretDeprecation("credentials"),
						Validators:         []validator.String{conflictsWithWO("credentials")},
					},
					"credentials_wo":         writeOnlyAttr("credentials"),
					"credentials_wo_version": writeOnlyVersionAttr("credentials"),
					"automatic_credentials": schema.Int64Attribute{
						Optional:    true,
						Description: "0 = disabled, explicit JSON credentials must be provided (default); 1 = enabled, use Application Default Credentials (ADC) to find credentials.",
					},
					"hns": schema.Int64Attribute{
						Optional:    true,
						Description: "Set to 1 if Hierarchical namespace is enabled for the bucket. " + enterpriseFeatureNote + ".",
					},
					"key_prefix": schema.StringAttribute{
						Optional:    true,
						Description: `If specified then the SFTPGo user will be restricted to objects starting with the specified prefix. The prefix must not start with "/" and must end with "/"`,
					},
					"storage_class": schema.StringAttribute{
						Optional:    true,
						Description: "The storage class to use when storing objects. Leave not set for default.",
					},
					"acl": schema.StringAttribute{
						Optional:    true,
						Description: "The ACL to apply to uploaded objects. Not set means the bucket default.",
					},
					"upload_part_size": schema.Int64Attribute{
						Optional:    true,
						Description: "The buffer size (in MB) to use for multipart uploads. The default value is 16MB. Not set means use the default.",
					},
					"upload_part_max_time": schema.Int64Attribute{
						Optional:    true,
						Description: "The maximum time allowed, in seconds, to upload a single chunk. The default value is 32. Not set means use the default.",
					},
				},
			},
			"azblobconfig": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Azure Blob Storage configuration details.",
				Attributes: map[string]schema.Attribute{
					"container": schema.StringAttribute{
						Optional:    true,
						Description: "Azure Blob Storage container name.",
					},
					"account_name": schema.StringAttribute{
						Optional:    true,
						Description: "Storage account name. Leave blank to use SAS URL.",
					},
					"account_key": schema.StringAttribute{
						Optional:           true,
						Sensitive:          true,
						Description:        "Plain text account key. " + secretDescriptionGeneric + " Mutually exclusive with `account_key_wo`.",
						DeprecationMessage: legacySecretDeprecation("account_key"),
						Validators:         []validator.String{conflictsWithWO("account_key")},
					},
					"account_key_wo":         writeOnlyAttr("account_key"),
					"account_key_wo_version": writeOnlyVersionAttr("account_key"),
					"sas_url": schema.StringAttribute{
						Optional:           true,
						Sensitive:          true,
						Description:        "Plain text SAS URL. " + secretDescriptionGeneric + " Mutually exclusive with `sas_url_wo`.",
						DeprecationMessage: legacySecretDeprecation("sas_url"),
						Validators:         []validator.String{conflictsWithWO("sas_url")},
					},
					"sas_url_wo":         writeOnlyAttr("sas_url"),
					"sas_url_wo_version": writeOnlyVersionAttr("sas_url"),
					"endpoint": schema.StringAttribute{
						Optional:    true,
						Description: `Optional endpoint. Default is "blob.core.windows.net". If you use the emulator the endpoint must include the protocol, for example "http://127.0.0.1:10000".`,
					},
					"key_prefix": schema.StringAttribute{
						Optional:    true,
						Description: `If specified then the SFTPGo user will be restricted to objects starting with the specified prefix. The prefix must not start with "/" and must end with "/"`,
					},
					"upload_part_size": schema.Int64Attribute{
						Optional:    true,
						Description: "The buffer size (in MB) to use for multipart uploads. If this value is not set, the default value (5MB) will be used.",
					},
					"upload_concurrency": schema.Int64Attribute{
						Optional:    true,
						Description: "How many parts are uploaded in parallel. Default: 5.",
					},
					"download_part_size": schema.Int64Attribute{
						Optional:    true,
						Description: "The buffer size (in MB) to use for multipart downloads. If this value is not set, the default value (5MB) will be used.",
					},
					"download_concurrency": schema.Int64Attribute{
						Optional:    true,
						Description: "How many parts are downloaded in parallel. Default: 5.",
					},
					"use_emulator": schema.BoolAttribute{
						Optional:    true,
						Description: "If true, the Azure Storage Emulator (Azurite) is used instead of the cloud service.",
					},
					"access_tier": schema.StringAttribute{
						Optional:    true,
						Description: "Blob Access Tier. Not set means the container default.",
					},
				},
			},
			"cryptconfig": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"passphrase": schema.StringAttribute{
						Optional:           true,
						Sensitive:          true,
						Description:        "Plain text passphrase. " + secretDescriptionGeneric + " Mutually exclusive with `passphrase_wo`.",
						DeprecationMessage: legacySecretDeprecation("passphrase"),
						Validators:         []validator.String{conflictsWithWO("passphrase")},
					},
					"passphrase_wo":         writeOnlyAttr("passphrase"),
					"passphrase_wo_version": writeOnlyVersionAttr("passphrase"),
					"read_buffer_size": schema.Int64Attribute{
						Optional:    true,
						Description: "Optional read buffer size, as MB, to use for downloads. Omit to disable buffering, that's fine in most use cases.",
						Validators: []validator.Int64{
							int64validator.Between(0, 10),
						},
					},
					"write_buffer_size": schema.Int64Attribute{
						Optional:    true,
						Description: "Optional write buffer size, as MB, to use for uploads. Omit to disable buffering, that's fine in most use cases.",
						Validators: []validator.Int64{
							int64validator.Between(0, 10),
						},
					},
				},
			},
			"sftpconfig": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Remote SFTP server configuration details.",
				Attributes: map[string]schema.Attribute{
					"endpoint": schema.StringAttribute{
						Required:    true,
						Description: "SFTP endpoint as host:port. Port is always required.",
						Validators: []validator.String{
							sftpEndPointValidator{},
						},
					},
					"username": schema.StringAttribute{
						Required:    true,
						Description: "Username for SFTP authentication.",
					},
					"password": schema.StringAttribute{
						Optional:           true,
						Sensitive:          true,
						Description:        "Plain text password. " + secretDescriptionGeneric + " Mutually exclusive with `password_wo`.",
						DeprecationMessage: legacySecretDeprecation("password"),
						Validators:         []validator.String{conflictsWithWO("password")},
					},
					"password_wo":         writeOnlyAttr("password"),
					"password_wo_version": writeOnlyVersionAttr("password"),
					"private_key": schema.StringAttribute{
						Optional:           true,
						Sensitive:          true,
						Description:        "Plain text private key. " + secretDescriptionGeneric + " Mutually exclusive with `private_key_wo`.",
						DeprecationMessage: legacySecretDeprecation("private_key"),
						Validators:         []validator.String{conflictsWithWO("private_key")},
					},
					"private_key_wo":         writeOnlyAttr("private_key"),
					"private_key_wo_version": writeOnlyVersionAttr("private_key"),
					"key_passphrase": schema.StringAttribute{
						Optional:           true,
						Sensitive:          true,
						Description:        "Plain text passphrase for the private key. " + secretDescriptionGeneric + " Mutually exclusive with `key_passphrase_wo`.",
						DeprecationMessage: legacySecretDeprecation("key_passphrase"),
						Validators:         []validator.String{conflictsWithWO("key_passphrase")},
					},
					"key_passphrase_wo":         writeOnlyAttr("key_passphrase"),
					"key_passphrase_wo_version": writeOnlyVersionAttr("key_passphrase"),
					"fingerprints": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "SHA256 fingerprints to validate when connecting to the external SFTP server. If not set any host key will be accepted: this is a security risk.",
					},
					"prefix": schema.StringAttribute{
						Required:    true,
						Description: `Similar to a chroot for local filesystem. Example: "/somedir/subdir".`,
					},
					"disable_concurrent_reads": schema.BoolAttribute{
						Optional:    true,
						Description: "Concurrent reads are safe to use and disabling them will degrade performance so they are enabled by default. Some servers automatically delete files once they are downloaded. Using concurrent reads is problematic with such servers.",
					},
					"buffer_size": schema.Int64Attribute{
						Optional:    true,
						Description: "The buffer size (in MB) to use for uploads/downloads. Buffering could improve performance for high latency networks. With buffering enabled upload resume is not supported and a file cannot be opened for both reading and writing at the same time. Not set means disabled.",
					},
					"equality_check_mode": schema.Int64Attribute{
						Optional:    true,
						Description: "Defines how to check if this config points to the same server as another config. By default both the endpoint and the username must match. 1 means that only the endpoint must match. If different configs point to the same server the renaming between the fs configs is allowed.",
					},
					"socks_proxy": schema.StringAttribute{
						Optional:    true,
						Description: "The address of the SOCKS proxy server, including schema, host, and port. Examples: socks5://127.0.0.1:1080, socks4://127.0.0.1:1080, socks4a://127.0.0.1:1080. " + enterpriseFeatureNote + ".",
					},
					"socks_username": schema.StringAttribute{
						Optional:    true,
						Description: "The optional SOCKS username. " + enterpriseFeatureNote + ".",
					},
					"socks_password": schema.StringAttribute{
						Optional:           true,
						Sensitive:          true,
						Description:        "Plain text SOCKS password. " + secretDescriptionGeneric + " Mutually exclusive with `socks_password_wo`. " + enterpriseFeatureNote + ".",
						DeprecationMessage: legacySecretDeprecation("socks_password"),
						Validators:         []validator.String{conflictsWithWO("socks_password")},
					},
					"socks_password_wo":         writeOnlyAttr("socks_password"),
					"socks_password_wo_version": writeOnlyVersionAttr("socks_password"),
				},
			},
			"ftpconfig": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Remote FTP server configuration details. " + enterpriseFeatureNote,
				Attributes: map[string]schema.Attribute{
					"endpoint": schema.StringAttribute{
						Required:    true,
						Description: "FTP endpoint as host:port.",
					},
					"username": schema.StringAttribute{
						Required:    true,
						Description: "Username for FTP authentication.",
					},
					"password": schema.StringAttribute{
						Optional:           true,
						Sensitive:          true,
						Description:        "Plain text password. " + secretDescriptionGeneric + " Mutually exclusive with `password_wo`.",
						DeprecationMessage: legacySecretDeprecation("password"),
						Validators:         []validator.String{conflictsWithWO("password")},
					},
					"password_wo":         writeOnlyAttr("password"),
					"password_wo_version": writeOnlyVersionAttr("password"),
					"tls_mode": schema.Int64Attribute{
						Optional:    true,
						Description: "0 disabled, 1 Explicit, 2 Implicit.",
					},
					"skip_tls_verify": schema.BoolAttribute{
						Optional:    true,
						Description: "If true, the TLS certificate of the FTP server is not verified.",
					},
				},
			},
			"httpconfig": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "HTTP/S remote filesystem configuration details.",
				Attributes: map[string]schema.Attribute{
					"endpoint": schema.StringAttribute{
						Required:    true,
						Description: "HTTP/S endpoint URL. SFTPGo uses this URL as base; for example for the `stat` API, SFTPGo appends `/stat/{name}`.",
					},
					"username": schema.StringAttribute{
						Optional:    true,
						Description: "Username for HTTP basic authentication.",
					},
					"password": schema.StringAttribute{
						Optional:           true,
						Sensitive:          true,
						Description:        "Plain text password. " + secretDescriptionGeneric + " Mutually exclusive with `password_wo`.",
						DeprecationMessage: legacySecretDeprecation("password"),
						Validators:         []validator.String{conflictsWithWO("password")},
					},
					"password_wo":         writeOnlyAttr("password"),
					"password_wo_version": writeOnlyVersionAttr("password"),
					"api_key": schema.StringAttribute{
						Optional:           true,
						Sensitive:          true,
						Description:        "Plain text API key. " + secretDescriptionGeneric + " Mutually exclusive with `api_key_wo`.",
						DeprecationMessage: legacySecretDeprecation("api_key"),
						Validators:         []validator.String{conflictsWithWO("api_key")},
					},
					"api_key_wo":         writeOnlyAttr("api_key"),
					"api_key_wo_version": writeOnlyVersionAttr("api_key"),
					"skip_tls_verify": schema.BoolAttribute{
						Optional:    true,
						Description: "If true, the TLS certificate of the HTTP endpoint is not verified. Use with caution.",
					},
					"equality_check_mode": schema.Int64Attribute{
						Optional:    true,
						Description: "Defines how to check if two configs point to the same server (enables renaming between matching configs). 0 = username and endpoint must match (default), 1 = only the endpoint must match.",
					},
				},
			},
		},
	}
}

func getComputedSchemaForVirtualFolders() dsschema.ListNestedAttribute {
	return dsschema.ListNestedAttribute{
		Computed:    true,
		Description: "Virtual folder.",
		NestedObject: dsschema.NestedAttributeObject{
			Attributes: map[string]dsschema.Attribute{
				"name": dsschema.StringAttribute{
					Computed:    true,
					Description: "Unique folder name",
				},
				"mapped_path": dsschema.StringAttribute{
					Computed:    true,
					Description: "Absolute path to a local directory. This is the folder root path for local storage provider. For non-local filesystems it will store temporary files.",
				},
				"virtual_path": dsschema.StringAttribute{
					Computed:    true,
					Description: "The folder will be available on this path.",
				},
				"description": dsschema.StringAttribute{
					Computed:    true,
					Description: "Optional description.",
				},
				"quota_size": dsschema.Int64Attribute{
					Computed:    true,
					Description: "Maximum size allowed as bytes. Not set means unlimited, -1 included in user quota",
				},
				"quota_files": dsschema.Int64Attribute{
					Computed:    true,
					Description: "Maximum number of files allowed. Not set means unlimited, -1 included in user quota",
				},
				"used_quota_size": dsschema.Int64Attribute{
					Computed:    true,
					Description: "Used quota as bytes.",
				},
				"used_quota_files": dsschema.Int64Attribute{
					Computed:    true,
					Description: "Used quota as number of files.",
				},
				"last_quota_update": dsschema.Int64Attribute{
					Computed:    true,
					Description: "Last quota update as unix timestamp in milliseconds",
				},
				"filesystem": getComputedSchemaForFilesystem(),
			},
		},
	}
}

func getSchemaForVirtualFolders() schema.SetNestedAttribute {
	return schema.SetNestedAttribute{
		Optional: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Required:    true,
					Description: "Unique folder name",
				},
				"virtual_path": schema.StringAttribute{
					Required:    true,
					Description: "The folder will be available on this path.",
				},
				"quota_size": schema.Int64Attribute{
					Required:    true,
					Description: "Maximum size allowed as bytes. Not set means unlimited, -1 included in user quota",
				},
				"quota_files": schema.Int64Attribute{
					Required:    true,
					Description: "Maximum number of files allowed. Not set means unlimited, -1 included in user quota",
				},
				"mapped_path": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Absolute path to a local directory. This is the folder root path for local storage provider. For non-local filesystems it will store temporary files.",
				},
				"description": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Optional description.",
				},
				"used_quota_size": schema.Int64Attribute{
					Computed:    true,
					Optional:    true,
					Description: "Used quota as bytes.",
				},
				"used_quota_files": schema.Int64Attribute{
					Computed:    true,
					Optional:    true,
					Description: "Used quota as number of files.",
				},
				"last_quota_update": schema.Int64Attribute{
					Computed:    true,
					Optional:    true,
					Description: "Last quota update as unix timestamp in milliseconds",
				},
				"filesystem": getComputedSchemaForFilesystem(),
			},
		},
	}
}

func getComputedSchemaForUserFilters(isGroup bool) dsschema.SingleNestedAttribute {
	result := dsschema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]dsschema.Attribute{
			"allowed_ip": dsschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: `Only connections from these IP/Mask are allowed. IP/Mask must be in CIDR notation as defined in RFC 4632 and RFC 4291 for example "192.0.2.0/24" or "2001:db8::/32"`,
			},
			"denied_ip": dsschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Connections from these IP/Mask are allowed. Denied rules will be evaluated before allowed ones.",
			},
			"denied_login_methods": dsschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Disabled login methods.",
			},
			"denied_protocols": dsschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Disabled protocols.",
			},
			"file_patterns": dsschema.ListNestedAttribute{
				Computed:    true,
				Description: `Filters based on shell patterns.`,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"path": dsschema.StringAttribute{
							Computed:    true,
							Description: "Virtual path, if no other specific filter is defined, the filter applies for sub directories too.",
						},
						"allowed_patterns": dsschema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "Files/directories with these, case insensitive, patterns are allowed. Allowed file patterns are evaluated before the denied ones.",
						},
						"denied_patterns": dsschema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "Files/directories with these, case insensitive, patterns are not allowed.",
						},
						"deny_policy": dsschema.Int64Attribute{
							Computed:    true,
							Description: "Set to 1 to hide denied files/directories in directory listing.",
						},
					},
				},
			},
			"max_upload_file_size": dsschema.Int64Attribute{
				Computed:    true,
				Description: "Max size allowed for a single upload. Unset means no limit.",
			},
			"tls_username": dsschema.StringAttribute{
				Computed:    true,
				Description: `TLS certificate attribute to use as username. For FTP clients it must match the name provided using the "USER" command. For WebDAV, if no username is provided, the CN will be used as username. For WebDAV clients it must match the implicit or provided username.`,
			},
			"external_auth_disabled": dsschema.BoolAttribute{
				Computed:    true,
				Description: "If set, external auth hook will not be executed.",
			},
			"pre_login_disabled": dsschema.BoolAttribute{
				Computed:    true,
				Description: "If set, external pre-login hook will not be executed.",
			},
			"check_password_disabled": dsschema.BoolAttribute{
				Computed:    true,
				Description: "If set, check password hook will not be executed.",
			},
			"disable_fs_checks": dsschema.BoolAttribute{
				Computed:    true,
				Description: "Disable checks for existence and automatic creation of home directory and virtual folders after user login.",
			},
			"web_client": dsschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Web Client/user REST API restrictions.",
			},
			"allow_api_key_auth": dsschema.BoolAttribute{
				Computed:    true,
				Description: "If set, API Key authentication is allowed.",
			},
			"user_type": dsschema.StringAttribute{
				Computed:    true,
				Description: "Hint for authentication plugins.",
			},
			"bandwidth_limits": dsschema.ListNestedAttribute{
				Computed:    true,
				Description: "Per-source bandwidth limits.",
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"sources": dsschema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: `Source networks in CIDR notation as defined in RFC 4632 and RFC 4291 for example "192.0.2.0/24" or "2001:db8::/32". The limit applies if the defined networks contain the client IP.`,
						},
						"upload_bandwidth": dsschema.Int64Attribute{
							Computed:    true,
							Description: "Maximum upload bandwidth as KB/s.",
						},
						"download_bandwidth": dsschema.Int64Attribute{
							Computed:    true,
							Description: "Maximum download bandwidth as KB/s.",
						},
					},
				},
			},
			"external_auth_cache_time": dsschema.Int64Attribute{
				Computed:    true,
				Description: "Defines the cache time, in seconds, for users authenticated using an external auth hook. Not set means no cache.",
			},
			"start_directory": dsschema.StringAttribute{
				Computed:    true,
				Description: `Alternate starting directory. If not set, the default is "/". This option is supported for SFTP/SCP, FTP and HTTP (WebClient/REST API) protocols. Relative paths will use this directory as base.`,
			},
			"two_factor_protocols": dsschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Defines protocols that require two factor authentication",
			},
			"ftp_security": dsschema.Int64Attribute{
				Computed:    true,
				Description: "FTP security mode. Set to 1 to require TLS for both data and control connection.",
			},
			"is_anonymous": dsschema.BoolAttribute{
				Computed:    true,
				Description: `If enabled the user can login with any password or no password at all. Anonymous users are supported for FTP and WebDAV protocols and permissions will be automatically set to "list" and "download" (read only)`,
			},
			"default_shares_expiration": dsschema.Int64Attribute{
				Computed:    true,
				Description: "Default expiration for newly created shares as number of days. Not set means no default expiration.",
			},
			"max_shares_expiration": dsschema.Int64Attribute{
				Computed:    true,
				Description: "Maximum allowed expiration, as a number of days, when a user creates or updates a share. Not set means that non-expiring shares are allowed.",
			},
			"password_expiration": dsschema.Int64Attribute{
				Computed:    true,
				Description: "The password expires after the defined number of days. Not set means no expiration",
			},
			"password_strength": dsschema.Int64Attribute{
				Computed:    true,
				Description: "Minimum password entropy enforced when a password is set. Not set means no per-user value: the primary group's password_strength is used, otherwise the system-level data_provider.password_validation default. A non-zero value overrides the system default (the override may be less strict than the system default). Values in the 50-70 range are suggested for common use cases.",
			},
			"password_policy": dsschema.SingleNestedAttribute{
				Computed:    true,
				Description: "Static password complexity requirements. Each field overrides the primary group's corresponding value, which in turn overrides the system-level data_provider.password_validation default (overrides may be less strict than the system default). Whenever possible, prefer using the entropy-based approach provided by password_strength. " + enterpriseFeatureNote,
				Attributes: map[string]dsschema.Attribute{
					"length": dsschema.Int64Attribute{
						Optional:    true,
						Description: "Minimum password length.",
					},
					"uppers": dsschema.Int64Attribute{
						Optional:    true,
						Description: "Minimum number of uppercase characters required.",
					},
					"lowers": dsschema.Int64Attribute{
						Optional:    true,
						Description: "Minimum number of lowercase characters required.",
					},
					"digits": dsschema.Int64Attribute{
						Optional:    true,
						Description: "Minimum number of digit characters required.",
					},
					"specials": dsschema.Int64Attribute{
						Optional:    true,
						Description: "Minimum number of special characters required.",
					},
				},
			},
			"access_time": dsschema.ListNestedAttribute{
				Computed:    true,
				Description: "Time periods in which access is allowed",
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"day_of_week": dsschema.Int64Attribute{
							Computed:    true,
							Description: "Day of week, 0 Sunday, 6 Saturday",
						},
						"from": dsschema.StringAttribute{
							Computed:    true,
							Description: "Start time in HH:MM format",
						},
						"to": dsschema.StringAttribute{
							Computed:    true,
							Description: "End time in HH:MM format",
						},
					},
				},
			},
			"enforce_secure_algorithms": dsschema.BoolAttribute{
				Computed:    true,
				Description: "If enabled, only secure algorithms are allowed. This setting is currently enforced for SSH/SFTP. " + enterpriseFeatureNote + ".",
			},
			"denied_share_paths": dsschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Virtual paths that cannot be shared. Shares for any listed path and its sub-paths are rejected. " + enterpriseFeatureNote + ".",
			},
			"denied_share_scopes": dsschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Share scopes that users are not allowed to use. Valid values: read, write, read_write. If all scopes are denied, sharing is completely disabled. " + enterpriseFeatureNote + ".",
			},
		},
	}
	if isGroup {
		result.Attributes["share_policy"] = dsschema.SingleNestedAttribute{
			Computed:    true,
			Description: "Share access rules. " + enterpriseFeatureNote,
			Attributes: map[string]dsschema.Attribute{
				"permissions": dsschema.Int64Attribute{
					Computed:    true,
					Description: "Bitmask representing the share permissions. 1 = Read, 2 = Write, 4 = Delete. Example: Read + Write is 3 (1 + 2).",
				},
				"mode": dsschema.Int64Attribute{
					Computed:    true,
					Description: "Policy mode. 1 = suggested (the group policy is pre-selected but can be removed by the user), 2 = enforced (the group policy is mandatory and cannot be changed by the user).",
				},
			},
		}
		return result
	}
	result.Attributes["require_password_change"] = dsschema.BoolAttribute{
		Computed:    true,
		Description: "If set, user must change their password from WebClient/REST API at next login.",
	}
	result.Attributes["tls_certs"] = dsschema.ListAttribute{
		ElementType: types.StringType,
		Computed:    true,
		Description: "TLS certificates for mutual authentication. If provided will be checked before TLS username.",
	}
	result.Attributes["additional_emails"] = dsschema.ListAttribute{
		ElementType: types.StringType,
		Computed:    true,
		Description: "Additional email addresses.",
	}
	result.Attributes["custom1"] = dsschema.StringAttribute{
		Computed:    true,
		Description: `An extra placeholder value available for use in group configurations. It can be referenced as %custom1%. Deprecated: use custom_placeholders instead. ` + enterpriseFeatureNote + ".",
	}
	result.Attributes["custom_placeholders"] = dsschema.ListAttribute{
		ElementType: types.StringType,
		Computed:    true,
		Description: "List of extra placeholders available for use in group configurations. Each placeholder can be referenced as %custom1%, %custom2%, and so on. " + enterpriseFeatureNote + ".",
	}
	return result
}

func getSchemaForUserFilters(isGroup bool) schema.SingleNestedAttribute {
	result := schema.SingleNestedAttribute{
		Optional: true,
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"allowed_ip": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: `Only connections from these IP/Mask are allowed. IP/Mask must be in CIDR notation as defined in RFC 4632 and RFC 4291 for example "192.0.2.0/24" or "2001:db8::/32"`,
			},
			"denied_ip": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Connections from these IP/Mask are allowed. Denied rules will be evaluated before allowed ones.",
			},
			"denied_login_methods": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: `Disabled login methods. Valid values: "publickey", "password", "password-over-SSH", "keyboard-interactive", "publickey+password", "publickey+keyboard-interactive", "TLSCertificate", "TLSCertificate+password"`,
				Validators: []validator.List{
					listvalidator.UniqueValues(),
					listvalidator.ValueStringsAre(stringvalidator.OneOf("publickey", "password", "password-over-SSH",
						"keyboard-interactive", "publickey+password", "publickey+keyboard-interactive", "TLSCertificate",
						"TLSCertificate+password")),
				},
			},
			"denied_protocols": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: `Disabled protocols. Valid values: SSH, FTP, DAV, HTTP`,
				Validators: []validator.List{
					listvalidator.UniqueValues(),
					listvalidator.ValueStringsAre(stringvalidator.OneOf("SSH", "FTP", "DAV", "HTTP")),
				},
			},
			"file_patterns": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							Required:    true,
							Description: "Virtual path, if no other specific filter is defined, the filter applies for sub directories too.",
						},
						"allowed_patterns": schema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "Files/directories with these, case insensitive, patterns are allowed. Allowed file patterns are evaluated before the denied ones.",
						},
						"denied_patterns": schema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "Files/directories with these, case insensitive, patterns are not allowed.",
						},
						"deny_policy": schema.Int64Attribute{
							Optional:    true,
							Description: "Set to 1 to hide denied files/directories in directory listing.",
						},
					},
				},
			},
			"max_upload_file_size": schema.Int64Attribute{
				Optional:    true,
				Description: "Max size allowed for a single upload. Unset means no limit.",
			},
			"tls_username": schema.StringAttribute{
				Optional:    true,
				Description: `TLS certificate attribute to use as username. For FTP clients it must match the name provided using the "USER" command. For WebDAV, if no username is provided, the CN will be used as username. For WebDAV clients it must match the implicit or provided username.`,
				Validators: []validator.String{
					stringvalidator.OneOf("CommonName"),
				},
			},
			"external_auth_disabled": schema.BoolAttribute{
				Optional:    true,
				Description: "If set, external auth hook will not be executed.",
			},
			"pre_login_disabled": schema.BoolAttribute{
				Optional:    true,
				Description: "If set, external pre-login hook will not be executed.",
			},
			"check_password_disabled": schema.BoolAttribute{
				Optional:    true,
				Description: "If set, check password hook will not be executed.",
			},
			"disable_fs_checks": schema.BoolAttribute{
				Optional:    true,
				Description: "Disable checks for existence and automatic creation of home directory and virtual folders after user login.",
			},
			"web_client": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: fmt.Sprintf("Web Client/user REST API restrictions. Valid values: %s. Only available in the Enterprise version: %s",
					strings.Join(client.WebClientOptions, ", "), strings.Join(client.EnterpriseWebClientOptions, ", ")),
				Validators: []validator.List{
					listvalidator.UniqueValues(),
					listvalidator.ValueStringsAre(stringvalidator.OneOf(client.WebClientOptions...)),
				},
			},
			"allow_api_key_auth": schema.BoolAttribute{
				Optional:    true,
				Description: "If set, API Key authentication is allowed.",
			},
			"user_type": schema.StringAttribute{
				Optional:    true,
				Description: "Hint for authentication plugins. Valid values: LDAPUser, OSUser",
				Validators: []validator.String{
					stringvalidator.OneOf("LDAPUser", "OSUser"),
				},
			},
			"bandwidth_limits": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"sources": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
							Description: `Source networks in CIDR notation as defined in RFC 4632 and RFC 4291 for example "192.0.2.0/24" or "2001:db8::/32". The limit applies if the defined networks contain the client IP.`,
						},
						"upload_bandwidth": schema.Int64Attribute{
							Optional:    true,
							Description: "Maximum upload bandwidth as KB/s.",
						},
						"download_bandwidth": schema.Int64Attribute{
							Optional:    true,
							Description: "Maximum download bandwidth as KB/s.",
						},
					},
				},
			},
			"external_auth_cache_time": schema.Int64Attribute{
				Optional:    true,
				Description: "Defines the cache time, in seconds, for users authenticated using an external auth hook. Not set means no cache.",
			},
			"start_directory": schema.StringAttribute{
				Optional:    true,
				Description: `Alternate starting directory. If not set, the default is "/". This option is supported for SFTP/SCP, FTP and HTTP (WebClient/REST API) protocols. Relative paths will use this directory as base.`,
			},
			"two_factor_protocols": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Defines protocols that require two factor authentication. Valid values: SSH, FTP, HTTP",
				Validators: []validator.List{
					listvalidator.UniqueValues(),
					listvalidator.ValueStringsAre(stringvalidator.OneOf("SSH", "FTP", "HTTP")),
				},
			},
			"ftp_security": schema.Int64Attribute{
				Optional:    true,
				Description: "FTP security mode. Set to 1 to require TLS for both data and control connection.",
			},
			"is_anonymous": schema.BoolAttribute{
				Optional:    true,
				Description: `If enabled the user can login with any password or no password at all. Anonymous users are supported for FTP and WebDAV protocols and permissions will be automatically set to "list" and "download" (read only)`,
			},
			"default_shares_expiration": schema.Int64Attribute{
				Optional:    true,
				Description: "Default expiration for newly created shares as number of days. Not set means no default expiration.",
			},
			"max_shares_expiration": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum allowed expiration, as a number of days, when a user creates or updates a share. Not set means that non-expiring shares are allowed.",
			},
			"password_expiration": schema.Int64Attribute{
				Optional:    true,
				Description: "The password expires after the defined number of days. Not set means no expiration",
			},
			"password_strength": schema.Int64Attribute{
				Optional:    true,
				Description: "Minimum password entropy enforced when a password is set. Not set means no per-user value: the primary group's password_strength is used, otherwise the system-level data_provider.password_validation default. A non-zero value overrides the system default (the override may be less strict than the system default). Values in the 50-70 range are suggested for common use cases.",
			},
			"password_policy": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Static password complexity requirements. Each field overrides the primary group's corresponding value, which in turn overrides the system-level data_provider.password_validation default (overrides may be less strict than the system default). Whenever possible, prefer using the entropy-based approach provided by password_strength. " + enterpriseFeatureNote,
				Attributes: map[string]schema.Attribute{
					"length": schema.Int64Attribute{
						Optional:    true,
						Description: "Minimum password length.",
					},
					"uppers": schema.Int64Attribute{
						Optional:    true,
						Description: "Minimum number of uppercase characters required.",
					},
					"lowers": schema.Int64Attribute{
						Optional:    true,
						Description: "Minimum number of lowercase characters required.",
					},
					"digits": schema.Int64Attribute{
						Optional:    true,
						Description: "Minimum number of digit characters required.",
					},
					"specials": schema.Int64Attribute{
						Optional:    true,
						Description: "Minimum number of special characters required.",
					},
				},
			},
			"access_time": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Time periods in which access is allowed",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"day_of_week": schema.Int64Attribute{
							Required:    true,
							Description: "Day of week, 0 Sunday, 6 Saturday",
							Validators: []validator.Int64{
								int64validator.Between(0, 6),
							},
						},
						"from": schema.StringAttribute{
							Required:    true,
							Description: "Start time in HH:MM format",
						},
						"to": schema.StringAttribute{
							Required:    true,
							Description: "End time in HH:MM format",
						},
					},
				},
			},
			"enforce_secure_algorithms": schema.BoolAttribute{
				Optional:    true,
				Description: "If enabled, only secure algorithms are allowed. This setting is currently enforced for SSH/SFTP. " + enterpriseFeatureNote + ".",
			},
			"denied_share_paths": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Virtual paths that cannot be shared. If a path is denied, shares for that path and any sub-path are rejected. " + enterpriseFeatureNote + ".",
				Validators: []validator.List{
					listvalidator.UniqueValues(),
				},
			},
			"denied_share_scopes": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Share scopes that users are not allowed to use. Valid values: read, write, read_write. If all scopes are denied, sharing is completely disabled. " + enterpriseFeatureNote + ".",
				Validators: []validator.List{
					listvalidator.UniqueValues(),
					listvalidator.ValueStringsAre(stringvalidator.OneOf("read", "write", "read_write")),
				},
			},
		},
	}
	if isGroup {
		result.Attributes["share_policy"] = schema.SingleNestedAttribute{
			Optional:    true,
			Description: "Share access rules. " + enterpriseFeatureNote,
			Attributes: map[string]schema.Attribute{
				"permissions": schema.Int64Attribute{
					Optional: true,
					MarkdownDescription: "Bitmask of permissions. Sum the values to combine permissions.\n\n" +
						"Supported values:\n" +
						"* `0`: None\n" +
						"* `1`: Read\n" +
						"* `2`: Write\n" +
						"* `4`: Delete\n" +
						"* `7`: All",
				},
				"mode": schema.Int64Attribute{
					Optional: true,
					MarkdownDescription: "Defines how the default share policy is applied.\n\n" +
						"Supported values:\n" +
						"* `1` (Suggested): The group is pre-selected but removable.\n" +
						"* `2` (Enforced): The association is mandatory.",
					Validators: []validator.Int64{
						int64validator.OneOf(1, 2),
					},
				},
			},
		}
		return result
	}
	result.Attributes["require_password_change"] = schema.BoolAttribute{
		Optional:    true,
		Description: "If set, user must change their password from WebClient/REST API at next login.",
	}
	result.Attributes["tls_certs"] = schema.ListAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Description: "TLS certificates for mutual authentication. If provided will be checked before TLS username.",
		Validators: []validator.List{
			listvalidator.UniqueValues(),
		},
	}
	result.Attributes["additional_emails"] = schema.ListAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Description: "Additional email addresses.",
		Validators: []validator.List{
			listvalidator.UniqueValues(),
		},
	}
	result.Attributes["custom1"] = schema.StringAttribute{
		Optional:    true,
		Description: `An extra placeholder value available for use in group configurations. It can be referenced as %custom1%. ` + enterpriseFeatureNote + ".",
	}
	result.Attributes["custom_placeholders"] = schema.ListAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Description: "List of extra placeholders available for use in group configurations. Each placeholder can be referenced as %custom1%, %custom2%, and so on. " + enterpriseFeatureNote + ".",
	}
	return result
}

// conflictsWithWO returns a validator that fails if the current attribute is
// set together with its `<name>_wo` sibling within the same nested object.
func conflictsWithWO(name string) validator.String {
	return stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(name + "_wo"))
}

// computedWOPlaceholder builds a Computed-only placeholder attribute for
// data sources. Values are always null in data source reads; it exists only
// so the data source schema matches the model struct shared with the
// resource.
func computedWOPlaceholder(isVersion bool) dsschema.Attribute {
	desc := "Write-only attribute placeholder. Always null in data source reads."
	if isVersion {
		desc = "Write-only trigger attribute placeholder. Always null in data source reads."
	}
	return dsschema.StringAttribute{
		Computed:    true,
		Description: desc,
	}
}

// writeOnlyAttr builds the `<name>_wo` write-only schema attribute. The
// write-only value and its sibling `<name>_wo_version` are mutually required:
// the version attribute is the only thing that lives in state, so setting the
// secret without the version would leave rotations undetectable.
func writeOnlyAttr(name string) schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Sensitive:   true,
		WriteOnly:   true,
		Description: "Write-only variant of `" + name + "`. " + writeOnlyDescriptionGeneric,
		Validators: []validator.String{
			stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName(name + "_wo_version")),
		},
	}
}

// writeOnlyVersionAttr builds the `<name>_wo_version` trigger attribute.
func writeOnlyVersionAttr(name string) schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Description: "Trigger attribute for `" + name + "_wo`. " + writeOnlyVersionDescGeneric,
		Validators: []validator.String{
			stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName(name + "_wo")),
		},
	}
}

// legacySecretDeprecation is the deprecation message appended to legacy
// sensitive attributes to nudge users toward their write-only counterparts.
func legacySecretDeprecation(name string) string {
	return "Storing secrets in plan/state is discouraged. Prefer the write-only variant `" + name + "_wo` (Terraform 1.11+)."
}

func preserveFsConfigPlanFields(ctx context.Context, fsPlan, fsState filesystem) (types.Object, diag.Diagnostics) {
	switch sdk.FilesystemProvider(fsState.Provider.ValueInt64()) {
	case sdk.S3FilesystemProvider:
		if fsPlan.S3Config != nil {
			fsState.S3Config.AccessSecret = fsPlan.S3Config.AccessSecret
			fsState.S3Config.SSECustomerKey = fsPlan.S3Config.SSECustomerKey
			fsState.S3Config.AccessSecretWOVersion = fsPlan.S3Config.AccessSecretWOVersion
			fsState.S3Config.SSECustomerKeyWOVersion = fsPlan.S3Config.SSECustomerKeyWOVersion
		}
	case sdk.GCSFilesystemProvider:
		if fsPlan.GCSConfig != nil {
			fsState.GCSConfig.Credentials = fsPlan.GCSConfig.Credentials
			fsState.GCSConfig.CredentialsWOVersion = fsPlan.GCSConfig.CredentialsWOVersion
		}
	case sdk.AzureBlobFilesystemProvider:
		if fsPlan.AzBlobConfig != nil {
			fsState.AzBlobConfig.AccountKey = fsPlan.AzBlobConfig.AccountKey
			fsState.AzBlobConfig.SASURL = fsPlan.AzBlobConfig.SASURL
			fsState.AzBlobConfig.AccountKeyWOVersion = fsPlan.AzBlobConfig.AccountKeyWOVersion
			fsState.AzBlobConfig.SASURLWOVersion = fsPlan.AzBlobConfig.SASURLWOVersion
		}
	case sdk.CryptedFilesystemProvider:
		if fsPlan.CryptConfig != nil {
			fsState.CryptConfig.Passphrase = fsPlan.CryptConfig.Passphrase
			fsState.CryptConfig.PassphraseWOVersion = fsPlan.CryptConfig.PassphraseWOVersion
		}
	case sdk.SFTPFilesystemProvider:
		if fsPlan.SFTPConfig != nil {
			fsState.SFTPConfig.Password = fsPlan.SFTPConfig.Password
			fsState.SFTPConfig.PrivateKey = fsPlan.SFTPConfig.PrivateKey
			fsState.SFTPConfig.KeyPassphrase = fsPlan.SFTPConfig.KeyPassphrase
			fsState.SFTPConfig.SocksPassword = fsPlan.SFTPConfig.SocksPassword
			fsState.SFTPConfig.PasswordWOVersion = fsPlan.SFTPConfig.PasswordWOVersion
			fsState.SFTPConfig.PrivateKeyWOVersion = fsPlan.SFTPConfig.PrivateKeyWOVersion
			fsState.SFTPConfig.KeyPassphraseWOVersion = fsPlan.SFTPConfig.KeyPassphraseWOVersion
			fsState.SFTPConfig.SocksPasswordWOVersion = fsPlan.SFTPConfig.SocksPasswordWOVersion
		}
	case sdk.HTTPFilesystemProvider:
		if fsPlan.HTTPConfig != nil {
			fsState.HTTPConfig.Password = fsPlan.HTTPConfig.Password
			fsState.HTTPConfig.APIKey = fsPlan.HTTPConfig.APIKey
			fsState.HTTPConfig.PasswordWOVersion = fsPlan.HTTPConfig.PasswordWOVersion
			fsState.HTTPConfig.APIKeyWOVersion = fsPlan.HTTPConfig.APIKeyWOVersion
		}
	case client.FTPFilesystemProvider:
		if fsPlan.FTPConfig != nil {
			fsState.FTPConfig.Password = fsPlan.FTPConfig.Password
			fsState.FTPConfig.PasswordWOVersion = fsPlan.FTPConfig.PasswordWOVersion
		}
	}

	return types.ObjectValueFrom(ctx, fsState.getTFAttributes(), fsState)
}

// applyFsConfigWriteOnly copies the write-only secret values from the
// configuration's filesystem block into the plan's filesystem block so the
// plan passed to toSFTPGo carries them. Only the nested config matching the
// configured provider is touched; the rest is left nil.
func applyFsConfigWriteOnly(ctx context.Context, fsConfig, fsPlan filesystem) (types.Object, diag.Diagnostics) {
	if fsPlan.S3Config != nil && fsConfig.S3Config != nil {
		fsPlan.S3Config.AccessSecretWO = fsConfig.S3Config.AccessSecretWO
		fsPlan.S3Config.SSECustomerKeyWO = fsConfig.S3Config.SSECustomerKeyWO
	}
	if fsPlan.GCSConfig != nil && fsConfig.GCSConfig != nil {
		fsPlan.GCSConfig.CredentialsWO = fsConfig.GCSConfig.CredentialsWO
	}
	if fsPlan.AzBlobConfig != nil && fsConfig.AzBlobConfig != nil {
		fsPlan.AzBlobConfig.AccountKeyWO = fsConfig.AzBlobConfig.AccountKeyWO
		fsPlan.AzBlobConfig.SASURLWO = fsConfig.AzBlobConfig.SASURLWO
	}
	if fsPlan.CryptConfig != nil && fsConfig.CryptConfig != nil {
		fsPlan.CryptConfig.PassphraseWO = fsConfig.CryptConfig.PassphraseWO
	}
	if fsPlan.SFTPConfig != nil && fsConfig.SFTPConfig != nil {
		fsPlan.SFTPConfig.PasswordWO = fsConfig.SFTPConfig.PasswordWO
		fsPlan.SFTPConfig.PrivateKeyWO = fsConfig.SFTPConfig.PrivateKeyWO
		fsPlan.SFTPConfig.KeyPassphraseWO = fsConfig.SFTPConfig.KeyPassphraseWO
		fsPlan.SFTPConfig.SocksPasswordWO = fsConfig.SFTPConfig.SocksPasswordWO
	}
	if fsPlan.FTPConfig != nil && fsConfig.FTPConfig != nil {
		fsPlan.FTPConfig.PasswordWO = fsConfig.FTPConfig.PasswordWO
	}
	if fsPlan.HTTPConfig != nil && fsConfig.HTTPConfig != nil {
		fsPlan.HTTPConfig.PasswordWO = fsConfig.HTTPConfig.PasswordWO
		fsPlan.HTTPConfig.APIKeyWO = fsConfig.HTTPConfig.APIKeyWO
	}
	return types.ObjectValueFrom(ctx, fsPlan.getTFAttributes(), fsPlan)
}

// preserveUnchangedFsSecrets mutates fsOut — a client.Filesystem already
// populated by the resource's toSFTPGo — to substitute a "preserve" sentinel
// for every secret whose rotation was not requested (i.e. both the legacy
// field and the corresponding write-only version in plan equal the ones in
// the previous state). The sentinel has Status=Redacted which SFTPGo's
// updateEncryptedSecrets treats as IsNotPlainAndNotEmpty → keep current
// server-side value. No-op when the provider is local (no secrets to
// preserve).
func preserveUnchangedFsSecrets(ctx context.Context, fsOut *client.Filesystem, planFs, prevFs types.Object) diag.Diagnostics {
	var plan, prev filesystem
	diags := planFs.As(ctx, &plan, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return diags
	}
	diags = prevFs.As(ctx, &prev, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return diags
	}

	switch sdk.FilesystemProvider(plan.Provider.ValueInt64()) {
	case sdk.S3FilesystemProvider:
		if plan.S3Config == nil || prev.S3Config == nil {
			return nil
		}
		if shouldPreserveSecret(plan.S3Config.AccessSecret, prev.S3Config.AccessSecret,
			plan.S3Config.AccessSecretWOVersion, prev.S3Config.AccessSecretWOVersion) {
			fsOut.S3Config.AccessSecret = preserveSecretSentinel()
		}
		if shouldPreserveSecret(plan.S3Config.SSECustomerKey, prev.S3Config.SSECustomerKey,
			plan.S3Config.SSECustomerKeyWOVersion, prev.S3Config.SSECustomerKeyWOVersion) {
			fsOut.S3Config.SSECustomerKey = preserveSecretSentinel()
		}
	case sdk.GCSFilesystemProvider:
		if plan.GCSConfig == nil || prev.GCSConfig == nil {
			return nil
		}
		if shouldPreserveSecret(plan.GCSConfig.Credentials, prev.GCSConfig.Credentials,
			plan.GCSConfig.CredentialsWOVersion, prev.GCSConfig.CredentialsWOVersion) {
			fsOut.GCSConfig.Credentials = preserveSecretSentinel()
		}
	case sdk.AzureBlobFilesystemProvider:
		if plan.AzBlobConfig == nil || prev.AzBlobConfig == nil {
			return nil
		}
		if shouldPreserveSecret(plan.AzBlobConfig.AccountKey, prev.AzBlobConfig.AccountKey,
			plan.AzBlobConfig.AccountKeyWOVersion, prev.AzBlobConfig.AccountKeyWOVersion) {
			fsOut.AzBlobConfig.AccountKey = preserveSecretSentinel()
		}
		if shouldPreserveSecret(plan.AzBlobConfig.SASURL, prev.AzBlobConfig.SASURL,
			plan.AzBlobConfig.SASURLWOVersion, prev.AzBlobConfig.SASURLWOVersion) {
			fsOut.AzBlobConfig.SASURL = preserveSecretSentinel()
		}
	case sdk.CryptedFilesystemProvider:
		if plan.CryptConfig == nil || prev.CryptConfig == nil {
			return nil
		}
		if shouldPreserveSecret(plan.CryptConfig.Passphrase, prev.CryptConfig.Passphrase,
			plan.CryptConfig.PassphraseWOVersion, prev.CryptConfig.PassphraseWOVersion) {
			fsOut.CryptConfig.Passphrase = preserveSecretSentinel()
		}
	case sdk.SFTPFilesystemProvider:
		if plan.SFTPConfig == nil || prev.SFTPConfig == nil {
			return nil
		}
		if shouldPreserveSecret(plan.SFTPConfig.Password, prev.SFTPConfig.Password,
			plan.SFTPConfig.PasswordWOVersion, prev.SFTPConfig.PasswordWOVersion) {
			fsOut.SFTPConfig.Password = preserveSecretSentinel()
		}
		if shouldPreserveSecret(plan.SFTPConfig.PrivateKey, prev.SFTPConfig.PrivateKey,
			plan.SFTPConfig.PrivateKeyWOVersion, prev.SFTPConfig.PrivateKeyWOVersion) {
			fsOut.SFTPConfig.PrivateKey = preserveSecretSentinel()
		}
		if shouldPreserveSecret(plan.SFTPConfig.KeyPassphrase, prev.SFTPConfig.KeyPassphrase,
			plan.SFTPConfig.KeyPassphraseWOVersion, prev.SFTPConfig.KeyPassphraseWOVersion) {
			fsOut.SFTPConfig.KeyPassphrase = preserveSecretSentinel()
		}
		if shouldPreserveSecret(plan.SFTPConfig.SocksPassword, prev.SFTPConfig.SocksPassword,
			plan.SFTPConfig.SocksPasswordWOVersion, prev.SFTPConfig.SocksPasswordWOVersion) {
			fsOut.SFTPConfig.SocksPassword = preserveSecretSentinel()
		}
	case sdk.HTTPFilesystemProvider:
		if plan.HTTPConfig == nil || prev.HTTPConfig == nil {
			return nil
		}
		if shouldPreserveSecret(plan.HTTPConfig.Password, prev.HTTPConfig.Password,
			plan.HTTPConfig.PasswordWOVersion, prev.HTTPConfig.PasswordWOVersion) {
			fsOut.HTTPConfig.Password = preserveSecretSentinel()
		}
		if shouldPreserveSecret(plan.HTTPConfig.APIKey, prev.HTTPConfig.APIKey,
			plan.HTTPConfig.APIKeyWOVersion, prev.HTTPConfig.APIKeyWOVersion) {
			fsOut.HTTPConfig.APIKey = preserveSecretSentinel()
		}
	case client.FTPFilesystemProvider:
		if plan.FTPConfig == nil || prev.FTPConfig == nil {
			return nil
		}
		if shouldPreserveSecret(plan.FTPConfig.Password, prev.FTPConfig.Password,
			plan.FTPConfig.PasswordWOVersion, prev.FTPConfig.PasswordWOVersion) {
			fsOut.FTPConfig.Password = preserveSecretSentinel()
		}
	}
	return nil
}

// shouldPreserveSecret reports whether a rotation was not requested for a
// secret attribute pair. The active channel in the current config drives the
// decision: if `<name>_wo_version` is set in plan, only its change triggers
// rotation; else if the legacy attribute is set in plan, its change triggers
// rotation; otherwise (neither in current config) the server-side value is
// preserved. This keeps WO↔legacy transitions correct — the side present in
// the new config is the authoritative one.
func shouldPreserveSecret(planLegacy, prevLegacy, planWOVersion, prevWOVersion types.String) bool {
	if !planWOVersion.IsNull() {
		return planWOVersion.Equal(prevWOVersion)
	}
	if !planLegacy.IsNull() {
		return planLegacy.Equal(prevLegacy)
	}
	return true
}

// preserveSecretSentinel returns a BaseSecret that SFTPGo's server-side
// updateEncryptedSecrets treats as "keep current" (IsNotPlainAndNotEmpty).
// Using Redacted avoids shipping any real secret material in the payload.
func preserveSecretSentinel() kms.BaseSecret {
	return kms.BaseSecret{Status: kms.SecretStatusRedacted}
}

// contains reports whether v is present in elems.
func contains[T comparable](elems []T, v T) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}
