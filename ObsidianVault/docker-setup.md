---
title: Docker setup
tags:
  - project
  - jojo-onepiece-simulator
  - docker
---

# Docker

## Frontend images

- **Dev**: `apps/frontend/Dockerfile.dev` — `node:22-alpine`, corepack pnpm, `pnpm install --frozen-lockfile`, expects source volume-mounted, `CMD pnpm expo start --web --port 8081`.
- **Prod**: `deployments/docker/Dockerfile.frontend` — multi-stage. Builder = `node:22-alpine` → `pnpm expo export -p web` → `dist/`. Runtime = `nginx:alpine`, non-root user uid 10001 (matches backend convention), serves via `deployments/docker/nginx.frontend.conf` (long-cache `/_expo/static/`, no-cache `sw.js`/`manifest.json`, SPA fallback).
- Build context for prod image = **repo root** (matches backend's `Dockerfile.backend` convention), not `apps/frontend`.
- Both need `RUN touch tamagui-web.css` after source COPY — see [[frontend-stack]] gotchas.
- Prod build needs `ARG`/`ENV` for `EXPO_PUBLIC_API_URL` / `EXPO_PUBLIC_SOCKET_URL` since Expo inlines them at build time.

## docker-compose.yml

- Backend service: added `CORS_ALLOWED_ORIGINS: http://localhost:${FRONTEND_PORT:-3000}` — otherwise backend deny-all blocks the browser.
- New `frontend` service: build context `..` (repo root), dockerfile `deployments/docker/Dockerfile.frontend`, port `${FRONTEND_PORT:-3000}:80`, `depends_on: backend`.

## Root .dockerignore gotcha

Root `.dockerignore` (`C:\code\JojoOnePieceSimulator2\.dockerignore`) needed `**/node_modules`, `**/dist`, `**/.expo`, `**/tamagui-web.css` added — without it local `node_modules` leaked into build context (repo root) and clashed with fresh `pnpm install` inside the image (`ERR_PNPM_ABORTED_REMOVE_MODULES_DIR_NO_TTY`). But excluding `tamagui-web.css` from context also means it must be created inside the Dockerfile via `touch`, not copied in.

## Verified

Prod image built + run standalone: `docker run` + curl confirmed `/`, `/manifest.json`, `/sw.js` all correct (200s, `sw.js` has `Cache-Control: no-cache`). NOT yet verified: `docker compose up` running backend+frontend together (deferred as a manual/optional check per plan).

Related: [[frontend-stack]], [[backend-contract]]
