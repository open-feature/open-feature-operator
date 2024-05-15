package v1beta1

import (
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func Test_FeatureFlagInProcessConfiguration_Merge(t *testing.T) {
	ff_old := &FeatureFlagInProcessConfiguration{
		Spec: FeatureFlagInProcessConfigurationSpec{
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
			EnvVarPrefix:          "PRE",
			Port:                  33,
			SocketPath:            "socket-path",
			Host:                  "host",
			TLS:                   true,
			OfflineFlagSourcePath: "path1",
			Selector:              "selector",
			Cache:                 "cache",
			CacheMaxSize:          12,
		},
	}

	ff_old.Spec.Merge(nil)

	require.Equal(t, &FeatureFlagInProcessConfiguration{
		Spec: FeatureFlagInProcessConfigurationSpec{
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
			EnvVarPrefix:          "PRE",
			Port:                  33,
			SocketPath:            "socket-path",
			Host:                  "host",
			TLS:                   true,
			OfflineFlagSourcePath: "path1",
			Selector:              "selector",
			Cache:                 "cache",
			CacheMaxSize:          12,
		},
	}, ff_old)

	ff_new := &FeatureFlagInProcessConfiguration{
		Spec: FeatureFlagInProcessConfigurationSpec{
			EnvVars: []v1.EnvVar{
				{
					Name:  "env3",
					Value: "val3",
				},
			},
			EnvVarPrefix:          "PRE_SECOND",
			Port:                  33,
			SocketPath:            "",
			Host:                  "host",
			TLS:                   true,
			OfflineFlagSourcePath: "",
			Selector:              "",
			Cache:                 "lru",
			CacheMaxSize:          1000,
		},
	}

	ff_old.Spec.Merge(&ff_new.Spec)

	require.Equal(t, ff_old, &FeatureFlagInProcessConfiguration{
		Spec: FeatureFlagInProcessConfigurationSpec{
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
			},
			EnvVarPrefix:          "PRE_SECOND",
			Port:                  33,
			SocketPath:            "socket-path",
			Host:                  "host",
			TLS:                   true,
			OfflineFlagSourcePath: "path1",
			Selector:              "selector",
			Cache:                 "cache",
			CacheMaxSize:          12,
		},
	})
}

func Test_FeatureFlagInProcessConfiguration_ToEnvVars(t *testing.T) {
	ff := FeatureFlagInProcessConfiguration{
		Spec: FeatureFlagInProcessConfigurationSpec{
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
			EnvVarPrefix:          "PRE",
			Port:                  33,
			SocketPath:            "socket-path",
			Host:                  "host",
			TLS:                   true,
			OfflineFlagSourcePath: "path1",
			Selector:              "selector",
			Cache:                 "cache",
			CacheMaxSize:          12,
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
			Name:  "PRE_HOST",
			Value: "host",
		},
		{
			Name:  "PRE_PORT",
			Value: "33",
		},
		{
			Name:  "PRE_TLS",
			Value: "true",
		},
		{
			Name:  "PRE_SOCKET_PATH",
			Value: "socket-path",
		},
		{
			Name:  "PRE_OFFLINE_FLAG_SOURCE_PATH",
			Value: "path1",
		},
		{
			Name:  "PRE_SOURCE_SELECTOR",
			Value: "selector",
		},
		{
			Name:  "PRE_CACHE",
			Value: "cache",
		},
		{
			Name:  "PRE_MAX_CACHE_SIZE",
			Value: "12",
		},
	}
	require.Equal(t, expected, ff.Spec.ToEnvVars())
}
