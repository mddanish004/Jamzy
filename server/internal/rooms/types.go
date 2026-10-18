package rooms

type RoomUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type Member struct {
	User     RoomUser `json:"user"`
	JoinedAt string   `json:"joinedAt"`
}

type DTO struct {
	ID                  string   `json:"id"`
	RoomCode            string   `json:"roomCode"`
	Host                RoomUser `json:"host"`
	Members             []Member `json:"members"`
	CurrentVideoID      *string  `json:"currentVideoId"`
	CurrentTitle        *string  `json:"currentTitle"`
	CurrentThumbnailURL *string  `json:"currentThumbnailUrl"`
	CurrentDuration     *float64 `json:"currentDuration"`
	IsPlaying           bool     `json:"isPlaying"`
	PlaybackPosition    float64  `json:"playbackPosition"`
	PlaybackUpdatedAt   *string  `json:"playbackUpdatedAt"`
}

type CreateRoomResponse struct {
	Room DTO `json:"room"`
}

type JoinRoomResponse struct {
	Room DTO `json:"room"`
}

type RoomResponse struct {
	Room DTO `json:"room"`
}

type LeaveRoomResponse struct {
	Left        bool `json:"left"`
	RoomDeleted bool `json:"roomDeleted"`
}

type PresenceMember struct {
	UserID   string
	Username string
}

type PresenceData struct {
	RoomID     string
	HostUserID string
	Members    []PresenceMember
}

type HostHandover struct {
	PreviousHostUserID   string
	PreviousHostUsername string
	NewHostUserID        string
}

type LeaveResult struct {
	RoomDeleted  bool
	LeftUserID   string
	HostHandover *HostHandover
}
