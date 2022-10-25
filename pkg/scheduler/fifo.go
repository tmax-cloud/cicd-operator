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
