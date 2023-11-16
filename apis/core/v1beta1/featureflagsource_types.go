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
	"os"
	"strconv"
	"strings"

	"github.com/open-feature/open-feature-operator/apis/core/v1beta1/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SidecarEnvVarPrefix              string = "SIDECAR_ENV_VAR_PREFIX"
	InputConfigurationEnvVarPrefix   string = "SIDECAR"
	SidecarMetricPortEnvVar          string = "METRICS_PORT"
	SidecarPortEnvVar                string = "PORT"
	SidecarSocketPathEnvVar          string = "SOCKET_PATH"
	SidecarEvaluatorEnvVar           string = "EVALUATOR"
	SidecarImageEnvVar               string = "IMAGE"
	SidecarVersionEnvVar             string = "TAG"
	SidecarProviderArgsEnvVar        string = "PROVIDER_ARGS"
	SidecarDefaultSyncProviderEnvVar string = "SYNC_PROVIDER"
	SidecarLogFormatEnvVar           string = "LOG_FORMAT"
	SidecarProbesEnabledVar          string = "PROBES_ENABLED"
	defaultSidecarEnvVarPrefix       string = "FLAGD"
	DefaultMetricPort                int32  = 8014
	defaultPort                      int32  = 8013
	defaultSocketPath                string = ""
	defaultEvaluator                 string = "json"
	defaultImage                     string = "ghcr.io/open-feature/flagd"
	// renovate: datasource=github-tags depName=open-feature/flagd/flagd
	defaultTag           string = "v0.7.0"
	defaultLogFormat     string = "json"
	defaultProbesEnabled bool   = true
)

// FeatureFlagSourceSpec defines the desired state of FeatureFlagSource
type FeatureFlagSourceSpec struct {
	// ManagemetPort defines the port to serve metrics on, defaults to 8014
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

	// Image allows for the sidecar image to be overridden, defaults to 'ghcr.io/open-feature/flagd'
	// +optional
	Image string `json:"image"`

	// Tag to be appended to the sidecar image, defaults to 'main'
	// +optional
	Tag string `json:"tag"`

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

	// Provider type - kubernetes, http(s), grpc(s) or filepath
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
func NewFeatureFlagSourceSpec() (*FeatureFlagSourceSpec, error) {
	fsc := &FeatureFlagSourceSpec{
		ManagementPort:      DefaultMetricPort,
		Port:                defaultPort,
		SocketPath:          defaultSocketPath,
		Evaluator:           defaultEvaluator,
		Image:               defaultImage,
		Tag:                 defaultTag,
		Sources:             []Source{},
		EnvVars:             []corev1.EnvVar{},
		SyncProviderArgs:    []string{},
		DefaultSyncProvider: common.SyncProviderKubernetes,
		EnvVarPrefix:        defaultSidecarEnvVarPrefix,
		LogFormat:           defaultLogFormat,
		RolloutOnChange:     nil,
		DebugLogging:        common.FalseVal(),
		OtelCollectorUri:    "",
	}

	// set default value derived from constant default
	probes := defaultProbesEnabled
	fsc.ProbesEnabled = &probes

	if metricsPort := os.Getenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarMetricPortEnvVar)); metricsPort != "" {
		metricsPortI, err := strconv.Atoi(metricsPort)
		if err != nil {
			return fsc, fmt.Errorf("unable to parse metrics port value %s to int32: %w", metricsPort, err)
		}
		fsc.ManagementPort = int32(metricsPortI)
	}

	if port := os.Getenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarPortEnvVar)); port != "" {
		portI, err := strconv.Atoi(port)
		if err != nil {
			return fsc, fmt.Errorf("unable to parse sidecar port value %s to int32: %w", port, err)
		}
		fsc.Port = int32(portI)
	}

	if socketPath := os.Getenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarSocketPathEnvVar)); socketPath != "" {
		fsc.SocketPath = socketPath
	}

	if evaluator := os.Getenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarEvaluatorEnvVar)); evaluator != "" {
		fsc.Evaluator = evaluator
	}

	if image := os.Getenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarImageEnvVar)); image != "" {
		fsc.Image = image
	}

	if tag := os.Getenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarVersionEnvVar)); tag != "" {
		fsc.Tag = tag
	}

	if syncProviderArgs := os.Getenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarProviderArgsEnvVar)); syncProviderArgs != "" {
		fsc.SyncProviderArgs = strings.Split(syncProviderArgs, ",") // todo: add documentation for this
	}

	if syncProvider := os.Getenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarDefaultSyncProviderEnvVar)); syncProvider != "" {
		fsc.DefaultSyncProvider = common.SyncProviderType(syncProvider)
	}

	if logFormat := os.Getenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarLogFormatEnvVar)); logFormat != "" {
		fsc.LogFormat = logFormat
	}

	if envVarPrefix := os.Getenv(SidecarEnvVarPrefix); envVarPrefix != "" {
		fsc.EnvVarPrefix = envVarPrefix
	}

	if probesEnabled := os.Getenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarProbesEnabledVar)); probesEnabled != "" {
		b, err := strconv.ParseBool(probesEnabled)
		if err != nil {
			return fsc, fmt.Errorf("unable to parse sidecar probes enabled %s to boolean: %w", probesEnabled, err)
		}
		fsc.ProbesEnabled = &b
	}

	return fsc, nil
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
	if new.Image != "" {
		fc.Image = new.Image
	}
	if new.Tag != "" {
		fc.Tag = new.Tag
	}
	if len(new.Sources) != 0 {
		fc.Sources = append(fc.Sources, new.Sources...)
	}
	if len(new.EnvVars) != 0 {
		fc.EnvVars = append(fc.EnvVars, new.EnvVars...)
	}
	if new.SyncProviderArgs != nil && len(new.SyncProviderArgs) > 0 {
		fc.SyncProviderArgs = append(fc.SyncProviderArgs, new.SyncProviderArgs...)
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

	if fc.ManagementPort != DefaultMetricPort {
		envs = append(envs, corev1.EnvVar{
			Name:  common.EnvVarKey(fc.EnvVarPrefix, SidecarMetricPortEnvVar),
			Value: fmt.Sprintf("%d", fc.ManagementPort),
		})
	}

	if fc.Port != defaultPort {
		envs = append(envs, corev1.EnvVar{
			Name:  common.EnvVarKey(fc.EnvVarPrefix, SidecarPortEnvVar),
			Value: fmt.Sprintf("%d", fc.Port),
		})
	}

	if fc.Evaluator != defaultEvaluator {
		envs = append(envs, corev1.EnvVar{
			Name:  common.EnvVarKey(fc.EnvVarPrefix, SidecarEvaluatorEnvVar),
			Value: fc.Evaluator,
		})
	}

	if fc.SocketPath != defaultSocketPath {
		envs = append(envs, corev1.EnvVar{
			Name:  common.EnvVarKey(fc.EnvVarPrefix, SidecarSocketPathEnvVar),
			Value: fc.SocketPath,
		})
	}

	if fc.LogFormat != defaultLogFormat {
		envs = append(envs, corev1.EnvVar{
			Name:  common.EnvVarKey(fc.EnvVarPrefix, SidecarLogFormatEnvVar),
			Value: fc.LogFormat,
		})
	}

	return envs
}
