/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"github.com/open-feature/open-feature-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FeatureFlagConfigurationSpec defines the desired state of FeatureFlagConfiguration
type FeatureFlagConfigurationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +optional
	// +nullable
	Provider *FeatureFlagProvider `json:"provider"`
	// FeatureFlagSpec is the json representation of the feature flag
	FeatureFlagSpec string `json:"featureFlagSpec,omitempty"`
}

type FeatureFlagProvider struct {
	// +kubebuilder:validation:Enum=flagD
	Name string `json:"name"`
	// +optional
	// +nullable
	Credentials *corev1.ObjectReference `json:"credentials"`
}

// FeatureFlagConfigurationStatus defines the observed state of FeatureFlagConfiguration
type FeatureFlagConfigurationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// FeatureFlagConfiguration is the Schema for the featureflagconfigurations API
type FeatureFlagConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FeatureFlagConfigurationSpec   `json:"spec,omitempty"`
	Status FeatureFlagConfigurationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FeatureFlagConfigurationList contains a list of FeatureFlagConfiguration
type FeatureFlagConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FeatureFlagConfiguration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FeatureFlagConfiguration{}, &FeatureFlagConfigurationList{})
}

func GetFfReference(ff *FeatureFlagConfiguration) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: ff.APIVersion,
		Kind:       ff.Kind,
		Name:       ff.Name,
		UID:        ff.UID,
		Controller: utils.TrueVal(),
	}
}

func GenerateFfConfigMap(name string, namespace string, references []metav1.OwnerReference, spec FeatureFlagConfigurationSpec) corev1.ConfigMap {
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
			"config.json": spec.FeatureFlagSpec,
		},
	}
}
