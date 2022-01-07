package main

import (
	"flag"
	"fmt"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/controllers"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/logrotate"
	"github.com/tmax-cloud/cicd-operator/pkg/chatops"
	"github.com/tmax-cloud/cicd-operator/pkg/chatops/plugins/approve"
	"github.com/tmax-cloud/cicd-operator/pkg/chatops/plugins/hold"
	"github.com/tmax-cloud/cicd-operator/pkg/chatops/plugins/trigger"
	"github.com/tmax-cloud/cicd-operator/pkg/dispatcher"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"github.com/tmax-cloud/cicd-operator/pkg/plugins/size"
	"github.com/tmax-cloud/cicd-operator/pkg/server"
	"io"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(cicdv1.AddToScheme(scheme))
	utilruntime.Must(tektonv1beta1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var healthAddr string
	flag.StringVar(&healthAddr, "health-addr", ":8888", "The address the health endpoint binds to.")
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
	if err := logrotate.StartRotate("0 0 1 * * ?"); err != nil {
		setupLog.Error(err, "")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     "0",
		HealthProbeBindAddress: healthAddr,
		Port:                   9443,
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
	cfgCtrl, err := controllers.NewConfigReconciler(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to initiate config reconciler")
		os.Exit(1)
	}
	go cfgCtrl.Start()
	cfgCtrl.Add(configs.ConfigMapNameCICDConfig, configs.ApplyControllerConfigChange)
	// Wait for initial config reconcile
	<-configs.ControllerInitCh

	// Init chat-ops
	co := chatops.New(mgr.GetClient())
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Init chat-ops plugins
	approveHandler := &approve.Handler{Client: mgr.GetClient()}
	triggerHandler := &trigger.Handler{Client: mgr.GetClient()}
	holdHandler := &hold.Handler{Client: mgr.GetClient()}

	co.RegisterCommandHandler(approve.CommandTypeApprove, approveHandler.HandleChatOps)
	co.RegisterCommandHandler(approve.CommandTypeGitLabApprove, approveHandler.HandleChatOps)
	co.RegisterCommandHandler(trigger.CommandTypeTest, triggerHandler.HandleChatOps)
	co.RegisterCommandHandler(trigger.CommandTypeRetest, triggerHandler.HandleChatOps)
	co.RegisterCommandHandler(hold.CommandTypeHold, holdHandler.HandleChatOps)

	// Create and start webhook server
	srv := server.New(mgr.GetClient(), mgr.GetConfig())
	// Add plugins for webhook
	server.AddPlugin([]git.EventType{git.EventTypePullRequest, git.EventTypePush}, &dispatcher.Dispatcher{Client: mgr.GetClient()})
	server.AddPlugin([]git.EventType{git.EventTypeIssueComment, git.EventTypePullRequestReview, git.EventTypePullRequestReviewComment}, co)
	server.AddPlugin([]git.EventType{git.EventTypePullRequest, git.EventTypePullRequestReview}, approveHandler)
	server.AddPlugin([]git.EventType{git.EventTypePullRequest}, &size.Size{Client: mgr.GetClient()})
	go srv.Start()

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
