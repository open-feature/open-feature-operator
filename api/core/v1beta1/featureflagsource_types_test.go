package v1beta1

import (
	"testing"

	"github.com/open-feature/open-feature-operator/api/core/v1beta1/common"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
			OFREPPort: 8016,
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
			OFREPPort: 8016,
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
			OFREPPort: 9090,
		},
	}

	ff_old.Spec.Merge(&ff_new.Spec)

	require.Equal(t, ff_old.Spec.EnvVarPrefix, "PREFIX")
	require.Equal(t, ff_old.Spec.Port, int32(331))
	require.Equal(t, ff_old.Spec.ManagementPort, int32(221))
	require.Equal(t, ff_old.Spec.SocketPath, "socket-path1")
	require.Equal(t, ff_old.Spec.Evaluator, "evaluator1")
	require.Equal(t, ff_old.Spec.LogFormat, "log1")
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
	require.Equal(t, ff_old.Spec.DefaultSyncProvider, common.SyncProviderFilepath)
	require.Equal(t, ff_old.Spec.RolloutOnChange, common.FalseVal())
	require.Equal(t, ff_old.Spec.ProbesEnabled, common.FalseVal())
	require.Equal(t, ff_old.Spec.DebugLogging, common.FalseVal())
	require.Equal(t, ff_old.Spec.OtelCollectorUri, "")
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
	require.Equal(t, int32(9090), ff_old.Spec.OFREPPort)
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
