package flagdproxy

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-logr/logr/testr"
	"github.com/open-feature/open-feature-operator/common"
	"github.com/open-feature/open-feature-operator/common/types"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	pullSecrets = []string{"test-pullSecret"}

	labels = map[string]string{
		"label1": "labelValue1",
		"label2": "labelValue2",
	}

	annotations = map[string]string{
		"annotation1": "annotationValue1",
		"annotation2": "annotationValue2",
	}

	testPort           = 88
	testManagementPort = 90
	testReplicaCount   = 1
	testDebugLogging   = true
	testNamespace      = "ns"
	testImage          = "image"
	testTag            = "tag"

	testEnvConfig = types.EnvConfig{
		PodNamespace:             testNamespace,
		FlagdProxyImage:          testImage,
		FlagdProxyTag:            testTag,
		FlagdProxyPort:           testPort,
		FlagdProxyManagementPort: testManagementPort,
		FlagdProxyReplicaCount:   testReplicaCount,
		FlagdProxyDebugLogging:   testDebugLogging,
	}

	expectedDeploymentReplicas = int32(testReplicaCount)
	expectedDeployment         = &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      FlagdProxyDeploymentName,
			Namespace: "ns",
			Labels: map[string]string{
				"app":                          FlagdProxyDeploymentName,
				"app.kubernetes.io/managed-by": common.ManagedByAnnotationValue,
				"app.kubernetes.io/version":    testTag,
			},
			ResourceVersion: "1",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       common.OperatorDeploymentName,
				},
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &expectedDeploymentReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": FlagdProxyDeploymentName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                          FlagdProxyDeploymentName,
						"app.kubernetes.io/name":       FlagdProxyDeploymentName,
						"app.kubernetes.io/managed-by": common.ManagedByAnnotationValue,
						"app.kubernetes.io/version":    "tag",
						"label1":                       "labelValue1",
						"label2":                       "labelValue2",
					},
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: FlagdProxyServiceAccountName,
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: pullSecrets[0]},
					},
					Containers: []corev1.Container{
						{
							Image: fmt.Sprintf("%s:%s", testImage, testTag),
							Name:  FlagdProxyDeploymentName,
							Ports: []corev1.ContainerPort{
								{
									Name:          "port",
									ContainerPort: int32(testPort),
								},
								{
									Name:          "management-port",
									ContainerPort: int32(testManagementPort),
								},
							},
							Args: []string{"start", "--management-port", fmt.Sprint(testManagementPort), "--debug"},
						},
					},
					TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
						{
							MaxSkew:           1,
							TopologyKey:       "kubernetes.io/hostname",
							WhenUnsatisfiable: corev1.DoNotSchedule,
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app.kubernetes.io/name":      FlagdProxyDeploymentName,
									common.ManagedByAnnotationKey: common.ManagedByAnnotationValue,
								},
							},
						},
					},
				},
			},
		},
	}
	expectedService = &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            FlagdProxyServiceName,
			Namespace:       testNamespace,
			ResourceVersion: "1",
			Labels: map[string]string{
				common.ManagedByAnnotationKey: common.ManagedByAnnotationValue,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       common.OperatorDeploymentName,
				},
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app.kubernetes.io/name":       FlagdProxyDeploymentName,
				"app.kubernetes.io/managed-by": common.ManagedByAnnotationValue,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "flagd-proxy",
					Port:       int32(testPort),
					TargetPort: intstr.FromInt(testPort),
				},
			},
		},
	}

	expectedPDBminAvailable = intstr.FromInt(0)
	expectedPDB             = &policyv1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            FlagdProxyPodDisruptionBudgetName,
			Namespace:       "ns",
			ResourceVersion: "1",
			Labels: map[string]string{
				common.ManagedByAnnotationKey: common.ManagedByAnnotationValue,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       common.OperatorDeploymentName,
				},
			},
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &expectedPDBminAvailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name":      FlagdProxyDeploymentName,
					common.ManagedByAnnotationKey: common.ManagedByAnnotationValue,
				},
			},
		},
	}
)

func TestNewFlagdProxyConfiguration(t *testing.T) {

	kpConfig := NewFlagdProxyConfiguration(types.EnvConfig{
		FlagdProxyPort:           8015,
		FlagdProxyManagementPort: 8016,
		FlagdProxyReplicaCount:   123,
	}, pullSecrets, labels, annotations)

	require.NotNil(t, kpConfig)
	require.Equal(t, &FlagdProxyConfiguration{
		Port:                   8015,
		ManagementPort:         8016,
		DebugLogging:           false,
		OperatorDeploymentName: common.OperatorDeploymentName,
		ImagePullSecrets:       pullSecrets,
		Replicas:               123,
		Labels:                 labels,
		Annotations:            annotations,
	}, kpConfig)
}

func TestNewFlagdProxyConfiguration_OverrideEnvVars(t *testing.T) {
	env := types.EnvConfig{
		FlagdProxyImage:          "my-image",
		FlagdProxyTag:            "my-tag",
		PodNamespace:             "my-namespace",
		FlagdProxyPort:           8080,
		FlagdProxyManagementPort: 8081,
		FlagdProxyDebugLogging:   true,
	}

	kpConfig := NewFlagdProxyConfiguration(env, pullSecrets, labels, annotations)

	require.NotNil(t, kpConfig)
	require.Equal(t, &FlagdProxyConfiguration{
		Port:                   8080,
		ManagementPort:         8081,
		DebugLogging:           true,
		Image:                  "my-image",
		Tag:                    "my-tag",
		Namespace:              "my-namespace",
		OperatorDeploymentName: common.OperatorDeploymentName,
		ImagePullSecrets:       pullSecrets,
		Labels:                 labels,
		Annotations:            annotations,
	}, kpConfig)
}

func TestNewFlagdProxyHandler(t *testing.T) {
	kpConfig := NewFlagdProxyConfiguration(types.EnvConfig{}, pullSecrets, labels, annotations)

	require.NotNil(t, kpConfig)

	fakeClient := fake.NewClientBuilder().Build()

	ph := NewFlagdProxyHandler(kpConfig, fakeClient, testr.New(t))

	require.NotNil(t, ph)

	require.Equal(t, kpConfig, ph.Config())
}

func TestFlagdProxyHandler_HandleFlagdProxy_ProxyExistsWithBadVersion(t *testing.T) {
	env := types.EnvConfig{
		PodNamespace: "ns",
	}
	kpConfig := NewFlagdProxyConfiguration(env, pullSecrets, labels, annotations)

	require.NotNil(t, kpConfig)

	fakeClient := fake.NewClientBuilder().WithObjects(createOFOTestDeployment(env.PodNamespace)).Build()

	ownerRef, err := getTestOFODeploymentOwnerRef(fakeClient, env.PodNamespace)
	require.Nil(t, err)

	proxyDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       kpConfig.Namespace,
			Name:            FlagdProxyDeploymentName,
			OwnerReferences: []metav1.OwnerReference{*ownerRef},
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": common.ManagedByAnnotationValue,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "my-container",
						},
					},
				},
			},
		},
	}

	err = fakeClient.Create(context.TODO(), proxyDeployment)
	require.Nil(t, err)

	ph := NewFlagdProxyHandler(kpConfig, fakeClient, testr.New(t))

	require.NotNil(t, ph)

	err = ph.HandleFlagdProxy(context.Background())

	require.Nil(t, err)

	deployment := &appsv1.Deployment{}
	err = fakeClient.Get(context.Background(), client.ObjectKey{
		Namespace: env.PodNamespace,
		Name:      FlagdProxyDeploymentName,
	}, deployment)

	require.Nil(t, err)
	require.NotNil(t, deployment)

	// verify that the existing deployment has been changed
	require.Equal(t, "flagd-proxy", deployment.Spec.Template.Spec.Containers[0].Name)
}

func TestFlagdProxyHandler_HandleFlagdProxy_ProxyExistsWithoutLabel(t *testing.T) {
	env := types.EnvConfig{
		PodNamespace: "ns",
	}
	kpConfig := NewFlagdProxyConfiguration(env, pullSecrets, labels, annotations)

	require.NotNil(t, kpConfig)

	proxyDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: kpConfig.Namespace,
			Name:      FlagdProxyDeploymentName,
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "my-container",
						},
					},
				},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().WithObjects(createOFOTestDeployment(env.PodNamespace), proxyDeployment).Build()

	ph := NewFlagdProxyHandler(kpConfig, fakeClient, testr.New(t))

	require.NotNil(t, ph)

	err := ph.HandleFlagdProxy(context.Background())

	require.ErrorContains(t, err, "not managed by OFO")

	deployment := &appsv1.Deployment{}
	err = fakeClient.Get(context.Background(), client.ObjectKey{
		Namespace: env.PodNamespace,
		Name:      FlagdProxyDeploymentName,
	}, deployment)

	require.Nil(t, err)
	require.NotNil(t, deployment)

	// verify that the existing deployment has not been changed
	require.Equal(t, "my-container", deployment.Spec.Template.Spec.Containers[0].Name)
}

func TestFlagdProxyHandler_HandleFlagdProxy_ProxyExistsWithNewestVersion(t *testing.T) {
	env := types.EnvConfig{
		PodNamespace: "ns",
	}
	kpConfig := NewFlagdProxyConfiguration(env, pullSecrets, labels, annotations)

	require.NotNil(t, kpConfig)

	fakeClient := fake.NewClientBuilder().WithObjects(createOFOTestDeployment(env.PodNamespace)).Build()

	ph := NewFlagdProxyHandler(kpConfig, fakeClient, testr.New(t))

	require.NotNil(t, ph)

	ownerRef, err := getTestOFODeploymentOwnerRef(fakeClient, env.PodNamespace)
	require.Nil(t, err)

	proxy := ph.newFlagdProxyDeployment(ownerRef)

	err = fakeClient.Create(context.TODO(), proxy)
	require.Nil(t, err)

	err = ph.HandleFlagdProxy(context.Background())

	require.Nil(t, err)

	deployment := &appsv1.Deployment{}
	err = fakeClient.Get(context.Background(), client.ObjectKey{
		Namespace: env.PodNamespace,
		Name:      FlagdProxyDeploymentName,
	}, deployment)

	require.Nil(t, err)
	require.NotNil(t, deployment)

	// verify that the existing deployment has not been changed
	require.Equal(t, "flagd-proxy", deployment.Spec.Template.Spec.Containers[0].Name)
}

func TestFlagdProxyHandler_HandleFlagdProxy_CreateProxy(t *testing.T) {
	kpConfig := NewFlagdProxyConfiguration(testEnvConfig, pullSecrets, labels, annotations)

	require.NotNil(t, kpConfig)

	fakeClient := fake.NewClientBuilder().WithObjects(createOFOTestDeployment(testNamespace)).Build()

	ph := NewFlagdProxyHandler(kpConfig, fakeClient, testr.New(t))

	require.NotNil(t, ph)

	// proxy does not exist
	deployment := &appsv1.Deployment{}
	err := fakeClient.Get(context.Background(), client.ObjectKey{
		Namespace: testNamespace,
		Name:      FlagdProxyDeploymentName,
	}, deployment)

	require.NotNil(t, err)

	err = ph.HandleFlagdProxy(context.Background())

	require.Nil(t, err)

	// proxy should exist
	deployment = &appsv1.Deployment{}
	err = fakeClient.Get(context.Background(), client.ObjectKey{
		Namespace: testNamespace,
		Name:      FlagdProxyDeploymentName,
	}, deployment)

	require.Nil(t, err)
	require.NotNil(t, deployment)

	require.Equal(t, expectedDeployment, deployment)

	service := &corev1.Service{}
	err = fakeClient.Get(context.Background(), client.ObjectKey{
		Namespace: testNamespace,
		Name:      FlagdProxyServiceName,
	}, service)

	require.Nil(t, err)
	require.NotNil(t, service)

	require.Equal(t, expectedService, service)

	pdb := &policyv1.PodDisruptionBudget{}
	err = fakeClient.Get(context.Background(), client.ObjectKey{
		Namespace: testNamespace,
		Name:      FlagdProxyPodDisruptionBudgetName,
	}, pdb)
	require.Nil(t, err)
	require.NotNil(t, pdb)

	require.Equal(t, expectedPDB, pdb)
}

func TestFlagdProxyHandler_HandleFlagdProxy_UpdateAllComponents(t *testing.T) {
	kpConfig := NewFlagdProxyConfiguration(testEnvConfig, pullSecrets, labels, annotations)
	require.NotNil(t, kpConfig)

	fakeClient := fake.NewClientBuilder().WithObjects(createOFOTestDeployment(testNamespace)).Build()

	ph := NewFlagdProxyHandler(kpConfig, fakeClient, testr.New(t))
	require.NotNil(t, ph)

	// Seed with slightly different values
	deploy := expectedDeployment.DeepCopy()
	deploy.ResourceVersion = ""
	deploy.Spec.Replicas = pointer.Int32(100000)
	require.Nil(t, fakeClient.Create(context.Background(), deploy))

	svc := expectedService.DeepCopy()
	svc.ResourceVersion = ""
	svc.Spec.Ports[0].Port = 100000
	require.Nil(t, fakeClient.Create(context.Background(), svc))

	pdb := expectedPDB.DeepCopy()
	pdb.ResourceVersion = ""
	minAvailable := intstr.FromInt(100000)
	pdb.Spec.MinAvailable = &minAvailable
	require.Nil(t, fakeClient.Create(context.Background(), pdb))

	// Run
	err := ph.HandleFlagdProxy(context.Background())
	require.Nil(t, err)

	// Get the updated resources
	require.Nil(t, fakeClient.Get(context.Background(), client.ObjectKey{
		Namespace: testNamespace,
		Name:      FlagdProxyDeploymentName,
	}, deploy))
	updatedExpectedDeployment := expectedDeployment.DeepCopy()
	updatedExpectedDeployment.ResourceVersion = "2"
	require.Equal(t, updatedExpectedDeployment, deploy)

	require.Nil(t, fakeClient.Get(context.Background(), client.ObjectKey{
		Namespace: testNamespace,
		Name:      FlagdProxyServiceName,
	}, svc))
	updatedExpectedService := expectedService.DeepCopy()
	updatedExpectedService.ResourceVersion = "2"
	require.Equal(t, updatedExpectedService, svc)

	// pdb := expectedPDB.DeepCopy()
	require.Nil(t, fakeClient.Get(context.Background(), client.ObjectKey{
		Namespace: testNamespace,
		Name:      FlagdProxyPodDisruptionBudgetName,
	}, pdb))
	updatedExpectedPDB := expectedPDB.DeepCopy()
	updatedExpectedPDB.ResourceVersion = "2"
	require.Equal(t, updatedExpectedPDB, pdb)
}

func createOFOTestDeployment(ns string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      common.OperatorDeploymentName,
		},
	}
}

func getTestOFODeploymentOwnerRef(c client.Client, ns string) (*metav1.OwnerReference, error) {
	d := &appsv1.Deployment{}
	if err := c.Get(context.TODO(), client.ObjectKey{Name: common.OperatorDeploymentName, Namespace: ns}, d); err != nil {
		return nil, err
	}
	return &metav1.OwnerReference{
		UID:        d.GetUID(),
		Name:       d.GetName(),
		APIVersion: d.APIVersion,
		Kind:       d.Kind,
	}, nil
}
