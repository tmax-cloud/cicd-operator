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

import "fmt"

// FakeMailEntity is a mail entity for the fake senders
type FakeMailEntity struct {
	Receivers []string
	Title     string
	Content   string
	IsHTML    bool
}

type fakeSender struct {
	Mails []FakeMailEntity
}

// NewFakeSender creates a new fakeSender
func NewFakeSender() *fakeSender {
	return &fakeSender{}
}

func (s *fakeSender) Send(to []string, subject string, content string, isHTML bool) error {
	if len(to) < 1 {
		return fmt.Errorf("receivers list is empty")
	}

	s.Mails = append(s.Mails, FakeMailEntity{
		Receivers: to,
		Title:     subject,
		Content:   content,
		IsHTML:    isHTML,
	})
	return nil
}
