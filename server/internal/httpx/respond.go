package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mddanish004/jamzy/server/internal/apperr"
	"github.com/mddanish004/jamzy/server/internal/db"
)

const maxBodyBytes = 100 * 1024

type Handler func(http.ResponseWriter, *http.Request) error

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(body)
}

func WriteUnauthorized(w http.ResponseWriter) {
	WriteJSON(w, 401, map[string]any{"authenticated": false, "message": "Unauthorized"})
}

func ReadJSON(w http.ResponseWriter, r *http.Request, dest any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	err := json.NewDecoder(r.Body).Decode(dest)
	if err == nil {
		return nil
	}

	if errors.Is(err, io.EOF) {
		return nil
	}

	var maxBytes *http.MaxBytesError
	if errors.As(err, &maxBytes) {
		return apperr.New("Request body too large", 413)
	}

	return apperr.New("Invalid request body", 400)
}

func (m *Middleware) Handle(fn Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			m.WriteError(w, err)
		}
	}
}

func (m *Middleware) WriteError(w http.ResponseWriter, err error) {
	var appError *apperr.Error
	if errors.As(err, &appError) {
		WriteJSON(w, appError.StatusCode, map[string]string{"message": appError.Message})
		return
	}

	var maxBytes *http.MaxBytesError
	if errors.As(err, &maxBytes) {
		WriteJSON(w, 413, map[string]string{"message": "Request body too large"})
		return
	}

	if constraint, ok := db.UniqueViolation(err); ok {
		m.Log.Error("Unique constraint violated", constraint)
		WriteJSON(w, 409, map[string]string{
			"message": fmt.Sprintf("A record with this %s already exists", constraintField(constraint)),
		})
		return
	}

	m.Log.Error("Unhandled error", err)
	WriteJSON(w, 500, map[string]string{"message": "Internal server error"})
}

func constraintField(constraint string) string {
	field := strings.TrimSuffix(constraint, "_key")
	field = strings.TrimSuffix(field, "_pkey")
	field = strings.TrimSuffix(field, "_idx")

	for _, prefix := range []string{"users_", "rooms_", "room_members_", "messages_"} {
		if strings.HasPrefix(field, prefix) {
			field = strings.TrimPrefix(field, prefix)
			break
		}
	}

	if field == "" {
		return "field"
	}
	return field
}
