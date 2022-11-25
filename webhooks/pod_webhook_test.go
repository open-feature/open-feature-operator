package webhooks

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	mutatePodNamespace           = "test-mutate-pod"
	defaultPodName               = "test-pod"
	defaultPodServiceAccountName = "test-pod-service-account"
	featureFlagConfigurationName = "test-feature-flag-configuration"
)

func setupMutatePodResources() {
	// Namespace and namespace resources
	ns := &corev1.Namespace{}
	ns.Name = mutatePodNamespace
	err := k8sClient.Create(testCtx, ns)
	Expect(err).ShouldNot(HaveOccurred())

	svcAccount := &corev1.ServiceAccount{}
	svcAccount.Namespace = mutatePodNamespace
	svcAccount.Name = defaultPodServiceAccountName
	err = k8sClient.Create(testCtx, svcAccount)
	Expect(err).ShouldNot(HaveOccurred())

	clusterRoleBinding := &v1.ClusterRoleBinding{}
	clusterRoleBinding.Namespace = mutatePodNamespace
	clusterRoleBinding.Name = clusterRoleBindingName
	clusterRoleBinding.APIVersion = "rbac.authorization.k8s.io/v1"
	clusterRoleBinding.RoleRef = v1.RoleRef{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     clusterRoleBindingName,
	}
	err = k8sClient.Create(testCtx, clusterRoleBinding)
	Expect(err).ShouldNot(HaveOccurred())

	ffConfig := &corev1alpha1.FeatureFlagConfiguration{}
	ffConfig.Namespace = mutatePodNamespace
	ffConfig.Name = featureFlagConfigurationName
	ffConfig.Spec.FlagDSpec = &corev1alpha1.FlagDSpec{Envs: []corev1.EnvVar{
		{Name: "LOG_LEVEL", Value: "dev"},
	}}
	ffConfig.Spec.FeatureFlagSpec = featureFlagSpec
	err = k8sClient.Create(testCtx, ffConfig)
	Expect(err).ShouldNot(HaveOccurred())
}

func testPod() *corev1.Pod {
	pod := &corev1.Pod{}
	pod.Namespace = mutatePodNamespace
	pod.Name = defaultPodName
	pod.Annotations = map[string]string{
		"openfeature.dev":                          "enabled",
		"openfeature.dev/featureflagconfiguration": featureFlagConfigurationName,
	}
	pod.Spec.Containers = []corev1.Container{
		{
			Name:  "container1",
			Image: "ubuntu",
		},
	}
	pod.Spec.ServiceAccountName = defaultPodServiceAccountName

	// In reality something like a Deployment would take ownership of pod creation.
	// A limitation of envtest is that inbuilt kubernetes controllers like deployment controllers aren't available.
	// Below simulates a pod that has ownership.
	pod.OwnerReferences = []metav1.OwnerReference{
		{
			Name:       "simulated-owner",
			Kind:       "deployment",
			APIVersion: "v1",
			UID:        "1f08bbbf-edb4-452a-9ffd-1898dd24c5b8",
		},
	}
	return pod
}

func getPod() *corev1.Pod {
	pod := &corev1.Pod{}
	name := types.NamespacedName{
		Namespace: mutatePodNamespace,
		Name:      defaultPodName,
	}
	err := k8sClient.Get(testCtx, name, pod)
	ExpectWithOffset(1, err).ShouldNot(HaveOccurred())
	return pod
}

func podMutationWebhookCleanup() {
	pod := &corev1.Pod{}
	pod.Namespace = mutatePodNamespace
	pod.Name = defaultPodName
	err := k8sClient.Delete(testCtx, pod, client.GracePeriodSeconds(0))
	Expect(err).ShouldNot(HaveOccurred())
}

var _ = Describe("pod mutation webhook", func() {

	It("should create flagd sidecar", func() {
		pod := testPod()
		err := k8sClient.Create(testCtx, pod)
		Expect(err).ShouldNot(HaveOccurred())

		pod = getPod()

		Expect(len(pod.Spec.Containers)).To(Equal(2))
		Expect(pod.Spec.Containers[1].Name).To(Equal("flagd"))
		Expect(pod.Spec.Containers[1].Image).To(Equal("ghcr.io/open-feature/flagd:" + FlagDTag))
		Expect(pod.Spec.Containers[1].Args).To(Equal([]string{
			"start", "--uri", "/etc/flagd/config.json", "--sync-provider", "kubernetes",
			"--sync-provider-args=namespace=" + mutatePodNamespace,
			"--sync-provider-args=featureflagconfiguration=" + featureFlagConfigurationName,
		}))
		Expect(pod.Spec.Containers[1].ImagePullPolicy).To(Equal(FlagDImagePullPolicy))
		Expect(pod.Spec.Containers[1].Env).To(Equal([]corev1.EnvVar{
			{Name: "LOG_LEVEL", Value: "dev"},
			{Name: flagdMetricPortEnvVar, Value: fmt.Sprintf("%d", flagdMetricsPort)},
		}))
		Expect(pod.Spec.Containers[1].Ports).To(Equal([]corev1.ContainerPort{
			{
				Name:          "metrics",
				Protocol:      "TCP",
				ContainerPort: 8014,
			},
		}))

		podMutationWebhookCleanup()
	})

	It("should not create flagd sidecar if openfeature.dev/featureflagconfiguration annotation isn't present", func() {
		pod := testPod()
		delete(pod.Annotations, "openfeature.dev/featureflagconfiguration")
		err := k8sClient.Create(testCtx, pod)
		Expect(err).ShouldNot(HaveOccurred())

		pod = getPod()

		Expect(len(pod.Spec.Containers)).To(Equal(1))

		podMutationWebhookCleanup()
	})

	It("should not create flagd sidecar if openfeature.dev annotation is disabled", func() {
		pod := testPod()
		pod.Annotations["openfeature.dev"] = "disabled"
		err := k8sClient.Create(testCtx, pod)
		Expect(err).ShouldNot(HaveOccurred())

		pod = getPod()

		Expect(len(pod.Spec.Containers)).To(Equal(1))

		podMutationWebhookCleanup()
	})

	It("should fail if pod has no owner references", func() {
		pod := testPod()
		pod.OwnerReferences = nil
		err := k8sClient.Create(testCtx, pod)
		Expect(err).Should(HaveOccurred())
	})

	It("should fail if service account not found", func() {
		pod := testPod()
		pod.Spec.ServiceAccountName = "foo"
		err := k8sClient.Create(testCtx, pod)
		Expect(err).Should(HaveOccurred())
	})

	It("should update cluster role binding's subjects", func() {
		pod := testPod()
		err := k8sClient.Create(testCtx, pod)
		Expect(err).ShouldNot(HaveOccurred())

		crb := &v1.ClusterRoleBinding{}
		err = k8sClient.Get(testCtx, client.ObjectKey{Name: clusterRoleBindingName}, crb)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(len(crb.Subjects)).Should(Equal(1))
		Expect(crb.Subjects[0].Kind).Should(Equal("ServiceAccount"))
		Expect(crb.Subjects[0].Namespace).Should(Equal(mutatePodNamespace))
		Expect(crb.Subjects[0].Name).Should(Equal(defaultPodServiceAccountName))

		podMutationWebhookCleanup()
	})

	It("should create config map if sync provider isn't kubernetes", func() {
		ffConfig := &corev1alpha1.FeatureFlagConfiguration{}
		err := k8sClient.Get(
			testCtx, client.ObjectKey{Name: featureFlagConfigurationName, Namespace: mutatePodNamespace}, ffConfig,
		)
		Expect(err).ShouldNot(HaveOccurred())

		ffConfig.Spec = corev1alpha1.FeatureFlagConfigurationSpec{
			SyncProvider: &corev1alpha1.FeatureFlagSyncProvider{
				Name: "not-kubernetes",
			},
		}
		err = k8sClient.Update(testCtx, ffConfig)
		Expect(err).ShouldNot(HaveOccurred())

		pod := testPod()
		err = k8sClient.Create(testCtx, pod)
		Expect(err).ShouldNot(HaveOccurred())

		cm := &corev1.ConfigMap{}
		err = k8sClient.Get(testCtx, client.ObjectKey{
			Name:      pod.Annotations["openfeature.dev/featureflagconfiguration"],
			Namespace: mutatePodNamespace,
		}, cm)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(cm.Name).To(Equal(pod.Annotations["openfeature.dev/featureflagconfiguration"]))
		Expect(cm.Namespace).To(Equal(mutatePodNamespace))
		Expect(cm.Annotations).To(Equal(map[string]string{
			"openfeature.dev/featureflagconfiguration": pod.Annotations["openfeature.dev/featureflagconfiguration"],
		}))
		Expect(len(cm.OwnerReferences)).To(Equal(2))
		Expect(cm.Data).To(Equal(map[string]string{
			"config.json": ffConfig.Spec.FeatureFlagSpec,
		}))

		podMutationWebhookCleanup()
		// reset FeatureFlagConfiguration
		ffConfig.Spec.SyncProvider = nil
		err = k8sClient.Update(testCtx, ffConfig)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("should not panic if flagDSpec isn't provided", func() {
		ffConfigName := "feature-flag-configuration-panic-test"
		ffConfig := &corev1alpha1.FeatureFlagConfiguration{}
		ffConfig.Namespace = mutatePodNamespace
		ffConfig.Name = ffConfigName
		ffConfig.Spec.FeatureFlagSpec = featureFlagSpec
		err := k8sClient.Create(testCtx, ffConfig)
		Expect(err).ShouldNot(HaveOccurred())

		pod := testPod()
		pod.Annotations["openfeature.dev/featureflagconfiguration"] = ffConfigName
		err = k8sClient.Create(testCtx, pod)
		Expect(err).ShouldNot(HaveOccurred())

		podMutationWebhookCleanup()
		err = k8sClient.Delete(testCtx, ffConfig, client.GracePeriodSeconds(0))
		Expect(err).ShouldNot(HaveOccurred())
	})
})
