package webhooks

import (
	"context"
	"fmt"
	"reflect"

	v1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
)

func (m *PodMutator) handleFeatureFlagConfigurationAnnotation(ctx context.Context, fcConfig *v1alpha1.FlagSourceConfigurationSpec, ffconfigAnnotation string, defaultNamespace string) error {
	for _, ffName := range parseList(ffconfigAnnotation) {
		ns, name := parseAnnotation(ffName, defaultNamespace)
		fsConfig := m.getFeatureFlag(ctx, ns, name)
		if reflect.DeepEqual(fsConfig, v1alpha1.FeatureFlagConfiguration{}) {
			return fmt.Errorf("FeatureFlagConfiguration %s not found", ffName)
		}
		if fsConfig.Spec.FlagDSpec != nil {
			if len(fsConfig.Spec.FlagDSpec.Envs) > 0 {
				fcConfig.EnvVars = append(fsConfig.Spec.FlagDSpec.Envs, fcConfig.EnvVars...)
			}
			if fsConfig.Spec.FlagDSpec.MetricsPort != 0 && fcConfig.MetricsPort == v1alpha1.DefaultMetricPort {
				fcConfig.MetricsPort = fsConfig.Spec.FlagDSpec.MetricsPort
			}
		}
		switch {
		case fsConfig.Spec.SyncProvider == nil:
			fcConfig.SyncProviders = append(fcConfig.SyncProviders, v1alpha1.SyncProvider{
				Provider: fcConfig.DefaultSyncProvider,
				Source:   ffName,
			})
		case v1alpha1.SyncProviderType(fsConfig.Spec.SyncProvider.Name).IsKubernetes():
			fcConfig.SyncProviders = append(fcConfig.SyncProviders, v1alpha1.SyncProvider{
				Provider: v1alpha1.SyncProviderKubernetes,
				Source:   ffName,
			})
		case v1alpha1.SyncProviderType(fsConfig.Spec.SyncProvider.Name).IsFilepath():
			fcConfig.SyncProviders = append(fcConfig.SyncProviders, v1alpha1.SyncProvider{
				Provider: v1alpha1.SyncProviderFilepath,
				Source:   ffName,
			})
		case v1alpha1.SyncProviderType(fsConfig.Spec.SyncProvider.Name).IsHttp():
			if fsConfig.Spec.SyncProvider.HttpSyncConfiguration == nil {
				return fmt.Errorf("FeatureFlagConfiguration %s is missing HttpSyncConfiguration", ffName)
			}
			fcConfig.SyncProviders = append(fcConfig.SyncProviders, v1alpha1.SyncProvider{
				Provider:            v1alpha1.SyncProviderHttp,
				Source:              fsConfig.Spec.SyncProvider.HttpSyncConfiguration.Target,
				HttpSyncBearerToken: fsConfig.Spec.SyncProvider.HttpSyncConfiguration.BearerToken,
			})
		default:
			return fmt.Errorf("FeatureFlagConfiguration %s configuration is unrecognized", ffName)
		}
	}
	return nil
}
