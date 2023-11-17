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
			Image:          "img",
			Tag:            "tag",
			Sources: []Source{
				{
					Source:     "src1",
					Provider:   common.SyncProviderGrpc,
					TLS:        true,
					CertPath:   "etc/cert.ca",
					ProviderID: "app",
					Selector:   "source=database",
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
			Image:          "img",
			Tag:            "tag",
			Sources: []Source{
				{
					Source:     "src1",
					Provider:   common.SyncProviderGrpc,
					TLS:        true,
					CertPath:   "etc/cert.ca",
					ProviderID: "app",
					Selector:   "source=database",
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
			Image:          "img1",
			Tag:            "tag1",
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
			Image:          "img1",
			Tag:            "tag1",
			Sources: []Source{
				{
					Source:     "src1",
					Provider:   common.SyncProviderGrpc,
					TLS:        true,
					CertPath:   "etc/cert.ca",
					ProviderID: "app",
					Selector:   "source=database",
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

func Test_FLagSourceConfiguration_NewFeatureFlagSourceSpec(t *testing.T) {
	//happy path
	t.Setenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarMetricPortEnvVar), "22")
	t.Setenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarPortEnvVar), "33")
	t.Setenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarSocketPathEnvVar), "val1")
	t.Setenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarEvaluatorEnvVar), "val2")
	t.Setenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarImageEnvVar), "val3")
	t.Setenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarVersionEnvVar), "val4")
	t.Setenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarProviderArgsEnvVar), "val11,val22")
	t.Setenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarDefaultSyncProviderEnvVar), "kubernetes")
	t.Setenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarLogFormatEnvVar), "val5")
	t.Setenv(SidecarEnvVarPrefix, "val6")
	t.Setenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarProbesEnabledVar), "true")

	fs, err := NewFeatureFlagSourceSpec()

	require.Nil(t, err)
	require.Equal(t, &FeatureFlagSourceSpec{
		ManagementPort:      22,
		Port:                33,
		SocketPath:          "val1",
		Evaluator:           "val2",
		Image:               "val3",
		Tag:                 "val4",
		Sources:             []Source{},
		EnvVars:             []v1.EnvVar{},
		SyncProviderArgs:    []string{"val11", "val22"},
		DefaultSyncProvider: common.SyncProviderKubernetes,
		EnvVarPrefix:        "val6",
		LogFormat:           "val5",
		ProbesEnabled:       common.TrueVal(),
		DebugLogging:        common.FalseVal(),
		OtelCollectorUri:    "",
	}, fs)

	//error paths
	t.Setenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarProbesEnabledVar), "blah")
	_, err = NewFeatureFlagSourceSpec()
	require.NotNil(t, err)

	t.Setenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarPortEnvVar), "blah")
	_, err = NewFeatureFlagSourceSpec()
	require.NotNil(t, err)

	t.Setenv(common.EnvVarKey(InputConfigurationEnvVarPrefix, SidecarMetricPortEnvVar), "blah")
	_, err = NewFeatureFlagSourceSpec()
	require.NotNil(t, err)
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
