package webhooks

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	featureFlagConfigurationNamespace = "test-validate-featureflagconfiguration"
)

var featureFlagSpec = `
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
    }
	`

func setupValidateFeatureFlagConfigurationResources() {
	ns := &corev1.Namespace{}
	ns.Name = featureFlagConfigurationNamespace
	err := k8sClient.Create(testCtx, ns)
	Expect(err).ShouldNot(HaveOccurred())
}

func featureflagconfigurationCleanup() {
	ffConfig := &corev1alpha1.FeatureFlagConfiguration{}
	ffConfig.Namespace = featureFlagConfigurationNamespace
	ffConfig.Name = featureFlagConfigurationName
	err := k8sClient.Delete(testCtx, ffConfig, client.GracePeriodSeconds(0))
	Expect(err).ShouldNot(HaveOccurred())
}

var _ = Describe("featureflagconfiguration validation webhook", func() {
	It("should pass when featureflagspec contains valid json", func() {
		ffConfig := &corev1alpha1.FeatureFlagConfiguration{}
		ffConfig.Namespace = featureFlagConfigurationNamespace
		ffConfig.Name = featureFlagConfigurationName
		ffConfig.Spec.FeatureFlagSpec = featureFlagSpec
		err := k8sClient.Create(testCtx, ffConfig)
		Expect(err).ShouldNot(HaveOccurred())

		featureflagconfigurationCleanup()
	})

	It("should fail when featureflagspec contains invalid json", func() {
		ffConfig := &corev1alpha1.FeatureFlagConfiguration{}
		ffConfig.Namespace = featureFlagConfigurationNamespace
		ffConfig.Name = featureFlagConfigurationName
		ffConfig.Spec.FeatureFlagSpec = `{"invalid":json}`
		err := k8sClient.Create(testCtx, ffConfig)
		Expect(err).Should(HaveOccurred())
	})

	It("should fail when featureflagspec doesn't conform to the schema", func() {
		ffConfig := &corev1alpha1.FeatureFlagConfiguration{}
		ffConfig.Namespace = featureFlagConfigurationNamespace
		ffConfig.Name = featureFlagConfigurationName
		ffConfig.Spec.FeatureFlagSpec = `
			{
				"flags":{
					"foo": {}
				}
			}
		`
		err := k8sClient.Create(testCtx, ffConfig)
		Expect(err).Should(HaveOccurred())
	})

	It("should check for existence of provider secret when service provider is given", func() {
		const credentialsName = "credentials-name"
		providerKeySecret := &corev1.Secret{}
		providerKeySecret.Name = credentialsName
		providerKeySecret.Namespace = featureFlagConfigurationNamespace
		err := k8sClient.Create(testCtx, providerKeySecret)
		Expect(err).ShouldNot(HaveOccurred())

		ffConfig := &corev1alpha1.FeatureFlagConfiguration{}
		ffConfig.Namespace = featureFlagConfigurationNamespace
		ffConfig.Name = featureFlagConfigurationName
		ffConfig.Spec.FeatureFlagSpec = featureFlagSpec
		ffConfig.Spec.ServiceProvider = &corev1alpha1.FeatureFlagServiceProvider{
			Name: "flagd",
			Credentials: &corev1.ObjectReference{
				Name:      credentialsName,
				Namespace: featureFlagConfigurationNamespace,
			},
		}
		err = k8sClient.Create(testCtx, ffConfig)
		Expect(err).ShouldNot(HaveOccurred())

		featureflagconfigurationCleanup()

		// cleanup secret
		err = k8sClient.Delete(testCtx, providerKeySecret)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("should fail if provider secret doesn't exist when service provider is given", func() {
		const credentialsName = "credentials-name"

		ffConfig := &corev1alpha1.FeatureFlagConfiguration{}
		ffConfig.Namespace = featureFlagConfigurationNamespace
		ffConfig.Name = featureFlagConfigurationName
		ffConfig.Spec.FeatureFlagSpec = featureFlagSpec
		ffConfig.Spec.ServiceProvider = &corev1alpha1.FeatureFlagServiceProvider{
			Name: "flagd",
			Credentials: &corev1.ObjectReference{
				Name:      credentialsName,
				Namespace: featureFlagConfigurationNamespace,
			},
		}
		err := k8sClient.Create(testCtx, ffConfig)
		Expect(err).Should(HaveOccurred())
	})
})
