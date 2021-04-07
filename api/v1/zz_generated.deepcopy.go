// +build !ignore_autogenerated

/*


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
	"github.com/operator-framework/operator-lib/status"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/pod"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

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
func (in *ApprovalAPIReqBody) DeepCopyInto(out *ApprovalAPIReqBody) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApprovalAPIReqBody.
func (in *ApprovalAPIReqBody) DeepCopy() *ApprovalAPIReqBody {
	if in == nil {
		return nil
	}
	out := new(ApprovalAPIReqBody)
	in.DeepCopyInto(out)
	return out
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
func (in *ApprovalSender) DeepCopyInto(out *ApprovalSender) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApprovalSender.
func (in *ApprovalSender) DeepCopy() *ApprovalSender {
	if in == nil {
		return nil
	}
	out := new(ApprovalSender)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApprovalSpec) DeepCopyInto(out *ApprovalSpec) {
	*out = *in
	if in.Sender != nil {
		in, out := &in.Sender, &out.Sender
		*out = new(ApprovalSender)
		**out = **in
	}
	if in.Users != nil {
		in, out := &in.Users, &out.Users
		*out = make([]string, len(*in))
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
		*out = make(status.Conditions, len(*in))
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
func (in *GitConfig) DeepCopyInto(out *GitConfig) {
	*out = *in
	in.Token.DeepCopyInto(&out.Token)
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
func (in *IntegrationConfigAPIReqRunPostBody) DeepCopyInto(out *IntegrationConfigAPIReqRunPostBody) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationConfigAPIReqRunPostBody.
func (in *IntegrationConfigAPIReqRunPostBody) DeepCopy() *IntegrationConfigAPIReqRunPostBody {
	if in == nil {
		return nil
	}
	out := new(IntegrationConfigAPIReqRunPostBody)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationConfigAPIReqRunPreBody) DeepCopyInto(out *IntegrationConfigAPIReqRunPreBody) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationConfigAPIReqRunPreBody.
func (in *IntegrationConfigAPIReqRunPreBody) DeepCopy() *IntegrationConfigAPIReqRunPreBody {
	if in == nil {
		return nil
	}
	out := new(IntegrationConfigAPIReqRunPreBody)
	in.DeepCopyInto(out)
	return out
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
	if in.PodTemplate != nil {
		in, out := &in.PodTemplate, &out.PodTemplate
		*out = new(pod.Template)
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
		*out = make(status.Conditions, len(*in))
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
func (in *IntegrationJobRefs) DeepCopyInto(out *IntegrationJobRefs) {
	*out = *in
	if in.Sender != nil {
		in, out := &in.Sender, &out.Sender
		*out = new(IntegrationJobSender)
		**out = **in
	}
	out.Base = in.Base
	if in.Pull != nil {
		in, out := &in.Pull, &out.Pull
		*out = new(IntegrationJobRefsPull)
		**out = **in
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
		*out = make([]string, len(*in))
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
func (in *TektonTask) DeepCopyInto(out *TektonTask) {
	*out = *in
	in.TaskRef.DeepCopyInto(&out.TaskRef)
	if in.Params != nil {
		in, out := &in.Params, &out.Params
		*out = make([]TektonTaskParam, len(*in))
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
func (in *TektonTaskParam) DeepCopyInto(out *TektonTaskParam) {
	*out = *in
	if in.ArrayVal != nil {
		in, out := &in.ArrayVal, &out.ArrayVal
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TektonTaskParam.
func (in *TektonTaskParam) DeepCopy() *TektonTaskParam {
	if in == nil {
		return nil
	}
	out := new(TektonTaskParam)
	in.DeepCopyInto(out)
	return out
}
