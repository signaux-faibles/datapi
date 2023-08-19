package imports

func pointer[T any](value T) *T {
	return &value
}
