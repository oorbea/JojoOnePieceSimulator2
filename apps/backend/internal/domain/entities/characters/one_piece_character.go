package characters

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// OnePieceCharacter is a One Piece-side character: physical conditioning,
// the three haki types, and Devil Fruit mastery.
type OnePieceCharacter struct {
	Character
	physicalForm    enums.PhysicalForm
	armamentHaki    enums.HakiLevel
	observationHaki enums.HakiLevel
	conquerorHaki   enums.HakiLevel
	fruitMastery    enums.FruitMastery
}

func NewOnePieceCharacter(
	character Character,
	physicalForm enums.PhysicalForm,
	armamentHaki enums.HakiLevel,
	observationHaki enums.HakiLevel,
	conquerorHaki enums.HakiLevel,
	fruitMastery enums.FruitMastery,
) (*OnePieceCharacter, error) {
	if !physicalForm.IsValid() {
		return nil, enums.ErrInvalidPhysicalForm
	}
	if !armamentHaki.IsValid() {
		return nil, enums.ErrInvalidHakiLevel
	}
	if !observationHaki.IsValid() {
		return nil, enums.ErrInvalidHakiLevel
	}
	if !conquerorHaki.IsValid() {
		return nil, enums.ErrInvalidHakiLevel
	}
	if !fruitMastery.IsValid() {
		return nil, enums.ErrInvalidFruitMastery
	}
	return &OnePieceCharacter{
		Character:       character,
		physicalForm:    physicalForm,
		armamentHaki:    armamentHaki,
		observationHaki: observationHaki,
		conquerorHaki:   conquerorHaki,
		fruitMastery:    fruitMastery,
	}, nil
}

func (o *OnePieceCharacter) PhysicalForm() enums.PhysicalForm { return o.physicalForm }
func (o *OnePieceCharacter) ArmamentHaki() enums.HakiLevel    { return o.armamentHaki }
func (o *OnePieceCharacter) ObservationHaki() enums.HakiLevel { return o.observationHaki }
func (o *OnePieceCharacter) ConquerorHaki() enums.HakiLevel   { return o.conquerorHaki }
func (o *OnePieceCharacter) FruitMastery() enums.FruitMastery { return o.fruitMastery }
