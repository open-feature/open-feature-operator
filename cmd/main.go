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
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	corev1beta1 "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	"github.com/open-feature/open-feature-operator/internal/common"
	"github.com/open-feature/open-feature-operator/internal/common/flagdinjector"
	"github.com/open-feature/open-feature-operator/internal/common/flagdproxy"
	"github.com/open-feature/open-feature-operator/internal/common/types"
	"github.com/open-feature/open-feature-operator/internal/common/utils"
	"github.com/open-feature/open-feature-operator/internal/controller/core/featureflagsource"
	"github.com/open-feature/open-feature-operator/internal/controller/core/flagd"
	flagdResources "github.com/open-feature/open-feature-operator/internal/controller/core/flagd/resources"
	webhooks "github.com/open-feature/open-feature-operator/internal/webhook"
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
	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	gatewayApiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	healthProbeBindAddressFlagName = "health-probe-bind-address"
	metricsBindAddressFlagName     = "metrics-bind-address"
	verboseFlagName                = "verbose"
	leaderElectFlagName            = "leader-elect"

	sidecarCpuLimitFlagName = "sidecar-cpu-limit"
	sidecarCpuLimitDefault  = "0.5"

	sidecarRamLimitFlagName = "sidecar-ram-limit"
	sidecarRamLimitDefault  = "64M"

	sidecarCpuRequestFlagName = "sidecar-cpu-request"
	sidecarCpuRequestDefault  = "0.2"

	sidecarRamRequestFlagName = "sidecar-ram-request"
	sidecarRamRequestDefault  = "32M"

	imagePullSecretFlagName    = "image-pull-secrets"
	imagePullSecretFlagDefault = ""

	labelsFlagName    = "labels"
	labelsFlagDefault = ""

	annotationsFlagName    = "annotations"
	annotationsFlagDefault = ""
)

var (
	scheme                                                                 = runtime.NewScheme()
	setupLog                                                               = ctrl.Log.WithName("setup")
	metricsAddr                                                            string
	metricsCertPath, metricsCertName, metricsCertKey                       string
	webhookCertPath, webhookCertName, webhookCertKey                       string
	secureMetrics                                                          bool
	enableHTTP2                                                            bool
	tlsOpts                                                                []func(*tls.Config)
	enableLeaderElection                                                   bool
	probeAddr                                                              string
	verbose                                                                bool
	sidecarCpuLimit, sidecarRamLimit, sidecarCpuRequest, sidecarRamRequest string
	imagePullSecrets                                                       string
	labels                                                                 string
	annotations                                                            string
)

// StringToMap transforms a string into a map[string]string
func StringToMap(s string) map[string]string {
	m := map[string]string{}
	for _, pair := range strings.Split(s, ",") {
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) == 2 {
			m[kv[0]] = kv[1]
		}
	}
	return m
}

// CommaSeparatedStringToSlice transforms a comma-separated string into a slice of strings
func CommaSeparatedStringToSlice(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(corev1beta1.AddToScheme(scheme))
	utilruntime.Must(gatewayApiv1.Install(scheme))
	//+kubebuilder:scaffold:scheme
}

//nolint:funlen,gocyclo,gocognit
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

	flag.BoolVar(&secureMetrics, "metrics-secure", true,
		"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
	flag.StringVar(&webhookCertPath, "webhook-cert-path", "", "The directory that contains the webhook certificate.")
	flag.StringVar(&webhookCertName, "webhook-cert-name", "tls.crt", "The name of the webhook certificate file.")
	flag.StringVar(&webhookCertKey, "webhook-cert-key", "tls.key", "The name of the webhook key file.")
	flag.StringVar(&metricsCertPath, "metrics-cert-path", "",
		"The directory that contains the metrics server certificate.")
	flag.StringVar(&metricsCertName, "metrics-cert-name", "tls.crt", "The name of the metrics server certificate file.")
	flag.StringVar(&metricsCertKey, "metrics-cert-key", "tls.key", "The name of the metrics server key file.")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")

	// the following default values are chosen as a result of load testing: https://github.com/open-feature/flagd/blob/main/tests/loadtest/README.MD#performance-observations
	flag.StringVar(&sidecarCpuLimit, sidecarCpuLimitFlagName, sidecarCpuLimitDefault, "sidecar CPU limit, in cores. (500m = .5 cores)")
	flag.StringVar(&sidecarRamLimit, sidecarRamLimitFlagName, sidecarRamLimitDefault, "sidecar memory limit, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)")
	flag.StringVar(&sidecarCpuRequest, sidecarCpuRequestFlagName, sidecarCpuRequestDefault, "sidecar CPU minimum, in cores. (500m = .5 cores)")
	flag.StringVar(&sidecarRamRequest, sidecarRamRequestFlagName, sidecarRamRequestDefault, "sidecar memory minimum, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)")
	flag.StringVar(&imagePullSecrets, imagePullSecretFlagName, imagePullSecretFlagDefault, "Comma-delimited list of secrets containing credentials to pull images.")
	flag.StringVar(&labels, labelsFlagName, labelsFlagDefault, "Map of labels to add to the deployed pods. Formatted like key1:value1,key2:value2,key3:value3")
	flag.StringVar(&annotations, annotationsFlagName, annotationsFlagDefault, "Map of annotations to add to the deployed pods. Formatted like key1:value1,key2:value2,key3:value3")

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

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	// Create watchers for metrics and webhooks certificates
	var metricsCertWatcher, webhookCertWatcher *certwatcher.CertWatcher

	// Initial webhook TLS options
	webhookTLSOpts := tlsOpts

	if len(webhookCertPath) > 0 {
		setupLog.Info("Initializing webhook certificate watcher using provided certificates",
			"webhook-cert-path", webhookCertPath, "webhook-cert-name", webhookCertName, "webhook-cert-key", webhookCertKey)

		var err error
		webhookCertWatcher, err = certwatcher.New(
			filepath.Join(webhookCertPath, webhookCertName),
			filepath.Join(webhookCertPath, webhookCertKey),
		)
		if err != nil {
			setupLog.Error(err, "Failed to initialize webhook certificate watcher")
			os.Exit(1)
		}

		webhookTLSOpts = append(webhookTLSOpts, func(config *tls.Config) {
			config.GetCertificate = webhookCertWatcher.GetCertificate
		})
	}

	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts: webhookTLSOpts,
		Port:    9443,
	})

	// Metrics endpoint is enabled in 'config/default/kustomization.yaml'. The Metrics options configure the server.
	// More info:
	// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.0/pkg/metrics/server
	// - https://book.kubebuilder.io/reference/metrics.html
	metricsServerOptions := metricsserver.Options{
		BindAddress:   metricsAddr,
		SecureServing: secureMetrics,
		TLSOpts:       tlsOpts,
	}

	if secureMetrics {
		// FilterProvider is used to protect the metrics endpoint with authn/authz.
		// These configurations ensure that only authorized users and service accounts
		// can access the metrics endpoint. The RBAC are configured in 'config/rbac/kustomization.yaml'. More info:
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.0/pkg/metrics/filters#WithAuthenticationAndAuthorization
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	// If the certificate is not specified, controller-runtime will automatically
	// generate self-signed certificates for the metrics server. While convenient for development and testing,
	// this setup is not recommended for production.
	//
	// - [METRICS-WITH-CERTS] at config/default/kustomization.yaml to generate and use certificates
	// managed by cert-manager for the metrics server.
	// - [PROMETHEUS-WITH-CERTS] at config/prometheus/kustomization.yaml for TLS certification.
	if len(metricsCertPath) > 0 {
		setupLog.Info("Initializing metrics certificate watcher using provided certificates",
			"metrics-cert-path", metricsCertPath, "metrics-cert-name", metricsCertName, "metrics-cert-key", metricsCertKey)

		var err error
		metricsCertWatcher, err = certwatcher.New(
			filepath.Join(metricsCertPath, metricsCertName),
			filepath.Join(metricsCertPath, metricsCertKey),
		)
		if err != nil {
			setupLog.Error(err, "to initialize metrics certificate watcher", "error", err)
			os.Exit(1)
		}

		metricsServerOptions.TLSOpts = append(metricsServerOptions.TLSOpts, func(config *tls.Config) {
			config.GetCertificate = metricsCertWatcher.GetCertificate
		})
	}

	resources, err := processResources()
	if err != nil {
		os.Exit(1)
	}

	disableCacheFor := []client.Object{&v1.ClusterRoleBinding{}}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: server.Options{
			BindAddress: metricsAddr,
		},

		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "131bf64c.openfeature.dev",
		Client: ctrlclient.Options{
			Cache: &ctrlclient.CacheOptions{
				DisableFor: disableCacheFor,
			},
		},
		WebhookServer: webhookServer,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if metricsCertWatcher != nil {
		setupLog.Info("Adding metrics certificate watcher to manager")
		if err := mgr.Add(metricsCertWatcher); err != nil {
			setupLog.Error(err, "unable to add metrics certificate watcher to manager")
			os.Exit(1)
		}
	}

	if webhookCertWatcher != nil {
		setupLog.Info("Adding webhook certificate watcher to manager")
		if err := mgr.Add(webhookCertWatcher); err != nil {
			setupLog.Error(err, "unable to add webhook certificate watcher to manager")
			os.Exit(1)
		}
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

	labelsMap := StringToMap(labels)
	annotationsMap := StringToMap(annotations)

	kph := flagdproxy.NewFlagdProxyHandler(
		flagdproxy.NewFlagdProxyConfiguration(
			env,
			CommaSeparatedStringToSlice(imagePullSecrets),
			labelsMap,
			annotationsMap,
		),
		mgr.GetClient(),
		ctrl.Log.WithName("FeatureFlagSource FlagdProxyHandler"),
	)

	flagSourceController := &featureflagsource.FeatureFlagSourceReconciler{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		Log:        ctrl.Log.WithName("FeatureFlagSource Controller"),
		FlagdProxy: kph,
		FlagdProxyBackoff: &utils.ExponentialBackoff{
			StartDelay: time.Second,
			MaxDelay:   time.Minute,
		},
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
		Scheme: mgr.GetScheme(),
		Log:    flagdControllerLogger,
	}

	flagdConfig := flagd.NewFlagdConfiguration(
		env,
		CommaSeparatedStringToSlice(imagePullSecrets),
		labelsMap,
		annotationsMap,
	)

	if err = (&flagd.FlagdReconciler{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		ResourceReconciler: flagdResourceReconciler,
		FlagdDeployment: &flagdResources.FlagdDeployment{
			Client:        mgr.GetClient(),
			Log:           flagdControllerLogger,
			FlagdInjector: flagdContainerInjector,
			FlagdConfig:   flagdConfig,
		},
		FlagdService: &flagdResources.FlagdService{
			FlagdConfig: flagdConfig,
		},
		FlagdIngress: &flagdResources.FlagdIngress{
			FlagdConfig: flagdConfig,
		},
		FlagdGatewayApiHttpRoute: &flagdResources.FlagdGatewayApiHttpRoute{
			FlagdConfig: flagdConfig,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Flagd")
		os.Exit(1)
	}

	if env.FlagsValidationEnabled {
		if err = (&webhooks.FeatureFlagCustomValidator{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create the validation webhook for FeatureFlag CRD", "webhook", "FeatureFlag")
			os.Exit(1)
		}
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
	if err := podMutator.InjectDecoder(admission.NewDecoder(mgr.GetScheme())); err != nil {
		setupLog.Error(err, "unable to inject decoder into mutating webhook")
		os.Exit(1)
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
