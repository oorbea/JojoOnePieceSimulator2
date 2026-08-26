---
name: cicd-picture-pipeline
description: "When building the CI/CD pipeline for JojoOnePieceSimulator2 backend, must compile/test both without and with the vips build tag, and validate the Docker image"
metadata: 
  node_type: memory
  type: project
  originSessionId: 92b94a62-7868-4330-9716-9f7a54304b12
  modified: 2026-08-02T11:31:45.951Z
---

**Resolved 2026-08-02**: `.github/workflows/ci.yml` (the old empty `.github/cicd.yml` was deleted — wrong path anyway, GitHub only reads `.github/workflows/`) now runs `go test ./...` and `go test -tags vips ./...` (libvips-dev installed on the runner) plus a `docker build -f deployments/docker/Dockerfile.backend .`. All 3 steps below are covered. See [[cicd-deployment]] in the repo's ObsidianVault for the full pipeline.

The image-compression pipeline (`apps/backend/internal/infrastructure/imaging`) is behind a `vips` Go build tag: `go build`/`go test ./...` (no tag) only compile the no-cgo stub, never the real libvips adapter.

**Why:** libvips needs cgo + a system `vips-dev`, which the Windows dev box doesn't have. The stub lets the rest of the module build/test everywhere; but a regression in the real adapter (`processor.go`) would silently ship undetected unless something builds it with the tag.

**How to apply:** when setting up the CI/CD pipeline for this repo, the job must run at least:
1. `go build ./...` && `go test ./...` (default, no tag — fast, no libvips needed)
2. Install `vips-dev` (apt: `libvips-dev`, or use an alpine/golang image) then `go build -tags vips ./...` && `go test -tags vips ./...` (or `go vet -tags vips ./...` at minimum) — this is the one that actually exercises govips/libvips linkage
3. `docker build -f deployments/docker/Dockerfile.backend .` from repo root — full end-to-end proof the production image (cgo + vips + codecs) still builds

Skipping step 2/3 means the vips-tagged code path is unverified by CI even though it's what ships in production.

See [[picture_compression_pipeline]] for the feature this build constraint belongs to, if that memory exists.
