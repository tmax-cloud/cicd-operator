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

package webhook

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/cli"
)

type command struct {
	*cobra.Command

	Config *cli.Configs
}

// New is a constructor of a webhook sub-command
func New(c *cli.Configs) cli.Command {
	cmd := &command{Config: c}
	cmd.Command = &cobra.Command{
		Use:   "webhook [IntegrationConfig]",
		Short: "Gets webhook information of an IntegrationConfig",
		Args:  cobra.ExactArgs(1),
		Run:   cmd.RunCommand,
	}

	return cmd
}

func (command *command) AddToCommand(cmd *cobra.Command) {
	cmd.AddCommand(command.Command)
}

func (command *command) RunCommand(_ *cobra.Command, args []string) {
	ic := args[0]

	// Run!
	client, ns, err := cli.GetClient(command.Config)
	if err != nil {
		log.Fatal(err)
	}

	cli.ExecAndHandleError(client.Get().
		Resource(cicdv1.IntegrationConfigKind).
		Namespace(ns).
		Name(ic).
		SubResource(cicdv1.IntegrationConfigAPIWebhookURL), printWebhookInfo)
}

func printWebhookInfo(raw []byte) {
	obj := &cicdv1.IntegrationConfigAPIReqWebhookURL{}

	if err := json.Unmarshal(raw, obj); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Webhook URL\t: %s\n", obj.URL)
	fmt.Printf("Webhook Secret\t: %s\n", obj.Secret)
}
