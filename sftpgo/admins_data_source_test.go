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
		Permissions:    []string{"add_users", "edit_users", "view_status", "quota_scans"},
		Email:          "admin@sftpgo.com",
		Description:    "Created from Terraform",
		AdditionalInfo: "TF",
		Filters: client.AdminFilters{
			AllowList:       []string{"127.0.0.1/8", "192.168.1.0/24"},
			AllowAPIKeyAuth: true,
			Preferences: client.AdminPreferences{
				HideUserPageSections:   96,
				DefaultUsersExpiration: 15,
			},
		},
	}
	_, err = c.CreateAdmin(admin)
	require.NoError(t, err)

	defer func() {
		err = c.DeleteAdmin(admin.Username)
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
					resource.TestCheckNoResourceAttr("data.sftpgo_admins.test", "admins.0.password"),
					resource.TestCheckNoResourceAttr("data.sftpgo_admins.test", "admins.0.email"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.0.status", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.0.permissions.#", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.0.permissions.0", "*"),
					resource.TestCheckResourceAttrSet("data.sftpgo_admins.test", "admins.0.created_at"),
					resource.TestCheckResourceAttrSet("data.sftpgo_admins.test", "admins.0.updated_at"),
					resource.TestCheckResourceAttrSet("data.sftpgo_admins.test", "admins.0.last_login"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.0.filters.%", "3"),
					resource.TestCheckNoResourceAttr("data.sftpgo_admins.test", "admins.0.filters.allow_api_key_auth"),
					resource.TestCheckNoResourceAttr("data.sftpgo_admins.test", "admins.0.filters.allow_list"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.0.filters.preferences.%", "2"),
					resource.TestCheckNoResourceAttr("data.sftpgo_admins.test", "admins.0.filters.preferences.default_users_expiration"),
					resource.TestCheckNoResourceAttr("data.sftpgo_admins.test", "admins.0.filters.preferences.hide_user_page_sections"),
					resource.TestCheckNoResourceAttr("data.sftpgo_admins.test", "admins.0.role"),
					resource.TestCheckNoResourceAttr("data.sftpgo_admins.test", "admins.0.groups"),
					// Check the admin created in the test case
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.username", admin.Username),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.id", admin.Username),
					resource.TestCheckNoResourceAttr("data.sftpgo_admins.test", "admins.1.password"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.email", admin.Email),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.status", "0"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.permissions.#",
						fmt.Sprintf("%d", len(admin.Permissions))),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.permissions.0", admin.Permissions[0]),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.permissions.1", admin.Permissions[1]),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.permissions.2", admin.Permissions[2]),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.permissions.3", admin.Permissions[3]),
					resource.TestCheckResourceAttrSet("data.sftpgo_admins.test", "admins.1.created_at"),
					resource.TestCheckResourceAttrSet("data.sftpgo_admins.test", "admins.1.updated_at"),
					resource.TestCheckResourceAttrSet("data.sftpgo_admins.test", "admins.1.last_login"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.filters.%", "3"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.filters.allow_list.#", "2"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.filters.allow_list.0", admin.Filters.AllowList[0]),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.filters.allow_list.1", admin.Filters.AllowList[1]),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.filters.allow_api_key_auth", "true"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.filters.preferences.%", "2"),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.filters.preferences.default_users_expiration",
						fmt.Sprintf("%d", admin.Filters.Preferences.DefaultUsersExpiration)),
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "admins.1.filters.preferences.hide_user_page_sections",
						fmt.Sprintf("%d", admin.Filters.Preferences.HideUserPageSections)),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.sftpgo_admins.test", "id", placeholderID),
				),
			},
		},
	})
}
