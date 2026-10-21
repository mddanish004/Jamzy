# Jamzy

Jamzy is a real-time music-listening app. You create a room, share a six-character code, and every listener sees the same YouTube video, at the same second, driven by one host. Chat and presence run alongside it.

**Live demo: [jamzy.mddanish.dev](https://jamzy.mddanish.dev/)**

[![CI](https://github.com/mddanish004/Jamzy/actions/workflows/ci.yml/badge.svg)](https://github.com/mddanish004/Jamzy/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.27-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![React](https://img.shields.io/badge/React-19.2-61DAFB?logo=react&logoColor=white)](https://react.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Neon-4169E1?logo=postgresql&logoColor=white)](https://neon.tech/)

The full system is a React single-page app, a Go API with a Socket.IO real-time layer, and PostgreSQL, deployed on Vercel, Render, and Neon. This document covers the whole repository; the backend and frontend have their own READMEs: [`server/README.md`](server/README.md) and [`client/README.md`](client/README.md).

## Why it exists

Watching a playlist together over a video call degrades audio quality, screen sharing burns bandwidth, and syncing separate browser tabs never quite lines up. Jamzy avoids all three by making the server the single authority on playback state: the host's controls are mutations on a database row, and every client derives its own playhead from the same timestamped snapshot. One clean websocket event replaces user-orchestrated tab alignment.


## Why not an existing tool

- **Spotify group sessions** are excellent for lean-back playlist listening, but there is no video, the catalog is limited to Spotify, and free accounts cannot host. Jamzy targets the YouTube catalog, which covers almost everything, including live streams.
- **Discord bots and screen share** either drop the video surface entirely or re-encode decent audio into a call stream. Neither gives you a structured queue, a room code you can share, or a persistent room.
- **Watch2gether-style sites** share video, but as a user you cannot inspect how any of it works. Jamzy is meant to be read: the sync protocol, the authorization model, and the rate-limiting boundaries are all visible in this repository.

If you only want uninterrupted background playlists, existing services are a better fit. Jamzy earns its keep as a working implementation of server-authoritative real-time sync with a full persistence and deployment story.

## Features

Playback and room

- Server-authoritative playback: play, pause, seek, and track changes are host-only server events. Each mutation is persisted to PostgreSQL as a timestamped snapshot, then broadcast to the room.
- Time-derived playhead: every client computes `playbackPosition + (now - playbackUpdatedAt)` locally, so bring-your-own-latency never delays the picture.
- Drift correction: a 5-second client interval re-derives the expected position and silently reseeks if the live player is more than 2 seconds off, adjusted for playback speed.
- 30-track playback history so the host can step back to previous tracks.
- Local-only volume and mute with persistence across reloads, plus 0.5x to 2x speed control.
- Host transfer with a 30-second disconnect grace period: a refresh or network blip never hands the room away. Otherwise the longest-standing online member is promoted, and the host can promote anyone manually.
- Persistent membership: rooms, members, and host role live in PostgreSQL, independent of any socket connection. A room is deleted only when its last member leaves.

Search and queue

- Server-side proxy over the YouTube Data API v3 has search, single-video resolution (watch URLs, Shorts, embeds, `youtu.be`), and playlist import of up to 50 tracks. API keys never reach the browser.
- Shared per-room queue of up to 50 items with host-only add, remove, reorder, and clear, plus automatic advance when the current track ends.
- Thumbnail URLs in track payloads are restricted to YouTube CDN hosts so no arbitrary URLs can be pointed at members.

Chat and presence

- Room chat with a 500-character cap, validated server-side, written to PostgreSQL before broadcast, and replayed on join.
- Emoji picker, auto-scroll, and an unread-messages pill.
- Ephemeral presence with multi-tab socket awareness and a live member roster marked with the current host.

Platform

- Clerk authentication: session JWTs verified by the backend on both REST requests and the socket handshake, with an allowed-origins (`azp`) check.
- Username onboarding with unique, validated handles.
- Per-user fixed-window rate limiting on every REST route and socket event, answered with `429` and `Retry-After`.
- Embedded SQL migrations applied automatically at startup under an advisory lock.
- 56 test functions across 12 Go test files, including an end-to-end suite and a timezone-regression test.

## Installation

Requirements: Go 1.27 or newer, Node.js 20 or newer, npm, and a PostgreSQL 14+ database such as [Neon](https://neon.tech/) (or a local instance). You also need free credentials: a [Clerk](https://clerk.com/) application (sign-in, publishable and secret keys) and a YouTube Data API v3 key from Google Cloud Console.

Clone and install:

```bash
git clone https://github.com/mddanish004/Jamzy.git
cd Jamzy

cd server && go mod download && cd ..
cd client && npm install && cd ..
```

Configure the backend by creating `server/.env`:

```env
PORT=5000
NODE_ENV=development
CORS_ORIGIN=http://localhost:5173
DATABASE_URL=postgresql://user:password@host/dbname?sslmode=require
CLERK_PUBLISHABLE_KEY=pk_test_or_live_key
CLERK_SECRET_KEY=sk_test_or_live_key
YOUTUBE_API_KEY=your_api_key
```

`CORS_ORIGIN` is a comma-separated list of allowed browser origins. `CLERK_AUTHORIZED_PARTIES` may also be set to a comma-separated list of origins allowed to mint session tokens; it defaults to `CORS_ORIGIN`.

Configure the frontend by creating `client/.env`:

```env
VITE_API_BASE_URL=http://localhost:5000/api
VITE_CLERK_PUBLISHABLE_KEY=pk_test_or_live_key
```

`VITE_API_BASE_URL` must end with `/api`; the socket URL is derived from it by stripping the `/api` suffix. Environment templates live in each app as `.env.example`.

## Running

Start the backend in one terminal:

```bash
cd server
go run ./cmd/api
```

On boot it connects to PostgreSQL, applies any pending embedded migrations, and listens on `0.0.0.0:$PORT` (port 5000 by default). `SIGINT` and `SIGTERM` shut it down gracefully.

Start the frontend in another terminal:

```bash
cd client
npm run dev
```

Then open `http://localhost:5173`, sign in through Clerk, complete the username onboarding, and create a room from the home screen. Open the room URL (or the six-character code) in a second browser or profile to act as a second listener. The room page shows the synchronized player, the queue panel, the roster, and chat.

Production builds and scripts are documented in the per-app READMEs. There are also repository-root scripts (`npm run dev:server`, `npm run dev:client`, `npm run build:client`, `npm test`) that wrap the Go and npm commands above.

## How it works

### Architecture

```mermaid
graph LR
    B["Browser<br/>React 19 SPA"]

    subgraph S["Go API server (Render)"]
        REST["REST /api/*<br/>net/http ServeMux"]
        AUTH["Auth middleware<br/>Clerk JWT + rate limits"]
        RT["Socket.IO server<br/>presence, queue, host transfer"]
        SVC["Domain services<br/>users, rooms, playback, messages, music"]
    end

    PG[("PostgreSQL<br/>Neon")]
    YT["YouTube Data API v3"]
    CLERK["Clerk"]

    B -- "REST + Bearer JWT" --> REST
    B -- "Socket.IO v4 over WebSocket" --> RT
    REST --> AUTH --> SVC
    RT --> CLERK & SVC
    SVC --> PG
    SVC --> YT
```

Persistent state (users, rooms, memberships, messages, playback snapshots) lives in PostgreSQL. Ephemeral state (socket presence, the queue, the host-transfer countdown) lives in memory in the Go process for the lifetime of a session, which keeps the real-time path free of database write amplification. The two states meet in one place: the `room:join` acknowledgment returns presence, the persisted playback snapshot, the in-memory queue, and recent chat history in a single round trip.

### The synchronization protocol

Playback mutations are server rules, not client suggestions. When the host plays, pauses, seeks, or loads a track, the server:

1. Verifies the sender's identity and that they are the room's current host,
2. Writes the new snapshot (track, position, playing flag) and a wall-clock timestamp `playback_updated_at` to the room row in PostgreSQL,
3. Acknowledges to the host and broadcasts a `playback:update` snapshot to every member.

Every client, including the host, then derives its own playhead continuously rather than waiting for per-frame updates:

```
expected position = playbackPosition + (now - playbackUpdatedAt)
```

A 5-second monitor compares the live YouTube player time against this expectation and reseeks when the difference exceeds 2 seconds, which absorbs tab throttling, buffering hiccups, and speed changes (the elapsed time is multiplied by the playback rate).

A time-and-timezone bug worth reading: the timestamp stored alongside playback was a plain PostgreSQL `timestamp`. Under a non-UTC database time zone, the write path produced zone-local literals while the driver UTC-normalized reads, which zeroed out the elapsed-time math and pinned every joiner to the last host control. The fix writes UTC wall clock explicitly (`timezone('UTC', now())`) in all three playback writers, and a committed regression test pins the database to `Asia/Kolkata` and fails against the old writes: [`server/internal/httpapi/playback_zone_test.go`](server/internal/httpapi/playback_zone_test.go).

### Authorization and safety

Identity comes from Clerk on every path. REST handlers resolve a verified session token to an application user; socket events carry the same token in the handshake. Authorization is server-side on every mutation: members can join, chat, and receive state; only the host can move playback or the queue. Payloads are validated server-side regardless of what the UI allows, rate limits are per user, and panics in realtime handlers are recovered so one bad event cannot take the process down.

## Tests

```bash
cd server
go test ./...
```

Unit tests sit next to the packages they cover (validation, rate limiting, middleware, host transfer, acknowledgment handling, room codes) and need no database. The repository root also has `npm test` aliased to this command.

The end-to-end suite drives the full stack, including a hand-written Socket.IO v4 client, against a real database. It is skipped unless `TEST_DATABASE_URL` is set:

```bash
TEST_DATABASE_URL=postgresql://user:pass@127.0.0.1:5432/jamzy_scratch \
  go test ./internal/httpapi/ -count=1
```

> [!WARNING]
> The e2e suite truncates the `messages`, `room_members`, `rooms`, and `users` tables. Point `TEST_DATABASE_URL` at a disposable local database only.

The client is checked by `npm run build` in `client/`, which runs the TypeScript compiler in strict mode before bundling. CI on every push runs the full gate: `gofmt`, `go vet`, `go build`, `go test`, and the client build ([`.github/workflows/ci.yml`](.github/workflows/ci.yml)).

## Configuration

Backend (`server/.env` or environment):

| Variable | Required | Default | Purpose |
| --- | --- | --- | --- |
| `DATABASE_URL` | yes | - | PostgreSQL connection string |
| `CLERK_SECRET_KEY` | yes | - | Verifies session tokens; without it every authenticated request is rejected |
| `CLERK_PUBLISHABLE_KEY` | no | - | Parity with the client; not used by the API |
| `YOUTUBE_API_KEY` | no | - | YouTube Data API key; music endpoints answer `503` without it |
| `PORT` | no | `5000` | HTTP listen port |
| `NODE_ENV` | no | `development` | Set to `production` for production behavior |
| `CORS_ORIGIN` | no | `http://localhost:5173` | Comma-separated allowed origins |
| `CLERK_AUTHORIZED_PARTIES` | no | `CORS_ORIGIN` | Origins allowed to mint session tokens (`azp` check) |

Frontend (`client/.env` or hosting dashboard):

| Variable | Required | Default | Purpose |
| --- | --- | --- | --- |
| `VITE_API_BASE_URL` | no | `http://localhost:5000/api` | Base URL of the backend; must end with `/api` |
| `VITE_CLERK_PUBLISHABLE_KEY` | no | - | Clerk publishable key; it is public by design |

## Deployment

The running deployment:

| Piece | Where | Notes |
| --- | --- | --- |
| Frontend | Vercel, [jamzy.mddanish.dev](https://jamzy.mddanish.dev/) | SPA rewrites and security headers in `client/vercel.json` |
| API | Render via the [`render.yaml`](render.yaml) blueprint | Go build, `/api/health` probe |
| Database | Neon serverless PostgreSQL | TLS via `sslmode=require` |
| Auth | Clerk | Custom production domain configured in the Clerk dashboard |
| CI | GitHub Actions | Format, vet, and tests on every push |

To reproduce it: create a Render web service from this repository using the blueprint and set the backend environment variables there; import the repository into Vercel with the root directory set to `client/` and set the client variables; point the app's Clerk instance at your domain and mirror that origin in `CORS_ORIGIN` and `CLERK_AUTHORIZED_PARTIES`.

## Repository layout

```
client/                      React 19 + Vite SPA
  src/components/            Player, chat, queue, search, landing modules
  src/hooks/                 Socket session, player control, user cache
  src/pages/                 Landing, onboarding, home, room
server/                      Go API (module github.com/mddanish004/jamzy/server)
  cmd/api/                   Wiring: config, pool, migrations, server, shutdown
  internal/                  Domain packages; see server/README.md for the map
  db/migrations/             Embedded SQL, applied on startup
  prisma/schema.prisma       Readable data model reference
render.yaml                  Render deployment blueprint
.github/workflows/ci.yml     CI gate
```

## Notes and limitations

> [!NOTE]
> On the free tiers of Render and Neon, services idle to sleep. The first request after a pause can take 30 to 60 seconds while the API instance boots and the database wakes.

> [!NOTE]
> Ephemeral room state is intentionally single-node: presence, queue, and host-transfer countdown live in the process memory. Scaling to multiple API instances would need a Socket.IO adapter (Redis) and a shared store for that state. This is a deliberate scope decision for a social app at personal scale, not an oversight.

> [!NOTE]
> Playhead derivation compares the browser clock to server wall clock. Clients with a badly skewed system clock can drift up until the next reseek.

Built and maintained by [mddanish004](https://github.com/mddanish004). ISC license, as declared in `package.json`.
