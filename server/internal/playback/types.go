package playback

type Track struct {
	VideoID      string `json:"videoId"`
	Title        string `json:"title"`
	ThumbnailURL string `json:"thumbnailUrl"`
	Duration     string `json:"duration"`
}

type Snapshot struct {
	VideoID      *string  `json:"videoId"`
	Title        *string  `json:"title"`
	ThumbnailURL *string  `json:"thumbnailUrl"`
	Duration     *float64 `json:"duration"`
	IsPlaying    bool     `json:"isPlaying"`
	Position     float64  `json:"position"`
	UpdatedAt    *string  `json:"updatedAt"`
	ServerTime   string   `json:"serverTime"`
}

type Action string

const (
	ActionPlay  Action = "play"
	ActionPause Action = "pause"
	ActionSeek  Action = "seek"
)

func ParseAction(raw any) (Action, bool) {
	value, ok := raw.(string)
	if !ok {
		return "", false
	}

	switch Action(value) {
	case ActionPlay, ActionPause, ActionSeek:
		return Action(value), true
	default:
		return "", false
	}
}
