package utils

<<<<<<< HEAD
import (
	configv1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

=======
>>>>>>> fc9391c (reworked to use FF CR as the webhook trigger)
func TrueVal() *bool {
	b := true
	return &b
}

func FalseVal() *bool {
	b := false
	return &b
}

func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func GenerateFfConfigMap(name string, namespace string, references []metav1.OwnerReference, spec configv1alpha1.FeatureFlagConfigurationSpec) corev1.ConfigMap {
	return corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				"openfeature.dev/featureflagconfiguration": name,
			},
			OwnerReferences: references,
		},
		Data: map[string]string{
			"config.yaml": spec.FeatureFlagSpec,
		},
	}
}
