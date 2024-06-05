package flagdproxy

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-logr/logr/testr"
	"github.com/open-feature/open-feature-operator/common"
	"github.com/open-feature/open-feature-operator/common/types"
	"github.com/stretchr/testify/require"
	appsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var pullSecrets = []string{"test-pullSecret"}

func TestNewFlagdProxyConfiguration(t *testing.T) {

	kpConfig := NewFlagdProxyConfiguration(types.EnvConfig{
		FlagdProxyPort:           8015,
		FlagdProxyManagementPort: 8016,
	}, pullSecrets)

	require.NotNil(t, kpConfig)
	require.Equal(t, &FlagdProxyConfiguration{
		Port:                   8015,
		ManagementPort:         8016,
		DebugLogging:           false,
		OperatorDeploymentName: common.OperatorDeploymentName,
		ImagePullSecrets:       pullSecrets,
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

	kpConfig := NewFlagdProxyConfiguration(env, pullSecrets)

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
	}, kpConfig)
}

func TestNewFlagdProxyHandler(t *testing.T) {
	kpConfig := NewFlagdProxyConfiguration(types.EnvConfig{}, pullSecrets)

	require.NotNil(t, kpConfig)

	fakeClient := fake.NewClientBuilder().Build()

	ph := NewFlagdProxyHandler(kpConfig, fakeClient, testr.New(t))

	require.NotNil(t, ph)

	require.Equal(t, kpConfig, ph.Config())
}

func TestDoesFlagdProxyExist(t *testing.T) {
	env := types.EnvConfig{
		PodNamespace: "ns",
	}

	deployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      FlagdProxyDeploymentName,
		},
		Spec: v1.DeploymentSpec{
			Template: v12.PodTemplateSpec{
				Spec: v12.PodSpec{
					Containers: []v12.Container{
						{
							Name: "my-container",
						},
					},
				},
			},
		},
	}

	kpConfig := NewFlagdProxyConfiguration(env, pullSecrets)

	require.NotNil(t, kpConfig)

	fakeClient := fake.NewClientBuilder().WithObjects().Build()

	ph := NewFlagdProxyHandler(kpConfig, fakeClient, testr.New(t))

	require.NotNil(t, ph)

	res, d, err := ph.doesFlagdProxyExist(context.TODO())
	require.Nil(t, err)
	require.Nil(t, d)
	require.False(t, res)

	err = fakeClient.Create(context.TODO(), deployment)
	require.Nil(t, err)

	res, d, err = ph.doesFlagdProxyExist(context.TODO())
	require.Nil(t, err)
	require.NotNil(t, d)
	require.True(t, res)
}

func TestFlagdProxyHandler_HandleFlagdProxy_ProxyExistsWithBadVersion(t *testing.T) {
	env := types.EnvConfig{
		PodNamespace: "ns",
	}
	kpConfig := NewFlagdProxyConfiguration(env, pullSecrets)

	require.NotNil(t, kpConfig)

	fakeClient := fake.NewClientBuilder().WithObjects(createOFOTestDeployment(env.PodNamespace)).Build()

	ownerRef, err := getTestOFODeploymentOwnerRef(fakeClient, env.PodNamespace)
	require.Nil(t, err)

	proxyDeployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       kpConfig.Namespace,
			Name:            FlagdProxyDeploymentName,
			OwnerReferences: []metav1.OwnerReference{*ownerRef},
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": common.ManagedByAnnotationValue,
			},
		},
		Spec: v1.DeploymentSpec{
			Template: v12.PodTemplateSpec{
				Spec: v12.PodSpec{
					Containers: []v12.Container{
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

	deployment := &v1.Deployment{}
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
	kpConfig := NewFlagdProxyConfiguration(env, pullSecrets)

	require.NotNil(t, kpConfig)

	proxyDeployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: kpConfig.Namespace,
			Name:      FlagdProxyDeploymentName,
		},
		Spec: v1.DeploymentSpec{
			Template: v12.PodTemplateSpec{
				Spec: v12.PodSpec{
					Containers: []v12.Container{
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

	require.Nil(t, err)

	deployment := &v1.Deployment{}
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
	kpConfig := NewFlagdProxyConfiguration(env, pullSecrets)

	require.NotNil(t, kpConfig)

	fakeClient := fake.NewClientBuilder().WithObjects(createOFOTestDeployment(env.PodNamespace)).Build()

	ph := NewFlagdProxyHandler(kpConfig, fakeClient, testr.New(t))

	require.NotNil(t, ph)

	ownerRef, err := getTestOFODeploymentOwnerRef(fakeClient, env.PodNamespace)
	require.Nil(t, err)

	proxy := ph.newFlagdProxyManifest(ownerRef)

	err = fakeClient.Create(context.TODO(), proxy)
	require.Nil(t, err)

	err = ph.HandleFlagdProxy(context.Background())

	require.Nil(t, err)

	deployment := &v1.Deployment{}
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
	env := types.EnvConfig{
		PodNamespace:             "ns",
		FlagdProxyImage:          "image",
		FlagdProxyTag:            "tag",
		FlagdProxyPort:           88,
		FlagdProxyManagementPort: 90,
		FlagdProxyDebugLogging:   true,
	}
	kpConfig := NewFlagdProxyConfiguration(env, pullSecrets)

	require.NotNil(t, kpConfig)

	fakeClient := fake.NewClientBuilder().WithObjects(createOFOTestDeployment(env.PodNamespace)).Build()

	ph := NewFlagdProxyHandler(kpConfig, fakeClient, testr.New(t))

	require.NotNil(t, ph)

	// proxy does not exist
	deployment := &v1.Deployment{}
	err := fakeClient.Get(context.Background(), client.ObjectKey{
		Namespace: env.PodNamespace,
		Name:      FlagdProxyDeploymentName,
	}, deployment)

	require.NotNil(t, err)

	err = ph.HandleFlagdProxy(context.Background())

	require.Nil(t, err)

	// proxy should exist
	deployment = &v1.Deployment{}
	err = fakeClient.Get(context.Background(), client.ObjectKey{
		Namespace: env.PodNamespace,
		Name:      FlagdProxyDeploymentName,
	}, deployment)

	require.Nil(t, err)
	require.NotNil(t, deployment)

	replicas := int32(1)
	args := []string{
		"start",
		"--management-port",
		fmt.Sprintf("%d", 90),
		"--debug",
	}

	expectedDeployment := &appsV1.Deployment{
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
				"app.kubernetes.io/version":    "tag",
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
		Spec: appsV1.DeploymentSpec{
			Replicas: &replicas,
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
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: FlagdProxyServiceAccountName,
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: pullSecrets[0]},
					},
					Containers: []corev1.Container{
						{
							Image: "image:tag",
							Name:  FlagdProxyDeploymentName,
							Ports: []corev1.ContainerPort{
								{
									Name:          "port",
									ContainerPort: int32(88),
								},
								{
									Name:          "management-port",
									ContainerPort: int32(90),
								},
							},
							Args: args,
						},
					},
				},
			},
		},
	}

	require.Equal(t, expectedDeployment, deployment)

	service := &corev1.Service{}
	err = fakeClient.Get(context.Background(), client.ObjectKey{
		Namespace: env.PodNamespace,
		Name:      FlagdProxyServiceName,
	}, service)

	require.Nil(t, err)
	require.NotNil(t, service)

	expectedService := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            FlagdProxyServiceName,
			Namespace:       "ns",
			ResourceVersion: "1",
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
					Port:       int32(88),
					TargetPort: intstr.FromInt(88),
				},
			},
		},
	}

	require.Equal(t, expectedService, service)
}

func createOFOTestDeployment(ns string) *v1.Deployment {
	return &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      common.OperatorDeploymentName,
		},
	}
}

func getTestOFODeploymentOwnerRef(c client.Client, ns string) (*metav1.OwnerReference, error) {
	d := &appsV1.Deployment{}
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
