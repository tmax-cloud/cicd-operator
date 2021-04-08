package webhook

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/cli"
	"log"
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
