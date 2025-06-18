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
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sftpgo/sdk"
	"github.com/sftpgo/sdk/kms"
	"github.com/stretchr/testify/require"
)

var (
	testFolder = client.BaseVirtualFolder{
		BaseVirtualFolder: sdk.BaseVirtualFolder{
			Name:        "tfolder",
			MappedPath:  filepath.Join(os.TempDir(), "tfolder"),
			Description: "desc",
		},
		FsConfig: client.Filesystem{
			Provider: 1,
			S3Config: sdk.S3FsConfig{
				BaseS3FsConfig: sdk.BaseS3FsConfig{
					Bucket:           "s3bucket",
					AccessKey:        "my key",
					Region:           "us-west-1",
					DownloadPartSize: 100,
					SkipTLSVerify:    true,
				},
				AccessSecret: kms.BaseSecret{
					Status:  kms.SecretStatusPlain,
					Payload: "s3secret",
				},
				SSECustomerKey: kms.BaseSecret{
					Status:  kms.SecretStatusPlain,
					Payload: "secretk3y",
				},
			},
		},
	}
)

func TestAccFoldersDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	c, err := getClient()
	require.NoError(t, err)
	_, err = c.CreateFolder(testFolder)
	require.NoError(t, err)

	defer func() {
		err = c.DeleteFolder(testFolder.Name)
		require.NoError(t, err)
	}()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: `data "sftpgo_folders" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of folders returned
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.#", "1"),
					// Check the folder fields
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.0.name", testFolder.Name),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.0.id", testFolder.Name),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.0.mapped_path", testFolder.MappedPath),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.0.description", testFolder.Description),
					resource.TestCheckResourceAttrSet("data.sftpgo_folders.test", "folders.0.used_quota_size"),
					resource.TestCheckResourceAttrSet("data.sftpgo_folders.test", "folders.0.used_quota_files"),
					resource.TestCheckResourceAttrSet("data.sftpgo_folders.test", "folders.0.last_quota_update"),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.provider",
						fmt.Sprintf("%d", testFolder.FsConfig.Provider)),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.s3config.bucket",
						testFolder.FsConfig.S3Config.Bucket),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.s3config.region",
						testFolder.FsConfig.S3Config.Region),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.s3config.access_key",
						testFolder.FsConfig.S3Config.AccessKey),
					resource.TestCheckResourceAttrSet("data.sftpgo_folders.test", "folders.0.filesystem.s3config.access_secret"),
					resource.TestCheckResourceAttrSet("data.sftpgo_folders.test", "folders.0.filesystem.s3config.sse_customer_key"),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.s3config.download_part_size",
						fmt.Sprintf("%d", testFolder.FsConfig.S3Config.DownloadPartSize)),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.s3config.skip_tls_verify",
						"true"),
					resource.TestCheckNoResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.gcsconfig"),
					resource.TestCheckNoResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.azblobconfig"),
					resource.TestCheckNoResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.cryptconfig"),
					resource.TestCheckNoResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.sftpconfig"),
					resource.TestCheckNoResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.httpconfig"),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "id", placeholderID),
				),
			},
		},
	})
}

func TestAccEnterpriseFoldersDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	c, err := getClient()
	require.NoError(t, err)
	if !c.IsEnterpriseEdition() {
		t.Skip("This test is supported only with the Enterprise edition")
	}
	folder1 := client.BaseVirtualFolder{
		BaseVirtualFolder: sdk.BaseVirtualFolder{
			Name:        "folder1",
			MappedPath:  filepath.Join(os.TempDir(), "folder1"),
			Description: "desc",
		},
		FsConfig: client.Filesystem{
			Provider: sdk.SFTPFilesystemProvider,
			SFTPConfig: client.SFTPFsConfig{
				BaseSFTPFsConfig: client.BaseSFTPFsConfig{
					Endpoint:    "127.0.0.1:22",
					Username:    "sftpuser",
					Socks5Proxy: "127.0.0.1:10080",
				},
				Password: kms.BaseSecret{
					Status:  kms.SecretStatusPlain,
					Payload: "sftps3secret",
				},
			},
		},
	}

	folder2 := client.BaseVirtualFolder{
		BaseVirtualFolder: sdk.BaseVirtualFolder{
			Name:       "folder2",
			MappedPath: filepath.Join(os.TempDir(), "folder2"),
		},
		FsConfig: client.Filesystem{
			Provider: sdk.GCSFilesystemProvider,
			GCSConfig: client.GCSFsConfig{
				BaseGCSFsConfig: client.BaseGCSFsConfig{
					Bucket:                "bucket",
					AutomaticCredentials:  1,
					HierarchicalNamespace: 1,
					UploadPartSize:        10,
					UploadPartMaxTime:     60,
				},
			},
		},
	}

	_, err = c.CreateFolder(folder1)
	require.NoError(t, err)
	_, err = c.CreateFolder(folder2)
	require.NoError(t, err)

	defer func() {
		err = c.DeleteFolder(folder1.Name)
		require.NoError(t, err)
		err = c.DeleteFolder(folder2.Name)
		require.NoError(t, err)
	}()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: `data "sftpgo_folders" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of folders returned
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.#", "2"),
					// Check the folder fields
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.0.name", folder1.Name),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.0.id", folder1.Name),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.0.mapped_path", folder1.MappedPath),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.0.description", folder1.Description),
					resource.TestCheckResourceAttrSet("data.sftpgo_folders.test", "folders.0.used_quota_size"),
					resource.TestCheckResourceAttrSet("data.sftpgo_folders.test", "folders.0.used_quota_files"),
					resource.TestCheckResourceAttrSet("data.sftpgo_folders.test", "folders.0.last_quota_update"),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.provider",
						strconv.Itoa(int(folder1.FsConfig.Provider))),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.sftpconfig.endpoint",
						folder1.FsConfig.SFTPConfig.Endpoint),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.sftpconfig.username",
						folder1.FsConfig.SFTPConfig.Username),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.sftpconfig.socks5_proxy",
						folder1.FsConfig.SFTPConfig.Socks5Proxy),
					resource.TestCheckNoResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.sftpconfig.socks5_username"),
					resource.TestCheckNoResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.sftpconfig.socks5_password"),
					resource.TestCheckNoResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.gcsconfig"),
					resource.TestCheckNoResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.azblobconfig"),
					resource.TestCheckNoResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.cryptconfig"),
					resource.TestCheckNoResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.s3config"),
					resource.TestCheckNoResourceAttr("data.sftpgo_folders.test", "folders.0.filesystem.httpconfig"),

					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.1.name", folder2.Name),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.1.id", folder2.Name),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.1.mapped_path", folder2.MappedPath),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.1.filesystem.provider",
						strconv.Itoa(int(folder2.FsConfig.Provider))),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.1.filesystem.gcsconfig.bucket",
						folder2.FsConfig.GCSConfig.Bucket),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.1.filesystem.gcsconfig.automatic_credentials", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.1.filesystem.gcsconfig.hns", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.1.filesystem.gcsconfig.upload_part_size", "10"),
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "folders.1.filesystem.gcsconfig.upload_part_max_time", "60"),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.sftpgo_folders.test", "id", placeholderID),
				),
			},
		},
	})
}
