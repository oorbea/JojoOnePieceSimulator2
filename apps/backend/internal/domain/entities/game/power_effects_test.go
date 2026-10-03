package game_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// --- fixtures -------------------------------------------------------------

// tuskFamily is prod's chain: Acto 1 -> 2 -> 3 -> 4 (ids 1..4).
func tuskFamily(t *testing.T) []*powers.Stand {
	t.Helper()
	a1 := mustEvolvedStand(t, 1, "Tusk: Acto 1", enums.Epic, nil)
	a2 := mustEvolvedStand(t, 2, "Tusk: Acto 2", enums.Epic, a1)
	a3 := mustEvolvedStand(t, 3, "Tusk: Acto 3", enums.Epic, a2)
	a4 := mustEvolvedStand(t, 4, "Tusk: Acto 4", enums.Epic, a3)
	return []*powers.Stand{a1, a2, a3, a4}
}

// pinned forces every weighted draw to one value, so a test controls exactly
// which stats the resolver starts from. Stand/fruit draws are forced through
// scriptRandom instead (they are uniform over the pool).
type pinned struct {
	spin    enums.SpinLevel
	hamon   enums.HamonLevel
	mastery enums.FruitMastery
	form    enums.PhysicalForm
}

func pinnedWeights(p pinned) game.AssignmentWeights {
	w := game.DefaultAssignmentWeights()
	w.NoStandWeight, w.NoDevilFruitWeight = 0, 0
	w.SpinLevelWeights = map[enums.SpinLevel]int{p.spin: 1}
	w.HamonLevelWeights = map[enums.HamonLevel]int{p.hamon: 1}
	mastery := p.mastery
	if mastery == enums.FruitMasteryNone {
		mastery = enums.FruitMasteryRegular
	}
	w.FruitMasteryWeights = map[enums.FruitMastery]int{mastery: 1}
	w.HakiSetWeights = map[game.HakiSet]int{game.HakiSetNone: 1}
	w.PhysicalFormWeights = map[enums.PhysicalForm]int{p.form: 1}
	return w
}

// scriptRandom answers IntN(n) with at[n] (0 when unset) and counts calls per
// n. Pool draws are IntN(poolSize) (NoStandWeight is 0) and an evolution
// tie-break is IntN(2), so a test picks "which stand" with at[len(pool)] and
// "which tied target" with at[2]. Pinned draws are IntN(1) and so always 0.
type scriptRandom struct {
	at    map[int]int
	calls map[int]int
}

func newScript(at map[int]int) *scriptRandom {
	return &scriptRandom{at: at, calls: map[int]int{}}
}

func (r *scriptRandom) IntN(n int) int {
	r.calls[n]++
	if n <= 0 {
		return 0
	}
	return r.at[n] % n
}

func names(stands []*powers.Stand, fruits []*powers.DevilFruit) map[string]string {
	m := map[string]string{}
	for _, s := range stands {
		m[s.ID().String()] = s.Name()
	}
	for _, f := range fruits {
		m[f.ID().String()] = f.Name()
	}
	return m
}

func build(t *testing.T, mangas []enums.Manga, w game.AssignmentWeights, rng game.RandomSource,
	stands []*powers.Stand, fruits []*powers.DevilFruit) *game.Loadout {
	t.Helper()
	pool := game.NewAvailablePowers(stands, game.LinkFruitEvolutions(fruits))
	l, err := game.NewLoadoutBuilder(mangas, w, rng).Build(pool)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return l
}

func describe(effects []game.PowerEffect, nm map[string]string) []string {
	out := make([]string, len(effects))
	for i, e := range effects {
		from, to := e.From, e.To
		if e.Kind == enums.EffectEvolution {
			from, to = nm[from], nm[to]
		}
		out[i] = fmt.Sprintf("%s %s %s>%s by %s:%s", e.Kind, e.Slot, from, to, e.CauseSlot, e.Cause)
	}
	return out
}

func assertEffects(t *testing.T, l *game.Loadout, nm map[string]string, want ...string) {
	t.Helper()
	got := describe(l.Effects(), nm)
	if len(got) == 0 && len(want) == 0 {
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("effects:\n got  %q\n want %q", got, want)
	}
}

var (
	jojo     = []enums.Manga{enums.Jojo}
	onePiece = []enums.Manga{enums.OnePiece}
	both     = []enums.Manga{enums.Jojo, enums.OnePiece}
)

// --- stand spin tiers -----------------------------------------------------

func TestPowerEffects_StandSpinTierRaisesSpin(t *testing.T) {
	tests := []struct {
		drawn    int
		wantSpin enums.SpinLevel
	}{
		{0, enums.SpinBasic},    // Acto 1
		{1, enums.SpinGolden},   // Acto 2
		{2, enums.SpinGolden},   // Acto 3
		{3, enums.SpinInfinite}, // Acto 4
	}
	for _, tc := range tests {
		stands := tuskFamily(t)
		l := build(t, jojo, pinnedWeights(pinned{spin: enums.SpinNone}), newScript(map[int]int{4: tc.drawn}), stands, nil)
		if l.Stand() != stands[tc.drawn] {
			t.Fatalf("drawn %d: expected %s, got %s", tc.drawn, stands[tc.drawn].Name(), l.Stand().Name())
		}
		if l.Spin() != tc.wantSpin {
			t.Fatalf("%s: expected spin %v, got %v", stands[tc.drawn].Name(), tc.wantSpin, l.Spin())
		}
		assertEffects(t, l, names(stands, nil),
			fmt.Sprintf("STAT_FLOOR SPIN NONE>%s by STAND:%s", tc.wantSpin, stands[tc.drawn].Name()))
	}
}

func TestPowerEffects_BallBreakerAndGoBeyondForceInfinite(t *testing.T) {
	for _, name := range []string{"Ball Breaker", "Soft & Wet: Go Beyond", "  BALL   breaker "} {
		stand := mustStand(t, 9, name, enums.Legendary)
		l := build(t, jojo, pinnedWeights(pinned{spin: enums.SpinBasic}), newScript(nil), []*powers.Stand{stand}, nil)
		if l.Spin() != enums.SpinInfinite {
			t.Fatalf("%q: expected INFINITE, got %v", name, l.Spin())
		}
	}
}

func TestPowerEffects_LegacyTuskNameDoesNotMatch(t *testing.T) {
	// The old table said "Tusk ACT4"; prod says "Tusk: Acto 4". Matching is
	// exact, so the old spelling must not match.
	stand := mustStand(t, 9, "Tusk ACT4", enums.Legendary)
	l := build(t, jojo, pinnedWeights(pinned{spin: enums.SpinNone}), newScript(nil), []*powers.Stand{stand}, nil)
	if l.Spin() != enums.SpinNone {
		t.Fatalf("expected no effect for an unknown name, got spin %v", l.Spin())
	}
}

// --- spin-driven stand evolution -----------------------------------------

func TestPowerEffects_SpinEvolvesStand(t *testing.T) {
	tests := []struct {
		name      string
		drawn     int
		spin      enums.SpinLevel
		tie       int
		wantStand string
		wantTies  int
		wantFx    []string
	}{
		{"acto1 + infinite -> acto4", 0, enums.SpinInfinite, 0, "Tusk: Acto 4", 0,
			[]string{"EVOLUTION STAND Tusk: Acto 1>Tusk: Acto 4 by SPIN:INFINITE"}},
		{"acto2 + infinite -> acto4", 1, enums.SpinInfinite, 0, "Tusk: Acto 4", 0,
			[]string{"EVOLUTION STAND Tusk: Acto 2>Tusk: Acto 4 by SPIN:INFINITE"}},
		{"acto3 + infinite -> acto4", 2, enums.SpinInfinite, 0, "Tusk: Acto 4", 0,
			[]string{"EVOLUTION STAND Tusk: Acto 3>Tusk: Acto 4 by SPIN:INFINITE"}},
		{"acto1 + golden tie -> acto2", 0, enums.SpinGolden, 0, "Tusk: Acto 2", 1,
			[]string{"EVOLUTION STAND Tusk: Acto 1>Tusk: Acto 2 by SPIN:GOLDEN"}},
		{"acto1 + golden tie -> acto3", 0, enums.SpinGolden, 1, "Tusk: Acto 3", 1,
			[]string{"EVOLUTION STAND Tusk: Acto 1>Tusk: Acto 3 by SPIN:GOLDEN"}},
		{"acto1 + basic stays", 0, enums.SpinBasic, 0, "Tusk: Acto 1", 0, nil},
		{"acto2 + golden stays (V1)", 1, enums.SpinGolden, 0, "Tusk: Acto 2", 0, nil},
		{"acto3 + golden stays (V1)", 2, enums.SpinGolden, 0, "Tusk: Acto 3", 0, nil},
		{"acto4 + infinite stays", 3, enums.SpinInfinite, 0, "Tusk: Acto 4", 0, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stands := tuskFamily(t)
			rng := newScript(map[int]int{4: tc.drawn, 2: tc.tie})
			l := build(t, jojo, pinnedWeights(pinned{spin: tc.spin}), rng, stands, nil)
			if l.Stand().Name() != tc.wantStand {
				t.Fatalf("expected %s, got %s", tc.wantStand, l.Stand().Name())
			}
			if l.Spin() != tc.spin {
				t.Fatalf("spin must be untouched, got %v", l.Spin())
			}
			if rng.calls[2] != tc.wantTies {
				t.Fatalf("expected %d tie-break draw(s), got %d", tc.wantTies, rng.calls[2])
			}
			assertEffects(t, l, names(stands, nil), tc.wantFx...)
			if l.DrawnStand() != stands[tc.drawn] {
				t.Fatalf("DrawnStand should be %s, got %s", stands[tc.drawn].Name(), l.DrawnStand().Name())
			}
		})
	}
}

func TestPowerEffects_EvolutionRespectsBannedStages(t *testing.T) {
	// Acto 4 banned: the pool only holds Acto 1..3.
	// The Spin outgrows every allowed stage, so it is the furthest-evolved one
	// that wins - in one hop and without a tie-break draw, whichever of the
	// equally ranked Acto 2 / Acto 3 would otherwise have been picked first.
	tests := []struct {
		name  string
		drawn int
		want  string
		hops  int
	}{
		{"acto2 + infinite -> acto3", 1, "Tusk: Acto 3", 1},
		{"acto1 + infinite -> acto3 in one hop", 0, "Tusk: Acto 3", 1},
		{"acto3 + infinite stays, nothing allowed above", 2, "Tusk: Acto 3", 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stands := tuskFamily(t)[:3]
			rng := newScript(map[int]int{3: tc.drawn, 2: 1})
			l := build(t, jojo, pinnedWeights(pinned{spin: enums.SpinInfinite}), rng, stands, nil)
			if l.Stand().Name() != tc.want {
				t.Fatalf("expected %s, got %s", tc.want, l.Stand().Name())
			}
			if len(l.Effects()) != tc.hops || rng.calls[2] != 0 {
				t.Fatalf("expected %d effect(s) and no tie-break, got %d / %d draws", tc.hops, len(l.Effects()), rng.calls[2])
			}
		})
	}
}

func TestPowerEffects_SoftAndWetEvolvesWithInfiniteSpin(t *testing.T) {
	sw := mustEvolvedStand(t, 10, "Soft & Wet", enums.Epic, nil)
	gb := mustEvolvedStand(t, 11, "Soft & Wet: Go Beyond", enums.Epic, sw)

	l := build(t, jojo, pinnedWeights(pinned{spin: enums.SpinInfinite}), newScript(map[int]int{2: 0}), []*powers.Stand{sw, gb}, nil)
	assertEffects(t, l, names([]*powers.Stand{sw, gb}, nil),
		"EVOLUTION STAND Soft & Wet>Soft & Wet: Go Beyond by SPIN:INFINITE")

	l = build(t, jojo, pinnedWeights(pinned{spin: enums.SpinGolden}), newScript(map[int]int{2: 0}), []*powers.Stand{sw, gb}, nil)
	if l.Stand() != sw {
		t.Fatalf("Soft & Wet must not evolve below INFINITE, got %s", l.Stand().Name())
	}

	// Drawing Go Beyond directly raises the spin instead.
	l = build(t, jojo, pinnedWeights(pinned{spin: enums.SpinNone}), newScript(map[int]int{2: 1}), []*powers.Stand{sw, gb}, nil)
	if l.Stand() != gb || l.Spin() != enums.SpinInfinite {
		t.Fatalf("expected Go Beyond with INFINITE spin, got %s/%v", l.Stand().Name(), l.Spin())
	}
}

func TestPowerEffects_UntieredFamilyNeverEvolves(t *testing.T) {
	e0 := mustEvolvedStand(t, 20, "Echoes Act 0", enums.Epic, nil)
	e1 := mustEvolvedStand(t, 21, "Echoes Act 1", enums.Epic, e0)
	l := build(t, jojo, pinnedWeights(pinned{spin: enums.SpinInfinite}), newScript(map[int]int{2: 0}), []*powers.Stand{e0, e1}, nil)
	if l.Stand() != e0 || len(l.Effects()) != 0 {
		t.Fatalf("expected Echoes Act 0 untouched, got %s with %d effects", l.Stand().Name(), len(l.Effects()))
	}
}

// --- fruit family: Gomu -> Nika ------------------------------------------

func gomuAndNika(t *testing.T) (*powers.DevilFruit, *powers.DevilFruit) {
	t.Helper()
	return mustDevilFruit(t, 30, "Gomu Gomu no mi", enums.Epic, enums.Paramecia),
		mustDevilFruit(t, 31, "Hito Hito no mi: Model Nika", enums.Legendary, enums.MythicalZoan)
}

func TestPowerEffects_AwakenedGomuBecomesNika(t *testing.T) {
	gomu, nika := gomuAndNika(t)
	fruits := []*powers.DevilFruit{gomu, nika}
	nm := names(nil, fruits)

	l := build(t, onePiece, pinnedWeights(pinned{mastery: enums.FruitMasteryAwakened, form: enums.PhysicalFormPrivate}),
		newScript(map[int]int{2: 0}), nil, fruits)
	if l.DevilFruit().ID() != nika.ID() || l.DrawnDevilFruit().ID() != gomu.ID() {
		t.Fatalf("expected Gomu drawn and Nika final, got drawn=%s final=%s", l.DrawnDevilFruit().Name(), l.DevilFruit().Name())
	}
	// Nika is a Mythical Zoan, so rule #2 also lifts the physical form.
	assertEffects(t, l, nm,
		"EVOLUTION DEVIL_FRUIT Gomu Gomu no mi>Hito Hito no mi: Model Nika by FRUIT_MASTERY:AWAKENED",
		"STAT_FLOOR PHYSICAL_FORM PRIVATE>MARINE_CAPTAIN by DEVIL_FRUIT:Hito Hito no mi: Model Nika",
	)

	l = build(t, onePiece, pinnedWeights(pinned{mastery: enums.FruitMasteryAdvanced}), newScript(map[int]int{2: 0}), nil, fruits)
	if l.DevilFruit().ID() != gomu.ID() || len(l.Effects()) != 0 {
		t.Fatalf("Gomu must not evolve below AWAKENED, got %s", l.DevilFruit().Name())
	}
}

func TestPowerEffects_NikaDrawnDirectlyForcesAwakened(t *testing.T) {
	gomu, nika := gomuAndNika(t)
	l := build(t, onePiece, pinnedWeights(pinned{mastery: enums.FruitMasteryRegular, form: enums.PhysicalFormMarineCaptain}),
		newScript(map[int]int{2: 1}), nil, []*powers.DevilFruit{gomu, nika})
	if l.DevilFruit().ID() != nika.ID() || l.FruitMastery() != enums.FruitMasteryAwakened {
		t.Fatalf("expected Nika with AWAKENED, got %s/%v", l.DevilFruit().Name(), l.FruitMastery())
	}
	assertEffects(t, l, names(nil, []*powers.DevilFruit{gomu, nika}),
		"STAT_FLOOR FRUIT_MASTERY REGULAR>AWAKENED by DEVIL_FRUIT:Hito Hito no mi: Model Nika")
}

func TestPowerEffects_BannedNikaKeepsGomu(t *testing.T) {
	gomu, _ := gomuAndNika(t)
	l := build(t, onePiece, pinnedWeights(pinned{mastery: enums.FruitMasteryAwakened}), newScript(nil), nil, []*powers.DevilFruit{gomu})
	if l.DevilFruit().ID() != gomu.ID() || len(l.Effects()) != 0 {
		t.Fatalf("expected Gomu untouched without Nika in the pool, got %s", l.DevilFruit().Name())
	}
}

func TestPowerEffects_ChoppersHitoHitoIsNotNika(t *testing.T) {
	chopper := mustDevilFruit(t, 40, "Hito Hito no mi", enums.Common, enums.Zoan)
	l := build(t, both, pinnedWeights(pinned{hamon: enums.HamonNone, form: enums.PhysicalFormPrivate}), newScript(nil), nil, []*powers.DevilFruit{chopper})
	if l.FruitMastery() == enums.FruitMasteryAwakened || l.Hamon() != enums.HamonNone {
		t.Fatalf("Chopper's fruit must not trigger Nika's rules, got mastery=%v hamon=%v", l.FruitMastery(), l.Hamon())
	}
	// It is a plain Zoan, though: rule #3.
	if l.PhysicalForm() != enums.PhysicalFormStrongFishman {
		t.Fatalf("expected a Zoan to raise the physical form to STRONG_FISHMAN, got %v", l.PhysicalForm())
	}
}

// --- stat floors ----------------------------------------------------------

func TestPowerEffects_ZoanPhysicalFormFloors(t *testing.T) {
	tests := []struct {
		name  string
		typ   enums.FruitType
		form  enums.PhysicalForm
		want  enums.PhysicalForm
		fires bool
	}{
		{"zoan raises to strong fishman", enums.Zoan, enums.PhysicalFormPrivate, enums.PhysicalFormStrongFishman, true},
		{"zoan leaves a higher form alone", enums.Zoan, enums.PhysicalFormViceAdmiral, enums.PhysicalFormViceAdmiral, false},
		{"mythical zoan raises to marine captain", enums.MythicalZoan, enums.PhysicalFormStrongFishman, enums.PhysicalFormMarineCaptain, true},
		{"ancient zoan raises to marine captain", enums.AncientZoan, enums.PhysicalFormPrivate, enums.PhysicalFormMarineCaptain, true},
		{"paramecia never raises", enums.Paramecia, enums.PhysicalFormPrivate, enums.PhysicalFormPrivate, false},
		{"logia never raises", enums.Logia, enums.PhysicalFormPrivate, enums.PhysicalFormPrivate, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fruit := mustDevilFruit(t, 50, "Some Fruit", enums.Epic, tc.typ)
			l := build(t, onePiece, pinnedWeights(pinned{form: tc.form}), newScript(nil), nil, []*powers.DevilFruit{fruit})
			if l.PhysicalForm() != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, l.PhysicalForm())
			}
			if (len(l.Effects()) == 1) != tc.fires {
				t.Fatalf("fires=%v but got effects %v", tc.fires, l.Effects())
			}
		})
	}
}

func TestPowerEffects_HermitPurpleRaisesHamon(t *testing.T) {
	hermit := mustStand(t, 60, "Hermit purple", enums.Common)
	nm := names([]*powers.Stand{hermit}, nil)

	l := build(t, jojo, pinnedWeights(pinned{hamon: enums.HamonNone}), newScript(nil), []*powers.Stand{hermit}, nil)
	if l.Hamon() != enums.HamonBasic {
		t.Fatalf("expected BASIC, got %v", l.Hamon())
	}
	assertEffects(t, l, nm, "STAT_FLOOR HAMON NONE>BASIC by STAND:Hermit purple")

	l = build(t, jojo, pinnedWeights(pinned{hamon: enums.HamonAdvanced}), newScript(nil), []*powers.Stand{hermit}, nil)
	assertEffects(t, l, nm)
}

func TestPowerEffects_PerfectHamonRaisesSpin(t *testing.T) {
	l := build(t, jojo, pinnedWeights(pinned{hamon: enums.HamonPerfect, spin: enums.SpinNone}), newScript(nil), nil, nil)
	if l.Spin() != enums.SpinBasic {
		t.Fatalf("expected spin BASIC, got %v", l.Spin())
	}
	assertEffects(t, l, nil, "STAT_FLOOR SPIN NONE>BASIC by HAMON:PERFECT")
}

func TestPowerEffects_CrossMangaRulesNeedBothMangas(t *testing.T) {
	kc := mustStand(t, 70, "King Crimson", enums.Mythical)
	stands := []*powers.Stand{kc}

	l := build(t, both, pinnedWeights(pinned{}), newScript(nil), stands, nil)
	if l.ObservationHaki() != enums.HakiYonkoCommander {
		t.Fatalf("both mangas: King Crimson should raise observation haki, got %v", l.ObservationHaki())
	}
	assertEffects(t, l, names(stands, nil), "STAT_FLOOR OBSERVATION_HAKI NONE>YONKO_COMMANDER by STAND:King Crimson")

	l = build(t, jojo, pinnedWeights(pinned{}), newScript(nil), stands, nil)
	if l.ObservationHaki() != enums.HakiNone || len(l.Effects()) != 0 {
		t.Fatalf("JoJo only: a One Piece stat must stay untouched, got %v / %v", l.ObservationHaki(), l.Effects())
	}
}

func TestPowerEffects_HamonCrossRules(t *testing.T) {
	// Hamon PERFECT, both mangas: spin (#5), armament haki (#9), physical form (#10), in table order.
	l := build(t, both, pinnedWeights(pinned{hamon: enums.HamonPerfect, spin: enums.SpinNone, form: enums.PhysicalFormPrivate}),
		newScript(nil), nil, nil)
	assertEffects(t, l, nil,
		"STAT_FLOOR SPIN NONE>BASIC by HAMON:PERFECT",
		"STAT_FLOOR ARMAMENT_HAKI NONE>PRIVATE by HAMON:PERFECT",
		"STAT_FLOOR PHYSICAL_FORM PRIVATE>MARINE_CAPTAIN by HAMON:PERFECT",
	)

	// Hamon ADVANCED: only the physical form.
	l = build(t, both, pinnedWeights(pinned{hamon: enums.HamonAdvanced, form: enums.PhysicalFormPrivate}), newScript(nil), nil, nil)
	assertEffects(t, l, nil, "STAT_FLOOR PHYSICAL_FORM PRIVATE>MARINE_CAPTAIN by HAMON:ADVANCED")

	// Hamon BASIC: nothing.
	l = build(t, both, pinnedWeights(pinned{hamon: enums.HamonBasic}), newScript(nil), nil, nil)
	assertEffects(t, l, nil)

	// JoJo only: just the same-manga spin rule.
	l = build(t, jojo, pinnedWeights(pinned{hamon: enums.HamonPerfect, spin: enums.SpinNone}), newScript(nil), nil, nil)
	assertEffects(t, l, nil, "STAT_FLOOR SPIN NONE>BASIC by HAMON:PERFECT")
}

func TestPowerEffects_NikaCascadesIntoHamonAndPhysicalForm(t *testing.T) {
	gomu, nika := gomuAndNika(t)
	fruits := []*powers.DevilFruit{gomu, nika}
	l := build(t, both, pinnedWeights(pinned{mastery: enums.FruitMasteryAwakened, hamon: enums.HamonNone, form: enums.PhysicalFormPrivate}),
		newScript(map[int]int{2: 0}), nil, fruits)
	assertEffects(t, l, names(nil, fruits),
		"EVOLUTION DEVIL_FRUIT Gomu Gomu no mi>Hito Hito no mi: Model Nika by FRUIT_MASTERY:AWAKENED",
		"STAT_FLOOR PHYSICAL_FORM PRIVATE>MARINE_CAPTAIN by DEVIL_FRUIT:Hito Hito no mi: Model Nika",
		"STAT_FLOOR HAMON NONE>ADVANCED by DEVIL_FRUIT:Hito Hito no mi: Model Nika",
	)
	if l.Hamon() != enums.HamonAdvanced {
		t.Fatalf("expected hamon ADVANCED, got %v", l.Hamon())
	}
}

func TestPowerEffects_ResolutionIsDeterministic(t *testing.T) {
	run := func() []string {
		stands := tuskFamily(t)
		l := build(t, jojo, pinnedWeights(pinned{spin: enums.SpinGolden, hamon: enums.HamonPerfect}),
			newScript(map[int]int{4: 0, 2: 1}), stands, nil)
		return describe(l.Effects(), names(stands, nil))
	}
	if a, b := run(), run(); !reflect.DeepEqual(a, b) {
		t.Fatalf("same inputs gave different effects:\n%q\n%q", a, b)
	}
}

// --- pool, families, linking ---------------------------------------------

func TestAvailablePowers_DrawRemovesTheWholeFamily(t *testing.T) {
	ws := mustEvolvedStand(t, 80, "Whitesnake", enums.Epic, nil)
	cm := mustEvolvedStand(t, 81, "C-MOON", enums.Epic, ws)
	mih := mustEvolvedStand(t, 82, "Made in Heaven", enums.Epic, cm)
	other := mustStand(t, 83, "The Hand", enums.Common)
	catalogue := []*powers.Stand{ws, cm, mih, other}

	pool := game.NewAvailablePowers(catalogue, nil)
	drawn, family, err := pool.DrawStand(1)
	if err != nil || drawn != cm {
		t.Fatalf("expected C-MOON, got %v / %v", drawn, err)
	}
	if len(family) != 3 {
		t.Fatalf("expected the 3-stage family back as candidates, got %d", len(family))
	}
	if left := pool.Stands(); len(left) != 1 || left[0] != other {
		t.Fatalf("expected only The Hand left in the pool, got %d entries", len(left))
	}

	// The rival team's pool, built from the same catalogue, is untouched.
	if rival := game.NewAvailablePowers(catalogue, nil); len(rival.Stands()) != 4 {
		t.Fatalf("expected an independent pool of 4, got %d", len(rival.Stands()))
	}
}

func TestAvailablePowers_DrawRemovesFruitFamily(t *testing.T) {
	gomu, nika := gomuAndNika(t)
	other := mustDevilFruit(t, 90, "Mera Mera no mi", enums.Legendary, enums.Logia)
	pool := game.NewAvailablePowers(nil, game.LinkFruitEvolutions([]*powers.DevilFruit{gomu, nika, other}))

	if _, family, err := pool.DrawDevilFruit(0); err != nil || len(family) != 2 {
		t.Fatalf("expected Gomu's 2-fruit family, got %d / %v", len(family), err)
	}
	if left := pool.DevilFruits(); len(left) != 1 || left[0].ID() != other.ID() {
		t.Fatalf("expected only Mera Mera left, got %d entries", len(left))
	}
}

func TestLinkFruitEvolutions(t *testing.T) {
	gomu, nika := gomuAndNika(t)
	chopper := mustDevilFruit(t, 40, "Hito Hito no mi", enums.Common, enums.Zoan)
	linked := game.LinkFruitEvolutions([]*powers.DevilFruit{gomu, nika, chopper})

	if linked[1].EvolvesFrom() == nil || linked[1].EvolvesFrom().ID() != gomu.ID() {
		t.Fatalf("expected Nika linked to Gomu")
	}
	if linked[0].EvolvesFrom() != nil || linked[2].EvolvesFrom() != nil {
		t.Fatalf("Gomu and Chopper's fruit have no parent")
	}
	if nika.EvolvesFrom() != nil {
		t.Fatalf("the catalogue's own fruit must not be mutated")
	}
	if got := game.FruitFamilyKey(linked[1]); got != gomu.ID() {
		t.Fatalf("expected Nika's family key to be Gomu's id, got %v", got)
	}

	// A banned/absent parent leaves the child unlinked rather than failing.
	if l := game.LinkFruitEvolutions([]*powers.DevilFruit{nika}); l[0].EvolvesFrom() != nil {
		t.Fatalf("expected Nika to stay unlinked without Gomu in the list")
	}
}

func TestCountFamilies(t *testing.T) {
	stands := tuskFamily(t)
	stands = append(stands, mustStand(t, 83, "The Hand", enums.Common))
	if got := game.CountStandFamilies(stands); got != 2 {
		t.Fatalf("expected 2 stand families, got %d", got)
	}
	gomu, nika := gomuAndNika(t)
	fruits := game.LinkFruitEvolutions([]*powers.DevilFruit{gomu, nika})
	if got := game.CountFruitFamilies(fruits); got != 1 {
		t.Fatalf("expected 1 fruit family, got %d", got)
	}
}

// --- validation -----------------------------------------------------------

func TestNewLoadoutFromSpec_EnforcesPowerEffectFloors(t *testing.T) {
	_, nika := gomuAndNika(t)
	kc := mustStand(t, 70, "King Crimson", enums.Mythical)
	zoan := mustDevilFruit(t, 50, "Some Zoan", enums.Epic, enums.Zoan)

	base := func() game.LoadoutSpec {
		return game.LoadoutSpec{
			Spin: enums.SpinNone, Hamon: enums.HamonNone, FruitMastery: enums.FruitMasteryNone,
			ArmamentHaki: enums.HakiNone, ObservationHaki: enums.HakiNone, ConquerorHaki: enums.HakiNone,
			PhysicalForm: enums.PhysicalFormPrivate,
		}
	}
	tests := []struct {
		name    string
		mutate  func(*game.LoadoutSpec)
		wantErr error
	}{
		{"nika below awakened", func(s *game.LoadoutSpec) {
			s.DevilFruit, s.FruitMastery, s.PhysicalForm = nika, enums.FruitMasteryRegular, enums.PhysicalFormMarineCaptain
		}, game.ErrPowerEffectFloorViolated},
		{"zoan with private form", func(s *game.LoadoutSpec) {
			s.DevilFruit, s.FruitMastery = zoan, enums.FruitMasteryRegular
		}, game.ErrPowerEffectFloorViolated},
		{"zoan with strong fishman form", func(s *game.LoadoutSpec) {
			s.DevilFruit, s.FruitMastery, s.PhysicalForm = zoan, enums.FruitMasteryRegular, enums.PhysicalFormStrongFishman
		}, nil},
		{"perfect hamon without spin", func(s *game.LoadoutSpec) { s.Hamon = enums.HamonPerfect }, game.ErrPowerEffectFloorViolated},
		{"king crimson cross rule ignored without both mangas", func(s *game.LoadoutSpec) { s.Stand = kc }, nil},
		{"king crimson cross rule enforced with both mangas", func(s *game.LoadoutSpec) {
			s.Stand, s.Mangas = kc, []enums.Manga{enums.Jojo, enums.OnePiece}
		}, game.ErrPowerEffectFloorViolated},
		{"king crimson satisfied", func(s *game.LoadoutSpec) {
			s.Stand, s.Mangas, s.ObservationHaki = kc, []enums.Manga{enums.Jojo, enums.OnePiece}, enums.HakiYonkoCommander
		}, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spec := base()
			tc.mutate(&spec)
			if _, err := game.NewLoadoutFromSpec(spec); err != tc.wantErr {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestPowerEffectRuleNames_AreNormalized(t *testing.T) {
	names := game.PowerEffectRuleNames()
	if len(names) == 0 {
		t.Fatal("expected rule names")
	}
	for _, n := range names {
		if n != normalizeForTest(n) {
			t.Fatalf("rule name %q is not normalized (lowercase, single spaces)", n)
		}
	}
}

func normalizeForTest(s string) string {
	out := ""
	space := false
	for _, r := range s {
		if r == ' ' {
			space = true
			continue
		}
		if space && out != "" {
			out += " "
		}
		space = false
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		out += string(r)
	}
	return out
}

func TestMissingPowerEffectNames(t *testing.T) {
	all := game.PowerEffectRuleNames()
	if got := game.MissingPowerEffectNames(nil, nil); len(got) != len(all) {
		t.Fatalf("an empty catalogue lacks every rule name, got %d of %d", len(got), len(all))
	}

	hermit := mustStand(t, 60, "  HERMIT   Purple ", enums.Common) // normalization applies to the catalogue side too
	missing := game.MissingPowerEffectNames([]*powers.Stand{hermit}, nil)
	for _, n := range missing {
		if n == "hermit purple" {
			t.Fatalf("hermit purple is in the catalogue but reported missing")
		}
	}
	if len(missing) != len(all)-1 {
		t.Fatalf("expected exactly one name to drop out, got %d missing of %d", len(missing), len(all))
	}
}
