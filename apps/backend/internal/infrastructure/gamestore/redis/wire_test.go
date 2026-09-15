package redis

import (
	"encoding/json"
	"strings"
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
	host.SetAvatar("avatars/host/thumb.webp", "https://accounts.google.com/host.jpg", 0.5, 0.5)
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

	// BattleIQ=0 deliberately, not some other value: 0 is the value most at
	// risk of being conflated with "absent" by an accidental omitempty on
	// a non-pointer field - see wireLoadout.BattleIQ's doc comment.
	loadout, err := game.NewLoadoutFromSpec(game.LoadoutSpec{
		Stand: stand, Spin: enums.SpinInfinite, Hamon: enums.HamonPerfect, FruitMastery: enums.FruitMasteryNone,
		ArmamentHaki: enums.HakiPrivate, ObservationHaki: enums.HakiPrivate, ConquerorHaki: enums.HakiPrivate,
		PhysicalForm: enums.PhysicalFormPrivate,
		BattleIQ:     game.NewBattleIQ(0),
	})
	if err != nil {
		t.Fatalf("NewLoadoutFromSpec: %v", err)
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
	// Regression guard for the exact TiedVotes-class bug (see
	// TestEncodeDecodeRoundTrip_TiedVotes's doc comment): battleIQ must
	// survive the wire round trip, and a present 0 must come back
	// present, not absent.
	if !loadout.BattleIQ().Present() {
		t.Fatalf("battleIQ lost across decode: got absent, want present with value 0")
	}
	if loadout.BattleIQ().Value() != 0 {
		t.Errorf("battleIQ value mismatch: got %d, want 0", loadout.BattleIQ().Value())
	}
}

// TestEncodeDecodeRoundTrip_TiedVotes guards the exact class of bug
// TestEncodeDecodeRoundTrip's RevealSpeed/SummaryDurationSeconds checks
// already guard for wireConfig: wireRound.TiedVotes was added after
// wireRound already existed and was missed, so every real Redis Save
// encoded a tied round's TiedVotes fine but the very next Get silently
// decoded it back to nil - Save persisted no error, decode raised no
// error, only the field was gone (a 2026-09-02 live two-browser
// walkthrough found the TIEBREAK vote-breakdown panel never rendering;
// hitting GET /api/v1/games/:id mid-TIEBREAK showed tiebreakUsed=true but
// no tiedVotes key at all). game.Snapshot()/Restore() alone (what
// game_response_test.go's TestNewGameStateResponse_TiedVotes and
// game_service_test.go's TestCastVote_Tie_DTOStateCarriesTiedVotes both
// exercise) never touches this package's separate wireRound type, so
// neither test could have caught it - only a round trip through this
// package's own encode/decode does.
func TestEncodeDecodeRoundTrip_TiedVotes(t *testing.T) {
	cfg, err := game.NewConfig(enums.Gauntlet, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo}, enums.Random, game.MaxGauntletPlayers, false, enums.Private, 30, game.PoolFilter{}, enums.Normal, game.DefaultSummaryDurationSeconds)
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
	builder := game.NewLoadoutBuilder(cfg.PowerMangas(), game.DefaultAssignmentWeights(), zeroWireRandom{})
	if err := g.AssignLoadouts(builder, pools); err != nil {
		t.Fatalf("AssignLoadouts: %v", err)
	}
	if err := g.OpenVoting(zeroWireRandom{}); err != nil {
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

	payload, err := encode(g, time.Now())
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	restored, err := decode(payload)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	rounds := restored.Rounds()
	if len(rounds) != 1 {
		t.Fatalf("rounds = %d, want 1", len(rounds))
	}
	round := rounds[0]
	if !round.TiebreakUsed {
		t.Fatal("TiebreakUsed lost across the wire round trip")
	}
	if len(round.TiedVotes) != 2 {
		t.Fatalf("TiedVotes = %+v, want the 2 votes that tied (lost across the wire round trip)", round.TiedVotes)
	}
	if round.TiedVotes[host.ID()] != "SURVIVE" {
		t.Errorf("TiedVotes[host] = %q, want SURVIVE", round.TiedVotes[host.ID()])
	}
	if round.TiedVotes[second.ID()] != "FALL" {
		t.Errorf("TiedVotes[second] = %q, want FALL", round.TiedVotes[second.ID()])
	}
}

// zeroWireRandom is a deterministic game.RandomSource - always 0.
type zeroWireRandom struct{}

func (zeroWireRandom) IntN(n int) int {
	if n <= 0 {
		return 0
	}
	return 0
}

var _ game.RandomSource = zeroWireRandom{}

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

// TestEncodeDecodeRoundTrip_PhaseEndsAt covers the durable phase deadline
// across the wire encoding, plus the legacy payload that predates the field.
func TestEncodeDecodeRoundTrip_PhaseEndsAt(t *testing.T) {
	g := buildTestGame(t)
	deadline := time.Now().Add(37 * time.Second).UTC().Truncate(time.Second)
	g.SetPhaseDeadline(deadline)

	payload, err := encode(g, time.Now())
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if !strings.Contains(string(payload), "phaseEndsAt") {
		t.Fatalf("encoded payload carries no phaseEndsAt: %s", payload)
	}

	restored, err := decode(payload)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	got, ok := restored.PhaseEndsAt()
	if !ok {
		t.Fatal("phase deadline lost across the wire encoding")
	}
	if !got.Equal(deadline) {
		t.Fatalf("PhaseEndsAt = %v, want %v", got, deadline)
	}

	// A payload written before this field existed must still decode cleanly
	// (the additive-omitempty rule on snapshotVersion), leaving no deadline.
	var env map[string]any
	if err := json.Unmarshal(payload, &env); err != nil {
		t.Fatalf("unmarshal into map: %v", err)
	}
	gameObj, ok := env["game"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected envelope shape, no game object: %+v", env)
	}
	delete(gameObj, "phaseEndsAt")
	legacy, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}
	legacyGame, err := decode(legacy)
	if err != nil {
		t.Fatalf("decode legacy (phaseEndsAt-less) payload: %v", err)
	}
	if _, ok := legacyGame.PhaseEndsAt(); ok {
		t.Fatal("a legacy payload must decode to no phase deadline")
	}
}

// TestEncodeDecodeRoundTrip_RevealReadySummaryReady guards the exact bug
// this pair of fields was added to fix (2026-09-14, live-tested): the
// sorteo/summary "saltar" skip vote (Game.revealReady/summaryReady) lived
// only on the in-memory *Game, with no field on game.Snapshot or wireGame.
// game.Snapshot()/Restore() alone never caught it (there was no field to
// drop), and MemoryGameStore never caught it either (it hands back the same
// live pointer, no round trip) - only encode/decode through THIS package
// exercises the bug: with REDIS_URL set, every withGame call does a real
// Get→Restore, so the vote a human had already cast was silently gone on
// the very next command, and RevealReadyComplete/SummaryReadyComplete could
// never reach "everyone's in" - the skip button did nothing until the full
// configured phase timer expired regardless of how many players pressed it.
func TestEncodeDecodeRoundTrip_RevealReadySummaryReady(t *testing.T) {
	cfg, err := game.NewConfig(enums.Gauntlet, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo}, enums.Random, game.MaxGauntletPlayers, false, enums.Private, 30, game.PoolFilter{}, enums.Normal, game.DefaultSummaryDurationSeconds)
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
	// Only the host has voted to skip the reveal - a real mid-vote snapshot,
	// not the all-voted case (which would collapse straight to
	// RevealReadyComplete()==true even if the count itself were lost).
	if err := g.MarkRevealReady(host.ID()); err != nil {
		t.Fatalf("MarkRevealReady: %v", err)
	}

	payload, err := encode(g, time.Now())
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	restored, err := decode(payload)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	ready, total := restored.RevealReadyProgress()
	if total != 2 {
		t.Fatalf("RevealReadyProgress total = %d, want 2", total)
	}
	if ready != 1 {
		t.Fatalf("RevealReadyProgress ready = %d, want 1 (host's vote lost across the wire round trip)", ready)
	}
	if restored.RevealReadyComplete() {
		t.Fatal("RevealReadyComplete = true, want false (only 1 of 2 humans ready)")
	}

	// Now the second human's vote completes it - proves the restored vote
	// isn't just decorative but actually participates in
	// RevealReadyComplete's count.
	if err := restored.MarkRevealReady(second.ID()); err != nil {
		t.Fatalf("MarkRevealReady on restored game: %v", err)
	}
	if !restored.RevealReadyComplete() {
		t.Fatal("RevealReadyComplete = false after both humans ready, want true")
	}

	// summaryReady mirrors revealReady exactly - same field, same bug class,
	// covered by advancing the restored game into SUMMARY and voting there.
	if err := restored.OpenSummary(); err != nil {
		t.Fatalf("OpenSummary: %v", err)
	}
	if err := restored.MarkSummaryReady(host.ID()); err != nil {
		t.Fatalf("MarkSummaryReady: %v", err)
	}

	payload2, err := encode(restored, time.Now())
	if err != nil {
		t.Fatalf("encode (summary): %v", err)
	}
	restored2, err := decode(payload2)
	if err != nil {
		t.Fatalf("decode (summary): %v", err)
	}
	sReady, sTotal := restored2.SummaryReadyProgress()
	if sTotal != 2 || sReady != 1 {
		t.Fatalf("SummaryReadyProgress = %d/%d, want 1/2 (summaryReady lost across the wire round trip)", sReady, sTotal)
	}
	if restored2.SummaryReadyComplete() {
		t.Fatal("SummaryReadyComplete = true, want false (only 1 of 2 humans ready)")
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

// TestEncodeDecodeRoundTrip_DisconnectedAbandonedAndFocalPoint guards the
// same class of bug TiedVotes/RevealReady already bit twice: a field that
// exists on game.ParticipantSnapshot but was never wired into
// wireParticipant silently resets on every Redis round trip while
// MemoryGameStore (which keeps the live pointer) never notices. Two fields
// are covered here at once because they share that exact failure mode:
//   - AvatarFocalX/AvatarFocalY were missed when wireParticipant was first
//     written (see its doc comment) - a non-0.5 value (buildTestGame's own
//     host defaults to 0.5/0.5, which would pass even with the bug) proves
//     the fix.
//   - DisconnectedAt/Abandoned are the new grace-period fields added
//     alongside this test.
func TestEncodeDecodeRoundTrip_DisconnectedAbandonedAndFocalPoint(t *testing.T) {
	g := buildTestGame(t)
	host, ok := g.Participant(game.ParticipantID{1})
	if !ok {
		t.Fatal("host missing before encode")
	}
	host.SetAvatar(host.AvatarThumbKey(), host.GooglePicture(), 0.2, 0.8)

	disconnectedAt := time.Now().Add(-90 * time.Second).UTC().Truncate(time.Second)
	if err := g.Disconnect(host.ID(), zeroWireRandom{}, disconnectedAt); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if err := g.Abandon(host.ID()); err != nil {
		t.Fatalf("Abandon: %v", err)
	}

	payload, err := encode(g, time.Now())
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	restored, err := decode(payload)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	got, ok := restored.Participant(game.ParticipantID{1})
	if !ok {
		t.Fatal("host missing after decode")
	}
	if got.AvatarFocalX() != 0.2 || got.AvatarFocalY() != 0.8 {
		t.Errorf("focal point = (%v, %v), want (0.2, 0.8) - lost across the wire round trip",
			got.AvatarFocalX(), got.AvatarFocalY())
	}
	if !got.Abandoned() {
		t.Error("expected Abandoned() == true after the wire round trip")
	}
	gotAt, ok := got.DisconnectedAt()
	if !ok {
		t.Fatal("expected DisconnectedAt to survive the wire round trip")
	}
	if !gotAt.Equal(disconnectedAt) {
		t.Errorf("DisconnectedAt = %v, want %v", gotAt, disconnectedAt)
	}

	// A payload written before these fields existed must still decode
	// cleanly to the safe default: connected, not abandoned.
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
		delete(pm, "avatarFocalX")
		delete(pm, "avatarFocalY")
		delete(pm, "disconnectedAt")
		delete(pm, "abandoned")
	}
	legacy, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}
	legacyGame, err := decode(legacy)
	if err != nil {
		t.Fatalf("decode legacy (grace-field-less) payload: %v", err)
	}
	legacyHost, ok := legacyGame.Participant(game.ParticipantID{1})
	if !ok {
		t.Fatal("host missing after decoding legacy payload")
	}
	if legacyHost.Abandoned() {
		t.Error("a legacy payload must decode to not-abandoned")
	}
	if _, ok := legacyHost.DisconnectedAt(); ok {
		t.Error("a legacy payload must decode to no disconnectedAt")
	}
}
