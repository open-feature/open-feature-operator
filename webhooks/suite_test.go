package webhooks

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	corev1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	corev1alpha2 "github.com/open-feature/open-feature-operator/apis/core/v1alpha2"
	corev1alpha3 "github.com/open-feature/open-feature-operator/apis/core/v1alpha3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	admissionv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg                 *rest.Config
	k8sClient           client.Client
	testEnv             *envtest.Environment
	testCtx, testCancel = context.WithCancel(context.Background())
)

const (
	podMutatingWebhookPath                        = "/mutate-v1-pod"
	validatingFeatureFlagConfigurationWebhookPath = "/validate-v1alpha1-featureflagconfiguration"
	featureFlagConfigurationNamespace             = "test-validate-featureflagconfiguration"
	featureFlagSpec                               = `
	{
      "flags": {
        "new-welcome-message": {
          "state": "ENABLED",
          "variants": {
            "on": true,
            "off": false
          },
          "defaultVariant": "on"
		}
      }
    }`
)

func strPtr(s string) *string { return &s }

func TestAPIs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	RegisterFailHandler(Fail)

	SetDefaultEventuallyTimeout(time.Second * 15)
	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))

	By("bootstrapping test environment")
	mutateFailPolicy := admissionv1.Ignore
	validateFailPolicy := admissionv1.Fail
	mutateSideEffects := admissionv1.SideEffectClassNoneOnDryRun
	validateSideEffects := admissionv1.SideEffectClassNone
	webhookInstallOptions := envtest.WebhookInstallOptions{
		MutatingWebhooks: []*admissionv1.MutatingWebhookConfiguration{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "mutating-webhook-configuration",
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "MutatingWebhookConfiguration",
					APIVersion: "admissionregistration.k8s.io/v1",
				},
				Webhooks: []admissionv1.MutatingWebhook{
					{
						Name:                    "mutate.openfeature.dev",
						AdmissionReviewVersions: []string{"v1"},
						FailurePolicy:           &mutateFailPolicy,
						ClientConfig: admissionv1.WebhookClientConfig{
							Service: &admissionv1.ServiceReference{
								Name:      "webhook-service",
								Namespace: "system",
								Path:      strPtr(podMutatingWebhookPath),
							},
						},
						Rules: []admissionv1.RuleWithOperations{
							{
								Operations: []admissionv1.OperationType{
									admissionv1.Create,
									admissionv1.Update,
								},
								Rule: admissionv1.Rule{
									APIGroups:   []string{""},
									APIVersions: []string{"v1"},
									Resources:   []string{"pods"},
								},
							},
						},
						SideEffects: &mutateSideEffects,
					},
					{
						Name:                    "validate.featureflagconfiguration.openfeature.dev",
						AdmissionReviewVersions: []string{"v1"},
						FailurePolicy:           &validateFailPolicy,
						ClientConfig: admissionv1.WebhookClientConfig{
							Service: &admissionv1.ServiceReference{
								Name:      "webhook-service",
								Namespace: "system",
								Path:      strPtr(validatingFeatureFlagConfigurationWebhookPath),
							},
						},
						Rules: []admissionv1.RuleWithOperations{
							{
								Operations: []admissionv1.OperationType{
									admissionv1.Create,
									admissionv1.Update,
								},
								Rule: admissionv1.Rule{
									APIGroups:   []string{"core.openfeature.dev"},
									APIVersions: []string{"v1alpha1"},
									Resources:   []string{"featureflagconfigurations"},
								},
							},
						},
						SideEffects: &validateSideEffects,
					},
				},
			},
		},
	}

	scheme := runtime.NewScheme()
	err := clientgoscheme.AddToScheme(scheme)
	Expect(err).ToNot(HaveOccurred())

	err = corev1alpha1.AddToScheme(scheme)
	Expect(err).ToNot(HaveOccurred())

	err = corev1alpha2.AddToScheme(scheme)
	Expect(err).ToNot(HaveOccurred())

	err = corev1alpha3.AddToScheme(scheme)
	Expect(err).ToNot(HaveOccurred())

	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		WebhookInstallOptions: webhookInstallOptions,
		Scheme:                scheme,
		CRDInstallOptions:     envtest.CRDInstallOptions{Scheme: scheme},
	}

	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	// deploy 'before' resources
	By("setting up previously existing pod (BackfillPermissions test)")
	setupPreviouslyExistingPods()

	// setup webhook server
	By("Setup webhook server")

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "localhost:8999",
		LeaderElection:     false,
		Host:               testEnv.WebhookInstallOptions.LocalServingHost,
		Port:               testEnv.WebhookInstallOptions.LocalServingPort,
		CertDir:            testEnv.WebhookInstallOptions.LocalServingCertDir,
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&corev1alpha1.FeatureFlagConfiguration{}).SetupWebhookWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&corev1alpha1.FlagSourceConfiguration{}).SetupWebhookWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&corev1.Pod{},
		fmt.Sprintf("%s/%s", OpenFeatureAnnotationPath, AllowKubernetesSyncAnnotation),
		OpenFeatureEnabledAnnotationIndex,
	)
	Expect(err).ToNot(HaveOccurred())

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

	// start webhook server
	By("running webhook server")

	go func() {
		err := mgr.Start(testCtx)
		Expect(err).ToNot(HaveOccurred())
	}()

	err = podMutator.BackfillPermissions(testCtx)
	Expect(err).ToNot(HaveOccurred())

	// wait for webhook to be ready to accept connections
	d := &net.Dialer{Timeout: time.Second}
	Eventually(func() error {
		serverURL := fmt.Sprintf("%s:%d", testEnv.WebhookInstallOptions.LocalServingHost, testEnv.WebhookInstallOptions.LocalServingPort)
		conn, err := tls.DialWithDialer(d, "tcp", serverURL, &tls.Config{
			InsecureSkipVerify: true,
		})
		if err != nil {
			return err
		}
		if err := conn.Close(); err != nil {
			return err
		}
		return nil
	}).Should(Succeed())

	// wait for ready state
	Eventually(func() error {
		return podMutator.IsReady(nil)
	}).Should(Succeed())

	By("setting up resources")
	setupMutatePodResources()
	setupValidateFeatureFlagConfigurationResources()

})

func setupValidateFeatureFlagConfigurationResources() {
	ns := &corev1.Namespace{}
	ns.Name = featureFlagConfigurationNamespace
	err := k8sClient.Create(testCtx, ns)
	Expect(err).ShouldNot(HaveOccurred())
}

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	testCancel()
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())

	testCtx, testCancel = context.WithCancel(context.Background())
})
