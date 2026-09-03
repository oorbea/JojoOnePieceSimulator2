# Deployments

## Local development

```
docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d
```

(`apps/backend/Makefile`'s `db-up`/`db-down`/`migrate-*` targets already do this.)
`docker-compose.dev.yml` is what publishes `backend`/`frontend` ports to the host -
the base `docker-compose.yml` doesn't, on purpose (see its comments).

## Administrative DB access

If you need to connect to Postgres from DBeaver through SSH/Tailscale, use
the dedicated override that binds Postgres only to `127.0.0.1` on the server.
The CD pipeline also applies this override automatically on every deploy.
If you need to run it manually, do so from `deployments/`:

```
docker compose -f docker-compose.yml -f docker-compose.tunnel.yml up -d postgres
```

Then point DBeaver's SSH tunnel at the server's Tailscale address and use:

- Remote host: `127.0.0.1`
- Remote port: `15432`
- Database: `jojo_one_piece_simulator`
- Username: `TrolloTron`
- Password: `Tutankamon.18`
- SSL mode: `disable`

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

The auth ones are the easiest to get wrong (a bad value fails at boot or,
worse, silently locks the owner out of every admin route), so they are spelled
out here too — they are ordinary `.env.example` entries, not extra secrets:

| Name | Kind | Purpose |
|---|---|---|
| `GOOGLE_CLIENT_ID` | `[SECRET]` | OAuth client id the backend verifies every Google ID token's `aud` against. Not secret in itself (client ids are public), but must be the *same* value as `EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID` or every login 401s. |
| `JWT_SECRET` | `[SECRET]` | HS256 signing key for this backend's own access tokens. **At least 32 characters** — `config.go` refuses to boot below that. Rotating it invalidates every issued token (no refresh tokens: everyone logs in again). |
| `ADMIN_EMAILS` | `[CONFIG]` | Comma-separated Google account emails promoted to `ADMIN` on login, matched case-insensitively. Empty means nobody is an admin, i.e. the whole admin UI is unreachable. |
| `STREAM_TICKET_TTL` | `[CONFIG]` | How long a minted SSE/WebSocket connection ticket stays redeemable (default `30s`). Not a secret — the whole point is that it's short-lived and single-use, unlike the JWT it replaced in `?token=`. See `ObsidianVault/stream-connection-tickets-2026-09-02.md`. |

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
