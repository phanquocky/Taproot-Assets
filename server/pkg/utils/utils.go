package utils

import "reflect"

// TernaryOpWithType return default value if value is nil.
func TernaryOpWithType[T any](value, defaultV T) T {
	if reflect.ValueOf(value).IsZero() {
		return defaultV
	}

	return value
}
