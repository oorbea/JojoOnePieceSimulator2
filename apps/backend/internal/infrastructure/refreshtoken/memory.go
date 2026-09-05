// Package refreshtoken provides infrastructure implementations of
// ports.IRefreshTokenStore. MemoryStore is the single-instance, in-process
// default (mirrors gamestore's Memory/Redis split); a Redis-backed adapter
// lives in the redis subpackage for multi-instance deployments.
package refreshtoken

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// refreshTokenBytes is how much entropy backs each minted refresh token -
// 32 bytes (256 bits) base64url-encoded to 43 characters, the same margin a
// cryptographically random session identifier would use.
const refreshTokenBytes = 32

type entry struct {
	token     ports.RefreshToken
	expiresAt time.Time
	used      bool
}

// Config configures MemoryStore.
type Config struct {
	// TTL bounds how long a minted refresh token stays redeemable.
	TTL time.Duration
}

// MemoryStore is an in-memory ports.IRefreshTokenStore, safe for concurrent
// use. Fine for a single backend instance - same reasoning already accepted
// for gamestore.MemoryGameStore and streamticket.MemoryStore.
//
// Unlike a stream ticket, a redeemed refresh token entry is NOT deleted:
// rotation requires detecting reuse (the same token redeemed twice), so the
// entry is kept - marked used - until its natural TTL expiry. Family
// liveness is tracked separately: families holds every FamilyID currently
// alive; its absence means the family was revoked (explicitly, or because a
// reuse was detected), and every token under it then fails with
// ErrRefreshInvalid regardless of its own used/expiry state.
type MemoryStore struct {
	mu       sync.Mutex
	tokens   map[string]entry
	families map[string]bool
	ttl      time.Duration
	nowFunc  func() time.Time
	randRead func([]byte) (int, error)
}

// NewMemoryStore builds an empty MemoryStore that mints refresh tokens
// valid for cfg.TTL.
func NewMemoryStore(cfg Config) *MemoryStore {
	return &MemoryStore{
		tokens:   make(map[string]entry),
		families: make(map[string]bool),
		ttl:      cfg.TTL,
		nowFunc:  time.Now,
		randRead: rand.Read,
	}
}

var _ ports.IRefreshTokenStore = (*MemoryStore)(nil)

// Issue mints a fresh single-use token for t, and marks t.FamilyID alive
// (idempotent - a no-op if it's already alive).
func (s *MemoryStore) Issue(_ context.Context, t ports.RefreshToken) (string, time.Time, error) {
	buf := make([]byte, refreshTokenBytes)
	if _, err := s.randRead(buf); err != nil {
		return "", time.Time{}, fmt.Errorf("generating refresh token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(buf)
	expiresAt := s.nowFunc().Add(s.ttl)

	s.mu.Lock()
	s.tokens[token] = entry{token: t, expiresAt: expiresAt}
	s.families[t.FamilyID] = true
	s.mu.Unlock()

	return token, expiresAt, nil
}

// Redeem atomically consumes token. Only one caller can ever observe
// !e.used and flip it under the same critical section, so exactly one
// concurrent Redeem of the same token succeeds; every other caller either
// sees the entry already used (reuse - revokes the family) or absent/
// expired/family-revoked (ErrRefreshInvalid).
func (s *MemoryStore) Redeem(_ context.Context, token string) (ports.RefreshToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.tokens[token]
	if !ok {
		return ports.RefreshToken{}, ports.ErrRefreshInvalid
	}
	if !s.families[e.token.FamilyID] {
		return ports.RefreshToken{}, ports.ErrRefreshInvalid
	}
	if e.used {
		delete(s.families, e.token.FamilyID)
		return ports.RefreshToken{}, ports.ErrRefreshReuse
	}
	if s.nowFunc().After(e.expiresAt) {
		return ports.RefreshToken{}, ports.ErrRefreshInvalid
	}

	e.used = true
	s.tokens[token] = e
	return e.token, nil
}

// RevokeFamily kills every token minted under familyID, present or future.
// Future Redeem calls against any of its tokens then fail with
// ErrRefreshInvalid, not ErrRefreshReuse - a revoked family and a replayed
// token are different situations, and only replay is ErrRefreshReuse.
func (s *MemoryStore) RevokeFamily(_ context.Context, familyID string) error {
	s.mu.Lock()
	delete(s.families, familyID)
	s.mu.Unlock()
	return nil
}

// DeleteExpired removes every token entry past its expiry, returning how
// many were removed. Exported only on the concrete type (not part of
// ports.IRefreshTokenStore) since Redis expires its own keys and has no use
// for it. Families have no independent expiry - they die only when
// explicitly revoked (or, for MemoryStore, on process restart).
func (s *MemoryStore) DeleteExpired() int {
	now := s.nowFunc()
	s.mu.Lock()
	defer s.mu.Unlock()

	n := 0
	for token, e := range s.tokens {
		if now.After(e.expiresAt) {
			delete(s.tokens, token)
			n++
		}
	}
	return n
}
