package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mddanish004/jamzy/server/internal/httpx"
	"github.com/mddanish004/jamzy/server/internal/logging"
	"github.com/mddanish004/jamzy/server/internal/music"
	"github.com/mddanish004/jamzy/server/internal/ratelimit"
	"github.com/mddanish004/jamzy/server/internal/rooms"
	"github.com/mddanish004/jamzy/server/internal/users"
)

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()

	limiter := ratelimit.NewLimiter()
	t.Cleanup(limiter.Close)

	log := logging.New()

	return New(Deps{
		Middleware: &httpx.Middleware{
			Log:         log,
			Limiter:     limiter,
			CORSOrigins: []string{"http://localhost:5173"},
		},
		Users: &users.Handlers{},
		Rooms: &rooms.Handlers{},
		Music: &music.Handlers{},
		Log:   log,
	})
}

func TestDocsRoutes(t *testing.T) {
	handler := newTestHandler(t)

	cases := []struct {
		path        string
		status      int
		contentType string
	}{
		{"/docs", 200, "text/html; charset=utf-8"},
		{"/docs/", 200, "text/html; charset=utf-8"},
		{"/docs/openapi.json", 200, "application/json; charset=utf-8"},
		{"/docs/missing", 404, "application/json; charset=utf-8"},
	}

	for _, tc := range cases {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest("GET", tc.path, nil))

		if recorder.Code != tc.status {
			t.Errorf("%s status = %d, want %d", tc.path, recorder.Code, tc.status)
		}
		if contentType := recorder.Header().Get("Content-Type"); contentType != tc.contentType {
			t.Errorf("%s content type = %q, want %q", tc.path, contentType, tc.contentType)
		}
	}
}

func TestProtectedRoutesRequireAuth(t *testing.T) {
	handler := newTestHandler(t)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/api/me", nil))

	if recorder.Code != 401 {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"authenticated":false`) {
		t.Fatalf("body = %q, want authenticated false", recorder.Body.String())
	}
}
