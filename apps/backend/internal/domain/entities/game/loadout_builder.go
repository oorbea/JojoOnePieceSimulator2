package game

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

var spinLevels = []enums.SpinLevel{enums.SpinNone, enums.SpinBasic, enums.SpinAdvanced, enums.SpinGolden, enums.SpinInfinite}
var hamonLevels = []enums.HamonLevel{enums.HamonNone, enums.HamonBasic, enums.HamonAdvanced, enums.HamonPerfect}
var fruitMasteryLevels = []enums.FruitMastery{enums.FruitMasteryRegular, enums.FruitMasteryAdvanced, enums.FruitMasteryAwakened}
var hakiLevels = []enums.HakiLevel{enums.HakiPrivate, enums.HakiViceAdmiral, enums.HakiYonkoCommander, enums.HakiYonkoPlus}
var physicalFormLevels = []enums.PhysicalForm{enums.PhysicalFormPrivate, enums.PhysicalFormViceAdmiral, enums.PhysicalFormYonkoCommander, enums.PhysicalFormYonkoPlus}

// LoadoutBuilder is the Template Method that assembles a Loadout for one
// participant. Which abilities get drawn at all is fixed by the mangas in
// play (JoJo draws Stand/Spin/Hamon, One Piece draws DevilFruit/
// FruitMastery/the three Hakis/PhysicalForm) and always in this order: the
// DevilFruit is drawn before its FruitMastery, since the latter depends on
// the former. The concrete draws are weighted random picks
// (RandomSource + AssignmentWeights); the hard invariants (fruit<->mastery
// coupling, RequiresSpin4) are re-checked by NewLoadout at the very end
// regardless of what the weighted draws produced.
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
		armamentHaki    = enums.HakiPrivate
		observationHaki = enums.HakiPrivate
		conquerorHaki   = enums.HakiPrivate
		physicalForm    = enums.PhysicalFormPrivate
		err             error
	)

	if b.hasManga(enums.Jojo) {
		if stand, err = b.drawStand(pool); err != nil {
			return nil, err
		}
		spin = b.drawSpin()
		hamon = b.drawHamon()
	}

	if b.hasManga(enums.OnePiece) {
		if devilFruit, err = b.drawDevilFruit(pool); err != nil {
			return nil, err
		}
		fruitMastery = b.drawFruitMastery(devilFruit)
		armamentHaki = b.drawHaki(b.weights.ArmamentHakiWeights)
		observationHaki = b.drawHaki(b.weights.ObservationHakiWeights)
		conquerorHaki = b.drawHaki(b.weights.ConquerorHakiWeights)
		physicalForm = b.drawPhysicalForm()
	}

	if stand != nil && HasTrait(&stand.Power, enums.RequiresSpin4) {
		spin = enums.SpinInfinite
	}

	return NewLoadout(stand, devilFruit, spin, hamon, fruitMastery, armamentHaki, observationHaki, conquerorHaki, physicalForm)
}

func (b *LoadoutBuilder) drawStand(pool *AvailablePowers) (*powers.Stand, error) {
	stands := pool.Stands()
	weights := make([]int, len(stands)+1)
	weights[0] = b.weights.NoStandWeight
	for i, s := range stands {
		weights[i+1] = b.weights.RarityWeight[s.Rarity()]
	}
	idx := weightedPick(b.rng, weights)
	if idx == 0 {
		return nil, nil
	}
	return pool.DrawStand(idx - 1)
}

func (b *LoadoutBuilder) drawDevilFruit(pool *AvailablePowers) (*powers.DevilFruit, error) {
	fruits := pool.DevilFruits()
	weights := make([]int, len(fruits)+1)
	weights[0] = b.weights.NoDevilFruitWeight
	for i, f := range fruits {
		weights[i+1] = b.weights.RarityWeight[f.Rarity()]
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

func (b *LoadoutBuilder) drawHaki(weightsByLevel map[enums.HakiLevel]int) enums.HakiLevel {
	weights := make([]int, len(hakiLevels))
	for i, l := range hakiLevels {
		weights[i] = weightsByLevel[l]
	}
	return hakiLevels[weightedPick(b.rng, weights)]
}

func (b *LoadoutBuilder) drawPhysicalForm() enums.PhysicalForm {
	weights := make([]int, len(physicalFormLevels))
	for i, l := range physicalFormLevels {
		weights[i] = b.weights.PhysicalFormWeights[l]
	}
	return physicalFormLevels[weightedPick(b.rng, weights)]
}
