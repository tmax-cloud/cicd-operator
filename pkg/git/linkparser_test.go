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
	"github.com/bmizerany/assert"
	"testing"
)

func TestParseLinkHeader(t *testing.T) {
	h := "<https://api.github.com/repositories/319253224/statuses/3196ccc37bcae94852079b04fcbfaf928341d6e9?per_page=100&page=3>; rel=\"prev\", <https://api.github.com/repositories/319253224/statuses/3196ccc37bcae94852079b04fcbfaf928341d6e9?per_page=100&page=1>; rel=\"first\""

	links := ParseLinkHeader(h)
	assert.Equal(t, true, links.Find("next") == nil, "No next rel")

	h = "<https://api.github.com/repositories/319253224/statuses/3196ccc37bcae94852079b04fcbfaf928341d6e9?per_page=100&page=2>; rel=\"prev\", <https://api.github.com/repositories/319253224/statuses/3196ccc37bcae94852079b04fcbfaf928341d6e9?per_page=100&page=4>; rel=\"next\", <https://api.github.com/repositories/319253224/statuses/3196ccc37bcae94852079b04fcbfaf928341d6e9?per_page=100&page=4>; rel=\"last\", <https://api.github.com/repositories/319253224/statuses/3196ccc37bcae94852079b04fcbfaf928341d6e9?per_page=100&page=1>; rel=\"first\""
	links = ParseLinkHeader(h)
	assert.Equal(t, false, links.Find("next") == nil, "Has next rel")
	assert.Equal(t, "https://api.github.com/repositories/319253224/statuses/3196ccc37bcae94852079b04fcbfaf928341d6e9?per_page=100&page=4", links.Find("next").URL, "Next URL")
}
