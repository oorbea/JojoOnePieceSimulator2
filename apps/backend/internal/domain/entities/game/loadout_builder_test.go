package game_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func TestLoadoutBuilder_JojoOnlyDrawsNoOnePieceAbilities(t *testing.T) {
	stand := mustStand(t, 1, "Star Platinum", enums.Legendary)
	pool := game.NewAvailablePowers([]*powers.Stand{stand}, nil)
	weights := game.DefaultAssignmentWeights()
	builder := game.NewLoadoutBuilder([]enums.Manga{enums.Jojo}, weights, &fakeRandom{seq: []int{1}})

	loadout, err := builder.Build(pool)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if loadout.DevilFruit() != nil || loadout.FruitMastery() != enums.FruitMasteryNone {
		t.Fatalf("expected no one piece abilities, got fruit=%v mastery=%v", loadout.DevilFruit(), loadout.FruitMastery())
	}
	if loadout.ArmamentHaki() != enums.HakiNone || loadout.ObservationHaki() != enums.HakiNone ||
		loadout.ConquerorHaki() != enums.HakiNone || loadout.PhysicalForm() != enums.PhysicalFormPrivate {
		t.Fatalf("expected zero-value one piece stats, got %+v", loadout)
	}
}

func TestLoadoutBuilder_OnePieceOnlyDrawsNoJojoAbilities(t *testing.T) {
	weights := game.DefaultAssignmentWeights()
	builder := game.NewLoadoutBuilder([]enums.Manga{enums.OnePiece}, weights, &fakeRandom{})
	pool := game.NewAvailablePowers(nil, nil)

	loadout, err := builder.Build(pool)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if loadout.Stand() != nil || loadout.Spin() != enums.SpinNone || loadout.Hamon() != enums.HamonNone {
		t.Fatalf("expected no jojo abilities, got %+v", loadout)
	}
}

func TestLoadoutBuilder_MixedAssignsBothMangas(t *testing.T) {
	stand := mustStand(t, 1, "Star Platinum", enums.Legendary)
	fruit := mustDevilFruit(t, 2, "Gomu Gomu no Mi", enums.Legendary, enums.Paramecia)
	pool := game.NewAvailablePowers([]*powers.Stand{stand}, []*powers.DevilFruit{fruit})
	weights := game.DefaultAssignmentWeights()
	builder := game.NewLoadoutBuilder(enums.Mangas(), weights, &fakeRandom{seq: []int{1}})

	loadout, err := builder.Build(pool)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if loadout.Stand() == nil || loadout.DevilFruit() == nil {
		t.Fatalf("expected both a stand and a devil fruit, got %+v", loadout)
	}
}

func TestLoadoutBuilder_FruitMasteryFollowsFruitPresence(t *testing.T) {
	fruit := mustDevilFruit(t, 1, "Gomu Gomu no Mi", enums.Legendary, enums.Paramecia)
	weights := game.DefaultAssignmentWeights()
	// index 0 is "no fruit"; index 1 is the only seeded fruit.
	builder := game.NewLoadoutBuilder([]enums.Manga{enums.OnePiece}, weights, &fakeRandom{seq: []int{1}})
	pool := game.NewAvailablePowers(nil, []*powers.DevilFruit{fruit})

	loadout, err := builder.Build(pool)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if loadout.DevilFruit() == nil {
		t.Fatalf("expected a devil fruit to be drawn")
	}
	if loadout.FruitMastery() == enums.FruitMasteryNone {
		t.Fatalf("expected fruit mastery >= REGULAR once a fruit is drawn, got NONE")
	}
}

func TestLoadoutBuilder_NoFruitForcesNoMastery(t *testing.T) {
	weights := game.DefaultAssignmentWeights()
	builder := game.NewLoadoutBuilder([]enums.Manga{enums.OnePiece}, weights, &fakeRandom{seq: []int{0}})
	pool := game.NewAvailablePowers(nil, nil)

	loadout, err := builder.Build(pool)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if loadout.DevilFruit() != nil || loadout.FruitMastery() != enums.FruitMasteryNone {
		t.Fatalf("expected no fruit and NONE mastery, got fruit=%v mastery=%v", loadout.DevilFruit(), loadout.FruitMastery())
	}
}

// recordingRandom is a game.RandomSource that always returns the top of the
// requested range (n-1) while recording the `n` passed to each IntN call, in
// call order. Since every draw step uses a distinct pool size or
// level-table length, the recorded sequence of `n`s is a fingerprint of the
// exact step order LoadoutBuilder.Build executed - this is what makes
// TestLoadoutBuilder_DrawOrder able to fail on an accidental reordering even
// though every other test here is order-agnostic (they use fakeRandom with a
// constant value applied regardless of which step is calling). Returning
// n-1 (rather than 0) matters here specifically: 0 would land every
// stand/devilFruit draw on their "none drawn" sentinel weight (always the
// first bucket), which short-circuits FruitMastery's draw entirely
// (loadout_builder.go's drawFruitMastery returns early when devilFruit is
// nil) and would silently drop a step from the recorded sequence.
type recordingRandom struct{ ns []int }

func (r *recordingRandom) IntN(n int) int {
	r.ns = append(r.ns, n)
	if n <= 0 {
		return 0
	}
	return n - 1
}

// TestLoadoutBuilder_DrawOrder pins the owner-mandated step order: Physical
// Form -> Stand -> Devil Fruit -> Fruit Mastery -> Hamon -> Haki Set ->
// Haki Mastery (once per haki the set draw landed on) -> Spin
// (RequiresSpin4 is a post-pass, not a draw, so it never shows up here).
// Pool sizes (2 stands, 3 fruits) are chosen so stand's IntN and
// devilFruit's IntN can't be confused with each other.
//
// n is the *total* weight weightedPick sums to for that draw, not the raw
// option count. recordingRandom always returns n-1, i.e. the *last* bucket
// of whatever table is being drawn from - for HakiSetWeights (weights.go)
// the last bucket is HakiSetConqueror, so exactly one HakiMastery draw
// follows (Armament/Observation are never drawn here), not three.
func TestLoadoutBuilder_DrawOrder(t *testing.T) {
	standA := mustStand(t, 1, "Star Platinum", enums.Legendary)
	standB := mustStand(t, 2, "The World", enums.Legendary)
	fruitA := mustDevilFruit(t, 1, "Gomu Gomu no Mi", enums.Legendary, enums.Paramecia)
	fruitB := mustDevilFruit(t, 2, "Mera Mera no Mi", enums.Legendary, enums.Paramecia)
	fruitC := mustDevilFruit(t, 3, "Yami Yami no Mi", enums.Legendary, enums.Paramecia)
	pool := game.NewAvailablePowers(
		[]*powers.Stand{standA, standB},
		[]*powers.DevilFruit{fruitA, fruitB, fruitC},
	)
	weights := game.DefaultAssignmentWeights()
	rng := &recordingRandom{}
	builder := game.NewLoadoutBuilder(enums.Mangas(), weights, rng)

	if _, err := builder.Build(pool); err != nil {
		t.Fatalf("Build: %v", err)
	}

	// PhysicalForm(6 levels, total 6) -> Stand(no-stand + 2, total 3) ->
	// DevilFruit(no-fruit + 3, total 4) -> FruitMastery(3 levels, total 3)
	// -> Hamon(total 25+35+35+5=100) -> HakiSet(total
	// 4+20+20+20+15+10+10+1=100, lands on Conqueror-only) ->
	// ConquerorMastery(4 levels, total 4) -> Spin(total 15+30+30+25=100).
	want := []int{6, 3, 4, 3, 100, 100, 4, 100}
	if len(rng.ns) != len(want) {
		t.Fatalf("expected %d draws, got %d: %v", len(want), len(rng.ns), rng.ns)
	}
	for i, n := range want {
		if rng.ns[i] != n {
			t.Fatalf("draw order mismatch at step %d: want IntN(%d), got IntN(%d) (full sequence %v)", i, n, rng.ns[i], rng.ns)
		}
	}
}

func TestLoadoutBuilder_DoesNotRepeatStandWithinTeam(t *testing.T) {
	standA := mustStand(t, 1, "Star Platinum", enums.Legendary)
	standB := mustStand(t, 2, "The World", enums.Legendary)
	pool := game.NewAvailablePowers([]*powers.Stand{standA, standB}, nil)
	weights := game.DefaultAssignmentWeights()
	builder := game.NewLoadoutBuilder([]enums.Manga{enums.Jojo}, weights, &fakeRandom{seq: []int{1}})

	l1, err := builder.Build(pool)
	if err != nil {
		t.Fatalf("Build (1st): %v", err)
	}
	l2, err := builder.Build(pool)
	if err != nil {
		t.Fatalf("Build (2nd): %v", err)
	}
	if l1.Stand() == nil || l2.Stand() == nil {
		t.Fatalf("expected both draws to produce a stand")
	}
	if l1.Stand().ID() == l2.Stand().ID() {
		t.Fatalf("expected distinct stands within the team, got the same one twice: %v", l1.Stand().ID())
	}
}
