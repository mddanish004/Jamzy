import { useCallback, useEffect, useRef, useState } from "react";

declare global {
  interface Window {
    YT?: typeof YT;
    onYouTubeIframeAPIReady?: () => void;
  }
}

let apiLoadingPromise: Promise<void> | null = null;

function loadYouTubeAPI(): Promise<void> {
  if (apiLoadingPromise) return apiLoadingPromise;

  apiLoadingPromise = new Promise<void>((resolve, reject) => {
    if (window.YT && window.YT.Player) {
      resolve();
      return;
    }

    const existingCallback = window.onYouTubeIframeAPIReady;
    window.onYouTubeIframeAPIReady = () => {
      existingCallback?.();
      resolve();
    };

    const tag = document.createElement("script");
    tag.src = "https://www.youtube.com/iframe_api";
    tag.onerror = () => {
      apiLoadingPromise = null;
      reject(new Error("Could not load the YouTube player script"));
    };
    document.head.appendChild(tag);
  });

  return apiLoadingPromise;
}

export type PlayerState =
  | "initialising"
  | "idle"
  | "buffering"
  | "playing"
  | "paused"
  | "ended"
  | "error";

function describeError(code: number): string {
  switch (code) {
    case 2:
      return "That video ID is not valid.";
    case 5:
      return "This video cannot be played in the HTML5 player.";
    case 100:
      return "That video was removed or is private.";
    case 101:
    case 150:
      return "The owner of this video does not allow it to be played on other sites. Try a different upload.";
    default:
      return `The YouTube player reported error ${code}.`;
  }
}

function readCurrentTime(ref: React.RefObject<YT.Player | null>): number {
  try {
    return ref.current?.getCurrentTime() ?? 0;
  } catch {
    return 0;
  }
}

export interface YouTubePlayerHandle {
  state: PlayerState;
  ready: boolean;
  error: string | null;
  currentTime: number;
  duration: number;
  loadVideo: (videoId: string, startSeconds?: number) => void;
  cueVideo: (videoId: string, startSeconds?: number) => void;
  play: () => void;
  pause: () => void;
  stopVideo: () => void;
  seekTo: (seconds: number) => void;
  getCurrentTime: () => number;
  setVolume: (volume: number) => void;
  getVolume: () => number;
  mute: () => void;
  unMute: () => void;
  isMuted: () => boolean;
  setPlaybackRate: (rate: number) => void;
  getPlaybackRate: () => number;
  setPlaybackQuality: (quality: string) => void;
  getPlaybackQuality: () => string;
  playbackQuality: string;
  availableQualityLevels: string[];
}

export interface YouTubePlayerOptions {
  onStateChange?: (state: PlayerState, currentTime: number) => void;
}

export function useYouTubePlayer(
  containerRef: React.RefObject<HTMLDivElement | null>,
  options: YouTubePlayerOptions = {},
): YouTubePlayerHandle {
  const playerRef = useRef<YT.Player | null>(null);
  const [state, setState] = useState<PlayerState>("initialising");
  const [ready, setReady] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [playbackQuality, setPlaybackQualityState] = useState<string>("auto");
  const [availableQualityLevels, setAvailableQualityLevels] = useState<string[]>([]);
  const pendingVideo = useRef<{
    videoId: string;
    startSeconds: number;
    paused: boolean;
  } | null>(null);

  const onStateChangeRef = useRef(options.onStateChange);
  useEffect(() => {
    onStateChangeRef.current = options.onStateChange;
  });

  useEffect(() => {
    const container = containerRef.current;
    if (!container) {
      setError(
        "Player container is missing. This is a bug — the container must always be rendered.",
      );
      setState("error");
      return;
    }

    let disposed = false;

    const mountPoint = document.createElement("div");
    container.appendChild(mountPoint);

    loadYouTubeAPI()
      .then(() => {
        if (disposed || !window.YT?.Player) return;

        playerRef.current = new window.YT.Player(mountPoint, {
          height: "100%",
          width: "100%",
          playerVars: {
            autoplay: 0,
            controls: 0,
            disablekb: 1,
            iv_load_policy: 3,
            modestbranding: 1,
            rel: 0,
            playsinline: 1,
            origin: window.location.origin,
          },
          events: {
            onReady: () => {
              if (disposed) return;
              setReady(true);
              setError(null);
              setState("idle");

              try {
                const q = playerRef.current?.getPlaybackQuality();
                if (q) setPlaybackQualityState(q);
                const levels = playerRef.current?.getAvailableQualityLevels();
                if (levels && levels.length > 0) setAvailableQualityLevels(levels);
              } catch {
              }

              const pending = pendingVideo.current;
              if (pending) {
                pendingVideo.current = null;
                if (pending.paused) {
                  playerRef.current?.cueVideoById(
                    pending.videoId,
                    pending.startSeconds,
                  );
                } else {
                  playerRef.current?.loadVideoById(
                    pending.videoId,
                    pending.startSeconds,
                  );
                }
              }
            },
            onPlaybackQualityChange: (e: YT.OnPlaybackQualityChangeEvent) => {
              if (disposed) return;
              if (e?.data) setPlaybackQualityState(e.data);
              try {
                const levels = playerRef.current?.getAvailableQualityLevels();
                if (levels && levels.length > 0) setAvailableQualityLevels(levels);
              } catch {
              }
            },
            onStateChange: (e: YT.OnStateChangeEvent) => {
              if (disposed) return;

              let next: PlayerState | null = null;
              switch (e.data) {
                case YT.PlayerState.PLAYING:
                  next = "playing";
                  try {
                    const q = playerRef.current?.getPlaybackQuality();
                    if (q) setPlaybackQualityState(q);
                    const levels = playerRef.current?.getAvailableQualityLevels();
                    if (levels && levels.length > 0) setAvailableQualityLevels(levels);
                  } catch {
                  }
                  break;
                case YT.PlayerState.PAUSED:
                  next = "paused";
                  break;
                case YT.PlayerState.ENDED:
                  next = "ended";
                  break;
                case YT.PlayerState.BUFFERING:
                  next = "buffering";
                  break;
                case YT.PlayerState.CUED:
                  next = "idle";
                  break;
                default:
                  break;
              }

              if (!next) return;
              setState(next);
              if (next !== "buffering") setError(null);

              onStateChangeRef.current?.(next, readCurrentTime(playerRef));
            },
            onError: (e: YT.OnErrorEvent) => {
              if (disposed) return;
              setError(describeError(e.data));
              setState("error");
            },
          },
        });
      })
      .catch((err: unknown) => {
        if (disposed) return;
        setError(
          err instanceof Error
            ? err.message
            : "Could not load the YouTube player",
        );
        setState("error");
      });

    return () => {
      disposed = true;
      setReady(false);
      try {
        playerRef.current?.stopVideo();
        playerRef.current?.destroy();
      } catch {
      }
      playerRef.current = null;
      container.innerHTML = "";
    };
  }, [containerRef]);

  useEffect(() => {
    if (!ready) return;

    const interval = setInterval(() => {
      const player = playerRef.current;
      if (!player) return;
      try {
        setCurrentTime(player.getCurrentTime());
        setDuration(player.getDuration());
      } catch {
      }
    }, 500);

    return () => clearInterval(interval);
  }, [ready]);

  const loadVideo = useCallback((videoId: string, startSeconds = 0) => {
    setError(null);
    setCurrentTime(startSeconds);
    setDuration(0);

    const player = playerRef.current;
    if (player) {
      setState("buffering");
      player.loadVideoById(videoId, startSeconds);
      return;
    }

    pendingVideo.current = { videoId, startSeconds, paused: false };
    setState("buffering");
  }, []);

  const cueVideo = useCallback((videoId: string, startSeconds = 0) => {
    setError(null);
    setCurrentTime(startSeconds);
    setDuration(0);

    const player = playerRef.current;
    if (player) {
      setState("idle");
      player.cueVideoById(videoId, startSeconds);
      return;
    }

    pendingVideo.current = { videoId, startSeconds, paused: true };
    setState("idle");
  }, []);

  const play = useCallback(() => {
    try {
      playerRef.current?.playVideo();
    } catch {
    }
  }, []);

  const pause = useCallback(() => {
    try {
      playerRef.current?.pauseVideo();
    } catch {
    }
  }, []);

  const stopVideo = useCallback(() => {
    pendingVideo.current = null;
    setError(null);
    setCurrentTime(0);
    setDuration(0);
    setState("idle");
    try {
      playerRef.current?.stopVideo();
      playerRef.current?.pauseVideo();
    } catch {
    }
  }, []);

  const seekTo = useCallback((seconds: number) => {
    try {
      playerRef.current?.seekTo(seconds, true);
      setCurrentTime(seconds);
    } catch {
    }
  }, []);

  const getCurrentTime = useCallback(() => readCurrentTime(playerRef), []);

  const setVolume = useCallback((volume: number) => {
    try {
      playerRef.current?.setVolume(volume);
    } catch {
    }
  }, []);

  const getVolume = useCallback(() => {
    try {
      return playerRef.current?.getVolume() ?? 100;
    } catch {
      return 100;
    }
  }, []);

  const mute = useCallback(() => {
    try {
      playerRef.current?.mute();
    } catch {
    }
  }, []);

  const unMute = useCallback(() => {
    try {
      playerRef.current?.unMute();
    } catch {
    }
  }, []);

  const isMuted = useCallback(() => {
    try {
      return playerRef.current?.isMuted() ?? false;
    } catch {
      return false;
    }
  }, []);

  const setPlaybackRate = useCallback((rate: number) => {
    try {
      playerRef.current?.setPlaybackRate(rate);
    } catch {
    }
  }, []);

  const getPlaybackRate = useCallback(() => {
    try {
      return playerRef.current?.getPlaybackRate() ?? 1;
    } catch {
      return 1;
    }
  }, []);

  const setPlaybackQuality = useCallback((quality: string) => {
    try {
      playerRef.current?.setPlaybackQuality(quality as YT.SuggestedVideoQuality);
      setPlaybackQualityState(quality);
    } catch {
    }
  }, []);

  const getPlaybackQuality = useCallback(() => {
    try {
      return playerRef.current?.getPlaybackQuality() ?? playbackQuality;
    } catch {
      return playbackQuality;
    }
  }, [playbackQuality]);

  return {
    state,
    ready,
    error,
    currentTime,
    duration,
    loadVideo,
    cueVideo,
    play,
    pause,
    stopVideo,
    seekTo,
    getCurrentTime,
    setVolume,
    getVolume,
    mute,
    unMute,
    isMuted,
    setPlaybackRate,
    getPlaybackRate,
    setPlaybackQuality,
    getPlaybackQuality,
    playbackQuality,
    availableQualityLevels,
  };
}
