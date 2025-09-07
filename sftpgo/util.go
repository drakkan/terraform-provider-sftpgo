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
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sftpgo/sdk"
)

const (
	computedSecretDescription = `SFTPGo secret formatted as string: "$<status>$<key>$<additional data length>$<additional data><payload>".`
	secretDescriptionGeneric  = `If you set a string in SFTPGo secret format, SFTPGo will keep the current secret on updates while the Terraform plan will save your value. Don't do this unless you are sure the values match (e.g because you imported an existing resource).`
	enterpriseFeatureNote     = `Available in the Enterprise edition`
)

func getComputedSchemaForFilesystem() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed:    true,
		Description: "Filesystem configuration.",
		Attributes: map[string]schema.Attribute{
			"provider": schema.Int64Attribute{
				Computed:    true,
				Description: "Provider. 0 = local filesystem, 1 = S3 Compatible, 2 = Google Cloud, 3 = Azure Blob, 4 = Local encrypted, 5 = SFTP, 6 = HTTP",
			},
			"osconfig": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"read_buffer_size": schema.Int64Attribute{
						Computed:    true,
						Description: "Optional read buffer size, as MB, to use for downloads.",
					},
					"write_buffer_size": schema.Int64Attribute{
						Computed:    true,
						Description: "Optional write buffer size, as MB, to use for uploads.",
					},
				},
			},
			"s3config": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"bucket": schema.StringAttribute{
						Computed: true,
					},
					"region": schema.StringAttribute{
						Computed: true,
					},
					"access_key": schema.StringAttribute{
						Computed: true,
					},
					"access_secret": schema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"sse_customer_key": schema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"key_prefix": schema.StringAttribute{
						Computed:    true,
						Description: `If specified then the SFTPGo user will be restricted to objects starting with this prefix.`,
					},
					"role_arn": schema.StringAttribute{
						Computed:    true,
						Description: "IAM Role ARN to assume.",
					},
					"session_token": schema.StringAttribute{
						Computed:    true,
						Description: "Optional Session token that is a part of temporary security credentials provisioned by AWS STS.",
					},
					"endpoint": schema.StringAttribute{
						Computed:    true,
						Description: "The endpoint is generally required for S3 compatible backends.",
					},
					"storage_class": schema.StringAttribute{
						Computed: true,
					},
					"acl": schema.StringAttribute{
						Computed:    true,
						Description: "The canned ACL to apply to uploaded objects. Empty means the bucket default.",
					},
					"upload_part_size": schema.Int64Attribute{
						Computed:    true,
						Description: "The buffer size (in MB) to use for multipart uploads.",
					},
					"upload_concurrency": schema.Int64Attribute{
						Computed:    true,
						Description: "How many parts are uploaded in parallel. Not set means the default (5).",
					},
					"download_part_size": schema.Int64Attribute{
						Computed:    true,
						Description: "The buffer size (in MB) to use for multipart downloads.",
					},
					"upload_part_max_time": schema.Int64Attribute{
						Computed:    true,
						Description: "The maximum time allowed, in seconds, to upload a single chunk. Not set means no timeout.",
					},
					"download_concurrency": schema.Int64Attribute{
						Computed:    true,
						Description: "How many parts are downloaded in parallel. Ignored for partial downloads.",
					},
					"download_part_max_time": schema.Int64Attribute{
						Computed:    true,
						Description: "The maximum time allowed, in seconds, to download a single chunk. Not set means no timeout.",
					},
					"force_path_style": schema.BoolAttribute{
						Computed:    true,
						Description: `If enabled path-style addressing is used, i.e. http://s3.amazonaws.com/BUCKET/KEY`,
					},
					"skip_tls_verify": schema.BoolAttribute{
						Computed:    true,
						Description: `If set the S3 client accepts any TLS certificate presented by the server and any host name in that certificate. In this mode, TLS is susceptible to man-in-the-middle attacks. This should be used only for testing.`,
					},
				},
			},
			"gcsconfig": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"bucket": schema.StringAttribute{
						Computed: true,
					},
					"key_prefix": schema.StringAttribute{
						Computed:    true,
						Description: `If specified then the SFTPGo user will be restricted to objects starting with this prefix.`,
					},
					"credentials": schema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"automatic_credentials": schema.Int64Attribute{
						Computed:    true,
						Description: "If set to 1 SFTPGo will use credentials from the environment",
					},
					"hns": schema.Int64Attribute{
						Computed:    true,
						Description: "1 if Hierarchical namespace support is enabled for the bucket. " + enterpriseFeatureNote + ".",
					},
					"storage_class": schema.StringAttribute{
						Computed: true,
					},
					"acl": schema.StringAttribute{
						Computed:    true,
						Description: "The ACL to apply to uploaded objects. Empty means the bucket default.",
					},
					"upload_part_size": schema.Int64Attribute{
						Computed:    true,
						Description: "The buffer size (in MB) to use for multipart uploads. The default value is 16MB. Not set means use the default.",
					},
					"upload_part_max_time": schema.Int64Attribute{
						Computed:    true,
						Description: "The maximum time allowed, in seconds, to upload a single chunk. The default value is 32. Not set means use the default.",
					},
				},
			},
			"azblobconfig": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"container": schema.StringAttribute{
						Computed: true,
					},
					"account_name": schema.StringAttribute{
						Computed: true,
					},
					"account_key": schema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"sas_url": schema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"endpoint": schema.StringAttribute{
						Computed:    true,
						Description: "Optional endpoint",
					},
					"key_prefix": schema.StringAttribute{
						Computed:    true,
						Description: `If specified then the SFTPGo user will be restricted to objects starting with this prefix.`,
					},
					"upload_part_size": schema.Int64Attribute{
						Computed:    true,
						Description: "The buffer size (in MB) to use for multipart uploads.",
					},
					"upload_concurrency": schema.Int64Attribute{
						Computed:    true,
						Description: "How many parts are uploaded in parallel.",
					},
					"download_part_size": schema.Int64Attribute{
						Computed:    true,
						Description: "The buffer size (in MB) to use for multipart downloads.",
					},
					"download_concurrency": schema.Int64Attribute{
						Computed:    true,
						Description: "How many parts are downloaded in parallel.",
					},
					"use_emulator": schema.BoolAttribute{
						Computed: true,
					},
					"access_tier": schema.StringAttribute{
						Computed: true,
					},
				},
			},
			"cryptconfig": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"passphrase": schema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"read_buffer_size": schema.Int64Attribute{
						Computed:    true,
						Description: "Optional read buffer size, as MB, to use for downloads.",
					},
					"write_buffer_size": schema.Int64Attribute{
						Computed:    true,
						Description: "Optional write buffer size, as MB, to use for uploads.",
					},
				},
			},
			"sftpconfig": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"endpoint": schema.StringAttribute{
						Computed:    true,
						Description: "SFTP endpoint as host:port.",
					},
					"username": schema.StringAttribute{
						Computed: true,
					},
					"password": schema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"private_key": schema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"key_passphrase": schema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"fingerprints": schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "SHA256 fingerprints to validate when connecting to the external SFTP server.",
					},
					"prefix": schema.StringAttribute{
						Computed:    true,
						Description: "Restrict access to this path.",
					},
					"disable_concurrent_reads": schema.BoolAttribute{
						Computed: true,
					},
					"buffer_size": schema.Int64Attribute{
						Computed:    true,
						Description: "The buffer size (in MB) to use for uploads/downloads.",
					},
					"equality_check_mode": schema.Int64Attribute{
						Computed: true,
					},
					"socks_proxy": schema.StringAttribute{
						Computed:    true,
						Description: "The address of the SOCKS proxy server, including schema, host, and port. Examples: socks5://127.0.0.1:1080, socks4://127.0.0.1:1080, socks4a://127.0.0.1:1080. " + enterpriseFeatureNote + ".",
					},
					"socks_username": schema.StringAttribute{
						Computed:    true,
						Description: "The optional SOCKS username. " + enterpriseFeatureNote + ".",
					},
					"socks_password": schema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription + " " + enterpriseFeatureNote + ".",
					},
				},
			},
			"httpconfig": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"endpoint": schema.StringAttribute{
						Computed: true,
					},
					"username": schema.StringAttribute{
						Computed: true,
					},
					"password": schema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"api_key": schema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"skip_tls_verify": schema.BoolAttribute{
						Computed: true,
					},
					"equality_check_mode": schema.Int64Attribute{
						Computed: true,
					},
				},
			},
		},
	}
}

func getSchemaForFilesystem() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Required:    true,
		Description: "Filesystem configuration.",
		Attributes: map[string]schema.Attribute{
			"provider": schema.Int64Attribute{
				Required:    true,
				Description: "Provider. 0 = local filesystem, 1 = S3 Compatible, 2 = Google Cloud, 3 = Azure Blob, 4 = Local encrypted, 5 = SFTP, 6 = HTTP",
				Validators: []validator.Int64{
					int64validator.Between(0, 6),
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
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"bucket": schema.StringAttribute{
						Required: true,
					},
					"region": schema.StringAttribute{
						Optional: true,
					},
					"access_key": schema.StringAttribute{
						Optional: true,
					},
					"access_secret": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text access secret. " + secretDescriptionGeneric,
					},
					"sse_customer_key": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text Server-Side encryption key. " + secretDescriptionGeneric,
					},
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
				},
			},
			"gcsconfig": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"bucket": schema.StringAttribute{
						Required: true,
					},
					"credentials": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text credentials. " + secretDescriptionGeneric,
					},
					"automatic_credentials": schema.Int64Attribute{
						Optional: true,
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
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"container": schema.StringAttribute{
						Optional: true,
					},
					"account_name": schema.StringAttribute{
						Optional: true,
					},
					"account_key": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text account key. " + secretDescriptionGeneric,
					},
					"sas_url": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text SAS URL. " + secretDescriptionGeneric,
					},
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
						Optional: true,
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
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text passphrase. " + secretDescriptionGeneric,
					},
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
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"endpoint": schema.StringAttribute{
						Required:    true,
						Description: "SFTP endpoint as host:port. Port is always required.",
						Validators: []validator.String{
							sftpEndPointValidator{},
						},
					},
					"username": schema.StringAttribute{
						Required: true,
					},
					"password": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text password. " + secretDescriptionGeneric,
					},
					"private_key": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text private key. " + secretDescriptionGeneric,
					},
					"key_passphrase": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text passphrase for the private key. " + secretDescriptionGeneric,
					},
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
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text SOCKS password. " + secretDescriptionGeneric + " " + enterpriseFeatureNote + ".",
					},
				},
			},
			"httpconfig": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"endpoint": schema.StringAttribute{
						Required: true,
					},
					"username": schema.StringAttribute{
						Optional: true,
					},
					"password": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text password. " + secretDescriptionGeneric,
					},
					"api_key": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text API key. " + secretDescriptionGeneric,
					},
					"skip_tls_verify": schema.BoolAttribute{
						Optional: true,
					},
					"equality_check_mode": schema.Int64Attribute{
						Optional: true,
					},
				},
			},
		},
	}
}

func getComputedSchemaForVirtualFolders() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Computed:    true,
		Description: "Virtual folder.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Computed:    true,
					Description: "Unique folder name",
				},
				"mapped_path": schema.StringAttribute{
					Computed:    true,
					Description: "Absolute path to a local directory. This is the folder root path for local storage provider. For non-local filesystems it will store temporary files.",
				},
				"virtual_path": schema.StringAttribute{
					Computed:    true,
					Description: "The folder will be available on this path.",
				},
				"description": schema.StringAttribute{
					Computed:    true,
					Description: "Optional description.",
				},
				"quota_size": schema.Int64Attribute{
					Computed:    true,
					Description: "Maximum size allowed as bytes. Not set means unlimited, -1 included in user quota",
				},
				"quota_files": schema.Int64Attribute{
					Computed:    true,
					Description: "Maximum number of files allowed. Not set means unlimited, -1 included in user quota",
				},
				"used_quota_size": schema.Int64Attribute{
					Computed:    true,
					Description: "Used quota as bytes.",
				},
				"used_quota_files": schema.Int64Attribute{
					Computed:    true,
					Description: "Used quota as number of files.",
				},
				"last_quota_update": schema.Int64Attribute{
					Computed:    true,
					Description: "Last quota update as unix timestamp in milliseconds",
				},
				"filesystem": getComputedSchemaForFilesystem(),
			},
		},
	}
}

func getSchemaForVirtualFolders() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
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

func getComputedSchemaForUserFilters(onlyBase bool) schema.SingleNestedAttribute {
	result := schema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"allowed_ip": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: `Only connections from these IP/Mask are allowed. IP/Mask must be in CIDR notation as defined in RFC 4632 and RFC 4291 for example "192.0.2.0/24" or "2001:db8::/32"`,
			},
			"denied_ip": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Connections from these IP/Mask are allowed. Denied rules will be evaluated before allowed ones.",
			},
			"denied_login_methods": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Disabled login methods.",
			},
			"denied_protocols": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Disabled protocols.",
			},
			"file_patterns": schema.ListNestedAttribute{
				Computed:    true,
				Description: `Filters based on shell patterns.`,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							Computed:    true,
							Description: "Virtual path, if no other specific filter is defined, the filter applies for sub directories too.",
						},
						"allowed_patterns": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "Files/directories with these, case insensitive, patterns are allowed. Allowed file patterns are evaluated before the denied ones.",
						},
						"denied_patterns": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "Files/directories with these, case insensitive, patterns are not allowed.",
						},
						"deny_policy": schema.Int64Attribute{
							Computed:    true,
							Description: "Set to 1 to hide denied files/directories in directory listing.",
						},
					},
				},
			},
			"max_upload_file_size": schema.Int64Attribute{
				Computed:    true,
				Description: "Max size allowed for a single upload. Unset means no limit.",
			},
			"tls_username": schema.StringAttribute{
				Computed:    true,
				Description: `TLS certificate attribute to use as username. For FTP clients it must match the name provided using the "USER" command. For WebDAV, if no username is provided, the CN will be used as username. For WebDAV clients it must match the implicit or provided username.`,
			},
			"external_auth_disabled": schema.BoolAttribute{
				Computed:    true,
				Description: "If set, external auth hook will not be executed.",
			},
			"pre_login_disabled": schema.BoolAttribute{
				Computed:    true,
				Description: "If set, external pre-login hook will not be executed.",
			},
			"check_password_disabled": schema.BoolAttribute{
				Computed:    true,
				Description: "If set, check password hook will not be executed.",
			},
			"disable_fs_checks": schema.BoolAttribute{
				Computed:    true,
				Description: "Disable checks for existence and automatic creation of home directory and virtual folders after user login.",
			},
			"web_client": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Web Client/user REST API restrictions.",
			},
			"allow_api_key_auth": schema.BoolAttribute{
				Computed:    true,
				Description: "If set, API Key authentication is allowed.",
			},
			"user_type": schema.StringAttribute{
				Computed:    true,
				Description: "Hint for authentication plugins.",
			},
			"bandwidth_limits": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Per-source bandwidth limits.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"sources": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: `Source networks in CIDR notation as defined in RFC 4632 and RFC 4291 for example "192.0.2.0/24" or "2001:db8::/32". The limit applies if the defined networks contain the client IP.`,
						},
						"upload_bandwidth": schema.Int64Attribute{
							Computed:    true,
							Description: "Maximum upload bandwidth as KB/s.",
						},
						"download_bandwidth": schema.Int64Attribute{
							Computed:    true,
							Description: "Maximum download bandwidth as KB/s.",
						},
					},
				},
			},
			"external_auth_cache_time": schema.Int64Attribute{
				Computed:    true,
				Description: "Defines the cache time, in seconds, for users authenticated using an external auth hook. Not set means no cache.",
			},
			"start_directory": schema.StringAttribute{
				Computed:    true,
				Description: `Alternate starting directory. If not set, the default is "/". This option is supported for SFTP/SCP, FTP and HTTP (WebClient/REST API) protocols. Relative paths will use this directory as base.`,
			},
			"two_factor_protocols": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Defines protocols that require two factor authentication",
			},
			"ftp_security": schema.Int64Attribute{
				Computed:    true,
				Description: "FTP security mode. Set to 1 to require TLS for both data and control connection.",
			},
			"is_anonymous": schema.BoolAttribute{
				Computed:    true,
				Description: `If enabled the user can login with any password or no password at all. Anonymous users are supported for FTP and WebDAV protocols and permissions will be automatically set to "list" and "download" (read only)`,
			},
			"default_shares_expiration": schema.Int64Attribute{
				Computed:    true,
				Description: "Default expiration for newly created shares as number of days. Not set means no default expiration.",
			},
			"max_shares_expiration": schema.Int64Attribute{
				Computed:    true,
				Description: "Maximum allowed expiration, as a number of days, when a user creates or updates a share. Not set means that non-expiring shares are allowed.",
			},
			"password_expiration": schema.Int64Attribute{
				Computed:    true,
				Description: "The password expires after the defined number of days. Not set means no expiration",
			},
			"password_strength": schema.Int64Attribute{
				Computed:    true,
				Description: "Minimum password strength. Not set means disabled, any password will be accepted. Values in the 50-70 range are suggested for common use cases.",
			},
			"access_time": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Time periods in which access is allowed",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"day_of_week": schema.Int64Attribute{
							Computed:    true,
							Description: "Day of week, 0 Sunday, 6 Saturday",
						},
						"from": schema.StringAttribute{
							Computed:    true,
							Description: "Start time in HH:MM format",
						},
						"to": schema.StringAttribute{
							Computed:    true,
							Description: "End time in HH:MM format",
						},
					},
				},
			},
			"enforce_secure_algorithms": schema.BoolAttribute{
				Computed:    true,
				Description: "If enabled, only secure algorithms are allowed. This setting is currently enforced for SSH/SFTP. " + enterpriseFeatureNote + ".",
			},
		},
	}
	if onlyBase {
		return result
	}
	result.Attributes["require_password_change"] = schema.BoolAttribute{
		Computed:    true,
		Description: "If set, user must change their password from WebClient/REST API at next login.",
	}
	result.Attributes["tls_certs"] = schema.ListAttribute{
		ElementType: types.StringType,
		Computed:    true,
		Description: "TLS certificates for mutual authentication. If provided will be checked before TLS username.",
	}
	result.Attributes["additional_emails"] = schema.ListAttribute{
		ElementType: types.StringType,
		Computed:    true,
		Description: "Additional email addresses.",
	}
	result.Attributes["custom1"] = schema.StringAttribute{
		Computed:    true,
		Description: `An extra placeholder value available for use in group configurations. It can be referenced as %custom1%. ` + enterpriseFeatureNote + ".",
	}
	return result
}

func getSchemaForUserFilters(onlyBase bool) schema.SingleNestedAttribute {
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
				Description: "Minimum password strength. Not set means disabled, any password will be accepted. Values in the 50-70 range are suggested for common use cases.",
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
		},
	}
	if onlyBase {
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
	return result
}

func preserveFsConfigPlanFields(ctx context.Context, fsPlan, fsState filesystem) (types.Object, diag.Diagnostics) {
	switch sdk.FilesystemProvider(fsState.Provider.ValueInt64()) {
	case sdk.S3FilesystemProvider:
		if fsPlan.S3Config != nil {
			fsState.S3Config.AccessSecret = fsPlan.S3Config.AccessSecret
			fsState.S3Config.SSECustomerKey = fsPlan.S3Config.SSECustomerKey
		}
	case sdk.GCSFilesystemProvider:
		if fsPlan.GCSConfig != nil {
			fsState.GCSConfig.Credentials = fsPlan.GCSConfig.Credentials
		}
	case sdk.AzureBlobFilesystemProvider:
		if fsPlan.AzBlobConfig != nil {
			fsState.AzBlobConfig.AccountKey = fsPlan.AzBlobConfig.AccountKey
			fsState.AzBlobConfig.SASURL = fsPlan.AzBlobConfig.SASURL
		}
	case sdk.CryptedFilesystemProvider:
		if fsPlan.CryptConfig != nil {
			fsState.CryptConfig.Passphrase = fsPlan.CryptConfig.Passphrase
		}
	case sdk.SFTPFilesystemProvider:
		if fsPlan.SFTPConfig != nil {
			fsState.SFTPConfig.Password = fsPlan.SFTPConfig.Password
			fsState.SFTPConfig.PrivateKey = fsPlan.SFTPConfig.PrivateKey
			fsState.SFTPConfig.KeyPassphrase = fsPlan.SFTPConfig.KeyPassphrase
			fsState.SFTPConfig.SocksPassword = fsPlan.SFTPConfig.SocksPassword
		}
	case sdk.HTTPFilesystemProvider:
		if fsPlan.HTTPConfig != nil {
			fsState.HTTPConfig.Password = fsPlan.HTTPConfig.Password
			fsState.HTTPConfig.APIKey = fsPlan.HTTPConfig.APIKey
		}
	}

	return types.ObjectValueFrom(ctx, fsState.getTFAttributes(), fsState)
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
