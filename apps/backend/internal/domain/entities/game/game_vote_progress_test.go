package game_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// lastVoteCast returns the most recently emitted game.VoteCast event, or
// fails the test if none was emitted.
func lastVoteCast(t *testing.T, events []game.DomainEvent) game.VoteCast {
	t.Helper()
	for i := len(events) - 1; i >= 0; i-- {
		if vc, ok := events[i].(game.VoteCast); ok {
			return vc
		}
	}
	t.Fatal("no VoteCast event found")
	return game.VoteCast{}
}

func TestGame_CastVote_EventCarriesHumanVoteProgress(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 2)
	assignAndOpenVoting(t, g)
	g.PullEvents() // drain setup events (GameStarted, LoadoutsAssigned, VotingOpened)

	if err := g.CastVote(players[0].ID(), "SURVIVE"); err != nil {
		t.Fatalf("CastVote 1: %v", err)
	}
	vc := lastVoteCast(t, g.PullEvents())
	if vc.HumanVotesCast != 1 || vc.HumanVoters != 2 {
		t.Fatalf("after first vote: got %d/%d, want 1/2", vc.HumanVotesCast, vc.HumanVoters)
	}

	if err := g.CastVote(players[1].ID(), "FALL"); err != nil {
		t.Fatalf("CastVote 2: %v", err)
	}
	vc = lastVoteCast(t, g.PullEvents())
	if vc.HumanVotesCast != 2 || vc.HumanVoters != 2 {
		t.Fatalf("after second vote: got %d/%d, want 2/2", vc.HumanVotesCast, vc.HumanVoters)
	}
}

func TestGame_BotVoteCast_CountsHumansOnly(t *testing.T) {
	cfg, err := game.NewConfig(enums.Versus, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo}, enums.Random, 3, true, enums.Private, 30, game.PoolFilter{}, enums.Normal, game.DefaultSummaryDurationSeconds)
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
	g.PullEvents() // drain up to (not including) OpenVoting's bot vote

	if err := g.OpenVoting(&fakeRandom{}); err != nil {
		t.Fatalf("OpenVoting: %v", err)
	}
	events := g.PullEvents()

	found := false
	for _, e := range events {
		if vc, ok := e.(game.VoteCast); ok {
			found = true
			if vc.ParticipantID != bot.ID() {
				t.Fatalf("expected the only VoteCast right after OpenVoting to be the bot's, got %v", vc.ParticipantID)
			}
			if vc.HumanVotesCast != 0 || vc.HumanVoters != 3 {
				t.Fatalf("bot's own VoteCast = %d/%d, want 0/3 (human-only, three humans, none voted yet)", vc.HumanVotesCast, vc.HumanVoters)
			}
		}
	}
	if !found {
		t.Fatal("expected the bot to have cast a VoteCast event on OpenVoting")
	}
}

func TestGame_CastVote_ProgressExcludesDisconnectedHuman(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 3)
	if err := g.Disconnect(players[2].ID(), &fakeRandom{}); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	assignAndOpenVoting(t, g)
	g.PullEvents()

	if err := g.CastVote(players[0].ID(), "SURVIVE"); err != nil {
		t.Fatalf("CastVote: %v", err)
	}
	vc := lastVoteCast(t, g.PullEvents())
	if vc.HumanVotesCast != 1 || vc.HumanVoters != 2 {
		t.Fatalf("got %d/%d, want 1/2 (disconnected third player excluded)", vc.HumanVotesCast, vc.HumanVoters)
	}
}

func TestGame_VoteProgress_AgreesWithVotingComplete(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 3)
	if err := g.Disconnect(players[2].ID(), &fakeRandom{}); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	assignAndOpenVoting(t, g)
	g.PullEvents()

	if err := g.CastVote(players[0].ID(), "SURVIVE"); err != nil {
		t.Fatalf("CastVote 1: %v", err)
	}
	vc := lastVoteCast(t, g.PullEvents())
	if (vc.HumanVotesCast == vc.HumanVoters) != g.VotingComplete() {
		t.Fatalf("progress %d/%d disagrees with VotingComplete()=%v after first vote",
			vc.HumanVotesCast, vc.HumanVoters, g.VotingComplete())
	}
	if g.VotingComplete() {
		t.Fatal("expected VotingComplete false with one connected human left unvoted")
	}

	if err := g.CastVote(players[1].ID(), "SURVIVE"); err != nil {
		t.Fatalf("CastVote 2: %v", err)
	}
	vc = lastVoteCast(t, g.PullEvents())
	if (vc.HumanVotesCast == vc.HumanVoters) != g.VotingComplete() {
		t.Fatalf("progress %d/%d disagrees with VotingComplete()=%v after second vote",
			vc.HumanVotesCast, vc.HumanVoters, g.VotingComplete())
	}
	if !g.VotingComplete() {
		t.Fatal("expected VotingComplete true once every connected human has voted")
	}
}

// TestGame_Revote_ProgressStartsFull documents the revote-reset semantics:
// the tie that opens TIEBREAK resets the round's Ballot (see Ballot.Reset),
// so the revote window's progress genuinely restarts at 0 rather than
// carrying every vote from the tied round over.
func TestGame_Revote_ProgressStartsFull(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 2)
	assignAndOpenVoting(t, g)

	if err := g.CastVote(players[0].ID(), "SURVIVE"); err != nil {
		t.Fatalf("CastVote 1: %v", err)
	}
	if err := g.CastVote(players[1].ID(), "FALL"); err != nil {
		t.Fatalf("CastVote 2: %v", err)
	}
	tied, err := g.CloseVoting()
	if err != nil || !tied {
		t.Fatalf("expected a tie on first close, tied=%v err=%v", tied, err)
	}
	if g.State() != enums.Tiebreak {
		t.Fatalf("state = %v, want TIEBREAK", g.State())
	}
	if g.VotingComplete() {
		t.Fatal("expected VotingComplete false right after the revote opens (ballot was reset)")
	}

	if err := g.CastVote(players[0].ID(), "SURVIVE"); err != nil {
		t.Fatalf("CastVote (revote) 1: %v", err)
	}
	vc := lastVoteCast(t, g.PullEvents())
	if vc.HumanVotesCast != 1 || vc.HumanVoters != 2 {
		t.Fatalf("first revote cast = %d/%d, want 1/2", vc.HumanVotesCast, vc.HumanVoters)
	}
	if g.VotingComplete() {
		t.Fatal("expected VotingComplete still false with one revote holdout")
	}
}
