package featureflagconfiguration

import (
	"context"
	"testing"

	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestFlagSourceConfigurationReconciler_Reconcile(t *testing.T) {
	const (
		testNamespace = "test-namespace"
		ffConfigName  = "test-config"
		cmName        = "test-cm"
	)

	tests := []struct {
		name         string
		ffConfig     *v1alpha1.FeatureFlagConfiguration
		cm           *corev1.ConfigMap
		wantProvider string
		wantCM       *corev1.ConfigMap
		cmDeleted    bool
	}{
		{
			name: "no provider set + no owner set -> ffconfig and cm will be updated",
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmName,
					Namespace: testNamespace,
					Annotations: map[string]string{
						"openfeature.dev/featureflagconfiguration": ffConfigName,
					},
				},
			},
			ffConfig:     createTestFFConfig(ffConfigName, testNamespace, cmName, ""),
			wantProvider: "flagd",
			wantCM: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmName,
					Namespace: testNamespace,
					Annotations: map[string]string{
						"openfeature.dev/featureflagconfiguration": ffConfigName,
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "core.openfeature.dev/v1alpha1",
							Kind:       "FeatureFlagConfiguration",
							Name:       ffConfigName,
						},
					},
				},
				Data: map[string]string{
					v1alpha1.FeatureFlagConfigurationConfigMapKey(testNamespace, cmName): "spec",
				},
			},
			cmDeleted: false,
		},
		{
			name: "one owner ref set -> cm will be deleted",
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmName,
					Namespace: testNamespace,
					Annotations: map[string]string{
						"openfeature.dev/featureflagconfiguration": ffConfigName,
					},
				},
			},
			ffConfig:     createTestFFConfig(ffConfigName, testNamespace, cmName, ""),
			wantProvider: "flagd",
			wantCM: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmName,
					Namespace: testNamespace,
					Annotations: map[string]string{
						"openfeature.dev/featureflagconfiguration": ffConfigName,
					},
				},
			},
			cmDeleted: true,
		},
	}

	err := v1alpha1.AddToScheme(scheme.Scheme)
	require.Nil(t, err)
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: testNamespace,
			Name:      ffConfigName,
		},
	}

	ctx := context.TODO()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(tt.ffConfig, tt.cm).Build()

			r := &FeatureFlagConfigurationReconciler{
				Client: fakeClient,
				Log:    ctrl.Log.WithName("featureflagconfiguration-controller"),
				Scheme: fakeClient.Scheme(),
			}

			if tt.cmDeleted {
				ffConfig2 := &v1alpha1.FeatureFlagConfiguration{}
				err = fakeClient.Get(ctx, types.NamespacedName{Name: ffConfigName, Namespace: testNamespace}, ffConfig2)
				require.Nil(t, err)

				cm2 := &corev1.ConfigMap{}
				err = fakeClient.Get(ctx, types.NamespacedName{Name: cmName, Namespace: testNamespace}, cm2)
				require.Nil(t, err)

				cm2.OwnerReferences = append(cm2.OwnerReferences, v1alpha1.GetFfReference(ffConfig2))
				err := r.Client.Update(ctx, cm2)
				require.Nil(t, err)
			}

			_, err = r.Reconcile(ctx, req)
			require.Nil(t, err)

			ffConfig2 := &v1alpha1.FeatureFlagConfiguration{}
			err = fakeClient.Get(ctx, types.NamespacedName{Name: ffConfigName, Namespace: testNamespace}, ffConfig2)
			require.Nil(t, err)

			require.Equal(t, tt.wantProvider, ffConfig2.Spec.ServiceProvider.Name)

			cm2 := &corev1.ConfigMap{}
			err = fakeClient.Get(ctx, types.NamespacedName{Name: cmName, Namespace: testNamespace}, cm2)

			if !tt.cmDeleted {
				require.Nil(t, err)
				require.Equal(t, tt.wantCM.Data, cm2.Data)
				require.Len(t, cm2.OwnerReferences, len(tt.wantCM.OwnerReferences))
				require.Equal(t, tt.wantCM.OwnerReferences[0].APIVersion, cm2.OwnerReferences[0].APIVersion)
				require.Equal(t, tt.wantCM.OwnerReferences[0].Name, cm2.OwnerReferences[0].Name)
				require.Equal(t, tt.wantCM.OwnerReferences[0].Kind, cm2.OwnerReferences[0].Kind)
			} else {
				require.NotNil(t, err)
			}
		})
	}
}

func createTestFFConfig(ffConfigName string, testNamespace string, cmName string, provider string) *v1alpha1.FeatureFlagConfiguration {
	fsConfig := &v1alpha1.FeatureFlagConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ffConfigName,
			Namespace: testNamespace,
		},
		Spec: v1alpha1.FeatureFlagConfigurationSpec{
			ServiceProvider: &v1alpha1.FeatureFlagServiceProvider{
				Name: provider,
			},
			FeatureFlagSpec: "spec",
		},
	}

	return fsConfig
}
