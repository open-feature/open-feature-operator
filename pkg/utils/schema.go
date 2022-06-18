package utils

import _ "embed"

//go:embed flagd-definitions.json
var schema string

func GetSchema() string {
	return schema
}
