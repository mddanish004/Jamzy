package playback

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mddanish004/jamzy/server/internal/apperr"
	"github.com/mddanish004/jamzy/server/internal/music"
	"github.com/mddanish004/jamzy/server/internal/timex"
)

const playbackColumns = `
	current_video_id, current_title, current_thumbnail_url, current_duration,
	is_playing, playback_position, playback_updated_at`

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) Get(ctx context.Context, roomCode string) (*Snapshot, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+playbackColumns+` FROM rooms WHERE room_code = $1`, roomCode)

	snapshot, err := scanSnapshot(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return snapshot, err
}

func (s *Service) SetTrack(ctx context.Context, roomCode string, userID string, track Track) (*Snapshot, error) {
	const query = `
		UPDATE rooms SET
			current_video_id = $2,
			current_title = $3,
			current_thumbnail_url = $4,
			current_duration = $5,
			is_playing = true,
			playback_position = 0,
			playback_updated_at = timezone('UTC', now()),
			playback_updated_by = $6,
			updated_at = now()
		WHERE room_code = $1
		RETURNING ` + playbackColumns

	row := s.pool.QueryRow(ctx, query,
		roomCode,
		track.VideoID,
		track.Title,
		track.ThumbnailURL,
		music.ParseISO8601DurationToSeconds(track.Duration),
		userID,
	)

	return requireSnapshot(row)
}

func (s *Service) ApplyControl(ctx context.Context, roomCode string, userID string, action Action, position float64) (*Snapshot, error) {
	var (
		currentVideoID *string
		isPlaying      bool
	)

	err := s.pool.QueryRow(ctx,
		`SELECT current_video_id, is_playing FROM rooms WHERE room_code = $1`, roomCode,
	).Scan(&currentVideoID, &isPlaying)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.New("Room not found", 404)
	}
	if err != nil {
		return nil, err
	}

	if currentVideoID == nil {
		return nil, apperr.New("Nothing is loaded in this room", 409)
	}

	safePosition := position
	if math.IsNaN(safePosition) || math.IsInf(safePosition, 0) || safePosition < 0 {
		safePosition = 0
	}

	nextIsPlaying := isPlaying
	switch action {
	case ActionPause:
		nextIsPlaying = false
	case ActionPlay:
		nextIsPlaying = true
	}

	const query = `
		UPDATE rooms SET
			is_playing = $2,
			playback_position = $3,
			playback_updated_at = timezone('UTC', now()),
			playback_updated_by = $4,
			updated_at = now()
		WHERE room_code = $1
		RETURNING ` + playbackColumns

	row := s.pool.QueryRow(ctx, query, roomCode, nextIsPlaying, safePosition, userID)

	return requireSnapshot(row)
}

func (s *Service) Clear(ctx context.Context, roomCode string, userID string) (*Snapshot, error) {
	const query = `
		UPDATE rooms SET
			current_video_id = NULL,
			current_title = NULL,
			current_thumbnail_url = NULL,
			current_duration = NULL,
			is_playing = false,
			playback_position = 0,
			playback_updated_at = timezone('UTC', now()),
			playback_updated_by = $2,
			updated_at = now()
		WHERE room_code = $1
		RETURNING ` + playbackColumns

	row := s.pool.QueryRow(ctx, query, roomCode, userID)

	return requireSnapshot(row)
}

func scanSnapshot(row pgx.Row) (*Snapshot, error) {
	var (
		videoID      *string
		title        *string
		thumbnailURL *string
		duration     *float64
		isPlaying    bool
		position     float64
		updatedAt    *time.Time
	)

	if err := row.Scan(&videoID, &title, &thumbnailURL, &duration, &isPlaying, &position, &updatedAt); err != nil {
		return nil, err
	}

	return &Snapshot{
		VideoID:      videoID,
		Title:        title,
		ThumbnailURL: thumbnailURL,
		Duration:     duration,
		IsPlaying:    isPlaying,
		Position:     position,
		UpdatedAt:    timex.ISOOrNil(updatedAt),
		ServerTime:   timex.ISO(time.Now()),
	}, nil
}

func requireSnapshot(row pgx.Row) (*Snapshot, error) {
	snapshot, err := scanSnapshot(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.New("Room not found", 404)
	}
	return snapshot, err
}
