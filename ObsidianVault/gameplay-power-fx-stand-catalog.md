---
title: "Per-power FX: curated Stand catalog (planned, 2026-09-03)"
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
  - feature
  - planned
  - design
---

# Per-power FX: curated Stand catalog (planned, 2026-09-03)

**Status: planning only, nothing here is built.** This is the owner's full curated list for
[[gameplay-power-fx]]'s registry — the "Seed catalog" section there had three placeholder examples
(Gomu Gomu no Mi, Holy's Stand, The World); this note replaces those Stand entries with the real,
complete list the owner dictated, kept in its own note because it's long enough to want its own
edit history separate from the mechanism/constraints note it plugs into. Read
[[gameplay-power-fx]] first for the shape this feeds (`PowerFx`, `cardEffect`/`avatarEffect`,
`RevealFxMaxMs`, the `.web`/`.native` split, reduced-motion rules) — this note only adds *content*
to that registry, not new mechanism.

## Sound: every entry here also gets a war-cry, including the ones below

**Owner decision (2026-09-03), not yet reflected in [[gameplay-power-fx]]'s `PowerFx` sketch**:
every Stand in this table gets a sound effect on top of its visual — normally the Stand's own
battle cry/war shout (e.g. "ORA ORA ORA", "MUDA MUDA MUDA"), one per entry below. **The actual audio
assets don't exist yet** — the owner will provide them when this feature is actually implemented,
not now. Until then this is a placeholder requirement, not a blocker on writing the registry shape:
`PowerFx` should grow an optional `soundEffect` field (asset id/path, TBD) alongside `cardEffect`/
`avatarEffect` when this is built.

**Tusk Act 1 is the one exception**: it gets *only* the sound (its own scream/cry), no visual effect
at all. Every other entry below gets both a visual and its own war-cry.

## The catalog

| Stand | Visual effect (owner's idea) | Sound |
|---|---|---|
| Star Platinum | Time stops | War-cry (TBD asset) |
| The World | Time stops | War-cry (TBD asset) |
| Holy's Stand | Brambles/thorns grow around the participant's profile picture, picture turns black-and-white | War-cry (TBD asset) |
| Crazy Diamond | Profile picture visibly breaks apart/demolished, then restores itself | War-cry (TBD asset) |
| Killer Queen (with Bites the Dust) | Explosion effect with a time-rewind | War-cry (TBD asset) |
| Gold Experience | Trees and plants grow around the picture | War-cry (TBD asset) |
| Gold Experience Requiem | "Infinite regression" effect | War-cry (TBD asset) |
| King Crimson | Time-skip effect | War-cry (TBD asset) |
| Chariot Requiem | Souls leaving the picture | War-cry (TBD asset) |
| Stone Free | Profile picture visibly frays/unravels into threads | War-cry (TBD asset) |
| Weather Report | Storm clouds | War-cry (TBD asset) |
| Weather Report — Heavy Weather | Rainbow, snails coming out | War-cry (TBD asset) |
| Whitesnake | A disc/record comes out of the profile picture | War-cry (TBD asset) |
| C-MOON | Everything flips upside down for a few seconds, with a transition in/out | War-cry (TBD asset) |
| Made in Heaven | Everything starts accelerating, the sun sweeps overhead faster and faster | War-cry (TBD asset) |
| Tusk Act 1 | **None — sound only** | Tusk's own scream (fan-made, TBD asset) |
| Tusk Act 2 | Profile picture spins/rotates | War-cry (TBD asset) |
| Tusk Act 3 | Holes open up, profile picture exits through a different, random hole | War-cry (TBD asset) |
| Tusk Act 4 | Breaks the "fourth wall", faces the viewer directly | War-cry (TBD asset) |
| Ball Breaker | A very large spin effect plus visible aging | War-cry (TBD asset) |
| Dirty Deeds Done Dirt Cheap | Disappears behind something — a curtain or wall-like cover | War-cry (TBD asset) |
| D4C Love Train | A beam of light hits the participant's profile picture | War-cry (TBD asset) |
| Soft & Wet | Bubbles | War-cry (TBD asset) |
| Soft & Wet — Go Beyond | Bubbles, with a spin | War-cry (TBD asset) |
| Paisley Park | A GPS/map-like effect, spreading out across the lobby | War-cry (TBD asset) |
| Wonder of U | Many Wonder of U figures start appearing across the screen, everything darkens with a calamity aura | War-cry (TBD asset) |

## Devil Fruits — now catalogued too

**Closed (2026-09-03, later same day)**: the owner's Devil Fruit list is in
[[gameplay-power-fx-devil-fruit-catalog]] now, same table shape as this note, sound-effect decision
included. Two of its entries are group-level ("every Ancient Zoan", "every other Mythical Zoan")
rather than per-fruit — see that note's own section on why that doesn't fit the current
`FRUIT_TYPE_FX_FALLBACK` sketch as-is.

## How this plugs into the registry sketch

Each row above becomes one `POWER_FX` entry keyed by the Stand's id, per [[gameplay-power-fx]]'s
`Record<string /* power id */, PowerFx>` shape. Most entries here only specify `cardEffect` (the
`PowerRevealCard` treatment) — whether each one *also* carries an `avatarEffect` (visible on the
participant's avatar for the rest of the sorteo, per [[gameplay-power-fx]]'s "Where does the effect
show?" answer) is an implementation detail to settle per-effect when this is actually built, not
decided here. Same for `extraHoldMs`: several of these (time-stop, time-skip, the C-MOON flip) are
strong candidates for adding to their slot's hold, capped at `RevealFxMaxMs`, but no numbers are
picked yet.

Related: [[gameplay-power-fx]], [[gameplay-power-fx-devil-fruit-catalog]],
[[game-match-assignment-frontend]], [[gameplay-game-modes]], [[gameplay-domain-design]].
