package webhooks

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	mutatePodNamespace             = "test-mutate-pod"
	defaultPodName                 = "test-pod"
	defaultPodServiceAccountName   = "test-pod-service-account"
	featureFlagConfigurationName   = "test-feature-flag-configuration"
	featureFlagConfigurationName2  = "test-feature-flag-configuration-2"
	flagSourceConfigurationName    = "test-flag-source-configuration"
	flagSourceConfigurationName2   = "test-flag-source-configuration-2"
	flagSourceConfigurationName3   = "test-flag-source-configuration-3"
	existingPod1Name               = "existing-pod-1"
	existingPod1ServiceAccountName = "existing-pod-1-service-account"
	existingPod2Name               = "existing-pod-2"
	existingPod2ServiceAccountName = "existing-pod-2-service-account"
)

// Sets up environment to simulate an upgrade, with an existing pod already in the cluster
func setupPreviouslyExistingPods() {
	ns := &corev1.Namespace{}
	ns.Name = mutatePodNamespace
	err := k8sClient.Create(testCtx, ns)
	Expect(err).ShouldNot(HaveOccurred())

	svcAccount := &corev1.ServiceAccount{}
	svcAccount.Namespace = mutatePodNamespace
	svcAccount.Name = existingPod1ServiceAccountName
	err = k8sClient.Create(testCtx, svcAccount)
	Expect(err).ShouldNot(HaveOccurred())

	svcAccount = &corev1.ServiceAccount{}
	svcAccount.Namespace = mutatePodNamespace
	svcAccount.Name = existingPod2ServiceAccountName
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

	existingPod := testPod(existingPod1Name, existingPod1ServiceAccountName, map[string]string{
		fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, EnabledAnnotation):                  "true",
		fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, AllowKubernetesSyncAnnotation):      "true",
	})
	err = k8sClient.Create(testCtx, existingPod)
	Expect(err).ShouldNot(HaveOccurred())

	existingPod = testPod(existingPod2Name, existingPod2ServiceAccountName, map[string]string{
		OpenFeatureAnnotationPrefix: "enabled",
		fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, AllowKubernetesSyncAnnotation):      "true",
	})
	err = k8sClient.Create(testCtx, existingPod)
	Expect(err).ShouldNot(HaveOccurred())
}

func setupMutatePodResources() {
	svcAccount := &corev1.ServiceAccount{}
	svcAccount.Namespace = mutatePodNamespace
	svcAccount.Name = defaultPodServiceAccountName
	err := k8sClient.Create(testCtx, svcAccount)
	Expect(err).ShouldNot(HaveOccurred())

	ffConfig := &v1alpha1.FeatureFlagConfiguration{}
	ffConfig.Namespace = mutatePodNamespace
	ffConfig.Name = featureFlagConfigurationName
	ffConfig.Spec.FlagDSpec = &v1alpha1.FlagDSpec{Envs: []corev1.EnvVar{
		{Name: "LOG_LEVEL", Value: "dev"},
	}}
	ffConfig.Spec.FeatureFlagSpec = featureFlagSpec
	err = k8sClient.Create(testCtx, ffConfig)
	Expect(err).ShouldNot(HaveOccurred())

	fsConfig := &v1alpha1.FlagSourceConfiguration{}
	fsConfig.Namespace = mutatePodNamespace
	fsConfig.Name = flagSourceConfigurationName
	fsConfig.Spec.Port = 8080
	fsConfig.Spec.Evaluator = "yaml"
	fsConfig.Spec.Image = "new-image"
	fsConfig.Spec.Tag = "latest"
	fsConfig.Spec.MetricsPort = 8081
	fsConfig.Spec.SocketPath = "/tmp/flag-source.sock"
	fsConfig.Spec.SyncProviderArgs = []string{
		"key3=val3",
	}
	fsConfig.Spec.LogFormat = "console"
	fsConfig.Spec.Sources = []v1alpha1.Source{
		{
			Source:   "not-real.com",
			Provider: "http",
		},
	}
	err = k8sClient.Create(testCtx, fsConfig)
	Expect(err).ShouldNot(HaveOccurred())

	ffConfig2 := &v1alpha1.FeatureFlagConfiguration{}
	ffConfig2.Namespace = mutatePodNamespace
	ffConfig2.Name = featureFlagConfigurationName2
	ffConfig2.Spec.FeatureFlagSpec = featureFlagSpec
	err = k8sClient.Create(testCtx, ffConfig2)
	Expect(err).ShouldNot(HaveOccurred())

	fsConfig2 := &v1alpha1.FlagSourceConfiguration{}
	fsConfig2.Namespace = mutatePodNamespace
	fsConfig2.Name = flagSourceConfigurationName2
	fsConfig2.Spec.Sources = []v1alpha1.Source{
		{
			Source: fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		},
		{
			Source:   fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName2),
			Provider: v1alpha1.SyncProviderFilepath,
		},
	}
	err = k8sClient.Create(testCtx, fsConfig2)
	Expect(err).ShouldNot(HaveOccurred())

	fsConfig3 := &v1alpha1.FlagSourceConfiguration{}
	fsConfig3.Namespace = mutatePodNamespace
	fsConfig3.Name = flagSourceConfigurationName3
	fsConfig3.Spec.Sources = []v1alpha1.Source{
		{
			Source:   fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName2),
			Provider: v1alpha1.SyncProviderKubernetes,
		},
		{
			Source:   "i don't exist",
			Provider: v1alpha1.SyncProviderFilepath,
		},
	}
	err = k8sClient.Create(testCtx, fsConfig3)
	Expect(err).ShouldNot(HaveOccurred())
}

func testPod(podName string, serviceAccountName string, annotations map[string]string) *corev1.Pod {
	pod := &corev1.Pod{}
	pod.Namespace = mutatePodNamespace
	pod.Name = podName
	pod.Annotations = annotations

	pod.Spec.Containers = []corev1.Container{
		{
			Name:  "container1",
			Image: "ubuntu",
		},
	}
	pod.Spec.ServiceAccountName = serviceAccountName

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

func getPod(podName string) *corev1.Pod {
	pod := &corev1.Pod{}
	name := types.NamespacedName{
		Namespace: mutatePodNamespace,
		Name:      podName,
	}
	err := k8sClient.Get(testCtx, name, pod)
	ExpectWithOffset(1, err).ShouldNot(HaveOccurred())
	return pod
}

func getRoleBinding(roleBindingName string) *v1.ClusterRoleBinding {
	roleBinding := &v1.ClusterRoleBinding{}
	name := types.NamespacedName{
		Namespace: mutatePodNamespace,
		Name:      roleBindingName,
	}
	err := k8sClient.Get(testCtx, name, roleBinding)
	ExpectWithOffset(1, err).ShouldNot(HaveOccurred())
	return roleBinding
}

func podMutationWebhookCleanup() {
	pod := &corev1.Pod{}
	pod.Namespace = mutatePodNamespace
	pod.Name = defaultPodName
	err := k8sClient.Delete(testCtx, pod, client.GracePeriodSeconds(0))
	Expect(err).ShouldNot(HaveOccurred())
}

var _ = Describe("pod mutation webhook", func() {
	It("should backfill role binding subjects when annotated pods already exist in the cluster", func() {
		// this integration test confirms the proper execution of the  podMutator.BackfillPermissions method
		// this method is responsible for backfilling the subjects of the open-feature-operator-flagd-kubernetes-sync
		// cluster role binding, for previously existing pods on startup
		// a retry is required on this test as the backfilling occurs asynchronously
		var finalError error
		for i := 0; i < 3; i++ {
			pod1 := getPod(existingPod1Name)
			pod2 := getPod(existingPod2Name)
			// Pod 1 and 2 must not have been mutated by the webhook (we want the rolebinding to be updated via BackfillPermissions)

			if len(pod1.Spec.Containers) != 1 {
				finalError = errors.New("pod1 has had a container injected, it should not be mutated by the webhook")
				time.Sleep(1 * time.Second)
				continue
			}
			if len(pod2.Spec.Containers) != 1 {
				finalError = errors.New("pod2 has had a container injected, it should not be mutated by the webhook")
				time.Sleep(1 * time.Second)
				continue
			}

			rb := getRoleBinding(clusterRoleBindingName)

			unexpectedServiceAccount := ""
			for _, subject := range rb.Subjects {
				if !reflect.DeepEqual(subject, v1.Subject{
					Kind:      "ServiceAccount",
					APIGroup:  "",
					Name:      existingPod1ServiceAccountName,
					Namespace: mutatePodNamespace,
				}) &&
					!reflect.DeepEqual(subject, v1.Subject{
						Kind:      "ServiceAccount",
						APIGroup:  "",
						Name:      existingPod2ServiceAccountName,
						Namespace: mutatePodNamespace,
					}) {
					unexpectedServiceAccount = subject.Name
				}
			}
			if unexpectedServiceAccount != "" {
				finalError = fmt.Errorf("unexpected subject found in role binding, name: %s", unexpectedServiceAccount)
				time.Sleep(1 * time.Second)
				continue
			}
			finalError = nil
			break
		}
		Expect(finalError).ShouldNot(HaveOccurred())
	})

	It("should update cluster role binding's subjects", func() {
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "enabled",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		})
		err := k8sClient.Create(testCtx, pod)
		Expect(err).ShouldNot(HaveOccurred())

		crb := &v1.ClusterRoleBinding{}
		err = k8sClient.Get(testCtx, client.ObjectKey{Name: clusterRoleBindingName}, crb)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(crb.Subjects).To(ContainElement(v1.Subject{
			Kind:      "ServiceAccount",
			APIGroup:  "",
			Name:      defaultPodServiceAccountName,
			Namespace: mutatePodNamespace,
		}))

		podMutationWebhookCleanup()
	})

	It("should create flagd sidecar", func() {
		flagConfig, _ := v1alpha1.NewFlagSourceConfigurationSpec()
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "enabled",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		})
		err := k8sClient.Create(testCtx, pod)
		Expect(err).ShouldNot(HaveOccurred())

		pod = getPod(defaultPodName)
		Expect(pod.Annotations["openfeature.dev/allowkubernetessync"]).To(Equal("true"))
		Expect(len(pod.Spec.Containers)).To(Equal(2))
		Expect(pod.Spec.Containers[1].Name).To(Equal("flagd"))
		Expect(pod.Spec.Containers[1].Image).To(Equal(fmt.Sprintf("%s:%s", flagConfig.Image, flagConfig.Tag)))
		Expect(pod.Spec.Containers[1].Args).To(Equal([]string{
			"start", "--uri", fmt.Sprintf("core.openfeature.dev/%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		}))
		Expect(pod.Spec.Containers[1].ImagePullPolicy).To(Equal(FlagDImagePullPolicy))
		Expect(pod.Spec.Containers[1].Env).To(Equal([]corev1.EnvVar{
			{Name: "FLAGD_LOG_LEVEL", Value: "dev"},
		}))
		Expect(pod.Spec.Containers[1].Ports).To(Equal([]corev1.ContainerPort{
			{
				Name:          "metrics",
				Protocol:      "TCP",
				ContainerPort: 8014,
			},
		}))

		// Validate probes. Default config will set them
		liveness := pod.Spec.Containers[1].LivenessProbe
		Expect(liveness).ToNot(BeNil())
		Expect(liveness.HTTPGet.Path).To(Equal(ProbeLiveness))

		readiness := pod.Spec.Containers[1].ReadinessProbe
		Expect(readiness).ToNot(BeNil())
		Expect(readiness.HTTPGet.Path).To(Equal(ProbeReadiness))

		podMutationWebhookCleanup()
	})

	It("should create flagd sidecar even if openfeature.dev/featureflagconfiguration annotation isn't present", func() {
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "enabled",
		})
		err := k8sClient.Create(testCtx, pod)
		Expect(err).ShouldNot(HaveOccurred())

		pod = getPod(defaultPodName)

		Expect(len(pod.Spec.Containers)).To(Equal(2))

		podMutationWebhookCleanup()
	})

	It("should not create flagd sidecar if openfeature.dev annotation is disabled", func() {
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "disabled",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		})
		err := k8sClient.Create(testCtx, pod)
		Expect(err).ShouldNot(HaveOccurred())

		pod = getPod(defaultPodName)

		Expect(len(pod.Spec.Containers)).To(Equal(1))

		podMutationWebhookCleanup()
	})

	It("should fail if pod has no owner references", func() {
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "enabled",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		})
		pod.OwnerReferences = nil
		err := k8sClient.Create(testCtx, pod)
		Expect(err).Should(HaveOccurred())
	})

	It("should fail if service account not found", func() {
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "enabled",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		})
		pod.Spec.ServiceAccountName = "foo"
		err := k8sClient.Create(testCtx, pod)
		Expect(err).Should(HaveOccurred())
	})

	It("should create config map if sync provider is filepath", func() {
		ffConfig := &v1alpha1.FeatureFlagConfiguration{}
		err := k8sClient.Get(
			testCtx, client.ObjectKey{Name: featureFlagConfigurationName, Namespace: mutatePodNamespace}, ffConfig,
		)
		Expect(err).ShouldNot(HaveOccurred())

		ffConfig.Spec = v1alpha1.FeatureFlagConfigurationSpec{
			SyncProvider: &v1alpha1.FeatureFlagSyncProvider{
				Name: string(v1alpha1.SyncProviderFilepath),
			},
		}
		err = k8sClient.Update(testCtx, ffConfig)
		Expect(err).ShouldNot(HaveOccurred())

		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "enabled",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		})
		err = k8sClient.Create(testCtx, pod)
		Expect(err).ShouldNot(HaveOccurred())
		cm := &corev1.ConfigMap{}
		err = k8sClient.Get(testCtx, client.ObjectKey{
			Name:      featureFlagConfigurationName,
			Namespace: mutatePodNamespace,
		}, cm)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(cm.Name).To(Equal(featureFlagConfigurationName))
		Expect(cm.Namespace).To(Equal(mutatePodNamespace))
		Expect(cm.Annotations).To(Equal(map[string]string{
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): featureFlagConfigurationName,
		}))
		Expect(len(cm.OwnerReferences)).To(Equal(2))

		Expect(cm.Data).To(Equal(map[string]string{
			fmt.Sprintf("%s_%s.flagd.json", mutatePodNamespace, featureFlagConfigurationName): ffConfig.Spec.FeatureFlagSpec,
		}))

		podMutationWebhookCleanup()
		// reset FeatureFlagConfiguration
		ffConfig.Spec.SyncProvider = nil
		err = k8sClient.Update(testCtx, ffConfig)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("should not panic if flagDSpec isn't provided", func() {
		ffConfigName := "feature-flag-configuration-panic-test"
		ffConfig := &v1alpha1.FeatureFlagConfiguration{}
		ffConfig.Namespace = mutatePodNamespace
		ffConfig.Name = ffConfigName
		ffConfig.Spec.FeatureFlagSpec = featureFlagSpec
		err := k8sClient.Create(testCtx, ffConfig)
		Expect(err).ShouldNot(HaveOccurred())

		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "enabled",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, ffConfigName),
		})
		err = k8sClient.Create(testCtx, pod)
		Expect(err).ShouldNot(HaveOccurred())

		podMutationWebhookCleanup()
		err = k8sClient.Delete(testCtx, ffConfig, client.GracePeriodSeconds(0))
		Expect(err).ShouldNot(HaveOccurred())
	})

	It(`should create flagd sidecar if openfeature.dev/enabled annotation is "true"`, func() {
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, EnabledAnnotation):                  "true",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		})
		err := k8sClient.Create(testCtx, pod)
		Expect(err).ShouldNot(HaveOccurred())

		pod = getPod(defaultPodName)

		Expect(len(pod.Spec.Containers)).To(Equal(2))

		podMutationWebhookCleanup()
	})

	It(`should only write non default flagsourceconfiguration env vars to the flagd container`, func() {
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "enabled",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
			"openfeature.dev/flagsourceconfiguration":                                             fmt.Sprintf("%s/%s", mutatePodNamespace, flagSourceConfigurationName),
		})
		err := k8sClient.Create(testCtx, pod)
		Expect(err).ShouldNot(HaveOccurred())

		pod = getPod(defaultPodName)
		Expect(pod.Spec.Containers[1].Env).To(Equal([]corev1.EnvVar{
			{Name: "FLAGD_METRICS_PORT", Value: "8081"},
			{Name: "FLAGD_PORT", Value: "8080"},
			{Name: "FLAGD_EVALUATOR", Value: "yaml"},
			{Name: "FLAGD_SOCKET_PATH", Value: "/tmp/flag-source.sock"},
			{Name: "FLAGD_LOG_FORMAT", Value: "console"},
		}))

		podMutationWebhookCleanup()
	})

	It(`should use env var configuration to overwrite flagsourceconfiguration defaults`, func() {
		os.Setenv(v1alpha1.SidecarEnvVarPrefix, "MY_SIDECAR")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarMetricPortEnvVar), "10")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarPortEnvVar), "20")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarSocketPathEnvVar), "socket")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarEvaluatorEnvVar), "evaluator")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarImageEnvVar), "image")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarVersionEnvVar), "version")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarDefaultSyncProviderEnvVar), "filepath")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarProviderArgsEnvVar), "key=value,key2=value2")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarLogFormatEnvVar), "yaml")

		// Override probes - disabled
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarProbesEnabledVar), "false")

		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "enabled",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		})
		err := k8sClient.Create(testCtx, pod)
		Expect(err).ShouldNot(HaveOccurred())

		pod = getPod(defaultPodName)
		Expect(pod.Spec.Containers[1].Env).To(Equal([]corev1.EnvVar{
			{Name: "MY_SIDECAR_METRICS_PORT", Value: "10"},
			{Name: "MY_SIDECAR_PORT", Value: "20"},
			{Name: "MY_SIDECAR_EVALUATOR", Value: "evaluator"},
			{Name: "MY_SIDECAR_SOCKET_PATH", Value: "socket"},
			{Name: "MY_SIDECAR_LOG_FORMAT", Value: "yaml"},
		}))
		Expect(pod.Spec.Containers[1].Image).To(Equal("image:version"))
		Expect(pod.Spec.Containers[1].Args).To(Equal([]string{
			"start",
			"--uri",
			"file:/etc/flagd/test-mutate-pod_test-feature-flag-configuration/test-mutate-pod_test-feature-flag-configuration.flagd.json",
			"--sync-provider-args",
			"key=value",
			"--sync-provider-args",
			"key2=value2",
		}))

		// Validate probes - disabled
		Expect(pod.Spec.Containers[1].LivenessProbe).To(BeNil())
		Expect(pod.Spec.Containers[1].ReadinessProbe).To(BeNil())

		podMutationWebhookCleanup()
	})

	It(`should overwrite env var configuration with flagsourceconfiguration values`, func() {
		os.Setenv(v1alpha1.SidecarEnvVarPrefix, "")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarMetricPortEnvVar), "")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarPortEnvVar), "")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarSocketPathEnvVar), "")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarEvaluatorEnvVar), "")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarImageEnvVar), "")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarVersionEnvVar), "")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarDefaultSyncProviderEnvVar), "")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarProviderArgsEnvVar), "key=value,key2=value2")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarLogFormatEnvVar), "")

		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix:               "enabled",
			"openfeature.dev/flagsourceconfiguration": fmt.Sprintf("%s/%s", mutatePodNamespace, flagSourceConfigurationName),
		})
		err := k8sClient.Create(testCtx, pod)
		Expect(err).ShouldNot(HaveOccurred())

		pod = getPod(defaultPodName)
		Expect(pod.Spec.Containers[1].Env).To(Equal([]corev1.EnvVar{
			{Name: "FLAGD_METRICS_PORT", Value: "8081"},
			{Name: "FLAGD_PORT", Value: "8080"},
			{Name: "FLAGD_EVALUATOR", Value: "yaml"},
			{Name: "FLAGD_SOCKET_PATH", Value: "/tmp/flag-source.sock"},
			{Name: "FLAGD_LOG_FORMAT", Value: "console"},
		}))
		Expect(pod.Spec.Containers[1].Image).To(Equal("new-image:latest"))
		Expect(pod.Spec.Containers[1].Args).To(Equal([]string{
			"start",
			"--uri",
			"not-real.com",
			"--sync-provider-args",
			"key=value",
			"--sync-provider-args",
			"key2=value2",
			"--sync-provider-args",
			"key3=val3",
		}))
		podMutationWebhookCleanup()
	})

	It("should create flagd sidecar using flagsourceconfiguration", func() {
		os.Setenv(v1alpha1.SidecarEnvVarPrefix, "")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarMetricPortEnvVar), "")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarPortEnvVar), "")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarSocketPathEnvVar), "")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarEvaluatorEnvVar), "")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarImageEnvVar), "")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarVersionEnvVar), "")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarDefaultSyncProviderEnvVar), "")
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarProviderArgsEnvVar), "")
		flagConfig, _ := v1alpha1.NewFlagSourceConfigurationSpec()
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, EnabledAnnotation): "true",
			"openfeature.dev/flagsourceconfiguration":                            fmt.Sprintf("%s/%s", mutatePodNamespace, flagSourceConfigurationName2),
		})
		err := k8sClient.Create(testCtx, pod)
		Expect(err).ShouldNot(HaveOccurred())

		pod = getPod(defaultPodName)
		Expect(pod.Annotations["openfeature.dev/allowkubernetessync"]).To(Equal("true"))
		Expect(len(pod.Spec.Containers)).To(Equal(2))
		Expect(pod.Spec.Containers[1].Name).To(Equal("flagd"))
		Expect(pod.Spec.Containers[1].Image).To(Equal(fmt.Sprintf("%s:%s", flagConfig.Image, flagConfig.Tag)))
		Expect(pod.Spec.Containers[1].Args).To(Equal([]string{
			"start",
			"--uri",
			"core.openfeature.dev/test-mutate-pod/test-feature-flag-configuration",
			"--uri",
			"file:/etc/flagd/test-mutate-pod_test-feature-flag-configuration-2/test-mutate-pod_test-feature-flag-configuration-2.flagd.json",
		}))
		Expect(pod.Spec.Containers[1].ImagePullPolicy).To(Equal(FlagDImagePullPolicy))
		Expect(pod.Spec.Containers[1].Ports).To(Equal([]corev1.ContainerPort{
			{
				Name:          "metrics",
				Protocol:      "TCP",
				ContainerPort: 8014,
			},
		}))

		podMutationWebhookCleanup()
	})

	It("should not create flagd sidecar if flagsourceconfiguration does not exist", func() {
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix:               "enabled",
			"openfeature.dev/flagsourceconfiguration": "im-not-real",
		})
		err := k8sClient.Create(testCtx, pod)
		Expect(err).Should(HaveOccurred())
	})

	It("should not create flagd sidecar if flagsourceconfiguration  contains a source that does not exist", func() {
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix:               "enabled",
			"openfeature.dev/flagsourceconfiguration": fmt.Sprintf("%s/%s", mutatePodNamespace, flagSourceConfigurationName3),
		})
		err := k8sClient.Create(testCtx, pod)
		Expect(err).Should(HaveOccurred())
	})

	It(`should use defaultSyncProvider if one isn't provided`, func() {
		os.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarDefaultSyncProviderEnvVar), "filepath")

		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "enabled",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FlagSourceConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, flagSourceConfigurationName2),
		})
		err := k8sClient.Create(testCtx, pod)
		Expect(err).ShouldNot(HaveOccurred())

		pod = getPod(defaultPodName)
		Expect(pod.Spec.Containers[1].Args).To(Equal([]string{
			"start",
			"--uri",
			"file:/etc/flagd/test-mutate-pod_test-feature-flag-configuration/test-mutate-pod_test-feature-flag-configuration.flagd.json",
			"--uri",
			"file:/etc/flagd/test-mutate-pod_test-feature-flag-configuration-2/test-mutate-pod_test-feature-flag-configuration-2.flagd.json",
		}))
		podMutationWebhookCleanup()
	})
})
