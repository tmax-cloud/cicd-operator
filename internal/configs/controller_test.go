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

package configs

import (
	"testing"

	"github.com/bmizerany/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestApplyControllerConfigChange(t *testing.T) {
	tc := map[string]controllerTestCase{
		"default": {ConfigMap: &corev1.ConfigMap{
			Data: map[string]string{},
		}, AssertFunc: func(t *testing.T, err error) {
			assert.Equal(t, true, err == nil)

			assert.Equal(t, 5, MaxPipelineRun)
			assert.Equal(t, false, EnableMail)
			assert.Equal(t, "", ExternalHostName)
			assert.Equal(t, "", ReportRedirectURITemplate)
			assert.Equal(t, "", SMTPHost)
			assert.Equal(t, "", SMTPUserSecret)
			assert.Equal(t, 120, CollectPeriod)
			assert.Equal(t, 120, IntegrationJobTTL)
			assert.Equal(t, "", IngressClass)
			assert.Equal(t, "", IngressHost)
		}},
		"noError": {ConfigMap: &corev1.ConfigMap{
			Data: map[string]string{
				"maxPipelineRun":            "2",
				"enableMail":                "true",
				"externalHostName":          "external.host.name",
				"reportRedirectUriTemplate": "https://asd/test",
				"smtpHost":                  "smtp.test.test",
				"smtpUserSecret":            "smtp-test",
				"collectPeriod":             "11",
				"integrationJobTTL":         "11",
				"ingressClass":              "test-cls",
				"ingressHost":               "test.host",
			},
		}, AssertFunc: func(t *testing.T, err error) {
			assert.Equal(t, true, err == nil)

			assert.Equal(t, 2, MaxPipelineRun)
			assert.Equal(t, true, EnableMail)
			assert.Equal(t, "external.host.name", ExternalHostName)
			assert.Equal(t, "https://asd/test", ReportRedirectURITemplate)
			assert.Equal(t, "smtp.test.test", SMTPHost)
			assert.Equal(t, "smtp-test", SMTPUserSecret)
			assert.Equal(t, 11, CollectPeriod)
			assert.Equal(t, 11, IntegrationJobTTL)
			assert.Equal(t, "test-cls", IngressClass)
			assert.Equal(t, "test.host", IngressHost)
		}},
		"errorOccur": {ConfigMap: &corev1.ConfigMap{
			Data: map[string]string{
				"enableMail":     "true",
				"smtpHost":       "",
				"smtpUserSecret": "",
			},
		}, AssertFunc: func(t *testing.T, err error) {
			assert.Equal(t, true, err != nil)
			assert.Equal(t, "email is enaled but smtp access info. is not given", err.Error())

			assert.Equal(t, true, EnableMail)
			assert.Equal(t, "", SMTPHost)
			assert.Equal(t, "", SMTPUserSecret)
		}},
	}

	for name, c := range tc {
		MaxPipelineRun = 0
		EnableMail = false
		ExternalHostName = ""
		ReportRedirectURITemplate = ""
		SMTPHost = ""
		SMTPUserSecret = ""
		CollectPeriod = 0
		IntegrationJobTTL = 0
		IngressClass = ""
		IngressHost = ""
		t.Run(name, func(t *testing.T) {
			err := ApplyControllerConfigChange(c.ConfigMap)
			c.AssertFunc(t, err)
		})
	}
}
