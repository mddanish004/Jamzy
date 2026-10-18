package ratelimit

import (
	"testing"
	"time"

	"github.com/mddanish004/jamzy/server/internal/config"
)

func rule(limit int, window time.Duration) config.RateLimitRule {
	return config.RateLimitRule{Limit: limit, Window: window}
}

func TestAllowsExactlyLimitHitsPerWindow(t *testing.T) {
	limiter := NewLimiter()
	defer limiter.Close()

	limit := rule(3, time.Minute)

	for i := 0; i < 3; i++ {
		if result := limiter.Consume("bucket:user:a", limit); !result.Allowed {
			t.Fatalf("hit %d was blocked", i+1)
		}
	}

	blocked := limiter.Consume("bucket:user:a", limit)
	if blocked.Allowed {
		t.Fatal("the fourth hit was allowed")
	}
	if blocked.RetryAfterMs <= 0 {
		t.Fatalf("RetryAfterMs = %d", blocked.RetryAfterMs)
	}
}

func TestCountsEachCallerSeparately(t *testing.T) {
	limiter := NewLimiter()
	defer limiter.Close()

	limit := rule(1, time.Minute)

	if result := limiter.Consume("bucket:user:a", limit); !result.Allowed {
		t.Fatal("first hit for caller a was blocked")
	}
	if result := limiter.Consume("bucket:user:a", limit); result.Allowed {
		t.Fatal("second hit for caller a was allowed")
	}
	if result := limiter.Consume("bucket:user:b", limit); !result.Allowed {
		t.Fatal("caller b was locked out by caller a")
	}
}

func TestCountsEachBucketSeparately(t *testing.T) {
	limiter := NewLimiter()
	defer limiter.Close()

	limit := rule(1, time.Minute)

	if result := limiter.Consume("search:user:a", limit); !result.Allowed {
		t.Fatal("search bucket blocked")
	}
	if result := limiter.Consume("chat:user:a", limit); !result.Allowed {
		t.Fatal("chat bucket was affected by the search bucket")
	}
}

func TestWindowRollsOver(t *testing.T) {
	limiter := NewLimiter()
	defer limiter.Close()

	limit := rule(1, 20*time.Millisecond)

	if result := limiter.Consume("bucket:user:a", limit); !result.Allowed {
		t.Fatal("first hit was blocked")
	}
	if result := limiter.Consume("bucket:user:a", limit); result.Allowed {
		t.Fatal("second hit was allowed inside the window")
	}

	time.Sleep(30 * time.Millisecond)

	if result := limiter.Consume("bucket:user:a", limit); !result.Allowed {
		t.Fatal("the caller was still blocked after the window rolled over")
	}
}

func TestConfiguredLimitsArePositive(t *testing.T) {
	rest := map[string]config.RateLimitRule{
		"me":            config.REST.Me,
		"createProfile": config.REST.CreateProfile,
		"createRoom":    config.REST.CreateRoom,
		"joinRoom":      config.REST.JoinRoom,
		"getRoom":       config.REST.GetRoom,
		"leaveRoom":     config.REST.LeaveRoom,
		"musicSearch":   config.REST.MusicSearch,
		"musicVideo":    config.REST.MusicVideo,
		"musicPlaylist": config.REST.MusicPlaylist,
	}

	sockets := map[string]config.RateLimitRule{
		"roomJoin":     config.Socket.RoomJoin,
		"chatSend":     config.Socket.ChatSend,
		"playback":     config.Socket.Playback,
		"queue":        config.Socket.Queue,
		"hostTransfer": config.Socket.HostTransfer,
	}

	for name, entry := range rest {
		if entry.Limit <= 0 || entry.Window <= 0 {
			t.Fatalf("REST limit %s is %+v", name, entry)
		}
	}

	for name, entry := range sockets {
		if entry.Limit <= 0 || entry.Window <= 0 {
			t.Fatalf("socket limit %s is %+v", name, entry)
		}
	}
}

func TestRoomCodeGuessingStaysSlowerThanOneTryPerSecond(t *testing.T) {
	perSecond := float64(config.REST.JoinRoom.Limit) / config.REST.JoinRoom.Window.Seconds()
	if perSecond >= 1 {
		t.Fatalf("joinRoom allows %.2f tries per second", perSecond)
	}
}

func TestRetryAfterIsAtLeastOneSecond(t *testing.T) {
	if seconds := RetryAfterSeconds(1); seconds != 1 {
		t.Fatalf("RetryAfterSeconds(1) = %d", seconds)
	}
	if seconds := RetryAfterSeconds(1200); seconds != 2 {
		t.Fatalf("RetryAfterSeconds(1200) = %d", seconds)
	}
}
