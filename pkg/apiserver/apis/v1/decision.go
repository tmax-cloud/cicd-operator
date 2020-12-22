package v1

import (
	"context"
	"encoding/json"
	"fmt"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/types"

	"github.com/tmax-cloud/cicd-operator/internal/wrapper"
)

// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="authorization.k8s.io",resources=subjectaccessreviews,verbs=get;list;watch;create;update;patch

func AddApproveApis(parent *wrapper.RouterWrapper, cli client.Client) error {
	approveWrapper := wrapper.New("/approve", []string{http.MethodPut}, approveHandler)
	if err := parent.Add(approveWrapper); err != nil {
		return err
	}

	k8sCliLock.Lock()
	defer k8sCliLock.Unlock()
	if k8sClient == nil {
		k8sClient = cli
	}

	return nil
}

func AddRejectApis(parent *wrapper.RouterWrapper, cli client.Client) error {
	approveWrapper := wrapper.New("/reject", []string{http.MethodPut}, rejectHandler)
	if err := parent.Add(approveWrapper); err != nil {
		return err
	}

	k8sCliLock.Lock()
	defer k8sCliLock.Unlock()
	if k8sClient == nil {
		k8sClient = cli
	}

	return nil
}

func approveHandler(w http.ResponseWriter, req *http.Request) {
	updateDecision(w, req, cicdv1.ApprovalResultApproved)
}

func rejectHandler(w http.ResponseWriter, req *http.Request) {
	updateDecision(w, req, cicdv1.ApprovalResultRejected)
}

func updateDecision(w http.ResponseWriter, req *http.Request, decision cicdv1.ApprovalResult) {
	// Get ns/approvalName
	vars := mux.Vars(req)

	ns, nsExist := vars["namespace"]
	approvalName, nameExist := vars["approvalName"]
	if !nsExist || !nameExist {
		_ = utils.RespondError(w, http.StatusBadRequest, "url is malformed")
		return
	}

	// Get decision reason
	userReq := &reqBody{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(userReq); err != nil {
		_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("body is not in json form or is malformed, err : %s", err.Error()))
		return
	}

	// Get user
	user, err := getUserName(req.Header)
	if err != nil {
		_ = utils.RespondError(w, http.StatusUnauthorized, fmt.Sprintf("forbidden user, err : %s", err.Error()))
		return
	}

	// Get corresponding Approval object
	if k8sClient == nil {
		msg := "k8sClient is not ready"
		log.Error(fmt.Errorf(msg), "")
		_ = utils.RespondError(w, http.StatusInternalServerError, msg)
		return
	}

	approval := &cicdv1.Approval{}
	if err := k8sClient.Get(context.Background(), types.NamespacedName{Name: approvalName, Namespace: ns}, approval); err != nil {
		_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("no Approval %s/%s is found", ns, approvalName))
		return
	}

	// If Approval is already in approved/rejected status, respond with error
	if approval.Status.Result == cicdv1.ApprovalResultApproved || approval.Status.Result == cicdv1.ApprovalResultRejected {
		_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("approval %s/%s is already in %s status", ns, approvalName, approval.Status.Result))
		return
	}

	// Check if the user is in the approver list
	approvers := approval.Spec.Users

	isApprover := false
	for _, a := range approvers {
		token := strings.Split(a, "=")
		if token[0] == user {
			isApprover = true
			break
		}
	}

	if !isApprover {
		_ = utils.RespondError(w, http.StatusUnauthorized, fmt.Sprintf("approval %s/%s is not requested to you", ns, approvalName))
		return
	}

	defer func() {
		if err := k8sClient.Status().Update(context.Background(), approval); err != nil {
			log.Error(err, "")
		}
	}()

	// Update status
	approval.Status.Result = decision
	approval.Status.Reason = userReq.Reason
	approval.Status.Approver = user
	approval.Status.DecisionTime = &metav1.Time{Time: time.Now()}

	_ = utils.RespondJSON(w, struct{}{})
}
