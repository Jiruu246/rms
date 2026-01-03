package utils

// Use with caustion, performance may be impacted
func Reduce[T any, R any](slice []T, accumulator func(R, T) R, initial R) R {
	result := initial
	for _, v := range slice {
		result = accumulator(result, v)
	}
	return result
}

func FindFirst[T any](slice []T, predicate func(T) bool) (T, bool) {
	for _, v := range slice {
		if predicate(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}
