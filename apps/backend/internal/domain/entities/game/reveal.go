package game

import (
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// RevealSlot names one step of the poder-a-poder reveal shown to clients
// while a Game sits in ASSIGNING, in the exact order LoadoutBuilder.Build
// draws them (see loadout_builder.go's doc comment - keep both in sync).
type RevealSlot byte

const (
	RevealPhysicalForm RevealSlot = iota
	RevealStand
	RevealDevilFruit
	RevealFruitMastery
	RevealHamon
	RevealArmamentHaki
	RevealObservationHaki
	RevealConquerorHaki
	RevealSpin
)

// revealSlotOrder is RevealSlots' backbone - every slot in draw order, each
// tagged with the manga that gates it and whether it's a "block" slot
// (Stand/DevilFruit art + a stat grid, needs a longer hold) vs a plain
// scalar enum chip.
var revealSlotOrder = []struct {
	slot  RevealSlot
	manga enums.Manga
	block bool
}{
	{RevealPhysicalForm, enums.OnePiece, false},
	{RevealStand, enums.Jojo, true},
	{RevealDevilFruit, enums.OnePiece, true},
	{RevealFruitMastery, enums.OnePiece, false},
	{RevealHamon, enums.Jojo, false},
	{RevealArmamentHaki, enums.OnePiece, false},
	{RevealObservationHaki, enums.OnePiece, false},
	{RevealConquerorHaki, enums.OnePiece, false},
	{RevealSpin, enums.Jojo, false},
}

// RevealSlots lists the slots a reveal shows for a lobby playing mangas, in
// draw order - a JoJo-only lobby never gets a PhysicalForm/DevilFruit/haki
// step, and vice versa. Mirrors LoadoutBuilder.hasManga's gating exactly, so
// a Loadout this builder can produce always has one value per slot returned
// here (never a slot with nothing to show - see RevealDuration's doc for why
// that matters).
func RevealSlots(mangas []enums.Manga) []RevealSlot {
	has := make(map[enums.Manga]struct{}, len(mangas))
	for _, m := range mangas {
		has[m] = struct{}{}
	}
	slots := make([]RevealSlot, 0, len(revealSlotOrder))
	for _, s := range revealSlotOrder {
		if _, ok := has[s.manga]; ok {
			slots = append(slots, s.slot)
		}
	}
	return slots
}

// Reveal timing constants, deliberately fixed regardless of what a slot's
// draw actually landed on (even a NONE/PRIVATE floor value still plays out
// its full spin+hold) - the frontend mirrors this table exactly (see
// loadout-reveal.ts) so both sides compute the identical RevealDuration
// without sharing code. Loosely modeled on the original terminal-based
// JoJoOnePiece_Simulator's loadingScreen/delay pacing (github.com/oorbea/
// JoJoOnePiece_Simulator): ~1.1s of suspense before a reveal, several
// seconds of hold after, a longer hold for the Stand/DevilFruit art blocks
// (V1 gave those a 10s hold for its own printed description, trimmed here
// since this UI never renders a power's description - see
// game-match-assignment-frontend.md's locale-gap note).
const (
	RevealIntroMs      = 1100
	RevealSpinMs       = 1650
	RevealHoldScalarMs = 2500
	RevealHoldBlockMs  = 4000
	RevealOutroMs      = 3300
)

// RevealDuration is how long the reveal overlay plays for a lobby playing
// mangas, before GameService opens voting. A pure function of mangas alone
// (never of the actual random draws) - see RevealSlots' doc: every slot
// that appears always spins for RevealSpinMs and holds for its own kind's
// duration, whether it landed on a real power or a NONE/PRIVATE floor. That
// is what lets the backend and the frontend arrive at the same number
// without exchanging anything beyond mangas itself.
func RevealDuration(mangas []enums.Manga) time.Duration {
	total := RevealIntroMs + RevealOutroMs
	for _, s := range revealSlotOrder {
		included := false
		for _, m := range mangas {
			if m == s.manga {
				included = true
				break
			}
		}
		if !included {
			continue
		}
		total += RevealSpinMs
		if s.block {
			total += RevealHoldBlockMs
		} else {
			total += RevealHoldScalarMs
		}
	}
	return time.Duration(total) * time.Millisecond
}
