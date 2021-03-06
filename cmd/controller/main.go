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
	"github.com/tmax-cloud/cicd-operator/internal/logrotate"
	"github.com/tmax-cloud/cicd-operator/pkg/apiserver"
	"github.com/tmax-cloud/cicd-operator/pkg/chatops"
	"github.com/tmax-cloud/cicd-operator/pkg/chatops/plugins/approve"
	"github.com/tmax-cloud/cicd-operator/pkg/chatops/plugins/trigger"
	"github.com/tmax-cloud/cicd-operator/pkg/collector"
	"github.com/tmax-cloud/cicd-operator/pkg/dispatcher"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"github.com/tmax-cloud/cicd-operator/pkg/server"
	"io"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	apiregv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
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
	utilruntime.Must(apiregv1.AddToScheme(scheme))
	utilruntime.Must(rbac.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

// +kubebuilder:rbac:groups="",resources=events,namespace=cicd-system,verbs=get;list;watch;create;update;patch

func main() {
	var metricsAddr string
	var healthAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&healthAddr, "health-addr", ":8888", "The address the health endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	flag.Parse()

	// Set log rotation
	logFile, err := logrotate.LogFile()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer func() {
		_ = logFile.Close()
	}()
	logWriter := io.MultiWriter(logFile, os.Stdout)
	ctrl.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(logWriter)))
	if err := logrotate.StartRotate(); err != nil {
		setupLog.Error(err, "")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		HealthProbeBindAddress: healthAddr,
		Port:                   9443,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "2787db31.tmax.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Add healthz handler
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to add healthz handler")
		os.Exit(1)
	}
	// Add readyz handler
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to add readyz handler")
		os.Exit(1)
	}

	// Start garbage collector
	gc, err := collector.New(mgr.GetClient())
	if err != nil {
		setupLog.Error(err, "error initializing garbage collector")
		os.Exit(1)
	}
	go gc.Start()

	// Config Controller
	cfgCtrl := &controllers.ConfigReconciler{Log: ctrl.Log.WithName("controllers").WithName("ConfigController"), Handlers: map[string]configs.Handler{}}
	go cfgCtrl.Start()
	cfgCtrl.Add(configs.ConfigMapNameCICDConfig, configs.ApplyControllerConfigChange)
	cfgCtrl.Add(configs.ConfigMapNameEmailTemplate, configs.ApplyEmailTemplateConfigChange)
	// Wait for initial config reconcile
	<-configs.InitCh

	// Controllers
	if err = (&controllers.IntegrationConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("IntegrationConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IntegrationConfig")
		os.Exit(1)
	}

	if err = controllers.NewIntegrationJobReconciler(mgr.GetClient(), mgr.GetScheme(), ctrl.Log.WithName("controllers").WithName("IntegrationJob")).SetupWithManager(mgr); err != nil {
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
	customRunController.AddKindHandler(cicdv1.CustomTaskKindEmail, &customs.EmailRunHandler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("EmailRun"),
		Scheme: mgr.GetScheme(),
	})
	customRunController.AddKindHandler(cicdv1.CustomTaskKindSlack, &customs.SlackRunHandler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("SlackRun"),
		Scheme: mgr.GetScheme(),
	})
	if err = customRunController.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CustomRun")
	}
	// +kubebuilder:scaffold:builder

	// Check for ingress first
	setupLog.Info("Waiting for ingress to be ready")
	ingCon := controllers.NewIngressController()
	go ingCon.Start()
	if err := ingCon.Wait(); err != nil {
		setupLog.Error(err, "error waiting for ingress")
		os.Exit(1)
	}

	// Init chat-ops
	co := chatops.New(mgr.GetClient())

	// Init plugins
	approveHandler := &approve.Handler{Client: mgr.GetClient()}
	triggerHandler := &trigger.Handler{Client: mgr.GetClient()}

	co.RegisterCommandHandler(approve.CommandTypeApprove, approveHandler.HandleChatOps)
	co.RegisterCommandHandler(approve.CommandTypeGitLabApprove, approveHandler.HandleChatOps)
	co.RegisterCommandHandler(trigger.CommandTypeTest, triggerHandler.HandleChatOps)
	co.RegisterCommandHandler(trigger.CommandTypeRetest, triggerHandler.HandleChatOps)

	// Create and start webhook server
	srv := server.New(mgr.GetClient(), mgr.GetConfig())
	// Add plugins for webhook
	server.AddPlugin([]git.EventType{git.EventTypePullRequest, git.EventTypePush}, &dispatcher.Dispatcher{Client: mgr.GetClient()})
	server.AddPlugin([]git.EventType{git.EventTypeIssueComment, git.EventTypePullRequestReview, git.EventTypePullRequestReviewComment}, co)
	server.AddPlugin([]git.EventType{git.EventTypePullRequest, git.EventTypePullRequestReview}, approveHandler)
	go srv.Start()

	// Start API aggregation server
	apiServer := apiserver.New(mgr.GetScheme())
	go apiServer.Start()

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
