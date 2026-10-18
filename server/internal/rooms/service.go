package rooms

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mddanish004/jamzy/server/internal/apperr"
	"github.com/mddanish004/jamzy/server/internal/db"
	"github.com/mddanish004/jamzy/server/internal/id"
	"github.com/mddanish004/jamzy/server/internal/timex"
	"github.com/mddanish004/jamzy/server/internal/users"
)

const roomColumns = `
	r.id, r.room_code, r.host_user_id, h.username,
	r.current_video_id, r.current_title, r.current_thumbnail_url, r.current_duration,
	r.is_playing, r.playback_position, r.playback_updated_at`

const roomMembersQuery = `
	SELECT rm.user_id, u.username, rm.joined_at
	FROM room_members rm
	JOIN users u ON u.id = rm.user_id
	WHERE rm.room_id = $1
	ORDER BY rm.joined_at ASC, rm.user_id ASC`

type roomRow struct {
	dto     DTO
	members []PresenceMember
}

type Service struct {
	pool  *pgxpool.Pool
	users *users.Service
}

func NewService(pool *pgxpool.Pool, usersService *users.Service) *Service {
	return &Service{pool: pool, users: usersService}
}

func (s *Service) CreateWithHost(ctx context.Context, clerkUserID string) (*DTO, error) {
	host, err := s.users.FindByClerkID(ctx, clerkUserID)
	if err != nil {
		return nil, err
	}
	if host == nil {
		return nil, apperr.New("User profile not found", 404)
	}

	for attempt := 0; attempt < RoomCodeMaxRetry; attempt++ {
		roomCode := GenerateRoomCode()

		roomID, joinedAt, err := s.createRoomWithMember(ctx, roomCode, host.ID)
		if err != nil {
			if _, ok := db.UniqueViolation(err); ok {
				continue
			}
			return nil, err
		}

		hostUser := RoomUser{ID: host.ID, Username: host.Username}

		return &DTO{
			ID:                  roomID,
			RoomCode:            roomCode,
			Host:                hostUser,
			Members:             []Member{{User: hostUser, JoinedAt: timex.ISO(joinedAt)}},
			CurrentVideoID:      nil,
			CurrentTitle:        nil,
			CurrentThumbnailURL: nil,
			CurrentDuration:     nil,
			IsPlaying:           false,
			PlaybackPosition:    0,
			PlaybackUpdatedAt:   nil,
		}, nil
	}

	return nil, apperr.New("Could not generate a unique room code, try again", 500)
}

func (s *Service) createRoomWithMember(ctx context.Context, roomCode string, hostUserID string) (string, time.Time, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", time.Time{}, err
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	roomID := id.New()

	if _, err := tx.Exec(ctx,
		`INSERT INTO rooms (id, room_code, host_user_id, created_at, updated_at)
		 VALUES ($1, $2, $3, now(), now())`,
		roomID, roomCode, hostUserID,
	); err != nil {
		return "", time.Time{}, err
	}

	var joinedAt time.Time
	if err := tx.QueryRow(ctx,
		`INSERT INTO room_members (room_id, user_id, joined_at)
		 VALUES ($1, $2, now())
		 RETURNING joined_at`,
		roomID, hostUserID,
	).Scan(&joinedAt); err != nil {
		return "", time.Time{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", time.Time{}, err
	}

	return roomID, joinedAt, nil
}

func (s *Service) Join(ctx context.Context, clerkUserID string, roomCode string) (*DTO, error) {
	user, err := s.users.FindByClerkID(ctx, clerkUserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperr.New("User profile not found", 404)
	}

	room, err := s.loadByCode(ctx, roomCode)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, apperr.New("Room not found", 404)
	}

	if isMember(room.members, user.ID) {
		return &room.dto, nil
	}

	if _, err := s.pool.Exec(ctx,
		`INSERT INTO room_members (room_id, user_id, joined_at)
		 VALUES ($1, $2, now())
		 ON CONFLICT (room_id, user_id) DO NOTHING`,
		room.dto.ID, user.ID,
	); err != nil {
		return nil, err
	}

	refreshed, err := s.loadByCode(ctx, roomCode)
	if err != nil {
		return nil, err
	}
	if refreshed == nil {
		return nil, apperr.New("Room not found", 404)
	}

	return &refreshed.dto, nil
}

func (s *Service) GetByCode(ctx context.Context, roomCode string) (*DTO, error) {
	room, err := s.loadByCode(ctx, roomCode)
	if err != nil || room == nil {
		return nil, err
	}
	return &room.dto, nil
}

func (s *Service) HostID(ctx context.Context, roomCode string) (*string, error) {
	var hostUserID string

	err := s.pool.QueryRow(ctx, `SELECT host_user_id FROM rooms WHERE room_code = $1`, roomCode).Scan(&hostUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &hostUserID, nil
}

func (s *Service) IDIfMember(ctx context.Context, roomCode string, userID string) (*string, error) {
	const query = `
		SELECT r.id
		FROM rooms r
		JOIN room_members rm ON rm.room_id = r.id AND rm.user_id = $2
		WHERE r.room_code = $1`

	var roomID string

	err := s.pool.QueryRow(ctx, query, roomCode, userID).Scan(&roomID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &roomID, nil
}

func (s *Service) MembersForPresence(ctx context.Context, roomCode string) (*PresenceData, error) {
	room, err := s.loadByCode(ctx, roomCode)
	if err != nil || room == nil {
		return nil, err
	}

	return &PresenceData{
		RoomID:     room.dto.ID,
		HostUserID: room.dto.Host.ID,
		Members:    room.members,
	}, nil
}

func (s *Service) TransferHost(ctx context.Context, roomID string, expectedHostUserID string, newHostUserID string) (bool, error) {
	tag, err := s.pool.Exec(ctx,
		`UPDATE rooms SET host_user_id = $3, updated_at = now()
		 WHERE id = $1 AND host_user_id = $2`,
		roomID, expectedHostUserID, newHostUserID,
	)
	if err != nil {
		return false, err
	}

	return tag.RowsAffected() == 1, nil
}

func (s *Service) Leave(ctx context.Context, clerkUserID string, roomCode string) (*LeaveResult, error) {
	user, err := s.users.FindByClerkID(ctx, clerkUserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperr.New("User profile not found", 404)
	}

	var roomID string
	var hostUserID string

	err = s.pool.QueryRow(ctx,
		`SELECT id, host_user_id FROM rooms WHERE room_code = $1`, roomCode,
	).Scan(&roomID, &hostUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.New("Room not found", 404)
	}
	if err != nil {
		return nil, err
	}

	var memberRoomID string
	err = s.pool.QueryRow(ctx,
		`SELECT room_id FROM room_members WHERE room_id = $1 AND user_id = $2`, roomID, user.ID,
	).Scan(&memberRoomID)
	if errors.Is(err, pgx.ErrNoRows) {
		return &LeaveResult{RoomDeleted: false}, nil
	}
	if err != nil {
		return nil, err
	}

	if hostUserID != user.ID {
		if _, err := s.pool.Exec(ctx,
			`DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`, roomID, user.ID,
		); err != nil {
			return nil, err
		}
		return &LeaveResult{RoomDeleted: false, LeftUserID: user.ID}, nil
	}

	successor, err := s.successor(ctx, roomID, user.ID)
	if err != nil {
		return nil, err
	}

	if successor == "" {
		if _, err := s.pool.Exec(ctx, `DELETE FROM rooms WHERE id = $1`, roomID); err != nil {
			return nil, err
		}
		return &LeaveResult{RoomDeleted: true, LeftUserID: user.ID}, nil
	}

	if err := s.handOver(ctx, roomID, successor, user.ID); err != nil {
		return nil, err
	}

	return &LeaveResult{
		RoomDeleted: false,
		LeftUserID:  user.ID,
		HostHandover: &HostHandover{
			PreviousHostUserID:   user.ID,
			PreviousHostUsername: user.Username,
			NewHostUserID:        successor,
		},
	}, nil
}

func (s *Service) handOver(ctx context.Context, roomID string, successorUserID string, leavingUserID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	if _, err := tx.Exec(ctx,
		`UPDATE rooms SET host_user_id = $2, updated_at = now() WHERE id = $1`,
		roomID, successorUserID,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx,
		`DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`, roomID, leavingUserID,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Service) successor(ctx context.Context, roomID string, excludeUserID string) (string, error) {
	var successorUserID string

	err := s.pool.QueryRow(ctx,
		`SELECT user_id FROM room_members
		 WHERE room_id = $1 AND user_id <> $2
		 ORDER BY joined_at ASC, user_id ASC
		 LIMIT 1`,
		roomID, excludeUserID,
	).Scan(&successorUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return successorUserID, nil
}

func (s *Service) loadByCode(ctx context.Context, roomCode string) (*roomRow, error) {
	const query = `
		SELECT ` + roomColumns + `
		FROM rooms r
		JOIN users h ON h.id = r.host_user_id
		WHERE r.room_code = $1`

	row := s.pool.QueryRow(ctx, query, roomCode)

	var (
		roomID       string
		code         string
		hostUserID   string
		hostUsername string
		videoID      *string
		title        *string
		thumbnailURL *string
		duration     *float64
		isPlaying    bool
		position     float64
		playbackAt   *time.Time
	)

	err := row.Scan(
		&roomID, &code, &hostUserID, &hostUsername,
		&videoID, &title, &thumbnailURL, &duration,
		&isPlaying, &position, &playbackAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	members, presence, err := s.loadMembers(ctx, roomID)
	if err != nil {
		return nil, err
	}

	return &roomRow{
		dto: DTO{
			ID:                  roomID,
			RoomCode:            code,
			Host:                RoomUser{ID: hostUserID, Username: hostUsername},
			Members:             members,
			CurrentVideoID:      videoID,
			CurrentTitle:        title,
			CurrentThumbnailURL: thumbnailURL,
			CurrentDuration:     duration,
			IsPlaying:           isPlaying,
			PlaybackPosition:    position,
			PlaybackUpdatedAt:   timex.ISOOrNil(playbackAt),
		},
		members: presence,
	}, nil
}

func (s *Service) loadMembers(ctx context.Context, roomID string) ([]Member, []PresenceMember, error) {
	rows, err := s.pool.Query(ctx, roomMembersQuery, roomID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	members := []Member{}
	presence := []PresenceMember{}

	for rows.Next() {
		var (
			userID   string
			username string
			joinedAt time.Time
		)
		if err := rows.Scan(&userID, &username, &joinedAt); err != nil {
			return nil, nil, err
		}

		members = append(members, Member{
			User:     RoomUser{ID: userID, Username: username},
			JoinedAt: timex.ISO(joinedAt),
		})
		presence = append(presence, PresenceMember{UserID: userID, Username: username})
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return members, presence, nil
}

func isMember(members []PresenceMember, userID string) bool {
	for _, member := range members {
		if member.UserID == userID {
			return true
		}
	}
	return false
}
