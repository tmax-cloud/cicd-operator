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
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	authorization "k8s.io/api/authorization/v1"
)

func TestGetUserName(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		u, err := GetUserName(http.Header{userHeader: []string{"user1"}})
		require.NoError(t, err)
		require.Equal(t, "user1", u)
	})

	t.Run("err", func(t *testing.T) {
		_, err := GetUserName(http.Header{})
		require.Error(t, err)
	})
}

func TestGetUserGroups(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		u, err := GetUserGroups(http.Header{groupHeader: []string{"group1"}})
		require.NoError(t, err)
		require.Equal(t, []string{"group1"}, u)
	})

	t.Run("err", func(t *testing.T) {
		_, err := GetUserGroups(http.Header{})
		require.Error(t, err)
	})
}

func TestGetUserExtras(t *testing.T) {
	extras := GetUserExtras(http.Header{extrasHeader + "extra1": []string{"val1"}, extrasHeader + "extra2": []string{"val2"}, "extra3": []string{"val3"}})
	require.Equal(t, map[string]authorization.ExtraValue{
		"extra1": []string{"val1"},
		"extra2": []string{"val2"},
	}, extras)
}
