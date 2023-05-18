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
			{
				Config: `
					resource "sftpgo_action" "test" {
						name = "test action"
						type = 1
						options = {
							http_config = {
								endpoint = "https://127.0.0.1:8082/notify"
								username = "myuser"
								password = "mynewpassword"
								headers = [
									{
										key = "Content-Type",
										value = "application/json"
									}
								]
								timeout = 20
								skip_tls_verify = true
								method = "POST"
								query_parameters = [
									{
										key = "q1"
										value = "val1"
									},
									{
										key = "q2"
										value = "val2"
									}
								]
								body = "{\"hello\":\"world\"}"
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
						"https://127.0.0.1:8082/notify"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.username", "myuser"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.password", "mynewpassword"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.headers.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.headers.0.key", "Content-Type"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.headers.0.value", "application/json"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.timeout", "20"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.skip_tls_verify", "true"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.method", "POST"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.body", "{\"hello\":\"world\"}"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.query_parameters.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.query_parameters.0.key", "q1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.query_parameters.0.value", "val1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.query_parameters.1.key", "q2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.query_parameters.1.value", "val2"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.idp_config"),
				),
			},
			{
				ResourceName:      "sftpgo_action.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
			{
				Config: `
					resource "sftpgo_action" "test" {
						name = "test action"
						type = 3
						options = {
							email_config = {
								recipients = ["example1@example.com", "example2@example.com"]
								subject = "test subject"
								body = "test body"
								attachments = ["/path1","/path2"]
							}
						}
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_action.test", "name", "test action"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "id", "test action"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "type", "3"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "7"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.recipients.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.recipients.0", "example1@example.com"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.recipients.1", "example2@example.com"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.subject", "test subject"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config.content_type"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.body", "test body"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.attachments.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.attachments.0", "/path1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.attachments.1", "/path2"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.idp_config"),
				),
			},
			{
				ResourceName:      "sftpgo_action.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: `
					resource "sftpgo_action" "test" {
						name = "test action"
						type = 3
						options = {
							email_config = {
								recipients = ["example3@example.com", "example4@example.com"]
								subject = "test subject1"
								content_type = 1
								body = "<p>test body1</p>"
								attachments = ["/path3","/path4"]
							}
						}
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_action.test", "name", "test action"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "id", "test action"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "type", "3"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "7"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.recipients.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.recipients.0", "example3@example.com"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.recipients.1", "example4@example.com"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.subject", "test subject1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.content_type", "1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.body", "<p>test body1</p>"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.attachments.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.attachments.0", "/path3"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.attachments.1", "/path4"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.idp_config"),
				),
			},
			{
				ResourceName:      "sftpgo_action.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: `
					resource "sftpgo_action" "test" {
						name = "test action"
						type = 8
						options = {
							retention_config = {
								folders = [
									{
										path = "/dir1",
										retention = 10,
										delete_empty_dirs = true
									},
									{
										path = "/dir2",
										retention = 15,
										ignore_user_permissions = true
									}
								]
							}
						}
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_action.test", "name", "test action"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "id", "test action"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "type", "8"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "7"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.retention_config.folders.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.retention_config.folders.0.path", "/dir1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.retention_config.folders.0.retention", "10"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.retention_config.folders.0.delete_empty_dirs", "true"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config.folders.0.ignore_user_permissions"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.retention_config.folders.1.path", "/dir2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.retention_config.folders.1.retention", "15"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config.folders.1.delete_empty_dirs"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.retention_config.folders.1.ignore_user_permissions", "true"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.idp_config"),
				),
			},
			{
				ResourceName:      "sftpgo_action.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: `
					resource "sftpgo_action" "test" {
						name = "test action"
						type = 9
						options = {
							fs_config = {
								type = 1
								renames = [
									{
										key = "/source1"
										value = "/target1"
									},
									{
										key = "/source2"
										value = "/target2"
									}
								]
							}
						}
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_action.test", "name", "test action"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "id", "test action"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "type", "9"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "7"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.%", "7"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.type", "1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.renames.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.renames.0.key", "/source1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.renames.0.value", "/target1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.renames.1.key", "/source2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.renames.1.value", "/target2"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.mkdirs"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.deletes"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.exist"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.copy"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.compress"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.idp_config"),
				),
			},
			{
				ResourceName:      "sftpgo_action.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: `
					resource "sftpgo_action" "test" {
						name = "test action"
						type = 9
						options = {
							fs_config = {
								type = 2
								deletes = ["/path1", "/path2"]
							}
						}
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_action.test", "name", "test action"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "id", "test action"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "type", "9"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "7"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.%", "7"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.type", "2"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.renames"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.deletes.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.deletes.0", "/path1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.deletes.1", "/path2"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.mkdirs"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.exist"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.copy"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.compress"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.idp_config"),
				),
			},
			{
				ResourceName:      "sftpgo_action.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: `
					resource "sftpgo_action" "test" {
						name = "test action"
						type = 9
						options = {
							fs_config = {
								type = 3
								mkdirs = ["/path1", "/path2"]
							}
						}
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_action.test", "name", "test action"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "id", "test action"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "type", "9"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "7"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.%", "7"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.type", "3"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.renames"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.mkdirs.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.mkdirs.0", "/path1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.mkdirs.1", "/path2"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.deletes"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.exist"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.copy"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.compress"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.idp_config"),
				),
			},
			{
				ResourceName:      "sftpgo_action.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: `
					resource "sftpgo_action" "test" {
						name = "test action"
						type = 9
						options = {
							fs_config = {
								type = 4
								exist = ["/path1", "/path2"]
							}
						}
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_action.test", "name", "test action"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "id", "test action"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "type", "9"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "7"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.%", "7"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.type", "4"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.renames"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.exist.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.exist.0", "/path1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.exist.1", "/path2"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.deletes"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.mkdirs"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.copy"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.compress"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.idp_config"),
				),
			},
			{
				Config: `
					resource "sftpgo_action" "test" {
						name = "test action"
						type = 9
						options = {
							fs_config = {
								type = 5
								compress = {
									name = "/test.zip"
									paths = ["/path1", "/path2"]
								}
							}
						}
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_action.test", "name", "test action"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "id", "test action"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "type", "9"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "7"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.%", "7"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.type", "5"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.renames"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.exist"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.deletes"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.mkdirs"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.copy"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.compress.%", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.compress.name", "/test.zip"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.compress.paths.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.compress.paths.0", "/path1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.compress.paths.1", "/path2"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.idp_config"),
				),
			},
			{
				Config: `
					resource "sftpgo_action" "test" {
						name = "test action"
						type = 9
						options = {
							fs_config = {
								type = 6
								copy = [
									{
										key = "/source1"
										value = "/target1"
									},
									{
										key = "/source2"
										value = "/target2"
									}
								]
							}
						}
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_action.test", "name", "test action"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "id", "test action"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "type", "9"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "7"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.%", "7"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.type", "6"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.renames"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.exist"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.deletes"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.mkdirs"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.copy.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.copy.0.key", "/source1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.copy.0.value", "/target1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.copy.1.key", "/source2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.copy.1.value", "/target2"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.compress"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.idp_config"),
				),
			},
			{
				Config: `
					resource "sftpgo_action" "test" {
						name = "test action"
						type = 11
						options = {
							pwd_expiration_config = {
								threshold = 10
							}
						}
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_action.test", "name", "test action"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "id", "test action"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "type", "11"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "7"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.pwd_expiration_config.threshold", "10"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.idp_config"),
				),
			},
			{
				Config: `
					resource "sftpgo_action" "test" {
						name = "test action"
						type = 13
						options = {
							idp_config = {
								mode = 0
								template_user = "{\"username:\":\"user\"}"
								template_admin = "{\"username:\":\"admin\"}"
							}
						}
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_action.test", "name", "test action"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "id", "test action"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "type", "13"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "7"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.idp_config.%", "3"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.idp_config.mode", "0"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.idp_config.template_user",
						"{\"username:\":\"user\"}"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.idp_config.template_admin",
						"{\"username:\":\"admin\"}"),
				),
			},
			{
				ResourceName:      "sftpgo_action.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
