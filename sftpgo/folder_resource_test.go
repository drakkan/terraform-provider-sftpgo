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
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/require"
)

func TestAccFolderResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: `
					resource "sftpgo_folder" "test" {
  					  name = "test folder"
  					  mapped_path    = "/tmp/test_folder"
  					  filesystem = {
    					provider = 3
    					azblobconfig = {
     		 			  container = "fake container"
    					  account_name = "my access key"
    					  account_key = "my secret"
    					  key_prefix = "prefix/"
						  upload_part_size = 100
    					}
  					  }
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_folder.test", "name", "test folder"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "id", "test folder"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "mapped_path", "/tmp/test_folder"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "description"),
					resource.TestCheckResourceAttrSet("sftpgo_folder.test", "used_quota_size"),
					resource.TestCheckResourceAttrSet("sftpgo_folder.test", "used_quota_files"),
					resource.TestCheckResourceAttrSet("sftpgo_folder.test", "last_quota_update"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.provider", "3"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.s3config"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.gcsconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.cryptconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.sftpconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.httpconfig"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.azblobconfig.container", "fake container"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.azblobconfig.account_name", "my access key"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.azblobconfig.account_key", "my secret"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.azblobconfig.key_prefix", "prefix/"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.azblobconfig.upload_part_size", "100"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.azblobconfig.sas_url"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sftpgo_folder.test",
				ImportState:       true,
				ImportStateVerify: false, // SFTPGo will not return plain text secrets
			},
			// Update and Read testing
			{
				Config: `
					resource "sftpgo_folder" "test" {
					  name = "test folder"
					  mapped_path    = "/tmp/folder"
					  description = "desc"
					  filesystem = {
					    provider = 0
					  }
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_folder.test", "name", "test folder"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "id", "test folder"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "mapped_path", "/tmp/folder"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "description", "desc"),
					resource.TestCheckResourceAttrSet("sftpgo_folder.test", "used_quota_size"),
					resource.TestCheckResourceAttrSet("sftpgo_folder.test", "used_quota_files"),
					resource.TestCheckResourceAttrSet("sftpgo_folder.test", "last_quota_update"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.provider", "0"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.osconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.s3config"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.gcsconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.cryptconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.sftpconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.httpconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.azblobconfig"),
				),
			},
			// Update and Read testing
			{
				Config: `
					resource "sftpgo_folder" "test" {
					  name = "test folder"
					  mapped_path    = "/tmp/folder"
					  filesystem = {
					    provider = 0
						osconfig = {
						  write_buffer_size = 5
						}
					  }
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_folder.test", "name", "test folder"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "id", "test folder"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "mapped_path", "/tmp/folder"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "description"),
					resource.TestCheckResourceAttrSet("sftpgo_folder.test", "used_quota_size"),
					resource.TestCheckResourceAttrSet("sftpgo_folder.test", "used_quota_files"),
					resource.TestCheckResourceAttrSet("sftpgo_folder.test", "last_quota_update"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.provider", "0"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.osconfig.write_buffer_size", "5"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.osconfig.read_buffer_size"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.s3config"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.gcsconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.cryptconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.sftpconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.httpconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.azblobconfig"),
				),
			},
			// Update and Read testing
			{
				Config: `
					resource "sftpgo_folder" "test" {
					  name = "test folder"
					  mapped_path    = "/tmp/folder"
					  filesystem = {
					    provider = 4
						cryptconfig = {
							passphrase = "pwd"
							read_buffer_size = 4
						  }
					  }
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_folder.test", "name", "test folder"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "id", "test folder"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "mapped_path", "/tmp/folder"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "description"),
					resource.TestCheckResourceAttrSet("sftpgo_folder.test", "used_quota_size"),
					resource.TestCheckResourceAttrSet("sftpgo_folder.test", "used_quota_files"),
					resource.TestCheckResourceAttrSet("sftpgo_folder.test", "last_quota_update"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.provider", "4"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.osconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.s3config"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.gcsconfig"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.cryptconfig.passphrase", "pwd"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.cryptconfig.read_buffer_size", "4"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.cryptconfig.write_buffer_size"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.sftpconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.httpconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.azblobconfig"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccEnterpriseFolderResource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	c, err := getClient()
	require.NoError(t, err)
	if !c.IsEnterpriseEdition() {
		t.Skip("This test is supported only with the Enterprise edition")
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: `
					resource "sftpgo_folder" "test" {
  					  name = "test folder"
  					  mapped_path    = "/tmp/test_folder"
  					  filesystem = {
    					provider = 5
    					sftpconfig = {
        				  endpoint = "127.0.0.1:22"
						  username = "root"
						  password = "sftppwd"
						  prefix = "/"
						  socks_proxy = "socks4://127.0.0.1:10080"
      				    }
  					  }
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_folder.test", "name", "test folder"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "id", "test folder"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "mapped_path", "/tmp/test_folder"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "description"),
					resource.TestCheckResourceAttrSet("sftpgo_folder.test", "used_quota_size"),
					resource.TestCheckResourceAttrSet("sftpgo_folder.test", "used_quota_files"),
					resource.TestCheckResourceAttrSet("sftpgo_folder.test", "last_quota_update"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.provider", "5"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.s3config"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.gcsconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.cryptconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.azblobconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.httpconfig"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.sftpconfig.endpoint", "127.0.0.1:22"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.sftpconfig.username", "root"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.sftpconfig.password", "sftppwd"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.sftpconfig.prefix", "/"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.sftpconfig.socks_proxy", "socks4://127.0.0.1:10080"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.sftpconfig.socks_username"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.sftpconfig.socks_password"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sftpgo_folder.test",
				ImportState:       true,
				ImportStateVerify: false, // SFTPGo will not return plain text secrets
			},
			// Update and Read testing
			{
				Config: `
					resource "sftpgo_folder" "test" {
					  name = "test folder"
					  mapped_path    = "/tmp/folder"
					  description = "desc"
					  filesystem = {
					    provider = 2
					  	  gcsconfig = {
					        bucket = "hns"
					        automatic_credentials = 1
					        hns = 1
							universe_domain = "googleapis.com"
			            }
					  }
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_folder.test", "name", "test folder"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "id", "test folder"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "mapped_path", "/tmp/folder"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "description", "desc"),
					resource.TestCheckResourceAttrSet("sftpgo_folder.test", "used_quota_size"),
					resource.TestCheckResourceAttrSet("sftpgo_folder.test", "used_quota_files"),
					resource.TestCheckResourceAttrSet("sftpgo_folder.test", "last_quota_update"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.provider", "2"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.osconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.s3config"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.cryptconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.sftpconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.httpconfig"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.test", "filesystem.azblobconfig"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.gcsconfig.bucket", "hns"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.gcsconfig.automatic_credentials", "1"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.gcsconfig.hns", "1"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "filesystem.gcsconfig.universe_domain", "googleapis.com"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccFolderResource_writeOnly(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "sftpgo_folder" "wo" {
					  name        = "wo_folder"
					  mapped_path = "/tmp/wo_folder"
					  filesystem = {
					    provider = 4
					    cryptconfig = {
					      passphrase_wo         = "initial_passphrase"
					      passphrase_wo_version = "1"
					    }
					  }
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_folder.wo", "filesystem.cryptconfig.passphrase_wo_version", "1"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.wo", "filesystem.cryptconfig.passphrase_wo"),
				),
			},
			{
				Config: `
					resource "sftpgo_folder" "wo" {
					  name        = "wo_folder"
					  mapped_path = "/tmp/wo_folder"
					  filesystem = {
					    provider = 4
					    cryptconfig = {
					      passphrase_wo         = "rotated_passphrase"
					      passphrase_wo_version = "2"
					    }
					  }
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_folder.wo", "filesystem.cryptconfig.passphrase_wo_version", "2"),
					resource.TestCheckNoResourceAttr("sftpgo_folder.wo", "filesystem.cryptconfig.passphrase_wo"),
				),
			},
		},
	})
}

func TestAccFolderResource_renameForcesReplace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "sftpgo_folder" "test" {
					  name = "rename_test_initial"
					  mapped_path = "/tmp/rename_test_folder"
					  filesystem = { provider = 0 }
					}`,
			},
			{
				Config: `
					resource "sftpgo_folder" "test" {
					  name = "rename_test_renamed"
					  mapped_path = "/tmp/rename_test_folder"
					  filesystem = { provider = 0 }
					}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sftpgo_folder.test", plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_folder.test", "name", "rename_test_renamed"),
					resource.TestCheckResourceAttr("sftpgo_folder.test", "id", "rename_test_renamed"),
				),
			},
		},
	})
}
