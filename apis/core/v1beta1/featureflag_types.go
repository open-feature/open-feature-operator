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

package v1beta1

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FeatureFlagSpec defines the desired state of FeatureFlag
type FeatureFlagSpec struct {
	// FlagSpec is the structured representation of the feature flag specification
	FlagSpec FlagSpec `json:"flagSpec,omitempty"`
}

type FlagSpec struct {
	Flags map[string]Flag `json:"flags"`
	// +optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Type=object
	Evaluators json.RawMessage `json:"$evaluators,omitempty"`
}

type Flag struct {
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

// FeatureFlagStatus defines the observed state of FeatureFlag
type FeatureFlagStatus struct {
}

// TODO change to `ff` when v1alpha* is removed
//+kubebuilder:resource:shortName="ffc"
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// FeatureFlag is the Schema for the featureflags API
type FeatureFlag struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FeatureFlagSpec   `json:"spec,omitempty"`
	Status FeatureFlagStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FeatureFlagList contains a list of FeatureFlag
type FeatureFlagList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FeatureFlag `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FeatureFlag{}, &FeatureFlagList{})
}
