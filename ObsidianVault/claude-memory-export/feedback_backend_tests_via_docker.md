---
name: feedback-backend-tests-via-docker
description: "Windows App Control policy blocks running `go test` binaries built in %TEMP% on this host — always run backend tests through Docker"
metadata: 
  node_type: memory
  type: feedback
  modified: 2026-08-17T09:21:15.934Z
  originSessionId: 487a3148-047e-4175-9c75-895c6bb671c3
---

Running `go test ./...` directly on this machine fails with `fork/exec ...:
Una directiva de Control de aplicaciones bloqueó este archivo` — a Windows
Application Control policy blocks executing test binaries built under
`%TEMP%`. Building the test binary inside the project directory instead
(`go test -c -o local/path/test.exe ./pkg && ./local/path/test.exe`) works
around it, but the user explicitly asked to always use Docker instead.

**Why:** the repo already anticipated this — `apps/backend/Makefile`'s
`test-docker`/`test-vips-docker`/`test-integration-docker` targets exist
specifically because "some host machines' Windows Application Control policy
randomly blocks host-run `go test` binaries" (comment right above them).

**How to apply:** run backend tests via
`docker compose -f deployments/docker-compose.yml -f deployments/docker-compose.test.yml run --rm backend-test go test ./...`
(equivalent to `make test-docker`, but `make` isn't installed in this Git Bash
setup) from `apps/backend`. Never reach for a bare host-side `go test`/`go
build`-then-run-binary workaround on this project — go straight to Docker.
`go build ./...` and `go vet ./...` (compile-only, no binary execution) are
fine to run directly on the host.
