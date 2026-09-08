package ports

import "fmt"

// optStringer renders an optional fmt.Stringer enum field as its String()
// value, or "" when the pointer is nil (the field was unset) - shared by
// every *Filters.Canonical() method in this package.
func optStringer[T fmt.Stringer](v *T) string {
	if v == nil {
		return ""
	}
	return (*v).String()
}

// optString renders an optional string field, or "" when nil.
func optString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
