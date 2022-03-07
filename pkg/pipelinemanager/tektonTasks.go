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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"gopkg.in/yaml.v2"
)

const (
	catalogURL = "https://raw.githubusercontent.com/tektoncd/catalog/master/task/%s/%s/%s.yaml"
)

func generateTektonTaskRunTask(j *cicdv1.Job, target *tektonv1beta1.PipelineTask, token string) ([]tektonv1beta1.TaskResourceBinding, error) {
	taskSpec := j.TektonTask
	// Ref local or catalog
	if taskSpec.TaskRef.Local != nil {
		target.TaskRef = taskSpec.TaskRef.Local
	} else if taskSpec.TaskRef.Catalog != "" {
		var catName, catVer, catUrl string
		if catTok := strings.Split(taskSpec.TaskRef.Catalog, "@"); len(catTok) == 2 {
			if catTok[0] == "private" {
				catUrl = "https://" + token + "@" + catTok[1][8:]
			} else if catTok[0] == "public" {
				catUrl = catTok[1]
			} else {
				catName = catTok[0]
				catVer = catTok[1]
			}
		} else {
			return nil, fmt.Errorf("catalog reference should either be in form of [name]@[version] or full url path for custom catalog")
		}
		// Fetch from catalog
		spec, err := fetchCatalog(catName, catVer, catUrl)
		if err != nil {
			return nil, err
		}
		target.TaskSpec = &tektonv1beta1.EmbeddedTask{TaskSpec: *spec}
	} else {
		return nil, fmt.Errorf("both local task and catalog are nil")
	}

	target.Params = append(target.Params, cicdv1.ConvertToTektonParams(taskSpec.Params)...)
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

func globalResourceName(jobName, resName string) string {
	return fmt.Sprintf("%s-%s", jobName, resName)
}

func fetchCatalog(catName, catVer, catUrl string) (*tektonv1beta1.TaskSpec, error) {
	var resp *http.Response
	var err error
	// Fetch
	if catUrl != "" {
		resp, err = http.Get(catUrl)
	} else {
		resp, err = http.Get(fmt.Sprintf(catalogURL, catName, catVer, catName))
	}
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
