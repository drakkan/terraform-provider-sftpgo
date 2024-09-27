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

func TestAccRulesDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	c, err := getClient()
	require.NoError(t, err)
	action1 := client.BaseEventAction{
		Name: "action1",
		Type: 4,
	}
	_, err = c.CreateAction(action1)
	require.NoError(t, err)
	rule := client.EventRule{
		Name:    "test rule",
		Status:  0,
		Trigger: 2,
		Conditions: client.EventRuleConditions{
			ProviderEvents: []string{"add", "update"},
			Options: client.ConditionOptions{
				ProviderObjects: []string{"user", "group"},
			},
		},
		Actions: []client.EventAction{
			{
				Name: action1.Name,
				Options: client.EventActionRelationOptions{
					StopOnFailure: true,
				},
			},
		},
	}
	_, err = c.CreateRule(rule)
	require.NoError(t, err)

	defer func() {
		err = c.DeleteRule(rule.Name)
		require.NoError(t, err)
		err = c.DeleteAction(action1.Name)
		require.NoError(t, err)
	}()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: `data "sftpgo_rules" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of actions returned
					resource.TestCheckResourceAttr("data.sftpgo_rules.test", "rules.#", "1"),
					// Check the created action
					resource.TestCheckResourceAttr("data.sftpgo_rules.test", "rules.0.name", rule.Name),
					resource.TestCheckResourceAttr("data.sftpgo_rules.test", "rules.0.id", rule.Name),
					resource.TestCheckResourceAttr("data.sftpgo_rules.test", "rules.0.status", fmt.Sprintf("%d", rule.Status)),
					resource.TestCheckNoResourceAttr("data.sftpgo_rules.test", "rules.0.description"),
					resource.TestCheckResourceAttr("data.sftpgo_rules.test", "rules.0.trigger", fmt.Sprintf("%d", rule.Trigger)),
					resource.TestCheckResourceAttr("data.sftpgo_rules.test", "rules.0.conditions.%", "5"),
					resource.TestCheckNoResourceAttr("data.sftpgo_rules.test", "rules.0.conditions.fs_events"),
					resource.TestCheckResourceAttr("data.sftpgo_rules.test", "rules.0.conditions.provider_events.#", "2"),
					resource.TestCheckResourceAttr("data.sftpgo_rules.test", "rules.0.conditions.provider_events.0", "add"),
					resource.TestCheckResourceAttr("data.sftpgo_rules.test", "rules.0.conditions.provider_events.1", "update"),
					resource.TestCheckNoResourceAttr("data.sftpgo_rules.test", "rules.0.conditions.schedules"),
					resource.TestCheckNoResourceAttr("data.sftpgo_rules.test", "rules.0.conditions.idp_login_event"),
					resource.TestCheckResourceAttr("data.sftpgo_rules.test", "rules.0.conditions.options.%", "10"),
					resource.TestCheckNoResourceAttr("data.sftpgo_rules.test", "rules.0.conditions.options.names"),
					resource.TestCheckNoResourceAttr("data.sftpgo_rules.test", "rules.0.conditions.options.group_names"),
					resource.TestCheckNoResourceAttr("data.sftpgo_rules.test", "rules.0.conditions.options.role_names"),
					resource.TestCheckNoResourceAttr("data.sftpgo_rules.test", "rules.0.conditions.options.fs_paths"),
					resource.TestCheckNoResourceAttr("data.sftpgo_rules.test", "rules.0.conditions.options.protocols"),
					resource.TestCheckResourceAttr("data.sftpgo_rules.test", "rules.0.conditions.options.provider_objects.#", "2"),
					resource.TestCheckResourceAttr("data.sftpgo_rules.test", "rules.0.conditions.options.provider_objects.0", "user"),
					resource.TestCheckResourceAttr("data.sftpgo_rules.test", "rules.0.conditions.options.provider_objects.1", "group"),
					resource.TestCheckNoResourceAttr("data.sftpgo_rules.test", "rules.0.conditions.options.min_size"),
					resource.TestCheckNoResourceAttr("data.sftpgo_rules.test", "rules.0.conditions.options.max_size"),
					resource.TestCheckNoResourceAttr("data.sftpgo_rules.test", "rules.0.conditions.options.concurrent_execution"),
					resource.TestCheckResourceAttr("data.sftpgo_rules.test", "rules.0.actions.#", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_rules.test", "rules.0.actions.0.name", action1.Name),
					resource.TestCheckResourceAttr("data.sftpgo_rules.test", "rules.0.actions.0.stop_on_failure", "true"),
					resource.TestCheckNoResourceAttr("data.sftpgo_rules.test", "rules.0.actions.0.is_failure_action"),
					resource.TestCheckNoResourceAttr("data.sftpgo_rules.test", "rules.0.actions.0.execute_sync"),
					resource.TestCheckResourceAttrSet("data.sftpgo_rules.test", "rules.0.created_at"),
					resource.TestCheckResourceAttrSet("data.sftpgo_rules.test", "rules.0.updated_at"),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.sftpgo_rules.test", "id", placeholderID),
				),
			},
		},
	})
}
