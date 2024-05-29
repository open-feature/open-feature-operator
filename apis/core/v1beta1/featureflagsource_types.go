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

// FeatureFlagSourceSpec defines the desired state of FeatureFlagSource
type FeatureFlagSourceSpec struct {
	// ManagemetPort defines the port to serve management on, defaults to 8014
	// +optional
	ManagementPort int32 `json:"managementPort"`

	// Port defines the port to listen on, defaults to 8013
	// +optional
	Port int32 `json:"port"`

	// SocketPath defines the unix socket path to listen on
	// +optional
	SocketPath string `json:"socketPath"`

	// Evaluator sets an evaluator, defaults to 'json'
	// +optional
	Evaluator string `json:"evaluator"`

	// SyncProviders define the syncProviders and associated configuration to be applied to the sidecar
	// +kubebuilder:validation:MinItems=1
	Sources []Source `json:"sources"`

	// EnvVars define the env vars to be applied to the sidecar, any env vars in FeatureFlag CRs
	// are added at the lowest index, all values will have the EnvVarPrefix applied, default FLAGD
	// +optional
	EnvVars []corev1.EnvVar `json:"envVars"`

	// SyncProviderArgs are string arguments passed to all sync providers, defined as key values separated by =
	// +optional
	SyncProviderArgs []string `json:"syncProviderArgs"`

	// DefaultSyncProvider defines the default sync provider
	// +optional
	DefaultSyncProvider common.SyncProviderType `json:"defaultSyncProvider"`

	// LogFormat allows for the sidecar log format to be overridden, defaults to 'json'
	// +optional
	LogFormat string `json:"logFormat"`

	// EnvVarPrefix defines the prefix to be applied to all environment variables applied to the sidecar, default FLAGD
	// +optional
	EnvVarPrefix string `json:"envVarPrefix"`

	// RolloutOnChange dictates whether annotated deployments will be restarted when configuration changes are
	// detected in this CR, defaults to false
	// +optional
	RolloutOnChange *bool `json:"rolloutOnChange"`

	// ProbesEnabled defines whether to enable liveness and readiness probes of flagd sidecar. Default true (enabled).
	// +optional
	ProbesEnabled *bool `json:"probesEnabled"`

	// DebugLogging defines whether to enable --debug flag of flagd sidecar. Default false (disabled).
	// +optional
	DebugLogging *bool `json:"debugLogging"`

	// OtelCollectorUri defines whether to enable --otel-collector-uri flag of flagd sidecar. Default false (disabled).
	// +optional
	OtelCollectorUri string `json:"otelCollectorUri"`

	// Resources defines flagd sidecar resources. Default to operator sidecar-cpu-* and sidecar-ram-* flags.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources"`
}

type Source struct {
	// Source is a URI of the flag sources
	Source string `json:"source"`

	// Provider type - kubernetes, http(s), grpc(s) or file
	// +optional
	Provider common.SyncProviderType `json:"provider"`

	// HttpSyncBearerToken is a bearer token. Used by http(s) sync provider only
	// +optional
	HttpSyncBearerToken string `json:"httpSyncBearerToken"`

	// TLS - Enable/Disable secure TLS connectivity. Currently used only by GRPC sync
	// +optional
	TLS bool `json:"tls"`

	// CertPath is a path of a certificate to be used by grpc TLS connection
	// +optional
	CertPath string `json:"certPath"`

	// ProviderID is an identifier to be used in grpc provider
	// +optional
	ProviderID string `json:"providerID"`

	// Selector is a flag configuration selector used by grpc provider
	// +optional
	Selector string `json:"selector,omitempty"`

	// Interval is a flag configuration interval in seconds used by http provider
	// +optional
	Interval uint32 `json:"interval,omitempty"`
}

// FeatureFlagSourceStatus defines the observed state of FeatureFlagSource
type FeatureFlagSourceStatus struct {
}

//+kubebuilder:resource:shortName="ffs"
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// FeatureFlagSource is the Schema for the FeatureFlagSources API
type FeatureFlagSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FeatureFlagSourceSpec   `json:"spec,omitempty"`
	Status FeatureFlagSourceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FeatureFlagSourceList contains a list of FeatureFlagSource
type FeatureFlagSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FeatureFlagSource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FeatureFlagSource{}, &FeatureFlagSourceList{})
}

//nolint:gocyclo
func (fc *FeatureFlagSourceSpec) Merge(new *FeatureFlagSourceSpec) {
	if new == nil {
		return
	}
	if new.ManagementPort != 0 {
		fc.ManagementPort = new.ManagementPort
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
	if len(new.Sources) != 0 {
		fc.Sources = append(fc.Sources, new.Sources...)
	}
	if len(new.EnvVars) != 0 {
		fc.EnvVars = append(fc.EnvVars, new.EnvVars...)
		fc.EnvVars = common.RemoveDuplicateEnvVars(fc.EnvVars)
	}
	if len(new.SyncProviderArgs) != 0 {
		fc.SyncProviderArgs = append(fc.SyncProviderArgs, new.SyncProviderArgs...)
		fc.SyncProviderArgs = common.RemoveDuplicatesFromSlice[string](fc.SyncProviderArgs)
	}
	if new.EnvVarPrefix != "" {
		fc.EnvVarPrefix = new.EnvVarPrefix
	}
	if new.DefaultSyncProvider != "" {
		fc.DefaultSyncProvider = new.DefaultSyncProvider
	}
	if new.LogFormat != "" {
		fc.LogFormat = new.LogFormat
	}
	if new.RolloutOnChange != nil {
		fc.RolloutOnChange = new.RolloutOnChange
	}
	if new.ProbesEnabled != nil {
		fc.ProbesEnabled = new.ProbesEnabled
	}
	if new.DebugLogging != nil {
		fc.DebugLogging = new.DebugLogging
	}
	if new.OtelCollectorUri != "" {
		fc.OtelCollectorUri = new.OtelCollectorUri
	}
}

func (fc *FeatureFlagSourceSpec) ToEnvVars() []corev1.EnvVar {
	envs := []corev1.EnvVar{}

	for _, envVar := range fc.EnvVars {
		envs = append(envs, corev1.EnvVar{
			Name:  common.EnvVarKey(fc.EnvVarPrefix, envVar.Name),
			Value: envVar.Value,
		})
	}

	// default values are always included in the envVars
	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, common.ManagementPortEnvVar),
		Value: fmt.Sprintf("%d", fc.ManagementPort),
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, common.PortEnvVar),
		Value: fmt.Sprintf("%d", fc.Port),
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, common.EvaluatorEnvVar),
		Value: fc.Evaluator,
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, common.LogFormatEnvVar),
		Value: fc.LogFormat,
	})

	// sets the FLAGD_RESOLVER var to "rpc" to configure the provider for RPC evaluation mode
	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, common.ResolverEnvVar),
		Value: common.RPCResolverType,
	})

	if fc.SocketPath != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  common.EnvVarKey(fc.EnvVarPrefix, common.SocketPathEnvVar),
			Value: fc.SocketPath,
		})
	}

	return envs
}
