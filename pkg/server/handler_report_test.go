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

package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	applyconfigurationscorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func Test_reportHandler_ServeHTTP(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	require.NoError(t, cicdv1.AddToScheme(scheme.Scheme))

	// Set templates
	templatePath := path.Join(os.TempDir(), "cicd-cont-test-report")
	require.NoError(t, os.MkdirAll(templatePath, os.ModePerm))
	defer func() {
		_ = os.RemoveAll(templatePath)
	}()

	require.NoError(t, os.Setenv("REPORT_TEMPLATE_PATH", templatePath))

	tc := map[string]struct {
		ns          string
		name        string
		job         string
		ij          *cicdv1.IntegrationJob
		redirectURL string
		template    []byte

		expectedCode    int
		expectedMessage string
	}{
		"normal": {
			ns:   "default",
			name: "test",
			job:  "job1",
			ij: &cicdv1.IntegrationJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test",
				},
				Status: cicdv1.IntegrationJobStatus{
					Jobs: []cicdv1.JobStatus{
						{Name: "job1", PodName: "pod1"},
					},
				},
			},
			template: []byte("TEST: {{.JobName}} {{.Log}}"),

			expectedCode:    http.StatusOK,
			expectedMessage: "TEST: test # Step : step-step-0\nThis is the log of the pod1 - step-step-0\n\n# Step : step-step-1\nThis is the log of the pod1 - step-step-1\n\n",
		},
		"noJobJobErr": {
			ns:   "default",
			name: "test",

			expectedCode:    http.StatusBadRequest,
			expectedMessage: "path is not in form of '/report/{namespace}/{jobName}/{jobJobName}'",
		},
		"getErr": {
			ns:   "default",
			name: "test",
			job:  "job1",

			expectedCode:    http.StatusBadRequest,
			expectedMessage: "cannot get IntegrationJob default/test",
		},
		"redirect": {
			ns:   "default",
			name: "test",
			job:  "job1",
			ij: &cicdv1.IntegrationJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test",
				},
				Status: cicdv1.IntegrationJobStatus{
					Jobs: []cicdv1.JobStatus{
						{Name: "job1", PodName: "pod1"},
					},
				},
			},
			redirectURL: "https://{{.Name}}",

			expectedCode: http.StatusMovedPermanently,
		},
		"redirectParseErr": {
			ns:   "default",
			name: "test",
			job:  "job1",
			ij: &cicdv1.IntegrationJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test",
				},
				Status: cicdv1.IntegrationJobStatus{
					Jobs: []cicdv1.JobStatus{
						{Name: "job1", PodName: "pod1"},
					},
				},
			},
			redirectURL: "{{...}}",

			expectedCode:    http.StatusBadRequest,
			expectedMessage: "cannot parse report redirection uri template",
		},
		"redirectExecErr": {
			ns:   "default",
			name: "test",
			job:  "job1",
			ij: &cicdv1.IntegrationJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test",
				},
				Status: cicdv1.IntegrationJobStatus{
					Jobs: []cicdv1.JobStatus{
						{Name: "job1", PodName: "pod1"},
					},
				},
			},
			redirectURL: "{{.NNAME}}",

			expectedCode:    http.StatusBadRequest,
			expectedMessage: "cannot execute report redirection uri template",
		},
		"jobStatusNil": {
			ns:   "default",
			name: "test",
			job:  "job1",
			ij: &cicdv1.IntegrationJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test",
				},
			},

			expectedCode:    http.StatusBadRequest,
			expectedMessage: "there is no job status job1 in IntegrationJob default/test",
		},
		"podLogErr": {
			ns:   "default",
			name: "test",
			job:  "job1",
			ij: &cicdv1.IntegrationJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test",
				},
				Status: cicdv1.IntegrationJobStatus{
					Jobs: []cicdv1.JobStatus{
						{Name: "job1", PodName: "pod2"},
					},
				},
			},
			template: []byte("TEST: {{.JobName}} {{.Log}}"),

			expectedCode:    http.StatusOK,
			expectedMessage: "TEST: test log does not exist... maybe the pod does not exist",
		},
		"templateParseErr": {
			ns:   "default",
			name: "test",
			job:  "job1",
			ij: &cicdv1.IntegrationJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test",
				},
				Status: cicdv1.IntegrationJobStatus{
					Jobs: []cicdv1.JobStatus{
						{Name: "job1", PodName: "pod1"},
					},
				},
			},
			template: []byte("TEST: {{.....JobName}} {{.Log}}"),

			expectedCode:    http.StatusBadRequest,
			expectedMessage: "cannot parse report template",
		},
		"templateExecErr": {
			ns:   "default",
			name: "test",
			job:  "job1",
			ij: &cicdv1.IntegrationJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test",
				},
				Status: cicdv1.IntegrationJobStatus{
					Jobs: []cicdv1.JobStatus{
						{Name: "job1", PodName: "pod1"},
					},
				},
			},
			template: []byte("TEST: {{.JobNameeeeeee}} {{.Log}}"),

			expectedCode:    http.StatusBadRequest,
			expectedMessage: "cannot execute report template",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			configs.ReportRedirectURITemplate = c.redirectURL
			require.NoError(t, ioutil.WriteFile(path.Join(templatePath, "template"), c.template, os.ModePerm))

			pod1 := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "step-step-0"},
						{Name: "step-step-1"},
					},
				},
			}

			fakeCli := ctrlfake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(pod1).Build()
			if c.ij != nil {
				require.NoError(t, fakeCli.Create(context.Background(), c.ij))
			}

			testSrv := httptest.NewServer(&fakeLogHandler{
				pods: map[string]struct{ containerLogs map[string]string }{
					"default/pod1": {
						containerLogs: map[string]string{
							"step-step-0": "This is the log of the pod1 - step-step-0",
							"step-step-1": "This is the log of the pod1 - step-step-1",
						},
					},
				},
			})
			handler := &reportHandler{k8sClient: fakeCli, podsGetter: &fakePodGetter{URL: testSrv.URL}}

			req := httptest.NewRequest(http.MethodPost, "https://test", nil)
			req = mux.SetURLVars(req, map[string]string{
				paramKeyNamespace: c.ns,
				paramKeyIJName:    c.name,
				paramKeyJobName:   c.job,
			})
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			b, err := ioutil.ReadAll(w.Body)
			require.NoError(t, err)

			require.Equal(t, c.expectedCode, w.Code)
			require.Containsf(t, string(b), c.expectedMessage, "contains")
		})
	}
}

func Test_reportHandler_getTemplate(t *testing.T) {
	// Set templates
	templatePath := path.Join(os.TempDir(), "cicd-cont-test-report")
	require.NoError(t, os.MkdirAll(templatePath, os.ModePerm))
	defer func() {
		_ = os.RemoveAll(templatePath)
	}()

	tc := map[string]struct {
		env      string
		template []byte

		errorOccurs  bool
		errorMessage string
	}{
		"normal": {
			env: templatePath,
		},
		"default": {
			errorOccurs:  true,
			errorMessage: "open /templates/report/template",
		},
		"parseErr": {
			env:          templatePath,
			template:     []byte("{{....}}"),
			errorOccurs:  true,
			errorMessage: "template: template:1: unexpected <.> in operand",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, os.Setenv("REPORT_TEMPLATE_PATH", c.env))
			require.NoError(t, ioutil.WriteFile(path.Join(templatePath, "template"), c.template, os.ModePerm))

			handler := &reportHandler{}
			_, err := handler.getTemplate()
			if c.errorOccurs {
				require.Error(t, err)
				require.Containsf(t, err.Error(), c.errorMessage, "contains")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_reportHandler_getPodLogs(t *testing.T) {
	tc := map[string]struct {
		name      string
		namespace string

		errorOccurs  bool
		errorMessage string
		expectedLog  string
	}{
		"normal": {
			name:      "pod1",
			namespace: "default",

			expectedLog: "# Step : step-step-0\nThis is the log of the pod1 - step-step-0\n\n# Step : step-step-1\nThis is the log of the pod1 - step-step-1\n\n# Step : step-step-2\n\n\n",
		},
		"emptyNs": {
			name:      "pod1",
			namespace: "",

			errorOccurs:  true,
			errorMessage: "podName and namespace should not be empty",
		},
		"getErr": {
			name:      "pod2",
			namespace: "default",

			errorOccurs:  true,
			errorMessage: "pods \"pod2\" not found",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			pod1 := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "step-step-0"},
						{Name: "step-step-1"},
						{Name: "step-step-2"},
					},
				},
			}

			fakeCli := ctrlfake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(pod1).Build()
			testSrv := httptest.NewServer(&fakeLogHandler{
				pods: map[string]struct{ containerLogs map[string]string }{
					"default/pod1": {
						containerLogs: map[string]string{
							"step-step-0": "This is the log of the pod1 - step-step-0",
							"step-step-1": "This is the log of the pod1 - step-step-1",
						},
					},
				},
			})
			handler := &reportHandler{k8sClient: fakeCli, podsGetter: &fakePodGetter{URL: testSrv.URL}}

			log, err := handler.getPodLogs(c.name, c.namespace, logf.Log)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expectedLog, log)
			}
		})
	}
}

func Test_reportHandler_getPodLog(t *testing.T) {
	tc := map[string]struct {
		name      string
		namespace string
		container string

		errorOccurs  bool
		errorMessage string
		expectedLog  string
	}{
		"normal": {
			name:      "pod1",
			namespace: "default",
			container: "step-step-0",

			expectedLog: "This is the log of the pod1 - step-step-0",
		},
		"streamErr": {
			name:      "pod2",
			namespace: "default",
			container: "container1",

			errorOccurs:  true,
			errorMessage: "the server could not find the requested resource (get pods pod2)",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			testSrv := httptest.NewServer(&fakeLogHandler{
				pods: map[string]struct{ containerLogs map[string]string }{
					"default/pod1": {
						containerLogs: map[string]string{
							"step-step-0": "This is the log of the pod1 - step-step-0",
						},
					},
				},
			})
			handler := &reportHandler{podsGetter: &fakePodGetter{URL: testSrv.URL}}

			log, err := handler.getPodLog(c.name, c.namespace, c.container)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expectedLog, log)
			}
		})
	}
}

type fakePodGetter struct {
	URL string
}

func (f *fakePodGetter) Pods(ns string) typedcorev1.PodInterface {
	return &fakeLogStreamer{URL: f.URL, ns: ns}
}

type fakeLogStreamer struct {
	URL string
	ns  string
}

func (f *fakeLogStreamer) Apply(_ context.Context, _ *applyconfigurationscorev1.PodApplyConfiguration, _ metav1.ApplyOptions) (*corev1.Pod, error) {
	return nil, nil
}

func (f *fakeLogStreamer) ApplyStatus(_ context.Context, _ *applyconfigurationscorev1.PodApplyConfiguration, _ metav1.ApplyOptions) (*corev1.Pod, error) {
	return nil, nil
}

func (f *fakeLogStreamer) EvictV1(_ context.Context, _ *policyv1.Eviction) error {
	return nil
}

func (f *fakeLogStreamer) EvictV1beta1(_ context.Context, _ *policyv1beta1.Eviction) error {
	return nil
}

func (f *fakeLogStreamer) ProxyGet(_, _, _, _ string, _ map[string]string) restclient.ResponseWrapper {
	return nil
}

func (f *fakeLogStreamer) Create(context.Context, *corev1.Pod, metav1.CreateOptions) (*corev1.Pod, error) {
	return nil, nil
}
func (f *fakeLogStreamer) Update(context.Context, *corev1.Pod, metav1.UpdateOptions) (*corev1.Pod, error) {
	return nil, nil
}
func (f *fakeLogStreamer) UpdateStatus(context.Context, *corev1.Pod, metav1.UpdateOptions) (*corev1.Pod, error) {
	return nil, nil
}
func (f *fakeLogStreamer) Delete(context.Context, string, metav1.DeleteOptions) error { return nil }
func (f *fakeLogStreamer) DeleteCollection(context.Context, metav1.DeleteOptions, metav1.ListOptions) error {
	return nil
}
func (f *fakeLogStreamer) Get(context.Context, string, metav1.GetOptions) (*corev1.Pod, error) {
	return nil, nil
}
func (f *fakeLogStreamer) List(context.Context, metav1.ListOptions) (*corev1.PodList, error) {
	return nil, nil
}
func (f *fakeLogStreamer) Watch(context.Context, metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}
func (f *fakeLogStreamer) Patch(context.Context, string, types.PatchType, []byte, metav1.PatchOptions, ...string) (result *corev1.Pod, err error) {
	return nil, nil
}
func (f *fakeLogStreamer) UpdateEphemeralContainers(_ context.Context, _ string, _ *corev1.Pod, _ metav1.UpdateOptions) (*corev1.Pod, error) {
	return nil, nil
}
func (f *fakeLogStreamer) Bind(context.Context, *corev1.Binding, metav1.CreateOptions) error {
	return nil
}
func (f *fakeLogStreamer) Evict(context.Context, *policyv1beta1.Eviction) error { return nil }
func (f *fakeLogStreamer) GetLogs(name string, opts *corev1.PodLogOptions) *restclient.Request {
	u, err := url.Parse(f.URL)
	if err != nil {
		panic(err)
	}
	rCli, err := restclient.NewRESTClient(u, "api/v1", restclient.ClientContentConfig{GroupVersion: schema.GroupVersion{Group: "", Version: "v1"}, Negotiator: runtime.NewClientNegotiator(serializer.NewCodecFactory(scheme.Scheme), schema.GroupVersion{})}, flowcontrol.NewFakeAlwaysRateLimiter(), http.DefaultClient)
	if err != nil {
		panic(err)
	}
	req := restclient.NewRequest(rCli)
	req.Verb(http.MethodGet).Namespace(f.ns).Name(name).Resource("pods").SubResource("log").VersionedParams(opts, scheme.ParameterCodec)
	return req
}

type fakeLogHandler struct {
	pods map[string]struct {
		containerLogs map[string]string
	}
}

func (f *fakeLogHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/v1")

	pathTokens := strings.Split(req.URL.Path, "/")
	cont := req.URL.Query().Get("container")

	if len(pathTokens) != 6 || pathTokens[1] != "namespaces" || pathTokens[3] != "pods" || pathTokens[5] != "log" || cont == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ns := pathTokens[2]
	name := pathTokens[4]

	key := fmt.Sprintf("%s/%s", ns, name)

	p, ok := f.pods[key]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	c, ok := p.containerLogs[cont]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_, _ = w.Write([]byte(c))
}
