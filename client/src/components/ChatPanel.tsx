import { useEffect, useRef, useState } from "react";
import type { ChatMessage } from "../services/types";
import type { ConnectionStatus } from "../hooks/useRoomSocket";
import { EmojiPicker } from "./ui/EmojiPicker";

interface ChatPanelProps {
  messages: ChatMessage[];
  currentUserId: string | undefined;
  connectionStatus: ConnectionStatus;
  memberCount: number;
  hostUserId?: string | null;
  sending: boolean;
  error: string | null;
  onSend: (content: string) => void;
}

const MESSAGE_MAX_LENGTH = 500;
const NEAR_BOTTOM_THRESHOLD = 80;

const EMOJI_ONLY = /^(?:\p{Extended_Pictographic}|\p{Emoji_Component}|️|‍)+$/u;

function isJumboEmoji(content: string): boolean {
  if (!EMOJI_ONLY.test(content)) return false;
  const graphemes = [...new Intl.Segmenter().segment(content)];
  return graphemes.length > 0 && graphemes.length <= 3;
}

function formatTime(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleTimeString(undefined, {
    hour: "numeric",
    minute: "2-digit",
  });
}

export function ChatPanel({
  messages,
  currentUserId,
  connectionStatus,
  memberCount,
  hostUserId,
  sending,
  error,
  onSend,
}: ChatPanelProps) {
  const [draft, setDraft] = useState("");
  const [hasNewBelow, setHasNewBelow] = useState(false);
  const [collapsed, setCollapsed] = useState(false);
  const listRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const wasNearBottom = useRef(true);
  const messageCount = messages.length;

  useEffect(() => {
    const list = listRef.current;
    if (!list) return;

    if (wasNearBottom.current) {
      list.scrollTop = list.scrollHeight;
      setHasNewBelow(false);
    } else {
      setHasNewBelow(true);
    }
  }, [messageCount]);

  function handleScroll() {
    const list = listRef.current;
    if (!list) return;
    const distanceFromBottom =
      list.scrollHeight - list.scrollTop - list.clientHeight;
    wasNearBottom.current = distanceFromBottom < NEAR_BOTTOM_THRESHOLD;
    if (wasNearBottom.current) setHasNewBelow(false);
  }

  function scrollToBottom() {
    const list = listRef.current;
    if (!list) return;
    list.scrollTop = list.scrollHeight;
    wasNearBottom.current = true;
    setHasNewBelow(false);
  }

  function handleEmojiSelect(emoji: string) {
    const input = inputRef.current;
    const start = input?.selectionStart ?? draft.length;
    const end = input?.selectionEnd ?? draft.length;
    const next = draft.slice(0, start) + emoji + draft.slice(end);

    if (next.length > MESSAGE_MAX_LENGTH) return;

    setDraft(next);

    requestAnimationFrame(() => {
      if (!input) return;
      const caret = start + emoji.length;
      input.focus();
      input.setSelectionRange(caret, caret);
    });
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const trimmed = draft.trim();
    if (trimmed.length === 0 || sending) return;
    onSend(trimmed);
    setDraft("");
  }

  const disconnected =
    connectionStatus === "disconnected" || connectionStatus === "reconnecting";

  const myUsername = [...messages]
    .reverse()
    .find((m) => m.userId === currentUserId)?.username;

  return (
    <div
      className={`flex flex-col overflow-hidden rounded-2xl border border-[#2c2b36] bg-surface-container-low shadow-xl ${
        collapsed ? "" : "h-[440px]"
      }`}
    >
      <div className="shrink-0 border-b border-[#2c2b36] bg-surface-container/70 px-space-md py-3.5 backdrop-blur-sm">
        <div className="flex items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <span className="material-symbols-outlined text-[18px] text-primary">
              chat_bubble
            </span>
            <span className="font-headline-sm text-headline-sm font-semibold uppercase tracking-tight text-on-surface">
              Chat
            </span>
            <span className="rounded-full border border-[#2c2b36] bg-surface-container-high px-2 py-0.5 font-label-sm text-label-sm font-semibold text-on-surface-variant">
              {memberCount}
            </span>
          </div>

          <div className="flex min-w-0 items-center gap-2">
            {currentUserId && (
              <span className="flex min-w-0 items-center gap-1 rounded-full border border-tertiary/20 bg-tertiary/10 px-2.5 py-0.5 font-label-sm text-label-sm font-semibold text-tertiary">
                <span
                  className={`size-1.5 shrink-0 rounded-full ${
                    disconnected
                      ? "bg-on-surface-variant"
                      : "bg-tertiary shadow-[0_0_6px_rgba(16,185,129,0.8)]"
                  }`}
                />
                <span className="min-w-0 truncate">{myUsername ?? "You"}</span>
                <span className="shrink-0">(You)</span>
              </span>
            )}

            <button
              type="button"
              onClick={() => setCollapsed((c) => !c)}
              className="text-on-surface-variant transition-colors hover:text-on-surface"
              aria-label={collapsed ? "Expand chat" : "Collapse chat"}
            >
              <span
                className={`material-symbols-outlined text-[18px] transition-transform duration-300 ${
                  collapsed ? "rotate-180" : ""
                }`}
              >
                expand_less
              </span>
            </button>
          </div>
        </div>
      </div>

      {!collapsed && (
        <>
          <div className="relative min-h-0 flex-1">
            <div
              ref={listRef}
              onScroll={handleScroll}
              className={`h-full overflow-y-auto px-space-md py-space-sm ${
                messages.length === 0
                  ? "flex flex-col items-center justify-center text-center"
                  : "flex flex-col items-start gap-space-sm"
              }`}
            >
              {messages.length === 0 ? (
                <div className="flex max-w-[240px] flex-col items-center gap-2">
                  <div className="grid size-12 place-items-center rounded-full border border-[#2c2b36] bg-surface-container text-2xl shadow-inner">
                    👋
                  </div>
                  <p className="mt-1 font-headline-sm text-headline-sm font-semibold text-on-surface">
                    No messages yet
                  </p>
                  <p className="font-body-sm text-body-sm leading-relaxed text-on-surface-variant">
                    Start the conversation or drop track recommendations here.
                  </p>
                </div>
              ) : (
                messages.map((msg) => {
                  const isMine = msg.userId === currentUserId;
                  const jumbo = isJumboEmoji(msg.content);
                  return (
                    <div
                      key={msg.id}
                      className={`flex w-full animate-fade-in flex-col ${
                        isMine ? "items-end" : "items-start"
                      }`}
                    >
                      <div className="mb-1 flex items-center gap-2">
                        <span className="font-label-sm text-label-sm text-on-surface-variant">
                          {formatTime(msg.createdAt)}
                        </span>
                        {isMine ? (
                          <span className="font-label-sm text-label-sm font-semibold text-primary">
                            You
                          </span>
                        ) : (
                          <span className="font-label-sm text-label-sm font-semibold text-secondary">
                            {msg.username}
                            {!jumbo && msg.userId === hostUserId && (
                              <span className="text-on-surface-variant"> · host</span>
                            )}
                          </span>
                        )}
                      </div>
                      <div
                        className={`max-w-[85%] text-left shadow-sm ${
                          jumbo
                            ? "px-1 py-0.5"
                            : isMine
                              ? "rounded-2xl rounded-tr-sm border border-[#2c2b36] bg-surface-container-high px-3.5 py-2"
                              : "rounded-2xl rounded-tl-sm border border-[#2c2b36] bg-surface-container px-3.5 py-2"
                        }`}
                      >
                        <p
                          className={`whitespace-pre-wrap break-words ${
                            jumbo
                              ? "text-3xl leading-tight"
                              : "font-body-sm text-body-sm text-on-surface"
                          }`}
                        >
                          {msg.content}
                        </p>
                      </div>
                    </div>
                  );
                })
              )}
            </div>

            {hasNewBelow && (
              <button
                type="button"
                onClick={scrollToBottom}
                className="absolute bottom-3 left-1/2 -translate-x-1/2 rounded-full bg-primary px-3 py-1.5 font-label-md text-label-md font-medium text-white shadow-lg transition-colors hover:bg-primary-container"
              >
                ↓ New messages
              </button>
            )}
          </div>

          <div className="shrink-0 space-y-1.5 border-t border-[#2c2b36] bg-surface-container-lowest/80 p-space-sm backdrop-blur-sm">
            {disconnected && (
              <p className="flex items-center gap-1.5 px-1 font-label-md text-label-md text-on-surface-variant">
                <span className="material-symbols-outlined text-[14px]">
                  reconnect
                </span>
                {connectionStatus === "reconnecting"
                  ? "Reconnecting…"
                  : "Disconnected — trying to reconnect…"}
              </p>
            )}
            {error && (
              <p className="flex items-center gap-1.5 px-1 font-label-md text-label-md text-error">
                <span className="material-symbols-outlined text-[14px]">error</span>
                {error}
              </p>
            )}

            <form onSubmit={handleSubmit} className="flex items-center gap-2">
              <EmojiPicker
                onSelect={handleEmojiSelect}
                disabled={connectionStatus !== "connected"}
              />

              <input
                ref={inputRef}
                type="text"
                value={draft}
                onChange={(e) => setDraft(e.target.value)}
                placeholder="Type a message…"
                maxLength={MESSAGE_MAX_LENGTH}
                autoComplete="off"
                disabled={connectionStatus !== "connected"}
                className="min-w-0 flex-1 rounded-xl border border-[#2c2b36] bg-surface-container px-3.5 py-2.5 font-body-sm text-body-sm text-on-surface transition-all placeholder:text-on-surface-variant/70 focus:border-primary/50 focus:bg-surface-container-high focus:outline-none disabled:opacity-50"
              />

              <button
                type="submit"
                disabled={
                  connectionStatus !== "connected" || draft.trim().length === 0
                }
                aria-label="Send message"
                title="Send message"
                className="grid size-9 shrink-0 place-items-center rounded-xl bg-gradient-to-br from-primary to-primary-container text-white shadow-[0_0_16px_rgba(251,113,133,0.4)] transition-all hover:scale-105 active:scale-95 disabled:pointer-events-none disabled:opacity-40 disabled:shadow-none"
              >
                <span className="material-symbols-outlined text-[18px]">send</span>
              </button>
            </form>
          </div>
        </>
      )}
    </div>
  );
}
