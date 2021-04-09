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
