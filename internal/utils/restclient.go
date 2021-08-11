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
	"context"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// RestClient is a group-version-resource specific rest client
type RestClient interface {
	Get(name string, getOpt *metav1.GetOptions, into runtime.Object) error
	List(listOpt *metav1.ListOptions, into runtime.Object) error
	Watch(listOpt *metav1.ListOptions) (watch.Interface, error)
	Create(obj runtime.Object, createOpt *metav1.CreateOptions) error
	Update(obj runtime.Object, opts *metav1.UpdateOptions) error
	Delete(name string, opt *metav1.DeleteOptions) error
}

type gvrClient struct {
	client rest.Interface

	ns       string
	resource string
}

// NewGroupVersionResourceClient creates a new group-version-resource specific rest client
func NewGroupVersionResourceClient(cfg *rest.Config, ns string, groupVersionResource schema.GroupVersionResource) (RestClient, error) {
	restCfg := rest.CopyConfig(cfg)
	if groupVersionResource.Group == "" {
		restCfg.APIPath = "/api"
	} else {
		restCfg.APIPath = "/apis"
	}
	gv := groupVersionResource.GroupVersion()
	restCfg.ContentConfig.GroupVersion = &gv
	restCfg.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	restCli, err := rest.RESTClientFor(restCfg)
	if err != nil {
		return nil, err
	}
	return &gvrClient{client: restCli, ns: ns, resource: groupVersionResource.Resource}, nil
}

func (s *gvrClient) Get(name string, getOpt *metav1.GetOptions, into runtime.Object) error {
	return s.client.Get().
		Namespace(s.ns).
		Resource(s.resource).
		Name(name).
		VersionedParams(getOpt, scheme.ParameterCodec).
		Do(context.Background()).
		Into(into)
}

func (s *gvrClient) List(listOpt *metav1.ListOptions, into runtime.Object) error {
	return s.client.Get().
		Namespace(s.ns).
		Resource(s.resource).
		VersionedParams(listOpt, scheme.ParameterCodec).
		Do(context.Background()).
		Into(into)
}

func (s *gvrClient) Watch(listOpt *metav1.ListOptions) (watch.Interface, error) {
	opt := listOpt.DeepCopy()
	opt.Watch = true
	return s.client.Get().
		Namespace(s.ns).
		Resource(s.resource).
		VersionedParams(opt, scheme.ParameterCodec).
		Watch(context.Background())
}

func (s *gvrClient) Create(obj runtime.Object, opt *metav1.CreateOptions) error {
	into := obj.DeepCopyObject()
	return s.client.Post().
		Namespace(s.ns).
		Resource(s.resource).
		VersionedParams(opt, scheme.ParameterCodec).
		Body(obj).
		Do(context.Background()).
		Into(into)
}

func (s *gvrClient) Update(obj runtime.Object, opts *metav1.UpdateOptions) error {
	into := obj.DeepCopyObject()
	objMeta, err := apimeta.Accessor(obj)
	if err != nil {
		return err
	}
	return s.client.Put().
		Namespace(s.ns).
		Resource(s.resource).
		Name(objMeta.GetName()).
		VersionedParams(opts, scheme.ParameterCodec).
		Body(obj).
		Do(context.Background()).
		Into(into)
}

func (s *gvrClient) Delete(name string, opt *metav1.DeleteOptions) error {
	return s.client.Delete().
		Namespace(s.ns).
		Resource(s.resource).
		Name(name).
		Body(opt).
		Do(context.Background()).
		Error()
}
