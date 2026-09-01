// Package gamestore provides infrastructure implementations of
// ports.IGameStore. MemoryGameStore is the only one today - a
// single-instance, in-process map. A Redis-backed adapter behind the same
// port is the next tanda; nothing outside this package should assume more
// than IGameStore promises.
package gamestore

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

type entry struct {
	game      *game.Game
	code      string
	updatedAt time.Time
	// ttl, when non-zero, overrides the reaper's olderThan for this entry
	// alone - set by SaveWithTTL so a terminal game expires on its own much
	// shorter schedule instead of lingering for the full lobby TTL.
	ttl time.Duration
}

// MemoryGameStore is an in-memory ports.IGameStore, safe for concurrent use.
// Fine for a single backend instance with no multi-instance fan-out need -
// the same reasoning as services.PictureEventHub.
type MemoryGameStore struct {
	mu      sync.RWMutex
	byID    map[game.GameID]*entry
	byCode  map[string]game.GameID
	nowFunc func() time.Time
}

// NewMemoryGameStore builds an empty MemoryGameStore.
func NewMemoryGameStore() *MemoryGameStore {
	return &MemoryGameStore{
		byID:    make(map[game.GameID]*entry),
		byCode:  make(map[string]game.GameID),
		nowFunc: time.Now,
	}
}

var _ ports.IGameStore = (*MemoryGameStore)(nil)

func (s *MemoryGameStore) Create(_ context.Context, code string, g *game.Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byCode[code]; exists {
		return ports.ErrGameCodeTaken
	}
	s.byID[g.ID()] = &entry{game: g, code: code, updatedAt: s.nowFunc()}
	s.byCode[code] = g.ID()
	return nil
}

func (s *MemoryGameStore) Get(_ context.Context, id game.GameID) (*game.Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.byID[id]
	if !ok {
		return nil, ports.ErrGameNotFound
	}
	return e.game, nil
}

func (s *MemoryGameStore) GetByCode(_ context.Context, code string) (*game.Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.byCode[code]
	if !ok {
		return nil, ports.ErrGameNotFound
	}
	e, ok := s.byID[id]
	if !ok {
		return nil, ports.ErrGameNotFound
	}
	return e.game, nil
}

func (s *MemoryGameStore) Code(_ context.Context, id game.GameID) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.byID[id]
	if !ok {
		return "", ports.ErrGameNotFound
	}
	return e.code, nil
}

func (s *MemoryGameStore) Save(ctx context.Context, g *game.Game) error {
	return s.SaveWithTTL(ctx, g, 0)
}

// SaveWithTTL implements ports.IGameStore. This store has no per-key
// expiry of its own - the Reaper drives expiry by calling DeleteExpired
// with the configured lobby TTL - so the override is recorded on the entry
// and honored there instead.
func (s *MemoryGameStore) SaveWithTTL(_ context.Context, g *game.Game, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.byID[g.ID()]
	if !ok {
		return ports.ErrGameNotFound
	}
	e.game = g
	e.updatedAt = s.nowFunc()
	e.ttl = ttl
	return nil
}

func (s *MemoryGameStore) Delete(_ context.Context, id game.GameID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.byID[id]
	if !ok {
		return nil
	}
	delete(s.byCode, e.code)
	delete(s.byID, id)
	return nil
}

func (s *MemoryGameStore) DeleteExpired(_ context.Context, olderThan time.Duration) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.nowFunc()
	removed := 0
	for id, e := range s.byID {
		// An entry saved with its own shorter TTL (a terminal game, see
		// SaveWithTTL) expires on that schedule instead of olderThan.
		lifetime := olderThan
		if e.ttl > 0 {
			lifetime = e.ttl
		}
		if e.updatedAt.Before(now.Add(-lifetime)) {
			delete(s.byCode, e.code)
			delete(s.byID, id)
			removed++
		}
	}
	return removed
}

// ListPublic implements ports.IGameStore. No SCAN ban applies here (that
// constraint is Redis-specific) - a plain map scan is fine for a
// single-instance in-memory store.
func (s *MemoryGameStore) ListPublic(_ context.Context, limit int) ([]*game.Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := make([]*entry, 0)
	for _, e := range s.byID {
		if e.game.IsPubliclyJoinable() {
			entries = append(entries, e)
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].updatedAt.After(entries[j].updatedAt)
	})
	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}
	out := make([]*game.Game, len(entries))
	for i, e := range entries {
		out[i] = e.game
	}
	return out, nil
}
