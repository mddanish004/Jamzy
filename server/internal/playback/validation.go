package playback

import (
	"errors"

	"github.com/mddanish004/jamzy/server/internal/music"
)

const (
	titleMaxLength    = 300
	durationMaxLength = 20
)

func ValidateTrack(raw any) (Track, error) {
	fields, ok := raw.(map[string]any)
	if !ok {
		return Track{}, errors.New("Track is required")
	}

	videoID := stringValue(fields["videoId"])
	if !music.ValidVideoID(videoID) {
		return Track{}, errors.New("Invalid video id")
	}

	return Track{
		VideoID:      videoID,
		Title:        boundedString(fields["title"], titleMaxLength),
		ThumbnailURL: music.SanitizeThumbnailURL(fields["thumbnailUrl"]),
		Duration:     boundedString(fields["duration"], durationMaxLength),
	}, nil
}

func stringValue(raw any) string {
	value, ok := raw.(string)
	if !ok {
		return ""
	}
	return value
}

func boundedString(raw any, max int) string {
	value := stringValue(raw)
	runes := []rune(value)
	if len(runes) > max {
		return string(runes[:max])
	}
	return value
}
