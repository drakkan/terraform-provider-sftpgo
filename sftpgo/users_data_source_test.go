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

func TestAccUsersDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	c, err := getClient()
	require.NoError(t, err)
	user := client.User{
		User: sdk.User{
			BaseUser: sdk.BaseUser{
				Username:       "test user",
				Status:         1,
				Email:          "user@sftpgo.com",
				ExpirationDate: 1680800030000,
				PublicKeys:     []string{"ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEUWwDwEWhTbF0MqAsp/oXK1HR2cElhM8oo1uVmL3ZeDKDiTm4ljMr92wfTgIGDqIoxmVqgYIkAOAhuykAVWBzc= user@host"},
				HomeDir:        filepath.Join(os.TempDir(), "user"),
				UID:            1001,
				MaxSessions:    10,
				Permissions: map[string][]string{
					"/": {"*"},
				},
				TotalDataTransfer: 25678,
				AdditionalInfo:    "Terraform",
				Groups: []sdk.GroupMapping{
					{
						Name: testGroup.Name,
						Type: sdk.GroupTypeSecondary,
					},
				},
				Role: testRole.Name,
			},
		},
		Password: "Cheiha0ahy7Ieghatiet4phei",
		FsConfig: client.Filesystem{
			Provider: 6,
			HTTPConfig: sdk.HTTPFsConfig{
				BaseHTTPFsConfig: sdk.BaseHTTPFsConfig{
					Endpoint: "http://127.0.0.1:8080",
				},
				APIKey: kms.BaseSecret{
					Status:  kms.SecretStatusPlain,
					Payload: "api key",
				},
			},
		},
		VirtualFolders: []client.VirtualFolder{
			{
				BaseVirtualFolder: client.BaseVirtualFolder{
					BaseVirtualFolder: sdk.BaseVirtualFolder{
						Name: testFolder.Name,
					},
				},
				VirtualPath: "/vpath",
				QuotaSize:   1000000,
				QuotaFiles:  100,
			},
		},
		Filters: client.UserFilters{
			BaseUserFilters: client.BaseUserFilters{
				DeniedProtocols:  []string{"SSH"},
				PasswordStrength: 75,
				AccessTime: []sdk.TimePeriod{
					{
						DayOfWeek: 1,
						From:      "12:03",
						To:        "14:05",
					},
				},
			},
			RequirePasswordChange: true,
		},
	}
	_, err = c.CreateRole(testRole)
	require.NoError(t, err)
	_, err = c.CreateFolder(testFolder)
	require.NoError(t, err)
	_, err = c.CreateGroup(testGroup)
	require.NoError(t, err)
	_, err = c.CreateUser(user)
	require.NoError(t, err)

	defer func() {
		err = c.DeleteUser(user.Username)
		require.NoError(t, err)
		err = c.DeleteGroup(testGroup.Name)
		require.NoError(t, err)
		err = c.DeleteFolder(testFolder.Name)
		require.NoError(t, err)
		err = c.DeleteRole(testRole.Name)
		require.NoError(t, err)
	}()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: `data "sftpgo_users" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of users returned
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.#", "1"),
					// Check the users fields
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.username", user.Username),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.id", user.Username),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.status", fmt.Sprintf("%d", user.Status)),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.email", user.Email),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.expiration_date", fmt.Sprintf("%d", user.ExpirationDate)),
					resource.TestCheckResourceAttrSet("data.sftpgo_users.test", "users.0.password"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.public_keys.0", user.PublicKeys[0]),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.home_dir", user.HomeDir),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.uid", fmt.Sprintf("%d", user.UID)),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.max_sessions", fmt.Sprintf("%d", user.MaxSessions)),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.permissions.%", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.permissions./", "*"),
					resource.TestCheckNoResourceAttr("data.sftpgo_users.test", "users.0.last_quota_update"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.total_data_transfer", fmt.Sprintf("%d", user.TotalDataTransfer)),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.additional_info", user.AdditionalInfo),
					resource.TestCheckResourceAttrSet("data.sftpgo_users.test", "users.0.created_at"),
					resource.TestCheckResourceAttrSet("data.sftpgo_users.test", "users.0.updated_at"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.groups.#", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.groups.0.name", testGroup.Name),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.groups.0.type", "2"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.role", testRole.Name),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.filters.denied_protocols.0", "SSH"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.filters.password_strength",
						fmt.Sprintf("%d", user.Filters.PasswordStrength)),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.filters.access_time.#", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.filters.access_time.0.day_of_week", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.filters.access_time.0.from", "12:03"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.filters.access_time.0.to", "14:05"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.filters.require_password_change", "true"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.filesystem.provider", fmt.Sprintf("%d", user.FsConfig.Provider)),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.filesystem.httpconfig.endpoint", user.FsConfig.HTTPConfig.Endpoint),
					resource.TestCheckResourceAttrSet("data.sftpgo_users.test", "users.0.filesystem.httpconfig.api_key"),
					resource.TestCheckNoResourceAttr("data.sftpgo_users.test", "users.0.filesystem.osconfig"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.virtual_folders.#", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.virtual_folders.0.name", testFolder.Name),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.virtual_folders.0.quota_size",
						fmt.Sprintf("%d", user.VirtualFolders[0].QuotaSize)),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.virtual_folders.0.quota_files",
						fmt.Sprintf("%d", user.VirtualFolders[0].QuotaFiles)),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "id", placeholderID),
				),
			},
		},
	})
}

func TestAccEnterpriseUsersDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}

	c, err := getClient()
	require.NoError(t, err)
	if !c.IsEnterpriseEdition() {
		t.Skip("This test is supported only with the Enterprise edition")
	}
	user1 := client.User{
		User: sdk.User{
			BaseUser: sdk.BaseUser{
				Username:   "user1",
				Status:     1,
				Email:      "user1@sftpgo.com",
				PublicKeys: []string{"ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEUWwDwEWhTbF0MqAsp/oXK1HR2cElhM8oo1uVmL3ZeDKDiTm4ljMr92wfTgIGDqIoxmVqgYIkAOAhuykAVWBzc= user@host"},
				HomeDir:    filepath.Join(os.TempDir(), "user1"),
				Permissions: map[string][]string{
					"/": {"*"},
				},
			},
		},
		Password: "Cheiha0ahy7Ieghatiet4phei",
		FsConfig: client.Filesystem{
			Provider: sdk.SFTPFilesystemProvider,
			SFTPConfig: client.SFTPFsConfig{
				BaseSFTPFsConfig: client.BaseSFTPFsConfig{
					Endpoint:       "127.0.0.1:2022",
					Username:       "testuser",
					Socks5Proxy:    "127.0.0.1:1080",
					Socks5Username: "socks_user",
				},
				Password: kms.BaseSecret{
					Status:  kms.SecretStatusPlain,
					Payload: "sftppass",
				},
				Socks5Password: kms.BaseSecret{
					Status:  kms.SecretStatusPlain,
					Payload: "sockspass",
				},
			},
		},
		Filters: client.UserFilters{
			BaseUserFilters: client.BaseUserFilters{
				EnforceSecureAlgorithms: true,
				WebClient: []string{"shares-require-email-auth",
					"wopi-disabled", "rest-api-disabled", sdk.WebClientInfoChangeDisabled},
			},
		},
	}

	user2 := client.User{
		User: sdk.User{
			BaseUser: sdk.BaseUser{
				Username: "user2",
				Status:   1,
				HomeDir:  filepath.Join(os.TempDir(), "user1"),
				Permissions: map[string][]string{
					"/": {"list", "download"},
				},
			},
		},
		FsConfig: client.Filesystem{
			Provider: sdk.GCSFilesystemProvider,
			GCSConfig: client.GCSFsConfig{
				BaseGCSFsConfig: client.BaseGCSFsConfig{
					Bucket:                "gcs_hns",
					KeyPrefix:             "users/user2/",
					AutomaticCredentials:  1,
					HierarchicalNamespace: 1,
				},
			},
		},
	}

	_, err = c.CreateUser(user1)
	require.NoError(t, err)
	_, err = c.CreateUser(user2)
	require.NoError(t, err)

	defer func() {
		err = c.DeleteUser(user1.Username)
		require.NoError(t, err)
		err = c.DeleteUser(user2.Username)
		require.NoError(t, err)
	}()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: `data "sftpgo_users" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of users returned
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.#", "2"),
					// Check the users fields
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.username", user1.Username),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.id", user1.Username),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.status", fmt.Sprintf("%d", user1.Status)),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.email", user1.Email),
					resource.TestCheckResourceAttrSet("data.sftpgo_users.test", "users.0.password"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.public_keys.0", user1.PublicKeys[0]),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.home_dir", user1.HomeDir),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.permissions.%", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.permissions./", "*"),
					resource.TestCheckNoResourceAttr("data.sftpgo_users.test", "users.0.last_quota_update"),
					resource.TestCheckResourceAttrSet("data.sftpgo_users.test", "users.0.created_at"),
					resource.TestCheckResourceAttrSet("data.sftpgo_users.test", "users.0.updated_at"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.groups.#", "0"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.filters.web_client.#", "4"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.filters.web_client.0", "shares-require-email-auth"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.filters.web_client.1", "wopi-disabled"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.filters.web_client.2", "rest-api-disabled"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.filters.web_client.3", sdk.WebClientInfoChangeDisabled),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.filters.enforce_secure_algorithms", "true"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.filesystem.provider", fmt.Sprintf("%d", user1.FsConfig.Provider)),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.filesystem.sftpconfig.endpoint", user1.FsConfig.SFTPConfig.Endpoint),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.filesystem.sftpconfig.username", user1.FsConfig.SFTPConfig.Username),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.0.filesystem.sftpconfig.socks5_proxy", user1.FsConfig.SFTPConfig.Socks5Proxy),
					resource.TestCheckResourceAttrSet("data.sftpgo_users.test", "users.0.filesystem.sftpconfig.socks5_password"),
					resource.TestCheckNoResourceAttr("data.sftpgo_users.test", "users.0.filesystem.osconfig"),
					// Check user2
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.1.username", user2.Username),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.1.id", user2.Username),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.1.status", fmt.Sprintf("%d", user2.Status)),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.1.home_dir", user2.HomeDir),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.1.permissions.%", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.1.permissions./", "list,download"),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.1.filesystem.provider", fmt.Sprintf("%d", user2.FsConfig.Provider)),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.1.filesystem.gcsconfig.bucket", user2.FsConfig.GCSConfig.Bucket),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.1.filesystem.gcsconfig.key_prefix", user2.FsConfig.GCSConfig.KeyPrefix),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.1.filesystem.gcsconfig.automatic_credentials", strconv.Itoa(user2.FsConfig.GCSConfig.AutomaticCredentials)),
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "users.1.filesystem.gcsconfig.hns", strconv.Itoa(user2.FsConfig.GCSConfig.HierarchicalNamespace)),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.sftpgo_users.test", "id", placeholderID),
				),
			},
		},
	})
}
