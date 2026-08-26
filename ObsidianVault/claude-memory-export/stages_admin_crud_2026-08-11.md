---
name: stages_admin_crud_2026-08-11
description: "Stage admin CRUD (backend tests + full frontend) shipped end-to-end, follows Stand's pattern with 3 deliberate deviations"
metadata: 
  node_type: memory
  type: project
  originSessionId: 6e088623-0619-4a53-9e7b-424c0cdf31be
  modified: 2026-08-11T11:52:36.224Z
---

Built the full admin CRUD for the `stage` entity end to end (backend test coverage + entire
`src/features/stages/` frontend feature, `/admin/stages` route, admin hub tile), mirroring the
Stand feature's architecture, UI/UX, and SSE patterns. Details, rationale, and a caught SSE bug are
written up in `ObsidianVault/game-stage-content.md`'s "Frontend admin CRUD + remaining backend
tests" section — read that before touching Stage again.

**Why:** backend production code for Stage (service/repo/routes/pictures/translations) already
existed from an earlier tanda, but had almost no test coverage, and the frontend admin screens had
been explicitly flagged as not built in `game-realtime-transport.md`.

**How to apply:** three deviations from copying Stand verbatim, each because Stage's contract
genuinely differs — don't "fix" these back to match Stand:
- Stage's translations use a sibling `shared/lib/stage-translations.ts`, not
  `power-translations.ts` — every locale is mandatory (no skills), the opposite of Power's
  "only en-GB mandatory" rule.
- `LocaleTabs`' `requiredLocale` prop now accepts `Locale | Locale[]` (was `Locale`-only) so Stage
  can star all three tabs; Stand/DevilFruit call sites are unaffected.
- No server-side pagination for Stages — catalogue is ~19 rows; manga filter goes to the backend's
  existing `?manga=`, free-text search is client-side over the result.

Also fixed a real bug while wiring SSE: `picture-events-bridge.tsx`'s kind-dispatch `else` treated
any non-STAND/DEVIL_FRUIT event as a profile event, so Stage picture-pipeline events were silently
invalidating the wrong query. Replaced with an explicit switch.
