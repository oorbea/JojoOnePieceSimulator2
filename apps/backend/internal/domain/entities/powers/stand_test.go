package powers_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func mustStand(t *testing.T, id byte, name string, evolvesFrom *powers.Stand) *powers.Stand {
	t.Helper()
	skills := []string{"skill"}
	power, err := powers.NewPower(powers.PowerID{id}, name, "description", enums.Common, &skills, "")
	if err != nil {
		t.Fatalf("NewPower(%q): %v", name, err)
	}
	stand, err := powers.NewStand(*power, enums.A, enums.A, enums.A, enums.A, enums.A, enums.A, evolvesFrom)
	if err != nil {
		t.Fatalf("NewStand(%q): %v", name, err)
	}
	return stand
}

func TestEvolutionDepth(t *testing.T) {
	starPlatinum := mustStand(t, 1, "Star Platinum", nil)
	if got := starPlatinum.EvolutionDepth(); got != 0 {
		t.Fatalf("base stand: EvolutionDepth() = %d, want 0", got)
	}

	silverChariot := mustStand(t, 2, "Silver Chariot", nil)
	chariotRequiem := mustStand(t, 3, "Chariot Requiem", silverChariot)
	if got := chariotRequiem.EvolutionDepth(); got != 1 {
		t.Fatalf("one-step evolution: EvolutionDepth() = %d, want 1", got)
	}

	echoesHuevo := mustStand(t, 4, "Echoes", nil)
	echoesAct1 := mustStand(t, 5, "Echoes ACT1", echoesHuevo)
	echoesAct2 := mustStand(t, 6, "Echoes ACT2", echoesAct1)
	echoesAct3 := mustStand(t, 7, "Echoes ACT3", echoesAct2)
	if got := echoesAct3.EvolutionDepth(); got != 3 {
		t.Fatalf("four-stage chain: EvolutionDepth() = %d, want 3", got)
	}
}
