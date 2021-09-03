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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestJobStatus_Equals(t *testing.T) {
	time1 := time.Now()
	time2 := time1.Add(10 * time.Hour)

	tc := map[string]struct {
		job1   *JobStatus
		job2   *JobStatus
		equals bool
	}{
		"equal": {
			job1: &JobStatus{
				State:          CommitStatusStateSuccess,
				Message:        "message1",
				StartTime:      &metav1.Time{Time: time1},
				CompletionTime: &metav1.Time{Time: time1},
			},
			job2: &JobStatus{
				State:          CommitStatusStateSuccess,
				Message:        "message1",
				StartTime:      &metav1.Time{Time: time1},
				CompletionTime: &metav1.Time{Time: time1},
			},
			equals: true,
		},
		"notEqual1": {
			job1: &JobStatus{
				State:          CommitStatusStateSuccess,
				Message:        "message2",
				StartTime:      &metav1.Time{Time: time1},
				CompletionTime: &metav1.Time{Time: time1},
			},
			job2: &JobStatus{
				State:          CommitStatusStateSuccess,
				Message:        "message1",
				StartTime:      &metav1.Time{Time: time1},
				CompletionTime: &metav1.Time{Time: time1},
			},
			equals: false,
		},
		"notEqual2": {
			job1: &JobStatus{
				State:          CommitStatusStateSuccess,
				Message:        "message1",
				StartTime:      &metav1.Time{Time: time1},
				CompletionTime: &metav1.Time{Time: time1},
			},
			job2: &JobStatus{
				State:          CommitStatusStateSuccess,
				Message:        "message1",
				StartTime:      &metav1.Time{Time: time2},
				CompletionTime: &metav1.Time{Time: time1},
			},
			equals: false,
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, c.equals, c.job1.Equals(c.job2))
		})
	}
}

func TestJobs_GetGraph(t *testing.T) {
	tc := map[string]struct {
		jobs Jobs

		errorOccurs  bool
		errorMessage string

		pres map[string][]string
	}{
		"normal": {
			jobs: Jobs{
				{Container: corev1.Container{Name: "job-1"}},
				{Container: corev1.Container{Name: "job-2"}, After: []string{"job-1"}},
				{Container: corev1.Container{Name: "job-3"}, After: []string{"job-1"}},
				{Container: corev1.Container{Name: "job-4"}, After: []string{"job-2", "job-3"}},
			},
			pres: map[string][]string{
				"job-1": nil,
				"job-2": {"job-1"},
				"job-3": {"job-1"},
				"job-4": {"job-2", "job-1", "job-3"},
			},
		},
		"cyclic": {
			jobs: Jobs{
				{Container: corev1.Container{Name: "job-1"}, After: []string{"job-2"}},
				{Container: corev1.Container{Name: "job-2"}, After: []string{"job-1"}},
			},
			errorOccurs:  true,
			errorMessage: "job graph is cyclic",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			graph, err := c.jobs.GetGraph()
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)

				for target, pres := range c.pres {
					require.Equal(t, pres, graph.GetPres(target))
				}
			}
		})
	}
}
