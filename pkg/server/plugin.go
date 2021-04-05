package server

import (
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
)

// Plugin is a webhook plugin interface, which handles git webhook payloads
type Plugin interface {
	Handle(*git.Webhook, *cicdv1.IntegrationConfig) error
}

var plugins = map[git.EventType][]Plugin{}

// HandleEvent passes webhook event to plugins
func HandleEvent(wh *git.Webhook, ic *cicdv1.IntegrationConfig) error {
	var retErr error
	plugins := getPlugins(wh.EventType)
	for _, p := range plugins {
		if err := p.Handle(wh, ic); err != nil {
			retErr = err
		}
	}
	return retErr
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
