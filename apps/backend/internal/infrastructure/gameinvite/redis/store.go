// Package redis provides a Redis-backed ports.IGameInviteStore, for
// multi-instance deployments where the process minting an invite (behind a
// normal authenticated POST) and the process redeeming it (a different
// visitor, possibly hours later against a different instance) may not be
// the same one. Mirrors streamticket/redis's Config/New/opContext shape.
package redis

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// inviteTokenBytes mirrors gameinvite.MemoryStore's entropy budget - 32
// bytes (256 bits), base64url-encoded to 43 characters.
const inviteTokenBytes = 32

// Config holds everything needed to reach the Redis instance backing the
// game invite store.
type Config struct {
	// URL is a redis:// or rediss:// connection string.
	URL string
	// DialTimeout bounds the initial connection + PING done by New.
	DialTimeout time.Duration
	// OpTimeout bounds every individual store operation. Fail-closed, like
	// streamticket/redis - an invite store is a source of truth for whether
	// a join may proceed, not an optional speedup.
	OpTimeout time.Duration
	// TTL is how long a minted invite stays redeemable.
	TTL time.Duration
}

// Store is a Redis-backed ports.IGameInviteStore. Fail-closed: every error
// is returned to the caller, never swallowed as a miss.
//
//	jojo:game-invite:<token> -> the envelope JSON (GameInvite)
//
// Shares the "jojo:" root with gamestore/redis and streamticket/redis but
// not infrastructure/cache's <ns>:<gen>: layout, for the same reason
// neither of those does: a cache generation bump must not be able to
// orphan a live invite.
type Store struct {
	client    *goredis.Client
	opTimeout time.Duration
	ttl       time.Duration
}

var _ ports.IGameInviteStore = (*Store)(nil)

func inviteKey(token string) string { return "jojo:game-invite:" + token }

// New connects to the Redis instance described by cfg and verifies
// reachability with a PING bounded by cfg.DialTimeout, so a misconfigured
// deploy fails at boot instead of silently minting invites nobody can ever
// redeem.
func New(ctx context.Context, cfg Config) (*Store, error) {
	opts, err := goredis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parsing redis url: %w", err)
	}
	client := goredis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, cfg.DialTimeout)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("pinging redis: %w", err)
	}

	return &Store{client: client, opTimeout: cfg.OpTimeout, ttl: cfg.TTL}, nil
}

func (s *Store) opContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, s.opTimeout)
}

// Issue implements ports.IGameInviteStore.
func (s *Store) Issue(ctx context.Context, inv ports.GameInvite) (string, time.Time, error) {
	buf := make([]byte, inviteTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", time.Time{}, fmt.Errorf("generating game invite: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(buf)

	payload, err := encode(inv)
	if err != nil {
		return "", time.Time{}, err
	}

	expiresAt := time.Now().Add(s.ttl)

	opCtx, cancel := s.opContext(ctx)
	defer cancel()

	// NX: a collision across 256 bits of entropy would mean something is
	// badly wrong (a broken RNG) - surfacing it as an error is correct,
	// never silently overwriting another caller's live invite.
	ok, err := s.client.SetNX(opCtx, inviteKey(token), payload, s.ttl).Result()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("storing game invite: %w", err)
	}
	if !ok {
		return "", time.Time{}, errors.New("game invite token collision")
	}
	return token, expiresAt, nil
}

// Lookup implements ports.IGameInviteStore. Unlike streamticket/redis's
// Redeem, this is a plain GET - no Lua script, no deletion: an invite is
// multi-use, so looking it up must never consume it.
func (s *Store) Lookup(ctx context.Context, token string) (ports.GameInvite, error) {
	opCtx, cancel := s.opContext(ctx)
	defer cancel()

	payload, err := s.client.Get(opCtx, inviteKey(token)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return ports.GameInvite{}, ports.ErrInviteInvalid
		}
		return ports.GameInvite{}, fmt.Errorf("looking up game invite: %w", err)
	}
	inv, err := decode(payload)
	if err != nil {
		return ports.GameInvite{}, fmt.Errorf("decoding game invite: %w", err)
	}
	return inv, nil
}

// Close releases the underlying Redis client.
func (s *Store) Close() error {
	return s.client.Close()
}
