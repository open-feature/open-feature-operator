package common

import "errors"

var ErrCannotCastFlagSourceConfiguration = errors.New("cannot cast FlagSourceConfiguration to v1alpha3")

func TrueVal() *bool {
	b := true
	return &b
}

func FalseVal() *bool {
	b := false
	return &b
}
