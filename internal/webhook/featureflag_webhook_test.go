package webhook

import (
	"encoding/json"
	"testing"

	"github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	"github.com/stretchr/testify/require"
)

func Test_validateFeatureFlagTargeting(t *testing.T) {
	tests := []struct {
		name    string
		in      v1beta1.Flags
		wantErr bool
	}{
		{
			name: "happy path",
			in: v1beta1.Flags{
				FlagsMap: map[string]v1beta1.Flag{
					"fractional": {
						State: "ENABLED",
						Variants: json.RawMessage(`{
							"clubs": "clubs",
							"diamonds": "diamonds",
							"hearts": "hearts",
							"spades": "spades",
							"none": "none"}
						`),
						DefaultVariant: "none",
						Targeting: json.RawMessage(`{
							"fractional": [
								["clubs", 25],
								["diamonds", 25],
								["hearts", 25],
								["spades", 25]
						  ]}
						`),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "happy path no targeting",
			in: v1beta1.Flags{
				FlagsMap: map[string]v1beta1.Flag{
					"fractional": {
						State: "ENABLED",
						Variants: json.RawMessage(`{
							"clubs": "clubs",
							"diamonds": "diamonds",
							"hearts": "hearts",
							"spades": "spades",
							"none": "none"}
						`),
						DefaultVariant: "none",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "fractional invalid bucketing",
			in: v1beta1.Flags{
				FlagsMap: map[string]v1beta1.Flag{
					"fractional-invalid-bucketing": {
						State: "ENABLED",
						Variants: json.RawMessage(`{
							"clubs": "clubs",
							"diamonds": "diamonds",
							"hearts": "hearts",
							"spades": "spades",
							"none": "none"}
						`),
						DefaultVariant: "none",
						Targeting: json.RawMessage(`{
							"fractional": [
								"invalid",
								["clubs", 25],
								["diamonds", 25],
								["hearts", 25],
								["spades", 25]
						  ]}
						`),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty variants",
			in: v1beta1.Flags{
				FlagsMap: map[string]v1beta1.Flag{
					"fractional-invalid-bucketing": {
						State:          "ENABLED",
						Variants:       json.RawMessage{},
						DefaultVariant: "on",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "fractional invalid weighting",
			in: v1beta1.Flags{
				FlagsMap: map[string]v1beta1.Flag{
					"fractional-invalid-weighting": {
						State: "ENABLED",
						Variants: json.RawMessage(`{
							"clubs": "clubs",
							"diamonds": "diamonds",
							"hearts": "hearts",
							"spades": "spades",
							"none": "none"}
						`),
						DefaultVariant: "none",
						Targeting: json.RawMessage(`{
							"fractional": [
								["clubs", 25],
								["diamonds", "25"],
								["hearts", 25],
								["spades", 25]
						  ]}
						`),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-ends-with-param",
			in: v1beta1.Flags{
				FlagsMap: map[string]v1beta1.Flag{
					"invalid-ends-with-param": {
						State: "ENABLED",
						Variants: json.RawMessage(`{
							"prefix": 1,
							"postfix": 2
						  }
						`),
						DefaultVariant: "none",
						Targeting: json.RawMessage(`{
							"if": [
							  {
								"ends_with": [{ "var": "id" }, 0]
							  },
							  "postfix",
							  "prefix"
							]
						  }
						`),
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				require.NotNil(t, validateFeatureFlagFlags(tt.in))
			} else {
				require.Nil(t, validateFeatureFlagFlags(tt.in))
			}
		})
	}
}
