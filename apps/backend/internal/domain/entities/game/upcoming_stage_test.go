package game_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
)

func TestUpcomingStage_VersusPicksAtAssignAndOpenVotingKeepsIt(t *testing.T) {
	g, _, _, _, _ := newVersusLobby(t, 1)
	assignOnly(t, g)

	if err := g.PrepareUpcomingStage(&fakeRandom{seq: []int{1}}); err != nil {
		t.Fatalf("PrepareUpcomingStage: %v", err)
	}
	upcoming, ok := g.UpcomingStage()
	if !ok || upcoming.Name() != "Battle Tendency" {
		t.Fatalf("UpcomingStage = %q (ok=%v), want Battle Tendency", upcoming.Name(), ok)
	}

	// A different rng must not matter: OpenVoting uses the announced stage.
	if err := g.OpenVoting(&fakeRandom{seq: []int{2}}); err != nil {
		t.Fatalf("OpenVoting: %v", err)
	}
	rounds := g.Rounds()
	if got := rounds[len(rounds)-1].Stage.Name(); got != "Battle Tendency" {
		t.Fatalf("round stage = %q, want the announced Battle Tendency", got)
	}
	if _, ok := g.UpcomingStage(); ok {
		t.Fatal("upcoming stage must be consumed once voting opens")
	}
}

func TestUpcomingStage_GauntletNeverSetsOne(t *testing.T) {
	g, _ := newGauntletGame(t, oneStage(t), 2)
	assignOnly(t, g)
	if err := g.PrepareUpcomingStage(&fakeRandom{}); err != nil {
		t.Fatalf("PrepareUpcomingStage: %v", err)
	}
	if _, ok := g.UpcomingStage(); ok {
		t.Fatal("Gauntlet keeps choosing its stage when voting opens")
	}
}

func TestUpcomingStage_SurvivesSnapshotRoundTrip(t *testing.T) {
	g, _, _, _, _ := newVersusLobby(t, 1)
	assignOnly(t, g)
	if err := g.PrepareUpcomingStage(&fakeRandom{seq: []int{2}}); err != nil {
		t.Fatalf("PrepareUpcomingStage: %v", err)
	}
	restored, err := game.Restore(g.Snapshot())
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	st, ok := restored.UpcomingStage()
	if !ok || st.Name() != "Stardust Crusaders" {
		t.Fatalf("restored UpcomingStage = %q (ok=%v), want Stardust Crusaders", st.Name(), ok)
	}
}
