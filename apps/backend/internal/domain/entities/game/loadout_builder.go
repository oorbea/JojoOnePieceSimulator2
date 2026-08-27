package game

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

var spinLevels = []enums.SpinLevel{enums.SpinNone, enums.SpinBasic, enums.SpinGolden, enums.SpinInfinite}
var hamonLevels = []enums.HamonLevel{enums.HamonNone, enums.HamonBasic, enums.HamonAdvanced, enums.HamonPerfect}
var fruitMasteryLevels = []enums.FruitMastery{enums.FruitMasteryRegular, enums.FruitMasteryAdvanced, enums.FruitMasteryAwakened}
var hakiMasteryLevels = []enums.HakiLevel{enums.HakiPrivate, enums.HakiViceAdmiral, enums.HakiYonkoCommander, enums.HakiYonkoPlus}
var physicalFormLevels = []enums.PhysicalForm{
	enums.PhysicalFormPrivate, enums.PhysicalFormStrongFishman, enums.PhysicalFormMarineCaptain,
	enums.PhysicalFormViceAdmiral, enums.PhysicalFormYonkoCommander, enums.PhysicalFormYonkoPlus,
}

// LoadoutBuilder is the Template Method that assembles a Loadout for one
// participant. Which abilities get drawn at all is fixed by the mangas in
// play (JoJo draws Stand/Spin/Hamon, One Piece draws PhysicalForm/DevilFruit/
// FruitMastery/the three Hakis), always in this fixed step order regardless
// of manga - each step is simply skipped when its manga isn't selected:
// PhysicalForm -> Stand -> DevilFruit -> FruitMastery -> Hamon -> Haki (set,
// then a mastery per haki present) -> Spin. DevilFruit is drawn before its
// FruitMastery, since the latter depends on the former; the RequiresSpin4
// override runs after Spin is drawn, last of all. The concrete draws are
// weighted random picks (RandomSource + AssignmentWeights) except Stand and
// DevilFruit, which are uniform over the pool - see AssignmentWeights' doc
// comment for why. The hard invariants (fruit<->mastery coupling,
// RequiresSpin4) are re-checked by NewLoadout at the very end regardless of
// what the weighted draws produced.
type LoadoutBuilder struct {
	mangas  map[enums.Manga]struct{}
	weights AssignmentWeights
	rng     RandomSource
}

// NewLoadoutBuilder builds a LoadoutBuilder for the given manga selection.
func NewLoadoutBuilder(mangas []enums.Manga, weights AssignmentWeights, rng RandomSource) *LoadoutBuilder {
	set := make(map[enums.Manga]struct{}, len(mangas))
	for _, m := range mangas {
		set[m] = struct{}{}
	}
	return &LoadoutBuilder{mangas: set, weights: weights, rng: rng}
}

func (b *LoadoutBuilder) hasManga(m enums.Manga) bool {
	_, ok := b.mangas[m]
	return ok
}

// Build draws a full Loadout for one participant. pool must not be shared
// with another team (see AvailablePowers) so a Stand/DevilFruit already
// drawn by a teammate cannot be drawn again.
func (b *LoadoutBuilder) Build(pool *AvailablePowers) (*Loadout, error) {
	var (
		stand           *powers.Stand
		devilFruit      *powers.DevilFruit
		spin            = enums.SpinNone
		hamon           = enums.HamonNone
		fruitMastery    = enums.FruitMasteryNone
		armamentHaki    = enums.HakiNone
		observationHaki = enums.HakiNone
		conquerorHaki   = enums.HakiNone
		physicalForm    = enums.PhysicalFormPrivate
		err             error
	)

	if b.hasManga(enums.OnePiece) {
		physicalForm = b.drawPhysicalForm()
	}

	if b.hasManga(enums.Jojo) {
		if stand, err = b.drawStand(pool); err != nil {
			return nil, err
		}
	}

	if b.hasManga(enums.OnePiece) {
		if devilFruit, err = b.drawDevilFruit(pool); err != nil {
			return nil, err
		}
		fruitMastery = b.drawFruitMastery(devilFruit)
	}

	if b.hasManga(enums.Jojo) {
		hamon = b.drawHamon()
	}

	if b.hasManga(enums.OnePiece) {
		set := b.drawHakiSet()
		if set.HasArmament() {
			armamentHaki = b.drawHakiMastery()
		}
		if set.HasObservation() {
			observationHaki = b.drawHakiMastery()
		}
		if set.HasConqueror() {
			conquerorHaki = b.drawHakiMastery()
		}
	}

	if b.hasManga(enums.Jojo) {
		spin = b.drawSpin()
	}

	if stand != nil && HasTrait(&stand.Power, enums.RequiresSpin4) {
		spin = enums.SpinInfinite
	}

	return NewLoadout(stand, devilFruit, spin, hamon, fruitMastery, armamentHaki, observationHaki, conquerorHaki, physicalForm)
}

// drawStand picks uniformly among "no stand" and every Stand in the pool -
// deliberately not weighted by rarity, matching V1's
// uniform_int_distribution over the whole power array (see
// AssignmentWeights' doc comment).
func (b *LoadoutBuilder) drawStand(pool *AvailablePowers) (*powers.Stand, error) {
	stands := pool.Stands()
	weights := make([]int, len(stands)+1)
	weights[0] = b.weights.NoStandWeight
	for i := range stands {
		weights[i+1] = 1
	}
	idx := weightedPick(b.rng, weights)
	if idx == 0 {
		return nil, nil
	}
	return pool.DrawStand(idx - 1)
}

// drawDevilFruit picks uniformly among "no fruit" and every DevilFruit in
// the pool - see drawStand's doc comment.
func (b *LoadoutBuilder) drawDevilFruit(pool *AvailablePowers) (*powers.DevilFruit, error) {
	fruits := pool.DevilFruits()
	weights := make([]int, len(fruits)+1)
	weights[0] = b.weights.NoDevilFruitWeight
	for i := range fruits {
		weights[i+1] = 1
	}
	idx := weightedPick(b.rng, weights)
	if idx == 0 {
		return nil, nil
	}
	return pool.DrawDevilFruit(idx - 1)
}

func (b *LoadoutBuilder) drawSpin() enums.SpinLevel {
	weights := make([]int, len(spinLevels))
	for i, l := range spinLevels {
		weights[i] = b.weights.SpinLevelWeights[l]
	}
	return spinLevels[weightedPick(b.rng, weights)]
}

func (b *LoadoutBuilder) drawHamon() enums.HamonLevel {
	weights := make([]int, len(hamonLevels))
	for i, l := range hamonLevels {
		weights[i] = b.weights.HamonLevelWeights[l]
	}
	return hamonLevels[weightedPick(b.rng, weights)]
}

// drawFruitMastery only runs once a DevilFruit has already been drawn -
// without one, mastery is always FruitMasteryNone.
func (b *LoadoutBuilder) drawFruitMastery(devilFruit *powers.DevilFruit) enums.FruitMastery {
	if devilFruit == nil {
		return enums.FruitMasteryNone
	}
	weights := make([]int, len(fruitMasteryLevels))
	for i, l := range fruitMasteryLevels {
		weights[i] = b.weights.FruitMasteryWeights[l]
	}
	return fruitMasteryLevels[weightedPick(b.rng, weights)]
}

// drawHakiSet picks which combination of Armament/Observation/Conqueror
// haki the player has at all, as a single correlated draw - see HakiSet's
// doc comment.
func (b *LoadoutBuilder) drawHakiSet() HakiSet {
	weights := make([]int, len(hakiSets))
	for i, s := range hakiSets {
		weights[i] = b.weights.HakiSetWeights[s]
	}
	return hakiSets[weightedPick(b.rng, weights)]
}

// drawHakiMastery picks the mastery level for one haki type the player has
// already been determined to have (see drawHakiSet) - never returns
// HakiNone.
func (b *LoadoutBuilder) drawHakiMastery() enums.HakiLevel {
	weights := make([]int, len(hakiMasteryLevels))
	for i, l := range hakiMasteryLevels {
		weights[i] = b.weights.HakiMasteryWeights[l]
	}
	return hakiMasteryLevels[weightedPick(b.rng, weights)]
}

func (b *LoadoutBuilder) drawPhysicalForm() enums.PhysicalForm {
	weights := make([]int, len(physicalFormLevels))
	for i, l := range physicalFormLevels {
		weights[i] = b.weights.PhysicalFormWeights[l]
	}
	return physicalFormLevels[weightedPick(b.rng, weights)]
}
