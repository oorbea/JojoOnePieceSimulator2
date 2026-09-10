package characters

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// JojoCharacter is a JoJo-side character: hamon/spin mastery plus a
// hand-authored battleIQ (a WAIS-IV-style score, 0-255). Unlike the
// game.BattleIQ value object a live loadout carries - which can be absent
// for a non-JoJo lobby - a JojoCharacter always has one: an admin
// authoring one picks the score directly, there is no "present/absent"
// state here.
type JojoCharacter struct {
	Character
	hamon    enums.HamonLevel
	spin     enums.SpinLevel
	battleIQ byte
}

func NewJojoCharacter(character Character, hamon enums.HamonLevel, spin enums.SpinLevel, battleIQ byte) (*JojoCharacter, error) {
	if !hamon.IsValid() {
		return nil, enums.ErrInvalidHamonLevel
	}
	if !spin.IsValid() {
		return nil, enums.ErrInvalidSpinLevel
	}
	return &JojoCharacter{
		Character: character,
		hamon:     hamon,
		spin:      spin,
		battleIQ:  battleIQ,
	}, nil
}

func (j *JojoCharacter) Hamon() enums.HamonLevel { return j.hamon }
func (j *JojoCharacter) Spin() enums.SpinLevel   { return j.spin }
func (j *JojoCharacter) BattleIQ() byte          { return j.battleIQ }
