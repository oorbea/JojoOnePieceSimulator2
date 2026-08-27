package game_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// newVersusLobby builds a Versus lobby with host on team A and other on
// team B - one per team, so it's valid for any teamSize >= 1.
func newVersusLobby(t *testing.T, teamSize int) (*game.Game, *game.Participant, *game.Participant, *game.Team, *game.Team) {
	t.Helper()
	cfg, err := game.NewConfig(enums.Versus, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo}, enums.Random, teamSize, false, enums.Private, 30, game.PoolFilter{})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	teamA := mustTeam(t, 100, "A")
	teamB := mustTeam(t, 101, "B")
	host := mustHumanParticipant(t, 1, 1, 100)
	other := mustHumanParticipant(t, 2, 2, 101)
	g, err := game.NewGame(game.GameID{1}, cfg, host, []*game.Team{teamA, teamB}, someStages(t))
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	if err := g.Join(other); err != nil {
		t.Fatalf("Join: %v", err)
	}
	return g, host, other, teamA, teamB
}

func TestSwitchTeam_SelfMove(t *testing.T) {
	g, _, other, teamA, teamB := newVersusLobby(t, 2)
	if err := g.SwitchTeam(other.ID(), other.ID(), teamA.ID()); err != nil {
		t.Fatalf("SwitchTeam: %v", err)
	}
	if other.TeamID() != teamA.ID() {
		t.Fatalf("expected other to be on team A, got %v", other.TeamID())
	}
	if !teamA.HasMember(other.ID()) {
		t.Fatalf("expected team A's member list to include other")
	}
	if teamB.HasMember(other.ID()) {
		t.Fatalf("expected team B's member list to no longer include other")
	}
}

func TestSwitchTeam_HostMovesSomeoneElse(t *testing.T) {
	g, host, other, teamA, _ := newVersusLobby(t, 2)
	if err := g.SwitchTeam(host.ID(), other.ID(), teamA.ID()); err != nil {
		t.Fatalf("SwitchTeam: %v", err)
	}
	if other.TeamID() != teamA.ID() {
		t.Fatalf("expected other to be on team A, got %v", other.TeamID())
	}
}

func TestSwitchTeam_NonHostCannotMoveSomeoneElse(t *testing.T) {
	g, host, other, _, teamB := newVersusLobby(t, 2)
	if err := g.SwitchTeam(other.ID(), host.ID(), teamB.ID()); err != game.ErrNotHost {
		t.Fatalf("expected ErrNotHost, got %v", err)
	}
}

func TestSwitchTeam_RejectsFullTeam(t *testing.T) {
	g, host, _, teamA, teamB := newVersusLobby(t, 1)
	// Both teams are already at capacity 1/1 - moving the host onto team B
	// (already holding other) must be rejected as full.
	if err := g.SwitchTeam(host.ID(), host.ID(), teamB.ID()); err != game.ErrTeamFull {
		t.Fatalf("expected ErrTeamFull, got %v", err)
	}
	if host.TeamID() != teamA.ID() {
		t.Fatalf("expected the rejected move to leave host on team A")
	}
}

func TestSwitchTeam_NoOpWhenAlreadyThere(t *testing.T) {
	g, host, _, teamA, _ := newVersusLobby(t, 2)
	if err := g.SwitchTeam(host.ID(), host.ID(), teamA.ID()); err != nil {
		t.Fatalf("expected a no-op success, got %v", err)
	}
}

func TestKick_HostOnly(t *testing.T) {
	g, host, other, _, _ := newVersusLobby(t, 2)
	if err := g.Kick(other.ID(), host.ID(), &fakeRandom{}); err != game.ErrNotHost {
		t.Fatalf("expected ErrNotHost, got %v", err)
	}
	if err := g.Kick(host.ID(), other.ID(), &fakeRandom{}); err != nil {
		t.Fatalf("Kick: %v", err)
	}
	if _, ok := g.Participant(other.ID()); ok {
		t.Fatalf("expected other to be removed from the game")
	}
}

func TestKick_CannotKickSelf(t *testing.T) {
	g, host, _, _, _ := newVersusLobby(t, 2)
	if err := g.Kick(host.ID(), host.ID(), &fakeRandom{}); err != game.ErrCannotKickSelf {
		t.Fatalf("expected ErrCannotKickSelf, got %v", err)
	}
}

func TestKick_EmitsPlayerKickedBeforePlayerLeft(t *testing.T) {
	g, host, other, _, _ := newVersusLobby(t, 2)
	g.PullEvents() // drain the PlayerJoined from newVersusLobby's Join
	if err := g.Kick(host.ID(), other.ID(), &fakeRandom{}); err != nil {
		t.Fatalf("Kick: %v", err)
	}
	events := g.PullEvents()
	if len(events) < 2 {
		t.Fatalf("expected at least PlayerKicked+PlayerLeft, got %+v", events)
	}
	if _, ok := events[0].(game.PlayerKicked); !ok {
		t.Fatalf("expected the first event to be PlayerKicked, got %T", events[0])
	}
}

func TestTransferHost_HostOnly(t *testing.T) {
	g, host, other, _, _ := newVersusLobby(t, 2)
	if err := g.TransferHost(other.ID(), host.ID()); err != game.ErrNotHost {
		t.Fatalf("expected ErrNotHost, got %v", err)
	}
	if err := g.TransferHost(host.ID(), other.ID()); err != nil {
		t.Fatalf("TransferHost: %v", err)
	}
	if g.HostID() != other.ID() {
		t.Fatalf("expected other to be the new host, got %v", g.HostID())
	}
}

func TestTransferHost_RejectsBot(t *testing.T) {
	// A Config with AllowBots so the bot actually seats, isolating this
	// test to TransferHost's own kind check rather than AddBot's.
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
	bot, err := game.NewBotParticipant(game.ParticipantID{9}, "bot", teamB.ID())
	if err != nil {
		t.Fatalf("NewBotParticipant: %v", err)
	}
	if err := g.AddBot(bot); err != nil {
		t.Fatalf("AddBot: %v", err)
	}
	if err := g.TransferHost(host.ID(), bot.ID()); err != game.ErrBotsNotAllowed {
		t.Fatalf("expected ErrBotsNotAllowed, got %v", err)
	}
}

func TestSetLocked_HostOnlyAndBlocksJoin(t *testing.T) {
	g, host, other, _, _ := newVersusLobby(t, 2)
	if err := g.SetLocked(other.ID(), true); err != game.ErrNotHost {
		t.Fatalf("expected ErrNotHost, got %v", err)
	}
	if err := g.SetLocked(host.ID(), true); err != nil {
		t.Fatalf("SetLocked: %v", err)
	}
	if !g.Locked() {
		t.Fatalf("expected the lobby to report locked")
	}
	third := mustHumanParticipant(t, 3, 3, 100)
	if err := g.Join(third); err != game.ErrLobbyLocked {
		t.Fatalf("expected ErrLobbyLocked, got %v", err)
	}
}

func TestReconfigure_HostOnlyAndLobbyOnly(t *testing.T) {
	g, host, other, teamA, teamB := newVersusLobby(t, 2)
	next, err := game.NewConfig(enums.Versus, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo}, enums.Random, 2, false, enums.Public, 45, game.PoolFilter{})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	if err := g.Reconfigure(other.ID(), next, []*game.Team{teamA, teamB}, someStages(t)); err != game.ErrNotHost {
		t.Fatalf("expected ErrNotHost, got %v", err)
	}
	if err := g.Reconfigure(host.ID(), next, []*game.Team{teamA, teamB}, someStages(t)); err != nil {
		t.Fatalf("Reconfigure: %v", err)
	}
	if g.Config().Visibility() != enums.Public {
		t.Fatalf("expected the new Config to be applied")
	}
}

func TestReconfigure_GauntletToVersusSplitsPlayers(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 4)
	next, err := game.NewConfig(enums.Versus, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo}, enums.Random, 2, false, enums.Private, 30, game.PoolFilter{})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	teamA := mustTeam(t, 200, "A")
	teamB := mustTeam(t, 201, "B")
	if err := g.Reconfigure(g.HostID(), next, []*game.Team{teamA, teamB}, someStages(t)); err != nil {
		t.Fatalf("Reconfigure: %v", err)
	}
	if teamA.Size() != 2 || teamB.Size() != 2 {
		t.Fatalf("expected an even 2/2 split, got A=%d B=%d", teamA.Size(), teamB.Size())
	}
	for _, p := range players {
		if p.TeamID() != teamA.ID() && p.TeamID() != teamB.ID() {
			t.Fatalf("participant %v not seated on either new team", p.ID())
		}
	}
}

func TestReconfigure_VersusToGauntletMergesPlayers(t *testing.T) {
	g, host, other, _, _ := newVersusLobby(t, 2)
	next, err := game.NewConfig(enums.Gauntlet, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo}, enums.Random, 10, false, enums.Private, 30, game.PoolFilter{})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	squad := mustTeam(t, 30, "Squad")
	if err := g.Reconfigure(host.ID(), next, []*game.Team{squad}, oneStage(t)); err != nil {
		t.Fatalf("Reconfigure: %v", err)
	}
	if squad.Size() != 2 {
		t.Fatalf("expected both players merged onto the single squad, got %d", squad.Size())
	}
	if host.TeamID() != squad.ID() || other.TeamID() != squad.ID() {
		t.Fatalf("expected both participants reseated on squad")
	}
}

func TestReconfigure_ShrinkingBelowSeatedHumansIsRejected(t *testing.T) {
	g, host, _, teamA, teamB := newVersusLobby(t, 2)
	third, err := game.NewHumanParticipant(game.ParticipantID{3}, user.UserID{3}, "p3", teamA.ID())
	if err != nil {
		t.Fatalf("NewHumanParticipant: %v", err)
	}
	if err := g.Join(third); err != nil {
		t.Fatalf("Join: %v", err)
	}
	// Team A now has 2 players (host + third). Shrinking teamSize to 1
	// would evict a seated human and must be rejected.
	shrunk, err := game.NewConfig(enums.Versus, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo}, enums.Random, 1, false, enums.Private, 30, game.PoolFilter{})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	before := g.Config()
	if err := g.Reconfigure(host.ID(), shrunk, []*game.Team{teamA, teamB}, someStages(t)); err != game.ErrConfigWouldEvictPlayers {
		t.Fatalf("expected ErrConfigWouldEvictPlayers, got %v", err)
	}
	if g.Config().TeamSize() != before.TeamSize() {
		t.Fatalf("expected the rejected Reconfigure to leave Config unchanged")
	}
}
