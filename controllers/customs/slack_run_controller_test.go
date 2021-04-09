package customs

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bmizerany/assert"
	"github.com/gorilla/mux"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/notification/slack"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"knative.dev/pkg/apis"
	"net/http"
	"net/http/httptest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

const (
	testSlackRunName = "test-slack-run"
	testRunNamespace = "default"
)

func TestSlackRunHandler_Handle(t *testing.T) {
	errCh := make(chan error, 10)
	// Launch test slack webhook server
	srv := newTestServer()

	s := runtime.NewScheme()
	utilruntime.Must(cicdv1.AddToScheme(s))
	utilruntime.Must(tektonv1beta1.AddToScheme(s))
	utilruntime.Must(tektonv1alpha1.AddToScheme(s))

	ij := &cicdv1.IntegrationJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ij-1",
			Namespace: testRunNamespace,
		},
		Spec: cicdv1.IntegrationJobSpec{
			Refs: cicdv1.IntegrationJobRefs{
				Base: cicdv1.IntegrationJobRefsBase{
					Ref: cicdv1.GitRef("refs/tags/v0.2.3"),
				},
			},
			Jobs: []cicdv1.Job{{}},
		},
	}
	ij.Spec.Jobs[0].Name = "test-job-1"

	run := &tektonv1alpha1.Run{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testSlackRunName,
			Namespace: testRunNamespace,
		},
		Spec: tektonv1alpha1.RunSpec{
			Ref: &tektonv1alpha1.TaskRef{
				APIVersion: "cicd.tmax.io/v1",
				Kind:       "SlackTask",
			},
			Params: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskSlackParamKeyWebhook, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: srv.URL}},
				{Name: cicdv1.CustomTaskSlackParamKeyMessage, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "$INTEGRATION_JOB_NAME/$JOB_NAME"}},
				{Name: cicdv1.CustomTaskSlackParamKeyIntegrationJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-ij-1"}},
				{Name: cicdv1.CustomTaskSlackParamKeyIntegrationJobJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-job-1"}},
			},
		},
	}

	fakeCli := fake.NewFakeClientWithScheme(s, run, ij)

	slackRunHandler := SlackRunHandler{Scheme: s, Client: fakeCli, Log: ctrl.Log}

	res, err := slackRunHandler.Handle(run)
	assert.Equal(t, true, err == nil)
	assert.Equal(t, ctrl.Result{}, res)

	resRun := &tektonv1alpha1.Run{}
	if err := fakeCli.Get(context.Background(), types.NamespacedName{Name: testSlackRunName, Namespace: testRunNamespace}, resRun); err != nil {
		t.Fatal(err)
	}

	cond := resRun.Status.GetCondition(apis.ConditionSucceeded)
	if cond == nil {
		t.Fatal("cond is nil")
	}

	t.Log(cond.Status)
	t.Log(cond.Reason)
	t.Log(cond.Message)

	assert.Equal(t, corev1.ConditionTrue, cond.Status)
	assert.Equal(t, "SentSlack", cond.Reason)
	assert.Equal(t, "", cond.Message)

	errCh <- nil

	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		} else {
			return
		}
	}
}

func newTestServer() *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			_ = req.Body.Close()
		}()
		userReq := slack.Message{}
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&userReq); err != nil {
			_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("body is not in json form or is malformed, err : %s", err.Error()))
			return
		}
	})

	return httptest.NewServer(router)
}
