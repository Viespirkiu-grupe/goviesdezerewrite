# System Design Spec

## High-Level Architecture

Current architecture is a pragmatic hybrid moving toward hexagonal design.

- Entrypoint: `main.go`
- Transport adapters: `internal/handlers/...` (Gin HTTP)
- Middleware: `internal/middleware`
- Core logic (extracted):
  - `internal/core/filequery` (range parsing, archive entry matching)
  - `internal/core/archivequery` (archive read use case)
- Infra adapter:
  - `internal/adapters/archive/ziparchive` (bridge to `ziputil`)
- Utilities:
  - `internal/utils` (path sharding, usage persistence, string similarity)
  - `internal/ziputil` (archive/attachment extraction machinery)

## Runtime Flow

1. Boot in `main.go`:
   - start tmp cleanup goroutine,
   - load config from env,
   - load usage cache,
   - ensure storage directory,
   - wire middleware + routes,
   - run Gin server.

2. Request path:
   - middleware logs request and enforces API key for non-GET methods,
   - handler executes use case and responds JSON/bytes.

## Request Flows

### Upload (`PUT /file/:filename`)

- Resolve sharded path from filename.
- Ensure parent directory exists.
- Overwrite destination file from request body stream.
- Recompute usage in memory and persist to `usage.json`.

### Download (`GET /file/:filename`)

- Resolve candidate paths (extension variants and extensionless fallbacks).
- If `extract` query present:
  - read archive bytes,
  - run `archivequery.ReadBestMatch` core use case,
  - select entry with `filequery.BestMatch` + similarity,
  - return extracted bytes.
- Else open local file stream.
- If `Range` present:
  - parse with `filequery.ParseRange`,
  - seek and stream limited window,
  - return `206` + `Content-Range`.
- Else stream full response (`200`).

### Delete (`DELETE /file/:filename`)

- Resolve candidate path,
- remove file,
- decrement usage,
- return deleted file metadata.

### Download by URL (`POST /download-url`)

- Fetch URL with `http.Get`.
- Stream response to temp file while computing md5.
- Move temp file into sharded final path by digest.
- Update usage and return digest/size.

## Design Rationale (What + Why)

- Keep HTTP layer thin enough to evolve quickly.
- Extract pure functions (`ParseRange`, `BestMatch`) for deterministic tests.
- Encapsulate archive read flow in core use case to reduce handler complexity.
- Keep local filesystem backend simple for easy Docker deployment.
- Use sharding-by-prefix to avoid huge flat directories.

## Known Design Tradeoffs

- Usage accounting is file-based and eventually consistent under concurrent process writes.
- `download` extraction currently buffers extracted file into memory.
- Temp cleanup strategy is broad (`find /tmp ... rm -rf`) and can affect unrelated temp files.
- Some `ziputil` functions are broad utility surface and not yet fully isolated behind ports.
