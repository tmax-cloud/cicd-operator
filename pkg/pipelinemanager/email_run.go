package pipelinemanager

import (
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"strconv"
)

func generateEmailRunTask(j cicdv1.Job, task *tektonv1beta1.PipelineTask) error {
	task.TaskRef = generateCustomTaskRef(cicdv1.CustomTaskKindEmail)

	// Get receivers
	receivers := j.Email.Receivers

	// Get title
	title := j.Email.Title

	// Get content
	content := j.Email.Content

	// Get isHTML
	isHTML := j.Email.IsHTML

	task.Params = append(task.Params, []tektonv1beta1.Param{
		{Name: cicdv1.CustomTaskEmailParamKeyReceivers, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeArray, ArrayVal: receivers}},
		{Name: cicdv1.CustomTaskEmailParamKeyTitle, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: title}},
		{Name: cicdv1.CustomTaskEmailParamKeyContent, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: content}},
		{Name: cicdv1.CustomTaskEmailParamKeyIsHTML, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: strconv.FormatBool(isHTML)}},
	}...)

	return nil
}
