---
title: "Seeding Part 6 (Stone Ocean) Stands — hand-authored, not catalogsync"
tags:
  - project
  - jojo-onepiece-simulator
  - decision
  - i18n
  - content
---

# catalog seed — Part 6 Stands (Stone Ocean)

## Status

Shipped 2026-09-28. `db/migrations/00022_seed_part6_stands.sql` — 26 new Stand rows (24 unique
wielders + the Weather Report → Weather Report: Heavy Weather 2-tier chain + the Whitesnake →
C-MOON → Made in Heaven 3-tier chain), all three locales (es-ES/en-GB/ca-ES), stats sourced from
jojowiki.com. No pre-existing row needed a format fix this time (grepped all prior migrations for
every name first — clean). Verified end-to-end against a fresh local stack: goose applies cleanly,
down/up round-trips clean, API returns all 26 rows with correct `evolvesFrom` chains (including the
3-deep one) in every locale.

Same overall approach as [[catalog-seed-part4-stands]]/[[catalog-seed-part5-stands]] — read those
for the general pipeline/rationale; this note only covers what's new/different for Part 6.

## What's new vs. Part 4/5

- **First 3-tier evolution chain**: `Whitesnake → C-MOON → Made in Heaven` (Enrico Pucci). Unlike
  the Killer Queen tiers (Part 4, shared stats across all 3) or Gold Experience (Part 5, single
  hop), all three of these have their own distinct wiki stat charts — verified independently rather
  than reused. `Made in Heaven`'s speed is a genuine wiki-infobox `∞`, encoded as `'INFINITE'`.
- **A 2-tier chain with a stat-sharing call**: `Weather Report → Weather Report: Heavy Weather`.
  Heavy Weather's own jojowiki infobox exists but every cell in it is an unrated `?` — treated the
  same as Part 4's Killer Queen precedent (evolved form has no real chart of its own → reuse the
  base form's exact stat chart) rather than encoding a chart of literal `?` marks as all-`NULL`.
- **Several Stands are non-combat/"reading-only" this round, more than any prior part** — Burning
  Down the House, Jail House Lock, Dragon's Dream, Sky High, Green, Green Grass of Home, Boiling
  Water Stand all have some or all of their 6 stats genuinely unrated (`∅`/`?` on the wiki, not
  "unknown to the researcher") → correctly literal `'NULL'`, not a guess.
- **One canonically unnamed Stand**: "Boiling Water Stand" (as the user wrote it) has no dedicated
  wiki infobox at all — it's listed under jojowiki's/fandom's unnamed-Stands catalogue. Wielder
  recorded as **Enrico Pucci** (he's the one who activates the DISC in-story to kill F.F.), all 6
  stats `'NULL'` (no rating exists anywhere for it, confirmed via multiple sources, not a guess).
- **One dual-role Stand**: Foo Fighters is simultaneously the Stand and its own user — the wiki's
  own "User" field literally names the Stand itself (as the plankton colony "F.F."), it isn't
  wielded by a separate human. Recorded `wielder: "F.F."`, with the dual-role nuance explained in
  the description prose rather than the wielder field, to keep that field name-shaped.
- **Zero wielder-mapping mistakes baked into the original task brief this time either** — every
  wielder was phrased as "verify" with no guessed name at all (learned from Part 5's "hedged guess
  still gets corrected" finding) — subagents still surfaced two genuinely nuanced cases worth
  recording (below), neither of which was a brief error, just canon nuance.

## Wielder nuances surfaced by the research step

- **Foo Fighters → F.F.** (the Stand is its own user; no separate human wielder in the wiki's own
  field) — see above.
- **Survivor → Guccio**, not the anonymous original "Mountain Lodge Owner" the Stand disc first
  belonged to — Guccio is the one who actually wields/uses it in the story, used as the catalogue's
  `wielder`.
- **Boiling Water Stand → Enrico Pucci** (activates the DISC, not the disc's unrevealed original
  owner) — see above.

## No new gotchas hit (mechanically)

Same fixed `gen.js` (`statVal()` always quotes, including `'NULL'` and `'INFINITE'`), zero
header-article violations across all 26 entries (regex check `^(Portador|Portadora|Usuari|
Usuària|User):\s+(en|na|l')` clean on first pass) — same discipline as Part 5. One new soft-quality
pass this round: ~19 skill lines came back over the ~140-char sanity threshold (still valid
single-line "Nombre: explicación" entries, just wordier than house style wants) — trimmed by hand
before generating SQL, no wiki-fact content lost, purely tightening prose.

## Rarity, not sourced from any wiki

Proposed by Claude from canon power level, anchored against Part 4/5's reference points (Star
Platinum/Crazy Diamond/Gold Experience/Metallica LEGENDARY; SP:TW/KQ:BtD/GER/King Crimson/Chariot
Requiem MYTHICAL; Killer Queen/Sticky Fingers/Silver Chariot EPIC; Moody Blues/Echoes Act 1
COMMON), then corrected by the owner in review — 13 of the initial 26 proposals were adjusted:

| Stand | Initial guess | Owner's call |
|---|---|---|
| Burning Down the House | RARE | COMMON |
| Weather Report | EPIC | LEGENDARY |
| Weather Report: Heavy Weather | LEGENDARY | MYTHICAL |
| Diver Down | RARE | EPIC |
| Whitesnake | LEGENDARY | EPIC |
| Survivor | RARE | COMMON |
| Yo-Yo Ma | RARE | COMMON |
| Green, Green Grass of Home | COMMON | EPIC |
| Manhattan Transfer | RARE | COMMON |
| Highway to Hell | COMMON | RARE |
| Marilyn Manson | RARE | COMMON |
| Bohemian Rhapsody | EPIC | LEGENDARY |
| Under World | EPIC | RARE |

Notably the owner rates Whitesnake (the DISC-extraction Stand) a full tier *below* its own
evolution C-MOON kept at LEGENDARY, rather than moving both up together as Claude's initial anchor
logic would have — the read is that Whitesnake's own combat utility is middling despite what it
unlocks later. Full table with every Stand is in the plan file this migration was built from (not
committed — plan files are ephemeral, this vault note is the durable record if the rarity reasoning
needs revisiting).

## Evolution chains

`stands.evolves_from_id`, same mechanism as every prior part. Two chains this round:
`Weather Report → Weather Report: Heavy Weather` (Wes Bluemarine — Heavy Weather shares Weather
Report's own stat chart per the reasoning above) and `Whitesnake → C-MOON → Made in Heaven` (Enrico
Pucci — three independently-stated stat charts, verified live as MYTHICAL's `evolvesFrom` correctly
resolves two levels deep through the API in every locale).

Related: [[catalog-seed-part5-stands]], [[catalog-seed-part4-stands]], [[catalog-seed-from-prod]],
[[content-authoring-stands-devil-fruits]], [[i18n-multi-language]]
