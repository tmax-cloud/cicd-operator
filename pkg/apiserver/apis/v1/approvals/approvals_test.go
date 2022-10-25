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
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/tmax-cloud/cicd-operator/internal/test"
	"github.com/tmax-cloud/cicd-operator/internal/wrapper"
)

func TestNewHandler(t *testing.T) {
	w := wrapper.New("/", nil, nil)
	w.SetRouter(mux.NewRouter())

	wNoRouter := wrapper.New("/", nil, nil)

	tc := map[string]struct {
		wrapper wrapper.RouterWrapper

		errorOccurs  bool
		errorMessage string
	}{
		"normal": {
			wrapper: w,
		},
		"approvalErr": {
			wrapper:      wNoRouter,
			errorOccurs:  true,
			errorMessage: "parent does not have a router",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			_, err := NewHandler(c.wrapper, nil, nil, &test.FakeLogger{})
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
