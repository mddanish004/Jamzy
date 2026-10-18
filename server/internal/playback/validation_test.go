package playback

import (
	"strings"
	"testing"
)

const validVideoID = "dQw4w9WgXcQ"

const validThumbnail = "https://i.ytimg.com/vi/dQw4w9WgXcQ/mqdefault.jpg"

func TestValidateTrackAcceptsARealSearchResult(t *testing.T) {
	track, err := ValidateTrack(map[string]any{
		"videoId":      validVideoID,
		"title":        "Some song",
		"thumbnailUrl": validThumbnail,
		"duration":     "PT3M45S",
	})
	if err != nil {
		t.Fatalf("ValidateTrack failed: %v", err)
	}

	if track.VideoID != validVideoID || track.Title != "Some song" || track.ThumbnailURL != validThumbnail || track.Duration != "PT3M45S" {
		t.Fatalf("ValidateTrack returned %+v", track)
	}
}

func TestValidateTrackRejectsNonVideoIDs(t *testing.T) {
	rejected := []any{
		"",
		"short",
		"waytoolongvideoid123",
		"../../../etc/passwd",
		"<script>alert(1)",
		"abcdefghijk;rm -rf",
		nil,
		42,
		map[string]any{"toString": validVideoID},
	}

	for _, videoID := range rejected {
		track, err := ValidateTrack(map[string]any{
			"videoId":      videoID,
			"title":        "t",
			"thumbnailUrl": "",
			"duration":     "",
		})
		if err == nil {
			t.Fatalf("ValidateTrack accepted videoId %v as %+v", videoID, track)
		}
	}
}

func TestValidateTrackRejectsNonObjectPayloads(t *testing.T) {
	for _, payload := range []any{nil, "track", 5, []any{}, true} {
		if track, err := ValidateTrack(payload); err == nil {
			t.Fatalf("ValidateTrack(%v) accepted as %+v", payload, track)
		}
	}
}

func TestValidateTrackStripsAHostileThumbnailButKeepsTheTrackPlayable(t *testing.T) {
	track, err := ValidateTrack(map[string]any{
		"videoId":      validVideoID,
		"title":        strings.Repeat("x", 1000),
		"thumbnailUrl": "https://attacker.example/track.gif",
		"duration":     "PT1M",
	})
	if err != nil {
		t.Fatalf("ValidateTrack failed: %v", err)
	}

	if track.ThumbnailURL != "" {
		t.Fatalf("hostile thumbnail survived as %q", track.ThumbnailURL)
	}
	if len([]rune(track.Title)) != titleMaxLength {
		t.Fatalf("title was not clamped: %d runes", len([]rune(track.Title)))
	}
	if track.VideoID != validVideoID {
		t.Fatalf("video id changed to %q", track.VideoID)
	}
}

func TestParseActionOnlyAcceptsTheDocumentedActions(t *testing.T) {
	for _, raw := range []any{"play", "pause", "seek"} {
		if _, ok := ParseAction(raw); !ok {
			t.Fatalf("ParseAction(%v) was rejected", raw)
		}
	}

	for _, raw := range []any{"stop", "PLAY", "", nil, 3, map[string]any{}} {
		if action, ok := ParseAction(raw); ok {
			t.Fatalf("ParseAction(%v) accepted as %q", raw, action)
		}
	}
}
