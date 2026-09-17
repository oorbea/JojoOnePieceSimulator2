// Package gameinvite provides infrastructure implementations of
// ports.IGameInviteStore. MemoryStore is the single-instance, in-process
// default (mirrors streamticket's Memory/Redis split); a Redis-backed
// adapter lives in the redis subpackage for multi-instance deployments.
package gameinvite

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// inviteTokenBytes mirrors streamticket's entropy budget - 32 bytes (256
// bits) base64url-encoded to 43 characters.
const inviteTokenBytes = 32

type entry struct {
	invite    ports.GameInvite
	expiresAt time.Time
}

// Config configures MemoryStore.
type Config struct {
	// TTL bounds how long a minted invite stays redeemable.
	TTL time.Duration
}

// MemoryStore is an in-memory ports.IGameInviteStore, safe for concurrent
// use. Fine for a single backend instance, same reasoning already accepted
// for streamticket.MemoryStore: an invite lives minutes, and losing the
// whole table on restart costs nothing worse than a share link needing to
// be re-minted.
type MemoryStore struct {
	mu       sync.RWMutex // Lookup only reads: unlike a burn, it never mutates.
	invites  map[string]entry
	ttl      time.Duration
	nowFunc  func() time.Time
	randRead func([]byte) (int, error)
}

// NewMemoryStore builds an empty MemoryStore that mints invites valid for
// cfg.TTL.
func NewMemoryStore(cfg Config) *MemoryStore {
	return &MemoryStore{
		invites:  make(map[string]entry),
		ttl:      cfg.TTL,
		nowFunc:  time.Now,
		randRead: rand.Read,
	}
}

var _ ports.IGameInviteStore = (*MemoryStore)(nil)

// Issue mints a fresh multi-use token for inv.
func (s *MemoryStore) Issue(_ context.Context, inv ports.GameInvite) (string, time.Time, error) {
	buf := make([]byte, inviteTokenBytes)
	if _, err := s.randRead(buf); err != nil {
		return "", time.Time{}, fmt.Errorf("generating game invite: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(buf)
	expiresAt := s.nowFunc().Add(s.ttl)

	s.mu.Lock()
	s.invites[token] = entry{invite: inv, expiresAt: expiresAt}
	s.mu.Unlock()

	return token, expiresAt, nil
}

// Lookup returns the invite named by token without consuming it - the
// multi-use property the port doc requires.
func (s *MemoryStore) Lookup(_ context.Context, token string) (ports.GameInvite, error) {
	s.mu.RLock()
	e, ok := s.invites[token]
	s.mu.RUnlock()

	if !ok {
		return ports.GameInvite{}, ports.ErrInviteInvalid
	}
	if s.nowFunc().After(e.expiresAt) {
		return ports.GameInvite{}, ports.ErrInviteInvalid
	}
	return e.invite, nil
}

// DeleteExpired removes every invite past its expiry, returning how many
// were removed. Exported only on the concrete type (not part of
// ports.IGameInviteStore), same reasoning as streamticket.MemoryStore's own
// DeleteExpired: Redis expires its own keys and has no use for it.
func (s *MemoryStore) DeleteExpired() int {
	now := s.nowFunc()
	s.mu.Lock()
	defer s.mu.Unlock()

	n := 0
	for token, e := range s.invites {
		if now.After(e.expiresAt) {
			delete(s.invites, token)
			n++
		}
	}
	return n
}
