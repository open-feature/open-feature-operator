package webhooks

import (
	"context"
	"fmt"
	"testing"

	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	corev1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestFeatureFlagConfigurationWebhook_validateFlagSourceConfiguration(t *testing.T) {
	const credentialsName = "credentials-name"
	const featureFlagConfigurationName = "test-feature-flag-configuration"

	tests := []struct {
		name   string
		obj    corev1alpha1.FeatureFlagConfiguration
		secret *corev1.Secret
		out    error
	}{
		{
			name: "valid without ServiceProvider",
			obj: corev1alpha1.FeatureFlagConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name:      featureFlagConfigurationName,
					Namespace: featureFlagConfigurationNamespace,
				},
				Spec: corev1alpha1.FeatureFlagConfigurationSpec{
					FeatureFlagSpec: featureFlagSpec,
				},
			},
			out: nil,
		},
		{
			name: "invalid json",
			obj: corev1alpha1.FeatureFlagConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name:      featureFlagConfigurationName,
					Namespace: featureFlagConfigurationNamespace,
				},
				Spec: corev1alpha1.FeatureFlagConfigurationSpec{
					FeatureFlagSpec: `{"invalid":json}`,
				},
			},
			out: fmt.Errorf("FeatureFlagSpec is not valid JSON: {\"invalid\":json}"),
		},
		{
			name: "invalid schema",
			obj: corev1alpha1.FeatureFlagConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name:      featureFlagConfigurationName,
					Namespace: featureFlagConfigurationNamespace,
				},
				Spec: corev1alpha1.FeatureFlagConfigurationSpec{
					FeatureFlagSpec: `{
						"flags":{
							"foo": {}
						}
					}`,
				},
			},
			out: fmt.Errorf("FeatureFlagSpec is not valid JSON: - flags.foo: Must validate one and only one schema (oneOf)\n- flags.foo: state is required\n- flags.foo: defaultVariant is required\n- flags.foo: Must validate all the schemas (allOf)\n"),
		},
		{
			name: "valid with ServiceProvider",
			obj: corev1alpha1.FeatureFlagConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name:      featureFlagConfigurationName,
					Namespace: featureFlagConfigurationNamespace,
				},
				Spec: corev1alpha1.FeatureFlagConfigurationSpec{
					FeatureFlagSpec: featureFlagSpec,
					ServiceProvider: &corev1alpha1.FeatureFlagServiceProvider{
						Name: "flagd",
						Credentials: &corev1.ObjectReference{
							Name:      credentialsName,
							Namespace: featureFlagConfigurationNamespace,
						},
					},
				},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      credentialsName,
					Namespace: featureFlagConfigurationNamespace,
				},
			},
			out: nil,
		},
		{
			name: "non-existing secret in ServiceProvider",
			obj: corev1alpha1.FeatureFlagConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name:      featureFlagConfigurationName,
					Namespace: featureFlagConfigurationNamespace,
				},
				Spec: corev1alpha1.FeatureFlagConfigurationSpec{
					FeatureFlagSpec: featureFlagSpec,
					ServiceProvider: &corev1alpha1.FeatureFlagServiceProvider{
						Name: "flagd",
						Credentials: &corev1.ObjectReference{
							Name:      credentialsName,
							Namespace: featureFlagConfigurationNamespace,
						},
					},
				},
			},
			out: fmt.Errorf("credentials secret not found"),
		},
	}

	err := v1alpha1.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := FeatureFlagConfigurationValidator{
				Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).Build(),
				Log:    ctrl.Log.WithName("webhook"),
			}

			if tt.secret != nil {
				err := validator.Client.Create(context.TODO(), tt.secret)
				require.Nil(t, err)
			}

			out := validator.validateFlagSourceConfiguration(context.TODO(), tt.obj)
			require.Equal(t, tt.out, out)
		})

	}
}
