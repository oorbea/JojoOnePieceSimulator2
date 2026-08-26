---
name: local-env-points-to-prod
description: "deployments/.env's EXPO_PUBLIC_API_URL defaults to the prod duckdns domain, not localhost — causes CORS errors when running the stack locally"
metadata: 
  node_type: memory
  type: project
  originSessionId: ac0eebf1-59f9-4d18-99fd-ff44fdacaf2f
  modified: 2026-08-04T11:32:19.017Z
---

`deployments/.env` (gitignored, local dev file) had `EXPO_PUBLIC_API_URL=https://jojo-one-piece-simulator.duckdns.org/api/v1` instead of `http://localhost:8080/api/v1`. `.env.example` leaves it blank, so this was set manually at some point (likely copy-pasted from prod config) and never corrected for local dev.

**Why:** `EXPO_PUBLIC_*` vars are baked into the frontend bundle at Docker build time (see [[frontend-stack]]). With the prod URL baked in, the browser calls the prod backend from `localhost:3000`, and prod's CORS `AllowedOrigins` (server-side config, correctly scoped to only the prod domain) rejects the request — surfaces as a CORS preflight error, but the real cause is wrong target URL, not a CORS bug.

**How to apply:** If `local-up` (or any local docker-compose run) throws CORS errors calling `jojo-one-piece-simulator.duckdns.org`, check `deployments/.env`'s `EXPO_PUBLIC_API_URL` first — it must be `http://localhost:8080/api/v1` for local dev. Fixed 2026-08-04; rebuild the frontend image after changing it (`docker compose ... up -d --build frontend`) since the var is compile-time, not runtime. Also note: switching the API target between prod/local may break the Google OAuth redirect_uri round-trip if it was registered for a specific origin — check that if login itself fails after this fix.
