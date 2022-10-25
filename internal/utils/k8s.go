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

package utils

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const nsEnv = "NAMESPACE"
const nsFilePathDefault = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

var nsFilePath = nsFilePathDefault

// Namespace can retrieve current namespace
func Namespace() string {
	if nsBytes, err := ioutil.ReadFile(nsFilePath); err == nil {
		return string(nsBytes)
	}

	// Fallback to env, default values
	// Not running in k8s cluster (maybe running locally)
	ns := os.Getenv(nsEnv)
	if ns == "" {
		ns = "cicd-system"
	}
	return ns
}

// CreateOrPatchObject patches the object if it exists, and creates one if not
func CreateOrPatchObject(obj, original, parent client.Object, cli client.Client, scheme *runtime.Scheme) error {
	if obj == nil {
		return fmt.Errorf("obj cannot be nil")
	}

	if parent != nil {
		if err := controllerutil.SetControllerReference(parent, obj, scheme); err != nil {
			return err
		}
	}

	// Update if resourceVersion exists, but create if not
	if obj.GetResourceVersion() != "" {
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
