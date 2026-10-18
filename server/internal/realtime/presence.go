package realtime

import (
	"sync"

	"github.com/mddanish004/jamzy/server/internal/rooms"
)

type presenceRoom struct {
	sockets map[string]map[string]struct{}
}

type presenceStore struct {
	mu    sync.Mutex
	rooms map[string]*presenceRoom
}

func newPresenceStore() *presenceStore {
	return &presenceStore{rooms: map[string]*presenceRoom{}}
}

func (p *presenceStore) room(roomCode string) *presenceRoom {
	entry, ok := p.rooms[roomCode]
	if !ok {
		entry = &presenceRoom{sockets: map[string]map[string]struct{}{}}
		p.rooms[roomCode] = entry
	}
	return entry
}

func (p *presenceStore) addSocket(roomCode string, userID string, socketID string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	sockets, ok := p.room(roomCode).sockets[userID]
	if !ok {
		sockets = map[string]struct{}{}
		p.rooms[roomCode].sockets[userID] = sockets
	}

	wasNew := len(sockets) == 0
	sockets[socketID] = struct{}{}

	return wasNew
}

func (p *presenceStore) removeSocket(roomCode string, userID string, socketID string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	entry, ok := p.rooms[roomCode]
	if !ok {
		return false
	}

	sockets, ok := entry.sockets[userID]
	if !ok {
		return false
	}

	delete(sockets, socketID)
	nowOffline := len(sockets) == 0

	if nowOffline {
		delete(entry.sockets, userID)
	}

	if len(entry.sockets) == 0 {
		delete(p.rooms, roomCode)
	}

	return nowOffline
}

func (p *presenceStore) isOnline(roomCode string, userID string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	entry, ok := p.rooms[roomCode]
	if !ok {
		return false
	}

	return len(entry.sockets[userID]) > 0
}

func (p *presenceStore) snapshot(roomCode string, members []rooms.PresenceMember, hostUserID string) PresenceSnapshot {
	p.mu.Lock()
	defer p.mu.Unlock()

	entry := p.rooms[roomCode]

	snapshot := PresenceSnapshot{
		Members:    make([]PresenceMember, 0, len(members)),
		HostUserID: hostUserID,
	}

	for _, member := range members {
		online := false
		if entry != nil {
			online = len(entry.sockets[member.UserID]) > 0
		}

		snapshot.Members = append(snapshot.Members, PresenceMember{
			UserID:   member.UserID,
			Username: member.Username,
			Online:   online,
		})
	}

	return snapshot
}
