package utils

import (
	"fmt"
	"io/ioutil"
	"os"
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
