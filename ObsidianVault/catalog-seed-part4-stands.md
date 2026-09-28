---
title: "Seeding Part 4 (Diamond is Unbreakable) Stands — hand-authored, not catalogsync"
tags:
  - project
  - jojo-onepiece-simulator
  - decision
  - i18n
  - content
---

# catalog seed — Part 4 Stands (Diamond is Unbreakable)

## Status

Shipped 2026-09-28. `db/migrations/00020_seed_part4_stands.sql` — 29 new Stand rows (24 unique
wielders + the Echoes 4-act chain + the Killer Queen 3-tier chain), all three locales
(es-ES/en-GB/ca-ES), stats sourced from jojowiki.com. Verified end-to-end against a fresh local
stack: goose applies cleanly, down/up round-trips clean, API returns all 29 with correct
`evolvesFrom` chains in every locale.

## How this differs from [[catalog-seed-from-prod]]

00017/00019 are **catalogsync** output: they mirror an existing prod row (created by an admin
through the normal CRUD UI) plus fill in missing translations. This migration is the opposite
direction — **brand-new game content, hand-authored into a migration directly**, no admin ever
typed these Stands in first. [[content-authoring-stands-devil-fruits]]'s "no bulk import, admin
types every row" norm now has a second documented exception (alongside catalogsync): a
Claude-authored content migration, reviewed by the owner before merging, same as catalogsync's own
"translation authoring stays human-reviewed" principle. Images are explicitly out of scope — the
owner adds them later in prod through the normal picture-upload flow, so every row ships with
`picture_status = NONE` (the column defaults).

## Content pipeline used

1. **5 parallel subagents**, one per batch of ~6 Stands (kept each evolution chain inside a single
   agent so it had full context for describing what changes act-to-act), each independently
   researching wielder + official stat chart via WebFetch/WebSearch against jojowiki.com /
   jojo.fandom.com — explicitly told not to trust memory. Wrote one JSON file per batch to a
   scratchpad dir, not the repo.
2. **Wielder corrections caught by the research step, not by me**: the task briefs guessed wrong on
   several (The Lock/Surface/Love Deluxe wielders were cross-swapped; Ratt/Harvest/Enigma/Aqua
   Necklace wielders were flat wrong). Every agent was told "the wiki wins over this brief's guess"
   and flagged its corrections in the hand-back — none required a second pass, all were internally
   consistent (e.g. one agent independently confirmed Highway Star → Yuya Fungami while another
   caught that the brief had wrongly assigned Fungami to Aqua Necklace).
3. **One stat conflict surfaced and resolved**: Echoes Act 2 Speed — one agent's web-search snippet
   said B, its own direct wiki-page fetch said D. Re-fetched the jojowiki.com page myself after all
   batches landed; D confirmed as the infobox value.
4. **A node script merged the 5 JSONs, validated shape** (stat enum membership, all 3 locales
   present, every `evolves_from` name resolves within the set), **generated UUIDv7 ids**, and
   rendered the goose SQL — this avoided ~90 rows of hand-escaped `'` in Postgres string literals.

## Gotchas hit

- **`stand_stat`'s `'NULL'` is a literal enum VALUE, not SQL NULL.** The columns are all `NOT NULL`
  (`00001_powers_and_stands.sql`). First generator pass emitted bare SQL `NULL` for stats the wiki
  doesn't rate (Echoes Act 0's egg form; a few Stands with no meaningful Range because they're bound
  to a fixed object — Atom Heart Father, Super Fly, Achtung Baby, Earth Wind and Fire, Stray Cat) →
  `ERROR: null value in column "attack_power" ... violates not-null constraint`. Fix: always
  single-quote the stat value, including the string `'NULL'` itself.
- **A subagent literally followed the ca-ES "en"/"na" article instruction into the header line
  itself** (`Portador: en Yoshikage Kira` instead of `Portador: Yoshikage Kira`) on 6 rows (all of
  batch B: Heaven's Door, the 3 Killer Queen tiers, Aqua Necklace, Bad Company) — the instruction
  was meant for names referenced *inside* the prose, not the fixed `Portador:`/`Usuari:` label line
  itself (compare `00019`'s real Anubis/Cream/Justice rows, which never article the header name).
  Caught by grepping the merged JSON for `^(Portador|Usuari...): (en|na|l')` before generating SQL.
- **Direct-migration writes bypass the repository cache invalidation the app boot path relies on.**
  `postgres.Migrate` only calls `cache.InvalidateCatalogNamespaces` when it detects it *applied*
  something on that boot (comparing goose version before/after — see [[catalog-seed-from-prod]]).
  A `goose down` + `goose up` round-trip run manually via `docker compose exec` (to re-seed after
  fixing the header typo) leaves the DB correct but Redis still serving the stale pre-fix
  translation — the backend restart afterward reported "no migrations to run" (version already at
  20) so it never re-invalidated either. Had to `redis-cli FLUSHALL` by hand to see the fix through
  the API. Not a bug in the shipped migration (a real admin edit still invalidates normally through
  the repository layer) — purely a footgun of manually replaying a migration mid-session for
  verification.
- **`stand_stat` enum has no `NULL`-as-"not applicable" ambiguity with `INFINITE`** — kept them
  distinct per spec: `INFINITE` only when the wiki literally shows an infinity symbol (none in this
  batch), `NULL` when the wiki has no rating at all for that stat.

## Rarity, not sourced from any wiki

The wiki doesn't grade Stand rarity — this project's own tier. Proposed by Claude from canon power
level (anchored against already-seeded reference points: Star Platinum/The World LEGENDARY, Star
Platinum: The World / Wonder of U MYTHICAL, Cream/Silver Chariot EPIC, Hierophant Green/Justice
RARE), then adjusted by the owner in review before implementation: The Hand LEGENDARY (not the
initial EPIC guess), Echoes Act 1 COMMON (not RARE), Echoes Act 3 RARE (not EPIC), Heaven's Door
LEGENDARY (not EPIC), Cheap Trick COMMON (not RARE). Full table with every Stand is in the plan file
this migration was built from (not committed — plan files are ephemeral, this vault note is the
durable record if the rarity reasoning needs revisiting).

## Evolution chains

`stands.evolves_from_id`, same mechanism the pre-existing Star Platinum → Star Platinum: The World
link uses. Two chains: `Echoes Act 0 → Act 1 → Act 2 → Act 3` (Koichi Hirose — Act 0 is the
undiscovered egg form, no stats, single "dormant" skill) and `Killer Queen → +Sheer Heart Attack →
+Bites the Dust` (Yoshikage Kira — per owner's instruction, all three tiers share Killer Queen's own
stat chart rather than Sheer Heart Attack's/Bites the Dust's separate wiki infobox stats; each
tier's description/skills layer on the newly-unlocked ability while keeping the prior ones
mentioned).

Related: [[catalog-seed-from-prod]], [[content-authoring-stands-devil-fruits]],
[[i18n-multi-language]], [[catalog-seed-part5-stands]] (same pipeline, Part 5),
[[catalog-seed-part6-stands]] (same pipeline, Part 6)
