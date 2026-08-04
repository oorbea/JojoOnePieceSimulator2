---
title: CI/CD + production deployment
tags:
  - project
  - jojo-onepiece-simulator
  - cicd
  - deployment
---

# CI/CD + production deployment (2026-08-02)

## What existed before
`.github/cicd.yml` was present but **empty** and in the wrong location — GitHub Actions only reads workflows from `.github/workflows/*.yml`, so it never ran anything. Deleted, replaced by two real workflows.

## `.github/workflows/ci.yml`
Triggers on PR → `main` and push → `develop`. Jobs: `backend` (go vet, `go test ./...`, and — closing the known gap, see [[cicd_picture_pipeline]] / ADR — `go test -tags vips ./...` with `libvips-dev` installed on the runner), `frontend` (pnpm typecheck + lint), `backend-image`/`frontend-image` (docker build of each prod Dockerfile, `type=gha` layer cache — split into two parallel jobs, was one sequential `images` job originally), `compose-config` (`docker compose ... config` against the prod override, with dummy env values — no Buildx needed, so it's its own fast job). A final `ci-success` job depends on all of them — it's the single required status check, so future new CI jobs never need a branch-protection change.

Branch protection on `main` applied via `gh api` (`required_status_checks.contexts=["ci-success"]`, strict=true) — PRs can't merge until CI is green.

### Gotchas hit during first real rollout (2026-08-02)
- **`pnpm/action-setup` needs a pinned version.** Repo has no root `package.json`, only `apps/frontend/package.json` — the action defaults to reading `packageManager` from `./package.json` at repo root, finds nothing, and fails with "No pnpm version is specified" even though `apps/frontend/package.json` has it. Fixed two ways together: added `"packageManager": "pnpm@11.18.0"` to `apps/frontend/package.json`, and set `package_json_file: apps/frontend/package.json` on the `pnpm/action-setup` step (job's `defaults.run.working-directory` only affects `run:` steps, not `uses:` actions).
- **Required status check name must match the job's exact display `name:`, not its id.** Branch protection was set with context `ci-success`, but the job's `name:` was `CI success` (space, different case) — GitHub reports checks by the display name, so it never matched and the PR sat forever on "Expected — Waiting for status to be reported" even though every job had gone green. Fixed by setting the job's `name: ci-success` to match exactly.
- **GitHub flags `cd.yml` as a possibly-malicious workflow**, requiring manual "Approve and run" by someone with write access before its first run (and again after any future edit to `cd.yml`). Expected, not a bug: the workflow reads *every* secret via `toJSON(secrets)` and ships them off-repo via scp/ssh — exactly the exfiltration pattern GitHub's static scanner watches for. The reference pipeline this was adapted from does the same thing. Living with occasional manual approval was the chosen tradeoff over listing ~45 secrets explicitly by name.

## `.github/workflows/cd.yml`
Triggers on push → `main` (i.e. after a PR merges once CI passed). Single job:
1. Connects to the target server over **Tailscale** (`tailscale/github-action`, OAuth client, `tag:ci`).
2. Writes `deployments/.env` from `toJSON(secrets)` + `toJSON(vars)` — every `[SECRET]`/`[CONFIG]` line in `deployments/.env.example` must exist as a same-named GitHub secret/variable. Excludes the CD-only infra credentials (Tailscale OAuth, SSH access) that the app itself never reads.
3. `scp`s that `.env` to `~/projects/JojoOnePieceSimulator2/deployments/.env` on the server.
4. `ssh`es in and: brings up `postgres`/`redis` first (covers first-ever deploy), records the current goose DB version, builds the new backend/frontend images (old containers keep serving), runs migrations against the **new** image via a one-off container *before* replacing anything, and rolls back (`goose down-to` + `git reset --hard`) if that migration fails. Only then does `up -d --remove-orphans backend frontend`, waits for the backend healthcheck, prunes old images.

Pattern adapted from a working reference pipeline in another repo (Mishkis-app/backend) — same Tailscale+SSH+pre-deploy-migration-with-rollback shape, adjusted for this repo's Go/goose/docker-compose stack instead of Python/alembic.

## Production topology
- Domain: `jojo-one-piece-simulator.duckdns.org`, fronted by an existing Nginx Proxy Manager instance on the server's `public-net` Docker network.
- Single host: `/` → `frontend:80`, `/api` → `backend:8080` (backend already mounts routes under `/api/v1`, so NPM's custom location forwards the path unchanged, no rewrite needed).
- Neither `backend` nor `frontend` publish ports to the host anymore in prod — see [[docker-setup]] for the compose base/dev/prod split this required.
- **Manual, not automatable**: the domain must be added as an authorized JS origin/redirect URI in the Google Cloud OAuth console, or login breaks in prod.

## GitHub secrets/variables needed
Every `deployments/.env.example` line ([SECRET]→secret, [CONFIG]→variable), plus CD-only: `TS_OAUTH_CLIENT_ID`, `TS_OAUTH_SECRET`, `SERVER_IP`, `SERVER_USER`, `SSH_PRIVATE_KEY`. Full list documented in `deployments/README.md` (new file).

## Verified locally before rollout
- `docker build -f deployments/docker/Dockerfile.backend .` succeeds (vips path compiles) and the runtime image has `wget` (busybox) for the healthcheck.
- `docker compose -f docker-compose.yml -f docker-compose.prod.yml config` — no ports published on backend/frontend.
- `docker compose -f docker-compose.yml -f docker-compose.dev.yml config` — ports restored for local dev.
- `go vet ./...` + `go test ./...` clean; frontend `pnpm typecheck` clean, `pnpm lint` — 0 errors (430 CRLF-only warnings, pre-existing, harmless).

## Verified end-to-end in production (2026-08-02)
First real PR → CI green → merge → CD deploy round-trip done successfully against the actual server. `docker compose ps` on the server shows `postgres`/`redis`/`backend` as `healthy`; `frontend` shows plain `Up` with no health column — expected, nginx has no `healthcheck:` defined in compose (only backend/postgres/redis do), not a failure.

Not yet verified: a real rollback (deliberately broken migration) — do this next if/when confidence in that path is needed before relying on it for a real incident.

Related: [[docker-setup]], [[ADR]], [[backend-contract]]
