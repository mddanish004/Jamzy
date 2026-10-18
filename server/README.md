# Jamzy API (Go)

The Jamzy backend: a Go service that serves the HTTP and Socket.IO contract the React
client in `../client` expects.

## Requirements

- Go 1.27 or newer
- PostgreSQL 14 or newer (the project develops against Neon and a local PostgreSQL)

## Configuration

Copy `.env.example` to `.env` and fill in the values. `.env` is read from the working
directory at startup, and real environment variables always win over the file.

| Variable | Required | Default | Purpose |
| --- | --- | --- | --- |
| `DATABASE_URL` | yes | - | PostgreSQL connection string |
| `NODE_ENV` | no | `development` | Set to `production` for production behaviour |
| `PORT` | no | `5000` | HTTP listen port |
| `CORS_ORIGIN` | no | `http://localhost:5173` | Comma-separated list of allowed origins |
| `CLERK_PUBLISHABLE_KEY` | no | - | Frontend key; exposed for parity, not used by the API |
| `CLERK_SECRET_KEY` | no | - | Clerk secret key used to verify session tokens |
| `CLERK_AUTHORIZED_PARTIES` | no | `CORS_ORIGIN` | Allowed `azp` claim values for session tokens |
| `YOUTUBE_API_KEY` | no | - | YouTube Data API key used by the music endpoints |

Without `CLERK_SECRET_KEY` every authenticated request is rejected. Without
`YOUTUBE_API_KEY` the music endpoints answer `503`.

## Running

```bash
cd server
go mod download
go run ./cmd/api
```

The service connects to PostgreSQL, applies any pending migrations, and then listens on
`0.0.0.0:$PORT`. `SIGINT` and `SIGTERM` trigger a graceful shutdown.

Build a binary with:

```bash
go build -o bin/jamzy-api ./cmd/api
```

## Migrations

The SQL in `internal/db/migrations` is the SQL Prisma originally generated for this project,
kept verbatim. `prisma/schema.prisma` is the readable form of the same model and is the
reference for the data model; the SQL is what the service executes. Migrations are embedded
into the binary and applied on startup by `internal/db/migrate.go`.

- Applied versions are recorded in `schema_migrations`.
- If the database already has Prisma's `_prisma_migrations` table, those versions are
  treated as applied, so an existing Prisma-managed database is adopted without replaying
  migrations.
- A PostgreSQL advisory lock serialises concurrent startups.
- Each migration runs inside a transaction.

To add a migration, drop a new `<timestamp>_<name>.sql` file into
`internal/db/migrations`. Files are applied in name order.

## REST API

All endpoints except `GET /api/health` require a Clerk session token in the
`Authorization: Bearer <token>` header. Every endpoint returns JSON.

| Method | Path | Rate limit | Description |
| --- | --- | --- | --- |
| GET | `/api/health` | - | Liveness probe |
| GET | `/api/me` | 60 / 60s | Resolve the authenticated application user |
| POST | `/api/users/profile` | 10 / 10m | Create or update the user profile |
| POST | `/api/rooms` | 15 / 5m | Create a room and become its host |
| POST | `/api/rooms/{roomCode}/join` | 30 / 5m | Join a room |
| GET | `/api/rooms/{roomCode}` | 60 / 5m | Fetch room state |
| POST | `/api/rooms/{roomCode}/leave` | 60 / 5m | Leave a room, handing over the host role if needed |
| GET | `/api/music/search` | 30 / 60s | Search YouTube |
| GET | `/api/music/video` | 60 / 60s | Resolve a single video |
| GET | `/api/music/playlist` | 15 / 60s | Resolve a playlist |

Rate limits are per authenticated user and per endpoint, held in memory, and reported with
`429` plus `Retry-After`.

## Socket.IO API

The Socket.IO server is mounted at `/socket.io/` and speaks the Socket.IO v4 protocol
(`EIO=4`). The client sends its Clerk token in the connect packet payload:
`{"token":"<session token>"}`. A rejected handshake answers with a `connect_error`.

Every client-to-server event below takes a payload and an acknowledgement callback. The
callback receives the result object, or an error message string on failure.

| Event (client to server) | Rate limit | Description |
| --- | --- | --- |
| `room:join` | 20 / 60s | Join a room and receive its current state |
| `room:leave` | - | Leave the current room |
| `playback:set` | 40 / 10s | Host starts a track |
| `playback:control` | 40 / 10s | Host pauses, resumes, or seeks |
| `playback:clear` | 40 / 10s | Host stops playback |
| `queue:add` | 100 / 10s | Append a track |
| `queue:prepend` | 100 / 10s | Insert a track at the front |
| `queue:remove` | 100 / 10s | Remove a queue entry |
| `queue:reorder` | 100 / 10s | Move a queue entry |
| `queue:advance` | 100 / 10s | Advance to the next track |
| `queue:clear` | 100 / 10s | Empty the queue |
| `chat:send` | 15 / 10s | Send a chat message |
| `host:transfer` | 10 / 60s | Host hands the role to another member |

| Event (server to client) | Description |
| --- | --- |
| `presence:update` | Member list and current host |
| `playback:update` | Playback snapshot (track, position, playing state) |
| `queue:update` | Full queue |
| `chat:new` | A new chat message |
| `host:update` | The host role changed |
| `host:pending` | The host disconnected; a handover countdown started |

Playback, presence, and queue state live in memory for the lifetime of the process. When a
host disconnects, the handover waits 30 seconds before promoting the longest-standing
remaining member, which gives the host a chance to reconnect.

## Testing

```bash
go test ./...
```

Unit tests cover validation, rate limiting, the middleware, host transfer, and ack
handling. They need no database.

The end-to-end suite drives the whole stack, including a hand-written Socket.IO v4 client,
against a real PostgreSQL database. It is skipped unless `TEST_DATABASE_URL` is set, and it
truncates the `messages`, `room_members`, `rooms`, and `users` tables, so point it at a
throwaway database:

```bash
TEST_DATABASE_URL=postgresql://jamzy_go:jamzy_go@127.0.0.1:5432/jamzy_go_dev \
  go test ./internal/httpapi/ -count=1
```

Migrations must have been applied once before running it, which happens automatically on
the first `go run ./cmd/api`.

## Layout

| Path | Responsibility |
| --- | --- |
| `cmd/api` | Wiring: config, pool, migrations, services, HTTP server, shutdown |
| `internal/apperr` | Error type carrying an HTTP status |
| `internal/auth` | Clerk token verification and the `Verifier` interface |
| `internal/config` | Environment loading and rate limit rules |
| `internal/db` | Connection pool, embedded migrations, migrator |
| `internal/httpapi` | Route table and dependency wiring |
| `internal/httpx` | Middleware: recovery, logging, security headers, CORS, auth, rate limits |
| `internal/id` | Collision-resistant identifiers, replacing Prisma's `cuid()` |
| `internal/logging` | Structured logger |
| `internal/messages` | Chat persistence |
| `internal/music` | YouTube lookup with caching |
| `internal/playback` | Playback snapshot persistence |
| `internal/ratelimit` | In-memory fixed-window limiter |
| `internal/realtime` | Socket.IO hub, presence, queue, host transfer |
| `internal/rooms` | Room and membership logic |
| `internal/textlen` | UTF-16 length measurement for input validation |
| `internal/timex` | ISO-8601 timestamp formatting |
| `internal/users` | User profiles |
| `prisma/schema.prisma` | The readable data model, the reference for the SQL migrations |
| `docs` | Backend notes |

## Design notes

The problems encountered while porting the backend and the trade-offs behind the decisions
are recorded in [`docs/PROBLEMS_AND_TRADEOFFS.md`](docs/PROBLEMS_AND_TRADEOFFS.md).
