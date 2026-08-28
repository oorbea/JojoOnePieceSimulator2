package game_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func TestGauntletMode_FallMajorityEndsInDefeat(t *testing.T) {
	g, players := newGauntletGame(t, someStages(t), 1)
	assignAndOpenVoting(t, g)

	if err := g.CastVote(players[0].ID(), "FALL"); err != nil {
		t.Fatalf("CastVote: %v", err)
	}
	if tied, err := g.CloseVoting(); err != nil || tied {
		t.Fatalf("CloseVoting: tied=%v err=%v", tied, err)
	}
	if err := g.CompleteRound(); err != nil {
		t.Fatalf("CompleteRound: %v", err)
	}
	if g.State() != enums.Finished {
		t.Fatalf("expected FINISHED after a FALL majority, got %v", g.State())
	}
	result, err := g.Result()
	if err != nil {
		t.Fatalf("Result: %v", err)
	}
	if result.Winner != game.OptionID("FALL") {
		t.Fatalf("expected a FALL outcome, got %v", result.Winner)
	}
	if result.RoundsPlayed != 1 {
		t.Fatalf("expected the run to stop at round 1, got %d rounds", result.RoundsPlayed)
	}
}

func TestGauntletMode_SurviveMajorityAdvances(t *testing.T) {
	g, players := newGauntletGame(t, someStages(t), 1) // 3 stages
	assignAndOpenVoting(t, g)

	if err := g.CastVote(players[0].ID(), "SURVIVE"); err != nil {
		t.Fatalf("CastVote: %v", err)
	}
	if tied, err := g.CloseVoting(); err != nil || tied {
		t.Fatalf("CloseVoting: tied=%v err=%v", tied, err)
	}
	if err := g.CompleteRound(); err != nil {
		t.Fatalf("CompleteRound: %v", err)
	}
	if g.State() != enums.Assigning {
		t.Fatalf("expected ASSIGNING for the next round, got %v", g.State())
	}
	if len(g.Rounds()) != 1 {
		t.Fatalf("expected 1 round recorded so far, got %d", len(g.Rounds()))
	}
}

func TestGauntletMode_ClearingEveryStageIsVictory(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 1) // single stage
	assignAndOpenVoting(t, g)

	if err := g.CastVote(players[0].ID(), "SURVIVE"); err != nil {
		t.Fatalf("CastVote: %v", err)
	}
	if _, err := g.CloseVoting(); err != nil {
		t.Fatalf("CloseVoting: %v", err)
	}
	if err := g.CompleteRound(); err != nil {
		t.Fatalf("CompleteRound: %v", err)
	}
	if g.State() != enums.Finished {
		t.Fatalf("expected FINISHED once the only stage is cleared, got %v", g.State())
	}
	result, err := g.Result()
	if err != nil {
		t.Fatalf("Result: %v", err)
	}
	if result.Winner != game.OptionID("SURVIVE") {
		t.Fatalf("expected a SURVIVE outcome, got %v", result.Winner)
	}
}

func TestGauntletMode_TieOpensRevoteThenExternalTiebreak(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 2)
	assignAndOpenVoting(t, g)

	if err := g.CastVote(players[0].ID(), "SURVIVE"); err != nil {
		t.Fatalf("CastVote: %v", err)
	}
	if err := g.CastVote(players[1].ID(), "FALL"); err != nil {
		t.Fatalf("CastVote: %v", err)
	}
	tied, err := g.CloseVoting()
	if err != nil || !tied {
		t.Fatalf("expected a tie on first close, tied=%v err=%v", tied, err)
	}
	if g.State() != enums.Tiebreak {
		t.Fatalf("expected TIEBREAK after the first tie, got %v", g.State())
	}

	// The revote also ties.
	tied, err = g.CloseVoting()
	if err != nil || !tied {
		t.Fatalf("expected a tie on the revote too, tied=%v err=%v", tied, err)
	}

	// The application layer resolves it externally (a coin flip today, an
	// LLM later - see ports.ITiebreaker) and feeds the winner back in.
	if err := g.ResolveTiebreak("SURVIVE"); err != nil {
		t.Fatalf("ResolveTiebreak: %v", err)
	}
	rounds := g.Rounds()
	last := rounds[len(rounds)-1]
	if last.Result == nil || !last.Result.DecidedByCoinFlip {
		t.Fatalf("expected the round to be marked as externally (coin-flip) decided")
	}
	if last.Result.Winner != "SURVIVE" {
		t.Fatalf("expected the tiebreak winner to be recorded, got %v", last.Result.Winner)
	}
}

func TestGauntletMode_LoadoutsNotReassignedBetweenRounds(t *testing.T) {
	if (game.GauntletMode{}).ReassignsEachRound() {
		t.Fatalf("expected Gauntlet to not reassign loadouts each round")
	}
}
