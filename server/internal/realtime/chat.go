package realtime

import (
	"context"
	"fmt"

	"github.com/mddanish004/jamzy/server/internal/config"
	"github.com/mddanish004/jamzy/server/internal/messages"
)

func (h *Hub) handleChatSend(conn *connection, payload any, reply replyFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), handshakeTimeout)
	defer cancel()

	roomCode, ok := h.guardMember(conn, payload, reply, "socket:chat", config.Socket.ChatSend, "chat")
	if !ok {
		return
	}

	content, err := messages.ValidateContent(field(payload, "content"))
	if err != nil {
		reply(errorAck{Ok: false, Message: err.Error()})
		return
	}

	roomID, err := h.rooms.IDIfMember(ctx, roomCode, conn.user.ID)
	if err != nil {
		h.replyError(reply, err, "Could not send message")
		return
	}
	if roomID == nil {
		reply(errorAck{Ok: false, Message: errNotAMember.Error()})
		return
	}

	message, err := h.messages.Create(ctx, *roomID, conn.user.ID, content)
	if err != nil {
		h.replyError(reply, err, "Could not send message")
		return
	}

	h.emitToRoom(roomCode, "chat:new", *message)
	h.log.Info(fmt.Sprintf("%s sent a message in %s", conn.user.Username, roomCode))

	reply(plainAck{Ok: true})
}
