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
	"fmt"

	"github.com/open-feature/open-feature-operator/apis/core/v1beta1/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InProcessConfigurationSpec defines the desired state of InProcessConfiguration
type InProcessConfigurationSpec struct {
	// Port defines the port to listen on, defaults to 8015
	// +kubebuilder:default:=8015
	// +optional
	Port int32 `json:"port"`

	// SocketPath defines the unix socket path to listen on
	// +optional
	SocketPath string `json:"socketPath"`

	// Host
	// +kubebuilder:default:=localhost
	// +optional
	Host string `json:"host"`

	// TLS
	// +kubebuilder:default:=false
	// +optional
	TLS bool `json:"tls"`

	// OfflineFlagSourcePath
	// +optional
	OfflineFlagSourcePath string `json:"offlineFlagSourcePath"`

	// Selector
	// +optional
	Selector string `json:"selector"`

	// Cache
	// +kubebuilder:default:="lru"
	// +kubebuilder:validation:Pattern="^(lru|disabled)$"
	// +kubebuilder:validation:Type:=string
	// +optional
	Cache string `json:"cache"`

	// CacheMaxSize
	// +kubebuilder:default:=1000
	// +optional
	CacheMaxSize int `json:"cacheMaxSize"`

	// EnvVars
	// +optional
	EnvVars []corev1.EnvVar `json:"envVars"`

	// EnvVarPrefix defines the prefix to be applied to all environment variables applied to the sidecar, default FLAGD
	// +optional
	// +kubebuilder:default:=FLAGD
	EnvVarPrefix string `json:"envVarPrefix"`
}

// InProcessConfigurationStatus defines the observed state of InProcessConfiguration
type InProcessConfigurationStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// InProcessConfiguration is the Schema for the inprocesconfigurations API
type InProcessConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InProcessConfigurationSpec   `json:"spec,omitempty"`
	Status InProcessConfigurationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InProcessConfigurationList contains a list of InProcessConfiguration
type InProcessConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InProcessConfiguration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InProcessConfiguration{}, &InProcessConfigurationList{})
}

func (fc *InProcessConfigurationSpec) Merge(new *InProcessConfigurationSpec) {
	if new == nil {
		return
	}
	if len(new.EnvVars) != 0 {
		fc.EnvVars = append(fc.EnvVars, new.EnvVars...)
		fc.EnvVars = common.RemoveDuplicateEnvVars(fc.EnvVars)
	}

	if new.Port != common.DefaultInProcessPort {
		fc.Port = new.Port
	}
	if new.SocketPath != "" {
		fc.SocketPath = new.SocketPath
	}
	if new.Host != common.DefaultHost {
		fc.Host = new.Host
	}
	if new.EnvVarPrefix != common.DefaultEnvVarPrefix {
		fc.EnvVarPrefix = new.EnvVarPrefix
	}
	if new.OfflineFlagSourcePath != "" {
		fc.OfflineFlagSourcePath = new.OfflineFlagSourcePath
	}
	if new.Selector != "" {
		fc.Selector = new.Selector
	}
	if new.Cache != common.DefaultCache {
		fc.Cache = new.Cache
	}
	if new.CacheMaxSize != int(common.DefaultCacheMaxSize) {
		fc.CacheMaxSize = new.CacheMaxSize
	}
	if new.TLS != common.DefaultTLS {
		fc.TLS = new.TLS
	}
}

func (fc *InProcessConfigurationSpec) ToEnvVars() []corev1.EnvVar {
	envs := []corev1.EnvVar{}

	// fill out the default values in case the values are empty
	fc.fillMissingDefaults()

	for _, envVar := range fc.EnvVars {
		envs = append(envs, corev1.EnvVar{
			Name:  common.EnvVarKey(fc.EnvVarPrefix, envVar.Name),
			Value: envVar.Value,
		})
	}

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, common.HostEnvVar),
		Value: fc.Host,
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, common.PortEnvVar),
		Value: fmt.Sprintf("%d", fc.Port),
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, common.TLSEnvVar),
		Value: fmt.Sprintf("%t", fc.TLS),
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, common.CacheEnvVar),
		Value: fc.Cache,
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, common.CacheMaxSizeEnvVar),
		Value: fmt.Sprintf("%d", fc.CacheMaxSize),
	})

	// sets the FLAGD_RESOLVER var to "in-process" to configure the provider for in-process evaluation mode
	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, common.ResolverEnvVar),
		Value: common.InProcessResolverType,
	})

	if fc.SocketPath != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  common.EnvVarKey(fc.EnvVarPrefix, common.SocketPathEnvVar),
			Value: fc.SocketPath,
		})
	}

	if fc.OfflineFlagSourcePath != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  common.EnvVarKey(fc.EnvVarPrefix, common.OfflineFlagSourcePathEnvVar),
			Value: fc.OfflineFlagSourcePath,
		})
	}

	if fc.Selector != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  common.EnvVarKey(fc.EnvVarPrefix, common.SelectorEnvVar),
			Value: fc.Selector,
		})
	}

	return envs
}

func (fc *InProcessConfigurationSpec) fillMissingDefaults() {
	if fc.EnvVarPrefix == "" {
		fc.EnvVarPrefix = common.DefaultEnvVarPrefix
	}

	if fc.Host == "" {
		fc.Host = common.DefaultHost
	}

	if fc.Port == 0 {
		fc.Port = common.DefaultInProcessPort
	}

	if fc.Cache == "" {
		fc.Cache = common.DefaultCache
	}

	if fc.CacheMaxSize == 0 {
		fc.CacheMaxSize = int(common.DefaultCacheMaxSize)
	}
}
