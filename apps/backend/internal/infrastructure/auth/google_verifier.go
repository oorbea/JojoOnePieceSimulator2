// Package auth provides infrastructure implementations of ports.
// IGoogleTokenVerifier and ports.ITokenIssuer - this is the only place the
// domain's Google/JWT concerns touch a concrete library.
package auth

import (
	"context"
	"fmt"

	"google.golang.org/api/idtoken"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// GoogleVerifier verifies Google-issued ID tokens against a single expected
// audience (the OAuth client id configured for this backend).
type GoogleVerifier struct {
	clientID string
}

var _ ports.IGoogleTokenVerifier = (*GoogleVerifier)(nil)

func NewGoogleVerifier(clientID string) *GoogleVerifier {
	return &GoogleVerifier{clientID: clientID}
}

// Verify validates rawIDToken's signature, issuer, audience and expiry, and
// extracts the caller's identity from its claims.
func (v *GoogleVerifier) Verify(ctx context.Context, rawIDToken string) (ports.GoogleIdentity, error) {
	payload, err := idtoken.Validate(ctx, rawIDToken, v.clientID)
	if err != nil {
		return ports.GoogleIdentity{}, fmt.Errorf("%w: %v", ports.ErrInvalidGoogleToken, err)
	}

	email, _ := payload.Claims["email"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)
	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	return ports.GoogleIdentity{
		Subject:       payload.Subject,
		Email:         email,
		EmailVerified: emailVerified,
		Name:          name,
		Picture:       picture,
	}, nil
}
