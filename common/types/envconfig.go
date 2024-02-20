package types

type EnvConfig struct {
	PodNamespace    string `envconfig:"POD_NAMESPACE" default:"open-feature-operator-system"`
	FlagdProxyImage string `envconfig:"FLAGD_PROXY_IMAGE" default:"ghcr.io/open-feature/flagd-proxy"`
	// renovate: datasource=github-tags depName=open-feature/flagd/flagd-proxy
	FlagdProxyTag            string `envconfig:"FLAGD_PROXY_TAG" default:"v0.5.0"`
	FlagdProxyPort           int    `envconfig:"FLAGD_PROXY_PORT" default:"8015"`
	FlagdProxyManagementPort int    `envconfig:"FLAGD_PROXY_MANAGEMENT_PORT" default:"8016"`
	FlagdProxyDebugLogging   bool   `envconfig:"FLAGD_PROXY_DEBUG_LOGGING" default:"false"`

	SidecarEnvVarPrefix   string `envconfig:"SIDECAR_ENV_VAR_PREFIX" default:"FLAGD"`
	SidecarManagementPort int    `envconfig:"SIDECAR_MANAGEMENT_PORT" default:"8014"`
	SidecarPort           int    `envconfig:"SIDECAR_PORT" default:"8013"`
	SidecarImage          string `envconfig:"SIDECAR_IMAGE" default:"ghcr.io/open-feature/flagd"`
	// renovate: datasource=github-tags depName=open-feature/flagd/flagd
	SidecarTag           string `envconfig:"SIDECAR_TAG" default:"v0.9.0"`
	SidecarSocketPath    string `envconfig:"SIDECAR_SOCKET_PATH" default:""`
	SidecarEvaluator     string `envconfig:"SIDECAR_EVALUATOR" default:"json"`
	SidecarProviderArgs  string `envconfig:"SIDECAR_PROVIDER_ARGS" default:""`
	SidecarSyncProvider  string `envconfig:"SIDECAR_SYNC_PROVIDER" default:"kubernetes"`
	SidecarLogFormat     string `envconfig:"SIDECAR_LOG_FORMAT" default:"json"`
	SidecarProbesEnabled bool   `envconfig:"SIDECAR_PROBES_ENABLED" default:"true"`
}
