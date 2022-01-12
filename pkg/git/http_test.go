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

package git

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient_CheckRateLimitGetResetTime(t *testing.T) {
	msg := fmt.Errorf("unixtime::000000000. Rate limit exceeded, code 403. Please increase the limit or wait until reset")
	tm := CheckRateLimitGetResetTime(msg)
	require.Equal(t, tm, 000000000)

	tm = CheckRateLimitGetResetTime(nil)
	require.Equal(t, tm, 0)
}

func TestClient_GetGapTime(t *testing.T) {
	require.Equal(t, time.Now().Unix()*-1, GetGapTime(0))
	require.Equal(t, 10-time.Now().Unix(), GetGapTime(10))
	require.Equal(t, 20-time.Now().Unix(), GetGapTime(20))
}
