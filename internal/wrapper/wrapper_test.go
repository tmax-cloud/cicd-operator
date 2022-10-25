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

package wrapper

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	w := New("/p", []string{"method1", "method2"}, func(w http.ResponseWriter, req *http.Request) {})
	require.Equal(t, "/p", w.subPath)
	require.Equal(t, []string{"method1", "method2"}, w.methods)
	require.NotNil(t, w.handler)
}

func Test_wrapper_Router(t *testing.T) {
	w := &wrapper{}
	require.Nil(t, w.Router())
	w.router = mux.NewRouter()
	require.NotNil(t, w.Router())
}

func Test_wrapper_SetRouter(t *testing.T) {
	w := &wrapper{}
	require.Nil(t, w.router)
	w.SetRouter(mux.NewRouter())
	require.NotNil(t, w.router)
}

func Test_wrapper_Children(t *testing.T) {
	w := &wrapper{}
	require.Nil(t, w.Children())
	w.children = append(w.children, &wrapper{})
	require.Len(t, w.Children(), 1)
}

func Test_wrapper_Parent(t *testing.T) {
	w := &wrapper{}
	require.Nil(t, w.Parent())
	w.parent = &wrapper{}
	require.NotNil(t, w.Parent())
}

func Test_wrapper_SetParent(t *testing.T) {
	w := &wrapper{}
	require.Nil(t, w.parent)
	w.SetParent(&wrapper{})
	require.NotNil(t, w.parent)
}

func Test_wrapper_Handler(t *testing.T) {
	w := &wrapper{}
	require.Nil(t, w.Handler())
	w.handler = func(w http.ResponseWriter, req *http.Request) {}
	require.NotNil(t, w.Handler())
}

func Test_wrapper_SubPath(t *testing.T) {
	w := &wrapper{}
	require.Empty(t, w.SubPath())
	w.subPath = "/test"
	require.Equal(t, "/test", w.SubPath())
}

func Test_wrapper_Methods(t *testing.T) {
	w := &wrapper{}
	require.Empty(t, w.Methods())
	w.methods = []string{"method1"}
	require.Equal(t, []string{"method1"}, w.Methods())
}

func Test_wrapper_Add(t *testing.T) {
	tc := map[string]struct {
		parent *wrapper
		child  *wrapper

		errorOccurs  bool
		errorMessage string
	}{
		"normal": {
			parent: &wrapper{router: mux.NewRouter()},
			child:  &wrapper{subPath: "/child", methods: []string{"method1", "method2"}},
		},
		"noMethods": {
			parent: &wrapper{router: mux.NewRouter()},
			child:  &wrapper{subPath: "/child"},
		},
		"nilChild": {
			parent:       &wrapper{router: mux.NewRouter()},
			errorOccurs:  true,
			errorMessage: "child is nil",
		},
		"hasParent": {
			parent:       &wrapper{router: mux.NewRouter()},
			child:        &wrapper{subPath: "/child", methods: []string{"method1", "method2"}, parent: &wrapper{}},
			errorOccurs:  true,
			errorMessage: "child already has parent",
		},
		"invalidSubPath": {
			parent:       &wrapper{router: mux.NewRouter()},
			child:        &wrapper{subPath: "child", methods: []string{"method1", "method2"}},
			errorOccurs:  true,
			errorMessage: "child subpath is not valid",
		},
		"nilRouter": {
			parent:       &wrapper{},
			child:        &wrapper{subPath: "/child", methods: []string{"method1", "method2"}},
			errorOccurs:  true,
			errorMessage: "parent does not have a router",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			err := c.parent.Add(c.child)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)

			}
		})
	}
}

func Test_wrapper_FullPath(t *testing.T) {
	root := &wrapper{subPath: "/api"}
	child := &wrapper{subPath: "/v1", parent: root}
	require.Equal(t, "/api/v1", child.FullPath())
}
