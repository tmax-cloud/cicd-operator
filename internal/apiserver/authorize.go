package apiserver

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strings"

	authorizationv1 "k8s.io/api/authorization/v1"
	authorization "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

// Authorizer authorizes an api request
type Authorizer interface {
	Authorize(h http.Handler) http.Handler
}

type authorizer struct {
	AuthCli *authorization.AuthorizationV1Client

	APIGroup   string
	APIVersion string
	Verb       string

	log logr.Logger
}

// NewAuthorizer instantiates a new authorizer
func NewAuthorizer(cli *authorization.AuthorizationV1Client, apiGroup, apiVersion, verb string) Authorizer {
	return &authorizer{
		AuthCli:    cli,
		APIGroup:   apiGroup,
		APIVersion: apiVersion,
		Verb:       verb,
		log:        logf.Log.WithName("authorizer"),
	}
}

func (a *authorizer) Authorize(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.TLS == nil || len(req.TLS.PeerCertificates) == 0 {
			msg := "is not https or there is no peer certificate"
			a.log.Info(msg)
			_ = utils.RespondError(w, http.StatusUnauthorized, msg)
			return
		}

		if err := a.reviewAccess(req); err != nil {
			a.log.Info(err.Error())
			_ = utils.RespondError(w, http.StatusForbidden, err.Error())
			return
		}

		h.ServeHTTP(w, req)
	})
}

func (a *authorizer) reviewAccess(req *http.Request) error {
	userName, err := GetUserName(req.Header)
	if err != nil {
		return err
	}

	userGroups, err := GetUserGroup(req.Header)
	if err != nil {
		return err
	}

	userExtras := GetUserExtras(req.Header)

	// URL : /apis/<ApiGroup>/<ApiVersion>/namespaces/<Namespace>/<Resource>/<ResourceName>/<SubResource>
	subPaths := strings.Split(req.URL.Path, "/")
	if len(subPaths) != 9 {
		return fmt.Errorf("URL should be in form of '/apis/<ApiGroup>/<ApiVersion>/namespaces/<Namespace>/<Resource>/<ResourceName>/<SubResource>'")
	}
	ns := subPaths[5]
	resourceType := subPaths[6]
	resourceName := subPaths[7]
	subResource := subPaths[8]

	r := &authorizationv1.SubjectAccessReview{
		Spec: authorizationv1.SubjectAccessReviewSpec{
			User:   userName,
			Groups: userGroups,
			Extra:  userExtras,
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Name:        resourceName,
				Namespace:   ns,
				Group:       a.APIGroup,
				Version:     a.APIVersion,
				Resource:    resourceType,
				Subresource: subResource,
				Verb:        a.Verb,
			},
		},
	}

	result, err := a.AuthCli.SubjectAccessReviews().Create(context.Background(), r, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	if result.Status.Allowed {
		return nil
	}

	return fmt.Errorf(result.Status.Reason)
}
