# Deployments

## Local development

```
docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d
```

(`apps/backend/Makefile`'s `db-up`/`db-down`/`migrate-*` targets already do this.)
`docker-compose.dev.yml` is what publishes `backend`/`frontend` ports to the host -
the base `docker-compose.yml` doesn't, on purpose (see its comments).

## Production

Deployed automatically by `.github/workflows/cd.yml` on every push to `main`
(after `.github/workflows/ci.yml` passes on the PR). Manually, it's the same
compose stack with the prod override:

```
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

Requirements on the server (`~/projects/JojoOnePieceSimulator2`, already cloned):

- Docker network `public-net` already created and shared with the reverse
  proxy (`docker network create public-net` if it's ever missing).
- Nginx Proxy Manager (or equivalent), also attached to `public-net`,
  proxying `jojo-one-piece-simulator.duckdns.org`:
  - `/` → `frontend:80`
  - `/api` → `backend:8080` (custom location; the backend already mounts its
    routes under `/api/v1`, so the path is forwarded as-is, no rewrite).
- Google Cloud Console OAuth client: add
  `https://jojo-one-piece-simulator.duckdns.org` as an authorized JavaScript
  origin / redirect URI, or login breaks in prod. **Not automatable — do
  this manually once.**

### GitHub repository secrets/variables

Every entry in `deployments/.env.example` marked `[SECRET]` or `[CONFIG]`
must exist in the GitHub repo with the same exact name — `[SECRET]` as a
repository **secret**, `[CONFIG]` as a repository **variable**. The CD
pipeline writes all of them into `deployments/.env` on the server before
deploying.

Plus these CD-only secrets (never written to `deployments/.env` — the app
doesn't read them):

| Secret | Purpose |
|---|---|
| `TS_OAUTH_CLIENT_ID` / `TS_OAUTH_SECRET` | Tailscale OAuth client, tag `tag:ci`, used by the CD runner to reach the server |
| `SERVER_IP` | Server's Tailscale IP/hostname |
| `SERVER_USER` | SSH user on the server |
| `SSH_PRIVATE_KEY` | SSH private key authorized on the server |

### Branch protection

CI must block merging to `main` until it passes:

```
gh api -X PUT repos/:owner/:repo/branches/main/protection \
  -F required_status_checks.strict=true \
  -F 'required_status_checks.contexts[]=ci-success' \
  -F enforce_admins=false \
  -F required_pull_request_reviews=null \
  -F restrictions=null
```

`ci-success` is a single job in `ci.yml` that depends on every other CI job,
so new CI jobs never require touching this setting again.
