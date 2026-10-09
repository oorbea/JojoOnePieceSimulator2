package game

import (
	"fmt"
	"math"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// ManualRules is everything the in-game manual states about the rules, as flat
// JSON-ready data computed from the same tables and the same resolver the game
// runs. cmd/typegen writes it to the frontend's contracts/rules.ts, so the
// manual cannot drift from the code: a rule change that is not regenerated
// fails the `contracts` CI job.
type ManualRules struct {
	// TieRule says how an evolution tie is broken: FURTHEST_EVOLVED.
	TieRule     string             `json:"tieRule"`
	Limits      ManualLimits       `json:"limits"`
	Evolutions  []ManualChain      `json:"evolutions"`
	StatFloors  []ManualStatFloor  `json:"statFloors"`
	Odds        ManualOdds         `json:"odds"`
	BattleIQ    []ManualBattleBand `json:"battleIQ"`
	Conventions []ManualConvention `json:"conventions"`
}

// ManualLimits are the numbers the manual quotes about how a game is played,
// straight from the game's own constants (config.go).
type ManualLimits struct {
	GauntletMinPlayers     int `json:"gauntletMinPlayers"`
	GauntletMaxPlayers     int `json:"gauntletMaxPlayers"`
	VersusMinTeamSize      int `json:"versusMinTeamSize"`
	VersusMaxTeamSize      int `json:"versusMaxTeamSize"`
	VersusTeamCount        int `json:"versusTeamCount"`
	VersusRounds           int `json:"versusRounds"`
	VotingDefaultSeconds   int `json:"votingDefaultSeconds"`
	VotingMinSeconds       int `json:"votingMinSeconds"`
	VotingMaxSeconds       int `json:"votingMaxSeconds"`
	VotingExtensionSeconds int `json:"votingExtensionSeconds"`
}

// ManualChain is one evolution family and what every (drawn stage x driver
// level) combination ends up as.
type ManualChain struct {
	Kind   string        `json:"kind"`   // STAND | DEVIL_FRUIT
	Driver string        `json:"driver"` // SPIN | FRUIT_MASTERY
	Stages []ManualStage `json:"stages"`
	Levels []string      `json:"levels"` // driver levels, ascending
	Rows   []ManualRow   `json:"rows"`   // one per drawn stage
}

// ManualStage is one stage of a chain and the driver level it requires ("" when
// the stage has no requirement, e.g. Soft & Wet before Go Beyond).
type ManualStage struct {
	Name string `json:"name"`
	Tier string `json:"tier,omitempty"`
}

// ManualRow is one drawn stage and its outcome per driver level.
type ManualRow struct {
	Drawn    string          `json:"drawn"`
	Outcomes []ManualOutcome `json:"outcomes"`
}

// ManualOutcome is what a drawn stage becomes at one driver level. Change is
// NONE, LEVEL_RAISED (the driver rose to the stage's tier) or EVOLVED.
type ManualOutcome struct {
	Level      string `json:"level"`
	Power      string `json:"power"`
	LevelAfter string `json:"levelAfter"`
	Change     string `json:"change"`
}

// ManualStatFloor is one "X raises Y to at least Z" rule.
type ManualStatFloor struct {
	Cross     bool          `json:"cross"`
	CauseSlot string        `json:"causeSlot"`
	When      ManualSubject `json:"when"`
	Target    string        `json:"target"`
	Floor     string        `json:"floor"`
}

// ManualOdds are the draw probabilities BEFORE power effects, which only ever
// raise stats. Percent values are rounded to two decimals.
type ManualOdds struct {
	// NoStandWeight / NoFruitWeight weigh "none" against a weight of 1 for every
	// candidate in the pool: P(none) = w / (w + poolSize).
	NoStandWeight int `json:"noStandWeight"`
	NoFruitWeight int `json:"noFruitWeight"`

	Spin         []ManualPct     `json:"spin"`
	Hamon        []ManualPct     `json:"hamon"`
	FruitMastery []ManualPct     `json:"fruitMastery"`
	PhysicalForm []ManualPct     `json:"physicalForm"`
	HakiMastery  []ManualPct     `json:"hakiMastery"`
	HakiSets     []ManualHakiSet `json:"hakiSets"`
	// HakiPresence is the chance of having each haki type at all.
	HakiPresence ManualHakiPresence `json:"hakiPresence"`
}

// ManualPct is one level and its probability in percent.
type ManualPct struct {
	Level   string  `json:"level"`
	Percent float64 `json:"percent"`
}

// ManualHakiSet is one combination of haki types and its probability.
type ManualHakiSet struct {
	Armament    bool    `json:"armament"`
	Observation bool    `json:"observation"`
	Conqueror   bool    `json:"conqueror"`
	Percent     float64 `json:"percent"`
}

// ManualHakiPresence is the chance of having each haki type, in percent.
type ManualHakiPresence struct {
	Armament    float64 `json:"armament"`
	Observation float64 `json:"observation"`
	Conqueror   float64 `json:"conqueror"`
}

// ManualBattleBand is one WAIS-IV band: its key, inclusive score range and
// the chance a JoJo draw lands in it.
type ManualBattleBand struct {
	Key     string  `json:"key"`
	Lo      int     `json:"lo"`
	Hi      int     `json:"hi"`
	Percent float64 `json:"percent"`
}

// ManualSubject is Subject with wire strings only.
type ManualSubject struct {
	Kind       string   `json:"kind"`
	Slot       string   `json:"slot,omitempty"`
	Min        string   `json:"min,omitempty"`
	FruitTypes []string `json:"fruitTypes,omitempty"`
	Names      []string `json:"names,omitempty"`
	Group      string   `json:"group,omitempty"`
}

// ManualConvention is Convention with wire strings only.
type ManualConvention struct {
	ID         string          `json:"id"`
	Category   string          `json:"category"`
	Actors     []ManualSubject `json:"actors"`
	Targets    []ManualSubject `json:"targets"`
	Immune     []ManualSubject `json:"immune"`
	Attenuated []ManualSubject `json:"attenuated"`
	Gap        int             `json:"gap"`
	Judgement  bool            `json:"judgement"`
}

// Evolution chains the manual draws. The catalogue stores a Stand's
// evolves_from link in the database, not in code, so the stage order is stated
// here (it matches prod's Tusk and Soft & Wet chains); each stage's tier comes
// from the real rule tables and every outcome from the real resolver. A tiered
// stage missing from these chains fails TestManualRules_CoverEveryTier.
var (
	manualStandChains = [][]string{
		{"Tusk: Act 1", "Tusk: Act 2", "Tusk: Act 3", "Tusk: Act 4"},
		{"Soft & Wet", convGoBeyond},
		{"Ball Breaker"},
	}
	manualFruitChains = [][]string{
		{convGomuGomu, convNika},
	}
)

// BuildManualRules computes ManualRules from the live rule tables.
func BuildManualRules() ManualRules {
	var chains []ManualChain
	for i, names := range manualStandChains {
		chains = append(chains, standChain(byte(10*(i+1)), names))
	}
	for i, names := range manualFruitChains {
		chains = append(chains, fruitChain(byte(100+10*i), names))
	}
	floors := make([]ManualStatFloor, 0, len(statFloorRules))
	for _, r := range statFloorRules {
		floors = append(floors, ManualStatFloor{
			Cross:     r.cross,
			CauseSlot: r.causeSlot.String(),
			When:      toManualSubject(r.when),
			Target:    r.target.String(),
			Floor:     slotLabel(r.target, r.floor),
		})
	}
	convs := CombatConventions()
	manualConvs := make([]ManualConvention, 0, len(convs))
	for _, c := range convs {
		manualConvs = append(manualConvs, ManualConvention{
			ID:         c.ID,
			Category:   string(c.Category),
			Actors:     toManualSubjects(c.Actors),
			Targets:    toManualSubjects(c.Targets),
			Immune:     toManualSubjects(c.Immune),
			Attenuated: toManualSubjects(c.Attenuated),
			Gap:        c.Gap,
			Judgement:  c.Judgement,
		})
	}
	w := DefaultAssignmentWeights()
	return ManualRules{
		TieRule: "FURTHEST_EVOLVED",
		Limits: ManualLimits{
			GauntletMinPlayers:     MinGauntletPlayers,
			GauntletMaxPlayers:     MaxGauntletPlayers,
			VersusMinTeamSize:      MinVersusTeamSize,
			VersusMaxTeamSize:      MaxVersusTeamSize,
			VersusTeamCount:        VersusTeamCount,
			VersusRounds:           VersusRounds,
			VotingDefaultSeconds:   DefaultVotingWindowSeconds,
			VotingMinSeconds:       MinVotingWindowSeconds,
			VotingMaxSeconds:       MaxVotingWindowSeconds,
			VotingExtensionSeconds: VotingExtensionSeconds,
		},
		Evolutions:  chains,
		StatFloors:  floors,
		Odds:        manualOdds(w),
		BattleIQ:    manualBattleIQ(w),
		Conventions: manualConvs,
	}
}

func toManualSubjects(in []Subject) []ManualSubject {
	out := make([]ManualSubject, 0, len(in))
	for _, s := range in {
		out = append(out, toManualSubject(s))
	}
	return out
}

func toManualSubject(s Subject) ManualSubject {
	m := ManualSubject{Kind: string(s.Kind), Min: s.Min, Names: s.Names, Group: s.Group}
	if s.Kind == SubjectSlotMin {
		m.Slot = s.Slot.String()
	}
	for _, t := range s.FruitTypes {
		m.FruitTypes = append(m.FruitTypes, t.String())
	}
	return m
}

func mustPower(id byte, name string) powers.Power {
	skills := []string{"skill"}
	p, err := powers.NewPower(powers.PowerID{id}, name, "description", enums.Epic, &skills, "")
	if err != nil {
		panic(fmt.Sprintf("manual rules: NewPower(%q): %v", name, err))
	}
	return *p
}

// standChain resolves every (drawn stage x Spin level) of one Stand chain with
// the real effect resolver.
func standChain(baseID byte, names []string) ManualChain {
	stands := make([]*powers.Stand, len(names))
	var parent *powers.Stand
	for i, n := range names {
		s, err := powers.NewStand(mustPower(baseID+byte(i), n), enums.A, enums.A, enums.A, enums.A, enums.A, enums.A, parent)
		if err != nil {
			panic(fmt.Sprintf("manual rules: NewStand(%q): %v", n, err))
		}
		stands[i], parent = s, s
	}
	chain := ManualChain{Kind: enums.SlotStand.String(), Driver: enums.SlotSpin.String()}
	for _, l := range spinLevels {
		chain.Levels = append(chain.Levels, l.String())
	}
	for _, s := range stands {
		stage := ManualStage{Name: s.Name()}
		if t, ok := standSpinTier(s); ok {
			stage.Tier = t.String()
		}
		chain.Stages = append(chain.Stages, stage)
	}
	for _, drawn := range stands {
		row := ManualRow{Drawn: drawn.Name()}
		for _, spin := range spinLevels {
			st := &effectState{stand: drawn, standFamily: stands}
			st.values[enums.SlotSpin] = int(spin)
			if _, err := resolvePowerEffects(st); err != nil {
				panic(fmt.Sprintf("manual rules: %v", err))
			}
			row.Outcomes = append(row.Outcomes, ManualOutcome{
				Level:      spin.String(),
				Power:      st.stand.Name(),
				LevelAfter: st.spin().String(),
				Change:     changeKind(st.stand != drawn, st.spin() != spin),
			})
		}
		chain.Rows = append(chain.Rows, row)
	}
	return chain
}

// fruitChain is standChain for Devil Fruits, driven by Fruit Mastery.
func fruitChain(baseID byte, names []string) ManualChain {
	raw := make([]*powers.DevilFruit, len(names))
	for i, n := range names {
		ft := enums.Paramecia
		if normalizePowerName(n) == nameNika {
			ft = enums.MythicalZoan
		}
		f, err := powers.NewDevilFruit(mustPower(baseID+byte(i), n), ft)
		if err != nil {
			panic(fmt.Sprintf("manual rules: NewDevilFruit(%q): %v", n, err))
		}
		raw[i] = f
	}
	fruits := LinkFruitEvolutions(raw)
	chain := ManualChain{Kind: enums.SlotDevilFruit.String(), Driver: enums.SlotFruitMastery.String()}
	for _, l := range fruitMasteryLevels {
		chain.Levels = append(chain.Levels, l.String())
	}
	for _, f := range fruits {
		stage := ManualStage{Name: f.Name()}
		if t, ok := fruitMasteryTier(f); ok {
			stage.Tier = t.String()
		}
		chain.Stages = append(chain.Stages, stage)
	}
	for _, drawn := range fruits {
		row := ManualRow{Drawn: drawn.Name()}
		for _, m := range fruitMasteryLevels {
			st := &effectState{fruit: drawn, fruitFamily: fruits}
			st.values[enums.SlotFruitMastery] = int(m)
			if _, err := resolvePowerEffects(st); err != nil {
				panic(fmt.Sprintf("manual rules: %v", err))
			}
			row.Outcomes = append(row.Outcomes, ManualOutcome{
				Level:      m.String(),
				Power:      st.fruit.Name(),
				LevelAfter: st.mastery().String(),
				Change:     changeKind(st.fruit != drawn, st.mastery() != m),
			})
		}
		chain.Rows = append(chain.Rows, row)
	}
	return chain
}

func changeKind(evolved, raised bool) string {
	switch {
	case evolved:
		return "EVOLVED"
	case raised:
		return "LEVEL_RAISED"
	default:
		return "NONE"
	}
}

func percent(w, total int) float64 {
	if total == 0 {
		return 0
	}
	return math.Round(float64(w)*10000/float64(total)) / 100
}

func pcts[L interface{ String() string }](levels []L, weight func(L) int) []ManualPct {
	total := 0
	for _, l := range levels {
		total += weight(l)
	}
	out := make([]ManualPct, 0, len(levels))
	for _, l := range levels {
		out = append(out, ManualPct{Level: l.String(), Percent: percent(weight(l), total)})
	}
	return out
}

func manualOdds(w AssignmentWeights) ManualOdds {
	hakiTotal := 0
	for _, s := range hakiSets {
		hakiTotal += w.HakiSetWeights[s]
	}
	var sets []ManualHakiSet
	var arm, obs, con int
	for _, s := range hakiSets {
		wt := w.HakiSetWeights[s]
		sets = append(sets, ManualHakiSet{
			Armament: s.HasArmament(), Observation: s.HasObservation(), Conqueror: s.HasConqueror(),
			Percent: percent(wt, hakiTotal),
		})
		if s.HasArmament() {
			arm += wt
		}
		if s.HasObservation() {
			obs += wt
		}
		if s.HasConqueror() {
			con += wt
		}
	}
	return ManualOdds{
		NoStandWeight: w.NoStandWeight,
		NoFruitWeight: w.NoDevilFruitWeight,
		Spin:          pcts(spinLevels, func(l enums.SpinLevel) int { return w.SpinLevelWeights[l] }),
		Hamon:         pcts(hamonLevels, func(l enums.HamonLevel) int { return w.HamonLevelWeights[l] }),
		FruitMastery:  pcts(fruitMasteryLevels, func(l enums.FruitMastery) int { return w.FruitMasteryWeights[l] }),
		PhysicalForm:  pcts(physicalFormLevels, func(l enums.PhysicalForm) int { return w.PhysicalFormWeights[l] }),
		HakiMastery:   pcts(hakiMasteryLevels, func(l enums.HakiLevel) int { return w.HakiMasteryWeights[l] }),
		HakiSets:      sets,
		HakiPresence: ManualHakiPresence{
			Armament:    percent(arm, hakiTotal),
			Observation: percent(obs, hakiTotal),
			Conqueror:   percent(con, hakiTotal),
		},
	}
}

func manualBattleIQ(w AssignmentWeights) []ManualBattleBand {
	total := 0
	for _, b := range battleIQBands {
		total += w.BattleIQBandWeights[b]
	}
	out := make([]ManualBattleBand, 0, len(battleIQBands))
	for _, b := range battleIQBands {
		lo, hi := b.Range()
		out = append(out, ManualBattleBand{
			Key: b.Key(), Lo: int(lo), Hi: int(hi), Percent: percent(w.BattleIQBandWeights[b], total),
		})
	}
	return out
}
