package flagd

import (
	"github.com/open-feature/open-feature-operator/common"
	"github.com/open-feature/open-feature-operator/common/types"
	resources "github.com/open-feature/open-feature-operator/controllers/core/flagd/common"
)

func NewFlagdConfiguration(env types.EnvConfig, imagePullSecrets []string, labels map[string]string, annotations map[string]string) resources.FlagdConfiguration {
	return resources.FlagdConfiguration{
		Image:                  env.FlagdImage,
		Tag:                    env.FlagdTag,
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
