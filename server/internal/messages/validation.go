package messages

import (
	"fmt"
	"strings"

	"github.com/mddanish004/jamzy/server/internal/apperr"
	"github.com/mddanish004/jamzy/server/internal/textlen"
)

func ValidateContent(raw any) (string, error) {
	value, ok := raw.(string)
	if !ok {
		return "", apperr.New("Message is required", 400)
	}

	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", apperr.New("Message cannot be empty", 400)
	}

	if textlen.UTF16(trimmed) > MaxLength {
		return "", apperr.New(fmt.Sprintf("Message must be at most %d characters", MaxLength), 400)
	}

	return trimmed, nil
}
