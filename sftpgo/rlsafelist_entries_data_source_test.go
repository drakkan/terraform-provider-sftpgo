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
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo/client"
)

func TestAccRateLimitersListDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	c, err := getClient()
	require.NoError(t, err)
	entry := client.IPListEntry{
		IPOrNet:     "192.168.7.0/25",
		Description: "",
		Type:        3,
		Mode:        1,
		Protocols:   0,
	}
	_, err = c.CreateIPListEntry(entry)
	require.NoError(t, err)

	defer func() {
		err = c.DeleteIPListEntry(entry.Type, entry.IPOrNet)
		require.NoError(t, err)
	}()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: `data "sftpgo_rlsafelist_entries" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of admins returned
					resource.TestCheckResourceAttr("data.sftpgo_rlsafelist_entries.test", "entries.#", "1"),
					// Check the created entry
					resource.TestCheckResourceAttr("data.sftpgo_rlsafelist_entries.test", "entries.0.ipornet", entry.IPOrNet),
					resource.TestCheckResourceAttr("data.sftpgo_rlsafelist_entries.test", "entries.0.id", entry.IPOrNet),
					resource.TestCheckNoResourceAttr("data.sftpgo_rlsafelist_entries.test", "entries.0.description"),
					resource.TestCheckResourceAttr("data.sftpgo_rlsafelist_entries.test", "entries.0.protocols", fmt.Sprintf("%d", entry.Protocols)),
					resource.TestCheckResourceAttrSet("data.sftpgo_rlsafelist_entries.test", "entries.0.created_at"),
					resource.TestCheckResourceAttrSet("data.sftpgo_rlsafelist_entries.test", "entries.0.updated_at"),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.sftpgo_rlsafelist_entries.test", "id", placeholderID),
				),
			},
		},
	})
}
