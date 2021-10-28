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
	"os"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// LoadKubeConfig parses kube config files
func LoadKubeConfig(cfg *Configs) (*rest.Config, string, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = cfg.KubeConfig

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
