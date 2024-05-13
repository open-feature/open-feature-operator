package v1beta1

import (
	"testing"

	"github.com/open-feature/open-feature-operator/apis/core/v1beta1/common"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func Test_FLagSourceConfiguration_Merge(t *testing.T) {
	ff_old := &FeatureFlagSource{
		Spec: FeatureFlagSourceSpec{
			EnvVars: []v1.EnvVar{
				{
					Name:  "env1",
					Value: "val1",
				},
				{
					Name:  "env2",
					Value: "val2",
				},
			},
			EnvVarPrefix:   "PRE",
			ManagementPort: 22,
			Port:           33,
			Evaluator:      "evaluator",
			SocketPath:     "socket-path",
			LogFormat:      "log",
			Sources: []Source{
				{
					Source:     "src1",
					Provider:   common.SyncProviderGrpc,
					TLS:        true,
					CertPath:   "etc/cert.ca",
					ProviderID: "app",
					Selector:   "source=database",
					Interval:   5,
				},
			},
			SyncProviderArgs:    []string{"arg1", "arg2"},
			DefaultSyncProvider: common.SyncProviderKubernetes,
			RolloutOnChange:     common.TrueVal(),
			ProbesEnabled:       common.TrueVal(),
			DebugLogging:        common.TrueVal(),
			OtelCollectorUri:    "",
		},
	}

	ff_old.Spec.Merge(nil)

	require.Equal(t, &FeatureFlagSource{
		Spec: FeatureFlagSourceSpec{
			EnvVars: []v1.EnvVar{
				{
					Name:  "env1",
					Value: "val1",
				},
				{
					Name:  "env2",
					Value: "val2",
				},
			},
			EnvVarPrefix:   "PRE",
			ManagementPort: 22,
			Port:           33,
			Evaluator:      "evaluator",
			SocketPath:     "socket-path",
			LogFormat:      "log",
			Sources: []Source{
				{
					Source:     "src1",
					Provider:   common.SyncProviderGrpc,
					TLS:        true,
					CertPath:   "etc/cert.ca",
					ProviderID: "app",
					Selector:   "source=database",
					Interval:   5,
				},
			},
			SyncProviderArgs:    []string{"arg1", "arg2"},
			DefaultSyncProvider: common.SyncProviderKubernetes,
			RolloutOnChange:     common.TrueVal(),
			ProbesEnabled:       common.TrueVal(),
			DebugLogging:        common.TrueVal(),
			OtelCollectorUri:    "",
		},
	}, ff_old)

	ff_new := &FeatureFlagSource{
		Spec: FeatureFlagSourceSpec{
			EnvVars: []v1.EnvVar{
				{
					Name:  "env3",
					Value: "val3",
				},
				{
					Name:  "env4",
					Value: "val4",
				},
			},
			EnvVarPrefix:   "PREFIX",
			ManagementPort: 221,
			Port:           331,
			Evaluator:      "evaluator1",
			SocketPath:     "socket-path1",
			LogFormat:      "log1",
			Sources: []Source{
				{
					Source:   "src2",
					Provider: common.SyncProviderFilepath,
				},
			},
			SyncProviderArgs:    []string{"arg3", "arg4"},
			DefaultSyncProvider: common.SyncProviderFilepath,
			RolloutOnChange:     common.FalseVal(),
			ProbesEnabled:       common.FalseVal(),
			DebugLogging:        common.FalseVal(),
			OtelCollectorUri:    "",
		},
	}

	ff_old.Spec.Merge(&ff_new.Spec)

	require.Equal(t, &FeatureFlagSource{
		Spec: FeatureFlagSourceSpec{
			EnvVars: []v1.EnvVar{
				{
					Name:  "env1",
					Value: "val1",
				},
				{
					Name:  "env2",
					Value: "val2",
				},
				{
					Name:  "env3",
					Value: "val3",
				},
				{
					Name:  "env4",
					Value: "val4",
				},
			},
			EnvVarPrefix:   "PREFIX",
			ManagementPort: 221,
			Port:           331,
			Evaluator:      "evaluator1",
			SocketPath:     "socket-path1",
			LogFormat:      "log1",
			Sources: []Source{
				{
					Source:     "src1",
					Provider:   common.SyncProviderGrpc,
					TLS:        true,
					CertPath:   "etc/cert.ca",
					ProviderID: "app",
					Selector:   "source=database",
					Interval:   5,
				},
				{
					Source:   "src2",
					Provider: common.SyncProviderFilepath,
				},
			},
			SyncProviderArgs:    []string{"arg1", "arg2", "arg3", "arg4"},
			DefaultSyncProvider: common.SyncProviderFilepath,
			RolloutOnChange:     common.FalseVal(),
			ProbesEnabled:       common.FalseVal(),
			DebugLogging:        common.FalseVal(),
			OtelCollectorUri:    "",
		},
	}, ff_old)
}

func Test_FLagSourceConfiguration_ToEnvVars(t *testing.T) {
	ff := FeatureFlagSource{
		Spec: FeatureFlagSourceSpec{
			EnvVars: []v1.EnvVar{
				{
					Name:  "env1",
					Value: "val1",
				},
				{
					Name:  "env2",
					Value: "val2",
				},
			},
			EnvVarPrefix:   "PRE",
			ManagementPort: 22,
			Port:           33,
			Evaluator:      "evaluator",
			SocketPath:     "socket-path",
			LogFormat:      "log",
		},
	}
	expected := []v1.EnvVar{
		{
			Name:  "PRE_env1",
			Value: "val1",
		},
		{
			Name:  "PRE_env2",
			Value: "val2",
		},
		{
			Name:  "PRE_MANAGEMENT_PORT",
			Value: "22",
		},
		{
			Name:  "PRE_PORT",
			Value: "33",
		},
		{
			Name:  "PRE_EVALUATOR",
			Value: "evaluator",
		},
		{
			Name:  "PRE_SOCKET_PATH",
			Value: "socket-path",
		},
		{
			Name:  "PRE_LOG_FORMAT",
			Value: "log",
		},
	}
	require.Equal(t, expected, ff.Spec.ToEnvVars())
}
