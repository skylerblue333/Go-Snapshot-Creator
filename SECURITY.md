# Security Policy

## Supported status

Sky Snapshot Creator is an engineering-beta component. It is not a production backup platform and currently has no authentication, authorization, TLS termination, durable audit log, encryption-at-rest layer, cloud-provider credentials, or persistent datastore.

## Security controls present

The HTTP server enforces bounded request bodies, strict JSON decoding, bounded/validated volume identifiers, bounded in-memory cardinality, explicit server timeouts, deterministic IDs, concurrency-safe state, dependency vulnerability scanning, and a non-root distroless runtime image.

## Reporting

Do not include secrets or exploit payloads in public issues. Report suspected vulnerabilities privately through the repository owner's GitHub security/contact channel when available.

## Deployment guidance

Do not expose this service directly to the public Internet without an authenticated reverse proxy, TLS, network policy, rate limiting, centralized audit logging, and a durable backend appropriate to the intended environment.
