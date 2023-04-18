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
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccAdminResource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	c, err := getClient()
	require.NoError(t, err)
	_, err = c.CreateFolder(testFolder)
	require.NoError(t, err)
	_, err = c.CreateGroup(testGroup)
	require.NoError(t, err)
	_, err = c.CreateRole(testRole)
	require.NoError(t, err)

	defer func() {
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
			// Create and Read testing
			{
				Config: `
					resource "sftpgo_admin" "test" {
					  username = "test"
					  status = 1
					  password = "pwd"
					  email = "admin@sftpgo.com"
					  permissions = ["add_users", "edit_users","del_users"]
 					  filters = {
    					allow_list = ["192.168.1.0/24"]
  					  }
					  preferences = {
						hide_user_page_sections = 5
					  }
					  role = "test role"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_admin.test", "username", "test"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "id", "test"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "password", "pwd"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "status", "1"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "email", "admin@sftpgo.com"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "permissions.#", "3"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "permissions.0", "add_users"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "permissions.1", "edit_users"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "permissions.2", "del_users"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "filters.%", "2"),
					resource.TestCheckNoResourceAttr("sftpgo_admin.test", "filters.allow_api_key_auth"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "filters.allow_list.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "filters.allow_list.0", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "preferences.%", "2"),
					resource.TestCheckNoResourceAttr("sftpgo_admin.test", "preferences.default_users_expiration"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "preferences.hide_user_page_sections", "5"),
					resource.TestCheckResourceAttrSet("sftpgo_admin.test", "created_at"),
					resource.TestCheckResourceAttrSet("sftpgo_admin.test", "updated_at"),
					resource.TestCheckResourceAttrSet("sftpgo_admin.test", "last_login"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "role", "test role"),
					resource.TestCheckNoResourceAttr("sftpgo_admin.test", "groups"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sftpgo_admin.test",
				ImportState:       true,
				ImportStateVerify: false, // SFTPGo will not return plain text password/secrets
			},
			// Update and Read testing
			{
				Config: `
					resource "sftpgo_admin" "test" {
						username = "test"
						status = 0
						password = "pwd1"
						permissions = ["*"]
						filters = {
						  allow_api_key_auth = true
						}
						preferences = {
						  default_users_expiration = 15
						}
						groups = [
							{
								name = "test group"
								options = {
									add_to_users_as = 1
								}
							}
						]
				  	}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_admin.test", "username", "test"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "password", "pwd1"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "status", "0"),
					resource.TestCheckNoResourceAttr("sftpgo_admin.test", "email"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "permissions.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "permissions.0", "*"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "filters.%", "2"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "filters.allow_api_key_auth", "true"),
					resource.TestCheckNoResourceAttr("sftpgo_admin.test", "filters.allow_list"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "preferences.%", "2"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "preferences.default_users_expiration", "15"),
					resource.TestCheckNoResourceAttr("sftpgo_admin.test", "preferences.hide_user_page_sections"),
					resource.TestCheckResourceAttrSet("sftpgo_admin.test", "created_at"),
					resource.TestCheckResourceAttrSet("sftpgo_admin.test", "updated_at"),
					resource.TestCheckResourceAttrSet("sftpgo_admin.test", "last_login"),
					resource.TestCheckNoResourceAttr("sftpgo_admin.test", "role"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "groups.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "groups.0.name", "test group"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "groups.0.options.add_to_users_as", "1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
