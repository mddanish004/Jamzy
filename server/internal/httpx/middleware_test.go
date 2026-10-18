package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mddanish004/jamzy/server/internal/config"
	"github.com/mddanish004/jamzy/server/internal/logging"
	"github.com/mddanish004/jamzy/server/internal/ratelimit"
)

func newMiddleware(t *testing.T) *Middleware {
	t.Helper()

	return &Middleware{
		Log:     logging.New(),
		Limiter: ratelimit.NewLimiter(),
	}
}

func TestRateLimitPassesThroughUntilTheCeilingThenAnswers429(t *testing.T) {
	middleware := newMiddleware(t)
	defer middleware.Limiter.Close()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := middleware.RateLimit("test", config.RateLimitRule{Limit: 2, Window: time.Minute})(next)

	for i := 0; i < 2; i++ {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/me", nil))

		if response.Code != http.StatusOK {
			t.Fatalf("hit %d answered %d", i+1, response.Code)
		}
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/me", nil))

	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("third hit answered %d", response.Code)
	}
	if response.Header().Get("Retry-After") == "" {
		t.Fatal("Retry-After header is missing")
	}

	var body map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body is not JSON: %v", err)
	}
	if body["message"] == "" {
		t.Fatal("response body has no message")
	}
}

func TestRateLimitDoesNotLeakOneCallersBudgetIntoAnothers(t *testing.T) {
	middleware := newMiddleware(t)
	defer middleware.Limiter.Close()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := middleware.RateLimit("test", config.RateLimitRule{Limit: 1, Window: time.Minute})(next)

	request := func(address string) int {
		req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
		req.RemoteAddr = address

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)

		return response.Code
	}

	if code := request("203.0.113.7:1234"); code != http.StatusOK {
		t.Fatalf("first request from 203.0.113.7 answered %d", code)
	}
	if code := request("203.0.113.7:1234"); code != http.StatusTooManyRequests {
		t.Fatalf("second request from 203.0.113.7 answered %d", code)
	}
	if code := request("198.51.100.4:1234"); code != http.StatusOK {
		t.Fatalf("first request from 198.51.100.4 answered %d", code)
	}
}
