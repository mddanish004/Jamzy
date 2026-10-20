package openapi

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSpecServesValidDocument(t *testing.T) {
	recorder := httptest.NewRecorder()

	Spec(recorder, httptest.NewRequest("GET", "/docs/openapi.json", nil))

	if recorder.Code != 200 {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json; charset=utf-8" {
		t.Fatalf("content type = %q, want application/json; charset=utf-8", contentType)
	}

	var document struct {
		OpenAPI string                     `json:"openapi"`
		Paths   map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &document); err != nil {
		t.Fatalf("decode spec: %v", err)
	}
	if !strings.HasPrefix(document.OpenAPI, "3.1") {
		t.Fatalf("openapi version = %q, want 3.1.x", document.OpenAPI)
	}

	for _, path := range []string{
		"/api/health",
		"/api/me",
		"/api/users/profile",
		"/api/rooms",
		"/api/rooms/{roomCode}",
		"/api/rooms/{roomCode}/join",
		"/api/rooms/{roomCode}/leave",
		"/api/music/search",
		"/api/music/video",
		"/api/music/playlist",
	} {
		if _, ok := document.Paths[path]; !ok {
			t.Errorf("spec is missing path %s", path)
		}
	}
	if len(document.Paths) != 10 {
		t.Errorf("paths = %d, want 10", len(document.Paths))
	}
}

func TestDocsServesReferencePage(t *testing.T) {
	recorder := httptest.NewRecorder()

	Docs(recorder, httptest.NewRequest("GET", "/docs", nil))

	if recorder.Code != 200 {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Fatalf("content type = %q, want text/html; charset=utf-8", contentType)
	}
	if !strings.Contains(recorder.Body.String(), "/docs/openapi.json") {
		t.Fatal("docs page does not reference /docs/openapi.json")
	}
}
