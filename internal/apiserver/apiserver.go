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

package apiserver

import (
	"fmt"
	"net/http"
	"strings"

	authorization "k8s.io/api/authorization/v1"
)

const (
	// APIGroup is a common group for api
	APIGroup = "cicdapi.tmax.io"

	// NamespaceParamKey is a common key for namespace var
	NamespaceParamKey = "namespace"

	userHeader   = "X-Remote-User"
	groupHeader  = "X-Remote-Group"
	extrasHeader = "X-Remote-Extra-"
)

// APIHandler is an api handler interface.
// Common functions should be defined, if needed
type APIHandler interface{}

// GetUserName extracts user name from the header
func GetUserName(header http.Header) (string, error) {
	for k, v := range header {
		if k == userHeader {
			return v[0], nil
		}
	}
	return "", fmt.Errorf("no header %s", userHeader)
}

// GetUserGroups extracts user group from the header
func GetUserGroups(header http.Header) ([]string, error) {
	for k, v := range header {
		if k == groupHeader {
			return v, nil
		}
	}
	return nil, fmt.Errorf("no header %s", groupHeader)
}

// GetUserExtras extracts user extras from the header
func GetUserExtras(header http.Header) map[string]authorization.ExtraValue {
	extras := map[string]authorization.ExtraValue{}

	for k, v := range header {
		if strings.HasPrefix(k, extrasHeader) {
			extras[strings.TrimPrefix(k, extrasHeader)] = v
		}
	}

	return extras
}
