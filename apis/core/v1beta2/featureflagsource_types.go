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

package v1beta2

import (
	"fmt"

	"github.com/open-feature/open-feature-operator/apis/core/v1beta2/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	MetricPortEnvVar            string = "MANAGEMENT_PORT"
	PortEnvVar                  string = "PORT"
	HostEnvVar                  string = "HOST"
	TLSEnvVar                   string = "TLS"
	SocketPathEnvVar            string = "SOCKET_PATH"
	OfflineFlagSourcePathEnvVar string = "OFFLINE_FLAG_SOURCE_PATH"
	SelectorEnvVar              string = "SOURCE_SELECTOR"
	CacheEnvVar                 string = "CACHE"
	CacheMaxSizeEnvVar          string = "MAX_CACHE_SIZE"
	ResolverEnvVar              string = "RESOLVER"
	EvaluatorEnvVar             string = "EVALUATOR"
	ImageEnvVar                 string = "IMAGE"
	VersionEnvVar               string = "TAG"
	ProviderArgsEnvVar          string = "PROVIDER_ARGS"
	DefaultSyncProviderEnvVar   string = "SYNC_PROVIDER"
	LogFormatEnvVar             string = "LOG_FORMAT"
	ProbesEnabledVar            string = "PROBES_ENABLED"
	defaultEnvVarPrefix         string = "FLAGD"
	DefaultManagementPort       int32  = 8014
	defaultPort                 int32  = 8013
	defaultEvaluator            string = "json"
	defaultLogFormat            string = "json"
	defaultProbesEnabled        bool   = true
	defaultTLS                  bool   = false
	defaultHost                 string = "localhost"
	defaultCache                string = "lru"
	defaultCacheMaxSize         int32  = 1000
)

// FeatureFlagSourceSpec defines the desired state of FeatureFlagSource
type FeatureFlagSourceSpec struct {
	RPC *RPCConf `json:"rpc,omitempty"`

	InProces *InProcessConf `json:"inProcess,omitempty"`

	// EnvVarPrefix defines the prefix to be applied to all environment variables applied to the sidecar, default FLAGD
	// +optional
	// +kubebuilder:default:=FLAGD
	EnvVarPrefix string `json:"envVarPrefix"`
}

type InProcessConf struct {
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
}

type RPCConf struct {
	// ManagemetPort defines the port to serve management on, defaults to 8014
	// +kubebuilder:default:=8014
	// +optional
	ManagementPort int32 `json:"managementPort"`

	// Port defines the port to listen on, defaults to 8013
	// +kubebuilder:default:=8013
	// +optional
	Port int32 `json:"port"`

	// SocketPath defines the unix socket path to listen on
	// +optional
	SocketPath string `json:"socketPath"`

	// Evaluator sets an evaluator, defaults to 'json'
	// +kubebuilder:default:=json
	// +optional
	Evaluator string `json:"evaluator"`

	// SyncProviders define the syncProviders and associated configuration to be applied to the sidecar
	// +kubebuilder:validation:MinItems=1
	// +optional
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
	// +kubebuilder:default:=kubernetes
	// +kubebuilder:validation:Enum:=kubernetes;file;http;grpc;flagd-proxy
	DefaultSyncProvider common.SyncProviderType `json:"defaultSyncProvider"`

	// LogFormat allows for the sidecar log format to be overridden, defaults to 'json'
	// +optional
	// +kubebuilder:default:=json
	LogFormat string `json:"logFormat"`

	// RolloutOnChange dictates whether annotated deployments will be restarted when configuration changes are
	// detected in this CR, defaults to false
	// +optional
	// +kubebuilder:default:=false
	RolloutOnChange bool `json:"rolloutOnChange"`

	// ProbesEnabled defines whether to enable liveness and readiness probes of flagd sidecar. Default true (enabled).
	// +optional
	// +kubebuilder:default:=true
	ProbesEnabled bool `json:"probesEnabled"`

	// DebugLogging defines whether to enable --debug flag of flagd sidecar. Default false (disabled).
	// +optional
	// +kubebuilder:default:=false
	DebugLogging bool `json:"debugLogging"`

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
	// +kubebuilder:default:=kubernetes
	// +kubebuilder:validation:Enum:=kubernetes;file;http;grpc;flagd-proxy
	Provider common.SyncProviderType `json:"provider"`

	// HttpSyncBearerToken is a bearer token. Used by http(s) sync provider only
	// +optional
	HttpSyncBearerToken string `json:"httpSyncBearerToken"`

	// TLS - Enable/Disable secure TLS connectivity. Currently used only by GRPC sync
	// +optional
	// +kubebuilder:default:=false
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
func (fc *FeatureFlagSourceSpec) MergeRPC(new *FeatureFlagSourceSpec) {
	if new == nil {
		return
	}

	fc.RPC.ManagementPort = new.RPC.ManagementPort
	fc.RPC.Port = new.RPC.Port
	fc.RPC.SocketPath = new.RPC.SocketPath
	fc.RPC.Evaluator = new.RPC.Evaluator
	fc.EnvVarPrefix = new.EnvVarPrefix
	fc.RPC.DefaultSyncProvider = new.RPC.DefaultSyncProvider
	fc.RPC.LogFormat = new.RPC.LogFormat
	fc.RPC.RolloutOnChange = new.RPC.RolloutOnChange
	fc.RPC.ProbesEnabled = new.RPC.ProbesEnabled
	fc.RPC.DebugLogging = new.RPC.DebugLogging
	fc.RPC.OtelCollectorUri = new.RPC.OtelCollectorUri

	if len(new.RPC.Sources) != 0 {
		fc.RPC.Sources = append(fc.RPC.Sources, new.RPC.Sources...)
	}
	if len(new.RPC.EnvVars) != 0 {
		fc.RPC.EnvVars = append(fc.RPC.EnvVars, new.RPC.EnvVars...)
	}
	if len(new.RPC.SyncProviderArgs) != 0 {
		fc.RPC.SyncProviderArgs = append(fc.RPC.SyncProviderArgs, new.RPC.SyncProviderArgs...)
	}

	fc.InProces = nil
}

//nolint:gocyclo
func (fc *FeatureFlagSourceSpec) MergeInProcess(new *FeatureFlagSourceSpec) {
	if new == nil {
		return
	}
	if len(new.InProces.EnvVars) != 0 {
		fc.InProces.EnvVars = append(fc.InProces.EnvVars, new.InProces.EnvVars...)
	}
	fc.InProces.Port = new.InProces.Port
	fc.InProces.SocketPath = new.InProces.SocketPath
	fc.InProces.Host = new.InProces.Host
	fc.EnvVarPrefix = new.EnvVarPrefix
	fc.InProces.OfflineFlagSourcePath = new.InProces.OfflineFlagSourcePath
	fc.InProces.Selector = new.InProces.Selector
	fc.InProces.Cache = new.InProces.Cache
	fc.InProces.CacheMaxSize = new.InProces.CacheMaxSize
	fc.InProces.TLS = new.InProces.TLS

	fc.RPC = nil
}

func (fc *FeatureFlagSourceSpec) ToEnvVarsRPC() []corev1.EnvVar {
	envs := []corev1.EnvVar{}

	for _, envVar := range fc.RPC.EnvVars {
		envs = append(envs, corev1.EnvVar{
			Name:  common.EnvVarKey(fc.EnvVarPrefix, envVar.Name),
			Value: envVar.Value,
		})
	}

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, MetricPortEnvVar),
		Value: fmt.Sprintf("%d", fc.RPC.ManagementPort),
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, PortEnvVar),
		Value: fmt.Sprintf("%d", fc.RPC.Port),
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, EvaluatorEnvVar),
		Value: fc.RPC.Evaluator,
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, SocketPathEnvVar),
		Value: fc.RPC.SocketPath,
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, LogFormatEnvVar),
		Value: fc.RPC.LogFormat,
	})

	return envs
}

func (fc *FeatureFlagSourceSpec) ToEnvVarsInProcess() []corev1.EnvVar {
	envs := []corev1.EnvVar{}

	for _, envVar := range fc.InProces.EnvVars {
		envs = append(envs, corev1.EnvVar{
			Name:  common.EnvVarKey(fc.EnvVarPrefix, envVar.Name),
			Value: envVar.Value,
		})
	}

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, HostEnvVar),
		Value: fc.InProces.Host,
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, PortEnvVar),
		Value: fmt.Sprintf("%d", fc.InProces.Port),
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, TLSEnvVar),
		Value: fmt.Sprintf("%t", fc.InProces.TLS),
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, SocketPathEnvVar),
		Value: fc.RPC.SocketPath,
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, OfflineFlagSourcePathEnvVar),
		Value: fc.InProces.OfflineFlagSourcePath,
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, SelectorEnvVar),
		Value: fc.InProces.Selector,
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, CacheEnvVar),
		Value: fc.InProces.Cache,
	})

	envs = append(envs, corev1.EnvVar{
		Name:  common.EnvVarKey(fc.EnvVarPrefix, CacheMaxSizeEnvVar),
		Value: fmt.Sprintf("%d", fc.InProces.CacheMaxSize),
	})

	return envs
}
