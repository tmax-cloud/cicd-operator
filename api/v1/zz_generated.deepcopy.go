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
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

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
func (in *IntegrationConfigJobs) DeepCopyInto(out *IntegrationConfigJobs) {
	*out = *in
	if in.PreSubmit != nil {
		in, out := &in.PreSubmit, &out.PreSubmit
		*out = make([]Job, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.PostSubmit != nil {
		in, out := &in.PostSubmit, &out.PostSubmit
		*out = make([]Job, len(*in))
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
	in.Jobs.DeepCopyInto(&out.Jobs)
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
func (in *IntegrationJobSpec) DeepCopyInto(out *IntegrationJobSpec) {
	*out = *in
	out.ConfigRef = in.ConfigRef
	if in.Jobs != nil {
		in, out := &in.Jobs, &out.Jobs
		*out = make([]Job, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Refs.DeepCopyInto(&out.Refs)
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
	if in.CompletionTime != nil {
		in, out := &in.CompletionTime, &out.CompletionTime
		*out = (*in).DeepCopy()
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
	if in.TektonTask != nil {
		in, out := &in.TektonTask, &out.TektonTask
		*out = new(TektonTask)
		(*in).DeepCopyInto(*out)
	}
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
func (in *JobTaskRef) DeepCopyInto(out *JobTaskRef) {
	*out = *in
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
	if in.Ref != nil {
		in, out := &in.Ref, &out.Ref
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SkipRef != nil {
		in, out := &in.SkipRef, &out.SkipRef
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
func (in *TektonTask) DeepCopyInto(out *TektonTask) {
	*out = *in
	out.TaskRef = in.TaskRef
	if in.Params != nil {
		in, out := &in.Params, &out.Params
		*out = make([]v1beta1.Param, len(*in))
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
		*out = make([]v1beta1.WorkspaceBinding, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
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
