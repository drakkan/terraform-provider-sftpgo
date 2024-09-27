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

func TestAccRuleResource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	c, err := getClient()
	require.NoError(t, err)
	action1 := client.BaseEventAction{
		Name: "action1",
		Type: 4,
	}
	action2 := client.BaseEventAction{
		Name: "action2",
		Type: 5,
	}
	_, err = c.CreateAction(action1)
	require.NoError(t, err)
	_, err = c.CreateAction(action2)
	require.NoError(t, err)

	defer func() {
		err = c.DeleteAction(action1.Name)
		require.NoError(t, err)
		err = c.DeleteAction(action2.Name)
		require.NoError(t, err)
	}()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: `
					resource "sftpgo_rule" "test" {
  					  name = "test rule"
					  status = 1
					  description = "test desc"
					  trigger = 1
					  conditions = {
						fs_events = ["upload"]
						options = {
						  group_names = [
							{
							    pattern = "group*"
							}
						  ]
						  role_names = [
							{
							    pattern = "role*"
								inverse_match = true
							}
						  ]
						  fs_paths = [
							{
							    pattern = "/*.txt"
								inverse_match = true
							},
							{
							    pattern = "/**/*.txt"
							}
						  ]
						  protocols = ["SFTP", "SCP"]
						  min_size = 1
						  max_size = 100
						}
					  }
					  actions = [
						{
							name = "action1"
							is_failure_action = true
						},
						{
							name = "action2"
							execute_sync = true
							stop_on_failure = true
						}
					  ]
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_rule.test", "name", "test rule"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "id", "test rule"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "status", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "description", "test desc"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "trigger", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.%", "5"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.fs_events.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.fs_events.0", "upload"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.provider_events"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.schedules"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.idp_login_event"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.%", "10"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.names"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.group_names.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.group_names.0.pattern", "group*"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.group_names.0.inverse_match"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.role_names.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.role_names.0.pattern", "role*"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.role_names.0.inverse_match", "true"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.fs_paths.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.fs_paths.0.pattern", "/*.txt"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.fs_paths.0.inverse_match", "true"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.fs_paths.1.pattern", "/**/*.txt"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.fs_paths.1.inverse_match"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.protocols.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.protocols.0", "SFTP"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.protocols.1", "SCP"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.event_statuses.#", "0"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.provider_objects"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.min_size", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.max_size", "100"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.concurrent_execution"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.0.name", "action1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.0.is_failure_action", "true"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "actions.0.execute_sync"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "actions.0.stop_on_failure"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.1.name", "action2"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "actions.1.is_failure_action"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.1.execute_sync", "true"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.1.stop_on_failure", "true"),
					resource.TestCheckResourceAttrSet("sftpgo_rule.test", "created_at"),
					resource.TestCheckResourceAttrSet("sftpgo_rule.test", "updated_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sftpgo_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: `
					resource "sftpgo_rule" "test" {
						name = "test rule"
					  	status = 0
					  	trigger = 3
					  	conditions = {
						  schedules = [
							{
								hour = "0"
								day_of_week = "*"
								day_of_month = "*"
								month = "*"
							}
						  ]
						  options = {
							names = [
								{
									pattern = "user*"
								}
							]
						  }
					  	}
					  	actions = [
						  {
							name = "action2"
						  }
					  	]
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_rule.test", "name", "test rule"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "id", "test rule"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "status", "0"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "trigger", "3"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.%", "5"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.fs_events"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.provider_events"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.schedules.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.schedules.0.hour", "0"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.schedules.0.day_of_week", "*"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.schedules.0.day_of_month", "*"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.schedules.0.month", "*"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.idp_login_event"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.%", "10"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.names.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.names.0.pattern", "user*"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.group_names"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.role_names"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.fs_paths"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.protocols"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.provider_objects"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.min_size"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.max_size"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.concurrent_execution"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.0.name", "action2"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "actions.0.is_failure_action"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "actions.0.execute_sync"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "actions.0.stop_on_failure"),
					resource.TestCheckResourceAttrSet("sftpgo_rule.test", "created_at"),
					resource.TestCheckResourceAttrSet("sftpgo_rule.test", "updated_at"),
				),
			},
			{
				ResourceName:      "sftpgo_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: `
					resource "sftpgo_rule" "test" {
  					  name = "test rule"
					  status = 1
					  description = "test provider event"
					  trigger = 2
					  conditions = {
						provider_events = ["add"]
						options = {
						  provider_objects = ["user"]
						}
					  }
					  actions = [
						{
							name = "action1"
						}
					  ]
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_rule.test", "name", "test rule"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "id", "test rule"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "status", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "description", "test provider event"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "trigger", "2"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.%", "5"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.provider_events.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.provider_events.0", "add"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.fs_events"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.schedules"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.idp_login_event"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.%", "10"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.names"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.provider_objects.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.provider_objects.0", "user"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.group_names.0.inverse_match"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.group_names"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.role_names"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.fs_paths"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.protocols"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.min_size"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.max_size"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.concurrent_execution"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.0.name", "action1"),
					resource.TestCheckResourceAttrSet("sftpgo_rule.test", "created_at"),
					resource.TestCheckResourceAttrSet("sftpgo_rule.test", "updated_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sftpgo_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: `
					resource "sftpgo_rule" "test" {
						name = "test rule"
					  	status = 0
					  	trigger = 7
					  	conditions = {
							idp_login_event = 0
					  	}
					  	actions = [
						  {
							name = "action2"
						  }
					  	]
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_rule.test", "name", "test rule"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "id", "test rule"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "status", "0"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "trigger", "7"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.%", "5"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.fs_events"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.provider_events"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.schedules"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.idp_login_event", "0"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.%", "10"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.names"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.group_names"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.role_names"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.fs_paths"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.protocols"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.provider_objects"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.min_size"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.max_size"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.concurrent_execution"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.0.name", "action2"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "actions.0.is_failure_action"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "actions.0.execute_sync"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "actions.0.stop_on_failure"),
					resource.TestCheckResourceAttrSet("sftpgo_rule.test", "created_at"),
					resource.TestCheckResourceAttrSet("sftpgo_rule.test", "updated_at"),
				),
			},
			{
				ResourceName:      "sftpgo_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: `
					resource "sftpgo_rule" "test" {
						name = "test rule"
					  	status = 1
					  	trigger = 7
					  	conditions = {
							idp_login_event = 1
					  	}
					  	actions = [
						  {
							name = "action2"
						  },
						  {
							name = "action1"
							is_failure_action = true
						  }
					  	]
				    }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_rule.test", "name", "test rule"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "id", "test rule"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "status", "1"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "description"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "trigger", "7"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.%", "5"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.fs_events"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.provider_events"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.schedules"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.idp_login_event", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.%", "10"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.names"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.group_names"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.role_names"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.fs_paths"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.protocols"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.provider_objects"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.min_size"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.max_size"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.concurrent_execution"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.0.name", "action2"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "actions.0.is_failure_action"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "actions.0.execute_sync"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "actions.0.stop_on_failure"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.1.name", "action1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.1.is_failure_action", "true"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "actions.1.execute_sync"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "actions.1.stop_on_failure"),
					resource.TestCheckResourceAttrSet("sftpgo_rule.test", "created_at"),
					resource.TestCheckResourceAttrSet("sftpgo_rule.test", "updated_at"),
				),
			},
			{
				Config: `
					resource "sftpgo_rule" "test" {
  					  name = "test rule"
					  status = 1
					  description = "test desc"
					  trigger = 1
					  conditions = {
						fs_events = ["upload"]
						options = {
						  group_names = [
							{
							    pattern = "group*"
							}
						  ]
						  role_names = [
							{
							    pattern = "role*"
								inverse_match = true
							}
						  ]
						  fs_paths = [
							{
							    pattern = "/*.txt"
								inverse_match = true
							},
							{
							    pattern = "/**/*.txt"
							}
						  ]
						  protocols = ["SFTP", "SCP"]
						  min_size = 1
						  max_size = 100
						  event_statuses = [1, 3]
						}
					  }
					  actions = [
						{
							name = "action1"
							is_failure_action = true
						},
						{
							name = "action2"
							execute_sync = true
							stop_on_failure = true
						}
					  ]
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sftpgo_rule.test", "name", "test rule"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "id", "test rule"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "status", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "description", "test desc"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "trigger", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.%", "5"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.fs_events.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.fs_events.0", "upload"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.provider_events"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.schedules"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.idp_login_event"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.%", "10"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.names"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.group_names.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.group_names.0.pattern", "group*"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.group_names.0.inverse_match"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.role_names.#", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.role_names.0.pattern", "role*"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.role_names.0.inverse_match", "true"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.fs_paths.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.fs_paths.0.pattern", "/*.txt"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.fs_paths.0.inverse_match", "true"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.fs_paths.1.pattern", "/**/*.txt"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.fs_paths.1.inverse_match"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.protocols.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.protocols.0", "SFTP"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.protocols.1", "SCP"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.event_statuses.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.event_statuses.0", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.event_statuses.1", "3"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.provider_objects"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.min_size", "1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "conditions.options.max_size", "100"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "conditions.options.concurrent_execution"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.#", "2"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.0.name", "action1"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.0.is_failure_action", "true"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "actions.0.execute_sync"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "actions.0.stop_on_failure"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.1.name", "action2"),
					resource.TestCheckNoResourceAttr("sftpgo_rule.test", "actions.1.is_failure_action"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.1.execute_sync", "true"),
					resource.TestCheckResourceAttr("sftpgo_rule.test", "actions.1.stop_on_failure", "true"),
					resource.TestCheckResourceAttrSet("sftpgo_rule.test", "created_at"),
					resource.TestCheckResourceAttrSet("sftpgo_rule.test", "updated_at"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
