package utils

import (
	"testing"
	"time"

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

func TestExponentialBackoff_Next(t *testing.T) {
	tests := []struct {
		name       string
		startDelay time.Duration
		maxDelay   time.Duration
		steps      int
		expected   time.Duration
	}{
		{name: "basic backoff", startDelay: 1 * time.Second, maxDelay: 16 * time.Second, steps: 3, expected: 4 * time.Second},
		{name: "max delay reached", startDelay: 1 * time.Second, maxDelay: 5 * time.Second, steps: 10, expected: 5 * time.Second},
		{name: "single step", startDelay: 500 * time.Millisecond, maxDelay: 10 * time.Second, steps: 1, expected: 500 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backoff := &ExponentialBackoff{StartDelay: tt.startDelay, MaxDelay: tt.maxDelay}
			var result time.Duration
			for i := 0; i < tt.steps; i++ {
				result = backoff.Next()
			}
			if result != tt.expected {
				t.Errorf("Expected delay after %d steps to be %v; got %v", tt.steps, tt.expected, result)
			}
		})
	}
}

func TestExponentialBackoff_Reset(t *testing.T) {
	backoff := &ExponentialBackoff{StartDelay: 1 * time.Second, MaxDelay: 10 * time.Second}

	// Increment the backoff a few times
	backoff.Next()
	backoff.Next()

	// Reset and check the counter
	backoff.Reset()
	if backoff.counter != 0 {
		t.Errorf("Expected counter to be reset to 0; got %d", backoff.counter)
	}
	if backoff.Next() != 1*time.Second {
		t.Errorf("Expected delay after reset to be %v; got %v", 1*time.Second, backoff.Next())
	}
}
