package v1beta1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_validateFeatureFlagTargeting(t *testing.T) {
	tests := []struct {
		name    string
		in      json.RawMessage
		wantErr bool
	}{
		{
			name: "happy path",
			in: json.RawMessage(`{
				"fractional": [
					{"var": "email"},
					[
					"red",
					25
					]
				]
				}`),
			wantErr: false,
		},
		{
			name: "invalid input",
			in: json.RawMessage(`{
				"fractional": [
					{"var": "email"},
					[
					"red",
					25d
					]
				]
				}`),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				require.NotNil(t, validateFeatureFlagTargeting(tt.in))
			} else {
				require.Nil(t, validateFeatureFlagTargeting(tt.in))
			}
		})
	}
}
