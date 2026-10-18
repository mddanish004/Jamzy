package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mddanish004/jamzy/server/internal/apperr"
	"github.com/mddanish004/jamzy/server/internal/db"
	"github.com/mddanish004/jamzy/server/internal/httpapi"
	"github.com/mddanish004/jamzy/server/internal/httpx"
	"github.com/mddanish004/jamzy/server/internal/logging"
	"github.com/mddanish004/jamzy/server/internal/messages"
	"github.com/mddanish004/jamzy/server/internal/music"
	"github.com/mddanish004/jamzy/server/internal/playback"
	"github.com/mddanish004/jamzy/server/internal/ratelimit"
	"github.com/mddanish004/jamzy/server/internal/realtime"
	"github.com/mddanish004/jamzy/server/internal/rooms"
	"github.com/mddanish004/jamzy/server/internal/users"
	"github.com/zishang520/engine.io/v2/types"
	"github.com/zishang520/socket.io/v2/socket"
)

const (
	videoID = "dQw4w9WgXcQ"
	ackWait = 3 * time.Second
	quiet   = 150 * time.Millisecond
)

type stubVerifier struct{}

func (stubVerifier) Verify(_ context.Context, token string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", apperr.New("Unauthorized", 401)
	}
	return "clerk-" + token, nil
}

type harness struct {
	t      *testing.T
	server *httptest.Server
}

func newHarness(t *testing.T) *harness {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()

	pool, err := db.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to %s: %v", databaseURL, err)
	}
	t.Cleanup(pool.Close)

	if _, err := pool.Exec(ctx, "TRUNCATE messages, room_members, rooms, users CASCADE"); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	log := logging.New()
	limiter := ratelimit.NewLimiter()
	t.Cleanup(limiter.Close)

	usersService := users.NewService(pool)
	roomsService := rooms.NewService(pool, usersService)

	hub := realtime.NewHub(
		roomsService,
		playback.NewService(pool),
		messages.NewService(pool),
		usersService,
		stubVerifier{},
		limiter,
		log,
	)

	options := socket.DefaultServerOptions()
	options.SetCors(&types.Cors{Origin: []any{"http://localhost:5173"}, Credentials: true})
	options.SetMaxHttpBufferSize(100_000)

	io := hub.Server(options)

	router := httpapi.New(httpapi.Deps{
		Middleware: &httpx.Middleware{
			Log:         log,
			Verifier:    stubVerifier{},
			Limiter:     limiter,
			CORSOrigins: []string{"http://localhost:5173"},
		},
		Users:  &users.Handlers{Service: usersService, Log: log},
		Rooms:  &rooms.Handlers{Service: roomsService, Hooks: hub, Log: log},
		Music:  &music.Handlers{Service: music.NewService("", log)},
		Socket: io.ServeHandler(nil),
		Log:    log,
	})

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	return &harness{t: t, server: server}
}

func (h *harness) request(method string, path string, token string, body any) (int, map[string]any) {
	h.t.Helper()

	var reader *strings.Reader

	if body == nil {
		reader = strings.NewReader("")
	} else {
		encoded, err := json.Marshal(body)
		if err != nil {
			h.t.Fatalf("encode body: %v", err)
		}
		reader = strings.NewReader(string(encoded))
	}

	request, err := http.NewRequest(method, h.server.URL+path, reader)
	if err != nil {
		h.t.Fatalf("build request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	response, err := h.server.Client().Do(request)
	if err != nil {
		h.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer response.Body.Close()

	payload := map[string]any{}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		h.t.Fatalf("%s %s: decode body: %v", method, path, err)
	}

	return response.StatusCode, payload
}

func (h *harness) expectStatus(method string, path string, token string, body any, want int) map[string]any {
	h.t.Helper()

	status, payload := h.request(method, path, token, body)

	if status != want {
		h.t.Fatalf("%s %s answered %d (%v), want %d", method, path, status, payload, want)
	}

	return payload
}

func TestHealthAndAuthenticationBoundary(t *testing.T) {
	h := newHarness(t)

	h.expectStatus(http.MethodGet, "/api/health", "", nil, 200)

	protected := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/me"},
		{http.MethodPost, "/api/users/profile"},
		{http.MethodPost, "/api/rooms"},
		{http.MethodPost, "/api/rooms/ABC234/join"},
		{http.MethodGet, "/api/rooms/ABC234"},
		{http.MethodPost, "/api/rooms/ABC234/leave"},
		{http.MethodGet, "/api/music/search?q=test"},
		{http.MethodGet, "/api/music/video?url=" + videoID},
		{http.MethodGet, "/api/music/playlist?url=PL1234567890abcdef"},
	}

	for _, route := range protected {
		payload := h.expectStatus(route.method, route.path, "", nil, 401)
		if payload["message"] != "Unauthorized" {
			t.Fatalf("%s %s answered %v", route.method, route.path, payload)
		}
	}

	h.expectStatus(http.MethodGet, "/api/does-not-exist", "", nil, 404)
}

func TestProfileRoomAndSocketFlow(t *testing.T) {
	h := newHarness(t)

	alice := h.expectStatus(http.MethodPost, "/api/users/profile", "alice", map[string]any{"username": "alice"}, 201)
	aliceUser := alice["user"].(map[string]any)
	aliceID := aliceUser["id"].(string)

	if aliceUser["clerkUserId"] != "clerk-alice" || aliceUser["username"] != "alice" {
		t.Fatalf("profile = %v", aliceUser)
	}

	repeated := h.expectStatus(http.MethodPost, "/api/users/profile", "alice", map[string]any{"username": "alice"}, 200)
	if repeated["user"].(map[string]any)["id"] != aliceID {
		t.Fatal("creating the same profile twice produced a different user")
	}

	h.expectStatus(http.MethodPost, "/api/users/profile", "alice", map[string]any{"username": "other"}, 200)
	h.expectStatus(http.MethodPost, "/api/users/profile", "bob", map[string]any{"username": "alice"}, 409)
	h.expectStatus(http.MethodPost, "/api/users/profile", "bob", map[string]any{}, 400)
	h.expectStatus(http.MethodPost, "/api/users/profile", "bob", map[string]any{"username": "ab"}, 400)

	me := h.expectStatus(http.MethodGet, "/api/me", "alice", nil, 200)
	if me["authenticated"] != true || me["onboardingComplete"] != true {
		t.Fatalf("me = %v", me)
	}

	bob := h.expectStatus(http.MethodPost, "/api/users/profile", "bob", map[string]any{"username": "bob"}, 201)
	bobID := bob["user"].(map[string]any)["id"].(string)

	carol := h.expectStatus(http.MethodPost, "/api/users/profile", "carol", map[string]any{"username": "carol"}, 201)
	carolID := carol["user"].(map[string]any)["id"].(string)

	created := h.expectStatus(http.MethodPost, "/api/rooms", "alice", nil, 201)
	room := created["room"].(map[string]any)
	roomCode := room["roomCode"].(string)

	if len(roomCode) != rooms.RoomCodeLength {
		t.Fatalf("room code = %q", roomCode)
	}
	if room["host"].(map[string]any)["id"] != aliceID {
		t.Fatalf("host = %v", room["host"])
	}
	if len(room["members"].([]any)) != 1 {
		t.Fatalf("members = %v", room["members"])
	}
	if room["playbackPosition"] != float64(0) || room["isPlaying"] != false {
		t.Fatalf("playback defaults = %v", room)
	}

	h.expectStatus(http.MethodGet, "/api/rooms/ZZZZZZ", "alice", nil, 404)
	h.expectStatus(http.MethodGet, "/api/rooms/ABC", "alice", nil, 400)
	h.expectStatus(http.MethodPost, "/api/rooms/ZZZZZZ/join", "bob", nil, 404)

	joined := h.expectStatus(http.MethodPost, "/api/rooms/"+roomCode+"/join", "bob", nil, 200)
	if len(joined["room"].(map[string]any)["members"].([]any)) != 2 {
		t.Fatalf("join produced %v", joined)
	}

	fetched := h.expectStatus(http.MethodGet, "/api/rooms/"+roomCode, "alice", nil, 200)
	members := fetched["room"].(map[string]any)["members"].([]any)
	if len(members) != 2 {
		t.Fatalf("members after join = %v", members)
	}
	if members[0].(map[string]any)["user"].(map[string]any)["id"] != aliceID {
		t.Fatalf("members are not ordered oldest-first: %v", members)
	}

	aliceSocket := dialSocket(t, h.server.URL, "alice")
	defer aliceSocket.close()

	bobSocket := dialSocket(t, h.server.URL, "bob")
	defer bobSocket.close()

	joinAck := aliceSocket.emitAck("room:join", map[string]any{"roomCode": roomCode})
	if joinAck["ok"] != true {
		t.Fatalf("join ack = %v", joinAck)
	}

	presence := joinAck["presence"].(map[string]any)
	if presence["hostUserId"] != aliceID {
		t.Fatalf("presence host = %v", presence)
	}
	if len(presence["members"].([]any)) != 2 {
		t.Fatalf("presence = %v", presence)
	}
	if len(joinAck["messages"].([]any)) != 0 || len(joinAck["queue"].([]any)) != 0 {
		t.Fatalf("join ack carried history: %v", joinAck)
	}

	bobSocket.emitAck("room:join", map[string]any{"roomCode": roomCode})

	bobSocket.waitFor("presence:update")

	if ack := bobSocket.emitAck("room:join", map[string]any{"roomCode": "ZZZZZZ"}); ack["ok"] != false {
		t.Fatalf("joining a missing room answered %v", ack)
	}
	if ack := bobSocket.emitAck("room:join", map[string]any{}); ack["message"] != "Room code is required" {
		t.Fatalf("joining without a code answered %v", ack)
	}

	if response := bobSocket.emitAck("playback:set", map[string]any{
		"roomCode": roomCode,
		"track":    map[string]any{"videoId": videoID, "title": "Song", "thumbnailUrl": "", "duration": "PT3M"},
	}); response["message"] != "Only the host can control playback" {
		t.Fatalf("a non-host set the track: %v", response)
	}

	setAck := aliceSocket.emitAck("playback:set", map[string]any{
		"roomCode": roomCode,
		"track":    map[string]any{"videoId": videoID, "title": "Song", "thumbnailUrl": "", "duration": "PT3M"},
	})
	if setAck["ok"] != true {
		t.Fatalf("playback:set answered %v", setAck)
	}

	snapshot := setAck["playback"].(map[string]any)
	if snapshot["videoId"] != videoID || snapshot["isPlaying"] != true {
		t.Fatalf("snapshot = %v", snapshot)
	}
	if snapshot["duration"] != float64(180) {
		t.Fatalf("duration = %v", snapshot["duration"])
	}
	if snapshot["serverTime"] == nil || snapshot["updatedAt"] == nil {
		t.Fatalf("snapshot is missing timestamps: %v", snapshot)
	}

	broadcast := bobSocket.waitFor("playback:update")
	if broadcast["videoId"] != videoID {
		t.Fatalf("broadcast = %v", broadcast)
	}

	if response := aliceSocket.emitAck("playback:set", map[string]any{
		"roomCode": roomCode,
		"track":    map[string]any{"videoId": "not-an-id", "title": "Song", "thumbnailUrl": "", "duration": "PT3M"},
	}); response["message"] != "Invalid video id" {
		t.Fatalf("a hostile track answered %v", response)
	}

	if response := aliceSocket.emitAck("playback:control", map[string]any{
		"roomCode": roomCode,
		"action":   "seek",
		"position": 42,
	}); response["ok"] != true {
		t.Fatalf("playback:control answered %v", response)
	}

	paused := aliceSocket.emitAck("playback:control", map[string]any{
		"roomCode": roomCode,
		"action":   "pause",
		"position": 42,
	})
	if paused["playback"].(map[string]any)["isPlaying"] != false {
		t.Fatalf("pause answered %v", paused)
	}

	if response := aliceSocket.emitAck("playback:control", map[string]any{
		"roomCode": roomCode,
		"action":   "stop",
		"position": 0,
	}); response["message"] != "Unknown playback action" {
		t.Fatalf("an unknown action answered %v", response)
	}

	aliceSocket.emitNoAck("queue:add", map[string]any{
		"roomCode": roomCode,
		"track":    map[string]any{"videoId": "aaaaaaaaaaa", "title": "Queued", "thumbnailUrl": "", "duration": "PT1M"},
	})

	queue := aliceSocket.waitForAny("queue:update")
	if len(queue.([]any)) != 1 {
		t.Fatalf("queue = %v", queue)
	}
	bobQueue := bobSocket.waitForAny("queue:update")
	if len(bobQueue.([]any)) != 1 {
		t.Fatalf("queue broadcast = %v", bobQueue)
	}

	advance := aliceSocket.emitAck("queue:advance", map[string]any{"roomCode": roomCode})
	if len(advance["queue"].([]any)) != 0 {
		t.Fatalf("advance answered %v", advance)
	}

	advanced := bobSocket.waitForMatching("playback:update", func(payload map[string]any) bool {
		return payload["videoId"] == "aaaaaaaaaaa"
	})
	if advanced["videoId"] != "aaaaaaaaaaa" {
		t.Fatalf("advanced playback = %v", advanced)
	}

	if chat := aliceSocket.emitAck("chat:send", map[string]any{}); chat["message"] != "Room code is required" {
		t.Fatalf("a chat without a room answered %v", chat)
	}

	if chat := aliceSocket.emitAck("chat:send", map[string]any{"roomCode": roomCode}); chat["message"] != "Message is required" {
		t.Fatalf("a blank chat answered %v", chat)
	}

	if chat := bobSocket.emitAck("chat:send", map[string]any{"roomCode": roomCode, "content": "hello"}); chat["ok"] != true {
		t.Fatalf("chat:send answered %v", chat)
	}

	message := aliceSocket.waitFor("chat:new")
	if message["content"] != "hello" || message["username"] != "bob" {
		t.Fatalf("chat:new = %v", message)
	}
	if message["createdAt"] == nil {
		t.Fatal("chat:new has no timestamp")
	}

	if response := aliceSocket.emitAck("host:transfer", map[string]any{"roomCode": roomCode, "userId": aliceID}); response["message"] != "You are already the host" {
		t.Fatalf("self transfer answered %v", response)
	}
	if response := aliceSocket.emitAck("host:transfer", map[string]any{"roomCode": roomCode, "userId": "nobody"}); response["message"] != "That person is not a member of this room" {
		t.Fatalf("unknown transfer target answered %v", response)
	}
	if response := bobSocket.emitAck("host:transfer", map[string]any{"roomCode": roomCode, "userId": bobID}); response["message"] != "Only the host can hand over the room" {
		t.Fatalf("non-host transfer answered %v", response)
	}

	if response := aliceSocket.emitAck("host:transfer", map[string]any{"roomCode": roomCode, "userId": bobID}); response["ok"] != true {
		t.Fatalf("host:transfer answered %v", response)
	}

	hostUpdate := bobSocket.waitFor("host:update")
	if hostUpdate["hostUserId"] != bobID || hostUpdate["reason"] != "manual" {
		t.Fatalf("host:update = %v", hostUpdate)
	}
	if hostUpdate["previousHostUserId"] != aliceID || hostUpdate["hostUsername"] != "bob" {
		t.Fatalf("host:update = %v", hostUpdate)
	}

	carolSocket := dialSocket(t, h.server.URL, "carol")
	defer carolSocket.close()

	h.expectStatus(http.MethodPost, "/api/rooms/"+roomCode+"/join", "carol", nil, 200)
	carolSocket.emitAck("room:join", map[string]any{"roomCode": roomCode})

	left := h.expectStatus(http.MethodPost, "/api/rooms/"+roomCode+"/leave", "bob", nil, 200)
	if left["left"] != true || left["roomDeleted"] != false {
		t.Fatalf("leave answered %v", left)
	}

	handover := carolSocket.waitFor("host:update")
	if handover["hostUserId"] != aliceID || handover["reason"] != "left" {
		t.Fatalf("handover = %v", handover)
	}
	if handover["previousHostUserId"] != bobID {
		t.Fatalf("handover = %v", handover)
	}
	if handover["hostUserId"] == carolID {
		t.Fatalf("the newest member inherited the room: %v", handover)
	}

	if ack := bobSocket.emitAck("chat:send", map[string]any{"roomCode": roomCode, "content": "still here"}); ack["message"] != "Not a member of this room" {
		t.Fatalf("a departed member could still chat: %v", ack)
	}

	aliceSocket.emitAck("chat:send", map[string]any{"roomCode": roomCode, "content": "after bob left"})
	carolSocket.waitFor("chat:new")

	carolSocket.close()

	bobSocket.close()
	aliceSocket.close()
}

func TestLeavingThroughRestStopsRoomEvents(t *testing.T) {
	h := newHarness(t)

	h.expectStatus(http.MethodPost, "/api/users/profile", "dave", map[string]any{"username": "dave"}, 201)
	h.expectStatus(http.MethodPost, "/api/users/profile", "erin", map[string]any{"username": "erin"}, 201)

	created := h.expectStatus(http.MethodPost, "/api/rooms", "dave", nil, 201)
	roomCode := created["room"].(map[string]any)["roomCode"].(string)
	h.expectStatus(http.MethodPost, "/api/rooms/"+roomCode+"/join", "erin", nil, 200)

	daveSocket := dialSocket(t, h.server.URL, "dave")
	defer daveSocket.close()

	erinSocket := dialSocket(t, h.server.URL, "erin")
	defer erinSocket.close()

	daveSocket.emitAck("room:join", map[string]any{"roomCode": roomCode})
	erinSocket.emitAck("room:join", map[string]any{"roomCode": roomCode})

	h.expectStatus(http.MethodPost, "/api/rooms/"+roomCode+"/leave", "dave", nil, 200)

	time.Sleep(2 * quiet)

	erinSocket.emitAck("chat:send", map[string]any{"roomCode": roomCode, "content": "anyone there"})

	daveSocket.expectNone("chat:new", quiet)

	if ack := daveSocket.emitAck("room:join", map[string]any{"roomCode": roomCode}); ack["message"] != "Not a member of this room" {
		t.Fatalf("a departed member rejoined their cached room: %v", ack)
	}

	erinSocket.waitFor("chat:new")
}

func TestDisconnectingHostStartsTheGraceCountdown(t *testing.T) {
	h := newHarness(t)

	h.expectStatus(http.MethodPost, "/api/users/profile", "frank", map[string]any{"username": "frank"}, 201)
	h.expectStatus(http.MethodPost, "/api/users/profile", "grace", map[string]any{"username": "grace"}, 201)

	created := h.expectStatus(http.MethodPost, "/api/rooms", "frank", nil, 201)
	roomCode := created["room"].(map[string]any)["roomCode"].(string)
	h.expectStatus(http.MethodPost, "/api/rooms/"+roomCode+"/join", "grace", nil, 200)

	frankSocket := dialSocket(t, h.server.URL, "frank")
	graceSocket := dialSocket(t, h.server.URL, "grace")
	defer graceSocket.close()

	frankSocket.emitAck("room:join", map[string]any{"roomCode": roomCode})
	graceSocket.emitAck("room:join", map[string]any{"roomCode": roomCode})

	frankSocket.close()

	pending := graceSocket.waitFor("host:pending")
	if pending["pending"] != true {
		t.Fatalf("host:pending = %v", pending)
	}
	if pending["hostUsername"] != "frank" || pending["successorUsername"] != "grace" {
		t.Fatalf("host:pending = %v", pending)
	}
	if pending["deadline"] == nil {
		t.Fatal("host:pending has no deadline")
	}

	returnedSocket := dialSocket(t, h.server.URL, "frank")
	defer returnedSocket.close()

	returnedSocket.emitAck("room:join", map[string]any{"roomCode": roomCode})

	cancelled := graceSocket.waitFor("host:pending")
	if cancelled["pending"] != false {
		t.Fatalf("the countdown was not cancelled: %v", cancelled)
	}
}

func TestSoloHostLeavingDeletesTheRoom(t *testing.T) {
	h := newHarness(t)

	h.expectStatus(http.MethodPost, "/api/users/profile", "henry", map[string]any{"username": "henry"}, 201)

	created := h.expectStatus(http.MethodPost, "/api/rooms", "henry", nil, 201)
	roomCode := created["room"].(map[string]any)["roomCode"].(string)

	left := h.expectStatus(http.MethodPost, "/api/rooms/"+roomCode+"/leave", "henry", nil, 200)
	if left["roomDeleted"] != true {
		t.Fatalf("leave answered %v", left)
	}

	h.expectStatus(http.MethodGet, "/api/rooms/"+roomCode, "henry", nil, 404)
	h.expectStatus(http.MethodPost, "/api/rooms/"+roomCode+"/leave", "henry", nil, 404)
}

func TestSocketHandshakeRejectsUnknownUsers(t *testing.T) {
	h := newHarness(t)

	wsURL := "ws" + strings.TrimPrefix(h.server.URL, "http") + "/socket.io/?EIO=4&transport=websocket"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	if _, _, err := conn.ReadMessage(); err != nil {
		t.Fatalf("read open packet: %v", err)
	}

	if err := conn.WriteMessage(websocket.TextMessage, []byte(`40{"token":"nosuchuser"}`)); err != nil {
		t.Fatalf("write connect packet: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(ackWait))

	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read connect error: %v", err)
	}

	if !strings.HasPrefix(string(data), "44") {
		t.Fatalf("expected a connect_error packet, got %q", data)
	}
	if !strings.Contains(string(data), "User profile not found") {
		t.Fatalf("connect_error = %q", data)
	}
}

type socketClient struct {
	t      *testing.T
	conn   *websocket.Conn
	mu     sync.Mutex
	write  sync.Mutex
	nextID uint64
	acks   map[uint64]chan json.RawMessage
	events map[string]chan json.RawMessage
	done   chan struct{}
}

func dialSocket(t *testing.T, serverURL string, token string) *socketClient {
	t.Helper()

	wsURL := "ws" + strings.TrimPrefix(serverURL, "http") + "/socket.io/?EIO=4&transport=websocket"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial %s: %v", wsURL, err)
	}

	client := &socketClient{
		t:      t,
		conn:   conn,
		acks:   map[uint64]chan json.RawMessage{},
		events: map[string]chan json.RawMessage{},
		done:   make(chan struct{}),
	}

	open := client.read()
	if !strings.HasPrefix(open, "0") {
		t.Fatalf("expected an engine.io open packet, got %q", open)
	}

	client.writeRaw("40" + `{"token":"` + token + `"}`)

	connected := client.read()
	if !strings.HasPrefix(connected, "40") {
		t.Fatalf("expected a socket.io connect packet, got %q", connected)
	}

	go client.readLoop()

	return client
}

func (c *socketClient) read() string {
	c.t.Helper()

	c.conn.SetReadDeadline(time.Now().Add(ackWait))

	_, data, err := c.conn.ReadMessage()
	if err != nil {
		c.t.Fatalf("read: %v", err)
	}

	return string(data)
}

func (c *socketClient) writeRaw(packet string) {
	c.write.Lock()
	defer c.write.Unlock()

	if err := c.conn.WriteMessage(websocket.TextMessage, []byte(packet)); err != nil {
		return
	}
}

func (c *socketClient) readLoop() {
	for {
		c.conn.SetReadDeadline(time.Time{})

		_, data, err := c.conn.ReadMessage()
		if err != nil {
			close(c.done)
			return
		}

		c.dispatch(string(data))
	}
}

func (c *socketClient) dispatch(payload string) {
	for _, packet := range strings.Split(payload, "\x1e") {
		c.dispatchPacket(packet)
	}
}

func (c *socketClient) dispatchPacket(packet string) {
	switch {
	case packet == "2":
		c.writeRaw("3")
	case strings.HasPrefix(packet, "42"):
		c.dispatchEvent(packet[2:])
	case strings.HasPrefix(packet, "43"):
		c.dispatchAck(packet[2:])
	case strings.HasPrefix(packet, "41"), strings.HasPrefix(packet, "44"):
	default:
	}
}

func (c *socketClient) dispatchEvent(rest string) {
	var args []json.RawMessage
	if err := json.Unmarshal([]byte(rest), &args); err != nil || len(args) < 2 {
		return
	}

	var name string
	if err := json.Unmarshal(args[0], &name); err != nil {
		return
	}

	c.mu.Lock()
	channel, ok := c.events[name]
	if !ok {
		channel = make(chan json.RawMessage, 32)
		c.events[name] = channel
	}
	c.mu.Unlock()

	select {
	case channel <- args[1]:
	default:
	}
}

func (c *socketClient) dispatchAck(rest string) {
	bracket := strings.Index(rest, "[")
	if bracket < 0 {
		return
	}

	id, err := strconv.ParseUint(rest[:bracket], 10, 64)
	if err != nil {
		return
	}

	var args []json.RawMessage
	if err := json.Unmarshal([]byte(rest[bracket:]), &args); err != nil || len(args) == 0 {
		return
	}

	c.mu.Lock()
	channel, ok := c.acks[id]
	if ok {
		delete(c.acks, id)
	}
	c.mu.Unlock()

	if ok {
		channel <- args[0]
	}
}

func (c *socketClient) eventChannel(name string) chan json.RawMessage {
	c.mu.Lock()
	defer c.mu.Unlock()

	channel, ok := c.events[name]
	if !ok {
		channel = make(chan json.RawMessage, 32)
		c.events[name] = channel
	}

	return channel
}

func (c *socketClient) emitNoAck(event string, payload any) {
	c.t.Helper()

	encoded, err := json.Marshal([]any{event, payload})
	if err != nil {
		c.t.Fatalf("encode %s: %v", event, err)
	}

	c.writeRaw("42" + string(encoded))
}

func (c *socketClient) emitAck(event string, payload any) map[string]any {
	c.t.Helper()

	c.mu.Lock()
	c.nextID++
	id := c.nextID
	channel := make(chan json.RawMessage, 1)
	c.acks[id] = channel
	c.mu.Unlock()

	encoded, err := json.Marshal([]any{event, payload})
	if err != nil {
		c.t.Fatalf("encode %s: %v", event, err)
	}

	c.writeRaw("42" + strconv.FormatUint(id, 10) + string(encoded))

	select {
	case raw := <-channel:
		response := map[string]any{}
		if err := json.Unmarshal(raw, &response); err != nil {
			c.t.Fatalf("ack for %s is not an object: %s", event, raw)
		}
		return response
	case <-time.After(ackWait):
		c.t.Fatalf("ack for %s timed out", event)
		return nil
	}
}

func (c *socketClient) waitFor(event string) map[string]any {
	c.t.Helper()

	channel := c.eventChannel(event)

	select {
	case raw := <-channel:
		payload := map[string]any{}
		if err := json.Unmarshal(raw, &payload); err != nil {
			c.t.Fatalf("%s payload is not an object: %s", event, raw)
		}
		return payload
	case <-time.After(ackWait):
		c.t.Fatalf("no %s event arrived", event)
		return nil
	}
}

func (c *socketClient) waitForAny(event string) any {
	c.t.Helper()

	channel := c.eventChannel(event)

	select {
	case raw := <-channel:
		var payload any
		if err := json.Unmarshal(raw, &payload); err != nil {
			c.t.Fatalf("%s payload is not valid JSON: %s", event, raw)
		}
		return payload
	case <-time.After(ackWait):
		c.t.Fatalf("no %s event arrived", event)
		return nil
	}
}

func (c *socketClient) waitForMatching(event string, match func(map[string]any) bool) map[string]any {
	c.t.Helper()

	deadline := time.Now().Add(ackWait)

	for {
		channel := c.eventChannel(event)

		remaining := time.Until(deadline)
		if remaining <= 0 {
			c.t.Fatalf("no matching %s event arrived", event)
			return nil
		}

		select {
		case raw := <-channel:
			payload := map[string]any{}
			if err := json.Unmarshal(raw, &payload); err != nil {
				continue
			}
			if match(payload) {
				return payload
			}
		case <-time.After(remaining):
			c.t.Fatalf("no matching %s event arrived", event)
			return nil
		}
	}
}

func (c *socketClient) expectNone(event string, wait time.Duration) {
	c.t.Helper()

	channel := c.eventChannel(event)

	select {
	case raw := <-channel:
		c.t.Fatalf("unexpected %s event: %s", event, raw)
	case <-time.After(wait):
	}
}

func (c *socketClient) close() {
	_ = c.conn.Close()
	<-c.done
}
