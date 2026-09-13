package dto

// ErrorResponse is the JSON body returned for any non-2xx response. Code is
// a stable machine-readable identifier (e.g. "STAND_NOT_FOUND") the
// frontend maps to a localized message - see shared/lib/toast.ts's
// showErrorToast. Error stays free-text English: it's the fallback shown
// when a client doesn't recognize the code, and what shows up in logs/curl.
// Details carries per-field validation failures (400s only) so the
// frontend can localize and attribute each one - see FieldError.
type ErrorResponse struct {
	Error   string       `json:"error"`
	Code    string       `json:"code,omitempty"`
	Details []FieldError `json:"details,omitempty"`
}

// FieldError is one field-level validation failure. Code is a stable i18n
// key (e.g. "validation.nameRequired") the frontend translates via
// t(code, {defaultValue: message}) - same degrade-to-English pattern as the
// top-level ErrorResponse.Code/Error. Message is the English fallback text,
// same wording these errors always had before Code existed.
type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Validation i18n codes shared across every request DTO's Validate(). Named
// ones reuse the exact keys the frontend's zod schemas already validate
// against client-side (shared/lib/zod.ts, *.types.ts) so no duplicate
// catalog entries are needed for the fields both layers check; the generic
// ones (ValInvalidValue, ValRequired, ValLocale*) cover the backend-only
// checks (enum parsing, malformed cursors, locale rules) the frontend never
// had a dedicated message for.
const (
	ValRequired              = "validation.required"
	ValNameRequired          = "validation.nameRequired"
	ValDescriptionRequired   = "validation.descriptionRequired"
	ValSkillsRequired        = "validation.skillsRequired"
	ValOrderNonNegative      = "validation.orderNonNegative"
	ValBattleIqRange         = "validation.battleIqRange"
	ValInvalidValue          = "validation.invalidValue"
	ValLocaleUnsupported     = "validation.localeUnsupported"
	ValLocaleDefaultRequired = "validation.localeDefaultRequired"
)
