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
	"github.com/stretchr/testify/assert"
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
