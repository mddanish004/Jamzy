package realtime

import (
	"context"
	"fmt"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mddanish004/jamzy/server/internal/apperr"
	"github.com/mddanish004/jamzy/server/internal/auth"
	"github.com/mddanish004/jamzy/server/internal/config"
	"github.com/mddanish004/jamzy/server/internal/logging"
	"github.com/mddanish004/jamzy/server/internal/messages"
	"github.com/mddanish004/jamzy/server/internal/playback"
	"github.com/mddanish004/jamzy/server/internal/ratelimit"
	"github.com/mddanish004/jamzy/server/internal/rooms"
	"github.com/mddanish004/jamzy/server/internal/users"
	"github.com/zishang520/socket.io/v2/socket"
)

const handshakeTimeout = 10 * time.Second

type connection struct {
	socket *socket.Socket
	user   users.PublicUser

	mu    sync.Mutex
	rooms map[string]socket.Room
}

func (c *connection) join(roomCode string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.rooms[roomCode] = roomName(roomCode)
}

func (c *connection) left(roomCode string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.rooms, roomCode)
}

func (c *connection) joined(roomCode string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.rooms[roomCode]
	return ok
}

func (c *connection) takeOtherRooms(roomCode string) []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	others := make([]string, 0, len(c.rooms))
	for code := range c.rooms {
		if code != roomCode {
			others = append(others, code)
		}
	}
	sort.Strings(others)

	return others
}

func (c *connection) takeAllRooms() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	all := make([]string, 0, len(c.rooms))
	for code := range c.rooms {
		all = append(all, code)
	}
	c.rooms = map[string]socket.Room{}
	sort.Strings(all)

	return all
}

type connectionRegistry struct {
	mu    sync.RWMutex
	conns map[socket.SocketId]*connection
}

func newConnectionRegistry() *connectionRegistry {
	return &connectionRegistry{conns: map[socket.SocketId]*connection{}}
}

func (r *connectionRegistry) add(conn *connection) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.conns[conn.socket.Id()] = conn
}

func (r *connectionRegistry) remove(id socket.SocketId) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.conns, id)
}

func (r *connectionRegistry) byUser(userID string) []*connection {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conns := make([]*connection, 0, len(r.conns))
	for _, conn := range r.conns {
		if conn.user.ID == userID {
			conns = append(conns, conn)
		}
	}

	return conns
}

type Hub struct {
	io       *socket.Server
	rooms    *rooms.Service
	playback *playback.Service
	messages *messages.Service
	users    *users.Service
	verifier auth.Verifier
	limiter  *ratelimit.Limiter
	log      logging.Logger

	presence *presenceStore
	queue    *queueStore
	pending  *pendingTransferStore
	conns    *connectionRegistry
}

func NewHub(
	roomsService *rooms.Service,
	playbackService *playback.Service,
	messagesService *messages.Service,
	usersService *users.Service,
	verifier auth.Verifier,
	limiter *ratelimit.Limiter,
	log logging.Logger,
) *Hub {
	return &Hub{
		rooms:    roomsService,
		playback: playbackService,
		messages: messagesService,
		users:    usersService,
		verifier: verifier,
		limiter:  limiter,
		log:      log,
		presence: newPresenceStore(),
		queue:    newQueueStore(),
		pending:  newPendingTransferStore(),
		conns:    newConnectionRegistry(),
	}
}

func (h *Hub) Server(options *socket.ServerOptions) *socket.Server {
	io := socket.NewServer(nil, options)
	h.io = io

	io.Use(h.authenticate)
	_ = io.On("connection", func(args ...any) {
		conn, ok := args[0].(*socket.Socket)
		if !ok {
			return
		}
		h.onConnection(conn)
	})

	return io
}

func (h *Hub) authenticate(client *socket.Socket, next func(*socket.ExtendedError)) {
	token := handshakeToken(client)
	if token == "" {
		next(socket.NewExtendedError("Missing auth token", nil))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), handshakeTimeout)
	defer cancel()

	clerkUserID, err := h.verifier.Verify(ctx, token)
	if err != nil {
		h.log.Warn("Socket authentication rejected", err)
		next(socket.NewExtendedError("Invalid token", nil))
		return
	}

	if clerkUserID == "" {
		next(socket.NewExtendedError("Invalid token: missing subject", nil))
		return
	}

	user, err := h.users.FindByClerkID(ctx, clerkUserID)
	if err != nil {
		h.log.Error("Socket authentication failed", err)
		next(socket.NewExtendedError("Authentication failed", nil))
		return
	}
	if user == nil {
		next(socket.NewExtendedError("User profile not found", nil))
		return
	}

	client.SetData(*user)
	next(nil)
}

func handshakeToken(client *socket.Socket) string {
	handshake := client.Handshake()
	if handshake == nil {
		return ""
	}

	auth, ok := handshake.Auth.(map[string]any)
	if !ok {
		return ""
	}

	token, _ := auth["token"].(string)
	return strings.TrimSpace(token)
}

func (h *Hub) onConnection(client *socket.Socket) {
	user, ok := client.Data().(users.PublicUser)
	if !ok {
		client.Disconnect(true)
		return
	}

	conn := &connection{
		socket: client,
		user:   user,
		rooms:  map[string]socket.Room{},
	}
	h.conns.add(conn)

	h.log.Info(fmt.Sprintf("Socket connected: %s (%s)", user.Username, client.Id()))

	h.on(client, "room:join", func(payload any, reply replyFunc) {
		h.handleRoomJoin(conn, payload, reply)
	})
	h.on(client, "playback:set", func(payload any, reply replyFunc) {
		h.handlePlaybackSet(conn, payload, reply)
	})
	h.on(client, "playback:control", func(payload any, reply replyFunc) {
		h.handlePlaybackControl(conn, payload, reply)
	})
	h.on(client, "playback:clear", func(payload any, reply replyFunc) {
		h.handlePlaybackClear(conn, payload, reply)
	})
	h.on(client, "queue:add", func(payload any, reply replyFunc) {
		h.handleQueueAdd(conn, payload, reply)
	})
	h.on(client, "queue:prepend", func(payload any, reply replyFunc) {
		h.handleQueuePrepend(conn, payload, reply)
	})
	h.on(client, "queue:remove", func(payload any, reply replyFunc) {
		h.handleQueueRemove(conn, payload, reply)
	})
	h.on(client, "queue:reorder", func(payload any, reply replyFunc) {
		h.handleQueueReorder(conn, payload, reply)
	})
	h.on(client, "queue:advance", func(payload any, reply replyFunc) {
		h.handleQueueAdvance(conn, payload, reply)
	})
	h.on(client, "queue:clear", func(payload any, reply replyFunc) {
		h.handleQueueClear(conn, payload, reply)
	})
	h.on(client, "chat:send", func(payload any, reply replyFunc) {
		h.handleChatSend(conn, payload, reply)
	})
	h.on(client, "host:transfer", func(payload any, reply replyFunc) {
		h.handleHostTransfer(conn, payload, reply)
	})

	_ = client.On("room:leave", func(args ...any) {
		payload, ack := splitArgs(args)
		reply := replyTo(ack)

		go func() {
			defer h.recoverEvent("room:leave")

			if err := h.handleRoomLeave(conn, payload); err != nil {
				h.log.Warn("room:leave rejected", err)
			}

			reply(nil)
		}()
	})

	_ = client.On("disconnect", func(args ...any) {
		reason := "unknown"
		if len(args) > 0 {
			reason = fmt.Sprint(args[0])
		}

		go func() {
			defer h.recoverEvent("disconnect")
			h.handleDisconnect(conn, reason)
		}()
	})
}

func (h *Hub) on(client *socket.Socket, event string, handler func(any, replyFunc)) {
	_ = client.On(event, func(args ...any) {
		payload, ack := splitArgs(args)
		reply := replyTo(ack)

		go func() {
			defer h.recoverEvent(event)
			handler(payload, reply)
		}()
	})
}

func (h *Hub) recoverEvent(event string) {
	recovered := recover()
	if recovered == nil {
		return
	}

	h.log.Error(
		fmt.Sprintf("Socket event %q panicked", event),
		fmt.Sprint(recovered),
		string(debug.Stack()),
	)
}

func (h *Hub) handleDisconnect(conn *connection, reason string) {
	h.conns.remove(conn.socket.Id())

	ctx, cancel := context.WithTimeout(context.Background(), handshakeTimeout)
	defer cancel()

	for _, roomCode := range conn.takeAllRooms() {
		if !h.presence.removeSocket(roomCode, conn.user.ID, string(conn.socket.Id())) {
			continue
		}

		h.broadcastPresence(ctx, roomCode)
		h.scheduleHostTransferIfHostLeft(ctx, roomCode, conn.user.ID, conn.user.Username)
	}

	h.log.Info(fmt.Sprintf("Socket disconnected: %s (%s) — %s", conn.user.Username, conn.socket.Id(), reason))
}

func (h *Hub) checkRate(bucket string, userID string, label string, rule config.RateLimitRule) error {
	result := h.limiter.Consume(bucket+":user:"+userID, rule)
	if result.Allowed {
		return nil
	}

	retryAfter := ratelimit.RetryAfterSeconds(result.RetryAfterMs)
	h.log.Warn("Socket rate limit exceeded", label)

	return fmt.Errorf("Slow down — try again in %ds.", retryAfter)
}

func (h *Hub) assertJoined(conn *connection, raw any) (string, error) {
	roomCode, err := rooms.ValidateRoomCode(raw)
	if err != nil {
		return "", err
	}

	if !conn.joined(roomCode) {
		return "", errNotAMember
	}

	return roomCode, nil
}

func (h *Hub) assertHost(ctx context.Context, roomCode string, userID string, action string) error {
	hostUserID, err := h.rooms.HostID(ctx, roomCode)
	if err != nil {
		return err
	}
	if hostUserID == nil {
		return apperr.New("Room not found", 404)
	}
	if *hostUserID != userID {
		return apperr.New(fmt.Sprintf("Only the host can %s", action), 403)
	}

	return nil
}

func (h *Hub) replyError(reply replyFunc, err error, fallback string) {
	if apperr.StatusCode(err) != 0 {
		reply(errorAck{Ok: false, Message: err.Error()})
		return
	}

	h.log.Error("Socket handler failed", err)
	reply(errorAck{Ok: false, Message: fallback})
}

func (h *Hub) EvictUserFromRoom(roomCode string, userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), handshakeTimeout)
	defer cancel()

	for _, conn := range h.conns.byUser(userID) {
		if !conn.joined(roomCode) {
			continue
		}

		conn.socket.Leave(roomName(roomCode))
		conn.left(roomCode)

		if h.presence.removeSocket(roomCode, userID, string(conn.socket.Id())) {
			h.broadcastPresence(ctx, roomCode)
		}
	}
}
