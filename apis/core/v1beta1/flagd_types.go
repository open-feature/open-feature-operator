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
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FlagdSpec defines the desired state of Flagd
type FlagdSpec struct {
	// Replicas defines the number of replicas to create for the service.
	// Default: 1
	// +optional
	// +kubebuilder:default=1
	Replicas *int32 `json:"replicas,omitempty"`

	// ServiceType represents the type of Service to create.
	// Must be one of: ClusterIP, NodePort, LoadBalancer, and ExternalName.
	// Default: ClusterIP
	// +optional
	// +kubebuilder:default=ClusterIP
	// +kubebuilder:validation:Enum:=ClusterIP;NodePort;LoadBalancer;ExternalName
	ServiceType v1.ServiceType `json:"serviceType,omitempty"`

	// ServiceAccountName the service account name for the flagd deployment
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// FeatureFlagSource references to a FeatureFlagSource from which the created flagd instance retrieves
	// the feature flag configurations
	FeatureFlagSource string `json:"featureFlagSource"`

	// Ingress
	// +optional
	Ingress IngressSpec `json:"ingress"`
}

// IngressSpec defines the options to be used when deploying the ingress for flagd
type IngressSpec struct {
	// Enabled enables/disables the ingress for flagd
	Enabled bool `json:"enabled,omitempty"`

	// Annotations the annotations to be added to the ingress
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Hosts list of hosts to be added to the ingress.
	// Empty string corresponds to rule with no host.
	Hosts []string `json:"hosts"`

	// TLS configuration for the ingress
	TLS []networkingv1.IngressTLS `json:"tls,omitempty"`

	// IngressClassName defines the name if the ingress class to be used for flagd
	// +optional
	IngressClassName *string `json:"ingressClassName,omitempty"`

	// PathType is the path type to be used for the ingress rules
	// +optional
	PathType networkingv1.PathType `json:"pathType,omitempty"`

	// FlagdPath is the path to be used for accessing the flagd flag evaluation API
	// Default: /flagd.evaluation.v1.Service
	// +optional
	FlagdPath string `json:"flagdPath,omitempty"`

	// OFREPPath is the path to be used for accessing the OFREP API
	// Default: /ofrep
	// +optional
	OFREPPath string `json:"ofrepPath,omitempty"`

	// SyncPath is the path to be used for accessing the sync API
	// Default: /flagd.sync.v1.Service
	// +optional
	SyncPath string `json:"syncPath,omitempty"`
}

// FlagdStatus defines the observed state of Flagd
type FlagdStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Flagd is the Schema for the flagds API
type Flagd struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FlagdSpec   `json:"spec,omitempty"`
	Status FlagdStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FlagdList contains a list of Flagd
type FlagdList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Flagd `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Flagd{}, &FlagdList{})
}
