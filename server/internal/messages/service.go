package messages

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mddanish004/jamzy/server/internal/id"
	"github.com/mddanish004/jamzy/server/internal/timex"
)

const recentLimit = 50

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) Create(ctx context.Context, roomID string, userID string, content string) (*DTO, error) {
	const query = `
		WITH inserted AS (
			INSERT INTO messages (id, room_id, user_id, content, created_at)
			VALUES ($1, $2, $3, $4, now())
			RETURNING id, room_id, user_id, content, created_at
		)
		SELECT i.id, i.room_id, i.user_id, i.content, i.created_at, u.username
		FROM inserted i
		JOIN users u ON u.id = i.user_id`

	var (
		messageID  string
		roomIDOut  string
		userIDOut  string
		contentOut string
		createdAt  time.Time
		username   string
	)

	err := s.pool.QueryRow(ctx, query, id.New(), roomID, userID, content).
		Scan(&messageID, &roomIDOut, &userIDOut, &contentOut, &createdAt, &username)
	if err != nil {
		return nil, err
	}

	return &DTO{
		ID:        messageID,
		RoomID:    roomIDOut,
		UserID:    userIDOut,
		Username:  username,
		Content:   contentOut,
		CreatedAt: timex.ISO(createdAt),
	}, nil
}

func (s *Service) Recent(ctx context.Context, roomID string) ([]DTO, error) {
	const query = `
		SELECT m.id, m.room_id, m.user_id, m.content, m.created_at, u.username
		FROM messages m
		JOIN users u ON u.id = m.user_id
		WHERE m.room_id = $1
		ORDER BY m.created_at DESC, m.id DESC
		LIMIT $2`

	rows, err := s.pool.Query(ctx, query, roomID, recentLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	descending := make([]DTO, 0, recentLimit)
	for rows.Next() {
		var (
			messageID  string
			roomIDOut  string
			userIDOut  string
			contentOut string
			createdAt  time.Time
			username   string
		)
		if err := rows.Scan(&messageID, &roomIDOut, &userIDOut, &contentOut, &createdAt, &username); err != nil {
			return nil, err
		}
		descending = append(descending, DTO{
			ID:        messageID,
			RoomID:    roomIDOut,
			UserID:    userIDOut,
			Username:  username,
			Content:   contentOut,
			CreatedAt: timex.ISO(createdAt),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	ascending := make([]DTO, 0, len(descending))
	for i := len(descending) - 1; i >= 0; i-- {
		ascending = append(ascending, descending[i])
	}

	return ascending, nil
}
