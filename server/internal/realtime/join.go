package realtime

import (
	"context"

	"github.com/mddanish004/jamzy/server/internal/config"
	"github.com/mddanish004/jamzy/server/internal/rooms"
)

func (h *Hub) handleRoomJoin(conn *connection, payload any, reply replyFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), handshakeTimeout)
	defer cancel()

	if err := h.checkRate("socket:join", conn.user.ID, "join", config.Socket.RoomJoin); err != nil {
		reply(errorAck{Ok: false, Message: err.Error()})
		return
	}

	roomCode, err := rooms.ValidateRoomCode(field(payload, "roomCode"))
	if err != nil {
		reply(errorAck{Ok: false, Message: err.Error()})
		return
	}

	roomData, err := h.rooms.MembersForPresence(ctx, roomCode)
	if err != nil {
		h.replyError(reply, err, "Join failed")
		return
	}
	if roomData == nil {
		reply(errorAck{Ok: false, Message: errRoomNotFound.Error()})
		return
	}

	if !isRoomMember(roomData.Members, conn.user.ID) {
		reply(errorAck{Ok: false, Message: errNotAMember.Error()})
		return
	}

	for _, previous := range conn.takeOtherRooms(roomCode) {
		conn.socket.Leave(roomName(previous))
		conn.left(previous)
		h.presence.removeSocket(previous, conn.user.ID, string(conn.socket.Id()))
		h.broadcastPresence(ctx, previous)
	}

	wasNewUser := h.presence.addSocket(roomCode, conn.user.ID, string(conn.socket.Id()))
	conn.socket.Join(roomName(roomCode))
	conn.join(roomCode)

	h.cancelPendingTransferOnReturn(ctx, roomCode, conn.user.ID, conn.user.Username, roomData.HostUserID)

	playbackSnapshot, err := h.playback.Get(ctx, roomCode)
	if err != nil {
		h.replyError(reply, err, "Join failed")
		return
	}

	history, err := h.messages.Recent(ctx, roomData.RoomID)
	if err != nil {
		h.replyError(reply, err, "Join failed")
		return
	}

	reply(joinAck{
		Ok:       true,
		Presence: h.presence.snapshot(roomCode, roomData.Members, roomData.HostUserID),
		Playback: playbackSnapshot,
		Messages: history,
		Queue:    h.queue.get(roomCode),
	})

	if wasNewUser {
		h.broadcastPresence(ctx, roomCode)
	}
}

func (h *Hub) handleRoomLeave(conn *connection, payload any) error {
	roomCode, err := rooms.ValidateRoomCode(field(payload, "roomCode"))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), handshakeTimeout)
	defer cancel()

	conn.socket.Leave(roomName(roomCode))
	conn.left(roomCode)

	if !h.presence.removeSocket(roomCode, conn.user.ID, string(conn.socket.Id())) {
		return nil
	}

	h.broadcastPresence(ctx, roomCode)
	h.scheduleHostTransferIfHostLeft(ctx, roomCode, conn.user.ID, conn.user.Username)

	return nil
}

func isRoomMember(members []rooms.PresenceMember, userID string) bool {
	for _, member := range members {
		if member.UserID == userID {
			return true
		}
	}
	return false
}
