package music

import "testing"

const validVideoID = "dQw4w9WgXcQ"

const validThumbnail = "https://i.ytimg.com/vi/dQw4w9WgXcQ/mqdefault.jpg"

func TestSanitizeThumbnailURLCannotPointAtAnotherOrigin(t *testing.T) {
	rejected := []any{
		"https://attacker.example/pixel.gif",
		"https://evil-ytimg.com/pixel.gif",
		"https://ytimg.com.attacker.example/pixel.gif",
		"http://i.ytimg.com/vi/abc/hq.jpg",
		"javascript:alert(1)",
		"data:image/svg+xml;base64,PHN2Zz48L3N2Zz4=",
		"//i.ytimg.com/vi/abc/hq.jpg",
		"http://169.254.169.254/latest/meta-data/",
		"file:///etc/passwd",
		"not a url",
		"",
		nil,
		12345,
	}

	for _, input := range rejected {
		if got := SanitizeThumbnailURL(input); got != "" {
			t.Fatalf("SanitizeThumbnailURL(%v) = %q, want empty", input, got)
		}
	}
}

func TestSanitizeThumbnailURLKeepsYouTubeImages(t *testing.T) {
	accepted := []string{
		validThumbnail,
		"https://i9.ytimg.com/vi/abc/hq.jpg",
		"https://img.youtube.com/vi/abc/0.jpg",
		"https://yt3.ggpht.com/a/default.jpg",
	}

	for _, input := range accepted {
		if got := SanitizeThumbnailURL(input); got != input {
			t.Fatalf("SanitizeThumbnailURL(%q) = %q", input, got)
		}
	}
}

func TestParseVideoIDHandlesTheSupportedLinkShapes(t *testing.T) {
	links := []string{
		validVideoID,
		"https://www.youtube.com/watch?v=" + validVideoID,
		"https://youtu.be/" + validVideoID,
		"https://music.youtube.com/watch?v=" + validVideoID,
		"https://www.youtube.com/embed/" + validVideoID,
		"https://www.youtube.com/shorts/" + validVideoID,
	}

	for _, link := range links {
		videoID, err := ParseVideoID(link)
		if err != nil {
			t.Fatalf("ParseVideoID(%q) failed: %v", link, err)
		}
		if videoID != validVideoID {
			t.Fatalf("ParseVideoID(%q) = %q", link, videoID)
		}
	}
}

func TestParseVideoIDRefusesOtherHosts(t *testing.T) {
	rejected := []any{
		"http://169.254.169.254/latest/meta-data/",
		"http://localhost:5000/api/me",
		"https://attacker.example/watch?v=" + validVideoID,
		"https://youtube.com.attacker.example/watch?v=" + validVideoID,
		"file:///etc/passwd",
		"",
		nil,
	}

	for _, input := range rejected {
		if videoID, err := ParseVideoID(input); err == nil {
			t.Fatalf("ParseVideoID(%v) accepted as %q", input, videoID)
		}
	}
}

func TestParsePlaylistIDHandlesTheSupportedLinkShapes(t *testing.T) {
	const playlistID = "PL1234567890abcdef"

	links := []string{
		playlistID,
		"https://www.youtube.com/playlist?list=" + playlistID,
		"https://youtube.com/watch?v=" + validVideoID + "&list=" + playlistID,
		"https://youtu.be/" + validVideoID + "?list=" + playlistID,
		"https://music.youtube.com/playlist?list=" + playlistID,
		"https://m.youtube.com/playlist?list=" + playlistID,
	}

	for _, link := range links {
		parsed, err := ParsePlaylistID(link)
		if err != nil {
			t.Fatalf("ParsePlaylistID(%q) failed: %v", link, err)
		}
		if parsed != playlistID {
			t.Fatalf("ParsePlaylistID(%q) = %q", link, parsed)
		}
	}
}

func TestParsePlaylistIDRefusesOtherHostsAndInvalidInput(t *testing.T) {
	const playlistID = "PL1234567890abcdef"

	rejected := []any{
		"http://attacker.example/playlist?list=" + playlistID,
		"https://youtube.com.attacker.example/playlist?list=" + playlistID,
		"https://www.youtube.com/watch?v=" + validVideoID,
		"https://www.youtube.com/playlist?list=",
		"file:///etc/passwd",
		"",
		nil,
		42,
	}

	for _, input := range rejected {
		if parsed, err := ParsePlaylistID(input); err == nil {
			t.Fatalf("ParsePlaylistID(%v) accepted as %q", input, parsed)
		}
	}
}

func TestIsPlaylistURLOnlyMatchesYouTubeLists(t *testing.T) {
	const playlistID = "PL1234567890abcdef"

	cases := map[string]bool{
		"https://www.youtube.com/playlist?list=" + playlistID:      true,
		"https://youtu.be/" + validVideoID + "?list=" + playlistID: true,
		"https://www.youtube.com/watch?v=" + validVideoID:          false,
		"https://attacker.example/playlist?list=" + playlistID:     false,
		"not a url": false,
	}

	for input, expected := range cases {
		if got := IsPlaylistURL(input); got != expected {
			t.Fatalf("IsPlaylistURL(%q) = %v", input, got)
		}
	}

	if IsPlaylistURL(nil) {
		t.Fatal("IsPlaylistURL(nil) must be false")
	}
}

func TestParseISO8601DurationToSeconds(t *testing.T) {
	cases := map[string]float64{
		"PT3M45S":  225,
		"PT1H2M3S": 3723,
		"P1DT1S":   86401,
		"PT0.5S":   0.5,
	}

	for input, expected := range cases {
		seconds := ParseISO8601DurationToSeconds(input)
		if seconds == nil {
			t.Fatalf("ParseISO8601DurationToSeconds(%q) returned nil", input)
		}
		if *seconds != expected {
			t.Fatalf("ParseISO8601DurationToSeconds(%q) = %v, want %v", input, *seconds, expected)
		}
	}

	for _, input := range []string{"", "3M45S", "PT", "garbage"} {
		if seconds := ParseISO8601DurationToSeconds(input); seconds != nil {
			t.Fatalf("ParseISO8601DurationToSeconds(%q) = %v, want nil", input, *seconds)
		}
	}
}
