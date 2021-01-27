package v1

import (
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/tmax-cloud/cicd-operator/internal/wrapper"
)

func addApprovalApis(parent *wrapper.RouterWrapper, cli client.Client) error {
	approvalWrapper := wrapper.New(fmt.Sprintf("/%s/{approvalName}", approvalKind), nil, nil)
	if err := parent.Add(approvalWrapper); err != nil {
		return err
	}

	approvalWrapper.Router.Use(authorize)

	if err := addApproveApis(approvalWrapper, cli); err != nil {
		return err
	}
	if err := addRejectApis(approvalWrapper, cli); err != nil {
		return err
	}
	return nil
}
