# Jamzy API (Go)

The Jamzy backend: one Go process that serves both the REST contract and the Socket.IO v4 real-time contract used by the React client in [`../client`](../client). Persistent state is PostgreSQL via `pgx` with hand-written SQL; ephemeral real-time state (presence, queue, host handover) is in process memory.

See the [root README](../README.md) for the project overview, the synchronization protocol, and deployment.

## Requirements

- Go 1.27 or newer
- PostgreSQL 14 or newer (developed against Neon and a local instance)
- A Clerk application: the secret key (`sk_...`) verifies session tokens; without it every authenticated request is rejected
- A YouTube Data API v3 key; the music endpoints answer `503` without it

## Configuration

Copy `.env.example` to `.env` and fill in the values. The file is read from the working directory at startup, and real environment variables always win over the file.

| Variable | Required | Default | Purpose |
| --- | --- | --- | --- |
| `DATABASE_URL` | yes | - | PostgreSQL connection string |
| `NODE_ENV` | no | `development` | Set to `production` for production behavior |
| `PORT` | no | `5000` | HTTP listen port |
| `CORS_ORIGIN` | no | `http://localhost:5173` | Comma-separated list of allowed origins |
| `CLERK_PUBLISHABLE_KEY` | no | - | Frontend key; carried for parity, not used by the API |
| `CLERK_SECRET_KEY` | no | - | Clerk secret key used to verify session tokens |
| `CLERK_AUTHORIZED_PARTIES` | no | `CORS_ORIGIN` | Allowed `azp` claim values for session tokens |
| `YOUTUBE_API_KEY` | no | - | YouTube Data API key used by the music endpoints |

## Running

```bash
cd server
go mod download
go run ./cmd/api
```

The service connects to PostgreSQL, applies any pending migrations, then listens on `0.0.0.0:$PORT`. `SIGINT` and `SIGTERM` trigger graceful shutdown. Build a binary with:

```bash
go build -o bin/jamzy-api ./cmd/api
```

This is exactly what the repository's `render.yaml` blueprint does on [Render](https://render.com): the API serves `/api/health` for the platform's health probe.

## Migrations

The SQL in `internal/db/migrations` is the SQL Prisma originally generated for this project, kept verbatim. [`prisma/schema.prisma`](prisma/schema.prisma) is the readable form of the same model and is the reference for the data model; the SQL is what the service executes. Migrations are embedded into the binary and applied on startup by `internal/db/migrate.go`.

- Applied versions are recorded in `schema_migrations`.
- If the database already has Prisma's `_prisma_migrations` table, those versions are treated as applied, so an existing Prisma-managed database is adopted without replaying migrations.
- A PostgreSQL advisory lock serializes concurrent startups.
- Each migration runs inside a transaction.

To add a migration, drop a new `<timestamp>_<name>.sql` file into `internal/db/migrations`. Files are applied in name order, and the matching change must also land in the schema file.

## REST API

All endpoints except `GET /api/health` require a Clerk session token in the `Authorization: Bearer <token>` header. Every endpoint returns JSON. Rate limits are per authenticated user and per endpoint, held in memory, answered with `429` plus `Retry-After`.

| Method | Path | Rate limit | Description |
| --- | --- | --- | --- |
| GET | `/api/health` | - | Public liveness probe |
| GET | `/api/me` | 60 / 60s | Resolve the authenticated application user |
| POST | `/api/users/profile` | 10 / 10m | Create or update the user profile (username onboarding) |
| POST | `/api/rooms` | 15 / 5m | Create a room and become its host |
| POST | `/api/rooms/{roomCode}/join` | 30 / 5m | Join a room |
| GET | `/api/rooms/{roomCode}` | 60 / 5m | Fetch room state |
| POST | `/api/rooms/{roomCode}/leave` | 60 / 5m | Leave a room, handing over the host role if needed |
| GET | `/api/music/search` | 30 / 60s | Search YouTube (up to 20 results) |
| GET | `/api/music/video` | 60 / 60s | Resolve a single video or pasted URL |
| GET | `/api/music/playlist` | 15 / 60s | Resolve a playlist into up to 50 playable tracks |

Room codes are six characters. Thumbnail URLs are accepted only from YouTube's CDN hosts, video IDs are validated against a strict 11-character pattern, and message contents are capped at 500 characters, all server-side regardless of what any client sends.

Engineering documentation and the interactive API reference are served at `/docs`, with the machine-readable OpenAPI document at `/docs/openapi.json`.

## Socket.IO API

The Socket.IO server is mounted at `/socket.io/` and speaks the Socket.IO v4 protocol (`EIO=4`). The client sends its Clerk token in the connect packet payload as `{"token":"<session token>"}`. A rejected handshake answers with `connect_error`. Socket events also carry `azp` validation of the token's authorized parties.

Every client-to-server event takes a payload plus an acknowledgment callback; the callback receives the result object, or an error string on failure.

| Event (client to server) | Rate limit | Authorization | Description |
| --- | --- | --- | --- |
| `room:join` | 20 / 60s | Member | Join a room, receive presence, playback, queue, and chat history in one ack |
| `room:leave` | - | Joined socket | Leave the current room |
| `playback:set` | 40 / 10s | Host only | Start a track |
| `playback:control` | 40 / 10s | Host only | Play, pause, or seek |
| `playback:clear` | 40 / 10s | Host only | Stop playback and clear the queue |
| `queue:add` | 100 / 10s | Host only | Append a track |
| `queue:prepend` | 100 / 10s | Host only | Insert a track at the front |
| `queue:remove` | 100 / 10s | Host only | Remove a queue entry |
| `queue:reorder` | 100 / 10s | Host only | Move a queue entry |
| `queue:advance` | 100 / 10s | Host only | Advance to the next track |
| `queue:clear` | 100 / 10s | Host only | Empty the queue |
| `chat:send` | 15 / 10s | Member | Send a chat message |
| `host:transfer` | 10 / 60s | Host only | Hand the host role to another member |

| Event (server to client) | Description |
| --- | --- |
| `presence:update` | Member list and current host |
| `playback:update` | Playback snapshot: track, position, playing state, wall-clock timestamp |
| `queue:update` | The full queue |
| `chat:new` | A persisted chat message |
| `host:update` | The host role changed |
| `host:pending` | The host disconnected; the handover countdown started |

Playback, presence, and queue state live in memory for the lifetime of the process. When a host's last socket drops, a 30-second countdown arms in a guarded store; reconnecting within the window cancels it, and expiry promotes the longest-standing remaining member in a transaction. Clients follow the resulting role through `host:update`, so playback permissions move with the role, not with the initial REST load.

## Testing

```bash
go test ./...
```

Unit tests sit beside the packages they cover: validation, rate limiting, middleware, host transfer, acknowledgment handling, room codes. They need no database.

The end-to-end suite drives the whole stack, including a hand-written Socket.IO v4 client, against a real PostgreSQL database. It is skipped unless `TEST_DATABASE_URL` is set, and it truncates the `messages`, `room_members`, `rooms`, and `users` tables:

```bash
TEST_DATABASE_URL=postgresql://user:pass@127.0.0.1:5432/jamzy_scratch \
  go test ./internal/httpapi/ -count=1
```

> [!WARNING]
> Point `TEST_DATABASE_URL` at a disposable local database only, and never at a database holding real data.

Migrations must have been applied once before running it, which happens automatically on the first `go run ./cmd/api`. One test in the suite, [`internal/httpapi/playback_zone_test.go`](internal/httpapi/playback_zone_test.go), pins the database session to a non-UTC time zone (`Asia/Kolkata`) and protects a real past regression: playback timestamps must be written as UTC wall clocks so playhead math stays correct for any database time zone.

## Layout

| Path | Responsibility |
| --- | --- |
| `cmd/api` | Wiring: config, pool, migrations, services, HTTP server, shutdown |
| `internal/apperr` | Error type carrying an HTTP status |
| `internal/auth` | Clerk token verification and the `Verifier` interface |
| `internal/config` | Environment loading and rate-limit rules |
| `internal/db` | Connection pool, embedded migrations, migrator |
| `internal/httpapi` | Route table and dependency wiring |
| `internal/httpx` | Middleware: recovery, logging, security headers, CORS, auth, rate limits |
| `internal/id` | Collision-resistant identifiers, replacing Prisma's `cuid()` |
| `internal/logging` | Structured logger |
| `internal/messages` | Chat persistence |
| `internal/music` | YouTube search, video resolver, playlist import with caching |
| `internal/playback` | Playback snapshot persistence |
| `internal/ratelimit` | In-memory fixed-window limiter |
| `internal/realtime` | Socket.IO hub, presence, queue, host transfer |
| `internal/rooms` | Room and membership logic |
| `internal/textlen` | UTF-16 length measurement for input validation |
| `internal/timex` | ISO-8601 timestamp formatting |
| `internal/users` | User profiles |
| `prisma/schema.prisma` | The readable data model, the reference for the SQL migrations |

## Design notes

A few decisions that shape the code, chosen deliberately rather than by default:

- **Routing and querying: standard library plus `pgx`.** `net/http`'s method-prefixed `ServeMux` and hand-written SQL in per-domain services replace a framework and an ORM. The API surface is small and the SQL is explicit, so there is little for either to manage away, and upgrading Go does not wait on a web framework release.
- **Ephemeral state stays out of PostgreSQL.** Presence, the queue, and the handover countdown are bounded maps guarded by mutexes. Persisting them would add write amplification and schema churn to state that is meaningless across restarts; the queue snapshot is handed to joiners in the `room:join` ack instead.
- **Playback timestamps are UTC wall clocks.** Client playhead math compares the browser clock to the server's stored timestamp, so the write path pins `timezone('UTC', now())` regardless of database time zone. The zone regression test above keeps it that way.
- **One structured error path.** Domain errors carry an HTTP status through `internal/apperr`, handlers stay thin, and internal failures surface as `500` without leaking internals to clients.
- **Panic containment.** Recovery middleware wraps HTTP routes, and the realtime paths recover per handler, so malformed payloads and logic bugs degrade to a single error response instead of a process exit.
