package common

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

type SyncProviderType string

const (
	SyncProviderKubernetes SyncProviderType = "kubernetes"
	SyncProviderFilepath   SyncProviderType = "file"
	SyncProviderGcs        SyncProviderType = "gcs"
	SyncProviderHttp       SyncProviderType = "http"
	SyncProviderGrpc       SyncProviderType = "grpc"
	SyncProviderFlagdProxy SyncProviderType = "flagd-proxy"
)

const (
	ManagementPortEnvVar        string = "MANAGEMENT_PORT"
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
	DefaultEnvVarPrefix         string = "FLAGD"
	DefaultManagementPort       int32  = 8014
	DefaultRPCPort              int32  = 8013
	DefaultInProcessPort        int32  = 8015
	DefaultEvaluator            string = "json"
	DefaultLogFormat            string = "json"
	DefaultProbesEnabled        bool   = true
	DefaultTLS                  bool   = false
	DefaultHost                 string = "localhost"
	DefaultCache                string = "lru"
	DefaultCacheMaxSize         int32  = 1000
	InProcessResolverType       string = "in-process"
	RPCResolverType             string = "rpc"
)

func (s SyncProviderType) IsKubernetes() bool {
	return s == SyncProviderKubernetes
}

func (s SyncProviderType) IsHttp() bool {
	return s == SyncProviderHttp
}

func (s SyncProviderType) IsGcs() bool {
	return s == SyncProviderGcs
}

func (s SyncProviderType) IsFilepath() bool {
	return s == SyncProviderFilepath
}

func (s SyncProviderType) IsGrpc() bool {
	return s == SyncProviderGrpc
}

func (s SyncProviderType) IsFlagdProxy() bool {
	return s == SyncProviderFlagdProxy
}

func TrueVal() *bool {
	b := true
	return &b
}

func FalseVal() *bool {
	b := false
	return &b
}

func EnvVarKey(prefix string, suffix string) string {
	return fmt.Sprintf("%s_%s", prefix, suffix)
}

// unique string used to create unique volume mount and file name
func FeatureFlagConfigurationId(namespace, name string) string {
	return EnvVarKey(namespace, name)
}

// unique key (and filename) for configMap data
func FeatureFlagConfigMapKey(namespace, name string) string {
	return fmt.Sprintf("%s.flagd.json", FeatureFlagConfigurationId(namespace, name))
}

func RemoveDuplicateEnvVars(input []corev1.EnvVar) []corev1.EnvVar {
	out := make([]corev1.EnvVar, 0, len(input))
	for i := len(input) - 1; i >= 0; i-- {
		if !isEnvVarNamePresent(out, input[i]) {
			out = append(out, input[i])
		}
	}
	return out
}

func isEnvVarNamePresent(slice []corev1.EnvVar, item corev1.EnvVar) bool {
	for _, i := range slice {
		if i.Name == item.Name {
			return true
		}
	}
	return false
}

func RemoveDuplicatesFromSlice[T comparable](input []T) []T {
	seen := make(map[T]bool)
	result := []T{}

	for _, item := range input {
		if _, ok := seen[item]; !ok {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}
