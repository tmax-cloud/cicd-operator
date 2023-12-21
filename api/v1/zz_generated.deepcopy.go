//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/pod"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AddTektonTaskParam) DeepCopyInto(out *AddTektonTaskParam) {
	*out = *in
	if in.TektonTask != nil {
		in, out := &in.TektonTask, &out.TektonTask
		*out = make([]TektonTaskDef, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AddTektonTaskParam.
func (in *AddTektonTaskParam) DeepCopy() *AddTektonTaskParam {
	if in == nil {
		return nil
	}
	out := new(AddTektonTaskParam)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Approval) DeepCopyInto(out *Approval) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Approval.
func (in *Approval) DeepCopy() *Approval {
	if in == nil {
		return nil
	}
	out := new(Approval)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Approval) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApprovalList) DeepCopyInto(out *ApprovalList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Approval, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApprovalList.
func (in *ApprovalList) DeepCopy() *ApprovalList {
	if in == nil {
		return nil
	}
	out := new(ApprovalList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ApprovalList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApprovalSpec) DeepCopyInto(out *ApprovalSpec) {
	*out = *in
	if in.Sender != nil {
		in, out := &in.Sender, &out.Sender
		*out = new(ApprovalUser)
		**out = **in
	}
	if in.Users != nil {
		in, out := &in.Users, &out.Users
		*out = make([]ApprovalUser, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApprovalSpec.
func (in *ApprovalSpec) DeepCopy() *ApprovalSpec {
	if in == nil {
		return nil
	}
	out := new(ApprovalSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApprovalStatus) DeepCopyInto(out *ApprovalStatus) {
	*out = *in
	if in.DecisionTime != nil {
		in, out := &in.DecisionTime, &out.DecisionTime
		*out = (*in).DeepCopy()
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApprovalStatus.
func (in *ApprovalStatus) DeepCopy() *ApprovalStatus {
	if in == nil {
		return nil
	}
	out := new(ApprovalStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApprovalUser) DeepCopyInto(out *ApprovalUser) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApprovalUser.
func (in *ApprovalUser) DeepCopy() *ApprovalUser {
	if in == nil {
		return nil
	}
	out := new(ApprovalUser)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitConfig) DeepCopyInto(out *GitConfig) {
	*out = *in
	if in.Token != nil {
		in, out := &in.Token, &out.Token
		*out = new(GitToken)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitConfig.
func (in *GitConfig) DeepCopy() *GitConfig {
	if in == nil {
		return nil
	}
	out := new(GitConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitToken) DeepCopyInto(out *GitToken) {
	*out = *in
	if in.ValueFrom != nil {
		in, out := &in.ValueFrom, &out.ValueFrom
		*out = new(GitTokenFrom)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitToken.
func (in *GitToken) DeepCopy() *GitToken {
	if in == nil {
		return nil
	}
	out := new(GitToken)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitTokenFrom) DeepCopyInto(out *GitTokenFrom) {
	*out = *in
	in.SecretKeyRef.DeepCopyInto(&out.SecretKeyRef)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitTokenFrom.
func (in *GitTokenFrom) DeepCopy() *GitTokenFrom {
	if in == nil {
		return nil
	}
	out := new(GitTokenFrom)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationConfig) DeepCopyInto(out *IntegrationConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationConfig.
func (in *IntegrationConfig) DeepCopy() *IntegrationConfig {
	if in == nil {
		return nil
	}
	out := new(IntegrationConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IntegrationConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationConfigJobs) DeepCopyInto(out *IntegrationConfigJobs) {
	*out = *in
	if in.PreSubmit != nil {
		in, out := &in.PreSubmit, &out.PreSubmit
		*out = make(Jobs, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.PostSubmit != nil {
		in, out := &in.PostSubmit, &out.PostSubmit
		*out = make(Jobs, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Periodic != nil {
		in, out := &in.Periodic, &out.Periodic
		*out = make(Periodics, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationConfigJobs.
func (in *IntegrationConfigJobs) DeepCopy() *IntegrationConfigJobs {
	if in == nil {
		return nil
	}
	out := new(IntegrationConfigJobs)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationConfigList) DeepCopyInto(out *IntegrationConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]IntegrationConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationConfigList.
func (in *IntegrationConfigList) DeepCopy() *IntegrationConfigList {
	if in == nil {
		return nil
	}
	out := new(IntegrationConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IntegrationConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationConfigSpec) DeepCopyInto(out *IntegrationConfigSpec) {
	*out = *in
	in.Git.DeepCopyInto(&out.Git)
	if in.Secrets != nil {
		in, out := &in.Secrets, &out.Secrets
		*out = make([]corev1.LocalObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.Workspaces != nil {
		in, out := &in.Workspaces, &out.Workspaces
		*out = make([]v1beta1.WorkspaceBinding, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Jobs.DeepCopyInto(&out.Jobs)
	if in.MergeConfig != nil {
		in, out := &in.MergeConfig, &out.MergeConfig
		*out = new(MergeConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.PodTemplate != nil {
		in, out := &in.PodTemplate, &out.PodTemplate
		*out = new(pod.Template)
		(*in).DeepCopyInto(*out)
	}
	in.IJManageSpec.DeepCopyInto(&out.IJManageSpec)
	if in.ParamConfig != nil {
		in, out := &in.ParamConfig, &out.ParamConfig
		*out = new(ParameterConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.TLSConfig != nil {
		in, out := &in.TLSConfig, &out.TLSConfig
		*out = new(TLSConfig)
		**out = **in
	}
	if in.When != nil {
		in, out := &in.When, &out.When
		*out = new(JobWhen)
		(*in).DeepCopyInto(*out)
	}
	if in.GolbalNotification != nil {
		in, out := &in.GolbalNotification, &out.GolbalNotification
		*out = new(Notification)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationConfigSpec.
func (in *IntegrationConfigSpec) DeepCopy() *IntegrationConfigSpec {
	if in == nil {
		return nil
	}
	out := new(IntegrationConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationConfigStatus) DeepCopyInto(out *IntegrationConfigStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationConfigStatus.
func (in *IntegrationConfigStatus) DeepCopy() *IntegrationConfigStatus {
	if in == nil {
		return nil
	}
	out := new(IntegrationConfigStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationJob) DeepCopyInto(out *IntegrationJob) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationJob.
func (in *IntegrationJob) DeepCopy() *IntegrationJob {
	if in == nil {
		return nil
	}
	out := new(IntegrationJob)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IntegrationJob) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationJobConfigRef) DeepCopyInto(out *IntegrationJobConfigRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationJobConfigRef.
func (in *IntegrationJobConfigRef) DeepCopy() *IntegrationJobConfigRef {
	if in == nil {
		return nil
	}
	out := new(IntegrationJobConfigRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationJobList) DeepCopyInto(out *IntegrationJobList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]IntegrationJob, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationJobList.
func (in *IntegrationJobList) DeepCopy() *IntegrationJobList {
	if in == nil {
		return nil
	}
	out := new(IntegrationJobList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IntegrationJobList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationJobManageSpec) DeepCopyInto(out *IntegrationJobManageSpec) {
	*out = *in
	if in.Timeout != nil {
		in, out := &in.Timeout, &out.Timeout
		*out = new(metav1.Duration)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationJobManageSpec.
func (in *IntegrationJobManageSpec) DeepCopy() *IntegrationJobManageSpec {
	if in == nil {
		return nil
	}
	out := new(IntegrationJobManageSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationJobRefs) DeepCopyInto(out *IntegrationJobRefs) {
	*out = *in
	if in.Sender != nil {
		in, out := &in.Sender, &out.Sender
		*out = new(IntegrationJobSender)
		**out = **in
	}
	out.Base = in.Base
	if in.Pulls != nil {
		in, out := &in.Pulls, &out.Pulls
		*out = make([]IntegrationJobRefsPull, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationJobRefs.
func (in *IntegrationJobRefs) DeepCopy() *IntegrationJobRefs {
	if in == nil {
		return nil
	}
	out := new(IntegrationJobRefs)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationJobRefsBase) DeepCopyInto(out *IntegrationJobRefsBase) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationJobRefsBase.
func (in *IntegrationJobRefsBase) DeepCopy() *IntegrationJobRefsBase {
	if in == nil {
		return nil
	}
	out := new(IntegrationJobRefsBase)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationJobRefsPull) DeepCopyInto(out *IntegrationJobRefsPull) {
	*out = *in
	out.Author = in.Author
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationJobRefsPull.
func (in *IntegrationJobRefsPull) DeepCopy() *IntegrationJobRefsPull {
	if in == nil {
		return nil
	}
	out := new(IntegrationJobRefsPull)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationJobRefsPullAuthor) DeepCopyInto(out *IntegrationJobRefsPullAuthor) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationJobRefsPullAuthor.
func (in *IntegrationJobRefsPullAuthor) DeepCopy() *IntegrationJobRefsPullAuthor {
	if in == nil {
		return nil
	}
	out := new(IntegrationJobRefsPullAuthor)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationJobSender) DeepCopyInto(out *IntegrationJobSender) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationJobSender.
func (in *IntegrationJobSender) DeepCopy() *IntegrationJobSender {
	if in == nil {
		return nil
	}
	out := new(IntegrationJobSender)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationJobSpec) DeepCopyInto(out *IntegrationJobSpec) {
	*out = *in
	out.ConfigRef = in.ConfigRef
	if in.Workspaces != nil {
		in, out := &in.Workspaces, &out.Workspaces
		*out = make([]v1beta1.WorkspaceBinding, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Jobs != nil {
		in, out := &in.Jobs, &out.Jobs
		*out = make(Jobs, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Refs.DeepCopyInto(&out.Refs)
	if in.PodTemplate != nil {
		in, out := &in.PodTemplate, &out.PodTemplate
		*out = new(pod.Template)
		(*in).DeepCopyInto(*out)
	}
	if in.Timeout != nil {
		in, out := &in.Timeout, &out.Timeout
		*out = new(metav1.Duration)
		**out = **in
	}
	if in.ParamConfig != nil {
		in, out := &in.ParamConfig, &out.ParamConfig
		*out = new(ParameterConfig)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationJobSpec.
func (in *IntegrationJobSpec) DeepCopy() *IntegrationJobSpec {
	if in == nil {
		return nil
	}
	out := new(IntegrationJobSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationJobStatus) DeepCopyInto(out *IntegrationJobStatus) {
	*out = *in
	if in.StartTime != nil {
		in, out := &in.StartTime, &out.StartTime
		*out = (*in).DeepCopy()
	}
	if in.CompletionTime != nil {
		in, out := &in.CompletionTime, &out.CompletionTime
		*out = (*in).DeepCopy()
	}
	if in.Jobs != nil {
		in, out := &in.Jobs, &out.Jobs
		*out = make([]JobStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationJobStatus.
func (in *IntegrationJobStatus) DeepCopy() *IntegrationJobStatus {
	if in == nil {
		return nil
	}
	out := new(IntegrationJobStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Job) DeepCopyInto(out *Job) {
	*out = *in
	in.Container.DeepCopyInto(&out.Container)
	if in.When != nil {
		in, out := &in.When, &out.When
		*out = new(JobWhen)
		(*in).DeepCopyInto(*out)
	}
	if in.After != nil {
		in, out := &in.After, &out.After
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.TektonTask != nil {
		in, out := &in.TektonTask, &out.TektonTask
		*out = new(TektonTask)
		(*in).DeepCopyInto(*out)
	}
	if in.Approval != nil {
		in, out := &in.Approval, &out.Approval
		*out = new(JobApproval)
		(*in).DeepCopyInto(*out)
	}
	in.NotificationMethods.DeepCopyInto(&out.NotificationMethods)
	if in.Notification != nil {
		in, out := &in.Notification, &out.Notification
		*out = new(Notification)
		(*in).DeepCopyInto(*out)
	}
	if in.TektonWhen != nil {
		in, out := &in.TektonWhen, &out.TektonWhen
		*out = make(v1beta1.WhenExpressions, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Results != nil {
		in, out := &in.Results, &out.Results
		*out = make([]v1beta1.TaskResult, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Job.
func (in *Job) DeepCopy() *Job {
	if in == nil {
		return nil
	}
	out := new(Job)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JobApproval) DeepCopyInto(out *JobApproval) {
	*out = *in
	if in.Approvers != nil {
		in, out := &in.Approvers, &out.Approvers
		*out = make([]ApprovalUser, len(*in))
		copy(*out, *in)
	}
	if in.ApproversConfigMap != nil {
		in, out := &in.ApproversConfigMap, &out.ApproversConfigMap
		*out = new(corev1.LocalObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JobApproval.
func (in *JobApproval) DeepCopy() *JobApproval {
	if in == nil {
		return nil
	}
	out := new(JobApproval)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JobStatus) DeepCopyInto(out *JobStatus) {
	*out = *in
	if in.StartTime != nil {
		in, out := &in.StartTime, &out.StartTime
		*out = (*in).DeepCopy()
	}
	if in.CompletionTime != nil {
		in, out := &in.CompletionTime, &out.CompletionTime
		*out = (*in).DeepCopy()
	}
	if in.Containers != nil {
		in, out := &in.Containers, &out.Containers
		*out = make([]v1beta1.StepState, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JobStatus.
func (in *JobStatus) DeepCopy() *JobStatus {
	if in == nil {
		return nil
	}
	out := new(JobStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JobTaskRef) DeepCopyInto(out *JobTaskRef) {
	*out = *in
	if in.Local != nil {
		in, out := &in.Local, &out.Local
		*out = new(v1beta1.TaskRef)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JobTaskRef.
func (in *JobTaskRef) DeepCopy() *JobTaskRef {
	if in == nil {
		return nil
	}
	out := new(JobTaskRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JobWhen) DeepCopyInto(out *JobWhen) {
	*out = *in
	if in.Branch != nil {
		in, out := &in.Branch, &out.Branch
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SkipBranch != nil {
		in, out := &in.SkipBranch, &out.SkipBranch
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Tag != nil {
		in, out := &in.Tag, &out.Tag
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SkipTag != nil {
		in, out := &in.SkipTag, &out.SkipTag
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JobWhen.
func (in *JobWhen) DeepCopy() *JobWhen {
	if in == nil {
		return nil
	}
	out := new(JobWhen)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in Jobs) DeepCopyInto(out *Jobs) {
	{
		in := &in
		*out = make(Jobs, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Jobs.
func (in Jobs) DeepCopy() Jobs {
	if in == nil {
		return nil
	}
	out := new(Jobs)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MergeConfig) DeepCopyInto(out *MergeConfig) {
	*out = *in
	in.Query.DeepCopyInto(&out.Query)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MergeConfig.
func (in *MergeConfig) DeepCopy() *MergeConfig {
	if in == nil {
		return nil
	}
	out := new(MergeConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MergeQuery) DeepCopyInto(out *MergeQuery) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.BlockLabels != nil {
		in, out := &in.BlockLabels, &out.BlockLabels
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Authors != nil {
		in, out := &in.Authors, &out.Authors
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SkipAuthors != nil {
		in, out := &in.SkipAuthors, &out.SkipAuthors
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Branches != nil {
		in, out := &in.Branches, &out.Branches
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SkipBranches != nil {
		in, out := &in.SkipBranches, &out.SkipBranches
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Checks != nil {
		in, out := &in.Checks, &out.Checks
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.OptionalChecks != nil {
		in, out := &in.OptionalChecks, &out.OptionalChecks
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MergeQuery.
func (in *MergeQuery) DeepCopy() *MergeQuery {
	if in == nil {
		return nil
	}
	out := new(MergeQuery)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NotiEmail) DeepCopyInto(out *NotiEmail) {
	*out = *in
	if in.Receivers != nil {
		in, out := &in.Receivers, &out.Receivers
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NotiEmail.
func (in *NotiEmail) DeepCopy() *NotiEmail {
	if in == nil {
		return nil
	}
	out := new(NotiEmail)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NotiSlack) DeepCopyInto(out *NotiSlack) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NotiSlack.
func (in *NotiSlack) DeepCopy() *NotiSlack {
	if in == nil {
		return nil
	}
	out := new(NotiSlack)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NotiWebhook) DeepCopyInto(out *NotiWebhook) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NotiWebhook.
func (in *NotiWebhook) DeepCopy() *NotiWebhook {
	if in == nil {
		return nil
	}
	out := new(NotiWebhook)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Notification) DeepCopyInto(out *Notification) {
	*out = *in
	if in.OnSuccess != nil {
		in, out := &in.OnSuccess, &out.OnSuccess
		*out = new(NotificationMethods)
		(*in).DeepCopyInto(*out)
	}
	if in.OnFailure != nil {
		in, out := &in.OnFailure, &out.OnFailure
		*out = new(NotificationMethods)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Notification.
func (in *Notification) DeepCopy() *Notification {
	if in == nil {
		return nil
	}
	out := new(Notification)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NotificationMethods) DeepCopyInto(out *NotificationMethods) {
	*out = *in
	if in.Email != nil {
		in, out := &in.Email, &out.Email
		*out = new(NotiEmail)
		(*in).DeepCopyInto(*out)
	}
	if in.Slack != nil {
		in, out := &in.Slack, &out.Slack
		*out = new(NotiSlack)
		**out = **in
	}
	if in.Webhook != nil {
		in, out := &in.Webhook, &out.Webhook
		*out = new(NotiWebhook)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NotificationMethods.
func (in *NotificationMethods) DeepCopy() *NotificationMethods {
	if in == nil {
		return nil
	}
	out := new(NotificationMethods)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ParameterConfig) DeepCopyInto(out *ParameterConfig) {
	*out = *in
	if in.ParamDefine != nil {
		in, out := &in.ParamDefine, &out.ParamDefine
		*out = make([]ParameterDefine, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ParamValue != nil {
		in, out := &in.ParamValue, &out.ParamValue
		*out = make([]ParameterValue, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ParameterConfig.
func (in *ParameterConfig) DeepCopy() *ParameterConfig {
	if in == nil {
		return nil
	}
	out := new(ParameterConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ParameterDefine) DeepCopyInto(out *ParameterDefine) {
	*out = *in
	if in.DefaultArray != nil {
		in, out := &in.DefaultArray, &out.DefaultArray
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ParameterDefine.
func (in *ParameterDefine) DeepCopy() *ParameterDefine {
	if in == nil {
		return nil
	}
	out := new(ParameterDefine)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ParameterValue) DeepCopyInto(out *ParameterValue) {
	*out = *in
	if in.ArrayVal != nil {
		in, out := &in.ArrayVal, &out.ArrayVal
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ParameterValue.
func (in *ParameterValue) DeepCopy() *ParameterValue {
	if in == nil {
		return nil
	}
	out := new(ParameterValue)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Periodic) DeepCopyInto(out *Periodic) {
	*out = *in
	in.Job.DeepCopyInto(&out.Job)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Periodic.
func (in *Periodic) DeepCopy() *Periodic {
	if in == nil {
		return nil
	}
	out := new(Periodic)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in Periodics) DeepCopyInto(out *Periodics) {
	{
		in := &in
		*out = make(Periodics, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Periodics.
func (in Periodics) DeepCopy() Periodics {
	if in == nil {
		return nil
	}
	out := new(Periodics)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TLSConfig) DeepCopyInto(out *TLSConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TLSConfig.
func (in *TLSConfig) DeepCopy() *TLSConfig {
	if in == nil {
		return nil
	}
	out := new(TLSConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TektonTask) DeepCopyInto(out *TektonTask) {
	*out = *in
	in.TaskRef.DeepCopyInto(&out.TaskRef)
	if in.Params != nil {
		in, out := &in.Params, &out.Params
		*out = make([]ParameterValue, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(v1beta1.TaskRunResources)
		(*in).DeepCopyInto(*out)
	}
	if in.Workspaces != nil {
		in, out := &in.Workspaces, &out.Workspaces
		*out = make([]v1beta1.WorkspacePipelineTaskBinding, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TektonTask.
func (in *TektonTask) DeepCopy() *TektonTask {
	if in == nil {
		return nil
	}
	out := new(TektonTask)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TektonTaskDef) DeepCopyInto(out *TektonTaskDef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TektonTaskDef.
func (in *TektonTaskDef) DeepCopy() *TektonTaskDef {
	if in == nil {
		return nil
	}
	out := new(TektonTaskDef)
	in.DeepCopyInto(out)
	return out
}
