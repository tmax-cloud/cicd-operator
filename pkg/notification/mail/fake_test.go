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

package mail

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFakeSender(t *testing.T) {
	f := NewFakeSender()
	require.Empty(t, f.Mails)
}

func Test_fakeSender_Send(t *testing.T) {
	tc := map[string]struct {
		to      []string
		subject string
		content string
		isHTML  bool

		errorOccurs  bool
		errorMessage string
		expectedMail FakeMailEntity
	}{
		"normal": {
			to:      []string{"admin@tmax.co.kr", "admin2@tmax.co.kr"},
			subject: "TEST EMAIL",
			content: "hahaha",
			isHTML:  false,
			expectedMail: FakeMailEntity{
				Receivers: []string{"admin@tmax.co.kr", "admin2@tmax.co.kr"},
				Title:     "TEST EMAIL",
				Content:   "hahaha",
				IsHTML:    false,
			},
		},
		"html": {
			to:      []string{"admin@tmax.co.kr", "admin2@tmax.co.kr"},
			subject: "TEST EMAIL",
			content: "<p>hahaha</p>",
			isHTML:  true,
			expectedMail: FakeMailEntity{
				Receivers: []string{"admin@tmax.co.kr", "admin2@tmax.co.kr"},
				Title:     "TEST EMAIL",
				Content:   "<p>hahaha</p>",
				IsHTML:    true,
			},
		},
		"emptyReceivers": {
			errorOccurs:  true,
			errorMessage: "receivers list is empty",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			f := &fakeSender{}

			err := f.Send(c.to, c.subject, c.content, c.isHTML)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				require.Len(t, f.Mails, 1)
				require.Equal(t, c.expectedMail, f.Mails[0])
			}
		})
	}
}
