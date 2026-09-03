// Package streamticket provides infrastructure implementations of
// ports.IStreamTicketStore. MemoryStore is the single-instance, in-process
// default (mirrors gamestore's Memory/Redis split); a Redis-backed adapter
// lives in the redis subpackage for multi-instance deployments.
package streamticket

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// ticketTokenBytes is how much entropy backs each minted ticket - 32 bytes
// (256 bits) base64url-encoded to 43 characters, the same margin a
// cryptographically random session identifier would use.
const ticketTokenBytes = 32

type entry struct {
	ticket    ports.StreamTicket
	expiresAt time.Time
}

// Config configures MemoryStore.
type Config struct {
	// TTL bounds how long a minted ticket stays redeemable.
	TTL time.Duration
}

// MemoryStore is an in-memory ports.IStreamTicketStore, safe for concurrent
// use. Fine for a single backend instance - the same reasoning already
// accepted for gamestore.MemoryGameStore and services.PictureEventHub: a
// ticket lives only seconds, and losing the whole table on restart costs
// nothing since every open stream is torn down on shutdown anyway.
type MemoryStore struct {
	mu       sync.Mutex // not RWMutex: Redeem always writes (it deletes).
	tickets  map[string]entry
	ttl      time.Duration
	nowFunc  func() time.Time
	randRead func([]byte) (int, error)
}

// NewMemoryStore builds an empty MemoryStore that mints tickets valid for
// cfg.TTL.
func NewMemoryStore(cfg Config) *MemoryStore {
	return &MemoryStore{
		tickets:  make(map[string]entry),
		ttl:      cfg.TTL,
		nowFunc:  time.Now,
		randRead: rand.Read,
	}
}

var _ ports.IStreamTicketStore = (*MemoryStore)(nil)

// Issue mints a fresh single-use token for t.
func (s *MemoryStore) Issue(_ context.Context, t ports.StreamTicket) (string, time.Time, error) {
	buf := make([]byte, ticketTokenBytes)
	if _, err := s.randRead(buf); err != nil {
		return "", time.Time{}, fmt.Errorf("generating stream ticket: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(buf)
	expiresAt := s.nowFunc().Add(s.ttl)

	s.mu.Lock()
	s.tickets[token] = entry{ticket: t, expiresAt: expiresAt}
	s.mu.Unlock()

	return token, expiresAt, nil
}

// Redeem atomically consumes token. The entry is deleted unconditionally
// before its expiry is even checked - that single critical section is what
// makes redemption single-use under concurrency: whichever goroutine's
// delete actually removes the map entry is the only one that can possibly
// succeed, and every other caller (including a legitimate expiry check)
// sees it already gone.
func (s *MemoryStore) Redeem(_ context.Context, token string) (ports.StreamTicket, error) {
	s.mu.Lock()
	e, ok := s.tickets[token]
	delete(s.tickets, token)
	s.mu.Unlock()

	if !ok {
		return ports.StreamTicket{}, ports.ErrTicketInvalid
	}
	if s.nowFunc().After(e.expiresAt) {
		return ports.StreamTicket{}, ports.ErrTicketInvalid
	}
	return e.ticket, nil
}

// DeleteExpired removes every ticket past its expiry, returning how many
// were removed. Exported only on the concrete type (not part of
// ports.IStreamTicketStore) since Redis expires its own keys and has no use
// for it - mirrors why gamestore.MemoryGameStore.DeleteExpired takes a ttl
// argument that Redis just ignores via TTL-on-write instead.
func (s *MemoryStore) DeleteExpired() int {
	now := s.nowFunc()
	s.mu.Lock()
	defer s.mu.Unlock()

	n := 0
	for token, e := range s.tickets {
		if now.After(e.expiresAt) {
			delete(s.tickets, token)
			n++
		}
	}
	return n
}
