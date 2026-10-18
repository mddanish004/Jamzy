package rooms

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/mddanish004/jamzy/server/internal/apperr"
	"github.com/mddanish004/jamzy/server/internal/auth"
	"github.com/mddanish004/jamzy/server/internal/httpx"
	"github.com/mddanish004/jamzy/server/internal/logging"
)

type Handlers struct {
	Service *Service
	Hooks   LeaveHooks
	Log     logging.Logger
}

func (h *Handlers) CreateRoom(w http.ResponseWriter, r *http.Request) error {
	clerkUserID, err := auth.ClerkUserID(r.Context())
	if err != nil {
		httpx.WriteUnauthorized(w)
		return nil
	}

	room, err := h.Service.CreateWithHost(r.Context(), clerkUserID)
	if err != nil {
		return err
	}

	httpx.WriteJSON(w, 201, CreateRoomResponse{Room: *room})
	return nil
}

func (h *Handlers) JoinRoom(w http.ResponseWriter, r *http.Request) error {
	clerkUserID, err := auth.ClerkUserID(r.Context())
	if err != nil {
		httpx.WriteUnauthorized(w)
		return nil
	}

	roomCode, err := ValidateRoomCode(r.PathValue("roomCode"))
	if err != nil {
		return err
	}

	room, err := h.Service.Join(r.Context(), clerkUserID, roomCode)
	if err != nil {
		return err
	}

	httpx.WriteJSON(w, 200, JoinRoomResponse{Room: *room})
	return nil
}

func (h *Handlers) GetRoom(w http.ResponseWriter, r *http.Request) error {
	if _, err := auth.ClerkUserID(r.Context()); err != nil {
		httpx.WriteUnauthorized(w)
		return nil
	}

	roomCode, err := ValidateRoomCode(r.PathValue("roomCode"))
	if err != nil {
		return err
	}

	room, err := h.Service.GetByCode(r.Context(), roomCode)
	if err != nil {
		return err
	}
	if room == nil {
		return apperr.New("Room not found", 404)
	}

	httpx.WriteJSON(w, 200, RoomResponse{Room: *room})
	return nil
}

func (h *Handlers) LeaveRoom(w http.ResponseWriter, r *http.Request) error {
	clerkUserID, err := auth.ClerkUserID(r.Context())
	if err != nil {
		httpx.WriteUnauthorized(w)
		return nil
	}

	roomCode, err := ValidateRoomCode(r.PathValue("roomCode"))
	if err != nil {
		return err
	}

	result, err := h.Service.Leave(r.Context(), clerkUserID, roomCode)
	if err != nil {
		return err
	}

	h.afterLeave(roomCode, result)

	httpx.WriteJSON(w, 200, LeaveRoomResponse{Left: true, RoomDeleted: result.RoomDeleted})
	return nil
}

func (h *Handlers) afterLeave(roomCode string, result *LeaveResult) {
	if h.Hooks == nil {
		return
	}

	handover := result.HostHandover

	h.runAsync(func() {
		if result.LeftUserID != "" {
			h.Hooks.EvictUserFromRoom(roomCode, result.LeftUserID)
		}

		if handover != nil {
			h.Hooks.ClearPendingTransfer(roomCode)
			h.Hooks.AnnounceHostChange(roomCode, *handover, HostChangeReasonLeft)
		}
	})
}

func (h *Handlers) runAsync(fn func()) {
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				h.Log.Error("Room broadcast failed", fmt.Sprint(recovered), string(debug.Stack()))
			}
		}()

		fn()
	}()
}
