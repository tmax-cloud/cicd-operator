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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApprovalStatus_GetDecisionTimeInZone(t *testing.T) {
	ttime, _ := time.Parse(time.RFC3339, "2021-08-27T07:20:50.52Z")
	testTime := metav1.NewTime(ttime)

	tc := map[string]struct {
		as   *ApprovalStatus
		zone string

		expectedTime string
		errorOccurs  bool
		errorMessage string
	}{
		"noDecisionTime": {
			as:           &ApprovalStatus{},
			errorOccurs:  true,
			errorMessage: "decision time is nil",
		},
		"Asia/Seoul": {
			as:           &ApprovalStatus{DecisionTime: &testTime},
			zone:         "Asia/Seoul",
			expectedTime: "2021-08-27 16:20:50.52 +0900 KST",
		},
		"America/New_York": {
			as:           &ApprovalStatus{DecisionTime: &testTime},
			zone:         "America/New_York",
			expectedTime: "2021-08-27 03:20:50.52 -0400 EDT",
		},
		"errorZone": {
			as:           &ApprovalStatus{DecisionTime: &testTime},
			zone:         "errZone",
			errorOccurs:  true,
			errorMessage: "unknown time zone errZone",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			result, err := c.as.GetDecisionTimeInZone(c.zone)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expectedTime, result.String())
			}
		})
	}
}
