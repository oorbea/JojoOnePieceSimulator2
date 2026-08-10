package game_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
)

func TestBallot_SimpleMajority(t *testing.T) {
	b, err := game.NewBallot([]game.OptionID{"SURVIVE", "FALL"})
	if err != nil {
		t.Fatalf("NewBallot: %v", err)
	}
	_ = b.Cast(game.ParticipantID{1}, "SURVIVE")
	_ = b.Cast(game.ParticipantID{2}, "SURVIVE")
	_ = b.Cast(game.ParticipantID{3}, "FALL")

	winner, tied := b.Tally()
	if tied || winner != "SURVIVE" {
		t.Fatalf("expected SURVIVE to win untied, got winner=%v tied=%v", winner, tied)
	}
}

func TestBallot_Tie(t *testing.T) {
	b, _ := game.NewBallot([]game.OptionID{"SURVIVE", "FALL"})
	_ = b.Cast(game.ParticipantID{1}, "SURVIVE")
	_ = b.Cast(game.ParticipantID{2}, "FALL")

	if _, tied := b.Tally(); !tied {
		t.Fatalf("expected a tie")
	}
}

func TestBallot_NullVotesDoNotCount(t *testing.T) {
	b, _ := game.NewBallot([]game.OptionID{"SURVIVE", "FALL"})
	_ = b.Cast(game.ParticipantID{1}, "SURVIVE")
	// participant 2 never votes - a null vote, e.g. disconnected/too slow.

	winner, tied := b.Tally()
	if tied || winner != "SURVIVE" {
		t.Fatalf("expected SURVIVE to win on emitted votes only, got winner=%v tied=%v", winner, tied)
	}
	if b.Count() != 1 {
		t.Fatalf("expected 1 emitted vote, got %d", b.Count())
	}
}

func TestBallot_ZeroVotesIsTied(t *testing.T) {
	b, _ := game.NewBallot([]game.OptionID{"SURVIVE", "FALL"})
	if _, tied := b.Tally(); !tied {
		t.Fatalf("expected zero emitted votes to be treated as a tie")
	}
}

func TestBallot_RecastOverwritesPreviousVote(t *testing.T) {
	b, _ := game.NewBallot([]game.OptionID{"SURVIVE", "FALL"})
	_ = b.Cast(game.ParticipantID{1}, "SURVIVE")
	_ = b.Cast(game.ParticipantID{1}, "FALL")

	winner, tied := b.Tally()
	if tied || winner != "FALL" {
		t.Fatalf("expected the last cast vote (FALL) to count, got winner=%v tied=%v", winner, tied)
	}
}

func TestBallot_InvalidOption(t *testing.T) {
	b, _ := game.NewBallot([]game.OptionID{"SURVIVE", "FALL"})
	if err := b.Cast(game.ParticipantID{1}, "MAYBE"); err != game.ErrInvalidBallotOption {
		t.Fatalf("expected ErrInvalidBallotOption, got %v", err)
	}
}
