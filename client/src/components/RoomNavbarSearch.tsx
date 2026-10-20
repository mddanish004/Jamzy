import { useEffect, useRef, useState } from "react";
import { useAuth } from "@clerk/react";
import {
  resolveYouTubeLink,
  resolveYouTubePlaylist,
  searchMusic,
} from "../services/api";
import { formatIsoDuration } from "../utils/duration";
import type { YouTubeSearchResult } from "../services/types";

interface RoomNavbarSearchProps {
  onSelect: (result: YouTubeSearchResult) => void;
  disabled?: boolean;
  canQueue?: boolean;
  onQueue?: (result: YouTubeSearchResult) => void;
  disabledMessage?: string;
  inputRef?: React.Ref<HTMLInputElement>;
}

interface VideoPreview {
  type: "video";
  track: YouTubeSearchResult;
}

interface PlaylistPreview {
  type: "playlist";
  title: string;
  tracks: YouTubeSearchResult[];
}

type LinkPreview = VideoPreview | PlaylistPreview;

const YOUTUBE_LINK_REGEX =
  /(^|\/\/|\s)((www|m|music)\.)?(youtube(-nocookie)?\.com|youtu\.be)\//i;

function isYouTubeLink(value: string): boolean {
  return YOUTUBE_LINK_REGEX.test(value);
}

function isYouTubePlaylistLink(value: string): boolean {
  return isYouTubeLink(value) && /[?&]list=/.test(value);
}

export function RoomNavbarSearch({
  onSelect,
  disabled = false,
  canQueue = false,
  onQueue,
  disabledMessage,
  inputRef,
}: RoomNavbarSearchProps) {
  const { getToken } = useAuth();
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<YouTubeSearchResult[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hasSearched, setHasSearched] = useState(false);
  const [isOpen, setIsOpen] = useState(false);

  const [linkPreview, setLinkPreview] = useState<LinkPreview | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [previewError, setPreviewError] = useState<string | null>(null);

  const containerRef = useRef<HTMLDivElement>(null);
  const localInputRef = useRef<HTMLInputElement>(null);

  function setInputEl(el: HTMLInputElement | null) {
    localInputRef.current = el;
    if (typeof inputRef === "function") {
      inputRef(el);
    } else if (inputRef) {
      (inputRef as React.RefObject<HTMLInputElement | null>).current = el;
    }
  }

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    }

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setIsOpen(false);
      }
    }

    document.addEventListener("mousedown", handleClickOutside);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, []);

  useEffect(() => {
    const trimmed = query.trim();
    if (!isYouTubeLink(trimmed)) {
      const timer = setTimeout(() => {
        setLinkPreview(null);
        setPreviewLoading(false);
        setPreviewError(null);
      }, 0);
      return () => clearTimeout(timer);
    }

    let cancelled = false;

    const timer = setTimeout(async () => {
      setPreviewLoading(true);
      setPreviewError(null);
      setIsOpen(true);
      try {
        if (isYouTubePlaylistLink(trimmed)) {
          const res = await resolveYouTubePlaylist(getToken, trimmed);
          if (cancelled) return;
          if (res.results.length === 0) {
            setPreviewError("No playable videos found in this playlist");
            setLinkPreview(null);
          } else {
            setLinkPreview({
              type: "playlist",
              title: res.playlistTitle || "YouTube Playlist",
              tracks: res.results,
            });
          }
        } else {
          const res = await resolveYouTubeLink(getToken, trimmed);
          if (cancelled) return;
          setLinkPreview({
            type: "video",
            track: res.result,
          });
        }
      } catch (err) {
        if (cancelled) return;
        const msg =
          err instanceof Error
            ? err.message
            : "Failed to resolve link preview";
        setPreviewError(msg);
        setLinkPreview(null);
      } finally {
        if (!cancelled) setPreviewLoading(false);
      }
    }, 400);

    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [query, getToken]);

  async function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    const trimmed = query.trim();
    if (!trimmed || busy || disabled) return;

    if (linkPreview && linkPreview.type === "video") {
      onSelect(linkPreview.track);
      setQuery("");
      setLinkPreview(null);
      setIsOpen(false);
      return;
    }

    if (linkPreview && linkPreview.type === "playlist") {
      handlePlayPlaylist(linkPreview);
      return;
    }

    setBusy(true);
    setError(null);
    setHasSearched(true);
    setIsOpen(true);

    try {
      const data = await searchMusic(getToken, trimmed);
      setResults(data.results);
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Search failed. Try again.";
      setError(message);
    } finally {
      setBusy(false);
    }
  }

  function handlePlayVideo(track: YouTubeSearchResult) {
    onSelect(track);
    setQuery("");
    setResults([]);
    setLinkPreview(null);
    setIsOpen(false);
  }

  function handleQueueVideo(track: YouTubeSearchResult) {
    if (!onQueue) return;
    onQueue(track);
    setQuery("");
    setResults([]);
    setLinkPreview(null);
    setIsOpen(false);
  }

  function handlePlayPlaylist(preview: PlaylistPreview) {
    if (preview.tracks.length === 0) return;
    const [first, ...rest] = preview.tracks;
    if (first) onSelect(first);
    if (onQueue && rest.length > 0) {
      rest.forEach((t) => onQueue(t));
    }
    setQuery("");
    setLinkPreview(null);
    setIsOpen(false);
  }

  function handleQueueAllPlaylist(preview: PlaylistPreview) {
    if (!onQueue) return;
    preview.tracks.forEach((t) => onQueue(t));
    setQuery("");
    setLinkPreview(null);
    setIsOpen(false);
  }

  function handleClearQuery() {
    setQuery("");
    setResults([]);
    setLinkPreview(null);
    setIsOpen(false);
    localInputRef.current?.focus();
  }

  const isPlaylist = isYouTubePlaylistLink(query.trim());
  const hasDropdownContent =
    isOpen &&
    (previewLoading ||
      previewError ||
      linkPreview !== null ||
      busy ||
      error !== null ||
      results.length > 0 ||
      (hasSearched && results.length === 0));

  return (
    <div ref={containerRef} className="relative w-full">
      <form onSubmit={handleSearch} className="relative flex w-full items-center">
        <div className="pointer-events-none absolute left-3.5 flex items-center">
          <span className="flex h-5 w-5 items-center justify-center rounded-md bg-secondary-container shadow-sm">
            <span className="material-symbols-outlined text-[14px] text-primary">
              smart_display
            </span>
          </span>
        </div>

        <input
          ref={setInputEl}
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onFocus={() => {
            if (query.trim().length > 0 || results.length > 0 || linkPreview) {
              setIsOpen(true);
            }
          }}
          placeholder="Search songs or paste YouTube link…"
          autoComplete="off"
          disabled={disabled}
          className="w-full rounded-xl border border-[#2c2b36] bg-surface-container-low py-2.5 pl-12 pr-24 font-body-md text-body-md text-on-surface transition-all placeholder:text-on-surface-variant/70 focus:border-primary/50 focus:bg-surface-container focus:outline-none disabled:opacity-50"
        />

        {query.length > 0 && (
          <button
            type="button"
            onClick={handleClearQuery}
            className="absolute right-16 grid size-6 place-items-center rounded-full text-on-surface-variant transition-colors hover:bg-surface-container-high hover:text-on-surface"
            aria-label="Clear search"
          >
            <span className="material-symbols-outlined text-[14px]">close</span>
          </button>
        )}

        <div className="pointer-events-none absolute right-3 flex items-center gap-1.5">
          <kbd className="pointer-events-none rounded border border-[#2c2b36] bg-surface-container-high px-2 py-0.5 font-label-sm text-label-sm uppercase tracking-wider text-on-surface-variant">
            ⌘K
          </kbd>
          {busy ? (
            <span className="pointer-events-none size-4 animate-spin rounded-full border-2 border-primary border-t-transparent" />
          ) : (
            <button
              type="submit"
              disabled={disabled || query.trim().length === 0}
              aria-label="Search"
              title="Search"
              className="pointer-events-auto disabled:pointer-events-none disabled:opacity-40"
            >
              <span className="material-symbols-outlined text-[18px] text-primary">
                search
              </span>
            </button>
          )}
        </div>
      </form>

      {hasDropdownContent && (
        <div className="jamzy-scrollbar absolute inset-x-0 top-full z-50 mt-2 max-h-[75vh] overflow-y-auto rounded-2xl border border-obsidian-border bg-surface-container p-2.5 shadow-2xl">
          <div className="mb-1.5 flex items-center justify-between px-1.5">
            <span className="font-label-sm text-label-sm uppercase tracking-wider text-on-surface-variant">
              {linkPreview
                ? isPlaylist
                  ? "Playlist preview"
                  : "Track preview"
                : results.length > 0
                  ? `Results (${results.length})`
                  : "Search"}
            </span>
            <button
              type="button"
              onClick={() => setIsOpen(false)}
              className="rounded-full px-1.5 py-0.5 font-label-sm text-label-sm text-on-surface-variant transition-colors hover:bg-surface-container-high hover:text-on-surface"
            >
              <span className="material-symbols-outlined text-[13px]">close</span>
            </button>
          </div>

          {previewLoading && (
            <div className="flex items-center gap-2.5 rounded-xl border border-obsidian-border bg-surface-container-low p-3 font-label-md text-label-md text-on-surface-variant">
              <span className="size-3.5 shrink-0 animate-spin rounded-full border-2 border-primary border-t-transparent" />
              <span>Resolving {isPlaylist ? "playlist" : "video"} preview…</span>
            </div>
          )}

          {previewError && (
            <div className="flex items-center justify-between rounded-xl border border-error/30 bg-error/10 px-3 py-2 font-label-md text-label-md text-error">
              <span>{previewError}</span>
              <button
                type="button"
                onClick={() => setPreviewError(null)}
                className="text-error transition-opacity hover:opacity-80"
                aria-label="Dismiss preview error"
              >
                <span className="material-symbols-outlined text-[14px]">close</span>
              </button>
            </div>
          )}

          {linkPreview && linkPreview.type === "video" && (
            <div className="relative overflow-hidden rounded-xl border border-obsidian-border bg-surface-container-low p-3 shadow-xl">
              <div className="flex items-start gap-3">
                <div className="relative shrink-0 overflow-hidden rounded-lg bg-black/40 shadow-md">
                  <img
                    src={linkPreview.track.thumbnailUrl}
                    alt=""
                    className="h-16 w-24 object-cover"
                  />
                  {linkPreview.track.duration && (
                    <span className="absolute bottom-1 right-1 rounded bg-black/85 px-1 py-0.5 text-[9px] font-semibold text-white">
                      {formatIsoDuration(linkPreview.track.duration)}
                    </span>
                  )}
                </div>

                <div className="min-w-0 flex-1">
                  <h4 className="line-clamp-2 font-body-md text-body-md font-semibold leading-snug text-on-surface">
                    {linkPreview.track.title}
                  </h4>
                  <p className="mt-1 truncate font-label-md text-label-md text-on-surface-variant">
                    {linkPreview.track.channelTitle || "YouTube"}
                  </p>

                  <div className="mt-2.5 flex items-center gap-2">
                    <button
                      type="button"
                      onClick={() => handlePlayVideo(linkPreview.track)}
                      disabled={disabled}
                      className="flex items-center gap-1.5 rounded-lg bg-gradient-to-br from-primary to-primary-container px-3 py-1.5 font-label-md text-label-md font-semibold text-white shadow-[0_0_16px_rgba(251,113,133,0.3)] transition-all hover:scale-[1.02] active:scale-95 disabled:opacity-50"
                    >
                      <span className="material-symbols-outlined msr-fill text-[14px]">
                        play_arrow
                      </span>
                      Play now
                    </button>

                    {canQueue && onQueue && (
                      <button
                        type="button"
                        onClick={() => handleQueueVideo(linkPreview.track)}
                        disabled={disabled}
                        className="flex items-center gap-1 rounded-lg border border-primary/40 bg-primary/10 px-3 py-1.5 font-label-md text-label-md font-semibold text-primary transition-all hover:bg-primary/20 active:scale-95 disabled:opacity-50"
                      >
                        <span className="material-symbols-outlined text-[14px]">
                          playlist_add
                        </span>
                        Queue
                      </button>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )}

          {linkPreview && linkPreview.type === "playlist" && (
            <div className="overflow-hidden rounded-xl border border-obsidian-border bg-surface-container-low p-3 shadow-xl">
              <div className="mb-2.5 flex items-center justify-between">
                <span className="font-label-md text-label-md font-medium text-on-surface-variant">
                  {linkPreview.tracks.length}{" "}
                  {linkPreview.tracks.length === 1 ? "track" : "tracks"}
                </span>
              </div>

              <div className="mb-3">
                <h3 className="truncate font-label-lg text-label-lg font-semibold text-on-surface">
                  {linkPreview.title}
                </h3>
                <p className="mt-0.5 font-label-md text-label-md text-on-surface-variant">
                  1st video plays immediately, remaining{" "}
                  {linkPreview.tracks.length - 1} queued for the room.
                </p>
              </div>

              <div className="mb-3 flex flex-wrap items-center gap-2">
                <button
                  type="button"
                  onClick={() => handlePlayPlaylist(linkPreview)}
                  disabled={disabled}
                  className="flex items-center gap-1.5 rounded-lg bg-gradient-to-br from-primary to-primary-container px-3 py-1.5 font-label-md text-label-md font-semibold text-white shadow-[0_0_16px_rgba(251,113,133,0.3)] transition-all hover:scale-[1.02] active:scale-95 disabled:opacity-50"
                >
                  <span className="material-symbols-outlined msr-fill text-[14px]">
                    play_arrow
                  </span>
                  Play playlist
                </button>

                {canQueue && onQueue && (
                  <button
                    type="button"
                    onClick={() => handleQueueAllPlaylist(linkPreview)}
                    disabled={disabled}
                    className="flex items-center gap-1 rounded-lg border border-primary/40 bg-primary/10 px-3 py-1.5 font-label-md text-label-md font-semibold text-primary transition-all hover:bg-primary/20 active:scale-95 disabled:opacity-50"
                  >
                    <span className="material-symbols-outlined text-[14px]">
                      playlist_add
                    </span>
                    Queue all ({linkPreview.tracks.length})
                  </button>
                )}
              </div>

              <div className="jamzy-scrollbar max-h-40 space-y-0.5 overflow-y-auto border-t border-obsidian-border/60 pr-1 pt-2">
                {linkPreview.tracks.map((t, idx) => {
                  const len = formatIsoDuration(t.duration);
                  return (
                    <div
                      key={t.videoId + idx}
                      className="flex items-center gap-2.5 rounded-lg px-2 py-1.5 transition-colors hover:bg-surface-container-high"
                    >
                      <span className="w-4 shrink-0 text-center font-label-sm text-label-sm text-on-surface-variant">
                        {idx + 1}
                      </span>
                      <img
                        src={t.thumbnailUrl}
                        alt=""
                        className="h-7 w-11 shrink-0 rounded bg-black/40 object-cover"
                      />
                      <div className="min-w-0 flex-1">
                        <p className="truncate font-label-md text-label-md font-medium text-on-surface">
                          {t.title}
                        </p>
                        {t.channelTitle && (
                          <p className="truncate font-label-sm text-label-sm text-on-surface-variant">
                            {t.channelTitle}
                          </p>
                        )}
                      </div>
                      {len && (
                        <span className="shrink-0 font-label-sm text-label-sm text-on-surface-variant">
                          {len}
                        </span>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {busy && (
            <div className="flex items-center justify-center gap-2.5 py-6 font-body-md text-body-md text-on-surface-variant">
              <span className="size-4 animate-spin rounded-full border-2 border-primary border-t-transparent" />
              <span>Searching YouTube…</span>
            </div>
          )}

          {error && (
            <p className="p-3 text-center font-body-md text-body-md text-error">
              {error}
            </p>
          )}

          {!busy && hasSearched && results.length === 0 && !error && !linkPreview && (
            <p className="py-6 text-center font-body-md text-body-md text-on-surface-variant">
              No results found. Try a different search.
            </p>
          )}

          {results.length > 0 && !busy && (
            <ul className="space-y-0.5">
              {results.map((result) => {
                const length = formatIsoDuration(result.duration);
                return (
                  <li
                    key={result.videoId}
                    className="group flex w-full items-center gap-3 rounded-xl p-2 transition-colors hover:bg-surface-container-low"
                  >
                    <img
                      src={result.thumbnailUrl}
                      alt=""
                      className="h-11 w-16 shrink-0 rounded-lg bg-surface-container-high object-cover"
                      loading="lazy"
                    />
                    <div className="min-w-0 flex-1">
                      <p className="truncate font-body-md text-body-md font-medium text-on-surface transition-colors group-hover:text-primary">
                        {result.title}
                      </p>
                      <p className="truncate font-label-md text-label-md text-on-surface-variant">
                        {result.channelTitle}
                      </p>
                    </div>
                    {length && (
                      <span className="hidden shrink-0 font-label-sm text-label-sm text-on-surface-variant sm:inline">
                        {length}
                      </span>
                    )}
                    <div className="flex shrink-0 items-center gap-1.5">
                      <button
                        onClick={() => handlePlayVideo(result)}
                        disabled={disabled}
                        className="rounded-lg border border-obsidian-border bg-surface-container-high px-2.5 py-1 font-label-sm text-label-sm font-semibold text-on-surface transition-colors hover:border-primary/40 hover:bg-surface-container-highest disabled:pointer-events-none disabled:opacity-50"
                      >
                        Play now
                      </button>
                      {canQueue && onQueue && (
                        <button
                          onClick={() => handleQueueVideo(result)}
                          disabled={disabled}
                          className="rounded-lg bg-primary/15 px-2.5 py-1 font-label-sm text-label-sm font-semibold text-primary transition-colors hover:bg-primary/25 disabled:pointer-events-none disabled:opacity-50"
                        >
                          + Queue
                        </button>
                      )}
                    </div>
                  </li>
                );
              })}
            </ul>
          )}

          {disabled && disabledMessage && (
            <p className="mt-1.5 flex items-center justify-center gap-1.5 text-center font-label-md text-label-md text-on-surface-variant">
              <span className="material-symbols-outlined text-[14px]">info</span>
              {disabledMessage}
            </p>
          )}
        </div>
      )}
    </div>
  );
}
