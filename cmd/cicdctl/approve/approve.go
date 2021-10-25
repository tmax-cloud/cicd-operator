/*
 Copyright 2021 The CI/CD Operator Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package approve

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/cli"
)

type command struct {
	ApproveCommand *cobra.Command
	RejectCommand  *cobra.Command

	Config *cli.Configs
}

// New is a constructor of the approve sub-command
func New(c *cli.Configs) cli.Command {
	cmd := &command{Config: c}
	cmd.ApproveCommand = approvalCommand(cicdv1.ApprovalAPIApprove, cmd.RunCommand)
	cmd.RejectCommand = approvalCommand(cicdv1.ApprovalAPIReject, cmd.RunCommand)
	return cmd
}

func approvalCommand(kind string, fn func(cmd *cobra.Command, args []string) error) *cobra.Command {
	return &cobra.Command{
		Use:   kind + " [Approval] [reason]",
		Short: strings.Title(kind) + "s an Approval",
		Args:  cobra.ExactArgs(2),
		RunE:  fn,
	}
}

func (command *command) AddToCommand(cmd *cobra.Command) {
	cmd.AddCommand(command.ApproveCommand)
	cmd.AddCommand(command.RejectCommand)
}

func (command *command) RunCommand(cmd *cobra.Command, args []string) error {
	approvals := args[0]
	reason := args[1]

	body, err := json.Marshal(&cicdv1.ApprovalAPIReqBody{
		Reason: reason,
	})
	if err != nil {
		return err
	}

	// Run!
	client, ns, err := cli.GetClient(command.Config)
	if err != nil {
		return err
	}

	return cli.ExecAndHandleError(client.Put().
		Resource(cicdv1.ApprovalKind).
		Namespace(ns).
		Name(approvals).
		SubResource(cmd.Name()).
		Body(body), func(_ []byte) error {
		fmt.Printf(strings.Title(cmd.Name())+"ed Approval %s/%s\n", ns, approvals)
		return nil
	})
}
