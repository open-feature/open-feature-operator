// Code generated by MockGen. DO NOT EDIT.
// Source: ./controllers/core/flagd/controller.go

// Package commonmock is a generated GoMock package.
package commonmock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1beta1 "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	resources "github.com/open-feature/open-feature-operator/controllers/core/flagd/resources"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// MockIFlagdResourceReconciler is a mock of IFlagdResourceReconciler interface.
type MockIFlagdResourceReconciler struct {
	ctrl     *gomock.Controller
	recorder *MockIFlagdResourceReconcilerMockRecorder
}

// MockIFlagdResourceReconcilerMockRecorder is the mock recorder for MockIFlagdResourceReconciler.
type MockIFlagdResourceReconcilerMockRecorder struct {
	mock *MockIFlagdResourceReconciler
}

// NewMockIFlagdResourceReconciler creates a new mock instance.
func NewMockIFlagdResourceReconciler(ctrl *gomock.Controller) *MockIFlagdResourceReconciler {
	mock := &MockIFlagdResourceReconciler{ctrl: ctrl}
	mock.recorder = &MockIFlagdResourceReconcilerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIFlagdResourceReconciler) EXPECT() *MockIFlagdResourceReconcilerMockRecorder {
	return m.recorder
}

// Reconcile mocks base method.
func (m *MockIFlagdResourceReconciler) Reconcile(ctx context.Context, flagd *v1beta1.Flagd, obj client.Object, resource resources.IFlagdResource) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reconcile", ctx, flagd, obj, resource)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reconcile indicates an expected call of Reconcile.
func (mr *MockIFlagdResourceReconcilerMockRecorder) Reconcile(ctx, flagd, obj, resource interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reconcile", reflect.TypeOf((*MockIFlagdResourceReconciler)(nil).Reconcile), ctx, flagd, obj, resource)
}
