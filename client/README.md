# Jamzy Web Client

The React single-page app for [Jamzy](../README.md): the landing page, Clerk sign-in and username onboarding, the room lobby, and the synchronized listening room with its player, queue, roster, and chat. It talks to the Go API over authenticated fetches and to the server's Socket.IO layer over WebSockets.

Deployed at [jamzy.mddanish.dev](https://jamzy.mddanish.dev/) on Vercel; see [Deployment](#deployment).

## Stack

| Category | Technology |
| --- | --- |
| Framework | [React 19.2](https://react.dev/) + [TypeScript 6](https://www.typescriptlang.org/) |
| Build tool | [Vite 8](https://vite.dev/) with the [React Compiler](https://react.dev/learn/react-compiler) enabled via Babel |
| Styling | [Tailwind CSS v4](https://tailwindcss.com/) with Material-3-style `@theme` tokens |
| Routing | [React Router 7](https://reactrouter.com/) |
| Real time | [socket.io-client](https://socket.io/) v4 |
| Authentication | [Clerk React](https://clerk.com/) with embedded `<SignIn />` / `<SignUp />` components |
| Media | YouTube IFrame Player API, loaded as a script |

Routes: the public landing page, `/signin` and `/signup`, `/onboarding` for first-time username selection, `/home` for the room lobby, and `/room/:roomCode` for the listening room. Route guards redirect unsigned or un-onboarded users before the room UI ever mounts.

## Requirements

- Node.js 20 or newer
- The Jamzy API running locally on port 5000, or its deployed URL
- A Clerk application publishable key

## Getting started

Copy the template and fill in two values:

```bash
cd client
cp .env.example .env
```

```env
VITE_API_BASE_URL=http://localhost:5000/api
VITE_CLERK_PUBLISHABLE_KEY=pk_test_or_live_key
```

`VITE_API_BASE_URL` must end with `/api`; the WebSocket base is derived from it by stripping that suffix. `VITE_CLERK_PUBLISHABLE_KEY` is the Clerk publishable key from your Clerk dashboard, which is public by design and safe to expose to the browser.

Install, then run the dev server:

```bash
npm install
npm run dev
```

The app is served at `http://localhost:5173`, which matches the API backend's default allowed CORS origin.

## Scripts

| Script | Command | Description |
| --- | --- | --- |
| `dev` | `vite` | Development server with hot module replacement |
| `build` | `tsc -b && vite build` | Strict type-check, then a production bundle into `dist/` |
| `preview` | `vite preview` | Serve the production build locally |
| `lint` | `eslint .` | ESLint across TypeScript and React files |

There are no client-side unit tests; correctness is verified by `npm run build`'s strict `tsc` pass and by the server's tests, which cover the API and socket contracts the client depends on.

## How it works

### Single-fetch user session (`CurrentUserProvider`)

The profile fetch for `/api/me` happens once per sign-in and is cached in React context. Route guards read the cached result to decide between the room app and the onboarding screen, so navigation never re-requests the profile. The onboarding response seeds the cache directly, avoiding a refetch race between route transition and profile creation.

### One acknowledgment per room entry (`useRoomSocket`)

The hook owns the Socket.IO connection lifecycle. It passes the Clerk session token in the handshake, which the server verifies before any room access. `room:join` is issued once per room and its acknowledgment carries everything needed to render the room in one round trip: the presence roster, the persisted playback snapshot, the in-memory queue, and recent chat history. Every server-to-client event (`presence:update`, `playback:update`, `queue:update`, `chat:new`, `host:update`, `host:pending`) updates a small piece of state, and reconnection re-syncs without a page reload.

### Playback derivation and drift correction (`useYouTubePlayer`)

The player controls the YouTube IFrame API. Rather than trusting the local player's timeline, the hook derives the authoritative playhead from the server snapshot:

```
expected = playbackPosition + (now - playbackUpdatedAt)
```

A 5-second interval compares the live player time to this expectation and silently reseeks when the difference exceeds 2 seconds. The elapsed time is multiplied by the playback rate, so speed changes stay in sync. While a remote update is being applied, local event listeners are suppressed so the correction cannot echo back to the server as a new host-editing event.

## Structure

```text
client/
├── public/                  Favicon and album artwork
├── src/
│   ├── components/
│   │   ├── landing/         Landing sections (hero, showcase, story, telemetry, footer)
│   │   ├── ui/              Shared atoms: logo, avatar, buttons, emoji picker, equalizer bars
│   │   ├── YouTubePlayer.tsx   Synchronized player with the full control deck
│   │   ├── ChatPanel.tsx    Persistent chat with scroll anchoring
│   │   ├── QueuePanel.tsx   Host-only queue: drag reorder, skip, clear
│   │   ├── RoomNavbarSearch.tsx  Header search with pasted-URL previews
│   │   ├── AuthStatusScreen.tsx  Auth error and retry states
│   │   └── ProtectedRoute.tsx   Auth and onboarding guards
│   ├── pages/               Landing, RootGuard, auth, onboarding, home, Room.tsx
│   ├── hooks/               useRoomSocket, useYouTubePlayer, useCurrentUser
│   ├── providers/           CurrentUserProvider and its context
│   ├── services/            Typed fetch client (api.ts) and shared DTOs (types.ts)
│   ├── config/env.ts        Runtime URL and key configuration
│   └── App.tsx              Route definitions
├── vercel.json              SPA rewrites and security response headers
└── vite.config.ts           Vite + Tailwind + React Compiler Babel setup
```

## Design system

Styling is Tailwind v4 with a Material-3-style token set in `src/index.css`, on a dark obsidian base:

| Token | Value | Role |
| --- | --- | --- |
| `--color-background` | `#0e0e10` | Obsidian canvas |
| `--color-primary` | `#fb7185` | Rose primary for interactive elements |
| `--color-secondary` | `#fda4af` | Soft blush supporting accents |
| `--color-tertiary` | `#10b981` | Emerald for live, synced, and online states |
| `--color-error` | `#ffb4ab` | Error foreground |

Type roles use a size scale that mirrors the token names (`label-sm` through `headline-xl`), set in Plus Jakarta Sans; Geist is reserved for code-style labels; Material Symbols Outlined provides icons. Responsive behavior is part of the layout work rather than a separate breakpoint pass: the room page collapses to a single column and the landing page reflows per section on narrow screens.

## Deployment

The client is a static SPA deployed to Vercel from this repository, with the root directory set to `client/`. [`vercel.json`](vercel.json) provides both pieces the SPA needs:

1. All deep links (for example a pasted `/room/A1B2C3` URL) rewrite to `/index.html` so React Router owns every path.
2. Security response headers (`X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, `Permissions-Policy`, `Strict-Transport-Security`) are applied to every route.

Set `VITE_API_BASE_URL` (the deployed API URL ending in `/api`) and `VITE_CLERK_PUBLISHABLE_KEY` in the Vercel project's environment variables, then redeploy so the new values are baked into the bundle. The Clerk instance must recognize the deployed domain; see the deployment section of the root README for the full wiring across Vercel, Render, and Clerk.
