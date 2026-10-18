package users

import (
	"net/http"
	"strings"

	"github.com/mddanish004/jamzy/server/internal/apperr"
	"github.com/mddanish004/jamzy/server/internal/auth"
	"github.com/mddanish004/jamzy/server/internal/db"
	"github.com/mddanish004/jamzy/server/internal/httpx"
	"github.com/mddanish004/jamzy/server/internal/logging"
)

type Handlers struct {
	Service *Service
	Log     logging.Logger
}

func (h *Handlers) GetMe(w http.ResponseWriter, r *http.Request) error {
	clerkUserID, err := auth.ClerkUserID(r.Context())
	if err != nil {
		httpx.WriteUnauthorized(w)
		return nil
	}

	user, err := h.Service.FindByClerkID(r.Context(), clerkUserID)
	if err != nil {
		return err
	}

	httpx.WriteJSON(w, 200, MeResponse{
		Authenticated:      true,
		User:               user,
		OnboardingComplete: user != nil,
	})
	return nil
}

func (h *Handlers) CreateProfile(w http.ResponseWriter, r *http.Request) error {
	clerkUserID, err := auth.ClerkUserID(r.Context())
	if err != nil {
		httpx.WriteUnauthorized(w)
		return nil
	}

	var body CreateProfileBody
	if err := httpx.ReadJSON(w, r, &body); err != nil {
		return err
	}

	username, err := ValidateUsername(body.Username)
	if err != nil {
		return err
	}

	existing, err := h.Service.FindByClerkID(r.Context(), clerkUserID)
	if err != nil {
		return err
	}
	if existing != nil {
		httpx.WriteJSON(w, 200, ProfileResponse{User: *existing})
		return nil
	}

	taken, err := h.Service.FindByUsername(r.Context(), username)
	if err != nil {
		return err
	}
	if taken != nil {
		return apperr.New("Username is already taken", 409)
	}

	user, err := h.Service.Create(r.Context(), clerkUserID, username)
	if err != nil {
		return h.recoverCreate(w, r, err, clerkUserID)
	}

	httpx.WriteJSON(w, 201, ProfileResponse{User: *user})
	return nil
}

func (h *Handlers) recoverCreate(w http.ResponseWriter, r *http.Request, err error, clerkUserID string) error {
	constraint, ok := db.UniqueViolation(err)
	if !ok {
		return err
	}

	if strings.Contains(constraint, "username") {
		return apperr.New("Username is already taken", 409)
	}

	if strings.Contains(constraint, "clerk_user_id") {
		existing, findErr := h.Service.FindByClerkID(r.Context(), clerkUserID)
		if findErr != nil {
			return findErr
		}
		if existing != nil {
			httpx.WriteJSON(w, 200, ProfileResponse{User: *existing})
			return nil
		}
	}

	return err
}
