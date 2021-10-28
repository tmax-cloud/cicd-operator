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

package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFakeLogger_Clear(t *testing.T) {
	l := &FakeLogger{
		Infos:     []string{"test", "test1"},
		Errors:    []error{fmt.Errorf("test"), fmt.Errorf("test1")},
		ErrorMsgs: []string{"test", "test1"},
	}
	l.Clear()
	require.Empty(t, l.Infos)
	require.Empty(t, l.Errors)
	require.Empty(t, l.ErrorMsgs)
}

func TestFakeLogger_Enabled(t *testing.T) {
	l := &FakeLogger{}
	require.True(t, l.Enabled())
}

func TestFakeLogger_Info(t *testing.T) {
	l := &FakeLogger{}
	l.Info("test-info")
	require.Len(t, l.Infos, 1)
	require.Equal(t, "test-info", l.Infos[0])
}

func TestFakeLogger_Error(t *testing.T) {
	l := &FakeLogger{}
	l.Error(fmt.Errorf("test-err"), "test-msg")
	require.Len(t, l.Errors, 1)
	require.Len(t, l.ErrorMsgs, 1)
	require.Equal(t, "test-err", l.Errors[0].Error())
	require.Equal(t, "test-msg", l.ErrorMsgs[0])
}

func TestFakeLogger_V(t *testing.T) {
	l := &FakeLogger{}
	require.Equal(t, l, l.V(0))
}

func TestFakeLogger_WithValues(t *testing.T) {
	l := &FakeLogger{}
	require.Equal(t, l, l.WithValues("val"))
}

func TestFakeLogger_WithName(t *testing.T) {
	l := &FakeLogger{}
	require.Equal(t, l, l.WithName("name"))
}
