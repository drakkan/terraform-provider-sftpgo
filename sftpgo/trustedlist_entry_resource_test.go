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

func TestAccEnterpriseTrustedListResource(t *testing.T) {
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
					resource "sftpgo_trustedlist_entry" "test" {
  					  ipornet = "172.16.4.0/24"
					  protocols = 0
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_trustedlist_entry.test", "ipornet", "172.16.4.0/24"),
					resource.TestCheckResourceAttr("sftpgo_trustedlist_entry.test", "id", "172.16.4.0/24"),
					resource.TestCheckNoResourceAttr("sftpgo_trustedlist_entry.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_trustedlist_entry.test", "protocols", "0"),
					resource.TestCheckResourceAttrSet("sftpgo_trustedlist_entry.test", "created_at"),
					resource.TestCheckResourceAttrSet("sftpgo_trustedlist_entry.test", "updated_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sftpgo_trustedlist_entry.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: `
					resource "sftpgo_trustedlist_entry" "test" {
					  ipornet = "172.16.4.0/24"
					  protocols = 7
					  description = "desc"
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_trustedlist_entry.test", "ipornet", "172.16.4.0/24"),
					resource.TestCheckResourceAttr("sftpgo_trustedlist_entry.test", "id", "172.16.4.0/24"),
					resource.TestCheckResourceAttr("sftpgo_trustedlist_entry.test", "description", "desc"),
					resource.TestCheckResourceAttr("sftpgo_trustedlist_entry.test", "protocols", "7"),
					resource.TestCheckResourceAttrSet("sftpgo_trustedlist_entry.test", "created_at"),
					resource.TestCheckResourceAttrSet("sftpgo_trustedlist_entry.test", "updated_at"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccEnterpriseTrustedListResource_renameForcesReplace(t *testing.T) {
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
			{
				Config: `
					resource "sftpgo_trustedlist_entry" "test" {
					  ipornet = "198.51.101.0/24"
					  protocols = 0
					}`,
			},
			{
				Config: `
					resource "sftpgo_trustedlist_entry" "test" {
					  ipornet = "198.51.101.1/32"
					  protocols = 0
					}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sftpgo_trustedlist_entry.test", plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_trustedlist_entry.test", "ipornet", "198.51.101.1/32"),
					resource.TestCheckResourceAttr("sftpgo_trustedlist_entry.test", "id", "198.51.101.1/32"),
				),
			},
		},
	})
}
