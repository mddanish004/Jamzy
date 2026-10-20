import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth, UserButton } from "@clerk/react";
import { createRoom, joinRoom } from "../services/api";
import { useCurrentUser } from "../hooks/useCurrentUser";
import { Logo } from "../components/ui/Logo";

const LAST_ROOM_KEY = "jamzy:lastRoom";

function readLastRoom(): string | null {
  try {
    return localStorage.getItem(LAST_ROOM_KEY);
  } catch {
    return null;
  }
}

export function Home() {
  const { getToken } = useAuth();
  const navigate = useNavigate();
  const currentUser = useCurrentUser();

  const [roomCode, setRoomCode] = useState("");
  const [creating, setCreating] = useState(false);
  const [joining, setJoining] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [lastRoom, setLastRoom] = useState<string | null>(readLastRoom);

  const username =
    currentUser.status === "authenticated"
      ? currentUser.me.user?.username
      : undefined;

  useEffect(() => {
    void getToken();
  }, [getToken]);

  function rememberRoom(code: string) {
    try {
      localStorage.setItem(LAST_ROOM_KEY, code);
    } catch {
    }
    setLastRoom(code);
  }

  async function handleCreateRoom() {
    setError(null);
    setCreating(true);
    try {
      const { room } = await createRoom(getToken);
      rememberRoom(room.roomCode);
      navigate(`/room/${room.roomCode}`, { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create room");
    } finally {
      setCreating(false);
    }
  }

  async function handleJoinRoom(e: React.FormEvent) {
    e.preventDefault();
    setError(null);

    const trimmed = roomCode.replace(/\s+/g, "").toUpperCase();
    if (trimmed.length === 0) {
      setError("Enter a room code");
      return;
    }

    setJoining(true);
    try {
      const { room } = await joinRoom(getToken, trimmed);
      rememberRoom(room.roomCode);
      navigate(`/room/${room.roomCode}`, { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to join room");
    } finally {
      setJoining(false);
    }
  }

  function formatCode(val: string): string {
    const cleaned = val
      .replace(/[^a-zA-Z0-9]/g, "")
      .toUpperCase()
      .slice(0, 6);
    return cleaned.split("").join(" ");
  }

  async function handlePasteCode() {
    try {
      const clipText = await navigator.clipboard.readText();
      if (clipText) {
        setRoomCode(formatCode(clipText));
      }
    } catch {
    }
  }

  const busy = creating || joining;

  return (
    <div className="relative flex min-h-dvh flex-col overflow-x-clip bg-surface-container-lowest font-body-md text-on-surface antialiased selection:bg-primary selection:text-white">
      <header className="sticky top-0 z-40 flex h-16 shrink-0 items-center justify-between border-b border-obsidian-border/60 bg-surface-container-lowest/85 px-gutter shadow-[0_1px_8px_rgba(0,0,0,0.25)] backdrop-blur-xl">
        <Link to="/" aria-label="Jamzy home" className="shrink-0">
          <Logo size={30} dot={false} />
        </Link>

        <div className="flex items-center gap-space-md">
          {username && (
            <span className="hidden font-label-md text-label-md text-on-surface-variant sm:inline">
              Hi,{" "}
              <span className="font-semibold text-on-surface">{username}</span>
            </span>
          )}

          {lastRoom && (
            <Link
              to={`/room/${lastRoom}`}
              className="flex items-center gap-1.5 rounded-lg border border-obsidian-border bg-surface-container-low px-space-md py-1.5 font-mono text-[13px] tracking-wider text-primary transition-colors hover:border-primary/40"
              title="Return to your recent room"
            >
              <span className="material-symbols-outlined text-[16px] text-on-surface-variant">
                tag
              </span>
              <span>{lastRoom}</span>
            </Link>
          )}

          <UserButton />
        </div>
      </header>

      <main className="relative min-h-0 flex-1">
        <div className="pointer-events-none absolute inset-0 overflow-hidden" aria-hidden="true">
          <div className="absolute -top-32 left-1/2 h-[480px] w-[720px] -translate-x-1/2 rounded-full bg-primary/10 opacity-70 blur-[140px]" />
          <div className="absolute top-[500px] left-1/4 h-[380px] w-[380px] rounded-full bg-secondary-container/15 opacity-50 blur-[130px]" />
          <div className="absolute top-[360px] right-1/4 h-[320px] w-[320px] rounded-full bg-tertiary/5 opacity-40 blur-[110px]" />
        </div>

        <div className="relative z-10 mx-auto flex w-full max-w-4xl flex-col items-center px-gutter pb-space-xl pt-space-xl">
          <div className="mx-auto mb-space-xl flex max-w-xl flex-col items-center text-center">
            <h1 className="relative mb-3 font-headline-xl-mobile text-headline-xl-mobile font-bold tracking-tight text-on-surface sm:text-[52px] sm:leading-[60px]">
              Listen together
              <span className="text-primary">.</span>
              <span className="absolute -inset-x-4 -bottom-2 h-[1px] bg-gradient-to-r from-transparent via-primary/30 to-transparent sm:-inset-x-8" />
            </h1>
            <p className="font-body-lg text-body-lg leading-relaxed text-on-surface-variant">
              Start a new room or join one with a code.
            </p>
          </div>

          {error && (
            <p className="mb-space-md w-full max-w-xl rounded-xl border border-error-container/60 bg-error/10 px-4 py-2 text-center font-body-md text-body-md text-error">
              {error}
            </p>
          )}

          <div className="flex w-full max-w-xl flex-col gap-space-lg">
            <section className="group relative overflow-hidden rounded-xl border border-obsidian-border bg-surface-container-low/95 p-space-lg shadow-md backdrop-blur-xl transition-all duration-300 hover:border-primary/40 hover:shadow-[0_12px_40px_-10px_rgba(251,113,133,0.18)]">
              <div className="pointer-events-none absolute -right-16 -top-16 h-36 w-36 rounded-full bg-gradient-to-br from-primary/20 via-primary-container/10 to-transparent blur-2xl transition-transform duration-500 group-hover:scale-125" />

              <div className="relative z-10 flex items-start gap-space-md">
                <div className="grid size-12 shrink-0 place-items-center rounded-xl border border-obsidian-border bg-surface-container-high text-primary shadow-sm transition-colors duration-300 group-hover:border-primary group-hover:bg-primary group-hover:text-on-primary">
                  <span className="material-symbols-outlined text-[26px]">
                    podcasts
                  </span>
                </div>
                <div className="flex flex-col">
                  <span className="mb-0.5 font-label-sm text-label-sm font-semibold uppercase tracking-widest text-primary">
                    Host session
                  </span>
                  <h2 className="font-headline-sm text-headline-sm tracking-tight text-on-surface">
                    Create a room
                  </h2>
                  <p className="mt-1 font-body-md text-body-md leading-normal text-on-surface-variant">
                    You'll be the host. Share the code to invite others.
                  </p>
                </div>
              </div>

              <div className="mt-space-md flex flex-wrap items-center gap-2 border-t border-obsidian-border/40 pt-space-md">
                {[
                  { icon: "graphic_eq", tone: "text-tertiary", label: "Real-time sync" },
                  { icon: "groups", tone: "text-primary", label: "Unlimited listeners" },
                  { icon: "tune", tone: "text-secondary", label: "Host-led queue" },
                ].map((badge) => (
                  <span
                    key={badge.label}
                    className="inline-flex items-center gap-1.5 rounded-lg border border-obsidian-border bg-surface-container px-2.5 py-1 font-label-md text-label-md text-on-surface-variant"
                  >
                    <span className={`material-symbols-outlined text-[15px] ${badge.tone}`}>
                      {badge.icon}
                    </span>
                    {badge.label}
                  </span>
                ))}
              </div>

              <div className="mt-space-lg">
                <button
                  type="button"
                  onClick={handleCreateRoom}
                  disabled={busy}
                  className="flex h-12 w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-primary via-primary-container to-secondary font-headline-sm text-[16px] font-semibold text-on-primary shadow-[0_0_24px_-4px_rgba(251,113,133,0.4)] transition-all duration-200 hover:brightness-105 hover:shadow-[0_0_32px_0px_rgba(251,113,133,0.55)] active:scale-[0.99] disabled:pointer-events-none disabled:opacity-60"
                >
                  {creating ? (
                    <>
                      <span className="material-symbols-outlined animate-spin text-[20px]">
                        progress_activity
                      </span>
                      <span>Spawning room…</span>
                    </>
                  ) : (
                    <>
                      <span className="material-symbols-outlined msr-fill text-[20px]">
                        music_note
                      </span>
                      <span>Create room</span>
                    </>
                  )}
                </button>
              </div>
            </section>

            <div className="relative my-1 flex items-center justify-center">
              <div className="h-[1px] w-full bg-gradient-to-r from-transparent via-obsidian-border to-transparent" />
              <div className="absolute flex items-center gap-1.5 rounded-full border border-obsidian-border bg-surface-container-low px-4 py-1 font-label-sm text-label-sm uppercase tracking-[0.25em] text-on-surface-variant shadow-sm">
                <span className="h-1 w-1 rounded-full bg-primary" />
                <span>OR</span>
                <span className="h-1 w-1 rounded-full bg-primary" />
              </div>
            </div>

            <section className="group relative overflow-hidden rounded-xl border border-obsidian-border bg-surface-container-low/95 p-space-lg shadow-md backdrop-blur-xl transition-all duration-300 hover:border-primary/40 hover:shadow-xl">
              <div className="relative z-10">
                <div className="mb-space-md flex items-start gap-space-md">
                  <div className="grid size-12 shrink-0 place-items-center rounded-xl border border-obsidian-border bg-surface-container-high text-on-surface-variant shadow-sm transition-colors duration-300 group-hover:border-primary/40 group-hover:text-primary">
                    <span className="material-symbols-outlined text-[26px]">key</span>
                  </div>
                  <div className="flex flex-col">
                    <span className="mb-0.5 font-label-sm text-label-sm font-semibold uppercase tracking-widest text-on-surface-variant">
                      Guest &amp; collaborator
                    </span>
                    <h2 className="font-headline-sm text-headline-sm tracking-tight text-on-surface">
                      Join a room
                    </h2>
                    <p className="mt-1 font-body-md text-body-md leading-normal text-on-surface-variant">
                      Enter the unique 6-character sanctuary access key.
                    </p>
                  </div>
                </div>

                <form onSubmit={handleJoinRoom} className="flex flex-col gap-space-md">
                  <div className="relative">
                    <label className="sr-only" htmlFor="room-code-input">
                      Room Code
                    </label>
                    <input
                      id="room-code-input"
                      value={roomCode}
                      onChange={(e) => setRoomCode(formatCode(e.target.value))}
                      onPaste={(e) => {
                        e.preventDefault();
                        setRoomCode(formatCode(e.clipboardData.getData("text")));
                      }}
                      autoComplete="off"
                      maxLength={11}
                      disabled={busy}
                      spellCheck={false}
                      placeholder="A B 7 2 K D"
                      className={`h-14 w-full rounded-xl border border-obsidian-border bg-surface-container-lowest px-space-md text-center font-mono text-headline-sm uppercase tracking-[0.28em] text-primary shadow-inner outline-none transition-all duration-200 placeholder:text-outline-variant/60 hover:border-[#3d3c4a] focus:border-primary focus:shadow-[0_0_0_2px_#fb7185,0_0_24px_-4px_rgba(251,113,133,0.3)] disabled:opacity-50 ${
                        error ? "animate-shake" : ""
                      }`}
                    />
                    <div className="absolute right-3 top-1/2 flex -translate-y-1/2 items-center">
                      <button
                        type="button"
                        onClick={handlePasteCode}
                        className="pointer-coarse:p-3 p-2 text-on-surface-variant transition-colors hover:text-primary"
                        title="Paste clipboard"
                        aria-label="Paste clipboard"
                      >
                        <span className="material-symbols-outlined text-[18px]">
                          content_paste
                        </span>
                      </button>
                    </div>
                  </div>

                  <button
                    type="submit"
                    disabled={busy}
                    className="flex h-12 w-full items-center justify-center gap-2 rounded-xl border border-obsidian-border bg-surface-container-high px-space-lg font-headline-sm text-[16px] font-semibold text-on-surface shadow-sm transition-all duration-200 hover:border-primary/30 hover:bg-surface-bright active:scale-[0.99] disabled:pointer-events-none disabled:opacity-40"
                  >
                    {joining ? (
                      <>
                        <span className="material-symbols-outlined animate-spin text-[20px]">
                          progress_activity
                        </span>
                        <span>Synchronizing…</span>
                      </>
                    ) : (
                      <>
                        <span className="material-symbols-outlined text-[20px]">
                          login
                        </span>
                        <span>Join room</span>
                      </>
                    )}
                  </button>
                </form>

                <div className="mt-space-md flex flex-wrap items-center justify-between gap-y-0.5 border-t border-obsidian-border/40 pt-space-md font-label-md text-label-md text-on-surface-variant">
                  {lastRoom ? (
                    <span className="flex items-center gap-1.5">
                      <span className="material-symbols-outlined text-[16px] text-tertiary">
                        history
                      </span>
                      <span>Recent room:</span>
                      <button
                        type="button"
                        onClick={() => setRoomCode(formatCode(lastRoom))}
                        className="font-mono text-primary transition-colors hover:text-secondary hover:underline"
                        title="Fill the code"
                      >
                        {lastRoom}
                      </button>
                    </span>
                  ) : (
                    <span className="flex items-center gap-1.5">
                      <span className="material-symbols-outlined text-[16px] text-tertiary">
                        history
                      </span>
                      <span>No rooms yet</span>
                    </span>
                  )}
                  <span className="font-label-sm text-label-sm tracking-wider text-outline">
                    INVITE-ONLY KEY
                  </span>
                </div>
              </div>
            </section>

            <div className="flex w-full items-center justify-center gap-1 py-space-sm opacity-35 transition-opacity hover:opacity-85">
              <svg className="h-6 w-48 text-primary" fill="currentColor" viewBox="0 0 192 24" aria-hidden="true">
                {[
                  [0, 8, 8], [6, 5, 14], [12, 10, 4], [18, 2, 20], [24, 6, 12],
                  [30, 9, 6], [36, 3, 18], [42, 7, 10], [48, 11, 2], [54, 4, 16],
                  [60, 8, 8], [66, 1, 22], [72, 6, 12], [78, 9, 6], [84, 4, 16],
                  [90, 7, 10], [96, 2, 20], [102, 8, 8], [108, 5, 14], [114, 9, 6],
                  [120, 3, 18], [126, 7, 10], [132, 10, 4], [138, 2, 20],
                  [144, 6, 12], [150, 8, 8], [156, 4, 16], [162, 10, 4],
                  [168, 6, 12], [174, 3, 18], [180, 8, 8], [186, 11, 2],
                ].map(([x, y, h], i) => (
                  <rect key={i} x={x} y={y} width={2} height={h} rx={1} />
                ))}
              </svg>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
