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
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"k8s.io/apimachinery/pkg/api/resource"
)

func Test_gitCheckout(t *testing.T) {
	tc := map[string]struct {
		cpuReq string
		memReq string

		expectedCpu resource.Quantity
		expectedMem resource.Quantity
	}{
		"normal": {
			cpuReq:      "200m",
			memReq:      "1Gi",
			expectedCpu: resource.MustParse("200m"),
			expectedMem: resource.MustParse("1Gi"),
		},
		"default": {
			expectedCpu: resource.MustParse("30m"),
			expectedMem: resource.MustParse("100Mi"),
		},
		"cpuErr": {
			cpuReq:      "200mmmmm",
			memReq:      "1Gi",
			expectedCpu: resource.MustParse("30m"),
			expectedMem: resource.MustParse("1Gi"),
		},
		"memErr": {
			cpuReq:      "200m",
			memReq:      "1GGGi",
			expectedCpu: resource.MustParse("200m"),
			expectedMem: resource.MustParse("100Mi"),
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			configs.GitCheckoutStepCPURequest = c.cpuReq
			configs.GitCheckoutStepMemRequest = c.memReq

			step := gitCheckout()
			require.Equal(t, c.expectedCpu, *step.Resources.Limits.Cpu())
			require.Equal(t, c.expectedCpu, *step.Resources.Requests.Cpu())
			require.Equal(t, c.expectedMem, *step.Resources.Limits.Memory())
			require.Equal(t, c.expectedMem, *step.Resources.Requests.Memory())
		})
	}
}
