package music

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mddanish004/jamzy/server/internal/apperr"
	"github.com/mddanish004/jamzy/server/internal/logging"
)

const (
	youtubeSearchEndpoint   = "https://www.googleapis.com/youtube/v3/search"
	youtubeVideosEndpoint   = "https://www.googleapis.com/youtube/v3/videos"
	youtubePlaylistEndpoint = "https://www.googleapis.com/youtube/v3/playlistItems"
	playlistMaxResults      = "50"
	requestTimeout          = 10 * time.Second
)

type Service struct {
	apiKey string
	client *http.Client
	log    logging.Logger
}

func NewService(apiKey string, log logging.Logger) *Service {
	return &Service{
		apiKey: apiKey,
		client: &http.Client{Timeout: requestTimeout},
		log:    log,
	}
}

type thumbnails struct {
	Default *thumbnail `json:"default"`
	Medium  *thumbnail `json:"medium"`
	High    *thumbnail `json:"high"`
}

type thumbnail struct {
	URL string `json:"url"`
}

type searchItem struct {
	ID struct {
		VideoID string `json:"videoId"`
	} `json:"id"`
	Snippet struct {
		Title        string     `json:"title"`
		ChannelTitle string     `json:"channelTitle"`
		Thumbnails   thumbnails `json:"thumbnails"`
	} `json:"snippet"`
}

type searchPayload struct {
	Items []searchItem `json:"items"`
}

type videoDetail struct {
	ID      string `json:"id"`
	Snippet struct {
		Title        string     `json:"title"`
		ChannelTitle string     `json:"channelTitle"`
		Thumbnails   thumbnails `json:"thumbnails"`
	} `json:"snippet"`
	ContentDetails struct {
		Duration string `json:"duration"`
	} `json:"contentDetails"`
	Status struct {
		Embeddable    *bool  `json:"embeddable"`
		PrivacyStatus string `json:"privacyStatus"`
	} `json:"status"`
}

type videosPayload struct {
	Items []videoDetail `json:"items"`
}

type playlistItem struct {
	Snippet struct {
		Title      string `json:"title"`
		ResourceID struct {
			Kind    string `json:"kind"`
			VideoID string `json:"videoId"`
		} `json:"resourceId"`
	} `json:"snippet"`
	Status struct {
		PrivacyStatus string `json:"privacyStatus"`
	} `json:"status"`
}

type playlistPayload struct {
	Items []playlistItem `json:"items"`
}

func (s *Service) Search(ctx context.Context, query string) ([]SearchResult, error) {
	if !s.configured() {
		s.log.Error("YouTube API key is not configured")
		return nil, apperr.New("Music search is not available", 503)
	}

	endpoint, err := url.Parse(youtubeSearchEndpoint)
	if err != nil {
		return nil, err
	}

	params := endpoint.Query()
	params.Set("part", "snippet")
	params.Set("type", "video")
	params.Set("maxResults", strconv.Itoa(SearchMaxResults))
	params.Set("q", query)
	params.Set("key", s.apiKey)
	endpoint.RawQuery = params.Encode()

	body, status, err := s.get(ctx, endpoint.String())
	if err != nil {
		s.log.Error("YouTube search request failed", err)
		return nil, apperr.New("Music search is temporarily unavailable", 502)
	}

	switch status {
	case http.StatusForbidden, http.StatusBadRequest:
		s.log.Error("YouTube API quota exceeded or API key invalid", status)
		return nil, apperr.New("Music search is temporarily unavailable", 503)
	}
	if status != http.StatusOK {
		s.log.Error("YouTube API error", status)
		return nil, apperr.New("Music search failed", 502)
	}

	var payload searchPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, apperr.New("Music search failed — invalid response", 502)
	}

	results := make([]SearchResult, 0, len(payload.Items))
	videoIDs := make([]string, 0, len(payload.Items))
	for _, item := range payload.Items {
		if item.ID.VideoID == "" {
			continue
		}
		videoIDs = append(videoIDs, item.ID.VideoID)
	}

	if len(videoIDs) == 0 {
		return []SearchResult{}, nil
	}

	durations := s.fetchDurations(ctx, videoIDs)

	for _, item := range payload.Items {
		if item.ID.VideoID == "" {
			continue
		}
		results = append(results, SearchResult{
			VideoID:      item.ID.VideoID,
			Title:        item.Snippet.Title,
			ChannelTitle: item.Snippet.ChannelTitle,
			ThumbnailURL: pickThumbnail(item.Snippet.Thumbnails),
			Duration:     durations[item.ID.VideoID],
		})
	}

	return results, nil
}

func (s *Service) Video(ctx context.Context, videoID string) (*SearchResult, error) {
	if !s.configured() {
		s.log.Error("YouTube API key is not configured")
		return nil, apperr.New("YouTube links are not available", 503)
	}

	endpoint, err := url.Parse(youtubeVideosEndpoint)
	if err != nil {
		return nil, err
	}

	params := endpoint.Query()
	params.Set("part", "snippet,contentDetails,status")
	params.Set("id", videoID)
	params.Set("key", s.apiKey)
	endpoint.RawQuery = params.Encode()

	body, status, err := s.get(ctx, endpoint.String())
	if err != nil {
		s.log.Error("YouTube video lookup failed", err)
		return nil, apperr.New("YouTube is temporarily unavailable", 502)
	}

	switch status {
	case http.StatusForbidden, http.StatusBadRequest:
		s.log.Error("YouTube API quota exceeded or API key invalid", status)
		return nil, apperr.New("YouTube is temporarily unavailable", 503)
	}
	if status != http.StatusOK {
		s.log.Error("YouTube API error", status)
		return nil, apperr.New("Could not load that video", 502)
	}

	var payload videosPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, apperr.New("Could not load that video — invalid response", 502)
	}

	if len(payload.Items) == 0 {
		return nil, apperr.New("That video does not exist or is private", 404)
	}

	item := payload.Items[0]
	if item.Status.Embeddable != nil && !*item.Status.Embeddable {
		return nil, apperr.New("The owner of this video does not allow it to be played on other sites", 422)
	}

	return &SearchResult{
		VideoID:      item.ID,
		Title:        item.Snippet.Title,
		ChannelTitle: item.Snippet.ChannelTitle,
		ThumbnailURL: pickThumbnail(item.Snippet.Thumbnails),
		Duration:     item.ContentDetails.Duration,
	}, nil
}

func (s *Service) Playlist(ctx context.Context, playlistID string) ([]SearchResult, error) {
	if !s.configured() {
		s.log.Error("YouTube API key is not configured")
		return nil, apperr.New("YouTube links are not available", 503)
	}

	endpoint, err := url.Parse(youtubePlaylistEndpoint)
	if err != nil {
		return nil, err
	}

	params := endpoint.Query()
	params.Set("part", "snippet,status")
	params.Set("playlistId", playlistID)
	params.Set("maxResults", playlistMaxResults)
	params.Set("key", s.apiKey)
	endpoint.RawQuery = params.Encode()

	body, status, err := s.get(ctx, endpoint.String())
	if err != nil {
		s.log.Error("YouTube playlist request failed", err)
		return nil, apperr.New("YouTube is temporarily unavailable", 502)
	}

	if status == http.StatusNotFound {
		return nil, apperr.New("Playlist not found or is private", 404)
	}
	switch status {
	case http.StatusForbidden, http.StatusBadRequest:
		s.log.Error("YouTube API quota exceeded or API key invalid", status)
		return nil, apperr.New("YouTube is temporarily unavailable", 503)
	}
	if status != http.StatusOK {
		s.log.Error("YouTube playlist API error", status)
		return nil, apperr.New("Could not load that playlist", 502)
	}

	var payload playlistPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, apperr.New("Could not load that playlist — invalid response", 502)
	}

	videoIDs := make([]string, 0, len(payload.Items))
	for _, item := range payload.Items {
		if item.Status.PrivacyStatus == "private" {
			continue
		}
		if item.Snippet.ResourceID.Kind != "youtube#video" {
			continue
		}
		if item.Snippet.ResourceID.VideoID == "" {
			continue
		}
		videoIDs = append(videoIDs, item.Snippet.ResourceID.VideoID)
	}

	if len(videoIDs) == 0 {
		return nil, apperr.New("No playable videos found in this playlist", 404)
	}

	details, err := s.details(ctx, videoIDs)
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(videoIDs))
	for _, videoID := range videoIDs {
		detail, ok := details[videoID]
		if !ok {
			continue
		}
		results = append(results, detail)
	}

	if len(results) == 0 {
		return nil, apperr.New("No playable videos found in this playlist", 404)
	}

	return results, nil
}

func (s *Service) details(ctx context.Context, videoIDs []string) (map[string]SearchResult, error) {
	endpoint, err := url.Parse(youtubeVideosEndpoint)
	if err != nil {
		return nil, err
	}

	params := endpoint.Query()
	params.Set("part", "snippet,contentDetails,status")
	params.Set("id", strings.Join(videoIDs, ","))
	params.Set("key", s.apiKey)
	endpoint.RawQuery = params.Encode()

	body, status, err := s.get(ctx, endpoint.String())
	if err != nil {
		s.log.Error("YouTube videos details request failed", err)
		return nil, apperr.New("YouTube is temporarily unavailable", 502)
	}

	switch status {
	case http.StatusForbidden, http.StatusBadRequest:
		s.log.Error("YouTube API quota exceeded or API key invalid", status)
		return nil, apperr.New("YouTube is temporarily unavailable", 503)
	}
	if status != http.StatusOK {
		s.log.Error("YouTube videos API error", status)
		return nil, apperr.New("Could not load playlist videos", 502)
	}

	var payload videosPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, apperr.New("Could not load playlist videos — invalid response", 502)
	}

	details := make(map[string]SearchResult, len(payload.Items))
	for _, item := range payload.Items {
		if item.Status.Embeddable != nil && !*item.Status.Embeddable {
			continue
		}
		details[item.ID] = SearchResult{
			VideoID:      item.ID,
			Title:        item.Snippet.Title,
			ChannelTitle: item.Snippet.ChannelTitle,
			ThumbnailURL: pickThumbnail(item.Snippet.Thumbnails),
			Duration:     item.ContentDetails.Duration,
		}
	}

	return details, nil
}

func (s *Service) fetchDurations(ctx context.Context, videoIDs []string) map[string]string {
	endpoint, err := url.Parse(youtubeVideosEndpoint)
	if err != nil {
		return map[string]string{}
	}

	params := endpoint.Query()
	params.Set("part", "contentDetails")
	params.Set("id", strings.Join(videoIDs, ","))
	params.Set("key", s.apiKey)
	endpoint.RawQuery = params.Encode()

	body, status, err := s.get(ctx, endpoint.String())
	if err != nil {
		s.log.Error("YouTube videos request failed", err)
		return map[string]string{}
	}
	if status != http.StatusOK {
		s.log.Error("YouTube videos API error", status)
		return map[string]string{}
	}

	var payload videosPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return map[string]string{}
	}

	durations := make(map[string]string, len(payload.Items))
	for _, item := range payload.Items {
		durations[item.ID] = item.ContentDetails.Duration
	}
	return durations
}

func (s *Service) get(ctx context.Context, endpoint string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "application/json")

	response, err := s.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 4*1024*1024))
	if err != nil {
		return nil, response.StatusCode, err
	}

	return body, response.StatusCode, nil
}

func (s *Service) configured() bool {
	return strings.TrimSpace(s.apiKey) != ""
}

func pickThumbnail(set thumbnails) string {
	if set.Medium != nil && set.Medium.URL != "" {
		return set.Medium.URL
	}
	if set.Default != nil && set.Default.URL != "" {
		return set.Default.URL
	}
	if set.High != nil && set.High.URL != "" {
		return set.High.URL
	}
	return ""
}
