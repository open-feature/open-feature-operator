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
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SidecarEnvVarPrefix        string = "SIDECAR_ENV_VAR_PREFIX"
	SidecarMetricPortEnvVar    string = "METRICS_PORT"
	SidecarPortEnvVar          string = "PORT"
	SidecarSocketPathEnvVar    string = "SOCKET_PATH"
	SidecarEvaluatorEnvVar     string = "EVALUATOR"
	SidecarImageEnvVar         string = "IMAGE"
	SidecarVersionEnvVar       string = "VERSION"
	SidecarProviderArgsEnvVar  string = "PROVIDER_ARGS"
	defaultMetricPort          int32  = 8014
	defaultPort                int32  = 8013
	defaultSocketPath          string = ""
	defaultEvaluator           string = "json"
	defaultImage               string = "ghcr.io/open-feature/flagd"
	defaultTag                 string = "main"
	defaultSidecarEnvVarPrefix string = "FLAGD"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FlagSourceConfigurationSpec defines the desired state of FlagSourceConfiguration
type FlagSourceConfigurationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// MetricsPort defines the port to serve metrics on, defaults to 8014
	// +optional
	MetricsPort int32 `json:"metricsPort"`

	// Port defines the port to listen on, defaults to 8013
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

	// Image allows for the sidecar image to be overridden, defaults to 'ghcr.io/open-feature/flagd'
	// +optional
	Image string `json:"image"`

	// Tag to be appended to the sidecar image, defaults to 'main'
	// +optional
	Tag string `json:"tag"`
}

func NewFlagSourceConfigurationSpec() (*FlagSourceConfigurationSpec, error) {
	var parseError error = nil
	fsc := &FlagSourceConfigurationSpec{
		MetricsPort:      defaultMetricPort,
		Port:             defaultPort,
		SocketPath:       defaultSocketPath,
		SyncProviderArgs: []string{},
		Evaluator:        defaultEvaluator,
		Image:            defaultImage,
		Tag:              defaultTag,
	}

	if metricsPort := os.Getenv(SidecarMetricPortEnvVar); metricsPort != "" {
		metricsPortI, err := strconv.Atoi(metricsPort)
		if err != nil {
			parseError = err
		} else {
			fsc.MetricsPort = int32(metricsPortI)
		}
	}

	if port := os.Getenv(SidecarPortEnvVar); port != "" {
		portI, err := strconv.Atoi(port)
		if err != nil {
			parseError = err
		} else {
			fsc.Port = int32(portI)
		}
	}

	if socketPath := os.Getenv(SidecarSocketPathEnvVar); socketPath != "" {
		fsc.SocketPath = socketPath
	}

	if evaluator := os.Getenv(SidecarEvaluatorEnvVar); evaluator != "" {
		fsc.Evaluator = evaluator
	}

	if image := os.Getenv(SidecarImageEnvVar); image != "" {
		fsc.Image = image
	}

	if tag := os.Getenv(SidecarVersionEnvVar); tag != "" {
		fsc.Tag = tag
	}

	if syncProviderArgs := os.Getenv(SidecarProviderArgsEnvVar); syncProviderArgs != "" {
		fsc.SyncProviderArgs = strings.Split(syncProviderArgs, ",") // todo: add documentation for this
	}

	return fsc, parseError
}

func (fc *FlagSourceConfigurationSpec) Merge(new *FlagSourceConfigurationSpec) {
	if new == nil {
		return
	}
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
	if new.SyncProviderArgs != nil && len(new.SyncProviderArgs) > 0 {
		fc.SyncProviderArgs = append(fc.SyncProviderArgs, new.SyncProviderArgs...)
	}
}

func (fc *FlagSourceConfigurationSpec) ToEnvVars() []corev1.EnvVar {
	envs := []corev1.EnvVar{}

	prefix := defaultSidecarEnvVarPrefix
	if p := os.Getenv(SidecarEnvVarPrefix); p != "" {
		prefix = p
	}

	if fc.MetricsPort != defaultMetricPort {
		envs = append(envs, corev1.EnvVar{
			Name:  fmt.Sprintf("%s_%s", prefix, SidecarMetricPortEnvVar),
			Value: fmt.Sprintf("%d", fc.MetricsPort),
		})
	}

	if fc.Port != defaultPort {
		envs = append(envs, corev1.EnvVar{
			Name:  fmt.Sprintf("%s_%s", prefix, SidecarPortEnvVar),
			Value: fmt.Sprintf("%d", fc.Port),
		})
	}

	if fc.Evaluator != defaultEvaluator {
		envs = append(envs, corev1.EnvVar{
			Name:  fmt.Sprintf("%s_%s", prefix, SidecarEvaluatorEnvVar),
			Value: fc.Evaluator,
		})
	}

	if fc.SocketPath != defaultSocketPath {
		envs = append(envs, corev1.EnvVar{
			Name:  fmt.Sprintf("%s_%s", prefix, SidecarSocketPathEnvVar),
			Value: fc.SocketPath,
		})
	}
	return envs
}

// FlagSourceConfigurationStatus defines the observed state of FlagSourceConfiguration
type FlagSourceConfigurationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:resource:shortName="fsc"
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// FlagSourceConfiguration is the Schema for the FlagSourceConfigurations API
type FlagSourceConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FlagSourceConfigurationSpec   `json:"spec,omitempty"`
	Status FlagSourceConfigurationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FlagSourceConfigurationList contains a list of FlagSourceConfiguration
type FlagSourceConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FlagSourceConfiguration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FlagSourceConfiguration{}, &FlagSourceConfigurationList{})
}
