package realtime

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mddanish004/jamzy/server/internal/config"
	"github.com/mddanish004/jamzy/server/internal/rooms"
	"github.com/mddanish004/jamzy/server/internal/timex"
)

const hostTransferGrace = 30 * time.Second

type pendingTransfer struct {
	timer      *time.Timer
	hostUserID string
	deadline   time.Time
}

type pendingTransferStore struct {
	mu      sync.Mutex
	pending map[string]*pendingTransfer
}

func newPendingTransferStore() *pendingTransferStore {
	return &pendingTransferStore{pending: map[string]*pendingTransfer{}}
}

func (s *pendingTransferStore) arm(roomCode string, hostUserID string, delay time.Duration, onFire func()) time.Time {
	s.cancel(roomCode)

	deadline := time.Now().Add(delay)

	timer := time.AfterFunc(delay, func() {
		s.mu.Lock()
		delete(s.pending, roomCode)
		s.mu.Unlock()

		onFire()
	})

	s.mu.Lock()
	s.pending[roomCode] = &pendingTransfer{
		timer:      timer,
		hostUserID: hostUserID,
		deadline:   deadline,
	}
	s.mu.Unlock()

	return deadline
}

func (s *pendingTransferStore) cancel(roomCode string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.pending[roomCode]
	if !ok {
		return "", false
	}

	entry.timer.Stop()
	delete(s.pending, roomCode)

	return entry.hostUserID, true
}

func (s *pendingTransferStore) isPending(roomCode string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.pending[roomCode]
	return ok
}

func (h *Hub) handleHostTransfer(conn *connection, payload any, reply replyFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), handshakeTimeout)
	defer cancel()

	roomCode, ok := h.guardMember(conn, payload, reply, "socket:hostTransfer", config.Socket.HostTransfer, "hostTransfer")
	if !ok {
		return
	}
	if !h.guardHost(ctx, conn, roomCode, reply, "hand over the room", "Could not transfer host") {
		return
	}

	targetUserID, _ := field(payload, "userId").(string)
	targetUserID = strings.TrimSpace(targetUserID)

	if targetUserID == "" {
		reply(errorAck{Ok: false, Message: "No member selected"})
		return
	}
	if targetUserID == conn.user.ID {
		reply(errorAck{Ok: false, Message: "You are already the host"})
		return
	}

	roomData, err := h.rooms.MembersForPresence(ctx, roomCode)
	if err != nil {
		h.replyError(reply, err, "Could not transfer host")
		return
	}
	if roomData == nil {
		reply(errorAck{Ok: false, Message: errRoomNotFound.Error()})
		return
	}

	target, ok := findMember(roomData.Members, targetUserID)
	if !ok {
		reply(errorAck{Ok: false, Message: "That person is not a member of this room"})
		return
	}

	applied, err := h.rooms.TransferHost(ctx, roomData.RoomID, conn.user.ID, targetUserID)
	if err != nil {
		h.replyError(reply, err, "Could not transfer host")
		return
	}
	if !applied {
		reply(errorAck{Ok: false, Message: "The host changed — try again"})
		return
	}

	h.pending.cancel(roomCode)

	h.log.Info(fmt.Sprintf(
		"Host transferred in %s: %s → %s (manual)",
		roomCode, conn.user.Username, target.Username,
	))

	h.announceHostChange(ctx, roomCode, rooms.HostHandover{
		PreviousHostUserID:   conn.user.ID,
		PreviousHostUsername: conn.user.Username,
	}, rooms.HostChangeReasonManual)

	reply(plainAck{Ok: true})
}

func (h *Hub) scheduleHostTransferIfHostLeft(ctx context.Context, roomCode string, departedUserID string, departedUsername string) {
	roomData, err := h.rooms.MembersForPresence(ctx, roomCode)
	if err != nil {
		h.log.Error("Host transfer scheduling failed", roomCode, err)
		return
	}
	if roomData == nil {
		return
	}
	if roomData.HostUserID != departedUserID {
		return
	}
	if h.presence.isOnline(roomCode, departedUserID) {
		return
	}

	successor, ok := h.pickSuccessor(roomCode, roomData, departedUserID)
	if !ok {
		return
	}

	deadline := h.pending.arm(roomCode, departedUserID, hostTransferGrace, func() {
		h.runScheduledTransfer(roomCode, departedUserID, departedUsername)
	})

	h.log.Info(fmt.Sprintf(
		"Host %s dropped in %s — transfer to %s in %s",
		departedUsername, roomCode, successor.Username, hostTransferGrace,
	))

	successorUsername := successor.Username
	deadlineISO := timex.ISO(deadline)

	h.emitToRoom(roomCode, "host:pending", HostPending{
		Pending:           true,
		HostUserID:        departedUserID,
		HostUsername:      departedUsername,
		SuccessorUsername: &successorUsername,
		Deadline:          &deadlineISO,
	})
}

func (h *Hub) runScheduledTransfer(roomCode string, expectedHostUserID string, expectedHostUsername string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			h.log.Error("Scheduled host transfer panicked", roomCode, fmt.Sprint(recovered))
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), handshakeTimeout)
	defer cancel()

	transferred := false

	defer func() {
		if !transferred {
			h.emitToRoom(roomCode, "host:pending", HostPending{
				Pending:      false,
				HostUserID:   expectedHostUserID,
				HostUsername: expectedHostUsername,
			})
		}
	}()

	roomData, err := h.rooms.MembersForPresence(ctx, roomCode)
	if err != nil {
		h.log.Error("Scheduled host transfer failed", roomCode, err)
		return
	}
	if roomData == nil {
		return
	}
	if roomData.HostUserID != expectedHostUserID {
		return
	}
	if h.presence.isOnline(roomCode, expectedHostUserID) {
		return
	}

	successor, ok := h.pickSuccessor(roomCode, roomData, expectedHostUserID)
	if !ok {
		return
	}

	applied, err := h.rooms.TransferHost(ctx, roomData.RoomID, expectedHostUserID, successor.UserID)
	if err != nil {
		h.log.Error("Scheduled host transfer failed", roomCode, err)
		return
	}
	if !applied {
		return
	}

	transferred = true

	h.log.Info(fmt.Sprintf(
		"Host transferred in %s: %s → %s (disconnect)",
		roomCode, expectedHostUsername, successor.Username,
	))

	h.announceHostChange(ctx, roomCode, rooms.HostHandover{
		PreviousHostUserID:   expectedHostUserID,
		PreviousHostUsername: expectedHostUsername,
	}, rooms.HostChangeReasonDisconnect)
}

func (h *Hub) cancelPendingTransferOnReturn(ctx context.Context, roomCode string, userID string, username string, currentHostUserID string) {
	if !h.pending.isPending(roomCode) {
		return
	}
	if currentHostUserID != userID {
		return
	}

	cancelledFor, ok := h.pending.cancel(roomCode)
	if !ok || cancelledFor != userID {
		return
	}

	h.log.Info(fmt.Sprintf("Host %s returned to %s — transfer cancelled", username, roomCode))

	h.emitToRoom(roomCode, "host:pending", HostPending{
		Pending:      false,
		HostUserID:   userID,
		HostUsername: username,
	})
}

func (h *Hub) ClearPendingTransfer(roomCode string) {
	h.pending.cancel(roomCode)
}

func (h *Hub) AnnounceHostChange(roomCode string, previous rooms.HostHandover, reason string) {
	ctx, cancel := context.WithTimeout(context.Background(), handshakeTimeout)
	defer cancel()

	h.announceHostChange(ctx, roomCode, previous, reason)
}

func (h *Hub) announceHostChange(ctx context.Context, roomCode string, previous rooms.HostHandover, reason string) {
	roomData, err := h.rooms.MembersForPresence(ctx, roomCode)
	if err != nil {
		h.log.Error("Host change broadcast failed", roomCode, err)
		return
	}
	if roomData == nil {
		return
	}

	hostUsername := ""
	for _, member := range roomData.Members {
		if member.UserID == roomData.HostUserID {
			hostUsername = member.Username
			break
		}
	}

	h.emitToRoom(roomCode, "host:update", HostUpdate{
		HostUserID:           roomData.HostUserID,
		HostUsername:         hostUsername,
		PreviousHostUserID:   previous.PreviousHostUserID,
		PreviousHostUsername: previous.PreviousHostUsername,
		Reason:               reason,
	})
	h.emitPresence(roomCode, roomData)
}

func (h *Hub) pickSuccessor(roomCode string, roomData *rooms.PresenceData, excludeUserID string) (rooms.PresenceMember, bool) {
	for _, member := range roomData.Members {
		if member.UserID == excludeUserID {
			continue
		}
		if h.presence.isOnline(roomCode, member.UserID) {
			return member, true
		}
	}

	return rooms.PresenceMember{}, false
}

func findMember(members []rooms.PresenceMember, userID string) (rooms.PresenceMember, bool) {
	for _, member := range members {
		if member.UserID == userID {
			return member, true
		}
	}

	return rooms.PresenceMember{}, false
}
