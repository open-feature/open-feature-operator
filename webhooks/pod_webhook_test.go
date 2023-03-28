package webhooks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-logr/logr/testr"
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha2"
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha3"
	"github.com/open-feature/open-feature-operator/controllers"
	"github.com/open-feature/open-feature-operator/pkg/utils"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"net/http"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"testing"
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
		asserts func(client client.Client)
	}{
		{
			name: "no annotated pod",
			mutator: &PodMutator{
				Client:                    NewClient(false),
				FlagDResourceRequirements: corev1.ResourceRequirements{},
				decoder:                   nil,
				Log:                       testr.New(t),
				ready:                     false,
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
								fmt.Sprintf("%s/%s", controllers.OpenFeatureAnnotationPrefix, EnabledAnnotation):                         "true",
								fmt.Sprintf("%s/%s", controllers.OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation):        fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
								fmt.Sprintf("%s/%s", controllers.OpenFeatureAnnotationPrefix, controllers.AllowKubernetesSyncAnnotation): "true",
							}},
					},
				),
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
								fmt.Sprintf("%s/%s", controllers.OpenFeatureAnnotationPrefix, EnabledAnnotation):                         "true",
								fmt.Sprintf("%s/%s", controllers.OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation):        fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
								fmt.Sprintf("%s/%s", controllers.OpenFeatureAnnotationPrefix, controllers.AllowKubernetesSyncAnnotation): "true",
							}},
					},
					&corev1.ServiceAccount{
						ObjectMeta: metav1.ObjectMeta{
							Name:      name,
							Namespace: ns,
							Annotations: map[string]string{
								fmt.Sprintf("%s/%s", controllers.OpenFeatureAnnotationPrefix, EnabledAnnotation):                         "true",
								fmt.Sprintf("%s/%s", controllers.OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation):        fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
								fmt.Sprintf("%s/%s", controllers.OpenFeatureAnnotationPrefix, controllers.AllowKubernetesSyncAnnotation): "true",
							}},
					},
					&rbac.ClusterRoleBinding{
						ObjectMeta: metav1.ObjectMeta{
							Name: "open-feature-operator-flagd-kubernetes-sync",
						},
					},
				),
			},
			wantErr: false,
			asserts: func(c client.Client) {
				crb := rbac.ClusterRoleBinding{}
				err := c.Get(context.TODO(), client.ObjectKey{Name: clusterRoleBindingName}, &crb)
				if err != nil {
					require.Fail(t, err.Error())
				}
				// after update, subjects should be 1
				require.Equal(t, 1, len(crb.Subjects))
			},
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
								fmt.Sprintf("%s/%s", controllers.OpenFeatureAnnotationPrefix, EnabledAnnotation):                         "true",
								fmt.Sprintf("%s/%s", controllers.OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation):        fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
								fmt.Sprintf("%s/%s", controllers.OpenFeatureAnnotationPrefix, controllers.AllowKubernetesSyncAnnotation): "true",
							}},
					},
					&corev1.ServiceAccount{
						ObjectMeta: metav1.ObjectMeta{
							Name:      name,
							Namespace: ns,
							Annotations: map[string]string{
								fmt.Sprintf("%s/%s", controllers.OpenFeatureAnnotationPrefix, EnabledAnnotation):                         "true",
								fmt.Sprintf("%s/%s", controllers.OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation):        fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
								fmt.Sprintf("%s/%s", controllers.OpenFeatureAnnotationPrefix, controllers.AllowKubernetesSyncAnnotation): "true",
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
			asserts: func(c client.Client) {
				crb := rbac.ClusterRoleBinding{}
				err := c.Get(context.TODO(), client.ObjectKey{Name: clusterRoleBindingName}, &crb)
				if err != nil {
					require.Fail(t, err.Error())
				}
				// after update, subjects should be 2
				require.Equal(t, 2, len(crb.Subjects))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.mutator
			if err := m.BackfillPermissions(context.TODO()); (err != nil) != tt.wantErr {
				t.Errorf("BackfillPermissions() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.asserts != nil {
				tt.asserts(m.Client)
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
				controllers.OpenFeatureAnnotationPrefix:                                                           "enabled",
				fmt.Sprintf("%s/%s", controllers.OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
			},
		},
	})
	require.Nil(t, err)
	goodAnnotatedPod, err := json.Marshal(corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "myAnnotatedPod",
			Annotations: map[string]string{
				controllers.OpenFeatureAnnotationPrefix:                                                           "enabled",
				fmt.Sprintf("%s/%s", controllers.OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation): fmt.Sprintf("%s/%s", mutatePodNamespace, featureFlagConfigurationName),
			},
			OwnerReferences: []metav1.OwnerReference{{UID: "123"}},
		},
	})
	require.Nil(t, err)

	tests := []struct {
		name     string
		mutator  *PodMutator
		req      admission.Request
		wantCode int32
	}{
		{
			name: "successful request pod not annotated",
			mutator: &PodMutator{
				Client:                    NewClient(false),
				FlagDResourceRequirements: corev1.ResourceRequirements{},
				decoder:                   decoder,
				Log:                       testr.New(t),
				ready:                     false,
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
		},
		{
			name: "forbidden request pod annotated but without owner",
			mutator: &PodMutator{
				Client:                    NewClient(false),
				FlagDResourceRequirements: corev1.ResourceRequirements{},
				decoder:                   decoder,
				Log:                       testr.New(t),
				ready:                     false,
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
			name: "forbidden request pod annotated with owner, but not registered",
			mutator: &PodMutator{
				Client:                    NewClient(false),
				FlagDResourceRequirements: corev1.ResourceRequirements{},
				decoder:                   decoder,
				Log:                       testr.New(t),
				ready:                     false,
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
			wantCode: http.StatusForbidden,
		},
		{
			name: "wrong request",
			mutator: &PodMutator{
				Client:                    NewClient(false),
				FlagDResourceRequirements: corev1.ResourceRequirements{},
				decoder:                   decoder,
				Log:                       testr.New(t),
				ready:                     false,
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
			m := tt.mutator
			got := m.Handle(context.TODO(), tt.req)
			if !reflect.DeepEqual(got.Result.Code, tt.wantCode) {
				t.Errorf("Handle() = %v, want %v", got.Result.Code, tt.wantCode)
			}
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
			annotations: map[string]string{controllers.OpenFeatureAnnotationPrefix: "enabled"},
			want:        true,
		},
		{
			name: "enabled",
			mutator: PodMutator{
				Log: testr.New(t),
			},
			annotations: map[string]string{fmt.Sprintf("%s/%s", controllers.OpenFeatureAnnotationPrefix, EnabledAnnotation): "true"},
			want:        true,
		}, {
			name: "disabled",
			mutator: PodMutator{
				Log: testr.New(t),
			},
			annotations: map[string]string{fmt.Sprintf("%s/%s", controllers.OpenFeatureAnnotationPrefix, EnabledAnnotation): "false"},
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

func TestPodMutator_createConfigMap(t *testing.T) {
	ownerUID := types.UID("123")
	tests := []struct {
		name      string
		mutator   *PodMutator
		namespace string
		confname  string
		pod       *corev1.Pod
		wantErr   error
	}{
		{
			name: "feature flag config not found",
			mutator: &PodMutator{
				Client: NewClient(false),
				Log:    testr.New(t),
			},
			namespace: "myns",
			confname:  "mypod",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{},
					},
				},
			},
			wantErr: errors.New("configuration myns/mypod not found"),
		},
		{
			name: "feature flag config found, config map created",
			mutator: &PodMutator{
				Client: NewClient(false, &v1alpha1.FeatureFlagConfiguration{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myconf",
						Namespace: "myns",
						UID:       ownerUID,
					},
				}),
				Log: testr.New(t),
			},
			namespace: "myns",
			confname:  "myconf",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.mutator
			err := controllers.CreateConfigMap(context.TODO(), m.Log, m.Client, tt.namespace, tt.confname, tt.pod.OwnerReferences)

			if tt.wantErr == nil {
				require.Nil(t, err)
				ffConfig := corev1.ConfigMap{}
				err := m.Client.Get(context.TODO(), client.ObjectKey{Name: tt.confname, Namespace: tt.namespace}, &ffConfig)
				require.Nil(t, err)
				require.Equal(t,
					map[string]string{
						"openfeature.dev/featureflagconfiguration": tt.confname,
					},
					ffConfig.Annotations)
				require.Equal(t, utils.FalseVal(), ffConfig.OwnerReferences[0].Controller)
				require.Equal(t, ownerUID, ffConfig.OwnerReferences[1].UID)

			} else {
				t.Log("checking error", err)
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
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
			got, got1 := controllers.ParseAnnotation(tt.s, tt.defaultNs)
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

func Test_podOwnerIsOwner(t *testing.T) {

	tests := []struct {
		name string
		pod  *corev1.Pod
		cm   corev1.ConfigMap
		want bool
	}{{
		name: "pod owner has same uid than the config map one",
		pod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{
						UID: "12345",
					},
				},
			},
		},
		cm: corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{
						UID: "12345",
					},
				},
			},
		},
		want: true,
	},
		{
			name: "pod and cm have different owners",
			pod:  &corev1.Pod{},
			cm: corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{
							UID: "12345",
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := controllers.SharedOwnership(tt.pod.OwnerReferences, tt.cm.OwnerReferences); got != tt.want {
				t.Errorf("podOwnerIsOwner() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setSecurityContext(t *testing.T) {
	user := int64(65532)
	group := int64(65532)
	want := &corev1.SecurityContext{
		// flagd does not require any additional capabilities, no bits set
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{
				"all",
			},
		},
		RunAsUser:  &user,
		RunAsGroup: &group,
		Privileged: utils.FalseVal(),
		// Prevents misconfiguration from allowing access to resources on host
		RunAsNonRoot: utils.TrueVal(),
		// Prevent container gaining more privileges than its parent process
		AllowPrivilegeEscalation: utils.FalseVal(),
		ReadOnlyRootFilesystem:   utils.TrueVal(),
		// SeccompProfile defines the systems calls that can be made by the container
		SeccompProfile: &corev1.SeccompProfile{
			Type: "RuntimeDefault",
		},
	}
	if got := setSecurityContext(); !reflect.DeepEqual(got, want) {
		t.Errorf("setSecurityContext() = %v, want %v", got, want)
	}

}

func NewClient(withIndexes bool, objs ...client.Object) client.Client {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme.Scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme.Scheme))
	utilruntime.Must(v1alpha2.AddToScheme(scheme.Scheme))
	utilruntime.Must(v1alpha3.AddToScheme(scheme.Scheme))

	annotationsSyncIndexer := func(obj client.Object) []string {
		res := obj.GetAnnotations()[fmt.Sprintf("%s/%s", controllers.OpenFeatureAnnotationPrefix, controllers.AllowKubernetesSyncAnnotation)]
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
			Build()
	}
	return fakeClient.Build()
}
