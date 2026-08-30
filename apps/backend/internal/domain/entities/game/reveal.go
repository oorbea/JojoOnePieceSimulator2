package game

import (
	"hash/fnv"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// RevealSlot names one step of the poder-a-poder reveal shown to clients
// while a Game sits in ASSIGNING, in the exact order LoadoutBuilder.Build
// draws them (see loadout_builder.go's doc comment - keep both in sync),
// with one exception: RevealHakiSet isn't a draw step at all, it's a
// synthetic slot summarizing which of the three haki types the loadout
// ends up with (see game.HakiSet), inserted right before the three
// individual level slots - the reveal tells its story as "which haki you
// have" before "how much of each", per the owner's request (2026-08-27).
type RevealSlot byte

const (
	RevealPhysicalForm RevealSlot = iota
	RevealStand
	RevealDevilFruit
	RevealFruitMastery
	RevealHamon
	RevealHakiSet
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
	{RevealHakiSet, enums.OnePiece, false},
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

// Reveal timing constants (ms). Rewritten 2026-08-30 (owner request) to
// reproduce the pacing of the original terminal-based
// JoJoOnePiece_Simulator's loadingScreen/delay (github.com/oorbea/
// JoJoOnePiece_Simulator's main.cc/loadingScreen.cc) instead of a flat
// per-slot hold - see ObsidianVault/game-match-assignment-frontend.md for
// the redesign this supersedes and gameplay-power-fx.md for the
// (not-yet-built) per-power extra hold RevealFxMaxMs reserves room for.
// The frontend mirrors this table exactly (loadout-reveal.ts) so both
// sides compute the identical RevealDuration without sharing code.
const (
	// RevealIntroMs is the lobby-wide "get ready" beat before the first
	// participant's turn begins.
	RevealIntroMs = 1100
	// RevealPlayerIntroMs is V1's "Le toca a X" beat at the start of each
	// participant's turn.
	RevealPlayerIntroMs = 1500
	// RevealNarratorMs is how long a slot's "before" narrator line (e.g.
	// "Ahora veamos si sabes usar el spin") holds before the roulette
	// starts spinning for that slot.
	RevealNarratorMs = 1200
	// RevealSpinBaseMs is one cycle of V1's loadingScreen (5*0.1s+3*0.2s =
	// 1.1s). RevealSpinCycles below decides how many cycles a given
	// spin plays, exactly like V1's generateRandomNumber(1, 2).
	RevealSpinBaseMs = 1100
	// RevealHoldScalarMs is the "after" hold for a plain scalar slot
	// (physical form, spin, hamon, fruit mastery, haki levels) - V1's
	// delay(2.5).
	RevealHoldScalarMs = 2500
	// RevealHoldFruitMs is the hold for a landed (non-empty) Devil Fruit -
	// V1's delay(5).
	RevealHoldFruitMs = 5000
	// RevealHoldStandMs is the hold for a landed (non-empty) Stand - V1's
	// delay(10).
	RevealHoldStandMs = 10000
	// RevealHoldEmptyMs is the hold for a Stand/DevilFruit slot that
	// landed on NONE - V1 gave the "no tienes fruta/no eres usuario de
	// stand" line the same 2.5s as any other scalar miss.
	RevealHoldEmptyMs = 2500
	// RevealPlayerOutroMs is the short beat between one participant's last
	// slot and the next participant's intro.
	RevealPlayerOutroMs = 800
	// RevealOutroMs is the lobby-wide dissolve into the filled-in roster
	// after the last participant's turn.
	RevealOutroMs = 3300
	// RevealFxMaxMs is the per-slot ceiling a future per-power special
	// effect (see gameplay-power-fx.md) may add to its own hold. Reserved
	// only - nothing adds to it yet, so it plays no part in RevealDuration
	// today.
	RevealFxMaxMs = 3000
)

// RevealPlayer is the minimal per-participant shape RevealDuration needs:
// whether their Loadout actually landed a Stand and/or a Devil Fruit, since
// those two slots hold for a different (longer) duration than a NONE/empty
// result, plus which of the three haki types they actually have - a
// participant with only Observation Haki must never get a roulette for
// Armament/Conqueror Haki at all (owner request, 2026-08-30: those slots
// don't exist for them, not "exist and land on NONE"). Order matters -
// RevealDuration walks players in the same order the reveal plays them in
// (join order, i.e. Game.Participants()).
type RevealPlayer struct {
	HasStand           bool
	HasDevilFruit      bool
	HasArmamentHaki    bool
	HasObservationHaki bool
	HasConquerorHaki   bool
}

// PlayerSlots is RevealSlots(mangas) further filtered down to the slots
// player actually gets: the three individual haki-level slots only appear
// for a haki type this player's Loadout actually has (see RevealPlayer's
// doc) - every other slot (including the synthetic RevealHakiSet summary)
// is unaffected, since PhysicalForm/Stand/DevilFruit/etc. always have a
// value to show (even a floor/NONE one) once their manga is selected.
// RevealDuration and the frontend's mirrored playerSlots both call this
// per participant, never the shared RevealSlots list, once a player has
// been assigned - so a lobby-wide "10 slots" only ever described a
// same-mangas lobby before this, and now varies per participant exactly
// like V1's own haki reveal (which never printed a type the roll didn't
// grant).
func PlayerSlots(mangas []enums.Manga, player RevealPlayer) []RevealSlot {
	all := RevealSlots(mangas)
	slots := make([]RevealSlot, 0, len(all))
	for _, s := range all {
		switch s {
		case RevealArmamentHaki:
			if !player.HasArmamentHaki {
				continue
			}
		case RevealObservationHaki:
			if !player.HasObservationHaki {
				continue
			}
		case RevealConquerorHaki:
			if !player.HasConquerorHaki {
				continue
			}
		}
		slots = append(slots, s)
	}
	return slots
}

// RevealSpinCycles picks how many loadingScreen-style spin cycles a given
// (participant, slot) plays - 1 or 2, mirroring V1's
// generateRandomNumber(1, 2) before every reveal. Deterministic (not
// actually random) so the backend's RevealDuration and the frontend's
// mirrored revealTimeline always agree on the same number without
// exchanging anything beyond the inputs both sides already have: the
// round/participant/slot identify a spin the same way on every client.
// Frontend mirror: loadout-reveal.ts's revealSpinCycles - keep both in
// sync.
func RevealSpinCycles(gameID GameID, roundIndex, participantIndex int, slot RevealSlot) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(gameID.String()))
	buf := []byte{byte(roundIndex), byte(participantIndex), byte(slot)}
	_, _ = h.Write(buf)
	if h.Sum32()%2 == 0 {
		return 1
	}
	return 2
}

// slotHoldMs is how long slot holds after landing for a given participant,
// honoring RevealPlayer's HasStand/HasDevilFruit for the two block slots.
func slotHoldMs(slot RevealSlot, player RevealPlayer) int {
	switch slot {
	case RevealStand:
		if player.HasStand {
			return RevealHoldStandMs
		}
		return RevealHoldEmptyMs
	case RevealDevilFruit:
		if player.HasDevilFruit {
			return RevealHoldFruitMs
		}
		return RevealHoldEmptyMs
	default:
		return RevealHoldScalarMs
	}
}

// RevealDuration is how long the reveal overlay plays for a lobby with the
// given mangas and players (in reveal order), at the given speed, before
// GameService opens voting. Deliberately NOT a pure function of mangas
// alone anymore (that was the pre-2026-08-30 design - see
// game-match-assignment-frontend.md's superseded note): the owner asked
// for V1's jugador-por-jugador pacing, where a Stand/Devil Fruit slot holds
// far longer when it actually lands a power than when it lands NONE. Both
// backend and frontend still arrive at the identical number without
// exchanging anything ad hoc, because both compute it from the same inputs
// already available on both sides: the lobby's mangas, each participant's
// own (already-assigned) Loadout, and the configured RevealSpeed - see
// GameConfigResponse.RevealSpeed and GameParticipantResponse.Loadout.
func RevealDuration(gameID GameID, roundIndex int, mangas []enums.Manga, players []RevealPlayer, speed enums.RevealSpeed) time.Duration {
	total := RevealIntroMs + RevealOutroMs
	for pi, p := range players {
		total += RevealPlayerIntroMs + RevealPlayerOutroMs
		for _, slot := range PlayerSlots(mangas, p) {
			spinMs := RevealSpinBaseMs * RevealSpinCycles(gameID, roundIndex, pi, slot)
			total += RevealNarratorMs + spinMs + slotHoldMs(slot, p)
		}
	}
	scaled := float64(total) * speed.Multiplier()
	return time.Duration(scaled) * time.Millisecond
}

// ResultDuration is how long GameService holds a Game in RESOLVING before
// advancing it (see Game.CompleteRound) - long enough for clients to read
// the round's outcome (winner, vote breakdown, coin flip) inline where the
// ballot used to be. Fixed, not host-configurable, same as the reveal.
const ResultDuration = 6 * time.Second
