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

package run

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

func Test_command_runPre(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/apis/cicdapi.tmax.io/v1/namespaces/default/integrationconfigs/test/runpre", func(w http.ResponseWriter, req *http.Request) {})
	srv := httptest.NewServer(router)

	tc := map[string]struct {
		cmd *command

		errorOccurs  bool
		errorMessage string
	}{
		"normal": {
			cmd: &command{
				Config: &cli.Configs{
					APIServer: srv.URL,
					Namespace: "default",
					Insecure:  true,
				},
				baseBranch: "master",
				headBranch: "feat/test",
			},
		},
		"haveBranch": {
			cmd: &command{
				Config: &cli.Configs{
					APIServer: srv.URL,
					Namespace: "default",
					Insecure:  true,
				},
				branch: "master",
			},
			errorOccurs:  true,
			errorMessage: "branch option cannot be used for pre",
		},
		"noHeadBranch": {
			cmd: &command{
				Config: &cli.Configs{
					APIServer: srv.URL,
					Namespace: "default",
					Insecure:  true,
				},
			},
			errorOccurs:  true,
			errorMessage: "head-branch option should be set for pre",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			err := c.cmd.runPre(nil, []string{"test"})
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_command_runPost(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/apis/cicdapi.tmax.io/v1/namespaces/default/integrationconfigs/test/runpost", func(w http.ResponseWriter, req *http.Request) {})
	srv := httptest.NewServer(router)

	tc := map[string]struct {
		cmd *command

		errorOccurs  bool
		errorMessage string
	}{
		"normal": {
			cmd: &command{
				Config: &cli.Configs{
					APIServer: srv.URL,
					Namespace: "default",
					Insecure:  true,
				},
			},
		},
		"haveBaseBranch": {
			cmd: &command{
				Config: &cli.Configs{
					APIServer: srv.URL,
					Namespace: "default",
					Insecure:  true,
				},
				baseBranch: "master",
			},
			errorOccurs:  true,
			errorMessage: "head-branch and base-branch options cannot be used for post",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			err := c.cmd.runPost(nil, []string{"test"})
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
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
	router.HandleFunc("/apis/cicdapi.tmax.io/v1/namespaces/default/integrationconfigs/test/runpre", func(w http.ResponseWriter, req *http.Request) {})
	router.HandleFunc("/apis/cicdapi.tmax.io/v1/namespaces/default/integrationconfigs/test/runpost", func(w http.ResponseWriter, req *http.Request) {})
	srv := httptest.NewServer(router)

	tc := map[string]struct {
		cfg       cli.Configs
		arguments []string
		subType   string

		errorOccurs  bool
		errorMessage string
	}{
		"pre": {
			cfg: cli.Configs{
				APIServer: srv.URL,
				Namespace: "default",
				Insecure:  true,
			},
			arguments: []string{"test"},
			subType:   "pre",
		},
		"post": {
			cfg: cli.Configs{
				APIServer: srv.URL,
				Namespace: "default",
				Insecure:  true,
			},
			arguments: []string{"test"},
			subType:   "post",
		},
		"cliErr": {
			cfg: cli.Configs{
				APIServer: "tcp:/~/test.com",
				Namespace: "default",
				Insecure:  true,
			},
			arguments:    []string{"test"},
			subType:      "pre",
			errorOccurs:  true,
			errorMessage: "host must be a URL or a host:port pair: \"tcp:///~/test.com\"",
		},
		"execErr": {
			cfg: cli.Configs{
				APIServer: srv.URL,
				Namespace: "default",
				Insecure:  true,
			},
			arguments:    []string{"test2"},
			subType:      "pre",
			errorOccurs:  true,
			errorMessage: "invalid character 'p' after top-level value",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			cmd := &command{Config: &c.cfg}
			err := cmd.RunCommand(c.arguments, c.subType)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
