package types

type EnvConfig struct {
	PodNamespace           string `envconfig:"POD_NAMESPACE" default:"open-feature-operator-system"`
	FlagdProxyImage        string `envconfig:"FLAGD_PROXY_IMAGE" default:"ghcr.io/open-feature/flagd-proxy"`
	FlagsValidationEnabled bool   `envconfig:"FLAGS_VALIDATION_ENABLED" default:"true"`
	FlagdProxyReplicaCount int    `envconfig:"FLAGD_PROXY_REPLICA_COUNT" default:"1"`
	// renovate: datasource=github-tags depName=open-feature/flagd/flagd-proxy
	FlagdProxyTag            string `envconfig:"FLAGD_PROXY_TAG" default:"v0.6.4"`
	FlagdProxyPort           int    `envconfig:"FLAGD_PROXY_PORT" default:"8015"`
	FlagdProxyManagementPort int    `envconfig:"FLAGD_PROXY_MANAGEMENT_PORT" default:"8016"`
	FlagdProxyDebugLogging   bool   `envconfig:"FLAGD_PROXY_DEBUG_LOGGING" default:"false"`

	FlagdImage string `envconfig:"FLAGD_IMAGE" default:"ghcr.io/open-feature/flagd"`
	// renovate: datasource=github-tags depName=open-feature/flagd/flagd
	FlagdTag            string `envconfig:"FLAGD_TAG" default:"v0.11.1"`
	FlagdPort           int    `envconfig:"FLAGD_PORT" default:"8013"`
	FlagdOFREPPort      int    `envconfig:"FLAGD_OFREP_PORT" default:"8016"`
	FlagdSyncPort       int    `envconfig:"FLAGD_SYNC_PORT" default:"8015"`
	FlagdManagementPort int    `envconfig:"FLAGD_MANAGEMENT_PORT" default:"8014"`
	FlagdDebugLogging   bool   `envconfig:"FLAGD_DEBUG_LOGGING" default:"false"`

	SidecarEnvVarPrefix   string `envconfig:"SIDECAR_ENV_VAR_PREFIX" default:"FLAGD"`
	SidecarManagementPort int    `envconfig:"SIDECAR_MANAGEMENT_PORT" default:"8014"`
	SidecarPort           int    `envconfig:"SIDECAR_PORT" default:"8013"`
	SidecarImage          string `envconfig:"SIDECAR_IMAGE" default:"ghcr.io/open-feature/flagd"`
	// renovate: datasource=github-tags depName=open-feature/flagd/flagd
	SidecarTag           string `envconfig:"SIDECAR_TAG" default:"v0.11.1"`
	SidecarSocketPath    string `envconfig:"SIDECAR_SOCKET_PATH" default:""`
	SidecarEvaluator     string `envconfig:"SIDECAR_EVALUATOR" default:"json"`
	SidecarProviderArgs  string `envconfig:"SIDECAR_PROVIDER_ARGS" default:""`
	SidecarSyncProvider  string `envconfig:"SIDECAR_SYNC_PROVIDER" default:"kubernetes"`
	SidecarLogFormat     string `envconfig:"SIDECAR_LOG_FORMAT" default:"json"`
	SidecarProbesEnabled bool   `envconfig:"SIDECAR_PROBES_ENABLED" default:"true"`
	// in-process configuration
	InProcessPort                  int    `envconfig:"IN_PROCESS_PORT" default:"8015"`
	InProcessSocketPath            string `envconfig:"IN_PROCESS_SOCKET_PATH" default:""`
	InProcessHost                  string `envconfig:"IN_PROCESS_HOST" default:"localhost"`
	InProcessTLS                   bool   `envconfig:"IN_PROCESS_TLS" default:"false"`
	InProcessOfflineFlagSourcePath string `envconfig:"IN_PROCESS_OFFLINE_FLAG_SOURCE_PATH" default:""`
	InProcessSelector              string `envconfig:"IN_PROCESS_SELECTOR" default:""`
	InProcessCache                 string `envconfig:"IN_PROCESS_CACHE" default:"lru"`
	InProcessEnvVarPrefix          string `envconfig:"IN_PROCESS_ENV_VAR_PREFIX" default:"FLAGD"`
	InProcessCacheMaxSize          int    `envconfig:"IN_PROCESS_CACHE_MAX_SIZE" default:"1000"`
}
