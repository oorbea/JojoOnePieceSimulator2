package services_test

import (
	"context"
	"sync"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

type recordingNotifier struct {
	mu    sync.Mutex
	calls []ports.GameNotification
}

func (r *recordingNotifier) NotifyGame(n ports.GameNotification) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, n)
}

func (r *recordingNotifier) kinds() []ports.GameNotificationKind {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]ports.GameNotificationKind, 0, len(r.calls))
	for _, c := range r.calls {
		out = append(out, c.Kind)
	}
	return out
}

func TestGameService_NotifiesPlayersOfStartAndVotingOpening(t *testing.T) {
	svc, deps := newTestGameService(t)
	notifier := &recordingNotifier{}
	svc.SetNotifier(notifier)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if len(notifier.kinds()) != 0 {
		t.Fatalf("notified during lobby creation: %v", notifier.kinds())
	}

	g, err = svc.StartGame(context.Background(), g.ID(), hostOf(t, g))
	if err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	if got := notifier.kinds(); len(got) != 1 || got[0] != ports.NotifyGameStarted {
		t.Fatalf("after StartGame notified %v, want exactly [GAME_STARTED]", got)
	}

	advanceReveal(deps, gauntletInput().PowerMangas)
	advanceSummary(deps)

	got := notifier.kinds()
	if len(got) != 2 || got[1] != ports.NotifyVotingOpened {
		t.Fatalf("after the reveal notified %v, want [GAME_STARTED VOTING_OPENED]", got)
	}

	for _, call := range notifier.calls {
		if call.GameID != g.ID() {
			t.Errorf("notification for game %s, want %s", call.GameID, g.ID())
		}
		if len(call.Recipients) != 1 || call.Recipients[0] != hostID {
			t.Errorf("recipients = %v, want only the host's user id", call.Recipients)
		}
	}
}

func TestGameService_WithoutANotifierStillPlays(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if _, err := svc.StartGame(context.Background(), g.ID(), hostOf(t, g)); err != nil {
		t.Fatalf("StartGame without a notifier: %v", err)
	}
}
