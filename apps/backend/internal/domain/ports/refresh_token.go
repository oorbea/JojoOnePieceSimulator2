package ports

import (
	"context"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// RefreshToken is what a redeemed refresh token proves: the identity of the
// user it was minted for, plus a FamilyID stable across the whole rotation
// chain (every token minted from one login, and every token minted by
// rotating it) so the whole chain can be revoked in one call - on logout,
// or the instant a reused (already-burned) token is redeemed again.
type RefreshToken struct {
	UserID   user.UserID
	Role     enums.UserRole
	FamilyID string
}

// IRefreshTokenStore mints and redeems long-lived, single-use, rotating
// refresh tokens. Issue is called at login (FamilyID == "" starts a new
// family) and again on every successful Redeem (same FamilyID, so the chain
// stays linked). Redeem is called once per refresh attempt and must never
// let two concurrent redemptions of the same token both succeed - if a
// token is redeemed a second time (replay), the ENTIRE family must be
// revoked, since that second redemption proves the token leaked.
type IRefreshTokenStore interface {
	// Issue mints a fresh single-use token for t, valid for the store's own
	// configured TTL.
	Issue(ctx context.Context, t RefreshToken) (token string, expiresAt time.Time, err error)

	// Redeem atomically consumes token: unknown, expired, or a token whose
	// family was revoked all return ErrRefreshInvalid. A token redeemed a
	// second time returns ErrRefreshReuse, and the whole family it belongs
	// to is revoked as a side effect.
	Redeem(ctx context.Context, token string) (RefreshToken, error)

	// RevokeFamily kills every token minted under familyID, present or
	// future. Called on logout, and internally by Redeem when it detects a
	// replay.
	RevokeFamily(ctx context.Context, familyID string) error
}
