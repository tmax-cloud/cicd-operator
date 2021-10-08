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
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
)

var reportPath = fmt.Sprintf("/report/{%s}/{%s}/{%s}", paramKeyNamespace, paramKeyIJName, paramKeyJobName)

const (
	templateConfigMapPathEnv     = "REPORT_TEMPLATE_PATH"
	templateConfigMapPathDefault = "/templates/report"
	templateConfigMapKey         = "template"

	errorLogNotExist = "log does not exist... maybe the pod does not exist"
)

type report struct {
	JobName    string
	JobJobName string
	JobStatus  *cicdv1.JobStatus
	Log        string
}

type reportHandler struct {
	k8sClient  client.Client
	podsGetter typedcorev1.PodsGetter
}

func (h *reportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqID := utils.RandomString(10)
	log := logger.WithValues("request", reqID)

	vars := mux.Vars(r)

	ns := vars[paramKeyNamespace]
	ijName := vars[paramKeyIJName]
	job := vars[paramKeyJobName]

	if ns == "" || ijName == "" || job == "" {
		logAndRespond(w, log, http.StatusBadRequest, fmt.Sprintf("req: %s, path is not in form of '%s'", reqID, reportPath),
			fmt.Sprintf("Bad request for path, path: %s", r.RequestURI))
		return
	}

	iJob := &cicdv1.IntegrationJob{}
	if err := h.k8sClient.Get(context.Background(), types.NamespacedName{Name: ijName, Namespace: ns}, iJob); err != nil {
		logAndRespond(w, log, http.StatusBadRequest, fmt.Sprintf("req: %s, cannot get IntegrationJob %s/%s", reqID, ns, ijName),
			fmt.Sprintf("Bad request for path, path: %s", r.RequestURI))
		return
	}

	// Redirect if it's enabled
	if configs.ReportRedirectURITemplate != "" {
		tmpl := template.New("")
		tmpl, err := tmpl.Parse(configs.ReportRedirectURITemplate)
		if err != nil {
			logAndRespond(w, log, http.StatusBadRequest, fmt.Sprintf("req: %s, cannot parse report redirection uri template", reqID),
				"Cannot parse report redirection uri template")
			return
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, iJob); err != nil {
			logAndRespond(w, log, http.StatusBadRequest, fmt.Sprintf("req: %s, cannot execute report redirection uri template", reqID),
				"Cannot execute report redirection uri template")
			return
		}

		// Redirect
		http.Redirect(w, r, buf.String(), http.StatusMovedPermanently)
		return
	}

	// Get Job Status
	jobStatus := h.getJobStatus(iJob, job)
	if jobStatus == nil {
		logAndRespond(w, log, http.StatusBadRequest,
			fmt.Sprintf("req: %s, there is no job status %s in IntegrationJob %s/%s", reqID, job, ns, ijName),
			fmt.Sprintf("Bad request for job, ns: %s, job: %s, jobJob: %s", ns, ijName, job))
		return
	}

	// Get Job-Job Log
	podLog, err := h.getPodLogs(jobStatus.PodName, ns, log)
	if err != nil {
		podLog = errorLogNotExist
	}

	// Get template
	tmpl, err := h.getTemplate()
	if err != nil {
		logAndRespond(w, log, http.StatusBadRequest, fmt.Sprintf("req: %s, cannot parse report template", reqID),
			"Cannot parse report template, err: "+err.Error())
		return
	}

	// Publish report
	// Maybe template.Execute writes header before returning 400...? so we do NOT call tmpl.Execute(w, ...) directly
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, report{JobName: ijName, JobJobName: job, JobStatus: jobStatus, Log: podLog}); err != nil {
		logAndRespond(w, log, http.StatusBadRequest, fmt.Sprintf("req: %s, cannot execute report template", reqID),
			"Cannot execute report template, err: "+err.Error())
		return
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		logAndRespond(w, log, http.StatusBadRequest, fmt.Sprintf("req: %s, cannot write result", reqID),
			"Cannot write result, err: "+err.Error())
		return
	}
}

func (h *reportHandler) getJobStatus(ij *cicdv1.IntegrationJob, job string) *cicdv1.JobStatus {
	for _, j := range ij.Status.Jobs {
		if j.Name == job {
			return &j
		}
	}
	return nil
}

func (h *reportHandler) getTemplate() (*template.Template, error) {
	templatePath := os.Getenv(templateConfigMapPathEnv)
	if templatePath == "" {
		templatePath = templateConfigMapPathDefault
	}

	return template.ParseFiles(path.Join(templatePath, templateConfigMapKey))
}

// +kubebuilder:rbac:groups="",resources=pods;pods/log,verbs=get;list;watch

func (h *reportHandler) getPodLogs(podName, namespace string, log logr.Logger) (string, error) {
	var logBuf bytes.Buffer

	if len(podName) == 0 || len(namespace) == 0 {
		return "", fmt.Errorf("podName and namespace should not be empty")
	}

	pod := &corev1.Pod{}
	if err := h.k8sClient.Get(context.Background(), types.NamespacedName{Name: podName, Namespace: namespace}, pod); err != nil {
		return "", err
	}

	for _, c := range pod.Spec.Containers {
		logBuf.WriteString("# Step : " + c.Name + "\n")
		l, err := h.getPodLog(podName, namespace, c.Name)
		if err != nil {
			log.Info(err.Error())
		}
		logBuf.WriteString(l + "\n\n")
	}

	return logBuf.String(), nil
}

func (h *reportHandler) getPodLog(podName, namespace, container string) (string, error) {
	podReq := h.podsGetter.Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{Container: container})
	podLogs, err := podReq.Stream(context.Background())
	if err != nil {
		return "", err
	}
	defer func() {
		_ = podLogs.Close()
	}()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
