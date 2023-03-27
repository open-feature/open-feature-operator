package utils

import "fmt"

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

// unique string used to create unique volume mount and file name
func FeatureFlagConfigurationId(namespace, name string) string {
	return fmt.Sprintf("%s_%s", namespace, name)
}

// unique key (and filename) for configMap data
func FeatureFlagConfigurationConfigMapKey(namespace, name string) string {
	return fmt.Sprintf("%s.flagd.json", FeatureFlagConfigurationId(namespace, name))
}
