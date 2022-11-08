package webhooks

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	"net"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var testCtx, testCancel = context.WithCancel(context.Background())

const (
	podMutatingWebhookPath                        = "/mutate-v1-pod"
	validatingFeatureFlagConfigurationWebhookPath = "/validate-v1alpha1-featureflagconfiguration"
)

func strPtr(s string) *string { return &s }

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	SetDefaultEventuallyTimeout(time.Minute)
	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
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

	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		WebhookInstallOptions: webhookInstallOptions,
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	scheme := runtime.NewScheme()
	err = clientgoscheme.AddToScheme(scheme)
	Expect(err).ToNot(HaveOccurred())

	err = v1alpha1.AddToScheme(scheme)
	Expect(err).ToNot(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	By("running webhook server")
	go func() {
		if err := run(testCtx, cfg, scheme, &testEnv.WebhookInstallOptions); err != nil {
			logf.Log.Error(err, "run webhook server")
		}
	}()
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

	By("setting up resources")
	setupMutatePodResources()
	setupValidateFeatureFlagConfigurationResources()
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	testCancel()
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())

	testCtx, testCancel = context.WithCancel(context.Background())
})
