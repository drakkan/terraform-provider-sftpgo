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

func TestAccActionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: `
					resource "sftpgo_action" "test" {
  					  name = "test action"
					  description = "test desc"
					  type = 4
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_action.test", "name", "test action"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "id", "test action"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "description", "test desc"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "type", "4"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "7"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.idp_config"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sftpgo_action.test",
				ImportState:       true,
				ImportStateVerify: true, // import verify will fail if we set any secret because it will be encrypted
			},
			// Update and Read testing
			{
				Config: `
					resource "sftpgo_action" "test" {
						name = "test action"
						type = 1
						options = {
							http_config = {
								endpoint = "http://127.0.0.1:8082/notify"
								username = "myuser"
								password = "mypassword"
								timeout = 10
								method = "GET"
							}
						}
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_action.test", "name", "test action"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "id", "test action"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "type", "1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "7"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.endpoint",
						"http://127.0.0.1:8082/notify"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.username", "myuser"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.password", "mypassword"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.timeout", "10"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.method", "GET"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.idp_config"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
