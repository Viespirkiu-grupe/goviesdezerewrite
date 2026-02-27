# Security and Operations Spec

## Security Model

### Authentication

- Controlled by `REQUIRE_API_KEY`.
- For non-GET requests, middleware requires `X-API-Key` exact match.
- GET requests are intentionally public in current model.

Implication:

- read endpoints are open unless external network controls are applied.

### Input Handling

- Upload uses path param as filename and writes to computed shard path.
- Download URL endpoint trusts remote URL from request body.
- Archive extraction uses caller-provided `extract` string for in-archive matching.

### Security Risks to Track

- default API key is hardcoded if env var missing,
- no request body size limits,
- URL fetch has no allowlist/denylist (SSRF risk),
- no rate limiting,
- no audit trail beyond request logs.

## Observability

Current:

- request logs with timestamp/method/path/status/latency,
- ad-hoc logs in download candidate checks and copy failures.

Missing but recommended:

- structured logging with correlation IDs,
- per-endpoint counters and latencies,
- error taxonomy and metrics,
- archive extraction timing and size metrics.

## Runtime and Deploy

Container build:

- multi-stage Dockerfile,
- static build: `CGO_ENABLED=0 go build -o app .`,
- runtime image: `alpine`, includes `p7zip`.

Required runtime configuration:

- `PORT`, `STORAGE_PATH`, `REQUIRE_API_KEY`, `API_KEY`.

Persistent volumes:

- storage root,
- working directory file `usage.json` (or relocate explicitly).

## Housekeeping Process

Background goroutine runs every minute:

- executes `find /tmp -mindepth 1 -mmin +5 -exec rm -rf {} +`

Operational caveat:

- this removes old temp items globally under `/tmp`, not just this service's files.
