package cli

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"os"
	"os/user"
	"path"
)

// LoadKubeConfig parses kube config files
func LoadKubeConfig(cfg *Configs) (*rest.Config, string, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = cfg.KubeConfig
	if _, ok := os.LookupEnv("HOME"); !ok {
		u, err := user.Current()
		if err != nil {
			return nil, "", fmt.Errorf("could not get current user: %v", err)
		}
		loadingRules.Precedence = append(loadingRules.Precedence, path.Join(u.HomeDir, clientcmd.RecommendedHomeDir, clientcmd.RecommendedFileName))
	}

	// Override
	override := &clientcmd.ConfigOverrides{
		AuthInfo: clientcmdapi.AuthInfo{
			ClientCertificate: cfg.CertFile,
			ClientKey:         cfg.KeyFile,
			Token:             cfg.BearerToken,
			Username:          cfg.Username,
			Password:          cfg.Password,
		},
		CurrentContext: cfg.Context,
		Context: clientcmdapi.Context{
			Cluster:   cfg.ClusterName,
			AuthInfo:  cfg.AuthInfoName,
			Namespace: cfg.Namespace,
		},
		ClusterInfo: clientcmdapi.Cluster{
			Server:                cfg.APIServer,
			InsecureSkipTLSVerify: cfg.Insecure,
			CertificateAuthority:  cfg.CAFile,
		},
	}

	rawCfg := clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, override, os.Stdin)
	cCfg, err := rawCfg.ClientConfig()
	if err != nil {
		return nil, "", err
	}

	ns, _, err := rawCfg.Namespace()
	if err != nil {
		return nil, "", err
	}

	cCfg.APIPath = "/apis"
	cCfg.GroupVersion = &schema.GroupVersion{
		Group:   apiGroup,
		Version: apiVersion,
	}
	cCfg.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	return cCfg, ns, nil
}
