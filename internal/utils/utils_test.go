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
	"github.com/stretchr/testify/require"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"testing"
)

func TestGetGitCli(t *testing.T) {
	tc := map[string]struct {
		ic *cicdv1.IntegrationConfig

		errorOccurs  bool
		errorMessage string
	}{
		"github": {
			ic: &cicdv1.IntegrationConfig{
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type: cicdv1.GitTypeGitHub,
					},
				},
			},
		},
		"gitlab": {
			ic: &cicdv1.IntegrationConfig{
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type: cicdv1.GitTypeGitLab,
					},
				},
			},
		},
		"fake": {
			ic: &cicdv1.IntegrationConfig{
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type: cicdv1.GitTypeFake,
					},
				},
			},
		},
		"wrongType": {
			ic: &cicdv1.IntegrationConfig{
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type: "wrongType",
					},
				},
			},
			errorOccurs:  true,
			errorMessage: "git type wrongType is not supported",
		},
		"initErr": {
			ic: &cicdv1.IntegrationConfig{
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type: cicdv1.GitTypeFake,
						Token: &cicdv1.GitToken{
							ValueFrom: &cicdv1.GitTokenFrom{
								SecretKeyRef: corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{Name: "test-not-exist"},
									Key:                  "token",
								},
							},
						},
					},
				},
			},
			errorOccurs:  true,
			errorMessage: "secrets \"test-not-exist\" not found",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			fakeCli := fake.NewClientBuilder().Build()
			_, err := GetGitCli(c.ic, fakeCli)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseApproversList(t *testing.T) {
	// Success test
	str := `admin@tmax.co.kr=admin@tmax.co.kr,test@tmax.co.kr
test2@tmax.co.kr=test2@tmax.co.kr`
	list, err := ParseApproversList(str)
	require.NoError(t, err)
	require.Len(t, list, 3, "list is not parsed well")
	require.Equal(t, "admin@tmax.co.kr=admin@tmax.co.kr", list[0], "list is not parsed well")
	require.Equal(t, "test@tmax.co.kr", list[1], "list is not parsed well")
	require.Equal(t, "test2@tmax.co.kr=test2@tmax.co.kr", list[2], "list is not parsed well")

	// Fail test
	str = "admin,,ttt"
	list, err = ParseApproversList(str)
	if err == nil {
		for i, l := range list {
			t.Logf("%d : %s", i, l)
		}
		t.Fatal("error not occur")
	}
}

func TestParseEmailFromUsers(t *testing.T) {
	// Include test
	users := []cicdv1.ApprovalUser{
		{Name: "aweilfjlwesfj"},
		{Name: "aweilfjlwesfj", Email: "aweiojweio"},
		{Name: "aweilfjlwesfj", Email: "admin@tmax.co.kr"},
		{Name: "asdij@oisdjf.sdfioj", Email: "test@tmax.co.kr"},
	}

	tos := ParseEmailFromUsers(users)

	require.Len(t, tos, 2, "list is not parsed well")
	require.Equal(t, "admin@tmax.co.kr", tos[0], "list is not parsed well")
	require.Equal(t, "test@tmax.co.kr", tos[1], "list is not parsed well")
}
