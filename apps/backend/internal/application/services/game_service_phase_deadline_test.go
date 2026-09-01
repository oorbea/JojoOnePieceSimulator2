package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// These tests cover the durable phase deadline: a backend restart mid-vote
// (or mid-reveal/summary/result) used to lose s.timers and the four
// deadline maps outright, wedging the game in that phase forever with no
// client-facing deadline and no timer that would ever fire to close it.
// The deadline now lives on the aggregate itself and a fresh GameService
// re-arms from it on first load.

// restartService simulates a process restart around a live game: it takes
// the aggregate through a real Snapshot/Restore round trip (so nothing
// process-local survives except what the snapshot carries), seeds a brand
// new store with the restored aggregate, and returns a brand new
// GameService whose clock starts at `at` - i.e. partway into the phase the
// old instance had armed.
func restartService(t *testing.T, old *gameTestDeps, g *game.Game, at time.Time) (*services.GameService, *gameTestDeps) {
	t.Helper()

	restored, err := game.Restore(g.Snapshot())
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}

	store := newFakeGameStore()
	if err := store.Create(context.Background(), "RESTRT", restored); err != nil {
		t.Fatalf("seeding restarted store: %v", err)
	}

	deps := &gameTestDeps{
		store:    store,
		stages:   old.stages,
		powers:   old.powers,
		weights:  old.weights,
		tiebreak: old.tiebreak,
		history:  old.history,
		rng:      old.rng,
		hub:      services.NewGameEventHub(),
		clock:    newFakeClockAt(at),
		users:    old.users,
	}
	return newGameServiceFromDeps(deps, deps.history), deps
}

// startedVotingGame drives a solo Gauntlet lobby all the way to an open
// voting window.
func startedVotingGame(t *testing.T) (*services.GameService, *gameTestDeps, *game.Game) {
	t.Helper()
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if _, err := svc.StartGame(context.Background(), g.ID(), hostOf(t, g)); err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	advanceReveal(deps, gauntletInput().PowerMangas)
	advanceSummary(deps)

	g, err = svc.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	if g.State() != enums.Voting {
		t.Fatalf("state = %v, want VOTING", g.State())
	}
	return svc, deps, g
}

// TestPhaseDeadline_PersistedForEveryTimedPhase walks a game through
// ASSIGNING, SUMMARY, VOTING and RESOLVING, asserting the aggregate carries
// the same deadline the service is serving out of its own in-process map at
// each step - so whichever phase a restart lands in, the snapshot has it.
func TestPhaseDeadline_PersistedForEveryTimedPhase(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	ctx := context.Background()

	g, _, err := svc.CreateGame(ctx, hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if _, ok := g.PhaseEndsAt(); ok {
		t.Fatal("a LOBBY game must carry no phase deadline")
	}

	g, err = svc.StartGame(ctx, g.ID(), hostOf(t, g))
	if err != nil {
		t.Fatalf("StartGame: %v", err)
	}

	check := func(state enums.GameState, inProcess func(game.GameID) (time.Time, bool)) {
		t.Helper()
		if g.State() != state {
			t.Fatalf("state = %v, want %v", g.State(), state)
		}
		persisted, ok := g.PhaseEndsAt()
		if !ok {
			t.Fatalf("%v: aggregate carries no phase deadline", state)
		}
		live, ok := inProcess(g.ID())
		if !ok {
			t.Fatalf("%v: service reports no in-process deadline", state)
		}
		if !persisted.Equal(live) {
			t.Fatalf("%v: persisted deadline %v != in-process deadline %v", state, persisted, live)
		}
	}

	check(enums.Assigning, svc.RevealEndsAt)

	advanceReveal(deps, gauntletInput().PowerMangas)
	if g, err = svc.GetGame(ctx, g.ID()); err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	check(enums.Summary, svc.SummaryEndsAt)

	advanceSummary(deps)
	if g, err = svc.GetGame(ctx, g.ID()); err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	check(enums.Voting, svc.VotingEndsAt)

	// A solo lobby's single vote closes the window immediately and parks the
	// game in RESOLVING for the result-display window.
	if g, err = svc.CastVote(ctx, g.ID(), hostOf(t, g), "SURVIVE"); err != nil {
		t.Fatalf("CastVote: %v", err)
	}
	check(enums.Resolving, svc.ResultEndsAt)
}

// TestPhaseDeadline_RearmUsesRemainingTime is the core restart regression:
// the new instance must resume the countdown where the old one left off,
// not restart a full fresh window.
func TestPhaseDeadline_RearmUsesRemainingTime(t *testing.T) {
	_, deps, g := startedVotingGame(t)

	deadline, ok := g.PhaseEndsAt()
	if !ok {
		t.Fatal("no persisted voting deadline to restart from")
	}

	// Restart 20s into a 30s window: 10s must remain.
	restartAt := deadline.Add(-10 * time.Second)
	svc2, deps2 := restartService(t, deps, g, restartAt)

	restored, err := svc2.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame after restart: %v", err)
	}
	if restored.State() != enums.Voting {
		t.Fatalf("state after restart = %v, want VOTING", restored.State())
	}

	got, ok := svc2.VotingEndsAt(g.ID())
	if !ok {
		t.Fatal("the restarted instance did not re-arm a voting deadline")
	}
	if !got.Equal(deadline) {
		t.Fatalf("re-armed deadline = %v, want the original %v (a full fresh window would be %v)",
			got, deadline, restartAt.Add(30*time.Second))
	}

	// Nine seconds is not enough - the window still has one to go.
	deps2.clock.Advance(9 * time.Second)
	after, err := svc2.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	if after.State() != enums.Voting {
		t.Fatalf("state after 9s of the remaining 10s = %v, want still VOTING", after.State())
	}

	// The tenth closes it, proving the re-armed timer is real and fires.
	deps2.clock.Advance(1 * time.Second)
	after, err = svc2.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	if after.State() == enums.Voting {
		t.Fatal("the re-armed voting timer never fired - the game is wedged in VOTING")
	}
}

// TestPhaseDeadline_ExpiredWindowClosesOnFirstLoad covers the restart that
// happens to come back up after the window it lost had already expired: the
// re-armed timer is due immediately and must close the window, not sit
// there.
func TestPhaseDeadline_ExpiredWindowClosesOnFirstLoad(t *testing.T) {
	_, deps, g := startedVotingGame(t)

	deadline, ok := g.PhaseEndsAt()
	if !ok {
		t.Fatal("no persisted voting deadline to restart from")
	}

	svc2, deps2 := restartService(t, deps, g, deadline.Add(5*time.Minute))
	if _, err := svc2.GetGame(context.Background(), g.ID()); err != nil {
		t.Fatalf("GetGame after restart: %v", err)
	}

	// An already-due timer is armed at zero; the fake clock fires due timers
	// on Advance, so a zero-length advance is enough to run it.
	deps2.clock.Advance(0)

	after, err := svc2.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	if after.State() == enums.Voting {
		t.Fatal("an already-expired persisted voting deadline left the game wedged in VOTING")
	}
}

// TestPhaseDeadline_LegacySnapshotArmsFullWindow covers a snapshot written
// before the deadline was persisted at all: rather than wedge, the game
// gets a fresh full window.
func TestPhaseDeadline_LegacySnapshotArmsFullWindow(t *testing.T) {
	_, deps, g := startedVotingGame(t)

	// A legacy snapshot simply has no deadline field.
	snap := g.Snapshot()
	snap.PhaseEndsAt = nil
	legacy, err := game.Restore(snap)
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if _, ok := legacy.PhaseEndsAt(); ok {
		t.Fatal("legacy snapshot should restore with no deadline")
	}

	startedAt := time.Unix(5000, 0)
	svc2, deps2 := restartService(t, deps, legacy, startedAt)

	if _, err := svc2.GetGame(context.Background(), g.ID()); err != nil {
		t.Fatalf("GetGame after restart: %v", err)
	}
	got, ok := svc2.VotingEndsAt(g.ID())
	if !ok {
		t.Fatal("a legacy snapshot left the game with no voting deadline at all (wedged)")
	}
	if want := startedAt.Add(30 * time.Second); !got.Equal(want) {
		t.Fatalf("legacy fallback deadline = %v, want a full fresh window ending %v", got, want)
	}

	deps2.clock.Advance(30 * time.Second)
	after, err := svc2.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	if after.State() == enums.Voting {
		t.Fatal("the legacy-fallback voting timer never fired")
	}
}

// TestPhaseDeadline_RearmIsNoOpWhenAlreadyArmed pins that the instance
// already driving a phase does not restart its own timer on every load -
// otherwise a client polling the state would push the deadline out forever.
func TestPhaseDeadline_RearmIsNoOpWhenAlreadyArmed(t *testing.T) {
	svc, deps, g := startedVotingGame(t)

	first, ok := svc.VotingEndsAt(g.ID())
	if !ok {
		t.Fatal("no voting deadline after opening the window")
	}

	deps.clock.Advance(5 * time.Second)
	if _, err := svc.GetGame(context.Background(), g.ID()); err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	if _, err := svc.Reconnect(context.Background(), g.ID(), hostOf(t, g)); err != nil {
		t.Fatalf("Reconnect: %v", err)
	}

	second, ok := svc.VotingEndsAt(g.ID())
	if !ok {
		t.Fatal("voting deadline vanished after a plain read")
	}
	if !second.Equal(first) {
		t.Fatalf("deadline moved from %v to %v - a re-arm restarted a timer this instance already held", first, second)
	}
}
