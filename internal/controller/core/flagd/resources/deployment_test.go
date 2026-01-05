// nolint:dupl
package resources

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	"github.com/open-feature/open-feature-operator/internal/common"
	commonfake "github.com/open-feature/open-feature-operator/internal/common/flagdinjector/fake"
	resources "github.com/open-feature/open-feature-operator/internal/controller/core/flagd/common"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var testFlagdConfig = resources.FlagdConfiguration{
	FlagdPort:              8013,
	OFREPPort:              8016,
	SyncPort:               8015,
	ManagementPort:         8014,
	DebugLogging:           false,
	Image:                  "flagd",
	Tag:                    "latest",
	OperatorNamespace:      "ofo-system",
	OperatorDeploymentName: "ofo",
}

func TestFlagdDeployment_getFlagdDeployment(t *testing.T) {
	err := api.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	flagdObj := &api.Flagd{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-flagd",
			Namespace: "my-namespace",
		},
		Spec: api.FlagdSpec{
			FeatureFlagSource: "my-flag-source",
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

	res, err := r.GetResource(context.Background(), flagdObj)

	require.Nil(t, err)
	require.NotNil(t, res)

	deploymentResult := res.(*appsv1.Deployment)

	require.Equal(t, flagdObj.Name, deploymentResult.Name)
	require.Equal(t, flagdObj.Namespace, deploymentResult.Namespace)
	require.Len(t, deploymentResult.OwnerReferences, 1)
	require.Equal(
		t,
		fmt.Sprintf("%s:%s", r.FlagdConfig.Image, r.FlagdConfig.Tag),
		deploymentResult.Spec.Template.Spec.Containers[0].Image,
	)
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
			FeatureFlagSource: "my-flag-source",
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

	deploymentResult, err := r.GetResource(context.Background(), flagdObj)

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
			FeatureFlagSource: "my-flag-source",
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

	deploymentResult, err := r.GetResource(context.Background(), flagdObj)

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
			FeatureFlagSource: "my-flag-source",
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

	deploymentResult, err := r.GetResource(context.Background(), flagdObj)

	require.NotNil(t, err)
	require.Nil(t, deploymentResult)
}

func Test_areDeploymentsEqual(t *testing.T) {
	type args struct {
		old client.Object
		new client.Object
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "has spec changed",
			args: args{
				old: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Replicas: intPtr(1),
					},
				},
				new: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Replicas: intPtr(2),
					},
				},
			},
			want: false,
		},
		{
			name: "has labels changed",
			args: args{
				old: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"key": "old",
						},
					},
				},
				new: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"key": "new",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "has annotations changed",
			args: args{
				old: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"key": "old",
						},
					},
				},
				new: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"key": "new",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "has not changed",
			args: args{
				old: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Replicas: intPtr(1),
					},
				},
				new: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Replicas: intPtr(1),
					},
				},
			},
			want: true,
		},
		{
			name: "old is not a deployment",
			args: args{
				old: &v1.ConfigMap{},
				new: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Replicas: intPtr(1),
					},
				},
			},
			want: false,
		},
		{
			name: "new is not a deployment",
			args: args{
				old: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Replicas: intPtr(1),
					},
				},
				new: &v1.ConfigMap{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &FlagdDeployment{}
			got := d.AreObjectsEqual(tt.args.old, tt.args.new)

			require.Equal(t, tt.want, got)
		})
	}
}

func Test_getLabels(t *testing.T) {
	const (
		flagdConfigTag = "latest"
		flagdName      = "test-flagd"
	)
	type args struct {
		flagdConfigLabels map[string]string
		flagdLabels       map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "no config labels, no flagd labels",
			args: args{
				flagdConfigLabels: nil,
				flagdLabels:       nil,
			},
			want: map[string]string{
				"app":                          flagdName,
				"app.kubernetes.io/name":       flagdName,
				"app.kubernetes.io/managed-by": common.ManagedByAnnotationValue,
				"app.kubernetes.io/version":    flagdConfigTag,
			},
		},
		{
			name: "unique config and flagd labels",
			args: args{
				flagdConfigLabels: map[string]string{
					"config-label": "config-value",
				},
				flagdLabels: map[string]string{
					"flagd-label": "flagd-value",
				},
			},
			want: map[string]string{
				"app":                          flagdName,
				"app.kubernetes.io/name":       flagdName,
				"app.kubernetes.io/managed-by": common.ManagedByAnnotationValue,
				"app.kubernetes.io/version":    flagdConfigTag,
				"config-label":                 "config-value",
				"flagd-label":                  "flagd-value",
			},
		},
		{
			name: "overlapping config and flagd labels",
			args: args{
				flagdConfigLabels: map[string]string{
					"overlapping": "config-value",
				},
				flagdLabels: map[string]string{
					"overlapping": "flagd-value",
				},
			},
			want: map[string]string{
				"app":                          flagdName,
				"app.kubernetes.io/name":       flagdName,
				"app.kubernetes.io/managed-by": common.ManagedByAnnotationValue,
				"app.kubernetes.io/version":    flagdConfigTag,
				"overlapping":                  "flagd-value",
			},
		},
		{
			name: "overlapping default labels",
			args: args{
				flagdConfigLabels: map[string]string{
					"app.kubernetes.io/name": "config-value",
				},
				flagdLabels: map[string]string{
					"app.kubernetes.io/name": "flagd-value",
				},
			},
			want: map[string]string{
				"app":                          flagdName,
				"app.kubernetes.io/name":       flagdName,
				"app.kubernetes.io/managed-by": common.ManagedByAnnotationValue,
				"app.kubernetes.io/version":    flagdConfigTag,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &FlagdDeployment{
				FlagdConfig: resources.FlagdConfiguration{
					Labels: tt.args.flagdConfigLabels,
					Tag:    flagdConfigTag,
				},
			}
			flagd := &api.Flagd{
				ObjectMeta: metav1.ObjectMeta{
					Name: flagdName,
				},
				Spec: api.FlagdSpec{
					PodLabels: tt.args.flagdLabels,
				},
			}

			got := r.getLabels(flagd)

			require.Equal(t, tt.want, got)
		})
	}
}

func Test_getAnnotations(t *testing.T) {
	const (
		flagdName = "test-flagd"
	)
	type args struct {
		flagdConfigAnnotations map[string]string
		flagdAnnotations       map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "no config annotations, no flagd annotations",
			args: args{
				flagdConfigAnnotations: nil,
				flagdAnnotations:       nil,
			},
			want: map[string]string{},
		},
		{
			name: "unique annotations and flagd annotations",
			args: args{
				flagdConfigAnnotations: map[string]string{
					"config-annotation": "config-value",
				},
				flagdAnnotations: map[string]string{
					"flagd-annotation": "flagd-value",
				},
			},
			want: map[string]string{
				"config-annotation": "config-value",
				"flagd-annotation":  "flagd-value",
			},
		},
		{
			name: "overlapping config and flagd labels",
			args: args{
				flagdConfigAnnotations: map[string]string{
					"overlapping": "config-value",
				},
				flagdAnnotations: map[string]string{
					"overlapping": "flagd-value",
				},
			},
			want: map[string]string{
				"overlapping": "flagd-value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &FlagdDeployment{
				FlagdConfig: resources.FlagdConfiguration{
					Annotations: tt.args.flagdConfigAnnotations,
				},
			}
			flagd := &api.Flagd{
				ObjectMeta: metav1.ObjectMeta{
					Name: flagdName,
				},
				Spec: api.FlagdSpec{
					PodAnnotations: tt.args.flagdAnnotations,
				},
			}

			got := r.getAnnotations(flagd)

			require.Equal(t, tt.want, got)
		})
	}
}

func intPtr(i int32) *int32 {
	return &i
}
