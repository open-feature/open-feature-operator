package v1beta1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_FLagSourceConfiguration_SyncProvider(t *testing.T) {
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
