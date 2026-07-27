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

func (p *Player) Username() string {
	return p.username
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

func (p *Player) DevilFruitMasteryToString() (string, error) {
	switch p.fruitMastery {
	case 0:
		return "No fruit", nil
	case 1:
		return "Regular", nil
	case 2:
		return "Advanced", nil
	case 3:
		return "Awakened", nil
	default:
		return "Unknown", errors.New(fmt.Sprintf("Unknown fruit mastery: %v", p.fruitMastery))
	}
}

func (p *Player) SpinToString() (string, error) {
	switch p.spin {
	case 0:
		return "No spin", nil
	case 1:
		return "Basic", nil
	case 2:
		return "Advanced", nil
	case 3:
		return "Golden", nil
	case 4:
		return "Infinite", nil
	default:
		return "Unknown", errors.New(fmt.Sprintf("Unknown spin: %v", p.spin))
	}
}
