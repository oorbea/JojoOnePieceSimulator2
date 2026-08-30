package redis

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// buildTestGame builds a small Gauntlet game with one participant carrying
// a Loadout whose Stand has an EvolvesFrom parent, so the wire round trip
// exercises the recursive embedding path.
func buildTestGame(t *testing.T) *game.Game {
	t.Helper()
	// enums.Swift (not enums.Normal, the zero value / restore fallback) so
	// the round trip actually exercises RevealSpeed surviving - a dropped
	// field would silently decode back to Normal and this test would never
	// notice (see wire.go's wireConfig.RevealSpeed doc for the bug this
	// guards against: toWire/fromWire once omitted it entirely). Same
	// reasoning for SummaryDurationSeconds: 90, not
	// game.DefaultSummaryDurationSeconds (60, also the restore fallback for
	// a zero value) - a dropped field would silently decode back to the
	// default and this test would never notice.
	cfg, err := game.NewConfig(enums.Gauntlet, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo}, enums.Random, game.MaxGauntletPlayers, false, enums.Private, 30, game.PoolFilter{}, enums.Swift, 90)
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
	parentPower, err := powers.NewPower(powers.PowerID{50}, "Star Platinum", "desc", enums.Legendary, &skills, "")
	if err != nil {
		t.Fatalf("NewPower(parent): %v", err)
	}
	parent, err := powers.NewStand(*parentPower, enums.A, enums.A, enums.A, enums.A, enums.A, enums.A, nil)
	if err != nil {
		t.Fatalf("NewStand(parent): %v", err)
	}
	childPower, err := powers.NewPower(powers.PowerID{51}, "Star Platinum: The World", "desc", enums.Legendary, &skills, "")
	if err != nil {
		t.Fatalf("NewPower(child): %v", err)
	}
	stand, err := powers.NewStand(*childPower, enums.Infinite, enums.Infinite, enums.Infinite, enums.Infinite, enums.Infinite, enums.Infinite, parent)
	if err != nil {
		t.Fatalf("NewStand(child): %v", err)
	}

	loadout, err := game.NewLoadout(stand, nil, enums.SpinInfinite, enums.HamonPerfect, enums.FruitMasteryNone,
		enums.HakiPrivate, enums.HakiPrivate, enums.HakiPrivate, enums.PhysicalFormPrivate)
	if err != nil {
		t.Fatalf("NewLoadout: %v", err)
	}
	host.AssignLoadout(loadout)

	return g
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	g := buildTestGame(t)
	now := time.Now()

	payload, err := encode(g, now)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	restored, err := decode(payload)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	if restored.ID() != g.ID() {
		t.Errorf("ID mismatch: got %s want %s", restored.ID(), g.ID())
	}
	host, ok := restored.Participant(game.ParticipantID{1})
	if !ok {
		t.Fatal("host missing after decode")
	}
	loadout := host.Loadout()
	if loadout == nil || loadout.Stand() == nil {
		t.Fatal("stand lost across decode")
	}
	if loadout.Stand().Name() != "Star Platinum: The World" {
		t.Errorf("stand name mismatch: got %q", loadout.Stand().Name())
	}
	if loadout.Stand().Rarity() != enums.Legendary {
		t.Errorf("stand rarity mismatch: got %v", loadout.Stand().Rarity())
	}
	parent := loadout.Stand().EvolvesFrom()
	if parent == nil || parent.Name() != "Star Platinum" {
		t.Fatalf("EvolvesFrom chain lost across decode: got %+v", parent)
	}
	if host.AvatarThumbKey() != "avatars/host/thumb.webp" {
		t.Errorf("avatar thumb key lost across decode: got %q", host.AvatarThumbKey())
	}
	if host.GooglePicture() != "https://accounts.google.com/host.jpg" {
		t.Errorf("google picture lost across decode: got %q", host.GooglePicture())
	}
	if restored.Config().RevealSpeed() != enums.Swift {
		t.Errorf("revealSpeed lost across decode: got %v, want %v", restored.Config().RevealSpeed(), enums.Swift)
	}
	if restored.Config().SummaryDurationSeconds() != 90 {
		t.Errorf("summaryDurationSeconds lost across decode: got %v, want 90", restored.Config().SummaryDurationSeconds())
	}
}

// TestDecodeToleratesMissingAvatarFields confirms a payload written before
// avatarThumbKey/googlePicture existed still decodes cleanly - the
// additive-omitempty-field rule documented on wire.go's snapshotVersion,
// exercised here rather than just asserted in a comment.
func TestDecodeToleratesMissingAvatarFields(t *testing.T) {
	g := buildTestGame(t)
	payload, err := encode(g, time.Now())
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	var env map[string]any
	if err := json.Unmarshal(payload, &env); err != nil {
		t.Fatalf("unmarshal into map: %v", err)
	}
	gameObj, ok := env["game"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected envelope shape, no game object: %+v", env)
	}
	participants, ok := gameObj["participants"].([]any)
	if !ok || len(participants) == 0 {
		t.Fatalf("unexpected game shape, no participants: %+v", gameObj)
	}
	for _, p := range participants {
		pm, ok := p.(map[string]any)
		if !ok {
			continue
		}
		delete(pm, "avatarThumbKey")
		delete(pm, "googlePicture")
	}
	legacy, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}

	restored, err := decode(legacy)
	if err != nil {
		t.Fatalf("decode legacy (avatar-less) payload: %v", err)
	}
	host, ok := restored.Participant(game.ParticipantID{1})
	if !ok {
		t.Fatal("host missing after decode")
	}
	if host.AvatarThumbKey() != "" || host.GooglePicture() != "" {
		t.Errorf("expected empty avatar fields on a legacy payload, got thumb=%q google=%q",
			host.AvatarThumbKey(), host.GooglePicture())
	}
}

func TestDecodeRejectsUnknownVersion(t *testing.T) {
	g := buildTestGame(t)
	payload, err := encode(g, time.Now())
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	var env map[string]any
	if err := json.Unmarshal(payload, &env); err != nil {
		t.Fatalf("unmarshal into map: %v", err)
	}
	env["v"] = 999
	bumped, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}

	if _, err := decode(bumped); err == nil {
		t.Fatal("expected error decoding an unknown snapshot version")
	}
}

func TestDecodeRejectsTruncatedJSON(t *testing.T) {
	g := buildTestGame(t)
	payload, err := encode(g, time.Now())
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	truncated := payload[:len(payload)/2]
	if _, err := decode(truncated); err == nil {
		t.Fatal("expected error decoding truncated JSON")
	}
}
