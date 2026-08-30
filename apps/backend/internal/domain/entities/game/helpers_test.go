package game_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// fakeRandom is a deterministic game.RandomSource: IntN pops the next
// queued value (mod n), cycling once exhausted. A nil/empty queue always
// returns 0.
type fakeRandom struct {
	seq []int
	i   int
}

func (f *fakeRandom) IntN(n int) int {
	if n <= 0 || len(f.seq) == 0 {
		return 0
	}
	v := f.seq[f.i%len(f.seq)]
	f.i++
	return v % n
}

func mustStand(t *testing.T, id byte, name string, rarity enums.PowerRarity) *powers.Stand {
	t.Helper()
	skills := []string{"skill"}
	power, err := powers.NewPower(powers.PowerID{id}, name, "description", rarity, &skills, "")
	if err != nil {
		t.Fatalf("NewPower(%q): %v", name, err)
	}
	stand, err := powers.NewStand(*power, enums.A, enums.A, enums.A, enums.A, enums.A, enums.A, nil)
	if err != nil {
		t.Fatalf("NewStand(%q): %v", name, err)
	}
	return stand
}

func mustDevilFruit(t *testing.T, id byte, name string, rarity enums.PowerRarity, fruitType enums.FruitType) *powers.DevilFruit {
	t.Helper()
	skills := []string{"skill"}
	power, err := powers.NewPower(powers.PowerID{id}, name, "description", rarity, &skills, "")
	if err != nil {
		t.Fatalf("NewPower(%q): %v", name, err)
	}
	fruit, err := powers.NewDevilFruit(*power, fruitType)
	if err != nil {
		t.Fatalf("NewDevilFruit(%q): %v", name, err)
	}
	return fruit
}

func mustHumanParticipant(t *testing.T, id, userID, team byte) *game.Participant {
	t.Helper()
	p, err := game.NewHumanParticipant(game.ParticipantID{id}, user.UserID{userID}, "player", game.TeamID{team})
	if err != nil {
		t.Fatalf("NewHumanParticipant: %v", err)
	}
	return p
}

func mustBotParticipant(t *testing.T, id, team byte) *game.Participant {
	t.Helper()
	p, err := game.NewBotParticipant(game.ParticipantID{id}, "bot", game.TeamID{team})
	if err != nil {
		t.Fatalf("NewBotParticipant: %v", err)
	}
	return p
}

func mustTeam(t *testing.T, id byte, name string) *game.Team {
	t.Helper()
	team, err := game.NewTeam(game.TeamID{id}, name, 0)
	if err != nil {
		t.Fatalf("NewTeam: %v", err)
	}
	return team
}

func mustStage(t *testing.T, id byte, manga enums.Manga, order int, name string) game.Stage {
	t.Helper()
	s, err := game.NewStage(game.StageID{id}, manga, order, name, "a test stage", "")
	if err != nil {
		t.Fatalf("NewStage: %v", err)
	}
	return s
}

func oneStage(t *testing.T) []game.Stage {
	return []game.Stage{mustStage(t, 1, enums.Jojo, 0, "Phantom Blood")}
}

func someStages(t *testing.T) []game.Stage {
	return []game.Stage{
		mustStage(t, 1, enums.Jojo, 0, "Phantom Blood"),
		mustStage(t, 2, enums.Jojo, 1, "Battle Tendency"),
		mustStage(t, 3, enums.Jojo, 2, "Stardust Crusaders"),
	}
}

// newGauntletGame builds a single-team Gauntlet Game with a host plus
// (players-1) additional joined human participants, all on the same team.
func newGauntletGame(t *testing.T, stages []game.Stage, players int) (*game.Game, []*game.Participant) {
	t.Helper()
	cfg, err := game.NewConfig(enums.Gauntlet, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo}, enums.Random, game.MaxGauntletPlayers, false, enums.Private, 30, game.PoolFilter{}, enums.Normal)
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	host := mustHumanParticipant(t, 1, 1, 100)
	team := mustTeam(t, 100, "Squad")
	g, err := game.NewGame(game.GameID{1}, cfg, host, []*game.Team{team}, stages)
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	participants := []*game.Participant{host}
	for i := 1; i < players; i++ {
		p := mustHumanParticipant(t, byte(i+1), byte(i+1), 100)
		if err := g.Join(p); err != nil {
			t.Fatalf("Join: %v", err)
		}
		participants = append(participants, p)
	}
	return g, participants
}

// assignAndOpenVoting drives a Gauntlet Game from LOBBY (Start must not
// have been called yet) through ASSIGNING into VOTING, with an empty
// power pool (the state-machine tests below don't care which powers get
// drawn).
func assignAndOpenVoting(t *testing.T, g *game.Game) {
	t.Helper()
	assignOnly(t, g)
	if err := g.OpenVoting(&fakeRandom{}); err != nil {
		t.Fatalf("OpenVoting: %v", err)
	}
}

// assignOnly drives a Gauntlet Game from LOBBY (Start must not have been
// called yet) through Start+AssignLoadouts, leaving it parked in ASSIGNING
// - the window the sorteo's MarkRevealReady/RevealReadyComplete operate in,
// before OpenVoting would move it on.
func assignOnly(t *testing.T, g *game.Game) {
	t.Helper()
	if err := g.Start(g.HostID()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	cfg := g.Config()
	builder := game.NewLoadoutBuilder(cfg.PowerMangas(), game.DefaultAssignmentWeights(), &fakeRandom{})
	pools := make(map[game.TeamID]*game.AvailablePowers, len(g.Teams()))
	for _, tm := range g.Teams() {
		pools[tm.ID()] = game.NewAvailablePowers(nil, nil)
	}
	if err := g.AssignLoadouts(builder, pools); err != nil {
		t.Fatalf("AssignLoadouts: %v", err)
	}
}
