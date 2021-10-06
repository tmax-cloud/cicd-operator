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

package approvals

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/apiserver"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/events"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="authorization.k8s.io",resources=subjectaccessreviews,verbs=get;list;watch;create;update;patch

func (h *handler) approveHandler(w http.ResponseWriter, req *http.Request) {
	h.updateDecision(w, req, cicdv1.ApprovalResultApproved)
}

func (h *handler) rejectHandler(w http.ResponseWriter, req *http.Request) {
	h.updateDecision(w, req, cicdv1.ApprovalResultRejected)
}

func (h *handler) updateDecision(w http.ResponseWriter, req *http.Request, decision cicdv1.ApprovalResult) {
	reqID := utils.RandomString(10)
	log := h.log.WithValues("request", reqID)

	// Get ns/approvalName
	vars := mux.Vars(req)

	ns, nsExist := vars[apiserver.NamespaceParamKey]
	approvalName, nameExist := vars[approvalNameParamKey]
	if !nsExist || !nameExist {
		_ = utils.RespondError(w, http.StatusBadRequest, "url is malformed")
		return
	}

	// Get decision reason
	userReq := &cicdv1.ApprovalAPIReqBody{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(userReq); err != nil {
		log.Info(err.Error())
		_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("req: %s, body is not in json form or is malformed, err : %s", reqID, err.Error()))
		return
	}

	// Get user
	user, err := apiserver.GetUserName(req.Header)
	if err != nil {
		log.Info(err.Error())
		_ = utils.RespondError(w, http.StatusUnauthorized, fmt.Sprintf("req: %s, forbidden user, err : %s", reqID, err.Error()))
		return
	}

	// Get corresponding Approval object
	approval := &cicdv1.Approval{}
	if err := h.k8sClient.Get(context.Background(), types.NamespacedName{Name: approvalName, Namespace: ns}, approval); err != nil {
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

		// Emit event
		_ = events.Emit(h.k8sClient, approval, corev1.EventTypeWarning, "ApproveNotAllowed", fmt.Sprintf("User: %s", user))

		_ = utils.RespondError(w, http.StatusForbidden, fmt.Sprintf("req: %s, approval %s/%s is not requested to you", reqID, ns, approvalName))
		return
	}

	defer func() {
		p := client.MergeFrom(original)
		if err := h.k8sClient.Status().Patch(context.Background(), approval, p); err != nil {
			log.Error(err, "")
		}
	}()

	// Update status
	approval.Status.Result = decision
	approval.Status.Reason = userReq.Reason
	approval.Status.Approver = user
	approval.Status.DecisionTime = &metav1.Time{Time: time.Now()}

	// Emit event
	_ = events.Emit(h.k8sClient, approval, corev1.EventTypeNormal, string(decision), fmt.Sprintf("User: %s, Reason: %s", user, userReq.Reason))

	_ = utils.RespondJSON(w, struct{}{})
}
