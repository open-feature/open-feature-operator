package common

import "fmt"

type SyncProviderType string

const (
	SyncProviderKubernetes SyncProviderType = "kubernetes"
	SyncProviderFilepath   SyncProviderType = "file"
	SyncProviderHttp       SyncProviderType = "http"
	SyncProviderGrpc       SyncProviderType = "grpc"
	SyncProviderFlagdProxy SyncProviderType = "flagd-proxy"
)

func (s SyncProviderType) IsKubernetes() bool {
	return s == SyncProviderKubernetes
}

func (s SyncProviderType) IsHttp() bool {
	return s == SyncProviderHttp
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
	return fmt.Sprintf("%s_%s", namespace, name)
}

// unique key (and filename) for configMap data
func FeatureFlagConfigMapKey(namespace, name string) string {
	return fmt.Sprintf("%s.flagd.json", FeatureFlagConfigurationId(namespace, name))
}
