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

type SyncProviderType string

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
	// INPUT_FLAGD_VERSION` is replaced in the `update-flagd` Makefile target
	defaultTag             string           = "INPUT_FLAGD_VERSION"
	defaultLogFormat       string           = "json"
	defaultProbesEnabled   bool             = true
	SyncProviderKubernetes SyncProviderType = "kubernetes"
	SyncProviderFilepath   SyncProviderType = "filepath"
	SyncProviderHttp       SyncProviderType = "http"
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

	// SyncProviderArgs are string arguments passed to all sync providers, defined as key values separated by =
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

	// DefaultSyncProvider defines the default sync provider
	// +optional
	DefaultSyncProvider SyncProviderType `json:"defaultSyncProvider"`

	// Sources defines the syncProviders and associated configuration to be applied to the sidecar
	// +kubebuilder:validation:MinItems=1
	Sources []Source `json:"sources"`

	// EnvVars define the env vars to be applied to the sidecar, any env vars in FeatureFlagConfiguration CRs
	// are added at the lowest index, all values will have the EnvVarPrefix applied
	// +optional
	EnvVars []corev1.EnvVar `json:"envVars"`

	// EnvVarPrefix defines the prefix to be applied to all environment variables applied to the sidecar, default FLAGD
	// +optional
	EnvVarPrefix string `json:"envVarPrefix"`

	// LogFormat allows for the sidecar log format to be overridden, defaults to 'json'
	// +optional
	LogFormat string `json:"logFormat"`

	// RolloutOnChange dictates whether annotated deployments will be restarted when configuration changes are
	// detected in this CR, defaults to false
	// +optional
	RolloutOnChange *bool `json:"rolloutOnChange"`

	// ProbesEnabled defines whether to enable liveness and readiness probes of flagd sidecar. Default true(enabled)
	// +optional
	ProbesEnabled *bool `json:"probesEnabled"`
}

type Source struct {
	Source string `json:"source"`
	// +optional
	Provider SyncProviderType `json:"provider"`
	// +optional
	HttpSyncBearerToken string `json:"httpSyncBearerToken"`
	// LogFormat allows for the sidecar log format to be overridden, defaults to 'json'
	// +optional
	LogFormat string `json:"logFormat"`
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

func NewFlagSourceConfigurationSpec() (*FlagSourceConfigurationSpec, error) {
	fsc := &FlagSourceConfigurationSpec{
		MetricsPort:         DefaultMetricPort,
		Port:                defaultPort,
		SocketPath:          defaultSocketPath,
		Evaluator:           defaultEvaluator,
		Image:               defaultImage,
		Tag:                 defaultTag,
		Sources:             []Source{},
		EnvVars:             []corev1.EnvVar{},
		SyncProviderArgs:    []string{},
		DefaultSyncProvider: SyncProviderKubernetes,
		EnvVarPrefix:        defaultSidecarEnvVarPrefix,
		LogFormat:           defaultLogFormat,
		RolloutOnChange:     nil,
	}

	// set default value derived from constant default
	probes := defaultProbesEnabled
	fsc.ProbesEnabled = &probes

	if metricsPort := os.Getenv(envVarKey(InputConfigurationEnvVarPrefix, SidecarMetricPortEnvVar)); metricsPort != "" {
		metricsPortI, err := strconv.Atoi(metricsPort)
		if err != nil {
			return fsc, fmt.Errorf("unable to parse metrics port value %s to int32: %w", metricsPort, err)
		}
		fsc.MetricsPort = int32(metricsPortI)
	}

	if port := os.Getenv(envVarKey(InputConfigurationEnvVarPrefix, SidecarPortEnvVar)); port != "" {
		portI, err := strconv.Atoi(port)
		if err != nil {
			return fsc, fmt.Errorf("unable to parse sidecar port value %s to int32: %w", port, err)
		}
		fsc.Port = int32(portI)
	}

	if socketPath := os.Getenv(envVarKey(InputConfigurationEnvVarPrefix, SidecarSocketPathEnvVar)); socketPath != "" {
		fsc.SocketPath = socketPath
	}

	if evaluator := os.Getenv(envVarKey(InputConfigurationEnvVarPrefix, SidecarEvaluatorEnvVar)); evaluator != "" {
		fsc.Evaluator = evaluator
	}

	if image := os.Getenv(envVarKey(InputConfigurationEnvVarPrefix, SidecarImageEnvVar)); image != "" {
		fsc.Image = image
	}

	if tag := os.Getenv(envVarKey(InputConfigurationEnvVarPrefix, SidecarVersionEnvVar)); tag != "" {
		fsc.Tag = tag
	}

	if syncProviderArgs := os.Getenv(envVarKey(InputConfigurationEnvVarPrefix, SidecarProviderArgsEnvVar)); syncProviderArgs != "" {
		fsc.SyncProviderArgs = strings.Split(syncProviderArgs, ",") // todo: add documentation for this
	}

	if syncProvider := os.Getenv(envVarKey(InputConfigurationEnvVarPrefix, SidecarDefaultSyncProviderEnvVar)); syncProvider != "" {
		fsc.DefaultSyncProvider = SyncProviderType(syncProvider)
	}

	if logFormat := os.Getenv(fmt.Sprintf("%s_%s", InputConfigurationEnvVarPrefix, SidecarLogFormatEnvVar)); logFormat != "" {
		fsc.LogFormat = logFormat
	}

	if envVarPrefix := os.Getenv(SidecarEnvVarPrefix); envVarPrefix != "" {
		fsc.EnvVarPrefix = envVarPrefix
	}

	if probesEnabled := os.Getenv(fmt.Sprintf("%s_%s", InputConfigurationEnvVarPrefix, SidecarProbesEnabledVar)); probesEnabled != "" {
		b, err := strconv.ParseBool(probesEnabled)
		if err != nil {
			return fsc, fmt.Errorf("unable to parse sidecar probes enabled %s to boolean: %w", probesEnabled, err)
		}
		fsc.ProbesEnabled = &b
	}

	return fsc, nil
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
}

func (fc *FlagSourceConfigurationSpec) ToEnvVars() []corev1.EnvVar {
	envs := []corev1.EnvVar{}

	for _, envVar := range fc.EnvVars {
		envs = append(envs, corev1.EnvVar{
			Name:  envVarKey(fc.EnvVarPrefix, envVar.Name),
			Value: envVar.Value,
		})
	}

	if fc.MetricsPort != DefaultMetricPort {
		envs = append(envs, corev1.EnvVar{
			Name:  envVarKey(fc.EnvVarPrefix, SidecarMetricPortEnvVar),
			Value: fmt.Sprintf("%d", fc.MetricsPort),
		})
	}

	if fc.Port != defaultPort {
		envs = append(envs, corev1.EnvVar{
			Name:  envVarKey(fc.EnvVarPrefix, SidecarPortEnvVar),
			Value: fmt.Sprintf("%d", fc.Port),
		})
	}

	if fc.Evaluator != defaultEvaluator {
		envs = append(envs, corev1.EnvVar{
			Name:  envVarKey(fc.EnvVarPrefix, SidecarEvaluatorEnvVar),
			Value: fc.Evaluator,
		})
	}

	if fc.SocketPath != defaultSocketPath {
		envs = append(envs, corev1.EnvVar{
			Name:  envVarKey(fc.EnvVarPrefix, SidecarSocketPathEnvVar),
			Value: fc.SocketPath,
		})
	}

	if fc.LogFormat != defaultLogFormat {
		envs = append(envs, corev1.EnvVar{
			Name:  envVarKey(fc.EnvVarPrefix, SidecarLogFormatEnvVar),
			Value: fc.LogFormat,
		})
	}

	return envs
}

func (s SyncProviderType) IsKubernetes() bool {
	return s == SyncProviderKubernetes
}

func (s SyncProviderType) IsHttp() bool {
	return s == SyncProviderHttp
}

func (s SyncProviderType) IsFilepath() bool {
	return s == SyncProviderFilepath
}

func envVarKey(prefix string, suffix string) string {
	return fmt.Sprintf("%s_%s", prefix, suffix)
}

func init() {
	SchemeBuilder.Register(&FlagSourceConfiguration{}, &FlagSourceConfigurationList{})
}
