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
	"path/filepath"
	"testing"

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sftpgo/sdk"
	"github.com/stretchr/testify/require"
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
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "8"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.user_inactivity_config"),
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
					  description = "test rotate log file"
					  type = 15
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_action.test", "name", "test action"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "id", "test action"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "description", "test rotate log file"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "type", "15"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "8"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.user_inactivity_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.idp_config"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sftpgo_action.test",
				ImportState:       true,
				ImportStateVerify: true, // import verify will fail if we set any secret because it will be encrypted
			},
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
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "8"),
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
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.user_inactivity_config"),
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
								endpoint = "http://127.0.0.1:8082/deletenotify"
								username = "myuser"
								password = "mypassword"
								timeout = 10
								method = "DELETE"
							}
						}
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_action.test", "name", "test action"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "id", "test action"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "type", "1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "8"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.endpoint",
						"http://127.0.0.1:8082/deletenotify"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.username", "myuser"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.password", "mypassword"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.timeout", "10"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.http_config.method", "DELETE"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.user_inactivity_config"),
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
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "8"),
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
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.user_inactivity_config"),
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
								bcc = ["example3@example.com", "example4@example.com"]
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
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "8"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.recipients.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.recipients.0", "example1@example.com"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.recipients.1", "example2@example.com"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.bcc.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.bcc.0", "example3@example.com"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.bcc.1", "example4@example.com"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.subject", "test subject"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config.content_type"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.body", "test body"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.attachments.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.attachments.0", "/path1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.attachments.1", "/path2"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.user_inactivity_config"),
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
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "8"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.recipients.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.recipients.0", "example3@example.com"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.recipients.1", "example4@example.com"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.bcc.#", "0"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.subject", "test subject1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.content_type", "1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.body", "<p>test body1</p>"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.attachments.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.attachments.0", "/path3"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.email_config.attachments.1", "/path4"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.user_inactivity_config"),
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
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "8"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.retention_config.folders.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.retention_config.folders.0.path", "/dir1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.retention_config.folders.0.retention", "10"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.retention_config.folders.0.delete_empty_dirs", "true"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.retention_config.folders.1.path", "/dir2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.retention_config.folders.1.retention", "15"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config.folders.1.delete_empty_dirs"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.user_inactivity_config"),
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
										update_modtime = true
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
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "8"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.type", "1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.renames.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.renames.0.key", "/source1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.renames.0.value", "/target1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.renames.0.update_modtime", "true"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.renames.1.key", "/source2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.renames.1.value", "/target2"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.renames.1.update_modtime"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.mkdirs"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.deletes"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.exist"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.copy"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.compress"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.user_inactivity_config"),
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
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "8"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
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
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.user_inactivity_config"),
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
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "8"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
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
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.user_inactivity_config"),
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
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "8"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
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
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.user_inactivity_config"),
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
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "8"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
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
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.user_inactivity_config"),
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
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "8"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
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
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.user_inactivity_config"),
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
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "8"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.pwd_expiration_config.threshold", "10"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.user_inactivity_config"),
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
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "8"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.user_inactivity_config"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.idp_config.%", "3"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.idp_config.mode", "0"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.idp_config.template_user",
						"{\"username:\":\"user\"}"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.idp_config.template_admin",
						"{\"username:\":\"admin\"}"),
				),
			},
			{
				Config: `
					resource "sftpgo_action" "test" {
						name = "test action"
						type = 14
						options = {
							user_inactivity_config = {
								disable_threshold = 10
								delete_threshold = 20
							}
						}
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_action.test", "name", "test action"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "id", "test action"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "type", "14"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "8"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.user_inactivity_config.disable_threshold", "10"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.user_inactivity_config.delete_threshold", "20"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.idp_config"),
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

func TestAccEnterpriseActionResource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	c, err := getClient()
	require.NoError(t, err)
	if !c.IsEnterpriseEdition() {
		t.Skip("This test is supported only with the Enterprise edition")
	}

	folder, err := c.CreateFolder(testFolder)
	require.NoError(t, err)

	f1 := client.BaseVirtualFolder{
		BaseVirtualFolder: sdk.BaseVirtualFolder{
			Name:       "target_folder",
			MappedPath: filepath.Join(os.TempDir(), "target_folder"),
		},
	}

	targetFolder, err := c.CreateFolder(f1)
	require.NoError(t, err)

	defer func() {
		err = c.DeleteFolder(folder.Name)
		require.NoError(t, err)
		err = c.DeleteFolder(targetFolder.Name)
		require.NoError(t, err)
	}()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: `
					resource "sftpgo_action" "test" {
						name = "test action"
						type = 9
						options = {
							fs_config = {
							  type = 7
							  folder = "tfolder"
							  target_folder = "target_folder"
                              pgp = {
							    mode = 2
           						paths = [
             					  {
               						key = "/{{.VirtualPath}}"
                                    value = "/{{.VirtualPath}}.pgp"
                                  }
                                ],
                                passphrase = "password",
                                private_key = <<EOF
-----BEGIN PGP PRIVATE KEY BLOCK-----

lQWGBGf1VooBDAC3zd3GOKs9dIEn2dCFVEHBPtbd1fEAb3PiGENySjnoVVyP9E50
kzEGZJjiebuFLzxdm+1oK82OwNex9cw7uTQaniKwET04J9MpgodhipmKjLyFnmjL
ibea8fg1xA1NhwCkuwLLYat8q0ISLlu4TSfxgR5Exnyn4S8mGHeCXupQ5JCbQp0P
N3anSu77soI56KHGLf4UyZ5robSXlvQqNtpHesGexKVpY2TwVheICs3PYRpgwpR8
+JrEyDu7ZECkCrlOwm0lblKOFZ6O2bKJa/0EvRDbFqd0WyLdJJrg8JuZovklZQ/1
5z2+qP9UIqiT4Bl+ZgMRIH15BH8W88TMdt9RpQgEmx7TZlm/oTWQcE31aNIpEgH/
8vSNsMBMiStEs1vV8bzsfCstzZLbomni0mXDd5/GqPXk25SGBTkA49PfOEg8dNJI
E8SgyKVYaVvb24xxDWarDDAigbuXXOCZFvsObu1/JdOVP5LXSS8GK/yM1KylaPBg
MU7R0ekbHhJG7MkAEQEAAf4HAwIGlR+9vLc0qP+3+aR6HLzTlhGTZGj/z8v1RWOE
6l1dWBlGmNOqkYPhOkq7GkxgYR7R4wcEhCen99qcgaDOEmPyWri2OwrGwrE0ZNp9
Ai1s4XoAB3cOdU2c36ecnPCq2ZWlAoN8R1w9z2M9rUjJXQubfWslBCIRJ7KbKTz+
GwaTj7nXov2Tb7OIv3NZtXs9edXcdhdqBRDu/l8RXQGsobJ2u2uRualq7NX6BUyx
ejRFkR+3Is6sBU0l7YOddS6/NpVPdjTyV9kkofsKkXtzeZ+Wpd5B+Hx0k2UQGPwr
SniHrGADKpYLhlnwDjlgUBnUroqt1oShaX+0mVJTT/7W98M2q9gq7gZijzxtm9jZ
Kg/Z+aE04pJXVU9fCSTeMDLedacGwG+21pojsPcHxBJFZSaQvw6ESGF+WMZj+rvJ
ajq9lqRW/olG/DNIdHUXf/beOL7cPjxuiAW6lXCheA6dj1G/YCY6LiObRFYcz6KT
hjWSuNnLlMvg3AozfkrZW4fBrDB9vUynz56ylLEGGN2wvplYjTurj6neT31BNmCl
8UH0AoZR14ONQSSOGrhBILwhyW7Ge46TDSKh2KOS/GA6K5QxkzrqkUZi0lH8jPoR
Ln8PeVHXpBTjTSuPC4nFkBTJi4JbU/7cqq/ZO5Vh9rktl5GWlp11F7yYQpwS0i4h
QpBnPzl1RmDV0UkbptXBq0afEft2BLlqpMMpkp5bIDzXzxeZ+UdX0/rygnuX8e+j
SfHvucAZPsKf+Tmqx/8lL/Bjc6DnXhxz54CUwr73e325pjgRqzZrpp/ec78+8FOd
LmSiABiFvVZyAiqz+3A6XvsCxdNu0tJT4vT/D3a1lzBwmArHToIpixOcyjoU9rjb
2GIa2Gb/M31ndX35BbdPsuiUDCXROXIklMggnqa248kbcsRnhTXnAjtC/Yy2/2bi
tZq6lzoHuGP0VQBx7H4AcDH8Hdy+1OxYH5wxyNhTAMalLFc/hm57CHm0Fg+G1jOm
IIGdEjBntgvOZ8dTj2nAH39FTjfOcK6/iMSX6H3yDvV33yOrtdOGRzgcy0XnUxr/
Vp+z9hEoIk2cHNXPA2HnwHMsA2fMTBkvAwMqxJzGh0MZ0IR6oAI826uFWqkTjUVJ
e+baj5HjyIM7jkMuTIyRueORyia7XmseigAGaWYvG0kF0g2eMw/UJY+2+Mu+nHLa
z/DwlxCAv+l734FTwd+KtJK8p1ENjHRlrF3U/LnoTgMww7+DNI3CEZZYqaXHSHLt
gqKp7rEJyNqrdj5NDcuG8jQNHqfQFnlS0qf5XVgSyPNTzLcANL7mwkMQbjaJCQKP
p7cKx3KQDDRop4F7c4SRzgoUju/mcbTQerQaVGVzdEtleSA8dGVzdEBleGFtcGxl
LmNvbT6JAdgEEwEIAEIWIQTviLHUA5rgAYE43yJlalAQeUjRigUCZ/VWigIbAwUJ
A8JnAAULCQgHAgMiAgEGFQoJCAsCBBYCAwECHgcCF4AACgkQZWpQEHlI0YphnAv+
JSLb412uqOELIUvMRCPWyFX+Y0tMJDfgp8ti0Lw8K7QCJgOwCdWj6hdvV5axzKzk
zxVnE1hY0WeDHhvttkYh6GyuelFQgAC3h3Hd90Qe1SQwN7fGwcGSIQE2Aos/fYdJ
DuW7YCzNyRxVdEr4j3tgkdDOgI1Gk1JDp7Yiz+93qSnskR9NNu7tIMAO/G7XieS/
pC7E4QsePbYpsMcwUCbB9XGVO9v985qjwD/JL6wN9QoH7VZyZWn/u9bY8rASUhmS
nqdu2QyXpo9vBOb+EnWBGTFA09s7E4EDgt8ccOU50Q1AZKO0953DUZ/WAk2zCjnY
lP8U55kwteHSNriAYYsZEGv4fLmGWy/Gt1n0mssgDi9wB4bgk5OxljCepWmkZYmX
oCRNUeksbvxZNCETQnVxBMjk/LPEtL+pnU9ntHsQ5lICuH/EB3aVfvDMq6wcX5hh
HRWazxtVz5XlAxKYNCwZYuyLaKl4e5MCIBXDty0gaIRXl7ty/YFLDAewmJ7IlDfR
nQWGBGf1VooBDACgCsdtvphpqbeVStn9yCkV+mw3tuj64qYHHsIUfJEB4iebi+gp
giMXJrFTDeoDtAL/6Z3Kt1TiBWPufZFdEbjn8aBE2FHXhNJhQjWDqng+KvvPaiVZ
VW4wh4nfvFcc6vON9PZdVyiTSVtHNbmiWyWHvd7rw0rn7/YJZD40owG+3Z+/kRaA
WBaTodSjTp7Xj8mOb9Hy11CiMDAuBxt5gCIch2Te5ee2ooVyfbF0QeL1PaEJ6Up3
09Oxaey/Ge3MKLtN79vSWvxJ5+b9n7xuwwpCMLx2NEG2VvXDplEaDHAuRjKQfSbT
GvVeMnfFHaGvugVXhY6gpbp99X0q6IlBKz86UkSAexvCE/Dafl6cDr0y39BC0Qaj
dKW9igFTfp/gS9Zu7gBqlGVk0pumzAwwPi9iD654RCrg3m3Vkh4z5eo9VgWI2QZu
eZcibeNkcx08IxJG31PHi0umLggRPNMmuAn6uwGYetrropQxV1TfpKqxt621hKrN
eTuyx/Xxz0kc+PsAEQEAAf4HAwKxU97EWyJgOP/AK9bmmml8hxGVGe3iuJxM2IE3
kiUbHKfgTby6YJW81r/fU+hCKxRMhscjJzuQfNyYE+0QfaP9WSlbzV+DpSHhZZRo
vRFQDwrBrS3XIHi2fRj3huZYqsmpmZB9IEDuHhqXUDepZ0Vw9DZITxg3gadHGeKY
vanMXygR6x2REsT8TryQNqk97zPWufnlItObutzE2VRC7lQmnL/pCdCwbJSSEEQt
PSB54RZPfAlQQj/EUFuwy2LmYoaNGy12eetkzMkEQ6CnPsH42pDnUmrDFxvQn53C
01ZSSrTVTjV3XWBaq9Is36BwW0EEMy9KSpT9hMzO5zfJ3riRMTII7bjoD92U2A2G
Z5Bf3WvFrPuUsTmbYi5Zi+AHH1YTP7Le0SO+nx9lTDb4k09FdHX8b81AyyL4tHRU
cKHkNCw0RjUzvXCX/G0EtEHyKDSZl1rzcUSxYLqBWpaBwQChKmbsPJxDZ2SwyyIT
dWDSMxWFRnej3xsqe61fd8qopiZKzlpdIrXTehrGVYqUjktZ6UunaEI7cWhOhZnk
HuMwQZsA2TvDUp5+MsVMdzmK8P/QwgV2PjiliCU01b/0q54Gf1fV0h+henH1ZDwo
Sb1xvfrg7OihaotyRWTUAQZ0oalVEPvfZpAgWuH15HgRAMYpWniKSZKZhO84SXEn
yzinu4e0KXBFiGi2zS7ZR9mvdUAmGS9RIWKWzFOXDjAeR/vbQGKHl2/+yzy90brC
icPMPAaUDu3Ndo3jUu9S4ngcB89R24YEc0ugb45NKtRqek1e+NpYfOYgkioVMBH7
sKB4KTA0M40hpAKV0u1bgWJdFJwU6DTY5alK3inESyJnhWRT99Becsj9uZrTFGvV
lQc6rMDxoCEFDuU0wiSD3lMe8pmXCaoe+hWbC/vK0BJbZrzvk2QmB+13VZYECMeT
mrjXuYXap82pm1cJqOw+yrCPspa6dm3rSyrCqR/iDjx7UIWbb53gRDeG1MLXksjP
sSVAvPVMhkgED6p7ZSDZ3BoHe+V8ZL2lpLtzAJ/IJuIIFCb2V0oOM+bd9J7PlfUH
KUZLlC4Ig5lsOeoLdwyLRndkm0rtGf3OU9Tk0BPmeUjrjw+eAR0Hk4RV43J9Damd
7/bJZp2yo54ReaC3h51JO3zF0xGeqOZJq92dGAeBhRxF82T5S6XP2wVSJRZ/DGH4
CUN8C/X/6tNcQAC8PN8oFXF8PoAQC5upiAWDhF8K8O1ceMMA9yiLawSu9ai952rY
IMcIh1LOrVMF2oY7Cl7V/s6h1etIWoJbDY9ClZ1hkYyeklTPMPCtZXcMmoc7itcI
Tc2Nu0ii8aluxnqlhU4cQ2lHeDJTa/0ac4kBvAQYAQgAJhYhBO+IsdQDmuABgTjf
ImVqUBB5SNGKBQJn9VaKAhsMBQkDwmcAAAoJEGVqUBB5SNGKmQcL/id8ymIVe77o
wtTHPWTqMrltfzaWRbE7SGu/+KhK94IL8hAFybSXRGfJtvUaf7GOohLSMuzEBfd+
Vi0HZlu7GG1xU4z92EsLeKlV9D3ASKwqa9ayjakEBl8vEcjl9J0W/fjCmwk4U+GZ
tIIS4WvPyh7sY6lXC34pnMStJxbZsMtrvZEwbCanBA0F5jfbPfZ40fyq/x5D9gbm
FHRA31QOh0ItoiAEkFFqCuh0YMGuHRftm2wVwq8hpdoFyxkg45onPQIHAeVu77jA
m+6+tp/XugfP/vzaJF1EoF+dFa5mql1MHeKvSqM3Ueln1DZwHmYX04YapyHZNEsy
21YT/KgHXpu339t7bXGKzsX4M2nWp/3fFj+n+MPw0rbQklp3oTTN7/5K6iv30xiT
9ZRMmVCokmu6oDH/9f0sr6cQs2WA+siLF5z7vskLEGATE07SWyj+g2MN7bj2KwAe
w6apA2HhaekP72VjBiusut8VwUcr9PrnX5b2uEtRtKYaRD/uiGirPQ==
=2nRD
-----END PGP PRIVATE KEY BLOCK-----
EOF,
                                public_key = <<EOF
-----BEGIN PGP PUBLIC KEY BLOCK-----

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
GRw45d4+RU0GqiutTB/J5RfVUzAbXAvG
=OvEA
-----END PGP PUBLIC KEY BLOCK-----
EOF
							  }
							}
						}
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_action.test", "name", "test action"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "id", "test action"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "type", "9"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.%", "8"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.http_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.cmd_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.email_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.retention_config"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.type", "7"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.pgp.mode", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.pgp.paths.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.pgp.paths.0.key", "/{{.VirtualPath}}"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.pgp.paths.0.value", "/{{.VirtualPath}}.pgp"),
					resource.TestCheckResourceAttrSet("sftpgo_action.test", "options.fs_config.pgp.private_key"),
					resource.TestCheckResourceAttrSet("sftpgo_action.test", "options.fs_config.pgp.passphrase"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.pgp.password"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.pgp.profile"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.folder", folder.Name),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.target_folder", targetFolder.Name),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.deletes"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.exist"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.copy"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.compress"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.user_inactivity_config"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.idp_config"),
				),
			},
			{
				ResourceName:      "sftpgo_action.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{"options.fs_config.pgp.private_key",
					"options.fs_config.pgp.passphrase", "options.fs_config.pgp.password"},
			},
			{
				Config: `
					resource "sftpgo_action" "test" {
						name = "test action"
						type = 9
						options = {
							fs_config = {
								type = 7,
								pgp = {
									mode = 1
									profile = 2
									password = "secret"
           							paths = [
             					      {
               						    key = "/{{.VirtualPath}}.pgp"
                                        value = "/{{.VirtualPath}}"
                                      }
                                   ]
								}
							}
						}
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_action.test", "name", "test action"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "id", "test action"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.type", "7"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.pgp.mode", "1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.pgp.profile", "2"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.pgp.paths.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.pgp.paths.0.key", "/{{.VirtualPath}}.pgp"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.fs_config.pgp.paths.0.value", "/{{.VirtualPath}}"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.pgp.private_key"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.pgp.passphrase"),
					resource.TestCheckResourceAttrSet("sftpgo_action.test", "options.fs_config.pgp.password"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.folder"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "options.fs_config.target_folder"),
				),
			},
			{
				ResourceName:      "sftpgo_action.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{"options.fs_config.pgp.private_key",
					"options.fs_config.pgp.passphrase", "options.fs_config.pgp.password"},
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
										path = "/"
										retention = 24
									}
								],
								archive_folder = "target_folder"
								archive_path = "/base"
							}
						}
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_action.test", "name", "test action"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "id", "test action"),
					resource.TestCheckNoResourceAttr("sftpgo_action.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "type", "8"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.retention_config.folders.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.retention_config.folders.0.path", "/"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.retention_config.folders.0.retention", "24"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.retention_config.archive_folder", "target_folder"),
					resource.TestCheckResourceAttr("sftpgo_action.test", "options.retention_config.archive_path", "/base"),
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
