package services

import (
	"context"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// CreateInvite mints a fresh multi-use "join this lobby" link for gameID,
// on behalf of userID, who must already be a seated human participant -
// any lobby member may share the room, not only the host (unlike
// RegenerateGameCode, which is host-only). Refuses once the lobby has left
// LOBBY: a link to a match already in progress has nothing useful to do.
func (s *GameService) CreateInvite(ctx context.Context, gameID game.GameID, userID user.UserID) (string, time.Time, error) {
	g, err := s.GetGame(ctx, gameID)
	if err != nil {
		return "", time.Time{}, err
	}
	if !isSeatedHuman(g, userID) {
		return "", time.Time{}, ports.ErrForbidden
	}
	if g.State() != enums.Lobby {
		return "", time.Time{}, ErrGameAlreadyStarted
	}
	code, err := s.store.Code(ctx, gameID)
	if err != nil {
		return "", time.Time{}, err
	}
	return s.invites.Issue(ctx, ports.GameInvite{GameID: gameID, Code: code, CreatedBy: userID})
}

// InviteStatus reports only whether token is still redeemable, for the
// PUBLIC status route an unfurl bot or an unauthenticated visitor may hit
// before ever logging in. It deliberately checks nothing about the target
// lobby beyond the invite's own validity and whether the host has since
// rotated the join code (see ports.IGameInviteStore's doc for how that
// revocation is detected) - state, fullness, and lock are all withheld,
// since exposing any of them here would leak a lobby fact to a caller who
// has proven nothing but possession of a URL. It never returns an error:
// every failure mode collapses into enums.InviteExpired.
func (s *GameService) InviteStatus(ctx context.Context, token string) enums.InviteStatus {
	inv, err := s.invites.Lookup(ctx, token)
	if err != nil {
		return enums.InviteExpired
	}
	currentCode, err := s.store.Code(ctx, inv.GameID)
	if err != nil || currentCode != inv.Code {
		return enums.InviteExpired
	}
	return enums.InviteValid
}

// InvitePreview returns a LobbyListing for the Game token names, for an
// authenticated caller who has not yet joined - the token is the
// credential, exactly as the join code itself is for PreviewByCode, and
// like PreviewByCode this works for PRIVATE lobbies too.
func (s *GameService) InvitePreview(ctx context.Context, token string) (LobbyListing, error) {
	inv, currentCode, err := s.resolveInvite(ctx, token)
	if err != nil {
		return LobbyListing{}, err
	}
	if currentCode != inv.Code {
		return LobbyListing{}, ErrInviteRevoked
	}
	g, err := s.store.Get(ctx, inv.GameID)
	if err != nil {
		return LobbyListing{}, err
	}
	return newLobbyListing(g), nil
}

// JoinByInvite seats userID as a new human participant in the Game token
// names. See JoinByCode's doc for the shared seating logic; JoinByInvite
// differs from it in three deliberate ways documented inline below.
func (s *GameService) JoinByInvite(ctx context.Context, token string, userID user.UserID) (*game.Game, error) {
	inv, err := s.invites.Lookup(ctx, token)
	if err != nil {
		return nil, err
	}
	return s.withGame(ctx, inv.GameID, func(g *game.Game) error {
		// Re-check the code from inside this Game's lock, serialized against
		// a concurrent RegenerateGameCode on this same instance - a rotated
		// code makes every invite minted against the old one dead, with no
		// second write anywhere (see ports.IGameInviteStore's doc).
		currentCode, err := s.store.Code(ctx, g.ID())
		if err != nil {
			return err
		}
		if currentCode != inv.Code {
			return ErrInviteRevoked
		}
		// Idempotent: unlike JoinByCode/JoinByID, a link may be clicked more
		// than once by someone already seated (a stale browser tab, a
		// double-tap) - re-erroring would strand the frontend with no game
		// id to route to, since AppError bodies carry none.
		if isSeatedHuman(g, userID) {
			return nil
		}
		// A specific, link-friendly error instead of letting g.Join's own
		// ErrInvalidStateTransition surface - the link's holder gets told
		// exactly why, not a generic state-machine error.
		if g.State() != enums.Lobby {
			return ErrGameAlreadyStarted
		}
		// Deliberately NOT gated by Config().Visibility() the way JoinByID
		// is (game.ErrLobbyPrivate): the invite token is itself the secret,
		// exactly like the join code is for JoinByCode, so a PRIVATE lobby
		// is just as joinable through a link as through its code.
		return s.joinLocked(ctx, g, userID)
	})
}

// resolveInvite looks up token and the invite's target game's *current*
// join code in one place, translating a vanished game the same way as an
// unknown token - callers must not distinguish the two outward (see
// ports.GameInvite's doc), only ErrInviteRevoked (code mismatch, i.e. a
// still-alive game whose code was rotated) is kept distinct.
func (s *GameService) resolveInvite(ctx context.Context, token string) (ports.GameInvite, string, error) {
	inv, err := s.invites.Lookup(ctx, token)
	if err != nil {
		return ports.GameInvite{}, "", err
	}
	currentCode, err := s.store.Code(ctx, inv.GameID)
	if err != nil {
		return ports.GameInvite{}, "", ports.ErrInviteInvalid
	}
	return inv, currentCode, nil
}
