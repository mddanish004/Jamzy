package users

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mddanish004/jamzy/server/internal/id"
)

const selectUserColumns = `id, clerk_user_id, username`

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) FindByClerkID(ctx context.Context, clerkUserID string) (*PublicUser, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+selectUserColumns+` FROM users WHERE clerk_user_id = $1`, clerkUserID)
	return scanUser(row)
}

func (s *Service) FindByUsername(ctx context.Context, username string) (*PublicUser, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+selectUserColumns+` FROM users WHERE username = $1`, username)
	return scanUser(row)
}

func (s *Service) Create(ctx context.Context, clerkUserID string, username string) (*PublicUser, error) {
	row := s.pool.QueryRow(ctx,
		`INSERT INTO users (id, clerk_user_id, username, created_at, updated_at)
		 VALUES ($1, $2, $3, now(), now())
		 RETURNING `+selectUserColumns,
		id.New(), clerkUserID, username,
	)
	return scanUser(row)
}

func scanUser(row pgx.Row) (*PublicUser, error) {
	var user PublicUser
	err := row.Scan(&user.ID, &user.ClerkUserID, &user.Username)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
