package customs

import (
	"context"
	"github.com/go-logr/logr"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/notification/slack"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/apis"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// SlackRunHandler handles slack custom task
type SlackRunHandler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// Handle sends slack to the webhook
func (a *SlackRunHandler) Handle(run *tektonv1alpha1.Run) (ctrl.Result, error) {
	ctx := context.Background()
	log := a.Log.WithValues("SlackRun", run.Name)

	original := run.DeepCopy()

	// New Condition default
	cond := run.Status.GetCondition(apis.ConditionSucceeded)
	if cond == nil {
		cond = &apis.Condition{
			Type:    apis.ConditionSucceeded,
			Status:  corev1.ConditionUnknown,
			Reason:  "Awaiting",
			Message: "Sending slack message",
		}
	}

	defer func() {
		run.Status.SetCondition(cond)
		p := client.MergeFrom(original)
		if err := a.Client.Status().Patch(ctx, run, p); err != nil {
			log.Error(err, "")
		}
	}()

	oldCond := run.Status.GetCondition(apis.ConditionSucceeded)
	// Ignore if it's completed (has completionTime)
	if run.Status.CompletionTime != nil || (oldCond != nil && oldCond.Status != corev1.ConditionUnknown) {
		return ctrl.Result{}, nil
	}

	// Set startTime
	if run.Status.StartTime == nil {
		run.Status.StartTime = &metav1.Time{Time: time.Now()}
	}

	// Send slack
	url, _, err := searchParam(run.Spec.Params, cicdv1.CustomTaskSlackParamKeyWebhook, tektonv1beta1.ParamTypeString)
	if err != nil {
		log.Error(err, "")
		cond.Status = corev1.ConditionFalse
		cond.Reason = "InsufficientParams"
		cond.Message = err.Error()
		return ctrl.Result{}, nil
	}

	message, _, err := searchParam(run.Spec.Params, cicdv1.CustomTaskSlackParamKeyMessage, tektonv1beta1.ParamTypeString)
	if err != nil {
		log.Error(err, "")
		cond.Status = corev1.ConditionFalse
		cond.Reason = "InsufficientParams"
		cond.Message = err.Error()
		return ctrl.Result{}, nil
	}

	ij, _, _ := searchParam(run.Spec.Params, cicdv1.CustomTaskSlackParamKeyIntegrationJob, tektonv1beta1.ParamTypeString)
	job, _, _ := searchParam(run.Spec.Params, cicdv1.CustomTaskSlackParamKeyIntegrationJobJob, tektonv1beta1.ParamTypeString)

	// Substitute variables
	compiledMessage, err := compileString(run.Namespace, ij, job, message, a.Client)
	if err != nil {
		log.Error(err, "")
		cond.Status = corev1.ConditionFalse
		cond.Reason = "CannotCompileMessage"
		cond.Message = err.Error()
		return ctrl.Result{}, nil
	}

	// Send!
	if err := slack.SendMessage(url, compiledMessage); err != nil {
		log.Error(err, "")
		cond.Status = corev1.ConditionFalse
		cond.Reason = "EmailError"
		cond.Message = err.Error()
	} else {
		cond.Status = corev1.ConditionTrue
		cond.Reason = "SentSlack"
		cond.Message = ""
	}

	run.Status.CompletionTime = &metav1.Time{Time: time.Now()}

	return ctrl.Result{}, nil
}
