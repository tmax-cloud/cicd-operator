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

// addApproveApis adds approve api
func addApproveApis(parent *wrapper.RouterWrapper, cli client.Client) error {
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

// addRejectApis adds reject api
func addRejectApis(parent *wrapper.RouterWrapper, cli client.Client) error {
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
	reqID := utils.RandomString(10)
	log := logger.WithValues("request", reqID)

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
		log.Info(err.Error())
		_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("req: %s, body is not in json form or is malformed, err : %s", reqID, err.Error()))
		return
	}

	// Get user
	user, err := getUserName(req.Header)
	if err != nil {
		log.Info(err.Error())
		_ = utils.RespondError(w, http.StatusUnauthorized, fmt.Sprintf("req: %s, forbidden user, err : %s", reqID, err.Error()))
		return
	}

	// Get corresponding Approval object
	if k8sClient == nil {
		msg := fmt.Errorf("req: %s, k8sClient is not ready", reqID)
		log.Info(msg.Error())
		_ = utils.RespondError(w, http.StatusInternalServerError, msg.Error())
		return
	}

	approval := &cicdv1.Approval{}
	if err := k8sClient.Get(context.Background(), types.NamespacedName{Name: approvalName, Namespace: ns}, approval); err != nil {
		log.Info(err.Error())
		_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("req: %s, no Approval %s/%s is found", reqID, ns, approvalName))
		return
	}
	original := approval.DeepCopy()

	// If Approval is already in approved/rejected status, respond with error
	if approval.Status.Result == cicdv1.ApprovalResultApproved || approval.Status.Result == cicdv1.ApprovalResultRejected {
		log.Info("approval is already decided")
		_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("req: %s, approval %s/%s is already in %s status", reqID, ns, approvalName, approval.Status.Result))
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
		log.Info(fmt.Sprintf("requested user (%s) is not an approver", user))
		_ = utils.RespondError(w, http.StatusForbidden, fmt.Sprintf("req: %s, approval %s/%s is not requested to you", reqID, ns, approvalName))
		return
	}

	defer func() {
		p := client.MergeFrom(original)
		if err := k8sClient.Status().Patch(context.Background(), approval, p); err != nil {
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
