package flagd

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	commonmock "github.com/open-feature/open-feature-operator/controllers/core/flagd/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"reflect"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

var testFlagdConfig = FlagdConfiguration{
	FlagdPort:              8013,
	OFREPPort:              8016,
	ManagementPort:         8014,
	DebugLogging:           false,
	Image:                  "flagd",
	Tag:                    "latest",
	OperatorNamespace:      "ofo-system",
	OperatorDeploymentName: "ofo",
}

type flagdMatcher struct {
	flagdObj api.Flagd
}

func (fm flagdMatcher) Matches(x interface{}) bool {
	flagd, ok := x.(*api.Flagd)
	if !ok {
		return false
	}
	return reflect.DeepEqual(fm.flagdObj.ObjectMeta, flagd.ObjectMeta) && reflect.DeepEqual(fm.flagdObj.Spec, flagd.Spec)
}

// String describes what the matcher matches.
func (fm flagdMatcher) String() string {
	return fmt.Sprintf("%v", fm.flagdObj)
}

func TestFlagdReconciler_Reconcile(t *testing.T) {
	err := api.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	flagdObj := &api.Flagd{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-flagd",
			Namespace: "my-namespace",
		},
		Spec: api.FlagdSpec{},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(flagdObj).Build()

	ctrl := gomock.NewController(t)

	deploymentReconciler := commonmock.NewMockIFlagdResource(ctrl)
	serviceReconciler := commonmock.NewMockIFlagdResource(ctrl)
	ingressReconciler := commonmock.NewMockIFlagdResource(ctrl)

	// deployment creation succeeds
	deploymentReconciler.EXPECT().Reconcile(gomock.Any(), flagdMatcher{flagdObj: *flagdObj}).Times(1).Return(nil)
	serviceReconciler.EXPECT().Reconcile(gomock.Any(), flagdMatcher{flagdObj: *flagdObj}).Times(1).Return(nil)
	ingressReconciler.EXPECT().Reconcile(gomock.Any(), flagdMatcher{flagdObj: *flagdObj}).Times(1).Return(nil)

	r := setupReconciler(fakeClient, deploymentReconciler, serviceReconciler, ingressReconciler)

	result, err := r.Reconcile(context.Background(), controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Namespace: flagdObj.Namespace,
			Name:      flagdObj.Name,
		},
	})

	require.Nil(t, err)
	require.Equal(t, controllerruntime.Result{}, result)
}

func TestFlagdReconciler_ReconcileResourceNotFound(t *testing.T) {
	err := api.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects().Build()

	r := setupReconciler(fakeClient, nil, nil, nil)

	result, err := r.Reconcile(context.Background(), controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "my-namespace",
			Name:      "my-flagd",
		},
	})

	require.Nil(t, err)
	require.Equal(t, controllerruntime.Result{}, result)
}

func TestFlagdReconciler_ReconcileFailDeployment(t *testing.T) {
	err := api.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	flagdObj := &api.Flagd{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-flagd",
			Namespace: "my-namespace",
		},
		Spec: api.FlagdSpec{},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(flagdObj).Build()

	ctrl := gomock.NewController(t)

	deploymentReconciler := commonmock.NewMockIFlagdResource(ctrl)

	// deployment creation succeeds
	deploymentReconciler.EXPECT().Reconcile(gomock.Any(), flagdMatcher{flagdObj: *flagdObj}).Times(1).Return(errors.New("oops"))

	r := setupReconciler(fakeClient, deploymentReconciler, nil, nil)

	result, err := r.Reconcile(context.Background(), controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Namespace: flagdObj.Namespace,
			Name:      flagdObj.Name,
		},
	})

	require.NotNil(t, err)
	require.Equal(t, controllerruntime.Result{}, result)
}

func TestFlagdReconciler_ReconcileFailService(t *testing.T) {
	err := api.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	flagdObj := &api.Flagd{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-flagd",
			Namespace: "my-namespace",
		},
		Spec: api.FlagdSpec{},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(flagdObj).Build()

	ctrl := gomock.NewController(t)

	deploymentReconciler := commonmock.NewMockIFlagdResource(ctrl)
	serviceReconciler := commonmock.NewMockIFlagdResource(ctrl)

	deploymentReconciler.EXPECT().Reconcile(gomock.Any(), flagdMatcher{flagdObj: *flagdObj}).Times(1).Return(nil)
	serviceReconciler.EXPECT().Reconcile(gomock.Any(), flagdMatcher{flagdObj: *flagdObj}).Times(1).Return(errors.New("oops"))

	r := setupReconciler(fakeClient, deploymentReconciler, serviceReconciler, nil)

	result, err := r.Reconcile(context.Background(), controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Namespace: flagdObj.Namespace,
			Name:      flagdObj.Name,
		},
	})

	require.NotNil(t, err)
	require.Equal(t, controllerruntime.Result{}, result)
}

func TestFlagdReconciler_ReconcileFailIngress(t *testing.T) {
	err := api.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	flagdObj := &api.Flagd{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-flagd",
			Namespace: "my-namespace",
		},
		Spec: api.FlagdSpec{},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(flagdObj).Build()

	ctrl := gomock.NewController(t)

	deploymentReconciler := commonmock.NewMockIFlagdResource(ctrl)
	serviceReconciler := commonmock.NewMockIFlagdResource(ctrl)
	ingressReconciler := commonmock.NewMockIFlagdResource(ctrl)

	deploymentReconciler.EXPECT().Reconcile(gomock.Any(), flagdMatcher{flagdObj: *flagdObj}).Times(1).Return(nil)
	serviceReconciler.EXPECT().Reconcile(gomock.Any(), flagdMatcher{flagdObj: *flagdObj}).Times(1).Return(nil)
	ingressReconciler.EXPECT().Reconcile(gomock.Any(), flagdMatcher{flagdObj: *flagdObj}).Times(1).Return(errors.New("oops"))

	r := setupReconciler(fakeClient, deploymentReconciler, serviceReconciler, ingressReconciler)

	result, err := r.Reconcile(context.Background(), controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Namespace: flagdObj.Namespace,
			Name:      flagdObj.Name,
		},
	})

	require.NotNil(t, err)
	require.Equal(t, controllerruntime.Result{}, result)
}

func setupReconciler(fakeClient client.WithWatch, deploymentReconciler *commonmock.MockIFlagdResource, serviceReconciler *commonmock.MockIFlagdResource, ingressReconciler *commonmock.MockIFlagdResource) *FlagdReconciler {
	return &FlagdReconciler{
		Client:          fakeClient,
		Scheme:          fakeClient.Scheme(),
		Log:             controllerruntime.Log.WithName("flagd controller"),
		FlagdConfig:     testFlagdConfig,
		FlagdDeployment: deploymentReconciler,
		FlagdService:    serviceReconciler,
		FlagdIngress:    ingressReconciler,
	}
}
