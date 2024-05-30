package flagd

import (
	"github.com/open-feature/open-feature-operator/common"
	"github.com/open-feature/open-feature-operator/common/types"
	resources "github.com/open-feature/open-feature-operator/controllers/core/flagd/common"
)

func NewFlagdConfiguration(env types.EnvConfig, imagePullSecret string) resources.FlagdConfiguration {
	return resources.FlagdConfiguration{
		Image:                  env.FlagdImage,
		Tag:                    env.FlagdTag,
		OperatorDeploymentName: common.OperatorDeploymentName,
		FlagdPort:              env.FlagdPort,
		OFREPPort:              env.FlagdOFREPPort,
		SyncPort:               env.FlagdSyncPort,
		ManagementPort:         env.FlagdManagementPort,
		DebugLogging:           env.FlagdDebugLogging,
		ImagePullSecret:        imagePullSecret,
	}
}
