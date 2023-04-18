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

func TestAccUserResource(t *testing.T) {
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
					resource "sftpgo_user" "test" {
  					  username = "test user"
  					  status      = 1
    				  password    = "secret pwd"
                      home_dir    = "/tmp/testuser"
    				  email       = "test@test.com"
    				  permissions = {
        				"/" = "*",
        				"/p1" = "list,download"
    				  }
    				  filesystem = {
      					provider = 1
      					s3config = {
        				  bucket = "bucket"
        				  region = "us-west-1"
        				  access_key = "key"
        				  access_secret = "secret payload"
      				    }
    				  }
    				  groups = [
      					{
        				  name = "test group"
        				  type = 3
                        }
                      ]
    				  virtual_folders = [
      					{
        				  name = "tfolder"
        				  virtual_path = "/vdir"
        				  quota_size = -1
        				  quota_files = -1
      					}
    				  ]
    				  filters = {
      					allowed_ip = ["192.168.1.0/24", "10.0.0.0/8"]
      					start_directory = "/start/dir"
      					file_patterns = [
        				  {
          					path = "/p1"
          					allowed_patterns = ["*.jpg","*.pdf"]
          					deny_policy = 1
        				  },
        				  {
          					path = "/p2"
          					denied_patterns = ["*.jpg","*.pdf"]
        				  },
        				  {
          					path = "/p3"
          					denied_patterns = ["*.abc"]
        				  }
      					]
      				    external_auth_disabled = true
      					bandwidth_limits = [
       	 				  {
          					sources = ["127.0.0.1/32","192.168.1.0/24"]
          					upload_bandwidth = 256
          					download_bandwidth = 128
        				  }
      					]
    				  }
					  role = "test role"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_user.test", "username", "test user"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "id", "test user"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "status", "1"),
					resource.TestCheckResourceAttrSet("sftpgo_user.test", "created_at"),
					resource.TestCheckResourceAttrSet("sftpgo_user.test", "updated_at"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "password", "secret pwd"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "home_dir", "/tmp/testuser"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "email", "test@test.com"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "permissions.%", "2"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "permissions./", "*"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "permissions./p1", "list,download"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filesystem.provider", "1"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filesystem.s3config.bucket", "bucket"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filesystem.s3config.region", "us-west-1"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filesystem.s3config.access_key", "key"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filesystem.s3config.access_secret", "secret payload"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "filesystem.gcsconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "description"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "additional_info"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "groups.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "groups.0.name", testGroup.Name),
					resource.TestCheckResourceAttr("sftpgo_user.test", "groups.0.type", "3"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "virtual_folders.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "virtual_folders.0.name", testFolder.Name),
					resource.TestCheckResourceAttr("sftpgo_user.test", "virtual_folders.0.virtual_path", "/vdir"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "virtual_folders.0.quota_size", "-1"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "virtual_folders.0.quota_files", "-1"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.allowed_ip.0", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.allowed_ip.1", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.start_directory", "/start/dir"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.file_patterns.#", "3"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.file_patterns.0.path", "/p1"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.file_patterns.0.allowed_patterns.0", "*.jpg"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.file_patterns.0.allowed_patterns.1", "*.pdf"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.file_patterns.0.deny_policy", "1"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "filters.file_patterns.0.denied_patterns"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.file_patterns.1.path", "/p2"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.file_patterns.1.denied_patterns.0", "*.jpg"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.file_patterns.1.denied_patterns.1", "*.pdf"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "filters.file_patterns.1.deny_policy"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "filters.file_patterns.1.allowed_patterns"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.file_patterns.2.path", "/p3"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.file_patterns.2.denied_patterns.0", "*.abc"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "filters.file_patterns.2.deny_policy"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "filters.file_patterns.2.allowed_patterns"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.external_auth_disabled", "true"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "filters.pre_login_disabled"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.bandwidth_limits.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.bandwidth_limits.0.sources.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.bandwidth_limits.0.sources.0", "127.0.0.1/32"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.bandwidth_limits.0.sources.1", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.bandwidth_limits.0.upload_bandwidth", "256"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.bandwidth_limits.0.download_bandwidth", "128"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "role", testRole.Name),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sftpgo_user.test",
				ImportState:       true,
				ImportStateVerify: false, // SFTPGo will not return plain text password/secrets
			},
			// Update and Read testing
			{
				Config: `
				resource "sftpgo_user" "test" {
				  username = "test user"
				  status      = 0
				  home_dir    = "/tmp/testuser"
				  additional_info = "info"
				  permissions = {
					"/" = "*",
					"/p2" = "list,download"
				  }
				  filesystem = {
					  provider = 0
				  }
				  groups = [
					  {
					  name = "test group"
					  type = 1
					}
				  ]
				  filters = {
					  denied_ip = ["192.168.1.0/24", "10.0.0.0/8"]
					  pre_login_disabled = true
					  denied_login_methods = ["publickey", "password-over-SSH"]
					  tls_username = "CommonName"
					  web_client = ["write-disabled"]
					  user_type = "LDAPUser"
					  data_transfer_limits = [
						{
							sources = ["2001:db8:abcd:0012::0/96"]
							upload_data_transfer = 100
							download_data_transfer = 200
						}
					  ]
					  ftp_security = 1
				  }
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_user.test", "username", "test user"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "id", "test user"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "status", "0"),
					resource.TestCheckResourceAttrSet("sftpgo_user.test", "created_at"),
					resource.TestCheckResourceAttrSet("sftpgo_user.test", "updated_at"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "password"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "home_dir", "/tmp/testuser"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "email"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "permissions.%", "2"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "permissions./", "*"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "permissions./p2", "list,download"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filesystem.provider", "0"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "filesystem.s3config"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "filesystem.gcsconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "additional_info", "info"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "groups.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "groups.0.name", testGroup.Name),
					resource.TestCheckResourceAttr("sftpgo_user.test", "groups.0.type", "1"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "virtual_folders"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "role"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.denied_ip.0", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.denied_ip.1", "10.0.0.0/8"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "filters.start_directory"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "filters.file_patterns"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "filters.external_auth_disabled"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.pre_login_disabled", "true"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "filters.bandwidth_limits"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.denied_login_methods.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.denied_login_methods.0", "publickey"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.denied_login_methods.1", "password-over-SSH"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.tls_username", "CommonName"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.web_client.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.web_client.0", "write-disabled"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.user_type", "LDAPUser"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.data_transfer_limits.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.data_transfer_limits.0.sources.0", "2001:db8:abcd:0012::0/96"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.data_transfer_limits.0.upload_data_transfer", "100"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.data_transfer_limits.0.download_data_transfer", "200"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "filters.data_transfer_limits.0.total_data_transfer"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "filters.is_anonymous"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.ftp_security", "1"),
				),
			},
			// Update and Read anonymous user testing
			{
				Config: `
				resource "sftpgo_user" "test" {
				  username = "test user"
				  status      = 1
				  home_dir    = "/tmp/testuser"
				  permissions = {
					"/" = "list,download"
				  }
				  filesystem = {
					  provider = 0
				  }
				  filters = {
					denied_protocols = ["SSH", "HTTP"]
					denied_login_methods = ["publickey", "password-over-SSH", "keyboard-interactive", "publickey+password", "publickey+keyboard-interactive", "TLSCertificate", "TLSCertificate+password"]
					is_anonymous = true
				  }
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_user.test", "username", "test user"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "id", "test user"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "status", "1"),
					resource.TestCheckResourceAttrSet("sftpgo_user.test", "created_at"),
					resource.TestCheckResourceAttrSet("sftpgo_user.test", "updated_at"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "password"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "home_dir", "/tmp/testuser"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "email"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "permissions.%", "1"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "permissions./", "list,download"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filesystem.provider", "0"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "filesystem.s3config"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "filesystem.gcsconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "description"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "additional_info"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "groups"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "virtual_folders"),
					resource.TestCheckNoResourceAttr("sftpgo_user.test", "role"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.is_anonymous", "true"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.denied_protocols.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.denied_protocols.0", "SSH"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.denied_protocols.1", "HTTP"),
					resource.TestCheckResourceAttr("sftpgo_user.test", "filters.denied_login_methods.#", "7"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
