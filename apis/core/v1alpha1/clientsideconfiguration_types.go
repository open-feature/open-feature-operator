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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClientSideConfigurationSpec defines the desired state of ClientSideConfiguration
type ClientSideConfigurationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ServiceAccountName      string                           `json:"serviceAccountName"`
	GatewayName             string                           `json:"gatewayName"`
	HTTPRouteHostname       string                           `json:"httpRouteHostname"`
	HTTPRouteName           string                           `json:"httpRouteName"`
	HTTPRouteMatches        []gatewayv1beta1.HTTPRouteMatch  `json:"httpRouteMatches,omitempty"`
	HTTPRouteFilters        []gatewayv1beta1.HTTPRouteFilter `json:"httpRouteFilters,omitempty"`
	GatewayListenerPort     int32                            `json:"gatewayListenerPort"`
	FlagSourceConfiguration string                           `json:"flagSourceConfiguration"`
	GatewayClassName        string                           `json:"gatewayClassName"`
	CorsAllowOrigin         string                           `json:"corsAllowOrigin"`
}

// ClientSideConfigurationStatus defines the observed state of ClientSideConfiguration
type ClientSideConfigurationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ClientSideConfiguration is the Schema for the clientsideconfigurations API
type ClientSideConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClientSideConfigurationSpec   `json:"spec,omitempty"`
	Status ClientSideConfigurationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClientSideConfigurationList contains a list of ClientSideConfiguration
type ClientSideConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClientSideConfiguration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClientSideConfiguration{}, &ClientSideConfigurationList{})
}
