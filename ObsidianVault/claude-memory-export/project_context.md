---
name: project_context
description: "JojoOnePieceSimulator2 is a solo hobby/learning project, now live in prod via CI/CD, with known weak spots in picture pipeline and auth"
metadata: 
  node_type: memory
  type: project
  originSessionId: 6d9b0ff3-71c5-41e7-bd5c-747a86a3fdfa
  modified: 2026-08-02T13:15:22.472Z
---

Solo hobby/learning project (JoJo + One Piece stand/devil-fruit simulator). **Prod deploy set up and verified live 2026-08-02**: CI (`.github/workflows/ci.yml`, required check `ci-success` on PRs→main) + CD (`.github/workflows/cd.yml`, push→main, Tailscale+SSH) deploy to a self-hosted server at `~/projects/JojoOnePieceSimulator2`, domain `jojo-one-piece-simulator.duckdns.org` via an existing Nginx Proxy Manager. First real PR→merge→deploy round-trip succeeded end-to-end. `docker-compose.yml` is now a base file with `docker-compose.dev.yml`/`docker-compose.prod.yml` overrides (compose merges `ports:` lists across `-f` files, so ports had to move out of the base entirely). GitHub flags `cd.yml` as a possibly-malicious workflow (dumps `toJSON(secrets)` off-repo) — needs manual "Approve and run" on first run and after any edit to that file, expected not a bug. Full gotcha list (pnpm/action-setup version resolution, required-check name matching) in ObsidianVault `cicd-deployment.md`/`docker-setup.md`.

Known weak spots (owner-flagged, not visible from code alone):
- Picture pipeline (libvips) is fragile — see [[cicd_picture_pipeline]] for the CI build-tag requirement.
- ~~Auth/session flow incomplete — Google login works, rest of session lifecycle is WIP.~~ **Corrected 2026-09-02** ([[auth-hardening-2026-09-02]]): the flow is complete and verified in prod. What was actually missing was test coverage (now added) and accurate docs (now fixed). JWT-in-query-string for SSE/WS, `localStorage` on web, and no refresh tokens are accepted trade-offs, not gaps.

**Why:** owner is sole contributor, building for learning/fun, not shipping to real users yet.
**How to apply:** don't assume production concerns (scaling, multi-tenant auth hardening) unless asked; flag auth/picture-pipeline fragility when touching those areas; ADRs/docs here are for the owner's own future reference, not team onboarding.
