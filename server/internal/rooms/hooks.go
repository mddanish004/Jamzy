package rooms

const (
	HostChangeReasonManual     = "manual"
	HostChangeReasonDisconnect = "disconnect"
	HostChangeReasonLeft       = "left"
)

type LeaveHooks interface {
	EvictUserFromRoom(roomCode string, userID string)
	ClearPendingTransfer(roomCode string)
	AnnounceHostChange(roomCode string, previous HostHandover, reason string)
}
