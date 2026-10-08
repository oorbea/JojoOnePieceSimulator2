package game

import (
	"errors"
	"sort"
	"strings"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

var (
	// ErrPowerEffectFloorViolated is returned when a Loadout holds a stat
	// below a floor one of its own powers (or another of its stats) demands -
	// e.g. Tusk: Act 4 with anything below SpinInfinite. The LoadoutBuilder
	// never produces one; this guards every other way of assembling a Loadout.
	ErrPowerEffectFloorViolated = errors.New("loadout violates a power effect floor")

	// ErrPowerEffectsDiverged is returned if effect resolution does not settle
	// within maxEffectPasses. Every rule only ever raises a value or moves a
	// power forward along its evolution chain, so this cannot happen with the
	// current table - it is a defensive bound, not an expected error.
	ErrPowerEffectsDiverged = errors.New("power effects did not settle")
)

const maxEffectPasses = 16

// PowerEffect records one thing a power did to a Loadout during assignment,
// in the order it happened. It exists so the sorteo can replay the drawn
// values first and animate each effect at the moment its trigger is revealed.
//
// From/To are enum wire strings for a STAT_FLOOR and PowerID strings for an
// EVOLUTION. Cause is the name of the power that triggered it, or the wire
// string of the stat value that did (e.g. "PERFECT" for a Hamon-driven rule).
type PowerEffect struct {
	Kind      enums.PowerEffectKind
	Slot      enums.LoadoutSlot
	From      string
	To        string
	CauseSlot enums.LoadoutSlot
	Cause     string
}

// Rule tables. Keyed by normalizePowerName(power.Name()) - the catalogue has
// no persisted traits column, so the rules match by name, exactly (never by
// prefix: "Hito Hito no mi" is Chopper's fruit, unrelated to Nika). Names are
// the prod catalogue's: "Tusk: Act N", not "Tusk ACT4". A rename in prod
// silently stops a rule from firing, which is what
// db/migrations/power_effect_names_test.go guards against for the seeded names.
const (
	nameHermitPurple = "hermit purple"
	nameKingCrimson  = "king crimson"
	nameGomuGomu     = "gomu gomu no mi"
	nameNika         = "hito hito no mi: model nika"
)

// standSpinTiers is the Spin level each Stand stage corresponds to. A Stand
// whose tier is above the drawn Spin raises the Spin to it; a Spin above the
// Stand's tier evolves the Stand to the highest allowed later stage whose tier
// the Spin still covers. Ported from the original JoJoOnePiece_Simulator's
// main.cc, where Tusk and Soft & Wet: Go Beyond behaved this way.
var standSpinTiers = map[string]enums.SpinLevel{
	"tusk: act 1":          enums.SpinBasic,
	"tusk: act 2":          enums.SpinGolden,
	"tusk: act 3":          enums.SpinGolden,
	"tusk: act 4":          enums.SpinInfinite,
	"ball breaker":          enums.SpinInfinite,
	"soft & wet: go beyond": enums.SpinInfinite,
}

// fruitMasteryTiers is the Fruit Mastery level a fruit stage requires.
var fruitMasteryTiers = map[string]enums.FruitMastery{
	nameNika: enums.FruitMasteryAwakened,
}

// fruitEvolvesFrom is the fruit-evolution relation (child -> parent), which
// the catalogue does not store for fruits. Applied by LinkFruitEvolutions.
var fruitEvolvesFrom = map[string]string{
	nameNika: nameGomuGomu,
}

func normalizePowerName(name string) string {
	return strings.Join(strings.Fields(strings.ToLower(name)), " ")
}

func standSpinTier(s *powers.Stand) (enums.SpinLevel, bool) {
	if s == nil {
		return enums.SpinNone, false
	}
	t, ok := standSpinTiers[normalizePowerName(s.Name())]
	return t, ok
}

func fruitMasteryTier(d *powers.DevilFruit) (enums.FruitMastery, bool) {
	if d == nil {
		return enums.FruitMasteryNone, false
	}
	t, ok := fruitMasteryTiers[normalizePowerName(d.Name())]
	return t, ok
}

// PowerEffectRuleNames lists every power name (normalized) the rule tables and
// the combat conventions refer to, for the catalogue test that checks they
// exist.
func PowerEffectRuleNames() []string {
	set := map[string]struct{}{
		nameHermitPurple: {},
		nameKingCrimson:  {},
	}
	for _, n := range ConventionPowerNames() {
		set[n] = struct{}{}
	}
	for n := range standSpinTiers {
		set[n] = struct{}{}
	}
	for n := range fruitMasteryTiers {
		set[n] = struct{}{}
	}
	for child, parent := range fruitEvolvesFrom {
		set[child] = struct{}{}
		set[parent] = struct{}{}
	}
	names := make([]string, 0, len(set))
	for n := range set {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// MissingPowerEffectNames returns the rule-table names no stand or devil fruit
// in the catalogue carries. Power effects match by exact name, so a power
// renamed in the catalogue silently stops triggering its rules - this is what
// the application logs about at startup.
func MissingPowerEffectNames(stands []*powers.Stand, fruits []*powers.DevilFruit) []string {
	present := make(map[string]struct{}, len(stands)+len(fruits))
	for _, s := range stands {
		present[normalizePowerName(s.Name())] = struct{}{}
	}
	for _, d := range fruits {
		present[normalizePowerName(d.Name())] = struct{}{}
	}
	var missing []string
	for _, n := range PowerEffectRuleNames() {
		if _, ok := present[n]; !ok {
			missing = append(missing, n)
		}
	}
	return missing
}

// StandFamilyKey identifies the evolution family a Stand belongs to: the ID of
// its root ancestor. Two Stands share a family iff they share this key.
func StandFamilyKey(s *powers.Stand) powers.PowerID {
	root := s
	seen := map[powers.PowerID]struct{}{}
	for root.EvolvesFrom() != nil {
		if _, ok := seen[root.ID()]; ok {
			break
		}
		seen[root.ID()] = struct{}{}
		root = root.EvolvesFrom()
	}
	return root.ID()
}

// FruitFamilyKey is StandFamilyKey for DevilFruits. Only meaningful on fruits
// already passed through LinkFruitEvolutions.
func FruitFamilyKey(d *powers.DevilFruit) powers.PowerID {
	root := d
	seen := map[powers.PowerID]struct{}{}
	for root.EvolvesFrom() != nil {
		if _, ok := seen[root.ID()]; ok {
			break
		}
		seen[root.ID()] = struct{}{}
		root = root.EvolvesFrom()
	}
	return root.ID()
}

// CountStandFamilies is how many distinct evolution families stands spans -
// the number of teammates the pool can actually serve, since drawing one
// stand removes its whole family from a team's pool.
func CountStandFamilies(stands []*powers.Stand) int {
	keys := make(map[powers.PowerID]struct{}, len(stands))
	for _, s := range stands {
		keys[StandFamilyKey(s)] = struct{}{}
	}
	return len(keys)
}

// CountFruitFamilies is CountStandFamilies for DevilFruits.
func CountFruitFamilies(fruits []*powers.DevilFruit) int {
	keys := make(map[powers.PowerID]struct{}, len(fruits))
	for _, d := range fruits {
		keys[FruitFamilyKey(d)] = struct{}{}
	}
	return len(keys)
}

// LinkFruitEvolutions returns fruits with each one's evolvesFrom attached per
// fruitEvolvesFrom. It must run on the full, unfiltered catalogue so a banned
// parent still links (the same way a stand's ancestors load whether or not
// they are banned); the PoolFilter is applied afterwards. Fruits are copied,
// never mutated.
func LinkFruitEvolutions(fruits []*powers.DevilFruit) []*powers.DevilFruit {
	byName := make(map[string]*powers.DevilFruit, len(fruits))
	for _, f := range fruits {
		byName[normalizePowerName(f.Name())] = f
	}
	linked := make(map[string]*powers.DevilFruit, len(fruitEvolvesFrom))
	var link func(f *powers.DevilFruit, depth int) *powers.DevilFruit
	link = func(f *powers.DevilFruit, depth int) *powers.DevilFruit {
		name := normalizePowerName(f.Name())
		parentName, ok := fruitEvolvesFrom[name]
		if !ok || depth > len(fruitEvolvesFrom) {
			return f
		}
		if l, done := linked[name]; done {
			return l
		}
		parent, found := byName[parentName]
		if !found {
			return f
		}
		l := f.WithEvolvesFrom(link(parent, depth+1))
		linked[name] = l
		return l
	}
	out := make([]*powers.DevilFruit, len(fruits))
	for i, f := range fruits {
		out[i] = link(f, 0)
	}
	return out
}

// effectState is the mutable scratch space the resolver works on: the drawn
// powers, the scalar stats (indexed by LoadoutSlot) and the family members
// still available as evolution targets.
type effectState struct {
	stand       *powers.Stand
	fruit       *powers.DevilFruit
	standFamily []*powers.Stand
	fruitFamily []*powers.DevilFruit
	values      [enums.SlotSpin + 1]int
	bothMangas  bool
}

func (st *effectState) spin() enums.SpinLevel { return enums.SpinLevel(st.values[enums.SlotSpin]) }
func (st *effectState) mastery() enums.FruitMastery {
	return enums.FruitMastery(st.values[enums.SlotFruitMastery])
}

func slotLabel(slot enums.LoadoutSlot, v int) string {
	switch slot {
	case enums.SlotPhysicalForm:
		return enums.PhysicalForm(v).String()
	case enums.SlotFruitMastery:
		return enums.FruitMastery(v).String()
	case enums.SlotHamon:
		return enums.HamonLevel(v).String()
	case enums.SlotArmamentHaki, enums.SlotObservationHaki, enums.SlotConquerorHaki:
		return enums.HakiLevel(v).String()
	case enums.SlotSpin:
		return enums.SpinLevel(v).String()
	default:
		return "UNKNOWN"
	}
}

// statFloorRule says: when trigger holds, target must be at least floor.
// cross rules (a power of one manga raising a stat of the other) only apply
// when both mangas are in play.
type statFloorRule struct {
	cross     bool
	causeSlot enums.LoadoutSlot
	target    enums.LoadoutSlot
	floor     int
	trigger   func(*effectState) (cause string, ok bool)
}

func standNamed(name string) func(*effectState) (string, bool) {
	return func(st *effectState) (string, bool) {
		if st.stand != nil && normalizePowerName(st.stand.Name()) == name {
			return st.stand.Name(), true
		}
		return "", false
	}
}

func fruitNamed(name string) func(*effectState) (string, bool) {
	return func(st *effectState) (string, bool) {
		if st.fruit != nil && normalizePowerName(st.fruit.Name()) == name {
			return st.fruit.Name(), true
		}
		return "", false
	}
}

func fruitTypeIn(types ...enums.FruitType) func(*effectState) (string, bool) {
	return func(st *effectState) (string, bool) {
		if st.fruit == nil {
			return "", false
		}
		for _, t := range types {
			if st.fruit.FruitType() == t {
				return st.fruit.Name(), true
			}
		}
		return "", false
	}
}

func slotAtLeast(slot enums.LoadoutSlot, min int) func(*effectState) (string, bool) {
	return func(st *effectState) (string, bool) {
		if st.values[slot] >= min {
			return slotLabel(slot, st.values[slot]), true
		}
		return "", false
	}
}

// statFloorRules is applied in this order, so that when two causes would both
// satisfy a floor the one revealed earliest in the sorteo gets the credit.
var statFloorRules = []statFloorRule{
	{false, enums.SlotStand, enums.SlotHamon, int(enums.HamonBasic), standNamed(nameHermitPurple)},
	{true, enums.SlotStand, enums.SlotObservationHaki, int(enums.HakiYonkoPlus), standNamed(nameKingCrimson)},
	{false, enums.SlotDevilFruit, enums.SlotPhysicalForm, int(enums.PhysicalFormMarineCaptain), fruitTypeIn(enums.MythicalZoan, enums.AncientZoan)},
	{false, enums.SlotDevilFruit, enums.SlotPhysicalForm, int(enums.PhysicalFormStrongFishman), fruitTypeIn(enums.Zoan)},
	{true, enums.SlotDevilFruit, enums.SlotHamon, int(enums.HamonAdvanced), fruitNamed(nameNika)},
	{false, enums.SlotHamon, enums.SlotSpin, int(enums.SpinBasic), slotAtLeast(enums.SlotHamon, int(enums.HamonPerfect))},
	{true, enums.SlotHamon, enums.SlotArmamentHaki, int(enums.HakiPrivate), slotAtLeast(enums.SlotHamon, int(enums.HamonPerfect))},
	{true, enums.SlotHamon, enums.SlotPhysicalForm, int(enums.PhysicalFormMarineCaptain), slotAtLeast(enums.SlotHamon, int(enums.HamonAdvanced))},
}

func (st *effectState) ruleActive(r statFloorRule) (string, bool) {
	if r.cross && !st.bothMangas {
		return "", false
	}
	return r.trigger(st)
}

func (st *effectState) applyFloor(r statFloorRule) (PowerEffect, bool) {
	cause, ok := st.ruleActive(r)
	if !ok {
		return PowerEffect{}, false
	}
	cur := st.values[r.target]
	if cur >= r.floor {
		return PowerEffect{}, false
	}
	st.values[r.target] = r.floor
	return PowerEffect{
		Kind:      enums.EffectStatFloor,
		Slot:      r.target,
		From:      slotLabel(r.target, cur),
		To:        slotLabel(r.target, r.floor),
		CauseSlot: r.causeSlot,
		Cause:     cause,
	}, true
}

// floorsViolated reports whether st breaks any floor: a stand/fruit stage
// whose tier its Spin/Mastery does not reach, or a stat rule whose target is
// below its floor.
func (st *effectState) floorsViolated() bool {
	if tier, ok := standSpinTier(st.stand); ok && st.spin() < tier {
		return true
	}
	if tier, ok := fruitMasteryTier(st.fruit); ok && st.mastery() < tier {
		return true
	}
	for _, r := range statFloorRules {
		if _, ok := st.ruleActive(r); ok && st.values[r.target] < r.floor {
			return true
		}
	}
	return false
}

// evoCandidate is one possible evolution target.
type evoCandidate struct {
	idx   int
	tier  int
	depth int
	id    string
}

// pickEvolution returns the idx of the candidate with the highest tier. When
// several candidates share that tier (Tusk: Act 1 + GOLDEN spin: Act 2 and
// Act 3 are both GOLDEN) the furthest-evolved one wins, so the outcome is
// always deterministic (owner decision 2026-10-09; V1 flipped a coin here) and
// no randomness is consumed. The same rule covers the Spin outgrowing every
// allowed stage (the ones above were banned): the Stand lands in one hop
// instead of wandering through its siblings. Candidates are sorted so the
// result never depends on catalogue order.
func pickEvolution(cands []evoCandidate) (int, bool) {
	if len(cands) == 0 {
		return 0, false
	}
	best := cands[0].tier
	for _, c := range cands[1:] {
		if c.tier > best {
			best = c.tier
		}
	}
	var top []evoCandidate
	for _, c := range cands {
		if c.tier == best {
			top = append(top, c)
		}
	}
	if len(top) == 1 {
		return top[0].idx, true
	}
	sort.Slice(top, func(i, j int) bool {
		if top[i].depth != top[j].depth {
			return top[i].depth < top[j].depth
		}
		return top[i].id < top[j].id
	})
	return top[len(top)-1].idx, true
}

func isStandDescendant(s, ancestor *powers.Stand) bool {
	seen := map[powers.PowerID]struct{}{}
	for cur := s.EvolvesFrom(); cur != nil; cur = cur.EvolvesFrom() {
		if cur.ID() == ancestor.ID() {
			return true
		}
		if _, ok := seen[cur.ID()]; ok {
			return false
		}
		seen[cur.ID()] = struct{}{}
	}
	return false
}

func isFruitDescendant(d, ancestor *powers.DevilFruit) bool {
	seen := map[powers.PowerID]struct{}{}
	for cur := d.EvolvesFrom(); cur != nil; cur = cur.EvolvesFrom() {
		if cur.ID() == ancestor.ID() {
			return true
		}
		if _, ok := seen[cur.ID()]; ok {
			return false
		}
		seen[cur.ID()] = struct{}{}
	}
	return false
}

// evolveStand moves the Stand forward along its chain once the Spin goes
// beyond the Stand's own tier (equal is "just right": Tusk: Act 2 with
// GOLDEN spin stays Act 2, as in V1), to the highest-tier descendant whose
// tier the Spin covers. Descendants only come from the team's pool, so a stage
// the lobby banned is never a target: with Act 4 banned, Act 2 + INFINITE
// settles on Act 3, the highest allowed stage not above the Spin.
func (st *effectState) evolveStand() (PowerEffect, bool) {
	if st.stand == nil {
		return PowerEffect{}, false
	}
	cur, _ := standSpinTier(st.stand)
	spin := st.spin()
	if spin <= cur {
		return PowerEffect{}, false
	}
	var cands []evoCandidate
	for i, s := range st.standFamily {
		tier, ok := standSpinTier(s)
		if !ok || tier < cur || tier > spin || !isStandDescendant(s, st.stand) {
			continue
		}
		cands = append(cands, evoCandidate{idx: i, tier: int(tier), depth: s.EvolutionDepth(), id: s.ID().String()})
	}
	idx, ok := pickEvolution(cands)
	if !ok {
		return PowerEffect{}, false
	}
	from := st.stand
	st.stand = st.standFamily[idx]
	return PowerEffect{
		Kind:      enums.EffectEvolution,
		Slot:      enums.SlotStand,
		From:      from.ID().String(),
		To:        st.stand.ID().String(),
		CauseSlot: enums.SlotSpin,
		Cause:     spin.String(),
	}, true
}

// evolveFruit is evolveStand for DevilFruits, driven by Fruit Mastery.
func (st *effectState) evolveFruit() (PowerEffect, bool) {
	if st.fruit == nil {
		return PowerEffect{}, false
	}
	cur, _ := fruitMasteryTier(st.fruit)
	mastery := st.mastery()
	if mastery <= cur {
		return PowerEffect{}, false
	}
	var cands []evoCandidate
	for i, d := range st.fruitFamily {
		tier, ok := fruitMasteryTier(d)
		if !ok || tier < cur || tier > mastery || !isFruitDescendant(d, st.fruit) {
			continue
		}
		cands = append(cands, evoCandidate{idx: i, tier: int(tier), depth: d.EvolutionDepth(), id: d.ID().String()})
	}
	idx, ok := pickEvolution(cands)
	if !ok {
		return PowerEffect{}, false
	}
	from := st.fruit
	st.fruit = st.fruitFamily[idx]
	return PowerEffect{
		Kind:      enums.EffectEvolution,
		Slot:      enums.SlotDevilFruit,
		From:      from.ID().String(),
		To:        st.fruit.ID().String(),
		CauseSlot: enums.SlotFruitMastery,
		Cause:     mastery.String(),
	}, true
}

func (st *effectState) floorStandTier() (PowerEffect, bool) {
	tier, ok := standSpinTier(st.stand)
	if !ok || st.spin() >= tier {
		return PowerEffect{}, false
	}
	from := st.spin()
	st.values[enums.SlotSpin] = int(tier)
	return PowerEffect{
		Kind:      enums.EffectStatFloor,
		Slot:      enums.SlotSpin,
		From:      from.String(),
		To:        tier.String(),
		CauseSlot: enums.SlotStand,
		Cause:     st.stand.Name(),
	}, true
}

func (st *effectState) floorFruitTier() (PowerEffect, bool) {
	tier, ok := fruitMasteryTier(st.fruit)
	if !ok || st.mastery() >= tier {
		return PowerEffect{}, false
	}
	from := st.mastery()
	st.values[enums.SlotFruitMastery] = int(tier)
	return PowerEffect{
		Kind:      enums.EffectStatFloor,
		Slot:      enums.SlotFruitMastery,
		From:      from.String(),
		To:        tier.String(),
		CauseSlot: enums.SlotDevilFruit,
		Cause:     st.fruit.Name(),
	}, true
}

// resolvePowerEffects applies every rule until none changes anything, and
// returns what it did in order. One pass runs fruit evolution, the fruit's
// mastery floor, stand evolution, the stand's spin floor, then the stat floor
// table; a change in one rule can enable another (Hamon PERFECT raises Spin,
// which can evolve Tusk), hence the repeat. Fully deterministic: it consumes
// no randomness.
func resolvePowerEffects(st *effectState) ([]PowerEffect, error) {
	var effects []PowerEffect
	for pass := 0; pass < maxEffectPasses; pass++ {
		changed := false
		steps := []func() (PowerEffect, bool){
			st.evolveFruit,
			st.floorFruitTier,
			st.evolveStand,
			st.floorStandTier,
		}
		for _, step := range steps {
			if e, ok := step(); ok {
				effects = append(effects, e)
				changed = true
			}
		}
		for _, r := range statFloorRules {
			if e, ok := st.applyFloor(r); ok {
				effects = append(effects, e)
				changed = true
			}
		}
		if !changed {
			return effects, nil
		}
	}
	return nil, ErrPowerEffectsDiverged
}
