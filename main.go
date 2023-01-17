/*
Copyright 2022.

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
	"context"
	"flag"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	corev1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	corev1alpha2 "github.com/open-feature/open-feature-operator/apis/core/v1alpha2"
	"github.com/open-feature/open-feature-operator/controllers"
	webhooks "github.com/open-feature/open-feature-operator/webhooks"
	//+kubebuilder:scaffold:imports
)

var (
	scheme                                                         = runtime.NewScheme()
	setupLog                                                       = ctrl.Log.WithName("setup")
	metricsAddr                                                    string
	enableLeaderElection                                           bool
	probeAddr                                                      string
	verbose                                                        bool
	flagDCpuLimit, flagDRamLimit, flagDCpuRequest, flagDRamRequest string
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(corev1alpha1.AddToScheme(scheme))
	utilruntime.Must(corev1alpha2.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&verbose, "verbose", true, "Disable verbose logging")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	// the following default values are chosen as a result of load testing: https://github.com/open-feature/flagd/blob/main/tests/loadtest/README.MD#performance-observations
	flag.StringVar(&flagDCpuLimit, "flagd-cpu-limit", "0.5", "flagd CPU limit, in cores. (500m = .5 cores)")
	flag.StringVar(&flagDRamLimit, "flagd-ram-limit", "64M", "flagd memory limit, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)")
	flag.StringVar(&flagDCpuRequest, "flagd-cpu-request", "0.2", "flagd CPU minimum, in cores. (500m = .5 cores)")
	flag.StringVar(&flagDRamRequest, "flagd-ram-request", "32M", "flagd memory minimum, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)")

	level := zapcore.InfoLevel
	if verbose {
		level = zapcore.DebugLevel
	}
	opts := zap.Options{
		Development: verbose,
		Level:       level,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	flagDCpuLimitResource, err := resource.ParseQuantity(flagDCpuLimit)
	if err != nil {
		setupLog.Error(err, "parse flagd cpu limit", "flagd-cpu-limit", flagDCpuLimit)
		os.Exit(1)
	}

	flagDRamLimitResource, err := resource.ParseQuantity(flagDRamLimit)
	if err != nil {
		setupLog.Error(err, "parse flagd ram limit", "flagd-ram-limit", flagDRamLimit)
		os.Exit(1)
	}

	flagDCpuRequestResource, err := resource.ParseQuantity(flagDCpuRequest)
	if err != nil {
		setupLog.Error(err, "parse flagd cpu request", "flagd-cpu-request", flagDCpuRequest)
		os.Exit(1)
	}

	flagDRamRequestResource, err := resource.ParseQuantity(flagDRamRequest)
	if err != nil {
		setupLog.Error(err, "parse flagd ram request", "flagd-ram-request", flagDRamRequest)
		os.Exit(1)
	}

	if flagDCpuRequestResource.Value() > flagDCpuLimitResource.Value() ||
		flagDRamRequestResource.Value() > flagDRamLimitResource.Value() {
		setupLog.Error(err, "flagd resource request is higher than the resource maximum")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "131bf64c.openfeature.dev",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// setup indexer for backfilling permissions on the flagd-kubernetes-sync role binding
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&corev1.Pod{},
		webhooks.OpenFeatureEnabledAnnotationPath,
		webhooks.OpenFeatureEnabledAnnotationIndex,
	); err != nil {
		setupLog.Error(err, "unable to create indexer", "webhook", webhooks.OpenFeatureEnabledAnnotationPath)
		os.Exit(1)
	}

	if err = (&controllers.FeatureFlagConfigurationReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "FeatureFlagConfiguration")
		os.Exit(1)
	}

	if err := (&corev1alpha1.FeatureFlagConfiguration{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "FeatureFlagConfiguration")
		os.Exit(1)
	}
	if err := (&corev1alpha2.FeatureFlagConfiguration{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "FeatureFlagConfiguration")
		os.Exit(1)
	}

	if err = (&controllers.FlagSourceConfigurationReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "FlagSourceConfiguration")
		os.Exit(1)
	}

	if err := (&corev1alpha1.FlagSourceConfiguration{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "FlagSourceConfiguration")
		os.Exit(1)
	}
	if err := (&corev1alpha2.FlagSourceConfiguration{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "FlagSourceConfiguration")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder
	hookServer := mgr.GetWebhookServer()
	podMutator := &webhooks.PodMutator{
		FlagDResourceRequirements: corev1.ResourceRequirements{
			Limits: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU:    flagDCpuLimitResource,
				corev1.ResourceMemory: flagDRamLimitResource,
			},
			Requests: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU:    flagDCpuRequestResource,
				corev1.ResourceMemory: flagDRamRequestResource,
			},
		},
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("mutating-pod-webhook"),
	}
	hookServer.Register("/mutate-v1-pod", &webhook.Admission{Handler: podMutator})
	hookServer.Register("/validate-v1alpha1-featureflagconfiguration", &webhook.Admission{Handler: &webhooks.FeatureFlagConfigurationValidator{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("validating-featureflagconfiguration-webhook"),
	}})

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	ctx := ctrl.SetupSignalHandler()
	errChan := make(chan error, 1)
	go func(chan error) {
		if err := mgr.Start(ctx); err != nil {
			errChan <- err
		}
	}(errChan)

	setupLog.Info("restoring flagd-kubernetes-sync cluster role binding subjects from current cluster state")
	go podMutator.BackfillPermissions(ctx)

	if err := <-errChan; err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
