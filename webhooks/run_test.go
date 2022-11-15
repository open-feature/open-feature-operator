package webhooks

import (
	"context"
	corev1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	corev1alpha2 "github.com/open-feature/open-feature-operator/apis/core/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

func run(ctx context.Context, cfg *rest.Config, scheme *runtime.Scheme, opts *envtest.WebhookInstallOptions) error {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "localhost:8999",
		LeaderElection:     false,
		Host:               opts.LocalServingHost,
		Port:               opts.LocalServingPort,
		CertDir:            opts.LocalServingCertDir,
	})
	if err != nil {
		return err
	}

	if err := (&corev1alpha1.FeatureFlagConfiguration{}).SetupWebhookWithManager(mgr); err != nil {
		return err
	}
	if err := (&corev1alpha2.FeatureFlagConfiguration{}).SetupWebhookWithManager(mgr); err != nil {
		return err
	}

	// +kubebuilder:scaffold:builder

	wh := mgr.GetWebhookServer()
	wh.Register(podMutatingWebhookPath, &webhook.Admission{Handler: &PodMutator{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("mutating-pod-webhook"),
	}})
	wh.Register(validatingFeatureFlagConfigurationWebhookPath, &webhook.Admission{
		Handler: &FeatureFlagConfigurationValidator{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("validating-featureflagconfiguration-webhook"),
		},
	})

	if err := mgr.Start(ctx); err != nil {
		return err
	}
	return nil
}
