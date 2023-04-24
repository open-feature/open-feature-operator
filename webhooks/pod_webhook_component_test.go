package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	controllercommon "github.com/open-feature/open-feature-operator/controllers/common"
	"reflect"
	"testing"
	"time"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/go-logr/logr/testr"
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	corev1alpha2 "github.com/open-feature/open-feature-operator/apis/core/v1alpha2"
	corev1alpha3 "github.com/open-feature/open-feature-operator/apis/core/v1alpha3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	k8sClient client.Client
	testCtx   context.Context
	mutator   *PodMutator
)

const (
	mutatePodNamespace                = "test-mutate-pod"
	defaultPodName                    = "test-pod"
	defaultPodServiceAccountName      = "test-pod-service-account"
	featureFlagConfigurationName      = "test-feature-flag-configuration"
	featureFlagConfigurationName2     = "test-feature-flag-configuration-2"
	flagSourceConfigurationName       = "test-flag-source-configuration"
	flagSourceConfigurationName2      = "test-flag-source-configuration-2"
	flagSourceConfigurationName3      = "test-flag-source-configuration-3"
	flagSourceConfigGrpc              = "test-flag-source-grpc"
	existingPod1Name                  = "existing-pod-1"
	existingPod1ServiceAccountName    = "existing-pod-1-service-account"
	existingPod2Name                  = "existing-pod-2"
	existingPod2ServiceAccountName    = "existing-pod-2-service-account"
	featureFlagConfigurationNamespace = "test-validate-featureflagconfiguration"
	featureFlagSpec                   = `
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

func TestPodMutationWebhook_Component(t *testing.T) {
	setupTests(t)
	t.Run("should backfill role binding subjects when annotated pods already exist in the cluster", func(t *testing.T) {
		// this integration test confirms the proper execution of the  podMutator.BackfillPermissions method
		// this method is responsible for backfilling the subjects of the open-feature-operator-flagd-kubernetes-sync
		// cluster role binding, for previously existing pods on startup
		// a retry is required on this test as the backfilling occurs asynchronously
		var finalError error
		for i := 0; i < 3; i++ {
			pod1 := getPod(existingPod1Name, t)
			pod2 := getPod(existingPod2Name, t)

			handleMutation(t, pod1)
			handleMutation(t, pod2)

			rb := getRoleBinding(clusterRoleBindingName, t)

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
		require.Nil(t, finalError)
	})

	t.Run("should update cluster role binding's subjects", func(t *testing.T) {
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "enabled",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		})
		err := k8sClient.Create(testCtx, pod)
		require.Nil(t, err)
		defer podMutationWebhookCleanup(t)

		handleMutation(t, pod)

		crb := &v1.ClusterRoleBinding{}
		err = k8sClient.Get(testCtx, client.ObjectKey{Name: clusterRoleBindingName}, crb)
		require.Nil(t, err)

		require.Contains(t, crb.Subjects, v1.Subject{
			Kind:      "ServiceAccount",
			APIGroup:  "",
			Name:      defaultPodServiceAccountName,
			Namespace: mutatePodNamespace,
		})
	})

	t.Run("should create flagd sidecar", func(t *testing.T) {
		flagConfig, _ := v1alpha1.NewFlagSourceConfigurationSpec()
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "enabled",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		})
		err := k8sClient.Create(testCtx, pod)
		defer podMutationWebhookCleanup(t)
		require.Nil(t, err)
		pod = handleMutation(t, pod)
		require.Equal(t, "true", pod.Annotations["openfeature.dev/allowkubernetessync"])
		require.Equal(t, 2, len(pod.Spec.Containers))
		require.Equal(t, pod.Spec.Containers[1].Name, "flagd")
		require.Equal(t, pod.Spec.Containers[1].Image, fmt.Sprintf("%s:%s", flagConfig.Image, flagConfig.Tag))
		require.Equal(t, pod.Spec.Containers[1].Args, []string{
			"start", "--sources", "[{\"uri\":\"test-mutate-pod/test-feature-flag-configuration\",\"provider\":\"kubernetes\"}]",
		})
		require.Equal(t, pod.Spec.Containers[1].ImagePullPolicy, FlagDImagePullPolicy)
		require.Equal(t, pod.Spec.Containers[1].Env, []corev1.EnvVar{
			{Name: "FLAGD_LOG_LEVEL", Value: "dev"},
		})

		require.Equal(t, []corev1.ContainerPort{
			{
				Name:          "metrics",
				ContainerPort: 8014,
			},
		}, pod.Spec.Containers[1].Ports)

		// Validate probes. Default config will set them
		liveness := pod.Spec.Containers[1].LivenessProbe
		require.NotNil(t, liveness)
		require.Equal(t, ProbeLiveness, liveness.HTTPGet.Path)

		readiness := pod.Spec.Containers[1].ReadinessProbe
		require.NotNil(t, readiness)
		require.Equal(t, ProbeReadiness, readiness.HTTPGet.Path)

	})

	t.Run("should create flagd sidecar even if openfeature.dev/featureflagconfiguration annotation isn't present", func(t *testing.T) {
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "enabled",
		})

		err := k8sClient.Create(testCtx, pod)
		require.Nil(t, err)
		defer podMutationWebhookCleanup(t)
		pod = handleMutation(t, pod)

		require.Equal(t, 2, len(pod.Spec.Containers))
	})

	t.Run("should not create flagd sidecar if openfeature.dev annotation is disabled", func(t *testing.T) {
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "disabled",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		})
		err := k8sClient.Create(testCtx, pod)
		require.Nil(t, err)
		defer podMutationWebhookCleanup(t)

		pod = handleMutation(t, pod)

		require.Equal(t, len(pod.Spec.Containers), 1)

	})

	t.Run("should fail if pod has no owner references", func(t *testing.T) {
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "enabled",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		})
		pod.OwnerReferences = nil
		err := k8sClient.Create(testCtx, pod)
		require.Nil(t, err)
		defer podMutationWebhookCleanup(t)
		handleError(t, pod)
	})

	t.Run("should fail if service account not found", func(t *testing.T) {
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "enabled",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		})
		pod.Spec.ServiceAccountName = "foo"
		err := k8sClient.Create(testCtx, pod)
		require.Nil(t, err)
		defer podMutationWebhookCleanup(t)
		handleError(t, pod)
	})

	t.Run("should create config map if sync provider is filepath", func(t *testing.T) {
		ffConfig := &v1alpha1.FeatureFlagConfiguration{}
		err := k8sClient.Get(
			testCtx, client.ObjectKey{Name: featureFlagConfigurationName, Namespace: mutatePodNamespace}, ffConfig,
		)
		require.Nil(t, err)

		ffConfig.Spec = v1alpha1.FeatureFlagConfigurationSpec{
			SyncProvider: &v1alpha1.FeatureFlagSyncProvider{
				Name: string(v1alpha1.SyncProviderFilepath),
			},
		}
		err = k8sClient.Update(testCtx, ffConfig)
		require.Nil(t, err)

		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "enabled",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		})
		err = k8sClient.Create(testCtx, pod)
		require.Nil(t, err)

		defer func() {
			podMutationWebhookCleanup(t)
			// reset FeatureFlagConfiguration
			ffConfig.Spec.SyncProvider = nil
			err = k8sClient.Update(testCtx, ffConfig)
			require.Nil(t, err)
		}()

		handleMutation(t, pod)

		cm := &corev1.ConfigMap{}
		err = k8sClient.Get(testCtx, client.ObjectKey{
			Name:      featureFlagConfigurationName,
			Namespace: mutatePodNamespace,
		}, cm)
		require.Nil(t, err)

		require.Equal(t, cm.Name, featureFlagConfigurationName)
		require.Equal(t, cm.Namespace, mutatePodNamespace)
		require.EqualValues(t, map[string]string{
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): featureFlagConfigurationName,
		}, cm.Annotations)
		require.Equal(t, 2, len(cm.OwnerReferences))

		require.Equal(t, cm.Data, map[string]string{
			fmt.Sprintf("%s_%s.flagd.json", mutatePodNamespace, featureFlagConfigurationName): ffConfig.Spec.FeatureFlagSpec,
		})

	})

	t.Run("should not panic if flagDSpec isn't provided", func(t *testing.T) {
		ffConfigName := "feature-flag-configuration-panic-test"
		ffConfig := &v1alpha1.FeatureFlagConfiguration{}
		ffConfig.Namespace = mutatePodNamespace
		ffConfig.Name = ffConfigName
		ffConfig.Spec.FeatureFlagSpec = featureFlagSpec
		err := k8sClient.Create(testCtx, ffConfig)
		require.Nil(t, err)

		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "enabled",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, ffConfigName),
		})
		err = k8sClient.Create(testCtx, pod)
		require.Nil(t, err)

		podMutationWebhookCleanup(t)
		err = k8sClient.Delete(testCtx, ffConfig, client.GracePeriodSeconds(0))
		require.Nil(t, err)
	})

	t.Run(`should create flagd sidecar if openfeature.dev/enabled annotation is "true"`, func(t *testing.T) {
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, EnabledAnnotation):                  "true",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		})
		err := k8sClient.Create(testCtx, pod)
		require.Nil(t, err)
		defer podMutationWebhookCleanup(t)

		pod = handleMutation(t, pod)

		require.Equal(t, 2, len(pod.Spec.Containers))

	})

	t.Run(`should only write non default flagsourceconfiguration env vars to the flagd container`, func(t *testing.T) {
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "enabled",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
			"openfeature.dev/flagsourceconfiguration":                                             fmt.Sprintf("%s/%s", mutatePodNamespace, flagSourceConfigurationName),
		})
		err := k8sClient.Create(testCtx, pod)
		require.Nil(t, err)
		defer podMutationWebhookCleanup(t)

		pod = handleMutation(t, pod)
		require.Equal(t, []corev1.EnvVar{
			{Name: "FLAGD_METRICS_PORT", Value: "8081"},
			{Name: "FLAGD_PORT", Value: "8080"},
			{Name: "FLAGD_EVALUATOR", Value: "yaml"},
			{Name: "FLAGD_SOCKET_PATH", Value: "/tmp/flag-source.sock"},
			{Name: "FLAGD_LOG_FORMAT", Value: "console"},
		},
			pod.Spec.Containers[1].Env)

	})

	t.Run(`should use env var configuration to overwrite flagsourceconfiguration defaults`, func(t *testing.T) {
		t.Setenv(v1alpha1.SidecarEnvVarPrefix, "MY_SIDECAR")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarMetricPortEnvVar), "10")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarPortEnvVar), "20")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarSocketPathEnvVar), "socket")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarEvaluatorEnvVar), "evaluator")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarImageEnvVar), "image")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarVersionEnvVar), "version")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarDefaultSyncProviderEnvVar), "filepath")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarProviderArgsEnvVar), "key=value,key2=value2")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarLogFormatEnvVar), "yaml")

		// Override probes - disabled
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarProbesEnabledVar), "false")

		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "enabled",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		})
		err := k8sClient.Create(testCtx, pod)
		require.Nil(t, err)
		defer podMutationWebhookCleanup(t)

		pod = handleMutation(t, pod)
		require.Equal(t, pod.Spec.Containers[1].Env, []corev1.EnvVar{
			{Name: "MY_SIDECAR_METRICS_PORT", Value: "10"},
			{Name: "MY_SIDECAR_PORT", Value: "20"},
			{Name: "MY_SIDECAR_EVALUATOR", Value: "evaluator"},
			{Name: "MY_SIDECAR_SOCKET_PATH", Value: "socket"},
			{Name: "MY_SIDECAR_LOG_FORMAT", Value: "yaml"},
		})
		require.Equal(t, pod.Spec.Containers[1].Image, "image:version")
		require.Equal(t, pod.Spec.Containers[1].Args, []string{
			"start",
			SourceConfigParam,
			"[{\"uri\":\"/etc/flagd/test-mutate-pod_test-feature-flag-configuration/test-mutate-pod_test-feature-flag-configuration.flagd.json\",\"provider\":\"file\"}]",
			"--sync-provider-args",
			"key=value",
			"--sync-provider-args",
			"key2=value2",
		})

		// Validate probes - disabled
		require.Nil(t, pod.Spec.Containers[1].LivenessProbe)
		require.Nil(t, pod.Spec.Containers[1].ReadinessProbe)

	})

	t.Run(`should overwrite env var configuration with flagsourceconfiguration values`, func(t *testing.T) {
		t.Setenv(v1alpha1.SidecarEnvVarPrefix, "")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarMetricPortEnvVar), "")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarPortEnvVar), "")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarSocketPathEnvVar), "")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarEvaluatorEnvVar), "")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarImageEnvVar), "")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarVersionEnvVar), "")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarDefaultSyncProviderEnvVar), "")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarProviderArgsEnvVar), "key=value,key2=value2")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarLogFormatEnvVar), "")

		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix:               "enabled",
			"openfeature.dev/flagsourceconfiguration": fmt.Sprintf("%s/%s", mutatePodNamespace, flagSourceConfigurationName),
		})
		err := k8sClient.Create(testCtx, pod)
		require.Nil(t, err)
		defer podMutationWebhookCleanup(t)

		pod = handleMutation(t, pod)
		require.Equal(t, pod.Spec.Containers[1].Env, []corev1.EnvVar{
			{Name: "FLAGD_METRICS_PORT", Value: "8081"},
			{Name: "FLAGD_PORT", Value: "8080"},
			{Name: "FLAGD_EVALUATOR", Value: "yaml"},
			{Name: "FLAGD_SOCKET_PATH", Value: "/tmp/flag-source.sock"},
			{Name: "FLAGD_LOG_FORMAT", Value: "console"},
		})
		require.Equal(t, pod.Spec.Containers[1].Image, "new-image:latest")
		require.Equal(t, pod.Spec.Containers[1].Args, []string{
			"start",
			SourceConfigParam,
			"[{\"uri\":\"not-real.com\",\"provider\":\"http\"}]",
			"--sync-provider-args",
			"key=value",
			"--sync-provider-args",
			"key2=value2",
			"--sync-provider-args",
			"key3=val3",
		})
	})

	t.Run("should create flagd sidecar using flagsourceconfiguration", func(t *testing.T) {
		t.Setenv(v1alpha1.SidecarEnvVarPrefix, "")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarMetricPortEnvVar), "")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarPortEnvVar), "")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarSocketPathEnvVar), "")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarEvaluatorEnvVar), "")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarImageEnvVar), "")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarVersionEnvVar), "")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarDefaultSyncProviderEnvVar), "")
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarProviderArgsEnvVar), "")
		flagConfig, _ := v1alpha1.NewFlagSourceConfigurationSpec()
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, EnabledAnnotation): "true",
			"openfeature.dev/flagsourceconfiguration":                            fmt.Sprintf("%s/%s", mutatePodNamespace, flagSourceConfigurationName2),
		})
		err := k8sClient.Create(testCtx, pod)
		require.Nil(t, err)
		defer podMutationWebhookCleanup(t)

		pod = handleMutation(t, pod)
		require.Equal(t, pod.Annotations["openfeature.dev/allowkubernetessync"], "true")
		require.Equal(t, len(pod.Spec.Containers), 2)
		require.Equal(t, pod.Spec.Containers[1].Name, "flagd")
		require.Equal(t, pod.Spec.Containers[1].Image, fmt.Sprintf("%s:%s", flagConfig.Image, flagConfig.Tag))
		require.Equal(t, pod.Spec.Containers[1].Args, []string{
			"start",
			SourceConfigParam,
			"[{\"uri\":\"test-mutate-pod/test-feature-flag-configuration\",\"provider\":\"kubernetes\"}," +
				"{\"uri\":\"/etc/flagd/test-mutate-pod_test-feature-flag-configuration-2/test-mutate-pod_test-feature-flag-configuration-2.flagd.json\",\"provider\":\"file\"}]",
		})
		require.Equal(t, pod.Spec.Containers[1].ImagePullPolicy, FlagDImagePullPolicy)
		require.Equal(t, pod.Spec.Containers[1].Ports, []corev1.ContainerPort{
			{
				Name:          "metrics",
				ContainerPort: 8014,
			},
		})

	})

	t.Run("should not create flagd sidecar if flagsourceconfiguration does not exist", func(t *testing.T) {
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix:               "enabled",
			"openfeature.dev/flagsourceconfiguration": "im-not-real",
		})
		err := k8sClient.Create(testCtx, pod)
		require.Nil(t, err)
		defer podMutationWebhookCleanup(t)
		handleError(t, pod)
	})

	t.Run("should not create flagd sidecar if flagsourceconfiguration  contains a source that does not exist", func(t *testing.T) {
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix:               "enabled",
			"openfeature.dev/flagsourceconfiguration": fmt.Sprintf("%s/%s", mutatePodNamespace, flagSourceConfigurationName3),
		})
		err := k8sClient.Create(testCtx, pod)
		require.Nil(t, err)
		defer podMutationWebhookCleanup(t)
		handleError(t, pod)
	})

	t.Run(`should use defaultSyncProvider if one isn't provided`, func(t *testing.T) {
		t.Setenv(fmt.Sprintf("%s_%s", v1alpha1.InputConfigurationEnvVarPrefix, v1alpha1.SidecarDefaultSyncProviderEnvVar), "filepath")

		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "enabled",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FlagSourceConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, flagSourceConfigurationName2),
		})
		err := k8sClient.Create(testCtx, pod)
		require.Nil(t, err)
		defer podMutationWebhookCleanup(t)

		pod = handleMutation(t, pod)
		require.Equal(t, pod.Spec.Containers[1].Args, []string{
			"start",
			SourceConfigParam,
			"[{\"uri\":\"/etc/flagd/test-mutate-pod_test-feature-flag-configuration/test-mutate-pod_test-feature-flag-configuration.flagd.json\",\"provider\":\"file\"}," +
				"{\"uri\":\"/etc/flagd/test-mutate-pod_test-feature-flag-configuration-2/test-mutate-pod_test-feature-flag-configuration-2.flagd.json\",\"provider\":\"file\"}]",
		})

	})

	t.Run("should create a valid grpc source configuration", func(t *testing.T) {
		pod := testPod(defaultPodName, defaultPodServiceAccountName, map[string]string{
			OpenFeatureAnnotationPrefix: "enabled",
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FlagSourceConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, flagSourceConfigGrpc),
		})
		err := k8sClient.Create(testCtx, pod)
		require.Nil(t, err)
		defer podMutationWebhookCleanup(t)

		pod = handleMutation(t, pod)
		require.Equal(t, pod.Spec.Containers[1].Args, []string{
			"start",
			SourceConfigParam,
			"[{\"uri\":\"grpc-service:9090\",\"provider\":\"grpc\",\"certPath\":\"/tmp/certs\",\"tls\":true,\"providerID\":\"myapp\",\"selector\":\"source=database\"}]",
		})
	})
}

func handleError(t *testing.T, pod *corev1.Pod) {
	_, res := triggerHandler(t, pod)
	require.False(t, res.Allowed)
}

// calls handle of the webhook and returns the mutated pod according to the resulting patch
func handleMutation(t *testing.T, pod *corev1.Pod) *corev1.Pod {

	rawPod, res := triggerHandler(t, pod)

	data, err := json.Marshal(res.Patches)
	assert.Nil(t, err)

	patch, err := jsonpatch.DecodePatch(data)
	assert.Nil(t, err)

	patchedPod, err := patch.Apply(rawPod)
	assert.Nil(t, err)

	newPod := &corev1.Pod{}
	err = json.Unmarshal(patchedPod, newPod)
	assert.Nil(t, err)

	return newPod
}

func triggerHandler(t *testing.T, pod *corev1.Pod) ([]byte, admission.Response) {
	rawPod, err := json.Marshal(pod)
	require.Nil(t, err)
	req := admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			UID: pod.UID,
			Object: runtime.RawExtension{
				Raw:    rawPod,
				Object: pod,
			},
		},
	}

	res := mutator.Handle(context.TODO(), req)
	return rawPod, res
}

func setupTests(t *testing.T) {

	utilruntime.Must(clientgoscheme.AddToScheme(scheme.Scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme.Scheme))
	utilruntime.Must(corev1alpha2.AddToScheme(scheme.Scheme))
	utilruntime.Must(corev1alpha3.AddToScheme(scheme.Scheme))

	annotationsSyncIndexer := func(obj client.Object) []string {
		res := obj.GetAnnotations()[fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, AllowKubernetesSyncAnnotation)]
		return []string{res}
	}

	featureflagIndexer := func(obj client.Object) []string {
		res := obj.GetAnnotations()["openfeature.dev/featureflagconfiguration"]
		return []string{res}
	}
	enabledIndexer := func(obj client.Object) []string {
		res := obj.GetAnnotations()["openfeature.dev/enabled"]
		return []string{res}
	}

	k8sClient = fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithIndex(
			&corev1.Pod{},
			"metadata.annotations.openfeature.dev/allowkubernetessync",
			annotationsSyncIndexer).
		WithIndex(
			&corev1.Pod{},
			"metadata.annotations.openfeature.dev/featureflagconfiguration",
			featureflagIndexer).
		WithIndex(
			&corev1.Pod{},
			"metadata.annotations.openfeature.dev/enabled",
			enabledIndexer).
		Build()

	decoder, err := admission.NewDecoder(scheme.Scheme)
	require.Nil(t, err)

	setupValidateFeatureFlagConfigurationResources(t)
	setupPreviouslyExistingPods(t)
	setupMutatePodResources(t)

	mutator = &PodMutator{
		Client:  k8sClient,
		decoder: decoder,
		Log:     testr.New(t),
		FlagdInjector: &controllercommon.FlagdContainerInjector{
			Client:                    k8sClient,
			Logger:                    testr.New(t),
			FlagDResourceRequirements: corev1.ResourceRequirements{},
		},
		ready: false,
	}

}

func setupValidateFeatureFlagConfigurationResources(t *testing.T) {
	ns := &corev1.Namespace{}
	ns.Name = featureFlagConfigurationNamespace
	err := k8sClient.Create(testCtx, ns)
	require.Nil(t, err)
}

// // Sets up environment to simulate an upgrade, with an existing pod already in the cluster
func setupPreviouslyExistingPods(t *testing.T) {
	ns := &corev1.Namespace{}
	ns.Name = mutatePodNamespace
	err := k8sClient.Create(testCtx, ns)
	require.Nil(t, err)

	svcAccount := &corev1.ServiceAccount{}
	svcAccount.Namespace = mutatePodNamespace
	svcAccount.Name = existingPod1ServiceAccountName
	err = k8sClient.Create(testCtx, svcAccount)
	require.Nil(t, err)

	svcAccount = &corev1.ServiceAccount{}
	svcAccount.Namespace = mutatePodNamespace
	svcAccount.Name = existingPod2ServiceAccountName
	err = k8sClient.Create(testCtx, svcAccount)
	require.Nil(t, err)

	clusterRoleBinding := &v1.ClusterRoleBinding{}
	clusterRoleBinding.Name = clusterRoleBindingName
	clusterRoleBinding.APIVersion = "rbac.authorization.k8s.io/v1"
	clusterRoleBinding.RoleRef = v1.RoleRef{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     clusterRoleBindingName,
	}
	err = k8sClient.Create(testCtx, clusterRoleBinding)
	require.Nil(t, err)

	existingPod := testPod(existingPod1Name, existingPod1ServiceAccountName, map[string]string{
		fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, EnabledAnnotation):                  "true",
		fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, AllowKubernetesSyncAnnotation):      "true",
	})
	err = k8sClient.Create(testCtx, existingPod)
	require.Nil(t, err)

	existingPod = testPod(existingPod2Name, existingPod2ServiceAccountName, map[string]string{
		OpenFeatureAnnotationPrefix: "enabled",
		fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
		fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, AllowKubernetesSyncAnnotation):      "true",
	})
	err = k8sClient.Create(testCtx, existingPod)
	require.Nil(t, err)
}

func setupMutatePodResources(t *testing.T) {
	svcAccount := &corev1.ServiceAccount{}
	svcAccount.Namespace = mutatePodNamespace
	svcAccount.Name = defaultPodServiceAccountName
	err := k8sClient.Create(testCtx, svcAccount)
	require.Nil(t, err)

	ffConfig := &v1alpha1.FeatureFlagConfiguration{}
	ffConfig.Namespace = mutatePodNamespace
	ffConfig.Name = featureFlagConfigurationName
	ffConfig.Spec.FlagDSpec = &v1alpha1.FlagDSpec{Envs: []corev1.EnvVar{
		{Name: "LOG_LEVEL", Value: "dev"},
	}}
	ffConfig.Spec.FeatureFlagSpec = featureFlagSpec
	err = k8sClient.Create(testCtx, ffConfig)
	require.Nil(t, err)

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
	require.Nil(t, err)

	ffConfig2 := &v1alpha1.FeatureFlagConfiguration{}
	ffConfig2.Namespace = mutatePodNamespace
	ffConfig2.Name = featureFlagConfigurationName2
	ffConfig2.Spec.FeatureFlagSpec = featureFlagSpec
	err = k8sClient.Create(testCtx, ffConfig2)
	require.Nil(t, err)

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
	require.Nil(t, err)

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
	require.Nil(t, err)

	fsConfigGrpc := &v1alpha1.FlagSourceConfiguration{}
	fsConfigGrpc.Namespace = mutatePodNamespace
	fsConfigGrpc.Name = flagSourceConfigGrpc
	fsConfigGrpc.Spec.Sources = []v1alpha1.Source{
		{
			Source:     "grpc-service:9090",
			Provider:   v1alpha1.SyncProviderGrpc,
			TLS:        true,
			ProviderID: "myapp",
			Selector:   "source=database",
			CertPath:   "/tmp/certs",
		},
	}
	err = k8sClient.Create(testCtx, fsConfigGrpc)
	require.Nil(t, err)
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

func getPod(podName string, t *testing.T) *corev1.Pod {
	pod := &corev1.Pod{}
	name := types.NamespacedName{
		Namespace: mutatePodNamespace,
		Name:      podName,
	}
	err := k8sClient.Get(testCtx, name, pod)
	require.Nil(t, err)
	return pod
}

func getRoleBinding(roleBindingName string, t *testing.T) *v1.ClusterRoleBinding {
	roleBinding := &v1.ClusterRoleBinding{}
	name := types.NamespacedName{
		Name: roleBindingName,
	}
	err := k8sClient.Get(testCtx, name, roleBinding)
	require.Nil(t, err)
	return roleBinding
}

func podMutationWebhookCleanup(t *testing.T) {
	pod := &corev1.Pod{}
	pod.Namespace = mutatePodNamespace
	pod.Name = defaultPodName
	err := k8sClient.Delete(testCtx, pod, client.GracePeriodSeconds(0))
	require.Nil(t, err)
	require.Eventually(t, func() bool {
		err = k8sClient.Get(testCtx, types.NamespacedName{
			Name: defaultPodName, Namespace: mutatePodNamespace,
		}, pod)
		return errors2.IsNotFound(err)
	}, time.Second, 10*time.Millisecond)
}
