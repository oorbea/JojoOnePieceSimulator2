---
title: "Feature (planned): Versus by inventory + characters (not implemented)"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - frontend
  - feature
  - gameplay
  - planned
---

# Versus by inventory + characters (planned, not implemented, 2026-08-27)

**Status: planning only.** Nothing in this note is built. This is a detailed brief so that when
implementation starts, the team only needs to settle concrete technicalities, not re-derive intent.
See [[gameplay-game-modes]] ("Versus ability sources") and [[gameplay-domain-design]] (known debt on
`ports.IInventory`) for the current state of the seam this feature fills.

## Premise

Today `enums.AbilitySource.Inventory` is rejected outright by `game.NewConfig`
(`game/config.go:85-86`, `game.ErrInventoryNotSupported`, mapped to HTTP 501
`INVENTORY_NOT_SUPPORTED` in `endpoints/errors.go:86`) — there is no persisted inventory. This
feature builds that system. Once it exists, Versus with `AbilitySource=INVENTORY` has **zero random
component**: every participant composes their own Loadout by hand, every round, from what they own.

## Four slots per round

Each participant picks, at the start of **each** of the 3 Versus rounds:

1. A **Stand**
2. A **Devil Fruit**
3. A **One Piece character**
4. A **JoJo character**

The One Piece character supplies `physicalForm`, all three Haki levels (Armament, Observation,
Conqueror), and `fruitMastery`. The JoJo character supplies `hamon`, `spin`, and a new `Loadout`
stat: **`battleIQ`**.

## Battle IQ

- Type: `byte`. Stored as the raw IQ score, **0-255, deliberately unbounded** (no validation range)
  — allows intentionally absurd values (e.g. 255 for a Pucci or a Kars).
- Display: the raw number **and** its category per the standard Wechsler classification (WAIS-IV):

  | Range | Category (en) |
  |---|---|
  | <70 | Extremely low |
  | 70–79 | Borderline |
  | 80–89 | Low average |
  | 90–109 | Average |
  | 110–119 | High average |
  | 120–129 | Superior |
  | ≥130 | Very superior |

  Everything above 130 collapses into "Very superior" — there's no higher WAIS-IV band, that's
  expected.
- Effect: enters `DefaultLoadoutEvaluator.Score` (scaled — the exact factor is a technicality for
  implementation time, needs to sit sensibly alongside the other summands), which in turn drives
  `BotVoter`'s comparison.

## Selection timing

At the **start of each round**, with a timed window — not a single draft before the match starts.
Lets a participant react to what the opposing team played the previous round. This mirrors the
existing voting-window pattern (`ports` timer owned by the application layer, not the domain) but is
a new kind of window; needs its own state/transition, not a reuse of `VOTING`.

## Uniqueness rules

- Within a single round, **no item repeats within a team** — same rule as today's Stand/Devil Fruit
  uniqueness (see [[gameplay-game-modes]] "Uniqueness"), extended to the 4 slots.
- For a single participant across the match, all 4 slots must be **distinct across the 3 rounds** —
  you cannot bring the same Stand/Fruit/character twice.
- A teammate **can** reuse in round 2 whatever another teammate on the same team used in round 1 —
  the no-repeat constraint is per-round-per-team, not per-match-per-team.

## Loadout invariants → auto-correct upward, never block

`NewLoadout`'s existing hard invariants (`game/loadout.go:45-97`, see [[gameplay-game-modes]] "Hard
invariants") stay in force even here — but in `INVENTORY` mode, mastery and spin come from the
chosen character, not from a random draw, so a selection can land on a combination the invariants
would otherwise reject. Decision: **auto-correct upward**, never block the pick:

- Picking a Stand with `REQUIRES_SPIN_4` (Tusk ACT4, Ball Breaker, Soft & Wet: Go Beyond) forces
  `SpinInfinite` for that round, even if the chosen JoJo character's own `spin` is lower.
- Picking a Devil Fruit while the chosen One Piece character's `fruitMastery` is `NONE` bumps it to
  `FruitMasteryRegular`.

Consequence to keep visible in the UI: the character stops being the sole source of those two stats
when an override kicks in — the selection screen must show which stat got bumped and why.

## Insufficient inventory → blocked at lobby entry

Playing Versus-by-inventory requires **≥3 distinct** of each of the 4 kinds (3 Stands, 3 Devil
Fruits, 3 One Piece characters, 3 JoJo characters) — otherwise a full 3-round match without
repetition isn't possible. If a player doesn't meet that bar, they **cannot join** the lobby at all;
the UI must say exactly what's missing (e.g. "needs 1 more Devil Fruit").

## Character model — Class Table Inheritance

Not a single `manga`-discriminated table (unlike `stages`, which *is* single-table). Decision: real
class table inheritance —

- `characters` — shared columns: id, name, `manga`, picture/renditions (same vips pipeline as Stands
  and Stages), `rarity`.
- `jojo_characters` — FK 1:1 to `characters`, JoJo-specific stat columns.
- `one_piece_characters` — FK 1:1 to `characters`, One Piece-specific stat columns.

In the Go domain: a base `Character` entity plus `JojoCharacter` / `OnePieceCharacter` concrete
types, mirroring the DB split.

- `rarity` reuses the same 5-level enum as Stands/Devil Fruits (`COMMON → RARE → EPIC → LEGENDARY →
  MYTHICAL`, see [[gameplay-domain-design]] once the Mythical tier lands) — one shared rarity concept
  across all drawable content.
- Content is **hand-authored by admins**, same as Stands/Devil Fruits — see
  [[content-authoring-stands-devil-fruits]]. No seed data, no scraper/importer. This implies **two**
  admin CRUDs (one per manga), reusing the existing Stand/Devil Fruit CRUD pattern
  ([[admin-search-and-filters]]) as the template.
- Lobby character bans go in a section of the lobby config **separate from** the existing
  `PoolFilter` (which already bans Stands/Devil Fruits) — not folded into it. Reason: `PoolFilter`'s
  banlist already carries a flagged WebSocket-frame-size risk
  (`endpoints/game_ws_endpoints.go:24`); adding characters to the same list would make that worse. A
  separate section means its own banlist and its own UI tab.

## Gachapon boxes

Beyond Stands and Devil Fruits, gachapon boxes will also be able to yield **characters**. This is the
**only** place rarity is meant to influence a draw's probability going forward — contrast with the
rest of the game, where random ability assignment (Gauntlet, Versus-`RANDOM`) is deliberately
**uniform**, ported 1:1 from the probabilities of the original [JoJoOnePiece_Simulator
V1](https://github.com/oorbea/JoJoOnePiece_Simulator) (see [[gameplay-game-modes]] once that port
lands). Gachapon boxes are the deliberate exception: higher rarity → lower draw probability. Exact
weight table is a technicality for implementation time.

**Open questions, deliberately left open** (owner said: document the mechanism, leave the economy
open):

- Currency / how boxes are earned (post-match reward? daily? purchase?) — undecided.
- What happens on a duplicate pull — undecided.
- Whether there's a pity system / pull guarantee — undecided.

## Code anchors (so implementation starts already localized)

- `ports/inventory.go:15` — the `IInventory` port, currently adapterless.
- `enums/ability_source.go:11-14` — `AbilitySource.Inventory`.
- `game/config.go:85-86` — the rejection to remove once an adapter exists.
- `game/errors.go:70` — `ErrInventoryNotSupported`.
- `endpoints/errors.go:86` — the 501 `INVENTORY_NOT_SUPPORTED` mapping.
- `application/services/game_service_test.go:514` — `TestCreateGame_InventoryAbilitySource_Rejected`,
  the test that pins today's rejection and will need to change.

## Implied `Loadout` changes

Adding `battleIQ` touches every place `Loadout` is represented, not just the domain type
(`game/loadout.go:31`):

- `game/snapshot.go:125` (`LoadoutSnapshot`) + its snapshot/restore functions.
- `infrastructure/gamestore/redis/wire.go:138` (`wireLoadout`).
- `api/dto/game_response.go:72` (`GameLoadoutResponse`).
- The reveal slot table (`game/reveal.go`) gains a `RevealBattleIQ` slot, gated to JoJo like Spin and
  Hamon.

Related: [[gameplay-game-modes]], [[gameplay-domain-design]], [[gameplay-application-layer]],
[[content-authoring-stands-devil-fruits]], [[admin-search-and-filters]], [[game-lobby-management]]
