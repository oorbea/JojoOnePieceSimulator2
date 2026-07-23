package entities

import (
	"errors"
	"fmt"
)

type Player struct {
	username        string
	physicalForm    byte
	stand           *Stand
	devilFruit      *DevilFruit
	fruitMastery    byte
	armamentHaki    byte
	observationHaki byte
	conquerorHaki   byte
	spin            byte
}

func onePieceStatToString(stat byte) (string, error) {
	switch stat {
	case 0:
		return "Private", nil
	case 1:
		return "Vice Admiral", nil
	case 2:
		return "Yonko Commander", nil
	case 3:
		return "Yonko+", nil
	default:
		return "Unknown", errors.New(fmt.Sprintf("Unknown stat: %v", stat))
	}
}

func (p *Player) PhysicalFormToString() (string, error) {
	return onePieceStatToString(p.physicalForm)
}

func (p *Player) ArmamentHakiToString() (string, error) {
	return onePieceStatToString(p.armamentHaki)
}

func (p *Player) ObservationHakiToString() (string, error) {
	return onePieceStatToString(p.observationHaki)
}

func (p *Player) ConquerorHakiToString() (string, error) {
	return onePieceStatToString(p.conquerorHaki)
}
