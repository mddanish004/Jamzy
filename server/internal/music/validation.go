package music

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/mddanish004/jamzy/server/internal/apperr"
	"github.com/mddanish004/jamzy/server/internal/textlen"
)

const (
	SearchQueryMinLength = 1
	SearchQueryMaxLength = 200
	SearchMaxResults     = 20
)

var (
	videoIDPattern    = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)
	playlistIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{2,100}$`)
	barePlaylistID    = regexp.MustCompile(`^(PL|UU|LL|FL|RD|OLAK)[A-Za-z0-9_-]{10,}$`)
	iso8601Duration   = regexp.MustCompile(`^P(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?)?$`)
)

var youtubeHosts = map[string]struct{}{
	"youtube.com":              {},
	"www.youtube.com":          {},
	"m.youtube.com":            {},
	"music.youtube.com":        {},
	"youtube-nocookie.com":     {},
	"www.youtube-nocookie.com": {},
}

var youtubeShortHosts = map[string]struct{}{
	"youtu.be":     {},
	"www.youtu.be": {},
}

var pathPrefixes = map[string]struct{}{
	"embed":  {},
	"shorts": {},
	"live":   {},
	"v":      {},
}

var thumbnailHostSuffixes = []string{"ytimg.com", "ggpht.com", "youtube.com"}

func ValidVideoID(videoID string) bool {
	return videoIDPattern.MatchString(videoID)
}

func ValidateSearchQuery(raw any) (string, error) {
	value, ok := raw.(string)
	if !ok {
		return "", apperr.New("Search query is required", 400)
	}

	trimmed := strings.TrimSpace(value)
	if len(trimmed) < SearchQueryMinLength {
		return "", apperr.New("Search query is required", 400)
	}

	if textlen.UTF16(trimmed) > SearchQueryMaxLength {
		return "", apperr.New("Search query must be at most 200 characters", 400)
	}

	return trimmed, nil
}

func ParseVideoID(raw any) (string, error) {
	value, ok := raw.(string)
	if !ok {
		return "", apperr.New("A YouTube link is required", 400)
	}

	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", apperr.New("A YouTube link is required", 400)
	}

	if videoIDPattern.MatchString(trimmed) {
		return trimmed, nil
	}

	parsed, err := parseURL(trimmed)
	if err != nil {
		return "", apperr.New("That does not look like a YouTube link", 400)
	}

	host := strings.ToLower(parsed.Hostname())
	segments := pathSegments(parsed.Path)

	candidate := ""

	if _, ok := youtubeShortHosts[host]; ok {
		if len(segments) > 0 {
			candidate = segments[0]
		}
	} else if _, ok := youtubeHosts[host]; ok {
		if queryID := parsed.Query().Get("v"); queryID != "" {
			candidate = queryID
		} else if len(segments) > 1 {
			if _, ok := pathPrefixes[segments[0]]; ok {
				candidate = segments[1]
			}
		}
	} else {
		return "", apperr.New("Only YouTube links are supported", 400)
	}

	if !videoIDPattern.MatchString(candidate) {
		return "", apperr.New("Could not find a video id in that link", 400)
	}

	return candidate, nil
}

func IsPlaylistURL(raw any) bool {
	value, ok := raw.(string)
	if !ok {
		return false
	}

	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}

	parsed, err := parseURL(trimmed)
	if err != nil {
		return false
	}

	host := strings.ToLower(parsed.Hostname())
	if !isYouTubeHost(host) {
		return false
	}

	return parsed.Query().Has("list")
}

func ParsePlaylistID(raw any) (string, error) {
	value, ok := raw.(string)
	if !ok {
		return "", apperr.New("A YouTube playlist link is required", 400)
	}

	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", apperr.New("A YouTube playlist link is required", 400)
	}

	if barePlaylistID.MatchString(trimmed) {
		return trimmed, nil
	}

	parsed, err := parseURL(trimmed)
	if err != nil {
		return "", apperr.New("That does not look like a YouTube link", 400)
	}

	if !isYouTubeHost(strings.ToLower(parsed.Hostname())) {
		return "", apperr.New("Only YouTube links are supported", 400)
	}

	candidate := parsed.Query().Get("list")
	if !playlistIDPattern.MatchString(candidate) {
		return "", apperr.New("Could not find a playlist id in that link", 400)
	}

	return candidate, nil
}

func ParseISO8601DurationToSeconds(raw string) *float64 {
	match := iso8601Duration.FindStringSubmatch(strings.TrimSpace(raw))
	if match == nil {
		return nil
	}

	total := 0.0
	total += part(match[1]) * 86400
	total += part(match[2]) * 3600
	total += part(match[3]) * 60
	total += part(match[4])

	if total <= 0 {
		return nil
	}

	return &total
}

func SanitizeThumbnailURL(raw any) string {
	value, ok := raw.(string)
	if !ok {
		return ""
	}

	trimmed := strings.TrimSpace(value)
	if trimmed == "" || len(trimmed) > 500 {
		return ""
	}

	parsed, err := parseURL(trimmed)
	if err != nil {
		return ""
	}

	if parsed.Scheme != "https" {
		return ""
	}

	host := strings.ToLower(parsed.Hostname())
	for _, suffix := range thumbnailHostSuffixes {
		if host == suffix || strings.HasSuffix(host, "."+suffix) {
			return parsed.String()
		}
	}

	return ""
}

func part(value string) float64 {
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func isYouTubeHost(host string) bool {
	if _, ok := youtubeHosts[host]; ok {
		return true
	}
	_, ok := youtubeShortHosts[host]
	return ok
}

func parseURL(value string) (*url.URL, error) {
	if !strings.Contains(value, "://") {
		value = "https://" + value
	}
	return url.Parse(value)
}

func pathSegments(path string) []string {
	segments := strings.Split(path, "/")
	filtered := make([]string, 0, len(segments))
	for _, segment := range segments {
		if segment != "" {
			filtered = append(filtered, segment)
		}
	}
	return filtered
}
