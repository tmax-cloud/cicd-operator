package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"k8s.io/client-go/rest"
	"os"
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
func ExecAndHandleError(req *rest.Request, fn func([]byte)) {
	raw, err := req.DoRaw(context.Background())
	if err != nil {
		resp := &utils.ErrorResponse{}

		if err2 := json.Unmarshal(raw, resp); err2 != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Println(resp.Message)
		os.Exit(1)
	}

	if fn != nil {
		fn(raw)
	}
}
