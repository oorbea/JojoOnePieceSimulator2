package game_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// mustEvolvedStand builds a Stand whose EvolvesFrom points at a distinct
// parent Stand, so the round trip test can assert the chain survives.
func mustEvolvedStand(t *testing.T, id byte, name string, rarity enums.PowerRarity, parent *powers.Stand) *powers.Stand {
	t.Helper()
	skills := []string{"skill"}
	power, err := powers.NewPower(powers.PowerID{id}, name, "description", rarity, &skills, "")
	if err != nil {
		t.Fatalf("NewPower(%q): %v", name, err)
	}
	stand, err := powers.NewStand(*power, enums.A, enums.A, enums.A, enums.A, enums.A, enums.A, parent)
	if err != nil {
		t.Fatalf("NewStand(%q): %v", name, err)
	}
	return stand
}

// buildMidMatchVersusGame builds a deliberately gnarly Versus Game: 2
// teams, one bot, host already reassigned after a disconnect, loadouts
// assigned (one with a Stand+EvolvesFrom parent, one with a DevilFruit), one
// already-resolved round, and a second round left mid-TIEBREAK.
func buildMidMatchVersusGame(t *testing.T) *game.Game {
	t.Helper()

	cfg, err := game.NewConfig(enums.Versus, []enums.Manga{enums.Jojo, enums.OnePiece}, []enums.Manga{enums.Jojo, enums.OnePiece}, enums.Random, 2, true, enums.Public, 45, game.PoolFilter{})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	host := mustHumanParticipant(t, 1, 1, 10)
	host.SetAvatar("avatars/host/thumb.webp", "https://accounts.google.com/host.jpg")
	teamA := mustTeam(t, 10, "Team A")
	teamB := mustTeam(t, 20, "Team B")
	g, err := game.NewGame(game.GameID{9}, cfg, host, []*game.Team{teamA, teamB}, someStages(t))
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}

	second := mustHumanParticipant(t, 2, 2, 10)
	if err := g.Join(second); err != nil {
		t.Fatalf("Join(second): %v", err)
	}
	third := mustHumanParticipant(t, 3, 3, 20)
	if err := g.Join(third); err != nil {
		t.Fatalf("Join(third): %v", err)
	}
	bot := mustBotParticipant(t, 4, 20)
	if err := g.AddBot(bot); err != nil {
		t.Fatalf("AddBot: %v", err)
	}

	// Host disconnects -> reassigns to `second` (only other connected
	// human on Team A).
	if err := g.Disconnect(host.ID(), &fakeRandom{}); err != nil {
		t.Fatalf("Disconnect(host): %v", err)
	}
	if g.HostID() != second.ID() {
		t.Fatalf("host not reassigned: got %s want %s", g.HostID(), second.ID())
	}

	parent := mustStand(t, 50, "Star Platinum", enums.Legendary)
	stand := mustEvolvedStand(t, 51, "Star Platinum: The World", enums.Legendary, parent)
	fruit := mustDevilFruit(t, 60, "Gomu Gomu no Mi", enums.Rare, enums.Paramecia)

	if err := g.Start(g.HostID()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	// seq{1} biases weightedPick away from index 0 ("no stand"/"no devil
	// fruit") so the assigned loadouts actually carry a Stand/DevilFruit to
	// round-trip.
	builder := game.NewLoadoutBuilder(cfg.PowerMangas(), game.DefaultAssignmentWeights(), &fakeRandom{seq: []int{1}})

	// Round 0: assign (Team A draws the evolved Stand, Team B the Devil
	// Fruit), vote with a clear winner, resolve.
	pools := map[game.TeamID]*game.AvailablePowers{
		teamA.ID(): game.NewAvailablePowers([]*powers.Stand{stand}, nil),
		teamB.ID(): game.NewAvailablePowers(nil, []*powers.DevilFruit{fruit}),
	}
	if err := g.AssignLoadouts(builder, pools); err != nil {
		t.Fatalf("AssignLoadouts (round 0): %v", err)
	}
	if err := g.OpenVoting(&fakeRandom{}); err != nil {
		t.Fatalf("OpenVoting (round 0): %v", err)
	}
	// Cast every participant's vote explicitly (CastVote does not check
	// Connected, and this overrides the bot's automatic vote too), so the
	// outcome is deterministic regardless of loadout-driven bot scoring:
	// 3-1 for Team A, a clean non-tied resolution.
	teamAOpt := game.OptionID(teamA.ID().String())
	teamBOpt0 := game.OptionID(teamB.ID().String())
	if err := g.CastVote(host.ID(), teamAOpt); err != nil {
		t.Fatalf("CastVote(host, r0): %v", err)
	}
	if err := g.CastVote(second.ID(), teamAOpt); err != nil {
		t.Fatalf("CastVote(second): %v", err)
	}
	if err := g.CastVote(third.ID(), teamAOpt); err != nil {
		t.Fatalf("CastVote(third): %v", err)
	}
	if err := g.CastVote(bot.ID(), teamBOpt0); err != nil {
		t.Fatalf("CastVote(bot, r0): %v", err)
	}
	if tied, err := g.CloseVoting(); err != nil || tied {
		t.Fatalf("CloseVoting (round 0): tied=%v err=%v", tied, err)
	}
	if err := g.CompleteRound(); err != nil {
		t.Fatalf("CompleteRound (round 0): %v", err)
	}

	// Round 1: assign again (Versus reassigns each round), tie the vote so
	// TiebreakUsed flips and the round is left genuinely mid-vote.
	pools = map[game.TeamID]*game.AvailablePowers{
		teamA.ID(): game.NewAvailablePowers(nil, nil),
		teamB.ID(): game.NewAvailablePowers(nil, nil),
	}
	if err := g.AssignLoadouts(builder, pools); err != nil {
		t.Fatalf("AssignLoadouts (round 1): %v", err)
	}
	if err := g.OpenVoting(&fakeRandom{}); err != nil {
		t.Fatalf("OpenVoting (round 1): %v", err)
	}
	// 2-2, forced regardless of automatic bot scoring, same reasoning as
	// round 0.
	teamBOpt := game.OptionID(teamB.ID().String())
	if err := g.CastVote(host.ID(), teamAOpt); err != nil {
		t.Fatalf("CastVote(host, r1): %v", err)
	}
	if err := g.CastVote(second.ID(), teamBOpt); err != nil {
		t.Fatalf("CastVote(second, r1): %v", err)
	}
	if err := g.CastVote(third.ID(), teamAOpt); err != nil {
		t.Fatalf("CastVote(third, r1): %v", err)
	}
	if err := g.CastVote(bot.ID(), teamBOpt); err != nil {
		t.Fatalf("CastVote(bot, r1): %v", err)
	}
	if tied, err := g.CloseVoting(); err != nil || !tied {
		t.Fatalf("CloseVoting (round 1) expected tie: tied=%v err=%v", tied, err)
	}

	return g
}

func TestSnapshotRestoreRoundTrip(t *testing.T) {
	g := buildMidMatchVersusGame(t)
	before := g.Snapshot()

	restored, err := game.Restore(before)
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}

	if restored.ID() != g.ID() {
		t.Errorf("ID mismatch: got %s want %s", restored.ID(), g.ID())
	}
	if restored.State() != g.State() {
		t.Errorf("State mismatch: got %v want %v", restored.State(), g.State())
	}
	if restored.HostID() != g.HostID() {
		t.Errorf("HostID mismatch: got %s want %s", restored.HostID(), g.HostID())
	}
	if len(restored.Participants()) != len(g.Participants()) {
		t.Fatalf("participant count mismatch: got %d want %d", len(restored.Participants()), len(g.Participants()))
	}
	for _, want := range g.Participants() {
		got, ok := restored.Participant(want.ID())
		if !ok {
			t.Fatalf("participant %s missing after restore", want.ID())
		}
		if got.Connected() != want.Connected() || got.Kind() != want.Kind() || got.TeamID() != want.TeamID() {
			t.Errorf("participant %s mismatch: got %+v want %+v", want.ID(), got, want)
		}
		if got.AvatarThumbKey() != want.AvatarThumbKey() || got.GooglePicture() != want.GooglePicture() {
			t.Errorf("participant %s avatar mismatch: got thumb=%q google=%q want thumb=%q google=%q",
				want.ID(), got.AvatarThumbKey(), got.GooglePicture(), want.AvatarThumbKey(), want.GooglePicture())
		}
		wl, gl := want.Loadout(), got.Loadout()
		if (wl == nil) != (gl == nil) {
			t.Fatalf("participant %s loadout nilness mismatch", want.ID())
		}
		if wl != nil {
			if (wl.Stand() == nil) != (gl.Stand() == nil) {
				t.Errorf("participant %s stand nilness mismatch", want.ID())
			}
			if wl.Stand() != nil && wl.Stand().Name() != gl.Stand().Name() {
				t.Errorf("participant %s stand mismatch: got %s want %s", want.ID(), gl.Stand().Name(), wl.Stand().Name())
			}
			if wl.Stand() != nil && wl.Stand().EvolvesFrom() != nil {
				if gl.Stand().EvolvesFrom() == nil || gl.Stand().EvolvesFrom().Name() != wl.Stand().EvolvesFrom().Name() {
					t.Errorf("participant %s stand EvolvesFrom lost across restore", want.ID())
				}
			}
			if (wl.DevilFruit() == nil) != (gl.DevilFruit() == nil) {
				t.Errorf("participant %s devil fruit nilness mismatch", want.ID())
			}
			if wl.FruitMastery() != gl.FruitMastery() || wl.Spin() != gl.Spin() {
				t.Errorf("participant %s ability levels mismatch", want.ID())
			}
		}
	}

	if len(restored.Rounds()) != len(g.Rounds()) {
		t.Fatalf("round count mismatch: got %d want %d", len(restored.Rounds()), len(g.Rounds()))
	}
	wantRounds, gotRounds := g.Rounds(), restored.Rounds()
	for i := range wantRounds {
		wr, gr := wantRounds[i], gotRounds[i]
		if wr.TiebreakUsed != gr.TiebreakUsed {
			t.Errorf("round %d TiebreakUsed mismatch: got %v want %v", i, gr.TiebreakUsed, wr.TiebreakUsed)
		}
		if (wr.Result == nil) != (gr.Result == nil) {
			t.Fatalf("round %d Result nilness mismatch", i)
		}
		if wr.Result != nil && (wr.Result.Winner != gr.Result.Winner || wr.Result.DecidedByCoinFlip != gr.Result.DecidedByCoinFlip) {
			t.Errorf("round %d Result mismatch: got %+v want %+v", i, gr.Result, wr.Result)
		}
		if gr.Ballot.Count() != wr.Ballot.Count() {
			t.Errorf("round %d ballot vote count mismatch: got %d want %d", i, gr.Ballot.Count(), wr.Ballot.Count())
		}
		wantVotes, gotVotes := wr.Ballot.Votes(), gr.Ballot.Votes()
		for pid, opt := range wantVotes {
			if gotVotes[pid] != opt {
				t.Errorf("round %d vote for %s mismatch: got %s want %s", i, pid, gotVotes[pid], opt)
			}
		}
		if len(wr.TiedVotes) != len(gr.TiedVotes) {
			t.Fatalf("round %d TiedVotes count mismatch: got %d want %d", i, len(gr.TiedVotes), len(wr.TiedVotes))
		}
		for pid, opt := range wr.TiedVotes {
			if gr.TiedVotes[pid] != opt {
				t.Errorf("round %d tied vote for %s mismatch: got %s want %s", i, pid, gr.TiedVotes[pid], opt)
			}
		}
	}

	// The restored game must still behave, not just hold data: tallying
	// the restored TIEBREAK round must reflect the restored votes, and
	// closing it again must not panic on a nil evaluator.
	if got := restored.State(); got != enums.Tiebreak {
		t.Fatalf("restored game not in TIEBREAK as expected: %v", got)
	}
	if restored.VotingComplete() != g.VotingComplete() {
		t.Errorf("VotingComplete mismatch after restore")
	}
	if _, err := restored.CloseVoting(); err != nil {
		t.Fatalf("CloseVoting on restored game: %v", err)
	}
}

func TestRestoreRejectsUnknownEnums(t *testing.T) {
	g := buildMidMatchVersusGame(t)
	base := g.Snapshot()

	t.Run("unknown state", func(t *testing.T) {
		s := base
		s.State = "SUPERPOSITION"
		if _, err := game.Restore(s); err == nil {
			t.Fatal("expected error for unknown state")
		}
	})

	t.Run("unknown stage manga", func(t *testing.T) {
		s := base
		s.Config.StageMangas = []string{"NARUTO"}
		if _, err := game.Restore(s); err == nil {
			t.Fatal("expected error for unknown stage manga")
		}
	})

	t.Run("unknown power manga", func(t *testing.T) {
		s := base
		s.Config.PowerMangas = []string{"NARUTO"}
		if _, err := game.Restore(s); err == nil {
			t.Fatal("expected error for unknown power manga")
		}
	})

	t.Run("unknown legacy manga fallback", func(t *testing.T) {
		s := base
		s.Config.StageMangas = nil
		s.Config.PowerMangas = nil
		s.Config.Mangas = []string{"NARUTO"}
		if _, err := game.Restore(s); err == nil {
			t.Fatal("expected error for unknown legacy manga")
		}
	})

	t.Run("unknown mode", func(t *testing.T) {
		s := base
		s.Config.Mode = "ROYALE"
		if _, err := game.Restore(s); err == nil {
			t.Fatal("expected error for unknown mode")
		}
	})
}

func TestRestoreRejectsMalformedBallot(t *testing.T) {
	g := buildMidMatchVersusGame(t)
	s := g.Snapshot()
	last := len(s.Rounds) - 1
	s.Rounds[last].Ballot.Options = s.Rounds[last].Ballot.Options[:1]
	if _, err := game.Restore(s); err == nil {
		t.Fatal("expected error restoring a ballot with fewer than 2 options")
	}
}

func TestRestoreOfZeroSnapshot(t *testing.T) {
	if _, err := game.Restore(game.Snapshot{}); err == nil {
		t.Fatal("expected error restoring a zero Snapshot (invalid mode/state/manga)")
	}
}

func TestSnapshotDoesNotDrainEvents(t *testing.T) {
	g, _ := newGauntletGame(t, oneStage(t), 1)
	if err := g.Start(g.HostID()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	_ = g.Snapshot()
	if len(g.PullEvents()) == 0 {
		t.Fatal("Snapshot must not drain PullEvents")
	}
}

// TestSnapshotRoundTrip_SplitMangaAxes pins that StageMangas/PowerMangas
// round-trip independently through Snapshot/Restore - a Snapshot never
// populates the legacy Mangas field, only the split ones.
func TestSnapshotRoundTrip_SplitMangaAxes(t *testing.T) {
	cfg, err := game.NewConfig(enums.Gauntlet, []enums.Manga{enums.Jojo, enums.OnePiece}, []enums.Manga{enums.Jojo}, enums.Random, 1, false, enums.Private, 30, game.PoolFilter{})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	host := mustHumanParticipant(t, 1, 1, 10)
	team := mustTeam(t, 10, "Squad")
	g, err := game.NewGame(game.GameID{1}, cfg, host, []*game.Team{team}, oneStage(t))
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}

	snap := g.Snapshot()
	if len(snap.Config.Mangas) != 0 {
		t.Fatalf("expected Snapshot to leave the legacy Mangas field empty, got %v", snap.Config.Mangas)
	}
	if got, want := snap.Config.StageMangas, []string{"JOJO", "ONE_PIECE"}; !equalStrings(got, want) {
		t.Fatalf("StageMangas = %v, want %v", got, want)
	}
	if got, want := snap.Config.PowerMangas, []string{"JOJO"}; !equalStrings(got, want) {
		t.Fatalf("PowerMangas = %v, want %v", got, want)
	}

	restored, err := game.Restore(snap)
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if got, want := restored.Config().StageMangas(), cfg.StageMangas(); !equalMangas(got, want) {
		t.Fatalf("restored StageMangas = %v, want %v", got, want)
	}
	if got, want := restored.Config().PowerMangas(), cfg.PowerMangas(); !equalMangas(got, want) {
		t.Fatalf("restored PowerMangas = %v, want %v", got, want)
	}
}

// TestRestore_LegacyMangasFallback pins the compatibility path for a
// Snapshot written before the StageMangas/PowerMangas split (e.g. a lobby
// still live in Redis across a deploy) - both axes fall back to the
// legacy Mangas field.
func TestRestore_LegacyMangasFallback(t *testing.T) {
	cfg, err := game.NewConfig(enums.Gauntlet, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo}, enums.Random, 1, false, enums.Private, 30, game.PoolFilter{})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	host := mustHumanParticipant(t, 1, 1, 10)
	team := mustTeam(t, 10, "Squad")
	g, err := game.NewGame(game.GameID{1}, cfg, host, []*game.Team{team}, oneStage(t))
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	snap := g.Snapshot()
	// Simulate a pre-split payload: only the legacy field is set.
	snap.Config.StageMangas = nil
	snap.Config.PowerMangas = nil
	snap.Config.Mangas = []string{"ONE_PIECE"}

	restored, err := game.Restore(snap)
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	want := []enums.Manga{enums.OnePiece}
	if got := restored.Config().StageMangas(); !equalMangas(got, want) {
		t.Fatalf("restored StageMangas = %v, want %v (legacy fallback)", got, want)
	}
	if got := restored.Config().PowerMangas(); !equalMangas(got, want) {
		t.Fatalf("restored PowerMangas = %v, want %v (legacy fallback)", got, want)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalMangas(a, b []enums.Manga) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
