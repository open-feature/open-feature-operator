package v1alpha1

import (
	"testing"

	"github.com/open-feature/open-feature-operator/pkg/utils"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func Test_FLagSourceConfiguration_SyncProvider(t *testing.T) {
	k := SyncProviderKubernetes
	f := SyncProviderFilepath
	h := SyncProviderHttp

	require.True(t, k.IsKubernetes())
	require.True(t, f.IsFilepath())
	require.True(t, h.IsHttp())

	require.False(t, f.IsKubernetes())
	require.False(t, h.IsFilepath())
	require.False(t, k.IsHttp())
}

func Test_FLagSourceConfiguration_envVarKey(t *testing.T) {
	require.Equal(t, "pre_suf", envVarKey("pre", "suf"))
}

func Test_FLagSourceConfiguration_ToEnvVars(t *testing.T) {
	ff := FlagSourceConfiguration{
		Spec: FlagSourceConfigurationSpec{
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
			EnvVarPrefix: "PRE",
			MetricsPort:  22,
			Port:         33,
			Evaluator:    "evaluator",
			SocketPath:   "socket-path",
			LogFormat:    "log",
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
			Name:  "PRE_METRICS_PORT",
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

func Test_FLagSourceConfiguration_Merge(t *testing.T) {
	ff_old := &FlagSourceConfiguration{
		Spec: FlagSourceConfigurationSpec{
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
			EnvVarPrefix: "PRE",
			MetricsPort:  22,
			Port:         33,
			Evaluator:    "evaluator",
			SocketPath:   "socket-path",
			LogFormat:    "log",
			Image:        "img",
			Tag:          "tag",
			Sources: []Source{
				{
					Source:              "src1",
					Provider:            SyncProviderKubernetes,
					HttpSyncBearerToken: "token1",
					LogFormat:           "log1",
				},
			},
			SyncProviderArgs:    []string{"arg1", "arg2"},
			DefaultSyncProvider: SyncProviderKubernetes,
			RolloutOnChange:     utils.TrueVal(),
			ProbesEnabled:       utils.TrueVal(),
		},
	}

	ff_old.Spec.Merge(nil)

	require.Equal(t, &FlagSourceConfiguration{
		Spec: FlagSourceConfigurationSpec{
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
			EnvVarPrefix: "PRE",
			MetricsPort:  22,
			Port:         33,
			Evaluator:    "evaluator",
			SocketPath:   "socket-path",
			LogFormat:    "log",
			Image:        "img",
			Tag:          "tag",
			Sources: []Source{
				{
					Source:              "src1",
					Provider:            SyncProviderKubernetes,
					HttpSyncBearerToken: "token1",
					LogFormat:           "log1",
				},
			},
			SyncProviderArgs:    []string{"arg1", "arg2"},
			DefaultSyncProvider: SyncProviderKubernetes,
			RolloutOnChange:     utils.TrueVal(),
			ProbesEnabled:       utils.TrueVal(),
		},
	}, ff_old)

	ff_new := &FlagSourceConfiguration{
		Spec: FlagSourceConfigurationSpec{
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
			EnvVarPrefix: "PREFIX",
			MetricsPort:  221,
			Port:         331,
			Evaluator:    "evaluator1",
			SocketPath:   "socket-path1",
			LogFormat:    "log1",
			Image:        "img1",
			Tag:          "tag1",
			Sources: []Source{
				{
					Source:              "src2",
					Provider:            SyncProviderFilepath,
					HttpSyncBearerToken: "token2",
					LogFormat:           "log2",
				},
			},
			SyncProviderArgs:    []string{"arg3", "arg4"},
			DefaultSyncProvider: SyncProviderFilepath,
			RolloutOnChange:     utils.FalseVal(),
			ProbesEnabled:       utils.FalseVal(),
		},
	}

	ff_old.Spec.Merge(&ff_new.Spec)

	require.Equal(t, &FlagSourceConfiguration{
		Spec: FlagSourceConfigurationSpec{
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
			EnvVarPrefix: "PREFIX",
			MetricsPort:  221,
			Port:         331,
			Evaluator:    "evaluator1",
			SocketPath:   "socket-path1",
			LogFormat:    "log1",
			Image:        "img1",
			Tag:          "tag1",
			Sources: []Source{
				{
					Source:              "src1",
					Provider:            SyncProviderKubernetes,
					HttpSyncBearerToken: "token1",
					LogFormat:           "log1",
				},
				{
					Source:              "src2",
					Provider:            SyncProviderFilepath,
					HttpSyncBearerToken: "token2",
					LogFormat:           "log2",
				},
			},
			SyncProviderArgs:    []string{"arg1", "arg2", "arg3", "arg4"},
			DefaultSyncProvider: SyncProviderFilepath,
			RolloutOnChange:     utils.FalseVal(),
			ProbesEnabled:       utils.FalseVal(),
		},
	}, ff_old)
}

func Test_FLagSourceConfiguration_NewFlagSourceConfigurationSpec(t *testing.T) {
	//happy path
	t.Setenv(envVarKey(InputConfigurationEnvVarPrefix, SidecarMetricPortEnvVar), "22")
	t.Setenv(envVarKey(InputConfigurationEnvVarPrefix, SidecarPortEnvVar), "33")
	t.Setenv(envVarKey(InputConfigurationEnvVarPrefix, SidecarSocketPathEnvVar), "val1")
	t.Setenv(envVarKey(InputConfigurationEnvVarPrefix, SidecarEvaluatorEnvVar), "val2")
	t.Setenv(envVarKey(InputConfigurationEnvVarPrefix, SidecarImageEnvVar), "val3")
	t.Setenv(envVarKey(InputConfigurationEnvVarPrefix, SidecarVersionEnvVar), "val4")
	t.Setenv(envVarKey(InputConfigurationEnvVarPrefix, SidecarProviderArgsEnvVar), "val11,val22")
	t.Setenv(envVarKey(InputConfigurationEnvVarPrefix, SidecarDefaultSyncProviderEnvVar), "kubernetes")
	t.Setenv(envVarKey(InputConfigurationEnvVarPrefix, SidecarLogFormatEnvVar), "val5")
	t.Setenv(SidecarEnvVarPrefix, "val6")
	t.Setenv(envVarKey(InputConfigurationEnvVarPrefix, SidecarProbesEnabledVar), "true")

	fs, err := NewFlagSourceConfigurationSpec()

	require.Nil(t, err)
	require.Equal(t, &FlagSourceConfigurationSpec{
		MetricsPort:         22,
		Port:                33,
		SocketPath:          "val1",
		Evaluator:           "val2",
		Image:               "val3",
		Tag:                 "val4",
		Sources:             []Source{},
		EnvVars:             []v1.EnvVar{},
		SyncProviderArgs:    []string{"val11", "val22"},
		DefaultSyncProvider: SyncProviderKubernetes,
		EnvVarPrefix:        "val6",
		LogFormat:           "val5",
		ProbesEnabled:       utils.TrueVal(),
	}, fs)

	//error paths
	t.Setenv(envVarKey(InputConfigurationEnvVarPrefix, SidecarProbesEnabledVar), "blah")
	_, err = NewFlagSourceConfigurationSpec()
	require.NotNil(t, err)

	t.Setenv(envVarKey(InputConfigurationEnvVarPrefix, SidecarPortEnvVar), "blah")
	_, err = NewFlagSourceConfigurationSpec()
	require.NotNil(t, err)

	t.Setenv(envVarKey(InputConfigurationEnvVarPrefix, SidecarMetricPortEnvVar), "blah")
	_, err = NewFlagSourceConfigurationSpec()
	require.NotNil(t, err)
}
