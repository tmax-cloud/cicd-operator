package utils

import (
	"fmt"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	authorization "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func Namespace() (string, error) {
	nsPath := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	if FileExists(nsPath) {
		// Running in k8s cluster
		nsBytes, err := ioutil.ReadFile(nsPath)
		if err != nil {
			return "", fmt.Errorf("could not read file %s", nsPath)
		}
		return string(nsBytes), nil
	} else {
		// Not running in k8s cluster (may be running locally)
		ns := os.Getenv("NAMESPACE")
		if ns == "" {
			ns = "cicd-system"
		}
		return ns, nil
	}
}

func AuthClient() (*authorization.AuthorizationV1Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	c, err := authorization.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func Client(scheme *runtime.Scheme) (client.Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}
	return c, nil
}
