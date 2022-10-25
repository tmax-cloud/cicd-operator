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

package main

import (
	"flag"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/tmax-cloud/cicd-operator/cmd/cicdctl/approve"
	"github.com/tmax-cloud/cicd-operator/cmd/cicdctl/run"
	"github.com/tmax-cloud/cicd-operator/cmd/cicdctl/webhook"
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
	webhook.New(configs).AddToCommand(cmd)

	// Set klog verbosity
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlag(flag.CommandLine.Lookup("v"))

	_ = cmd.Execute()
}
