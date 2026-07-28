package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// jwtClaims is our own access token's payload: the standard registered
// claims (sub, iss, iat, exp) plus a "role" claim, which is what
// distinguishes an admin token from a regular one - both share the same
// format and signing key.
type jwtClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// JWTIssuer issues and parses HS256 access tokens for ports.ITokenIssuer.
type JWTIssuer struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

var _ ports.ITokenIssuer = (*JWTIssuer)(nil)

func NewJWTIssuer(secret []byte, issuer string, ttl time.Duration) *JWTIssuer {
	return &JWTIssuer{secret: secret, issuer: issuer, ttl: ttl}
}

// Issue signs a new access token for u.
func (j *JWTIssuer) Issue(u *user.User) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(j.ttl)

	claims := jwtClaims{
		Role: u.Role().String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID().String(),
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(j.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("signing access token: %w", err)
	}
	return token, expiresAt, nil
}

// Parse validates rawToken's signature, algorithm, issuer and expiry, and
// extracts its claims.
func (j *JWTIssuer) Parse(rawToken string) (ports.Claims, error) {
	var claims jwtClaims
	_, err := jwt.ParseWithClaims(rawToken, &claims, func(t *jwt.Token) (any, error) {
		return j.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(j.issuer),
		jwt.WithExpirationRequired())
	if err != nil {
		return ports.Claims{}, fmt.Errorf("%w: %v", ports.ErrUnauthenticated, err)
	}

	userID, err := user.ParseUserID(claims.Subject)
	if err != nil {
		return ports.Claims{}, fmt.Errorf("%w: %v", ports.ErrUnauthenticated, err)
	}
	role, err := enums.ParseUserRole(claims.Role)
	if err != nil {
		return ports.Claims{}, fmt.Errorf("%w: %v", ports.ErrUnauthenticated, err)
	}

	return ports.Claims{UserID: userID, Role: role}, nil
}
