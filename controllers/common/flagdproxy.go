package common

import (
	"github.com/open-feature/open-feature-operator/pkg/utils"
	"os"
)

const (
	ManagedByAnnotationValue      = "open-feature-operator"
	EnvVarPodNamespace            = "POD_NAMESPACE"
	EnvVarProxyImage              = "FLAGD_PROXY_IMAGE"
	EnvVarProxyTag                = "FLAGD_PROXY_TAG"
	EnvVarProxyPort               = "FLAGD_PROXY_PORT"
	EnvVarProxyMetricsPort        = "FLAGD_PROXY_METRICS_PORT"
	EnvVarProxyDebugLogging       = "FLAGD_PROXY_DEBUG_LOGGING"
	DefaultFlagdProxyImage        = "ghcr.io/open-feature/flagd-proxy"
	DefaultFlagdProxyTag          = "v0.2.0" //FLAGD_PROXY_TAG_RENOVATE
	DefaultFlagdProxyPort         = 8015
	DefaultFlagdProxyMetricsPort  = 8016
	DefaultFlagdProxyDebugLogging = false
	DefaultFlagdProxyNamespace    = "open-feature-operator-system"

	FlagdProxyDeploymentName     = "flagd-proxy"
	FlagdProxyServiceAccountName = "open-feature-operator-flagd-proxy"
	FlagdProxyServiceName        = "flagd-proxy-svc"

	OperatorDeploymentName = "open-feature-operator-controller-manager"
)

type FlagdProxyConfiguration struct {
	Port                   int
	MetricsPort            int
	DebugLogging           bool
	Image                  string
	Tag                    string
	Namespace              string
	OperatorDeploymentName string
}

func NewFlagdProxyConfiguration() (*FlagdProxyConfiguration, error) {
	config := &FlagdProxyConfiguration{
		Image:                  DefaultFlagdProxyImage,
		Tag:                    DefaultFlagdProxyTag,
		Namespace:              DefaultFlagdProxyNamespace,
		OperatorDeploymentName: OperatorDeploymentName,
	}
	ns, ok := os.LookupEnv(EnvVarPodNamespace)
	if ok {
		config.Namespace = ns
	}
	kpi, ok := os.LookupEnv(EnvVarProxyImage)
	if ok {
		config.Image = kpi
	}
	kpt, ok := os.LookupEnv(EnvVarProxyTag)
	if ok {
		config.Tag = kpt
	}
	port, err := utils.GetIntEnvVar(EnvVarProxyPort, DefaultFlagdProxyPort)
	if err != nil {
		return config, err
	}
	config.Port = port

	metricsPort, err := utils.GetIntEnvVar(EnvVarProxyMetricsPort, DefaultFlagdProxyMetricsPort)
	if err != nil {
		return config, err
	}
	config.MetricsPort = metricsPort

	kpDebugLogging, err := utils.GetBoolEnvVar(EnvVarProxyDebugLogging, DefaultFlagdProxyDebugLogging)
	if err != nil {
		return config, err
	}
	config.DebugLogging = kpDebugLogging

	return config, nil
}
