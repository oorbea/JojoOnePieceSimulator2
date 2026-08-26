package dto_test

import (
	"context"
	"testing"

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
	cfg, err := game.NewConfig(enums.Gauntlet, []enums.Manga{enums.Jojo}, enums.Random, game.MaxGauntletPlayers, false, enums.Private, 30, game.PoolFilter{})
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
		noStageText, esResolver, noFruitText, nil)
	if err != nil {
		t.Fatalf("NewGameStateResponse (es-ES): %v", err)
	}
	caResp, err := dto.NewGameStateResponse(context.Background(), g, "ABC123", host.ID(),
		noPictures, noPictures, noPictures, noPictures,
		noStageText, caResolver, noFruitText, nil)
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
		noStageText, empty, empty, nil)
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
	cfg, err := game.NewConfig(enums.Versus, []enums.Manga{enums.Jojo}, enums.Random, 2, true, enums.Private, 30, game.PoolFilter{})
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
		noStageText, noFruitText, noFruitText, nil)
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
