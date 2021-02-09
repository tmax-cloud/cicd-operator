package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"html/template"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
)

var reportPath = fmt.Sprintf("/report/{%s}/{%s}/{%s}", paramKeyNamespace, paramKeyJobName, paramKeyJobJobName)

const (
	templateConfigMapName = "report-template"
	templateConfigMapKey  = "template"

	errorLogNotExist = "log does not exist... maybe the pod does not exist"
)

type report struct {
	JobName    string
	JobJobName string
	JobStatus  *cicdv1.JobStatus
	Log        string
}

type reportHandler struct {
	k8sClient client.Client
	clientSet *kubernetes.Clientset
}

func (h *reportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqID := utils.RandomString(10)
	log := logger.WithValues("request", reqID)

	vars := mux.Vars(r)

	ns, nsExist := vars[paramKeyNamespace]
	jobName, jobNameExist := vars[paramKeyJobName]
	jobJobName, jobJobNameExist := vars[paramKeyJobJobName]

	if !nsExist || !jobNameExist || !jobJobNameExist {
		logAndRespond(w, log, http.StatusBadRequest, fmt.Sprintf("req: %s, path is not in form of '%s'", reqID, reportPath),
			fmt.Sprintf("Bad request for path, path: %s", r.RequestURI))
		return
	}

	iJob := &cicdv1.IntegrationJob{}
	if err := h.k8sClient.Get(context.Background(), types.NamespacedName{Name: jobName, Namespace: ns}, iJob); err != nil {
		logAndRespond(w, log, http.StatusBadRequest, fmt.Sprintf("req: %s, cannot get IntegrationJob %s/%s", reqID, ns, jobName),
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
	var jobStatus *cicdv1.JobStatus
	for _, j := range iJob.Status.Jobs {
		if j.Name == jobJobName {
			jobStatus = &j
			break
		}
	}
	if jobStatus == nil {
		logAndRespond(w, log, http.StatusBadRequest,
			fmt.Sprintf("req: %s, there is no job status %s in IntegrationJob %s/%s", reqID, jobJobName, ns, jobName),
			fmt.Sprintf("Bad request for job, ns: %s, job: %s, jobJob: %s", ns, jobName, jobJobName))
		return
	}

	// Get Job-Job Log
	podLog, err := h.getPodLogs(jobStatus.PodName, ns, log)
	if err != nil {
		podLog = errorLogNotExist
	}

	// Get template
	templateStr, err := h.getTemplateString()
	if err != nil {
		logAndRespond(w, log, http.StatusBadRequest, fmt.Sprintf("req: %s, cannot get report template", reqID),
			"Cannot get report template")
		return
	}

	tmpl := template.New("")
	tmpl, err = tmpl.Parse(templateStr)
	if err != nil {
		logAndRespond(w, log, http.StatusBadRequest, fmt.Sprintf("req: %s, cannot parse report template", reqID),
			"Cannot parse report template")
		return
	}

	// Publish report
	if err := tmpl.Execute(w, report{JobName: jobName, JobJobName: jobJobName, JobStatus: jobStatus, Log: podLog}); err != nil {
		logAndRespond(w, log, http.StatusBadRequest, fmt.Sprintf("req: %s, cannot execute report template", reqID),
			"Cannot execute report template")
		return
	}
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
	podReq := h.clientSet.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{Container: container})
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

func (h *reportHandler) getTemplateString() (string, error) {
	ns, err := utils.Namespace()
	if err != nil {
		return "", err
	}
	cm := &corev1.ConfigMap{}
	if err := h.k8sClient.Get(context.Background(), types.NamespacedName{Name: templateConfigMapName, Namespace: ns}, cm); err != nil {
		return "", err
	}

	templateString, templateFound := cm.Data[templateConfigMapKey]
	if !templateFound {
		return "", err
	}

	return templateString, nil
}
