// Package gamestore provides infrastructure implementations of
// ports.IGameStore. MemoryGameStore is the only one today - a
// single-instance, in-process map. A Redis-backed adapter behind the same
// port is the next tanda; nothing outside this package should assume more
// than IGameStore promises.
package gamestore

import (
	"context"
	"sync"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

type entry struct {
	game      *game.Game
	code      string
	updatedAt time.Time
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

func (s *MemoryGameStore) Save(_ context.Context, g *game.Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.byID[g.ID()]
	if !ok {
		return ports.ErrGameNotFound
	}
	e.game = g
	e.updatedAt = s.nowFunc()
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

	cutoff := s.nowFunc().Add(-olderThan)
	removed := 0
	for id, e := range s.byID {
		if e.updatedAt.Before(cutoff) {
			delete(s.byCode, e.code)
			delete(s.byID, id)
			removed++
		}
	}
	return removed
}
