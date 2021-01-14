package v1

import (
	"context"
	"fmt"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	authorization "k8s.io/api/authorization/v1"
)

const (
	UserHeader   = "X-Remote-User"
	GroupHeader  = "X-Remote-Group"
	ExtrasHeader = "X-Remote-Extra-"
)

func Authorize(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if err := authorize(req); err != nil {
			_ = utils.RespondError(w, http.StatusUnauthorized, err.Error())
			return
		}

		if err := reviewAccess(req); err != nil {
			_ = utils.RespondError(w, http.StatusForbidden, err.Error())
			return
		}

		h.ServeHTTP(w, req)
	})
}

func authorize(req *http.Request) error {
	if req.TLS == nil || len(req.TLS.PeerCertificates) == 0 {
		return fmt.Errorf("is not https or there is no peer certificate")
	}
	return nil
}

func reviewAccess(req *http.Request) error {
	userName, err := getUserName(req.Header)
	if err != nil {
		return err
	}

	userGroups, err := getUserGroup(req.Header)
	if err != nil {
		return err
	}

	userExtras := getUserExtras(req.Header)

	// URL : /apis/cicdapi.tmax.io/v1/namespaces/default/approvals/test-approval/approve
	subPaths := strings.Split(req.URL.Path, "/")
	if len(subPaths) != 9 {
		return fmt.Errorf("URL should be in form of '/apis/cicdapi.tmax.io/v1/namespaces/<namespace>/approvals/<approval-name>/[approve|reject]'")
	}
	subResource := subPaths[8]

	vars := mux.Vars(req)

	ns, nsExist := vars["namespace"]
	approvalName, nameExist := vars["approvalName"]
	if !nsExist || !nameExist {
		return fmt.Errorf("url is malformed")
	}

	r := &authorization.SubjectAccessReview{
		Spec: authorization.SubjectAccessReviewSpec{
			User:   userName,
			Groups: userGroups,
			Extra:  userExtras,
			ResourceAttributes: &authorization.ResourceAttributes{
				Name:        approvalName,
				Namespace:   ns,
				Group:       ApiGroup,
				Version:     ApiVersion,
				Resource:    ApprovalKind,
				Subresource: subResource,
				Verb:        "update",
			},
		},
	}

	authCliLock.Lock()
	defer authCliLock.Unlock()

	if authClient == nil {
		cli, err := utils.AuthClient()
		if err != nil {
			return err
		}
		authClient = cli
	}

	result, err := authClient.SubjectAccessReviews().Create(context.Background(), r, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	if result.Status.Allowed {
		return nil
	}

	return fmt.Errorf(result.Status.Reason)
}

func getUserName(header http.Header) (string, error) {
	for k, v := range header {
		if k == UserHeader {
			return v[0], nil
		}
	}
	return "", fmt.Errorf("no header %s", UserHeader)
}

func getUserGroup(header http.Header) ([]string, error) {
	for k, v := range header {
		if k == UserHeader {
			return v, nil
		}
	}
	return nil, fmt.Errorf("no header %s", GroupHeader)
}

func getUserExtras(header http.Header) map[string]authorization.ExtraValue {
	extras := map[string]authorization.ExtraValue{}

	for k, v := range header {
		if strings.HasPrefix(k, ExtrasHeader) {
			extras[strings.TrimPrefix(k, ExtrasHeader)] = v
		}
	}

	return extras
}
