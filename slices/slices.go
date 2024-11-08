package slices

import (
	"sort"
)

func Uniqualize[T comparable](input []T) []T {
	// с указателями не работает. жаль. sort.Slice(input, func(i, j int) bool { return input[i] < input[j] })
	res := make([]T, 0, len(input))
	for _, val := range input {
		found := false
		for _, v := range res {
			if v == val {
				found = true
				break
			}
		}
		if !found {
			res = append(res, val)
		}
	}
	return res
}

func MergeSlices(slice1 []string, slice2 []string) []string {
	return Uniqualize(append(slice1, slice2...))
}

func MapKeys[T any](m map[string]T) []string {
	res := make([]string, 0, len(m))
	for i := range m {
		res = append(res, i)
	}
	sort.Strings(res)
	return res
}
