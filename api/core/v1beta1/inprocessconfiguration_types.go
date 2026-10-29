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

	"github.com/open-feature/open-feature-operator/api/core/v1beta1/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// InProcessConfigurationSpec defines the desired state of InProcessConfiguration
type InProcessConfigurationSpec struct {
	// Port defines the port to listen on, defaults to 8015
	// +optional
	Port *int32 `json:"port,omitempty"`

	// SocketPath defines the unix socket path to listen on
	// +optional
	SocketPath *string `json:"socketPath,omitempty"`

	// Host
	// +optional
	Host *string `json:"host,omitempty"`

	// TLS
	// +optional
	TLS *bool `json:"tls,omitempty"`

	// OfflineFlagSourcePath
	// +optional
	OfflineFlagSourcePath *string `json:"offlineFlagSourcePath,omitempty"`

	// Selector
	// +optional
	Selector *string `json:"selector,omitempty"`

	// Cache
	// +kubebuilder:validation:Pattern="^(lru|disabled)$"
	// +kubebuilder:validation:Type:=string
	// +optional
	Cache *string `json:"cache,omitempty"`

	// CacheMaxSize
	// +optional
	CacheMaxSize *int `json:"cacheMaxSize,omitempty"`

	// EnvVars
	// +optional
	EnvVars []corev1.EnvVar `json:"envVars"`

	// EnvVarPrefix defines the prefix to be applied to all environment variables applied to the sidecar, default FLAGD
	// +optional
	EnvVarPrefix *string `json:"envVarPrefix,omitempty"`
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

	if new.Port != nil {
		fc.Port = new.Port
	}
	if new.SocketPath != nil {
		fc.SocketPath = new.SocketPath
	}
	if new.Host != nil {
		fc.Host = new.Host
	}
	if new.EnvVarPrefix != nil {
		fc.EnvVarPrefix = new.EnvVarPrefix
	}
	if new.OfflineFlagSourcePath != nil {
		fc.OfflineFlagSourcePath = new.OfflineFlagSourcePath
	}
	if new.Selector != nil {
		fc.Selector = new.Selector
	}
	if new.Cache != nil {
		fc.Cache = new.Cache
	}
	if new.CacheMaxSize != nil {
		fc.CacheMaxSize = new.CacheMaxSize
	}
	if new.TLS != nil {
		fc.TLS = new.TLS
	}
}

func (fc *InProcessConfigurationSpec) ToEnvVars() []corev1.EnvVar {
	envs := []corev1.EnvVar{}
	prefix := ptr.Deref(fc.EnvVarPrefix, common.DefaultEnvVarPrefix)

	for _, envVar := range fc.EnvVars {
		newEnvVar := corev1.EnvVar{
			Name: common.EnvVarKey(prefix, envVar.Name),
		}

		if envVar.Value != "" {
			newEnvVar.Value = envVar.Value
		} else if envVar.ValueFrom != nil {
			newEnvVar.ValueFrom = envVar.ValueFrom
		}

		envs = append(envs, newEnvVar)
	}

	// default values are always included in the envVars
	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(prefix, common.HostEnvVar),
		Value: ptr.Deref(fc.Host, common.DefaultHost),
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(prefix, common.PortEnvVar),
		Value: fmt.Sprintf("%d", ptr.Deref(fc.Port, common.DefaultInProcessPort)),
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(prefix, common.TLSEnvVar),
		Value: fmt.Sprintf("%t", ptr.Deref(fc.TLS, common.DefaultTLS)),
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(prefix, common.CacheEnvVar),
		Value: ptr.Deref(fc.Cache, common.DefaultCache),
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(prefix, common.CacheMaxSizeEnvVar),
		Value: fmt.Sprintf("%d", ptr.Deref(fc.CacheMaxSize, int(common.DefaultCacheMaxSize))),
	})

	// sets the FLAGD_RESOLVER var to "in-process" to configure the provider for in-process evaluation mode
	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(prefix, common.ResolverEnvVar),
		Value: common.InProcessResolverType,
	})

	if v := ptr.Deref(fc.SocketPath, ""); v != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  common.EnvVarKey(prefix, common.SocketPathEnvVar),
			Value: v,
		})
	}

	if v := ptr.Deref(fc.OfflineFlagSourcePath, ""); v != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  common.EnvVarKey(prefix, common.OfflineFlagSourcePathEnvVar),
			Value: v,
		})
	}

	if v := ptr.Deref(fc.Selector, ""); v != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  common.EnvVarKey(prefix, common.SelectorEnvVar),
			Value: v,
		})
	}

	return envs
}
