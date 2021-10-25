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

func Test_command_AddToCommand(t *testing.T) {
	cmd := &command{}
	cmd.Command = &cobra.Command{}

	cob := &cobra.Command{}
	cmd.AddToCommand(cob)
	require.Len(t, cob.Commands(), 1)
}

func Test_command_RunCommand(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/apis/cicdapi.tmax.io/v1/namespaces/default/integrationconfigs/test/webhookurl", func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(`{"url": "test-url", "secret": "test-secret"}`))
	})
	srv := httptest.NewServer(router)

	tc := map[string]struct {
		cfg       cli.Configs
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
			arguments: []string{"test"},
		},
		"cliErr": {
			cfg: cli.Configs{
				APIServer: "tcp:/~/test.com",
				Namespace: "default",
				Insecure:  true,
			},
			arguments:    []string{"test"},
			errorOccurs:  true,
			errorMessage: "host must be a URL or a host:port pair: \"tcp:///~/test.com\"",
		},
		"execErr": {
			cfg: cli.Configs{
				APIServer: srv.URL,
				Namespace: "default",
				Insecure:  true,
			},
			arguments:    []string{"test222"},
			errorOccurs:  true,
			errorMessage: "invalid character 'p' after top-level value",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			cmd := &command{Config: &c.cfg}
			cob := &cobra.Command{Use: "webhook"}
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

func Test_printWebhookInfo(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		require.NoError(t, printWebhookInfo([]byte(`{"url": "test-url", "secret": "test-secret"}`)))
	})

	t.Run("unmarshalErr", func(t *testing.T) {
		require.Error(t, printWebhookInfo([]byte("aaa")))
	})
}
