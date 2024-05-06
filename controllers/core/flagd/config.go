package flagd

import (
	"github.com/open-feature/open-feature-operator/common"
	"github.com/open-feature/open-feature-operator/common/types"
)

type FlagdConfiguration struct {
	FlagdPort      int
	OFREPPort      int
	ManagementPort int
	DebugLogging   bool
	Image          string
	Tag            string

	OperatorNamespace      string
	OperatorDeploymentName string
}

func NewFlagdConfiguration(env types.EnvConfig) FlagdConfiguration {
	return FlagdConfiguration{
		Image:                  env.FlagdImage,
		Tag:                    env.FlagdTag,
		OperatorDeploymentName: common.OperatorDeploymentName,
		FlagdPort:              env.FlagdPort,
		OFREPPort:              env.FlagdOFREPPort,
		ManagementPort:         env.FlagdManagementPort,
		DebugLogging:           env.FlagdDebugLogging,
	}
}
