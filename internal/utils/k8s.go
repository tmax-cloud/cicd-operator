package utils

import (
	"context"
	"fmt"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	authorization "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Namespace can retrieve current namespace
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

// AuthClient is a K8s client for Authorization
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

// Client is a k8s client
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

// CreateOrPatchObject patches the object if it exists, and creates one if not
func CreateOrPatchObject(obj, original, parent runtime.Object, cli client.Client, scheme *runtime.Scheme) error {
	if obj == nil {
		return fmt.Errorf("obj cannot be nil")
	}

	objMeta, objOk := obj.(metav1.Object)
	if !objOk {
		return fmt.Errorf("obj cannot be casted to metav1.Object")
	}

	if parent != nil {
		parentMeta, parentOk := parent.(metav1.Object)
		if !parentOk {
			return fmt.Errorf("parent cannot be casted to metav1.Object")
		}

		if err := controllerutil.SetControllerReference(parentMeta, objMeta, scheme); err != nil {
			return err
		}
	}

	// Update if resourceVersion exists, but create if not
	if objMeta.GetResourceVersion() != "" {
		if original == nil {
			return fmt.Errorf("original object exists but not passed")
		}

		p := client.MergeFrom(original)
		if err := cli.Patch(context.Background(), obj, p); err != nil {
			return err
		}
	} else {
		if err := cli.Create(context.Background(), obj); err != nil {
			return err
		}
	}

	return nil
}
