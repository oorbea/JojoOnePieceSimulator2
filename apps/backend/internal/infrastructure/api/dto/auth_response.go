package dto

import "time"

// LoginResponse is the JSON body returned by POST /auth/google,
// /auth/refresh. The refresh token itself normally travels only in the
// HttpOnly cookie set alongside this body (see cookies.go); RefreshToken is
// populated here only when the caller sent
// X-Refresh-Token-Transport: header (native clients, which have no cookie
// jar), and omitted entirely for the default web/cookie flow.
type LoginResponse struct {
	AccessToken string       `json:"accessToken"`
	TokenType   string       `json:"tokenType"`
	ExpiresAt   time.Time    `json:"expiresAt"`
	User        UserResponse `json:"user"`
	// RefreshToken is only set when the caller opted into header transport
	// via X-Refresh-Token-Transport: header.
	RefreshToken string `json:"refreshToken,omitempty"`
}
