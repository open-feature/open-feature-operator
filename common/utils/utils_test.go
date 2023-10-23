package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_FeatureFlagConfigurationId(t *testing.T) {
	require.Equal(t, "namespace_name", FeatureFlagConfigurationId("namespace", "name"))
}

func Test_FeatureFlagConfigurationConfigMapKey(t *testing.T) {
	require.Equal(t, "namespace_name.flagd.json", FeatureFlagConfigurationConfigMapKey("namespace", "name"))
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

func Test_parseAnnotation(t *testing.T) {
	tests := []struct {
		name      string
		s         string
		defaultNs string
		wantNs    string
		want      string
	}{
		{
			name:      "no namespace",
			s:         "test",
			defaultNs: "ofo",
			wantNs:    "ofo",
			want:      "test",
		},
		{
			name:      "namespace",
			s:         "myns/test",
			defaultNs: "ofo",
			wantNs:    "myns",
			want:      "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ParseAnnotation(tt.s, tt.defaultNs)
			if got != tt.wantNs {
				t.Errorf("parseAnnotation() got = %v, want %v", got, tt.wantNs)
			}
			if got1 != tt.want {
				t.Errorf("parseAnnotation() got1 = %v, want %v", got1, tt.want)
			}
		})
	}
}
