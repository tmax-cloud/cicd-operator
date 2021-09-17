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

package size

import (
	"fmt"
	"strings"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	labelPrefix = "size/"
)

type prSize int

const (
	sizeXS = prSize(iota)
	sizeS
	sizeM
	sizeL
	sizeXL
	sizeXXL
)

var labels = [...]string{"XS", "S", "M", "L", "XL", "XXL"}

var log = logf.Log.WithName("size-plugin")

// Size plugin finds out how many lines are changed by the pull request and label the size to the pull request
type Size struct {
	Client client.Client
}

// Name returns a name of size plugin
func (s *Size) Name() string {
	return "size"
}

// Handle handles a pull request event and set size label to the pull request
func (s *Size) Handle(wh *git.Webhook, config *cicdv1.IntegrationConfig) error {
	// Filter only PullRequest event's open/synchronize action
	pr := wh.PullRequest
	if wh.EventType != git.EventTypePullRequest || pr == nil || (pr.Action != git.PullRequestActionOpen && pr.Action != git.PullRequestActionReOpen && pr.Action != git.PullRequestActionSynchronize) {
		return nil
	}

	// Get the number of change lines
	gitCli, err := utils.GetGitCli(config, s.Client)
	if err != nil {
		return err
	}

	// Get diffs of the pull request
	diff, err := gitCli.GetPullRequestDiff(pr.ID)
	if err != nil {
		return err
	}

	numLines := 0
	for _, c := range diff.Changes {
		numLines += c.Changes
	}

	// Determine the size
	properLabel := determineProperSizeLabel(numLines)

	// Check old size label
	currentLabels := getSizeLabels(pr.Labels)

	// Delete the old label if it exists && not proper
	hasProperLabel := false
	for _, l := range currentLabels {
		// Is it a proper label?
		if l == properLabel {
			hasProperLabel = true
			continue
		}
		// If not, delete it!
		if err := gitCli.DeleteLabel(git.IssueTypePullRequest, wh.PullRequest.ID, l); err != nil {
			return err
		}
	}
	if hasProperLabel {
		return nil
	}

	log.Info(fmt.Sprintf("Setting size label %s to %s/%s's PR#%d", properLabel, config.Namespace, config.Name, pr.ID), "changes", numLines)

	// Set a new size label
	if err := gitCli.SetLabel(git.IssueTypePullRequest, wh.PullRequest.ID, properLabel); err != nil {
		return err
	}

	return nil
}

func getSizeLabels(labels []git.IssueLabel) []string {
	var sizes []string

	for _, l := range labels {
		if strings.HasPrefix(l.Name, labelPrefix) {
			sizes = append(sizes, l.Name)
		}
	}

	return sizes
}

func determineProperSizeLabel(numLines int) string {
	var size prSize
	switch {
	case numLines <= configs.PluginSizeS:
		size = sizeXS
	case numLines <= configs.PluginSizeM:
		size = sizeS
	case numLines <= configs.PluginSizeL:
		size = sizeM
	case numLines <= configs.PluginSizeXL:
		size = sizeL
	case numLines <= configs.PluginSizeXXL:
		size = sizeXL
	default:
		size = sizeXXL
	}
	return fmt.Sprintf("%s%s", labelPrefix, labels[size])
}
