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
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func TestConfigs_AddFlags(t *testing.T) {
	arguments := []string{
		"--kubeconfig", "test-dir/config",
		"--context", "test-ctx",
		"--certificate-authority", "test-ca-dir.crt",
		"--insecure-skip-tls-verify", "true",
		"--server", "https://test-cluster",
		"--username", "test-username",
		"--password", "test-pw",
		"--token", "test-bearer-token",
		"--client-certificate", "test-client.crt",
		"--client-key", "test-client.key",
		"--cluster", "test-cluster",
		"--namespace", "test-ns",
		"--user", "test-user",
	}
	expectedConfig := Configs{
		KubeConfig:   "test-dir/config",
		Context:      "test-ctx",
		CAFile:       "test-ca-dir.crt",
		Insecure:     true,
		APIServer:    "https://test-cluster",
		Username:     "test-username",
		Password:     "test-pw",
		BearerToken:  "test-bearer-token",
		CertFile:     "test-client.crt",
		KeyFile:      "test-client.key",
		ClusterName:  "test-cluster",
		Namespace:    "test-ns",
		AuthInfoName: "test-user",
	}

	flagSet := pflag.NewFlagSet("test-flags", pflag.ContinueOnError)
	flagSet.AddFlag(&pflag.Flag{})

	cfg := &Configs{}
	cfg.AddFlags(flagSet)

	require.NoError(t, flagSet.Parse(arguments))
	require.Equal(t, expectedConfig, *cfg)
}
