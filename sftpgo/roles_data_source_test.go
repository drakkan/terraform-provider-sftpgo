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

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo/client"
)

var (
	testRole = client.Role{
		Name:        "test role",
		Description: "just a test role",
	}
)

func TestAccRolesDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	c, err := getClient()
	require.NoError(t, err)
	_, err = c.CreateRole(testRole)
	require.NoError(t, err)

	defer func() {
		err = c.DeleteRole(testRole.Name)
		require.NoError(t, err)
	}()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: `data "sftpgo_roles" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of roles returned
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.#", "1"),
					// Check the folder fields
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.name", testRole.Name),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.id", testRole.Name),
					resource.TestCheckResourceAttrSet("data.sftpgo_roles.test", "roles.0.created_at"),
					resource.TestCheckResourceAttrSet("data.sftpgo_roles.test", "roles.0.updated_at"),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "id", placeholderID),
				),
			},
		},
	})
}

func TestAccEnterpriseRolesDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	c, err := getClient()
	require.NoError(t, err)
	if !c.IsEnterpriseEdition() {
		t.Skip("This test is supported only with the Enterprise edition")
	}
	role := client.Role{
		Name:              "isolated role",
		Description:       "just a test role",
		ResourceIsolation: 1,
		Settings: client.RoleSettings{
			StorageAllowlist: client.RoleStorageAllowlist{
				AllowedProviders: []int{0, 1},
				Local: client.LocalRoleScope{
					AllowedPaths: []string{"/tmp/tenant1"},
					UsersBaseDir: "/tmp/tenant1/homes",
				},
				S3: client.S3RoleScope{
					DefaultAllow: true,
					AllowedBuckets: []client.S3BucketRef{
						{
							Bucket:    "bucket1",
							KeyPrefix: "tenant1/",
							Endpoint:  "https://s3.example.com",
						},
					},
					DeniedBuckets: []client.S3BucketRef{
						{
							Bucket: "secret",
						},
					},
				},
				Azure: client.AzureRoleScope{
					DefaultAllow: true,
					AllowedContainers: []client.AzureContainerRef{
						{
							Account:   "account1",
							Container: "container1",
							KeyPrefix: "tenant1/",
							Endpoint:  "https://azure.example.com",
						},
					},
					DeniedContainers: []client.AzureContainerRef{
						{
							Account:   "account1",
							Container: "secret",
						},
					},
				},
				GCS: client.GCSRoleScope{
					DefaultAllow: true,
					AllowedBuckets: []client.GCSBucketRef{
						{
							Bucket:         "bucket1",
							KeyPrefix:      "tenant1/",
							UniverseDomain: "googleapis.com",
						},
					},
					DeniedBuckets: []client.GCSBucketRef{
						{
							Bucket: "bucket-secret",
						},
					},
				},
				SFTP: client.SFTPRoleScope{
					EndpointRoleScope: client.EndpointRoleScope{
						DefaultAllow:     true,
						AllowedEndpoints: []string{"127.0.0.1:2022"},
						DeniedEndpoints:  []string{"10.0.0.1:22"},
					},
					AllowedProxies: []string{"127.0.0.1:1080"},
				},
				FTP: client.EndpointRoleScope{
					DefaultAllow:     true,
					AllowedEndpoints: []string{"127.0.0.1:2121"},
					DeniedEndpoints:  []string{"10.0.0.1:21"},
				},
				HTTP: client.EndpointRoleScope{
					DefaultAllow:     true,
					AllowedEndpoints: []string{"https://127.0.0.1:9999/tenant1"},
					DeniedEndpoints:  []string{"https://10.0.0.1/private"},
				},
			},
		},
	}
	_, err = c.CreateRole(role)
	require.NoError(t, err)

	defer func() {
		err = c.DeleteRole(role.Name)
		require.NoError(t, err)
	}()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: `data "sftpgo_roles" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.#", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.name", role.Name),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.resource_isolation", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.allowed_providers.#", "2"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.local.allowed_paths.0", "/tmp/tenant1"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.local.users_base_dir", "/tmp/tenant1/homes"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.s3.default_allow", "true"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.s3.allowed_buckets.0.bucket", "bucket1"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.s3.allowed_buckets.0.key_prefix", "tenant1/"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.s3.allowed_buckets.0.endpoint", "https://s3.example.com"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.s3.denied_buckets.0.bucket", "secret"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.azure.default_allow", "true"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.azure.allowed_containers.0.account", "account1"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.azure.allowed_containers.0.container", "container1"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.azure.allowed_containers.0.key_prefix", "tenant1/"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.azure.allowed_containers.0.endpoint", "https://azure.example.com"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.azure.denied_containers.0.container", "secret"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.gcs.default_allow", "true"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.gcs.allowed_buckets.0.bucket", "bucket1"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.gcs.allowed_buckets.0.key_prefix", "tenant1/"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.gcs.allowed_buckets.0.universe_domain", "googleapis.com"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.gcs.denied_buckets.0.bucket", "bucket-secret"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.sftp.default_allow", "true"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.sftp.allowed_endpoints.0", "127.0.0.1:2022"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.sftp.denied_endpoints.0", "10.0.0.1:22"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.sftp.allowed_proxies.0", "127.0.0.1:1080"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.ftp.default_allow", "true"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.ftp.allowed_endpoints.0", "127.0.0.1:2121"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.ftp.denied_endpoints.0", "10.0.0.1:21"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.http.default_allow", "true"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.http.allowed_endpoints.0", "https://127.0.0.1:9999/tenant1"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "roles.0.settings.storage_allowlist.http.denied_endpoints.0", "https://10.0.0.1/private"),
					resource.TestCheckResourceAttr("data.sftpgo_roles.test", "id", placeholderID),
				),
			},
		},
	})
}
