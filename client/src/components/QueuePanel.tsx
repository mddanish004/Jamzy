import { useState } from "react";
import { formatIsoDuration } from "../utils/duration";
import type { QueueItem } from "../services/types";

interface QueuePanelProps {
  queue: QueueItem[];
  isHost: boolean;
  hasPlayback: boolean;
  onRemove: (index: number) => void;
  onReorder: (fromIndex: number, toIndex: number) => void;
  onClear: () => void;
  onSkip: () => void;
  onFindSongs: () => void;
}

export function QueuePanel({
  queue,
  isHost,
  hasPlayback,
  onRemove,
  onReorder,
  onClear,
  onSkip,
  onFindSongs,
}: QueuePanelProps) {
  const [dragging, setDragging] = useState<number | null>(null);
  const [dragOver, setDragOver] = useState<number | null>(null);

  function handleDragStart(index: number) {
    setDragging(index);
  }

  function handleDragEnter(index: number) {
    if (dragging === null || dragging === index) return;
    setDragOver(index);
  }

  function handleDrop(index: number) {
    if (dragging === null || dragging === index) {
      setDragging(null);
      setDragOver(null);
      return;
    }
    onReorder(dragging, index);
    setDragging(null);
    setDragOver(null);
  }

  function handleDragEnd() {
    setDragging(null);
    setDragOver(null);
  }

  function moveUp(index: number) {
    if (index <= 0) return;
    onReorder(index, index - 1);
  }

  function moveDown(index: number) {
    if (index >= queue.length - 1) return;
    onReorder(index, index + 1);
  }

  return (
    <div className="flex min-h-[300px] flex-col overflow-hidden rounded-2xl border border-[#2c2b36] bg-surface-container-low shadow-xl">
      <div className="flex shrink-0 items-center justify-between border-b border-[#2c2b36] bg-surface-container/70 px-space-md py-3.5 backdrop-blur-sm">
        <div className="flex items-center gap-2">
          <span className="material-symbols-outlined text-[18px] text-primary">
            queue_music
          </span>
          <span className="font-headline-sm text-headline-sm font-semibold uppercase tracking-tight text-on-surface">
            Up next
          </span>
          <span className="rounded-full border border-[#2c2b36] bg-surface-container-high px-2 py-0.5 font-label-sm text-label-sm font-semibold text-on-surface-variant">
            {queue.length} {queue.length === 1 ? "track" : "tracks"}
          </span>
        </div>

        {isHost && queue.length > 0 && (
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={onSkip}
              className="pointer-coarse:py-2 flex items-center gap-1 font-label-sm text-label-sm text-on-surface-variant transition-colors hover:text-primary"
              aria-label="Skip to next track"
              title="Skip to next track"
            >
              <span className="material-symbols-outlined text-[16px]">
                skip_next
              </span>
              <span>Skip</span>
            </button>
            <button
              type="button"
              onClick={onClear}
              className="pointer-coarse:py-2 flex items-center gap-1 font-label-sm text-label-sm text-on-surface-variant transition-colors hover:text-error"
              aria-label="Clear queue"
              title="Clear the whole queue"
            >
              <span className="material-symbols-outlined text-[16px]">
                delete_sweep
              </span>
              <span>Clear</span>
            </button>
          </div>
        )}
      </div>

      {queue.length === 0 ? (
        <div className="flex flex-1 flex-col items-center justify-center p-space-md text-center">
          <div className="flex max-w-[260px] flex-col items-center gap-space-sm">
            <div className="grid size-12 place-items-center rounded-xl border border-[#2c2b36] bg-surface-container-high text-on-surface-variant">
              <span className="material-symbols-outlined text-[24px]">
                library_music
              </span>
            </div>
            <div className="flex flex-col gap-1">
              <p className="font-headline-sm text-headline-sm font-semibold text-on-surface">
                Queue is empty
              </p>
              <p className="font-body-sm text-body-sm leading-relaxed text-on-surface-variant">
                {isHost
                  ? "Add songs from the search bar above to build your synchronized queue."
                  : "Waiting for the host to add tracks."}
              </p>
            </div>
            {isHost && (
              <button
                type="button"
                onClick={onFindSongs}
                className="pointer-coarse:py-2 mt-1 flex items-center gap-1.5 rounded-lg border border-[#2c2b36] bg-surface-container px-3 py-1.5 font-label-sm text-label-sm font-semibold text-primary shadow-sm transition-all hover:border-primary/40 hover:bg-surface-container-high"
              >
                <span className="material-symbols-outlined text-[14px]">search</span>
                <span>Find songs</span>
              </button>
            )}
          </div>
        </div>
      ) : (
        <ul className="jamzy-scrollbar max-h-72 min-h-0 flex-1 space-y-0.5 overflow-y-auto p-space-sm pr-1">
          {queue.map((item, index) => {
            const length = formatIsoDuration(item.duration);
            const isFirst = index === 0;
            const isDraggingThis = dragging === index;
            const isDragTarget = dragOver === index;

            return (
              <li
                key={`${item.videoId}-${index}`}
                draggable={isHost}
                onDragStart={() => handleDragStart(index)}
                onDragEnter={() => handleDragEnter(index)}
                onDragOver={(e) => e.preventDefault()}
                onDrop={() => handleDrop(index)}
                onDragEnd={handleDragEnd}
                className={`group flex items-center gap-2.5 rounded-xl px-2 py-2 transition-all ${
                  isDraggingThis ? "opacity-40" : ""
                } ${
                  isDragTarget
                    ? "bg-primary/5 ring-1 ring-primary/60"
                    : "hover:bg-surface-container-high"
                } ${isFirst ? "border-l-2 border-primary" : ""}`}
              >
                <span
                  className={`w-5 shrink-0 text-center font-label-sm text-label-sm ${
                    isFirst ? "font-semibold text-primary" : "text-on-surface-variant"
                  } ${isHost ? "cursor-grab active:cursor-grabbing" : ""}`}
                  aria-hidden="true"
                >
                  {index + 1}
                </span>

                <img
                  src={item.thumbnailUrl}
                  alt=""
                  className="h-9 w-12 shrink-0 rounded-md bg-surface-container-high object-cover"
                  loading="lazy"
                />

                <div className="min-w-0 flex-1">
                  <p className="truncate font-label-md text-label-md font-medium text-on-surface">
                    {item.title}
                  </p>
                  {isFirst && (
                    <p className="font-label-sm text-label-sm font-semibold uppercase tracking-wider text-primary/80">
                      Up next{length ? ` · ${length}` : ""}
                    </p>
                  )}
                </div>

                {isHost && (
                  <div className="flex shrink-0 items-center gap-0.5 transition-opacity sm:opacity-0 sm:group-hover:opacity-100">
                    <button
                      type="button"
                      onClick={(e) => {
                        e.stopPropagation();
                        moveUp(index);
                      }}
                      disabled={index === 0}
                      className="pointer-coarse:size-9 grid size-7 place-items-center rounded-md text-on-surface-variant transition-all hover:bg-surface-container-highest hover:text-on-surface active:scale-95 disabled:pointer-events-none disabled:opacity-20"
                      aria-label="Move up in queue"
                      title="Move up"
                    >
                      <span className="material-symbols-outlined text-[16px]">
                        keyboard_arrow_up
                      </span>
                    </button>
                    <button
                      type="button"
                      onClick={(e) => {
                        e.stopPropagation();
                        moveDown(index);
                      }}
                      disabled={index === queue.length - 1}
                      className="pointer-coarse:size-9 grid size-7 place-items-center rounded-md text-on-surface-variant transition-all hover:bg-surface-container-highest hover:text-on-surface active:scale-95 disabled:pointer-events-none disabled:opacity-20"
                      aria-label="Move down in queue"
                      title="Move down"
                    >
                      <span className="material-symbols-outlined text-[16px]">
                        keyboard_arrow_down
                      </span>
                    </button>
                    <button
                      type="button"
                      onClick={(e) => {
                        e.stopPropagation();
                        onRemove(index);
                      }}
                      className="pointer-coarse:size-9 grid size-7 place-items-center rounded-md text-on-surface-variant transition-all hover:bg-error/15 hover:text-error active:scale-95"
                      aria-label={`Remove ${item.title} from queue`}
                      title="Remove from queue"
                    >
                      <span className="material-symbols-outlined text-[16px]">
                        close
                      </span>
                    </button>
                  </div>
                )}
              </li>
            );
          })}
        </ul>
      )}

      <div className="flex shrink-0 items-center justify-between border-t border-[#2c2b36] bg-surface-container-lowest/70 px-space-md py-2.5 font-label-sm text-label-sm text-on-surface-variant backdrop-blur-sm">
        <span className="flex items-center gap-1">
          <span className="material-symbols-outlined text-[13px] text-tertiary">
            security
          </span>
          <span>Host-led queue</span>
        </span>
        <span>{queue.length > 0 && hasPlayback ? "Auto-advance on" : "Auto-advance ready"}</span>
      </div>
    </div>
  );
}
