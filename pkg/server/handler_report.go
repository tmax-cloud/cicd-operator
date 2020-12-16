package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gorilla/mux"
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
	TemplateConfigMapName = "report-template"
	TemplateConfigMapKey  = "template"

	ErrorLogNotExist = "log does not exist... maybe the pod does not exist"
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
	vars := mux.Vars(r)

	ns, nsExist := vars[paramKeyNamespace]
	jobName, jobNameExist := vars[paramKeyJobName]
	jobJobName, jobJobNameExist := vars[paramKeyJobJobName]

	if !nsExist || !jobNameExist || !jobJobNameExist {
		_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("path is not in form of '%s'", reportPath))
		log.Info("Bad request for path", "path", r.RequestURI)
		return
	}

	iJob := &cicdv1.IntegrationJob{}
	if err := h.k8sClient.Get(context.TODO(), types.NamespacedName{Name: jobName, Namespace: ns}, iJob); err != nil {
		_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("cannot get IntegrationJob %s/%s", ns, jobName))
		log.Info("Bad request for path", "path", r.RequestURI)
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
		_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("there is no job status %s in IntegrationJob %s/%s", jobJobName, ns, jobName))
		log.Info("Bad request for job", "ns", ns, "job", jobName, "jobJob", jobJobName)
		return
	}

	// Get Job-Job Log
	podLog, err := h.getPodLogs(jobStatus.PodName, ns)
	if err != nil {
		podLog = ErrorLogNotExist
	}

	// Get template
	templateStr, err := h.getTemplateString()
	if err != nil {
		_ = utils.RespondError(w, http.StatusBadRequest, "cannot get report template")
		log.Info("Cannot get report template")
		return
	}

	tmpl := template.New("")
	tmpl, err = tmpl.Parse(templateStr)
	if err != nil {
		_ = utils.RespondError(w, http.StatusBadRequest, "cannot parse report template")
		log.Info("Cannot parse report template")
		return
	}

	// Publish report
	if err := tmpl.Execute(w, report{JobName: jobName, JobJobName: jobJobName, JobStatus: jobStatus, Log: podLog}); err != nil {
		_ = utils.RespondError(w, http.StatusBadRequest, "cannot execute report template")
		log.Info("Cannot execute report template")
		return
	}
}

func (h *reportHandler) getPodLogs(podName, namespace string) (string, error) {
	var logBuf bytes.Buffer

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
	podLogs, err := podReq.Stream()
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

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
	if err := h.k8sClient.Get(context.TODO(), types.NamespacedName{Name: TemplateConfigMapName, Namespace: ns}, cm); err != nil {
		return "", err
	}

	templateString, templateFound := cm.Data[TemplateConfigMapKey]
	if !templateFound {
		return "", err
	}

	return templateString, nil
}
