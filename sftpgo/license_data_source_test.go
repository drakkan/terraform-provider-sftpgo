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

func TestAccLicenseDataSource(t *testing.T) {
	licenseKey1 := os.Getenv("SFTPGO_LICENSE_KEY1")
	licenseKey2 := os.Getenv("SFTPGO_LICENSE_KEY2")
	if licenseKey1 == "" || licenseKey2 == "" {
		t.Skip("Skipping license key environment variables unset")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: `data "sftpgo_license_info" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sftpgo_license_info.test", "license.key"),
					resource.TestCheckResourceAttrSet("data.sftpgo_license_info.test", "license.type"),
					resource.TestCheckResourceAttrSet("data.sftpgo_license_info.test", "license.valid_from"),
					resource.TestCheckResourceAttrSet("data.sftpgo_license_info.test", "license.valid_to"),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.sftpgo_license_info.test", "id", placeholderID),
				),
			},
		},
	})
}
