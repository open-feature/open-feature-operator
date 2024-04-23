package v1beta1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_validateFeatureFlagTargeting(t *testing.T) {
	// happy path
	in := json.RawMessage(`{
		"fractional": [
			{"var": "email"},
			[
			"red",
			25
			]
		]
		}`)

	require.Nil(t, validateFeatureFlagTargeting(in))

	// failure path
	in = json.RawMessage(`{
		"fractional": [
			{"var": "email"},
			[
			"red",
			25d
			]
		]
		}`)

	require.NotNil(t, validateFeatureFlagTargeting(in))
}
