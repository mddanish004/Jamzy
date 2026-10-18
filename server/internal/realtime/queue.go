package realtime

import "sync"

const maxQueueSize = 50

type queueStore struct {
	mu     sync.Mutex
	queues map[string][]QueueItem
}

func newQueueStore() *queueStore {
	return &queueStore{queues: map[string][]QueueItem{}}
}

func (q *queueStore) items(roomCode string) []QueueItem {
	queued, ok := q.queues[roomCode]
	if !ok {
		queued = []QueueItem{}
		q.queues[roomCode] = queued
	}
	return queued
}

func copyItems(items []QueueItem) []QueueItem {
	copied := make([]QueueItem, len(items))
	copy(copied, items)
	return copied
}

func (q *queueStore) enqueue(roomCode string, item QueueItem) []QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	items := q.items(roomCode)
	if len(items) < maxQueueSize {
		items = append(items, item)
		q.queues[roomCode] = items
	}

	return copyItems(items)
}

func (q *queueStore) prepend(roomCode string, item QueueItem) []QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	items := q.items(roomCode)
	if len(items) < maxQueueSize {
		items = append([]QueueItem{item}, items...)
		q.queues[roomCode] = items
	}

	return copyItems(items)
}

func (q *queueStore) dequeue(roomCode string) (QueueItem, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	items, ok := q.queues[roomCode]
	if !ok || len(items) == 0 {
		return QueueItem{}, false
	}

	next := items[0]
	q.queues[roomCode] = items[1:]

	return next, true
}

func (q *queueStore) remove(roomCode string, index int) []QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	items := q.items(roomCode)
	if index >= 0 && index < len(items) {
		items = append(items[:index], items[index+1:]...)
		q.queues[roomCode] = items
	}

	return copyItems(items)
}

func (q *queueStore) reorder(roomCode string, fromIndex int, toIndex int) []QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	items := q.items(roomCode)
	length := len(items)

	if fromIndex < 0 || fromIndex >= length || toIndex < 0 || toIndex >= length || fromIndex == toIndex {
		return copyItems(items)
	}

	moved := items[fromIndex]
	items = append(items[:fromIndex], items[fromIndex+1:]...)
	items = append(items[:toIndex], append([]QueueItem{moved}, items[toIndex:]...)...)
	q.queues[roomCode] = items

	return copyItems(items)
}

func (q *queueStore) get(roomCode string) []QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	return copyItems(q.items(roomCode))
}

func (q *queueStore) clear(roomCode string) []QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	delete(q.queues, roomCode)

	return []QueueItem{}
}
