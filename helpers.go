package centreon

// nilToEmpty returns an empty, non-nil slice when s is nil, so it marshals to
// `[]` rather than `null` for API fields that reject a null array.
func nilToEmpty[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
