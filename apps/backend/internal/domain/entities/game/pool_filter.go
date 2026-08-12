package game

import (
	"errors"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// ErrInvalidPoolFilter is returned when NewPoolFilter is given an invalid or
// duplicated rarity/fruit type, or a nil banned PowerID.
var ErrInvalidPoolFilter = errors.New("invalid pool filter")

// PoolFilter is a host-configured restriction over the Stand/DevilFruit
// catalog a lobby draws Loadouts from: category allowlists (empty means "no
// restriction, everything of that kind is allowed") plus an explicit
// banlist that always wins regardless of category. It is a value object -
// applying it to a catalog is a pure function, no ports, no context.
type PoolFilter struct {
	standRarities []enums.PowerRarity
	fruitRarities []enums.PowerRarity
	fruitTypes    []enums.FruitType
	banned        []powers.PowerID
}

// NewPoolFilter validates and builds a PoolFilter. Any nil slice is treated
// as "no restriction" for that dimension.
func NewPoolFilter(
	standRarities []enums.PowerRarity,
	fruitRarities []enums.PowerRarity,
	fruitTypes []enums.FruitType,
	banned []powers.PowerID,
) (PoolFilter, error) {
	standRaritiesCopy, err := dedupRarities(standRarities)
	if err != nil {
		return PoolFilter{}, err
	}
	fruitRaritiesCopy, err := dedupRarities(fruitRarities)
	if err != nil {
		return PoolFilter{}, err
	}

	seenTypes := make(map[enums.FruitType]struct{}, len(fruitTypes))
	fruitTypesCopy := make([]enums.FruitType, 0, len(fruitTypes))
	for _, t := range fruitTypes {
		if !t.IsValid() {
			return PoolFilter{}, ErrInvalidPoolFilter
		}
		if _, dup := seenTypes[t]; dup {
			continue
		}
		seenTypes[t] = struct{}{}
		fruitTypesCopy = append(fruitTypesCopy, t)
	}

	seenBanned := make(map[powers.PowerID]struct{}, len(banned))
	bannedCopy := make([]powers.PowerID, 0, len(banned))
	for _, id := range banned {
		if id.IsNil() {
			return PoolFilter{}, ErrInvalidPoolFilter
		}
		if _, dup := seenBanned[id]; dup {
			continue
		}
		seenBanned[id] = struct{}{}
		bannedCopy = append(bannedCopy, id)
	}

	return PoolFilter{
		standRarities: standRaritiesCopy,
		fruitRarities: fruitRaritiesCopy,
		fruitTypes:    fruitTypesCopy,
		banned:        bannedCopy,
	}, nil
}

func dedupRarities(rarities []enums.PowerRarity) ([]enums.PowerRarity, error) {
	seen := make(map[enums.PowerRarity]struct{}, len(rarities))
	out := make([]enums.PowerRarity, 0, len(rarities))
	for _, r := range rarities {
		if !r.IsValid() {
			return nil, ErrInvalidPoolFilter
		}
		if _, dup := seen[r]; dup {
			continue
		}
		seen[r] = struct{}{}
		out = append(out, r)
	}
	return out, nil
}

// StandRarities, FruitRarities, FruitTypes, Banned return copies of the
// filter's configured restrictions.
func (f PoolFilter) StandRarities() []enums.PowerRarity {
	return append([]enums.PowerRarity(nil), f.standRarities...)
}
func (f PoolFilter) FruitRarities() []enums.PowerRarity {
	return append([]enums.PowerRarity(nil), f.fruitRarities...)
}
func (f PoolFilter) FruitTypes() []enums.FruitType {
	return append([]enums.FruitType(nil), f.fruitTypes...)
}
func (f PoolFilter) Banned() []powers.PowerID {
	return append([]powers.PowerID(nil), f.banned...)
}

func (f PoolFilter) isBanned(id powers.PowerID) bool {
	for _, b := range f.banned {
		if b == id {
			return true
		}
	}
	return false
}

// AllowsStand reports whether s passes this filter's rarity allowlist and
// banlist.
func (f PoolFilter) AllowsStand(s *powers.Stand) bool {
	if s == nil {
		return false
	}
	if f.isBanned(s.ID()) {
		return false
	}
	if len(f.standRarities) == 0 {
		return true
	}
	for _, r := range f.standRarities {
		if r == s.Rarity() {
			return true
		}
	}
	return false
}

// AllowsDevilFruit reports whether d passes this filter's rarity allowlist,
// fruit-type allowlist, and banlist.
func (f PoolFilter) AllowsDevilFruit(d *powers.DevilFruit) bool {
	if d == nil {
		return false
	}
	if f.isBanned(d.ID()) {
		return false
	}
	if len(f.fruitRarities) > 0 {
		found := false
		for _, r := range f.fruitRarities {
			if r == d.Rarity() {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(f.fruitTypes) > 0 {
		found := false
		for _, t := range f.fruitTypes {
			if t == d.FruitType() {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Apply filters stands and devilFruits down to what this PoolFilter allows,
// preserving order.
func (f PoolFilter) Apply(stands []*powers.Stand, devilFruits []*powers.DevilFruit) ([]*powers.Stand, []*powers.DevilFruit) {
	filteredStands := make([]*powers.Stand, 0, len(stands))
	for _, s := range stands {
		if f.AllowsStand(s) {
			filteredStands = append(filteredStands, s)
		}
	}
	filteredFruits := make([]*powers.DevilFruit, 0, len(devilFruits))
	for _, d := range devilFruits {
		if f.AllowsDevilFruit(d) {
			filteredFruits = append(filteredFruits, d)
		}
	}
	return filteredStands, filteredFruits
}
