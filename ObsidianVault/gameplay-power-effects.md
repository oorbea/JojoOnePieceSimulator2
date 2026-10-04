---
title: "Feature: power effects (stat floors + evolutions) - shipped 2026-10-04"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - frontend
  - feature
  - gameplay
---

# Power effects: a power raises another stat, or evolves

**Status: shipped 2026-10-04**, on `develop`, in 7 atomic commits (domain, wire, DTO/contracts, reveal timing, sorteo UI, a copy fix, and the catalogue-name test). Verified live against the local stack with a restricted pool (see [[#Live verification]]).

Replaces the old `REQUIRES_SPIN_4` trait ("3 Stands force Spin INFINITE", matched by name in `power_traits.go`, deleted). That rule was the only one the vault had listed; the owner asked for the rest of V1's behaviour plus new cases, so this is the general mechanism. Rule list below is what the owner approved in the iterative session of 2026-10-03.

## The prod bug that started it

The old table matched `"Tusk ACT4"`; prod's catalogue names it `Tusk: Acto 4` (`Tusk: Acto 1..4`). **The Spin-INFINITE rule never fired for Tusk in prod.** Same class of bug V1 had with `"Soft & Wet: Go Beyond!"` (the `!` made its check dead). Names are now matched exactly after trim + lowercase + collapsing spaces, and `db/migrations/power_effect_names_test.go` fails when a rule name is in neither a seed migration nor the explicit prod-only allowlist.

## Rules

Rule storage is a **code table keyed by power name** (`game/power_effects.go`) - an explicit owner decision over adding persisted columns. Trade-off accepted: an admin renaming one of these powers in prod silently stops its rule. Guards: the catalogue test (seeded names), the prod-only allowlist (Tusk Acto 1-4, Ball Breaker, Soft & Wet: Go Beyond), and a one-time `log.Printf` from `GameService.loadDrawCatalog` listing rule names missing from the live catalogue.

**Stand spin tiers** (port of V1 `main.cc:82-126`): Tusk Acto 1 = BASIC, Acto 2 = GOLDEN, Acto 3 = GOLDEN, Acto 4 = INFINITE, Ball Breaker = INFINITE, Soft & Wet: Go Beyond = INFINITE (Soft & Wet itself has no tier).
- Spin below the stand's tier -> Spin is raised to it.
- Spin **above** the stand's tier -> the stand evolves to the highest-tier allowed descendant the Spin covers. Equal is "just right" (Acto 2 + GOLDEN stays Acto 2, as in V1). Acto 1 + GOLDEN is a genuine tie between Acto 2 and Acto 3 -> 50/50, like V1. If the Spin outgrows every *allowed* stage (the host banned the top ones), the furthest-evolved one wins deterministically in one hop, no tie-break draw (Acto 2 + INFINITE with Acto 4 banned -> Acto 3).
- Soft & Wet + INFINITE -> Go Beyond, with V1's typo bug fixed.

**Fruit family Gomu -> Nika**: `Gomu Gomu no mi` + AWAKENED mastery -> `Hito Hito no mi: Model Nika`. Nika drawn directly with lower mastery -> mastery raised to AWAKENED, and its reveal starts from Gomu (like Tusk Acto 4 starts from Acto 1). `Hito Hito no mi` (Chopper's) and `...: Model Daibutsu` are unrelated - exact match, not prefix. Fruits have no `evolves_from` in the DB; the relation lives in the code table and is attached at load time by `game.LinkFruitEvolutions` (returns copies; must run on the **unfiltered** list so a banned Gomu still parents Nika).

**Stat floors** (resolved to a fixpoint, only ever raise):

| # | Trigger | Floor |
|---|---|---|
| 2 | MYTHICAL_ZOAN / ANCIENT_ZOAN fruit | Physical Form >= MARINE_CAPTAIN |
| 3 | ZOAN fruit | Physical Form >= STRONG_FISHMAN |
| 4 | Stand `Hermit purple` | Hamon >= BASIC |
| 5 | Hamon PERFECT | Spin >= BASIC |
| 6 | Stand `King Crimson` | Observation Haki >= YONKO_COMMANDER (cross-manga) |
| 7 | Fruit `Hito Hito no mi: Model Nika` | Hamon >= ADVANCED (cross-manga) |
| 9 | Hamon PERFECT | Armament Haki >= PRIVATE (cross-manga) |
| 10 | Hamon >= ADVANCED | Physical Form >= MARINE_CAPTAIN (cross-manga) |

Cross-manga rules (a power of one manga raising a stat of the other) only apply when **both** mangas are in play. Proposed but not taken: #8 Pika Pika -> Hamon, and the whole Battle IQ group (#11/#12).

**Team uniqueness by family**: drawing a Stand/fruit removes its **whole evolution family** from that team's pool (no Whitesnake + C-MOON, no Gomu + Nika). `AvailablePowers.DrawStand/DrawDevilFruit` return the removed family as the evolution candidates, so an evolution target can never already be held by a teammate, and a banned stage is never a candidate (the pool is already filtered). `checkPoolSufficiency` counts **families**, not entries - ten stages of one chain can serve one teammate.

## Architecture

- `game/power_effects.go`: tables, `PowerEffect{Kind, Slot, From, To, CauseSlot, Cause}`, `resolvePowerEffects` (passes until stable, cap 16 -> `ErrPowerEffectsDiverged`; per pass: fruit evolution, fruit mastery floor, stand evolution, stand spin floor, then the ordered stat-floor table). `rng.IntN` is consumed **only** for an exact-tier tie, so `TestLoadoutBuilder_DrawOrder` still holds.
- `LoadoutBuilder.Build` resolves after every draw. `Loadout` carries `effects`; `Stand()/DevilFruit()` are the **final** ones, `DrawnStand()/DrawnDevilFruit()` the pre-evolution ones (found on the final power's ancestor chain by the first evolution effect's `From`).
- `NewLoadoutFromSpec` enforces the floors (`ErrPowerEffectFloorViolated`, wire code `POWER_EFFECT_FLOOR_VIOLATED`, replaces `SPIN_4_REQUIRED`); **restore skips them** (`newLoadout(spec, false)`), closing the debt in [[game-lobby-persistence]] where a changed rule table made in-flight lobbies unrestorable. Future inventory mode ([[gameplay-versus-inventory-characters]]) gets the floors for free.
- New enums `PowerEffectKind` (STAT_FLOOR, EVOLUTION) and `LoadoutSlot` replace `PowerTrait` on the wire.
- Persistence: `LoadoutSnapshot.Effects` -> `wireLoadout.Effects` (`omitempty`, **no `snapshotVersion` bump**, an unparseable effect is dropped on restore) and `powersnap.DevilFruitSnapshot.EvolvesFrom`. DTO: `GameLoadoutResponse.effects` (always `[]`, never null) and `DevilFruitResponse.evolvesFrom` (null from the catalogue endpoints). `LoadoutEffectResponse` had to be added to `cmd/typegen/registry.go` - typegen panics on an unregistered struct.

## Sorteo (reveal)

Each effect plays its own beat **at the moment its trigger shows up**, after replaying the drawn values: the loadout the stage draws is `displayLoadoutAt(final, effectsApplied)` (every later effect undone via its `from`). A stat floor is an intro (cause line, drawn value) + a hold (`EffectLevelUp`: new value in gold, "LEVELLED UP" stamp). An evolution reuses `PowerRevealCard`'s evolve beats (flash/stamp/sound), walking the **final** power's chain from the effect's `from` to its `to`.

- Anchors (`power-effects.ts`'s `effectAnchors`): after the latest, in reveal order, of the slot the effect changes, the slot that triggered it, and **the previous effect's anchor** - so applied effects are always a prefix of the list. Consequence seen live: a Nika -> Hamon floor can play *after* the King Crimson -> Observation one even though Hamon is revealed first. Intentional.
- A haki type an effect grants from nothing has no level slot (`RevealPlayerFor` / `revealPlayerFor` use the drawn presence); its effect is shown on the `hakiSet` summary.
- Durations: `RevealEffectIntroMs` 1500, `RevealEffectLandMs` 2500; an EVOLUTION adds `RevealEvolvingMs` + `RevealEvolveStepMs` per intermediate stage. A drawn fruit with a chain now gets the stand-style evolve beat on its own slot (`FruitEvolutionSteps`). Backend `game.RevealPlayerFor` and frontend `revealPlayerFor` are the single derivation of timeline inputs; a golden scenario (152 400 ms, `GameID{1}`) is pinned in both `power_effects_test.go` and `loadout-reveal.test.ts` - change both together.

## Gotchas hit while building it

- Evolution trigger semantics: `spin > stand's tier` triggers, destination is the highest allowed descendant `<=` spin. An earlier `tier > current` candidate filter left Acto 2 + INFINITE stuck when Acto 4 was banned; and without the "furthest wins when the Spin outgrows every allowed stage" rule, Acto 1 + INFINITE with Acto 4 banned took two hops depending on the tie draw.
- `TestLoadoutBuilder_MatchesV1Distributions` now draws JoJo and One Piece separately: with both mangas the cross floors deliberately shift Physical Form/Haki away from the raw V1 tables the test pins.
- In the stage, the cause line shows twice during an evolution (stage narrator + the card's `causeLine`, since the card is a Modal covering the stage) - tests use `getAllByText`.
- Frontend Docker verify from Git Bash: bind-mounting `$(pwd)` silently mounts nothing (path conversion); use `MSYS_NO_PATHCONV=1` and `C:/...`.
- A long Python/heredoc edit inside one `Bash` call broke the shell once; writing the script to a file first is reliable.

## Live verification

Lobby with the pool restricted to Gomu/Nika + King Crimson/Hermit purple, both mangas, one human, viewed in Chrome. Observed in order: a fruit evolve beat, "Nika raises your Fruit Mastery to Awakened" + "LEVELLED UP", "King Crimson raises your Observation Haki to Yonko Commander", "Nika raises your Hamon to Advanced". That run exposed the fruit slot saying "your **stand** is evolving" (fixed, own copy key). Not seen live: a **spin-driven Tusk** evolution (Tusk is prod-only, not in the local seed) - covered by unit tests, worth eyeballing against prod data. The only console errors were the pre-existing audio `AbortError: play() interrupted by pause()`.

## Open / deliberately not done

- Lobby UI sufficiency (`pool-stats.ts` `poolShortfalls`) still counts stand **entries**; stands could be family-aware client-side but the backend's check is authoritative. Fruits can't be (the DTO has no fruit `evolvesFrom` from the catalogue).
- Inventory mode should call the same resolver instead of re-implementing auto-correct (see [[gameplay-versus-inventory-characters]]).
- A rule name renamed in prod is only caught by the startup log line, not at admin-edit time.

Related: [[gameplay-game-modes]], [[gameplay-domain-design]], [[game-match-assignment-frontend]], [[game-lobby-persistence]], [[catalog-seed-from-prod]]
