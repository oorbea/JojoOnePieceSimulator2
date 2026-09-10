package ports

import (
	"reflect"
	"strings"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// TestCanonical_FieldCount guards against the exact drift Canonical() exists
// to prevent: a field added to a *Filters struct without extending its
// Canonical() method. Canonical() joins one segment per field with "|", so
// segment count must always equal the struct's field count - add a field to
// StandFilters/DevilFruitFilters/StageFilters without touching Canonical()
// and this fails immediately, instead of silently producing two different
// filter combinations that hash to the same cache key / cursor fingerprint.
func TestCanonical_FieldCount(t *testing.T) {
	rarity := enums.Common
	attackPower := enums.A
	evolvesFrom := "Star Platinum"
	search := "star"
	fruitType := enums.Paramecia
	manga := enums.Jojo
	hamon := enums.HamonBasic
	spin := enums.SpinBasic
	battleIQ := byte(130)
	physicalForm := enums.PhysicalFormPrivate
	hakiLevel := enums.HakiPrivate
	fruitMastery := enums.FruitMasteryRegular

	cases := []struct {
		name      string
		filters   interface{ Canonical() string }
		numFields int
	}{
		{
			name: "StandFilters",
			filters: StandFilters{
				Rarity: &rarity, AttackPower: &attackPower, Speed: &attackPower,
				AttackRange: &attackPower, Endurance: &attackPower, Precision: &attackPower,
				Potential: &attackPower, EvolvesFrom: &evolvesFrom, Search: &search,
			},
			numFields: reflect.TypeOf(StandFilters{}).NumField(),
		},
		{
			name:      "DevilFruitFilters",
			filters:   DevilFruitFilters{Rarity: &rarity, FruitType: &fruitType, Search: &search},
			numFields: reflect.TypeOf(DevilFruitFilters{}).NumField(),
		},
		{
			name:      "StageFilters",
			filters:   StageFilters{Manga: &manga, Search: &search},
			numFields: reflect.TypeOf(StageFilters{}).NumField(),
		},
		{
			name: "JojoCharacterFilters",
			filters: JojoCharacterFilters{
				Rarity: &rarity, Hamon: &hamon, Spin: &spin, BattleIQ: &battleIQ, Search: &search,
			},
			numFields: reflect.TypeOf(JojoCharacterFilters{}).NumField(),
		},
		{
			name: "OnePieceCharacterFilters",
			filters: OnePieceCharacterFilters{
				Rarity: &rarity, PhysicalForm: &physicalForm, ArmamentHaki: &hakiLevel,
				ObservationHaki: &hakiLevel, ConquerorHaki: &hakiLevel, FruitMastery: &fruitMastery, Search: &search,
			},
			numFields: reflect.TypeOf(OnePieceCharacterFilters{}).NumField(),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := strings.Count(c.filters.Canonical(), "|") + 1
			if got != c.numFields {
				t.Errorf("Canonical() has %d segments, want %d (one per struct field) - "+
					"a field was added to %s without extending its Canonical() method",
					got, c.numFields, c.name)
			}
		})
	}
}

// TestCanonical_DiffersPerField locks the actual anti-drift property:
// changing any one field's value must change the canonical string, or two
// distinct filter combinations would collide onto the same cache key /
// cursor fingerprint.
func TestCanonical_DiffersPerField(t *testing.T) {
	rarity := enums.Common
	otherRarity := enums.Legendary
	search := "star"

	base := StandFilters{Rarity: &rarity, Search: &search}
	changed := StandFilters{Rarity: &otherRarity, Search: &search}
	if base.Canonical() == changed.Canonical() {
		t.Error("StandFilters.Canonical() unchanged after changing Rarity")
	}

	baseFruit := DevilFruitFilters{Rarity: &rarity}
	changedFruit := DevilFruitFilters{Rarity: &otherRarity}
	if baseFruit.Canonical() == changedFruit.Canonical() {
		t.Error("DevilFruitFilters.Canonical() unchanged after changing Rarity")
	}

	mangaA := enums.Jojo
	mangaB := enums.OnePiece
	baseStage := StageFilters{Manga: &mangaA}
	changedStage := StageFilters{Manga: &mangaB}
	if baseStage.Canonical() == changedStage.Canonical() {
		t.Error("StageFilters.Canonical() unchanged after changing Manga")
	}

	iqA := byte(130)
	iqB := byte(200)
	baseJojo := JojoCharacterFilters{BattleIQ: &iqA}
	changedJojo := JojoCharacterFilters{BattleIQ: &iqB}
	if baseJojo.Canonical() == changedJojo.Canonical() {
		t.Error("JojoCharacterFilters.Canonical() unchanged after changing BattleIQ")
	}

	formA := enums.PhysicalFormPrivate
	formB := enums.PhysicalFormYonkoPlus
	baseOnePiece := OnePieceCharacterFilters{PhysicalForm: &formA}
	changedOnePiece := OnePieceCharacterFilters{PhysicalForm: &formB}
	if baseOnePiece.Canonical() == changedOnePiece.Canonical() {
		t.Error("OnePieceCharacterFilters.Canonical() unchanged after changing PhysicalForm")
	}
}
