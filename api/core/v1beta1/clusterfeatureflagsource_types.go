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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterFeatureFlagSourceSpec defines the desired state of ClusterFeatureFlagSource.
// It is cluster-scoped and its Sources[].source values MUST be qualified with an
// explicit namespace in the form <namespace>/<name>; bare names are rejected by
// the schema because a cluster-scoped CR has no ambient namespace to resolve them.
type ClusterFeatureFlagSourceSpec struct {
	FeatureFlagSourceSpec `json:",inline"`
}

// ClusterFeatureFlagSourceStatus defines the observed state of ClusterFeatureFlagSource
type ClusterFeatureFlagSourceStatus struct{}

//+kubebuilder:resource:scope=Cluster,shortName="cffs"
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:validation:XValidation:rule="self.spec.sources.all(s, !has(s.provider) || !(s.provider in ['kubernetes', 'file', 'flagd-proxy']) || s.source.matches('^[^/]+/[^/]+$'))",message="kubernetes, file and flagd-proxy sources must use the form <namespace>/<name>"

// ClusterFeatureFlagSource is the cluster-scoped Schema for the ClusterFeatureFlagSources API.
// Intended for cluster admins to publish shared flagd sidecar configurations that can be
// referenced from any namespace via the openfeature.dev/clusterfeatureflagsource annotation.
type ClusterFeatureFlagSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterFeatureFlagSourceSpec   `json:"spec,omitempty"`
	Status ClusterFeatureFlagSourceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterFeatureFlagSourceList contains a list of ClusterFeatureFlagSource
type ClusterFeatureFlagSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterFeatureFlagSource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterFeatureFlagSource{}, &ClusterFeatureFlagSourceList{})
}
