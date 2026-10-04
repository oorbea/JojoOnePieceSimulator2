package enums

import "errors"

// LoadoutSlot names one ability of a Loadout. It exists so a PowerEffect can
// say which slot it changed (and which slot caused it) in a form that
// survives the wire. Declared in the sorteo's reveal order, which is also the
// order the LoadoutBuilder resolves them in - see game.revealSlotOrder.
type LoadoutSlot byte

const (
	SlotPhysicalForm LoadoutSlot = iota
	SlotStand
	SlotDevilFruit
	SlotFruitMastery
	SlotHamon
	SlotArmamentHaki
	SlotObservationHaki
	SlotConquerorHaki
	SlotSpin
)

func (s LoadoutSlot) String() string {
	switch s {
	case SlotPhysicalForm:
		return "PHYSICAL_FORM"
	case SlotStand:
		return "STAND"
	case SlotDevilFruit:
		return "DEVIL_FRUIT"
	case SlotFruitMastery:
		return "FRUIT_MASTERY"
	case SlotHamon:
		return "HAMON"
	case SlotArmamentHaki:
		return "ARMAMENT_HAKI"
	case SlotObservationHaki:
		return "OBSERVATION_HAKI"
	case SlotConquerorHaki:
		return "CONQUEROR_HAKI"
	case SlotSpin:
		return "SPIN"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidLoadoutSlot = errors.New("invalid loadout slot")

func (s LoadoutSlot) IsValid() bool {
	switch s {
	case SlotPhysicalForm, SlotStand, SlotDevilFruit, SlotFruitMastery, SlotHamon,
		SlotArmamentHaki, SlotObservationHaki, SlotConquerorHaki, SlotSpin:
		return true
	default:
		return false
	}
}

func ParseLoadoutSlot(str string) (LoadoutSlot, error) {
	switch str {
	case "PHYSICAL_FORM":
		return SlotPhysicalForm, nil
	case "STAND":
		return SlotStand, nil
	case "DEVIL_FRUIT":
		return SlotDevilFruit, nil
	case "FRUIT_MASTERY":
		return SlotFruitMastery, nil
	case "HAMON":
		return SlotHamon, nil
	case "ARMAMENT_HAKI":
		return SlotArmamentHaki, nil
	case "OBSERVATION_HAKI":
		return SlotObservationHaki, nil
	case "CONQUEROR_HAKI":
		return SlotConquerorHaki, nil
	case "SPIN":
		return SlotSpin, nil
	default:
		return SlotPhysicalForm, ErrInvalidLoadoutSlot
	}
}
