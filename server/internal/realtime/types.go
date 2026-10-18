package realtime

import (
	"github.com/mddanish004/jamzy/server/internal/messages"
	"github.com/mddanish004/jamzy/server/internal/playback"
)

type PresenceMember struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Online   bool   `json:"online"`
}

type PresenceSnapshot struct {
	Members    []PresenceMember `json:"members"`
	HostUserID string           `json:"hostUserId"`
}

type HostUpdate struct {
	HostUserID           string `json:"hostUserId"`
	HostUsername         string `json:"hostUsername"`
	PreviousHostUserID   string `json:"previousHostUserId"`
	PreviousHostUsername string `json:"previousHostUsername"`
	Reason               string `json:"reason"`
}

type HostPending struct {
	Pending           bool    `json:"pending"`
	HostUserID        string  `json:"hostUserId"`
	HostUsername      string  `json:"hostUsername"`
	SuccessorUsername *string `json:"successorUsername,omitempty"`
	Deadline          *string `json:"deadline,omitempty"`
}

type QueueItem struct {
	VideoID      string `json:"videoId"`
	Title        string `json:"title"`
	ThumbnailURL string `json:"thumbnailUrl"`
	Duration     string `json:"duration"`
}

type joinAck struct {
	Ok       bool               `json:"ok"`
	Presence PresenceSnapshot   `json:"presence"`
	Playback *playback.Snapshot `json:"playback,omitempty"`
	Messages []messages.DTO     `json:"messages"`
	Queue    []QueueItem        `json:"queue"`
}

type playbackAck struct {
	Ok       bool              `json:"ok"`
	Playback playback.Snapshot `json:"playback"`
}

type queueAck struct {
	Ok    bool        `json:"ok"`
	Queue []QueueItem `json:"queue"`
}

type plainAck struct {
	Ok bool `json:"ok"`
}

type errorAck struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}
