package realtime

import (
	"context"

	"github.com/mddanish004/jamzy/server/internal/playback"
	"github.com/mddanish004/jamzy/server/internal/rooms"
	"github.com/zishang520/engine.io/v2/types"
	"github.com/zishang520/socket.io/v2/socket"
)

const roomPrefix = "room:"

func roomName(roomCode string) socket.Room {
	return socket.Room(roomPrefix + roomCode)
}

func (h *Hub) emitToRoom(roomCode string, event string, payload any) {
	ns := h.io.Sockets()
	room := roomName(roomCode)
	ids := ns.Adapter().Sockets(types.NewSet(room))
	clients := ns.Sockets()
	for _, id := range ids.Keys() {
		client, ok := clients.Load(id)
		if !ok {
			continue
		}
		if err := client.Emit(event, payload); err != nil {
			h.log.Error("Room broadcast failed", roomCode, event, err)
		}
	}
}

func (h *Hub) emitPresence(roomCode string, data *rooms.PresenceData) {
	h.emitToRoom(roomCode, "presence:update", h.presence.snapshot(roomCode, data.Members, data.HostUserID))
}

func (h *Hub) broadcastPresence(ctx context.Context, roomCode string) {
	data, err := h.rooms.MembersForPresence(ctx, roomCode)
	if err != nil {
		h.log.Error("Presence broadcast failed", roomCode, err)
		return
	}
	if data == nil {
		return
	}

	h.emitPresence(roomCode, data)
}

func (h *Hub) broadcastPlayback(roomCode string, snapshot playback.Snapshot) {
	h.emitToRoom(roomCode, "playback:update", snapshot)
}

func (h *Hub) broadcastQueue(roomCode string, queue []QueueItem) {
	h.emitToRoom(roomCode, "queue:update", queue)
}
