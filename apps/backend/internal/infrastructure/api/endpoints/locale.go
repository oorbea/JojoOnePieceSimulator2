package endpoints

import (
	"context"
	"net/http"
	"strings"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// localeContextKey is an unexported type so context values set here can
// never collide with keys from other packages.
type localeContextKey struct{}

func withLocale(ctx context.Context, locale enums.Locale) context.Context {
	return context.WithValue(ctx, localeContextKey{}, locale)
}

// LocaleFromRequest returns the locale resolved by resolveLocale for r,
// defaulting to enums.EnGB if the middleware was somehow skipped.
func LocaleFromRequest(r *http.Request) enums.Locale {
	if locale, ok := r.Context().Value(localeContextKey{}).(enums.Locale); ok {
		return locale
	}
	return enums.EnGB
}

// resolveLocale determines the caller's locale for read routes: an explicit
// "?lang=" query override wins, otherwise the first supported language tag
// in the standard Accept-Language header, otherwise enums.EnGB. It never
// rejects a request - an unrecognized value just falls through to the
// default, exactly like every other public read route.
func resolveLocale(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale := enums.EnGB
		if q := r.URL.Query().Get("lang"); q != "" {
			if parsed, err := enums.ParseLocale(q); err == nil {
				locale = parsed
			}
		} else if parsed, ok := parseAcceptLanguage(r.Header.Get("Accept-Language")); ok {
			locale = parsed
		}
		next.ServeHTTP(w, r.WithContext(withLocale(r.Context(), locale)))
	})
}

// parseAcceptLanguage picks the first tag in an Accept-Language header
// (ignoring quality values - our locale set is tiny enough that a simple
// first-match is fine) that maps onto a supported enums.Locale, matching
// language-only prefixes too (e.g. "es" or "es-MX" both resolve to es-ES,
// "ca" to ca-ES).
func parseAcceptLanguage(header string) (enums.Locale, bool) {
	for _, part := range strings.Split(header, ",") {
		tag := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		if tag == "" {
			continue
		}
		if locale, err := enums.ParseLocale(tag); err == nil {
			return locale, true
		}
		lang := strings.ToLower(strings.SplitN(tag, "-", 2)[0])
		switch lang {
		case "en":
			return enums.EnGB, true
		case "es":
			return enums.EsES, true
		case "ca":
			return enums.CaES, true
		}
	}
	return enums.EnGB, false
}
