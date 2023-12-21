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

package integrationconfigs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/apiserver"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"github.com/tmax-cloud/cicd-operator/pkg/server"
	"k8s.io/apimachinery/pkg/types"
)

const (
	defaultBranch = "master"
)

func (h *handler) runPreHandler(w http.ResponseWriter, req *http.Request) {
	h.runHandler(w, req, git.EventTypePullRequest)
}

func (h *handler) runPostHandler(w http.ResponseWriter, req *http.Request) {
	h.runHandler(w, req, git.EventTypePush)
}

func (h *handler) runHandler(w http.ResponseWriter, req *http.Request, et git.EventType) {
	reqID := utils.RandomString(10)
	log := h.log.WithValues("request", reqID)

	// Get ns/resource name
	vars := mux.Vars(req)

	ns, nsExist := vars[apiserver.NamespaceParamKey]
	resName, nameExist := vars[icParamKey]
	if !nsExist || !nameExist {
		log.Info("url is malformed")
		_ = utils.RespondError(w, http.StatusBadRequest, "url is malformed")
		return
	}

	// Get user
	user, err := apiserver.GetUserName(req.Header)
	if err != nil {
		log.Info(err.Error())
		_ = utils.RespondError(w, http.StatusUnauthorized, fmt.Sprintf("req: %s, forbidden user, err : %s", reqID, err.Error()))
		return
	}
	userEscaped := regexp.MustCompile("[^-A-Za-z0-9_.]").ReplaceAllString(user, "_")

	// Get IntegrationConfig
	ic := &cicdv1.IntegrationConfig{}
	if err := h.k8sClient.Get(context.Background(), types.NamespacedName{Name: resName, Namespace: ns}, ic); err != nil {
		log.Info(err.Error())
		_ = utils.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("req: %s, cannot get IntegrationConfig %s/%s", reqID, ns, resName))
		return
	}

	gitHost, err := ic.Spec.Git.GetGitHost()
	if err != nil {
		log.Info(err.Error())
		_ = utils.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("req: %s, cannot get IntegrationConfig %s/%s's git host", reqID, ns, resName))
		return
	}

	// Build webhook
	wh := &git.Webhook{
		EventType: et,
		Repo: git.Repository{
			Name: ic.Spec.Git.Repository,
			URL:  fmt.Sprintf("%s/%s", gitHost, ic.Spec.Git.Repository),
		},
	}

	userReqPre := &cicdv1.IntegrationConfigAPIReqRunPreBody{}
	userReqPost := &cicdv1.IntegrationConfigAPIReqRunPostBody{}

	switch et {
	case git.EventTypePullRequest:
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(userReqPre); err != nil {
			log.Info(err.Error())
			_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("req: %s, cannot build pull_request webhook", reqID))
			return
		}
		pr, err := buildPullRequestWebhook(userReqPre, userEscaped)
		if err != nil {
			log.Info(err.Error())
			_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("req: %s, cannot build pull_request webhook", reqID))
			return
		}
		wh.PullRequest = pr
	case git.EventTypePush:
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(userReqPost); err != nil {
			log.Info(err.Error())
			_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("req: %s, cannot build pull_request webhook", reqID))
			return
		}
		push, err := buildPushWebhook(userReqPost)
		if err != nil {
			log.Info(err.Error())
			_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("req: %s, cannot build push webhook", reqID))
			return
		}
		wh.Push = push
	}
	wh.Sender = git.User{
		Name: fmt.Sprintf("trigger-%s-end", userEscaped),
	}

	// Update IntegrationConfig based on the request
	switch et {
	case git.EventTypePullRequest:
		updatedIC, err := updateIntegrationConfigPre(ic, userReqPre, et)
		if err != nil {
			log.Info(err.Error())
			_ = utils.RespondError(w, http.StatusBadRequest, "cannot update pull_request integrationconfig. please check jobName is valid")
			return
		}
		ic = updatedIC
	case git.EventTypePush:
		updatedIC, err := updateIntegrationConfigPost(ic, userReqPost, et)
		if err != nil {
			log.Info(err.Error())
			_ = utils.RespondError(w, http.StatusBadRequest, "cannot update push integrationconfig. please check jobName is valid")
			return
		}
		ic = updatedIC
	}

	// Trigger Run!
	if err := server.HandleEvent(wh, ic, "dispatcher"); err != nil {
		log.Info(err.Error())
		_ = utils.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("req: %s, cannot handle event, err : %s", reqID, err.Error()))
		return
	}

	_ = utils.RespondJSON(w, struct{}{})
}

func updateIntegrationConfigPost(ic *cicdv1.IntegrationConfig, userReqBody *cicdv1.IntegrationConfigAPIReqRunPostBody, et git.EventType) (*cicdv1.IntegrationConfig, error) {
	targetJob := &ic.Spec.Jobs.PostSubmit
	var existingJob *cicdv1.Job

	for _, addParams := range userReqBody.AddTektonTaskParams {
		jobName := addParams.JobName
		if jobName == "" {
			return nil, errors.New("JobName must be set")
		}

		existingJob = nil
		for i, job := range *targetJob {
			if job.Name == jobName {
				existingJob = &(*targetJob)[i]
				break
			}
		}

		if existingJob == nil {
			return nil, fmt.Errorf("job with name '%s' not found", jobName)
		}

		for _, taskDef := range addParams.TektonTask {
			// Check if a parameter with the same name already exists
			for i, existingParam := range existingJob.TektonTask.Params {
				if existingParam.Name == taskDef.Name {
					existingJob.TektonTask.Params = append(existingJob.TektonTask.Params[:i], existingJob.TektonTask.Params[i+1:]...)
					break
				}
			}

			existingJob.TektonTask.Params = append(existingJob.TektonTask.Params, cicdv1.ParameterValue{
				Name:      taskDef.Name,
				StringVal: taskDef.StringVal,
			})
		}
	}

	return ic, nil
}

func updateIntegrationConfigPre(ic *cicdv1.IntegrationConfig, userReqBody *cicdv1.IntegrationConfigAPIReqRunPreBody, et git.EventType) (*cicdv1.IntegrationConfig, error) {
	targetJob := &ic.Spec.Jobs.PreSubmit
	var existingJob *cicdv1.Job

	for _, addParams := range userReqBody.AddTektonTaskParams {
		jobName := addParams.JobName
		if jobName == "" {
			return nil, errors.New("JobName must be set")
		}

		existingJob = nil
		for i, job := range *targetJob {
			if job.Name == jobName {
				existingJob = &(*targetJob)[i]
				break
			}
		}

		if existingJob == nil {
			return nil, fmt.Errorf("job with name '%s' not found", jobName)
		}

		for _, taskDef := range addParams.TektonTask {
			// Check if a parameter with the same name already exists
			for i, existingParam := range existingJob.TektonTask.Params {
				if existingParam.Name == taskDef.Name {
					existingJob.TektonTask.Params = append(existingJob.TektonTask.Params[:i], existingJob.TektonTask.Params[i+1:]...)
					break
				}
			}

			existingJob.TektonTask.Params = append(existingJob.TektonTask.Params, cicdv1.ParameterValue{
				Name:      taskDef.Name,
				StringVal: taskDef.StringVal,
			})
		}
	}

	return ic, nil
}

func buildPullRequestWebhook(userReq *cicdv1.IntegrationConfigAPIReqRunPreBody, user string) (*git.PullRequest, error) {
	baseBranch := userReq.BaseBranch
	headBranch := userReq.HeadBranch
	if baseBranch == "" {
		baseBranch = defaultBranch
	}
	if headBranch == "" {
		return nil, fmt.Errorf("head_branch must be set")
	}

	return &git.PullRequest{
		State:  git.PullRequestStateOpen,
		Action: git.PullRequestActionOpen,
		Author: git.User{
			Name: fmt.Sprintf("trigger-%s-end", user),
		},
		Base: git.Base{
			Ref: baseBranch,
			Sha: git.FakeSha,
		},
		Head: git.Head{
			Ref: headBranch,
			Sha: git.FakeSha,
		},
	}, nil
}

func buildPushWebhook(userReq *cicdv1.IntegrationConfigAPIReqRunPostBody) (*git.Push, error) {
	branch := userReq.Branch
	if branch == "" {
		branch = defaultBranch
	}

	return &git.Push{
		Ref: branch,
		Sha: git.FakeSha,
	}, nil
}
