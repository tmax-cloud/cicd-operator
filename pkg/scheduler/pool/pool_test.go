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

package pool

import (
	"fmt"
	"testing"
	"time"

	"github.com/bmizerany/assert"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/structs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestJobPool_SyncJob(t *testing.T) {
	ch := make(chan struct{}, 1)
	p := New(ch, testCompare)

	now := time.Now()
	testJob1 := jobForTest("1", "default", now)
	testJob2 := jobForTest("2", "default", now)
	testJob3 := jobForTest("3", "default", now)
	testJob4 := jobForTest("4", "default", now)
	testJob5 := jobForTest("5", "default", now)
	testJob6 := jobForTest("6", "default", now)
	testJob7 := jobForTest("6", "l2c-system", now)

	p.SyncJob(testJob1)
	p.SyncJob(testJob2)
	p.SyncJob(testJob3)
	p.SyncJob(testJob4)
	p.SyncJob(testJob5)
	p.SyncJob(testJob6)
	p.SyncJob(testJob7)

	// Initial
	assert.Equal(t, 7, p.pending.Len(), "state transition isn't done properly")
	assert.Equal(t, 0, p.running.Len(), "state transition isn't done properly")

	// 3 Running
	testJob3.Status.State = cicdv1.IntegrationJobStateRunning
	p.SyncJob(testJob3)
	assert.Equal(t, 6, p.pending.Len(), "state transition isn't done properly")
	assert.Equal(t, 1, p.running.Len(), "state transition isn't done properly")

	// 3 Completed
	testJob3.Status.State = cicdv1.IntegrationJobStateCompleted
	p.SyncJob(testJob3)
	assert.Equal(t, 6, p.pending.Len(), "state transition isn't done properly")
	assert.Equal(t, 0, p.running.Len(), "state transition isn't done properly")
}

func testCompare(_a, _b structs.Item) bool {
	if _a == nil || _b == nil {
		return false
	}
	a, aOk := _a.(*JobNode)
	b, bOk := _b.(*JobNode)
	if !aOk || !bOk {
		return false
	}

	return a.CreationTimestamp.Time.Before(b.CreationTimestamp.Time) || fmt.Sprintf("%s_%s", a.Namespace, a.Name) < fmt.Sprintf("%s_%s", b.Namespace, b.Name)
}

func jobForTest(name, namespace string, created time.Time) *cicdv1.IntegrationJob {
	return &cicdv1.IntegrationJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: metav1.Time{Time: created},
		},
		Spec: cicdv1.IntegrationJobSpec{
			Timeout: &metav1.Duration{
				Duration: 1,
			},
		},
		Status: cicdv1.IntegrationJobStatus{
			State: cicdv1.IntegrationJobStatePending,
		},
	}
}
