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
	"context"
	"fmt"

	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (p *pipelineManager) handleNotification(jobStatus *cicdv1.JobStatus, ij *cicdv1.IntegrationJob, cfg *cicdv1.IntegrationConfig) error {
	// Get jobSpec spec
	jobSpec := getSpecFromStatus(jobStatus, ij.Spec.ConfigRef.Type, cfg)
	if jobSpec == nil {
		return fmt.Errorf("no jobSpec %s exists in the config", jobStatus.Name)
	}

	if jobSpec.Notification == nil {
		return nil
	}

	// Get noti method
	noti := getNotificationFromSpec(jobStatus, jobSpec)
	if noti == nil {
		return nil
	}

	// Handle email
	if noti.Email != nil {
		if err := p.handleEmailNotification(ij, jobSpec, noti.Email); err != nil {
			return err
		}
	}

	// Handle slack
	if noti.Slack != nil {
		if err := p.handleSlackNotification(ij, jobSpec, noti.Slack); err != nil {
			return err
		}
	}

	return nil
}

func (p *pipelineManager) handleEmailNotification(ij *cicdv1.IntegrationJob, job *cicdv1.Job, email *cicdv1.NotiEmail) error {
	runSpec := commonNotiRun(cicdv1.CustomTaskKindEmail, ij.Name, job.Name, ij.Namespace)
	runSpec.Spec.Params = generateEmailRunParams(ij, job, email)

	if err := controllerutil.SetOwnerReference(ij, runSpec, p.Scheme); err != nil {
		return err
	}

	// Check if it exists
	run := &tektonv1alpha1.Run{}
	if err := p.Client.Get(context.Background(), types.NamespacedName{Name: runSpec.Name, Namespace: runSpec.Namespace}, run); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		// Create if not exist
		if err := p.Client.Create(context.Background(), runSpec); err != nil {
			return err
		}
	}

	return nil
}

func (p *pipelineManager) handleSlackNotification(ij *cicdv1.IntegrationJob, job *cicdv1.Job, slack *cicdv1.NotiSlack) error {
	runSpec := commonNotiRun(cicdv1.CustomTaskKindSlack, ij.Name, job.Name, ij.Namespace)
	runSpec.Spec.Params = generateSlackRunParams(ij, job, slack)

	if err := controllerutil.SetOwnerReference(ij, runSpec, p.Scheme); err != nil {
		return err
	}

	// Check if it exists
	run := &tektonv1alpha1.Run{}
	if err := p.Client.Get(context.Background(), types.NamespacedName{Name: runSpec.Name, Namespace: runSpec.Namespace}, run); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		// Create if not exist
		if err := p.Client.Create(context.Background(), runSpec); err != nil {
			return err
		}
	}

	return nil
}

func commonNotiRun(kind, ijName, jobName, ns string) *tektonv1alpha1.Run {
	t := kind
	switch t {
	case cicdv1.CustomTaskKindEmail:
		t = "email"
	case cicdv1.CustomTaskKindSlack:
		t = "slack"
	}
	return &tektonv1alpha1.Run{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s-%s", ijName, jobName, t),
			Namespace: ns,
		},
		Spec: tektonv1alpha1.RunSpec{
			Ref: generateCustomTaskRef(kind),
		},
	}
}

func getSpecFromStatus(jobStatus *cicdv1.JobStatus, t cicdv1.JobType, cfg *cicdv1.IntegrationConfig) *cicdv1.Job {
	var jobs []cicdv1.Job

	switch t {
	case cicdv1.JobTypePreSubmit:
		jobs = cfg.Spec.Jobs.PreSubmit
	case cicdv1.JobTypePostSubmit:
		jobs = cfg.Spec.Jobs.PostSubmit
	default:
		return nil
	}

	for _, j := range jobs {
		if j.Name == jobStatus.Name {
			return &j
		}
	}

	return nil
}

func getNotificationFromSpec(jobStatus *cicdv1.JobStatus, jobSpec *cicdv1.Job) *cicdv1.NotificationMethods {
	switch jobStatus.State {
	case cicdv1.CommitStatusStateSuccess:
		return jobSpec.Notification.OnSuccess
	case cicdv1.CommitStatusStateFailure, cicdv1.CommitStatusStateError:
		return jobSpec.Notification.OnFailure
	default:
		return nil
	}
}
