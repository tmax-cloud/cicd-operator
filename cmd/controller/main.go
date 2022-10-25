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

package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tmax-cloud/cicd-operator/controllers/customs"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/logrotate"
	"github.com/tmax-cloud/cicd-operator/pkg/collector"
	"github.com/tmax-cloud/cicd-operator/pkg/notification/mail"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	apiregv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	opts := zap.Options{
		Development: false,
	}
	opts.BindFlags(flag.CommandLine)

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
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts), zap.WriteTo(logWriter)))
	if err := logrotate.StartRotate("0 0 1 * * ?"); err != nil {
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

	// Config Controller
	// Initiate first, before any other components start
	cfgCtrl, err := controllers.NewConfigReconciler(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to initiate config reconciler")
		os.Exit(1)
	}
	go cfgCtrl.Start()
	cfgCtrl.Add(configs.ConfigMapNameCICDConfig, configs.ApplyControllerConfigChange)
	cfgCtrl.Add(configs.ConfigMapNamePluginConfig, configs.ApplyPluginConfigChange)
	cfgCtrl.Add(configs.ConfigMapNameBlockerConfig, configs.ApplyBlockerConfigChange)
	// Wait for initial config reconcile
	<-configs.ControllerInitCh
	<-configs.BlockerInitCh

	// Start garbage collector
	gc, err := collector.New(mgr.GetClient())
	if err != nil {
		setupLog.Error(err, "error initializing garbage collector")
		os.Exit(1)
	}
	go gc.Start()

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
		Client:     mgr.GetClient(),
		Log:        ctrl.Log.WithName("controllers").WithName("Approval"),
		Scheme:     mgr.GetScheme(),
		MailSender: mail.NewSender(mgr.GetClient()),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Approval")
		os.Exit(1)
	}
	customRunController := &controllers.CustomRunReconciler{
		Client:          mgr.GetClient(),
		Log:             ctrl.Log.WithName("controllers").WithName("CustomRun"),
		Scheme:          mgr.GetScheme(),
		KindHandlerMap:  map[string]controllers.KindHandler{},
		HandlerChildren: map[string][]client.Object{},
	}
	// Add custom Run handlers
	customRunController.AddKindHandler(cicdv1.CustomTaskKindApproval, &customs.ApprovalRunHandler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ApprovalRun"),
		Scheme: mgr.GetScheme(),
	}, &cicdv1.Approval{})
	customRunController.AddKindHandler(cicdv1.CustomTaskKindEmail, &customs.EmailRunHandler{
		Client:     mgr.GetClient(),
		Log:        ctrl.Log.WithName("controllers").WithName("EmailRun"),
		Scheme:     mgr.GetScheme(),
		MailSender: mail.NewSender(mgr.GetClient()),
	})
	customRunController.AddKindHandler(cicdv1.CustomTaskKindSlack, &customs.SlackRunHandler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("SlackRun"),
		Scheme: mgr.GetScheme(),
	})
	customRunController.AddKindHandler(cicdv1.CustomTaskKindWebHook, &customs.WebhookRunHandler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("WebhookRun"),
		Scheme: mgr.GetScheme(),
	})
	if err = customRunController.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CustomRun")
	}
	// +kubebuilder:scaffold:builder

	// Start webhook expose controller
	setupLog.Info("Starting webhook expose controller")
	exposeCon, err := controllers.NewExposeController(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create expose controller")
		os.Exit(1)
	}
	go exposeCon.Start(nil)

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
