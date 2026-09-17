package ports

import (
	"context"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
)

// GameInvite is what a redeemed invite token proves: which Game it points
// at, the join code that was current when it was minted (the revocation
// mechanism - see IGameInviteStore's doc), and who minted it.
type GameInvite struct {
	GameID    game.GameID
	Code      string
	CreatedBy user.UserID
}

// IGameInviteStore mints and looks up short-lived invite tokens for a
// shareable "join this lobby" link. Unlike IStreamTicketStore, Lookup is
// NOT a burn: an invite link is deliberately multi-use (it may be pasted
// into a group chat), so every redemption only reads the entry, never
// deletes it. Do not "fix" this into a single-use Redeem - that would break
// the product requirement.
//
// Revocation on join-code rotation is achieved without any store-side
// invalidation call: GameInvite.Code freezes the join code at mint time,
// and the caller (GameService.JoinByInvite) compares it against the game's
// *current* code (IGameStore.Code) at redemption time. A rotated code
// makes every previously minted invite for that game compare unequal and
// therefore dead, with no second write needed inside RegenerateGameCode's
// critical section - see ObsidianVault/ADR.md for why a per-game token set
// (which would need that second write, and fails open on a Redis error)
// was rejected in favour of this comparison.
type IGameInviteStore interface {
	// Issue mints a fresh multi-use token for inv, valid for the store's own
	// configured TTL.
	Issue(ctx context.Context, inv GameInvite) (token string, expiresAt time.Time, err error)

	// Lookup returns the invite named by token without consuming it. Unknown
	// or expired both return ErrInviteInvalid without distinguishing which.
	Lookup(ctx context.Context, token string) (GameInvite, error)
}
