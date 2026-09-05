package endpoints

import (
	"net/http"
	"time"
)

// CookieConfig configures the refresh-token cookie. Both setRefreshCookie
// and clearRefreshCookie MUST derive every attribute from the same
// CookieConfig value - a mismatched Path/Domain/Secure/SameSite between set
// and clear is the classic bug that leaves a cookie the browser won't let
// you delete.
type CookieConfig struct {
	Name     string
	Path     string
	Secure   bool
	SameSite http.SameSite
}

// setRefreshCookie sets the HttpOnly refresh-token cookie carrying token,
// expiring alongside it.
func setRefreshCookie(w http.ResponseWriter, cfg CookieConfig, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.Name,
		Value:    token,
		Path:     cfg.Path,
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
		Expires:  expiresAt,
		// Domain intentionally omitted: host-only cookie. duckdns.org is on
		// the Public Suffix List, so a Domain attribute would be rejected or
		// dangerous; host-only is both correct and safest here.
	})
}

// clearRefreshCookie deletes the refresh-token cookie set by
// setRefreshCookie, using the exact same attributes so the browser
// recognizes it as the same cookie.
func clearRefreshCookie(w http.ResponseWriter, cfg CookieConfig) {
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.Name,
		Value:    "",
		Path:     cfg.Path,
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
		MaxAge:   -1,
	})
}
