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

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestRegisterControllerConfigUpdateChan(t *testing.T) {
	controllerConfigUpdateChan = nil
	ch := make(chan struct{})
	RegisterControllerConfigUpdateChan(ch)
	require.Len(t, controllerConfigUpdateChan, 1)
}

func TestApplyControllerConfigChange(t *testing.T) {
	tc := map[string]controllerTestCase{
		"default": {ConfigMap: &corev1.ConfigMap{
			Data: map[string]string{},
		}, AssertFunc: func(t *testing.T, err error) {
			require.NoError(t, err)

			require.Equal(t, 5, MaxPipelineRun)
			require.False(t, EnableMail)
			require.Equal(t, "", ExternalHostName)
			require.Equal(t, "", ReportRedirectURITemplate)
			require.Equal(t, "", SMTPHost)
			require.Equal(t, "", SMTPUserSecret)
			require.Equal(t, 120, CollectPeriod)
			require.Equal(t, 120, IntegrationJobTTL)
			require.Equal(t, "", IngressClass)
			require.Equal(t, "", IngressHost)
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
			require.NoError(t, err)

			require.Equal(t, 2, MaxPipelineRun)
			require.True(t, EnableMail)
			require.Equal(t, "external.host.name", ExternalHostName)
			require.Equal(t, "https://asd/test", ReportRedirectURITemplate)
			require.Equal(t, "smtp.test.test", SMTPHost)
			require.Equal(t, "smtp-test", SMTPUserSecret)
			require.Equal(t, 11, CollectPeriod)
			require.Equal(t, 11, IntegrationJobTTL)
			require.Equal(t, "test-cls", IngressClass)
			require.Equal(t, "test.host", IngressHost)
		}},
		"errorOccur": {ConfigMap: &corev1.ConfigMap{
			Data: map[string]string{
				"enableMail":     "true",
				"smtpHost":       "",
				"smtpUserSecret": "",
			},
		}, AssertFunc: func(t *testing.T, err error) {
			require.Error(t, err)
			require.Equal(t, "email is enaled but smtp access info. is not given", err.Error())

			require.True(t, EnableMail)
			require.Equal(t, "", SMTPHost)
			require.Equal(t, "", SMTPUserSecret)
		}},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
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

			ch := make(chan struct{}, 1)
			controllerConfigUpdateChan = append(controllerConfigUpdateChan, ch)
			go func() {
				<-ch
			}()

			err := ApplyControllerConfigChange(c.ConfigMap)
			c.AssertFunc(t, err)
		})
	}
}
