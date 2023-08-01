package webhooks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/open-feature/open-feature-operator/controllers/common/constant"
	commonmock "github.com/open-feature/open-feature-operator/controllers/common/mock"
	"net/http"
	"reflect"
	"testing"

	"github.com/go-logr/logr/testr"
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha2"
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha3"
	"github.com/open-feature/open-feature-operator/pkg/utils"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestOpenFeatureEnabledAnnotationIndex(t *testing.T) {

	tests := []struct {
		name string
		o    client.Object
		want []string
	}{
		{
			name: "no annotations",
			o:    &corev1.Pod{},
			want: []string{"false"},
		}, {
			name: "annotated wrong",
			o:    &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{"test/ann": "nope", "openfeature.dev/allowkubernetessync": "false"}}},
			want: []string{"false"},
		}, {
			name: "annotated with enabled index",
			o:    &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{"openfeature.dev/allowkubernetessync": "true"}}},
			want: []string{"true"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OpenFeatureEnabledAnnotationIndex(tt.o); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OpenFeatureEnabledAnnotationIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
		setup   func(injector *commonmock.MockIFlagdContainerInjector)
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
								fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, EnabledAnnotation):                  "true",
								fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
								fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, AllowKubernetesSyncAnnotation):      "true",
							}},
					},
				),
			},
			setup: func(injector *commonmock.MockIFlagdContainerInjector) {
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
								fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, EnabledAnnotation):                  "true",
								fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
								fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, AllowKubernetesSyncAnnotation):      "true",
							}},
					},
					&corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      pod + "-2",
							Namespace: ns,
							Annotations: map[string]string{
								fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, EnabledAnnotation):                  "true",
								fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
								fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, AllowKubernetesSyncAnnotation):      "true",
							}},
					},
				),
			},
			setup: func(injector *commonmock.MockIFlagdContainerInjector) {
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
								fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, EnabledAnnotation):                  "true",
								fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
								fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, AllowKubernetesSyncAnnotation):      "true",
							}},
						Spec: corev1.PodSpec{ServiceAccountName: "my-service-account"},
					},
					&corev1.ServiceAccount{
						ObjectMeta: metav1.ObjectMeta{
							Name:      name,
							Namespace: ns,
							Annotations: map[string]string{
								fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, EnabledAnnotation):                  "true",
								fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
								fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, AllowKubernetesSyncAnnotation):      "true",
							}},
					},
					&rbac.ClusterRoleBinding{
						ObjectMeta: metav1.ObjectMeta{
							Name: "open-feature-operator-flagd-kubernetes-sync",
						},
					},
				),
			},
			setup: func(injector *commonmock.MockIFlagdContainerInjector) {
				injector.EXPECT().EnableClusterRoleBinding(context.Background(), ns, "my-service-account").Times(1)
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
								fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, EnabledAnnotation):                  "true",
								fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
								fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, AllowKubernetesSyncAnnotation):      "true",
							}},
					},
					&corev1.ServiceAccount{
						ObjectMeta: metav1.ObjectMeta{
							Name:      name,
							Namespace: ns,
							Annotations: map[string]string{
								fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, EnabledAnnotation):                  "true",
								fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
								fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, AllowKubernetesSyncAnnotation):      "true",
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
			setup: func(injector *commonmock.MockIFlagdContainerInjector) {
				injector.EXPECT().EnableClusterRoleBinding(context.Background(), ns, "").Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			mockInjector := commonmock.NewMockIFlagdContainerInjector(ctrl)

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
				OpenFeatureAnnotationPrefix: "enabled",
				fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
			},
		},
	})
	require.Nil(t, err)

	antPod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myAnnotatedPod",
			Namespace: mutatePodNamespace,
			Annotations: map[string]string{
				OpenFeatureAnnotationPrefix: "enabled",
				fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
			},
			OwnerReferences: []metav1.OwnerReference{{UID: "123"}},
		},
		Spec: corev1.PodSpec{ServiceAccountName: defaultPodServiceAccountName},
	}
	goodAnnotatedPod, err := json.Marshal(antPod)
	require.Nil(t, err)

	tests := []struct {
		name     string
		mutator  *PodMutator
		req      admission.Request
		wantCode int32
		allow    bool
		setup    func(mockInjector *commonmock.MockIFlagdContainerInjector)
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
		},
		{
			name: "forbidden request pod annotated with owner, but cluster role binding cannot be enabled",
			mutator: &PodMutator{
				Client: NewClient(false,
					&v1alpha1.FeatureFlagConfiguration{
						ObjectMeta: metav1.ObjectMeta{
							Name:      featureFlagConfigurationName,
							Namespace: mutatePodNamespace,
						},
						Spec: v1alpha1.FeatureFlagConfigurationSpec{
							FlagDSpec: &v1alpha1.FlagDSpec{Envs: []corev1.EnvVar{
								{Name: "LOG_LEVEL", Value: "dev"},
							}},
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
			setup: func(mockInjector *commonmock.MockIFlagdContainerInjector) {
				mockInjector.EXPECT().
					EnableClusterRoleBinding(
						gomock.Any(),
						antPod.Namespace,
						antPod.Spec.ServiceAccountName,
					).Return(errors.New("error")).Times(1)
			},
			wantCode: http.StatusForbidden,
		},
		{
			name: "forbidden request pod annotated with owner, but flagd proxy is not ready",
			mutator: &PodMutator{
				Client: NewClient(false,
					&v1alpha1.FeatureFlagConfiguration{
						ObjectMeta: metav1.ObjectMeta{
							Name:      featureFlagConfigurationName,
							Namespace: mutatePodNamespace,
						},
						Spec: v1alpha1.FeatureFlagConfigurationSpec{
							FlagDSpec: &v1alpha1.FlagDSpec{Envs: []corev1.EnvVar{
								{Name: "LOG_LEVEL", Value: "dev"},
							}},
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
			setup: func(mockInjector *commonmock.MockIFlagdContainerInjector) {
				mockInjector.EXPECT().
					EnableClusterRoleBinding(
						gomock.Any(),
						antPod.Namespace,
						antPod.Spec.ServiceAccountName,
					).Return(nil).Times(1)

				mockInjector.EXPECT().
					InjectFlagd(
						gomock.Any(),
						gomock.AssignableToTypeOf(&antPod.ObjectMeta),
						gomock.AssignableToTypeOf(&antPod.Spec),
						gomock.AssignableToTypeOf(&v1alpha1.FlagSourceConfigurationSpec{}),
					).Return(constant.ErrFlagdProxyNotReady).Times(1)
			},
			wantCode: http.StatusForbidden,
		},
		{
			name: "forbidden request pod annotated with owner, but feature flag configuration is not available",
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
			setup: func(mockInjector *commonmock.MockIFlagdContainerInjector) {
				mockInjector.EXPECT().
					EnableClusterRoleBinding(
						gomock.Any(),
						antPod.Namespace,
						antPod.Spec.ServiceAccountName,
					).Return(nil).Times(1)

				mockInjector.EXPECT().
					InjectFlagd(
						gomock.Any(),
						gomock.AssignableToTypeOf(&antPod.ObjectMeta),
						gomock.AssignableToTypeOf(&antPod.Spec),
						gomock.AssignableToTypeOf(&v1alpha1.FlagSourceConfigurationSpec{}),
					).Return(constant.ErrFlagdProxyNotReady).Times(1)
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name: "happy path: request pod annotated configured for env var",
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
						ObjectMeta: metav1.ObjectMeta{Name: clusterRoleBindingName},
						Subjects:   nil,
						RoleRef:    rbac.RoleRef{},
					},
					&v1alpha1.FeatureFlagConfiguration{
						ObjectMeta: metav1.ObjectMeta{
							Name:      featureFlagConfigurationName,
							Namespace: mutatePodNamespace,
						},
						Spec: v1alpha1.FeatureFlagConfigurationSpec{
							FlagDSpec: &v1alpha1.FlagDSpec{Envs: []corev1.EnvVar{
								{Name: "LOG_LEVEL", Value: "dev"},
							}},
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
						Raw:    goodAnnotatedPod,
						Object: &antPod,
					},
				},
			},
			setup: func(mockInjector *commonmock.MockIFlagdContainerInjector) {
				mockInjector.EXPECT().
					EnableClusterRoleBinding(
						gomock.Any(),
						antPod.Namespace,
						antPod.Spec.ServiceAccountName,
					).Return(nil).Times(1)

				mockInjector.EXPECT().
					InjectFlagd(
						gomock.Any(),
						gomock.AssignableToTypeOf(&antPod.ObjectMeta),
						gomock.AssignableToTypeOf(&antPod.Spec),
						gomock.AssignableToTypeOf(&v1alpha1.FlagSourceConfigurationSpec{}),
					).Return(nil).Times(1)
			},
			allow: true,
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
			mockFlagdInjector := commonmock.NewMockIFlagdContainerInjector(ctrl)

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

func TestPodMutator_checkOFEnabled(t *testing.T) {

	tests := []struct {
		name        string
		mutator     PodMutator
		annotations map[string]string
		want        bool
	}{
		{
			name: "deprecated enabled",
			mutator: PodMutator{
				Log: testr.New(t),
			},
			annotations: map[string]string{OpenFeatureAnnotationPrefix: "enabled"},
			want:        true,
		},
		{
			name: "enabled",
			mutator: PodMutator{
				Log: testr.New(t),
			},
			annotations: map[string]string{fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, EnabledAnnotation): "true"},
			want:        true,
		}, {
			name: "disabled",
			mutator: PodMutator{
				Log: testr.New(t),
			},
			annotations: map[string]string{fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, EnabledAnnotation): "false"},
			want:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &tt.mutator
			if got := m.checkOFEnabled(tt.annotations); got != tt.want {
				t.Errorf("checkOFEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseAnnotation(t *testing.T) {
	tests := []struct {
		name      string
		s         string
		defaultNs string
		wantNs    string
		want      string
	}{
		{
			name:      "no namespace",
			s:         "test",
			defaultNs: "ofo",
			wantNs:    "ofo",
			want:      "test",
		},
		{
			name:      "namespace",
			s:         "myns/test",
			defaultNs: "ofo",
			wantNs:    "myns",
			want:      "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := utils.ParseAnnotation(tt.s, tt.defaultNs)
			if got != tt.wantNs {
				t.Errorf("parseAnnotation() got = %v, want %v", got, tt.wantNs)
			}
			if got1 != tt.want {
				t.Errorf("parseAnnotation() got1 = %v, want %v", got1, tt.want)
			}
		})
	}
}

func Test_parseList(t *testing.T) {

	tests := []struct {
		name string
		s    string
		want []string
	}{
		{
			name: "empty string",
			s:    "",
			want: []string{},
		}, {
			name: "nice list with spaces",
			s:    "annotation1, annotation2,    annotation4 , annotation3,",
			want: []string{"annotation1", "annotation2", "annotation4", "annotation3"},
		}, {
			name: "list with no spaces",
			s:    "annotation1, annotation2,annotation4, annotation3",
			want: []string{"annotation1", "annotation2", "annotation4", "annotation3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseList(tt.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func NewClient(withIndexes bool, objs ...client.Object) client.Client {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme.Scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme.Scheme))
	utilruntime.Must(v1alpha2.AddToScheme(scheme.Scheme))
	utilruntime.Must(v1alpha3.AddToScheme(scheme.Scheme))

	annotationsSyncIndexer := func(obj client.Object) []string {
		res := obj.GetAnnotations()[fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, AllowKubernetesSyncAnnotation)]
		return []string{res}
	}

	featureflagIndexer := func(obj client.Object) []string {
		res := obj.GetAnnotations()["openfeature.dev/featureflagconfiguration"]
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
				"metadata.annotations.openfeature.dev/featureflagconfiguration",
				featureflagIndexer).
			Build()
	}
	return fakeClient.Build()
}
