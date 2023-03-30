package featureflagconfiguration

import (
	"context"
	"testing"

	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	"github.com/open-feature/open-feature-operator/pkg/utils"
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
					utils.FeatureFlagConfigurationConfigMapKey(testNamespace, cmName): "spec",
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
			// set up k8s fake client
			fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(tt.ffConfig, tt.cm).Build()

			r := &FeatureFlagConfigurationReconciler{
				Client: fakeClient,
				Log:    ctrl.Log.WithName("featureflagconfiguration-controller"),
				Scheme: fakeClient.Scheme(),
			}

			if tt.cmDeleted {
				// if configMap should be deleted, we need to set ffConfig as the only OwnerRef before executing the Reconcile function
				ffConfig := &v1alpha1.FeatureFlagConfiguration{}
				err = fakeClient.Get(ctx, types.NamespacedName{Name: ffConfigName, Namespace: testNamespace}, ffConfig)
				require.Nil(t, err)

				cm := &corev1.ConfigMap{}
				err = fakeClient.Get(ctx, types.NamespacedName{Name: cmName, Namespace: testNamespace}, cm)
				require.Nil(t, err)

				cm.OwnerReferences = append(cm.OwnerReferences, ffConfig.GetReference())
				err := r.Client.Update(ctx, cm)
				require.Nil(t, err)
			}

			// reconcile
			_, err = r.Reconcile(ctx, req)
			require.Nil(t, err)

			ffConfig := &v1alpha1.FeatureFlagConfiguration{}
			err = fakeClient.Get(ctx, types.NamespacedName{Name: ffConfigName, Namespace: testNamespace}, ffConfig)
			require.Nil(t, err)

			// check that the provider name is set correctly
			require.Equal(t, tt.wantProvider, ffConfig.Spec.ServiceProvider.Name)

			cm := &corev1.ConfigMap{}
			err = fakeClient.Get(ctx, types.NamespacedName{Name: cmName, Namespace: testNamespace}, cm)

			if !tt.cmDeleted {
				// if configMap should not be deleted, check the correct values
				require.Nil(t, err)
				require.Equal(t, tt.wantCM.Data, cm.Data)
				require.Len(t, cm.OwnerReferences, len(tt.wantCM.OwnerReferences))
				require.Equal(t, tt.wantCM.OwnerReferences[0].APIVersion, cm.OwnerReferences[0].APIVersion)
				require.Equal(t, tt.wantCM.OwnerReferences[0].Name, cm.OwnerReferences[0].Name)
				require.Equal(t, tt.wantCM.OwnerReferences[0].Kind, cm.OwnerReferences[0].Kind)
			} else {
				// if configMap should be deleted, we expect error
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
