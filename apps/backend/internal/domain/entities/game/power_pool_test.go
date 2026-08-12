package game_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func TestAvailablePowers_DrawStandOutOfRange(t *testing.T) {
	pool := game.NewAvailablePowers(nil, nil)
	if _, err := pool.DrawStand(0); err != game.ErrPowerPoolExhausted {
		t.Fatalf("expected ErrPowerPoolExhausted, got %v", err)
	}
}

func TestAvailablePowers_DrawRemovesFromPool(t *testing.T) {
	stand := mustStand(t, 1, "Star Platinum", enums.Legendary)
	pool := game.NewAvailablePowers([]*powers.Stand{stand}, nil)

	drawn, err := pool.DrawStand(0)
	if err != nil {
		t.Fatalf("DrawStand: %v", err)
	}
	if drawn.ID() != stand.ID() {
		t.Fatalf("expected to draw the seeded stand")
	}
	if len(pool.Stands()) != 0 {
		t.Fatalf("expected pool to be empty after the draw, got %d", len(pool.Stands()))
	}
	if _, err := pool.DrawStand(0); err != game.ErrPowerPoolExhausted {
		t.Fatalf("expected ErrPowerPoolExhausted on a second draw, got %v", err)
	}
}

func TestAvailablePowers_DrawDevilFruitOutOfRange(t *testing.T) {
	pool := game.NewAvailablePowers(nil, nil)
	if _, err := pool.DrawDevilFruit(0); err != game.ErrPowerPoolExhausted {
		t.Fatalf("expected ErrPowerPoolExhausted, got %v", err)
	}
}
