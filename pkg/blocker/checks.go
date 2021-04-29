package blocker

import (
	"fmt"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"sort"
	"strings"
)

// checkConditionsSimple checks labels, approved, author, branch conditions for a PR to be in a merge pool
func checkConditionsSimple(q cicdv1.MergeQuery, pr *git.PullRequest) (bool, string) {
	var messages []string

	// Check labels
	labels := map[string]struct{}{}
	for _, l := range pr.Labels {
		labels[l.Name] = struct{}{}
	}
	if q.ApproveRequired { // Check 'approved' label if approval is required
		q.Labels = append(q.Labels, "approved")
	}
	passLabelChecks, labelCheckMsg := checkLabels(labels, q)
	if labelCheckMsg != "" {
		messages = append(messages, labelCheckMsg)
	}

	// Check author
	passAuthorCheck, authorCheckMsg := checkAuthor(pr.Sender.Name, q)
	if authorCheckMsg != "" {
		messages = append(messages, authorCheckMsg)
	}

	// Check branch
	passBranchCheck, branchCheckMsg := checkBranch(pr.Base.Ref, q)
	if branchCheckMsg != "" {
		messages = append(messages, branchCheckMsg)
	}

	return passLabelChecks && passAuthorCheck && passBranchCheck, strings.Join(messages, " ")
}

// checkConditionsFull is a checkConditionsSimple + commit status check + merge conflict check
// Return: status / removeFromMergePool / description
func checkConditionsFull(q cicdv1.MergeQuery, pr *PullRequest) (bool, bool, string) {
	var messages []string

	// Check labels (, approved), branch, author
	simpleResult, simpleMessage := checkConditionsSimple(q, &pr.PullRequest)
	if simpleMessage != "" {
		messages = append(messages, simpleMessage)
	}
	if !simpleResult {
		return false, true, strings.Join(messages, " ")
	}

	// Check merge conflict
	passMergeConflict := pr.Mergeable
	if !passMergeConflict {
		messages = append(messages, "Merge conflicts exist.")
	}

	// Check commit statuses
	passCommitStatus, commitStatusMsg := checkChecks(pr.Statuses, q)
	if commitStatusMsg != "" {
		messages = append(messages, commitStatusMsg)
	}

	return simpleResult && passMergeConflict && passCommitStatus, false, strings.Join(messages, " ")
}

func checkBranch(b string, q cicdv1.MergeQuery) (bool, string) {
	branch := strings.TrimPrefix(b, "refs/heads/")
	isProperBranch := true
	msg := ""
	if len(q.Branches) > 0 && !containsString(branch, q.Branches) {
		isProperBranch = false
		msg = fmt.Sprintf("Branch [%s] is not in branches query.", branch)
	}
	if containsString(branch, q.SkipBranches) {
		isProperBranch = false
		msg = fmt.Sprintf("Branch [%s] is in skipBranches query.", branch)
	}

	return isProperBranch, msg
}

func checkAuthor(author string, q cicdv1.MergeQuery) (bool, string) {
	isProperAuthor := true
	msg := ""
	if len(q.Authors) > 0 && !containsString(author, q.Authors) {
		isProperAuthor = false
		msg = fmt.Sprintf("Author [%s] is not in authors query.", author)
	}
	if containsString(author, q.SkipAuthors) {
		isProperAuthor = false
		msg = fmt.Sprintf("Author [%s] is in skipAuthors query.", author)
	}

	return isProperAuthor, msg
}

func checkLabels(labels map[string]struct{}, q cicdv1.MergeQuery) (bool, string) {
	isProperLabels := true
	msg := ""

	if len(q.Labels) > 0 {
		var missing []string
		for _, l := range q.Labels {
			_, exist := labels[l]
			if !exist {
				isProperLabels = false
				missing = append(missing, l)
			}
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			msg = fmt.Sprintf("Label [%s] is required.", strings.Join(missing, ","))
		}
	}
	if len(q.BlockLabels) > 0 {
		var blocking []string
		for _, l := range q.BlockLabels {
			_, exist := labels[l]
			if exist {
				isProperLabels = false
				blocking = append(blocking, l)
			}
		}
		if len(blocking) > 0 {
			if msg != "" {
				msg += " "
			}
			sort.Strings(blocking)
			msg += fmt.Sprintf("Label [%s] is blocking the merge.", strings.Join(blocking, ","))
		}
	}

	return isProperLabels, msg
}

func checkChecks(statuses map[string]git.CommitStatus, q cicdv1.MergeQuery) (bool, string) {
	var unmetChecks []string
	passAllRequiredChecks := true
	if len(q.Checks) > 0 {
		// Check for the required checks
		for _, c := range q.Checks {
			s, exist := statuses[c]
			if exist {
				if s.State != "success" {
					passAllRequiredChecks = false
					unmetChecks = append(unmetChecks, c)
				}
			} else {
				// Handle if the check is not registered yet
				passAllRequiredChecks = false
				unmetChecks = append(unmetChecks, c)
			}
		}
	} else {
		// Check for the other checks
		for context, s := range statuses {
			if context == blockerContext {
				continue
			}
			if s.State != "success" && !containsString(context, q.OptionalChecks) {
				passAllRequiredChecks = false
				unmetChecks = append(unmetChecks, context)
			}
		}
	}

	msg := ""
	if !passAllRequiredChecks {
		sort.Strings(unmetChecks)
		msg = fmt.Sprintf("Checks [%s] are not successful.", strings.Join(unmetChecks, ","))
	}

	return passAllRequiredChecks, msg
}

func containsString(needle string, arr []string) bool {
	for _, e := range arr {
		if e == needle {
			return true
		}
	}
	return false
}
