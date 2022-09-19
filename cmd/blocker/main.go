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

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/controllers"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/logrotate"
	"github.com/tmax-cloud/cicd-operator/pkg/blocker"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
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
	// +kubebuilder:scaffold:scheme
}

func main() {
	var healthAddr string
	opts := zap.Options{
		Development: false,
	}
	opts.BindFlags(flag.CommandLine)

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
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts), zap.WriteTo(logWriter)))
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
		setupLog.Error(err, "unable to start Manager")
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
	cfgCtrl.Add(configs.ConfigMapNameBlockerConfig, configs.ApplyBlockerConfigChange)
	// Wait for initial config reconcile
	<-configs.BlockerInitCh

	// Blocker
	b := blocker.New(mgr.GetClient())
	go b.Start()
	go b.StartBlockerStatusServer()

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
