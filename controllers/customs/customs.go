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

package customs

import (
	"bytes"
	"context"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"html/template"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const (
	varIntegrationJobName = "$INTEGRATION_JOB_NAME"
	varJobName            = "$JOB_NAME"
)

func compileString(namespace, ijName, job, content string, client client.Client) (string, error) {
	// Get IntegrationJob
	ij := &cicdv1.IntegrationJob{}
	if err := client.Get(context.Background(), types.NamespacedName{Name: ijName, Namespace: namespace}, ij); err != nil {
		return "", err
	}

	content = strings.ReplaceAll(content, varIntegrationJobName, ijName)
	content = strings.ReplaceAll(content, varJobName, job)

	// Compile template
	compiledContent := &bytes.Buffer{}
	contentTemplate := template.New("contentTemplate")
	contentTemplate, err := contentTemplate.Parse(content)
	if err != nil {
		return "", err
	}
	if err := contentTemplate.Execute(compiledContent, ij); err != nil {
		return "", err
	}

	return compiledContent.String(), nil
}
