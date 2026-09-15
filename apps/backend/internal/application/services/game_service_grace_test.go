package services_test

import (
	"context"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// twoPlayerLobby builds a solo-created Gauntlet lobby with a second human
// joined, both still in LOBBY.
func twoPlayerLobby(t *testing.T) (*services.GameService, *gameTestDeps, *game.Game, game.ParticipantID) {
	t.Helper()
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	g, code, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.JoinByCode(context.Background(), code, joinerID)
	if err != nil {
		t.Fatalf("JoinByCode: %v", err)
	}
	var joinerPID game.ParticipantID
	for _, p := range g.Participants() {
		if p.ID() != g.HostID() {
			joinerPID = p.ID()
		}
	}
	if joinerPID.IsNil() {
		t.Fatalf("could not find joiner's ParticipantID")
	}
	return svc, deps, g, joinerPID
}

func TestGrace_LobbyDisconnectFreesSeatAfterWindow(t *testing.T) {
	svc, deps, g, joinerPID := twoPlayerLobby(t)
	ctx := context.Background()

	if _, err := svc.Disconnect(ctx, g.ID(), joinerPID); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}

	// Still within the grace window: the seat must still be there.
	deps.clock.Advance(services.LobbyDisconnectGrace - 1)
	g, err := svc.GetGame(ctx, g.ID())
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	if _, ok := g.Participant(joinerPID); !ok {
		t.Fatalf("seat freed too early, before the grace window elapsed")
	}

	// Past the window: a real Leave must have run.
	deps.clock.Advance(2 * services.LobbyDisconnectGrace)
	g, err = svc.GetGame(ctx, g.ID())
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	if _, ok := g.Participant(joinerPID); ok {
		t.Fatalf("expected the disconnected LOBBY seat to be freed (Leave) after the grace window")
	}
}

func TestGrace_ReconnectBeforeExpiryCancelsTimer(t *testing.T) {
	svc, deps, g, joinerPID := twoPlayerLobby(t)
	ctx := context.Background()

	if _, err := svc.Disconnect(ctx, g.ID(), joinerPID); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if _, err := svc.Reconnect(ctx, g.ID(), joinerPID); err != nil {
		t.Fatalf("Reconnect: %v", err)
	}

	// Advance well past what the grace window would have been - a cancelled
	// timer must never fire the Leave.
	deps.clock.Advance(10 * services.LobbyDisconnectGrace)
	g, err := svc.GetGame(ctx, g.ID())
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	p, ok := g.Participant(joinerPID)
	if !ok {
		t.Fatalf("reconnected seat must not have been freed")
	}
	if !p.Connected() {
		t.Fatalf("expected the reconnected seat to be connected")
	}
}

func TestGrace_LobbyDisconnectPromotedToMatchGraceOnStart(t *testing.T) {
	svc, deps, g, joinerPID := twoPlayerLobby(t)
	ctx := context.Background()

	if _, err := svc.Disconnect(ctx, g.ID(), joinerPID); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	// Start the match before the LOBBY grace window would have elapsed.
	if _, err := svc.StartGame(ctx, g.ID(), hostOf(t, g)); err != nil {
		t.Fatalf("StartGame: %v", err)
	}

	// Past the LOBBY grace window, but well within the longer match window -
	// the seat must have been promoted to the match grace, not freed.
	deps.clock.Advance(2 * services.LobbyDisconnectGrace)
	g, err := svc.GetGame(ctx, g.ID())
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	if _, ok := g.Participant(joinerPID); !ok {
		t.Fatalf("a LOBBY disconnect that got promoted to an in-match one must not expire under the LOBBY window")
	}

	// Past the full match window: now it should be abandoned in place, not
	// removed - same ParticipantID, still seated.
	deps.clock.Advance(services.MatchDisconnectGrace)
	g, err = svc.GetGame(ctx, g.ID())
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	p, ok := g.Participant(joinerPID)
	if !ok {
		t.Fatalf("expected the seat to still be present (Abandon keeps the seat), got removed")
	}
	if !p.Abandoned() {
		t.Fatalf("expected the seat to be abandoned after the full match grace window")
	}
}

func TestGrace_RestartRearmsFromPersistedDisconnectedAt(t *testing.T) {
	svc, deps, g, joinerPID := twoPlayerLobby(t)
	ctx := context.Background()

	if _, err := svc.Disconnect(ctx, g.ID(), joinerPID); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	// Advance partway into the grace window before the "restart".
	deps.clock.Advance(services.LobbyDisconnectGrace / 2)
	g, err := svc.GetGame(ctx, g.ID())
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}

	// Simulate a process restart: a brand new GameService/clock over a store
	// seeded from a real Snapshot/Restore round trip, exactly like
	// restartService does for phase timers.
	restarted, restartedDeps := restartService(t, deps, g, deps.clock.Now())

	// The new process has no timer at all yet - GetGame's read path is what
	// re-arms it from the persisted disconnectedAt (see
	// rearmGraceTimersLocked). Only after that does advancing the clock past
	// the remaining half of the original window actually fire it.
	if _, err := restarted.GetGame(ctx, g.ID()); err != nil {
		t.Fatalf("GetGame after restart (re-arm): %v", err)
	}
	restartedDeps.clock.Advance(services.LobbyDisconnectGrace)
	final, err := restarted.GetGame(ctx, g.ID())
	if err != nil {
		t.Fatalf("GetGame after restart: %v", err)
	}
	if _, ok := final.Participant(joinerPID); ok {
		t.Fatalf("expected the grace timer to have been re-armed and fired after restart")
	}
}

func TestGrace_TerminalGameArmsNoTimer(t *testing.T) {
	svc, deps, _ := finishedGauntlet(t)
	ctx := context.Background()
	id := deps.finishedID

	g, err := svc.GetGame(ctx, id)
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	host := hostOf(t, g)

	if _, err := svc.Disconnect(ctx, id, host); err != nil {
		t.Fatalf("Disconnect on a finished game: %v", err)
	}

	// Even advancing well past the match grace window must not touch a
	// terminal game - no Abandon/Leave should ever run against it.
	deps.clock.Advance(10 * services.MatchDisconnectGrace)
	g, err = svc.GetGame(ctx, id)
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	if g.State() != enums.Finished {
		t.Fatalf("expected state to remain FINISHED, got %v", g.State())
	}
	p, ok := g.Participant(host)
	if !ok {
		t.Fatalf("expected the host's seat to remain present in a finished game")
	}
	if p.Abandoned() {
		t.Fatalf("a terminal game must never grow an Abandon timer")
	}
}
