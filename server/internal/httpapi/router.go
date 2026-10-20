package httpapi

import (
	"net/http"

	"github.com/mddanish004/jamzy/server/internal/config"
	"github.com/mddanish004/jamzy/server/internal/httpx"
	"github.com/mddanish004/jamzy/server/internal/logging"
	"github.com/mddanish004/jamzy/server/internal/music"
	"github.com/mddanish004/jamzy/server/internal/openapi"
	"github.com/mddanish004/jamzy/server/internal/rooms"
	"github.com/mddanish004/jamzy/server/internal/users"
)

const socketPath = "/socket.io/"

type Deps struct {
	Middleware *httpx.Middleware
	Users      *users.Handlers
	Rooms      *rooms.Handlers
	Music      *music.Handlers
	Socket     http.Handler
	Log        logging.Logger
}

func New(deps Deps) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, 200, map[string]any{
			"success": true,
			"message": "API is running",
		})
	})

	mux.HandleFunc("GET /docs", openapi.Docs)
	mux.HandleFunc("GET /docs/{$}", openapi.Docs)
	mux.HandleFunc("GET /docs/openapi.json", openapi.Spec)

	mux.Handle("GET /api/me", deps.Middleware.Route("me", config.REST.Me, deps.Users.GetMe))
	mux.Handle("POST /api/users/profile", deps.Middleware.Route("createProfile", config.REST.CreateProfile, deps.Users.CreateProfile))

	mux.Handle("POST /api/rooms", deps.Middleware.Route("createRoom", config.REST.CreateRoom, deps.Rooms.CreateRoom))
	mux.Handle("POST /api/rooms/{roomCode}/join", deps.Middleware.Route("joinRoom", config.REST.JoinRoom, deps.Rooms.JoinRoom))
	mux.Handle("GET /api/rooms/{roomCode}", deps.Middleware.Route("getRoom", config.REST.GetRoom, deps.Rooms.GetRoom))
	mux.Handle("POST /api/rooms/{roomCode}/leave", deps.Middleware.Route("leaveRoom", config.REST.LeaveRoom, deps.Rooms.LeaveRoom))

	mux.Handle("GET /api/music/search", deps.Middleware.Route("musicSearch", config.REST.MusicSearch, deps.Music.Search))
	mux.Handle("GET /api/music/video", deps.Middleware.Route("musicVideo", config.REST.MusicVideo, deps.Music.ResolveVideo))
	mux.Handle("GET /api/music/playlist", deps.Middleware.Route("musicPlaylist", config.REST.MusicPlaylist, deps.Music.ResolvePlaylist))

	if deps.Socket != nil {
		mux.Handle(socketPath, deps.Socket)
	}

	mux.HandleFunc("/", deps.Middleware.NotFound)

	return deps.Middleware.Recover(
		deps.Middleware.RequestLog(
			deps.Middleware.SecurityHeaders(
				deps.Middleware.CORS(mux),
			),
		),
	)
}
