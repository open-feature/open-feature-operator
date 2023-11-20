package webhooks

import (
	"fmt"
	"strings"

	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	apicommon "github.com/open-feature/open-feature-operator/apis/core/v1beta1/common"
	"github.com/open-feature/open-feature-operator/common"
	"github.com/open-feature/open-feature-operator/common/constant"
	corev1 "k8s.io/api/core/v1"
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
	val, ok := pod.ObjectMeta.Annotations[fmt.Sprintf("openfeature.dev/%s", constant.AllowKubernetesSyncAnnotation)]
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
	val, ok := annotations[fmt.Sprintf("%s/%s", constant.OpenFeatureAnnotationPrefix, constant.EnabledAnnotation)]
	return ok && val == "true"
}

func NewFeatureFlagSourceSpec(env common.EnvConfig) *api.FeatureFlagSourceSpec {
	f := false
	args := strings.Split(env.SidecarProviderArgs, ",")
	// use empty array when arguments are not set
	if len(args) == 1 && args[0] == "" {
		args = []string{}
	}
	return &api.FeatureFlagSourceSpec{
		ManagementPort:      int32(env.SidecarManagementPort),
		Port:                int32(env.SidecarPort),
		SocketPath:          env.SidecarSocketPath,
		Evaluator:           env.SidecarEvaluator,
		Sources:             []api.Source{},
		EnvVars:             []corev1.EnvVar{},
		SyncProviderArgs:    args,
		DefaultSyncProvider: apicommon.SyncProviderType(env.SidecarSyncProvider),
		EnvVarPrefix:        env.SidecarEnvVarPrefix,
		LogFormat:           env.SidecarLogFormat,
		RolloutOnChange:     nil,
		DebugLogging:        &f,
		OtelCollectorUri:    "",
		ProbesEnabled:       &env.SidecarProbesEnabled,
	}
}
