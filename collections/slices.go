package collections

func Uniqualize[T comparable](input []T) []T {
	// с указателями не работает. жаль. sort.Slice(input, func(i, j int) bool { return input[i] < input[j] })
	seen := make(map[T]bool, len(input))
	res := make([]T, 0, len(input))
	for _, val := range input {
		if _, ok := seen[val]; !ok {
			seen[val] = true
			res = append(res, val)
		}
	}
	return res
}

func MergeSlices[T comparable](slice1 []T, slice2 []T) []T {
	return Uniqualize(append(slice1, slice2...))
}
