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

package server

import (
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
)

// Plugin is a webhook plugin interface, which handles git webhook payloads
type Plugin interface {
	Name() string
	Handle(*git.Webhook, *cicdv1.IntegrationConfig) error
}

var plugins = map[git.EventType][]Plugin{}

// HandleEvent passes webhook event to plugins
func HandleEvent(wh *git.Webhook, ic *cicdv1.IntegrationConfig, wantedPlugins ...string) error {
	var retErr error
	plugins := getPlugins(wh.EventType)
	for _, p := range plugins {
		if len(wantedPlugins) == 0 || contains(wantedPlugins, p.Name()) {
			if err := p.Handle(wh, ic); err != nil {
				retErr = err
			}
		}
	}
	return retErr
}

func contains(list []string, needle string) bool {
	for _, s := range list {
		if s == needle {
			return true
		}
	}
	return false
}

// AddPlugin adds handler for specific events
func AddPlugin(events []git.EventType, p Plugin) {
	for _, ev := range events {
		addPlugin(ev, p)
	}
}

func addPlugin(ev git.EventType, p Plugin) {
	_, exist := plugins[ev]
	if !exist {
		plugins[ev] = []Plugin{}
	}
	plugins[ev] = append(plugins[ev], p)
}

func getPlugins(ev git.EventType) []Plugin {
	return plugins[ev]
}
