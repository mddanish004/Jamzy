package messages

import (
	"strings"
	"testing"
)

func TestValidateContentTrimsAndAccepts(t *testing.T) {
	content, err := ValidateContent("  hello  ")
	if err != nil {
		t.Fatalf("ValidateContent failed: %v", err)
	}
	if content != "hello" {
		t.Fatalf("ValidateContent = %q", content)
	}
}

func TestValidateContentRejectsEmptyOversizedAndNonString(t *testing.T) {
	rejected := []any{
		"",
		"   ",
		" \t\n ",
		strings.Repeat("x", 501),
		nil,
		7,
		map[string]any{},
		[]any{},
	}

	for _, input := range rejected {
		if content, err := ValidateContent(input); err == nil {
			t.Fatalf("ValidateContent(%v) accepted as %q", input, content)
		}
	}
}

func TestValidateContentAcceptsExactlyTheMaximumLength(t *testing.T) {
	if _, err := ValidateContent(strings.Repeat("x", MaxLength)); err != nil {
		t.Fatalf("a message at the maximum length must be accepted: %v", err)
	}
}

func TestValidateContentMeasuresNonLatinTextInCodeUnits(t *testing.T) {
	message := strings.Repeat("é", MaxLength)
	if _, err := ValidateContent(message); err != nil {
		t.Fatalf("a %d unit message must be accepted: %v", MaxLength, err)
	}

	if _, err := ValidateContent(message + "x"); err == nil {
		t.Fatal("a message beyond the maximum length must be rejected")
	}
}
