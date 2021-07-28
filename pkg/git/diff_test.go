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
	"github.com/stretchr/testify/require"
	"testing"
)

type diffTestCase struct {
	diffString        string
	expectedAdditions int
	expectedDeletions int
	errorOccurs       bool
	errorMessage      string
}

func TestGetChangedLinesFromDiff(t *testing.T) {
	tc := map[string]diffTestCase{
		"hunk": {
			diffString:        "@@ -1,7 +1,7 @@\n \u003c!DOCTYPE html\u003e\n \u003chtml\u003e\n     \u003chead\u003e\n-        \u003ctitle\u003eTomcatMavenApp\u003c/title\u003e\n+        \u003ctitle\u003eTomcatMavenAppaaaa - add commit3\u003c/title\u003e\n         \u003cmeta http-equiv=\"Content-Type\" content=\"text/html; charset=UTF-8\"\u003e\n     \u003c/head\u003e\n     \u003cbody\u003e\n",
			expectedAdditions: 1,
			expectedDeletions: 1,
		},
		"file": {
			diffString:        "--- /dev/null\n+++ b/LICENSE\n@@ -0,0 +1,21 @@\n+The MIT License (MIT)\n+\n+Copyright (c) 2018 Administrator\n+\n+Permission is hereby granted, free of charge, to any person obtaining a copy\n+of this software and associated documentation files (the \"Software\"), to deal\n+in the Software without restriction, including without limitation the rights\n+to use, copy, modify, merge, publish, distribute, sublicense, and/or sell\n+copies of the Software, and to permit persons to whom the Software is\n+furnished to do so, subject to the following conditions:\n+\n+The above copyright notice and this permission notice shall be included in all\n+copies or substantial portions of the Software.\n+\n+THE SOFTWARE IS PROVIDED \"AS IS\", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR\n+IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,\n+FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE\n+AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER\n+LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,\n+OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE\n+SOFTWARE.\n",
			expectedAdditions: 21,
			expectedDeletions: 0,
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			additions, deletions, err := GetChangedLinesFromDiff(c.diffString)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expectedAdditions, additions)
				require.Equal(t, c.expectedDeletions, deletions)
			}
		})
	}
}
