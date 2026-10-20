import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useAuth, UserButton } from "@clerk/react";
import { getRoom, leaveRoom } from "../services/api";
import { useRoomSocket, type ConnectionStatus } from "../hooks/useRoomSocket";
import { useCurrentUser } from "../hooks/useCurrentUser";
import { RoomNavbarSearch } from "../components/RoomNavbarSearch";
import { YouTubePlayer } from "../components/YouTubePlayer";
import { ChatPanel } from "../components/ChatPanel";
import { QueuePanel } from "../components/QueuePanel";
import { Avatar } from "../components/ui/Avatar";
import type { RoomDTO, YouTubeSearchResult } from "../services/types";

const LAST_ROOM_KEY = "jamzy:lastRoom";

export function Room() {
  const { roomCode } = useParams<{ roomCode: string }>();
  const { getToken } = useAuth();
  const navigate = useNavigate();
  const currentUser = useCurrentUser();

  const [room, setRoom] = useState<RoomDTO | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [peopleOpen, setPeopleOpen] = useState(false);

  useEffect(() => {
    if (!roomCode) return;

    let cancelled = false;

    (async () => {
      setLoading(true);
      setError(null);
      try {
        const { room: data } = await getRoom(getToken, roomCode.toUpperCase());
        if (!cancelled) setRoom(data);
      } catch (err) {
        if (!cancelled) {
          const message =
            err instanceof Error ? err.message : "Failed to load room";
          setError(message);
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [roomCode, getToken]);

  useEffect(() => {
    if (!roomCode) return;
    try {
      localStorage.setItem(LAST_ROOM_KEY, roomCode.toUpperCase());
    } catch {
    }
  }, [roomCode]);

  const {
    presence,
    connectionStatus,
    playback,
    playbackError,
    setTrack,
    sendControl,
    clearTrack,
    messages,
    chatError,
    chatSending,
    sendChatMessage,
    queue,
    addToQueue,
    prependToQueue,
    removeFromQueue,
    reorderQueue,
    clearQueue,
    advanceQueue,
    hostUserId,
    hostNotice,
    dismissHostNotice,
    hostPending,
    hostError,
    transferHost,
  } = useRoomSocket(
    room ? room.roomCode : null,
    room ? room.members : [],
    currentUser.status === "authenticated" ? currentUser.me.user ?? undefined : undefined,
  );

  const socketReady = connectionStatus === "connected";
  const searchInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        searchInputRef.current?.focus();
        searchInputRef.current?.select();
      }
    }
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

  const [previousTracks, setPreviousTracks] = useState<YouTubeSearchResult[]>([]);
  const [prevRoomCode, setPrevRoomCode] = useState(roomCode);
  if (roomCode !== prevRoomCode) {
    setPrevRoomCode(roomCode);
    setPreviousTracks([]);
  }
  const lastActiveVideoRef = useRef<YouTubeSearchResult | null>(null);
  const isNavigatingPreviousRef = useRef(false);

  useEffect(() => {
    lastActiveVideoRef.current = null;
    isNavigatingPreviousRef.current = false;
  }, [roomCode]);

  useEffect(() => {
    if (!playback?.videoId) {
      lastActiveVideoRef.current = null;
      return;
    }

    const currentVideo: YouTubeSearchResult = {
      videoId: playback.videoId,
      title: playback.title || "Untitled",
      channelTitle: "",
      thumbnailUrl: playback.thumbnailUrl || "",
      duration: playback.duration ? `PT${playback.duration}S` : "",
    };

    if (
      lastActiveVideoRef.current &&
      lastActiveVideoRef.current.videoId !== playback.videoId
    ) {
      if (!isNavigatingPreviousRef.current) {
        const prev = lastActiveVideoRef.current;
        setPreviousTracks((history) => [...history.slice(-29), prev]);
      } else {
        isNavigatingPreviousRef.current = false;
      }
    }

    lastActiveVideoRef.current = currentVideo;
  }, [playback?.videoId, playback?.title, playback?.thumbnailUrl, playback?.duration]);

  const myUserId =
    currentUser.status === "authenticated" ? currentUser.me.user?.id : undefined;
  const effectiveHostId = hostUserId ?? room?.host.id ?? null;
  const isHost = Boolean(myUserId && effectiveHostId === myUserId);
  const canControlPlayback = socketReady && isHost;

  const handlePreviousTrack = useCallback(() => {
    if (!canControlPlayback || previousTracks.length === 0) return;
    const newHistory = [...previousTracks];
    const trackToPlay = newHistory.pop();
    if (!trackToPlay) return;

    isNavigatingPreviousRef.current = true;
    setPreviousTracks(newHistory);

    if (lastActiveVideoRef.current) {
      prependToQueue(lastActiveVideoRef.current);
    }

    setTrack(trackToPlay);
  }, [canControlPlayback, previousTracks, prependToQueue, setTrack]);

  useEffect(() => {
    if (!hostNotice) return;
    const timer = setTimeout(dismissHostNotice, 7000);
    return () => clearTimeout(timer);
  }, [hostNotice, dismissHostNotice]);

  async function handleCopyCode() {
    if (!room) return;
    try {
      await navigator.clipboard.writeText(room.roomCode);
      setCopied(true);
      setTimeout(() => setCopied(false), 2200);
    } catch {
    }
  }

  async function handleLeave() {
    if (!roomCode) return;
    navigate("/home", { replace: true });
    leaveRoom(getToken, roomCode.toUpperCase()).catch((err) => {
      console.error("leaveRoom failed:", err);
    });
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-background font-body-md text-on-surface antialiased">
        <div className="flex min-h-screen flex-col items-center justify-center gap-3">
          <span className="material-symbols-outlined animate-pulse text-[32px] text-primary">
            graphic_eq
          </span>
          <p className="font-label-md text-label-md text-on-surface-variant">
            Loading room…
          </p>
        </div>
      </div>
    );
  }

  if (error || !room) {
    return (
      <div className="min-h-screen bg-background font-body-md text-on-surface antialiased">
        <main className="flex min-h-screen flex-col items-center justify-center px-5">
          <div className="w-full max-w-md space-y-4 rounded-2xl border border-obsidian-border bg-surface-container-low p-8 text-center shadow-2xl">
            <span className="material-symbols-outlined text-[32px] text-error">
              search_off
            </span>
            <p className="text-body-md font-body-md text-error">
              {error ?? "Room not found"}
            </p>
            <button
              type="button"
              onClick={() => navigate("/home", { replace: true })}
              className="flex h-11 w-full items-center justify-center rounded-xl border border-obsidian-border bg-surface-container-high px-5 font-label-md text-label-md text-on-surface transition-colors hover:bg-surface-bright"
            >
              Back to lobby
            </button>
          </div>
        </main>
      </div>
    );
  }

  return (
    <div className="min-h-dvh bg-background font-body-md text-on-surface antialiased selection:bg-primary selection:text-white">
      <header className="sticky top-0 z-50 border-b border-obsidian-border/60 bg-background/95 backdrop-blur-xl">
        <section className="flex w-full flex-col items-center justify-between gap-space-md px-gutter py-space-sm md:flex-row">
          <div className="flex w-full items-center gap-space-md lg:max-w-xl">
            <RoomNavbarSearch
              onSelect={setTrack}
              disabled={!canControlPlayback}
              canQueue={canControlPlayback && !!playback?.videoId}
              onQueue={addToQueue}
              inputRef={searchInputRef}
              disabledMessage={
                !socketReady
                  ? "Connecting to the room — search unlocks in a moment."
                  : "Only the host can search and add music."
              }
            />
          </div>

          <div className="flex w-full flex-wrap items-center justify-between gap-space-sm md:w-auto md:justify-end">
            <ConnectionCapsule status={connectionStatus} />

            <button
              type="button"
              onClick={() => setPeopleOpen((o) => !o)}
              className="flex items-center gap-1.5 rounded-full border border-obsidian-border bg-surface-container-low px-2.5 py-1.5 font-label-md text-label-md text-on-surface-variant shadow-sm transition-colors hover:border-primary/40 hover:text-on-surface pointer-coarse:px-3 pointer-coarse:py-2.5"
              aria-label={`${presence.length} in room — view people`}
            >
              <span className="material-symbols-outlined text-[16px]">group</span>
              <span className="font-semibold">{presence.length}</span>
            </button>

            <div className="flex items-center gap-2 rounded-full border border-obsidian-border bg-surface-container-low py-1 pl-space-sm pr-1.5">
              <div className="flex items-center gap-1 font-label-md text-label-md tracking-widest text-on-surface">
                <span className="text-[11px] font-medium tracking-normal text-on-surface-variant">
                  CODE:
                </span>
                <span className="font-bold text-primary">{room.roomCode}</span>
              </div>
              <button
                type="button"
                onClick={handleCopyCode}
                title="Copy room invite code"
                className={`pointer-coarse:px-3 pointer-coarse:py-2 flex items-center gap-1 rounded-full border border-obsidian-border/60 bg-surface-container-high px-2.5 py-1 font-label-sm text-label-sm transition-all active:scale-95 hover:bg-surface-bright ${
                  copied ? "text-tertiary" : "text-on-surface"
                }`}
              >
                <span className="material-symbols-outlined text-[14px]">
                  {copied ? "check" : "content_copy"}
                </span>
                <span>{copied ? "Copied!" : "Copy"}</span>
              </button>
            </div>

            <div className="relative flex items-center justify-center rounded-full border border-primary/30 bg-surface-container-high p-0.5">
              <UserButton />
            </div>
            <button
              type="button"
              onClick={handleLeave}
              title="Leave room"
              className="group flex items-center gap-1.5 rounded-xl border border-obsidian-border bg-surface-container-low px-3.5 py-2 font-label-md text-label-md text-on-surface-variant transition-all hover:bg-surface-container-highest hover:text-error pointer-coarse:py-2.5"
            >
              <span className="material-symbols-outlined text-[18px] transition-transform group-hover:translate-x-0.5">
                logout
              </span>
              <span className="font-medium">Leave</span>
            </button>
          </div>
        </section>
      </header>

      {peopleOpen && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setPeopleOpen(false)} />
          <div className="fixed inset-x-gutter bottom-4 z-50 max-h-[70dvh] overflow-y-auto rounded-2xl border border-obsidian-border bg-surface-container p-3 shadow-2xl sm:inset-x-auto sm:bottom-auto sm:right-gutter sm:top-[4.75rem] sm:w-64 sm:overflow-y-auto">
            <p className="mb-3 flex items-center justify-between font-label-sm text-label-sm uppercase tracking-wider text-on-surface-variant">
              <span>People in room</span>
              <span className="rounded-full bg-surface-container-high px-2 py-0.5 text-on-surface-variant">
                {presence.length}
              </span>
            </p>
            <ul className="jamzy-scrollbar max-h-72 space-y-0.5 overflow-y-auto">
              {presence.map((member) => {
                const memberIsHost = member.userId === effectiveHostId;
                return (
                  <li
                    key={member.userId}
                    className="flex items-center gap-2.5 rounded-lg px-2 py-2 transition-colors hover:bg-surface-container-low"
                  >
                    <Avatar name={member.username} size={28} online={member.online} />
                    <span className="flex-1 truncate text-body-md font-body-md font-medium text-on-surface">
                      {member.username}
                    </span>
                    {isHost && !memberIsHost && member.online && (
                      <button
                        type="button"
                        onClick={() => {
                          transferHost(member.userId);
                          setPeopleOpen(false);
                        }}
                        className="rounded-full border border-obsidian-border px-2 py-0.5 font-label-sm text-label-sm uppercase tracking-wider text-on-surface-variant transition-colors hover:border-primary hover:text-primary pointer-coarse:px-3 pointer-coarse:py-1.5"
                      >
                        Make host
                      </button>
                    )}
                    {memberIsHost && (
                      <span className="rounded-full bg-primary/20 px-1.5 py-0.5 font-label-sm text-label-sm uppercase tracking-wider text-primary">
                        Host
                      </span>
                    )}
                  </li>
                );
              })}
            </ul>
          </div>
        </>
      )}

      <main className="relative px-gutter pb-space-xl pt-space-md">
        <div className="pointer-events-none fixed inset-0 overflow-hidden" aria-hidden="true">
          <div className="absolute -top-32 left-1/2 h-[480px] w-[720px] -translate-x-1/2 rounded-full bg-primary/10 opacity-70 blur-[140px]" />
          <div className="absolute top-[420px] left-1/4 h-[380px] w-[380px] rounded-full bg-secondary-container/15 opacity-50 blur-[130px]" />
          <div className="absolute top-[280px] right-1/4 h-[320px] w-[320px] rounded-full bg-tertiary/5 opacity-40 blur-[110px]" />
        </div>

        <div className="relative z-10 mx-auto w-full max-w-[1600px]">
          <HostPendingBanner
            key={hostPending?.deadline}
            pending={hostPending}
          />

          {hostNotice && (
            <div className="mb-space-md mt-space-md flex items-center gap-3 rounded-xl border border-primary/30 bg-primary/10 px-4 py-3">
              <span className="material-symbols-outlined text-[18px] text-primary">
                headphones
              </span>
              <p className="flex-1 font-body-sm text-body-sm text-on-surface">
                <span className="font-semibold">
                  {hostNotice.hostUserId === myUserId
                    ? "You are"
                    : `${hostNotice.hostUsername} is`}
                </span>{" "}
                now the host
                {hostNotice.reason === "disconnect"
                  ? ` — ${hostNotice.previousHostUsername} disconnected.`
                  : hostNotice.reason === "left"
                    ? ` — ${hostNotice.previousHostUsername} left the room.`
                    : "."}
              </p>
              <button
                type="button"
                onClick={dismissHostNotice}
                className="pointer-coarse:px-2.5 pointer-coarse:py-1.5 rounded-full px-2 py-0.5 font-label-sm text-label-sm text-on-surface-variant transition-colors hover:bg-surface-container-high hover:text-on-surface"
                aria-label="Dismiss"
              >
                <span className="material-symbols-outlined text-[14px]">close</span>
              </button>
            </div>
          )}

          {hostError && (
            <p className="mb-space-md mt-space-md flex items-center gap-2 rounded-xl border border-error-container/50 bg-error/10 px-4 py-2 font-body-sm text-body-sm text-error">
              <span className="material-symbols-outlined text-[16px]">error</span>
              {hostError}
            </p>
          )}

          <div className="mt-space-md grid grid-cols-1 items-start gap-gutter lg:grid-cols-12">
            <div className="flex flex-col gap-space-md lg:col-span-8">
              <YouTubePlayer
                playback={playback}
                onControl={canControlPlayback ? sendControl : null}
                onClear={clearTrack}
                onEnded={isHost ? advanceQueue : undefined}
                onSkip={isHost ? advanceQueue : undefined}
                onPrevious={isHost ? handlePreviousTrack : undefined}
                canPrevious={isHost && previousTracks.length > 0}
                hasQueue={queue.length > 0}
                syncError={playbackError}
                onFocusSearch={() => searchInputRef.current?.focus()}
                hostUsername={room.host.username}
              />

              <div className="grid grid-cols-1 gap-space-sm pt-space-xs md:grid-cols-3">
                <Featurette
                  icon="equalizer"
                  tone="text-primary"
                  label="Drift-corrected sync"
                  sub={playback?.isPlaying ? "Re-sync every 5s" : "Shared playhead"}
                />
                <Featurette
                  icon="group"
                  tone="text-secondary"
                  label="Host-led queue"
                  sub={`${queue.length} track${queue.length === 1 ? "" : "s"} queued`}
                />
                <Featurette
                  icon="lock_open"
                  tone="text-primary-container"
                  label="Private sanctuary"
                  sub={`Invite key · ${room.roomCode}`}
                />
              </div>
            </div>

            <div className="flex flex-col gap-space-md lg:col-span-4">
              <ChatPanel
                messages={messages}
                currentUserId={myUserId}
                connectionStatus={connectionStatus}
                memberCount={presence.length}
                hostUserId={effectiveHostId}
                sending={chatSending}
                error={chatError}
                onSend={sendChatMessage}
              />

              <QueuePanel
                queue={queue}
                isHost={isHost}
                hasPlayback={!!playback}
                onRemove={removeFromQueue}
                onReorder={reorderQueue}
                onClear={clearQueue}
                onSkip={advanceQueue}
                onFindSongs={() => searchInputRef.current?.focus()}
              />
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}

function Featurette({
  icon,
  tone,
  label,
  sub,
}: {
  icon: string;
  tone: string;
  label: string;
  sub: string;
}) {
  return (
    <div className="flex items-center gap-3 rounded-xl border border-obsidian-border bg-surface-container-lowest p-space-sm shadow-md">
      <div className="grid size-10 shrink-0 place-items-center rounded-lg bg-surface-container shadow-inner">
        <span className={`material-symbols-outlined text-[20px] ${tone}`}>
          {icon}
        </span>
      </div>
      <div className="flex min-w-0 flex-col">
        <span className="truncate font-label-md text-label-md font-medium text-on-surface">
          {label}
        </span>
        <span className="truncate font-label-sm text-label-sm text-on-surface-variant">
          {sub}
        </span>
      </div>
    </div>
  );
}

export function ConnectionCapsule({
  status,
}: {
  status: ConnectionStatus;
}) {
  const styles: Record<
    ConnectionStatus,
    { color: string; glow: boolean; label: string }
  > = {
    connecting: { color: "bg-on-surface-variant", glow: false, label: "Connecting…" },
    connected: { color: "bg-tertiary", glow: true, label: "Live · in sync" },
    reconnecting: { color: "bg-on-surface-variant", glow: false, label: "Reconnecting…" },
    disconnected: { color: "bg-error", glow: false, label: "Disconnected" },
  };
  const { color, label, glow } = styles[status];
  return (
    <div className="flex items-center gap-2 rounded-full border border-obsidian-border bg-surface-container-low px-space-sm py-1.5 shadow-sm">
      {glow ? (
        <span className="relative flex size-2.5">
          <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-tertiary opacity-75" />
          <span className="relative inline-flex h-2.5 w-2.5 rounded-full bg-tertiary" />
        </span>
      ) : (
        <span className={`relative size-2.5 rounded-full ${color} animate-pulse`} />
      )}
      <span
        className={`font-label-md text-label-md font-semibold tracking-wide ${
          status === "connected" ? "text-tertiary" : "text-on-surface-variant"
        }`}
      >
        {label}
      </span>
    </div>
  );
}

function HostPendingBanner({
  pending,
}: {
  pending: {
    hostUsername: string;
    successorUsername?: string;
    deadline?: string;
  } | null;
}) {
  const [secondsLeft, setSecondsLeft] = useState(() =>
    remainingSeconds(pending?.deadline),
  );

  useEffect(() => {
    if (!pending) return;
    const timer = setInterval(
      () => setSecondsLeft(remainingSeconds(pending.deadline)),
      1000,
    );
    return () => clearInterval(timer);
  }, [pending]);

  if (!pending) return null;

  return (
    <div className="mt-space-md flex items-center gap-3 rounded-xl border border-on-surface-variant/30 bg-surface-container px-4 py-3">
      <span className="h-1.5 w-1.5 shrink-0 animate-pulse rounded-full bg-on-surface-variant" />
      <p className="font-body-sm text-body-sm text-on-surface">
        <span className="font-semibold">{pending.hostUsername}</span> lost
        connection.
        {pending.successorUsername
          ? ` Handing the room to ${pending.successorUsername}`
          : " Waiting for them to return"}
        {secondsLeft > 0 ? ` in ${secondsLeft}s…` : "…"}
      </p>
    </div>
  );
}

function remainingSeconds(deadline?: string): number {
  if (!deadline) return 0;
  const ms = new Date(deadline).getTime() - Date.now();
  return ms > 0 ? Math.ceil(ms / 1000) : 0;
}
