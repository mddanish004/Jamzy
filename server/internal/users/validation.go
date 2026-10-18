package users

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mddanish004/jamzy/server/internal/apperr"
	"github.com/mddanish004/jamzy/server/internal/textlen"
)

const (
	UsernameMinLength = 3
	UsernameMaxLength = 20
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func ValidateUsername(raw any) (string, error) {
	value, ok := raw.(string)
	if !ok {
		return "", apperr.New("Username is required", 400)
	}

	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", apperr.New("Username is required", 400)
	}

	if len(trimmed) < UsernameMinLength {
		return "", apperr.New(fmt.Sprintf("Username must be at least %d characters", UsernameMinLength), 400)
	}

	if textlen.UTF16(trimmed) > UsernameMaxLength {
		return "", apperr.New(fmt.Sprintf("Username must be at most %d characters", UsernameMaxLength), 400)
	}

	if !usernamePattern.MatchString(trimmed) {
		return "", apperr.New("Username can only contain letters, numbers, and underscores", 400)
	}

	return trimmed, nil
}
