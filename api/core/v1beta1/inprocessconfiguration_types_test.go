package v1beta1

import (
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func Test_InProcessConfiguration_Merge(t *testing.T) {
	ff_old := &InProcessConfiguration{
		Spec: InProcessConfigurationSpec{
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

	require.Equal(t, &InProcessConfiguration{
		Spec: InProcessConfigurationSpec{
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

	ff_new := &InProcessConfiguration{
		Spec: InProcessConfigurationSpec{
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

	require.Equal(t, ff_old.Spec.EnvVarPrefix, "PRE_SECOND")
	require.Equal(t, ff_old.Spec.Port, int32(33))
	require.Equal(t, ff_old.Spec.SocketPath, "socket-path")
	require.Equal(t, ff_old.Spec.Host, "host")
	require.Equal(t, ff_old.Spec.TLS, true)
	require.Equal(t, ff_old.Spec.OfflineFlagSourcePath, "path1")
	require.Equal(t, ff_old.Spec.Selector, "selector")
	require.Equal(t, ff_old.Spec.Cache, "cache")
	require.Equal(t, ff_old.Spec.CacheMaxSize, 12)
	require.Len(t, ff_old.Spec.EnvVars, 3)
	require.Contains(t, ff_old.Spec.EnvVars, v1.EnvVar{
		Name:  "env1",
		Value: "val1",
	})
	require.Contains(t, ff_old.Spec.EnvVars, v1.EnvVar{
		Name:  "env2",
		Value: "val2",
	})
	require.Contains(t, ff_old.Spec.EnvVars, v1.EnvVar{
		Name:  "env3",
		Value: "val3",
	})
}

func Test_InProcessConfiguration_ToEnvVars(t *testing.T) {
	ff := InProcessConfiguration{
		Spec: InProcessConfigurationSpec{
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
                    Name: "configMapKeyRef",
                    ValueFrom: &v1.EnvVarSource{
                        ConfigMapKeyRef: &v1.ConfigMapKeySelector{
                            LocalObjectReference: v1.LocalObjectReference{
                                Name: "configMapName",
                            },
                        },
                    },
                },
                {
                    Name: "fieldRef",
                    ValueFrom: &v1.EnvVarSource{
                        FieldRef: &v1.ObjectFieldSelector{
                            FieldPath: "fieldPath",
                        },
                    },
                },
                {
                    Name: "resourceFieldRef",
                    ValueFrom: &v1.EnvVarSource{
                        ResourceFieldRef: &v1.ResourceFieldSelector{
                            ContainerName: "containerName",
                            Resource:      "resourceField",
                        },
                    },
                },
                {
                    Name: "secretKeyRef",
                    ValueFrom: &v1.EnvVarSource{
                        SecretKeyRef: &v1.SecretKeySelector{
                            LocalObjectReference: v1.LocalObjectReference{
                                Name: "secretName",
                            },
                        },
                    },
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
            Name: "PRE_configMapKeyRef",
            ValueFrom: &v1.EnvVarSource{
                ConfigMapKeyRef: &v1.ConfigMapKeySelector{
                    LocalObjectReference: v1.LocalObjectReference{
                        Name: "configMapName",
                    },
                },
            },
        },
        {
            Name: "PRE_fieldRef",
            ValueFrom: &v1.EnvVarSource{
                FieldRef: &v1.ObjectFieldSelector{
                    FieldPath: "fieldPath",
                },
            },
        },
        {
            Name: "PRE_resourceFieldRef",
            ValueFrom: &v1.EnvVarSource{
                ResourceFieldRef: &v1.ResourceFieldSelector{
                    ContainerName: "containerName",
                    Resource:      "resourceField",
                },
            },
        },
        {
            Name: "PRE_secretKeyRef",
            ValueFrom: &v1.EnvVarSource{
                SecretKeyRef: &v1.SecretKeySelector{
                    LocalObjectReference: v1.LocalObjectReference{
                        Name: "secretName",
                    },
                },
            },
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
			Name:  "PRE_CACHE",
			Value: "cache",
		},
		{
			Name:  "PRE_MAX_CACHE_SIZE",
			Value: "12",
		},
		{
			Name:  "PRE_RESOLVER",
			Value: "in-process",
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
	}
	require.Equal(t, expected, ff.Spec.ToEnvVars())
}
