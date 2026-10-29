package webhook

import (
	"context"
	"errors"
	"fmt"
	"strings"

	api "github.com/open-feature/open-feature-operator/api/core/v1beta1"
	apicommon "github.com/open-feature/open-feature-operator/api/core/v1beta1/common"
	"github.com/open-feature/open-feature-operator/internal/common"
	"github.com/open-feature/open-feature-operator/internal/common/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func OpenFeatureEnabledAnnotationIndex(o client.Object) []string {
	pod, ok := o.(*corev1.Pod)
	if !ok {
		return []string{"false"}
	}
	if pod.ObjectMeta.Annotations == nil {
		return []string{
			"false",
		}
	}
	val, ok := pod.ObjectMeta.Annotations[fmt.Sprintf("openfeature.dev/%s", common.AllowKubernetesSyncAnnotation)]
	if ok && val == "true" {
		return []string{
			"true",
		}
	}
	return []string{
		"false",
	}
}

func parseList(s string) []string {
	out := []string{}
	ss := strings.Split(s, ",")
	for i := 0; i < len(ss); i++ {
		newS := strings.TrimSpace(ss[i])
		if newS != "" { //function should not add empty values
			out = append(out, newS)
		}
	}
	return out
}

func containsK8sProvider(sources []api.Source) bool {
	for _, source := range sources {
		if source.Provider.IsKubernetes() {
			return true
		}
	}
	return false
}

func checkOFEnabled(annotations map[string]string) bool {
	val, ok := annotations[fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.EnabledAnnotation)]
	return ok && val == "true"
}

func NewFeatureFlagSourceSpec(env types.EnvConfig) *api.FeatureFlagSourceSpec {
	f := false
	args := strings.Split(env.SidecarProviderArgs, ",")
	// use empty array when arguments are not set
	if len(args) == 1 && args[0] == "" {
		args = []string{}
	}
	return &api.FeatureFlagSourceSpec{
		ManagementPort:      ptr.To(int32(env.SidecarManagementPort)),
		Port:                ptr.To(int32(env.SidecarPort)),
		SocketPath:          ptr.To(env.SidecarSocketPath),
		Evaluator:           ptr.To(env.SidecarEvaluator),
		Sources:             []api.Source{},
		EnvVars:             []corev1.EnvVar{},
		SyncProviderArgs:    args,
		DefaultSyncProvider: ptr.To(apicommon.SyncProviderType(env.SidecarSyncProvider)),
		EnvVarPrefix:        ptr.To(env.SidecarEnvVarPrefix),
		LogFormat:           ptr.To(env.SidecarLogFormat),
		RolloutOnChange:     nil,
		DebugLogging:        &f,
		OtelCollectorUri:    ptr.To(""),
		ProbesEnabled:       &env.SidecarProbesEnabled,
		OFREPPort:           ptr.To(apicommon.DefaultOFREPPort),
	}
}

func NewInProcessConfigurationSpec(env types.EnvConfig) *api.InProcessConfigurationSpec {
	return &api.InProcessConfigurationSpec{
		Port:                  ptr.To(int32(env.InProcessPort)),
		SocketPath:            ptr.To(env.InProcessSocketPath),
		Host:                  ptr.To(env.InProcessHost),
		TLS:                   ptr.To(env.InProcessTLS),
		OfflineFlagSourcePath: ptr.To(env.InProcessOfflineFlagSourcePath),
		Selector:              ptr.To(env.InProcessSelector),
		Cache:                 ptr.To(env.InProcessCache),
		CacheMaxSize:          ptr.To(env.InProcessCacheMaxSize),
		EnvVarPrefix:          ptr.To(env.InProcessEnvVarPrefix),
	}
}

func shouldUseSidecar(annotations map[string]string) bool {
	_, ok := annotations[fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.FeatureFlagSourceAnnotation)]
	return ok
}

func shouldUseInProcess(annotations map[string]string) bool {
	_, ok := annotations[fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.InProcessConfigurationAnnotation)]
	return ok
}

func (m *PodMutator) getFeatureFlagSource(ctx context.Context, namespace string, name string) (*api.FeatureFlagSource, error) {
	fcConfig := &api.FeatureFlagSource{}
	if err := m.Client.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, fcConfig); err != nil {
		return nil, err
	}
	return fcConfig, nil
}

func (m *PodMutator) getInProcessConfiguration(ctx context.Context, namespace string, name string) (*api.InProcessConfiguration, error) {
	fcConfig := &api.InProcessConfiguration{}
	if err := m.Client.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, fcConfig); err != nil {
		return nil, err
	}
	return fcConfig, nil
}

// ErrCrossNamespaceRef is returned when a namespaced annotation token references
// a resource in a namespace other than the pod's own. Cross-namespace resolution
// via namespaced annotations was removed in 0.10.0; use a ClusterFeatureFlagSource
// or ClusterInProcessConfiguration and the corresponding cluster-scoped annotation.
var ErrCrossNamespaceRef = errors.New("cross-namespace source reference is not permitted; create a ClusterFeatureFlagSource or ClusterInProcessConfiguration and use the corresponding cluster-scoped annotation instead")

// resolveNamespacedRef parses a single annotation token intended for a namespaced
// resource reference. Accepted forms:
//   - bare name -> (podNS, name, nil)
//   - <podNS>/<name> -> (podNS, name, nil), kept for back-compat with pre-0.10.0 manifests
//
// Any <other-ns>/<name> form is rejected with ErrCrossNamespaceRef.
// An empty token returns (podNS, "", nil); downstream client.Get will fail on the
// empty name. This matches the historical behavior of utils.ParseAnnotation.
func resolveNamespacedRef(token, podNS string) (namespace, name string, err error) {
	parts := strings.SplitN(token, "/", 2)
	if len(parts) == 1 {
		return podNS, parts[0], nil
	}
	if parts[0] != podNS {
		return "", "", fmt.Errorf("%w: %q", ErrCrossNamespaceRef, token)
	}
	return podNS, parts[1], nil
}

func shouldUseClusterFeatureFlagSource(annotations map[string]string) bool {
	_, ok := annotations[fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.ClusterFeatureFlagSourceAnnotation)]
	return ok
}

func shouldUseClusterInProcessConfiguration(annotations map[string]string) bool {
	_, ok := annotations[fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.ClusterInProcessConfigurationAnnotation)]
	return ok
}

func (m *PodMutator) getClusterFeatureFlagSource(ctx context.Context, name string) (*api.ClusterFeatureFlagSource, error) {
	cffs := &api.ClusterFeatureFlagSource{}
	if err := m.Client.Get(ctx, client.ObjectKey{Name: name}, cffs); err != nil {
		return nil, err
	}
	return cffs, nil
}

func (m *PodMutator) getClusterInProcessConfiguration(ctx context.Context, name string) (*api.ClusterInProcessConfiguration, error) {
	cipc := &api.ClusterInProcessConfiguration{}
	if err := m.Client.Get(ctx, client.ObjectKey{Name: name}, cipc); err != nil {
		return nil, err
	}
	return cipc, nil
}
