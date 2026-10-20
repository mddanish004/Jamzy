import { useCallback, useEffect, useRef, useState } from "react";
import { io as ioClient, type Socket } from "socket.io-client";
import { useAuth } from "@clerk/react";
import type {
  ChatMessage,
  PlaybackAction,
  PlaybackSnapshot,
  PresenceSnapshot,
  PresenceMember,
  YouTubeSearchResult,
  QueueItem,
  QueueSnapshot,
  HostPendingPayload,
  HostUpdatePayload,
} from "../services/types";
import { clientConfig } from "../config/env";

const SOCKET_URL = clientConfig.socketUrl;

export type ConnectionStatus =
  | "connecting"
  | "connected"
  | "reconnecting"
  | "disconnected";

interface RoomJoinAck {
  ok: boolean;
  presence?: PresenceSnapshot;
  playback?: PlaybackSnapshot;
  messages?: ChatMessage[];
  queue?: QueueSnapshot;
  message?: string;
}

interface PlaybackAck {
  ok: boolean;
  playback?: PlaybackSnapshot;
  message?: string;
}

interface HostTransferAck {
  ok: boolean;
  message?: string;
}

function mergeMessages(
  prev: ChatMessage[],
  incoming: ChatMessage[],
): ChatMessage[] {
  const byId = new Map(prev.map((m) => [m.id, m]));
  for (const m of incoming) {
    byId.set(m.id, m);
    if (!m.id.startsWith("optimistic:")) {
      for (const [key, existing] of byId) {
        if (
          key.startsWith("optimistic:") &&
          existing.userId === m.userId &&
          existing.content === m.content
        ) {
          byId.delete(key);
          break;
        }
      }
    }
  }
  return Array.from(byId.values()).sort((a, b) =>
    a.createdAt.localeCompare(b.createdAt),
  );
}

export interface RoomSocket {
  presence: PresenceMember[];
  connectionStatus: ConnectionStatus;
  playback: PlaybackSnapshot | null;
  playbackError: string | null;
  setTrack: (track: YouTubeSearchResult) => void;
  sendControl: (action: PlaybackAction, position: number) => void;
  clearTrack: () => void;
  messages: ChatMessage[];
  chatError: string | null;
  chatSending: boolean;
  sendChatMessage: (content: string) => void;
  queue: QueueItem[];
  addToQueue: (track: YouTubeSearchResult) => void;
  prependToQueue: (track: YouTubeSearchResult) => void;
  removeFromQueue: (index: number) => void;
  reorderQueue: (fromIndex: number, toIndex: number) => void;
  clearQueue: () => void;
  advanceQueue: () => void;
  hostUserId: string | null;
  hostNotice: HostUpdatePayload | null;
  dismissHostNotice: () => void;
  hostPending: HostPendingPayload | null;
  hostError: string | null;
  transferHost: (userId: string) => void;
}

export function useRoomSocket(
  roomCode: string | null,
  initialMembers: { user: { id: string; username: string } }[],
  currentUser?: { id: string; username: string } | null,
): RoomSocket {
  const { getToken } = useAuth();
  const [presence, setPresence] = useState<PresenceMember[]>(
    initialMembers.map((m) => ({
      userId: m.user.id,
      username: m.user.username,
      online: false,
    })),
  );
  const [connectionStatus, setConnectionStatus] =
    useState<ConnectionStatus>("connecting");
  const [playback, setPlayback] = useState<PlaybackSnapshot | null>(null);
  const [playbackError, setPlaybackError] = useState<string | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [chatError, setChatError] = useState<string | null>(null);
  const [chatSending, setChatSending] = useState(false);
  const [queue, setQueue] = useState<QueueItem[]>([]);
  const [hostUserId, setHostUserId] = useState<string | null>(null);
  const [hostNotice, setHostNotice] = useState<HostUpdatePayload | null>(null);
  const [hostPending, setHostPending] = useState<HostPendingPayload | null>(
    null,
  );
  const [hostError, setHostError] = useState<string | null>(null);
  const socketRef = useRef<Socket | null>(null);

  const getTokenRef = useRef(getToken);
  useEffect(() => {
    getTokenRef.current = getToken;
  });

  const currentUserRef = useRef(currentUser);
  useEffect(() => {
    currentUserRef.current = currentUser;
  });

  useEffect(() => {
    if (!roomCode) return;

    let cancelled = false;

    const applyJoinAck = (res: RoomJoinAck) => {
      if (cancelled) return;
      if (!res.ok) {
        setPlaybackError(res.message ?? "Could not join the room");
        return;
      }
      if (res.presence) {
        setPresence(res.presence.members);
        setHostUserId(res.presence.hostUserId);
      }
      if (res.playback) setPlayback(res.playback);
      if (res.queue) setQueue(res.queue);
      if (res.messages)
        setMessages((prev) => mergeMessages(prev, res.messages!));
    };

    (async () => {
      const token = await getTokenRef.current();
      if (cancelled || !token) return;

      const socket = ioClient(SOCKET_URL, {
        auth: { token },
        transports: ["websocket"],
      });
      socketRef.current = socket;

      socket.on("connect", () => {
        setConnectionStatus("connected");
        socket.emit("room:join", { roomCode }, applyJoinAck);
      });

      socket.on("disconnect", () => {
        setConnectionStatus("disconnected");
      });

      socket.on("connect_error", (err: Error) => {
        setPlaybackError(err.message);
      });

      socket.io.on("reconnect_attempt", () => {
        setConnectionStatus("reconnecting");
      });

      socket.io.on("reconnect", () => {
        setConnectionStatus("connected");
        socket.emit("room:join", { roomCode }, applyJoinAck);
      });

      socket.on("presence:update", (snapshot: PresenceSnapshot) => {
        setPresence(snapshot.members);
        setHostUserId(snapshot.hostUserId);
      });

      socket.on("host:update", (payload: HostUpdatePayload) => {
        setHostUserId(payload.hostUserId);
        setHostNotice(payload);
        setHostPending(null);
        setHostError(null);
      });

      socket.on("host:pending", (payload: HostPendingPayload) => {
        setHostPending(payload.pending ? payload : null);
      });

      socket.on("playback:update", (snapshot: PlaybackSnapshot) => {
        setPlayback(snapshot);
        setPlaybackError(null);
      });

      socket.on("queue:update", (snapshot: QueueSnapshot) => {
        setQueue(snapshot);
      });

      socket.on("chat:new", (incoming: ChatMessage) => {
        setMessages((prev) => mergeMessages(prev, [incoming]));
      });
    })();

    return () => {
      cancelled = true;
      const socket = socketRef.current;
      if (socket) {
        socket.disconnect();
        socketRef.current = null;
      }
    };
  }, [roomCode]);

  const emitPlayback = useCallback(
    (event: string, payload: Record<string, unknown>) => {
      const socket = socketRef.current;
      if (!socket || !socket.connected) {
        setPlaybackError("Not connected to the room yet");
        return;
      }
      socket.emit(event, payload, (res: PlaybackAck) => {
        if (res?.ok) {
          setPlaybackError(null);
          if (res.playback) setPlayback(res.playback);
        } else {
          setPlaybackError(res?.message ?? "Playback update failed");
        }
      });
    },
    [],
  );

  const setTrack = useCallback(
    (track: YouTubeSearchResult) => {
      if (!roomCode) return;
      emitPlayback("playback:set", { roomCode, track });
    },
    [emitPlayback, roomCode],
  );

  const sendControl = useCallback(
    (action: PlaybackAction, position: number) => {
      if (!roomCode) return;
      emitPlayback("playback:control", { roomCode, action, position });
    },
    [emitPlayback, roomCode],
  );

  const clearTrack = useCallback(() => {
    if (!roomCode) return;
    emitPlayback("playback:clear", { roomCode });
  }, [emitPlayback, roomCode]);

  const sendChatMessage = useCallback(
    (content: string) => {
      const socket = socketRef.current;
      if (!roomCode) return;
      if (!socket || !socket.connected) {
        setChatError("Not connected to the room yet");
        return;
      }

      const optimisticId = `optimistic:${Date.now()}`;
      const me = currentUserRef.current;
      const optimisticMsg: ChatMessage = {
        id: optimisticId,
        roomId: roomCode,
        userId: me?.id ?? "",
        username: me?.username ?? "",
        content,
        createdAt: new Date().toISOString(),
      };
      setMessages((prev) => mergeMessages(prev, [optimisticMsg]));

      setChatSending(true);
      socket.emit("chat:send", { roomCode, content }, (res: { ok: boolean; message?: string }) => {
        setChatSending(false);
        if (!res?.ok) {
          setMessages((prev) => prev.filter((m) => m.id !== optimisticId));
          setChatError(res?.message ?? "Could not send message");
        } else {
          setChatError(null);
        }
      });
    },
    [roomCode],
  );

  const emitQueue = useCallback(
    (event: string, payload: Record<string, unknown>) => {
      const socket = socketRef.current;
      if (!socket || !socket.connected) return;
      socket.emit(event, payload);
    },
    [],
  );

  const addToQueue = useCallback(
    (track: YouTubeSearchResult) => {
      if (!roomCode) return;
      emitQueue("queue:add", { roomCode, track });
    },
    [emitQueue, roomCode],
  );

  const prependToQueue = useCallback(
    (track: YouTubeSearchResult) => {
      if (!roomCode) return;
      emitQueue("queue:prepend", { roomCode, track });
    },
    [emitQueue, roomCode],
  );

  const removeFromQueue = useCallback(
    (index: number) => {
      if (!roomCode) return;
      emitQueue("queue:remove", { roomCode, index });
    },
    [emitQueue, roomCode],
  );

  const reorderQueue = useCallback(
    (fromIndex: number, toIndex: number) => {
      if (!roomCode) return;
      emitQueue("queue:reorder", { roomCode, fromIndex, toIndex });
    },
    [emitQueue, roomCode],
  );

  const clearQueue = useCallback(() => {
    if (!roomCode) return;
    emitQueue("queue:clear", { roomCode });
  }, [emitQueue, roomCode]);

  const advanceQueue = useCallback(() => {
    if (!roomCode) return;
    emitQueue("queue:advance", { roomCode });
  }, [emitQueue, roomCode]);

  const transferHost = useCallback(
    (userId: string) => {
      const socket = socketRef.current;
      if (!roomCode) return;
      if (!socket || !socket.connected) {
        setHostError("Not connected to the room yet");
        return;
      }
      socket.emit(
        "host:transfer",
        { roomCode, userId },
        (res: HostTransferAck) => {
          setHostError(res?.ok ? null : (res?.message ?? "Could not transfer host"));
        },
      );
    },
    [roomCode],
  );

  const dismissHostNotice = useCallback(() => setHostNotice(null), []);

  return {
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
  };
}
