package v1

import (
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/tmax-cloud/cicd-operator/internal/wrapper"
)

func AddApprovalApis(parent *wrapper.RouterWrapper, cli client.Client) error {
	approvalWrapper := wrapper.New(fmt.Sprintf("/%s/{approvalName}", ApprovalKind), nil, nil)
	if err := parent.Add(approvalWrapper); err != nil {
		return err
	}

	approvalWrapper.Router.Use(Authorize)

	if err := AddApproveApis(approvalWrapper, cli); err != nil {
		return err
	}
	if err := AddRejectApis(approvalWrapper, cli); err != nil {
		return err
	}
	return nil
}
