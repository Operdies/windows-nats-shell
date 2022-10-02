package utils

import (
	"gopkg.in/yaml.v3"
)

func EncodeAny[T any](value T) []byte {
	result, _ := yaml.Marshal(value)
	return result
}

func DecodeAny[T any](buffer []byte) T {
	var result T
	yaml.Unmarshal(buffer, &result)
	return result
}
