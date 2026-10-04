---
name: playtest-fixes-2026-10-03
description: "9-item playtest batch shipped on develop 2026-10-03 (bots UI, majority skip, +10s, rewatch sorteo, power card, Versus stage announced in sorteo); Versus round-2 'no sorteo' bug NOT reproduced"
metadata:
  node_type: memory
  type: project
  originSessionId: a614c548-9bdd-4683-8000-c556ca0db3ac
  modified: 2026-10-03T17:40:36.681Z
---

Shipped directly on `develop` (no branches), one commit per item, verified once at the end (backend + contracts + frontend all green). Full detail: vault `playtest-fixes-2026-10-03.md`.

**Open:** the owner's "Versus ronda 2: aparece un stage pero no hay sorteo" could not be reproduced with 1 human + bot. Fixed the real adjacent cause (stale previous-round `StageBanner` during ASSIGNING). If it comes back with 2+ humans, check `shouldReveal`/`assignmentSeq` and `live.*` fields not reset on `ROUND_RESOLVED`.

**Live-verified 2026-10-03 with 3 dev-login tabs (joseph/jotaro/polnareff, Gauntlet):** majority-skip + both toasts, rewatch lands at the same slot as non-skippers, +10s (VOTING and TIEBREAK, non-host has no button), participant modal (no raw `enums.`/`hakiSet`, "Ver ficha completa" opens catalogue card), add/remove bot, Reconfigure ghost-bot fix. That pass found a real bug in `useSkipNotice` (summary toast never fired: VOTING_OPENED nulls `summaryEndsAt` before the state flips) - fixed in d9be24f. Mobile 375px could not be tested (`resize_window` is a no-op here; same-origin iframe blocked) - only verified by squeezing the modal container via DOM. Hover card can't host clicks (`pointerEvents:none`), so the power-card button lives in `LoadoutModal` only.

**Browser-testing gotchas:** the first click/type after `navigate` is silently lost while the page hydrates (wait ~2s or repeat it, then verify); background tabs freeze toast timers/screenshots, so read the DOM (`[data-sonner-toast]`) instead of trusting a screenshot; new tabs inherit the shared cookie session, so log out + `/dev-login` per tab to get distinct accounts.

**Why:** typegen needs every payload registered in `cmd/typegen/registry.go` or `ws.ts` crashes the whole app at runtime ([[dev_auth_bypass_2026-09-25]] testing caught it); feature barrels that export containers break presentational tests. See also [[feedback_no_coauthor_atomic_commits]] and [[feedback_frontend_verify_cache]].
