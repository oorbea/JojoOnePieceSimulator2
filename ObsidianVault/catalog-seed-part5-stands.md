---
title: "Seeding Part 5 (Vento Aureo/Golden Wind) Stands — hand-authored, not catalogsync"
tags:
  - project
  - jojo-onepiece-simulator
  - decision
  - i18n
  - content
---

# catalog seed — Part 5 Stands (Golden Wind)

## Status

Shipped 2026-09-28. `db/migrations/00021_seed_part5_stands.sql` — 26 new Stand rows (25 unique
wielders + the Gold Experience → Gold Experience: Requiem chain), all three locales
(es-ES/en-GB/ca-ES), stats sourced from jojowiki.com. Also fixes house-style violations in the
pre-existing `Chariot Requiem` row (translations only, stats/rarity/evolution untouched). Verified
end-to-end against a fresh local stack: goose applies cleanly, down/up round-trips clean, API
returns all 26 new rows plus the fixed Chariot Requiem text in every locale.

Same overall approach as [[catalog-seed-part4-stands]] (Part 4) — read that note for the general
pipeline/rationale; this note only covers what's new/different for Part 5.

## What's new vs. Part 4

- **One evolution chain instead of two**: `Gold Experience → Gold Experience: Requiem` (all other 24
  Stands + King Crimson are standalone rows). Per product decision, **Epitaph is folded into King
  Crimson's own description/skills, not a separate row** — it's Doppio's borrowed use of King
  Crimson's precognition, not an independent Stand in this catalogue's model.
- **A format-only fix migrated alongside new content**: `Chariot Requiem` (from 00017/00019) had
  house-style violations — es-ES header said `Usuario:` instead of `Portador:`, its skills had no
  explanations; en-GB/ca-ES had no header line at all and paragraph-length skills; ca-ES had typos
  ("Fitxa/Fletxa" inconsistency, "pròpis" for "propis"). Fixed via `UPDATE power_translations ...
  WHERE power_id = ... AND locale = ... AND EXISTS (SELECT 1 FROM powers WHERE id = ... AND
  updated_at = '<original dump timestamp>'::timestamptz)` — same admin-edit-wins guard pattern as
  00019's catalogsync updates: if an admin already touched the row since the last dump, this no-ops
  instead of clobbering their edit. Stats/rarity/name/evolution untouched on purpose. The `-- +goose
  Down` intentionally does NOT revert this fix (the old text was the buggy one — nothing worth
  rolling back to).
- **Zero wielder-mapping mistakes baked into the original task brief this time** — every wielder
  question was phrased as "verify carefully" rather than a flat guess, and 5 parallel subagents still
  caught real corrections (see below), showing the "wiki wins over the brief" instruction earns its
  keep even when the brief tries to hedge.

## Wielder corrections caught by the research step

- Sticky Fingers → **Bruno Bucciarati** (brief guessed Leone Abbacchio)
- Moody Blues → **Leone Abbacchio** (brief guessed Guido Mista) — Sticky Fingers/Moody Blues wielders
  were swapped in the brief, same failure shape as Part 4's Lock/Surface/Love Deluxe cross-swap.
- Mr. President → **Coco Jumbo** (a turtle, not "a small monkey" as the brief guessed)
- Kraft Work → **Sale** (brief guessed Formaggio, who actually wields Little Feet)
- Metallica → **Risotto Nero** (brief guessed Melone, who actually wields Baby Face)
- Rolling Stones → **Scolippi**, confirmed a genuine living Stand user (child onset), not a Stand
  bound to an inanimate statue with no human wielder as the brief's phrasing hinted — it only
  manifests as autonomous/stone-like and acts beyond Scolippi's conscious control.

## No new gotchas hit

Unlike Part 4 (enum-literal `'NULL'` bug, ca-ES header-article leak, stale-Redis-after-manual-replay
footgun — all documented in [[catalog-seed-part4-stands]]), this run reused the same fixed `gen.js`
(`statVal()` always quotes) and the subagent prompts explicitly warned against the header-article
mistake up front — the merge/validate script's regex check
(`^(Portador|Portadora|Usuari|Usuària|User):\s+(en|na|l')`) found zero violations across all 27
entries (26 new + the Chariot Requiem fix) on the first pass. The stale-Redis footgun still applies
in principle to any manual `goose down`/`up` replay done for verification — `redis-cli FLUSHALL` was
run once after the round-trip test here too, same as Part 4.

## Rarity, not sourced from any wiki

Proposed by Claude from canon power level, anchored against Part 4's reference points
(Star Platinum/Crazy Diamond/The Hand/Heaven's Door LEGENDARY; Star Platinum: The World / Killer
Queen: Bites the Dust / Chariot Requiem MYTHICAL; Killer Queen/Cream/Silver Chariot EPIC), then
corrected by the owner in review: Moody Blues COMMON (not the initial RARE guess), King Crimson
MYTHICAL (not LEGENDARY), Metallica LEGENDARY (not EPIC) — the owner's read is that Metallica's
utility (detects/tracks all metal within its territory, extremely hard to escape once caught) plus
Risotto Nero's role as a top *Squadra Esecuzioni* enforcer outweighs its "just a detection Stand"
surface reading. Full table with every Stand is in the plan file this migration was built from (not
committed — plan files are ephemeral, this vault note is the durable record if the rarity reasoning
needs revisiting).

Related: [[catalog-seed-part4-stands]], [[catalog-seed-from-prod]],
[[content-authoring-stands-devil-fruits]], [[i18n-multi-language]], [[catalog-seed-part6-stands]]
(same pipeline, Part 6)
