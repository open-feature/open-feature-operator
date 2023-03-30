package utils

import (
	"fmt"
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

func GetIntEnvVar(key string) (int, error) {
	val, err := strconv.Atoi(key)
	if err != nil {
		return 0, fmt.Errorf("could not parse %s env var to int: %w", key, err)
	}
	return val, nil
}

func GetBoolEnvVar(key string) (bool, error) {
	val, err := strconv.ParseBool(key)
	if err != nil {
		return false, fmt.Errorf("could not parse %s env var to bool: %w", key, err)
	}
	return val, nil
}
