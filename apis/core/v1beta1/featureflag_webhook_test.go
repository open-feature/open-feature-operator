package v1beta1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_validateFeatureFlagTargeting(t *testing.T) {
	tests := []struct {
		name    string
		in      Flags
		wantErr bool
	}{
		{
			name: "happy path",
			in: Flags{
				FlagsMap: map[string]Flag{
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
			in: Flags{
				FlagsMap: map[string]Flag{
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
			name: "flactional invalid bucketing",
			in: Flags{
				FlagsMap: map[string]Flag{
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
			in: Flags{
				FlagsMap: map[string]Flag{
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
			name: "flactional invalid weighting",
			in: Flags{
				FlagsMap: map[string]Flag{
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
			in: Flags{
				FlagsMap: map[string]Flag{
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
