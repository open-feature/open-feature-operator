package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_FLagSourceConfiguration_EnvVarKey(t *testing.T) {
	require.Equal(t, "pre_suf", EnvVarKey("pre", "suf"))
}

func Test_FLagSourceConfiguration_FeatureFlagConfigurationId(t *testing.T) {
	require.Equal(t, "pre_suf", FeatureFlagConfigurationId("pre", "suf"))
}

func Test_FLagSourceConfiguration_FeatureFlagConfigMapKey(t *testing.T) {
	require.Equal(t, "pre_suf.flagd.json", FeatureFlagConfigMapKey("pre", "suf"))
}

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

func Test_FLagSourceConfiguration_envVarKey(t *testing.T) {
	require.Equal(t, "pre_suf", EnvVarKey("pre", "suf"))
}
