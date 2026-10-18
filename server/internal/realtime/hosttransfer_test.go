package realtime

import (
	"testing"
	"time"

	"github.com/mddanish004/jamzy/server/internal/rooms"
)

func TestPendingTransferFiresOnceAndForgetsItself(t *testing.T) {
	store := newPendingTransferStore()

	fired := make(chan string, 1)

	store.arm("ABC123", "user-1", 20*time.Millisecond, func() {
		fired <- "user-1"
	})

	if !store.isPending("ABC123") {
		t.Fatal("a freshly armed transfer is not pending")
	}

	select {
	case hostUserID := <-fired:
		if hostUserID != "user-1" {
			t.Fatalf("fired for %q", hostUserID)
		}
	case <-time.After(time.Second):
		t.Fatal("the transfer never fired")
	}

	if store.isPending("ABC123") {
		t.Fatal("a fired transfer is still pending")
	}
}

func TestCancelReportsWhoTheTimerWasArmedFor(t *testing.T) {
	store := newPendingTransferStore()

	store.arm("ABC123", "user-1", time.Minute, func() {
		t.Error("a cancelled transfer fired")
	})

	cancelledFor, ok := store.cancel("ABC123")
	if !ok || cancelledFor != "user-1" {
		t.Fatalf("cancel = %q, %v", cancelledFor, ok)
	}

	if store.isPending("ABC123") {
		t.Fatal("the entry survived its own cancellation")
	}
	if _, ok := store.cancel("ABC123"); ok {
		t.Fatal("cancelling twice reported a pending transfer")
	}
}

func TestArmingAgainReplacesThePendingTransfer(t *testing.T) {
	store := newPendingTransferStore()

	fired := make(chan struct{}, 2)

	store.arm("ABC123", "user-1", time.Minute, func() { fired <- struct{}{} })
	store.arm("ABC123", "user-2", 20*time.Millisecond, func() { fired <- struct{}{} })

	select {
	case <-fired:
	case <-time.After(time.Second):
		t.Fatal("the replacement transfer never fired")
	}

	time.Sleep(20 * time.Millisecond)

	if len(fired) != 0 {
		t.Fatal("the replaced transfer also fired")
	}
}

func TestPickSuccessorSkipsTheDepartingHostAndOfflineMembers(t *testing.T) {
	hub := &Hub{presence: newPresenceStore()}

	hub.presence.addSocket("ABC123", "user-2", "socket-2")

	roomData := &rooms.PresenceData{
		RoomID:     "room-1",
		HostUserID: "user-1",
		Members: []rooms.PresenceMember{
			{UserID: "user-1", Username: "host"},
			{UserID: "user-3", Username: "offline"},
			{UserID: "user-2", Username: "successor"},
		},
	}

	successor, ok := hub.pickSuccessor("ABC123", roomData, "user-1")
	if !ok {
		t.Fatal("no successor was picked")
	}
	if successor.UserID != "user-2" {
		t.Fatalf("successor = %q", successor.UserID)
	}
}

func TestPickSuccessorGivesUpWhenNobodyElseIsOnline(t *testing.T) {
	hub := &Hub{presence: newPresenceStore()}

	roomData := &rooms.PresenceData{
		Members: []rooms.PresenceMember{
			{UserID: "user-1", Username: "host"},
			{UserID: "user-3", Username: "offline"},
		},
	}

	if successor, ok := hub.pickSuccessor("ABC123", roomData, "user-1"); ok {
		t.Fatalf("picked %q", successor.UserID)
	}
}

func TestPresenceSnapshotMarksOnlyConnectedMembersOnline(t *testing.T) {
	store := newPresenceStore()

	members := []rooms.PresenceMember{
		{UserID: "user-1", Username: "one"},
		{UserID: "user-2", Username: "two"},
	}

	if !store.addSocket("ABC123", "user-1", "socket-1") {
		t.Fatal("the first socket for a user was not treated as new")
	}
	if store.addSocket("ABC123", "user-1", "socket-2") {
		t.Fatal("a second socket for the same user was treated as new")
	}

	snapshot := store.snapshot("ABC123", members, "user-1")
	if len(snapshot.Members) != 2 {
		t.Fatalf("snapshot has %d members", len(snapshot.Members))
	}
	if !snapshot.Members[0].Online {
		t.Fatal("user-1 is not marked online")
	}
	if snapshot.Members[1].Online {
		t.Fatal("user-2 is marked online without a socket")
	}
	if snapshot.HostUserID != "user-1" {
		t.Fatalf("host = %q", snapshot.HostUserID)
	}

	if store.removeSocket("ABC123", "user-1", "socket-1") {
		t.Fatal("a user with another socket open went offline")
	}
	if !store.removeSocket("ABC123", "user-1", "socket-2") {
		t.Fatal("the user stayed online after their last socket dropped")
	}
	if store.isOnline("ABC123", "user-1") {
		t.Fatal("the user is still online")
	}
}

func TestQueueStoreBoundsAndMutations(t *testing.T) {
	store := newQueueStore()

	if len(store.get("ABC123")) != 0 {
		t.Fatal("a fresh queue is not empty")
	}

	first := QueueItem{VideoID: "aaaaaaaaaaa"}
	second := QueueItem{VideoID: "bbbbbbbbbbb"}
	third := QueueItem{VideoID: "ccccccccccc"}

	store.enqueue("ABC123", first)
	store.enqueue("ABC123", second)
	store.prepend("ABC123", third)

	queue := store.get("ABC123")
	if len(queue) != 3 || queue[0].VideoID != "ccccccccccc" {
		t.Fatalf("queue = %+v", queue)
	}

	store.reorder("ABC123", 0, 2)
	queue = store.get("ABC123")
	if queue[2].VideoID != "ccccccccccc" {
		t.Fatalf("reorder produced %+v", queue)
	}

	store.remove("ABC123", -1)
	store.remove("ABC123", 99)
	if len(store.get("ABC123")) != 3 {
		t.Fatal("an out-of-range index changed the queue")
	}

	next, ok := store.dequeue("ABC123")
	if !ok || next.VideoID != "aaaaaaaaaaa" {
		t.Fatalf("dequeue = %+v, %v", next, ok)
	}

	store.clear("ABC123")
	if len(store.get("ABC123")) != 0 {
		t.Fatal("clear left items behind")
	}
	if _, ok := store.dequeue("ABC123"); ok {
		t.Fatal("dequeue on an empty queue reported an item")
	}
}

func TestQueueStoreRefusesToGrowPastItsCeiling(t *testing.T) {
	store := newQueueStore()

	for i := 0; i < maxQueueSize+10; i++ {
		store.enqueue("ABC123", QueueItem{VideoID: "aaaaaaaaaaa"})
	}

	if len(store.get("ABC123")) != maxQueueSize {
		t.Fatalf("queue length = %d", len(store.get("ABC123")))
	}
}
