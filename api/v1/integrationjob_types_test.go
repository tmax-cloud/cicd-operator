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

package v1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIntegrationJobStatus_SetDefaults(t *testing.T) {
	tc := map[string]struct {
		currentState  IntegrationJobState
		expectedState IntegrationJobState
	}{
		"default": {
			expectedState: IntegrationJobStatePending,
		},
		"alreadySet": {
			currentState:  IntegrationJobStateCompleted,
			expectedState: IntegrationJobStateCompleted,
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			ijStatus := &IntegrationJobStatus{State: c.currentState}
			ijStatus.SetDefaults()
			require.Equal(t, c.expectedState, ijStatus.State)
		})
	}
}

func TestIntegrationJob_GetReportServerAddress(t *testing.T) {
	configs.CurrentExternalHostName = "test.host.com"
	ij := &IntegrationJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ij",
			Namespace: "test-ns",
		},
	}
	require.Equal(t, "http://test.host.com/report/test-ns/test-ij/test-job", ij.GetReportServerAddress("test-job"))
}

func TestIntegrationJob_IsCompleted(t *testing.T) {
	tc := map[string]struct {
		completionTime *metav1.Time
		expectedResult bool
	}{
		"completed": {
			completionTime: &metav1.Time{Time: time.Now()},
			expectedResult: true,
		},
		"uncompleted": {
			completionTime: nil,
			expectedResult: false,
		},
	}
	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			ij := &IntegrationJob{
				Status: IntegrationJobStatus{
					CompletionTime: c.completionTime,
				},
			}
			completed := ij.IsCompleted()
			require.Equal(t, c.expectedResult, completed)
		})
	}
}
