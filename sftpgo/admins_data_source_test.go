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
	"github.com/stretchr/testify/require"

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo/client"
)

func TestAccAdminsDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	c, err := getClient()
	require.NoError(t, err)
	admin := client.Admin{
		Username:       "test_admin",
		Password:       "test_password",
		Permissions:    []string{"add_users", "edit_users", "view_status", "quota_scans", "manage_folders", "disable_mfa"},
		Email:          "admin@sftpgo.com",
		Description:    "Created from Terraform",
		AdditionalInfo: "TF",
		Filters: client.AdminFilters{
			AllowList:             []string{"127.0.0.1/8", "192.168.1.0/24"},
			AllowAPIKeyAuth:       true,
			RequirePasswordChange: true,
			RequireTwoFactor:      true,
			Preferences: client.AdminPreferences{
				HideUserPageSections:   96,
				DefaultUsersExpiration: 15,
			},
		},
		Groups: []client.AdminGroupMapping{
			{
				Name: testGroup.Name,
				Options: client.AdminGroupMappingOptions{
					AddToUsersAs: 2,
				},
			},
		},
		Role: testRole.Name,
	}
	_, err = c.CreateRole(testRole)
	require.NoError(t, err)
	_, err = c.CreateFolder(testFolder)
	require.NoError(t, err)
	_, err = c.CreateGroup(testGroup)
	require.NoError(t, err)
	_, err = c.CreateAdmin(admin)
	require.NoError(t, err)

	defer func() {
		err = c.DeleteAdmin(admin.Username)
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
				Config: `data "sftpgo_admins" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of admins returned
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.#", "2"),
					// The first admin is the default one
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.0.username", "admin"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.0.id", "admin"),
					resource.TestCheckResourceAttrSet("data.sftpgo_admins.test", "admins.0.password"),
					resource.TestCheckNoResourceAttr("data.sftpgo_admins.test", "admins.0.email"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.0.status", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.0.permissions.#", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.0.permissions.0", "*"),
					resource.TestCheckResourceAttrSet("data.sftpgo_admins.test", "admins.0.created_at"),
					resource.TestCheckResourceAttrSet("data.sftpgo_admins.test", "admins.0.updated_at"),
					resource.TestCheckResourceAttrSet("data.sftpgo_admins.test", "admins.0.last_login"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.0.filters.%", "4"),
					resource.TestCheckNoResourceAttr("data.sftpgo_admins.test", "admins.0.filters.allow_api_key_auth"),
					resource.TestCheckNoResourceAttr("data.sftpgo_admins.test", "admins.0.filters.require_password_change"),
					resource.TestCheckNoResourceAttr("data.sftpgo_admins.test", "admins.0.filters.require_two_factor"),
					resource.TestCheckNoResourceAttr("data.sftpgo_admins.test", "admins.0.filters.allow_list"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.0.preferences.%", "2"),
					resource.TestCheckNoResourceAttr("data.sftpgo_admins.test", "admins.0.preferences.default_users_expiration"),
					resource.TestCheckNoResourceAttr("data.sftpgo_admins.test", "admins.0.preferences.hide_user_page_sections"),
					resource.TestCheckNoResourceAttr("data.sftpgo_admins.test", "admins.0.role"),
					resource.TestCheckNoResourceAttr("data.sftpgo_admins.test", "admins.0.groups"),
					// Check the admin created in the test case
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.username", admin.Username),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.id", admin.Username),
					resource.TestCheckResourceAttrSet("data.sftpgo_admins.test", "admins.1.password"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.email", admin.Email),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.status", "0"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.permissions.#",
						fmt.Sprintf("%d", len(admin.Permissions))),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.permissions.0", admin.Permissions[0]),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.permissions.1", admin.Permissions[1]),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.permissions.2", admin.Permissions[2]),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.permissions.3", admin.Permissions[3]),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.permissions.4", admin.Permissions[4]),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.permissions.5", admin.Permissions[5]),
					resource.TestCheckResourceAttrSet("data.sftpgo_admins.test", "admins.1.created_at"),
					resource.TestCheckResourceAttrSet("data.sftpgo_admins.test", "admins.1.updated_at"),
					resource.TestCheckResourceAttrSet("data.sftpgo_admins.test", "admins.1.last_login"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.filters.%", "4"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.filters.allow_list.#", "2"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.filters.allow_list.0", admin.Filters.AllowList[0]),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.filters.allow_list.1", admin.Filters.AllowList[1]),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.filters.allow_api_key_auth", "true"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.filters.require_password_change", "true"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.filters.require_two_factor", "true"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.preferences.%", "2"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.preferences.default_users_expiration",
						fmt.Sprintf("%d", admin.Filters.Preferences.DefaultUsersExpiration)),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.preferences.hide_user_page_sections",
						fmt.Sprintf("%d", admin.Filters.Preferences.HideUserPageSections)),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.role", testRole.Name),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.groups.#", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.groups.0.name", testGroup.Name),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.groups.0.options.add_to_users_as", "2"),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "id", placeholderID),
				),
			},
		},
	})
}
