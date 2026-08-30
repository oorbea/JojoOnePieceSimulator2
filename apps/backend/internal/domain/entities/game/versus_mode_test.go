package game_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func TestVersusMode_PlaysExactlyThreeRoundsAndTracksWins(t *testing.T) {
	cfg, err := game.NewConfig(enums.Versus, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo}, enums.Random, 1, false, enums.Private, 30, game.PoolFilter{}, enums.Normal, game.DefaultSummaryDurationSeconds)
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	teamA := mustTeam(t, 100, "A")
	teamB := mustTeam(t, 101, "B")
	host := mustHumanParticipant(t, 1, 1, 100)
	rival := mustHumanParticipant(t, 2, 2, 101)
	g, err := game.NewGame(game.GameID{1}, cfg, host, []*game.Team{teamA, teamB}, someStages(t))
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	if err := g.Join(rival); err != nil {
		t.Fatalf("Join: %v", err)
	}
	if err := g.Start(g.HostID()); err != nil {
		t.Fatalf("Start: %v", err)
	}

	optionA := game.OptionID(teamA.ID().String())
	optionB := game.OptionID(teamB.ID().String())
	// Team A wins rounds 1 and 3, team B wins round 2 -> 2-1 for A.
	winners := []game.OptionID{optionA, optionB, optionA}

	for i, winner := range winners {
		builder := game.NewLoadoutBuilder(cfg.PowerMangas(), game.DefaultAssignmentWeights(), &fakeRandom{})
		pools := map[game.TeamID]*game.AvailablePowers{
			teamA.ID(): game.NewAvailablePowers(nil, nil),
			teamB.ID(): game.NewAvailablePowers(nil, nil),
		}
		if err := g.AssignLoadouts(builder, pools); err != nil {
			t.Fatalf("round %d AssignLoadouts: %v", i, err)
		}
		roundsBefore := len(g.Rounds())
		if err := g.OpenVoting(&fakeRandom{seq: []int{i}}); err != nil {
			t.Fatalf("round %d OpenVoting: %v", i, err)
		}
		if len(g.Rounds()) != roundsBefore+1 {
			t.Fatalf("round %d: expected a new round to be recorded", i)
		}
		if err := g.CastVote(host.ID(), winner); err != nil {
			t.Fatalf("round %d CastVote host: %v", i, err)
		}
		if err := g.CastVote(rival.ID(), winner); err != nil {
			t.Fatalf("round %d CastVote rival: %v", i, err)
		}
		if tied, err := g.CloseVoting(); err != nil || tied {
			t.Fatalf("round %d CloseVoting: tied=%v err=%v", i, tied, err)
		}
		if err := g.CompleteRound(); err != nil {
			t.Fatalf("round %d CompleteRound: %v", i, err)
		}
	}

	if g.State() != enums.Finished {
		t.Fatalf("expected FINISHED after 3 rounds, got %v", g.State())
	}
	if len(g.Rounds()) != game.VersusRounds {
		t.Fatalf("expected exactly %d rounds, got %d", game.VersusRounds, len(g.Rounds()))
	}
	result, err := g.Result()
	if err != nil {
		t.Fatalf("Result: %v", err)
	}
	if result.Winner != optionA {
		t.Fatalf("expected team A (2 round wins) to win the match, got %v", result.Winner)
	}
}

func TestVersusMode_ReassignsLoadoutsEachRound(t *testing.T) {
	if !(game.VersusMode{}).ReassignsEachRound() {
		t.Fatalf("expected Versus to reassign loadouts every round")
	}
}

func TestVersusMode_StageIsRandomPerRound(t *testing.T) {
	cfg, err := game.NewConfig(enums.Versus, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo}, enums.Random, 1, false, enums.Private, 30, game.PoolFilter{}, enums.Normal, game.DefaultSummaryDurationSeconds)
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	teamA := mustTeam(t, 100, "A")
	teamB := mustTeam(t, 101, "B")
	host := mustHumanParticipant(t, 1, 1, 100)
	rival := mustHumanParticipant(t, 2, 2, 101)
	stages := someStages(t)
	g, err := game.NewGame(game.GameID{1}, cfg, host, []*game.Team{teamA, teamB}, stages)
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	if err := g.Join(rival); err != nil {
		t.Fatalf("Join: %v", err)
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
	// Force the "random" pick to select stages[2] (Stardust Crusaders).
	if err := g.OpenVoting(&fakeRandom{seq: []int{2}}); err != nil {
		t.Fatalf("OpenVoting: %v", err)
	}
	rounds := g.Rounds()
	if rounds[0].Stage.Name() != stages[2].Name() {
		t.Fatalf("expected the rng-selected stage %q, got %q", stages[2].Name(), rounds[0].Stage.Name())
	}
}
