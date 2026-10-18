package config

import "time"

type RateLimitRule struct {
	Limit  int
	Window time.Duration
}

type RESTLimits struct {
	Me            RateLimitRule
	CreateProfile RateLimitRule
	CreateRoom    RateLimitRule
	JoinRoom      RateLimitRule
	GetRoom       RateLimitRule
	LeaveRoom     RateLimitRule
	MusicSearch   RateLimitRule
	MusicVideo    RateLimitRule
	MusicPlaylist RateLimitRule
}

type SocketLimits struct {
	RoomJoin     RateLimitRule
	ChatSend     RateLimitRule
	Playback     RateLimitRule
	Queue        RateLimitRule
	HostTransfer RateLimitRule
}

var REST = RESTLimits{
	Me:            RateLimitRule{Limit: 60, Window: 60 * time.Second},
	CreateProfile: RateLimitRule{Limit: 10, Window: 10 * time.Minute},
	CreateRoom:    RateLimitRule{Limit: 15, Window: 5 * time.Minute},
	JoinRoom:      RateLimitRule{Limit: 30, Window: 5 * time.Minute},
	GetRoom:       RateLimitRule{Limit: 60, Window: 5 * time.Minute},
	LeaveRoom:     RateLimitRule{Limit: 60, Window: 5 * time.Minute},
	MusicSearch:   RateLimitRule{Limit: 30, Window: 60 * time.Second},
	MusicVideo:    RateLimitRule{Limit: 60, Window: 60 * time.Second},
	MusicPlaylist: RateLimitRule{Limit: 15, Window: 60 * time.Second},
}

var Socket = SocketLimits{
	RoomJoin:     RateLimitRule{Limit: 20, Window: 60 * time.Second},
	ChatSend:     RateLimitRule{Limit: 15, Window: 10 * time.Second},
	Playback:     RateLimitRule{Limit: 40, Window: 10 * time.Second},
	Queue:        RateLimitRule{Limit: 100, Window: 10 * time.Second},
	HostTransfer: RateLimitRule{Limit: 10, Window: 60 * time.Second},
}
