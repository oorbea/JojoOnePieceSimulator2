package game_test

import (
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// newVersusGame builds a 1v1 Versus Game with host on teamA and rival on
// teamB, still in LOBBY.
func newVersusGame(t *testing.T) (g *game.Game, host, rival *game.Participant, teamA, teamB *game.Team) {
	t.Helper()
	cfg, err := game.NewConfig(enums.Versus, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo}, enums.Random, 1, false, enums.Private, 30, game.PoolFilter{}, enums.Normal, game.DefaultSummaryDurationSeconds)
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	teamA = mustTeam(t, 100, "A")
	teamB = mustTeam(t, 101, "B")
	host = mustHumanParticipant(t, 1, 1, 100)
	rival = mustHumanParticipant(t, 2, 2, 101)
	g, err = game.NewGame(game.GameID{1}, cfg, host, []*game.Team{teamA, teamB}, someStages(t))
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	if err := g.Join(rival); err != nil {
		t.Fatalf("Join: %v", err)
	}
	return g, host, rival, teamA, teamB
}

func hasEvent[T game.DomainEvent](events []game.DomainEvent) bool {
	for _, e := range events {
		if _, ok := e.(T); ok {
			return true
		}
	}
	return false
}

func TestGame_Disconnect_EmitsPlayerDisconnected(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 2)
	g.PullEvents()

	if err := g.Disconnect(players[1].ID(), &fakeRandom{}, time.Now()); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	events := g.PullEvents()
	if !hasEvent[game.PlayerDisconnected](events) {
		t.Fatalf("expected PlayerDisconnected event, got %#v", events)
	}
}

func TestGame_Reconnect_EmitsPlayerReconnectedAndReelectsHost(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 2)
	rng := &fakeRandom{seq: []int{0}}

	// Host disconnects, leaving the other player as the only active human -
	// reassignHost promotes them immediately.
	if err := g.Disconnect(g.HostID(), rng, time.Now()); err != nil {
		t.Fatalf("Disconnect host: %v", err)
	}
	if g.HostID() != players[1].ID() {
		t.Fatalf("expected host reassigned to %v, got %v", players[1].ID(), g.HostID())
	}

	// Now the (new) host also disconnects: nobody is left connected, so
	// hostID goes nil (host is disconnected but not yet abandoned, so
	// hasActiveHuman still holds and the game does not abort).
	if err := g.Disconnect(players[1].ID(), rng, time.Now()); err != nil {
		t.Fatalf("Disconnect remaining host: %v", err)
	}
	if !g.HostID().IsNil() {
		t.Fatalf("expected nil host with nobody connected, got %v", g.HostID())
	}
	if g.State() == enums.Aborted {
		t.Fatalf("a disconnected-but-not-abandoned lobby must not abort")
	}
	g.PullEvents()

	// The first one to reconnect must both regain control AND re-elect a
	// host, or the lobby is stuck with no host and no way to get one.
	if err := g.Reconnect(players[1].ID(), rng); err != nil {
		t.Fatalf("Reconnect: %v", err)
	}
	if g.HostID() != players[1].ID() {
		t.Fatalf("expected reconnecting player to become host, got %v", g.HostID())
	}
	events := g.PullEvents()
	if !hasEvent[game.PlayerReconnected](events) {
		t.Fatalf("expected PlayerReconnected event, got %#v", events)
	}
	if !hasEvent[game.HostReassigned](events) {
		t.Fatalf("expected HostReassigned event on reconnect-into-hostless-lobby, got %#v", events)
	}
}

func TestGame_Abandon_VersusDoesNotShrinkTeam(t *testing.T) {
	g, _, rival, teamA, _ := newVersusGame(t)
	if err := g.Start(g.HostID()); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if err := g.Disconnect(rival.ID(), &fakeRandom{}, time.Now()); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if err := g.Abandon(rival.ID()); err != nil {
		t.Fatalf("Abandon: %v", err)
	}

	if g.State() == enums.Aborted {
		t.Fatalf("abandoning one of two active humans must not abort a Versus match")
	}
	if teamA.Size() != 1 {
		t.Fatalf("abandoning must not touch the abandoned team's size, got %d", teamA.Size())
	}
	for _, tm := range g.Teams() {
		if tm.Size() == 0 {
			t.Fatalf("an abandoned seat must still occupy its team slot, team %v is empty", tm.ID())
		}
	}
}

func TestGame_Abandon_AbortsWhenLastActiveHuman(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 1)

	if err := g.Disconnect(players[0].ID(), &fakeRandom{}, time.Now()); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if g.State() == enums.Aborted {
		t.Fatalf("a disconnected-but-not-abandoned solo lobby must not abort yet")
	}
	if err := g.Abandon(players[0].ID()); err != nil {
		t.Fatalf("Abandon: %v", err)
	}
	if g.State() != enums.Aborted {
		t.Fatalf("expected ABORTED once the only human is abandoned, got %v", g.State())
	}
}

func TestGame_Abandon_RetainsUserIDAndLoadout(t *testing.T) {
	g, _, rival, _, _ := newVersusGame(t)
	if err := g.Start(g.HostID()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	builder := game.NewLoadoutBuilder(g.Config().PowerMangas(), game.DefaultAssignmentWeights(), &fakeRandom{})
	pools := map[game.TeamID]*game.AvailablePowers{}
	for _, tm := range g.Teams() {
		pools[tm.ID()] = game.NewAvailablePowers(nil, nil)
	}
	if err := g.AssignLoadouts(builder, pools); err != nil {
		t.Fatalf("AssignLoadouts: %v", err)
	}
	before, ok := g.Participant(rival.ID())
	if !ok {
		t.Fatalf("rival not found before Abandon")
	}
	beforeUserID := before.UserID()
	beforeLoadout := before.Loadout()

	if err := g.Disconnect(rival.ID(), &fakeRandom{}, time.Now()); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if err := g.Abandon(rival.ID()); err != nil {
		t.Fatalf("Abandon: %v", err)
	}

	after, ok := g.Participant(rival.ID())
	if !ok {
		t.Fatalf("abandoned seat must still be present in the Game")
	}
	if after.UserID() == nil || *after.UserID() != *beforeUserID {
		t.Fatalf("abandoned seat must keep its userID, got %v want %v", after.UserID(), beforeUserID)
	}
	if after.Loadout() != beforeLoadout {
		t.Fatalf("abandoned seat must keep its assigned loadout")
	}
	if !after.Abandoned() {
		t.Fatalf("expected Abandoned() true")
	}
}

func TestGame_CastBotVotes_VersusAutoVotesAbandonedSeatEvenWithoutBots(t *testing.T) {
	g, host, rival, _, teamB := newVersusGame(t)
	// AllowBots is false in newVersusGame's config - an abandoned seat must
	// still auto-vote, since it is not a bot in the config sense.
	if g.Config().AllowBots() {
		t.Fatalf("test setup assumption violated: expected AllowBots() == false")
	}
	if err := g.Start(g.HostID()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	builder := game.NewLoadoutBuilder(g.Config().PowerMangas(), game.DefaultAssignmentWeights(), &fakeRandom{})
	pools := map[game.TeamID]*game.AvailablePowers{}
	for _, tm := range g.Teams() {
		pools[tm.ID()] = game.NewAvailablePowers(nil, nil)
	}
	if err := g.AssignLoadouts(builder, pools); err != nil {
		t.Fatalf("AssignLoadouts: %v", err)
	}

	if err := g.Disconnect(rival.ID(), &fakeRandom{}, time.Now()); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if err := g.Abandon(rival.ID()); err != nil {
		t.Fatalf("Abandon: %v", err)
	}

	if err := g.OpenVoting(&fakeRandom{}); err != nil {
		t.Fatalf("OpenVoting: %v", err)
	}
	rounds := g.Rounds()
	round := rounds[len(rounds)-1]
	if !round.Ballot.HasVoted(rival.ID()) {
		t.Fatalf("expected the abandoned seat to have an auto-cast vote")
	}
	optA := game.OptionID(g.Teams()[0].ID().String())
	optB := game.OptionID(teamB.ID().String())
	votes := round.Ballot.Votes()
	if votes[rival.ID()] != optA && votes[rival.ID()] != optB {
		t.Fatalf("expected abandoned seat to auto-vote for a valid option, got %v", votes[rival.ID()])
	}

	// The human-only progress counters must not count the abandoned seat.
	if err := g.CastVote(host.ID(), optA); err != nil {
		t.Fatalf("CastVote host: %v", err)
	}
	if !g.VotingComplete() {
		t.Fatalf("expected VotingComplete once the sole remaining active human has voted")
	}
}

func TestGame_CastBotVotes_GauntletDoesNotAutoVoteAbandonedSeat(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 2)
	assignOnly(t, g)

	if err := g.Disconnect(players[1].ID(), &fakeRandom{}, time.Now()); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if err := g.Abandon(players[1].ID()); err != nil {
		t.Fatalf("Abandon: %v", err)
	}
	if err := g.OpenVoting(&fakeRandom{}); err != nil {
		t.Fatalf("OpenVoting: %v", err)
	}

	rounds := g.Rounds()
	round := rounds[len(rounds)-1]
	if round.Ballot.HasVoted(players[1].ID()) {
		t.Fatalf("Gauntlet must never auto-vote for an abandoned seat")
	}
	// But the abandoned seat's loadout must still count for the squad.
	p, ok := g.Participant(players[1].ID())
	if !ok {
		t.Fatalf("abandoned seat must still be seated")
	}
	if p.Loadout() == nil {
		t.Fatalf("abandoned seat must still carry its assigned loadout")
	}
}

func TestGame_Reconnect_OverwritesAutoCastVote(t *testing.T) {
	g, _, rival, _, teamA := newVersusGame(t)
	if err := g.Start(g.HostID()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	builder := game.NewLoadoutBuilder(g.Config().PowerMangas(), game.DefaultAssignmentWeights(), &fakeRandom{})
	pools := map[game.TeamID]*game.AvailablePowers{}
	for _, tm := range g.Teams() {
		pools[tm.ID()] = game.NewAvailablePowers(nil, nil)
	}
	if err := g.AssignLoadouts(builder, pools); err != nil {
		t.Fatalf("AssignLoadouts: %v", err)
	}
	if err := g.Disconnect(rival.ID(), &fakeRandom{}, time.Now()); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if err := g.Abandon(rival.ID()); err != nil {
		t.Fatalf("Abandon: %v", err)
	}
	if err := g.OpenVoting(&fakeRandom{}); err != nil {
		t.Fatalf("OpenVoting: %v", err)
	}

	if err := g.Reconnect(rival.ID(), &fakeRandom{}); err != nil {
		t.Fatalf("Reconnect: %v", err)
	}
	optA := game.OptionID(teamA.ID().String())
	if err := g.CastVote(rival.ID(), optA); err != nil {
		t.Fatalf("CastVote after reconnect: %v", err)
	}
	rounds := g.Rounds()
	votes := rounds[len(rounds)-1].Ballot.Votes()
	if votes[rival.ID()] != optA {
		t.Fatalf("expected reconnected player's own vote to win, got %v", votes[rival.ID()])
	}
}
