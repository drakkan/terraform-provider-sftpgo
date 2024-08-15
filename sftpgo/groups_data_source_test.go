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
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sftpgo/sdk"
	"github.com/sftpgo/sdk/kms"
	"github.com/stretchr/testify/require"
)

var (
	testGroup = sdk.Group{
		BaseGroup: sdk.BaseGroup{
			Name: "test group",
		},
		UserSettings: sdk.GroupUserSettings{
			BaseGroupUserSettings: sdk.BaseGroupUserSettings{
				Permissions: map[string][]string{
					"/":   {"*"},
					"/p1": {"list", "download"},
				},
				ExpiresIn: 10,
				Filters: sdk.BaseUserFilters{
					AllowedIP:         []string{"172.16.0.0/16"},
					MaxUploadFileSize: 10000000,
					Hooks: sdk.HooksFilter{
						ExternalAuthDisabled: true,
					},
				},
			},
			FsConfig: sdk.Filesystem{
				Provider: 4,
				CryptConfig: sdk.CryptFsConfig{
					Passphrase: kms.BaseSecret{
						Status:  kms.SecretStatusPlain,
						Payload: "secret passphrase",
					},
					OSFsConfig: sdk.OSFsConfig{
						ReadBufferSize:  5,
						WriteBufferSize: 0,
					},
				},
			},
		},
		VirtualFolders: []sdk.VirtualFolder{
			{
				BaseVirtualFolder: sdk.BaseVirtualFolder{
					Name: testFolder.Name,
				},
				VirtualPath: "/vpath",
				QuotaSize:   -1,
				QuotaFiles:  -1,
			},
		},
	}
)

func TestAccGroupsDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	c, err := getClient()
	require.NoError(t, err)
	_, err = c.CreateFolder(testFolder)
	require.NoError(t, err)
	_, err = c.CreateGroup(testGroup)
	require.NoError(t, err)

	defer func() {
		err = c.DeleteGroup(testGroup.Name)
		require.NoError(t, err)
		err = c.DeleteFolder(testFolder.Name)
		require.NoError(t, err)
	}()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: `data "sftpgo_groups" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of groups returned
					resource.TestCheckResourceAttr("data.sftpgo_groups.test", "groups.#", "1"),
					// Check the groups fields
					resource.TestCheckResourceAttr("data.sftpgo_groups.test", "groups.0.name", testGroup.Name),
					resource.TestCheckResourceAttr("data.sftpgo_groups.test", "groups.0.id", testGroup.Name),
					resource.TestCheckResourceAttrSet("data.sftpgo_groups.test", "groups.0.created_at"),
					resource.TestCheckResourceAttrSet("data.sftpgo_groups.test", "groups.0.updated_at"),
					resource.TestCheckNoResourceAttr("data.sftpgo_groups.test", "groups.0.description"),
					resource.TestCheckResourceAttr("data.sftpgo_groups.test", "groups.0.user_settings.permissions.%", "2"),
					resource.TestCheckResourceAttr("data.sftpgo_groups.test", "groups.0.user_settings.permissions./", "*"),
					resource.TestCheckResourceAttr("data.sftpgo_groups.test", "groups.0.user_settings.permissions./p1", "list,download"),
					resource.TestCheckResourceAttr("data.sftpgo_groups.test", "groups.0.user_settings.expires_in", "10"),
					resource.TestCheckNoResourceAttr("data.sftpgo_groups.test", "groups.0.user_settings.home_dir"),
					resource.TestCheckNoResourceAttr("data.sftpgo_groups.test", "groups.0.user_settings.max_sessions"),
					resource.TestCheckResourceAttr("data.sftpgo_groups.test", "groups.0.user_settings.filters.allowed_ip.#", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_groups.test", "groups.0.user_settings.filters.allowed_ip.0", "172.16.0.0/16"),
					resource.TestCheckResourceAttr("data.sftpgo_groups.test", "groups.0.user_settings.filters.max_upload_file_size", "10000000"),
					resource.TestCheckResourceAttr("data.sftpgo_groups.test", "groups.0.user_settings.filters.external_auth_disabled", "true"),
					resource.TestCheckNoResourceAttr("data.sftpgo_groups.test", "groups.0.user_settings.filters.pre_login_disabled"),
					resource.TestCheckNoResourceAttr("data.sftpgo_groups.test", "groups.0.user_settings.filters.web_client"),
					resource.TestCheckResourceAttr("data.sftpgo_groups.test", "groups.0.user_settings.filesystem.provider", "4"),
					resource.TestCheckResourceAttrSet("data.sftpgo_groups.test", "groups.0.user_settings.filesystem.cryptconfig.passphrase"),
					resource.TestCheckResourceAttr("data.sftpgo_groups.test", "groups.0.user_settings.filesystem.cryptconfig.read_buffer_size", "5"),
					resource.TestCheckNoResourceAttr("data.sftpgo_groups.test", "groups.0.user_settings.filesystem.cryptconfig.write_buffer_size"),
					resource.TestCheckNoResourceAttr("data.sftpgo_groups.test", "groups.0.user_settings.filesystem.sftpconfig"),
					resource.TestCheckNoResourceAttr("data.sftpgo_groups.test", "groups.0.user_settings.filesystem.httpconfig"),
					resource.TestCheckResourceAttr("data.sftpgo_groups.test", "groups.0.virtual_folders.#", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_groups.test", "groups.0.virtual_folders.0.name", testFolder.Name),
					resource.TestCheckResourceAttr("data.sftpgo_groups.test", "groups.0.virtual_folders.0.virtual_path",
						testGroup.VirtualFolders[0].VirtualPath),
					resource.TestCheckResourceAttr("data.sftpgo_groups.test", "groups.0.virtual_folders.0.quota_size",
						fmt.Sprintf("%d", testGroup.VirtualFolders[0].QuotaSize)),
					resource.TestCheckResourceAttr("data.sftpgo_groups.test", "groups.0.virtual_folders.0.quota_files",
						fmt.Sprintf("%d", testGroup.VirtualFolders[0].QuotaFiles)),
					resource.TestCheckResourceAttrSet("data.sftpgo_groups.test", "groups.0.virtual_folders.0.used_quota_size"),
					resource.TestCheckResourceAttrSet("data.sftpgo_groups.test", "groups.0.virtual_folders.0.used_quota_files"),
					resource.TestCheckResourceAttrSet("data.sftpgo_groups.test", "groups.0.virtual_folders.0.last_quota_update"),
					resource.TestCheckResourceAttr("data.sftpgo_groups.test", "groups.0.virtual_folders.0.filesystem.provider",
						fmt.Sprintf("%d", testFolder.FsConfig.Provider)),
					resource.TestCheckResourceAttr("data.sftpgo_groups.test", "groups.0.virtual_folders.0.filesystem.s3config.bucket",
						testFolder.FsConfig.S3Config.Bucket),
					resource.TestCheckResourceAttrSet("data.sftpgo_groups.test", "groups.0.virtual_folders.0.filesystem.s3config.access_secret"),
					resource.TestCheckResourceAttrSet("data.sftpgo_groups.test", "groups.0.virtual_folders.0.filesystem.s3config.sse_customer_key"),
					resource.TestCheckNoResourceAttr("data.sftpgo_groups.test", "groups.0.virtual_folders.0.filesystem.gcsconfig"),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.sftpgo_groups.test", "id", placeholderID),
				),
			},
		},
	})
}
