---
name: sorteo-v1-pacing-2026-08-30
description: "Sorteo rewritten to V1's jugador-por-jugador pacing + big power cards + synced skip, per-power FX planned only"
metadata: 
  node_type: memory
  type: project
  originSessionId: cf791089-d705-4cee-95a1-0210ee8886ab
  modified: 2026-08-30T10:16:04.193Z
---

Rewrote the power-assignment (sorteo) phase to replay `JoJoOnePiece_Simulator` V1's terminal
pacing and copy: one participant's full turn at a time (not all lanes spinning simultaneously
like the 2026-08-17 redesign), a full-screen `PowerRevealCard` when a Stand/Devil Fruit lands
(art+description+skills, extracted `PowerBlock` shared with `loadout-modal.tsx`), a
`RevealNarrator` band with V1's before/after lines translated into all three locales, a
per-lobby `RevealSpeed` config (Relaxed/Normal/Swift), and a synchronized skip
(`MarkRevealReady`/`REVEAL_READY`) any connected human can trigger — not host-only.

Backend `game.RevealDuration` deliberately stopped being a pure function of `mangas` alone; it now
also depends on each participant's landed Stand/DevilFruit and `RevealSpeed`. `RevealSpinCycles`
(FNV-1a hash) keeps backend/frontend spin-cycle counts deterministic without exchanging anything
extra — mirrored bit-for-bit in `loadout-reveal.ts`.

Explicitly **not** implemented: per-power special visual effects (Gomu Gomu no Mi bounce, Holy's
Stand greyscale+brambles, The World time-stop). Documented as a planning-only TODO in
[[gameplay-power-fx]] per the owner's explicit request — curated registry + rarity/fruitType
fallback, shown on the reveal card and the sorteo-scoped avatar only, capped extra hold time.

Full technical writeup: `ObsidianVault/game-match-assignment-frontend.md`'s 2026-08-30 section.
Backend+frontend both verified green via Docker (`go test ./...`, `pnpm typecheck && lint &&
jest --maxWorkers=2`). Six atomic commits, no co-author trailer per [[feedback_no_coauthor_atomic_commits]].
