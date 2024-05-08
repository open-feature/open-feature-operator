package flagd

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	"github.com/open-feature/open-feature-operator/common"
	resourcemock "github.com/open-feature/open-feature-operator/controllers/core/flagd/resources/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestResourceReconciler_Reconcile_CreateResource(t *testing.T) {
	err := api.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()

	r := &ResourceReconciler{
		Client: fakeClient,
		Scheme: fakeClient.Scheme(),
		Log:    controllerruntime.Log.WithName("resource-reconciler"),
	}

	flagdObj := &api.Flagd{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-flagd",
			Namespace: "my-namespace",
		},
	}

	ctrl := gomock.NewController(t)
	mockRes := resourcemock.NewMockIFlagdResource(ctrl)
	mockRes.EXPECT().
		GetResource(gomock.Any(), flagdMatcher{flagdObj: *flagdObj}).
		Times(1).
		Return(&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: flagdObj.Namespace,
				Name:      flagdObj.Name,
			},
			Data: map[string]string{},
		}, nil)

	err = r.Reconcile(
		context.Background(),
		flagdObj,
		&corev1.ConfigMap{},
		mockRes,
	)

	require.Nil(t, err)

	result := &corev1.ConfigMap{}
	err = fakeClient.Get(context.Background(), client.ObjectKey{
		Namespace: flagdObj.Namespace,
		Name:      flagdObj.Name,
	}, result)

	require.Nil(t, err)
}

func TestResourceReconciler_Reconcile_UpdateManagedResource(t *testing.T) {
	err := api.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	flagdObj := &api.Flagd{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-flagd",
			Namespace: "my-namespace",
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: flagdObj.Namespace,
			Name:      flagdObj.Name,
			Labels: map[string]string{
				common.ManagedByAnnotationKey: common.ManagedByAnnotationValue,
			},
		},
		Data: map[string]string{},
	}).Build()

	r := &ResourceReconciler{
		Client: fakeClient,
		Scheme: fakeClient.Scheme(),
		Log:    controllerruntime.Log.WithName("resource-reconciler"),
	}

	ctrl := gomock.NewController(t)
	mockRes := resourcemock.NewMockIFlagdResource(ctrl)
	mockRes.EXPECT().
		GetResource(gomock.Any(), flagdMatcher{flagdObj: *flagdObj}).
		Times(1).
		Return(&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: flagdObj.Namespace,
				Name:      flagdObj.Name,
			},
			Data: map[string]string{
				"foo": "bar",
			},
		}, nil)

	mockRes.EXPECT().AreObjectsEqual(gomock.Any(), gomock.Any()).Return(false)

	err = r.Reconcile(
		context.Background(),
		flagdObj,
		&corev1.ConfigMap{},
		mockRes,
	)

	require.Nil(t, err)

	result := &corev1.ConfigMap{}
	err = fakeClient.Get(context.Background(), client.ObjectKey{
		Namespace: flagdObj.Namespace,
		Name:      flagdObj.Name,
	}, result)

	require.Nil(t, err)

	// verify the resource was updated
	require.Equal(t, "bar", result.Data["foo"])
}

func TestResourceReconciler_Reconcile_CannotCreateResource(t *testing.T) {
	err := api.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	flagdObj := &api.Flagd{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-flagd",
			Namespace: "my-namespace",
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: flagdObj.Namespace,
			Name:      flagdObj.Name,
			Labels: map[string]string{
				common.ManagedByAnnotationKey: common.ManagedByAnnotationValue,
			},
		},
		Data: map[string]string{},
	}).Build()

	r := &ResourceReconciler{
		Client: fakeClient,
		Scheme: fakeClient.Scheme(),
		Log:    controllerruntime.Log.WithName("resource-reconciler"),
	}

	ctrl := gomock.NewController(t)
	mockRes := resourcemock.NewMockIFlagdResource(ctrl)
	mockRes.EXPECT().
		GetResource(gomock.Any(), flagdMatcher{flagdObj: *flagdObj}).
		Times(1).
		Return(nil, errors.New("oops"))

	err = r.Reconcile(
		context.Background(),
		flagdObj,
		&corev1.ConfigMap{},
		mockRes,
	)

	require.NotNil(t, err)
}

func TestResourceReconciler_Reconcile_UnmanagedResourceAlreadyExists(t *testing.T) {
	err := api.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	flagdObj := &api.Flagd{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-flagd",
			Namespace: "my-namespace",
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: flagdObj.Namespace,
			Name:      flagdObj.Name,
		},
		Data: map[string]string{},
	}).Build()

	r := &ResourceReconciler{
		Client: fakeClient,
		Scheme: fakeClient.Scheme(),
		Log:    controllerruntime.Log.WithName("resource-reconciler"),
	}
	ctrl := gomock.NewController(t)
	mockRes := resourcemock.NewMockIFlagdResource(ctrl)

	err = r.Reconcile(
		context.Background(),
		flagdObj,
		&corev1.ConfigMap{},
		mockRes,
	)

	require.NotNil(t, err)
}
