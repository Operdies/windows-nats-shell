package utils

import (
	"gopkg.in/yaml.v3"
)

func Contains[T comparable](haystack []T, needle T) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}

func Select[T1 any, T2 any](source []T1, selector func(T1) T2) []T2 {
	result := make([]T2, len(source))
	for i, item := range source {
		r := selector(item)
		result[i] = r
	}
	return result
}

func Filter[T1 any](source []T1, filter func(T1) bool) []T1 {
	result := make([]T1, len(source))
	k := 0
	for _, item := range source {
		if filter(item) {
			result[k] = item
			k = k + 1
		}
	}
	return result[:k]
}

func EncodeAny[T any](value T) []byte {
	result, _ := yaml.Marshal(value)
	return result
}

func DecodeAny[T any](buffer []byte) T {
	var result T
	yaml.Unmarshal(buffer, &result)
	return result
}
