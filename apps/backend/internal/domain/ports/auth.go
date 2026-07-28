package ports

import (
	"context"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// GoogleIdentity is the subset of a verified Google ID token the domain
// cares about.
type GoogleIdentity struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
}

// IGoogleTokenVerifier verifies a raw Google ID token and extracts the
// caller's identity, keeping the concrete verification strategy (JWKS
// fetch/cache, audience check, ...) out of the domain layer.
type IGoogleTokenVerifier interface {
	Verify(ctx context.Context, rawIDToken string) (GoogleIdentity, error)
}

// Claims is the authenticated identity carried by our own access tokens.
type Claims struct {
	UserID user.UserID
	Role   enums.UserRole
}

// ITokenIssuer issues and parses the backend's own access tokens, so the
// domain and application layers never depend on a specific token format.
type ITokenIssuer interface {
	Issue(u *user.User) (token string, expiresAt time.Time, err error)
	Parse(token string) (Claims, error)
}
