package v1beta1

import (
	"testing"

	"github.com/open-feature/open-feature-operator/api/core/v1beta1/common"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
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
			EnvVarPrefix:          ptr.To("PRE"),
			Port:                  ptr.To(int32(33)),
			SocketPath:            ptr.To("socket-path"),
			Host:                  ptr.To("host"),
			TLS:                   ptr.To(true),
			OfflineFlagSourcePath: ptr.To("path1"),
			Selector:              ptr.To("selector"),
			Cache:                 ptr.To("cache"),
			CacheMaxSize:          ptr.To(12),
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
			EnvVarPrefix:          ptr.To("PRE"),
			Port:                  ptr.To(int32(33)),
			SocketPath:            ptr.To("socket-path"),
			Host:                  ptr.To("host"),
			TLS:                   ptr.To(true),
			OfflineFlagSourcePath: ptr.To("path1"),
			Selector:              ptr.To("selector"),
			Cache:                 ptr.To("cache"),
			CacheMaxSize:          ptr.To(12),
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
			EnvVarPrefix: ptr.To("PRE_SECOND"),
			Selector:     ptr.To(""),
			Cache:        ptr.To("lru"),
			CacheMaxSize: ptr.To(1000),
		},
	}

	ff_old.Spec.Merge(&ff_new.Spec)

	require.Equal(t, "PRE_SECOND", *ff_old.Spec.EnvVarPrefix)
	require.Equal(t, int32(33), *ff_old.Spec.Port)
	require.Equal(t, "socket-path", *ff_old.Spec.SocketPath)
	require.Equal(t, "host", *ff_old.Spec.Host)
	require.True(t, *ff_old.Spec.TLS)
	require.Equal(t, "path1", *ff_old.Spec.OfflineFlagSourcePath)
	require.Equal(t, "", *ff_old.Spec.Selector)
	require.Equal(t, "lru", *ff_old.Spec.Cache)
	require.Equal(t, 1000, *ff_old.Spec.CacheMaxSize)
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

func Test_InProcessConfiguration_Merge_UnsetKeepsBase(t *testing.T) {
	base := &InProcessConfigurationSpec{
		Port:  ptr.To(int32(9999)),
		Host:  ptr.To("flagd.example"),
		TLS:   ptr.To(true),
		Cache: ptr.To("disabled"),
	}
	base.Merge(&InProcessConfigurationSpec{Selector: ptr.To("app=x")})
	require.Equal(t, int32(9999), *base.Port)
	require.Equal(t, "flagd.example", *base.Host)
	require.True(t, *base.TLS)
	require.Equal(t, "disabled", *base.Cache)
	require.Equal(t, "app=x", *base.Selector)
}

func Test_InProcessConfiguration_Merge_ExplicitDefaultOverrides(t *testing.T) {
	base := &InProcessConfigurationSpec{
		Port:         ptr.To(int32(9999)),
		Host:         ptr.To("flagd.example"),
		TLS:          ptr.To(true),
		Cache:        ptr.To("disabled"),
		CacheMaxSize: ptr.To(5),
		EnvVarPrefix: ptr.To("CUSTOM"),
	}
	base.Merge(&InProcessConfigurationSpec{
		Port:         ptr.To(common.DefaultInProcessPort),
		Host:         ptr.To(common.DefaultHost),
		TLS:          ptr.To(common.DefaultTLS),
		Cache:        ptr.To(common.DefaultCache),
		CacheMaxSize: ptr.To(int(common.DefaultCacheMaxSize)),
		EnvVarPrefix: ptr.To(common.DefaultEnvVarPrefix),
	})
	require.Equal(t, common.DefaultInProcessPort, *base.Port)
	require.Equal(t, common.DefaultHost, *base.Host)
	require.False(t, *base.TLS)
	require.Equal(t, common.DefaultCache, *base.Cache)
	require.Equal(t, int(common.DefaultCacheMaxSize), *base.CacheMaxSize)
	require.Equal(t, common.DefaultEnvVarPrefix, *base.EnvVarPrefix)
}

func Test_InProcessConfiguration_Merge_EnvVarsByName(t *testing.T) {
	base := &InProcessConfigurationSpec{EnvVars: []v1.EnvVar{
		{Name: "A", Value: "early"},
		{Name: "B", Value: "keep"},
	}}
	base.Merge(&InProcessConfigurationSpec{EnvVars: []v1.EnvVar{{Name: "A", Value: "late"}}})
	require.ElementsMatch(t, []v1.EnvVar{
		{Name: "A", Value: "late"},
		{Name: "B", Value: "keep"},
	}, base.EnvVars)
}

func Test_InProcessConfiguration_ToEnvVars_UnsetUsesDefaults(t *testing.T) {
	require.Equal(t, []v1.EnvVar{
		{Name: "FLAGD_HOST", Value: "localhost"},
		{Name: "FLAGD_PORT", Value: "8015"},
		{Name: "FLAGD_TLS", Value: "false"},
		{Name: "FLAGD_CACHE", Value: "lru"},
		{Name: "FLAGD_MAX_CACHE_SIZE", Value: "1000"},
		{Name: "FLAGD_RESOLVER", Value: "in-process"},
	}, (&InProcessConfigurationSpec{}).ToEnvVars())
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
			EnvVarPrefix:          ptr.To("PRE"),
			Port:                  ptr.To(int32(33)),
			SocketPath:            ptr.To("socket-path"),
			Host:                  ptr.To("host"),
			TLS:                   ptr.To(true),
			OfflineFlagSourcePath: ptr.To("path1"),
			Selector:              ptr.To("selector"),
			Cache:                 ptr.To("cache"),
			CacheMaxSize:          ptr.To(12),
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
