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
	"fmt"
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap/zapcore"
	appsV1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	corev1beta1 "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	"github.com/open-feature/open-feature-operator/common"
	"github.com/open-feature/open-feature-operator/common/flagdinjector"
	"github.com/open-feature/open-feature-operator/common/flagdproxy"
	"github.com/open-feature/open-feature-operator/common/types"
	"github.com/open-feature/open-feature-operator/controllers/core/featureflagsource"
	"github.com/open-feature/open-feature-operator/controllers/core/flagd"
	webhooks "github.com/open-feature/open-feature-operator/webhooks"
)

const (
	healthProbeBindAddressFlagName = "health-probe-bind-address"
	metricsBindAddressFlagName     = "metrics-bind-address"
	verboseFlagName                = "verbose"
	leaderElectFlagName            = "leader-elect"
	sidecarCpuLimitFlagName        = "sidecar-cpu-limit"
	sidecarRamLimitFlagName        = "sidecar-ram-limit"
	sidecarCpuRequestFlagName      = "sidecar-cpu-request"
	sidecarRamRequestFlagName      = "sidecar-ram-request"
	sidecarCpuLimitDefault         = "0.5"
	sidecarRamLimitDefault         = "64M"
	sidecarCpuRequestDefault       = "0.2"
	sidecarRamRequestDefault       = "32M"
)

var (
	scheme                                                                 = runtime.NewScheme()
	setupLog                                                               = ctrl.Log.WithName("setup")
	metricsAddr                                                            string
	enableLeaderElection                                                   bool
	probeAddr                                                              string
	verbose                                                                bool
	sidecarCpuLimit, sidecarRamLimit, sidecarCpuRequest, sidecarRamRequest string
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(corev1beta1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

//nolint:funlen,gocyclo
func main() {
	var env types.EnvConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("Failed to process env var: %s", err)
	}

	flag.StringVar(&metricsAddr, metricsBindAddressFlagName, ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, healthProbeBindAddressFlagName, ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&verbose, verboseFlagName, true, "Disable verbose logging")
	flag.BoolVar(&enableLeaderElection, leaderElectFlagName, false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	// the following default values are chosen as a result of load testing: https://github.com/open-feature/flagd/blob/main/tests/loadtest/README.MD#performance-observations
	flag.StringVar(&sidecarCpuLimit, sidecarCpuLimitFlagName, sidecarCpuLimitDefault, "sidecar CPU limit, in cores. (500m = .5 cores)")
	flag.StringVar(&sidecarRamLimit, sidecarRamLimitFlagName, sidecarRamLimitDefault, "sidecar memory limit, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)")
	flag.StringVar(&sidecarCpuRequest, sidecarCpuRequestFlagName, sidecarCpuRequestDefault, "sidecar CPU minimum, in cores. (500m = .5 cores)")
	flag.StringVar(&sidecarRamRequest, sidecarRamRequestFlagName, sidecarRamRequestDefault, "sidecar memory minimum, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)")

	flag.Parse()

	level := zapcore.InfoLevel
	if verbose {
		level = zapcore.DebugLevel
	}
	opts := zap.Options{
		Development: verbose,
		Level:       level,
	}
	opts.BindFlags(flag.CommandLine)

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	resources, err := processResources()
	if err != nil {
		os.Exit(1)
	}

	disableCacheFor := []client.Object{&v1.ClusterRoleBinding{}}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "131bf64c.openfeature.dev",
		ClientDisableCacheFor:  disableCacheFor,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// setup indexer for backfilling permissions on the flagd-kubernetes-sync role binding
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&corev1.Pod{},
		fmt.Sprintf("%s/%s", common.PodOpenFeatureAnnotationPath, common.AllowKubernetesSyncAnnotation),
		webhooks.OpenFeatureEnabledAnnotationIndex,
	); err != nil {
		setupLog.Error(
			err,
			"unable to create indexer",
			"webhook",
			fmt.Sprintf("%s/%s", common.PodOpenFeatureAnnotationPath, common.AllowKubernetesSyncAnnotation),
		)
		os.Exit(1)
	}

	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&appsV1.Deployment{},
		fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPath, common.FeatureFlagSourceAnnotation),
		common.FeatureFlagSourceIndex,
	); err != nil {
		setupLog.Error(
			err,
			"unable to create indexer",
			"webhook",
			fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPath, common.FeatureFlagSourceAnnotation),
		)
		os.Exit(1)
	}

	kph := flagdproxy.NewFlagdProxyHandler(
		flagdproxy.NewFlagdProxyConfiguration(env),
		mgr.GetClient(),
		ctrl.Log.WithName("FeatureFlagSource FlagdProxyHandler"),
	)

	flagSourceController := &featureflagsource.FeatureFlagSourceReconciler{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		Log:        ctrl.Log.WithName("FeatureFlagSource Controller"),
		FlagdProxy: kph,
	}
	if err = flagSourceController.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "FeatureFlagSource")
		os.Exit(1)
	}

	flagdContainerInjector := &flagdinjector.FlagdContainerInjector{
		Client:                    mgr.GetClient(),
		Logger:                    ctrl.Log.WithName("flagd-container injector"),
		FlagdProxyConfig:          kph.Config(),
		FlagdResourceRequirements: *resources,
		Image:                     env.SidecarImage,
		Tag:                       env.SidecarTag,
	}

	flagdControllerLogger := ctrl.Log.WithName("Flagd Controller")

	flagdResourceReconciler := &flagd.ResourceReconciler{
		Client: mgr.GetClient(),
		Log:    flagdControllerLogger,
	}
	flagdConfig := flagd.NewFlagdConfiguration(env)

	if err = (&flagd.FlagdReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		FlagdDeployment: &flagd.FlagdDeployment{
			Client:             mgr.GetClient(),
			Log:                flagdControllerLogger,
			FlagdInjector:      flagdContainerInjector,
			FlagdConfig:        flagdConfig,
			ResourceReconciler: flagdResourceReconciler,
		},
		FlagdService: &flagd.FlagdService{
			Client:             mgr.GetClient(),
			Log:                flagdControllerLogger,
			FlagdConfig:        flagdConfig,
			ResourceReconciler: flagdResourceReconciler,
		},
		FlagdIngress: &flagd.FlagdIngress{
			Client:             mgr.GetClient(),
			Log:                flagdControllerLogger,
			FlagdConfig:        flagdConfig,
			ResourceReconciler: flagdResourceReconciler,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Flagd")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder
	hookServer := mgr.GetWebhookServer()
	podMutator := &webhooks.PodMutator{
		Client:           mgr.GetClient(),
		Log:              ctrl.Log.WithName("mutating-pod-webhook"),
		FlagdProxyConfig: kph.Config(),
		Env:              env,
		FlagdInjector:    flagdContainerInjector,
	}
	hookServer.Register("/mutate-v1-pod", &webhook.Admission{Handler: podMutator})

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", podMutator.IsReady); err != nil {
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
	// backfill can be handled asynchronously, so we do not need to block via the channel
	go func() {
		if err := podMutator.BackfillPermissions(ctx); err != nil {
			setupLog.Error(err, "podMutator backfill permissions error")
		}
	}()

	if err := <-errChan; err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func processResources() (*corev1.ResourceRequirements, error) {
	cpuLimitResource, err := resource.ParseQuantity(sidecarCpuLimit)
	if err != nil {
		setupLog.Error(err, "parse sidecar cpu limit", sidecarCpuLimitFlagName, sidecarCpuLimit)
		return nil, err
	}

	ramLimitResource, err := resource.ParseQuantity(sidecarRamLimit)
	if err != nil {
		setupLog.Error(err, "parse sidecar ram limit", sidecarRamLimitFlagName, sidecarRamLimit)
		return nil, err
	}

	cpuRequestResource, err := resource.ParseQuantity(sidecarCpuRequest)
	if err != nil {
		setupLog.Error(err, "parse sidecar cpu request", sidecarCpuRequestFlagName, sidecarCpuRequest)
		return nil, err
	}

	ramRequestResource, err := resource.ParseQuantity(sidecarRamRequest)
	if err != nil {
		setupLog.Error(err, "parse sidecar ram request", sidecarRamRequestFlagName, sidecarRamRequest)
		return nil, err
	}

	if cpuRequestResource.Value() > cpuLimitResource.Value() ||
		ramRequestResource.Value() > ramLimitResource.Value() {
		setupLog.Error(err, "sidecar resource request is higher than the resource maximum")
		return nil, err
	}

	return &corev1.ResourceRequirements{
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    cpuLimitResource,
			corev1.ResourceMemory: ramLimitResource,
		},
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    cpuRequestResource,
			corev1.ResourceMemory: ramRequestResource,
		},
	}, nil
}
