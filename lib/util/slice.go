package util

import (
	"iter"
	"maps"
)

func SeqToSlice[T any](seq iter.Seq[T]) []T {
	var result []T
	for value := range seq {
		result = append(result, value)
	}
	return result
}

func MapKeys[K comparable, V any](m map[K]V) []K {
	seq := maps.Keys(m)
	return SeqToSlice(seq)
}
