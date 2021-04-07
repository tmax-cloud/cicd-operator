package approve

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/cli"
	"log"
	"strings"
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

func approvalCommand(kind string, fn func(cmd *cobra.Command, args []string)) *cobra.Command {
	return &cobra.Command{
		Use:   kind + " [Approval] [reason]",
		Short: strings.Title(kind) + "s an Approval",
		Args:  cobra.ExactArgs(2),
		Run:   fn,
	}
}

func (command *command) AddToCommand(cmd *cobra.Command) {
	cmd.AddCommand(command.ApproveCommand)
	cmd.AddCommand(command.RejectCommand)
}

func (command *command) RunCommand(cmd *cobra.Command, args []string) {
	approvals := args[0]
	reason := args[1]

	body, err := json.Marshal(&cicdv1.ApprovalAPIReqBody{
		Reason: reason,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Run!
	client, ns, err := cli.GetClient(command.Config)
	if err != nil {
		log.Fatal(err)
	}

	cli.ExecAndHandleError(client.Put().
		Resource(cicdv1.ApprovalKind).
		Namespace(ns).
		Name(approvals).
		SubResource(cmd.Name()).
		Body(body))

	fmt.Printf(strings.Title(cmd.Name())+"ed Approval %s/%s\n", ns, approvals)
}
