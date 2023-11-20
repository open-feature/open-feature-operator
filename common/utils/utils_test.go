package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_FeatureFlagId(t *testing.T) {
	require.Equal(t, "namespace_name", FeatureFlagId("namespace", "name"))
}

func Test_FeatureFlagConfigMapKey(t *testing.T) {
	require.Equal(t, "namespace_name.flagd.json", FeatureFlagConfigMapKey("namespace", "name"))
}

func Test_FalseVal(t *testing.T) {
	f := false
	require.Equal(t, &f, FalseVal())
}

func Test_TrueVal(t *testing.T) {
	tt := true
	require.Equal(t, &tt, TrueVal())
}

func Test_ContainsString(t *testing.T) {
	slice := []string{"str1", "str2"}
	require.True(t, ContainsString(slice, "str1"))
	require.False(t, ContainsString(slice, "some"))
}

func Test_ParseAnnotations(t *testing.T) {
	s1, s2 := ParseAnnotation("some/anno", "default")
	require.Equal(t, "some", s1)
	require.Equal(t, "anno", s2)

	s1, s2 = ParseAnnotation("anno", "default")
	require.Equal(t, "default", s1)
	require.Equal(t, "anno", s2)
}
