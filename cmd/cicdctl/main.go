package main

import (
	"flag"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/tmax-cloud/cicd-operator/cmd/cicdctl/approve"
	"github.com/tmax-cloud/cicd-operator/cmd/cicdctl/run"
	"github.com/tmax-cloud/cicd-operator/pkg/cli"
	"k8s.io/klog"
)

func main() {
	cmd := &cobra.Command{
		Use:   "cicdctl [Command]",
		Short: "cicdctl runs CI/CD operator-related tasks",
	}

	configs := &cli.Configs{}
	configs.AddFlags(cmd.PersistentFlags())

	approve.New(configs).AddToCommand(cmd)
	run.New(configs).AddToCommand(cmd)

	// Set klog verbosity
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlag(flag.CommandLine.Lookup("v"))

	_ = cmd.Execute()
}
