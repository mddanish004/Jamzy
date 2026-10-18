package music

import (
	"net/http"

	"github.com/mddanish004/jamzy/server/internal/auth"
	"github.com/mddanish004/jamzy/server/internal/httpx"
)

type Handlers struct {
	Service *Service
}

func (h *Handlers) Search(w http.ResponseWriter, r *http.Request) error {
	if _, err := auth.ClerkUserID(r.Context()); err != nil {
		httpx.WriteUnauthorized(w)
		return nil
	}

	query, err := ValidateSearchQuery(r.URL.Query().Get("q"))
	if err != nil {
		return err
	}

	results, err := h.Service.Search(r.Context(), query)
	if err != nil {
		return err
	}

	httpx.WriteJSON(w, 200, SearchResponse{Results: results})
	return nil
}

func (h *Handlers) ResolveVideo(w http.ResponseWriter, r *http.Request) error {
	if _, err := auth.ClerkUserID(r.Context()); err != nil {
		httpx.WriteUnauthorized(w)
		return nil
	}

	videoID, err := ParseVideoID(r.URL.Query().Get("url"))
	if err != nil {
		return err
	}

	result, err := h.Service.Video(r.Context(), videoID)
	if err != nil {
		return err
	}

	httpx.WriteJSON(w, 200, VideoResponse{Result: *result})
	return nil
}

func (h *Handlers) ResolvePlaylist(w http.ResponseWriter, r *http.Request) error {
	if _, err := auth.ClerkUserID(r.Context()); err != nil {
		httpx.WriteUnauthorized(w)
		return nil
	}

	playlistID, err := ParsePlaylistID(r.URL.Query().Get("url"))
	if err != nil {
		return err
	}

	results, err := h.Service.Playlist(r.Context(), playlistID)
	if err != nil {
		return err
	}

	httpx.WriteJSON(w, 200, SearchResponse{Results: results})
	return nil
}
