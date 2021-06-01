package main

import (
	"fmt"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/controllers"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/logrotate"
	"github.com/tmax-cloud/cicd-operator/pkg/blocker"
	"io"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(cicdv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
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
		Scheme:             scheme,
		MetricsBindAddress: "0",
		Port:               9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start Manager")
		os.Exit(1)
	}

	// Config Controller
	cfgCtrl := &controllers.ConfigReconciler{Log: ctrl.Log.WithName("controllers").WithName("ConfigController"), Handlers: map[string]configs.Handler{}}
	go cfgCtrl.Start()
	cfgCtrl.Add(configs.ConfigMapNameBlockerConfig, configs.ApplyBlockerConfigChange)
	// Wait for initial config reconcile
	<-configs.InitCh

	// Blocker
	b := blocker.New(mgr.GetClient())
	go b.Start()
	go b.StartBlockerStatusServer()

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
