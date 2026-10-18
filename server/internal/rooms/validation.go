package rooms

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"github.com/mddanish004/jamzy/server/internal/apperr"
)

const (
	RoomCodeLength   = 6
	RoomCodeMaxRetry = 10
	roomCodeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
)

var roomCodePattern = regexp.MustCompile(`^[A-Z0-9]+$`)

func ValidateRoomCode(raw any) (string, error) {
	value, ok := raw.(string)
	if !ok {
		return "", apperr.New("Room code is required", 400)
	}

	normalized := strings.ToUpper(strings.TrimSpace(value))

	if normalized == "" {
		return "", apperr.New("Room code is required", 400)
	}

	if len(normalized) != RoomCodeLength {
		return "", apperr.New(fmt.Sprintf("Room code must be %d characters", RoomCodeLength), 400)
	}

	if !roomCodePattern.MatchString(normalized) {
		return "", apperr.New("Room code can only contain letters and numbers", 400)
	}

	return normalized, nil
}

func GenerateRoomCode() string {
	limit := big.NewInt(int64(len(roomCodeAlphabet)))

	var builder strings.Builder
	builder.Grow(RoomCodeLength)

	for i := 0; i < RoomCodeLength; i++ {
		index, err := rand.Int(rand.Reader, limit)
		if err != nil {
			builder.WriteByte(roomCodeAlphabet[i%len(roomCodeAlphabet)])
			continue
		}
		builder.WriteByte(roomCodeAlphabet[index.Int64()])
	}

	return builder.String()
}
