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

package gitlab

import (
	"encoding/json"
	"github.com/bmizerany/assert"
	"testing"
	"time"
)

type testTimeStruct struct {
	Time *gitlabTime `json:"time"`
}

func TestGitlabTime_UnmarshalJSON(t *testing.T) {
	buf := []byte("{\"time\": \"2021-02-03 07:13:23 UTC\"}")

	tt := &testTimeStruct{}
	if err := json.Unmarshal(buf, tt); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "2021-02-03 07:13:23 +0000 UTC", tt.Time.UTC().String())
}

func TestGitlabTime_MarshalJSON(t *testing.T) {
	targetTime, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2021-02-03 07:13:23 +0000 UTC")
	if err != nil {
		t.Fatal(err)
	}
	tt := &testTimeStruct{
		Time: &gitlabTime{
			Time: targetTime,
		},
	}

	result, err := json.Marshal(tt)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "{\"time\":\"2021-02-03 07:13:23 UTC\"}", string(result))
}
