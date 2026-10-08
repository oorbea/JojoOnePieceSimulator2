---
title: "Catalogue catch-up 2026-10-08: Part 7/8 translations, 78 canon fruits, rarity rebalance"
tags:
  - project
  - jojo-onepiece-simulator
  - decision
  - i18n
  - content
---

# Catalogue catch-up 2026-10-08 (migrations 00024-00028)

## Status

Written and verified on a throwaway Postgres (up from 23, down to 23, up again, with fixture rows
imitating the admin-created prod rows) + backend `go build/vet/test` + frontend typecheck/lint/game
jest, all in Docker. **Not committed, not deployed.** The owner supplied the prod state as two
DBeaver CSV exports (no credentials ever entered the repo or the chat).

## Why it was needed

`catalogsync`'s snapshot of 2026-09-28 was stale: admins had since typed **all of Part 7 and Part 8**
into prod by hand, es-ES only. The repo's seeds only knew Parts 3-6 + 75 fruits, so the diff had to
be taken against a fresh prod read, not against the migrations (owner's warning: "hay poderes
metidos por admins que quizas no estan en ninguna migracion"). Prod read used: `SELECT kind, name,
rarity, locales (locale:descLen/skillCount)` for the inventory, then `id, name, description, skills`
(es-ES) for the 51 Stands lacking en-GB/ca-ES.

## What the migrations do

| File | Content |
|---|---|
| `00024_catalog_fixups_20261008.sql` | renames + Stand rarity fixes, every UPDATE guarded by old name / old rarity so an admin edit wins |
| `00025_stand_translations_and_killer_queen_jojolion.sql` | en-GB + ca-ES for 51 Stands (`ON CONFLICT DO NOTHING`, gated on the row existing) + new Stand `Killer Queen (JoJolion)` (EPIC, stats all `'NULL'`: wiki unrated) |
| `00026_seed_paramecia_fruits.sql` | 67 canon Paramecia |
| `00027_seed_zoan_fruits.sql` | 11 Zoan/Mythical Zoan (incl. canon `Batto Batto no mi`), deletes the game-only `Batto Batto no mi: Model Vampire` |
| `00028_fruit_rarity_rebalance.sql` | fruit rarities, see below |

Renames: `Tusk: Acto N` -> `Tusk: Act N`, `THE WORLD` -> `The World (Steel Ball Run)`, `Sugar Mountain's
Spring` -> `Sugar Mountain`, `Milagro Man` -> `Milagroman`, `Ozone Baby` -> `Ozon Baby`, `Tatoo You!`
-> `TATOO YOU!`, `Hermit purple` -> `Hermit Purple`, `Dirty Deeds Done Dirt Cheap: Love Train` ->
`D4C: Love Train`.

**The Tusk rename is load-bearing**: `game/power_effects.go` keys its Spin rules by exact lowercased
name ([[gameplay-power-effects]]). Code + tests + frontend fixtures moved to `tusk: act N` in the same
change; `power_effect_names_test.go`'s `prodOnlyRuleNames` lost the four Tusk entries because the
rename migration now "seeds" those names (the test fails loudly the other way round otherwise).
Ship the code and 00024 together - old code against new data (or the reverse) silently disables the
Tusk rule.

## Owner decisions

- Part 9 (The JOJOLands) out for now (still publishing, most players haven't read it). Wrecking Ball
  is not a Stand. Unnamed Stands/fruits are not seeded.
- A flashback appearance is not a new Stand, but an alternate-universe version is: two Worlds (Part 3 /
  SBR) and two Killer Queens (Part 4 / JoJolion) exist; Tusk (JoJolion) does not.
- Stand rarity: both Worlds MYTHICAL, base D4C LEGENDARY (Love Train stays MYTHICAL), High Priestess
  RARE, Ticket to Ride stays LEGENDARY.
- Fruit rarity rule: mirror the Stand proportions (~7% MYTHICAL / ~11% LEGENDARY / ~19% EPIC).
  MYTHICAL = Nika, Gura Gura, Yami Yami, Ope Ope, Uo Uo Seiryu, Magu Magu, Pika Pika, Hie Hie, Soru
  Soru (Yamata no Orochi demoted to LEGENDARY). Ancient Zoans + Mori Mori -> EPIC, Tama Tama -> RARE,
  the four new Mythical Zoans (Pegasus/Onyudo/Kirin/Ratatoskr) LEGENDARY. Before this only one fruit
  was MYTHICAL and 31/75 were LEGENDARY. Rarity does not affect the draw (uniform); it only shows in
  the UI and in the bots' `rarityBonus` (1/2/4/8).

## Pipeline used (reusable)

1. 4 research agents (canon lists per part/type, diffed against the repo's *and then* prod's names).
   WebFetch gets HTTP 402 from `onepiece.fandom.com`; `curl` against the MediaWiki API
   (`api.php?action=parse&page=...&prop=wikitext`) works. jojowiki.com works with WebFetch.
2. 5 translation agents (51 Stands from the es-ES CSV text) + 6 fruit-authoring agents (3 locales each,
   house style: 1-sentence description with no header, exactly 3 `Name: explanation` skills) + 1 agent for
   Killer Queen (JoJolion). All wrote JSON into the session scratchpad, never the repo.
3. `gen_sql.py` (scratchpad, not committed) validates (3 skills with a colon, no article after the
   `Usuari:`/`User:` header, no `;` at a line end, no `$$`/`${`) and renders the SQL with UUIDv7 ids.
   Same trust model as [[catalog-seed-part6-stands]]: Claude-authored, owner-reviewed.

## Gotchas

- Git Bash rewrites `/path` arguments for `docker.exe` (`docker cp x pgtest:/fix.sql` became
  `C:/Program Files/Git/fix.sql`); `MSYS_NO_PATHCONV=1` plus `docker exec -i ... < file` avoids it.
- The goose CLI is not in the module graph (`go run .../cmd/goose` fails on go.sum): install it with
  `cd /tmp && go install github.com/pressly/goose/v3/cmd/goose@v3.27.3` inside the backend-test
  container to apply migrations against a scratch DB.
- A "faithful" translation preserves the source's own typos (Born This Way says "Kyo Nijimura" in its
  last skill while the header says Kei) - left as is, flag for an admin edit.
- Wiki thin spots, written conservatively: Gabu Gabu (abilities "currently unknown"), Mato Mato/Fuku
  Fuku (not named in the manga but listed as canon Paramecia), Uma Uma/Albatross/Tanuki/Hound (generic
  Zoan wording), Bata Bata.
- Recount: the Paramecia agent said 92 then 93 canon named fruits; 26 present + 67 missing = 93.

Related: [[catalog-seed-from-prod]], [[catalog-seed-part6-stands]], [[content-authoring-stands-devil-fruits]],
[[gameplay-power-effects]], [[i18n-multi-language]]
