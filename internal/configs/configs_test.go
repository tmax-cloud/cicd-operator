package configs

import (
	"github.com/bmizerany/assert"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestGetVars(t *testing.T) {
	configMapData := map[string]string{
		"testInt1":    "1",
		"testInt2":    "12",
		"testString1": "tetest",
		"testString2": "tetest2",
		"testBool1":   "true",
		"testBool2":   "false",
	}

	var testInt1, testInt2, testInt3 int
	var testString1, testString2, testString3 string
	var testBool1, testBool2, testBool3 bool

	testConfig := map[string]operatorConfig{
		"testInt1":    {Type: cfgTypeInt, IntVal: &testInt1, IntDefault: 11},
		"testInt2":    {Type: cfgTypeInt, IntVal: &testInt2, IntDefault: 11},
		"testInt3":    {Type: cfgTypeInt, IntVal: &testInt3, IntDefault: 11},
		"testString1": {Type: cfgTypeString, StringVal: &testString1, StringDefault: "def"},
		"testString2": {Type: cfgTypeString, StringVal: &testString2, StringDefault: "def"},
		"testString3": {Type: cfgTypeString, StringVal: &testString3, StringDefault: "def"},
		"testBool1":   {Type: cfgTypeBool, BoolVal: &testBool1, BoolDefault: true},
		"testBool2":   {Type: cfgTypeBool, BoolVal: &testBool2, BoolDefault: true},
		"testBool3":   {Type: cfgTypeBool, BoolVal: &testBool3, BoolDefault: true},
	}

	getVars(configMapData, testConfig)

	assert.Equal(t, 1, testInt1, "testInt1")
	assert.Equal(t, 12, testInt2, "testInt2")
	assert.Equal(t, 11, testInt3, "testInt3")

	assert.Equal(t, "tetest", testString1, "testString1")
	assert.Equal(t, "tetest2", testString2, "testString2")
	assert.Equal(t, "def", testString3, "testString3")

	assert.Equal(t, true, testBool1, "testBool1")
	assert.Equal(t, false, testBool2, "testBool2")
	assert.Equal(t, true, testBool3, "testBool3")
}

type controllerTestCase struct {
	ConfigMap  *corev1.ConfigMap
	AssertFunc func(*testing.T, error)
}
