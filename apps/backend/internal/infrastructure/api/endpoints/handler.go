package endpoints

import "net/http"

// HandlerFunc is like http.HandlerFunc but returns an error instead of
// writing its own error response, so error-to-status mapping lives in one
// place (handleError) instead of being duplicated at every call site.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Wrap adapts a HandlerFunc into an http.HandlerFunc, centralizing error
// handling: any error returned by h is translated into an HTTP response by
// handleError.
func Wrap(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			handleError(w, err)
		}
	}
}
