---
title: "Norma: verificar en Docker local antes de dar por acabado"
tags:
  - norma
  - jojo-onepiece-simulator
  - workflow
---

# Norma: los checks equivalentes a CI, siempre en Docker local (2026-08-26)

A petición explícita del owner: **una tanda no se da por terminada hasta pasar, en Docker local,
los checks que correrían en CI** - solo los relevantes según lo tocado (cambio solo-frontend ⇒ sin
suite de backend; solo-backend ⇒ sin la de frontend).

## Por qué Docker, no el host directamente

- Backend: Windows App Control bloquea binarios de `go test` recién compilados en esta máquina - ver
  [[feedback_backend_tests_via_docker]] (memoria de Claude) / `docker-setup.md`. `apps/backend`'s
  `Makefile` ya envuelve esto (`make test-docker`, `test-vips-docker`, `test-integration-docker`).
- Frontend: no hay un servicio `frontend-test` en `docker-compose.test.yml` (ese fichero solo cubre
  backend hoy) - el patrón que funcionó (2026-08-26): montar `apps/frontend` en `node:22-alpine`
  (misma versión que `.github/workflows/ci.yml`), pero **copiando el código fuente a un volumen
  propio del contenedor** en vez de bind-montar directo - un `pnpm install` bind-montado sobre el
  `node_modules` real de Windows lo corrompería con binarios de Linux. Ver el comando exacto abajo.

## Comandos

Backend (desde `apps/backend`, o con las rutas `-f` ajustadas):
```
docker compose -f ../../deployments/docker-compose.yml -f ../../deployments/docker-compose.test.yml \
  run --rm backend-test go test ./...
```
(`go build ./... && go vet ./...` primero si el cambio es grande - `make test-docker` ya solo cubre
`go test`.)

Frontend (desde la raíz del repo, PowerShell/Git Bash - un volumen nombrado cachea `node_modules`
entre ejecuciones para que las siguientes tandas no reinstalen desde cero):
```bash
docker run --rm -v "$(pwd):/repo:ro" -v jojo-frontend-work:/work node:22-alpine sh -c '
  corepack enable && corepack prepare pnpm@11.18.0 --activate
  mkdir -p /work/repo
  # primera vez: copia completa + pnpm install --frozen-lockfile
  # tandas siguientes: solo re-sincronizar src/ sobre /work/repo/frontend (rsync-like cp),
  # conservando node_modules ya instalado, y pnpm install de nuevo (rápido, usa el store cacheado)
  cd /work/repo/frontend
  CI=true pnpm install --frozen-lockfile --prefer-offline
  pnpm typecheck && pnpm lint && pnpm test:ci
'
```
No existe todavía un `docker-compose` dedicado para esto - candidato a añadir un servicio
`frontend-test` real a `docker-compose.test.yml` si esta norma se usa a menudo (evitaría reinventar
el volumen/copy cada vez). Anotado como mejora pendiente, no bloqueante.

**Flakiness conocida bajo Docker con muchos workers**: `pnpm test:ci` con el paralelismo por
defecto de Jest mostró 2 tests fallando por timing (`use-loadout-reveal.test.tsx`,
`tooltip.test.tsx`) que pasan limpio en aislamiento o con `--maxWorkers=2` - mismo patrón que el
flake de `StageCard` ya documentado en [[game-match-assignment-frontend]]. Si un test falla solo en
la corrida completa y pasa en aislamiento, es candidato a este mismo patrón antes de asumir una
regresión real.

## Contratos generados (2026-09-02)

Desde [[contratos-tipos-generados]], `/verify` tiene un §4 propio: si el diff toca
`apps/backend/cmd/typegen/**`, `dto/**`, `enums/**`, `apierr/**`, o
`apps/frontend/src/shared/contracts/**`, regenerar y diffear es obligatorio (`make types-check-docker`,
o el equivalente `docker compose ... run --rm typegen go run ./cmd/typegen` + `git diff --exit-code`
en el host). Un cambio en esas rutas de backend implica **también** correr la suite de frontend
aunque el diff todavía no tenga ningún fichero `apps/frontend/**` - regenerar debe producir uno.

## Mecanismo para no olvidarlo

- Skill `/verify` (`.claude/skills/verify/SKILL.md`) - detecta qué se tocó (diff contra la rama
  base) y lanza los comandos Docker correspondientes.
- Hook `Stop` en `settings.json` que recuerda los comandos exactos si hay cambios sin verificar al
  intentar terminar - configurado vía el skill `update-config`, no a mano.
- CI (`.github/workflows/ci.yml`) tiene su propio job `contracts` que hace lo mismo en cada push -
  ver [[contratos-tipos-generados]].

Related: [[docker-setup]], [[zettelkasten-workflow]], [[game-vote-buttons-2026-08-26]], [[contratos-tipos-generados]].
