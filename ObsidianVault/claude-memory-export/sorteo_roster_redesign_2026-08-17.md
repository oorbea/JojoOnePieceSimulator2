---
name: sorteo-roster-redesign-2026-08-17
description: "Sorteo reel geometry bug fixed + redesigned, haki wording, per-locale loadout text, voting roster shows only avatar+username with hover card/modal"
metadata: 
  node_type: memory
  type: project
  modified: 2026-08-17T09:20:42.651Z
  originSessionId: 487a3148-047e-4175-9c75-895c6bb671c3
---

Shipped 2026-08-17, verified live via `local-up` + `claude-in-chrome` (real Postgres/Redis stack,
real admin-created Stand/DevilFruit fixtures, real Google session) as well as full unit test suites
on both sides. 6 commits on `develop`.

- **Root cause of "the ruleta always lands blank"**: `power-roulette.tsx`'s clipping window used
  `justify="center"` on a strip taller than the window, so `translateY=0` showed the strip's middle
  row (not the top), and the old `restY` math landed past the strip's real end for any pool >3
  candidates - i.e. always, in practice. Fixed + redesigned (3-row window, overshoot-spring landing,
  edge fade, per-lane stagger). Geometry split into `lib/reel-geometry.ts`, unit-tested
  (`reel-geometry.test.ts`) so this exact invariant can't silently regress.
- Haki wording (es-ES/ca-ES only): "Haki de Armadura"/"Haki de Observación"/"Haki del Rey" (ca-ES:
  "Haki d'Armadura"/"Haki d'Observació"/"Haki del Rei").
- Backend: a loadout's Stand/DevilFruit `description`/`skills` are now re-resolved per viewer locale
  at serialization (`GameEndpoints.standTextResolver`/`devilFruitTextResolver`, mirroring the
  existing `StageTextResolver` pattern) instead of frozen to en-GB by `RepoPowerPool`. Also carries
  each participant's avatar (own upload or Google picture) through to the game snapshot.
- Voting roster (`MatchRoster`) now shows only avatar+username per participant
  (`ParticipantTile`/`ParticipantAvatar`) - full breakdown moved to a 1.5s hover card (web) /
  long-press card (native) or a tap-opened near-fullscreen `LoadoutModal`.

**Why:** full context and the exact bug mechanics are in
`ObsidianVault/game-match-assignment-frontend.md`'s "Reel geometry bug + redesign..." section
(dated 2026-08-17) - this memory is just a pointer, don't duplicate details here.

**How to apply:** if the ruleta or voting roster come up again, read that vault section first
before re-deriving anything. [[game_lobby_feature_closed_2026-08-13]]
