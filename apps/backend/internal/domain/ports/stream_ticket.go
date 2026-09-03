package ports

import (
	"context"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// TicketPurpose distinguishes which stream a StreamTicket may open, so a
// ticket minted for one stream can never be redeemed against another.
type TicketPurpose string

const (
	TicketPurposeEvents TicketPurpose = "events"
	TicketPurposeGameWS TicketPurpose = "game-ws"
)

// StreamTicket is what a redeemed ticket proves: the identity of the caller
// that minted it, at mint time, plus which stream (and, for a game socket,
// which game) it may open. A short TTL and single-use redemption are the
// store's job, not this type's.
type StreamTicket struct {
	UserID   user.UserID
	Role     enums.UserRole
	Purpose  TicketPurpose
	Resource string // game id for TicketPurposeGameWS, "" for TicketPurposeEvents
}

// IStreamTicketStore mints and redeems short-lived, single-use tickets that
// authenticate a browser EventSource/WebSocket connection, replacing a full
// access token in the URL's query string. Issue is called from a normal
// authenticated POST (over Authorization: Bearer); Redeem is called once,
// from the stream handler itself, and must never let two concurrent
// redemptions of the same token both succeed.
type IStreamTicketStore interface {
	// Issue mints a fresh single-use token for t, valid for the store's own
	// configured TTL. Mirrors ITokenIssuer.Issue's (token, expiresAt, error)
	// shape.
	Issue(ctx context.Context, t StreamTicket) (token string, expiresAt time.Time, err error)

	// Redeem atomically consumes token: unknown, expired, and
	// already-redeemed all return ErrTicketInvalid without distinguishing
	// which, and the token is gone afterwards regardless of the outcome.
	Redeem(ctx context.Context, token string) (StreamTicket, error)
}
