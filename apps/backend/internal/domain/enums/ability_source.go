package enums

import "errors"

// AbilitySource selects how a Versus match draws player Loadouts: freshly
// randomized, or pulled from the player's owned-power inventory. Only
// Random is implemented today - Inventory is reserved for when a player
// inventory exists (see ports.IInventory).
type AbilitySource byte

const (
	Random AbilitySource = iota
	Inventory
)

func (s AbilitySource) String() string {
	switch s {
	case Random:
		return "RANDOM"
	case Inventory:
		return "INVENTORY"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidAbilitySource = errors.New("invalid ability source")

func (s AbilitySource) IsValid() bool {
	switch s {
	case Random, Inventory:
		return true
	default:
		return false
	}
}

func ParseAbilitySource(str string) (AbilitySource, error) {
	switch str {
	case "RANDOM":
		return Random, nil
	case "INVENTORY":
		return Inventory, nil
	default:
		return Random, ErrInvalidAbilitySource
	}
}
