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
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"k8s.io/client-go/rest"
)

const (
	apiGroup   = "cicdapi.tmax.io"
	apiVersion = "v1"
)

// Command is a cobra wrapper for cli
type Command interface {
	AddToCommand(command *cobra.Command)
}

// GetClient loads kube config and returns a kubernetes rest client
func GetClient(cfg *Configs) (*rest.RESTClient, string, error) {
	c, ns, err := LoadKubeConfig(cfg)
	if err != nil {
		return nil, "", err
	}

	if c.BearerToken != "" {
		c.CertData = nil
		c.KeyData = nil
	}

	client, err := rest.RESTClientFor(c)
	if err != nil {
		return nil, "", err
	}

	return client, ns, err
}

// ExecAndHandleError executes the request req and handles error
func ExecAndHandleError(req *rest.Request, fn func([]byte) error) error {
	raw, err := req.DoRaw(context.Background())
	if err != nil {
		resp := &utils.ErrorResponse{}
		if err2 := json.Unmarshal(raw, resp); err2 != nil {
			return err2
		}
		return fmt.Errorf(resp.Message)
	}

	if fn != nil {
		return fn(raw)
	}
	return nil
}
