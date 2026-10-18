package realtime

import "errors"

var (
	errNotAMember   = errors.New("Not a member of this room")
	errRoomNotFound = errors.New("Room not found")
)
