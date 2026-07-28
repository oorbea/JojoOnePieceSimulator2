package endpoints

import (
	"context"
	"net/http"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// claimsContextKey is an unexported type so context values set here can
// never collide with keys from other packages.
type claimsContextKey struct{}

func withClaims(ctx context.Context, claims ports.Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey{}, claims)
}

func claimsFromContext(ctx context.Context) (ports.Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey{}).(ports.Claims)
	return claims, ok
}

// ClaimsFromRequest returns the authenticated caller's claims, populated by
// RequireAuth. Only meaningful for handlers mounted behind RequireAuth.
func ClaimsFromRequest(r *http.Request) (ports.Claims, bool) {
	return claimsFromContext(r.Context())
}
