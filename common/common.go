package common

import (
	"context"
	"fmt"
	"time"

	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	appsV1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ReconcileErrorInterval      = 10 * time.Second
	ReconcileSuccessInterval    = 120 * time.Second
	FinalizerName               = "featureflag.core.openfeature.dev/finalizer"
	OpenFeatureAnnotationPath   = "spec.template.metadata.annotations.openfeature.dev/openfeature.dev"
	FeatureFlagSourceAnnotation = "featureflagsource"
	OpenFeatureAnnotationRoot   = "openfeature.dev"
)

type EnvConfig struct {
	PodNamespace    string `envconfig:"POD_NAMESPACE" default:"open-feature-operator-system"`
	FlagdProxyImage string `envconfig:"FLAGD_PROXY_IMAGE" default:"ghcr.io/open-feature/flagd-proxy"`
	// renovate: datasource=github-tags depName=open-feature/flagd/flagd-proxy
	FlagdProxyTag            string `envconfig:"FLAGD_PROXY_TAG" default:"v0.3.0"`
	FlagdProxyPort           int    `envconfig:"FLAGD_PROXY_PORT" default:"8015"`
	FlagdProxyManagementPort int    `envconfig:"FLAGD_PROXY_MANAGEMENT_PORT" default:"8016"`
	FlagdProxyDebugLogging   bool   `envconfig:"FLAGD_PROXY_DEBUG_LOGGING" default:"false"`

	SidecarEnvVarPrefix   string `envconfig:"SIDECAR_ENV_VAR_PREFIX" default:"FLAGD"`
	SidecarManagementPort int    `envconfig:"SIDECAR_MANAGEMENT_PORT" default:"8014"`
	SidecarPort           int    `envconfig:"SIDECAR_PORT" default:"8013"`
	SidecarImage          string `envconfig:"SIDECAR_IMAGE" default:"ghcr.io/open-feature/flagd"`
	// renovate: datasource=github-tags depName=open-feature/flagd/flagd-proxy
	SidecarTag           string `envconfig:"SIDECAR_TAG" default:"v0.7.0"`
	SidecarSocketPath    string `envconfig:"SIDECAR_SOCKET_PATH" default:""`
	SidecarEvaluator     string `envconfig:"SIDECAR_EVALUATOR" default:"json"`
	SidecarProviderArgs  string `envconfig:"SIDECAR_PROVIDER_ARGS" default:""`
	SidecarSyncProvider  string `envconfig:"SIDECAR_SYNC_PROVIDER" default:"kubernetes"`
	SidecarLogFormat     string `envconfig:"SIDECAR_LOG_FORMAT" default:"json"`
	SidecarProbesEnabled bool   `envconfig:"SIDECAR_PROBES_ENABLED" default:"true"`
}

func FeatureFlagSourceIndex(o client.Object) []string {
	deployment, ok := o.(*appsV1.Deployment)
	if !ok {
		return []string{
			"false",
		}
	}

	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		return []string{
			"false",
		}
	}
	if _, ok := deployment.Spec.Template.ObjectMeta.Annotations[fmt.Sprintf("openfeature.dev/%s", FeatureFlagSourceAnnotation)]; ok {
		return []string{
			"true",
		}
	}
	return []string{
		"false",
	}
}

func FindFlagConfig(ctx context.Context, c client.Client, namespace string, name string) (*api.FeatureFlag, error) {
	ffConfig := &api.FeatureFlag{}
	if err := c.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, ffConfig); err != nil {
		return nil, err
	}
	return ffConfig, nil
}

// SharedOwnership returns true if any of the owner references match in the given slices
func SharedOwnership(ownerReferences1, ownerReferences2 []metav1.OwnerReference) bool {
	for _, owner1 := range ownerReferences1 {
		for _, owner2 := range ownerReferences2 {
			if owner1.UID == owner2.UID {
				return true
			}
		}
	}
	return false
}
