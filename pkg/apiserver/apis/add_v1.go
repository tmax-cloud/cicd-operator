package apis

import v1 "github.com/tmax-cloud/cicd-operator/pkg/apiserver/apis/v1"

func init() {
	AddAPIFuncs = append(AddAPIFuncs, v1.AddV1Apis)
}
