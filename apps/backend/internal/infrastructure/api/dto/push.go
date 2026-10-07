package dto

import (
	"net/url"
	"regexp"
	"strings"
)

const (
	// maxPushEndpointLen / maxPushKeyLen bound what a client may store; real
	// endpoints are a few hundred bytes and keys under 100.
	maxPushEndpointLen = 2048
	maxPushKeyLen      = 256
)

// allowedPushHosts are the push services browsers actually use. The server
// POSTs a message to whatever endpoint a client registers, so accepting an
// arbitrary URL would turn this API into an SSRF primitive (any logged-in
// user could make the backend call internal addresses). Pinning the host to
// the real push services closes that; a new browser vendor is a one-line add.
// Each entry matches the host itself and any subdomain of it.
var allowedPushHosts = []string{
	"fcm.googleapis.com",        // Chrome, Edge, Brave, Opera, Samsung Internet
	"push.services.mozilla.com", // Firefox (updates.push..., updates-autopush...)
	"push.apple.com",            // Safari / iOS web push (web.push.apple.com)
	"notify.windows.com",        // legacy Edge (WNS)
}

var pushKeyPattern = regexp.MustCompile(`^[A-Za-z0-9_\-+/=]+$`)

// PushConfigResponse is returned by GET /users/me/push. When push is not
// configured on the server (no VAPID keys) Enabled is false and the client
// hides the feature.
type PushConfigResponse struct {
	Enabled   bool   `json:"enabled"`
	PublicKey string `json:"publicKey"`
}

// PushKeys are the two values PushSubscription.toJSON() carries besides the
// endpoint.
type PushKeys struct {
	P256dh string `json:"p256dh"`
	Auth   string `json:"auth"`
}

// PushSubscribeRequest is the JSON body of POST /push/subscriptions - the
// browser's PushSubscription.toJSON() minus the fields the server ignores.
type PushSubscribeRequest struct {
	Endpoint string   `json:"endpoint"`
	Keys     PushKeys `json:"keys"`
}

// Validate checks the endpoint against the push-service allowlist and the
// keys' shape, collecting every problem.
func (r PushSubscribeRequest) Validate() error {
	var errs []FieldError
	if msg := validatePushEndpoint(r.Endpoint); msg != "" {
		errs = append(errs, FieldError{Field: "endpoint", Code: ValInvalidValue, Message: "endpoint: " + msg})
	}
	if !validPushKey(r.Keys.P256dh) {
		errs = append(errs, FieldError{Field: "keys.p256dh", Code: ValInvalidValue, Message: "keys.p256dh: must be a base64 key of at most 256 characters"})
	}
	if !validPushKey(r.Keys.Auth) {
		errs = append(errs, FieldError{Field: "keys.auth", Code: ValInvalidValue, Message: "keys.auth: must be a base64 secret of at most 256 characters"})
	}
	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}

// PushUnsubscribeRequest is the JSON body of DELETE /push/subscriptions.
type PushUnsubscribeRequest struct {
	Endpoint string `json:"endpoint"`
}

// Validate only requires a plausible endpoint: unsubscribing an unknown or
// foreign one is a harmless no-op, so the allowlist isn't needed here.
func (r PushUnsubscribeRequest) Validate() error {
	if r.Endpoint == "" || len(r.Endpoint) > maxPushEndpointLen {
		return &ValidationError{Errors: []FieldError{{Field: "endpoint", Code: ValRequired, Message: "endpoint is required"}}}
	}
	return nil
}

func validPushKey(key string) bool {
	return key != "" && len(key) <= maxPushKeyLen && pushKeyPattern.MatchString(key)
}

// validatePushEndpoint returns "" for an acceptable endpoint, else why not.
func validatePushEndpoint(raw string) string {
	if raw == "" {
		return "is required"
	}
	if len(raw) > maxPushEndpointLen {
		return "is too long"
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "must be a valid URL"
	}
	if u.Scheme != "https" {
		return "must use https"
	}
	if u.User != nil {
		return "must not contain credentials"
	}
	if port := u.Port(); port != "" && port != "443" {
		return "must use the default https port"
	}
	host := strings.ToLower(u.Hostname())
	for _, allowed := range allowedPushHosts {
		if host == allowed || strings.HasSuffix(host, "."+allowed) {
			return ""
		}
	}
	return "is not a recognised push service"
}
