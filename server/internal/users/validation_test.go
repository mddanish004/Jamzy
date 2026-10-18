package users

import (
	"strings"
	"testing"
)

func TestValidateUsernameTrimsAndAccepts(t *testing.T) {
	cases := map[string]string{
		"  john123 ": "john123",
		"a_b_C_9":    "a_b_C_9",
		"__proto__":  "__proto__",
		"john\n ":    "john",
	}

	for input, expected := range cases {
		username, err := ValidateUsername(input)
		if err != nil {
			t.Fatalf("ValidateUsername(%q) failed: %v", input, err)
		}
		if username != expected {
			t.Fatalf("ValidateUsername(%q) = %q, want %q", input, username, expected)
		}
	}
}

func TestValidateUsernameRejectsOutsideTheCharacterSet(t *testing.T) {
	rejected := []any{
		"ab",
		strings.Repeat("a", 21),
		"<script>",
		"john doe",
		"john\x00",
		"jo\nhn",
		"üser",
		"__proto__!",
		"",
		"   ",
		nil,
		123,
		map[string]any{"username": "john"},
		[]any{"john"},
	}

	for _, input := range rejected {
		if username, err := ValidateUsername(input); err == nil {
			t.Fatalf("ValidateUsername(%v) accepted as %q", input, username)
		}
	}
}

func TestValidateUsernameCountsUnicodeTheWayJavaScriptDoes(t *testing.T) {
	if _, err := ValidateUsername(strings.Repeat("a", 20)); err != nil {
		t.Fatalf("20 characters should be accepted: %v", err)
	}

	if _, err := ValidateUsername("日本語です"); err == nil {
		t.Fatal("non-ASCII characters must not pass the charset check")
	}
}
