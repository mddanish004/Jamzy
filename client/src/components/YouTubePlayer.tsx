import { useCallback, useEffect, useRef, useState } from "react";
import { useYouTubePlayer } from "../hooks/useYouTubePlayer";
import { formatSeconds } from "../utils/duration";
import type { PlaybackAction, PlaybackSnapshot } from "../services/types";
import type { PlayerState } from "../hooks/useYouTubePlayer";
import { EqualizerBars } from "./ui/EqualizerBars";

interface YouTubePlayerProps {
  playback: PlaybackSnapshot | null;
  onControl: ((action: PlaybackAction, position: number) => void) | null;
  onClear: () => void;
  onEnded?: () => void;
  onSkip?: () => void;
  onPrevious?: () => void;
  canPrevious?: boolean;
  hasQueue?: boolean;
  syncError: string | null;
  onFocusSearch?: () => void;
  hostUsername?: string;
}

const stateLabels: Record<PlayerState, string> = {
  initialising: "Starting player…",
  idle: "Ready",
  buffering: "Buffering…",
  playing: "Playing",
  paused: "Paused",
  ended: "Ended",
  error: "Error",
};

function sessionLabel(state: PlayerState): string {
  return stateLabels[state].toUpperCase();
}

const DRIFT_TOLERANCE_SECONDS = 2;
const DRIFT_CHECK_INTERVAL_MS = 5000;
const ECHO_SUPPRESSION_MS = 2000;

const SPEED_OPTIONS = [0.5, 0.75, 1, 1.25, 1.5, 2];

function formatQualityBadge(rawQuality: string): string {
  switch (rawQuality) {
    case "highres":
    case "hd2160":
      return "4K";
    case "hd1440":
      return "1440p";
    case "hd1080":
      return "1080p";
    case "hd720":
      return "720p";
    case "large":
      return "480p";
    case "medium":
      return "360p";
    case "small":
      return "240p";
    case "tiny":
      return "144p";
    case "auto":
    case "default":
      return "Auto";
    default:
      return rawQuality ? rawQuality.toUpperCase() : "Auto";
  }
}

function positionAtSnapshot(playback: PlaybackSnapshot): number {
  if (!playback.isPlaying || !playback.updatedAt) return playback.position;
  const elapsedMs =
    Date.parse(playback.serverTime) - Date.parse(playback.updatedAt);
  if (!Number.isFinite(elapsedMs)) return playback.position;
  return playback.position + Math.max(0, elapsedMs / 1000);
}

export function YouTubePlayer({
  playback,
  onControl,
  onClear,
  onEnded,
  onSkip,
  onPrevious,
  canPrevious = false,
  hasQueue = false,
  syncError,
  onFocusSearch,
  hostUsername,
}: YouTubePlayerProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const stageCardRef = useRef<HTMLDivElement>(null);

  const suppressEmitUntil = useRef(0);
  const anchor = useRef<{ position: number; atMs: number } | null>(null);
  const appliedVideoId = useRef<string | null>(null);
  const [scrubbing, setScrubbing] = useState<number | null>(null);

  const [isLooping, setIsLooping] = useState(false);
  const isLoopingRef = useRef(isLooping);
  useEffect(() => {
    isLoopingRef.current = isLooping;
  }, [isLooping]);

  const [showShortcutsModal, setShowShortcutsModal] = useState(false);
  const shortcutsModalRef = useRef<HTMLDivElement>(null);

  const [volume, setVolumeState] = useState<number>(() => {
    const saved = localStorage.getItem("jamzy_volume");
    if (saved !== null) {
      const parsed = Number(saved);
      if (!Number.isNaN(parsed) && parsed >= 0 && parsed <= 100) return parsed;
    }
    return 80;
  });

  const [isMuted, setIsMuted] = useState<boolean>(() => {
    return localStorage.getItem("jamzy_muted") === "true";
  });

  const [playbackRate, setPlaybackRateState] = useState<number>(1);
  const [showSpeedMenu, setShowSpeedMenu] = useState(false);
  const speedMenuRef = useRef<HTMLDivElement>(null);

  const playerRef = useRef<{
    seekTo: (pos: number) => void;
    play: () => void;
  } | null>(null);

  const handleLocalStateChange = useCallback(
    (state: PlayerState, at: number) => {
      if (state === "ended") {
        if (isLoopingRef.current) {
          playerRef.current?.seekTo(0);
          playerRef.current?.play();
          if (onControl) {
            onControl("seek", 0);
            onControl("play", 0);
          }
          return;
        }
        onEnded?.();
        return;
      }

      if (state === "playing" && appliedVideoId.current === playback?.videoId) {
        anchor.current = { position: at, atMs: Date.now() };
      }
      if (Date.now() < suppressEmitUntil.current) {
        suppressEmitUntil.current = Date.now() + ECHO_SUPPRESSION_MS;
        return;
      }

      if (!onControl) return;
      if (!playback?.videoId) return;
      if (state === "playing") onControl("play", at);
      else if (state === "paused") onControl("pause", at);
    },
    [onControl, onEnded, playback?.videoId],
  );

  const player = useYouTubePlayer(containerRef, {
    onStateChange: handleLocalStateChange,
  });

  useEffect(() => {
    playerRef.current = player;
  }, [player]);

  const {
    ready,
    loadVideo,
    cueVideo,
    play,
    pause,
    stopVideo,
    seekTo,
    getCurrentTime,
    setVolume,
    mute,
    unMute,
    setPlaybackRate,
    playbackQuality,
  } = player;
  const videoId = playback?.videoId ?? null;

  useEffect(() => {
    if (!playback || !playback.videoId) {
      if (appliedVideoId.current !== null) {
        appliedVideoId.current = null;
        anchor.current = null;
        suppressEmitUntil.current = Date.now() + ECHO_SUPPRESSION_MS;
        stopVideo();
      }
      return;
    }

    if (!ready) return;

    const target = positionAtSnapshot(playback);
    suppressEmitUntil.current = Date.now() + ECHO_SUPPRESSION_MS;
    anchor.current = { position: target, atMs: Date.now() };

    if (appliedVideoId.current !== playback.videoId) {
      appliedVideoId.current = playback.videoId;
      if (playback.isPlaying) {
        loadVideo(playback.videoId, target);
      } else {
        cueVideo(playback.videoId, target);
      }
      return;
    }

    if (Math.abs(getCurrentTime() - target) > DRIFT_TOLERANCE_SECONDS) {
      seekTo(target);
    }

    if (playback.isPlaying) play();
    else pause();
  }, [
    ready,
    playback,
    loadVideo,
    cueVideo,
    play,
    pause,
    stopVideo,
    seekTo,
    getCurrentTime,
  ]);

  useEffect(() => {
    if (!ready || !playback?.videoId) return;

    const interval = setInterval(() => {
      if (appliedVideoId.current !== playback.videoId) return;

      if (!playback.isPlaying) {
        if (player.state === "playing") {
          suppressEmitUntil.current = Date.now() + ECHO_SUPPRESSION_MS;
          pause();
        }
        return;
      }

      if (player.state === "paused" || player.state === "idle") {
        suppressEmitUntil.current = Date.now() + ECHO_SUPPRESSION_MS;
        play();
        return;
      }

      if (player.state !== "playing") return;

      const current = anchor.current;
      if (!current) return;
      const expected =
        current.position +
        ((Date.now() - current.atMs) / 1000) * playbackRate;
      if (Math.abs(getCurrentTime() - expected) > DRIFT_TOLERANCE_SECONDS) {
        suppressEmitUntil.current = Date.now() + ECHO_SUPPRESSION_MS;
        seekTo(expected);
      }
    }, DRIFT_CHECK_INTERVAL_MS);

    return () => clearInterval(interval);
  }, [
    ready,
    playback,
    player.state,
    getCurrentTime,
    seekTo,
    play,
    pause,
    playbackRate,
  ]);

  useEffect(() => {
    if (!ready) return;
    if (isMuted) {
      mute();
    } else {
      unMute();
      setVolume(volume);
    }
  }, [ready, volume, isMuted, mute, unMute, setVolume]);

  const handleVolumeChange = useCallback(
    (newVol: number) => {
      const clamped = Math.max(0, Math.min(100, newVol));
      setVolumeState(clamped);
      localStorage.setItem("jamzy_volume", String(clamped));
      if (clamped === 0) {
        setIsMuted(true);
        localStorage.setItem("jamzy_muted", "true");
        mute();
      } else {
        if (isMuted) {
          setIsMuted(false);
          localStorage.setItem("jamzy_muted", "false");
          unMute();
        }
        setVolume(clamped);
      }
    },
    [isMuted, mute, unMute, setVolume],
  );

  const toggleMute = useCallback(() => {
    setIsMuted((prev) => {
      const next = !prev;
      localStorage.setItem("jamzy_muted", String(next));
      if (next) {
        mute();
      } else {
        unMute();
        const effectiveVol = volume === 0 ? 50 : volume;
        if (volume === 0) {
          setVolumeState(50);
          localStorage.setItem("jamzy_volume", "50");
        }
        setVolume(effectiveVol);
      }
      return next;
    });
  }, [mute, unMute, setVolume, volume]);

  useEffect(() => {
    if (!ready) return;
    setPlaybackRate(playbackRate);
  }, [ready, playbackRate, setPlaybackRate]);

  const handleSpeedChange = useCallback(
    (rate: number) => {
      setPlaybackRateState(rate);
      setPlaybackRate(rate);
      anchor.current = { position: getCurrentTime(), atMs: Date.now() };
    },
    [getCurrentTime, setPlaybackRate],
  );

  useEffect(() => {
    if (!showSpeedMenu) return;
    function handleClickOutside(e: MouseEvent) {
      if (
        speedMenuRef.current &&
        !speedMenuRef.current.contains(e.target as Node)
      ) {
        setShowSpeedMenu(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [showSpeedMenu]);

  const duration = player.duration || playback?.duration || 0;
  const displayedTime = scrubbing ?? player.currentTime;
  const isPlaying = player.state === "playing";
  const controlsDisabled = !ready || !videoId || !onControl;

  const needsTapToPlay =
    Boolean(videoId) &&
    ready &&
    playback?.isPlaying === true &&
    (player.state === "paused" || player.state === "idle");

  const [actionRipple, setActionRipple] = useState<"play" | "pause" | null>(null);
  const rippleTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const triggerActionRipple = useCallback((action: "play" | "pause") => {
    if (rippleTimerRef.current) clearTimeout(rippleTimerRef.current);
    setActionRipple(action);
    rippleTimerRef.current = setTimeout(() => {
      setActionRipple(null);
    }, 650);
  }, []);

  function handleTapToPlay() {
    suppressEmitUntil.current = Date.now() + ECHO_SUPPRESSION_MS;
    play();
    triggerActionRipple("play");
  }

  function handleClear() {
    suppressEmitUntil.current = Date.now() + ECHO_SUPPRESSION_MS;
    appliedVideoId.current = null;
    anchor.current = null;
    stopVideo();
    onClear();
  }

  function handleToggle() {
    if (!onControl) return;
    const at = getCurrentTime();
    suppressEmitUntil.current = Date.now() + ECHO_SUPPRESSION_MS;
    if (isPlaying) {
      pause();
      onControl("pause", at);
      triggerActionRipple("pause");
    } else {
      play();
      onControl("play", at);
      triggerActionRipple("play");
    }
  }

  function handleSurfaceClick() {
    if (needsTapToPlay) {
      handleTapToPlay();
    } else if (!controlsDisabled) {
      handleToggle();
    }
  }

  function commitSeek(value: number) {
    const clamped = Math.max(0, Math.min(duration || 0, value));
    setScrubbing(null);
    seekTo(clamped);
    anchor.current = { position: clamped, atMs: Date.now() };
    suppressEmitUntil.current = Date.now() + ECHO_SUPPRESSION_MS;
    onControl?.("seek", clamped);
  }

  function handleJump(seconds: number) {
    if (controlsDisabled) return;
    commitSeek(displayedTime + seconds);
  }

  function handleResync() {
    if (!playback?.videoId || !ready) return;
    const target = positionAtSnapshot(playback);
    suppressEmitUntil.current = Date.now() + ECHO_SUPPRESSION_MS;
    seekTo(target);
    anchor.current = { position: target, atMs: Date.now() };
  }

  function handleFullscreen() {
    const el = stageCardRef.current;
    if (!el) return;
    if (document.fullscreenElement) {
      void document.exitFullscreen().catch(() => {});
    } else {
      void el.requestFullscreen?.().catch(() => {});
    }
  }

  useEffect(() => {
    if (!showShortcutsModal) return;
    function handleClickOutside(e: MouseEvent) {
      if (
        shortcutsModalRef.current &&
        !shortcutsModalRef.current.contains(e.target as Node)
      ) {
        setShowShortcutsModal(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [showShortcutsModal]);

  const handleToggleRef = useRef(handleToggle);
  useEffect(() => { handleToggleRef.current = handleToggle; });

  const handleTapToPlayRef = useRef(handleTapToPlay);
  useEffect(() => { handleTapToPlayRef.current = handleTapToPlay; });

  const handleJumpRef = useRef(handleJump);
  useEffect(() => { handleJumpRef.current = handleJump; });

  const handleVolumeChangeRef = useRef(handleVolumeChange);
  useEffect(() => { handleVolumeChangeRef.current = handleVolumeChange; });

  const toggleMuteRef = useRef(toggleMute);
  useEffect(() => { toggleMuteRef.current = toggleMute; });

  const handleSpeedChangeRef = useRef(handleSpeedChange);
  useEffect(() => { handleSpeedChangeRef.current = handleSpeedChange; });

  const onSkipRef = useRef(onSkip);
  useEffect(() => { onSkipRef.current = onSkip; });

  const onPreviousRef = useRef(onPrevious);
  useEffect(() => { onPreviousRef.current = onPrevious; });

  const canPreviousRef = useRef(canPrevious);
  useEffect(() => { canPreviousRef.current = canPrevious; });

  const volumeRef = useRef(volume);
  useEffect(() => { volumeRef.current = volume; });

  const playbackRateRef = useRef(playbackRate);
  useEffect(() => { playbackRateRef.current = playbackRate; });

  const controlsDisabledRef = useRef(controlsDisabled);
  useEffect(() => { controlsDisabledRef.current = controlsDisabled; });

  const needsTapToPlayRef = useRef(needsTapToPlay);
  useEffect(() => { needsTapToPlayRef.current = needsTapToPlay; });

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (
        e.target instanceof HTMLElement &&
        (e.target.tagName === "INPUT" ||
          e.target.tagName === "TEXTAREA" ||
          e.target.tagName === "SELECT" ||
          e.target.isContentEditable ||
          Boolean(e.target.closest("input, textarea, select, [contenteditable='true']")))
      ) {
        return;
      }

      if (e.ctrlKey || e.altKey || e.metaKey) return;

      const key = e.key.toLowerCase();

      if (e.code === "Space" || key === " ") {
        e.preventDefault();
        if (!controlsDisabledRef.current) {
          handleToggleRef.current();
        } else if (needsTapToPlayRef.current) {
          handleTapToPlayRef.current();
        }
        return;
      }

      if (e.code === "ArrowLeft") {
        e.preventDefault();
        if (!controlsDisabledRef.current) {
          handleJumpRef.current(-10);
        }
        return;
      }

      if (e.code === "ArrowRight") {
        e.preventDefault();
        if (!controlsDisabledRef.current) {
          handleJumpRef.current(10);
        }
        return;
      }

      if (key === "q") {
        e.preventDefault();
        const currentRate = playbackRateRef.current;
        const idx = SPEED_OPTIONS.indexOf(currentRate);
        const nextIdx = idx > 0 ? idx - 1 : 0;
        handleSpeedChangeRef.current(SPEED_OPTIONS[nextIdx]);
        return;
      }

      if (key === "e") {
        e.preventDefault();
        const currentRate = playbackRateRef.current;
        const idx = SPEED_OPTIONS.indexOf(currentRate);
        const nextIdx =
          idx >= 0 && idx < SPEED_OPTIONS.length - 1
            ? idx + 1
            : idx === -1
              ? 2
              : SPEED_OPTIONS.length - 1;
        handleSpeedChangeRef.current(SPEED_OPTIONS[nextIdx]);
        return;
      }

      if (key === "t") {
        e.preventDefault();
        if (!controlsDisabledRef.current && onSkipRef.current) {
          onSkipRef.current();
        }
        return;
      }

      if (key === "y") {
        e.preventDefault();
        if (
          !controlsDisabledRef.current &&
          canPreviousRef.current &&
          onPreviousRef.current
        ) {
          onPreviousRef.current();
        }
        return;
      }

      if (key === "r") {
        e.preventDefault();
        if (!controlsDisabledRef.current) {
          setIsLooping((prev) => !prev);
        }
        return;
      }

      if (key === "a") {
        e.preventDefault();
        handleVolumeChangeRef.current(volumeRef.current - 5);
        return;
      }

      if (key === "d") {
        e.preventDefault();
        handleVolumeChangeRef.current(volumeRef.current + 5);
        return;
      }

      if (key === "m") {
        e.preventDefault();
        toggleMuteRef.current();
        return;
      }
    }

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

  const message = syncError ?? player.error;
  const hdQuality =
    playbackQuality === "hd720" ||
    playbackQuality === "hd1080" ||
    playbackQuality === "hd1440" ||
    playbackQuality === "hd2160" ||
    playbackQuality === "highres";

  return (
    <section
      aria-label="Media stage and playback console"
      className="flex flex-col gap-space-sm"
    >
      <div className="flex items-center justify-between px-space-xs">
        <div className="flex min-w-0 items-center gap-space-sm">
          {videoId && isPlaying ? (
            <EqualizerBars className="h-2.5" />
          ) : (
            <span className="size-2 shrink-0 animate-pulse rounded-full bg-primary shadow-[0_0_8px_rgba(251,113,133,0.8)]" />
          )}
          <h2 className="shrink-0 font-headline-sm text-headline-sm font-semibold uppercase tracking-tight text-on-surface">
            Now playing
          </h2>
          {isLooping && videoId && (
            <span className="hidden shrink-0 items-center gap-1 rounded-full border border-primary/30 bg-primary/10 px-2 py-0.5 font-label-sm text-label-sm text-primary sm:inline-flex">
              <span className="material-symbols-outlined text-[12px]">repeat</span>
              Loop
            </span>
          )}
          {videoId && playback?.title && (
            <span className="hidden min-w-0 truncate font-label-sm text-label-sm text-on-surface-variant md:inline">
              · {playback.title}
            </span>
          )}
        </div>

        <div className="relative shrink-0" ref={shortcutsModalRef}>
          <button
            type="button"
            onClick={() => setShowShortcutsModal((prev) => !prev)}
            aria-label="Keyboard shortcuts"
            className="pointer-coarse:hidden flex items-center gap-1.5 rounded-lg border border-[#2c2b36] bg-surface-container-low px-2.5 py-1 font-label-sm text-label-sm text-on-surface-variant transition-all hover:bg-surface-container-high hover:text-on-surface"
          >
            <span className="material-symbols-outlined text-[15px]">keyboard</span>
            <span className="hidden sm:inline">Shortcuts</span>
          </button>

          {showShortcutsModal && (
            <div className="absolute right-0 top-full z-40 mt-2 w-80 rounded-xl border border-obsidian-border bg-surface-container p-4 shadow-2xl">
              <div className="mb-3 flex items-center justify-between border-b border-obsidian-border pb-2.5">
                <span className="font-label-lg text-label-lg uppercase tracking-wider text-on-surface">
                  Keyboard shortcuts
                </span>
                <button
                  type="button"
                  onClick={() => setShowShortcutsModal(false)}
                  aria-label="Close shortcuts"
                  className="rounded-full p-1 text-on-surface-variant transition-colors hover:bg-surface-container-high hover:text-on-surface"
                >
                  <span className="material-symbols-outlined text-[16px]">close</span>
                </button>
              </div>

              <div className="space-y-1">
                {[
                  ["Play / Pause", ["Space"], "min-w-12"],
                  ["10s Forward / Back", ["←", "→"], "h-6 w-6"],
                  ["Playback Speed", ["Q", "E"], "h-6 w-6"],
                  ["Skip Video", ["T"], "h-6 w-6"],
                  ["Previous Video", ["Y"], "h-6 w-6"],
                  ["Loop Video", ["R"], "h-6 w-6"],
                  ["Volume Down / Up", ["A", "D"], "h-6 w-6"],
                  ["Mute / Unmute", ["M"], "h-6 w-6"],
                ].map(([label, keys, size]) => (
                  <div
                    key={label as string}
                    className="flex items-center justify-between rounded-lg px-2 py-1.5 transition-colors hover:bg-surface-container-low"
                  >
                    <span className="font-label-md text-label-md font-medium text-on-surface">
                      {label}
                    </span>
                    <span className="flex items-center gap-1.5">
                      {(keys as string[]).map((k) => (
                        <kbd
                          key={k}
                          className={`${size} inline-flex items-center justify-center rounded-md border border-obsidian-border bg-surface-container-highest px-2.5 py-0.5 font-label-md text-label-md font-semibold text-primary shadow-sm`}
                        >
                          {k}
                        </kbd>
                      ))}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>

      <div
        ref={stageCardRef}
        className="group relative flex aspect-[16/9] w-full flex-col justify-between overflow-hidden rounded-2xl border border-[#2c2b36] bg-surface-container-lowest p-space-md shadow-2xl"
      >
        <div className="pointer-events-none absolute -top-24 left-1/2 h-[420px] w-[420px] -translate-x-1/2 rounded-full bg-primary/15 blur-[110px]" />
        <div className="pointer-events-none absolute -bottom-20 right-10 h-72 w-72 rounded-full bg-secondary-container/20 blur-[90px]" />
        <div className="pointer-events-none absolute left-1/4 top-1/3 h-48 w-48 rounded-full bg-primary-container/10 blur-[80px]" />

        <div
          ref={containerRef}
          className={`absolute inset-0 ${videoId ? "" : "invisible"}`}
        />

        {videoId && (
          <div
            onClick={handleSurfaceClick}
            className={`absolute inset-0 z-10 flex items-center justify-center ${
              !controlsDisabled || needsTapToPlay ? "cursor-pointer" : ""
            }`}
          >
            {actionRipple && (
              <div className="pointer-events-none flex h-20 w-20 animate-ping items-center justify-center rounded-full bg-black/70 text-white shadow-2xl backdrop-blur-md [animation-duration:500ms]">
                <span className="material-symbols-outlined msr-fill text-[40px] text-primary">
                  {actionRipple === "play" ? "play_arrow" : "pause"}
                </span>
              </div>
            )}

            {!actionRipple && (needsTapToPlay || (!isPlaying && !controlsDisabled)) && (
              <button
                type="button"
                onClick={(e) => {
                  e.stopPropagation();
                  if (needsTapToPlay) handleTapToPlay();
                  else handleToggle();
                }}
                className="flex items-center gap-2 rounded-full bg-primary px-5 py-2.5 font-label-lg text-label-lg text-on-primary shadow-[0_0_30px_rgba(251,113,133,0.45)] transition-all hover:scale-105 hover:bg-primary-container active:scale-95"
              >
                <span className="material-symbols-outlined msr-fill text-[18px]">
                  play_arrow
                </span>
                Tap to play
              </button>
            )}
          </div>
        )}

        {videoId && (
          <div className="pointer-events-none relative z-20 flex items-start justify-between gap-space-sm">
            <div className="pointer-events-auto flex min-w-0 items-center gap-2 rounded-md border border-[#2c2b36] bg-surface-container/70 px-2.5 py-1 backdrop-blur-md">
              <span className="material-symbols-outlined shrink-0 text-[16px] text-primary">
                smart_display
              </span>
              <span className="min-w-0 truncate font-label-sm text-label-sm text-on-surface">
                {playback?.title || "Untitled"}
              </span>
            </div>

            <div className="flex shrink-0 items-center gap-2">
              <span
                title={`Stream quality: ${formatQualityBadge(playbackQuality)} (auto-selected by YouTube)`}
                className="rounded-md border border-[#2c2b36] bg-surface-container/70 px-2 py-0.5 font-label-sm text-label-sm text-on-surface-variant backdrop-blur-md"
              >
                {formatQualityBadge(playbackQuality)}
                {hdQuality && (
                  <span className="ml-1 rounded bg-primary/20 px-1 py-px text-[9px] font-bold uppercase text-primary">
                    HD
                  </span>
                )}
              </span>
              <button
                type="button"
                onClick={(e) => {
                  e.stopPropagation();
                  handleClear();
                }}
                disabled={!onControl}
                aria-label="Clear the stage"
                title="Clear the stage"
                className="pointer-coarse:size-10 pointer-events-auto grid size-8 place-items-center rounded-md text-on-surface-variant transition-colors hover:text-primary disabled:pointer-events-none disabled:opacity-40"
              >
                <span className="material-symbols-outlined text-[20px]">close</span>
              </button>
              <button
                type="button"
                onClick={(e) => {
                  e.stopPropagation();
                  handleFullscreen();
                }}
                aria-label="Toggle fullscreen"
                title="Fullscreen"
                className="pointer-coarse:size-10 pointer-events-auto grid size-8 place-items-center rounded-md text-on-surface-variant transition-colors hover:text-primary"
              >
                <span className="material-symbols-outlined text-[20px]">fullscreen</span>
              </button>
            </div>
          </div>
        )}

        {!videoId && (
          <div className="relative z-10 mx-auto my-auto flex max-w-lg flex-col items-center justify-center px-space-md py-space-lg text-center">
            <div className="relative mb-space-md">
              <div className="grid size-20 place-items-center rounded-2xl border border-primary/25 bg-surface-container-low shadow-[0_0_45px_rgba(251,113,133,0.22)]">
                <span className="material-symbols-outlined text-[40px] text-primary">
                  graphic_eq
                </span>
              </div>
              <div className="absolute -right-1 -top-1 grid size-6 place-items-center rounded-full bg-primary-container text-white shadow-[0_0_12px_rgba(244,63,94,0.6)]">
                <span className="material-symbols-outlined text-[14px]">music_note</span>
              </div>
            </div>
            <h2 className="mb-2 font-headline-md text-headline-md font-semibold tracking-tight text-on-surface">
              The stage is waiting
            </h2>
            <p className="mx-auto mb-space-md max-w-md font-body-md text-body-md text-on-surface-variant">
              Search for a song or paste a YouTube link to start listening
              together with pristine spatial sync.
            </p>
            {onFocusSearch && (
              <button
                type="button"
                onClick={onFocusSearch}
                className="group/btn inline-flex items-center gap-2 rounded-full border border-[#2c2b36] bg-surface-container-low px-space-md py-2 font-label-md text-label-md text-on-surface shadow-lg transition-all hover:border-primary/30 hover:bg-surface-container hover:text-primary"
              >
                <span className="material-symbols-outlined text-[18px] text-primary transition-transform group-hover/btn:-translate-y-0.5">
                  north
                </span>
                <span>Focus search bar (⌘K)</span>
              </button>
            )}
          </div>
        )}

        <div className="pointer-events-none relative z-20 flex flex-wrap items-center justify-between gap-y-1 font-label-sm text-label-sm text-on-surface-variant">
          <div className="flex items-center gap-2">
            <span className="size-1.5 rounded-full bg-primary" />
            <span>Room host: {hostUsername ?? "—"}</span>
          </div>
          <span>Auto re-sync every {DRIFT_CHECK_INTERVAL_MS / 1000}s</span>
        </div>
      </div>

      <div className="flex w-full flex-col gap-space-sm rounded-2xl border border-[#2c2b36] bg-surface-container-low p-space-md shadow-xl">
        <div className="flex w-full flex-col gap-1.5">
          <input
            type="range"
            min={0}
            max={duration || 0}
            step={1}
            value={Math.min(displayedTime, duration || 0)}
            disabled={controlsDisabled || duration === 0}
            onChange={(e) => setScrubbing(Number(e.target.value))}
            onMouseUp={(e) => commitSeek(Number(e.currentTarget.value))}
            onTouchEnd={(e) => commitSeek(Number(e.currentTarget.value))}
            onKeyUp={(e) => {
              const key = e.key;
              if (
                key === "ArrowLeft" ||
                key === "ArrowRight" ||
                key === "Home" ||
                key === "End"
              ) {
                commitSeek(Number(e.currentTarget.value));
              }
            }}
            className="jamzy-range h-2 w-full"
            aria-label="Seek"
          />

          <div className="flex w-full items-center justify-between font-label-sm text-label-sm text-on-surface-variant">
            <span className="tabular-nums">{formatSeconds(displayedTime)}</span>
            {message ? (
              <span className="flex items-center gap-1 font-semibold text-error">
                <span className="material-symbols-outlined text-[13px]">
                  sync_problem
                </span>
                SYNC LOST
              </span>
            ) : videoId ? (
              <span className="font-semibold tracking-wide text-primary">
                SESSION {sessionLabel(player.state)}
              </span>
            ) : (
              <span className="tracking-wide text-on-surface-variant/70">
                STAGE EMPTY
              </span>
            )}
            <span className="tabular-nums">
              {duration ? formatSeconds(duration) : "--:--"}
            </span>
          </div>

          {message && (
            <p className="font-body-sm text-body-sm text-error">{message}</p>
          )}
        </div>

        <div className="flex flex-wrap items-center justify-between gap-y-2.5 pt-1 max-sm:flex-col max-sm:items-center">
          <div className="flex items-center gap-space-xs">
            <button
              type="button"
              onClick={() => setIsLooping((prev) => !prev)}
              disabled={controlsDisabled}
              aria-label={isLooping ? "Disable looping" : "Loop current video"}
              title={isLooping ? "Looper: ON (R)" : "Looper: OFF (R)"}
              className={`pointer-coarse:size-11 grid size-9 place-items-center rounded-full transition-all active:scale-95 disabled:pointer-events-none disabled:opacity-30 ${
                isLooping
                  ? "bg-primary/15 text-primary"
                  : "text-on-surface-variant hover:bg-surface-container-high hover:text-primary"
              }`}
            >
              <span className="material-symbols-outlined text-[20px]">repeat</span>
            </button>

            <div className="relative" ref={speedMenuRef}>
              <button
                type="button"
                onClick={() => setShowSpeedMenu((prev) => !prev)}
                aria-label={`Playback speed: ${playbackRate}x`}
                title="Playback speed (Q: slower, E: faster)"
                className="pointer-coarse:px-3 pointer-coarse:py-2 flex items-center gap-1 rounded-full px-2 py-1 font-label-md text-label-md text-on-surface-variant transition-all hover:bg-surface-container-high hover:text-primary active:scale-95"
              >
                <span className="material-symbols-outlined text-[20px]">speed</span>
                <span>{playbackRate === 1 ? "1x" : `${playbackRate}x`}</span>
              </button>

              {showSpeedMenu && (
                <div className="absolute bottom-full left-0 z-30 mb-2 w-36 rounded-xl border border-obsidian-border bg-surface-container p-1.5 shadow-2xl backdrop-blur-xl">
                  <div className="px-2 py-1 font-label-sm text-label-sm uppercase tracking-wider text-on-surface-variant">
                    Speed
                  </div>
                  {SPEED_OPTIONS.map((rate) => (
                    <button
                      key={rate}
                      type="button"
                      onClick={() => {
                        handleSpeedChange(rate);
                        setShowSpeedMenu(false);
                      }}
                      className={`flex w-full items-center justify-between rounded-lg px-2 py-1 font-label-md text-label-md transition-colors ${
                        playbackRate === rate
                          ? "bg-primary/15 font-semibold text-primary"
                          : "text-on-surface-variant hover:bg-surface-container-high hover:text-on-surface"
                      }`}
                    >
                      <span>{rate === 1 ? "1.0x Normal" : `${rate}x`}</span>
                      {playbackRate === rate && (
                        <span className="material-symbols-outlined text-[14px]">
                          check
                        </span>
                      )}
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>

          <div className="flex items-center gap-space-md">
            <button
              type="button"
              onClick={onPrevious}
              disabled={!onPrevious || !canPrevious || controlsDisabled}
              aria-label="Previous video"
              title={canPrevious ? "Previous video (Y)" : "No previous video"}
              className="grid size-10 place-items-center rounded-full text-on-surface transition-all hover:bg-surface-container-high hover:text-primary active:scale-95 disabled:pointer-events-none disabled:opacity-30 pointer-coarse:size-11"
            >
              <span className="material-symbols-outlined text-[24px]">
                skip_previous
              </span>
            </button>

            <button
              type="button"
              onClick={handleToggle}
              disabled={controlsDisabled}
              aria-label={isPlaying ? "Pause" : "Play"}
              title={isPlaying ? "Pause (Space)" : "Play (Space)"}
              className="grid size-14 place-items-center rounded-full bg-gradient-to-br from-primary via-primary-container to-[#be123c] text-white shadow-[0_0_30px_rgba(251,113,133,0.45)] transition-all hover:scale-105 hover:shadow-[0_0_36px_rgba(251,113,133,0.65)] active:scale-95 disabled:pointer-events-none disabled:opacity-40 disabled:shadow-none"
            >
              <span className="material-symbols-outlined msr-fill text-[32px]">
                {isPlaying ? "pause" : "play_arrow"}
              </span>
            </button>

            <button
              type="button"
              onClick={onSkip}
              disabled={!onSkip || controlsDisabled}
              aria-label="Skip to next video"
              title={hasQueue ? "Skip to next video (T)" : "Skip current video (T)"}
              className="grid size-10 place-items-center rounded-full text-on-surface transition-all hover:bg-surface-container-high hover:text-primary active:scale-95 disabled:pointer-events-none disabled:opacity-30 pointer-coarse:size-11"
            >
              <span className="material-symbols-outlined text-[24px]">skip_next</span>
            </button>
          </div>

          <div className="flex items-center gap-space-sm">
            <div className="group/vol flex items-center gap-2">
              <button
                type="button"
                onClick={toggleMute}
                aria-label={isMuted || volume === 0 ? "Unmute" : "Mute"}
                title={isMuted || volume === 0 ? "Unmute (M)" : `Mute (M) · Vol: ${volume}% (A / D)`}
                className="pointer-coarse:p-2.5 text-on-surface-variant transition-colors hover:text-primary"
              >
                <span className="material-symbols-outlined text-[20px]">
                  {isMuted || volume === 0
                    ? "volume_off"
                    : volume < 50
                      ? "volume_down"
                      : "volume_up"}
                </span>
              </button>

              <input
                type="range"
                min={0}
                max={100}
                step={1}
                value={isMuted ? 0 : volume}
                onChange={(e) => handleVolumeChange(Number(e.target.value))}
                className="jamzy-range h-1.5 w-20 sm:w-24"
                aria-label="Volume"
                title="Volume (A: down, D: up)"
              />
            </div>

            <div className="mx-1 h-4 w-[1px] bg-[#2c2b36]" />

            <button
              type="button"
              onClick={handleResync}
              disabled={!videoId || !ready}
              aria-label="Re-sync to the room"
              title="Re-anchor to the room's playhead (local catch-up)"
              className={`pointer-coarse:size-11 grid size-9 place-items-center rounded-full transition-colors hover:bg-surface-container-high disabled:pointer-events-none disabled:opacity-30 ${
                message ? "text-error" : "text-tertiary"
              }`}
            >
              <span className="material-symbols-outlined text-[20px]">sync</span>
            </button>
          </div>
        </div>
      </div>
    </section>
  );
}
