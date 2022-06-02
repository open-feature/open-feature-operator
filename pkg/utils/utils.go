package utils

import (
	configv1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

func GetFfReference(ff *configv1alpha1.FeatureFlagConfiguration) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: ff.APIVersion,
		Kind:       ff.Kind,
		Name:       ff.Name,
		UID:        ff.UID,
		Controller: TrueVal(),
	}
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
