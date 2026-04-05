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
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sftpgo/sdk"
)

const (
	computedSecretDescription = `SFTPGo secret formatted as string: "$<status>$<key>$<additional data length>$<additional data><payload>".`
	secretDescriptionGeneric  = `If you set a string in SFTPGo secret format, SFTPGo will keep the current secret on updates while the Terraform plan will save your value. Don't do this unless you are sure the values match (e.g because you imported an existing resource).`
	writeOnlyNote             = `This is a write-only attribute, it will not be stored in the state. It is intended to be used with ephemeral values.`
	enterpriseFeatureNote     = `Available in the Enterprise edition`
)

func getComputedSchemaForFilesystem() dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		Computed:    true,
		Description: "Filesystem configuration.",
		Attributes: map[string]dschema.Attribute{
			"provider": dschema.Int64Attribute{
				Computed:    true,
				Description: "Provider. 0 = local filesystem, 1 = S3 Compatible, 2 = Google Cloud, 3 = Azure Blob, 4 = Local encrypted, 5 = SFTP, 6 = HTTP",
			},
			"osconfig": dschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dschema.Attribute{
					"read_buffer_size": dschema.Int64Attribute{
						Computed:    true,
						Description: "Optional read buffer size, as MB, to use for downloads.",
					},
					"write_buffer_size": dschema.Int64Attribute{
						Computed:    true,
						Description: "Optional write buffer size, as MB, to use for uploads.",
					},
				},
			},
			"s3config": dschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dschema.Attribute{
					"bucket": dschema.StringAttribute{
						Computed: true,
					},
					"region": dschema.StringAttribute{
						Computed: true,
					},
					"access_key": dschema.StringAttribute{
						Computed: true,
					},
					"access_secret": dschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"sse_customer_key": dschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"key_prefix": dschema.StringAttribute{
						Computed:    true,
						Description: `If specified then the SFTPGo user will be restricted to objects starting with this prefix.`,
					},
					"role_arn": dschema.StringAttribute{
						Computed:    true,
						Description: "IAM Role ARN to assume.",
					},
					"session_token": dschema.StringAttribute{
						Computed:    true,
						Description: "Optional Session token that is a part of temporary security credentials provisioned by AWS STS.",
					},
					"endpoint": dschema.StringAttribute{
						Computed:    true,
						Description: "The endpoint is generally required for S3 compatible backends.",
					},
					"storage_class": dschema.StringAttribute{
						Computed: true,
					},
					"acl": dschema.StringAttribute{
						Computed:    true,
						Description: "The canned ACL to apply to uploaded objects. Empty means the bucket default.",
					},
					"upload_part_size": dschema.Int64Attribute{
						Computed:    true,
						Description: "The buffer size (in MB) to use for multipart uploads.",
					},
					"upload_concurrency": dschema.Int64Attribute{
						Computed:    true,
						Description: "How many parts are uploaded in parallel. Not set means the default (5).",
					},
					"download_part_size": dschema.Int64Attribute{
						Computed:    true,
						Description: "The buffer size (in MB) to use for multipart downloads.",
					},
					"upload_part_max_time": dschema.Int64Attribute{
						Computed:    true,
						Description: "The maximum time allowed, in seconds, to upload a single chunk. Not set means no timeout.",
					},
					"download_concurrency": dschema.Int64Attribute{
						Computed:    true,
						Description: "How many parts are downloaded in parallel. Ignored for partial downloads.",
					},
					"download_part_max_time": dschema.Int64Attribute{
						Computed:    true,
						Description: "The maximum time allowed, in seconds, to download a single chunk. Not set means no timeout.",
					},
					"force_path_style": dschema.BoolAttribute{
						Computed:    true,
						Description: `If enabled path-style addressing is used, i.e. http://s3.amazonaws.com/BUCKET/KEY`,
					},
					"skip_tls_verify": dschema.BoolAttribute{
						Computed:    true,
						Description: `If set the S3 client accepts any TLS certificate presented by the server and any host name in that certificate. In this mode, TLS is susceptible to man-in-the-middle attacks. This should be used only for testing.`,
					},
				},
			},
			"gcsconfig": dschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dschema.Attribute{
					"bucket": dschema.StringAttribute{
						Computed: true,
					},
					"key_prefix": dschema.StringAttribute{
						Computed:    true,
						Description: `If specified then the SFTPGo user will be restricted to objects starting with this prefix.`,
					},
					"universe_domain": dschema.StringAttribute{
						Computed:    true,
						Description: `The universe domain to use for Google Cloud API requests. If omitted or empty, the default public domain (googleapis.com) is used. Set this value if you need to connect to a custom Google Cloud environment, such as Google Distributed Cloud or a Sovereign Cloud. ` + enterpriseFeatureNote,
					},
					"credentials": dschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"automatic_credentials": dschema.Int64Attribute{
						Computed:    true,
						Description: "If set to 1 SFTPGo will use credentials from the environment",
					},
					"hns": dschema.Int64Attribute{
						Computed:    true,
						Description: "1 if Hierarchical namespace support is enabled for the bucket. " + enterpriseFeatureNote + ".",
					},
					"storage_class": dschema.StringAttribute{
						Computed: true,
					},
					"acl": dschema.StringAttribute{
						Computed:    true,
						Description: "The ACL to apply to uploaded objects. Empty means the bucket default.",
					},
					"upload_part_size": dschema.Int64Attribute{
						Computed:    true,
						Description: "The buffer size (in MB) to use for multipart uploads. The default value is 16MB. Not set means use the default.",
					},
					"upload_part_max_time": dschema.Int64Attribute{
						Computed:    true,
						Description: "The maximum time allowed, in seconds, to upload a single chunk. The default value is 32. Not set means use the default.",
					},
				},
			},
			"azblobconfig": dschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dschema.Attribute{
					"container": dschema.StringAttribute{
						Computed: true,
					},
					"account_name": dschema.StringAttribute{
						Computed: true,
					},
					"account_key": dschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"sas_url": dschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"endpoint": dschema.StringAttribute{
						Computed:    true,
						Description: "Optional endpoint",
					},
					"key_prefix": dschema.StringAttribute{
						Computed:    true,
						Description: `If specified then the SFTPGo user will be restricted to objects starting with this prefix.`,
					},
					"upload_part_size": dschema.Int64Attribute{
						Computed:    true,
						Description: "The buffer size (in MB) to use for multipart uploads.",
					},
					"upload_concurrency": dschema.Int64Attribute{
						Computed:    true,
						Description: "How many parts are uploaded in parallel.",
					},
					"download_part_size": dschema.Int64Attribute{
						Computed:    true,
						Description: "The buffer size (in MB) to use for multipart downloads.",
					},
					"download_concurrency": dschema.Int64Attribute{
						Computed:    true,
						Description: "How many parts are downloaded in parallel.",
					},
					"use_emulator": dschema.BoolAttribute{
						Computed: true,
					},
					"access_tier": dschema.StringAttribute{
						Computed: true,
					},
				},
			},
			"cryptconfig": dschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dschema.Attribute{
					"passphrase": dschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"read_buffer_size": dschema.Int64Attribute{
						Computed:    true,
						Description: "Optional read buffer size, as MB, to use for downloads.",
					},
					"write_buffer_size": dschema.Int64Attribute{
						Computed:    true,
						Description: "Optional write buffer size, as MB, to use for uploads.",
					},
				},
			},
			"sftpconfig": dschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dschema.Attribute{
					"endpoint": dschema.StringAttribute{
						Computed:    true,
						Description: "SFTP endpoint as host:port.",
					},
					"username": dschema.StringAttribute{
						Computed: true,
					},
					"password": dschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"private_key": dschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"key_passphrase": dschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"fingerprints": dschema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "SHA256 fingerprints to validate when connecting to the external SFTP server.",
					},
					"prefix": dschema.StringAttribute{
						Computed:    true,
						Description: "Restrict access to this path.",
					},
					"disable_concurrent_reads": dschema.BoolAttribute{
						Computed: true,
					},
					"buffer_size": dschema.Int64Attribute{
						Computed:    true,
						Description: "The buffer size (in MB) to use for uploads/downloads.",
					},
					"equality_check_mode": dschema.Int64Attribute{
						Computed: true,
					},
					"socks_proxy": dschema.StringAttribute{
						Computed:    true,
						Description: "The address of the SOCKS proxy server, including schema, host, and port. Examples: socks5://127.0.0.1:1080, socks4://127.0.0.1:1080, socks4a://127.0.0.1:1080. " + enterpriseFeatureNote + ".",
					},
					"socks_username": dschema.StringAttribute{
						Computed:    true,
						Description: "The optional SOCKS username. " + enterpriseFeatureNote + ".",
					},
					"socks_password": dschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription + " " + enterpriseFeatureNote + ".",
					},
				},
			},
			"ftpconfig": dschema.SingleNestedAttribute{
				Computed:    true,
				Description: enterpriseFeatureNote,
				Attributes: map[string]dschema.Attribute{
					"endpoint": dschema.StringAttribute{
						Computed:    true,
						Description: "FTP endpoint as host:port.",
					},
					"username": dschema.StringAttribute{
						Computed: true,
					},
					"password": dschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"tls_mode": dschema.Int64Attribute{
						Computed:    true,
						Description: "0 disabled, 1 Explicit, 2 Implicit.",
					},
					"skip_tls_verify": dschema.BoolAttribute{
						Computed: true,
					},
				},
			},
			"httpconfig": dschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dschema.Attribute{
					"endpoint": dschema.StringAttribute{
						Computed: true,
					},
					"username": dschema.StringAttribute{
						Computed: true,
					},
					"password": dschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"api_key": dschema.StringAttribute{
						Computed:    true,
						Description: computedSecretDescription,
					},
					"skip_tls_verify": dschema.BoolAttribute{
						Computed: true,
					},
					"equality_check_mode": dschema.Int64Attribute{
						Computed: true,
					},
				},
			},
		},
	}
}

func getSchemaForFilesystem() rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		Required:    true,
		Description: "Filesystem configuration.",
		Attributes: map[string]rschema.Attribute{
			"provider": rschema.Int64Attribute{
				Required:    true,
				Description: "Provider. 0 = local filesystem, 1 = S3 Compatible, 2 = Google Cloud, 3 = Azure Blob, 4 = Local encrypted, 5 = SFTP, 6 = HTTP, 7 = FTP",
				Validators: []validator.Int64{
					int64validator.Between(0, 7),
				},
			},
			"osconfig": rschema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]rschema.Attribute{
					"read_buffer_size": rschema.Int64Attribute{
						Optional:    true,
						Description: "Optional read buffer size, as MB, to use for downloads. Omit to disable buffering, that's fine in most use cases.",
						Validators: []validator.Int64{
							int64validator.Between(0, 10),
						},
					},
					"write_buffer_size": rschema.Int64Attribute{
						Optional:    true,
						Description: "Optional write buffer size, as MB, to use for uploads. Omit to disable no buffering, that's fine in most use cases.",
						Validators: []validator.Int64{
							int64validator.Between(0, 10),
						},
					},
				},
			},
			"s3config": rschema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]rschema.Attribute{
					"bucket": rschema.StringAttribute{
						Required: true,
					},
					"region": rschema.StringAttribute{
						Optional: true,
					},
					"access_key": rschema.StringAttribute{
						Optional: true,
					},
					"access_secret": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text access secret. " + secretDescriptionGeneric,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("access_secret_wo")),
						},
					},
					"access_secret_wo": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						WriteOnly:   true,
						Description: "Write-only plain text access secret. " + writeOnlyNote,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("access_secret")),
						},
					},
					"access_secret_wo_version": rschema.Int64Attribute{
						Optional:    true,
						Description: "Version for the write-only access secret.",
					},
					"sse_customer_key": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text Server-Side encryption key. " + secretDescriptionGeneric,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("sse_customer_key_wo")),
						},
					},
					"sse_customer_key_wo": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						WriteOnly:   true,
						Description: "Write-only plain text Server-Side encryption key. " + writeOnlyNote,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("sse_customer_key")),
						},
					},
					"sse_customer_key_wo_version": rschema.Int64Attribute{
						Optional:    true,
						Description: "Version for the write-only Server-Side encryption key.",
					},
					"key_prefix": rschema.StringAttribute{
						Optional:    true,
						Description: `If specified then the SFTPGo user will be restricted to objects starting with the specified prefix. The prefix must not start with "/" and must end with "/"`,
					},
					"role_arn": rschema.StringAttribute{
						Optional:    true,
						Description: "Optional IAM Role ARN to assume.",
					},
					"session_token": rschema.StringAttribute{
						Optional:    true,
						Description: "Optional Session token that is a part of temporary security credentials provisioned by AWS STS.",
					},
					"endpoint": rschema.StringAttribute{
						Optional:    true,
						Description: "The endpoint is generally required for S3 compatible backends. For AWS S3, leave not set to use the default endpoint for the specified region.",
					},
					"storage_class": rschema.StringAttribute{
						Optional:    true,
						Description: "The storage class to use when storing objects. Leave not set for default.",
					},
					"acl": rschema.StringAttribute{
						Optional:    true,
						Description: "The canned ACL to apply to uploaded objects. Not set means the bucket default.",
					},
					"upload_part_size": rschema.Int64Attribute{
						Optional:    true,
						Description: "The buffer size (in MB) to use for multipart uploads. If this value is not set, the default value (5MB) will be used.",
					},
					"upload_concurrency": rschema.Int64Attribute{
						Optional:    true,
						Description: "How many parts are uploaded in parallel. Not set means the default (5).",
					},
					"download_part_size": rschema.Int64Attribute{
						Optional:    true,
						Description: "The buffer size (in MB) to use for multipart downloads. If this value is not set, the default value (5MB) will be used.",
					},
					"upload_part_max_time": rschema.Int64Attribute{
						Optional:    true,
						Description: "The maximum time allowed, in seconds, to upload a single chunk. Not set means no timeout.",
					},
					"download_concurrency": rschema.Int64Attribute{
						Optional:    true,
						Description: "How many parts are downloaded in parallel. Not set means the default (5). Ignored for partial downloads.",
					},
					"download_part_max_time": rschema.Int64Attribute{
						Optional:    true,
						Description: "The maximum time allowed, in seconds, to download a single chunk. Not set means no timeout. Ignored for partial downloads.",
					},
					"force_path_style": rschema.BoolAttribute{
						Optional:    true,
						Description: `If set path-style addressing is used, i.e. http://s3.amazonaws.com/BUCKET/KEY`,
					},
					"skip_tls_verify": rschema.BoolAttribute{
						Optional:    true,
						Description: `If set the S3 client accepts any TLS certificate presented by the server and any host name in that certificate. In this mode, TLS is susceptible to man-in-the-middle attacks. This should be used only for testing.`,
					},
				},
			},
			"gcsconfig": rschema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]rschema.Attribute{
					"bucket": rschema.StringAttribute{
						Required: true,
					},
					"universe_domain": rschema.StringAttribute{
						Optional:    true,
						Description: `The universe domain to use for Google Cloud API requests. If omitted or empty, the default public domain (googleapis.com) is used. Set this value if you need to connect to a custom Google Cloud environment, such as Google Distributed Cloud or a Sovereign Cloud. ` + enterpriseFeatureNote,
					},
					"credentials": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text credentials. " + secretDescriptionGeneric,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("credentials_wo")),
						},
					},
					"credentials_wo": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						WriteOnly:   true,
						Description: "Write-only plain text credentials. " + writeOnlyNote,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("credentials")),
						},
					},
					"credentials_wo_version": rschema.Int64Attribute{
						Optional:    true,
						Description: "Version for the write-only credentials.",
					},
					"automatic_credentials": rschema.Int64Attribute{
						Optional: true,
					},
					"hns": rschema.Int64Attribute{
						Optional:    true,
						Description: "Set to 1 if Hierarchical namespace is enabled for the bucket. " + enterpriseFeatureNote + ".",
					},
					"key_prefix": rschema.StringAttribute{
						Optional:    true,
						Description: `If specified then the SFTPGo user will be restricted to objects starting with the specified prefix. The prefix must not start with "/" and must end with "/"`,
					},
					"storage_class": rschema.StringAttribute{
						Optional:    true,
						Description: "The storage class to use when storing objects. Leave not set for default.",
					},
					"acl": rschema.StringAttribute{
						Optional:    true,
						Description: "The ACL to apply to uploaded objects. Not set means the bucket default.",
					},
					"upload_part_size": rschema.Int64Attribute{
						Optional:    true,
						Description: "The buffer size (in MB) to use for multipart uploads. The default value is 16MB. Not set means use the default.",
					},
					"upload_part_max_time": rschema.Int64Attribute{
						Optional:    true,
						Description: "The maximum time allowed, in seconds, to upload a single chunk. The default value is 32. Not set means use the default.",
					},
				},
			},
			"azblobconfig": rschema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]rschema.Attribute{
					"container": rschema.StringAttribute{
						Optional: true,
					},
					"account_name": rschema.StringAttribute{
						Optional: true,
					},
					"account_key": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text account key. " + secretDescriptionGeneric,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("account_key_wo")),
						},
					},
					"account_key_wo": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						WriteOnly:   true,
						Description: "Write-only plain text account key. " + writeOnlyNote,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("account_key")),
						},
					},
					"account_key_wo_version": rschema.Int64Attribute{
						Optional:    true,
						Description: "Version for the write-only account key.",
					},
					"sas_url": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text SAS URL. " + secretDescriptionGeneric,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("sas_url_wo")),
						},
					},
					"sas_url_wo": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						WriteOnly:   true,
						Description: "Write-only plain text SAS URL. " + writeOnlyNote,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("sas_url")),
						},
					},
					"sas_url_wo_version": rschema.Int64Attribute{
						Optional:    true,
						Description: "Version for the write-only SAS URL.",
					},
					"endpoint": rschema.StringAttribute{
						Optional:    true,
						Description: `Optional endpoint. Default is "blob.core.windows.net". If you use the emulator the endpoint must include the protocol, for example "http://127.0.0.1:10000".`,
					},
					"key_prefix": rschema.StringAttribute{
						Optional:    true,
						Description: `If specified then the SFTPGo user will be restricted to objects starting with the specified prefix. The prefix must not start with "/" and must end with "/"`,
					},
					"upload_part_size": rschema.Int64Attribute{
						Optional:    true,
						Description: "The buffer size (in MB) to use for multipart uploads. If this value is not set, the default value (5MB) will be used.",
					},
					"upload_concurrency": rschema.Int64Attribute{
						Optional:    true,
						Description: "How many parts are uploaded in parallel. Default: 5.",
					},
					"download_part_size": rschema.Int64Attribute{
						Optional:    true,
						Description: "The buffer size (in MB) to use for multipart downloads. If this value is not set, the default value (5MB) will be used.",
					},
					"download_concurrency": rschema.Int64Attribute{
						Optional:    true,
						Description: "How many parts are downloaded in parallel. Default: 5.",
					},
					"use_emulator": rschema.BoolAttribute{
						Optional: true,
					},
					"access_tier": rschema.StringAttribute{
						Optional:    true,
						Description: "Blob Access Tier. Not set means the container default.",
					},
				},
			},
			"cryptconfig": rschema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]rschema.Attribute{
					"passphrase": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text passphrase. " + secretDescriptionGeneric,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("passphrase_wo")),
						},
					},
					"passphrase_wo": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						WriteOnly:   true,
						Description: "Write-only plain text passphrase. " + writeOnlyNote,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("passphrase")),
						},
					},
					"passphrase_wo_version": rschema.Int64Attribute{
						Optional:    true,
						Description: "Version for the write-only passphrase.",
					},
					"read_buffer_size": rschema.Int64Attribute{
						Optional:    true,
						Description: "Optional read buffer size, as MB, to use for downloads. Omit to disable buffering, that's fine in most use cases.",
						Validators: []validator.Int64{
							int64validator.Between(0, 10),
						},
					},
					"write_buffer_size": rschema.Int64Attribute{
						Optional:    true,
						Description: "Optional write buffer size, as MB, to use for uploads. Omit to disable buffering, that's fine in most use cases.",
						Validators: []validator.Int64{
							int64validator.Between(0, 10),
						},
					},
				},
			},
			"sftpconfig": rschema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]rschema.Attribute{
					"endpoint": rschema.StringAttribute{
						Required:    true,
						Description: "SFTP endpoint as host:port. Port is always required.",
						Validators: []validator.String{
							sftpEndPointValidator{},
						},
					},
					"username": rschema.StringAttribute{
						Required: true,
					},
					"password": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text password. " + secretDescriptionGeneric,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("password_wo")),
						},
					},
					"password_wo": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						WriteOnly:   true,
						Description: "Write-only plain text password. " + writeOnlyNote,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("password")),
						},
					},
					"password_wo_version": rschema.Int64Attribute{
						Optional:    true,
						Description: "Version for the write-only password.",
					},
					"private_key": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text private key. " + secretDescriptionGeneric,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("private_key_wo")),
						},
					},
					"private_key_wo": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						WriteOnly:   true,
						Description: "Write-only plain text private key. " + writeOnlyNote,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("private_key")),
						},
					},
					"private_key_wo_version": rschema.Int64Attribute{
						Optional:    true,
						Description: "Version for the write-only private key.",
					},
					"key_passphrase": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text passphrase for the private key. " + secretDescriptionGeneric,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("key_passphrase_wo")),
						},
					},
					"key_passphrase_wo": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						WriteOnly:   true,
						Description: "Write-only plain text passphrase for the private key. " + writeOnlyNote,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("key_passphrase")),
						},
					},
					"key_passphrase_wo_version": rschema.Int64Attribute{
						Optional:    true,
						Description: "Version for the write-only passphrase for the private key.",
					},
					"fingerprints": rschema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "SHA256 fingerprints to validate when connecting to the external SFTP server. If not set any host key will be accepted: this is a security risk.",
					},
					"prefix": rschema.StringAttribute{
						Required:    true,
						Description: `Similar to a chroot for local filesystem. Example: "/somedir/subdir".`,
					},
					"disable_concurrent_reads": rschema.BoolAttribute{
						Optional:    true,
						Description: "Concurrent reads are safe to use and disabling them will degrade performance so they are enabled by default. Some servers automatically delete files once they are downloaded. Using concurrent reads is problematic with such servers.",
					},
					"buffer_size": rschema.Int64Attribute{
						Optional:    true,
						Description: "The buffer size (in MB) to use for uploads/downloads. Buffering could improve performance for high latency networks. With buffering enabled upload resume is not supported and a file cannot be opened for both reading and writing at the same time. Not set means disabled.",
					},
					"equality_check_mode": rschema.Int64Attribute{
						Optional:    true,
						Description: "Defines how to check if this config points to the same server as another config. By default both the endpoint and the username must match. 1 means that only the endpoint must match. If different configs point to the same server the renaming between the fs configs is allowed.",
					},
					"socks_proxy": rschema.StringAttribute{
						Optional:    true,
						Description: "The address of the SOCKS proxy server, including schema, host, and port. Examples: socks5://127.0.0.1:1080, socks4://127.0.0.1:1080, socks4a://127.0.0.1:1080. " + enterpriseFeatureNote + ".",
					},
					"socks_username": rschema.StringAttribute{
						Optional:    true,
						Description: "The optional SOCKS username. " + enterpriseFeatureNote + ".",
					},
					"socks_password": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text SOCKS password. " + secretDescriptionGeneric + " " + enterpriseFeatureNote + ".",
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("socks_password_wo")),
						},
					},
					"socks_password_wo": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						WriteOnly:   true,
						Description: "Write-only plain text SOCKS password. " + writeOnlyNote + " " + enterpriseFeatureNote + ".",
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("socks_password")),
						},
					},
					"socks_password_wo_version": rschema.Int64Attribute{
						Optional:    true,
						Description: "Version for the write-only SOCKS password. " + enterpriseFeatureNote + ".",
					},
				},
			},
			"ftpconfig": rschema.SingleNestedAttribute{
				Optional:    true,
				Description: enterpriseFeatureNote,
				Attributes: map[string]rschema.Attribute{
					"endpoint": rschema.StringAttribute{
						Required:    true,
						Description: "FTP endpoint as host:port.",
					},
					"username": rschema.StringAttribute{
						Required: true,
					},
					"password": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text password. " + secretDescriptionGeneric,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("password_wo")),
						},
					},
					"password_wo": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						WriteOnly:   true,
						Description: "Write-only plain text password. " + writeOnlyNote,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("password")),
						},
					},
					"password_wo_version": rschema.Int64Attribute{
						Optional:    true,
						Description: "Version for the write-only password.",
					},
					"tls_mode": rschema.Int64Attribute{
						Optional:    true,
						Description: "0 disabled, 1 Explicit, 2 Implicit.",
					},
					"skip_tls_verify": rschema.BoolAttribute{
						Optional: true,
					},
				},
			},
			"httpconfig": rschema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]rschema.Attribute{
					"endpoint": rschema.StringAttribute{
						Required: true,
					},
					"username": rschema.StringAttribute{
						Optional: true,
					},
					"password": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text password. " + secretDescriptionGeneric,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("password_wo")),
						},
					},
					"password_wo": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						WriteOnly:   true,
						Description: "Write-only plain text password. " + writeOnlyNote,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("password")),
						},
					},
					"password_wo_version": rschema.Int64Attribute{
						Optional:    true,
						Description: "Version for the write-only password.",
					},
					"api_key": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Plain text API key. " + secretDescriptionGeneric,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("api_key_wo")),
						},
					},
					"api_key_wo": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						WriteOnly:   true,
						Description: "Write-only plain text API key. " + writeOnlyNote,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("api_key")),
						},
					},
					"api_key_wo_version": rschema.Int64Attribute{
						Optional:    true,
						Description: "Version for the write-only API key.",
					},
					"skip_tls_verify": rschema.BoolAttribute{
						Optional: true,
					},
					"equality_check_mode": rschema.Int64Attribute{
						Optional: true,
					},
				},
			},
		},
	}
}

func getComputedSchemaForVirtualFolders() dschema.ListNestedAttribute {
	return dschema.ListNestedAttribute{
		Computed:    true,
		Description: "Virtual folder.",
		NestedObject: dschema.NestedAttributeObject{
			Attributes: map[string]dschema.Attribute{
				"name": dschema.StringAttribute{
					Computed:    true,
					Description: "Unique folder name",
				},
				"mapped_path": dschema.StringAttribute{
					Computed:    true,
					Description: "Absolute path to a local directory. This is the folder root path for local storage provider. For non-local filesystems it will store temporary files.",
				},
				"virtual_path": dschema.StringAttribute{
					Computed:    true,
					Description: "The folder will be available on this path.",
				},
				"description": dschema.StringAttribute{
					Computed:    true,
					Description: "Optional description.",
				},
				"quota_size": dschema.Int64Attribute{
					Computed:    true,
					Description: "Maximum size allowed as bytes. Not set means unlimited, -1 included in user quota",
				},
				"quota_files": dschema.Int64Attribute{
					Computed:    true,
					Description: "Maximum number of files allowed. Not set means unlimited, -1 included in user quota",
				},
				"used_quota_size": dschema.Int64Attribute{
					Computed:    true,
					Description: "Used quota as bytes.",
				},
				"used_quota_files": dschema.Int64Attribute{
					Computed:    true,
					Description: "Used quota as number of files.",
				},
				"last_quota_update": dschema.Int64Attribute{
					Computed:    true,
					Description: "Last quota update as unix timestamp in milliseconds",
				},
				"filesystem": getComputedSchemaForFilesystem(),
			},
		},
	}
}

func getSchemaForVirtualFolders() rschema.ListNestedAttribute {
	return rschema.ListNestedAttribute{
		Optional: true,
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"name": rschema.StringAttribute{
					Required:    true,
					Description: "Unique folder name",
				},
				"virtual_path": rschema.StringAttribute{
					Required:    true,
					Description: "The folder will be available on this path.",
				},
				"quota_size": rschema.Int64Attribute{
					Required:    true,
					Description: "Maximum size allowed as bytes. Not set means unlimited, -1 included in user quota",
				},
				"quota_files": rschema.Int64Attribute{
					Required:    true,
					Description: "Maximum number of files allowed. Not set means unlimited, -1 included in user quota",
				},
				"mapped_path": rschema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Absolute path to a local directory. This is the folder root path for local storage provider. For non-local filesystems it will store temporary files.",
				},
				"description": rschema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Optional description.",
				},
				"used_quota_size": rschema.Int64Attribute{
					Computed:    true,
					Optional:    true,
					Description: "Used quota as bytes.",
				},
				"used_quota_files": rschema.Int64Attribute{
					Computed:    true,
					Optional:    true,
					Description: "Used quota as number of files.",
				},
				"last_quota_update": rschema.Int64Attribute{
					Computed:    true,
					Optional:    true,
					Description: "Last quota update as unix timestamp in milliseconds",
				},
				"filesystem": getComputedSchemaForFilesystem(),
			},
		},
	}
}

func getComputedSchemaForUserFilters(isGroup bool) dschema.SingleNestedAttribute {
	result := dschema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]dschema.Attribute{
			"allowed_ip": dschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: `Only connections from these IP/Mask are allowed. IP/Mask must be in CIDR notation as defined in RFC 4632 and RFC 4291 for example "192.0.2.0/24" or "2001:db8::/32"`,
			},
			"denied_ip": dschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Connections from these IP/Mask are allowed. Denied rules will be evaluated before allowed ones.",
			},
			"denied_login_methods": dschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Disabled login methods.",
			},
			"denied_protocols": dschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Disabled protocols.",
			},
			"file_patterns": dschema.ListNestedAttribute{
				Computed:    true,
				Description: `Filters based on shell patterns.`,
				NestedObject: dschema.NestedAttributeObject{
					Attributes: map[string]dschema.Attribute{
						"path": dschema.StringAttribute{
							Computed:    true,
							Description: "Virtual path, if no other specific filter is defined, the filter applies for sub directories too.",
						},
						"allowed_patterns": dschema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "Files/directories with these, case insensitive, patterns are allowed. Allowed file patterns are evaluated before the denied ones.",
						},
						"denied_patterns": dschema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "Files/directories with these, case insensitive, patterns are not allowed.",
						},
						"deny_policy": dschema.Int64Attribute{
							Computed:    true,
							Description: "Set to 1 to hide denied files/directories in directory listing.",
						},
					},
				},
			},
			"max_upload_file_size": dschema.Int64Attribute{
				Computed:    true,
				Description: "Max size allowed for a single upload. Unset means no limit.",
			},
			"tls_username": dschema.StringAttribute{
				Computed:    true,
				Description: `TLS certificate attribute to use as username. For FTP clients it must match the name provided using the "USER" command. For WebDAV, if no username is provided, the CN will be used as username. For WebDAV clients it must match the implicit or provided username.`,
			},
			"external_auth_disabled": dschema.BoolAttribute{
				Computed:    true,
				Description: "If set, external auth hook will not be executed.",
			},
			"pre_login_disabled": dschema.BoolAttribute{
				Computed:    true,
				Description: "If set, external pre-login hook will not be executed.",
			},
			"check_password_disabled": dschema.BoolAttribute{
				Computed:    true,
				Description: "If set, check password hook will not be executed.",
			},
			"disable_fs_checks": dschema.BoolAttribute{
				Computed:    true,
				Description: "Disable checks for existence and automatic creation of home directory and virtual folders after user login.",
			},
			"web_client": dschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Web Client/user REST API restrictions.",
			},
			"allow_api_key_auth": dschema.BoolAttribute{
				Computed:    true,
				Description: "If set, API Key authentication is allowed.",
			},
			"user_type": dschema.StringAttribute{
				Computed:    true,
				Description: "Hint for authentication plugins.",
			},
			"bandwidth_limits": dschema.ListNestedAttribute{
				Computed:    true,
				Description: "Per-source bandwidth limits.",
				NestedObject: dschema.NestedAttributeObject{
					Attributes: map[string]dschema.Attribute{
						"sources": dschema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: `Source networks in CIDR notation as defined in RFC 4632 and RFC 4291 for example "192.0.2.0/24" or "2001:db8::/32". The limit applies if the defined networks contain the client IP.`,
						},
						"upload_bandwidth": dschema.Int64Attribute{
							Computed:    true,
							Description: "Maximum upload bandwidth as KB/s.",
						},
						"download_bandwidth": dschema.Int64Attribute{
							Computed:    true,
							Description: "Maximum download bandwidth as KB/s.",
						},
					},
				},
			},
			"external_auth_cache_time": dschema.Int64Attribute{
				Computed:    true,
				Description: "Defines the cache time, in seconds, for users authenticated using an external auth hook. Not set means no cache.",
			},
			"start_directory": dschema.StringAttribute{
				Computed:    true,
				Description: `Alternate starting directory. If not set, the default is "/". This option is supported for SFTP/SCP, FTP and HTTP (WebClient/REST API) protocols. Relative paths will use this directory as base.`,
			},
			"two_factor_protocols": dschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Defines protocols that require two factor authentication",
			},
			"ftp_security": dschema.Int64Attribute{
				Computed:    true,
				Description: "FTP security mode. Set to 1 to require TLS for both data and control connection.",
			},
			"is_anonymous": dschema.BoolAttribute{
				Computed:    true,
				Description: `If enabled the user can login with any password or no password at all. Anonymous users are supported for FTP and WebDAV protocols and permissions will be automatically set to "list" and "download" (read only)`,
			},
			"default_shares_expiration": dschema.Int64Attribute{
				Computed:    true,
				Description: "Default expiration for newly created shares as number of days. Not set means no default expiration.",
			},
			"max_shares_expiration": dschema.Int64Attribute{
				Computed:    true,
				Description: "Maximum allowed expiration, as a number of days, when a user creates or updates a share. Not set means that non-expiring shares are allowed.",
			},
			"password_expiration": dschema.Int64Attribute{
				Computed:    true,
				Description: "The password expires after the defined number of days. Not set means no expiration",
			},
			"password_strength": dschema.Int64Attribute{
				Computed:    true,
				Description: "Minimum password strength. Not set means disabled, any password will be accepted. Values in the 50-70 range are suggested for common use cases.",
			},
			"password_policy": dschema.SingleNestedAttribute{
				Computed:    true,
				Description: "Static password complexity requirements. Whenever possible, prefer using the entropy-based approach provided by password_strength. " + enterpriseFeatureNote,
				Attributes: map[string]dschema.Attribute{
					"length": dschema.Int64Attribute{
						Optional: true,
					},
					"uppers": dschema.Int64Attribute{
						Optional: true,
					},
					"lowers": dschema.Int64Attribute{
						Optional: true,
					},
					"digits": dschema.Int64Attribute{
						Optional: true,
					},
					"specials": dschema.Int64Attribute{
						Optional: true,
					},
				},
			},
			"access_time": dschema.ListNestedAttribute{
				Computed:    true,
				Description: "Time periods in which access is allowed",
				NestedObject: dschema.NestedAttributeObject{
					Attributes: map[string]dschema.Attribute{
						"day_of_week": dschema.Int64Attribute{
							Computed:    true,
							Description: "Day of week, 0 Sunday, 6 Saturday",
						},
						"from": dschema.StringAttribute{
							Computed:    true,
							Description: "Start time in HH:MM format",
						},
						"to": dschema.StringAttribute{
							Computed:    true,
							Description: "End time in HH:MM format",
						},
					},
				},
			},
			"enforce_secure_algorithms": dschema.BoolAttribute{
				Computed:    true,
				Description: "If enabled, only secure algorithms are allowed. This setting is currently enforced for SSH/SFTP. " + enterpriseFeatureNote + ".",
			},
		},
	}
	if isGroup {
		result.Attributes["share_policy"] = dschema.SingleNestedAttribute{
			Computed:    true,
			Description: "Share access rules. " + enterpriseFeatureNote,
			Attributes: map[string]dschema.Attribute{
				"permissions": dschema.Int64Attribute{
					Computed: true,
				},
				"mode": dschema.Int64Attribute{
					Computed: true,
				},
			},
		}
		return result
	}
	result.Attributes["require_password_change"] = dschema.BoolAttribute{
		Computed:    true,
		Description: "If set, user must change their password from WebClient/REST API at next login.",
	}
	result.Attributes["tls_certs"] = dschema.ListAttribute{
		ElementType: types.StringType,
		Computed:    true,
		Description: "TLS certificates for mutual authentication. If provided will be checked before TLS username.",
	}
	result.Attributes["additional_emails"] = dschema.ListAttribute{
		ElementType: types.StringType,
		Computed:    true,
		Description: "Additional email addresses.",
	}
	result.Attributes["custom1"] = dschema.StringAttribute{
		Computed:    true,
		Description: `An extra placeholder value available for use in group configurations. It can be referenced as %custom1%. Deprecated: use custom_placeholders instead. ` + enterpriseFeatureNote + ".",
	}
	result.Attributes["custom_placeholders"] = dschema.ListAttribute{
		ElementType: types.StringType,
		Computed:    true,
		Description: "List of extra placeholders available for use in group configurations. Each placeholder can be referenced as %custom1%, %custom2%, and so on. " + enterpriseFeatureNote + ".",
	}
	return result
}

func getSchemaForUserFilters(isGroup bool) rschema.SingleNestedAttribute {
	result := rschema.SingleNestedAttribute{
		Optional: true,
		Computed: true,
		Attributes: map[string]rschema.Attribute{
			"allowed_ip": rschema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: `Only connections from these IP/Mask are allowed. IP/Mask must be in CIDR notation as defined in RFC 4632 and RFC 4291 for example "192.0.2.0/24" or "2001:db8::/32"`,
			},
			"denied_ip": rschema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Connections from these IP/Mask are allowed. Denied rules will be evaluated before allowed ones.",
			},
			"denied_login_methods": rschema.ListAttribute{
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
			"denied_protocols": rschema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: `Disabled protocols. Valid values: SSH, FTP, DAV, HTTP`,
				Validators: []validator.List{
					listvalidator.UniqueValues(),
					listvalidator.ValueStringsAre(stringvalidator.OneOf("SSH", "FTP", "DAV", "HTTP")),
				},
			},
			"file_patterns": rschema.ListNestedAttribute{
				Optional: true,
				NestedObject: rschema.NestedAttributeObject{
					Attributes: map[string]rschema.Attribute{
						"path": rschema.StringAttribute{
							Required:    true,
							Description: "Virtual path, if no other specific filter is defined, the filter applies for sub directories too.",
						},
						"allowed_patterns": rschema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "Files/directories with these, case insensitive, patterns are allowed. Allowed file patterns are evaluated before the denied ones.",
						},
						"denied_patterns": rschema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "Files/directories with these, case insensitive, patterns are not allowed.",
						},
						"deny_policy": rschema.Int64Attribute{
							Optional:    true,
							Description: "Set to 1 to hide denied files/directories in directory listing.",
						},
					},
				},
			},
			"max_upload_file_size": rschema.Int64Attribute{
				Optional:    true,
				Description: "Max size allowed for a single upload. Unset means no limit.",
			},
			"tls_username": rschema.StringAttribute{
				Optional:    true,
				Description: `TLS certificate attribute to use as username. For FTP clients it must match the name provided using the "USER" command. For WebDAV, if no username is provided, the CN will be used as username. For WebDAV clients it must match the implicit or provided username.`,
				Validators: []validator.String{
					stringvalidator.OneOf("CommonName"),
				},
			},
			"external_auth_disabled": rschema.BoolAttribute{
				Optional:    true,
				Description: "If set, external auth hook will not be executed.",
			},
			"pre_login_disabled": rschema.BoolAttribute{
				Optional:    true,
				Description: "If set, external pre-login hook will not be executed.",
			},
			"check_password_disabled": rschema.BoolAttribute{
				Optional:    true,
				Description: "If set, check password hook will not be executed.",
			},
			"disable_fs_checks": rschema.BoolAttribute{
				Optional:    true,
				Description: "Disable checks for existence and automatic creation of home directory and virtual folders after user login.",
			},
			"web_client": rschema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: fmt.Sprintf("Web Client/user REST API restrictions. Valid values: %s. Only available in the Enterprise version: %s",
					strings.Join(client.WebClientOptions, ", "), strings.Join(client.EnterpriseWebClientOptions, ", ")),
				Validators: []validator.List{
					listvalidator.UniqueValues(),
					listvalidator.ValueStringsAre(stringvalidator.OneOf(client.WebClientOptions...)),
				},
			},
			"allow_api_key_auth": rschema.BoolAttribute{
				Optional:    true,
				Description: "If set, API Key authentication is allowed.",
			},
			"user_type": rschema.StringAttribute{
				Optional:    true,
				Description: "Hint for authentication plugins. Valid values: LDAPUser, OSUser",
				Validators: []validator.String{
					stringvalidator.OneOf("LDAPUser", "OSUser"),
				},
			},
			"bandwidth_limits": rschema.ListNestedAttribute{
				Optional: true,
				NestedObject: rschema.NestedAttributeObject{
					Attributes: map[string]rschema.Attribute{
						"sources": rschema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
							Description: `Source networks in CIDR notation as defined in RFC 4632 and RFC 4291 for example "192.0.2.0/24" or "2001:db8::/32". The limit applies if the defined networks contain the client IP.`,
						},
						"upload_bandwidth": rschema.Int64Attribute{
							Optional:    true,
							Description: "Maximum upload bandwidth as KB/s.",
						},
						"download_bandwidth": rschema.Int64Attribute{
							Optional:    true,
							Description: "Maximum download bandwidth as KB/s.",
						},
					},
				},
			},
			"external_auth_cache_time": rschema.Int64Attribute{
				Optional:    true,
				Description: "Defines the cache time, in seconds, for users authenticated using an external auth hook. Not set means no cache.",
			},
			"start_directory": rschema.StringAttribute{
				Optional:    true,
				Description: `Alternate starting directory. If not set, the default is "/". This option is supported for SFTP/SCP, FTP and HTTP (WebClient/REST API) protocols. Relative paths will use this directory as base.`,
			},
			"two_factor_protocols": rschema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Defines protocols that require two factor authentication. Valid values: SSH, FTP, HTTP",
				Validators: []validator.List{
					listvalidator.UniqueValues(),
					listvalidator.ValueStringsAre(stringvalidator.OneOf("SSH", "FTP", "HTTP")),
				},
			},
			"ftp_security": rschema.Int64Attribute{
				Optional:    true,
				Description: "FTP security mode. Set to 1 to require TLS for both data and control connection.",
			},
			"is_anonymous": rschema.BoolAttribute{
				Optional:    true,
				Description: `If enabled the user can login with any password or no password at all. Anonymous users are supported for FTP and WebDAV protocols and permissions will be automatically set to "list" and "download" (read only)`,
			},
			"default_shares_expiration": rschema.Int64Attribute{
				Optional:    true,
				Description: "Default expiration for newly created shares as number of days. Not set means no default expiration.",
			},
			"max_shares_expiration": rschema.Int64Attribute{
				Optional:    true,
				Description: "Maximum allowed expiration, as a number of days, when a user creates or updates a share. Not set means that non-expiring shares are allowed.",
			},
			"password_expiration": rschema.Int64Attribute{
				Optional:    true,
				Description: "The password expires after the defined number of days. Not set means no expiration",
			},
			"password_strength": rschema.Int64Attribute{
				Optional:    true,
				Description: "Minimum password strength. Not set means disabled, any password will be accepted. Values in the 50-70 range are suggested for common use cases.",
			},
			"password_policy": rschema.SingleNestedAttribute{
				Optional:    true,
				Description: "Static password complexity requirements. Whenever possible, prefer using the entropy-based approach provided by password_strength. " + enterpriseFeatureNote,
				Attributes: map[string]rschema.Attribute{
					"length": rschema.Int64Attribute{
						Optional: true,
					},
					"uppers": rschema.Int64Attribute{
						Optional: true,
					},
					"lowers": rschema.Int64Attribute{
						Optional: true,
					},
					"digits": rschema.Int64Attribute{
						Optional: true,
					},
					"specials": rschema.Int64Attribute{
						Optional: true,
					},
				},
			},
			"access_time": rschema.ListNestedAttribute{
				Optional:    true,
				Description: "Time periods in which access is allowed",
				NestedObject: rschema.NestedAttributeObject{
					Attributes: map[string]rschema.Attribute{
						"day_of_week": rschema.Int64Attribute{
							Required:    true,
							Description: "Day of week, 0 Sunday, 6 Saturday",
							Validators: []validator.Int64{
								int64validator.Between(0, 6),
							},
						},
						"from": rschema.StringAttribute{
							Required:    true,
							Description: "Start time in HH:MM format",
						},
						"to": rschema.StringAttribute{
							Required:    true,
							Description: "End time in HH:MM format",
						},
					},
				},
			},
			"enforce_secure_algorithms": rschema.BoolAttribute{
				Optional:    true,
				Description: "If enabled, only secure algorithms are allowed. This setting is currently enforced for SSH/SFTP. " + enterpriseFeatureNote + ".",
			},
		},
	}
	if isGroup {
		result.Attributes["share_policy"] = rschema.SingleNestedAttribute{
			Optional:    true,
			Description: "Share access rules. " + enterpriseFeatureNote,
			Attributes: map[string]rschema.Attribute{
				"permissions": rschema.Int64Attribute{
					Optional: true,
					MarkdownDescription: "Bitmask of permissions. Sum the values to combine permissions.\n\n" +
						"Supported values:\n" +
						"* `0`: None\n" +
						"* `1`: Read\n" +
						"* `2`: Write\n" +
						"* `4`: Delete\n" +
						"* `7`: All",
				},
				"mode": rschema.Int64Attribute{
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
	result.Attributes["require_password_change"] = rschema.BoolAttribute{
		Optional:    true,
		Description: "If set, user must change their password from WebClient/REST API at next login.",
	}
	result.Attributes["tls_certs"] = rschema.ListAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Description: "TLS certificates for mutual authentication. If provided will be checked before TLS username.",
		Validators: []validator.List{
			listvalidator.UniqueValues(),
		},
	}
	result.Attributes["additional_emails"] = rschema.ListAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Description: "Additional email addresses.",
		Validators: []validator.List{
			listvalidator.UniqueValues(),
		},
	}
	result.Attributes["custom1"] = rschema.StringAttribute{
		Optional:    true,
		Description: `An extra placeholder value available for use in group configurations. It can be referenced as %custom1%. ` + enterpriseFeatureNote + ".",
	}
	result.Attributes["custom_placeholders"] = rschema.ListAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Description: "List of extra placeholders available for use in group configurations. Each placeholder can be referenced as %custom1%, %custom2%, and so on. " + enterpriseFeatureNote + ".",
	}
	return result
}

func preserveFsConfigPlanFields(ctx context.Context, fsPlan, fsState filesystem) (types.Object, diag.Diagnostics) {
	switch sdk.FilesystemProvider(fsState.Provider.ValueInt64()) {
	case sdk.S3FilesystemProvider:
		if fsPlan.S3Config != nil {
			fsState.S3Config.AccessSecret = fsPlan.S3Config.AccessSecret
			fsState.S3Config.AccessSecretWoVersion = fsPlan.S3Config.AccessSecretWoVersion
			fsState.S3Config.SSECustomerKey = fsPlan.S3Config.SSECustomerKey
			fsState.S3Config.SSECustomerKeyWoVersion = fsPlan.S3Config.SSECustomerKeyWoVersion
		}
	case sdk.GCSFilesystemProvider:
		if fsPlan.GCSConfig != nil {
			fsState.GCSConfig.Credentials = fsPlan.GCSConfig.Credentials
			fsState.GCSConfig.CredentialsWoVersion = fsPlan.GCSConfig.CredentialsWoVersion
		}
	case sdk.AzureBlobFilesystemProvider:
		if fsPlan.AzBlobConfig != nil {
			fsState.AzBlobConfig.AccountKey = fsPlan.AzBlobConfig.AccountKey
			fsState.AzBlobConfig.AccountKeyWoVersion = fsPlan.AzBlobConfig.AccountKeyWoVersion
			fsState.AzBlobConfig.SASURL = fsPlan.AzBlobConfig.SASURL
			fsState.AzBlobConfig.SASURLWoVersion = fsPlan.AzBlobConfig.SASURLWoVersion
		}
	case sdk.CryptedFilesystemProvider:
		if fsPlan.CryptConfig != nil {
			fsState.CryptConfig.Passphrase = fsPlan.CryptConfig.Passphrase
			fsState.CryptConfig.PassphraseWoVersion = fsPlan.CryptConfig.PassphraseWoVersion
		}
	case sdk.SFTPFilesystemProvider:
		if fsPlan.SFTPConfig != nil {
			fsState.SFTPConfig.Password = fsPlan.SFTPConfig.Password
			fsState.SFTPConfig.PasswordWoVersion = fsPlan.SFTPConfig.PasswordWoVersion
			fsState.SFTPConfig.PrivateKey = fsPlan.SFTPConfig.PrivateKey
			fsState.SFTPConfig.PrivateKeyWoVersion = fsPlan.SFTPConfig.PrivateKeyWoVersion
			fsState.SFTPConfig.KeyPassphrase = fsPlan.SFTPConfig.KeyPassphrase
			fsState.SFTPConfig.KeyPassphraseWoVersion = fsPlan.SFTPConfig.KeyPassphraseWoVersion
			fsState.SFTPConfig.SocksPassword = fsPlan.SFTPConfig.SocksPassword
			fsState.SFTPConfig.SocksPasswordWoVersion = fsPlan.SFTPConfig.SocksPasswordWoVersion
		}
	case sdk.HTTPFilesystemProvider:
		if fsPlan.HTTPConfig != nil {
			fsState.HTTPConfig.Password = fsPlan.HTTPConfig.Password
			fsState.HTTPConfig.PasswordWoVersion = fsPlan.HTTPConfig.PasswordWoVersion
			fsState.HTTPConfig.APIKey = fsPlan.HTTPConfig.APIKey
			fsState.HTTPConfig.APIKeyWoVersion = fsPlan.HTTPConfig.APIKeyWoVersion
		}
	case client.FTPFilesystemProvider:
		if fsPlan.FTPConfig != nil {
			fsState.FTPConfig.Password = fsPlan.FTPConfig.Password
			fsState.FTPConfig.PasswordWoVersion = fsPlan.FTPConfig.PasswordWoVersion
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
