package v1beta1

import (
	"testing"
	"time"

	"github.com/open-feature/open-feature-operator/api/core/v1beta1/common"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
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
			EnvVarPrefix:   ptr.To("PRE"),
			ManagementPort: ptr.To(int32(22)),
			Port:           ptr.To(int32(33)),
			Evaluator:      ptr.To("evaluator"),
			SocketPath:     ptr.To("socket-path"),
			LogFormat:      ptr.To("log"),
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
			DefaultSyncProvider: ptr.To(common.SyncProviderKubernetes),
			RolloutOnChange:     common.TrueVal(),
			ProbesEnabled:       common.TrueVal(),
			DebugLogging:        common.TrueVal(),
			OtelCollectorUri:    ptr.To(""),
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU: resource.MustParse("100m"),
				},
			},
			ContextValues: map[string]string{
				"env": "staging",
			},
			HeaderToContextMappings: map[string]string{
				"X-Tenant": "tenant",
			},
			CORS:      []string{"http://localhost:3000"},
			OFREPPort: ptr.To(int32(8016)),
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
			EnvVarPrefix:   ptr.To("PRE"),
			ManagementPort: ptr.To(int32(22)),
			Port:           ptr.To(int32(33)),
			Evaluator:      ptr.To("evaluator"),
			SocketPath:     ptr.To("socket-path"),
			LogFormat:      ptr.To("log"),
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
			DefaultSyncProvider: ptr.To(common.SyncProviderKubernetes),
			RolloutOnChange:     common.TrueVal(),
			ProbesEnabled:       common.TrueVal(),
			DebugLogging:        common.TrueVal(),
			OtelCollectorUri:    ptr.To(""),
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU: resource.MustParse("100m"),
				},
			},
			ContextValues: map[string]string{
				"env": "staging",
			},
			HeaderToContextMappings: map[string]string{
				"X-Tenant": "tenant",
			},
			CORS:      []string{"http://localhost:3000"},
			OFREPPort: ptr.To(int32(8016)),
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
			EnvVarPrefix:   ptr.To("PREFIX"),
			ManagementPort: ptr.To(int32(221)),
			Port:           ptr.To(int32(331)),
			Evaluator:      ptr.To("evaluator1"),
			SocketPath:     ptr.To("socket-path1"),
			LogFormat:      ptr.To("log1"),
			Sources: []Source{
				{
					Source:   "src2",
					Provider: common.SyncProviderFilepath,
				},
			},
			SyncProviderArgs:    []string{"arg3", "arg4"},
			DefaultSyncProvider: ptr.To(common.SyncProviderFilepath),
			RolloutOnChange:     common.FalseVal(),
			ProbesEnabled:       common.FalseVal(),
			DebugLogging:        common.FalseVal(),
			OtelCollectorUri:    ptr.To(""),
			Resources: v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceCPU: resource.MustParse("200m"),
				},
			},
			ContextValues: map[string]string{
				"env":    "production",
				"region": "us-east-1",
			},
			HeaderToContextMappings: map[string]string{
				"X-Tenant": "tenant-override",
				"X-Region": "region",
			},
			CORS:      []string{"https://app.example.com", "https://admin.example.com"},
			OFREPPort: ptr.To(int32(9090)),
		},
	}

	ff_old.Spec.Merge(&ff_new.Spec)

	require.Equal(t, "PREFIX", *ff_old.Spec.EnvVarPrefix)
	require.Equal(t, int32(331), *ff_old.Spec.Port)
	require.Equal(t, int32(221), *ff_old.Spec.ManagementPort)
	require.Equal(t, "socket-path1", *ff_old.Spec.SocketPath)
	require.Equal(t, "evaluator1", *ff_old.Spec.Evaluator)
	require.Equal(t, "log1", *ff_old.Spec.LogFormat)
	require.Equal(t, ff_old.Spec.Sources, []Source{
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
	})
	require.Equal(t, ff_old.Spec.SyncProviderArgs, []string{"arg1", "arg2", "arg3", "arg4"})
	require.Equal(t, common.SyncProviderFilepath, *ff_old.Spec.DefaultSyncProvider)
	require.Equal(t, ff_old.Spec.RolloutOnChange, common.FalseVal())
	require.Equal(t, ff_old.Spec.ProbesEnabled, common.FalseVal())
	require.Equal(t, ff_old.Spec.DebugLogging, common.FalseVal())
	require.Equal(t, "", *ff_old.Spec.OtelCollectorUri)
	require.Equal(t, ff_old.Spec.Resources.Requests[v1.ResourceCPU], resource.MustParse("100m"))
	require.Equal(t, ff_old.Spec.Resources.Limits[v1.ResourceCPU], resource.MustParse("200m"))
	require.Len(t, ff_old.Spec.EnvVars, 4)
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
	require.Contains(t, ff_old.Spec.EnvVars, v1.EnvVar{
		Name:  "env4",
		Value: "val4",
	})

	// context values are merged additively, with new overriding old
	require.Equal(t, "production", ff_old.Spec.ContextValues["env"])
	require.Equal(t, "us-east-1", ff_old.Spec.ContextValues["region"])

	// header mappings are merged additively, with new overriding old
	require.Equal(t, "tenant-override", ff_old.Spec.HeaderToContextMappings["X-Tenant"])
	require.Equal(t, "region", ff_old.Spec.HeaderToContextMappings["X-Region"])

	// CORS is replaced entirely
	require.Equal(t, []string{"https://app.example.com", "https://admin.example.com"}, ff_old.Spec.CORS)

	// OFREPPort is overridden
	require.Equal(t, int32(9090), *ff_old.Spec.OFREPPort)
}

func Test_FLagSourceConfiguration_Merge_UnsetKeepsBase(t *testing.T) {
	base := &FeatureFlagSourceSpec{
		Port:          ptr.To(int32(9999)),
		LogFormat:     ptr.To("console"),
		ProbesEnabled: ptr.To(false),
		OFREPPort:     ptr.To(int32(9090)),
	}
	base.Merge(&FeatureFlagSourceSpec{Evaluator: ptr.To("json")})
	require.Equal(t, int32(9999), *base.Port)
	require.Equal(t, "console", *base.LogFormat)
	require.False(t, *base.ProbesEnabled)
	require.Equal(t, int32(9090), *base.OFREPPort)
	require.Equal(t, "json", *base.Evaluator)
}

func Test_FLagSourceConfiguration_Merge_ExplicitDefaultOverrides(t *testing.T) {
	base := &FeatureFlagSourceSpec{
		Port:          ptr.To(int32(9999)),
		LogFormat:     ptr.To("console"),
		ProbesEnabled: ptr.To(false),
		SocketPath:    ptr.To("/tmp/flagd.sock"),
		OFREPPort:     ptr.To(int32(9090)),
	}
	base.Merge(&FeatureFlagSourceSpec{
		Port:          ptr.To(common.DefaultRPCPort),
		LogFormat:     ptr.To(common.DefaultLogFormat),
		ProbesEnabled: ptr.To(true),
		SocketPath:    ptr.To(""),
		OFREPPort:     ptr.To(common.DefaultOFREPPort),
	})
	require.Equal(t, common.DefaultRPCPort, *base.Port)
	require.Equal(t, common.DefaultLogFormat, *base.LogFormat)
	require.True(t, *base.ProbesEnabled)
	require.Equal(t, "", *base.SocketPath)
	require.Equal(t, common.DefaultOFREPPort, *base.OFREPPort)
}

func Test_FLagSourceConfiguration_ToEnvVars_UnsetUsesDefaults(t *testing.T) {
	require.Equal(t, []v1.EnvVar{
		{Name: "FLAGD_MANAGEMENT_PORT", Value: "8014"},
		{Name: "FLAGD_PORT", Value: "8013"},
		{Name: "FLAGD_EVALUATOR", Value: "json"},
		{Name: "FLAGD_LOG_FORMAT", Value: "json"},
		{Name: "FLAGD_RESOLVER", Value: "rpc"},
	}, (&FeatureFlagSourceSpec{}).ToEnvVars())
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
				{
					Name:  "AZURE_STORAGE_ACCOUNT",
					Value: "account123",
				},
				{
					Name:  "AZURE_STORAGE_KEY",
					Value: "key456",
				},
				{
					Name:  "AWS_ACCESS_KEY_ID",
					Value: "AKIAIOSFODNN7EXAMPLE",
				},
				{
					Name:  "AWS_REGION",
					Value: "us-east-1",
				},
				{
					Name:  "GOOGLE_APPLICATION_CREDENTIALS",
					Value: "/var/run/secrets/gcp/key.json",
				},
			},
			EnvVarPrefix:   ptr.To("PRE"),
			ManagementPort: ptr.To(int32(22)),
			Port:           ptr.To(int32(33)),
			Evaluator:      ptr.To("evaluator"),
			SocketPath:     ptr.To("socket-path"),
			LogFormat:      ptr.To("log"),
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
			Name:  "AZURE_STORAGE_ACCOUNT",
			Value: "account123",
		},
		{
			Name:  "AZURE_STORAGE_KEY",
			Value: "key456",
		},
		{
			Name:  "AWS_ACCESS_KEY_ID",
			Value: "AKIAIOSFODNN7EXAMPLE",
		},
		{
			Name:  "AWS_REGION",
			Value: "us-east-1",
		},
		{
			Name:  "GOOGLE_APPLICATION_CREDENTIALS",
			Value: "/var/run/secrets/gcp/key.json",
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
			Name:  "PRE_LOG_FORMAT",
			Value: "log",
		},
		{
			Name:  "PRE_RESOLVER",
			Value: "rpc",
		},
		{
			Name:  "PRE_SOCKET_PATH",
			Value: "socket-path",
		},
	}
	require.Equal(t, expected, ff.Spec.ToEnvVars())
}

func Test_FLagSourceConfiguration_Merge_Keepalive(t *testing.T) {
	base := &FeatureFlagSourceSpec{}
	// override sets both keepalive fields
	permit := true
	base.Merge(&FeatureFlagSourceSpec{
		KeepAliveMinTime:             &metav1.Duration{Duration: 45 * time.Second},
		KeepAlivePermitWithoutStream: &permit,
	})
	require.NotNil(t, base.KeepAliveMinTime)
	require.Equal(t, 45*time.Second, base.KeepAliveMinTime.Duration)
	require.NotNil(t, base.KeepAlivePermitWithoutStream)
	require.True(t, *base.KeepAlivePermitWithoutStream)

	// nil override must not clobber existing values
	base.Merge(&FeatureFlagSourceSpec{})
	require.NotNil(t, base.KeepAliveMinTime)
	require.Equal(t, 45*time.Second, base.KeepAliveMinTime.Duration)
	require.NotNil(t, base.KeepAlivePermitWithoutStream)
}

func Test_FLagSourceConfiguration_ToEnvVars_Keepalive(t *testing.T) {
	// unset: neither keepalive env var is emitted
	unset := FeatureFlagSourceSpec{EnvVarPrefix: ptr.To("FLAGD")}
	for _, e := range unset.ToEnvVars() {
		require.NotContains(t, e.Name, "KEEP_ALIVE")
	}

	// set: both env vars emitted with formatted values
	permit := true
	set := FeatureFlagSourceSpec{
		EnvVarPrefix:                 ptr.To("FLAGD"),
		KeepAliveMinTime:             &metav1.Duration{Duration: 45 * time.Second},
		KeepAlivePermitWithoutStream: &permit,
	}
	got := set.ToEnvVars()
	require.Contains(t, got, v1.EnvVar{Name: "FLAGD_KEEP_ALIVE_MIN_TIME", Value: "45s"})
	require.Contains(t, got, v1.EnvVar{Name: "FLAGD_KEEP_ALIVE_PERMIT_WITHOUT_STREAM", Value: "true"})

	// set false: explicitly disabling permit-without-stream still emits the env var
	permitFalse := false
	setFalse := FeatureFlagSourceSpec{
		EnvVarPrefix:                 ptr.To("FLAGD"),
		KeepAlivePermitWithoutStream: &permitFalse,
	}
	gotFalse := setFalse.ToEnvVars()
	require.Contains(t, gotFalse, v1.EnvVar{Name: "FLAGD_KEEP_ALIVE_PERMIT_WITHOUT_STREAM", Value: "false"})

	// mixed: only the set field is emitted
	mixed := FeatureFlagSourceSpec{
		EnvVarPrefix:     ptr.To("FLAGD"),
		KeepAliveMinTime: &metav1.Duration{Duration: 1*time.Minute + 30*time.Second},
	}
	gotMixed := mixed.ToEnvVars()
	require.Contains(t, gotMixed, v1.EnvVar{Name: "FLAGD_KEEP_ALIVE_MIN_TIME", Value: "1m30s"})
	for _, e := range gotMixed {
		require.NotEqual(t, "FLAGD_KEEP_ALIVE_PERMIT_WITHOUT_STREAM", e.Name)
	}
}
