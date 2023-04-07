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

package v1alpha3

import (
	appsV1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FlagdSpec defines the desired state of Flagd
type FlagdSpec struct {
	// Important: Run "make" to regenerate code after modifying this file
	// +optional
	DeploymentSpec appsV1.DeploymentSpec `json:"deploymentSpec"`
	// +optional
	ServiceAccountName string `json:"serviceAccountName"`
	// +optional
	Service                 string `json:"service"`
	FlagSourceConfiguration string `json:"flagSourceConfiguration"`
}

// FlagdStatus defines the observed state of Flagd
type FlagdStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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