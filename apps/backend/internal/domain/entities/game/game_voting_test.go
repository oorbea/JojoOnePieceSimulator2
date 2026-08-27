package game_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func TestGame_VotingComplete_FalseUntilAllHumansVote(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 2)
	assignAndOpenVoting(t, g)

	if g.VotingComplete() {
		t.Fatalf("expected VotingComplete to be false before anyone votes")
	}
	if err := g.CastVote(players[0].ID(), "SURVIVE"); err != nil {
		t.Fatalf("CastVote: %v", err)
	}
	if g.VotingComplete() {
		t.Fatalf("expected VotingComplete to be false with one human left unvoted")
	}
	if err := g.CastVote(players[1].ID(), "SURVIVE"); err != nil {
		t.Fatalf("CastVote: %v", err)
	}
	if !g.VotingComplete() {
		t.Fatalf("expected VotingComplete to be true once every human has voted")
	}
}

func TestGame_VotingComplete_TrueWhenLastHoldoutDisconnects(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 2)
	assignAndOpenVoting(t, g)

	if err := g.CastVote(players[0].ID(), "SURVIVE"); err != nil {
		t.Fatalf("CastVote: %v", err)
	}
	if g.VotingComplete() {
		t.Fatalf("expected VotingComplete to still be false")
	}
	if err := g.Disconnect(players[1].ID(), &fakeRandom{}); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if !g.VotingComplete() {
		t.Fatalf("expected VotingComplete to be true once the only holdout disconnects")
	}
}

func TestGame_VotingComplete_NotBlockedByBots(t *testing.T) {
	cfg, err := game.NewConfig(enums.Versus, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo}, enums.Random, 2, true, enums.Private, 30, game.PoolFilter{})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	teamA := mustTeam(t, 100, "A")
	teamB := mustTeam(t, 101, "B")
	host := mustHumanParticipant(t, 1, 1, 100)
	g, err := game.NewGame(game.GameID{1}, cfg, host, []*game.Team{teamA, teamB}, someStages(t))
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	bot := mustBotParticipant(t, 2, 100)
	if err := g.AddBot(bot); err != nil {
		t.Fatalf("AddBot: %v", err)
	}
	humanB1 := mustHumanParticipant(t, 3, 3, 101)
	humanB2 := mustHumanParticipant(t, 4, 4, 101)
	if err := g.Join(humanB1); err != nil {
		t.Fatalf("Join humanB1: %v", err)
	}
	if err := g.Join(humanB2); err != nil {
		t.Fatalf("Join humanB2: %v", err)
	}
	if err := g.Start(g.HostID()); err != nil {
		t.Fatalf("Start: %v", err)
	}

	builder := game.NewLoadoutBuilder(cfg.PowerMangas(), game.DefaultAssignmentWeights(), &fakeRandom{})
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

	if g.VotingComplete() {
		t.Fatalf("expected VotingComplete false before any human votes")
	}
	teamAOption := game.OptionID(teamA.ID().String())
	if err := g.CastVote(host.ID(), teamAOption); err != nil {
		t.Fatalf("CastVote host: %v", err)
	}
	if g.VotingComplete() {
		t.Fatalf("expected VotingComplete false with 2 humans left in team B")
	}
	if err := g.CastVote(humanB1.ID(), teamAOption); err != nil {
		t.Fatalf("CastVote humanB1: %v", err)
	}
	if err := g.CastVote(humanB2.ID(), teamAOption); err != nil {
		t.Fatalf("CastVote humanB2: %v", err)
	}
	if !g.VotingComplete() {
		t.Fatalf("expected VotingComplete true once every human (not the bot) has voted")
	}
}
