# Hexagonal Refactor Roadmap

This roadmap captures incremental, test-first steps to move from handler-centric logic to explicit ports/use-cases/adapters.

## Current State

Already extracted:

- core pure logic: `filequery`.
- core use-case: `archivequery`.
- archive adapter: `ziparchive`.

Still handler-heavy:

- local file lookup/open,
- usage accounting orchestration,
- URL fetch + digest + move flow.

## Target Shape

- `internal/core/...` contains only business policies and use-cases.
- `internal/ports/...` defines interfaces for storage, clock, downloader, usage repository.
- `internal/adapters/...` implements ports (filesystem, HTTP client, archive libs, persistence).
- `internal/handlers/...` maps HTTP <-> use-case DTOs only.

## Step Plan

### Step 1: File Read Use Case

- Add `core/filedownload` use case with port interfaces:
  - `Finder` (resolve candidate path),
  - `Opener` (open stream + stat),
  - `ArchiveReader` (optional extract).
- Keep range parsing in `filequery`.
- Add tests for behavior mapping independent of Gin.

### Step 2: File Write/Delete Use Cases

- Add `core/filewrite` and `core/filedelete`.
- Move usage mutation behind a `UsageStore` port.
- Add unit tests for success and failure branches.

### Step 3: URL Ingest Use Case

- Add `core/urlingest` with ports:
  - `Downloader` (`Do(req)`),
  - `TempStore`,
  - `BlobStore`,
  - `Hasher`.
- Add SSRF controls in adapter layer (allowlist/block private ranges).

### Step 4: Error Taxonomy

- Define typed domain errors in core.
- Centralize HTTP status mapping in handler layer.
- Standardize JSON error body (`code`, `message`).

### Step 5: Operational Hardening

- Replace broad `/tmp` cleanup with service-owned temp directory cleanup.
- Add structured logs and request IDs.
- Add request size/time limits.

## Verification Strategy Per Step

- Characterization tests first for current behavior.
- New core tests for use-cases and ports.
- Keep `go test ./...` green at each commit.
- Avoid cross-step big-bang refactor.

## Definition of Done

- Handlers contain no filesystem/archive business logic.
- Core packages have high test coverage and no Gin imports.
- Infra libraries (`ziputil`, `os`, `net/http`) are adapter-only dependencies.
