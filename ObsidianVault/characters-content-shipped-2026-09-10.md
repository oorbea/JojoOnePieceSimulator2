---
title: "Feature: JoJo/One Piece character content + battleIQ (2026-09-10)"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - frontend
  - feature
  - gameplay
---

# Character content + battleIQ (Phase 1, 2026-09-10)

Phase 1 of [[gameplay-versus-inventory-characters]], deliberately scoped down to *content + one
new `Loadout` stat only* — no persisted inventory, no gachapon, no per-round 4-slot selection UI,
and the existing `AbilitySource=INVENTORY` rejection (`game/config.go`) was left untouched. Result:
admins can author JoJo/One Piece characters end to end (incl. the vips picture pipeline), players
can browse them in a public read-only catalogue, and `battleIQ` is a real stat on `Loadout`,
already assigned in Gauntlet and Versus-random.

## `battleIQ`

- Value object `game.BattleIQ{value byte; present bool}` — absent-vs-present-zero matters (0 is a
  legitimate, absurd score), so it's not a bare `byte`. Present only for JoJo-manga lobbies.
- Drawn by weighted WAIS-IV band (`2/7/16/50/16/7/2` frequencies for Extremely-low through
  Very-superior), then, only within the top band, exponential decay with a 30-point half-life
  (`weight(x) = 0.5^((x-130)/30)`) so 255 is roughly 25× less likely than 130 rather than equally
  likely — every other band is narrow enough to just sample uniformly.
- Scored 0-6 by band in `DefaultLoadoutEvaluator`, the same sublinear band→score shape
  `standStatScore` already uses for Stand stats — never the raw 0-255 value (a 255 would otherwise
  be worth ~5× the rest of the score combined).
- WAIS-IV category label is derived **client-side** (`shared/lib/battle-iq.ts`); only the raw
  number crosses the wire (`GameLoadoutResponse.BattleIQ *int`). `enums.battleIQCategory.*` is the
  one `enums.*` i18n namespace with no generated enum behind it — noted in that file in case the
  backend ever exposes the band itself.
- Carried through every place a `Loadout` is represented: snapshot, Redis wire encoding (regression
  test pins that a present-0 score survives an encode/decode round trip — `omitempty` on a bare
  `*byte` would otherwise eat it), the API DTO, and reveal (`RevealBattleIQ`, appended as the last
  reveal ordinal on both backend and frontend — inserting instead of appending would have
  desynced every later slot's reveal-spin cycle on both sides).

## Character content — CTI, separate from `powers`

`characters` + `character_translations` (description-only, en-GB mandatory — the Stands/Devil
Fruits rule, not Stages' all-three-mandatory rule) + `jojo_characters`/`one_piece_characters`
subtype tables, discriminated by `manga` the same way `power_kind` discriminates
`stands`/`devil_fruits`. Five stat enums (`haki_level`, `physical_form`, `fruit_mastery`,
`hamon_level`, `spin_level`) existed in Go since the original game-domain work but had never been
persisted anywhere — this shipped their first Postgres types.

Two admin CRUDs (`JojoCharacterService`/`OnePieceCharacterService`, `/jojo-characters` +
`/one-piece-characters`), a shared `characterPicturePublisher` for the picture worker (both kinds
share `CharacterID`, unlike Stand/DevilFruit's `PowerID` split), and `enums.PictureSubjectKind`
gained `JOJO_CHARACTER`/`ONE_PIECE_CHARACTER` appended at the end (same append-not-insert rule as
every other ordinal-sensitive enum in this codebase).

## Frontend: one feature, one screen, manga-exclusive filter

One `features/characters` slice covers both kinds — the two only really diverge in their stat
descriptor (`lib/character-stats.ts`'s `JOJO_STAT_ROWS`/`ONE_PIECE_STAT_ROWS`), so
`character-stat-block.tsx` is the *only* component that knows which kind it's rendering;
`characters-screen.tsx`, `character-card.tsx`, `character-detail.tsx` are all generic over it.

The manga filter is **exclusive** (JoJo or One Piece, default JoJo, no "all") — unlike Stages'
manga filter, which is just a `WHERE` clause against one table. A JojoCharacter and an
OnePieceCharacter are two separate backend endpoints with two independent keyset paginations, so
an "all" option would mean merging two cursors; not worth it for Phase 1. `CharactersContainer`
mounts both kinds' list queries and both kinds' `react-hook-form` instances unconditionally
(conditionally calling hooks isn't legal) and only the active one is `enabled`/rendered/submitted.

The create/edit form asks **which manga first**, as its own required step with no default in
create mode; the field set below it (JoJo's hamon/spin/battleIq vs One Piece's physicalForm/3
hakis/fruitMastery) swaps entirely based on that answer, and in edit mode the manga is locked
(changing it would mean destroying the row's subtype half). Same pattern `stage-form-modal.tsx`
uses for a Stage's manga, applied here to *which stats* the form even shows instead of just a
tag on the row.

## Related

[[gameplay-versus-inventory-characters]] (the parent brief — still planning-only past this
slice), [[catalogo-publico-stands-devil-fruits-stages]] (the `readOnly` discriminated-union
pattern this catalogue reuses), [[gameplay-game-modes]], [[gameplay-domain-design]].
