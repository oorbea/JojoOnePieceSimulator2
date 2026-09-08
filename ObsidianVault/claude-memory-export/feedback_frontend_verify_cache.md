---
name: feedback-frontend-verify-cache
description: "Frontend Docker verification must reuse jest cache and run targeted tests during iteration, not the full suite every time"
metadata: 
  node_type: memory
  type: feedback
  originSessionId: c2426392-7b43-42ed-90d4-f380a3d7c380
  modified: 2026-09-08T12:12:03.888Z
---

Reuse caches when verifying frontend changes in Docker, and don't run the full `pnpm jest` suite (~90s+) on every intermediate fix — run only the affected test files, then do one full-suite pass at the end before committing.

**Why:** user explicitly flagged (2026-09-08) that verification was taking too long — repeatedly running the entire 60+ suite / 1200+ test frontend suite in Docker for every small fix during iteration (e.g. fixing one lint error, one type error) wastes real wall-clock time the user is waiting on.

**How to apply:**
- Pass `--cacheDirectory=/work/.jest-cache` (or similar path on the persistent named volume, e.g. `jojo-frontend-work`) to `pnpm jest` so jest's transform cache survives across container runs, not just `node_modules`/pnpm store.
- During iteration (fixing a typecheck/lint/test failure), run `pnpm jest --ci --maxWorkers=2 --cacheDirectory=... <specific-test-file-names>` and `pnpm exec eslint <specific-changed-files>` instead of the full `pnpm lint`/`pnpm jest` — this cut a ~90s run down to ~50s even including install, and would be far faster for truly targeted single-file checks.
- Only run the full `pnpm typecheck && pnpm lint && pnpm jest` (or `pnpm test:ci`) once, as the final check before committing — this is still required per [[feedback_backend_tests_via_docker]]'s sibling norm and `norma-verificacion-docker.md`, just not on every intermediate iteration.
- The `node_modules`/pnpm-store caching via the named Docker volume (`jojo-frontend-work`) was already in place and working (install steps were already fast, ~1-2s) — the actual bottleneck was always the jest run itself, not the install.
