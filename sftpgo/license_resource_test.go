// Copyright (C) 2025 Nicola Murino
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
)

func TestAccLicenseResource(t *testing.T) {
	licenseKey1 := os.Getenv("SFTPGO_LICENSE_KEY1")
	licenseKey2 := os.Getenv("SFTPGO_LICENSE_KEY2")
	if licenseKey1 == "" || licenseKey2 == "" {
		t.Skip("Skipping license key environment variables unset")
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: `
					resource "sftpgo_license" "test" {
  					  key = "` + licenseKey1 + `"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_license.test", "key", licenseKey1),
					resource.TestCheckResourceAttrSet("sftpgo_license.test", "type"),
					resource.TestCheckResourceAttrSet("sftpgo_license.test", "valid_from"),
					resource.TestCheckResourceAttrSet("sftpgo_license.test", "valid_to"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sftpgo_license.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: `
								resource "sftpgo_license" "test" {
			  					  key = "` + licenseKey2 + `"
								}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_license.test", "key", licenseKey2),
					resource.TestCheckResourceAttrSet("sftpgo_license.test", "type"),
					resource.TestCheckResourceAttrSet("sftpgo_license.test", "valid_from"),
					resource.TestCheckResourceAttrSet("sftpgo_license.test", "valid_to"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
