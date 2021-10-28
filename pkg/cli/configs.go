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

package cli

import (
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	flagKubeConfig = "kubeconfig"
)

// Configs is configs struct for accessing k8s api server
type Configs struct {
	KubeConfig string

	CAFile   string
	CertFile string
	KeyFile  string

	BearerToken  string
	Username     string
	Password     string
	AuthInfoName string
	Namespace    string

	ClusterName string
	Context     string
	APIServer   string

	Insecure bool
}

// AddFlags adds flags to a command line. Flags are sampled from kubectl's flags
func (c *Configs) AddFlags(f *pflag.FlagSet) {
	f.StringVar(&c.KubeConfig, flagKubeConfig, c.KubeConfig, "Path to the kubeconfig file to use for CLI requests.")

	flags := clientcmd.RecommendedConfigOverrideFlags("")

	f.StringVar(&c.Context, flags.CurrentContext.LongName, c.ClusterName, flags.CurrentContext.Description)

	f.StringVar(&c.CAFile, flags.ClusterOverrideFlags.CertificateAuthority.LongName, c.CAFile, flags.ClusterOverrideFlags.CertificateAuthority.Description)
	f.BoolVar(&c.Insecure, flags.ClusterOverrideFlags.InsecureSkipTLSVerify.LongName, c.Insecure, flags.ClusterOverrideFlags.InsecureSkipTLSVerify.Description)
	f.StringVarP(&c.APIServer, flags.ClusterOverrideFlags.APIServer.LongName, "s", c.APIServer, flags.ClusterOverrideFlags.APIServer.Description)

	f.StringVar(&c.Username, flags.AuthOverrideFlags.Username.LongName, c.Username, flags.AuthOverrideFlags.Username.Description)
	f.StringVar(&c.Password, flags.AuthOverrideFlags.Password.LongName, c.Password, flags.AuthOverrideFlags.Password.Description)
	f.StringVar(&c.BearerToken, flags.AuthOverrideFlags.Token.LongName, c.BearerToken, flags.AuthOverrideFlags.Token.Description)
	f.StringVar(&c.CertFile, flags.AuthOverrideFlags.ClientCertificate.LongName, c.CertFile, flags.AuthOverrideFlags.ClientCertificate.Description)
	f.StringVar(&c.KeyFile, flags.AuthOverrideFlags.ClientKey.LongName, c.KeyFile, flags.AuthOverrideFlags.ClientKey.Description)

	f.StringVar(&c.ClusterName, flags.ContextOverrideFlags.ClusterName.LongName, c.ClusterName, flags.ContextOverrideFlags.ClusterName.Description)
	f.StringVarP(&c.Namespace, flags.ContextOverrideFlags.Namespace.LongName, flags.ContextOverrideFlags.Namespace.ShortName, c.Namespace, flags.ContextOverrideFlags.Namespace.Description)
	f.StringVar(&c.AuthInfoName, flags.ContextOverrideFlags.AuthInfoName.LongName, c.AuthInfoName, flags.ContextOverrideFlags.AuthInfoName.Description)
}
