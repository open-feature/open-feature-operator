package webhooks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/go-logr/logr/testr"
	"github.com/golang/mock/gomock"
	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	apicommon "github.com/open-feature/open-feature-operator/apis/core/v1beta1/common"
	"github.com/open-feature/open-feature-operator/common"
	flagdinjectorfake "github.com/open-feature/open-feature-operator/common/flagdinjector/fake"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	mutatePodNamespace           = "test-mutate-pod"
	defaultPodServiceAccountName = "test-pod-service-account"
	featureFlagSourceName        = "test-feature-flag-source"
	inProcessConfigurationName   = "test-feature-flag-in-process-configuration"
)

func TestPodMutator_BackfillPermissions(t *testing.T) {
	const (
		ns   = "mynamespace"
		pod  = "mypod"
		name = "default"
	)

	tests := []struct {
		name    string
		mutator *PodMutator
		wantErr bool
		setup   func(injector *flagdinjectorfake.MockFlagdContainerInjector)
	}{
		{
			name: "no annotated pod",
			mutator: &PodMutator{
				Client:  NewClient(false),
				decoder: nil,
				Log:     testr.New(t),
				ready:   false,
			},
			wantErr: true,
		},
		{
			name: "pod is annotated",
			mutator: &PodMutator{
				Log: testr.New(t),
				Client: NewClient(true,
					&corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      pod,
							Namespace: ns,
							Annotations: map[string]string{
								fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.EnabledAnnotation):             "true",
								fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.FeatureFlagSourceAnnotation):   fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagSourceName),
								fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.AllowKubernetesSyncAnnotation): "true",
							}},
					},
				),
			},
			setup: func(injector *flagdinjectorfake.MockFlagdContainerInjector) {
				injector.EXPECT().EnableClusterRoleBinding(
					gomock.Any(),
					ns,
					"",
				).Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name: "pod is annotated, ClusterRoleBinding cannot be enabled; continue with other pods",
			mutator: &PodMutator{
				Log: testr.New(t),
				Client: NewClient(true,
					&corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      pod + "-1",
							Namespace: ns,
							Annotations: map[string]string{
								fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.EnabledAnnotation):             "true",
								fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.FeatureFlagSourceAnnotation):   fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagSourceName),
								fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.AllowKubernetesSyncAnnotation): "true",
							}},
					},
					&corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      pod + "-2",
							Namespace: ns,
							Annotations: map[string]string{
								fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.EnabledAnnotation):             "true",
								fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.FeatureFlagSourceAnnotation):   fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagSourceName),
								fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.AllowKubernetesSyncAnnotation): "true",
							}},
					},
				),
			},
			setup: func(injector *flagdinjectorfake.MockFlagdContainerInjector) {
				// make the mock return an error - in this case we still expect the number of invocations
				// to match the number of pods
				injector.EXPECT().EnableClusterRoleBinding(
					gomock.Any(),
					ns,
					"",
				).Return(errors.New("error")).Times(2)
			},
			wantErr: false,
		},
		{
			name: "Subjects exists: no backfill",
			mutator: &PodMutator{
				Log: testr.New(t),
				Client: NewClient(true,
					&corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      pod,
							Namespace: ns,
							Annotations: map[string]string{
								fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.EnabledAnnotation):             "true",
								fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.FeatureFlagSourceAnnotation):   fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagSourceName),
								fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.AllowKubernetesSyncAnnotation): "true",
							}},
						Spec: corev1.PodSpec{ServiceAccountName: "my-service-account"},
					},
					&corev1.ServiceAccount{
						ObjectMeta: metav1.ObjectMeta{
							Name:      name,
							Namespace: ns,
							Annotations: map[string]string{
								fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.EnabledAnnotation):             "true",
								fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.FeatureFlagSourceAnnotation):   fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagSourceName),
								fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.AllowKubernetesSyncAnnotation): "true",
							}},
					},
					&rbac.ClusterRoleBinding{
						ObjectMeta: metav1.ObjectMeta{
							Name: "open-feature-operator-flagd-kubernetes-sync",
						},
					},
				),
			},
			setup: func(injector *flagdinjectorfake.MockFlagdContainerInjector) {
				injector.EXPECT().EnableClusterRoleBinding(context.TODO(), ns, "my-service-account").Times(1)
			},
			wantErr: false,
		},
		{
			name: "Subjects does not exist: backfill",
			mutator: &PodMutator{
				Log: testr.New(t),
				Client: NewClient(true,
					&corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      pod,
							Namespace: ns,
							Annotations: map[string]string{
								fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.EnabledAnnotation):             "true",
								fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.FeatureFlagSourceAnnotation):   fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagSourceName),
								fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.AllowKubernetesSyncAnnotation): "true",
							}},
					},
					&corev1.ServiceAccount{
						ObjectMeta: metav1.ObjectMeta{
							Name:      name,
							Namespace: ns,
							Annotations: map[string]string{
								fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.EnabledAnnotation):             "true",
								fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.FeatureFlagSourceAnnotation):   fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagSourceName),
								fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.AllowKubernetesSyncAnnotation): "true",
							}},
					},
					&rbac.ClusterRoleBinding{
						ObjectMeta: metav1.ObjectMeta{
							Name: "open-feature-operator-flagd-kubernetes-sync",
						},
						Subjects: []rbac.Subject{
							{
								Kind:      "ServiceAccount",
								Name:      "new",
								Namespace: ns,
							},
						},
					},
				),
			},
			wantErr: false,
			setup: func(injector *flagdinjectorfake.MockFlagdContainerInjector) {
				injector.EXPECT().EnableClusterRoleBinding(context.TODO(), ns, "").Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			mockInjector := flagdinjectorfake.NewMockFlagdContainerInjector(ctrl)

			if tt.setup != nil {
				tt.setup(mockInjector)
			}
			m := tt.mutator
			m.FlagdInjector = mockInjector
			if err := m.BackfillPermissions(context.TODO()); (err != nil) != tt.wantErr {
				t.Errorf("BackfillPermissions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPodMutator_Handle(t *testing.T) {
	decoder, err := admission.NewDecoder(scheme.Scheme)
	require.Nil(t, err)

	goodPod, err := json.Marshal(corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "myPod"},
	})
	require.Nil(t, err)

	badAnnotatedPod, err := json.Marshal(corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "myAnnotatedPod",
			Annotations: map[string]string{
				fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.EnabledAnnotation):           "true",
				fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.FeatureFlagSourceAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagSourceName),
			},
		},
	})
	require.Nil(t, err)

	antPod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myAnnotatedPod",
			Namespace: mutatePodNamespace,
			Annotations: map[string]string{
				fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.EnabledAnnotation):           "true",
				fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.FeatureFlagSourceAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagSourceName),
			},
			OwnerReferences: []metav1.OwnerReference{{UID: "123"}},
		},
		Spec: corev1.PodSpec{ServiceAccountName: defaultPodServiceAccountName},
	}
	goodAnnotatedPod, err := json.Marshal(antPod)
	require.Nil(t, err)

	inProcessPod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myAnnotatedPod",
			Namespace: mutatePodNamespace,
			Annotations: map[string]string{
				fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.EnabledAnnotation):                "true",
				fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.InProcessConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, inProcessConfigurationName),
			},
			OwnerReferences: []metav1.OwnerReference{{UID: "123"}},
		},
		Spec: corev1.PodSpec{ServiceAccountName: defaultPodServiceAccountName},
	}

	goodInProcessAnnotatedPod, err := json.Marshal(inProcessPod)
	require.Nil(t, err)

	missingAnnotationPod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myNotAnnotatedPod",
			Namespace: mutatePodNamespace,
			Annotations: map[string]string{
				fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.EnabledAnnotation): "true",
			},
			OwnerReferences: []metav1.OwnerReference{{UID: "123"}},
		},
		Spec: corev1.PodSpec{ServiceAccountName: defaultPodServiceAccountName},
	}

	missingPod, err := json.Marshal(missingAnnotationPod)
	require.Nil(t, err)

	tests := []struct {
		name     string
		mutator  *PodMutator
		req      admission.Request
		wantCode int32
		allow    bool
		setup    func(mockInjector *flagdinjectorfake.MockFlagdContainerInjector)
	}{
		{
			name: "successful request pod not annotated",
			mutator: &PodMutator{
				Client:  NewClient(false),
				decoder: decoder,
				Log:     testr.New(t),
				ready:   false,
			},
			req: admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					UID: "123",
					Object: runtime.RawExtension{
						Raw:    goodPod,
						Object: &corev1.Pod{},
					},
				},
			},
			wantCode: http.StatusOK,
			allow:    true,
		},
		{
			name: "forbidden request pod annotated but without owner",
			mutator: &PodMutator{
				Client:  NewClient(false),
				decoder: decoder,
				Log:     testr.New(t),
				ready:   false,
			},
			req: admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					UID: "123",
					Object: runtime.RawExtension{
						Raw:    badAnnotatedPod,
						Object: &corev1.Pod{},
					},
				},
			},
			wantCode: http.StatusForbidden,
			allow:    false,
		},
		{
			name: "forbidden request pod annotated with owner, but cluster role binding cannot be enabled",
			mutator: &PodMutator{
				Client: NewClient(false,
					&api.FeatureFlagSource{
						ObjectMeta: metav1.ObjectMeta{
							Name:      featureFlagSourceName,
							Namespace: mutatePodNamespace,
						},
						Spec: api.FeatureFlagSourceSpec{
							Sources: []api.Source{
								{Provider: apicommon.SyncProviderKubernetes},
							},
						},
					},
				),
				decoder: decoder,
				Log:     testr.New(t),
				ready:   false,
			},
			req: admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					UID: "123",
					Object: runtime.RawExtension{
						Raw:    goodAnnotatedPod,
						Object: &corev1.Pod{},
					},
				},
			},
			setup: func(mockInjector *flagdinjectorfake.MockFlagdContainerInjector) {
				mockInjector.EXPECT().
					EnableClusterRoleBinding(
						gomock.Any(),
						antPod.Namespace,
						antPod.Spec.ServiceAccountName,
					).Return(errors.New("error")).Times(1)
			},
			wantCode: http.StatusForbidden,
			allow:    false,
		},
		{
			name: "forbidden request pod annotated with owner, but flagd proxy is not ready",
			mutator: &PodMutator{
				Client: NewClient(false,
					&api.FeatureFlagSource{
						ObjectMeta: metav1.ObjectMeta{
							Name:      featureFlagSourceName,
							Namespace: mutatePodNamespace,
						},
						Spec: api.FeatureFlagSourceSpec{},
					},
				),
				decoder: decoder,
				Log:     testr.New(t),
				ready:   false,
			},
			req: admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					UID: "123",
					Object: runtime.RawExtension{
						Raw:    goodAnnotatedPod,
						Object: &corev1.Pod{},
					},
				},
			},
			setup: func(mockInjector *flagdinjectorfake.MockFlagdContainerInjector) {
				mockInjector.EXPECT().
					InjectFlagd(
						gomock.Any(),
						gomock.AssignableToTypeOf(&antPod.ObjectMeta),
						gomock.AssignableToTypeOf(&antPod.Spec),
						gomock.AssignableToTypeOf(&api.FeatureFlagSourceSpec{}),
					).Return(common.ErrFlagdProxyNotReady).Times(1)
			},
			wantCode: http.StatusForbidden,
		},
		{
			name: "forbidden request pod annotated with owner, but FeatureFlagSource is not available",
			mutator: &PodMutator{
				Client:  NewClient(false),
				decoder: decoder,
				Log:     testr.New(t),
				ready:   false,
			},
			req: admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					UID: "123",
					Object: runtime.RawExtension{
						Raw:    goodAnnotatedPod,
						Object: &corev1.Pod{},
					},
				},
			},
			wantCode: http.StatusNotFound,
		},
		{
			name: "happy path rpc: request pod annotated configured for env var",
			mutator: &PodMutator{
				Client: NewClient(true,
					&antPod,
					&corev1.ServiceAccount{
						ObjectMeta: metav1.ObjectMeta{
							Name:      defaultPodServiceAccountName,
							Namespace: mutatePodNamespace,
						},
					},
					&rbac.ClusterRoleBinding{
						ObjectMeta: metav1.ObjectMeta{Name: common.ClusterRoleBindingName},
						Subjects:   nil,
						RoleRef:    rbac.RoleRef{},
					},
					&api.FeatureFlagSource{
						ObjectMeta: metav1.ObjectMeta{
							Name:      featureFlagSourceName,
							Namespace: mutatePodNamespace,
						},
						Spec: api.FeatureFlagSourceSpec{},
					},
				),
				decoder: decoder,
				Log:     testr.New(t),
			},
			req: admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					UID: "123",
					Object: runtime.RawExtension{
						Raw:    goodAnnotatedPod,
						Object: &antPod,
					},
				},
			},
			setup: func(mockInjector *flagdinjectorfake.MockFlagdContainerInjector) {
				mockInjector.EXPECT().
					InjectFlagd(
						gomock.Any(),
						gomock.AssignableToTypeOf(&antPod.ObjectMeta),
						gomock.AssignableToTypeOf(&antPod.Spec),
						gomock.AssignableToTypeOf(&api.FeatureFlagSourceSpec{}),
					).Return(nil).Times(1)
			},
			allow: true,
		},
		{
			name: "happy path in-process: request pod annotated configured for env var",
			mutator: &PodMutator{
				Client: NewClient(true,
					&inProcessPod,
					&corev1.ServiceAccount{
						ObjectMeta: metav1.ObjectMeta{
							Name:      defaultPodServiceAccountName,
							Namespace: mutatePodNamespace,
						},
					},
					&rbac.ClusterRoleBinding{
						ObjectMeta: metav1.ObjectMeta{Name: common.ClusterRoleBindingName},
						Subjects:   nil,
						RoleRef:    rbac.RoleRef{},
					},
					&api.InProcessConfiguration{
						ObjectMeta: metav1.ObjectMeta{
							Name:      inProcessConfigurationName,
							Namespace: mutatePodNamespace,
						},
						Spec: api.InProcessConfigurationSpec{
							EnvVars: []corev1.EnvVar{
								{
									Name:  "env1",
									Value: "val1",
								},
								{
									Name:  "env2",
									Value: "val2",
								},
							},
						},
					},
				),
				decoder: decoder,
				Log:     testr.New(t),
			},
			req: admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					UID: "123",
					Object: runtime.RawExtension{
						Raw:    goodInProcessAnnotatedPod,
						Object: &inProcessPod,
					},
				},
			},
			setup: func(mockInjector *flagdinjectorfake.MockFlagdContainerInjector) {
				mockInjector.EXPECT().
					InjectFlagd(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).Return(nil).Times(0)
			},
			allow: true,
		},
		{
			name: "ofo enabled but annotation missing",
			mutator: &PodMutator{
				Client: NewClient(true,
					&inProcessPod,
					&corev1.ServiceAccount{
						ObjectMeta: metav1.ObjectMeta{
							Name:      defaultPodServiceAccountName,
							Namespace: mutatePodNamespace,
						},
					},
					&rbac.ClusterRoleBinding{
						ObjectMeta: metav1.ObjectMeta{Name: common.ClusterRoleBindingName},
						Subjects:   nil,
						RoleRef:    rbac.RoleRef{},
					},
				),
				decoder: decoder,
				Log:     testr.New(t),
			},
			req: admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					UID: "123",
					Object: runtime.RawExtension{
						Raw:    missingPod,
						Object: &missingAnnotationPod,
					},
				},
			},
			setup: func(mockInjector *flagdinjectorfake.MockFlagdContainerInjector) {
				mockInjector.EXPECT().
					InjectFlagd(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).Return(nil).Times(0)
			},
			wantCode: http.StatusForbidden,
			allow:    false,
		},
		{
			name: "wrong request",
			mutator: &PodMutator{
				Client:  NewClient(false),
				decoder: decoder,
				Log:     testr.New(t),
				ready:   false,
			},
			req: admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					UID: "123",
					Object: runtime.RawExtension{
						Raw:    []byte{'1'},
						Object: &corev1.ConfigMap{},
					},
				},
			},
			wantCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			mockFlagdInjector := flagdinjectorfake.NewMockFlagdContainerInjector(ctrl)

			m := tt.mutator

			if tt.setup != nil {
				tt.setup(mockFlagdInjector)
			}
			m.FlagdInjector = mockFlagdInjector

			got := m.Handle(context.TODO(), tt.req)

			if tt.wantCode != 0 && !reflect.DeepEqual(got.Result.Code, tt.wantCode) {
				t.Errorf("Handle() = %v, want %v", got.Result.Code, tt.wantCode)
			}

			require.Equal(t, tt.allow, got.Allowed)
		})
	}
}

func NewClient(withIndexes bool, objs ...client.Object) client.Client {
	utilruntime.Must(scheme.AddToScheme(scheme.Scheme))
	utilruntime.Must(api.AddToScheme(scheme.Scheme))

	annotationsSyncIndexer := func(obj client.Object) []string {
		res := obj.GetAnnotations()[fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.AllowKubernetesSyncAnnotation)]
		return []string{res}
	}

	featureflagIndexer := func(obj client.Object) []string {
		res := obj.GetAnnotations()["openfeature.dev/featureflag"]
		return []string{res}
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(objs...)

	if withIndexes {
		return fakeClient.
			WithIndex(
				&corev1.Pod{},
				"metadata.annotations.openfeature.dev/allowkubernetessync",
				annotationsSyncIndexer).
			WithIndex(
				&corev1.Pod{},
				"metadata.annotations.openfeature.dev/featureflag",
				featureflagIndexer).
			Build()
	}
	return fakeClient.Build()
}

func TestPodMutator_IsReady(t *testing.T) {

	podMutator := PodMutator{
		ready: true,
	}

	require.Nil(t, podMutator.IsReady(nil))

	podMutator.ready = false

	require.NotNil(t, podMutator.IsReady(nil))
}
