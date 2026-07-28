---
title: Backend API contract
tags:
  - project
  - jojo-onepiece-simulator
  - backend
---

# Backend API contract

Go/chi backend, `apps/backend`. Base path `/api/v1`, port 8080 (`apps/backend/internal/config/config.go:187`).

## Auth

- `POST /api/v1/auth/google` `{idToken}` → `{accessToken, tokenType:"Bearer", expiresAt, user{id,email,username,completeName,picture,role}}`
- 201 first login, 200 returning.
- **No refresh token, no cookies.** 24h JWT. 401 anywhere → force re-login, no silent refresh.
- All other routes: `Authorization: Bearer <token>` required.
- Writes additionally require `role === "ADMIN"`.

## Caching

Read routes emit `ETag` + `Cache-Control: private` + `Vary: Authorization`, honor `If-None-Match` → `304`. (`apps/backend/internal/infrastructure/api/endpoints/cache_headers.go`)

Backend also has a caching layer for Stand repo + picture storage (ETag/Cache-Control) and background image processing with `pictureStatus` tracking, returning `202 Accepted` on picture upload.

## JSON shape

camelCase. Errors: `{error, details?[]}`.

## Enums (strings)

- rarity: `COMMON|RARE|EPIC|LEGENDARY`
- stand stats: `E|D|C|B|A|INFINITE|NULL`
- fruitType: `PARAMECIA|ZOAN|LOGIA|SPECIAL_PARAMECIA|ANCIENT_ZOAN|MYTHICAL_ZOAN`
- role: `REGULAR|ADMIN`
- pictureStatus: `NONE|PENDING|READY|FAILED`

## CORS

Deny-all unless `CORS_ALLOWED_ORIGINS` set explicitly (`config.go:230`). Frontend origin must be added — see [[docker-setup]].

## Websockets

None exist. Do not build real-time features against this backend yet.

## Domain entities seen in backend history

DevilFruit (CRUD, filtering, picture handling), Stand (with cache layer), pictures (background processing + status).

Related: [[frontend-stack]]
