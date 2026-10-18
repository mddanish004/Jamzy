package realtime

import (
	"context"

	"github.com/mddanish004/jamzy/server/internal/config"
)

const controlPlayback = "control playback"

func (h *Hub) guardMember(
	conn *connection,
	payload any,
	reply replyFunc,
	bucket string,
	rule config.RateLimitRule,
	label string,
) (string, bool) {
	if err := h.checkRate(bucket, conn.user.ID, label, rule); err != nil {
		reply(errorAck{Ok: false, Message: err.Error()})
		return "", false
	}

	roomCode, err := h.assertJoined(conn, field(payload, "roomCode"))
	if err != nil {
		reply(errorAck{Ok: false, Message: err.Error()})
		return "", false
	}

	return roomCode, true
}

func (h *Hub) guardHost(ctx context.Context, conn *connection, roomCode string, reply replyFunc, action string, fallback string) bool {
	if err := h.assertHost(ctx, roomCode, conn.user.ID, action); err != nil {
		h.replyError(reply, err, fallback)
		return false
	}

	return true
}
