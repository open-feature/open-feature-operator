package flagdproxy

import (
	"context"
	"testing"

	"github.com/go-logr/logr/testr"
	"github.com/open-feature/open-feature-operator/common/types"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNewFlagdProxyConfiguration(t *testing.T) {
	kpConfig := NewFlagdProxyConfiguration(types.EnvConfig{
		FlagdProxyPort:           8015,
		FlagdProxyManagementPort: 8016,
	})

	require.NotNil(t, kpConfig)
	require.Equal(t, &FlagdProxyConfiguration{
		Port:                   8015,
		ManagementPort:         8016,
		DebugLogging:           false,
		OperatorDeploymentName: operatorDeploymentName,
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

	kpConfig := NewFlagdProxyConfiguration(env)

	require.NotNil(t, kpConfig)
	require.Equal(t, &FlagdProxyConfiguration{
		Port:                   8080,
		ManagementPort:         8081,
		DebugLogging:           true,
		Image:                  "my-image",
		Tag:                    "my-tag",
		Namespace:              "my-namespace",
		OperatorDeploymentName: operatorDeploymentName,
	}, kpConfig)
}

func TestNewFlagdProxyHandler(t *testing.T) {
	kpConfig := NewFlagdProxyConfiguration(types.EnvConfig{})

	require.NotNil(t, kpConfig)

	fakeClient := fake.NewClientBuilder().Build()

	ph := NewFlagdProxyHandler(kpConfig, fakeClient, testr.New(t))

	require.NotNil(t, ph)

	require.Equal(t, kpConfig, ph.Config())
}

func TestFlagdProxyHandler_HandleFlagdProxy_ProxyExists(t *testing.T) {
	env := types.EnvConfig{
		PodNamespace: "ns",
	}
	kpConfig := NewFlagdProxyConfiguration(env)

	require.NotNil(t, kpConfig)

	fakeClient := fake.NewClientBuilder().WithObjects(&v1.Deployment{
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
	}).Build()

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

func TestFlagdProxyHandler_HandleFlagdProxy_CreateProxy(t *testing.T) {
	env := types.EnvConfig{
		PodNamespace: "ns",
	}
	kpConfig := NewFlagdProxyConfiguration(env)

	require.NotNil(t, kpConfig)

	fakeClient := fake.NewClientBuilder().Build()

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

	require.Equal(t, FlagdProxyDeploymentName, deployment.Spec.Template.Spec.Containers[0].Name)
}
