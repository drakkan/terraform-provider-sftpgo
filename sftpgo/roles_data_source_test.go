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
