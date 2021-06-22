// Copyright 2021 Workrise Technologies Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package eiam

import (
	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	eiam "github.com/rigup/ephemeral-iam/internal"
	"github.com/rigup/ephemeral-iam/pkg/options"
)

// RootCommand is the top level cobra command.
var RootCommand *eiam.RootCommand

// NewEphemeralIamCommand returns cobra.Command to run eiam command.
func NewEphemeralIamCommand() (*eiam.RootCommand, error) {
	cmds := &eiam.RootCommand{Command: cobra.Command{
		Use:   "eiam",
		Short: "Utility for granting short-lived, privileged access to GCP APIs.",
		Long: dedent.Dedent(`
			╭────────────────────────────────────────────────────────────╮
			│                                                            │
			│                        ephemeral-iam                       │
			│  ──────────────────────────────────────────────────────    │
			│  A CLI tool for temporarily escalating GCP IAM privileges  │
			│  to perform high privilege tasks.                          │
			│                                                            │
			│           https://github.com/rigup/ephemeral-iam           │
			│                                                            │
			╰────────────────────────────────────────────────────────────╯
			
			
			╭────────────────────── Example usage ───────────────────────╮
			│                                                            │
			│                   Start privileged session                 │
			│  ──────────────────────────────────────────────────────    │
			│  $ eiam assume-privileges \                                │
			│      -s example-svc@my-project.iam.gserviceaccount.com \   │
			│      --reason "Emergency security patch (JIRA-1234)"       │
			│                                                            │
			│                                                            │
			│                                                            │
			│                     Run gcloud command                     │
			│  ──────────────────────────────────────────────────────    │
			│  $ eiam gcloud compute instances list --format=json \      │
			│      -s example@my-project.iam.gserviceaccount.com \       │
			│      -R "Reason"                                           │
			│                                                            │
			╰────────────────────────────────────────────────────────────╯
		
			Please report any bugs or feature requests by opening a new
			issue at https://github.com/rigup/ephemeral-iam/issues
		`),
		SilenceErrors: true,
		SilenceUsage:  true,
	}}

	cmds.ResetFlags()

	cmds.AddCommand(newCmdAssumePrivileges())
	cmds.AddCommand(newCmdCloudSQLProxy())
	cmds.AddCommand(newCmdConfig())
	cmds.AddCommand(newCmdDefaultServiceAccounts())
	cmds.AddCommand(newCmdGcloud())
	cmds.AddCommand(newCmdKubectl())
	cmds.AddCommand(newCmdListServiceAccounts())
	cmds.AddCommand(newCmdPlugins())
	cmds.AddCommand(newCmdQueryPermissions())
	cmds.AddCommand(newCmdVersion())
	if err := cmds.LoadPlugins(); err != nil {
		return nil, err
	}
	options.AddPersistentFlags(cmds.PersistentFlags())

	RootCommand = cmds

	return cmds, nil
}
