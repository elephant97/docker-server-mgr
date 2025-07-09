package utils

func ListToMap[T any, K comparable, V any](
	slice []T,
	keyFunc func(T) K,
	valueFunc func(T) V,
) map[K]V {
	result := make(map[K]V)

	for _, item := range slice {
		key := keyFunc(item)
		value := valueFunc(item)
		result[key] = value
	}

	return result
}
