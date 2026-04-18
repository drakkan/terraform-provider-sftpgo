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
	"path/filepath"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sftpgo/sdk"
	"github.com/sftpgo/sdk/kms"
	"github.com/stretchr/testify/require"

	"github.com/drakkan/terraform-provider-sftpgo/sftpgo/client"
)

func TestAccActionsDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	c, err := getClient()
	require.NoError(t, err)
	action := client.BaseEventAction{
		Name:        "action",
		Description: "desc",
		Type:        2,
		Options: client.EventActionOptions{
			CmdConfig: client.EventActionCommandConfig{
				Cmd:     "/bin/true",
				Args:    []string{"arg1", "arg2"},
				Timeout: 20,
				EnvVars: []client.KeyValue{
					{
						Key:   "ENV1",
						Value: "VAL1",
					},
				},
			},
		},
	}
	_, err = c.CreateAction(action)
	require.NoError(t, err)

	defer func() {
		err = c.DeleteAction(action.Name)
		require.NoError(t, err)
	}()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: `data "sftpgo_actions" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of actions returned
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.#", "1"),
					// Check the created action
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.name", action.Name),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.id", action.Name),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.description", action.Description),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.type", fmt.Sprintf("%d", action.Type)),
					resource.TestCheckNoResourceAttr("data.sftpgo_actions.test", "actions.0.options.http_config"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.options.cmd_config.cmd",
						action.Options.CmdConfig.Cmd),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.options.cmd_config.args.#", "2"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.options.cmd_config.args.0",
						action.Options.CmdConfig.Args[0]),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.options.cmd_config.args.1",
						action.Options.CmdConfig.Args[1]),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.options.cmd_config.timeout",
						fmt.Sprintf("%d", action.Options.CmdConfig.Timeout)),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.options.cmd_config.env_vars.#", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.options.cmd_config.env_vars.0.key",
						action.Options.CmdConfig.EnvVars[0].Key),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.options.cmd_config.env_vars.0.value",
						action.Options.CmdConfig.EnvVars[0].Value),
					resource.TestCheckNoResourceAttr("data.sftpgo_actions.test", "actions.0.options.email_config"),
					resource.TestCheckNoResourceAttr("data.sftpgo_actions.test", "actions.0.options.retention_config"),
					resource.TestCheckNoResourceAttr("data.sftpgo_actions.test", "actions.0.options.fs_config"),
					resource.TestCheckNoResourceAttr("data.sftpgo_actions.test", "actions.0.options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("data.sftpgo_actions.test", "actions.0.options.user_inactivity_config"),
					resource.TestCheckNoResourceAttr("data.sftpgo_actions.test", "actions.0.options.idp_config"),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "id", placeholderID),
				),
			},
		},
	})
}

func TestAccEnterpriseActionsDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	c, err := getClient()
	require.NoError(t, err)
	if !c.IsEnterpriseEdition() {
		t.Skip("This test is supported only with the Enterprise edition")
	}

	folder, err := c.CreateFolder(testFolder)
	require.NoError(t, err)

	f1 := client.BaseVirtualFolder{
		BaseVirtualFolder: sdk.BaseVirtualFolder{
			Name:       "target_folder",
			MappedPath: filepath.Join(os.TempDir(), "target_folder"),
		},
	}

	targetFolder, err := c.CreateFolder(f1)
	require.NoError(t, err)

	action := client.BaseEventAction{
		Name:        "action pgp",
		Description: "desc",
		Type:        client.ActionTypeFilesystem,
		Options: client.EventActionOptions{
			FsConfig: client.EventActionFilesystemConfig{
				Type: client.FilesystemActionPGP,
				PGP: client.EventActionPGP{
					Mode:    1,
					Profile: 1,
					Paths: []client.KeyValue{
						{
							Key:   "/{{.VirtualPath}}",
							Value: "/{{.VirtualPath}}.pgp",
						},
					},
					Password: kms.BaseSecret{
						Status:  kms.SecretStatusPlain,
						Payload: "password",
					},
				},
				Folder:       folder.Name,
				TargetFolder: targetFolder.Name,
			},
		},
	}
	_, err = c.CreateAction(action)
	require.NoError(t, err)

	action1 := client.BaseEventAction{
		Name: "metadata check",
		Type: client.ActionTypeFilesystem,
		Options: client.EventActionOptions{
			FsConfig: client.EventActionFilesystemConfig{
				Type: client.FilesystemActionMetadataCheck,
				MetadataCheck: client.EventActionMetadataCheck{
					Path: "/test",
					Metadata: client.KeyValue{
						Key:   "k",
						Value: "v",
					},
					Timeout: 10,
				},
			},
		},
	}
	_, err = c.CreateAction(action1)
	require.NoError(t, err)

	action2 := client.BaseEventAction{
		Name: "z_copy_extended",
		Type: client.ActionTypeFilesystem,
		Options: client.EventActionOptions{
			FsConfig: client.EventActionFilesystemConfig{
				Type: client.FilesystemActionCopy,
				Copy: []client.CopyConfig{
					{
						KeyValue:               client.KeyValue{Key: "/src", Value: "/dst"},
						OnSourceCopied:         2,
						OnSourceCopiedMovePath: "/archive",
						MaxRetries:             4,
					},
				},
				ContinueOnError: true,
			},
		},
	}
	_, err = c.CreateAction(action2)
	require.NoError(t, err)

	action3 := client.BaseEventAction{
		Name: "z_event_report",
		Type: client.ActionTypeEventReport,
		Options: client.EventActionOptions{
			EventReportConfig: client.EventActionEventReportConfig{
				TimeWindow:   30,
				FsActions:    []string{"upload", "download"},
				Statuses:     []int32{1, 2},
				SplitReports: true,
			},
		},
	}
	_, err = c.CreateAction(action3)
	require.NoError(t, err)

	action4 := client.BaseEventAction{
		Name: "z_email_report",
		Type: client.ActionTypeEmail,
		Options: client.EventActionOptions{
			EmailConfig: client.EventActionEmailConfig{
				Recipients:        []string{"ops@example.com"},
				Subject:           "s",
				Body:              "b",
				AttachEventReport: true,
			},
		},
	}
	_, err = c.CreateAction(action4)
	require.NoError(t, err)

	defer func() {
		err = c.DeleteAction(action.Name)
		require.NoError(t, err)
		err = c.DeleteAction(action1.Name)
		require.NoError(t, err)
		err = c.DeleteAction(action2.Name)
		require.NoError(t, err)
		err = c.DeleteAction(action3.Name)
		require.NoError(t, err)
		err = c.DeleteAction(action4.Name)
		require.NoError(t, err)
		err = c.DeleteFolder(folder.Name)
		require.NoError(t, err)
		err = c.DeleteFolder(targetFolder.Name)
		require.NoError(t, err)
	}()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: `data "sftpgo_actions" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of actions returned
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.#", "5"),
					// Check the created action
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.name", action.Name),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.id", action.Name),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.description", action.Description),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.type", fmt.Sprintf("%d", action.Type)),
					resource.TestCheckNoResourceAttr("data.sftpgo_actions.test", "actions.0.options.http_config"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.options.fs_config.folder",
						action.Options.FsConfig.Folder),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.options.fs_config.target_folder",
						action.Options.FsConfig.TargetFolder),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.options.fs_config.pgp.mode",
						strconv.Itoa(action.Options.FsConfig.PGP.Mode)),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.options.fs_config.pgp.profile",
						strconv.Itoa(action.Options.FsConfig.PGP.Profile)),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.options.fs_config.pgp.paths.#", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.options.fs_config.pgp.paths.0.key",
						action.Options.FsConfig.PGP.Paths[0].Key),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.0.options.fs_config.pgp.paths.0.value",
						action.Options.FsConfig.PGP.Paths[0].Value),
					resource.TestCheckResourceAttrSet("data.sftpgo_actions.test", "actions.0.options.fs_config.pgp.password"),
					resource.TestCheckNoResourceAttr("data.sftpgo_actions.test", "actions.0.options.fs_config.pgp.public_key"),
					resource.TestCheckNoResourceAttr("data.sftpgo_actions.test", "actions.0.options.fs_config.pgp.private_key"),
					resource.TestCheckNoResourceAttr("data.sftpgo_actions.test", "actions.0.options.fs_config.pgp.passphrase"),
					resource.TestCheckNoResourceAttr("data.sftpgo_actions.test", "actions.0.options.email_config"),
					resource.TestCheckNoResourceAttr("data.sftpgo_actions.test", "actions.0.options.retention_config"),
					resource.TestCheckNoResourceAttr("data.sftpgo_actions.test", "actions.0.options.cmd_config"),
					resource.TestCheckNoResourceAttr("data.sftpgo_actions.test", "actions.0.options.pwd_expiration_config"),
					resource.TestCheckNoResourceAttr("data.sftpgo_actions.test", "actions.0.options.user_inactivity_config"),
					resource.TestCheckNoResourceAttr("data.sftpgo_actions.test", "actions.0.options.idp_config"),

					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.1.name", action1.Name),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.1.type", fmt.Sprintf("%d", action1.Type)),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.1.options.fs_config.type", "8"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.1.options.fs_config.metadata_check.path", "/test"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.1.options.fs_config.metadata_check.metadata.key", "k"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.1.options.fs_config.metadata_check.metadata.value", "v"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.1.options.fs_config.metadata_check.timeout", "10"),

					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.2.name", action2.Name),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.2.type", fmt.Sprintf("%d", action2.Type)),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.2.options.fs_config.type", "6"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.2.options.fs_config.copy.#", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.2.options.fs_config.copy.0.key", "/src"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.2.options.fs_config.copy.0.value", "/dst"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.2.options.fs_config.copy.0.on_source_copied", "2"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.2.options.fs_config.copy.0.on_source_copied_move_path", "/archive"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.2.options.fs_config.copy.0.max_retries", "4"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.2.options.fs_config.continue_on_error", "true"),

					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.3.name", action3.Name),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.3.type", fmt.Sprintf("%d", action3.Type)),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.3.options.event_report_config.time_window", "30"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.3.options.event_report_config.fs_actions.#", "2"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.3.options.event_report_config.fs_actions.0", "upload"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.3.options.event_report_config.fs_actions.1", "download"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.3.options.event_report_config.statuses.#", "2"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.3.options.event_report_config.statuses.0", "1"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.3.options.event_report_config.statuses.1", "2"),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.3.options.event_report_config.split_reports", "true"),

					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.4.name", action4.Name),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.4.type", fmt.Sprintf("%d", action4.Type)),
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "actions.4.options.email_config.attach_event_report", "true"),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.sftpgo_actions.test", "id", placeholderID),
				),
			},
		},
	})
}
