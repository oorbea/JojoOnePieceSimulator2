package game_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// TestGame_ResolveRound_HoldsInResolvingUntilCompleteRound proves the split
// added 2026-08-28: a clear winner parks the Game in RESOLVING with the
// Round's Result already set, and only CompleteRound advances it further -
// the pause that gives clients a window to render the round's outcome
// before the next sorteo starts.
func TestGame_ResolveRound_HoldsInResolvingUntilCompleteRound(t *testing.T) {
	g, players := newGauntletGame(t, someStages(t), 1)
	assignAndOpenVoting(t, g)

	if err := g.CastVote(players[0].ID(), "SURVIVE"); err != nil {
		t.Fatalf("CastVote: %v", err)
	}
	if tied, err := g.CloseVoting(); err != nil || tied {
		t.Fatalf("CloseVoting: tied=%v err=%v", tied, err)
	}
	if g.State() != enums.Resolving {
		t.Fatalf("state after a clear winner = %v, want RESOLVING", g.State())
	}
	rounds := g.Rounds()
	last := rounds[len(rounds)-1]
	if last.Result == nil || last.Result.Winner != "SURVIVE" {
		t.Fatalf("expected Result to already be set while RESOLVING, got %+v", last.Result)
	}

	if err := g.CompleteRound(); err != nil {
		t.Fatalf("CompleteRound: %v", err)
	}
	if g.State() != enums.Assigning {
		t.Fatalf("state after CompleteRound = %v, want ASSIGNING", g.State())
	}
}

// TestGame_CompleteRound_OnlyValidFromResolving guards CompleteRound
// against being called from any other state (e.g. twice in a row, or
// before a round has even resolved).
func TestGame_CompleteRound_OnlyValidFromResolving(t *testing.T) {
	g, players := newGauntletGame(t, someStages(t), 1)
	assignAndOpenVoting(t, g)

	if err := g.CompleteRound(); err != game.ErrInvalidStateTransition {
		t.Fatalf("CompleteRound while VOTING: err = %v, want ErrInvalidStateTransition", err)
	}

	if err := g.CastVote(players[0].ID(), "SURVIVE"); err != nil {
		t.Fatalf("CastVote: %v", err)
	}
	if _, err := g.CloseVoting(); err != nil {
		t.Fatalf("CloseVoting: %v", err)
	}
	if err := g.CompleteRound(); err != nil {
		t.Fatalf("first CompleteRound: %v", err)
	}
	if err := g.CompleteRound(); err != game.ErrInvalidStateTransition {
		t.Fatalf("second CompleteRound: err = %v, want ErrInvalidStateTransition", err)
	}
}

// TestGame_CloseVoting_TieCapturesTiedVotesThenResetsBallot proves the
// owner's explicit call (2026-08-28): the votes that tied are preserved on
// Round.TiedVotes before the ballot is wiped for the revote, so a client
// can still render what tied instead of just a "tie - revote" label.
func TestGame_CloseVoting_TieCapturesTiedVotesThenResetsBallot(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 2)
	assignAndOpenVoting(t, g)

	if err := g.CastVote(players[0].ID(), "SURVIVE"); err != nil {
		t.Fatalf("CastVote(0): %v", err)
	}
	if err := g.CastVote(players[1].ID(), "FALL"); err != nil {
		t.Fatalf("CastVote(1): %v", err)
	}
	tied, err := g.CloseVoting()
	if err != nil || !tied {
		t.Fatalf("CloseVoting: tied=%v err=%v", tied, err)
	}
	if g.State() != enums.Tiebreak {
		t.Fatalf("state after the first tie = %v, want TIEBREAK", g.State())
	}

	rounds := g.Rounds()
	round := rounds[len(rounds)-1]
	if len(round.TiedVotes) != 2 {
		t.Fatalf("TiedVotes = %+v, want the 2 votes that tied", round.TiedVotes)
	}
	if round.TiedVotes[players[0].ID()] != "SURVIVE" || round.TiedVotes[players[1].ID()] != "FALL" {
		t.Fatalf("TiedVotes = %+v, want each player's original tied vote preserved", round.TiedVotes)
	}
	if round.Ballot.Count() != 0 {
		t.Fatalf("Ballot.Count() after the tie = %d, want 0 (revote starts empty)", round.Ballot.Count())
	}
}
