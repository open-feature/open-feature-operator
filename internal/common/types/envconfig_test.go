package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnvConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		env     EnvConfig
		wantErr bool
	}{
		{name: "empty", env: EnvConfig{}},
		{name: "valid", env: EnvConfig{SidecarImagePullPolicy: "Always", FlagdImagePullPolicy: "IfNotPresent", FlagdProxyImagePullPolicy: "Never"}},
		{name: "invalid sidecar", env: EnvConfig{SidecarImagePullPolicy: "always"}, wantErr: true},
		{name: "invalid flagd", env: EnvConfig{FlagdImagePullPolicy: "sometimes"}, wantErr: true},
		{name: "invalid proxy", env: EnvConfig{FlagdProxyImagePullPolicy: "ifnotpresent"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.env.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
