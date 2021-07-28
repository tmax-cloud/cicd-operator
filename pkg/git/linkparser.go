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

// LinkHeader is a header list for "Link"
type LinkHeader []LinkHeaderEntry

// Find finds a rel from the LinkHeader
func (l LinkHeader) Find(rel string) *LinkHeaderEntry {
	for _, h := range l {
		if h.Rel == rel {
			return &h
		}
	}
	return nil
}

// LinkHeaderEntry is an entry for a LinkHeader
type LinkHeaderEntry struct {
	URL string
	Rel string
}

// ParseLinkHeader parses url, rel from the header "Link"
func ParseLinkHeader(h string) LinkHeader {
	var result LinkHeader

	if h == "" {
		return result
	}

	// Entry
	for _, e := range strings.Split(h, ",") {
		entry := LinkHeaderEntry{}
		// Tokens
		for _, t := range strings.Split(strings.TrimSpace(e), ";") {
			token := strings.TrimSpace(t)

			if token == "" {
				continue
			}

			if token[0] == '<' && token[len(token)-1] == '>' {
				entry.URL = strings.TrimSpace(strings.Trim(token, "<>"))
			}

			var key, val string
			parts := strings.SplitN(token, "=", 2)
			key = parts[0]
			switch len(parts) {
			case 1:
				val = ""
			case 2:
				val = strings.Trim(parts[1], "\"")
			default:
				continue
			}

			if strings.ToLower(key) == "rel" {
				entry.Rel = val
			}
		}
		result = append(result, entry)
	}

	return result
}
