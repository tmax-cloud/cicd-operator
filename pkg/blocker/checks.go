package blocker

import (
	"fmt"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"strings"
)

func checkConditions(q cicdv1.MergeQuery, pr *git.PullRequest) (bool, string) {
	var messages []string

	// Check labels
	var labels []string
	for _, l := range pr.Labels {
		labels = append(labels, l.Name)
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

func checkLabels(labels []string, q cicdv1.MergeQuery) (bool, string) {
	isProperLabels := true
	msg := ""
	for _, l := range q.Labels {
		if !containsString(l, labels) {
			isProperLabels = false
			msg = fmt.Sprintf("Label [%s] is required.", l)
		}
	}
	for _, l := range q.BlockLabels {
		if containsString(l, labels) {
			isProperLabels = false
			msg = fmt.Sprintf("Label [%s] is blocking the merge.", l)
		}
	}

	return isProperLabels, msg
}

func checkChecks(statuses []git.CommitStatus, q cicdv1.MergeQuery) (bool, string) {
	var unmetChecks []string
	passAllRequiredChecks := true
	if len(q.Checks) > 0 {
		// Check for the required checks
		for _, c := range q.Checks {
			for _, s := range statuses {
				if s.Context == blockerContext {
					continue
				}
				if s.Context == c {
					if s.State != "success" {
						passAllRequiredChecks = false
						unmetChecks = append(unmetChecks, s.Context)
					}
					break
				}
			}
		}
	} else {
		// Check for the other checks
		for _, s := range statuses {
			if s.Context == blockerContext {
				continue
			}
			if s.State != "success" && !containsString(s.Context, q.OptionalChecks) {
				passAllRequiredChecks = false
				unmetChecks = append(unmetChecks, s.Context)
			}
		}
	}

	msg := ""
	if !passAllRequiredChecks {
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

func containsInt(needle int, arr []int) bool {
	for _, e := range arr {
		if e == needle {
			return true
		}
	}
	return false
}
