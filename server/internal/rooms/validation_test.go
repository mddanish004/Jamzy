package rooms

import (
	"strings"
	"testing"
)

func TestRoomCodesHaveTheDocumentedShape(t *testing.T) {
	for i := 0; i < 200; i++ {
		code := GenerateRoomCode()
		if len(code) != RoomCodeLength {
			t.Fatalf("code %q has length %d", code, len(code))
		}
		for _, char := range code {
			if !strings.ContainsRune("ABCDEFGHJKMNPQRSTUVWXYZ23456789", char) {
				t.Fatalf("code %q contains %q", code, char)
			}
		}
	}
}

func TestRoomCodesAreNotPredictableFromAPinnedSource(t *testing.T) {
	seen := map[string]struct{}{}
	for i := 0; i < 50; i++ {
		seen[GenerateRoomCode()] = struct{}{}
	}

	if len(seen) < 2 {
		t.Fatalf("room codes repeated across 50 draws: %v", seen)
	}
}

func TestRoomCodesDoNotRepeatAcrossALargeSample(t *testing.T) {
	const draws = 5000

	seen := make(map[string]struct{}, draws)
	for i := 0; i < draws; i++ {
		seen[GenerateRoomCode()] = struct{}{}
	}

	if len(seen) != draws {
		t.Fatalf("expected %d distinct codes, got %d", draws, len(seen))
	}
}

func TestValidateRoomCodeNormalizesCaseAndWhitespace(t *testing.T) {
	for _, input := range []string{"ab72kd", "  ab72kd  ", "AB72KD"} {
		code, err := ValidateRoomCode(input)
		if err != nil {
			t.Fatalf("ValidateRoomCode(%q) failed: %v", input, err)
		}
		if code != "AB72KD" {
			t.Fatalf("ValidateRoomCode(%q) = %q", input, code)
		}
	}
}

func TestValidateRoomCodeRejectsMalformedInput(t *testing.T) {
	rejected := []any{
		"ABC12",
		"ABC1234",
		"AB-2KD",
		"AB 2KD",
		"../../etc",
		"%2e%2e",
		"",
		"   ",
		nil,
		42,
		map[string]any{"roomCode": "ABC123"},
		[]any{"ABC123"},
		true,
	}

	for _, input := range rejected {
		if code, err := ValidateRoomCode(input); err == nil {
			t.Fatalf("ValidateRoomCode(%v) accepted as %q", input, code)
		}
	}
}
