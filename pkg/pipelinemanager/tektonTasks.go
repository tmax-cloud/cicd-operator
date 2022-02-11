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

package pipelinemanager

import (
	"context"
	"encoding/json"
	"fmt"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"strings"
)

const (
	catalogURL = "https://raw.githubusercontent.com/tektoncd/catalog/master/task/%s/%s/%s.yaml"
)

func (p *pipelineManager) generateTektonTaskRunTask(j *cicdv1.Job, namespace string, target *tektonv1beta1.PipelineTask, volumes []corev1.Volume, volumeMounts []corev1.VolumeMount) ([]tektonv1beta1.TaskResourceBinding, error) {
	taskSpec := j.TektonTask
	var spec *tektonv1beta1.TaskSpec
	var err error
	// Ref local or catalog
	if taskSpec.TaskRef.Local != nil {
		// target.TaskRef = taskSpec.TaskRef.Local
		taskRef := taskSpec.TaskRef.Local

		spec, err = p.fetchLocalTask(taskRef, namespace)
		if err != nil {
			return nil, err
		}
	} else if taskSpec.TaskRef.Catalog != "" {
		catTok := strings.Split(taskSpec.TaskRef.Catalog, "@")
		if len(catTok) != 2 {
			return nil, fmt.Errorf("catalog reference should be in form of [name]@[version]")
		}
		catName := catTok[0]
		catVer := catTok[1]

		// Fetch from catalog
		spec, err = fetchCatalog(catName, catVer)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("both local task and catalog are nil")
	}
	// Manual VolumeMounts
	spec.Volumes = append(spec.Volumes, volumes...)
	for i := range spec.Steps {
		spec.Steps[i].VolumeMounts = append(spec.Steps[i].VolumeMounts, volumeMounts...)
	}
	target.TaskSpec = &tektonv1beta1.EmbeddedTask{TaskSpec: *spec}
	// Params
	for _, p := range taskSpec.Params {
		v := tektonv1beta1.ArrayOrString{}

		if p.ArrayVal != nil {
			v.Type = tektonv1beta1.ParamTypeArray
			v.ArrayVal = append(v.ArrayVal, p.ArrayVal...)
		} else {
			v.Type = tektonv1beta1.ParamTypeString
			v.StringVal = p.StringVal
		}

		target.Params = append(target.Params, tektonv1beta1.Param{
			Name:  p.Name,
			Value: v,
		})
	}

	// Resources
	var resources []tektonv1beta1.TaskResourceBinding
	if taskSpec.Resources != nil {
		target.Resources = &tektonv1beta1.PipelineTaskResources{}
		// Input
		for _, res := range taskSpec.Resources.Inputs {
			globalName := globalResourceName(j.Name, res.Name)
			target.Resources.Inputs = append(target.Resources.Inputs, tektonv1beta1.PipelineTaskInputResource{
				Name:     res.Name,
				Resource: globalName,
			})

			resSpec := res.DeepCopy()
			resSpec.Name = globalName

			// Append to resources
			resources = append(resources, *resSpec)
		}

		// Output
		for _, res := range taskSpec.Resources.Outputs {
			globalName := globalResourceName(j.Name, res.Name)
			target.Resources.Outputs = append(target.Resources.Outputs, tektonv1beta1.PipelineTaskOutputResource{
				Name:     res.Name,
				Resource: globalName,
			})

			resSpec := res.DeepCopy()
			resSpec.Name = globalName

			// Append to resources
			resources = append(resources, *resSpec)
		}
	}

	// Workspaces
	target.Workspaces = append(target.Workspaces, taskSpec.Workspaces...)

	return resources, nil
}

func (p *pipelineManager) fetchLocalTask(taskRef *tektonv1beta1.TaskRef, namespace string) (*tektonv1beta1.TaskSpec, error) {
	var spec *tektonv1beta1.TaskSpec
	if taskRef.Kind == tektonv1beta1.NamespacedTaskKind {
		task := &tektonv1beta1.Task{}
		if err := p.Client.Get(context.Background(), types.NamespacedName{Name: taskRef.Name, Namespace: namespace}, task); err != nil {
			return nil, err
		}
		spec = &task.Spec
	} else if taskRef.Kind == tektonv1beta1.ClusterTaskKind {
		clusterTask := &tektonv1beta1.ClusterTask{}
		if err := p.Client.Get(context.Background(), types.NamespacedName{Name: taskRef.Name, Namespace: ""}, clusterTask); err != nil {
			return nil, err
		}
		spec = &clusterTask.Spec
	} else {
		return nil, fmt.Errorf("task kind should be task or clustertask")
	}
	return spec, nil
}

func globalResourceName(jobName, resName string) string {
	return fmt.Sprintf("%s-%s", jobName, resName)
}

func fetchCatalog(catName, catVer string) (*tektonv1beta1.TaskSpec, error) {
	// Fetch
	resp, err := http.Get(fmt.Sprintf(catalogURL, catName, catVer, catName))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("error: %d, msg: %s", resp.StatusCode, string(respBody))
	}

	// Unmarshal
	var body interface{}
	if err := yaml.Unmarshal(respBody, &body); err != nil {
		return nil, err
	}
	body = convert(body)
	jsonString, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	task := &tektonv1beta1.Task{}
	if err := json.Unmarshal(jsonString, task); err != nil {
		return nil, err
	}

	return &task.Spec, nil
}

func convert(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convert(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convert(v)
		}
	}
	return i
}
