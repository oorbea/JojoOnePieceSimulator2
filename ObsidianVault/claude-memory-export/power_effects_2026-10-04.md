---
name: power-effects-2026-10-04
description: "Power effects shipped on develop 2026-10-04 (Tusk/Soft&Wet Spin evolutions, Gomu->Nika, stat floors, family uniqueness, sorteo beats); 8 commits, vault note gameplay-power-effects.md; owner decisions and what is left open"
metadata:
  node_type: memory
  type: project
  originSessionId: ed8af644-cd12-4697-b394-d5d6a7abee89
  modified: 2026-10-03T22:34:59.637Z
---

Shipped on `develop` (not merged to main yet): a code table by power **name** (owner chose it over DB columns) makes powers raise other stats or evolve at draw time, shown as beats in the sorteo. Full rules/architecture/gotchas: `ObsidianVault/gameplay-power-effects.md`. Commits are atomic and have **no Co-Authored-By** (owner rule, also in [[feedback_no_coauthor_atomic_commits]]); worked directly on `develop`, no branches/worktrees (owner said so for this task).

**Why it came up**: vault only listed 3 Stands -> Spin INFINITE; owner asked to check V1 (`oorbea/JoJoOnePiece_Simulator` `main.cc`) for more and to propose new cases (incl. cross-manga). V1 had per-ACT Tusk spin floors + Spin-driven Tusk evolution; owner wanted all of it plus Soft & Wet -> Go Beyond (V1's typo bug fixed), Gomu -> Nika at AWAKENED, and Zoan/Hermit purple/King Crimson/Hamon floors.

**Prod bug found by reading the prod catalogue (owner ran the SQL in DBeaver - I did not touch prod)**: prod names are `Tusk: Acto 1..4`, the old code matched `"Tusk ACT4"`, so the Tusk rule never fired in prod.

**Owner decisions worth remembering**: no two stages of one evolution chain on a team (Whitesnake + C-MOON); bans -> evolve to the highest *allowed* stage the Spin covers; Nika must be revealed from Gomu like Tusk Act 4 from Act 1 (never "directly"); effects are animated at the moment of their trigger, evolution-style; Nika drawn directly raises mastery to AWAKENED. Rejected: the Battle IQ group, Pika Pika -> Hamon.

**How to apply**: renaming a rule's power in the admin silently disables its rule (only `power_effect_names_test.go` for seeded names + a startup log line guard it). Spin-driven Tusk evolution was unit-tested but not seen live (Tusk isn't in the local seed) - worth a look against prod-shaped data. Live-testing the sorteo: restrict the lobby pool via the API `poolFilter.banned` to the few powers you want, one human, both mangas; log `document.body.innerText` lines in the page to catch beats shorter than the screenshot latency.
