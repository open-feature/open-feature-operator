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

// FeatureFlagInProcessSourceSpec defines the desired state of FeatureFlagInProcessSource
type FeatureFlagInProcessSourceSpec struct {
	// Port defines the port to listen on, defaults to 8013
	// +kubebuilder:default:=8013
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
	// +kubebuilder:default:=lru
	// +kubebuilder:validation:Enum:=lru;disabled
	// +optional
	Cache string `json:"cache"`

	// CacheMaxSize
	// +kubebuilder:default:=1000
	// +optional
	CacheMaxSize int `json:"cacheMaxSize"`

	//EnvVars
	// +optional
	EnvVars []corev1.EnvVar `json:"envVars"`

	// EnvVarPrefix defines the prefix to be applied to all environment variables applied to the sidecar, default FLAGD
	// +optional
	// +kubebuilder:default:=FLAGD
	EnvVarPrefix string `json:"envVarPrefix"`
}

// FeatureFlagInProcessSourceStatus defines the observed state of FeatureFlagInProcessSource
type FeatureFlagInProcessSourceStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// FeatureFlagInProcessSource is the Schema for the featureflaginprocesssources API
type FeatureFlagInProcessSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FeatureFlagInProcessSourceSpec   `json:"spec,omitempty"`
	Status FeatureFlagInProcessSourceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FeatureFlagInProcessSourceList contains a list of FeatureFlagInProcessSource
type FeatureFlagInProcessSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FeatureFlagInProcessSource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FeatureFlagInProcessSource{}, &FeatureFlagInProcessSourceList{})
}

func (fc *FeatureFlagInProcessSourceSpec) Merge(new *FeatureFlagInProcessSourceSpec) {
	if new == nil {
		return
	}
	if len(new.EnvVars) != 0 {
		fc.EnvVars = append(fc.EnvVars, new.EnvVars...)
	}
	if new.Port != common.DefaultPort {
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

func (fc *FeatureFlagInProcessSourceSpec) ToEnvVars() []corev1.EnvVar {
	envs := []corev1.EnvVar{}

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
		Name:  common.EnvVarKey(fc.EnvVarPrefix, common.SocketPathEnvVar),
		Value: fc.SocketPath,
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, common.OfflineFlagSourcePathEnvVar),
		Value: fc.OfflineFlagSourcePath,
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, common.SelectorEnvVar),
		Value: fc.Selector,
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, common.CacheEnvVar),
		Value: fc.Cache,
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, common.CacheMaxSizeEnvVar),
		Value: fmt.Sprintf("%d", fc.CacheMaxSize),
	})

	return envs
}
