# JojoOnePieceSimulator2

Online game about JoJo's Bizarre Adventure and One Piece — players get a randomized loadout (Stand or Devil Fruit) and play through game modes like Gauntlet and Versus.

## Structure

- `apps/backend/` — Go API (hexagonal architecture: domain / application / infrastructure). See its [README](apps/backend/README.md).
- `apps/frontend/` — Expo + React Native Web PWA. See its [README](apps/frontend/README.md).
- `deployments/` — Docker Compose stack (local, tunnel, prod) and CI/CD notes. See its [README](deployments/README.md).
- `ObsidianVault/` — project knowledge base (architecture decisions, conventions, norms). Check it before implementing non-trivial changes.

## Getting started

```
docker compose -f deployments/docker-compose.yml -f deployments/docker-compose.dev.yml up -d --build
```

Or run `.\start.ps1` for an interactive menu wrapping the same compose stack plus goose migrations and common Go commands.

Frontend on `:3000`, backend on `:8080`. See `deployments/README.md` for env vars and production deployment.
