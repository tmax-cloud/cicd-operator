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

import "strings"
import "github.com/sourcegraph/go-diff/diff"

// GetChangedLinesFromDiff parses a diffString and returns the number of changed lines (additions / deletions)
func GetChangedLinesFromDiff(diffString string) (int, int, error) {
	var hunks []*diff.Hunk
	var err error

	// diffString can be either a hunk (without a file header) or a multi-file diff
	if strings.HasPrefix(diffString, "@@") {
		hunks, err = diff.ParseHunks([]byte(diffString))
		if err != nil {
			return 0, 0, err
		}
	} else {
		files, err := diff.ParseMultiFileDiff([]byte(diffString))
		if err != nil {
			return 0, 0, err
		}
		for _, f := range files {
			hunks = append(hunks, f.Hunks...)
		}
	}

	added := 0
	deleted := 0
	for _, h := range hunks {
		added += int(h.Stat().Added) + int(h.Stat().Changed)
		deleted += int(h.Stat().Deleted) + int(h.Stat().Changed)
	}

	return added, deleted, nil
}
