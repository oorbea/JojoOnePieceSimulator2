package dto

// ErrorResponse is the JSON body returned for any non-2xx response. Code is
// a stable machine-readable identifier (e.g. "STAND_NOT_FOUND") the
// frontend maps to a localized message - see shared/lib/toast.ts's
// showErrorToast. Error stays free-text English: it's the fallback shown
// when a client doesn't recognize the code, and what shows up in logs/curl.
type ErrorResponse struct {
	Error   string   `json:"error"`
	Code    string   `json:"code,omitempty"`
	Details []string `json:"details,omitempty"`
}
