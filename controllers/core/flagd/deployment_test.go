package flagd

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	commonfake "github.com/open-feature/open-feature-operator/common/flagdinjector/fake"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestFlagdDeployment_getFlagdDeployment(t *testing.T) {
	err := api.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	flagdObj := &api.Flagd{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-flagd",
			Namespace: "my-namespace",
		},
		Spec: api.FlagdSpec{
			FeatureFlagSourceRef: v1.ObjectReference{
				Name:      "my-flag-source",
				Namespace: "my-namespace",
			},
		},
	}

	flagSource := &api.FeatureFlagSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-flag-source",
			Namespace: "my-namespace",
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(flagSource, flagdObj).Build()

	ctrl := gomock.NewController(t)

	fakeFlagdInjector := commonfake.NewMockFlagdContainerInjector(ctrl)
	fakeFlagdInjector.EXPECT().
		InjectFlagd(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Times(1).
		DoAndReturn(func(
			ctx context.Context,
			objectMeta *metav1.ObjectMeta,
			podSpec *v1.PodSpec,
			flagSourceConfig *api.FeatureFlagSourceSpec,
		) error {
			// simulate the injection of a container into the podspec
			podSpec.Containers = []v1.Container{
				{
					Name: "flagd",
				},
			}
			return nil
		})

	r := &FlagdDeployment{
		Client:        fakeClient,
		Log:           controllerruntime.Log.WithName("test"),
		FlagdInjector: fakeFlagdInjector,
		FlagdConfig:   testFlagdConfig,
	}

	deploymentResult, err := r.getFlagdDeployment(context.Background(), flagdObj)

	require.Nil(t, err)
	require.NotNil(t, deploymentResult)

	require.Equal(t, flagdObj.Name, deploymentResult.Name)
	require.Equal(t, flagdObj.Namespace, deploymentResult.Namespace)
	require.Len(t, deploymentResult.OwnerReferences, 1)
	require.Equal(t, []v1.ContainerPort{
		{
			Name:          "management",
			ContainerPort: int32(r.FlagdConfig.ManagementPort),
		},
		{
			Name:          "flagd",
			ContainerPort: int32(r.FlagdConfig.FlagdPort),
		},
		{
			Name:          "ofrep",
			ContainerPort: int32(r.FlagdConfig.OFREPPort),
		},
		{
			Name:          "sync",
			ContainerPort: int32(r.FlagdConfig.SyncPort),
		},
	}, deploymentResult.Spec.Template.Spec.Containers[0].Ports)
}

func TestFlagdDeployment_getFlagdDeployment_ErrorInInjector(t *testing.T) {
	err := api.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	flagdObj := &api.Flagd{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-flagd",
			Namespace: "my-namespace",
		},
		Spec: api.FlagdSpec{
			FeatureFlagSourceRef: v1.ObjectReference{
				Name:      "my-flag-source",
				Namespace: "my-namespace",
			},
		},
	}

	flagSource := &api.FeatureFlagSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-flag-source",
			Namespace: "my-namespace",
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(flagSource, flagdObj).Build()

	ctrl := gomock.NewController(t)

	fakeFlagdInjector := commonfake.NewMockFlagdContainerInjector(ctrl)
	fakeFlagdInjector.EXPECT().
		InjectFlagd(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Times(1).
		Return(errors.New("oops"))

	r := &FlagdDeployment{
		Client:        fakeClient,
		Log:           controllerruntime.Log.WithName("test"),
		FlagdInjector: fakeFlagdInjector,
		FlagdConfig:   testFlagdConfig,
	}

	deploymentResult, err := r.getFlagdDeployment(context.Background(), flagdObj)

	require.NotNil(t, err)
	require.Nil(t, deploymentResult)
}

func TestFlagdDeployment_getFlagdDeployment_ContainerNotInjected(t *testing.T) {
	err := api.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	flagdObj := &api.Flagd{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-flagd",
			Namespace: "my-namespace",
		},
		Spec: api.FlagdSpec{
			FeatureFlagSourceRef: v1.ObjectReference{
				Name:      "my-flag-source",
				Namespace: "my-namespace",
			},
		},
	}

	flagSource := &api.FeatureFlagSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-flag-source",
			Namespace: "my-namespace",
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(flagSource, flagdObj).Build()

	ctrl := gomock.NewController(t)

	fakeFlagdInjector := commonfake.NewMockFlagdContainerInjector(ctrl)
	fakeFlagdInjector.EXPECT().
		InjectFlagd(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Times(1).
		Return(nil)

	r := &FlagdDeployment{
		Client:        fakeClient,
		Log:           controllerruntime.Log.WithName("test"),
		FlagdInjector: fakeFlagdInjector,
		FlagdConfig:   testFlagdConfig,
	}

	deploymentResult, err := r.getFlagdDeployment(context.Background(), flagdObj)

	require.NotNil(t, err)
	require.Nil(t, deploymentResult)
}

func TestFlagdDeployment_getFlagdDeployment_FlagSourceNotFound(t *testing.T) {
	err := api.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	flagdObj := &api.Flagd{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-flagd",
			Namespace: "my-namespace",
		},
		Spec: api.FlagdSpec{
			FeatureFlagSourceRef: v1.ObjectReference{
				Name:      "my-flag-source",
				Namespace: "my-namespace",
			},
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(flagdObj).Build()

	ctrl := gomock.NewController(t)

	fakeFlagdInjector := commonfake.NewMockFlagdContainerInjector(ctrl)

	r := &FlagdDeployment{
		Client:        fakeClient,
		Log:           controllerruntime.Log.WithName("test"),
		FlagdInjector: fakeFlagdInjector,
		FlagdConfig:   testFlagdConfig,
	}

	deploymentResult, err := r.getFlagdDeployment(context.Background(), flagdObj)

	require.NotNil(t, err)
	require.Nil(t, deploymentResult)
}
