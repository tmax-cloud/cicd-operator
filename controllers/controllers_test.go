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

package controllers

import (
	"net/http"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

type fakeManager struct {
	scheme *runtime.Scheme
}

func (f *fakeManager) Add(manager.Runnable) error {
	return nil
}
func (f *fakeManager) Elected() <-chan struct{} {
	return make(chan struct{})
}
func (f *fakeManager) SetFields(interface{}) error {
	return nil
}
func (f *fakeManager) AddMetricsExtraHandler(_ string, _ http.Handler) error {
	return nil
}
func (f *fakeManager) AddHealthzCheck(_ string, _ healthz.Checker) error {
	return nil
}
func (f *fakeManager) AddReadyzCheck(_ string, _ healthz.Checker) error {
	return nil
}
func (f *fakeManager) Start(<-chan struct{}) error {
	return nil
}
func (f *fakeManager) GetConfig() *rest.Config {
	return nil
}
func (f *fakeManager) GetScheme() *runtime.Scheme {
	return f.scheme
}
func (f *fakeManager) GetClient() client.Client {
	return nil
}
func (f *fakeManager) GetFieldIndexer() client.FieldIndexer {
	return nil
}
func (f *fakeManager) GetCache() cache.Cache {
	return nil
}
func (f *fakeManager) GetEventRecorderFor(_ string) record.EventRecorder {
	return nil
}
func (f *fakeManager) GetRESTMapper() meta.RESTMapper {
	return nil
}
func (f *fakeManager) GetAPIReader() client.Reader {
	return nil
}
func (f *fakeManager) GetWebhookServer() *webhook.Server {
	return nil
}
func (f *fakeManager) GetLogger() logr.Logger {
	return log.Log
}

type fakeLogger struct {
	info     []string
	error    []error
	errorMsg []string
}

func (f *fakeLogger) Clear() {
	f.info = nil
	f.error = nil
	f.errorMsg = nil
}

func (f *fakeLogger) Info(msg string, _ ...interface{}) {
	f.info = append(f.info, msg)
}
func (f *fakeLogger) Enabled() bool { return true }
func (f *fakeLogger) Error(err error, msg string, _ ...interface{}) {
	f.error = append(f.error, err)
	f.errorMsg = append(f.errorMsg, msg)
}
func (f *fakeLogger) V(_ int) logr.InfoLogger                 { return f }
func (f *fakeLogger) WithValues(_ ...interface{}) logr.Logger { return f }
func (f *fakeLogger) WithName(_ string) logr.Logger           { return f }
