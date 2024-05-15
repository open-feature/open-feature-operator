package common

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func Test_FeatureFlagSource_SyncProvider(t *testing.T) {
	k := SyncProviderKubernetes
	f := SyncProviderFilepath
	h := SyncProviderHttp
	g := SyncProviderGrpc

	require.True(t, k.IsKubernetes())
	require.True(t, f.IsFilepath())
	require.True(t, h.IsHttp())
	require.True(t, g.IsGrpc())

	require.False(t, f.IsKubernetes())
	require.False(t, h.IsFilepath())
	require.False(t, k.IsGrpc())
	require.False(t, g.IsHttp())
}

func Test_FLagSourceConfiguration_EnvVarKey(t *testing.T) {
	require.Equal(t, "pre_suf", EnvVarKey("pre", "suf"))
}

func Test_FLagSourceConfiguration_FeatureFlagConfigurationId(t *testing.T) {
	require.Equal(t, "pre_suf", FeatureFlagConfigurationId("pre", "suf"))
}

func Test_FLagSourceConfiguration_FeatureFlagConfigMapKey(t *testing.T) {
	require.Equal(t, "pre_suf.flagd.json", FeatureFlagConfigMapKey("pre", "suf"))
}

func Test_RemoveDuplicateEnvVars(t *testing.T) {
	input1 := []corev1.EnvVar{
		{
			Name:  "key1",
			Value: "val1",
		},
		{
			Name:  "key2",
			Value: "val2",
		},
		{
			Name:  "key1",
			Value: "val3",
		},
	}
	input2 := []corev1.EnvVar{
		{
			Name:  "key1",
			Value: "val1",
		},
		{
			Name:  "key2",
			Value: "val2",
		},
		{
			Name:  "key3",
			Value: "val3",
		},
	}

	require.Equal(t, RemoveDuplicateEnvVars(input1), []corev1.EnvVar{
		{
			Name:  "key1",
			Value: "val3",
		},
		{
			Name:  "key2",
			Value: "val2",
		},
	})

	require.Equal(t, RemoveDuplicateEnvVars(input2), []corev1.EnvVar{
		{
			Name:  "key1",
			Value: "val1",
		},
		{
			Name:  "key2",
			Value: "val2",
		},
		{
			Name:  "key3",
			Value: "val3",
		},
	})
}

func Test_RemoveDuplicateGenerics(t *testing.T) {
	input1 := []string{
		"some", "input", "duplicate", "some",
	}
	input2 := []int{
		1, 2, 3, 4, 2,
	}

	require.Equal(t, RemoveDuplicatesGeneric(input1), []string{
		"some", "input", "duplicate",
	})

	require.Equal(t, RemoveDuplicatesGeneric(input2), []int{
		1, 2, 3, 4,
	})
}
