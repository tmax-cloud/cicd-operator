package server

import (
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
)

type Plugin interface {
	Handle(*git.Webhook, *cicdv1.IntegrationConfig) error
}

var plugins = map[git.EventType][]Plugin{}

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

func GetPlugins(ev git.EventType) []Plugin {
	return plugins[ev]
}
