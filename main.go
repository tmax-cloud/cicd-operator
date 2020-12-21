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

package main

import (
	"flag"
	"fmt"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tmax-cloud/cicd-operator/controllers/customs"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/pkg/dispatcher"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"github.com/tmax-cloud/cicd-operator/pkg/scheduler"
	"github.com/tmax-cloud/cicd-operator/pkg/server"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(cicdv1.AddToScheme(scheme))
	utilruntime.Must(tektonv1beta1.AddToScheme(scheme))
	utilruntime.Must(tektonv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	flag.StringVar(&configs.ExternalHostName, "external-hostname", "", "External hostname for webhook server (default is ingress hostname)")
	flag.IntVar(&configs.MaxPipelineRun, "max-pipeline-run", 5, "Max number of pipelineruns that can run simultaneously")

	// Email setting
	flag.BoolVar(&configs.EnableMail, "enable-mail", false, "Whether or not to enable sendMail feature")
	flag.StringVar(&configs.SMTPHost, "smtp-host", "", "SMTP host e.g., mail.tmax.co.kr:25")
	flag.StringVar(&configs.SMTPUserSecret, "smtp-user-secret", "", "Secret name whose type is basic auth and contains username/password")

	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	// Check email setting
	if configs.EnableMail {
		if configs.SMTPHost == "" || configs.SMTPUserSecret == "" {
			setupLog.Error(fmt.Errorf("email feature is enabled but following options are not given: smtp-host, smtp-user-secret"), "")
			os.Exit(1)
		}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "2787db31.tmax.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.IntegrationConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("IntegrationConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IntegrationConfig")
		os.Exit(1)
	}
	if err = (&controllers.IntegrationJobReconciler{
		Client:    mgr.GetClient(),
		Log:       ctrl.Log.WithName("controllers").WithName("IntegrationJob"),
		Scheme:    mgr.GetScheme(),
		Scheduler: scheduler.New(mgr.GetClient(), mgr.GetScheme()),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IntegrationJob")
		os.Exit(1)
	}
	if err = (&controllers.ApprovalReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Approval"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Approval")
		os.Exit(1)
	}
	customRunController := &controllers.CustomRunReconciler{
		Client:          mgr.GetClient(),
		Log:             ctrl.Log.WithName("controllers").WithName("CustomRun"),
		Scheme:          mgr.GetScheme(),
		KindHandlerMap:  map[string]controllers.KindHandler{},
		HandlerChildren: map[string][]runtime.Object{},
	}
	// Add custom Run handlers
	customRunController.AddKindHandler(cicdv1.CustomTaskKindApproval, &customs.ApprovalRunHandler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ApprovalRun"),
		Scheme: mgr.GetScheme(),
	}, &cicdv1.Approval{})
	if err = customRunController.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CustomRun")
	}
	// +kubebuilder:scaffold:builder

	// Check for ingress first
	setupLog.Info("Waiting for ingress to be ready")
	if err := controllers.WaitIngressReady(); err != nil {
		setupLog.Error(err, "error while waiting ingress ready")
		os.Exit(1)
	}

	// Create and start webhook server
	srv := server.New(mgr.GetClient(), mgr.GetConfig())

	// Add plugins for webhook
	server.AddPlugin([]git.EventType{git.EventTypePullRequest, git.EventTypePush}, &dispatcher.Dispatcher{Client: mgr.GetClient()})

	go srv.Start()

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
