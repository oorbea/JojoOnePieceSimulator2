package services_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// TestVotingEndsAt_TracksThenClearsOnEarlyClose mirrors RevealEndsAt's own
// happy-path test: the deadline is recorded the instant voting opens, and
// cleared the instant the window closes early because every connected
// human voted.
func TestVotingEndsAt_TracksThenClearsOnEarlyClose(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.StartGame(context.Background(), g.ID(), g.HostID())
	if err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	advanceReveal(deps, gauntletInput().Mangas)
	g, err = svc.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame after reveal: %v", err)
	}
	if g.State() != enums.Voting {
		t.Fatalf("state after reveal = %v, want VOTING", g.State())
	}

	endsAt, ok := svc.VotingEndsAt(g.ID())
	if !ok {
		t.Fatal("VotingEndsAt: want ok once voting is open")
	}
	want := deps.clock.Now().Add(30 * time.Second)
	if !endsAt.Equal(want) {
		t.Fatalf("VotingEndsAt = %v, want %v", endsAt, want)
	}

	// FALL ends a Gauntlet run outright (unlike SURVIVE, which would just
	// advance to the fixture's second stage and reopen voting immediately,
	// since Gauntlet doesn't reassign Loadouts past round 1 - see
	// TestGauntlet_ClearAllStages_Victory) - so this is the case that
	// actually leaves no further voting window to track.
	if _, err := svc.CastVote(context.Background(), g.ID(), g.HostID(), "FALL"); err != nil {
		t.Fatalf("CastVote: %v", err)
	}
	if _, ok := svc.VotingEndsAt(g.ID()); ok {
		t.Fatal("VotingEndsAt after the sole human voted FALL: want not-ok, the run finished")
	}
}

// TestVotingEndsAt_AbsentDuringReveal proves the reveal and voting
// deadlines never coexist for the same Game.
func TestVotingEndsAt_AbsentDuringReveal(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if _, err := svc.StartGame(context.Background(), g.ID(), g.HostID()); err != nil {
		t.Fatalf("StartGame: %v", err)
	}

	if _, ok := svc.RevealEndsAt(g.ID()); !ok {
		t.Fatal("RevealEndsAt: want ok right after StartGame")
	}
	if _, ok := svc.VotingEndsAt(g.ID()); ok {
		t.Fatal("VotingEndsAt: want not-ok during the reveal - the two deadlines must never coexist")
	}
}

// TestVotingEndsAt_ClearedOnTimerExpiry covers the window closing on its
// own timer (nobody votes in time). Zero votes is a tie (see
// game.Ballot.Tally), so the first expiry only opens a fresh revote window
// (still ok, freshly rescheduled) - the deadline is only truly gone once
// the second expiry's tie is handed to the tiebreaker and that resolves the
// run outright (deps.tiebreak.winner is set to FALL for exactly that
// reason - see TestVotingEndsAt_TracksThenClearsOnEarlyClose's doc on why
// FALL, not SURVIVE, is what actually stops the game from reopening voting
// again on its own).
func TestVotingEndsAt_ClearedOnTimerExpiry(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if _, err := svc.StartGame(context.Background(), g.ID(), g.HostID()); err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	advanceReveal(deps, gauntletInput().Mangas)
	deps.tiebreak.winner = "FALL"

	if _, ok := svc.VotingEndsAt(g.ID()); !ok {
		t.Fatal("VotingEndsAt: want ok once voting is open")
	}

	// Nobody votes; the 30s window timer fires on its own -> zero votes is a
	// tie -> a fresh revote window opens with its own deadline.
	deps.clock.Advance(30 * time.Second)
	if _, ok := svc.VotingEndsAt(g.ID()); !ok {
		t.Fatal("VotingEndsAt after the first expiry: want ok (the revote window just opened)")
	}

	// Nobody votes the revote either -> second tie -> the tiebreaker
	// resolves the round as FALL, which ends the run.
	deps.clock.Advance(30 * time.Second)
	if _, ok := svc.VotingEndsAt(g.ID()); ok {
		t.Fatal("VotingEndsAt after the revote also expired: want not-ok, the run finished")
	}
}

// TestVotingEndsAt_RevoteReschedulesDeadline checks that opening the
// revote window (the first tie) records a fresh deadline rather than
// leaving the stale one from the original window.
func TestVotingEndsAt_RevoteReschedulesDeadline(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	g, code, err := svc.CreateGame(context.Background(), hostID, versusInput(1))
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.JoinByCode(context.Background(), code, joinerID)
	if err != nil {
		t.Fatalf("JoinByCode: %v", err)
	}
	g, err = svc.StartGame(context.Background(), g.ID(), g.HostID())
	if err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	advanceReveal(deps, versusInput(1).Mangas)
	g, err = svc.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame after reveal: %v", err)
	}

	teamA := game.OptionID(g.Teams()[0].ID().String())
	teamB := game.OptionID(g.Teams()[1].ID().String())
	var joinerParticipant game.ParticipantID
	for _, p := range g.Participants() {
		if p.ID() != g.HostID() {
			joinerParticipant = p.ID()
		}
	}

	// Advance the clock a few seconds before the tie, so a stale deadline
	// carried over from the first window would be visibly wrong against the
	// fresh one the revote must record.
	deps.clock.Advance(5 * time.Second)

	if _, err := svc.CastVote(context.Background(), g.ID(), g.HostID(), teamA); err != nil {
		t.Fatalf("host vote: %v", err)
	}
	g, err = svc.CastVote(context.Background(), g.ID(), joinerParticipant, teamB)
	if err != nil {
		t.Fatalf("joiner vote: %v", err)
	}
	if g.State() != enums.Tiebreak {
		t.Fatalf("state after tie = %v, want TIEBREAK", g.State())
	}

	endsAt, ok := svc.VotingEndsAt(g.ID())
	if !ok {
		t.Fatal("VotingEndsAt: want ok once the revote window opens")
	}
	want := deps.clock.Now().Add(30 * time.Second)
	if !endsAt.Equal(want) {
		t.Fatalf("VotingEndsAt (revote) = %v, want %v (fresh, not carried over from the first window)", endsAt, want)
	}
}

// TestAbortGame_DuringVoting_ClearsVotingEndsAt mirrors
// TestAbortGame_DuringReveal_NeverOpensVoting: aborting mid-vote clears the
// deadline and leaves no timer behind to resurrect the (deleted) Game.
func TestAbortGame_DuringVoting_ClearsVotingEndsAt(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.StartGame(context.Background(), g.ID(), g.HostID())
	if err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	advanceReveal(deps, gauntletInput().Mangas)
	g, err = svc.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame after reveal: %v", err)
	}
	if g.State() != enums.Voting {
		t.Fatalf("state after reveal = %v, want VOTING", g.State())
	}

	if _, err := svc.AbortGame(context.Background(), g.ID(), g.HostID()); err != nil {
		t.Fatalf("AbortGame: %v", err)
	}
	if _, ok := svc.VotingEndsAt(g.ID()); ok {
		t.Fatal("VotingEndsAt after abort: want not-ok, the voting timer should be cancelled")
	}

	// Advance well past the voting window - if the timer had survived the
	// abort, this is where it would fire against a Game already deleted.
	deps.clock.Advance(30 * time.Second)

	if _, err := svc.GetGame(context.Background(), g.ID()); !errors.Is(err, ports.ErrGameNotFound) {
		t.Fatalf("GetGame after abort+advance: err = %v, want ErrGameNotFound", err)
	}
}

// TestCastVote_PublishesHumanVoteProgress is the end-to-end proof that the
// domain event's HumanVotesCast/HumanVoters survive GameService.publish.
func TestCastVote_PublishesHumanVoteProgress(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	g, code, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if _, err := svc.JoinByCode(context.Background(), code, joinerID); err != nil {
		t.Fatalf("JoinByCode: %v", err)
	}
	g, err = svc.StartGame(context.Background(), g.ID(), g.HostID())
	if err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	advanceReveal(deps, gauntletInput().Mangas)
	g, err = svc.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame after reveal: %v", err)
	}

	var joiner game.ParticipantID
	for _, p := range g.Participants() {
		if p.ID() != g.HostID() {
			joiner = p.ID()
		}
	}

	sub, unsub := deps.hub.Subscribe(g.ID())
	var mu sync.Mutex
	var voteCastEvents []game.VoteCast
	done := make(chan struct{})
	go func() {
		for e := range sub {
			if vc, ok := e.Event.(game.VoteCast); ok {
				mu.Lock()
				voteCastEvents = append(voteCastEvents, vc)
				mu.Unlock()
			}
		}
		close(done)
	}()

	if _, err := svc.CastVote(context.Background(), g.ID(), g.HostID(), "SURVIVE"); err != nil {
		t.Fatalf("host vote: %v", err)
	}
	if _, err := svc.CastVote(context.Background(), g.ID(), joiner, "SURVIVE"); err != nil {
		t.Fatalf("joiner vote: %v", err)
	}

	unsub()
	<-done

	mu.Lock()
	defer mu.Unlock()
	if len(voteCastEvents) != 2 {
		t.Fatalf("published VoteCast events = %d, want 2", len(voteCastEvents))
	}
	if voteCastEvents[0].HumanVotesCast != 1 || voteCastEvents[0].HumanVoters != 2 {
		t.Fatalf("first VoteCast = %d/%d, want 1/2", voteCastEvents[0].HumanVotesCast, voteCastEvents[0].HumanVoters)
	}
	if voteCastEvents[1].HumanVotesCast != 2 || voteCastEvents[1].HumanVoters != 2 {
		t.Fatalf("second VoteCast = %d/%d, want 2/2", voteCastEvents[1].HumanVotesCast, voteCastEvents[1].HumanVoters)
	}
}
