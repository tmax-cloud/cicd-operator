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

package blocker

import (
	"encoding/json"
	"fmt"
	"github.com/bmizerany/assert"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"net/http"
	"net/http/httptest"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
)

func TestBlocker_statusServer(t *testing.T) {
	cli := statusServerTestConfig()

	b := New(cli)
	b.Pools["api.github.com/tmax-cloud/cicd-operator"] = NewPRPool(testICNamespace, testICName)

	srv := httptest.NewServer(b.newRouter())

	// TEST 1
	resp, err := http.Get(fmt.Sprintf("%s/status", srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	resultBytes, _ := ioutil.ReadAll(resp.Body)
	var result []statusListEntity
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, resp.StatusCode, "Successful request")
	assert.Equal(t, 1, len(result), "One result pool")

	// TEST 2
	resp, err = http.Get(fmt.Sprintf("%s/status/api.github.com/tmax-cloud/cicd-operator", srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	resultBytes, _ = ioutil.ReadAll(resp.Body)

	t.Log(string(resultBytes))
	assert.Equal(t, 200, resp.StatusCode, "Successful request")
}

func statusServerTestConfig() client.Client {
	if _, exist := os.LookupEnv("CI"); !exist {
		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	}

	s := runtime.NewScheme()
	utilruntime.Must(cicdv1.AddToScheme(s))
	return fake.NewClientBuilder().WithScheme(s).Build()
}
