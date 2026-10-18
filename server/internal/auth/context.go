package auth

import (
	"context"

	"github.com/mddanish004/jamzy/server/internal/apperr"
)

type contextKey struct{}

type Identity struct {
	ClerkUserID string
}

func WithIdentity(ctx context.Context, identity Identity) context.Context {
	return context.WithValue(ctx, contextKey{}, identity)
}

func ClerkUserID(ctx context.Context) (string, error) {
	identity, ok := ctx.Value(contextKey{}).(Identity)
	if !ok || identity.ClerkUserID == "" {
		return "", apperr.New("Unauthorized", 401)
	}
	return identity.ClerkUserID, nil
}

func ErrUnauthorized() error {
	return apperr.New("Unauthorized", 401)
}
