package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// finishedGauntlet drives a solo Gauntlet lobby to FINISHED by voting FALL,
// which ends the run on the first round.
func finishedGauntlet(t *testing.T) (*services.GameService, *gameTestDeps, enums.GameState) {
	t.Helper()
	svc, deps, g := startedVotingGame(t)
	ctx := context.Background()

	if _, err := svc.CastVote(ctx, g.ID(), hostOf(t, g), "FALL"); err != nil {
		t.Fatalf("CastVote: %v", err)
	}
	advanceResult(deps)

	final, err := svc.GetGame(ctx, g.ID())
	if err != nil {
		t.Fatalf("GetGame after finish: %v", err)
	}
	if final.State() != enums.Finished {
		t.Fatalf("state = %v, want FINISHED", final.State())
	}
	deps.finishedID = g.ID()
	return svc, deps, final.State()
}

// TestFinished_GameStaysReadable is the regression test against the old
// behavior: finalizeLocked used to Delete the game from the store the
// instant it reached FINISHED, which made a result screen impossible - the
// terminal frame and the client's follow-up read raced, and the read lost.
func TestFinished_GameStaysReadable(t *testing.T) {
	svc, deps, _ := finishedGauntlet(t)
	id := deps.finishedID

	// Readable through the service...
	if _, err := svc.GetGame(context.Background(), id); err != nil {
		t.Fatalf("GetGame on a finished game: %v, want it to still be readable", err)
	}
	// ...and actually present in the store, not just cached anywhere.
	if _, err := deps.store.Get(context.Background(), id); err != nil {
		t.Fatalf("store.Get on a finished game: %v, want it to still be there", err)
	}
}

// TestFinished_UsesShortTTL pins that the terminal save uses the dedicated
// short TTL rather than the store's default lobby TTL.
func TestFinished_UsesShortTTL(t *testing.T) {
	_, deps, _ := finishedGauntlet(t)

	got := deps.store.ttlFor(deps.finishedID)
	if got != services.FinishedGameTTL {
		t.Fatalf("finished game saved with TTL %v, want services.FinishedGameTTL (%v)", got, services.FinishedGameTTL)
	}
	if services.FinishedGameTTL >= 2*time.Hour {
		t.Fatalf("FinishedGameTTL = %v, want something much shorter than the 2h lobby TTL", services.FinishedGameTTL)
	}
}

// TestFinished_HistoryRecordedOnce covers the post-finish Reconnect/
// Disconnect path: both are legal against a terminal game and both re-enter
// withGame's Finished branch, which must not re-issue the history inserts on
// every client reload.
func TestFinished_HistoryRecordedOnce(t *testing.T) {
	svc, deps, _ := finishedGauntlet(t)
	id := deps.finishedID
	ctx := context.Background()

	g, err := svc.GetGame(ctx, id)
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	host := hostOf(t, g)

	if before := len(deps.history.all()); before != 1 {
		t.Fatalf("history records right after finishing = %d, want 1", before)
	}

	// Both of these re-enter the Finished branch of withGame.
	if _, err := svc.Reconnect(ctx, id, host); err != nil {
		t.Fatalf("Reconnect after finish: %v", err)
	}
	if _, err := svc.Disconnect(ctx, id, host); err != nil {
		t.Fatalf("Disconnect after finish: %v", err)
	}

	if got := len(deps.history.all()); got != 1 {
		t.Fatalf("history records after a post-finish reconnect+disconnect = %d, want still 1", got)
	}
	// And the game is still readable throughout - a reconnect must not
	// evict it.
	if _, err := svc.GetGame(ctx, id); err != nil {
		t.Fatalf("GetGame after post-finish presence churn: %v", err)
	}
}
