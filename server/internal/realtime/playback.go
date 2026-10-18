package realtime

import (
	"context"
	"fmt"
	"strconv"

	"github.com/mddanish004/jamzy/server/internal/config"
	"github.com/mddanish004/jamzy/server/internal/playback"
)

func (h *Hub) handlePlaybackSet(conn *connection, payload any, reply replyFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), handshakeTimeout)
	defer cancel()

	roomCode, ok := h.guardMember(conn, payload, reply, "socket:playback", config.Socket.Playback, "playback")
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

	snapshot, err := h.playback.SetTrack(ctx, roomCode, conn.user.ID, track)
	if err != nil {
		h.replyError(reply, err, "Playback update failed")
		return
	}

	h.log.Info(fmt.Sprintf("%s set %s in %s", conn.user.Username, track.VideoID, roomCode))
	h.broadcastPlayback(roomCode, *snapshot)

	reply(playbackAck{Ok: true, Playback: *snapshot})
}

func (h *Hub) handlePlaybackControl(conn *connection, payload any, reply replyFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), handshakeTimeout)
	defer cancel()

	roomCode, ok := h.guardMember(conn, payload, reply, "socket:playback", config.Socket.Playback, "playback")
	if !ok {
		return
	}
	if !h.guardHost(ctx, conn, roomCode, reply, controlPlayback, "Playback update failed") {
		return
	}

	action, ok := playback.ParseAction(field(payload, "action"))
	if !ok {
		reply(errorAck{Ok: false, Message: "Unknown playback action"})
		return
	}

	snapshot, err := h.playback.ApplyControl(ctx, roomCode, conn.user.ID, action, numeric(field(payload, "position")))
	if err != nil {
		h.replyError(reply, err, "Playback update failed")
		return
	}

	h.broadcastPlayback(roomCode, *snapshot)

	reply(playbackAck{Ok: true, Playback: *snapshot})
}

func (h *Hub) handlePlaybackClear(conn *connection, payload any, reply replyFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), handshakeTimeout)
	defer cancel()

	roomCode, ok := h.guardMember(conn, payload, reply, "socket:playback", config.Socket.Playback, "playback")
	if !ok {
		return
	}
	if !h.guardHost(ctx, conn, roomCode, reply, controlPlayback, "Playback update failed") {
		return
	}

	snapshot, err := h.playback.Clear(ctx, roomCode, conn.user.ID)
	if err != nil {
		h.replyError(reply, err, "Playback update failed")
		return
	}

	empty := h.queue.clear(roomCode)

	h.broadcastPlayback(roomCode, *snapshot)
	h.broadcastQueue(roomCode, empty)

	reply(playbackAck{Ok: true, Playback: *snapshot})
}

func numeric(raw any) float64 {
	switch value := raw.(type) {
	case float64:
		return value
	case string:
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0
		}
		return parsed
	case bool:
		if value {
			return 1
		}
		return 0
	default:
		return 0
	}
}
