# r2-worker-proxy

A small Cloudflare Worker that fronts one R2 bucket via a native binding
(`env.BUCKET.put/get/delete/list`), so `apps/backend` can reach R2 through
Cloudflare's generic `*.workers.dev` pool instead of R2's S3 API endpoint
(`<account>.r2.cloudflarestorage.com`).

**Why this exists:** an ISP can have a broken route to that one specific
Cloudflare anycast prefix while every other Cloudflare-fronted domain
works fine - see `ObsidianVault/storage-fallback-chain.md`'s 2026-09-13
incident. This Worker's traffic to R2 happens entirely inside
Cloudflare's own network, never touching that prefix.

Optional for `apps/backend`: it's only used when `R2_WORKER_URL` and
`R2_WORKER_SECRET` are both set (see
`internal/infrastructure/storage/workerproxy`); unset, the backend talks
to R2's S3 API directly as before.

## Wire protocol

See `src/index.ts`'s top comment - it must stay in lockstep with
`apps/backend/internal/infrastructure/storage/workerproxy/backend.go`.

## Commands

```
npm install
npm test         # vitest + Miniflare, regenerates worker-configuration.d.ts first
npm run typecheck
npm run dev       # wrangler dev, local Miniflare server
npm run deploy    # wrangler deploy (normally done by CD, not by hand)
```

`worker-configuration.d.ts` is generated from `wrangler.toml` by
`wrangler types` (run automatically before `test`/`typecheck` via
`pretest`/`pretypecheck`) - never hand-edited or committed.

## Setup (one-time, per Cloudflare account)

1. `R2_PROXY_SECRET` - a long random value, set as a GitHub secret with
   the same value the backend's `R2_WORKER_SECRET` uses. CD's
   `wrangler-action` sets it as a Worker secret on deploy.
2. `CLOUDFLARE_API_TOKEN` - a token with Workers Scripts:Edit + R2
   Storage:Edit permissions, as a GitHub secret.
3. `CLOUDFLARE_ACCOUNT_ID` - same value as the backend's `R2_ACCOUNT_ID`.

Once deployed, set the backend's `R2_WORKER_URL` (the Worker's
`https://jojo-r2-proxy.<subdomain>.workers.dev` address) and
`R2_WORKER_SECRET` as GitHub secrets too - they flow into
`deployments/.env` the same way every other secret does (see
`deployments/README.md`).
