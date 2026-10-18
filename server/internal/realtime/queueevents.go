package realtime

import (
	"context"
	"fmt"
	"math"

	"github.com/mddanish004/jamzy/server/internal/config"
	"github.com/mddanish004/jamzy/server/internal/playback"
)

func (h *Hub) handleQueueAdd(conn *connection, payload any, reply replyFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), handshakeTimeout)
	defer cancel()

	roomCode, ok := h.guardMember(conn, payload, reply, "socket:queue", config.Socket.Queue, "queue")
	if !ok {
		return
	}
	if !h.guardHost(ctx, conn, roomCode, reply, controlPlayback, "Playback update failed") {
		return
	}

	track, ok := h.validTrack(payload, reply)
	if !ok {
		return
	}

	queue := h.queue.enqueue(roomCode, toQueueItem(track))

	h.log.Info(fmt.Sprintf("%s queued %s in %s", conn.user.Username, track.VideoID, roomCode))
	h.broadcastQueue(roomCode, queue)

	reply(queueAck{Ok: true, Queue: queue})
}

func (h *Hub) handleQueuePrepend(conn *connection, payload any, reply replyFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), handshakeTimeout)
	defer cancel()

	roomCode, ok := h.guardMember(conn, payload, reply, "socket:queue", config.Socket.Queue, "queue")
	if !ok {
		return
	}
	if !h.guardHost(ctx, conn, roomCode, reply, controlPlayback, "Playback update failed") {
		return
	}

	track, ok := h.validTrack(payload, reply)
	if !ok {
		return
	}

	queue := h.queue.prepend(roomCode, toQueueItem(track))

	h.log.Info(fmt.Sprintf("%s prepended %s in %s", conn.user.Username, track.VideoID, roomCode))
	h.broadcastQueue(roomCode, queue)

	reply(queueAck{Ok: true, Queue: queue})
}

func (h *Hub) handleQueueRemove(conn *connection, payload any, reply replyFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), handshakeTimeout)
	defer cancel()

	roomCode, ok := h.guardMember(conn, payload, reply, "socket:queue", config.Socket.Queue, "queue")
	if !ok {
		return
	}
	if !h.guardHost(ctx, conn, roomCode, reply, controlPlayback, "Playback update failed") {
		return
	}

	queue := h.queue.remove(roomCode, queueIndex(field(payload, "index")))
	h.broadcastQueue(roomCode, queue)

	reply(queueAck{Ok: true, Queue: queue})
}

func (h *Hub) handleQueueReorder(conn *connection, payload any, reply replyFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), handshakeTimeout)
	defer cancel()

	roomCode, ok := h.guardMember(conn, payload, reply, "socket:queue", config.Socket.Queue, "queue")
	if !ok {
		return
	}
	if !h.guardHost(ctx, conn, roomCode, reply, controlPlayback, "Playback update failed") {
		return
	}

	queue := h.queue.reorder(roomCode, queueIndex(field(payload, "fromIndex")), queueIndex(field(payload, "toIndex")))
	h.broadcastQueue(roomCode, queue)

	reply(queueAck{Ok: true, Queue: queue})
}

func (h *Hub) handleQueueAdvance(conn *connection, payload any, reply replyFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), handshakeTimeout)
	defer cancel()

	roomCode, ok := h.guardMember(conn, payload, reply, "socket:queue", config.Socket.Queue, "queue")
	if !ok {
		return
	}
	if !h.guardHost(ctx, conn, roomCode, reply, controlPlayback, "Playback update failed") {
		return
	}

	next, hasNext := h.queue.dequeue(roomCode)
	queue := h.queue.get(roomCode)

	if hasNext {
		snapshot, err := h.playback.SetTrack(ctx, roomCode, conn.user.ID, playback.Track(next))
		if err != nil {
			h.replyError(reply, err, "Playback update failed")
			return
		}

		h.log.Info(fmt.Sprintf("%s advanced queue → %s in %s", conn.user.Username, next.VideoID, roomCode))
		h.broadcastPlayback(roomCode, *snapshot)
	} else {
		snapshot, err := h.playback.Clear(ctx, roomCode, conn.user.ID)
		if err != nil {
			h.replyError(reply, err, "Playback update failed")
			return
		}

		h.log.Info(fmt.Sprintf("%s queue exhausted in %s", conn.user.Username, roomCode))
		h.broadcastPlayback(roomCode, *snapshot)
	}

	h.broadcastQueue(roomCode, queue)

	reply(queueAck{Ok: true, Queue: queue})
}

func (h *Hub) handleQueueClear(conn *connection, payload any, reply replyFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), handshakeTimeout)
	defer cancel()

	roomCode, ok := h.guardMember(conn, payload, reply, "socket:queue", config.Socket.Queue, "queue")
	if !ok {
		return
	}
	if !h.guardHost(ctx, conn, roomCode, reply, controlPlayback, "Playback update failed") {
		return
	}

	queue := h.queue.clear(roomCode)
	h.broadcastQueue(roomCode, queue)

	reply(queueAck{Ok: true, Queue: queue})
}

func (h *Hub) validTrack(payload any, reply replyFunc) (playback.Track, bool) {
	track, err := playback.ValidateTrack(field(payload, "track"))
	if err != nil {
		reply(errorAck{Ok: false, Message: err.Error()})
		return playback.Track{}, false
	}

	return track, true
}

func toQueueItem(track playback.Track) QueueItem {
	return QueueItem{
		VideoID:      track.VideoID,
		Title:        track.Title,
		ThumbnailURL: track.ThumbnailURL,
		Duration:     track.Duration,
	}
}

func queueIndex(raw any) int {
	value, ok := raw.(float64)
	if !ok {
		return -1
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return -1
	}
	if value < math.MinInt32 || value > math.MaxInt32 {
		return -1
	}

	return int(value)
}
