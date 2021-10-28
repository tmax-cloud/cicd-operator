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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/tmax-cloud/cicd-operator/pkg/cli"
)

func TestNew(t *testing.T) {
	cfg := &cli.Configs{
		KubeConfig: "test",
	}
	cmdIf := New(cfg)
	cmd, ok := cmdIf.(*command)
	require.True(t, ok)

	require.Equal(t, cfg, cmd.Config)
}

func Test_approvalCommand(t *testing.T) {
	cmd := approvalCommand("approve", func(_ *cobra.Command, _ []string) error { return nil })
	require.Equal(t, "approve [Approval] [reason]", cmd.Use)
	require.Equal(t, "Approves an Approval", cmd.Short)
}

func Test_command_AddToCommand(t *testing.T) {
	cmd := &command{}
	cmd.ApproveCommand = &cobra.Command{}
	cmd.RejectCommand = &cobra.Command{}
	cob := &cobra.Command{}
	cmd.AddToCommand(cob)

	require.Len(t, cob.Commands(), 2)
}

func Test_command_RunCommand(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/apis/cicdapi.tmax.io/v1/namespaces/default/approvals/test-approval/approve", func(w http.ResponseWriter, req *http.Request) {})
	srv := httptest.NewServer(router)

	tc := map[string]struct {
		cfg       cli.Configs
		command   string
		arguments []string

		errorOccurs  bool
		errorMessage string
	}{
		"normal": {
			cfg: cli.Configs{
				APIServer: srv.URL,
				Namespace: "default",
				Insecure:  true,
			},
			command:   "approve",
			arguments: []string{"test-approval", "just just"},
		},
		"cliErr": {
			cfg: cli.Configs{
				APIServer: "tcp:/~/test.com",
				Namespace: "default",
				Insecure:  true,
			},
			command:      "approve",
			arguments:    []string{"test-approval", "just just"},
			errorOccurs:  true,
			errorMessage: "host must be a URL or a host:port pair: \"tcp:///~/test.com\"",
		},
		"execErr": {
			cfg: cli.Configs{
				APIServer: srv.URL,
				Namespace: "default",
				Insecure:  true,
			},
			command:      "approve",
			arguments:    []string{"test-approval222", "just just"},
			errorOccurs:  true,
			errorMessage: "invalid character 'p' after top-level value",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			cmd := &command{Config: &c.cfg}
			cob := &cobra.Command{Use: c.command}
			err := cmd.RunCommand(cob, c.arguments)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
