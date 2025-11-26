package repo

// deref safely de-references AWS SDK pointers.
func deref[T any](ptr *T) T {
	var zero T
	if ptr == nil {
		return zero
	}
	return *ptr
}
