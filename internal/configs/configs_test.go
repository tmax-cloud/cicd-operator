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

package configs

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func Test_getVars(t *testing.T) {
	configMapData := map[string]string{
		"testInt1":    "1",
		"testInt2":    "12",
		"testInt5":    "test",
		"testString1": "tetest",
		"testString2": "tetest2",
		"testBool1":   "true",
		"testBool2":   "false",
		"testBool5":   "test",
	}

	var testInt1, testInt2, testInt3, testInt5 int
	var testString1, testString2, testString3 string
	var testBool1, testBool2, testBool3, testBool5 bool

	testConfig := map[string]operatorConfig{
		"testInt1":    {Type: cfgTypeInt, IntVal: &testInt1, IntDefault: 11},
		"testInt2":    {Type: cfgTypeInt, IntVal: &testInt2, IntDefault: 11},
		"testInt3":    {Type: cfgTypeInt, IntVal: &testInt3, IntDefault: 11},
		"testInt4":    {Type: cfgTypeInt},
		"testInt5":    {Type: cfgTypeInt, IntVal: &testInt5},
		"testString1": {Type: cfgTypeString, StringVal: &testString1, StringDefault: "def"},
		"testString2": {Type: cfgTypeString, StringVal: &testString2, StringDefault: "def"},
		"testString3": {Type: cfgTypeString, StringVal: &testString3, StringDefault: "def"},
		"testString4": {Type: cfgTypeString},
		"testBool1":   {Type: cfgTypeBool, BoolVal: &testBool1, BoolDefault: true},
		"testBool2":   {Type: cfgTypeBool, BoolVal: &testBool2, BoolDefault: true},
		"testBool3":   {Type: cfgTypeBool, BoolVal: &testBool3, BoolDefault: true},
		"testBool4":   {Type: cfgTypeBool},
		"testBool5":   {Type: cfgTypeBool, BoolVal: &testBool5},
	}

	getVars(configMapData, testConfig)

	require.Equal(t, 1, testInt1, "testInt1")
	require.Equal(t, 12, testInt2, "testInt2")
	require.Equal(t, 11, testInt3, "testInt3")

	require.Equal(t, "tetest", testString1, "testString1")
	require.Equal(t, "tetest2", testString2, "testString2")
	require.Equal(t, "def", testString3, "testString3")

	require.True(t, testBool1)
	require.False(t, testBool2)
	require.True(t, testBool3)
}

type controllerTestCase struct {
	ConfigMap  *corev1.ConfigMap
	AssertFunc func(*testing.T, error)
}
