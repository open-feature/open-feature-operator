package common

import (
	"context"
	"github.com/go-logr/logr/testr"
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	"github.com/open-feature/open-feature-operator/pkg/constant"
	"github.com/stretchr/testify/require"
	appsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestFlagdContainerInjector_InjectFlagd(t *testing.T) {

	namespace := "my-namespace"

	_ = v1alpha1.AddToScheme(scheme.Scheme)

	serviceAccount := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: namespace,
		},
	}

	cbr := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: constant.ClusterRoleBindingName,
		},
	}

	ffConfig := &v1alpha1.FeatureFlagConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "server-side",
			Namespace: namespace,
		},
		Spec: v1alpha1.FeatureFlagConfigurationSpec{},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).WithObjects(ffConfig, serviceAccount, cbr)

	client := fakeClient.Build()

	fi := &FlagdContainerInjector{
		Client:                    client,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagDResourceRequirements: getResourceRequirements(),
	}

	deployment := appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-deployment",
			Namespace: namespace,
		},
		Spec: appsV1.DeploymentSpec{},
	}

	debugLogging := true
	probesEnabled := true

	flagSourceConfig := &v1alpha1.FlagSourceConfigurationSpec{
		MetricsPort: 8014,
		Port:        8013,
		Image:       "flagd",
		Tag:         "0.5.0",
		Sources: []v1alpha1.Source{
			{
				Source:   "my-namespace/server-side",
				Provider: "kubernetes",
			},
		},
		EnvVars: []v1.EnvVar{
			{
				Name:  "my-env-var",
				Value: "my-value",
			},
		},
		EnvVarPrefix:  "flagd",
		ProbesEnabled: &probesEnabled,
		DebugLogging:  &debugLogging,
	}
	err := fi.InjectFlagd(context.Background(), &deployment.ObjectMeta, &deployment.Spec.Template.Spec, flagSourceConfig)

	require.Nil(t, err)

	expectedDeployment := getExpectedDeployment(namespace)

	require.Equal(t, expectedDeployment, deployment)
}

func getExpectedDeployment(namespace string) appsV1.Deployment {
	return appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-deployment",
			Namespace: namespace,
			Annotations: map[string]string{
				"openfeature.dev/allowkubernetessync": "true",
			},
		},
		Spec: appsV1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "flagd",
							Image: "flagd:0.5.0",
							Args: []string{
								"start",
								"--sources",
								"[{\"uri\":\"my-namespace/server-side\",\"provider\":\"kubernetes\"}]",
								"--debug",
							},
							WorkingDir: "",
							Ports: []v1.ContainerPort{
								{
									Name:          "metrics",
									ContainerPort: int32(8014),
								},
							},
							Env: []v1.EnvVar{
								{
									Name:  "flagd_my-env-var",
									Value: "my-value",
								},
								{
									Name:  "flagd_EVALUATOR",
									Value: "",
								},
								{
									Name:  "flagd_LOG_FORMAT",
									Value: "",
								},
							},
							Resources: getResourceRequirements(),
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/healthz",
										Port:   intstr.IntOrString{Type: 0, IntVal: 8014, StrVal: ""},
										Host:   "",
										Scheme: "HTTP",
									},
								},
								InitialDelaySeconds: 5,
								TimeoutSeconds:      0,
								PeriodSeconds:       0,
								SuccessThreshold:    0,
								FailureThreshold:    0,
							},
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/readyz",
										Port:   intstr.IntOrString{Type: 0, IntVal: 8014, StrVal: ""},
										Host:   "",
										Scheme: "HTTP",
									},
								},
								InitialDelaySeconds: 5,
								TimeoutSeconds:      0,
								PeriodSeconds:       0,
								SuccessThreshold:    0,
								FailureThreshold:    0,
							},
							VolumeMounts:             []v1.VolumeMount{},
							TerminationMessagePath:   "",
							TerminationMessagePolicy: "",
							ImagePullPolicy:          "Always",
							SecurityContext: &v1.SecurityContext{
								Capabilities: &v1.Capabilities{
									Drop: []v1.Capability{
										"all",
									},
								},
								Privileged:               boolPtr(false),
								RunAsUser:                intPtr(65532),
								RunAsGroup:               intPtr(65532),
								RunAsNonRoot:             boolPtr(true),
								ReadOnlyRootFilesystem:   boolPtr(true),
								AllowPrivilegeEscalation: boolPtr(false),
								SeccompProfile: &v1.SeccompProfile{
									Type: "RuntimeDefault",
								},
							},
						},
					},
				},
			},
		},
	}
}

func intPtr(i int64) *int64 {
	return &i
}

func getProxyConfig() *FlagdProxyConfiguration {
	return &FlagdProxyConfiguration{
		Port:         8013,
		MetricsPort:  8014,
		DebugLogging: false,
		Image:        "flagd",
		Tag:          "0.5.0",
		Namespace:    "my-namespace",
	}
}

func getResourceRequirements() v1.ResourceRequirements {
	cpuReq, _ := resource.ParseQuantity("0.2")
	cpuLimit, _ := resource.ParseQuantity("0.5")
	memReq, _ := resource.ParseQuantity("10M")
	memLimit, _ := resource.ParseQuantity("20M")
	return v1.ResourceRequirements{
		Limits: map[v1.ResourceName]resource.Quantity{
			v1.ResourceCPU:    cpuLimit,
			v1.ResourceMemory: memLimit,
		},
		Requests: map[v1.ResourceName]resource.Quantity{
			v1.ResourceCPU:    cpuReq,
			v1.ResourceMemory: memReq,
		},
	}
}

func boolPtr(b bool) *bool {
	return &b
}
