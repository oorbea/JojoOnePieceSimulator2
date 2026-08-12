package game_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func TestNewGame_GauntletRequiresStages(t *testing.T) {
	cfg, err := game.NewConfig(enums.Gauntlet, []enums.Manga{enums.Jojo}, enums.Random, 5, false, enums.Private, 30, game.PoolFilter{})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	host := mustHumanParticipant(t, 1, 1, 100)
	team := mustTeam(t, 100, "Squad")
	if _, err := game.NewGame(game.GameID{1}, cfg, host, []*game.Team{team}, nil); err != game.ErrNoStagesAvailable {
		t.Fatalf("expected ErrNoStagesAvailable, got %v", err)
	}
}

func TestGame_StartRejectedOutsideLobby(t *testing.T) {
	g, _ := newGauntletGame(t, oneStage(t), 1)
	assignAndOpenVoting(t, g)

	if err := g.Start(g.HostID()); err != game.ErrInvalidStateTransition {
		t.Fatalf("expected ErrInvalidStateTransition, got %v", err)
	}
}

func TestGame_OnlyHostCanStart(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 2)
	if err := g.Start(players[1].ID()); err != game.ErrNotHost {
		t.Fatalf("expected ErrNotHost, got %v", err)
	}
}

func TestGame_OnlyHostCanAbort(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 2)
	if err := g.Abort(players[1].ID()); err != game.ErrNotHost {
		t.Fatalf("expected ErrNotHost, got %v", err)
	}
	if err := g.Abort(g.HostID()); err != nil {
		t.Fatalf("Abort: %v", err)
	}
	if g.State() != enums.Aborted {
		t.Fatalf("expected ABORTED, got %v", g.State())
	}
}

func TestGame_VersusRequiresEqualTeamsToStart(t *testing.T) {
	cfg, err := game.NewConfig(enums.Versus, []enums.Manga{enums.Jojo}, enums.Random, 2, false, enums.Private, 30, game.PoolFilter{})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	host := mustHumanParticipant(t, 1, 1, 100)
	teamA := mustTeam(t, 100, "A")
	teamB := mustTeam(t, 101, "B")
	g, err := game.NewGame(game.GameID{1}, cfg, host, []*game.Team{teamA, teamB}, someStages(t))
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	// Only team A has a player (the host) - team B is empty. An empty team
	// is ErrNotEnoughPlayers, not a size mismatch (see the unequal-but-
	// non-empty case below).
	if err := g.Start(g.HostID()); err != game.ErrNotEnoughPlayers {
		t.Fatalf("expected ErrNotEnoughPlayers, got %v", err)
	}

	other := mustHumanParticipant(t, 2, 2, 101)
	if err := g.Join(other); err != nil {
		t.Fatalf("Join: %v", err)
	}
	other2 := mustHumanParticipant(t, 3, 3, 100)
	if err := g.Join(other2); err != nil {
		t.Fatalf("Join: %v", err)
	}
	// Now team A has 2 players, team B has 1 - unequal but both non-empty.
	if err := g.Start(g.HostID()); err != game.ErrTeamSizeMismatch {
		t.Fatalf("expected ErrTeamSizeMismatch, got %v", err)
	}
}

func TestGame_GauntletRejectsBots(t *testing.T) {
	g, _ := newGauntletGame(t, oneStage(t), 1)
	bot := mustBotParticipant(t, 50, 100)
	if err := g.AddBot(bot); err != game.ErrBotsNotAllowed {
		t.Fatalf("expected ErrBotsNotAllowed, got %v", err)
	}
}

func TestGame_HostReassignedOnDisconnect(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 2)
	if err := g.Disconnect(g.HostID(), &fakeRandom{seq: []int{0}}); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if g.HostID() != players[1].ID() {
		t.Fatalf("expected host reassigned to the remaining player, got %v", g.HostID())
	}
}

func TestGame_AbortsWhenNoHumansRemain(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 1)
	if err := g.Disconnect(players[0].ID(), &fakeRandom{}); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if g.State() != enums.Aborted {
		t.Fatalf("expected ABORTED when no connected humans remain, got %v", g.State())
	}
}

func TestGame_VersusTeamEmptiedInLobbyDoesNotAbort(t *testing.T) {
	// An empty Versus team is normal and recoverable in LOBBY (more
	// players can still join, or SwitchTeam can empty a team on the way
	// to an even split) - only a mid-match empty team aborts. See
	// TestGame_VersusAbortsWhenTeamEmptiesMidMatch below.
	cfg, err := game.NewConfig(enums.Versus, []enums.Manga{enums.Jojo}, enums.Random, 1, false, enums.Private, 30, game.PoolFilter{})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	host := mustHumanParticipant(t, 1, 1, 100)
	other := mustHumanParticipant(t, 2, 2, 101)
	teamA := mustTeam(t, 100, "A")
	teamB := mustTeam(t, 101, "B")
	g, err := game.NewGame(game.GameID{1}, cfg, host, []*game.Team{teamA, teamB}, someStages(t))
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	if err := g.Join(other); err != nil {
		t.Fatalf("Join: %v", err)
	}
	if err := g.Leave(other.ID(), &fakeRandom{}); err != nil {
		t.Fatalf("Leave: %v", err)
	}
	if g.State() != enums.Lobby {
		t.Fatalf("expected LOBBY (not aborted) when a versus team empties before starting, got %v", g.State())
	}
}

func TestGame_VersusAbortsWhenTeamEmptiesMidMatch(t *testing.T) {
	cfg, err := game.NewConfig(enums.Versus, []enums.Manga{enums.Jojo}, enums.Random, 1, false, enums.Private, 30, game.PoolFilter{})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	host := mustHumanParticipant(t, 1, 1, 100)
	other := mustHumanParticipant(t, 2, 2, 101)
	teamA := mustTeam(t, 100, "A")
	teamB := mustTeam(t, 101, "B")
	g, err := game.NewGame(game.GameID{1}, cfg, host, []*game.Team{teamA, teamB}, someStages(t))
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	if err := g.Join(other); err != nil {
		t.Fatalf("Join: %v", err)
	}
	if err := g.Start(g.HostID()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	builder := game.NewLoadoutBuilder(cfg.Mangas(), game.DefaultAssignmentWeights(), &fakeRandom{})
	pools := map[game.TeamID]*game.AvailablePowers{
		teamA.ID(): game.NewAvailablePowers(nil, nil),
		teamB.ID(): game.NewAvailablePowers(nil, nil),
	}
	if err := g.AssignLoadouts(builder, pools); err != nil {
		t.Fatalf("AssignLoadouts: %v", err)
	}
	if err := g.OpenVoting(&fakeRandom{}); err != nil {
		t.Fatalf("OpenVoting: %v", err)
	}
	if err := g.Leave(other.ID(), &fakeRandom{}); err != nil {
		t.Fatalf("Leave: %v", err)
	}
	if g.State() != enums.Aborted {
		t.Fatalf("expected ABORTED when a versus team empties mid-match, got %v", g.State())
	}
}

func TestGame_EventsEmittedAndDrained(t *testing.T) {
	// NewGame seats the host directly (not via Join), so it emits nothing -
	// a second participant joining is what should produce a PlayerJoined
	// event.
	g, players := newGauntletGame(t, oneStage(t), 2)
	events := g.PullEvents()
	if len(events) == 0 {
		t.Fatalf("expected at least a PlayerJoined event for the second player")
	}
	found := false
	for _, e := range events {
		if joined, ok := e.(game.PlayerJoined); ok && joined.ParticipantID == players[1].ID() {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a PlayerJoined event for %v, got %+v", players[1].ID(), events)
	}
	if len(g.PullEvents()) != 0 {
		t.Fatalf("expected a second PullEvents call to be empty")
	}
}
