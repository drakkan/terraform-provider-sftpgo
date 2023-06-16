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
			Filters: sdk.UserFilters{
				BaseUserFilters: sdk.BaseUserFilters{
					DeniedProtocols:  []string{"SSH"},
					PasswordStrength: 75,
				},
				RequirePasswordChange: true,
			},
			FsConfig: sdk.Filesystem{
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
			VirtualFolders: []sdk.VirtualFolder{
				{
					BaseVirtualFolder: sdk.BaseVirtualFolder{
						Name: testFolder.Name,
					},
					VirtualPath: "/vpath",
					QuotaSize:   1000000,
					QuotaFiles:  100,
				},
			},
		},
		Password: "Cheiha0ahy7Ieghatiet4phei",
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
