package dto_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

// buildLoadoutTestGame builds a one-participant Gauntlet game whose host
// already carries a Loadout with a Stand - built directly via AssignLoadout
// (like redis/wire_test.go's buildTestGame), never through the weighted
// LoadoutBuilder draw, so which Stand ends up in the loadout is exact and
// deterministic instead of RNG-dependent.
func buildLoadoutTestGame(t *testing.T) (*game.Game, *powers.Stand) {
	t.Helper()
	cfg, err := game.NewConfig(enums.Gauntlet, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo}, enums.Random, game.MaxGauntletPlayers, false, enums.Private, 30, game.PoolFilter{})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	host, err := game.NewHumanParticipant(game.ParticipantID{1}, user.UserID{1}, "host", game.TeamID{10})
	if err != nil {
		t.Fatalf("NewHumanParticipant: %v", err)
	}
	host.SetAvatar("avatars/host/thumb.webp", "https://accounts.google.com/host.jpg")
	team, err := game.NewTeam(game.TeamID{10}, "Squad", 0)
	if err != nil {
		t.Fatalf("NewTeam: %v", err)
	}
	stage, err := game.NewStage(game.StageID{1}, enums.Jojo, 0, "Phantom Blood", "a test stage", "")
	if err != nil {
		t.Fatalf("NewStage: %v", err)
	}
	g, err := game.NewGame(game.GameID{1}, cfg, host, []*game.Team{team}, []game.Stage{stage})
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}

	skills := []string{"ORA ORA ORA"}
	power, err := powers.NewPower(powers.PowerID{50}, "Star Platinum", "A being beyond human.", enums.Legendary, &skills, "")
	if err != nil {
		t.Fatalf("NewPower: %v", err)
	}
	stand, err := powers.NewStand(*power, enums.A, enums.A, enums.A, enums.A, enums.A, enums.A, nil)
	if err != nil {
		t.Fatalf("NewStand: %v", err)
	}
	loadout, err := game.NewLoadout(stand, nil, enums.SpinInfinite, enums.HamonPerfect, enums.FruitMasteryNone,
		enums.HakiPrivate, enums.HakiPrivate, enums.HakiPrivate, enums.PhysicalFormPrivate)
	if err != nil {
		t.Fatalf("NewLoadout: %v", err)
	}
	host.AssignLoadout(loadout)

	return g, stand
}

// zeroRandom is the simplest game.RandomSource: always picks index 0. Good
// enough for OpenVoting's stage pick and LoadoutBuilder's draws in tests
// that don't care which stage/power comes out, only that voting opens.
type zeroRandom struct{}

func (zeroRandom) IntN(int) int { return 0 }

// buildTiedRoundGame builds a 2-player Gauntlet game with an open round
// tied 1-1 (host SURVIVE, second player FALL), for
// TestNewGameStateResponse_TiedVotes below.
func buildTiedRoundGame(t *testing.T) (*game.Game, *game.Participant) {
	t.Helper()
	cfg, err := game.NewConfig(enums.Gauntlet, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo}, enums.Random, game.MaxGauntletPlayers, false, enums.Private, 30, game.PoolFilter{})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	host, err := game.NewHumanParticipant(game.ParticipantID{1}, user.UserID{1}, "host", game.TeamID{10})
	if err != nil {
		t.Fatalf("NewHumanParticipant host: %v", err)
	}
	second, err := game.NewHumanParticipant(game.ParticipantID{2}, user.UserID{2}, "second", game.TeamID{10})
	if err != nil {
		t.Fatalf("NewHumanParticipant second: %v", err)
	}
	team, err := game.NewTeam(game.TeamID{10}, "Squad", 0)
	if err != nil {
		t.Fatalf("NewTeam: %v", err)
	}
	stage, err := game.NewStage(game.StageID{1}, enums.Jojo, 0, "Phantom Blood", "a test stage", "")
	if err != nil {
		t.Fatalf("NewStage: %v", err)
	}
	g, err := game.NewGame(game.GameID{1}, cfg, host, []*game.Team{team}, []game.Stage{stage})
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	if err := g.Join(second); err != nil {
		t.Fatalf("Join: %v", err)
	}
	if err := g.Start(g.HostID()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	pools := map[game.TeamID]*game.AvailablePowers{team.ID(): game.NewAvailablePowers(nil, nil)}
	builder := game.NewLoadoutBuilder(cfg.PowerMangas(), game.DefaultAssignmentWeights(), zeroRandom{})
	if err := g.AssignLoadouts(builder, pools); err != nil {
		t.Fatalf("AssignLoadouts: %v", err)
	}
	if err := g.OpenVoting(zeroRandom{}); err != nil {
		t.Fatalf("OpenVoting: %v", err)
	}
	if err := g.CastVote(host.ID(), "SURVIVE"); err != nil {
		t.Fatalf("CastVote host: %v", err)
	}
	if err := g.CastVote(second.ID(), "FALL"); err != nil {
		t.Fatalf("CastVote second: %v", err)
	}
	if tied, err := g.CloseVoting(); err != nil || !tied {
		t.Fatalf("CloseVoting: tied=%v err=%v", tied, err)
	}
	return g, host
}

// TestNewGameStateResponse_TiedVotes guards the owner's explicit call
// (2026-08-28): once a tie opens a revote, the tied round's TiedVotes is
// revealed in the response even though the round hasn't resolved (no
// Result yet) - unlike Votes, which stays hidden until Result is set.
func TestNewGameStateResponse_TiedVotes(t *testing.T) {
	g, host := buildTiedRoundGame(t)
	noFruitText := func(_ context.Context, _ powers.PowerID) (ports.PowerContent, error) {
		return ports.PowerContent{}, nil
	}

	resp, err := dto.NewGameStateResponse(context.Background(), g, "ABC123", host.ID(),
		noPictures, noPictures, noPictures, noPictures,
		noStageText, noFruitText, noFruitText, dto.GameStateDeadlines{})
	if err != nil {
		t.Fatalf("NewGameStateResponse: %v", err)
	}

	round := resp.Game.Rounds[len(resp.Game.Rounds)-1]
	if round.Result != nil {
		t.Fatalf("Result = %+v, want nil while still TIEBREAK", round.Result)
	}
	if len(round.TiedVotes) != 2 {
		t.Fatalf("TiedVotes = %+v, want the 2 votes that tied", round.TiedVotes)
	}
	if round.TiedVotes[host.ID().String()] != "SURVIVE" {
		t.Errorf("TiedVotes[host] = %q, want SURVIVE", round.TiedVotes[host.ID().String()])
	}

	raw, err := json.Marshal(resp.Game)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(raw), "tiedVotes") {
		t.Fatalf("expected tiedVotes in the marshaled response: %s", raw)
	}
}

func noPictures(_ context.Context, key string) (string, error) { return key, nil }

func noStageText(_ context.Context, _ game.StageID) (string, error) { return "", nil }

// TestNewGameStateResponse_LoadoutStandText_PerViewerLocale is the
// regression test for the sorteo backend change: RepoPowerPool freezes a
// loadout's Stand description+skills to en-GB at draw time (see
// infrastructure/game/repo_power_pool.go), since a live Game is one instance
// shared by every participant and can only ever carry one baked-in locale.
// NewGameStateResponse must re-resolve that text per call via
// resolveStandText instead of trusting what's frozen on the domain Stand -
// two calls for the SAME *game.Game, differing only in resolver, must
// produce different Stand text.
func TestNewGameStateResponse_LoadoutStandText_PerViewerLocale(t *testing.T) {
	g, stand := buildLoadoutTestGame(t)
	host := g.Participants()[0]

	esResolver := func(_ context.Context, id powers.PowerID) (ports.PowerContent, error) {
		if id != stand.ID() {
			t.Fatalf("resolveStandText called with unexpected id %s", id)
		}
		return ports.PowerContent{Description: "Un ser más allá de lo humano.", Skills: []string{"Parada de tiempo"}}, nil
	}
	caResolver := func(_ context.Context, id powers.PowerID) (ports.PowerContent, error) {
		if id != stand.ID() {
			t.Fatalf("resolveStandText called with unexpected id %s", id)
		}
		return ports.PowerContent{Description: "Un ésser més enllà de l'humà.", Skills: []string{"Aturada de temps"}}, nil
	}
	noFruitText := func(_ context.Context, _ powers.PowerID) (ports.PowerContent, error) {
		return ports.PowerContent{}, nil
	}

	esResp, err := dto.NewGameStateResponse(context.Background(), g, "ABC123", host.ID(),
		noPictures, noPictures, noPictures, noPictures,
		noStageText, esResolver, noFruitText, dto.GameStateDeadlines{})
	if err != nil {
		t.Fatalf("NewGameStateResponse (es-ES): %v", err)
	}
	caResp, err := dto.NewGameStateResponse(context.Background(), g, "ABC123", host.ID(),
		noPictures, noPictures, noPictures, noPictures,
		noStageText, caResolver, noFruitText, dto.GameStateDeadlines{})
	if err != nil {
		t.Fatalf("NewGameStateResponse (ca-ES): %v", err)
	}

	esStand := esResp.Game.Participants[0].Loadout.Stand
	caStand := caResp.Game.Participants[0].Loadout.Stand
	if esStand == nil || caStand == nil {
		t.Fatal("expected both responses to carry the loadout's Stand")
	}
	if esStand.Description != "Un ser más allá de lo humano." {
		t.Errorf("es-ES description = %q", esStand.Description)
	}
	if len(esStand.Skills) != 1 || esStand.Skills[0] != "Parada de tiempo" {
		t.Errorf("es-ES skills = %v", esStand.Skills)
	}
	if caStand.Description != "Un ésser més enllà de l'humà." {
		t.Errorf("ca-ES description = %q", caStand.Description)
	}
	if esStand.Description == caStand.Description {
		t.Fatal("expected the two locale-bound resolvers to produce different Stand text for the identical Stand")
	}
	// Name is a proper noun and must NOT be re-resolved - it stays whatever
	// the domain Stand (frozen at draw time) already carries.
	if esStand.Name != "Star Platinum" || caStand.Name != "Star Platinum" {
		t.Errorf("Stand name changed across locales: es=%q ca=%q", esStand.Name, caStand.Name)
	}
}

// TestNewGameStateResponse_LoadoutStandText_FallsBackOnMissingTranslation
// mirrors stageTextResolver's own defense-in-depth doc: a resolver with
// nothing for a locale must not error or panic, only surface whatever
// content it has (here, none - an empty PowerContent).
func TestNewGameStateResponse_LoadoutStandText_FallsBackOnMissingTranslation(t *testing.T) {
	g, _ := buildLoadoutTestGame(t)
	host := g.Participants()[0]

	empty := func(_ context.Context, _ powers.PowerID) (ports.PowerContent, error) {
		return ports.PowerContent{}, nil
	}

	resp, err := dto.NewGameStateResponse(context.Background(), g, "ABC123", host.ID(),
		noPictures, noPictures, noPictures, noPictures,
		noStageText, empty, empty, dto.GameStateDeadlines{})
	if err != nil {
		t.Fatalf("NewGameStateResponse: %v", err)
	}
	stand := resp.Game.Participants[0].Loadout.Stand
	if stand == nil {
		t.Fatal("expected the loadout's Stand to still be present")
	}
	if stand.Description != "" {
		t.Errorf("description = %q, want empty", stand.Description)
	}
	if len(stand.Skills) != 0 {
		t.Errorf("skills = %v, want empty (never nil)", stand.Skills)
	}
}

// TestNewGameStateResponse_ParticipantAvatar covers resolveParticipantAvatar's
// three cases through the public seam: a self-uploaded thumbnail wins over
// the Google-synced picture, a Google-only user falls back to it, and a bot
// (never called SetAvatar) resolves to "".
func TestNewGameStateResponse_ParticipantAvatar(t *testing.T) {
	// Versus + AllowBots: bots are Versus-only (see Game.AddBot), so a bot
	// fixture needs that combination even though the interesting assertions
	// here have nothing to do with team play.
	cfg, err := game.NewConfig(enums.Versus, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo}, enums.Random, 2, true, enums.Private, 30, game.PoolFilter{})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	host, err := game.NewHumanParticipant(game.ParticipantID{1}, user.UserID{1}, "host", game.TeamID{10})
	if err != nil {
		t.Fatalf("NewHumanParticipant: %v", err)
	}
	host.SetAvatar("avatars/host/thumb.webp", "https://accounts.google.com/host.jpg")
	teamA, err := game.NewTeam(game.TeamID{10}, "Team A", 0)
	if err != nil {
		t.Fatalf("NewTeam(A): %v", err)
	}
	teamB, err := game.NewTeam(game.TeamID{20}, "Team B", 0)
	if err != nil {
		t.Fatalf("NewTeam(B): %v", err)
	}
	stage, err := game.NewStage(game.StageID{1}, enums.Jojo, 0, "Phantom Blood", "a test stage", "")
	if err != nil {
		t.Fatalf("NewStage: %v", err)
	}
	g, err := game.NewGame(game.GameID{1}, cfg, host, []*game.Team{teamA, teamB}, []game.Stage{stage})
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}

	googleOnly, err := game.NewHumanParticipant(game.ParticipantID{2}, user.UserID{2}, "googleOnly", game.TeamID{20})
	if err != nil {
		t.Fatalf("NewHumanParticipant(googleOnly): %v", err)
	}
	googleOnly.SetAvatar("", "https://accounts.google.com/only.jpg")
	if err := g.Join(googleOnly); err != nil {
		t.Fatalf("Join(googleOnly): %v", err)
	}

	bot, err := game.NewBotParticipant(game.ParticipantID{3}, "Bot 1", game.TeamID{20})
	if err != nil {
		t.Fatalf("NewBotParticipant: %v", err)
	}
	if err := g.AddBot(bot); err != nil {
		t.Fatalf("AddBot: %v", err)
	}

	presignThumb := func(_ context.Context, key string) (string, error) {
		return "https://cdn.example/" + key, nil
	}
	noFruitText := func(_ context.Context, _ powers.PowerID) (ports.PowerContent, error) {
		return ports.PowerContent{}, nil
	}

	resp, err := dto.NewGameStateResponse(context.Background(), g, "ABC123", host.ID(),
		noPictures, noPictures, noPictures, presignThumb,
		noStageText, noFruitText, noFruitText, dto.GameStateDeadlines{})
	if err != nil {
		t.Fatalf("NewGameStateResponse: %v", err)
	}

	byID := map[string]string{}
	for _, p := range resp.Game.Participants {
		byID[p.ID] = p.AvatarThumb
	}
	if got := byID[host.ID().String()]; got != "https://cdn.example/avatars/host/thumb.webp" {
		t.Errorf("host avatar = %q, want the presigned upload URL", got)
	}
	if got := byID[googleOnly.ID().String()]; got != "https://accounts.google.com/only.jpg" {
		t.Errorf("google-only avatar = %q, want the raw Google URL, never presigned", got)
	}
	if got := byID[bot.ID().String()]; got != "" {
		t.Errorf("bot avatar = %q, want empty", got)
	}
}

// TestNewGameStateResponse_Deadlines guards GameStateDeadlines end to end:
// each field round-trips to its own RFC3339 response field, independently
// of the other, and a zero-value GameStateDeadlines marshals neither key at
// all (the omitempty on both).
func TestNewGameStateResponse_Deadlines(t *testing.T) {
	g, _ := buildLoadoutTestGame(t)
	host := g.Participants()[0]
	noFruitText := func(_ context.Context, _ powers.PowerID) (ports.PowerContent, error) {
		return ports.PowerContent{}, nil
	}

	votingEndsAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	resp, err := dto.NewGameStateResponse(context.Background(), g, "ABC123", host.ID(),
		noPictures, noPictures, noPictures, noPictures,
		noStageText, noFruitText, noFruitText,
		dto.GameStateDeadlines{VotingEndsAt: &votingEndsAt})
	if err != nil {
		t.Fatalf("NewGameStateResponse: %v", err)
	}
	if resp.Game.RevealEndsAt != nil {
		t.Errorf("RevealEndsAt = %v, want nil (only VotingEndsAt was set)", *resp.Game.RevealEndsAt)
	}
	if resp.Game.VotingEndsAt == nil || *resp.Game.VotingEndsAt != votingEndsAt.Format(time.RFC3339) {
		t.Errorf("VotingEndsAt = %v, want %v", resp.Game.VotingEndsAt, votingEndsAt.Format(time.RFC3339))
	}

	raw, err := json.Marshal(resp.Game)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	empty, err := dto.NewGameStateResponse(context.Background(), g, "ABC123", host.ID(),
		noPictures, noPictures, noPictures, noPictures,
		noStageText, noFruitText, noFruitText, dto.GameStateDeadlines{})
	if err != nil {
		t.Fatalf("NewGameStateResponse (zero deadlines): %v", err)
	}
	emptyRaw, err := json.Marshal(empty.Game)
	if err != nil {
		t.Fatalf("Marshal (zero deadlines): %v", err)
	}
	if strings.Contains(string(emptyRaw), "revealEndsAt") || strings.Contains(string(emptyRaw), "votingEndsAt") {
		t.Fatalf("zero-value GameStateDeadlines marshaled a deadline key: %s", emptyRaw)
	}
	if !strings.Contains(string(raw), "votingEndsAt") {
		t.Fatalf("expected votingEndsAt in the marshaled response: %s", raw)
	}

	resultEndsAt := time.Date(2026, 1, 1, 12, 0, 6, 0, time.UTC)
	resultResp, err := dto.NewGameStateResponse(context.Background(), g, "ABC123", host.ID(),
		noPictures, noPictures, noPictures, noPictures,
		noStageText, noFruitText, noFruitText,
		dto.GameStateDeadlines{ResultEndsAt: &resultEndsAt})
	if err != nil {
		t.Fatalf("NewGameStateResponse (ResultEndsAt): %v", err)
	}
	if resultResp.Game.RevealEndsAt != nil || resultResp.Game.VotingEndsAt != nil {
		t.Errorf("RevealEndsAt/VotingEndsAt should stay nil when only ResultEndsAt was set")
	}
	if resultResp.Game.ResultEndsAt == nil || *resultResp.Game.ResultEndsAt != resultEndsAt.Format(time.RFC3339) {
		t.Errorf("ResultEndsAt = %v, want %v", resultResp.Game.ResultEndsAt, resultEndsAt.Format(time.RFC3339))
	}
	if strings.Contains(string(emptyRaw), "resultEndsAt") {
		t.Fatalf("zero-value GameStateDeadlines marshaled resultEndsAt: %s", emptyRaw)
	}
}
