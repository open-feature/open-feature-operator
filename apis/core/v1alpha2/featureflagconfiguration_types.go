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

package v1alpha2

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FeatureFlagConfigurationSpec defines the desired state of FeatureFlagConfiguration
type FeatureFlagConfigurationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ServiceProvider [DEPRECATED]: superseded by FlagSourceConfiguration
	// +optional
	// +nullable
	ServiceProvider *FeatureFlagServiceProvider `json:"serviceProvider"`
	// SyncProvider [DEPRECATED]: superseded by FlagSourceConfiguration
	// +optional
	// +nullable
	SyncProvider *FeatureFlagSyncProvider `json:"syncProvider"`
	// FlagDSpec [DEPRECATED]: superseded by FlagSourceConfiguration
	// +optional
	// +nullable
	FlagDSpec *FlagDSpec `json:"flagDSpec"`
	// FeatureFlagSpec is the structured representation of the feature flag specification
	FeatureFlagSpec FeatureFlagSpec `json:"featureFlagSpec,omitempty"`
	// Resources defines flagd sidecar resources. Default to operator sidecar-cpu-* and sidecar-ram-* flags.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources"`
}

type FlagDSpec struct {
	// +optional
	Envs []corev1.EnvVar `json:"envs"`
}

type FeatureFlagSpec struct {
	Flags map[string]FlagSpec `json:"flags"`
	// +optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Type=object
	Evaluators json.RawMessage `json:"$evaluators,omitempty"`
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
	// +optional
	// +nullable
	HttpSyncConfiguration *HttpSyncConfiguration `json:"httpSyncConfiguration"`
}

// HttpSyncConfiguration defines the desired configuration for a http sync
type HttpSyncConfiguration struct {
	// Target is the target url for flagd to poll
	Target string `json:"target"`
	// +optional
	BearerToken string `json:"bearerToken,omitempty"`
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

//+kubebuilder:resource:shortName="ff"
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
