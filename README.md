# Sky Snapshot Creator

Focused Go snapshot-control API for the SKYCOIN4444 engineering portfolio.

## Status

**Engineering beta.** The service provides an in-memory snapshot registry with deterministic IDs, bounded capacity, idempotent creation, input validation, health/readiness endpoints, a Prometheus-style metric, race-tested concurrency, vulnerability scanning, and a non-root container. It does **not** perform cloud/block-device snapshots, persist state, replicate data, authenticate callers, or represent a production backup system.

## API

- `GET /health` — liveness.
- `GET /ready` — readiness.
- `GET /metrics` — current in-memory snapshot count.
- `POST /api/v1/snapshots` — create an idempotent snapshot record from `{ "volume": "vol-123" }`.
- `GET /api/v1/snapshots/{id}` — fetch one snapshot record.

Volume identifiers are trimmed, limited to 128 characters, and restricted to letters, digits, `.`, `_`, `:`, and `-`. Request bodies are capped at 4 KiB. `MAX_SNAPSHOTS` defaults to 1000 and must be between 1 and 100000.

## Local verification

Requires Go 1.25.x:

```bash
gofmt -w .
go vet ./...
go test -race -count=1 ./...
go build ./cmd/server
```

Run with:

```bash
BIND_ADDR=:8080 MAX_SNAPSHOTS=1000 go run ./cmd/server
```

## Container

```bash
docker build -t sky-snapshot .
docker run --rm -p 8080:8080 sky-snapshot
```

The runtime image is distroless and runs as UID/GID 65532 rather than root.

## Architecture

The HTTP boundary uses the Go standard library. Snapshot records live in a mutex-protected map. IDs are stable SHA-256-derived identifiers from the validated volume value, which makes repeated requests idempotent for the same volume within a process. The service deliberately models snapshot **control metadata**, not actual storage snapshots.

## SKYCOIN4444 integration

Use this as a stable service boundary for development workflows that need bounded snapshot metadata or orchestration examples. A real ecosystem backup component should replace the in-memory store with durable state and connect to a verified cloud/storage provider rather than copying this implementation into a flagship app.

## Security and limits

See `SECURITY.md`. Authentication, authorization, TLS termination, durable audit logging, encryption at rest, provider credentials, and disaster-recovery validation are outside this repository's current evidence boundary.

## License

See `LICENSE`.
