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
	"encoding/json"
	"errors"
	"fmt"
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
	ServiceProvider *FeatureFlagServiceProvider `json:"serviceProvider"`
	// +optional
	// +nullable
	SyncProvider *FeatureFlagSyncProvider `json:"syncProvider"`
	// +optional
	// +nullable
	FlagDSpec *FlagDSpec `json:"flagDSpec"`
	// FeatureFlagSpec is the json representation of the feature flag specification
	// Deprecated: use FeatureFlagSpecV2
	FeatureFlagSpec *string `json:"featureFlagSpec,omitempty"`
	// FeatureFlagSpec is the structured representation of the feature flag specification
	FeatureFlagSpecV2 *FeatureFlagSpec `json:"featureFlagSpecV2,omitempty"`
}

func (ffcs FeatureFlagConfigurationSpec) FeatureFlagSpecJSON() (string, error) {
	var ffcsJson string
	if ffcs.FeatureFlagSpecV2 != nil { // prioritise V2
		ffcsJsonB, err := json.Marshal(ffcs.FeatureFlagSpecV2)
		if err != nil {
			return "", fmt.Errorf("FeatureFlagSpecV2: %w", err)
		}

		ffcsJson = string(ffcsJsonB)
	} else if ffcs.FeatureFlagSpec != nil {
		ffcsJson = *ffcs.FeatureFlagSpec
	} else {
		return "", errors.New("FeatureFlagSpecV2 and FeatureFlagSpec are empty")
	}

	return ffcsJson, nil
}

type FlagDSpec struct {
	// +optional
	MetricsPort int32 `json:"metricsPort"`
	// +optional
	Envs []corev1.EnvVar `json:"envs"`
}

type FeatureFlagSpec struct {
	Flags map[string]FlagSpec `json:"flags"`
}

type FlagSpec struct {
	// +kubebuilder:validation:Enum=ENABLED;DISABLED
	State string `json:"state"`
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Type=object
	Variants       json.RawMessage `json:"variants"`
	DefaultVariant string          `json:"defaultVariant"`
	// +optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Type=object
	// Targeting is the json targeting rule
	Targeting json.RawMessage `json:"targeting,omitempty"`
}

type FeatureFlagSyncProvider struct {
	Name string `json:"name"`
}

func (ffsp FeatureFlagSyncProvider) IsKubernetes() bool {
	return ffsp.Name == "kubernetes"
}

type FeatureFlagServiceProvider struct {
	// +kubebuilder:validation:Enum=flagd
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

func GenerateFfConfigMap(
	name string, namespace string, references []metav1.OwnerReference, spec FeatureFlagConfigurationSpec,
) (corev1.ConfigMap, error) {
	configJson, err := spec.FeatureFlagSpecJSON()
	if err != nil {
		return corev1.ConfigMap{}, err
	}

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
			"config.json": configJson,
		},
	}, nil
}
