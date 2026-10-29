package flagd

import (
	"github.com/open-feature/open-feature-operator/internal/common"
	"github.com/open-feature/open-feature-operator/internal/common/types"
	resources "github.com/open-feature/open-feature-operator/internal/controller/core/flagd/common"
	corev1 "k8s.io/api/core/v1"
)

func NewFlagdConfiguration(env types.EnvConfig, imagePullSecrets []string, labels map[string]string, annotations map[string]string) resources.FlagdConfiguration {
	return resources.FlagdConfiguration{
		Image:                  env.FlagdImage,
		Tag:                    env.FlagdTag,
		ImagePullPolicy:        corev1.PullPolicy(env.FlagdImagePullPolicy),
		OperatorDeploymentName: common.OperatorDeploymentName,
		FlagdPort:              env.FlagdPort,
		OFREPPort:              env.FlagdOFREPPort,
		SyncPort:               env.FlagdSyncPort,
		ManagementPort:         env.FlagdManagementPort,
		DebugLogging:           env.FlagdDebugLogging,
		ImagePullSecrets:       imagePullSecrets,
		Labels:                 labels,
		Annotations:            annotations,
	}
}
