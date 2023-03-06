package common

import "fmt"

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
	DefaultSidecarEnvVarPrefix       string = "FLAGD"
	DefaultMetricPort                int32  = 8014
	DefaultPort                      int32  = 8013
	DefaultSocketPath                string = ""
	DefaultEvaluator                 string = "json"
	DefaultImage                     string = "ghcr.io/open-feature/flagd"
	// `INPUT FLAGD VERSION` is replaced in the `update-flagd` Makefile target
	DefaultTag             string = "INPUT_FLAGD_VERSION"
	DefaultLogFormat       string = "json"
	SyncProviderKubernetes string = "kubernetes"
	SyncProviderFilepath   string = "filepath"
	SyncProviderHttp       string = "http"
	defaultSyncProvider           = SyncProviderKubernetes
)

func EnvVarKey(prefix string, suffix string) string {
	return fmt.Sprintf("%s_%s", prefix, suffix)
}
