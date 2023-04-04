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
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAdminResource(t *testing.T) {
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
    preferences = {
      hide_user_page_sections = 5
    }
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_admin.test", "username", "test"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "password", "pwd"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "status", "1"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "email", "admin@sftpgo.com"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "permissions.#", "3"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "permissions.0", "add_users"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "permissions.1", "edit_users"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "permissions.2", "del_users"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "filters.%", "3"),
					resource.TestCheckNoResourceAttr("sftpgo_admin.test", "filters.allow_api_key_auth"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "filters.allow_list.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "filters.allow_list.0", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "filters.preferences.%", "2"),
					resource.TestCheckNoResourceAttr("sftpgo_admin.test", "filters.preferences.default_users_expiration"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "filters.preferences.hide_user_page_sections", "5"),
					resource.TestCheckResourceAttrSet("sftpgo_admin.test", "created_at"),
					resource.TestCheckResourceAttrSet("sftpgo_admin.test", "updated_at"),
					resource.TestCheckResourceAttrSet("sftpgo_admin.test", "last_login"),
				),
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
					  		preferences = {
								default_users_expiration = 15
					  		}
						}
				  	}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_admin.test", "username", "test"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "password", "pwd1"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "status", "0"),
					resource.TestCheckNoResourceAttr("sftpgo_admin.test", "email"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "permissions.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "permissions.0", "*"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "filters.%", "3"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "filters.allow_api_key_auth", "true"),
					resource.TestCheckNoResourceAttr("sftpgo_admin.test", "filters.allow_list"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "filters.preferences.%", "2"),
					resource.TestCheckResourceAttr("sftpgo_admin.test", "filters.preferences.default_users_expiration", "15"),
					resource.TestCheckNoResourceAttr("sftpgo_admin.test", "filters.preferences.hide_user_page_sections"),
					resource.TestCheckResourceAttrSet("sftpgo_admin.test", "created_at"),
					resource.TestCheckResourceAttrSet("sftpgo_admin.test", "updated_at"),
					resource.TestCheckResourceAttrSet("sftpgo_admin.test", "last_login"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
