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
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	FlagdMetricPortEnvVar string = "FLAGD_METRICS_PORT"
	FlagdPortEnvVar       string = "FLAGD_PORT"
	FlagdSocketPathEnvVar string = "FLAGD_SOCKET_PATH"
	FlagdEvaluatorEnvVar  string = "FLAGD_EVALUATOR"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FlagdConfigurationSpec defines the desired state of FlagdConfiguration
type FlagdConfigurationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// MetricsPort defines the port to serve metrics on, defaults to 8013
	// +optional
	MetricsPort int32 `json:"metricsPort"`

	// Port defines the port to listen on, defaults to 8014
	// +optional
	Port int32 `json:"port"`

	// SocketPath defines the unix socket path to listen on
	// +optional
	SocketPath string `json:"socketPath"`

	//SyncProviderArgs are string arguments passed to all sync providers, defined as key values separated by =
	// +optional
	SyncProviderArgs []string `json:"syncProviderArgs"`

	// Evaluator sets an evaluator, defaults to 'json'
	// +optional
	Evaluator string `json:"evaluator"`

	// Image allows for the flagd image to be overridden, defaults to 'ghcr.io/open-feature/flagd'
	// +optional
	Image string `json:"image"`

	// Tag to be appended to the flagd image, defaults to 'main'
	// +optional
	Tag string `json:"tag"`
}

func NewFlagdConfigurationSpec() *FlagdConfigurationSpec {
	var tag = "main"
	if os.Getenv("FLAGD_VERSION") != "" {
		tag = os.Getenv("FLAGD_VERSION")
	}
	return &FlagdConfigurationSpec{
		MetricsPort:      8014,
		Port:             8013,
		SocketPath:       "",
		SyncProviderArgs: []string{},
		Evaluator:        "json",
		Image:            "ghcr.io/open-feature/flagd",
		Tag:              tag,
	}
}

func (fc *FlagdConfigurationSpec) Merge(new *FlagdConfigurationSpec) {
	if new.MetricsPort != 0 {
		fc.MetricsPort = new.MetricsPort
	}
	if new.Port != 0 {
		fc.Port = new.Port
	}
	if new.SocketPath != "" {
		fc.SocketPath = new.SocketPath
	}
	if new.Evaluator != "" {
		fc.Evaluator = new.Evaluator
	}
	if new.Image != "" {
		fc.Image = new.Image
	}
	if new.Tag != "" {
		fc.Tag = new.Tag
	}
	if new.SyncProviderArgs != nil {
		for k, v := range new.SyncProviderArgs {
			fc.SyncProviderArgs[k] = v
		}
	}
}

func (fc *FlagdConfigurationSpec) ToEnvVars() []corev1.EnvVar {
	envs := []corev1.EnvVar{
		{
			Name:  FlagdMetricPortEnvVar,
			Value: fmt.Sprintf("%d", fc.MetricsPort),
		},
		{
			Name:  FlagdPortEnvVar,
			Value: fmt.Sprintf("%d", fc.Port),
		},
		{
			Name:  FlagdEvaluatorEnvVar,
			Value: fc.Evaluator,
		},
	}
	if fc.SocketPath != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  FlagdSocketPathEnvVar,
			Value: fc.SocketPath,
		})
	}
	return envs
}

// FlagdConfigurationStatus defines the observed state of FlagdConfiguration
type FlagdConfigurationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// FlagdConfiguration is the Schema for the flagdconfigurations API
type FlagdConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FlagdConfigurationSpec   `json:"spec,omitempty"`
	Status FlagdConfigurationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FlagdConfigurationList contains a list of FlagdConfiguration
type FlagdConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FlagdConfiguration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FlagdConfiguration{}, &FlagdConfigurationList{})
}
