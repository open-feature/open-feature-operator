package v1beta1

import (
	"fmt"

	"github.com/open-feature/open-feature-operator/apis/core/v1beta1/common"
	"github.com/open-feature/open-feature-operator/apis/core/v1beta2"
	v1beta2common "github.com/open-feature/open-feature-operator/apis/core/v1beta2/common"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts the src v1beta1.FeatureFlagSource to the hub version (v1beta1.FeatureFlagSource)
//
//nolint:gocyclo
func (src *FeatureFlagSource) ConvertTo(dstRaw conversion.Hub) error {
	dst, ok := dstRaw.(*v1beta2.FeatureFlagSource)

	if !ok {
		return fmt.Errorf("type %T %s", dstRaw, "unable to convert to v1beta2.FeatureFlagSource")
	}

	// Copy equal stuff to new object
	dst.ObjectMeta = src.ObjectMeta

	dst.Spec.EnvVarPrefix = src.Spec.EnvVarPrefix
	dst.Spec.RPC = &v1beta2.RPCConf{
		ManagementPort:      src.Spec.ManagementPort,
		Port:                src.Spec.Port,
		SocketPath:          src.Spec.SocketPath,
		Evaluator:           src.Spec.Evaluator,
		DefaultSyncProvider: v1beta2common.SyncProviderType(src.Spec.DefaultSyncProvider),
		LogFormat:           src.Spec.LogFormat,
		ProbesEnabled:       *src.Spec.ProbesEnabled,
		DebugLogging:        *src.Spec.DebugLogging,
		OtelCollectorUri:    src.Spec.OtelCollectorUri,
	}
	dst.Spec.RPC.Resources.Limits = src.Spec.Resources.Limits
	dst.Spec.RPC.Resources.Requests = src.Spec.Resources.Requests
	dst.Spec.RPC.Resources.Claims = make([]corev1.ResourceClaim, len(src.Spec.Resources.Claims))
	copy(dst.Spec.RPC.Resources.Claims, src.Spec.Resources.Claims)
	dst.Spec.RPC.SyncProviderArgs = make([]string, len(src.Spec.SyncProviderArgs))
	copy(dst.Spec.RPC.SyncProviderArgs, src.Spec.SyncProviderArgs)
	dst.Spec.RPC.EnvVars = make([]corev1.EnvVar, len(src.Spec.EnvVars))
	copy(dst.Spec.RPC.EnvVars, src.Spec.EnvVars)
	dst.Spec.RPC.Sources = make([]v1beta2.Source, len(src.Spec.Sources))
	for idx, item := range src.Spec.Sources {
		dst.Spec.RPC.Sources[idx] = v1beta2.Source{
			Source:              item.Source,
			Provider:            v1beta2common.SyncProviderType(item.Provider),
			HttpSyncBearerToken: item.HttpSyncBearerToken,
			TLS:                 item.TLS,
			CertPath:            item.CertPath,
			ProviderID:          item.ProviderID,
			Selector:            item.Selector,
			Interval:            item.Interval,
		}
	}

	return nil
}

// ConvertFrom converts from the hub version (v1beta2.FeatureFlagSource) to this version (v1beta1.FeatureFlagSource)
//
//nolint:gocyclo
func (dst *FeatureFlagSource) ConvertFrom(srcRaw conversion.Hub) error {
	src, ok := srcRaw.(*v1beta2.FeatureFlagSource)

	if !ok {
		return fmt.Errorf("type %T %s", srcRaw, "unable to convert from v1beta2.FeatureFlagSource")
	}

	// Copy equal stuff to new object
	dst.ObjectMeta = src.ObjectMeta

	dst.Spec.EnvVarPrefix = src.Spec.EnvVarPrefix
	dst.Spec.ManagementPort = src.Spec.RPC.ManagementPort
	dst.Spec.Port = src.Spec.RPC.Port
	dst.Spec.SocketPath = src.Spec.RPC.SocketPath
	dst.Spec.Evaluator = src.Spec.RPC.Evaluator
	dst.Spec.DefaultSyncProvider = common.SyncProviderType(src.Spec.RPC.DefaultSyncProvider)
	dst.Spec.LogFormat = src.Spec.RPC.LogFormat
	dst.Spec.ProbesEnabled = &src.Spec.RPC.ProbesEnabled
	dst.Spec.DebugLogging = &src.Spec.RPC.DebugLogging
	dst.Spec.OtelCollectorUri = src.Spec.RPC.OtelCollectorUri
	dst.Spec.Resources.Limits = src.Spec.RPC.Resources.Limits
	dst.Spec.Resources.Requests = src.Spec.RPC.Resources.Requests
	dst.Spec.Resources.Claims = make([]corev1.ResourceClaim, len(src.Spec.RPC.Resources.Claims))
	copy(dst.Spec.Resources.Claims, src.Spec.RPC.Resources.Claims)
	dst.Spec.SyncProviderArgs = make([]string, len(src.Spec.RPC.SyncProviderArgs))
	copy(dst.Spec.SyncProviderArgs, src.Spec.RPC.SyncProviderArgs)
	dst.Spec.EnvVars = make([]corev1.EnvVar, len(src.Spec.RPC.EnvVars))
	copy(dst.Spec.EnvVars, src.Spec.RPC.EnvVars)
	dst.Spec.Sources = make([]Source, len(src.Spec.RPC.Sources))
	for idx, item := range src.Spec.RPC.Sources {
		dst.Spec.Sources[idx] = Source{
			Source:              item.Source,
			Provider:            common.SyncProviderType(item.Provider),
			HttpSyncBearerToken: item.HttpSyncBearerToken,
			TLS:                 item.TLS,
			CertPath:            item.CertPath,
			ProviderID:          item.ProviderID,
			Selector:            item.Selector,
			Interval:            item.Interval,
		}
	}

	return nil
}
