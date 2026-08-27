---
title: "Feature: Gauntlet and Versus game modes (domain layer, 2026-08-10)"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - feature
  - gameplay
---

# Gameplay: Gauntlet and Versus game modes (2026-08-10)

## Status

Domain + application layers are built (application landed 2026-08-10, right after the domain pass
— see [[gameplay-application-layer]] for the technical design). Redis lobby storage, websockets,
DTOs, and HTTP/WS routes are **not built yet** — the application layer runs today against an
in-memory `IGameStore` and a hardcoded stage catalog stub, both explicitly temporary. This note is
the gameplay rulebook: what was agreed with the owner, precisely, so it never needs re-deriving.

## Modes

Two modes, both built on the same round loop (assign → vote → resolve):

- **Gauntlet** — cooperative. 1 team, 1–10 players. Rounds = JoJo parts and/or One Piece sagas,
  played through until the squad falls or clears everything.
- **Versus** — competitive. 2 teams of 1–5 players each, **always equal size**. Exactly 3 rounds,
  each with a freshly random Stage and freshly reassigned Loadouts.

At creation, the host picks which manga(s) are in play: JoJo only, One Piece only, or both. That
selection drives which abilities get drawn (see below) and, for Gauntlet, which Stages exist to
play.

## Abilities per manga

- **JoJo**: Stand, Spin, Hamon.
- **One Piece**: Physical Form, Devil Fruit, Fruit Mastery, and all three Haki types (Armament,
  Observation, Conqueror).
- **Both**: everything above.

"No stand" and "no devil fruit" are both valid draws — a player can end up with neither ability
from their manga's pool. Whether that happens (and how skewed towards rarity/rare-power-lucky
draws it is) is controlled entirely by the weighting adapter behind
`ports.IAssignmentWeights` — the domain has no opinion on the actual probabilities, only on the
invariants below.

## Hard invariants (enforced regardless of how a Loadout was produced)

1. **Fruit before mastery, and they're coupled**: a Devil Fruit is drawn first; its Fruit Mastery
   is drawn second and depends on the result. No fruit ⇒ Fruit Mastery must be `NONE`. Any fruit
   (any `FruitType`) ⇒ Fruit Mastery must be at least `REGULAR` — never `NONE` with a fruit
   present.
2. **Spin and Hamon are independent of the Stand and of each other.** A player can have all three,
   any two, just one, or none — e.g. a pure Hamon user with no Stand is valid, same as a Stand user
   with no Spin/Hamon at all.
3. **Exception**: a small, named set of Stands force Spin to `INFINITE` (spin level 4) regardless
   of what was drawn: **Tusk ACT4**, **Ball Breaker**, **Soft & Wet: Go Beyond**. This is modeled
   as a `PowerTrait` (`REQUIRES_SPIN_4`), matched by Stand name today — see
   [[gameplay-domain-design]] for why, and what replaces the name match later.

These are re-checked by `Loadout` construction itself, not just by the builder that produces one —
so any future ability source (inventory-based Versus, admin-forced test loadouts, ...) gets them
for free.

## Uniqueness

A Stand or Devil Fruit **never repeats within the same team** in the same game (Gauntlet) or the
same round (Versus). The rival team in Versus is a separate draw pool and can repeat whatever the
first team got. Bots draw from the same pool as humans and count towards this uniqueness.

## Voting

Every round is decided by a vote among the relevant participants:

- **Gauntlet**: the whole squad votes `SURVIVE` or `FALL` for the current Stage. Absolute majority
  of `FALL` ends the run in defeat immediately, at that round. Absolute majority of `SURVIVE`
  advances to the next Stage. Clearing every Stage (all parts/sagas selected for the game) is
  victory.
- **Versus**: every participant votes for whichever team (including their own) they think wins the
  round — no self-vote restriction, honesty is assumed since both teams are equal size. Whichever
  team gets the majority of votes wins the round. After 3 rounds, whichever team won more rounds
  wins the match (a tie across 3 rounds is structurally impossible, since each round always
  resolves to a definite, non-tied winner — see tiebreak below).

**Voting window**: up to 30s (a timer owned entirely by the application layer, not the domain), but
closed **as soon as every connected human has voted** — nobody waits out the rest of the window
once nothing is left to collect. A participant who disconnects, or simply doesn't vote in time,
counts as a **null vote**: the round is resolved from the votes that *were* emitted, never from the
full player count. Votes can be freely changed until the window closes — the last cast vote is the
one that counts. Bots vote immediately when the window opens (see Bots below), so they never block
the early-close check.

**Ties and coin flips**: a tie (including the zero-votes case) opens exactly one revote window.
Ties can be changed freely, same as the first vote. If it's still tied after the revote, the
round is decided externally: today, a 50/50 coin flip; later, an LLM call. Both live behind the
same seam (`ports.ITiebreaker`) so swapping the implementation touches nothing else.

**2026-08-26**: the revote window now starts from a genuinely empty ballot — `Game.CloseVoting`
resets the round's `Ballot` (and re-casts every bot's vote for it) the moment the first tie opens
`TIEBREAK`, instead of leaving every vote from the tied round still on file. Before this, a revote
opened already reading "everyone voted" and the very first changed vote resolved the whole thing —
see [[game-vote-buttons-2026-08-26]] for the fix and why.

## Host

The game's creator is the host: they set the config (mode, manga selection, ability source, team
size, bot toggle) and are the only one who can start the game or abort it. If the host disconnects
mid-game, the host role is reassigned **randomly** to another connected human participant. If no
connected human remains, the game aborts.

## Aborting

A game aborts (no winner, no loser) when:

- No connected human participant remains (bots alone don't keep a game alive).
- The host explicitly cancels it.
- In Versus, either team drops to zero players (human or bot).

## Bots

Bots only exist in Versus, only to fill uneven teams (never in Gauntlet). A bot:

- Receives a Loadout under the exact same rules and weights as a human, and counts towards the
  no-repeat-within-team uniqueness rule.
- Votes automatically, by comparing the aggregate power of each team's current Loadouts (a
  heuristic scoring function, swappable without touching the game mode or the state machine) and
  voting for whichever side scores higher.

## Versus ability sources

Two submodes for how Versus draws abilities, selected per game:

- `RANDOM` — same random-draw logic as Gauntlet (weights + uniqueness), reassigned every round.
- `INVENTORY` — pulls from a player's owned-power inventory instead. **Not implemented.** There is
  no persisted inventory (no schema, no unlock/gacha flow) yet — selecting this submode is rejected
  outright (`game.ErrInventoryNotSupported`) until that system exists. The domain design already
  has the seam for it (`ports.IInventory`, `enums.AbilitySource`) so building it later doesn't
  require touching the game/round/voting logic at all. Full planned design (4-slot per-round
  selection, characters, gachapon boxes, Battle IQ): [[gameplay-versus-inventory-characters]].

## Deliberately not done here

- Stage catalog content (which JoJo parts, which One Piece sagas exist, at what granularity — the
  owner chose **sagas**, ~11 of them, not the finer ~31-arc breakdown) is meant to be admin-managed
  CRUD data. There is still no schema/migration/admin CRUD for it — the application layer runs
  today against a **hardcoded stub** (`infrastructure/game.StaticStageCatalog`) so the feature is
  runnable end-to-end; see [[gameplay-application-layer]].
- Run-based progression between Gauntlet rounds (leveling up a surviving squad) — there's a named,
  empty hook for it (`GauntletMode.afterRound`) but no behavior yet.
- Redis lobby storage and websocket-driven realtime voting/assignment — the application layer runs
  today against an **in-memory** `ports.IGameStore` and exposes no transport at all (no HTTP/WS
  routes). See [[gameplay-application-layer]] and [[backend-contract]] for the current state.

**2026-08-14**: a round that (re)assigns Loadouts (every Versus round; Gauntlet's first) no longer
opens voting immediately after — `GameService.scheduleRevealDelay` holds the Game in `ASSIGNING`
until the frontend's sorteo overlay would plausibly have finished, so voting can never open before
anyone has seen their loadout. Rounds that don't reassign (Gauntlet after round 1) still open voting
immediately, since there's nothing new to reveal. See [[gameplay-application-layer]] for the
mechanism and [[game-match-assignment-frontend]] for the frontend half.

Related: [[gameplay-domain-design]] (technical design, patterns, file map), [[backend-contract]],
[[gameplay-application-layer]], [[game-match-assignment-frontend]], [[ADR]]
