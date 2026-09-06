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
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccRoleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: `
					resource "sftpgo_role" "test" {
  					  name = "test role"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_role.test", "name", "test role"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "id", "test role"),
					resource.TestCheckNoResourceAttr("sftpgo_role.test", "description"),
					resource.TestCheckResourceAttrSet("sftpgo_role.test", "created_at"),
					resource.TestCheckResourceAttrSet("sftpgo_role.test", "updated_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sftpgo_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// External deletion test
			{
				PreConfig: func() {
					c, err := getClient()
					assert.NoError(t, err)
					err = c.DeleteRole("test role")
					assert.NoError(t, err)
				},
				Config: `
					resource "sftpgo_role" "test" {
					  name = "test role"
					  description = "desc"
				    }`,
			},
			// Update and Read testing
			{
				Config: `
					resource "sftpgo_role" "test" {
					  name = "test role"
					  description = "desc"
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_role.test", "name", "test role"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "id", "test role"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "description", "desc"),
					resource.TestCheckResourceAttrSet("sftpgo_role.test", "created_at"),
					resource.TestCheckResourceAttrSet("sftpgo_role.test", "updated_at"),
				),
			},
			// External deletion test as last step
			{
				PreConfig: func() {
					c, err := getClient()
					assert.NoError(t, err)
					err = c.DeleteRole("test role")
					assert.NoError(t, err)
				},
				Config: `
					resource "sftpgo_role" "test" {
					  name = "test role"
					  description = "desc"
				    }`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccRoleResource_renameForcesReplace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "sftpgo_role" "test" {
					  name = "rename_test_initial"
					}`,
			},
			{
				Config: `
					resource "sftpgo_role" "test" {
					  name = "rename_test_renamed"
					}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sftpgo_role.test", plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_role.test", "name", "rename_test_renamed"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "id", "rename_test_renamed"),
				),
			},
		},
	})
}

func TestAccEnterpriseRoleResource(t *testing.T) {
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
			// Create and Read testing, every allowlist field is set
			{
				Config: `
					resource "sftpgo_role" "test" {
					  name = "test isolated role"
					  description = "desc"
					  resource_isolation = 1
					  settings = {
						storage_allowlist = {
						  allowed_providers = [0, 1, 2, 3, 4, 5, 6, 7]
						  local = {
							allowed_paths = ["/tmp/tenant1", "/srv/tenant1"]
							users_base_dir = "/tmp/tenant1/homes"
						  }
						  s3 = {
							allowed_buckets = [
							  {
								bucket = "bucket1"
								key_prefix = "tenant1/"
								endpoint = "https://s3.example.com"
							  },
							  {
								bucket = "bucket2"
							  },
							  {
								endpoint = "https://minio.example.com:9000"
							  }
							]
						  }
						  azure = {
							allowed_containers = [
							  {
								account = "account1"
								container = "container1"
								key_prefix = "tenant1/"
								endpoint = "https://azure.example.com"
							  }
							]
						  }
						  gcs = {
							allowed_buckets = [
							  {
								bucket = "bucket1"
								key_prefix = "tenant1/"
								universe_domain = "googleapis.com"
							  }
							]
						  }
						  sftp = {
							allowed_endpoints = ["127.0.0.1:2022", "sftp.example.com:22"]
							allowed_proxies = ["socks5://127.0.0.1:1080", "socks5://proxy.example.com:1080"]
						  }
						  ftp = {
							allowed_endpoints = ["127.0.0.1:2121"]
						  }
						  http = {
							allowed_endpoints = ["https://127.0.0.1:9999/tenant1"]
						  }
						}
					  }
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_role.test", "name", "test isolated role"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "id", "test isolated role"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "description", "desc"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "resource_isolation", "1"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.allowed_providers.#", "8"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.allowed_providers.0", "0"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.allowed_providers.7", "7"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.local.allowed_paths.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.local.allowed_paths.0", "/tmp/tenant1"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.local.allowed_paths.1", "/srv/tenant1"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.local.users_base_dir", "/tmp/tenant1/homes"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.s3.allowed_buckets.#", "3"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.s3.allowed_buckets.0.bucket", "bucket1"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.s3.allowed_buckets.0.key_prefix", "tenant1/"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.s3.allowed_buckets.0.endpoint", "https://s3.example.com"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.s3.allowed_buckets.1.bucket", "bucket2"),
					resource.TestCheckNoResourceAttr("sftpgo_role.test", "settings.storage_allowlist.s3.allowed_buckets.1.key_prefix"),
					resource.TestCheckNoResourceAttr("sftpgo_role.test", "settings.storage_allowlist.s3.allowed_buckets.1.endpoint"),
					// an entry naming the endpoint alone grants every bucket it serves
					resource.TestCheckNoResourceAttr("sftpgo_role.test", "settings.storage_allowlist.s3.allowed_buckets.2.bucket"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.s3.allowed_buckets.2.endpoint", "https://minio.example.com:9000"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.azure.allowed_containers.0.account", "account1"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.azure.allowed_containers.0.container", "container1"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.azure.allowed_containers.0.key_prefix", "tenant1/"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.azure.allowed_containers.0.endpoint", "https://azure.example.com"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.gcs.allowed_buckets.0.bucket", "bucket1"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.gcs.allowed_buckets.0.key_prefix", "tenant1/"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.gcs.allowed_buckets.0.universe_domain", "googleapis.com"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.sftp.allowed_endpoints.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.sftp.allowed_endpoints.0", "127.0.0.1:2022"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.sftp.allowed_endpoints.1", "sftp.example.com:22"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.sftp.allowed_proxies.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.sftp.allowed_proxies.1", "socks5://proxy.example.com:1080"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.ftp.allowed_endpoints.0", "127.0.0.1:2121"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.http.allowed_endpoints.0", "https://127.0.0.1:9999/tenant1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sftpgo_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing: the scopes left out of the allowlist are
			// cleared
			{
				Config: `
					resource "sftpgo_role" "test" {
					  name = "test isolated role"
					  description = "desc"
					  resource_isolation = 1
					  settings = {
						storage_allowlist = {
						  allowed_providers = [0, 1]
						  local = {
							allowed_paths = ["/tmp/tenant1"]
							users_base_dir = "/tmp/tenant1/homes"
						  }
						  s3 = {
							allowed_buckets = [
							  {
								bucket = "bucket1"
							  }
							]
						  }
						  ftp = {
							allowed_endpoints = ["127.0.0.1:2121"]
						  }
						}
					  }
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.allowed_providers.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.local.allowed_paths.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.local.users_base_dir", "/tmp/tenant1/homes"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.s3.allowed_buckets.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.s3.allowed_buckets.0.bucket", "bucket1"),
					resource.TestCheckNoResourceAttr("sftpgo_role.test", "settings.storage_allowlist.s3.allowed_buckets.0.key_prefix"),
					resource.TestCheckNoResourceAttr("sftpgo_role.test", "settings.storage_allowlist.azure"),
					resource.TestCheckNoResourceAttr("sftpgo_role.test", "settings.storage_allowlist.gcs"),
					resource.TestCheckNoResourceAttr("sftpgo_role.test", "settings.storage_allowlist.sftp"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.ftp.allowed_endpoints.0", "127.0.0.1:2121"),
					resource.TestCheckNoResourceAttr("sftpgo_role.test", "settings.storage_allowlist.http"),
				),
			},
			// The base directory of the users is required while the role isolates
			// its resources
			{
				Config: `
					resource "sftpgo_role" "test" {
					  name = "test isolated role"
					  description = "desc"
					  resource_isolation = 1
					  settings = {
						storage_allowlist = {
						  allowed_providers = [0]
						  local = {
							allowed_paths = ["/tmp/tenant1"]
						  }
						}
					  }
					}`,
				ExpectError: regexp.MustCompile("users base directory is required"),
			},
			// Isolation off, the allowlist is kept and enforces nothing
			{
				Config: `
					resource "sftpgo_role" "test" {
					  name = "test isolated role"
					  description = "desc"
					  resource_isolation = 0
					  settings = {
						storage_allowlist = {
						  allowed_providers = [0]
						  local = {
							allowed_paths = ["/tmp/tenant1", "/tmp/shared"]
						  }
						}
					  }
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_role.test", "resource_isolation", "0"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.local.allowed_paths.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_role.test", "settings.storage_allowlist.local.allowed_paths.1", "/tmp/shared"),
					resource.TestCheckNoResourceAttr("sftpgo_role.test", "settings.storage_allowlist.s3"),
					resource.TestCheckNoResourceAttr("sftpgo_role.test", "settings.storage_allowlist.azure"),
					resource.TestCheckNoResourceAttr("sftpgo_role.test", "settings.storage_allowlist.gcs"),
					resource.TestCheckNoResourceAttr("sftpgo_role.test", "settings.storage_allowlist.sftp"),
					resource.TestCheckNoResourceAttr("sftpgo_role.test", "settings.storage_allowlist.ftp"),
					resource.TestCheckNoResourceAttr("sftpgo_role.test", "settings.storage_allowlist.http"),
				),
			},
			// Remove the settings
			{
				Config: `
					resource "sftpgo_role" "test" {
					  name = "test isolated role"
					  description = "desc"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_role.test", "resource_isolation", "0"),
					resource.TestCheckNoResourceAttr("sftpgo_role.test", "settings"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
