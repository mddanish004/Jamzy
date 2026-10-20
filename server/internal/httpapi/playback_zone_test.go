package httpapi_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

const zoneTestTz = "Asia/Kolkata"

func (h *harness) pinDatabaseTimezone(t *testing.T) {
	t.Helper()

	ctx := context.Background()

	database := ""
	if err := h.pool.QueryRow(ctx, "SELECT current_database()").Scan(&database); err != nil {
		t.Fatalf("select current_database(): %v", err)
	}

	previous := ""
	if err := h.pool.QueryRow(ctx, "SHOW timezone").Scan(&previous); err != nil {
		t.Fatalf("show timezone: %v", err)
	}

	if _, err := h.pool.Exec(ctx,
		fmt.Sprintf("ALTER DATABASE %q SET timezone TO %q", database, zoneTestTz),
	); err != nil {
		t.Skipf("cannot pin a non-UTC database timezone: %v", err)
	}
	h.pool.Reset()

	t.Cleanup(func() {
		_, _ = h.pool.Exec(ctx, fmt.Sprintf("ALTER DATABASE %q SET timezone TO %q", database, previous))
	})
}

func derivedPosition(snapshot map[string]any) float64 {
	position, _ := snapshot["position"].(float64)
	playing, _ := snapshot["isPlaying"].(bool)
	if !playing {
		return position
	}

	updatedAt, _ := snapshot["updatedAt"].(string)
	serverTime, _ := snapshot["serverTime"].(string)
	if updatedAt == "" || serverTime == "" {
		return position
	}

	updated, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return position
	}
	stamped, err := time.Parse(time.RFC3339Nano, serverTime)
	if err != nil {
		return position
	}

	elapsed := stamped.Sub(updated).Seconds()
	if elapsed < 0 {
		return position
	}
	return position + elapsed
}

func TestGuestRejoinPlaybackSnapshotIsTimezoneStable(t *testing.T) {
	h := newHarness(t)
	h.pinDatabaseTimezone(t)

	h.expectStatus(http.MethodPost, "/api/users/profile", "alice", map[string]any{"username": "alice"}, 201)
	h.expectStatus(http.MethodPost, "/api/users/profile", "bob", map[string]any{"username": "bob"}, 201)

	created := h.expectStatus(http.MethodPost, "/api/rooms", "alice", nil, 201)
	roomCode := created["room"].(map[string]any)["roomCode"].(string)
	h.expectStatus(http.MethodPost, "/api/rooms/"+roomCode+"/join", "bob", nil, 200)

	host := dialSocket(t, h.server.URL, "alice")
	if ack := host.emitAck("room:join", map[string]any{"roomCode": roomCode}); ack["ok"] != true {
		t.Fatalf("host room:join answered %v", ack)
	}

	setAck := host.emitAck("playback:set", map[string]any{
		"roomCode": roomCode,
		"track":    map[string]any{"videoId": videoID, "title": "Song", "thumbnailUrl": "", "duration": "PT3M"},
	})
	if setAck["ok"] != true {
		t.Fatalf("playback:set answered %v", setAck)
	}

	time.Sleep(3 * time.Second)

	rejoined := dialSocket(t, h.server.URL, "bob")
	join := rejoined.emitAck("room:join", map[string]any{"roomCode": roomCode})
	snapshot, ok := join["playback"].(map[string]any)
	if !ok {
		t.Fatalf("join ack carried no playback: %v", join)
	}

	derived := derivedPosition(snapshot)
	if derived < 2.4 || derived > 5 {
		t.Fatalf("rejoined guest derives playhead %.2fs from snapshot %v, want ~3s", derived, snapshot)
	}

	if ack := host.emitAck("playback:control", map[string]any{
		"roomCode": roomCode,
		"action":   "pause",
		"position": 30,
	}); ack["ok"] != true {
		t.Fatalf("playback:control answered %v", ack)
	}
	time.Sleep(300 * time.Millisecond)

	pausedGuest := dialSocket(t, h.server.URL, "bob")
	pausedJoin := pausedGuest.emitAck("room:join", map[string]any{"roomCode": roomCode})
	pausedSnapshot, ok := pausedJoin["playback"].(map[string]any)
	if !ok {
		t.Fatalf("join ack carried no playback: %v", pausedJoin)
	}
	if pausedDerived := derivedPosition(pausedSnapshot); pausedDerived < 29.5 || pausedDerived > 30.5 {
		t.Fatalf("rejoined guest derives paused playhead %.2fs, want ~30s", pausedDerived)
	}
}
