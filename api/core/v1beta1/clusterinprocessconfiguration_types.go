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

// ClusterInProcessConfigurationSpec defines the desired state of ClusterInProcessConfiguration.
// It is cluster-scoped and structurally identical to InProcessConfigurationSpec.
type ClusterInProcessConfigurationSpec struct {
	InProcessConfigurationSpec `json:",inline"`
}

// ClusterInProcessConfigurationStatus defines the observed state of ClusterInProcessConfiguration
type ClusterInProcessConfigurationStatus struct{}

//+kubebuilder:resource:scope=Cluster,shortName="cipc"
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ClusterInProcessConfiguration is the cluster-scoped Schema for the ClusterInProcessConfigurations API.
// Intended for cluster admins to publish shared in-process flagd configuration that can be referenced
// from any namespace via the openfeature.dev/clusterinprocessconfiguration annotation.
type ClusterInProcessConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterInProcessConfigurationSpec   `json:"spec,omitempty"`
	Status ClusterInProcessConfigurationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterInProcessConfigurationList contains a list of ClusterInProcessConfiguration
type ClusterInProcessConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterInProcessConfiguration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterInProcessConfiguration{}, &ClusterInProcessConfigurationList{})
}
