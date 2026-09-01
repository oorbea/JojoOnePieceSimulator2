package services_test

// Test coverage for §4's power-pool hardening pass (see
// ObsidianVault/game-lobby-todo.md §4): checkPoolSufficiency and beginRound's
// per-round PoolFilter.Apply were audited as already mode-agnostic and
// covering both power kinds - this file closes the pure test-coverage gap,
// no production code changes. Assertions never pin which power a
// participant landed (the loadout builder has a deliberate "no power"
// weight, see game.LoadoutBuilder) - they only assert that no participant,
// in any round, ever holds a banned power id.

import (
	"context"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// assertNoBannedPower fails t if any participant in g currently holds a
// Loadout referencing one of bannedIDs.
func assertNoBannedPower(t *testing.T, g *game.Game, bannedIDs ...powers.PowerID) {
	t.Helper()
	banned := make(map[powers.PowerID]struct{}, len(bannedIDs))
	for _, id := range bannedIDs {
		banned[id] = struct{}{}
	}
	for _, p := range g.Participants() {
		loadout := p.Loadout()
		if loadout == nil {
			continue
		}
		if s := loadout.Stand(); s != nil {
			if _, ok := banned[s.ID()]; ok {
				t.Fatalf("participant %s was assigned banned Stand %q", p.ID(), s.Name())
			}
		}
		if d := loadout.DevilFruit(); d != nil {
			if _, ok := banned[d.ID()]; ok {
				t.Fatalf("participant %s was assigned banned DevilFruit %q", p.ID(), d.Name())
			}
		}
	}
}

func TestPoolFilter_Gauntlet_BannedStandNeverAssigned(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	bannedStandID := deps.powers.stands[0].ID()
	poolFilter, err := game.NewPoolFilter(nil, nil, nil, []powers.PowerID{bannedStandID})
	if err != nil {
		t.Fatalf("NewPoolFilter: %v", err)
	}

	input := gauntletInput()
	input.PoolFilter = poolFilter

	g, _, err := svc.CreateGame(context.Background(), hostID, input)
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.StartGame(context.Background(), g.ID(), g.HostID())
	if err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	// Gauntlet assigns Loadouts synchronously inside beginRound, before the
	// reveal-delay timer even starts, so the assignment can be checked
	// straight off StartGame's returned Game.
	assertNoBannedPower(t, g, bannedStandID)

	advanceReveal(deps, input.PowerMangas)
	advanceSummary(deps)
	g, err = svc.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	assertNoBannedPower(t, g, bannedStandID)
}

func TestPoolFilter_Versus_BannedPowersNeverAssignedAcrossRounds(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	bannedStandID := deps.powers.stands[0].ID()
	bannedFruitID := deps.powers.fruits[0].ID()
	poolFilter, err := game.NewPoolFilter(nil, nil, nil, []powers.PowerID{bannedStandID, bannedFruitID})
	if err != nil {
		t.Fatalf("NewPoolFilter: %v", err)
	}

	input := versusInput(1)
	input.PoolFilter = poolFilter

	g, code, err := svc.CreateGame(context.Background(), hostID, input)
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.JoinByCode(context.Background(), code, joinerID)
	if err != nil {
		t.Fatalf("JoinByCode: %v", err)
	}
	g, err = svc.StartGame(context.Background(), g.ID(), g.HostID())
	if err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	assertNoBannedPower(t, g, bannedStandID, bannedFruitID)

	advanceReveal(deps, input.PowerMangas)
	advanceSummary(deps)
	g, err = svc.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame after reveal: %v", err)
	}

	teamA := g.Teams()[0].ID()
	optionA := game.OptionID(teamA.String())

	var joinerParticipant game.ParticipantID
	for _, p := range g.Participants() {
		if p.ID() != g.HostID() {
			joinerParticipant = p.ID()
		}
	}

	// Versus reassigns Loadouts every round (VersusMode.ReassignsEachRound)
	// - checking at the top of every iteration proves beginRound's per-
	// round PoolFilter.Apply, not just the first assignment.
	for round := 0; round < game.VersusRounds; round++ {
		assertNoBannedPower(t, g, bannedStandID, bannedFruitID)

		g, err = svc.CastVote(context.Background(), g.ID(), g.HostID(), optionA)
		if err != nil {
			t.Fatalf("round %d host vote: %v", round, err)
		}
		g, err = svc.CastVote(context.Background(), g.ID(), joinerParticipant, optionA)
		if err != nil {
			t.Fatalf("round %d joiner vote: %v", round, err)
		}
		advanceResult(deps)

		if round < game.VersusRounds-1 {
			advanceReveal(deps, input.PowerMangas)
			advanceSummary(deps)
			g, err = svc.GetGame(context.Background(), g.ID())
			if err != nil {
				t.Fatalf("GetGame after round %d reveal: %v", round, err)
			}
		}
	}

	if g.State() != enums.Finished {
		t.Fatalf("state = %v, want FINISHED", g.State())
	}
}

func TestPoolFilter_Gauntlet_EveryStandBanned_StartGameFails(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	banned := make([]powers.PowerID, 0, len(deps.powers.stands))
	for _, s := range deps.powers.stands {
		banned = append(banned, s.ID())
	}
	poolFilter, err := game.NewPoolFilter(nil, nil, nil, banned)
	if err != nil {
		t.Fatalf("NewPoolFilter: %v", err)
	}

	input := gauntletInput()
	input.PoolFilter = poolFilter

	g, _, err := svc.CreateGame(context.Background(), hostID, input)
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if _, err := svc.StartGame(context.Background(), g.ID(), g.HostID()); err != game.ErrPoolTooSmall {
		t.Fatalf("StartGame err = %v, want ErrPoolTooSmall", err)
	}

	// A too-small filtered pool must leave the Game untouched in LOBBY,
	// never stranded in ASSIGNING with no way back.
	g, err = svc.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	if g.State() != enums.Lobby {
		t.Fatalf("state = %v, want LOBBY", g.State())
	}
}

func TestPoolFilter_Versus_EveryDevilFruitBanned_StartGameFails(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	banned := make([]powers.PowerID, 0, len(deps.powers.fruits))
	for _, d := range deps.powers.fruits {
		banned = append(banned, d.ID())
	}
	poolFilter, err := game.NewPoolFilter(nil, nil, nil, banned)
	if err != nil {
		t.Fatalf("NewPoolFilter: %v", err)
	}

	// Both power mangas enabled (versusInput's default) proves the
	// second-kind (DevilFruit) exhaustion path specifically, distinct from
	// the Stand exhaustion path TestPoolFilter_Gauntlet_EveryStandBanned_
	// StartGameFails already covers.
	input := versusInput(1)
	input.PoolFilter = poolFilter

	g, code, err := svc.CreateGame(context.Background(), hostID, input)
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.JoinByCode(context.Background(), code, joinerID)
	if err != nil {
		t.Fatalf("JoinByCode: %v", err)
	}
	if _, err := svc.StartGame(context.Background(), g.ID(), g.HostID()); err != game.ErrPoolTooSmall {
		t.Fatalf("StartGame err = %v, want ErrPoolTooSmall", err)
	}
}

func TestPoolFilter_JojoOnly_EveryDevilFruitBanned_StartGameSucceeds(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	banned := make([]powers.PowerID, 0, len(deps.powers.fruits))
	for _, d := range deps.powers.fruits {
		banned = append(banned, d.ID())
	}
	poolFilter, err := game.NewPoolFilter(nil, nil, nil, banned)
	if err != nil {
		t.Fatalf("NewPoolFilter: %v", err)
	}

	input := gauntletInput()
	// gauntletInput already selects only Jojo for both StageMangas and
	// PowerMangas - explicit here so the intent ("fruit exhaustion is
	// irrelevant when OnePiece isn't a selected power manga") reads clearly
	// regardless of gauntletInput's own defaults ever changing.
	input.PowerMangas = []enums.Manga{enums.Jojo}
	input.PoolFilter = poolFilter

	g, _, err := svc.CreateGame(context.Background(), hostID, input)
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if _, err := svc.StartGame(context.Background(), g.ID(), g.HostID()); err != nil {
		t.Fatalf("StartGame: %v, want success", err)
	}
}

func TestPoolFilter_ChecksActualOccupancy_NotConfiguredTeamSize(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	// gauntletInput's TeamSize is 5, but this fake catalog only ever has 2
	// Stands - if StartGame checked the configured capacity rather than the
	// 2 participants actually seated, it would wrongly reject this.
	input := gauntletInput()
	if len(deps.powers.stands) != 2 {
		t.Fatalf("test fixture assumption broken: want exactly 2 fake Stands, got %d", len(deps.powers.stands))
	}

	g, code, err := svc.CreateGame(context.Background(), hostID, input)
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.JoinByCode(context.Background(), code, joinerID)
	if err != nil {
		t.Fatalf("JoinByCode: %v", err)
	}
	if len(g.Teams()[0].Members()) != 2 {
		t.Fatalf("seated members = %d, want 2", len(g.Teams()[0].Members()))
	}

	if _, err := svc.StartGame(context.Background(), g.ID(), g.HostID()); err != nil {
		t.Fatalf("StartGame: %v, want success (2 seats <= 2 available Stands, even though TeamSize=%d)", err, input.TeamSize)
	}
}

func TestPoolFilter_EditLobbyConfig_BanlistHonoredOnStart(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	// Created with an empty PoolFilter (nothing restricted).
	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}

	bannedStandID := deps.powers.stands[0].ID()
	poolFilter, err := game.NewPoolFilter(nil, nil, nil, []powers.PowerID{bannedStandID})
	if err != nil {
		t.Fatalf("NewPoolFilter: %v", err)
	}

	// EditLobbyConfig is the path the frontend actually uses for a config
	// update (PATCH /{id}/config) - ConfigUpdateInput is CreateGameInput's
	// alias, so this mirrors gauntletInput with only PoolFilter changed.
	update := gauntletInput()
	update.PoolFilter = poolFilter
	g, err = svc.EditLobbyConfig(context.Background(), g.ID(), g.HostID(), update)
	if err != nil {
		t.Fatalf("EditLobbyConfig: %v", err)
	}

	g, err = svc.StartGame(context.Background(), g.ID(), g.HostID())
	if err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	assertNoBannedPower(t, g, bannedStandID)
}
