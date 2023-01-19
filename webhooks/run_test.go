package webhooks

import (
	"context"

	corev1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	corev1alpha2 "github.com/open-feature/open-feature-operator/apis/core/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
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

	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&corev1.Pod{},
		OpenFeatureEnabledAnnotationPath,
		OpenFeatureEnabledAnnotationIndex,
	); err != nil {
		return err
	}

	// +kubebuilder:scaffold:builder

	wh := mgr.GetWebhookServer()
	podMutator := &PodMutator{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("mutating-pod-webhook"),
	}
	wh.Register(podMutatingWebhookPath, &webhook.Admission{Handler: podMutator})
	wh.Register(validatingFeatureFlagConfigurationWebhookPath, &webhook.Admission{
		Handler: &FeatureFlagConfigurationValidator{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("validating-featureflagconfiguration-webhook"),
		},
	})

	// podMutator.BackfillPermissions is dependent upon mgr.Start executing correctly
	// due to its time.Sleep within the retry loop, mgr.Start will always fail before podMutator.BackfillPermissions
	// times out, resulting in the most relevant error being passed into the errChan
	// mgr.Start will also only output an error value of nil once the context is closed (this occurs when the test suite is terminating)
	errChan := make(chan error, 1)
	go func() {
		if err := podMutator.BackfillPermissions(ctx); err != nil {
			errChan <- err
		}
	}()
	go func() {
		errChan <- mgr.Start(ctx)
	}()

	return <-errChan
}
