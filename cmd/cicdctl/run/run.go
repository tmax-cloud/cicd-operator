package run

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/cli"
	"log"
)

const (
	subTypePre  = "pre"
	subTypePost = "post"
)

type command struct {
	*cobra.Command

	Config *cli.Configs

	branch     string
	headBranch string
	baseBranch string
}

// New is a constructor of a run sub-command
func New(c *cli.Configs) cli.Command {
	cmd := &command{Config: c}
	cmd.Command = &cobra.Command{
		Use:   "run",
		Short: "Triggers jobs of an IntegrationConfig",
	}

	preCommand := &cobra.Command{
		Use:   "pre [IntegrationConfig]",
		Short: "Triggers preSubmit jobs of an IntegrationConfig",
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			if cmd.branch != "" {
				log.Fatal("branch option cannot be used for pre")
			}
			if cmd.headBranch == "" {
				log.Fatal("head-branch option should be set for pre")
			}
			cmd.RunCommand(args, subTypePost)
		},
	}
	preCommand.Flags().StringVar(&cmd.baseBranch, "base-branch", "", "Base branch for the PullRequest event")
	preCommand.Flags().StringVar(&cmd.headBranch, "head-branch", "", "Head branch for the PullRequest event")
	cmd.Command.AddCommand(preCommand)

	postCommand := &cobra.Command{
		Use:   "post [IntegrationConfig]",
		Short: "Triggers postSubmit jobs of an IntegrationConfig",
		Args:  cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			if cmd.headBranch != "" || cmd.baseBranch != "" {
				log.Fatal("head-branch and base-branch options cannot be used for post")
			}
			cmd.RunCommand(args, subTypePre)
		},
	}
	postCommand.Flags().StringVar(&cmd.branch, "branch", "", "Branch for the Push event")
	cmd.Command.AddCommand(postCommand)

	return cmd
}

func (command *command) AddToCommand(cmd *cobra.Command) {
	cmd.AddCommand(command.Command)
}

func (command *command) RunCommand(args []string, subType string) {
	ic := args[0]

	var subResource string
	var obj interface{}

	switch subType {
	case subTypePre:
		subResource = cicdv1.IntegrationConfigAPIRunPre
		obj = cicdv1.IntegrationConfigAPIReqRunPreBody{
			BaseBranch: command.baseBranch,
			HeadBranch: command.headBranch,
		}
	case subTypePost:
		subResource = cicdv1.IntegrationConfigAPIRunPost
		obj = cicdv1.IntegrationConfigAPIReqRunPostBody{
			Branch: command.branch,
		}
	}

	body, err := json.Marshal(obj)
	if err != nil {
		log.Fatal(err)
	}

	// Run!
	client, ns, err := cli.GetClient(command.Config)
	if err != nil {
		log.Fatal(err)
	}

	cli.ExecAndHandleError(client.Post().
		Resource(cicdv1.IntegrationConfigKind).
		Namespace(ns).
		Name(ic).
		SubResource(subResource).
		Body(body))

	fmt.Printf("Triggered %s jobs for IntegrationConfig %s/%s\n", subType, ns, ic)
}
