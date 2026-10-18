package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/clerk/clerk-sdk-go/v2"
)

func TestVerifyFetchesKeySetWhenCacheIsCold(t *testing.T) {
	var requestedPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"keys":[]}`))
	}))
	defer srv.Close()

	url := srv.URL
	key := "sk_test_stub"
	clerk.SetBackend(clerk.NewBackend(&clerk.BackendConfig{
		URL:        &url,
		HTTPClient: srv.Client(),
		Key:        &key,
	}))
	t.Cleanup(func() { clerk.SetBackend(nil) })

	verifier := NewClerkVerifier(key, nil)

	_, err := verifier.Verify(context.Background(), tokenWithKeyID(t, "test-key-id"))

	if requestedPath != "/jwks" {
		t.Fatalf("expected the verifier to fetch the key set, requested %q", requestedPath)
	}
	if err == nil {
		t.Fatal("expected verification to fail: the stub key set publishes no keys")
	}
	if !strings.Contains(err.Error(), "missing json web key") {
		t.Fatalf("expected the fetched key set to be reported as empty, got %v", err)
	}
}

func TestVerifyRejectsTokenWithoutKeyID(t *testing.T) {
	verifier := NewClerkVerifier("sk_test_stub", nil)

	header, err := json.Marshal(map[string]string{"alg": "RS256"})
	if err != nil {
		t.Fatalf("marshal header: %v", err)
	}
	token := base64.RawURLEncoding.EncodeToString(header) + ".payload.signature"

	if _, err := verifier.Verify(context.Background(), token); err == nil {
		t.Fatal("expected a token without a key id to be rejected")
	}
}

func tokenWithKeyID(t *testing.T, keyID string) string {
	t.Helper()

	header, err := json.Marshal(map[string]string{"alg": "RS256", "kid": keyID})
	if err != nil {
		t.Fatalf("marshal header: %v", err)
	}

	return base64.RawURLEncoding.EncodeToString(header) + ".payload.signature"
}
