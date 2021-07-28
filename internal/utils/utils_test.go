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

package utils

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestParseApproversList(t *testing.T) {
	// Success test
	str := `admin@tmax.co.kr=admin@tmax.co.kr,test@tmax.co.kr
test2@tmax.co.kr=test2@tmax.co.kr`
	list, err := ParseApproversList(str)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 3, len(list), "list is not parsed well")
	assert.Equal(t, "admin@tmax.co.kr=admin@tmax.co.kr", list[0], "list is not parsed well")
	assert.Equal(t, "test@tmax.co.kr", list[1], "list is not parsed well")
	assert.Equal(t, "test2@tmax.co.kr=test2@tmax.co.kr", list[2], "list is not parsed well")

	// Fail test
	str = "admin,,ttt"
	list, err = ParseApproversList(str)
	if err == nil {
		for i, l := range list {
			t.Logf("%d : %s", i, l)
		}
		t.Fatal("error not occur")
	}
}

func TestParseEmailFromUsers(t *testing.T) {
	// Include test
	users := []string{
		"aweilfjlwesfj",
		"aweilfjlwesfj=aweiojweio",
		"aweilfjlwesfj=admin@tmax.co.kr",
		"asdij@oisdjf.sdfioj=test@tmax.co.kr",
	}

	tos := ParseEmailFromUsers(users)

	assert.Equal(t, 2, len(tos), "list is not parsed well")
	assert.Equal(t, "admin@tmax.co.kr", tos[0], "list is not parsed well")
	assert.Equal(t, "test@tmax.co.kr", tos[1], "list is not parsed well")
}
