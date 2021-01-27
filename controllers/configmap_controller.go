/*


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

package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"strconv"
)

const (
	configNameConfig        = "cicd-config"
	configNameEmailTemplate = "email-template"
)

// ConfigReconciler reconciles a Approval object
type ConfigReconciler struct {
	client typedcorev1.ConfigMapInterface
	Log    logr.Logger

	GcChan   chan struct{}
	InitChan chan struct{}

	Init bool
}

// Start starts the config map reconciler
func (r *ConfigReconciler) Start() {
	var err error
	r.client, err = newConfigMapClient()
	if err != nil {
		r.Log.Error(err, "")
		os.Exit(1)
	}

	// Get first to check the ConfigMap's existence
	_, err = r.client.Get(context.Background(), configNameConfig, metav1.GetOptions{})
	if err != nil {
		r.Log.Error(err, "")
		os.Exit(1)
	}

	_, err = r.client.Get(context.Background(), configNameEmailTemplate, metav1.GetOptions{})
	if err != nil {
		r.Log.Error(err, "")
		os.Exit(1)
	}

	for {
		r.watch()
	}
}

func (r *ConfigReconciler) watch() {
	log := r.Log.WithName("config controller")

	watcher, err := r.client.Watch(context.Background(), metav1.ListOptions{
		FieldSelector: "metadata.name!=2787db31.tmax.io",
	})
	if err != nil {
		log.Error(err, "")
		return
	}

	for ev := range watcher.ResultChan() {
		cm, ok := ev.Object.(*corev1.ConfigMap)
		if ok {
			if err := r.Reconcile(cm); err != nil {
				log.Error(err, "")
			}
		}
	}
}

type cfgType int

const (
	cfgTypeString cfgType = iota
	cfgTypeInt
	cfgTypeBool
)

type operatorConfig struct {
	Type cfgType

	StringVal     *string
	StringDefault string

	IntVal     *int
	IntDefault int

	BoolVal     *bool
	BoolDefault bool
}

// Reconcile reconciles ConfigMap
func (r *ConfigReconciler) Reconcile(cm *corev1.ConfigMap) error {
	r.Log.Info("Config is changed")

	if cm == nil {
		return nil
	}

	switch cm.Name {
	case configNameConfig:
		if err := r.reconcileConfig(cm); err != nil {
			return err
		}
	case configNameEmailTemplate:
		if err := r.reconcileEmailTemplate(cm); err != nil {
			return err
		}
	}

	if !r.Init {
		r.Init = true
		r.InitChan <- struct{}{}
	}

	return nil
}

func (r *ConfigReconciler) reconcileConfig(cm *corev1.ConfigMap) error {
	vars := map[string]operatorConfig{
		"maxPipelineRun":            {Type: cfgTypeInt, IntVal: &configs.MaxPipelineRun, IntDefault: 5},         // Max PipelineRun count
		"enableMail":                {Type: cfgTypeBool, BoolVal: &configs.EnableMail, BoolDefault: false},      // Enable Mail
		"externalHostName":          {Type: cfgTypeString, StringVal: &configs.ExternalHostName},                // External Hostname
		"reportRedirectUriTemplate": {Type: cfgTypeString, StringVal: &configs.ReportRedirectURITemplate},       // RedirectUriTemplate for report access
		"smtpHost":                  {Type: cfgTypeString, StringVal: &configs.SMTPHost},                        // SMTP Host
		"smtpUserSecret":            {Type: cfgTypeString, StringVal: &configs.SMTPUserSecret},                  // SMTP Cred
		"collectPeriod":             {Type: cfgTypeInt, IntVal: &configs.CollectPeriod, IntDefault: 120},        // GC period
		"integrationJobTTL":         {Type: cfgTypeInt, IntVal: &configs.IntegrationJobTTL, IntDefault: 120},    // GC threshold
		"ingressClass":              {Type: cfgTypeString, StringVal: &configs.IngressClass, StringDefault: ""}, // Ingress class
	}

	getVars(cm.Data, vars)

	// Check SMTP config.s
	if configs.EnableMail && (configs.SMTPHost == "" || configs.SMTPUserSecret == "") {
		return fmt.Errorf("email is enaled but smtp access info. is not given")
	}

	// Reconfigure GC
	r.GcChan <- struct{}{}

	return nil
}

func (r *ConfigReconciler) reconcileEmailTemplate(cm *corev1.ConfigMap) error {
	vars := map[string]operatorConfig{
		"request-title":   {Type: cfgTypeString, StringVal: &configs.ApprovalRequestMailTitle, StringDefault: "[CI/CD] Approval '{{.Name}}' is requested to you"},
		"request-content": {Type: cfgTypeString, StringVal: &configs.ApprovalRequestMailContent, StringDefault: "{{.Name}}"},
		"result-title":    {Type: cfgTypeString, StringVal: &configs.ApprovalResultMailTitle, StringDefault: "[CI/CD] Approval is {{.Status.Result}}"},
		"result-content":  {Type: cfgTypeString, StringVal: &configs.ApprovalResultMailContent, StringDefault: "{{.Name}}"},
	}

	getVars(cm.Data, vars)

	return nil
}

func getVars(data map[string]string, vars map[string]operatorConfig) {
	for key, c := range vars {
		v, exist := data[key]
		switch c.Type {
		case cfgTypeString:
			if c.StringVal == nil {
				continue
			}
			if exist {
				*c.StringVal = v
			} else {
				*c.StringVal = c.StringDefault
			}
		case cfgTypeInt:
			if c.IntVal == nil {
				continue
			}
			if exist {
				i, err := strconv.Atoi(v)
				if err != nil {
					continue
				}
				*c.IntVal = i
			} else {
				*c.IntVal = c.IntDefault
			}
		case cfgTypeBool:
			if c.BoolVal == nil {
				continue
			}
			if exist {
				b, err := strconv.ParseBool(v)
				if err != nil {
					continue
				}
				*c.BoolVal = b
			} else {
				*c.BoolVal = c.BoolDefault
			}
		}
	}
}

func newConfigMapClient() (typedcorev1.ConfigMapInterface, error) {
	conf, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return nil, err
	}

	namespace, err := utils.Namespace()
	if err != nil {
		return nil, err
	}

	return clientSet.CoreV1().ConfigMaps(namespace), nil
}
