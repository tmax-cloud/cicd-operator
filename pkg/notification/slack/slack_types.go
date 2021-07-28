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

package slack

// Message is a slack message
type Message struct {
	Text   string         `json:"text"`
	Blocks []MessageBlock `json:"blocks"`
}

// MessageBlock is a slack message block
type MessageBlock struct {
	Type string    `json:"type"`
	Text BlockText `json:"text"`
}

// BlockText is an actual text
type BlockText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
