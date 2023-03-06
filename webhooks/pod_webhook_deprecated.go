package webhooks

import (
	"context"
	"fmt"
	"reflect"

	api "github.com/open-feature/open-feature-operator/apis/core/v1alpha3"
	apicommon "github.com/open-feature/open-feature-operator/apis/core/v1alpha3/common"
)

func (m *PodMutator) handleFeatureFlagConfigurationAnnotation(ctx context.Context, fcConfig *api.FlagSourceConfigurationSpec, ffconfigAnnotation string, defaultNamespace string) error {
	for _, ffName := range parseList(ffconfigAnnotation) {
		ns, name := parseAnnotation(ffName, defaultNamespace)
		fsConfig := m.getFeatureFlag(ctx, ns, name)
		if reflect.DeepEqual(fsConfig, api.FeatureFlagConfiguration{}) {
			return fmt.Errorf("FeatureFlagConfiguration %s not found", ffName)
		}
		if fsConfig.Spec.FlagDSpec != nil {
			if len(fsConfig.Spec.FlagDSpec.Envs) > 0 {
				fcConfig.EnvVars = append(fsConfig.Spec.FlagDSpec.Envs, fcConfig.EnvVars...)
			}
			if fsConfig.Spec.FlagDSpec.MetricsPort != 0 && fcConfig.MetricsPort == apicommon.DefaultMetricPort {
				fcConfig.MetricsPort = fsConfig.Spec.FlagDSpec.MetricsPort
			}
		}
		switch {
		case fsConfig.Spec.SyncProvider == nil:
			fcConfig.Sources = append(fcConfig.Sources, api.Source{
				Provider: fcConfig.DefaultSyncProvider,
				Source:   ffName,
			})
		case fsConfig.Spec.SyncProvider.Name == apicommon.SyncProviderKubernetes:
			fcConfig.Sources = append(fcConfig.Sources, api.Source{
				Provider: apicommon.SyncProviderKubernetes,
				Source:   ffName,
			})
		case fsConfig.Spec.SyncProvider.Name == apicommon.SyncProviderFilepath:
			fcConfig.Sources = append(fcConfig.Sources, api.Source{
				Provider: apicommon.SyncProviderFilepath,
				Source:   ffName,
			})
		case fsConfig.Spec.SyncProvider.Name == apicommon.SyncProviderHttp:
			if fsConfig.Spec.SyncProvider.HttpSyncConfiguration == nil {
				return fmt.Errorf("FeatureFlagConfiguration %s is missing HttpSyncConfiguration", ffName)
			}
			fcConfig.Sources = append(fcConfig.Sources, api.Source{
				Provider:            apicommon.SyncProviderHttp,
				Source:              fsConfig.Spec.SyncProvider.HttpSyncConfiguration.Target,
				HttpSyncBearerToken: fsConfig.Spec.SyncProvider.HttpSyncConfiguration.BearerToken,
			})
		default:
			return fmt.Errorf("FeatureFlagConfiguration %s configuration is unrecognized", ffName)
		}
	}
	return nil
}
