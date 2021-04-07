package apiserver

import (
	"fmt"
	authorization "k8s.io/api/authorization/v1"
	"net/http"
	"strings"
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

// GetUserGroup extracts user group from the header
func GetUserGroup(header http.Header) ([]string, error) {
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
