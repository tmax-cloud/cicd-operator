package customs

import (
	"context"
	"github.com/go-logr/logr"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/mail"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/apis"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type EmailRunHandler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (a *EmailRunHandler) Handle(run *tektonv1alpha1.Run) (ctrl.Result, error) {
	ctx := context.Background()
	log := a.Log.WithValues("EmailRun", run.Namespace)

	original := run.DeepCopy()

	// New Condition default
	cond := run.Status.GetCondition(apis.ConditionSucceeded)
	if cond == nil {
		cond = &apis.Condition{
			Type:    apis.ConditionSucceeded,
			Status:  corev1.ConditionUnknown,
			Reason:  "Waiting",
			Message: "Sending email",
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

	// Send mail
	_, receivers, err := searchParam(run.Spec.Params, cicdv1.CustomTaskEmailParamKeyReceivers, tektonv1beta1.ParamTypeArray)
	if err != nil {
		log.Error(err, "")
		return ctrl.Result{}, err
	}

	title, _, err := searchParam(run.Spec.Params, cicdv1.CustomTaskEmailParamKeyTitle, tektonv1beta1.ParamTypeString)
	if err != nil {
		log.Error(err, "")
		return ctrl.Result{}, err
	}

	content, _, err := searchParam(run.Spec.Params, cicdv1.CustomTaskEmailParamKeyContent, tektonv1beta1.ParamTypeString)
	if err != nil {
		log.Error(err, "")
		return ctrl.Result{}, err
	}

	isHtmlStr, _, _ := searchParam(run.Spec.Params, cicdv1.CustomTaskEmailParamKeyIsHtml, tektonv1beta1.ParamTypeString)

	isHtml := false
	if isHtmlStr == "true" {
		isHtml = true
	}

	if err := mail.Send(receivers, title, content, isHtml, a.Client); err != nil {
		log.Error(err, "")
		cond.Status = corev1.ConditionFalse
		cond.Reason = "EmailError"
		cond.Message = err.Error()
	} else {
		cond.Status = corev1.ConditionTrue
		cond.Reason = "SentMail"
		cond.Message = ""
	}

	run.Status.CompletionTime = &metav1.Time{Time: time.Now()}

	return ctrl.Result{}, nil
}
