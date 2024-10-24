package utils

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

func TrueVal() *bool {
	b := true
	return &b
}

func FalseVal() *bool {
	b := false
	return &b
}

func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func ParseAnnotation(s string, defaultNs string) (string, string) {
	ss := strings.Split(s, "/")
	if len(ss) == 2 {
		return ss[0], ss[1]
	}
	return defaultNs, s
}

// unique string used to create unique volume mount and file name
func FeatureFlagId(namespace, name string) string {
	return fmt.Sprintf("%s_%s", namespace, name)
}

// unique key (and filename) for configMap data
func FeatureFlagConfigMapKey(namespace, name string) string {
	return fmt.Sprintf("%s.flagd.json", FeatureFlagId(namespace, name))
}

type ExponentialBackoff struct {
	StartDelay time.Duration
	MaxDelay   time.Duration
	counter    int64
}

func (e *ExponentialBackoff) Next() time.Duration {
	val := atomic.AddInt64(&e.counter, 1)

	delay := e.StartDelay * (1 << (val - 1))
	if delay > e.MaxDelay {
		delay = e.MaxDelay
	}
	return delay
}

func (e *ExponentialBackoff) Reset() {
	e.counter = 0
}
