package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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

func GetIntEnvVar(key string, defaultVal int) (int, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal, nil
	}
	valInt, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("could not parse %s env var to int: %w", key, err)
	}
	return valInt, nil
}

func GetBoolEnvVar(key string, defaultVal bool) (bool, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal, nil
	}
	valBool, err := strconv.ParseBool(val)
	if err != nil {
		return false, fmt.Errorf("could not parse %s env var to bool: %w", key, err)
	}
	return valBool, nil
}

// unique string used to create unique volume mount and file name
func FeatureFlagId(namespace, name string) string {
	return fmt.Sprintf("%s_%s", namespace, name)
}

// unique key (and filename) for configMap data
func FeatureFlagConfigMapKey(namespace, name string) string {
	return fmt.Sprintf("%s.flagd.json", FeatureFlagId(namespace, name))
}
