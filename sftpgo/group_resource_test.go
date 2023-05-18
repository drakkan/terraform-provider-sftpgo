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

func TestAccGroupResource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	c, err := getClient()
	require.NoError(t, err)
	_, err = c.CreateFolder(testFolder)
	require.NoError(t, err)

	defer func() {
		err = c.DeleteFolder(testFolder.Name)
		require.NoError(t, err)
	}()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: `
					resource "sftpgo_group" "test" {
  					  name = "test group"
  					  description = "dsc"
					  user_settings = {
						home_dir = "/tmp/home"
						max_sessions = 10
						"permissions" = {
							"/dir1" = "list,download"
							"/dir2" = "list,upload"
						}
						quota_size = 40960000
						quota_files = 100
						filters = {
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
							]
							two_factor_protocols = ["SSH","HTTP"]
						}
						filesystem = {
						  provider = 5
						  sftpconfig = {
							endpoint = "127.0.0.1:22"
							username = "root"
							password = "sftppwd"
							prefix = "/"
							equality_check_mode = 1
							fingerprints = ["SHA256:RFzBCUItH9LZS0cKB5UE6ceAYhBD5C8GeOBip8Z11+4"]
						  }
					    }
					  }
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_group.test", "name", "test group"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "id", "test group"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "description", "dsc"),
					resource.TestCheckResourceAttrSet("sftpgo_group.test", "created_at"),
					resource.TestCheckResourceAttrSet("sftpgo_group.test", "updated_at"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.home_dir", "/tmp/home"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.max_sessions", "10"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.permissions.%", "2"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.permissions./dir1", "list,download"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.permissions./dir2", "list,upload"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.quota_size", "40960000"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.quota_files", "100"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.upload_bandwidth"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.file_patterns.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.file_patterns.0.path", "/p1"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.file_patterns.0.allowed_patterns.0", "*.jpg"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.file_patterns.0.allowed_patterns.1", "*.pdf"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.file_patterns.0.deny_policy", "1"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filters.file_patterns.0.denied_patterns"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.file_patterns.1.path", "/p2"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.file_patterns.1.denied_patterns.0", "*.jpg"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.file_patterns.1.denied_patterns.1", "*.pdf"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filters.file_patterns.1.deny_policy"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filters.file_patterns.1.allowed_patterns"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.two_factor_protocols.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.two_factor_protocols.0", "SSH"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.two_factor_protocols.1", "HTTP"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filters.max_upload_file_size"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filters.check_password_disabled"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filesystem.provider", "5"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filesystem.sftpconfig.endpoint", "127.0.0.1:22"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filesystem.sftpconfig.username", "root"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filesystem.sftpconfig.password", "sftppwd"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filesystem.sftpconfig.prefix", "/"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filesystem.sftpconfig.equality_check_mode", "1"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filesystem.sftpconfig.private_key"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filesystem.sftpconfig.fingerprints.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filesystem.sftpconfig.fingerprints.0",
						"SHA256:RFzBCUItH9LZS0cKB5UE6ceAYhBD5C8GeOBip8Z11+4"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filesystem.sftpconfig.disable_concurrent_reads"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filesystem.s3config"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filesystem.gcsconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "virtual_folders"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sftpgo_group.test",
				ImportState:       true,
				ImportStateVerify: false, // SFTPGo will not return plain text secrets
			},
			// Update and Read testing
			{
				Config: `
				resource "sftpgo_group" "test" {
				  name = "test group"
				  user_settings = {
					home_dir = "/tmp/home/local"
					max_sessions = 5
					upload_bandwidth = 128
					filters = {
						two_factor_protocols = ["HTTP"]
						max_upload_file_size = 1024
						check_password_disabled = true
						bandwidth_limits = [
							{
								sources = ["127.0.0.1/32"]
								upload_bandwidth = 256
								download_bandwidth = 64
							}
						]
					}
					filesystem = {
					  provider = 0
					}
				  }
				  virtual_folders = [
					{
						name = "tfolder"
						virtual_path = "/f1"
						quota_size = 0
						quota_files = 0
					}
				  ]
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_group.test", "name", "test group"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "id", "test group"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "description"),
					resource.TestCheckResourceAttrSet("sftpgo_group.test", "created_at"),
					resource.TestCheckResourceAttrSet("sftpgo_group.test", "updated_at"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.home_dir", "/tmp/home/local"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.max_sessions", "5"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.permissions"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.quota_size"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.quota_files"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.upload_bandwidth", "128"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filters.file_patterns"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.two_factor_protocols.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.two_factor_protocols.0", "HTTP"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.max_upload_file_size", "1024"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.check_password_disabled", "true"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.bandwidth_limits.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.bandwidth_limits.0.sources.0", "127.0.0.1/32"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.bandwidth_limits.0.upload_bandwidth", "256"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.bandwidth_limits.0.download_bandwidth", "64"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filesystem.provider", "0"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filesystem.osconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filesystem.s3config"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filesystem.gcsconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filesystem.cryptconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filesystem.sftpconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filesystem.httpconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filesystem.azblobconfig"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "virtual_folders.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "virtual_folders.0.name", testFolder.Name),
					resource.TestCheckResourceAttr("sftpgo_group.test", "virtual_folders.0.virtual_path", "/f1"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "virtual_folders.0.quota_size", "0"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "virtual_folders.0.quota_files", "0"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "virtual_folders.0.filesystem.provider", "1"),
				),
			},
			// Update and Read testing
			{
				Config: `
				resource "sftpgo_group" "test" {
				  name = "test group"
				  user_settings = {
					home_dir = "/tmp/home/local"
					filters = {
						two_factor_protocols = ["SSH"]
					}
					filesystem = {
					  provider = 4
					  cryptconfig = {
						passphrase = "pwd"
					    write_buffer_size = 5
					  }
					}
				  }
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_group.test", "name", "test group"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "id", "test group"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "description"),
					resource.TestCheckResourceAttrSet("sftpgo_group.test", "created_at"),
					resource.TestCheckResourceAttrSet("sftpgo_group.test", "updated_at"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.home_dir", "/tmp/home/local"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.max_sessions"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.permissions"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.quota_size"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.quota_files"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.upload_bandwidth"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filters.file_patterns"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.two_factor_protocols.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.two_factor_protocols.0", "SSH"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filters.max_upload_file_size"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filters.check_password_disabled"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filters.bandwidth_limits.#", "0"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filesystem.provider", "4"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filesystem.cryptconfig.passphrase", "pwd"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "user_settings.filesystem.cryptconfig.write_buffer_size", "5"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filesystem.cryptconfig.read_buffer_size"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filesystem.osconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filesystem.s3config"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filesystem.gcsconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filesystem.sftpconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filesystem.httpconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_group.test", "user_settings.filesystem.azblobconfig"),
					resource.TestCheckResourceAttr("sftpgo_group.test", "virtual_folders.#", "0"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
