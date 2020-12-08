package scheduler

import (
	"fmt"
	"github.com/tmax-cloud/cicd-operator/pkg/scheduler/pool"
	"github.com/tmax-cloud/cicd-operator/pkg/structs"
)

func fifoCompare(_a, _b structs.Item) bool {
	if _a == nil || _b == nil {
		return false
	}
	a, aOk := _a.(*pool.JobNode)
	b, bOk := _b.(*pool.JobNode)
	if !aOk || !bOk {
		return false
	}

	return a.CreationTimestamp.Time.Before(b.CreationTimestamp.Time) || fmt.Sprintf("%s_%s", a.Namespace, a.Name) < fmt.Sprintf("%s_%s", b.Namespace, b.Name)
}
